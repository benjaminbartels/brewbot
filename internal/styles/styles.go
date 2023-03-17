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
	styles map[string]Style
}

type Style struct {
	Name                      string `json:"name"`
	Number                    string `json:"number"`
	Category                  string `json:"category"`
	CategoryNumber            string `json:"categorynumber"`
	OverallImpression         string `json:"overallimpression"`
	Aroma                     string `json:"aroma"`
	Appearance                string `json:"appearance"`
	Flavor                    string `json:"flavor"`
	Mouthfeel                 string `json:"mouthfeel"`
	Comments                  string `json:"comments"`
	History                   string `json:"history"`
	CharacteristicIngredients string `json:"characteristicingredients"`
	StyleComparison           string `json:"stylecomparison"`
	IBUMin                    string `json:"ibumin"`
	IBUMax                    string `json:"ibumax"`
	OGMin                     string `json:"ogmin"`
	OGMax                     string `json:"ogmax"`
	FGMin                     string `json:"fgmin"`
	FGMax                     string `json:"fgmax"`
	ABVMin                    string `json:"abvmin"`
	ABVMax                    string `json:"abvmax"`
	SRMMin                    string `json:"srmmin"`
	SRMMax                    string `json:"srmmax"`
	CommercialExamples        string `json:"commercialexamples"`
	Tags                      string `json:"tags"`
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
	slice := make([]Style, 0, len(s.styles))
	for _, style := range s.styles {
		slice = append(slice, style)
	}

	return slice[rand.Intn(len(slice))]
}

func (s *StyleSource) Get(ctx context.Context, number string) *Style {
	style, ok := s.styles[number]
	if !ok {
		return nil
	}

	return &style
}
