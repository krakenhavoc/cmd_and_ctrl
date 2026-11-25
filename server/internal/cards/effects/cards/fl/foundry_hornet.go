package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foundry Hornet — {3}{B} Creature — Insect 2/3 (Aether Revolt, #5536):
//
//	"Flying
//	 When this creature enters, if you control a creature with a +1/+1
//	 counter on it, creatures your opponents control get -1/-1 until end of
//	 turn."
//
// An intervening "if" (CR 603.4), read as the ability would trigger and
// again as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e5342529-dc1a-4ebb-a0fd-843e01147402",
		Name:            "Foundry Hornet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control a creature with a +1/+1 counter on it",
				func(g *game.Game, you uuid.UUID) bool {
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == you && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0 {
							return true
						}
					}
					return false
				},
				effects.WhenThisEnters("Foundry Hornet — creatures your opponents control get -1/-1 until end of turn",
					effects.Do(effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.OpponentControls()),
						Power:     -1,
						Toughness: -1,
						Label:     "Foundry Hornet — -1/-1 until end of turn",
					}))),
		},
	})
}
