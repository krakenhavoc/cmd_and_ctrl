package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pearl-Ear, Imperial Advisor — Legendary Creature — Fox Advisor {1}{W}{W},
// 3/4 (Secrets of Strixhaven Commander, #4834):
//
//	"Lifelink
//	 Enchantment spells you cast have affinity for Auras. (They cost {1} less
//	 to cast for each Aura you control.)
//	 Whenever you cast an Aura spell that targets a modified permanent you
//	 control, draw a card. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// The affinity is the AffinityFor cost reduction (CR 702.41a) given to other
// spells: a cost modifier over the controller's enchantment spells, counting
// the Auras on the battlefield that controller controls (an Aura spell still on
// the stack is not one). The draw is a cast trigger over the targets the spell
// announced (CR 601.2c): a target that is a permanent its caster controls and
// is modified by CR 700.9, Kodama of the West Tree's reading
// (b07IsModified: a counter, an Equipment, or an Aura its controller
// controls). The Aura being cast is not yet attached, so it does not
// modify its own target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9bdc1e63-1cf9-4ddb-a7d5-f2a2eb91bdfd",
		Name:            "Pearl-Ear, Imperial Advisor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		CostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsYouControl(HasSubtype("Aura")),
				"Enchantment spells you cast have affinity for Auras.",
				YourSpell(), socB10EnchantmentSpell()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, socB10AuraSpellTargetsYourModifiedPermanent,
				"Pearl-Ear, Imperial Advisor — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
