package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Uncivil Unrest — Enchantment {4}{R} (EDHREC rank 2954):
//
//	"Nontoken creatures you control have riot. (They enter with your
//	 choice of a +1/+1 counter or haste.)
//	 If a creature you control with a +1/+1 counter on it would deal
//	 damage to a permanent or player, it deals double that damage
//	 instead."
//
// The Gruul damage doubler. Both halves are CR 614 replacements on
// the enchantment:
//
//   - Riot is an entry replacement over every nontoken creature
//     entering under the controller's control
//     (b28PermanentsYouControlEnterWithACounter with a nontoken
//     test), so the counter is on the creature before its enters
//     trigger fires and a Hardened Scales applies on top.
//   - The doubling is a damage replacement gated on the SOURCE: a
//     creature the controller controls that has a +1/+1 counter at
//     the moment the damage would be dealt, combat and noncombat
//     alike, to any permanent or player
//     (b28DoubleDamageFromCounteredCreaturesYouControl). CR 616:
//     with a second doubler the affected player orders them.
//
// DECLARED SIMPLIFICATION, weaker than printed: riot always gives
// the +1/+1 counter; the haste option is not offered. The choice is
// made as the creature enters, which means a prompt inside the entry
// pipeline and a PendingChoiceKind to carry the answer; #478 gave the
// effect-side entries a resume, so the blocker is now only that the
// prompt itself does not exist. A counter every time is one of the two
// printed outcomes, never a third, and it is the half the second
// ability wants.
func init() {
	Register(Spec{
		OracleID:     "bd655e8b-f192-4635-9e23-357b6f89ef8f",
		Name:         "Uncivil Unrest",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Riot always puts a +1/+1 counter on the creature — the haste option isn't offered."},
		Replacements: []game.ReplacementEffect{
			b28PermanentsYouControlEnterWithACounter("Uncivil Unrest: riot — a +1/+1 counter",
				func(c game.Card) bool { return c.IsCreature() && !IsToken(c) }),
			b28DoubleDamageFromCounteredCreaturesYouControl("Uncivil Unrest: double damage from a creature with a +1/+1 counter"),
		},
	})
}
