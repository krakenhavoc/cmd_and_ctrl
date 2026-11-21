package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunderhawk Gunship — Artifact — Vehicle {6}, 6/6 (Warhammer 40,000
// Commander, S82 GAP batch 186, #5948):
//
//	"Flying
//	 When this Vehicle enters, create two 2/2 white Astartes Warrior
//	 creature tokens with vigilance.
//	 Whenever this Vehicle attacks, attacking creatures you control gain
//	 flying until end of turn.
//	 Crew 2"
//
// The attack trigger grants flying to the attacking creatures you control
// as it resolves (CR 611.2c): the set is locked then, so a creature put onto
// the battlefield attacking afterwards does not get it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d64a4ad2-c80e-4a7f-93f2-b702916dba07",
		Name:            "Thunderhawk Gunship",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Thunderhawk Gunship"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Thunderhawk Gunship — create two 2/2 white Astartes Warrior creature tokens with vigilance",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white Astartes Warrior with vigilance"), N: 2})),
			effects.WheneverThisAttacks("Thunderhawk Gunship — attacking creatures you control gain flying until end of turn",
				effects.Do(effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.IsAttackingCreature(), effects.YouControl()),
					Keywords: []string{"flying"},
					Label:    "Thunderhawk Gunship — flying",
				})),
		},
	})
}
