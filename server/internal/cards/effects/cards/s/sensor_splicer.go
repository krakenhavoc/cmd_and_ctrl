package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sensor Splicer — Creature — Phyrexian Artificer {4}{W}, 1/1 (S82 GAP, #5948):
//
//	"When this creature enters, create a 3/3 colorless Phyrexian Golem
//	 artifact creature token.
//	 Golem creatures you control have vigilance."
//
// "Golem" is a word of the card's text, read through it (CR 612.2).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9676da93-1e17-4f6c-b610-a242b78c71ab",
		Name:         "Sensor Splicer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sensor Splicer — create a 3/3 colorless Phyrexian Golem artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 colorless Phyrexian Golem artifact"), N: 1})),
		},
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Golem"))
			}, "vigilance"),
		},
	})
}
