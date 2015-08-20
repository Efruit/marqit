package trader

import (
	"github.com/Efruit/marqit/bank"
	xchg "github.com/Efruit/marqit/exchange"
)

//var Traders = make(map[LicenseID]Trader)

type Trader interface {
	xchg.Statuser
	Init(xchg.LicenseID)     // You've been licensed, here's your ID
	GetID() xchg.LicenseID   // What's your ID?
	GetAcct() bank.AccountID // What's your bank account?
}

type AuctionAsker interface { // Must implement to sell
	NewBid(xchg.AuctionID, xchg.Bid)                             // A new bid has been placed on your auction.
	UpdateAuction(xchg.AuctionID, xchg.TradeStatus, xchg.Reason) // Signals the other party's tradestatus
	ConfirmAuction(xchg.AuctionID) xchg.TradeStatus              // The market will periodically check up on transactions to make sure nobody's filing under a false identity
}

type AuctionBidder interface { // Must implement to buy
	NewAuction(xchg.Auction)                             // A new auction has been placed
	UpdateBid(xchg.BidID, xchg.TradeStatus, xchg.Reason) // The bid's tradestatus has been changed. i.e. you have won the auction, awaiting payment, etc.
	ConfirmBid(xchg.BidID) xchg.TradeStatus              // The market will periodically check up on transactions to make sure nobody's filing under a false identity
}

type Dealer interface { // Must implement to ???
	DealerList([]xchg.Dealer)
	UpdateDeal(xchg.Deal, xchg.TradeStatus)
	ConfirmDeal(xchg.Deal) xchg.TradeStatus
}
