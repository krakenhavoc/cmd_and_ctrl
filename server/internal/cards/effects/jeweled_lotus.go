package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jeweled Lotus — Artifact {0} (Commander Legends):
//
//	"{T}, Sacrifice this artifact: Add three mana of any one color. Spend
//	 this mana only to cast your commander."
//
// Lotus Bloom's mana ability (one colour pick for three mana) with Queen's
// Diamond's CR 106.6 spend restriction game.ManaRestrictYourCommander: the
// mana pays for a spell that is a commander its caster owns (CR 903.3d) and
// for nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23216e59-f147-4bb7-a698-12777c2f8584",
		Name:         "Jeweled Lotus",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced:     OneColorOfAmount(3),
			Restrictions: []string{ManaRestrictCast, game.ManaRestrictYourCommander},
			Label:        "{T}, Sacrifice this artifact: Add three mana of any one color. Spend this mana only to cast your commander",
		}},
	})
}
