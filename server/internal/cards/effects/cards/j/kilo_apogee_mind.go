package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kilo, Apogee Mind — Legendary Artifact Creature — Robot Artificer
// {U}{R}{W}, 3/3 (Edge of Eternities Commander):
//
//	"Haste
//	 Whenever Kilo becomes tapped, proliferate."
//
// Attentive Sunscribe's becomes-tapped trigger (the tap event and the
// attack that taps, ThisBecameTapped) over Flux Channeler's proliferate.
// A vigilance attack does not tap it, and a tap paid to station goes on
// the stack above the station ability, as the ruling says.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a884166-de5a-43ba-8855-4999f39f2059",
		Name:            "Kilo, Apogee Mind",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Kilo, Apogee Mind — proliferate",
				effects.Do(effects.Proliferate{})),
		},
	})
}
