package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lashwhip Predator — Creature — Plant Beast {4}{G}{G}, 5/7 (Edge of
// Eternities, #4461):
//
//	"This spell costs {2} less to cast if your opponents control three or
//	 more creatures.
//	 Reach"
//
// Dreaded Bat-Cloud's conditional self-reduction (CostsLess, generic mana
// only, CR 601.2f). "Your opponents control" is the creatures of all your
// opponents together, so one creature each from three opponents is
// enough; a creature you control never counts. Read at announce.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "56035e66-e730-436d-8de2-4ded7fd1b51e",
		Name:            "Lashwhip Predator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if your opponents control three or more creatures.",
				func(q game.CostQuery) bool {
					return q.Game != nil && lashwhipCreaturesYourOpponentsControl(q.Game, q.Controller) >= 3
				}),
		},
	})
}

// lashwhipCreaturesYourOpponentsControl counts the creatures controlled by
// players other than you.
func lashwhipCreaturesYourOpponentsControl(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you && c.IsCreature() {
			n++
		}
	}
	return n
}
