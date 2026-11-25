package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tenth District Veteran — Creature — Human Soldier {2}{W}, 2/3 (Ravnica
// Allegiance):
//
//	"Vigilance
//	 Whenever this creature attacks, untap another target creature you control."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c25e4b69-4ea3-43ce-ba6f-b918ebfc4980",
		Name:            "Tenth District Veteran",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Tenth District Veteran — untap another target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok || id == uuid.Nil {
						return nil
					}
					return effects.UntapTarget{Target: id}.Apply(ctx)
				}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
