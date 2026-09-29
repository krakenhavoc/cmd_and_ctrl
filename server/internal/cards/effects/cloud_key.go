package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloud Key — Artifact {3}:
//
//	"As this artifact enters, choose artifact, creature, enchantment,
//	 instant, or sorcery.
//	 Spells you cast of the chosen type cost {1} less to cast."
//
// The choice is CR 614.12's "choose <A>, <B>, …" over words PRINTED on
// the card — the same shape the Siege cycle's ChooseOptionAsEnters
// uses, just five options instead of two and read back by a cost
// predicate instead of an ability gate. ChosenOptionOf is the reader
// choose_option.go built for exactly that ("for an effect that has to
// say the word"); this is the first card that needs it.
//
// Until the controller answers, ChosenOptionOf returns "", which
// cloudKeyMatchesChosenType reads as "match nothing" — the safe
// direction, since an empty type read as "every type" would make the
// Key a strictly-better Cloud Key for the one moment between entering
// and answering.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a838818-d590-4374-9a63-d9e6381a0f0d",
		Name:         "Cloud Key",
		Completeness: CompletenessFull,
		AsEnters: ChooseOptionAsEnters("Cloud Key",
			"artifact", "creature", "enchantment", "instant", "sorcery"),
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Spells you cast of the chosen type cost {1} less to cast.",
				YourSpell(), cloudKeyChosenTypeSpell()),
		},
	})
}

// cloudKeyChosenTypeSpell reads the word Cloud Key's controller chose
// as it entered and matches the spell being priced against it.
func cloudKeyChosenTypeSpell() CostPredicate {
	return func(q game.CostQuery) bool {
		return cloudKeyMatchesChosenType(ChosenOptionOf(q.Game, q.Source.InstanceID), q.Card)
	}
}

func cloudKeyMatchesChosenType(chosen string, c game.Card) bool {
	switch chosen {
	case "artifact":
		return c.IsArtifact()
	case "creature":
		return c.IsCreature()
	case "enchantment":
		return c.IsEnchantment()
	case "instant":
		return c.IsInstant()
	case "sorcery":
		return c.IsSorcery()
	default:
		return false
	}
}
