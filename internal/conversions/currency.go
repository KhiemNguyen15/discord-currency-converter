package conversions

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/viper"
)

type currencyData struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

func ConvertCurrency(from string, to string, value float64) (float64, error) {
	appID := viper.GetString("open_exchange_rates.app_id")

	data, err := getCurrencyData(appID)
	if err != nil {
		return 0, err
	}

	baseRate, found := data[from]
	if !found {
		return 0, nil
	}

	conversionRate, found := data[to]
	if !found {
		return 0, nil
	}

	return value / baseRate * conversionRate, nil
}

func getCurrencyData(appID string) (map[string]float64, error) {
	url := fmt.Sprintf(
		"https://openexchangerates.org/api/latest.json?app_id=%s",
		appID,
	)

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error while getting currency data: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code: %v", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error while reading response body: %v", err)
	}

	var data currencyData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("Error while unmarshalling JSON body to struct: %v", err)
	}

	return data.Rates, nil
}
