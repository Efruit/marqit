package manager

import (
	"github.com/Efruit/marqit/ticker"
)

type Ticker interface {
	RegisterTicker(string, ticker.Ticker) // Add a ticker
	Tickers() []ticker.Ticker
}
