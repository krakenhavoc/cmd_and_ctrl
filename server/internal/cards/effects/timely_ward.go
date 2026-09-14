package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Timely Ward — Enchantment — Aura for {2}{W} (EDHREC rank 1682):
//
//	"You may cast this spell as though it had flash if it targets a
//	 commander.
//	 Enchant creature
//	 Enchanted creature has indestructible."
//
// A permanent Swiftfoot-style shield for a commander: unlike a
// protection spell it does not wear off at end of turn, so the
// creature stays through every board wipe until somebody answers the
// Aura.
//
// Indestructible is honoured — S25 (#77) put it in the keyword table
// with a consumer in the destruction path — so the grant is real and
// not a badge. What it does NOT stop is the same list Darksteel Plate's
// file carries: sacrifice, exile, -X/-X to zero toughness, and the
// legend rule.
//
// ONE SIMPLIFICATION, strictly weaker: the conditional flash is not
// offered. "As though it had flash IF IT TARGETS A COMMANDER" is a
// cast-timing permission that depends on the target chosen during
// the cast it is gating — the engine's flash gate reads a keyword off
// the card before any target exists, and Spec.PrintedKeywords is
// unconditional, so declaring flash here would let the Aura be cast
// at instant speed onto ANY creature. That is stronger than printed,
// which is the one direction a simplification must never go. So the
// Aura is sorcery-speed only, which is a worse card than the printed
// one and never a better one.
func init() {
	Register(Spec{
		OracleID:     "8e589dca-d31e-4d2f-81b1-18e5f9198795",
		Name:         "Timely Ward",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"It can't be cast at instant speed on a commander — it's sorcery-speed only."},
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			GrantToAttached("indestructible"),
		},
	})
}
