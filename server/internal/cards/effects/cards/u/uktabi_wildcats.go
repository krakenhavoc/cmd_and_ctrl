package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Uktabi Wildcats — Creature — Cat {4}{G}, */* (Mirage, S82 OLD b808):
//
//	"Uktabi Wildcats's power and toughness are each equal to the number of
//	 Forests you control.
//	 {G}, Sacrifice a Forest: Regenerate this creature."
//
// The size is a layer 7a characteristic-defining ability (CR 604.3) counting
// every Forest its controller controls, a Forest-typed nonbasic land included;
// sacrificing a Forest for the regeneration lowers it at once. The regeneration
// is the shared "<cost>: Regenerate this creature" body behind a mana and a
// sacrifice-a-Forest cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd25afe0-a69a-4e0b-b364-b88d3a522ed3",
		Name:         "Uktabi Wildcats",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.OldB808PTEqualToForestsYouControl("Uktabi Wildcats's power and toughness are each equal to the number of Forests you control."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, Sacrifice a Forest: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost: effects.Plus(effects.ManaCost("{G}"),
				effects.SacrificeN(1, "a Forest", effects.Subtype("Forest"))),
			Effect: effects.RegenerateThisSource,
		}},
	})
}
