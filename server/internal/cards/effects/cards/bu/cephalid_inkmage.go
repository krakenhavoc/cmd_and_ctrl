package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cephalid Inkmage — Creature — Octopus Wizard {2}{U}, 2/2 (Foundations,
// #4465):
//
//	"When this creature enters, surveil 3.
//	 Threshold — This creature can't be blocked as long as there are seven
//	 or more cards in your graveyard."
//
// The entry trigger surveils three. Threshold is Nightwhorl Hermit's
// block rule, its seven-card count read as blockers are declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d979a0b-ba04-4abb-8f78-c707ce15351a",
		Name:         "Cephalid Inkmage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cephalid Inkmage — surveil 3", effects.Do(effects.Surveil{N: 3})),
		},
		BlockRules: []game.BlockRule{effects.CantBeBlockedWhile(effects.OnSelf(), func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			return effects.FraThresholdMet(g, c.Controller)
		})},
	})
}
