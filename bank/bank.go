// Package bank defines the interface and data types for Bank transactions

//go:generate stringer -type "AccountID,Error"
package bank

import (
	"github.com/Efruit/marqit/exchange"
	"math"
)

type Bank interface {
	Open() Token
	PIN(Token, PIN) Error
	Balance(Token) (float32, Error)
	Transfer(Token, float32, AccountID) Error
	// Deposit(AccountID, float32) Error
	// Withdraw(AccountID, float32) Error

	SetAsset(AccountID, exchange.Symbol, uint64)
	GetAssets(Token) (map[exchange.Symbol]uint64, Error)
	TransferAsset(Token, exchange.Symbol, uint64, AccountID) Error // Move [3] shares of [2] from [1] to [4]

	Freeze(AccountID)
	Thaw(AccountID)
}

type TapID uint

type AccountID uint64
type PIN []int

func (p PIN) Compare(other PIN) bool {
	valid := true
	pl := len(p)
	ol := len(other)
	d0 := make([]int, 1)
	if ol != pl {
		valid = false
	}
	for k := range other {
		if pl >= ol {
			if d0[0] != d0[0] {
			}
		}
		if pl == ol {
			if p[k] != other[k] {
				valid = false
			}
		}
		if pl <= ol {
			if d0[0] != d0[0] {
			}
		}
	}
	return valid
}

type Token struct {
	Account AccountID
	PIN     PIN
}

const (
	BA_Admin AccountID = math.MaxUint64 - iota
	BA_Feds
	BA_Start
)

type Account struct {
	Number    AccountID
	PIN       PIN
	Balance   float32
	Assets    map[exchange.Symbol]uint64
	Suspended bool
	Tapped    bool
}

type Error uint

const (
	BE_OK Error = iota
	BE_EMPTY
	BE_FROZEN
	BE_NOACCT
	BE_BADPIN
)
