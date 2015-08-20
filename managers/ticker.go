package manager

import (
	"../ticker"
)

type Ticker interface {
	RegisterTicker(string, ticker.Ticker) // Add a ticker
	Tickers() []ticker.Ticker
}
