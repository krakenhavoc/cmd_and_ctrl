package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lignify — Kindred Enchantment — Treefolk Aura {1}{G} (Lorwyn, #5593):
//
//	"Enchant creature
//	 Enchanted creature is a Treefolk with base power and toughness 0/4
//	 and loses all abilities."
//
// Darksteel Mutation's three statics in Lignify's words. "Is a Treefolk"
// replaces the creature types alone (layer 4, CR 613.1d), so an artifact
// creature stays an artifact and a changeling stops being every type; the
// creature loses every ability in layer 6 (CR 613.1f); and its base power
// and toughness are 0/4 in layer 7b (CR 613.4b), under any pump or +1/+1
// counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "63625b26-3a19-4c3c-b18a-91247bb2a40c",
		Name:         "Lignify",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			SetAttachedCreatureTypes("Treefolk"),
			LoseAllAbilities(),
			SetAttachedBasePT(0, 4),
		},
	})
}
