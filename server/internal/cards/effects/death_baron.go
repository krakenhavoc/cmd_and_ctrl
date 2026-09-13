package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Death Baron — Creature — Zombie Wizard {1}{B}{B}, 2/2 (EDHREC rank
// 1735):
//
//	"Skeletons you control and other Zombies you control get +1/+1
//	 and have deathtouch."
//
// The Zombie lord. Two statics over one AppliesTo
// (b16DeathBaronApplies): a Layer 7c +1/+1 and a Layer 6 deathtouch
// grant. "Other" applies to the Zombies only — the Baron is a Zombie,
// not a Skeleton, so it never buffs itself as printed; a Skeleton
// Zombie Baron (a changeling copy) would, also as printed. Effective
// subtypes, so a changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "99024aa8-5687-4d38-8a4b-feef42d6c1ff",
		Name:         "Death Baron",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(b16DeathBaronApplies, 1, 1),
			b16GrantKeywords(b16DeathBaronApplies, "deathtouch"),
		},
	})
}
