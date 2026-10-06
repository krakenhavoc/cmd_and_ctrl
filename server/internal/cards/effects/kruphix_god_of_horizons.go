package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kruphix, God of Horizons — Legendary Enchantment Creature — God
// {3}{G}{U}, 4/7:
//
//	"Indestructible
//	 As long as your devotion to green and blue is less than seven,
//	 Kruphix isn't a creature.
//	 You have no maximum hand size.
//	 If you would lose unspent mana, that mana becomes colorless
//	 instead."
//
// The conversion is Spec.ManaPool's ManaPoolBecomesColorless (#2166):
// the mana stays in the pool as {C} (so it is kept at every later
// boundary too, as the ruling says), and keeps whatever spend
// restrictions it already had. The other lines are the shared God
// clause and the hand-size static.
func init() {
	Register(Spec{
		OracleID:        "b2cefcd6-4b81-479c-86ff-1695b836972c",
		Name:            "Kruphix, God of Horizons",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		NoMaxHandSize:   true,
		Static:          []game.StaticAbility{godUnlessDevotionToColors(7, "G", "U")},
		ManaPool: []game.ManaPoolStatic{
			{Kind: game.ManaPoolBecomesColorless},
		},
	})
}
