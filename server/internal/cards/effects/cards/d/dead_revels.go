package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dead Revels — Sorcery {3}{B} (Ravnica Allegiance):
//
//	"Spectacle {1}{B} (You may cast this spell for its spectacle cost rather
//	 than its mana cost if an opponent lost life this turn.)
//	 Return up to two target creature cards from your graveyard to your hand."
//
// Spectacle is the shared alternative cost; the body is Morbid Plunder's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4354eb0a-28c5-4f3f-8e78-fd7eec230c7b",
		Name:             "Dead Revels",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Spectacle("{1}{B}")},
		Targets:          effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.Creature(), effects.YouOwn()).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
