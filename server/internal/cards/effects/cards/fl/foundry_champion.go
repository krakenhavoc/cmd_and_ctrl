package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foundry Champion — Creature — Elemental Soldier {4}{R}{W}, 4/4 (Gatecrash,
// S82 OLD #5565):
//
//	"When this creature enters, it deals damage to any target equal to the
//	 number of creatures you control.
//	 {R}: This creature gets +1/+0 until end of turn.
//	 {W}: This creature gets +0/+1 until end of turn."
//
// Spawn of Thraxes's enter trigger with the creatures counted as it resolves
// (the Champion itself included while it is still out), and Beacon Hawk's and
// Bellows Lizard's pump rows with no {T}, so each may be activated any number
// of times. The Champion is the damage source, as it last existed if it has
// left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1c44856-ca53-4b6e-902e-8d9c6fe1f61c",
		Name:         "Foundry Champion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Foundry Champion — it deals damage to any target equal to the number of creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						n := effects.B04CreaturesControlled(g, ctx.Controller())
						return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetAny()),
		},
		Activated: []effects.ActivatedAbility{
			effects.ThisGetsRow("{R}", 1, 0, "Foundry Champion"),
			effects.ThisGetsRow("{W}", 0, 1, "Foundry Champion"),
		},
	})
}
