package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lurking Chupacabra — Creature — Beast Horror {3}{B}, 2/3 (Ixalan, #5532):
//
//	"Whenever a creature you control explores, target creature an opponent
//	 controls gets -2/-2 until end of turn."
//
// Merfolk Cave-Diver's explore trigger with Bat'leth's -2/-2: the target is
// chosen as the trigger goes on the stack (CR 603.3d), so with no opposing
// creature it is removed; one gone or no longer legal by resolution is
// unaffected (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Lurking Chupacabra — target creature an opponent controls gets -2/-2 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "178e73f9-fc6e-4425-870e-45f96644010a",
		Name:         "Lurking Chupacabra",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.On(game.EventExplored, effects.ACreatureYouControlExplored, label,
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.BoostUntilEOT{Target: id, Power: -2, Toughness: -2, Label: "Lurking Chupacabra — -2/-2 until end of turn"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
		},
	})
}
