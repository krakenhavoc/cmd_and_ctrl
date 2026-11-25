package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rubblebelt Rioters — Creature — Human Berserker {1}{R}{G}, 0/4 (War of the Spark, #5520):
//
//	"Haste
//	 Whenever this creature attacks, it gets +X/+0 until end of turn, where
//	 X is the greatest power among creatures you control."
//
// Pathbreaker Ibex's X: the greatest power among creatures you control as the
// trigger resolves, the Rioters themselves counted (the text says "creatures
// you control"). The pump goes to the Rioters only while they are still on the
// battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d750f1d1-a3f9-4180-a903-2bd3f4cb9674",
		Name:            "Rubblebelt Rioters",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Rubblebelt Rioters — it gets +X/+0 until end of turn, where X is the greatest power among creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.BoostUntilEOT{
						Target: item.SourceCardID,
						Power:  effects.B42GreatestPowerControlledBy(g, item.Controller),
						Label:  "Rubblebelt Rioters — +X/+0",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
