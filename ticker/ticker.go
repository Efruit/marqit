package ticker

import (
	"github.com/Efruit/marqit/exchange"
)

type Ticker interface {
	Update(exchange.Stock)
	News(Article)
	GetTick() []Tick
	GetNews(exchange.Symbol) []Article
	Register(Watcher)
}

type Tick struct {
	Symbol exchange.Symbol

	Bid     float32
	BidSize uint64

	Ask     float32
	AskSize uint64

	LastSale float32
	LastSize uint64

	QuoteAt exchange.Stamp
	TradeAt exchange.Stamp

	Volume uint
}

type Watcher interface {
	Update(Tick)
	News([]Article)
}

type Article struct {
	Mood   int
	Symbol exchange.Symbol
}
