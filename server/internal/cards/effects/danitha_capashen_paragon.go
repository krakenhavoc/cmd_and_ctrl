package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Danitha Capashen, Paragon — Legendary Creature — Human Knight
// {2}{W}, 2/2:
//
//	"First strike, vigilance, lifelink
//	 Aura and Equipment spells you cast cost {1} less to cast."
//
// The three keywords ride PrintedKeywords. The discount names two
// subtypes rather than a card type — HasSubtype("Aura") for an
// enchantment spell, HasSubtype("Equipment") for an artifact spell —
// so it's a small closure local to this card rather than a shared
// ArtifactOrEnchantmentSpell()-style predicate, which would also (and
// wrongly) discount a plain enchantment or a plain artifact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4b6377da-83e7-4519-9582-16a9c16b8faa",
		Name:            "Danitha Capashen, Paragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance", "lifelink"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Aura and Equipment spells you cast cost {1} less to cast.",
				YourSpell(), auraOrEquipmentSpell()),
		},
	})
}

func auraOrEquipmentSpell() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.HasSubtype("Aura") || q.Card.HasSubtype("Equipment")
	}
}
