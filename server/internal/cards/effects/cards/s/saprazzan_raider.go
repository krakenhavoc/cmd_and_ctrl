package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saprazzan Raider — Creature — Merfolk {2}{U}, 1/2 (Mercadian Masques, S82 OLD batch 688, #5619):
//
//	"When this creature becomes blocked, return it to its owner's hand."
//
// The once-per-attacker "becomes blocked" event (CR 509.1h) over Fleeting Effigy's return to
// hand, which acts only while the Raider is still the permanent the ability came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca011fb1-fdc6-4eec-9344-c247339292ab",
		Name:         "Saprazzan Raider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesBlocked && ev.CardID == source.InstanceID
			}, "Saprazzan Raider — return it to its owner's hand",
				effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand))),
		},
	})
}
