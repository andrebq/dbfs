package blob

import (
	"encoding/json"

	"github.com/andrebq/dbfs/internal/tuple"
)

func (c Chunk) MarshalYAML() (interface{}, error) {
	return struct {
		Start int64  `yaml:"start"`
		End   int64  `yaml:"stop"`
		Size  int64  `yaml:"size"`
		Ref   string `yaml:"ref"`
	}{
		Start: c.Start,
		End:   c.End,
		Size:  int64(c.Size),
		Ref:   c.Ref.String(),
	}, nil
}

func (c Chunk) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Start int64  `json:"start"`
		End   int64  `json:"stop"`
		Size  int64  `json:"size"`
		Ref   string `json:"ref"`
	}{
		Start: c.Start,
		End:   c.End,
		Size:  int64(c.Size),
		Ref:   c.Ref.String(),
	})
}

func (c Chunk) MarshalBinary() ([]byte, error) {
	return tuple.MarshalBinary(tuple.Pairs{}.Add("start", c.Start).Add("end", c.End).Add("size", c.Size).Add("ref", c.Ref).Named())
}

func (t Tree) MarshalBinary() ([]byte, error) {
	tup := tuple.Pairs{}.Add("leaves", t.Leaves).Add("branches", t.Branches).Named()
	return tuple.MarshalBinary(tup)
}
