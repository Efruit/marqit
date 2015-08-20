package main

import (
	"./bank"
	bank1 "./bank/v1"
	"./broker/irrational"
	"./exchange"
	"./managers"
	market1 "./market/v1"
	"./trader/netplayer"
	// "code.google.com/p/gocircuit/src/circuit/kit/debug"
	"github.com/inconshreveable/log15"
	"math/rand"
	"time"
)

func main() {
	log15.Info("Starting Marqit")
	// debug.InstallCtrlCPanic()

	Mkt := market1.NewMarket1()
	Bnk, AKey := bank1.NewBank1()
	log15.Info("Bank created", "admin", AKey)

	Mkt.RegisterBank("the bank", Bnk)
	Mkt.RegisterTrader("t.irrational", irrational.NewIrrational)
	Mkt.RegisterTrader("t.netplayer", netplayer.NewNetPlayer)
	Mkt.SetCount(manager.TraderCfg{Type: "t.irrational", Count: 10}, manager.TraderCfg{Type: "t.netplayer", Count: 1})
	Mkt.SpawnTraders()

	NStocks := 2 + rand.Intn(30)
	log15.Info("Creating stocks", "n", NStocks)
	for i := 0; i <= NStocks; i++ {
		symbol := ""
		sl := 1 + rand.Intn(3)
		for n := 0; n <= sl; n++ {
			symbol += string(byte(65 + rand.Intn(26)))
		}

		s := exchange.Stock{
			Symbol:  exchange.Symbol(symbol),
			Shares:  uint(rand.Int31n(100000)),
			Trading: true,
			Mode:    exchange.M_AUCTION,
		}
		log15.Info("Adding", "symbol", s.Symbol, "shares", s.Shares)
		Mkt.AddStock(s)
	}

	for _, v := range Mkt.List() {
		recip := 2 * (1 + rand.Intn(5))
		log15.Info("Spreading initial", "symbol", v.Symbol, "volume", v.Shares, "recipients", recip)
		NTraders := len(Mkt.Traders())
		for i := 0; i <= recip; i++ {
			act := rand.Intn(NTraders - 1)
			ba := Mkt.Traders()[act].GetAcct()
			log15.Info("Initial", "recipient", ba, "count", uint64(int(v.Shares)/recip))
			Bnk.SetAsset(ba, v.Symbol, uint64(int(v.Shares)/recip))
		}
	}

	for k := range Mkt.Traders() {
		act := Mkt.Traders()[k].GetAcct()
		amt := float32(rand.Int31n(100000))
		log15.Info("Granting funds", "recipient", act, "ammount", amt)
		Bnk.Transfer(bank.Token{Account: bank.BA_Start, PIN: AKey}, amt, act)
	}

	Mkt.Start(10*time.Second, 15*time.Second, 5)
}
