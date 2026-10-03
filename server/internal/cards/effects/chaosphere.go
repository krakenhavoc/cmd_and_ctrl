package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chaosphere — World Enchantment {2}{R}:
//
//	"Creatures with flying can block only creatures with flying.
//	 Creatures without flying have reach. (They can block creatures
//	 with flying.)"
//
// The first line is Departed Soulkeeper's block rule over every
// creature with flying, every player's (#750): such a blocker may not
// be paired with an attacker without flying. The second is a layer-6
// keyword grant to every creature without flying, so a ground creature
// can block a flier. Both read flying as layer 6 leaves it, so a
// creature that loses flying to a Gravity Sphere blocks as a ground
// creature and gets reach.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed01d5d7-8f34-47b3-9ca8-d82c242d38b4",
		Name:         "Chaosphere",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnMatching(HasKeyword("flying")), OnMatching(WithoutKeyword("flying")),
				"creatures with flying can block only creatures with flying"),
		},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, _ *game.Card) bool {
				return target.IsCreature() && WithoutKeyword("flying")(g, uuid.Nil, *target)
			}, "reach"),
		},
	})
}
