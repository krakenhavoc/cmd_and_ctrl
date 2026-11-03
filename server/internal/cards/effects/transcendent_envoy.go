package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Transcendent Envoy — Enchantment Creature — Griffin {1}{W}, 1/2
// (Secrets of Strixhaven Commander, #4834):
//
//	"Flying
//	 Aura spells you cast cost {1} less to cast."
//
// A cost reducer (CR 601.2f) over the controller's own Aura spells, the
// Danitha Capashen shape with the Equipment half left out. An Aura spell is a
// spell whose card has the Aura subtype.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f08f7814-4585-41dd-b82e-ed2ef5e24242",
		Name:            "Transcendent Envoy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Aura spells you cast cost {1} less to cast.",
				YourSpell(), socB14AuraSpell()),
		},
	})
}
