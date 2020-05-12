package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const cacheTTL = 5 // seconds

// Get quote from big list of quotes
func getQuote(pairs []currencyPair, pair string) (value, bid, ask float64) {
	for _, v := range pairs {
		if v.Symbol == pair {
			ask, _ = strconv.ParseFloat(v.MinAsk, 64)
			bid, _ = strconv.ParseFloat(v.MaxBid, 64)
			value = (ask + bid) / 2
		}
	}

	return value, bid, ask
}

var cacheData []byte
var cacheTs int64

func getData() (result string, err error) {
	data := []byte{}
	if time.Now().Unix()-cacheTTL < cacheTs {
		data = cacheData
	} else {
		response, err := http.Get("https://api.livecoin.net/exchange/maxbid_minask")
		if nil != err {
			return "", err
		}

		data, err = ioutil.ReadAll(response.Body)
		if nil != err {
			return "", err
		}
	}
	var pairs currencyPairs

	{
		err := json.Unmarshal(data, &pairs)
		if nil != err {
			return "", err
		}
	}

	BtcUsd, _, _ := getQuote(pairs.CurrencyPairs, "BTC/USD")
	EthUsd, _, _ := getQuote(pairs.CurrencyPairs, "ETH/USD")
	OtnBtc, OtnBtcBid, OtnBtcAsk := getQuote(pairs.CurrencyPairs, "OTN/BTC")
	OtnEth, OtnEthBid, OtnEthAsk := getQuote(pairs.CurrencyPairs, "OTN/ETH")
	OtnUsd, _, _ := getQuote(pairs.CurrencyPairs, "OTN/USD")

	result += fmt.Sprintf("\nBTC/USD = $%.3f", BtcUsd)
	result += fmt.Sprintf("\nETH/USD = $%.3f", EthUsd)
	result += fmt.Sprintf("\nOTN/BTC = $%.3f (%.6f - %.6f)", OtnBtc*BtcUsd, OtnBtcBid, OtnBtcAsk)
	result += fmt.Sprintf("\nOTN/ETH = $%.3f (%.6f - %.6f)", OtnEth*EthUsd, OtnEthBid, OtnEthAsk)
	result += fmt.Sprintf("\nOTN/USD = $%.3f", OtnUsd)

	return result, nil
}
