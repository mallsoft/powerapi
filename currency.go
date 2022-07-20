package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Currency struct {
	Success   bool               `json:"success"`
	Timestamp int                `json:"timestamp"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
}

var ConversionRate float64

func updateCurrency() {
	token := os.Getenv("CURRENCY_API_KEY")

	url := "https://api.apilayer.com/exchangerates_data/latest?symbols=NOK&base=EUR"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", token)

	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var currency Currency
	err = json.Unmarshal(body, &currency)
	if err != nil {
		panic(err)
	}

	ConversionRate = currency.Rates["NOK"]
}
