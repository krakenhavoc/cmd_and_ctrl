package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leyline of Abundance — Enchantment {2}{G}{G} (Enchanting Tales):
//
//	"If this card is in your opening hand, you may begin the game with it on
//	 the battlefield.
//	 Whenever you tap a creature for mana, add an additional {G}.
//	 {6}{G}{G}: Put a +1/+1 counter on each creature you control."
//
// The opening-hand clause (CR 103.6a) is Spec.OpeningHand (ADR 0133). The mana
// is Badgermole Cub's triggered mana ability (CR 605.1b): it fires for a
// creature of yours tapped for mana, read through the layers, and adds {G}.
// The activation is Abandoned Air Temple's one placement per creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1197a595-3afc-4f18-b868-bbb096771922",
		Name:         "Leyline of Abundance",
		Completeness: effects.CompletenessFull,
		OpeningHand:  effects.BeginTheGameOnTheBattlefield(),
		ManaTriggers: []game.ManaTrigger{{
			Label: "Leyline of Abundance — add an additional {G}",
			AppliesTo: func(prod game.ManaProduced, source *game.Card, _ *game.Game) bool {
				return prod.Source.IsCreature() && prod.Controller == source.Controller
			},
			Produced: func(game.ManaProduced, *game.Card, *game.Game) string { return "{G}" },
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}{G}{G}: Put a +1/+1 counter on each creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{6}{G}{G}"),
			Effect:  effects.B11PutCounterOnEachCreatureYouControl,
		}},
	})
}
