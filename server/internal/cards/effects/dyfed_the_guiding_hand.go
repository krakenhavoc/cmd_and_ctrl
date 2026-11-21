package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ab23a9e4-2d62-4b01-9d52-593de40811b0",
		Name:         "Dyfed, the Guiding Hand",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Create two tapped Powerstone tokens.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 2},
				Effect:  Do(CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 2}),
			},
			{
				Label:   "−X: Untap X target artifacts.",
				Cost:    LoyaltyMinusX(),
				Targets: targetsCountedByX(TargetPermanent("X target artifacts", Artifact())),
				Effect:  b22UntapEachLegalTarget,
			},
			{
				Label:   "−6: Search your library for an artifact card, put it onto the battlefield, then shuffle.",
				Cost:    LoyaltyCost(-6),
				Purpose: game.Purpose{Tutors: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsArtifact() },
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Reason:    "Dyfed, the Guiding Hand — search for an artifact card to put onto the battlefield",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
