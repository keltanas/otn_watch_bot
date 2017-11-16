package main

type currencyPairs struct {
	CurrencyPairs []currencyPair `json:"currencyPairs"`
}

type currencyPair struct {
	Symbol string `json:"symbol"`
	MaxBid string `json:"maxBid"`
	MinAsk string `json:"minAsk"`
}
