package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orc Sureshot — {3}{B} Creature — Orc Archer 4/2 (Fate Reforged):
//
//	"Whenever another creature you control enters, target creature an opponent
//	 controls gets -1/-1 until end of turn."
//
// One trigger per creature that enters under your control, the Sureshot's own
// entry excluded; the target is chosen as it goes on the stack (CR 603.3d) and
// re-checked as it resolves (CR 608.2b). A creature that enters at the same
// time as the Sureshot does trigger it (CR 603.6a): both are on the battlefield
// then.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c56726a6-d424-46a0-b3b3-2b44fbcbc63a",
		Name:         "Orc Sureshot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
					return ok && c.IsCreature()
				}, "Orc Sureshot — target creature an opponent controls gets -1/-1 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.BoostUntilEOT{Target: id, Power: -1, Toughness: -1, Label: "Orc Sureshot — -1/-1"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
