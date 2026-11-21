package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resistance Squad — Creature — Human Soldier {2}{W}, 3/2 (Innistrad:
// Crimson Vow, #5495):
//
//	"When this creature enters, if you control another Human, draw a
//	 card."
//
// An intervening "if" (CR 603.4): checked as the Squad enters, so with no
// other Human it never triggers, and again as the trigger resolves, so a
// Human that left in response draws nothing. Humans are read through the
// layers (a changeling or a token counts); the Squad never counts itself
// ("another").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a16eeef9-e03c-4ebb-b986-325252a9491f",
		Name:         "Resistance Squad",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.S82VowB02ControlsAnotherHuman,
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.Self(ev, source, game.Characteristic{}, g) && effects.S82VowB02ControlsAnother(g, source.Controller, source.InstanceID, source.TextSubtype("Human"))
				}, "Resistance Squad — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
