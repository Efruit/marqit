// Package broker used to do something. That something is no longer done.
package broker

import (
	"github.com/Efruit/marqit/exchange"
	"github.com/Efruit/marqit/trader"
)

type Broker interface {
	trader.Trader
	Add(exchange.Order)
	Summary(exchange.SumLen) exchange.Summary
	Place(exchange.Order) exchange.OrderID
	Order(exchange.OrderID) exchange.Order
}
