package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kozilek's Sentinel — Creature — Eldrazi Drone {1}{R}, 1/4 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Whenever you cast a colorless spell, this creature gets +1/+0 until end of
//	 turn."
//
// Skittering Cicada's colorless-spell trigger (judged on the spell as cast,
// through the layers), with each cast a separate +1/+0.
//
// No simplification.
func init() {
	const name = "Kozilek's Sentinel"
	effects.Register(effects.Spec{
		OracleID:        "86ae05d6-f1ba-456d-bbc0-f35cd0d0537d",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered:       []game.TriggeredAbility{effects.OldB306ThisGetsPowerWhenYouCastColorless(name)},
	})
}
