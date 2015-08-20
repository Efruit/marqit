package irrational

import (
	"github.com/Efruit/marqit/bank"
	xchg "github.com/Efruit/marqit/exchange"
	"github.com/Efruit/marqit/managers"
	"github.com/Efruit/marqit/ticker"
	"github.com/Efruit/marqit/trader"
	"github.com/inconshreveable/log15"
	"github.com/oleiade/lane"
	"math/rand"
	"time"
)

var NewIrrational manager.TraderMaker = func(m manager.MiniMarket) trader.Trader {
	return &Irrational{mkt: m, log: log15.New("module", "irrational")}
}

var _ trader.AuctionBidder = &Irrational{}
var _ trader.AuctionAsker = &Irrational{}

type Irrational struct {
	log  log15.Logger
	mode xchg.Mode
	id   xchg.LicenseID
	act  bank.Token
	mkt  manager.MiniMarket
	cl   chan bool
	cl2  bool
	tq   *lane.Queue
	lp   map[xchg.Symbol]float32
	ls   map[xchg.Symbol]uint64
	tra  map[xchg.AuctionID]xchg.TradeStatus
	trb  map[xchg.BidID]xchg.TradeStatus
	lo   float32
}

func (i *Irrational) Init(license xchg.LicenseID) {
	i.log.Info("Licensed", "license", license)
	i.log = log15.New("license", license)
	i.id = license
	//i.tc = make(chan Tick)
	i.cl = make(chan bool)
	i.tq = lane.NewQueue()
	i.lp = make(map[xchg.Symbol]float32)
	i.ls = make(map[xchg.Symbol]uint64)
	i.tra = make(map[xchg.AuctionID]xchg.TradeStatus)
	i.trb = make(map[xchg.BidID]xchg.TradeStatus)
	i.act = i.mkt.Bank().Open()
	i.log.Info("Bank account opened", "token", i.act)
	go i.Loop()
}

func (i Irrational) GetID() xchg.LicenseID   { return i.id }
func (i Irrational) GetAcct() bank.AccountID { return i.act.Account }

func (i *Irrational) Status(chg xchg.Status) {
	switch chg {
	case xchg.S_OPEN:
		mr, mn, mode := i.mkt.RunID()
		i.log.Info("Open", "run", mr.String(), "number", mn, "mode", mode)
		i.mode = mode
		i.cl2 = true
		i.cl <- true
		if b, err := i.mkt.Bank().Balance(i.act); err != bank.BE_OK {
			i.log.Error("Bank error", "error", err)
		} else {
			i.lo = b
		}
	case xchg.S_CLOSE:
		i.cl2 = false
		i.cl <- false
		if b, err := i.mkt.Bank().Balance(i.act); err != bank.BE_OK {
			i.log.Error("Bank error", "error", err)
		} else {
			i.log.Info("Closing balance", "amt", i.lo-b)
		}
	case xchg.S_PAUSE:
		i.cl2 = false
		i.cl <- false
	case xchg.S_RESUME:
		i.cl2 = true
		i.cl <- true
	case xchg.S_QUOTA_REACHED:
		i.cl2 = false
		i.cl <- false
	}
}

func (i *Irrational) Loop() {
	open := i.cl2
	for {
		if !open {
			open = <-i.cl
			continue
		}

		time.Sleep(time.Duration(rand.Intn(800)) * time.Millisecond)

		if open = i.cl2; !open {
			continue
		}
	operate:
		switch rand.Intn(2) {
		case 0: // Buy
			stocks := i.mkt.List()
			stl := len(stocks)
			stockn := rand.Intn(stl - 1)

			stock := stocks[stockn].Symbol

			price := i.lp[stock] + rand.Float32()
			if rand.Intn(2) == 1 || price <= 2 {
				price = price + rand.Float32()
			} else {
				price = price - rand.Float32()
			}
			i.log.Info("Searching for auction", "symbol", stock, "price", price)
			for _, v := range i.mkt.Auctions() {
				if v.Stock == stock {
					if b, err := i.mkt.Bank().Balance(i.act); err != bank.BE_OK {
						i.log.Error("Bank error", "error", err)
						break operate

					} else {
						if b <= (price * float32(v.Number)) {
							// i.log.Info("Can't afford", "symbol", stock, "price", price, "shares", v.Number)
							continue
						}
					}
					// i.log.Info("Bidding on auction", "auction", v.ID)
					b := xchg.Bid{
						Bidder: i.id,
						ID:     xchg.BidID{Auction: v.ID},
						Number: v.Number - uint64(rand.Int31n(int32(v.Number))),
						Price:  price,
					}
					o, err := i.mkt.Bid(b)
					if err != nil {
						i.log.Error("Bidding error", "error", err.Error(), "bidid", b.ID)
						break operate
					}
					i.log.Info("Bid placed", "id", o, "stock", stock, "price", price)
					i.trb[o] = xchg.TS_SUBMITTED
					break
				}
			}
		case 1: // Sell
			a, err := i.mkt.Bank().GetAssets(i.act)
			if err != bank.BE_OK {
				i.log.Error("Bank error", "error", err)

				break operate
			}

			for k, v := range a {
				price := i.lp[k] + rand.Float32()
				if rand.Intn(2) == 1 || price <= 2 {
					price = price + rand.Float32()
				} else {
					price = price - rand.Float32()
				}
				b := xchg.Auction{
					Seller: i.id,
					Stock:  k,
					Number: v - uint64(rand.Int31n(int32(v))),
					Ask:    price,
					Length: time.Duration(rand.Intn(10000)+500) * time.Millisecond,
				}
				i.log.Info("Asking", "symbol", k, "number", b.Number, "price", price)
				o, err := i.mkt.Ask(b)
				if err != nil {
					i.log.Error("Asking error", "error", err.Error())
					break operate
				}
				i.tra[o] = xchg.TS_SUBMITTED
			}
		case 2:
			// Do nothing.
		}
		open = i.cl2
	}
}

func (i *Irrational) ConfirmBid(b xchg.BidID) xchg.TradeStatus {
	return i.trb[b]
}

func (i *Irrational) ConfirmAuction(b xchg.AuctionID) xchg.TradeStatus {
	return i.tra[b]
}
func (i *Irrational) Update(t ticker.Tick) {
	i.lp[t.Symbol] = t.LastSale
	i.ls[t.Symbol] = t.LastSize
}

func (i *Irrational) NewAuction(xchg.Auction) {} // TODO

type dts [2]xchg.TradeStatus

var bidresponses = map[dts](func(i *Irrational, a xchg.Auction, b xchg.Bid, t dts, r xchg.Reason)){
	{xchg.TS_SUBMITTED, xchg.TS_ACKNOWLEDGED}: func(i *Irrational, a xchg.Auction, b xchg.Bid, s dts, r xchg.Reason) {
		other := i.mkt.Licensee(a.Seller)
		i.log.Info("Paying", "other", a.Seller, "ammount", b.Price)
		i.mkt.Bank().Transfer(i.act, b.Price, other.GetAcct())
		i.trb[b.ID] = xchg.TS_PAID
		other.(trader.AuctionAsker).UpdateAuction(a.ID, xchg.TS_PAID, xchg.R_OK)
	},
	{xchg.TS_ACKNOWLEDGED, xchg.TS_PAID}: func(*Irrational, xchg.Auction, xchg.Bid, dts, xchg.Reason) {},
	{xchg.TS_SUBMITTED, xchg.TS_ABORTED}: func(*Irrational, xchg.Auction, xchg.Bid, dts, xchg.Reason) {}, // Seller canceled.

}

func (i *Irrational) UpdateBid(b xchg.BidID, ts xchg.TradeStatus, r xchg.Reason) {
	if v, ok := i.trb[b]; !ok {
		// not our bid
		// error
		i.log.Error("Invalid bid updated", "bid", b, "me", v, "other", ts, "reason", "Bid not submitted")
	} else {
		// our bid
		a, err := i.mkt.Auction(b.Auction)
		if err != nil {
			i.log.Error("Invalid bid updated", "bid", b, "me", v, "other", ts, "reason", r)
		}
		bb := a.Bids[b.Bid]
		f, ok := bidresponses[dts{v, ts}]
		if !ok {
			// bad transition
			i.log.Error("Bad transition", "me", v, "other", ts, "reason", r)
		} else {
			i.log.Info("Transitioning", "bid", b, "me", v, "other", ts)
			f(i, a, bb, dts{v, ts}, r)
		}

	}
}

func (i *Irrational) NewBid(xchg.AuctionID, xchg.Bid)                             {} // TODO
func (i *Irrational) UpdateAuction(xchg.AuctionID, xchg.TradeStatus, xchg.Reason) {} // TODO
