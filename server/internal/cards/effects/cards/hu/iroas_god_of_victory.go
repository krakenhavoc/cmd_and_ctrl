package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iroas, God of Victory — Legendary Enchantment Creature — God
// {2}{R}{W}, 7/4 (Secret Lair Drop, #5462):
//
//	"Indestructible
//	 As long as your devotion to red and white is less than seven, Iroas
//	 isn't a creature.
//	 Creatures you control have menace.
//	 Prevent all damage that would be dealt to attacking creatures you
//	 control."
//
// Mogis' God clause with red and white (GodUnlessDevotionToColors, CR 700.5).
// The menace is Goblin War Drums' grant over the creatures its controller
// controls; Iroas himself gets it only while he is a creature. The prevention
// is Rune-Tail's Essence's static with a narrower recipient: a creature its
// controller controls that is attacking (CR 506.4) as the damage would be
// dealt, combat or not, from any source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "36041cf2-159f-4809-9377-930ac4db545f",
		Name:            "Iroas, God of Victory",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{
			effects.GodUnlessDevotionToColors(7, "R", "W"),
			effects.TribalKeywordGrant(effects.TribeFilter{YoursOnly: true}, "menace"),
		},
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToAttackingCreaturesYouControl,
				Label: "Iroas, God of Victory — prevent all damage that would be dealt to attacking creatures you control",
			}),
		},
	})
}
