package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cad63b31-667b-4518-bc49-be4ffb82bc63",
		Name:         "Vraska, Soul of Stone",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.IsArtifact() && target.Controller == source.Controller
			}, "vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(),
				"Vraska, Soul of Stone — create a Sculpture Treasure token",
				effects.Do(effects.CreateToken{Template: effects.SculptureTreasureToken(), N: 1})),
		},
	})
}
