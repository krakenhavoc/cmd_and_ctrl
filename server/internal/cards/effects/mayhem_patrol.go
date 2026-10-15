package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "f77a9160-1d37-476d-8908-29ddcc982328",
		Name:             "Mayhem Patrol",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"menace"},
		AlternativeCosts: []game.AlternativeCost{Blitz("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Mayhem Patrol — target creature gets +1/+0",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return BoostUntilEOT{Target: t.ID, Power: 1, Label: "Mayhem Patrol — +1/+0"}.Apply(ctx)
						}
					}
					return nil
				}), TargetCreature("target creature")),
		},
	})
}
