package game

import "sync"

// pay_action_token.go — #6189: the token a payment has an opponent
// create (PayActionOpponentCreatesToken, Varchild's War-Riders'
// "Cumulative upkeep—Have an opponent create a 1/1 red Survivor creature
// token").
//
// The payment rides a prompt as data, so it names its token by KEY and
// not by template: a whole Card on a PendingChoice would be a second
// copy of the printed token on every restore point, and the token rows
// live in the catalog (cards/effects/tokens_table.go), which the game
// package cannot see. The catalog files its row here at init, the way it
// files a conjurable template (conjure.go), and the payment looks it up
// as it is paid. A key is an on-disk identity: never renamed, never
// reused (ADR 0041 P9).

var (
	payActionTokenMu sync.RWMutex
	payActionTokens  = map[string]Card{}
)

// RegisterPayActionToken files the printed token template a
// PayActionOpponentCreatesToken creates under `key`. Called from the
// catalog's init; a second template under one key panics, as a duplicate
// registration does.
func RegisterPayActionToken(key string, template Card) {
	if key == "" {
		panic("game: RegisterPayActionToken needs a key")
	}
	payActionTokenMu.Lock()
	defer payActionTokenMu.Unlock()
	if _, dup := payActionTokens[key]; dup {
		panic("game: pay-action token " + key + " registered twice")
	}
	payActionTokens[key] = template
}

// PayActionTokenForEffect is the template filed under `key`.
func PayActionTokenForEffect(key string) (Card, bool) {
	payActionTokenMu.RLock()
	defer payActionTokenMu.RUnlock()
	c, ok := payActionTokens[key]
	return c, ok
}
