package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hell's Thunder — Creature — Elemental {1}{R}{R}, 4/4 (Shards of Alara,
// #5588):
//
//	"Flying, haste
//	 At the beginning of the end step, sacrifice this creature.
//	 Unearth {4}{R}"
//
// Blistering Firecat's end-step sacrifice, at every player's end step,
// with the shared Unearth row (CR 702.84a). An unearthed Hell's Thunder
// meets both at the same end step: whichever resolves first, unearth's
// replacement exiles it instead of letting it go to the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5aad3485-47c1-4d87-b340-61d7003cd786",
		Name:            "Hell's Thunder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Hell's Thunder — sacrifice it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{4}{R}")},
	})
}
