package manager

import (
	"github.com/Efruit/marqit/bank"
)

type Bank interface {
	RegisterBank(string, bank.Bank)
	Bank() bank.Bank
}
