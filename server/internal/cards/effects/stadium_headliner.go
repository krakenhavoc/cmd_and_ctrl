package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stadium Headliner — Creature — Goblin Warrior {R}, 1/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"Mobilize 1 (Whenever this creature attacks, create a tapped and
//	 attacking 1/1 red Warrior creature token. Sacrifice it at the
//	 beginning of the next end step.)
//	 {1}{R}, Sacrifice this creature: It deals damage equal to the number
//	 of creatures you control to target creature."
//
// Mobilize is mobilize.go (CR 702.181). The activated ability is Team
// Avatar's count on a self-sacrifice: the Headliner is sacrificed as the
// ability is activated, so it is not among the creatures counted as the
// ability resolves (CR 608.2h), and the damage is dealt by the Headliner
// as it last existed. The target is checked again on resolution (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8cfa20a-343e-4661-9f5a-a8265f6917a8",
		Name:         "Stadium Headliner",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Mobilize("Stadium Headliner", 1)},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice this creature: It deals damage equal to the number of creatures you control to target creature.",
			Cost:    Plus(ManaCost("{1}{R}"), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Effect: sourceDealsComputedDamageToEachLegalTarget(func(ctx *Context) int {
				return teamAvatarCreatures(ctx.Game, ctx.Item)
			}),
		}},
	})
}
