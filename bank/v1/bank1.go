package v1

import (
	bank ".."
	"../../exchange"
	"github.com/inconshreveable/log15"
	"math/rand"
	"sync"
)

var BnLog = log15.New("module", "bank", "version", "v1")

type Bank1 struct {
	sync.RWMutex
	accounts map[bank.AccountID]bank.Account
	a        uint
	apin     bank.PIN
}

var _ bank.Bank = &Bank1{}

func NewBank1() (bank.Bank, bank.PIN) {
	apin := (*Bank1).generatePIN(nil)
	return &Bank1{
		accounts: make(map[bank.AccountID]bank.Account),
		apin:     apin,
	}, apin
}

func (_ *Bank1) generatePIN() bank.PIN {
	length := rand.Intn(10) + 10
	p := make(bank.PIN, length)
	for k := range p {
		p[k] = rand.Intn(2048)
	}
	return p
}

func (b *Bank1) verify(t bank.Token) bool {
	if v, ok := b.accounts[t.Account]; !ok && !(t.Account == bank.BA_Start || t.Account == bank.BA_Admin || t.Account == bank.BA_Feds) {
		BnLog.Info("Access attempted on nonexistant bank account", "token", t)
		return false
	} else {
		valid := v.PIN.Compare(t.PIN) || b.apin.Compare(t.PIN)
		if !valid {
			BnLog.Info("Access denied on existing bank account", "token", t, "actual", v.PIN)
		}
		return valid
	}
}

func (b *Bank1) isAdmin(t bank.Token) bool {
	return b.apin.Compare(t.PIN)
}

func (b *Bank1) Open() bank.Token {
	b.a++
	a := bank.Account{
		Number:  bank.AccountID(b.a),
		PIN:     b.generatePIN(),
		Balance: 0.0,
		Assets:  make(map[exchange.Symbol]uint64),
	}

	b.accounts[a.Number] = a
	return bank.Token{a.Number, a.PIN}
}

func (b *Bank1) PIN(t bank.Token, n bank.PIN) bank.Error {
	if !b.verify(t) {
		return bank.BE_BADPIN
	}

	b1 := b.accounts[t.Account]
	b1.PIN = n
	b.accounts[t.Account] = b1
	return bank.BE_OK
}

func (b *Bank1) Balance(t bank.Token) (float32, bank.Error) {
	if !b.verify(t) {
		return 0.0, bank.BE_BADPIN
	}
	return b.accounts[t.Account].Balance, bank.BE_OK
}

// func (b *Bank1) Deposit(a bank.AccountID, c float32) bank.Error {
// 	if c < 0 {
// 		panic("bank: deposit amt < 0")
// 	}
// 	if b.frozen[a] {
// 		return bank.BE_FROZEN
// 	} else {
// 		b.balance[a] += c
// 		return bank.BE_OK
// 	}
// }

// func (b *Bank1) Withdraw(a bank.AccountID, c float32) bank.Error {
// 	if c < 0 {
// 		panic("bank: deposit amt < 0")
// 	}
// 	if b.frozen[a] {
// 		return bank.BE_FROZEN
// 	} else if b.balance[a]-c <= 0 {
// 		return bank.BE_EMPTY
// 	} else {
// 		b.balance[a] -= c
// 		return bank.BE_OK
// 	}
// }

func (b *Bank1) Transfer(a1 bank.Token, c float32, a2 bank.AccountID) bank.Error {
	if !b.verify(a1) {
		return bank.BE_BADPIN
	}
	if _, ok := b.accounts[a2]; !ok {
		return bank.BE_NOACCT
	}
	if c < 0 {
		panic("bank: transfer amt < 0")
	}
	if b.isAdmin(a1) && (a1.Account == bank.BA_Admin || a1.Account == bank.BA_Start) {
		BnLog.Info("Admin/Start account deducted", "from", a1, "amt", c, "to", a2)
		bb := b.accounts[a2]
		bb.Balance += c
		b.accounts[a2] = bb
		return bank.BE_OK
	}
	if b.accounts[a1.Account].Suspended || b.accounts[a2].Suspended {
		return bank.BE_FROZEN
	} else if b.accounts[a1.Account].Balance-c <= 0 {
		return bank.BE_EMPTY
	} else {
		b1 := b.accounts[a1.Account]
		b2 := b.accounts[a2]
		b1.Balance -= c
		b2.Balance += c
		b.accounts[a1.Account] = b1
		b.accounts[a2] = b2
		BnLog.Info("Transferring", "from", a1.Account, "from.new", b1.Balance, "to", a2, "to.new", b2.Balance, "amt", c)
		return bank.BE_OK
	}
}

func (b *Bank1) SetAsset(a bank.AccountID, s exchange.Symbol, n uint64) {
	if v, ok := b.accounts[a]; !ok {
		return
	} else if v.Assets == nil {
		v.Assets = make(map[exchange.Symbol]uint64)
		v.Assets[s] = n
		b.accounts[a] = v
	} else {
		v.Assets[s] = n
		b.accounts[a] = v
	}
}

func (b *Bank1) GetAssets(a bank.Token) (map[exchange.Symbol]uint64, bank.Error) {
	if !b.verify(a) {
		return map[exchange.Symbol]uint64{}, bank.BE_BADPIN
	} else {
		return b.accounts[a.Account].Assets, bank.BE_OK
	}
}

func (b *Bank1) TransferAsset(from bank.Token, s exchange.Symbol, n uint64, to bank.AccountID) bank.Error {
	if !b.verify(from) {
		return bank.BE_BADPIN
	}

	if _, ok := b.accounts[to]; !ok {
		return bank.BE_NOACCT
	}

	if b.accounts[from.Account].Suspended || b.accounts[to].Suspended {
		return bank.BE_FROZEN
	} else if b.accounts[from.Account].Assets == nil {
		return bank.BE_EMPTY
	} else if b.accounts[from.Account].Assets[s]-n <= 0 {
		return bank.BE_EMPTY
	} else {
		b1 := b.accounts[from.Account]
		b2 := b.accounts[to]

		b1.Assets[s] -= n
		b2.Assets[s] += n

		b.accounts[from.Account] = b1
		b.accounts[to] = b2
		return bank.BE_OK
	}
}

func (b *Bank1) Freeze(a bank.AccountID) {
	b1 := b.accounts[a]
	b1.Suspended = true
	b.accounts[a] = b1
}

func (b *Bank1) Thaw(a bank.AccountID) {
	b1 := b.accounts[a]
	b1.Suspended = false
	b.accounts[a] = b1
}
