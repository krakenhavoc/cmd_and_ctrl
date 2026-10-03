package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Xolatoyac, the Smiling Flood — Legendary Creature — Salamander Serpent (6/6) for {4}{G}{U}:
//
//	"Whenever Xolatoyac enters or attacks, put a flood counter on target land. That land is an Island in addition to its other types for as long as it has a flood counter on it.
//	 At the beginning of your end step, untap each permanent you control with a counter on it."
//
// ADR 0109 §2 (#1604): Aquitect's Will's flood counter
// (FloodTargetLand) on an enters-or-attacks trigger. The end-step
// untap reads each permanent you control as it resolves; one with any
// counter on it, of any kind, untaps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c0fc559-857f-4239-97fa-379ced3dfc1f",
		Name:         "Xolatoyac, the Smiling Flood",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEntersOrAttacks("Xolatoyac, the Smiling Flood — put a flood counter on target land",
				func(g *game.Game, item *game.StackItem) error {
					return FloodTargetLand(NewContext(g, item), "Xolatoyac, the Smiling Flood")
				}), TargetPermanent("target land", Land())),
			AtYourEndStep("Xolatoyac, the Smiling Flood — untap each permanent you control with a counter on it",
				func(g *game.Game, item *game.StackItem) error {
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != item.Controller || !c.Tapped || !b64HasAnyCounter()(g, item.Controller, c) {
							continue
						}
						if err := g.UntapTargetForEffect(c.InstanceID); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
