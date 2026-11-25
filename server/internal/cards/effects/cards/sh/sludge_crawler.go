package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sludge Crawler — Creature — Eldrazi Drone {B}, 1/1 (Battle for Zendikar,
// #5547):
//
//	"Devoid (This card has no color.)
//	 Ingest (Whenever this creature deals combat damage to a player, that
//	 player exiles the top card of their library.)
//	 {2}: This creature gets +1/+1 until end of turn."
//
// Ingest is OldB308Ingest; the pump stacks with each activation.
//
// No simplification.
func init() {
	const name = "Sludge Crawler"
	effects.Register(effects.Spec{
		OracleID:        "dd05294b-2c5b-48c2-a9f6-8a528baa3dff",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered:       []game.TriggeredAbility{effects.OldB308Ingest(name)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1,
					Label: name + " — +1/+1 until end of turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
