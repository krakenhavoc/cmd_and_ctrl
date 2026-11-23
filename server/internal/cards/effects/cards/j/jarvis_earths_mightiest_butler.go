package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jarvis, Earth's Mightiest Butler — Legendary Creature — Human Advisor
// {2}{U}, 1/4 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever you cast a Hero spell, draw a card."
//
// A Hero spell is a spell with the Hero subtype, read from the spell on
// the stack as it is cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63e17cbd-6dae-42d2-a074-7924db1477bf",
		Name:         "Jarvis, Earth's Mightiest Butler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.HasSubtype("Hero"), "Jarvis — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
