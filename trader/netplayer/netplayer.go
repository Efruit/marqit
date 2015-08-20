package netplayer

import (
	"github.com/Efruit/marqit/bank"
	xchg "github.com/Efruit/marqit/exchange"
	"github.com/Efruit/marqit/managers"
	"github.com/Efruit/marqit/trader"
	"github.com/inconshreveable/log15"
)

var NpLog = log15.New("module", "netplayer")

var NewNetPlayer manager.TraderMaker = func(m manager.MiniMarket) trader.Trader {
	return &NetPlayer{Market: m}
}

type NetPlayer struct {
	Market  manager.MiniMarket
	License xchg.LicenseID
	Account bank.Token
	trx     map[xchg.OrderID]xchg.TradeStatus
}

var _ trader.Trader = &NetPlayer{}

func (n *NetPlayer) Init(l xchg.LicenseID) {
	bank := n.Market.Bank()
	if bank == nil {
		panic("netplayer: bank == nil")
	}
	n.License = l
	n.Account = bank.Open()
	NpLog.Info("Licensed", "license", n.License)
	NpLog.Info("Bank account opened", "account", n.Account)
}

func (n NetPlayer) GetID() xchg.LicenseID                   { return n.License }
func (n NetPlayer) GetAcct() bank.AccountID                 { return n.Account.Account }
func (n *NetPlayer) Status(xchg.Status)                     {}
func (n NetPlayer) Confirm(o xchg.OrderID) xchg.TradeStatus { return n.trx[o] }
