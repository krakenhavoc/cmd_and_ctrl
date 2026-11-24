package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daemogoth Titan — Creature — Demon {B/G}{B/G}{B/G}{B/G}, 11/10 (Strixhaven:
// School of Mages, #5502):
//
//	"Whenever this creature attacks or blocks, sacrifice a creature."
//
// Morkrut Necropod's attacks-or-blocks trigger (one trigger however many
// attackers it blocks, CR 509.3a) with any creature its controller controls as
// the candidate set, the Titan included: the choice is made as the ability
// resolves, and it is mandatory (CR 701.21a), so a lone Titan sacrifices
// itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a52e73bb-47cb-4912-9b81-4ae40641eabb",
		Name:         "Daemogoth Titan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
				},
				"Daemogoth Titan — sacrifice a creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificeChoice{
						Player:     item.Controller,
						Candidates: effects.PermanentsControlledByMatching(g, item.Controller, effects.Creature()),
						Question:   "Daemogoth Titan — sacrifice a creature",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
