package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vraska, Soul of Stone — Legendary Creature — Gorgon Wizard {U}{R}{W},
// 3/3:
//
//	"Artifact creatures you control have vigilance.
//	 Whenever you cast a noncreature spell, create a 1/1 colorless
//	 Sculpture Treasure artifact creature token with '{T}, Sacrifice
//	 this token: Add one mana of any color.'"
//
// The vigilance is a layer 6 grant (it covers the Sculptures it makes).
// The token is a catalog template (SculptureTreasureToken) and an
// artifact creature, so its mana ability waits out summoning sickness.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cad63b31-667b-4518-bc49-be4ffb82bc63",
		Name:         "Vraska, Soul of Stone",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.IsArtifact() && target.Controller == source.Controller
			}, "vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(),
				"Vraska, Soul of Stone — create a Sculpture Treasure token",
				Do(CreateToken{Template: SculptureTreasureToken(), N: 1})),
		},
	})
}
