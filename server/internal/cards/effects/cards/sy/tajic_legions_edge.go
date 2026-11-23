package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tajic, Legion's Edge — Legendary Creature — Human Soldier {1}{R}{W}, 3/2
// (Ravnica Remastered, #5948):
//
//	"Haste
//	 Mentor (Whenever this creature attacks, put a +1/+1 counter on target
//	 attacking creature with lesser power.)
//	 Prevent all noncombat damage that would be dealt to other creatures you
//	 control.
//	 {R}{W}: Tajic gains first strike until end of turn."
//
// Mentor is Blade Instructor's trigger. The prevention is Blessed Sanctuary's
// CR 615.1a replacement, for another creature the Tajic's controller controls
// (Crystal Barricade's recipient): damage to Tajic itself and combat damage to
// the others still happen. The grant is Endling's, pinned to this Tajic.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ae6f21a2-e6b6-4793-8343-e27310c0bea1",
		Name:            "Tajic, Legion's Edge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{effects.Mentor("Tajic, Legion's Edge — mentor")},
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToAnotherCreatureYouControl,
				Damage: effects.NoncombatDamage,
				Label:  "Tajic, Legion's Edge — prevent all noncombat damage that would be dealt to other creatures you control",
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}{W}: Tajic gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{R}{W}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Tajic, Legion's Edge — first strike until end of turn"),
		}},
	})
}
