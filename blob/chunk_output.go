package blob

import "encoding/json"

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
