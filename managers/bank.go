package manager

import (
	"../bank"
)

type Bank interface {
	RegisterBank(string, bank.Bank)
	Bank() bank.Bank
}
