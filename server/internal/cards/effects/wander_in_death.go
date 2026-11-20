package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wander in Death — Sorcery {2}{B}
// (Amonkhet, #5535):
//
//	"Return up to two target creature cards from your graveyard to your hand.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Macabre Reconstruction's return, plus cycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f25c5451-58fa-4916-8342-ae88c5e406cc",
		Name:         "Wander in Death",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("up to two target creature cards from your graveyard", Creature(), YouOwn()).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
