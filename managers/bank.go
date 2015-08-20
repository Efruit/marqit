// Package manager defines the interfaces for managing Banks, Tickers, Markets, etc.
package manager

import (
	"github.com/Efruit/marqit/bank"
)

type Bank interface {
	RegisterBank(string, bank.Bank)
	Bank() bank.Bank
}
