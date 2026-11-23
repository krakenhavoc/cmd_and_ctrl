package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cori Mountain Monastery — Land (EDHREC rank 3577):
//
//	"This land enters tapped unless you control a Plains or an
//	 Island.
//	 {T}: Add {R}.
//	 {3}{R}, {T}: Exile the top card of your library. Until the end
//	 of your next turn, you may play that card."
//
// Tarkir: Dragonstorm's Jeskai utility land. The tapped entry is the
// checkland's CR 614 self-replacement reading the post-layer land
// types (a Hallowed Fountain turns it on); the activated ability is
// Prosper's Mystic Arcanum — the top card exiled with permission to
// PLAY it, land included, lasting until the end of the controller's
// next turn (b20ExileTopUntilEndOfNextTurn, ADR 0063's seat-turn
// duration).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35c60b66-8c85-432e-90fe-99c19d21ed15",
		Name:         "Cori Mountain Monastery",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedUnless(effects.YouControlLandTyped("plains", "island"))},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}, {T}: Exile the top card of your library. Until the end of your next turn, you may play that card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{R}"), effects.TapCost()),
			Effect:  effects.B20ExileTopUntilEndOfNextTurn,
		}},
	})
}
