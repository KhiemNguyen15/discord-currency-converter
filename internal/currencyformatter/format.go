package currencyformatter

import (
	"encoding/csv"
	"math"
	"os"
	"strconv"

	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

var tagMap = make(map[language.Tag]currency.Unit)

func LoadTagMap() {
	languageTags := getLanguageTags()

	for _, tag := range languageTags {
		cur, _ := currency.FromTag(tag)
		tagMap[tag] = cur
	}
}

func FormatCurrency(isoCode string, value float64) string {
	cur, _ := currency.ParseISO(isoCode)
	langTag, _ := getKey(tagMap, cur)

	scale, incCents := currency.Cash.Rounding(cur)
	incFloat := math.Pow10(-scale) * float64(incCents)
	incFmt := strconv.FormatFloat(incFloat, 'f', scale, 64)
	dec := number.Decimal(value, number.Scale(scale), number.IncrementString(incFmt))

	p := message.NewPrinter(langTag)

	return p.Sprintf("%3v%v %v", currency.Symbol(cur), dec, cur)
}

func getKey(
	m map[language.Tag]currency.Unit,
	value currency.Unit,
) (key language.Tag, ok bool) {
	for k, v := range m {
		if v == value {
			key = k
			ok = true
			return
		}
	}
	return
}

func getLanguageTags() []language.Tag {
	file, _ := os.Open("bcp47.csv")
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	data, _ := reader.ReadAll()

	var languageTags []language.Tag
	for _, row := range data {
		for _, col := range row {
			languageTags = append(languageTags, language.MustParse(col))
		}
	}

	return languageTags
}
