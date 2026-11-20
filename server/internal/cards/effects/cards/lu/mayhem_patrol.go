package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mayhem Patrol — Creature — Devil Warrior {1}{R}, 1/2:
//
//	"Menace
//	 Whenever this creature attacks, target creature gets +1/+0 until end
//	 of turn.
//	 Blitz {1}{R}"
//
// Blitz (CR 702.152, #3772). The attack trigger targets any creature,
// the Patrol itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "f77a9160-1d37-476d-8908-29ddcc982328",
		Name:             "Mayhem Patrol",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"menace"},
		AlternativeCosts: []game.AlternativeCost{effects.Blitz("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Mayhem Patrol — target creature gets +1/+0",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: "Mayhem Patrol — +1/+0"}.Apply(ctx)
						}
					}
					return nil
				}), effects.TargetCreature("target creature")),
		},
	})
}
