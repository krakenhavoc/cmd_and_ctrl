package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bramble Wurm — Creature — Wurm {6}{G}, 7/6 (Innistrad: Crimson Vow):
//
//	"Reach, trample
//	 When this creature enters, you gain 5 life.
//	 {2}{G}, Exile this card from your graveyard: You gain 5 life."
//
// The graveyard ability exiles the card as a cost (CR 602.2b) and gains life
// as it resolves, so the source having gone changes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "db4f67a9-6b47-4abe-9367-718da9e0609e",
		Name:            "Bramble Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bramble Wurm — you gain 5 life", effects.Do(effects.GainLife{Amount: 5})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, Exile this card from your graveyard: You gain 5 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.Do(effects.GainLife{Amount: 5}),
		}},
	})
}
