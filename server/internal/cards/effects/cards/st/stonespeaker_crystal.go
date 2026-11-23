package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stonespeaker Crystal — Artifact {4} (EDHREC rank 3147):
//
//	"{T}: Add {C}{C}.
//	 {2}, {T}, Sacrifice this artifact: Exile any number of target
//	 players' graveyards. Draw a card."
//
// A Hedron Archive that answers graveyards instead of drawing two.
// The mana ability is a plain Sol Ring shape; the sacrifice ability
// is Tormod's Crypt widened to "any number of target players" — a
// zero-to-unbounded player clause, so the activator may name nobody,
// everyone, or just the reanimator — with the draw after the exile,
// in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a494fcee-6885-434c-aad5-6f83640c4472",
		Name:         "Stonespeaker Crystal",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice Stonespeaker Crystal: Exile any number of target players' graveyards. Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("any number of target players").WithCount(0, 0),
			Effect:  effects.B30ExileTargetGraveyardsThenDraw,
		}},
	})
}
