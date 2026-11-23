package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sister of Silence — Creature — Human Knight {4}{U}, 3/3 (Warhammer 40,000
// Commander, S82 GAP batch 195, #5948):
//
//	"Flash
//	 Psychic Abomination — When this creature enters, counter target instant
//	 spell, sorcery spell, activated ability, or triggered ability."
//
// The target is a stack object chosen as the trigger goes on the stack: an
// instant or a sorcery spell (a creature spell is not one), an activated
// ability (a mana ability never uses the stack, and a loyalty ability is an
// activated ability) or a triggered ability, and it is re-checked as the trigger
// resolves (CR 608.2b). One that has left the stack is not countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fcc38b0b-d54a-4ecf-96f7-e09709227735",
		Name:            "Sister of Silence",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Sister of Silence — counter target instant spell, sorcery spell, activated ability, or triggered ability",
				effects.GapB195CounterTheTargetItem), effects.GapB195SisterTarget()),
		},
	})
}
