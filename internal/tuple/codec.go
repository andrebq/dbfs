package tuple

import (
	"bytes"
	"encoding/json"
	"errors"
	"sort"

	"github.com/vmihailenco/msgpack/v4"
)

type (
	// Named represents a tuple whose items are pairs of tuples
	// where the first pair is the name of the field
	//
	// and the second pair is the value of that field
	Named struct {
		pairs []Indexed
	}

	// Index represents a tuple of positional dependendent items
	Indexed []interface{}

	Pairs struct {
		pairs []Indexed
	}

	Content interface {
		anchor()
	}
)

func MarshalBinary(c Content) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := msgpack.NewEncoder(buf)
	err := enc.Encode(c)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p Pairs) Named() Named {
	p.pairs = nil
	sort.Sort(p)
	return Named{pairs: p.pairs}
}

func (p Pairs) Add(name string, value interface{}) Pairs {
	p.pairs = append(p.pairs, Indexed{name, value})
	return p
}

func (p Pairs) Len() int {
	return len(p.pairs)
}

func (p Pairs) Less(a, b int) bool {
	return p.pairs[a][0].(string) < p.pairs[b][0].(string)
}

func (p Pairs) Swap(a, b int) {
	p.pairs[a], p.pairs[b] = p.pairs[b], p.pairs[a]
}

func (n Named) anchor()   {}
func (i Indexed) anchor() {}

func (n Named) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Pairs []Indexed `json:"pairs"`
	}{
		Pairs: n.pairs,
	})
}

func (n *Named) UnmarshalJSON(val []byte) error {
	tmp := struct {
		Pairs []Indexed `json:"pairs"`
	}{}
	err := json.Unmarshal(val, &tmp)
	if err != nil {
		return err
	}
	n.pairs = tmp.Pairs
	return nil
}

func (n Named) EncodeMsgpack(enc *msgpack.Encoder) error {
	err := enc.EncodeArrayLen(len(n.pairs))
	if err != nil {
		return err
	}
	for _, p := range n.pairs {
		err = enc.Encode(p)
		if err != nil {
			return err
		}
	}
	return err
}

func (n *Named) DecodeMsgpack(dec *msgpack.Decoder) error {
	sz, err := dec.DecodeArrayLen()
	if err != nil {
		return nil
	}
	if sz < 0 {
		return errors.New("negative length")
	}
	n.pairs = make([]Indexed, sz)
	for i := range n.pairs {
		var pair Indexed
		err = dec.Decode(&pair)
		if err != nil {
			n.pairs = nil
			return err
		}
		n.pairs[i] = pair
	}
	return nil
}

func (i Indexed) MarshalJSON() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Indexed) UnmarshalJSON(value []byte) error {
	return json.Unmarshal(value, i)
}

func (i Indexed) EncodeMsgpack(enc *msgpack.Encoder) error {
	err := enc.EncodeArrayLen(len(i))
	if err != nil {
		return err
	}
	for _, item := range i {
		err = enc.Encode(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Indexed) DecodeMsgpack(dec *msgpack.Decoder) error {
	sz, err := dec.DecodeArrayLen()
	if err != nil {
		return err
	}
	if sz < 0 {
		return errors.New("negative length")
	}
	*i = make([]interface{}, sz)
	for idx := range *i {
		val, err := dec.DecodeInterface()
		if err != nil {
			*i = nil
			return err
		}
		(*i)[idx] = val
	}
	return nil
}
