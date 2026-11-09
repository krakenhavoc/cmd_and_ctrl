package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Fourth Sphere — Plane — Phyrexia (Planechase Anthology, #4870):
//
//	"At the beginning of your upkeep, sacrifice a nonblack creature.
//	 Whenever chaos ensues, create a 2/2 black Zombie creature token."
//
// A plane (ADR 0216 §6). The planar controller chooses which nonblack
// creature they control to sacrifice, judged by color as it resolves; with
// none, nothing happens (CR 701.21a). The Zombie is the planar
// controller's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13e801cd-e537-4ea6-8c3d-21b142bdc026",
		Name:         "The Fourth Sphere",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("The Fourth Sphere — sacrifice a nonblack creature",
				func(g *game.Game, item *game.StackItem) error {
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && !b9IsBlack(c) {
							ids = append(ids, c.InstanceID)
						}
					}
					return SacrificeChoice{
						Player:     item.Controller,
						Candidates: ids,
						Question:   "The Fourth Sphere — sacrifice a nonblack creature",
					}.Apply(NewContext(g, item))
				}),
			WheneverChaosEnsues("The Fourth Sphere — create a 2/2 black Zombie creature token",
				Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1})),
		},
	})
}
