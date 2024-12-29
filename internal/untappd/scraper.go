package untappd

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-errors/errors"
	"github.com/gocolly/colly"
)

const base = "https://untappd.com/v/"

type Patron struct {
	Name     string
	CheckIns int
	Rank     int
}

type Menu struct {
	Name  string
	Items []MenuItem
}

type MenuItem struct {
	Name    string
	Brewery string
	Style   string
	ABV     string
	IBU     string
}

func Scrape(path string) ([]Menu, []Patron, error) {
	c := colly.NewCollector()

	var rankCtr int

	menus := []Menu{}
	patrons := []Patron{}

	c.OnHTML(`.menu-section`, func(e *colly.HTMLElement) {
		menuName := e.DOM.Find("div.menu-section-header h4").Clone()
		menuName.Find("span").Remove()

		menu := Menu{
			Name:  strings.TrimSpace(menuName.Text()),
			Items: []MenuItem{},
		}

		e.ForEach("li.menu-item", func(_ int, el *colly.HTMLElement) {
			item := MenuItem{}
			item.Name = normalizeWhitespace(strings.TrimSpace(el.DOM.Find("h5 a").Text()))
			item.Brewery = strings.TrimSpace(el.DOM.Find("h6 span a").First().Text())
			item.Style = strings.TrimSpace(el.DOM.Find("h5 em").Text())
			abvAndIbu := strings.TrimSpace(el.DOM.Find("h6 span").Text())
			parts := strings.Split(abvAndIbu, "•")
			item.ABV = strings.TrimSpace(strings.Split(parts[0], " ")[0])
			item.IBU = "N/A"

			if len(parts) > 1 {
				ibuParts := strings.Split(parts[1], "IBU")
				if len(ibuParts) > 0 {
					item.IBU = strings.TrimSpace(ibuParts[0])
				}
			}

			menu.Items = append(menu.Items, item)
		})

		menus = append(menus, menu)
	})

	patronRegEx := regexp.MustCompile(`^(.*) \((\d+) check-ins\)$`)

	c.OnHTML(`a[data-href=":loyal/drinkers"]`, func(e *colly.HTMLElement) {
		title := e.Attr("original-title")
		if title == "" {
			title = e.Attr("title")
		}

		if title == "" {
			return
		}

		rankCtr++

		p := Patron{
			Rank: rankCtr,
		}

		matches := patronRegEx.FindStringSubmatch(title)
		if len(matches) == 3 {
			p.Name = matches[1]

			checkIns, err := strconv.Atoi(matches[2])
			if err != nil {
				panic(err)
			}

			p.CheckIns = checkIns

			patrons = append(patrons, p)
		}
	})

	u, err := url.JoinPath(base, path)
	if err != nil {
		return nil, nil, errors.WrapPrefix(err, "could not join path", 0)
	}

	if err := c.Visit(u); err != nil {
		return nil, nil, errors.WrapPrefix(err, "could not visit url", 0)
	}

	return menus, patrons, nil
}

func normalizeWhitespace(input string) string {
	re := regexp.MustCompile(`\s+`)

	return strings.TrimSpace(re.ReplaceAllString(input, " "))
}
