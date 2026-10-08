package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cackling Counterpart — Instant for {1}{U}{U}:
//
//	"Create a token that's a copy of target creature you control.
//	 Flashback {5}{U}{U}"
//
// The first flashback card with a target clause, and the reason the
// cast path had to ride the Board's ordinary prompt chain rather
// than the bare-payload route the exile impulse button uses: a
// graveyard cast of this has to open the targeting UI exactly as a
// hand cast does.
//
// The token is a copy of the copiable values only (CR 707.2), which
// CreateTokenCopy already gets right — counters, damage and auras on
// the original are not copied, and the copy carries the original's
// oracle ID so every catalog hook keyed on it comes along.
//
// The caveat this card used to carry (a copied card's on-enter hook was
// skipped) closed with #762, when token creation became a replaceable
// event with a real entry: the token's CR 614.12 Spec.AsEnters clause
// and its ETB triggers both fire. Relm's Sketching, which shares this
// TokenCopyOfSingleTarget body, has shipped Full on that basis (#2550).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "9e2adca5-f39c-4a09-bcce-8238ebac2c4a",
		Name:             "Cackling Counterpart",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{U}{U}")},
		Targets:          TargetCreature("target creature you control", YouControl()),
		OnResolve:        TokenCopyOfSingleTarget,
	})
}
