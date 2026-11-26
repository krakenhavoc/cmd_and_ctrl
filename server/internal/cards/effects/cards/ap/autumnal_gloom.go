package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Autumnal Gloom // Ancient of the Equinox — Enchantment {2}{G}
// (Shadows over Innistrad, #5543):
//
//	"{B}: Mill a card.
//	 Delirium — At the beginning of your end step, if there are four or
//	 more card types among cards in your graveyard, transform this
//	 enchantment."
//
// Back, Creature — Treefolk 4/4: "Trample, hexproof".
//
// The end-step trigger is an intervening if (CR 603.4) on the controller's
// own graveyard: checked as the step begins and again as the trigger
// resolves. The transform is the in-place verb (CR 701.27), ignored if the
// enchantment has already turned over since the trigger went on the stack
// (CR 701.27f).
//
// No simplification.
const AutumnalGloomOracleID = "245aa786-d607-414f-ab55-b003b2b9faf4"

func init() {
	effects.Register(effects.Spec{
		OracleID:     AutumnalGloomOracleID,
		Name:         "Autumnal Gloom",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: Mill a card.",
			Cost:    effects.ManaCost("{B}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.Do(effects.MillCards{N: 1}),
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.B24CondDelirium{},
				effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, effects.YouHaveDeliriumWhen),
					"Autumnal Gloom — delirium: transform this enchantment", effects.Do(effects.TransformThis{}))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        AutumnalGloomOracleID + "#1",
		Name:            "Ancient of the Equinox",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "hexproof"},
	})
}
