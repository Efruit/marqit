//go:generate stringer -type "Reason"
package exchange

type Dealership interface { // NASDAQ Model
	RegisterDealer(Dealer)
	Resolve(Symbol) []Dealer
	Dealers() []Dealer
}

type Dealer interface {
	SellPrice(Symbol) (float32, error)
	BuyPrice(Symbol) (float32, error)

	Stocks() []Symbol

	Sell(Deal) Response
	Buy(Deal) Response
}

type Response struct {
	OK     bool
	Reason Reason
}

type Reason uint

const (
	R_OK           Reason = 0
	R_NONE         Reason = iota + 1
	R_NOT_TRADING         // Default cop-out
	R_PROHIBITED          // I can't sell what I have, period.
	R_RESTRICTED          // I can't sell what I have TO YOU.
	R_PRICE_LOW           // Probably not issued when buying.
	R_PRICE_HIGH          // Probably not for selling either.
	R_UNACCEPTABLE        // Like ^ but more vauge. Try changing something and get back to me.
	R_CANT_PAY            // I can not pay
)

type Deal struct {
	Buyer  LicenseID
	Volume uint64
}
