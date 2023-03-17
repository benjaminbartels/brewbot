package styles

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"

	"github.com/pkg/errors"
)

var _ StyleRepo = (*StyleSource)(nil)

type StyleSource struct {
	styles []Style
}

type Style struct {
	Name         string `json:"name"`
	Number       string `json:"number"`
	CategoryName string `json:"category"`
}

func NewStyleRepo(fileName string) (*StyleSource, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "could open %s", fileName)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrapf(err, "could read %s", fileName)
	}

	s := StyleSource{}

	err = json.Unmarshal(data, &s.styles)
	if err != nil {
		return nil, errors.Wrap(err, "could unmarshal styles")
	}

	return &s, nil
}

func (s *StyleSource) Random(ctx context.Context) Style {
	return s.styles[rand.Intn(len(s.styles))]
}
