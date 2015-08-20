package market

import (
	"github.com/Efruit/marqit/exchange"
	"github.com/Efruit/marqit/managers"
	"github.com/nu7hatch/gouuid"
	"time"
)

type Exchange interface {
	manager.Bank
	manager.Broker
	manager.Ticker
	exchange.Dealership
	exchange.Auctioner

	Init()                                     // Perform initial setup
	RunID() (uuid.UUID, uint64, exchange.Mode) // Tell the Run identifier, open number, and exchange model.

	AddStock(exchange.Stock)                           // Define a stock
	IPO(exchange.Stock, float32, []exchange.LicenseID) // IPO a stock, volume specified by the exchange.Stock.Number, price by [2], initial holder(s) by [3]. If [3] is empty, the entirety will be given to a random trader.
	List() []exchange.Stock                            // Retrieve the active stock list

	Start(time.Duration, time.Duration, uint) // Run the simulation with a day length of [1] and a day count of [2]
	Open()                                    // Open the market
	Pause()                                   // Stop processing queue items
	Resume()                                  // Continue the queue
	Close(Normal bool) exchange.Summary       // Sound the closing bell. Normal specifies the nature of the closing.

	Status() bool // Are we trading?
}
