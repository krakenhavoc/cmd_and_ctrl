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
	// FlipTwoIgnoreOne is Krark's Thumb's "If you would flip a coin,
	// instead flip two coins and ignore one" (#6152, CR 705): a
	// replacement of each coin the player flips. Each instance doubles the
	// coins (two Thumbs flip four and keep one), and the player keeps the
	// coin they want: on a called flip a won coin, which is never worse
	// than a lost one, and on an uncalled flip the face the effect asked
	// for (FlipCoinsPreferringForEffect). The ignored coins never happened:
	// they emit no event, so nothing triggers on them.
	FlipTwoIgnoreOne bool
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

// maxThumbDoublings caps the coins one flip can become: 2^6 = 64 coins a
// coin, past any table's count of Thumbs and well inside the stream.
const maxThumbDoublings = 6

// coinsPerFlipLocked is how many coins `player` flips for each coin an
// effect asks for: 2^k for k FlipTwoIgnoreOne rules on permanents they
// control (#6152). Caller must hold g.mu.
func (g *Game) coinsPerFlipLocked(player uuid.UUID) int {
	if CatalogCoinFlipRules == nil || g.Battlefield == nil || player == uuid.Nil {
		return 1
	}
	k := 0
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
			if r.FlipTwoIgnoreOne {
				k++
			}
		}
	}
	if k > maxThumbDoublings {
		k = maxThumbDoublings
	}
	return 1 << k
}
