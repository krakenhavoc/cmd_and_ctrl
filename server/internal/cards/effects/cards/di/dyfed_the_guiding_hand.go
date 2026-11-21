package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dyfed, the Guiding Hand — Legendary Planeswalker — Dyfed {3}{U}{U}, loyalty 4
// (Mystery Booster Commander, #5948):
//
//	"+1: Create two tapped Powerstone tokens. (They're artifacts with "{T}:
//	 Add {C}. This mana can't be spent to cast a nonartifact spell.")
//	 −X: Untap X target artifacts.
//	 −6: Search your library for an artifact card, put it onto the battlefield,
//	 then shuffle.
//	 Dyfed, the Guiding Hand can be your commander."
//
// The +1 is Ashnod's tapped Powerstone, twice. The −X announces X with the
// activation (CR 107.3a), no more than Dyfed's loyalty; its clause is exactly X
// target artifacts (CountFromX), each re-checked as it resolves (CR 608.2b), and
// X = 0 untaps nothing. The −6 searches for any artifact card, not a
// "may" (the search is mandatory, a failure to find is legal), puts it onto the
// battlefield and shuffles. "Can be your commander" is deck-building, read from
// the oracle text by the deck validator.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab23a9e4-2d62-4b01-9d52-593de40811b0",
		Name:         "Dyfed, the Guiding Hand",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Create two tapped Powerstone tokens.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 2},
				Effect:  effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 2}),
			},
			{
				Label:   "−X: Untap X target artifacts.",
				Cost:    effects.LoyaltyMinusX(),
				Targets: effects.TargetsCountedByX(effects.TargetPermanent("X target artifacts", effects.Artifact())),
				Effect:  effects.B22UntapEachLegalTarget,
			},
			{
				Label:   "−6: Search your library for an artifact card, put it onto the battlefield, then shuffle.",
				Cost:    effects.LoyaltyCost(-6),
				Purpose: game.Purpose{Tutors: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsArtifact() },
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Reason:    "Dyfed, the Guiding Hand — search for an artifact card to put onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
