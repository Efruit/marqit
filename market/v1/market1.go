package v1

import (
	"github.com/Efruit/marqit/bank"
	"github.com/Efruit/marqit/broker"
	x "github.com/Efruit/marqit/exchange"
	"github.com/Efruit/marqit/managers"
	"github.com/Efruit/marqit/market"
	"github.com/Efruit/marqit/ticker"
	"github.com/Efruit/marqit/trader"
	"github.com/inconshreveable/log15"
	"github.com/nu7hatch/gouuid"
	"gopkg.in/validator.v2"
	"math/rand"
	"time"
)

type Auctioner struct {
	m        *Market1
	auctions map[x.AuctionID]x.Auction
	old      []x.Auction
	tickers  map[x.AuctionID]*x.Timer
}

func (m *Auctioner) setup() {
	if m.auctions == nil {
		m.auctions = make(map[x.AuctionID]x.Auction)
	}
	if m.tickers == nil {
		m.tickers = make(map[x.AuctionID]*x.Timer)
	}
}

func (m *Auctioner) pauseAll() {
	for k := range m.tickers {
		m.tickers[k].Pause()
		MkLog.Info("Paused", "timer", k)
	}
}

func (m *Auctioner) resumeAll() {
	for k := range m.tickers {
		m.tickers[k].Resume()
		MkLog.Info("Resumed", "timer", k)
	}
}

func (m *Auctioner) Auctions() (a []x.Auction) {
	for _, v := range m.auctions {
		a = append(a, v)
	}
	return
}
func (m *Auctioner) Auction(a x.AuctionID) (x.Auction, error) {
	v, ok := m.auctions[a]
	if ok {
		return v, nil
	} else {
		for _, v := range m.old {
			if v.ID == a {
				return v, nil
			}
		}
	}
	return x.Auction{}, x.NoSuchEntity
}

func (m *Auctioner) Bid(b x.Bid) (x.BidID, error) {
	if !m.m.Status() {
		return x.BidID{}, x.NotTrading
	}
	err := validator.NewValidator().WithTag("place").Validate(b)
	if err != nil {
		return x.BidID{}, err
	}
	_, ok := m.m.Licensee(b.Bidder).(trader.AuctionBidder)
	if !ok {
		return x.BidID{}, x.UnqualifiedBidder
	}
	v, ok := m.auctions[b.ID.Auction]
	if !ok {
		return x.BidID{}, x.NoSuchEntity
	}

	b.Placed = time.Now()
	b.ID.Bid = uint64(len(v.Bids))
	v.Bids = append(v.Bids, b)
	m.auctions[b.ID.Auction] = v
	//TODO: Notify owner
	//TODO: Queue verification with Security
	return b.ID, nil
}

func (m *Auctioner) Ask(a x.Auction) (x.AuctionID, error) {
	if !m.m.Status() {
		return x.AuctionID{}, x.NotTrading
	}
	err := validator.NewValidator().WithTag("ask").Validate(a)
	if err != nil {
		return x.AuctionID{}, err
	}

	_, ok := m.m.Licensee(a.Seller).(trader.AuctionAsker)
	if !ok {
		return x.AuctionID{}, x.UnqualifiedAsker
	}

	aid, _ := uuid.NewV4()
	a.ID = x.AuctionID{*aid}
	m.auctions[a.ID] = a
	go func() {
		MkLog.Info("Starting keeper", "aid", a.ID, "time", a.Length)
		m.tickers[a.ID] = x.NewTimer(a.Length)
		select {
		case _, ok := <-m.tickers[a.ID].C:
			if !ok {
			} else {
				err_ := m.Finish(a.ID)
				if err_ != nil {
					MkLog.Error("Closing error", "aid", a.ID)
				}
			}
		}
		delete(m.tickers, a.ID)

	}()
	return a.ID, nil
}

func (m *Auctioner) Cancel(a x.AuctionID) error {
	v, ok := m.auctions[a]
	if !ok {
		return x.NoSuchEntity
	}
	v.Open = false
	//TODO: Queue verification with Security
	m.auctions[a] = v
	sent := map[x.LicenseID]struct{}{}
	for _, t := range v.Bids {
		if _, ok := sent[t.Bidder]; !ok {

			go func() {
				bidder, ok := m.m.Licensee(t.Bidder).(trader.AuctionBidder)
				if !ok {
					MkLog.Error("Error notifying Bidder", "license", v.Seller, "reason", "Bidder unqualified to recieve updates")
					return
				}
				bidder.UpdateBid(t.ID, x.TS_ABORTED, x.R_OK)
			}()
			sent[t.Bidder] = struct{}{}
		}
	}
	m.tickers[a].Stop()
	close(m.tickers[a].C)
	return nil
}
func (m *Auctioner) Finish(a x.AuctionID) error {
	v, ok := m.auctions[a]
	if !ok {
		return x.NoSuchEntity
	}
	v.Open = false
	//TODO: Notify watchers
	v.Winner = v.Win()
	if len(v.Bids) == 0 || v.Winner == (x.Bid{}) {
		MkLog.Info("Auction canceled", "id", a, "reason", "0 Bids or No winning bid")
	} else {
		go func() {
			asker, ok := m.m.Licensee(v.Seller).(trader.AuctionAsker)
			if !ok {
				MkLog.Error("Error notifying Seller", "license", v.Seller, "reason", "Seller unqualified to recieve updates")
				return
			}
			bidder, ok := m.m.Licensee(v.Winner.Bidder).(trader.AuctionBidder)
			if !ok {
				MkLog.Error("Error notifying Bidder", "license", v.Winner.Bidder, "reason", "Bidder unqualified to recieve updates")
				return
			}
			asker.UpdateAuction(a, x.TS_ACKNOWLEDGED, x.R_OK)
			bidder.UpdateBid(v.Winner.ID, x.TS_ACKNOWLEDGED, x.R_OK)
		}()
	}
	//TODO: Queue verification with Security
	m.auctions[a] = v
	return nil
}

type Dealer struct{}

func (_ *Dealer) RegisterDealer(x.Dealer)     { return }              // TODO
func (_ *Dealer) Resolve(x.Symbol) []x.Dealer { return []x.Dealer{} } // TODO
func (_ *Dealer) Dealers() []x.Dealer         { return []x.Dealer{} } // TODO

type Ticker struct{}

func (_ *Ticker) RegisterTicker(string, ticker.Ticker) {}                           // TODO
func (_ *Ticker) Tickers() []ticker.Ticker             { return []ticker.Ticker{} } // TODO

type Market1 struct {
	Auctioner
	Dealer
	Ticker
	rid          uuid.UUID
	rdc          uint64
	bank         bank.Bank
	brokers      []broker.Broker
	tickers      []ticker.Ticker
	trading      bool
	traders      map[x.LicenseID]trader.Trader
	traderbases  map[string]manager.TraderMaker
	tradercounts map[string]uint64
	stock        map[x.Symbol]x.Stock
}

var MkLog = log15.New("module", "market", "version", "v1")

func NewMarket1() market.Exchange {
	m := &Market1{}
	m.Init()
	return m
}

func (m *Market1) Init() {
	MkLog.Info("First run, setting new RunID")
	rid, _ := uuid.NewV4()
	m.rid = *rid
	MkLog.Info("Init", "runid", m.rid.String())
	m.traders = make(map[x.LicenseID]trader.Trader)
	m.traderbases = make(map[string]manager.TraderMaker)
	m.tradercounts = make(map[string]uint64)
	m.stock = make(map[x.Symbol]x.Stock)
	m.Auctioner.m = m //mmmmmmmm
	m.setup()
}

func (m *Market1) SpawnTraders() {
	for k, v := range m.tradercounts {
		MkLog.Info("Preparing to spawn", "number", v, "type", k)
		for i := 1; i <= int(v); i++ {
			var lid x.LicenseID = x.LicenseID(rand.Uint32())
			for {
				_, ok := m.traders[lid]
				if !ok {
					break
				} else {
					lid = x.LicenseID(rand.Uint32())

				}
				//lid = LicenseID(rand.Uint32())
			}
			MkLog.Info("Spawning trader", "total", k, "number", i, "id", lid)
			t := m.traderbases[k](m)
			t.Init(lid)
			m.traders[lid] = t
		}
	}
}

func (m *Market1) CreateTrader(ty string, lid x.LicenseID) {
	MkLog.Info("Spawning pre-licensed trader", "type", ty, "id", lid)
	t := m.traderbases[ty](m)
	t.Init(lid)
	m.traders[lid] = t
}

func (m *Market1) RunID() (uuid.UUID, uint64, x.Mode) {
	return m.rid, m.rdc, x.M_BOTH
}

func (m *Market1) AddBroker(b broker.Broker) { m.brokers = append(m.brokers, b) }
func (m *Market1) Brokers() []broker.Broker  { return m.brokers }

func (m *Market1) AddStock(s x.Stock) {
	m.stock[s.Symbol] = s
}
func (m *Market1) IPO(x.Stock, float32, []x.LicenseID) {}

func (m *Market1) SetCount(tc ...manager.TraderCfg) {
	for _, v := range tc {
		MkLog.Info("Setting", "number", v.Count, "x", v.Type)
		m.tradercounts[v.Type] = v.Count
	}
}

func (m *Market1) RegisterTrader(s string, t manager.TraderMaker) {
	MkLog.Info("Adding trader base type", "type", s)
	m.traderbases[s] = t
}
func (m *Market1) Traders() (t []trader.Trader) {
	for _, v := range m.traders {
		t = append(t, v)
	}
	return
}
func (m *Market1) Licensee(lid x.LicenseID) trader.Trader { return m.traders[lid] }

func (m *Market1) RegisterBroker(string, manager.BrokerMaker) {} // TODO

func (m *Market1) RegisterBank(_ string, b bank.Bank) { m.bank = b }
func (m Market1) Bank() bank.Bank                     { return m.bank }

func (m *Market1) Start(daylen, pauselen time.Duration, weeklen uint) {
	MkLog.Info("Starting week")
	for i1 := uint(0); i1 <= weeklen; i1++ {
		var ds x.Summary
		m.rdc++
		m.Open()
		select {
		case <-time.After(daylen):
			ds = m.Close(true)
		}

		select {
		case <-time.After(pauselen):
			MkLog.Info("Summary", "data", ds.String())
		}
	}
}

func (m *Market1) Retire(x.LicenseID) {} // TODO

func (m *Market1) Open() {
	MkLog.Info("Opening market")
	m.trading = true
	m.resumeAll()
	for _, v := range m.traders {
		go v.Status(x.S_OPEN)
	}
}

func (m *Market1) Pause() {
	MkLog.Info("Pausing trading")
	m.trading = false
	m.pauseAll()
	for _, v := range m.traders {
		go v.Status(x.S_PAUSE)
	}
}

func (m *Market1) Resume() {
	MkLog.Info("Resuming trading")
	m.trading = true
	m.resumeAll()
	for _, v := range m.traders {
		go v.Status(x.S_RESUME)
	}
}

func (m *Market1) Close(normal bool) x.Summary {
	MkLog.Info("Closing market")
	m.trading = false
	m.pauseAll()
	var ds []x.Summary
	for _, v := range m.brokers {
		go v.Status(x.S_CLOSE)
	}
	for _, v := range m.brokers {
		ds = append(ds, v.Summary(x.SL_DAY))
	}

	for _, v := range m.traders {
		go v.Status(x.S_CLOSE)
	}
	return x.Combine(ds)
}

func (m *Market1) Status() bool { return m.trading }

func (m *Market1) List() (ret []x.Stock) {
	for _, v := range m.stock {
		ret = append(ret, v)
	}
	return
}
