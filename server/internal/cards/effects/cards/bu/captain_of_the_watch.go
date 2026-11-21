package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain of the Watch — Creature — Human Soldier {4}{W}{W}, 3/3 (Magic 2010, #5585):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)
//	 Other Soldier creatures you control get +1/+1 and have vigilance.
//	 When this creature enters, create three 1/1 white Soldier creature
//	 tokens."
//
// Vigilance rides PrintedKeywords. The lord half is Yotian Tactician's
// anthem and Commander Mustard's keyword grant over the other Soldiers you
// control, so an opponent's Soldiers are not affected. The enters trigger
// creates three 1/1 white Soldier tokens, which the lord then pumps.
//
// No simplification.
func init() {
	soldiers := effects.TribeFilter{Tribes: []string{"Soldier"}, Others: true, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:        "9f2af06d-8fb7-4276-bb92-3559a6d1fa18",
		Name:            "Captain of the Watch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Tokens: 3, CreatureTokens: 3},
		Static: []game.StaticAbility{
			effects.TribalAnthem(soldiers, 1, 1),
			effects.TribalKeywordGrant(soldiers, "vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Captain of the Watch — create three 1/1 white Soldier creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 3})),
		},
	})
}
