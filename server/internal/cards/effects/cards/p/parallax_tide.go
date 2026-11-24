package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Parallax Tide — Enchantment {2}{U}{U} (Secret Lair Drop):
//
//	"Fading 5
//	 Remove a fade counter from this enchantment: Exile target land.
//	 When this enchantment leaves the battlefield, each player returns to the
//	 battlefield all cards they own exiled with it."
//
// Fading is the shared pair (fading.go). The land is exiled linked to this
// object (ExileTarget.Linked, ADR 0155 §8) and the leave trigger returns every
// card still exiled with it, together, under its owner's control (Savior of
// Ollenbock's shape); a Tide that returns is a new object with no link.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e36a92be-537d-4a63-858a-afaab064fbf7",
		Name:         "Parallax Tide",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Fading(5)},
		Triggered: []game.TriggeredAbility{
			effects.FadingUpkeep(),
			effects.On(game.EventLTB, effects.Self, "Parallax Tide — each player returns the cards they own exiled with it", effects.ReturnAllExiledWithThisTogether),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove a fade counter from this enchantment: Exile target land.",
			Cost:    effects.RemoveCountersFromThis(game.CounterFade, 1),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  effects.AvB7ExileTargetLinked,
		}},
	})
}
