package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Biting-Palm Ninja — Creature — Human Ninja {2}{B}, 3/3 (Neon Dynasty):
//
//	"Ninjutsu {2}{B}
//	 This creature enters with a menace counter on it.
//	 Whenever this creature deals combat damage to a player, you may remove
//	 a menace counter from it. When you do, that player reveals their hand
//	 and you choose a nonland card from it. Exile that card."
//
// The menace counter is a keyword counter (CR 122.1b) put on as it enters,
// including when ninjutsu puts it onto the battlefield. The trigger is the
// shared remove-a-counter-then-reflexive shape (s82_neo_b06_helpers.go); the
// reflexive trigger reveals the damaged player's hand to the table and
// exiles the nonland card you choose (exiling is not discarding).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c22447b-7a7f-4a27-8d3b-1bb664ea94c3",
		Name:         "Biting-Palm Ninja",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{2}{B}")},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterMenace, 1, "Biting-Palm Ninja: enters with a menace counter on it"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverThisDealsCombatDamageToAPlayer("Biting-Palm Ninja — you may remove a menace counter from it; when you do, that player reveals their hand and you exile a nonland card from it",
				effects.S82NeoB06RemoveKeywordCounterThen(game.CounterMenace,
					"Biting-Palm Ninja — that player reveals their hand; you choose a nonland card from it and exile it",
					effects.S82NeoB06RevealHandExileBody)),
				"Biting-Palm Ninja — remove a menace counter? (When you do, that player reveals their hand and you exile a nonland card from it.)"),
		},
	})
}
