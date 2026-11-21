package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brutal Cathar // Moonrage Brute — {2}{W} Creature — Human Soldier
// Werewolf 2/2 // Creature — Werewolf 3/3 (#2586, ADR 0132):
//
//	Front: "When this creature enters or transforms into Brutal Cathar,
//	        exile target creature an opponent controls until this
//	        creature leaves the battlefield.
//	        Daybound"
//	Back:  "First strike
//	        Ward—Pay 3 life.
//	        Nightbound"
//
// One ability with two conditions: this permanent entering, or this
// permanent turning over TO its front face (EventTransform, Amount 0 —
// Revealing Eye's reading, CR 701.27e). Cast at night it enters on
// Moonrage Brute and never triggers; turning back to day triggers it
// then. Each trigger exiles with its own "until this leaves" record
// (CR 610.3), so the first exile ends when the permanent leaves, not when
// it merely turns over.
//
// No simplification.
func init() {
	const oracle = "1ed2d8e0-462b-468e-8fd3-1f3c6d99fb8a"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Brutal Cathar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventTransform},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				if ev.CardID != source.InstanceID {
					return false
				}
				return ev.Kind == game.EventETB || (ev.Kind == game.EventTransform && ev.Amount == 0)
			},
			Targets: effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			Key:     "Brutal Cathar — exile target creature an opponent controls until this creature leaves the battlefield",
			Effect: effects.ExileChosenTargetUntilThisLeaves(
				"Brutal Cathar — the exiled creature returns when Brutal Cathar leaves the battlefield"),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Moonrage Brute",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "nightbound"},
		Triggered:       []game.TriggeredAbility{effects.Ward(effects.WardLife(3), "Moonrage Brute — ward—pay 3 life")},
	})
}
