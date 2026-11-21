package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witch's Oven — Artifact {1} (EDHREC rank 3105):
//
//	"{T}, Sacrifice a creature: Create a Food token. If the
//	 sacrificed creature's toughness was 4 or greater, create two
//	 Food tokens instead. (They're artifacts with "{2}, {T},
//	 Sacrifice this token: You gain 3 life.")"
//
// The one-mana sacrifice outlet that pays in Food. A CR 602 ability
// with a tap and a sacrifice-a-creature cost (Goblin Bombardment's
// clause — the activator's own creatures, the source excluded by
// type); the creature is gone by the time the ability resolves, so
// which one it was is read back off the event log
// (b17PermanentSacrificedToPay, Jarad's read) and its toughness as
// it last stood is its printed toughness plus its counters
// (b29LastKnownToughnessOffBattlefield).
//
// Toughness is the sacrificed creature's last-known value, post-layer
// and with counters (PermanentInfo.Toughness, CR 603.10), so an anthem or
// an Equipment bonus counts toward the two-Food threshold.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fa0fe02-2452-4386-8e0c-165757b0f0a3",
		Name:         "Witch's Oven",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a creature: Create a Food token. If the sacrificed creature's toughness was 4 or greater, create two Food tokens instead.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeACreature()),
			Effect:  effects.B29FoodForSacrificedCreature,
		}},
	})
}
