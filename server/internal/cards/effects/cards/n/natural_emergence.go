package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Natural Emergence — Enchantment {2}{R}{G} (Planeshift, S82 OLD b643, #5615):
//
//	"When this enchantment enters, return a red or green enchantment you control to
//	 its owner's hand.
//	 Lands you control are 2/2 creatures with first strike. They're still lands."
//
// The enters trigger is Marsh Crocodile's mandatory return over enchantments; this
// one is red and green, so it may return itself. The static is Living Plane's two
// layers over the lands its controller controls: layer 4 adds the creature type
// and keeps every other type, layer 7b sets 2/2 so counters and pumps still apply,
// and layer 6 grants first strike. The lands are summoning sick the turn they
// enter, as any creature is (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce19d2dd-4b2c-493e-8963-ad97649248e2",
		Name:         "Natural Emergence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Natural Emergence — return a red or green enchantment you control to its owner's hand", effects.Do(effects.ReturnOneYouControl{
				Match:    effects.OldB643EnchantmentOfEitherColor("R", "G"),
				Question: "Natural Emergence — return a red or green enchantment you control to its owner's hand",
			})),
		},
		Static: effects.OldB643YourLandsAre22FirstStrikeCreatures("Lands you control are 2/2 creatures with first strike. They're still lands."),
	})
}
