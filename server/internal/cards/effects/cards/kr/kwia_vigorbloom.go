package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kwia Vigorbloom — Legendary Creature — Elder Sphinx {3}{G}{W}{W},
// 6/6:
//
//	"Flying, vigilance, lifelink, ward {2}
//	 Whenever you gain life, create a colorless artifact token named
//	 Lotus with "{T}, Sacrifice this token: Add three mana of any one
//	 color." This ability triggers only once each turn."
//
// The once-per-turn limit is Lazav's (b11TriggeredThisTurn): the trigger
// is declined at the harvest when one from this Kwia already went on the
// stack this turn. The Lotus is a catalog token (LotusToken) whose three
// mana are one colour pick, the Gilded Lotus shape.
//
// No simplification.
func init() {
	const label = "Kwia Vigorbloom — create a Lotus token"
	effects.Register(effects.Spec{
		OracleID:        "07a28621-e617-46b0-af3e-2efb03b5056a",
		Name:            "Kwia Vigorbloom",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Kwia Vigorbloom — ward {2}"),
			effects.On(game.EventChangeLife, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.YouGainedLife(ev, source, lki, g) && !effects.B11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, effects.Do(effects.CreateToken{Template: effects.LotusToken(), N: 1})),
		},
	})
}
