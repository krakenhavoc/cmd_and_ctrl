package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chittering Witch — Creature — Human Warlock {3}{B}, 2/2 (Tarkir: Dragonstorm Commander)
// Commander):
//
//	"When this creature enters, create a number of 1/1 black Rat
//	 creature tokens equal to the number of opponents you have.
//	 {1}{B}, Sacrifice a creature: Target creature gets -2/-2 until end
//	 of turn."
//
// The Rats are counted as the trigger resolves, over the opponents still
// in the game (the printed ruling). The ability is Plagued Rusalka's with
// the Witch able to sacrifice herself as the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b1aa532-4f19-407e-83a5-13b3160ad8c2",
		Name:         "Chittering Witch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Chittering Witch — create a number of 1/1 black Rat creature tokens equal to the number of opponents you have",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return CreateToken{Template: TokenCard("1/1 black Rat"), N: len(ctx.Opponents())}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice a creature: Target creature gets -2/-2 until end of turn.",
			Purpose: ForTargets(HarmUntilEOT(0, -2, -2)),
			Cost:    Plus(ManaCost("{1}{B}"), SacrificeACreature()),
			Targets: TargetCreature("target creature"),
			Effect:  boostTargetUntilEOT("Chittering Witch — -2/-2", -2, -2),
		}},
	})
}
