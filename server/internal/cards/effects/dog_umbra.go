package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dog Umbra — Enchantment — Aura {1}{W} (Modern Horizons 3, #5948):
//
//	"Flash
//	 Enchant creature
//	 As long as another player controls enchanted creature, it can't attack
//	 or block. Otherwise, this Aura has umbra armor. (If enchanted creature
//	 would be destroyed, instead remove all damage from it and destroy this
//	 Aura.)"
//
// Two halves of one conditional, both re-read at every recompute and every
// destruction, so a Control Magic on the enchanted creature flips the Aura
// between them. "Another player" is anyone but the Aura's controller. The
// restriction is Pacifism's, gated with RestrictAttachedWhile; umbra armor
// is the shared CR 702.89 replacement, applying only while the enchanted
// creature's controller is the Aura's controller.
//
// No simplification.
func init() {
	umbra := UmbraArmor()
	baseApplies := umbra.AppliesTo
	umbra.AppliesTo = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
		if src == nil || !baseApplies(ev, g, src) {
			return false
		}
		return g.AffectedPermanentControllerForEffect(ev) == src.Controller
	}
	Register(Spec{
		OracleID:        "e475c96f-6cf8-4139-92fd-f68ecaae4893",
		Name:            "Dog Umbra",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			RestrictAttachedWhile(func(host *game.Card, _ *game.Game, source *game.Card) bool {
				return host.Controller != source.Controller
			}, game.CantAttackOrBlock),
		},
		Replacements: []game.ReplacementEffect{umbra},
	})
}
