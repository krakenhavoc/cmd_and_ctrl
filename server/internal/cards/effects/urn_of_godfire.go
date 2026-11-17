package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urn of Godfire — Artifact {1} (March of the Machine):
//
//	"{2}: Add one mana of any color.
//	 {6}, {T}, Sacrifice this artifact: Destroy target creature or
//	 enchantment."
//
// The first is a mana ability (CR 605.1a) paying generic mana, as on Barrels of
// Blasting Jelly. The second is a stack ability: the sacrifice is a cost, so the
// target is destroyed even if the Urn is gone by resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "420290d9-7bb9-4e5a-82ec-c777152ccc72",
		Name:         "Urn of Godfire",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Mana: "{2}"},
			Produced: "{W|U|B|R|G}",
			Label:    "{2}: Add one mana of any color.",
			Answers:  game.AnswerValue,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{6}, {T}, Sacrifice this artifact: Destroy target creature or enchantment.",
			Cost:    Plus(ManaCost("{6}"), TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target creature or enchantment", Or(Creature(), Enchantment())),
			Effect:  destroyFirstLegalCardTarget,
		}},
	})
}
