//go:generate stringer -type "Status,Mode,Action,SumLen,TradeStatus"
package exchange

import (
	"github.com/nu7hatch/gouuid"
	"time"
)

type Lister interface {
	List() []Stock
}

type Statuser interface {
	Status(Status)
}

type Status uint

const (
	S_OPEN Status = iota
	S_CLOSE
	S_PAUSE
	S_RESUME
	S_QUOTA_REACHED
)

type LicenseID uint

type Trade struct {
	Seller   LicenseID
	Volume   uint64
	Buyer    LicenseID
	Duration time.Time
}

type Mode uint

const (
	M_AUCTION Mode = iota
	M_DEALER
	M_BOTH
)

type TraderCfg struct {
	Type  string
	Count uint64
}

type Symbol string

type Stock struct {
	Name    string
	Trading bool
	Symbol  Symbol
	Sectors []Sector
	Shares  uint
	Mode    Mode
}

type Sector uint

type AuctionID struct{ uuid.UUID }

func (a AuctionID) String() string {
	return a.UUID.String()
}

type Duration int
type Stamp struct {
	Day  int
	Time int
}

type Order struct {
	Buyer      LicenseID
	Action     Action
	Price      float32
	Target     Symbol
	Size       uint
	Expiration Duration
	ID         OrderID
}

type OrderID uuid.UUID

type Action uint

const (
	A_NONE Action = iota
	A_BID
	A_ASK
)

type SumLen uint

const (
	SL_DAY SumLen = iota
	SL_WEEK
	SL_CYCLE
)

type Summary struct {
	Traders uint64
	Volume  uint64
	Trades  []Trade
}

func (s Summary) String() string {
	return ""
}

func Combine([]Summary) Summary {
	return Summary{}
}

type TradeStatus uint

const (
	TS_UNKNOWN      TradeStatus = iota // I know nothing of this item.
	TS_SUBMITTED                       // I posted this item.
	TS_ACKNOWLEDGED                    // I have won the bid, or finalized the deal; and am committed to paying.
	TS_PAID                            // I have paid or transfered the assets.
	TS_RECIEVED                        // I have recieved payment.
	TS_CLOSED                          // I have completed this interaction successfully.
	TS_ABORTED                         // I broke off this interaction.
	TS_ERROR_ME                        // I have made some error.
	TS_ERROR_YOU                       // The other party has made some error.
	TS_ROLLBACK                        // I have attempted to repair my error.
	TS_FAULT_ME                        // I can not complete this transaction.
	TS_FAULT_YOU                       // The other party can not complete this transaction.
)
