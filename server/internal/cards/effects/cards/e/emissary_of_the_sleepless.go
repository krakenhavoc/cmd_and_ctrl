package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emissary of the Sleepless — Creature — Spirit {4}{W}, 2/4 (Shadows over
// Innistrad, #5543):
//
//	"Flying
//	 When this creature enters, if a creature died this turn, create a 1/1
//	 white Spirit creature token with flying."
//
// The intervening "if" (CR 603.4) reads the table-wide tally of creatures
// that died this turn, tokens included, as the ability triggers and again
// as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f21bbacb-0436-46fa-b361-582b766d3964",
		Name:            "Emissary of the Sleepless",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondAnyCreatureDied{}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && g.TurnTally.CreaturesDied > 0
			}, "Emissary of the Sleepless — create a 1/1 white Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1}))),
		},
	})
}
