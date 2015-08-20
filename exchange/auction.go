package exchange

import (
	"errors"
	"strconv"
	"time"
)

var NoSuchEntity = errors.New("No such entity")
var NoPermission = errors.New("Permission denied")
var NotTrading = errors.New("Trading currently prohibited")
var UnqualifiedBidder = errors.New("Bidder is unqualified")
var UnqualifiedAsker = errors.New("Asker is unqualified")

type Auctioner interface { // NYSE Model
	Auctions() []Auction
	Auction(AuctionID) (Auction, error)
	Ask(Auction) (AuctionID, error)
	Bid(Bid) (BidID, error)
	Finish(AuctionID) error
	Cancel(AuctionID) error
}

type Auction struct {
	ID         AuctionID
	Seller     LicenseID
	Stock      Symbol
	Open       bool
	Length     time.Duration `ask:"min=1"`
	Ask        float32       `ask:"min=0"`
	Number     uint64        `ask:"min=1"`
	Bids       []Bid
	Adjudicate func([]Bid) Bid
	Winner     Bid
}

type Bid struct {
	ID     BidID
	Placed time.Time
	Bidder LicenseID
	Number uint64  `place:"min=1"`
	Price  float32 `place:"min=0"`
}

func (a Auction) Win() Bid {
	var Best Bid
	if a.Adjudicate == nil {
		for _, v := range a.Bids {
			if v.Price > Best.Price && v.Placed.After(Best.Placed) {
				Best = v
			}
		}
	} else {
		Best = a.Adjudicate(a.Bids)
	}
	return Best
}

type BidID struct {
	Auction AuctionID
	Bid     uint64
}

func (b BidID) String() string {
	return b.Auction.String() + "#" + strconv.FormatUint(b.Bid, 10)
}
