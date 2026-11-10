package game

import "github.com/google/uuid"

// coin_flip_rules.go — a printed static that decides how a player's coins
// land: Edgar, King of Figaro's "Two-Headed Coin — The first time you flip
// one or more coins each turn, those coins come up heads and you win those
// flips."
//
// Like OpponentEffectProtection it is data on the card's definition, read
// off the battlefield through CatalogAbilityKey every time a player flips,
// so it is derived rather than stored: it ends the moment the permanent
// leaves or loses its abilities, and two of them do not change the answer.
// "The first time each turn" is read off the turn's event log: no
// EventFlipCoin by the player has been emitted since the turn began (the
// flip that is about to happen has emitted none yet).

// CoinFlipRule is one printed rule over the coins its controller flips.
type CoinFlipRule struct {
	// Label names the rule for the log ("Two-Headed Coin").
	Label string
	// FirstTimeEachTurn limits the rule to the first flip action each turn
	// in which the player flips one or more coins.
	FirstTimeEachTurn bool
	// HeadsAndWin makes every coin of the flip come up heads and be won by
	// the player. Called and uncalled flips alike: an uncalled flip has no
	// winner, and this rule is what lets one be won (the printed ruling).
	HeadsAndWin bool
}

// CatalogCoinFlipRules returns the rules a permanent with the catalog key
// carries, or nil. Set from CardDef at boot like every other per-slot hook.
var CatalogCoinFlipRules func(key string) []CoinFlipRule

// coinsComeUpHeadsLocked reports whether the coins `player` is about to
// flip are made heads and won by a rule on a permanent they control. Caller
// must hold g.mu.
func (g *Game) coinsComeUpHeadsLocked(player uuid.UUID) bool {
	if CatalogCoinFlipRules == nil || g.Battlefield == nil || player == uuid.Nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		if src.Controller != player {
			continue
		}
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, r := range CatalogCoinFlipRules(key) {
			if !r.HeadsAndWin {
				continue
			}
			if r.FirstTimeEachTurn && g.flippedCoinsThisTurnLocked(player) {
				continue
			}
			return true
		}
	}
	return false
}

// flippedCoinsThisTurnLocked reports whether the player has flipped a coin
// since the turn began. Caller must hold g.mu.
func (g *Game) flippedCoinsThisTurnLocked(player uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == EventFlipCoin && ev.Actor == player {
			return true
		}
	}
	return false
}
