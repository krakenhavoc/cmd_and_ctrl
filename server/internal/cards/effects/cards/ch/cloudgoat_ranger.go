package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloudgoat Ranger — Creature — Giant Warrior Ranger {3}{W}{W}, 3/3
// (Kaldheim Commander, #5506):
//
//	"When this creature enters, create three 1/1 white Kithkin Soldier
//	 creature tokens.
//	 Tap three untapped Kithkin you control: This creature gets +2/+0 and
//	 gains flying until end of turn."
//
// The cost is tap-others (#758) over Kithkin; it is not the {T} symbol
// (CR 302.6), so Kithkin that arrived this turn may pay it, and it does
// not target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1c13a19-46cd-474b-a46b-a483be6b91bb",
		Name:         "Cloudgoat Ranger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cloudgoat Ranger — create three 1/1 white Kithkin Soldier tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Kithkin Soldier"), N: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "Tap three untapped Kithkin you control: This creature gets +2/+0 and gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant,
				Pump: &game.Pump{Power: 2, Keywords: []string{"flying"}}},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: effects.TargetPermanent("three untapped Kithkin you control", effects.HasSubtype("Kithkin")),
				Label:  "three untapped Kithkin you control",
			}},
			Effect: effects.ThisCreatureUntilEOT("Cloudgoat Ranger — +2/+0 and flying", 2, 0, "flying"),
		}},
	})
}
