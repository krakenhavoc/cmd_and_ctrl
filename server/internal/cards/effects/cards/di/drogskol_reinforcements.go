package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drogskol Reinforcements — Creature — Spirit Soldier {3}{W}, 2/2
// (Innistrad: Crimson Vow Commander):
//
//	"Melee (Whenever this creature attacks, it gets +1/+1 until end of turn
//	 for each opponent you attacked this combat.)
//	 Other Spirits you control have melee.
//	 Prevent all noncombat damage that would be dealt to Spirits you
//	 control."
//
// Melee is the engine keyword (CR 702.121). The grant is a layer 6 keyword
// grant to every other Spirit its controller has (Skyhunter Strike
// Force's shape). The prevention is Blessed Sanctuary's noncombat
// prevention narrowed to Spirits; the Reinforcements is itself a Spirit and
// is covered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6d174448-52f5-4439-9586-3abc5af940d7",
		Name:            "Drogskol Reinforcements",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordMelee},
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.InstanceID != source.InstanceID && target.HasSubtypeWord("Spirit")
			}, game.KeywordMelee),
		},
		Replacements: []game.ReplacementEffect{
			effects.VowB12PreventNoncombatToSpirits("Drogskol Reinforcements — prevent all noncombat damage that would be dealt to Spirits you control"),
		},
	})
}
