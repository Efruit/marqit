package broker

import (
	"../exchange"
	"../trader"
)

type Broker interface {
	trader.Trader
	Add(exchange.Order)
	Summary(exchange.SumLen) exchange.Summary
	Place(exchange.Order) exchange.OrderID
	Order(exchange.OrderID) exchange.Order
}
