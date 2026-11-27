package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soulquake — Sorcery {3}{U}{U}{B}{B} (Alara Reborn, #5586):
//
//	"Return all creatures on the battlefield and all creature cards in
//	 graveyards to their owners' hands."
//
// Evacuation's mass bounce for the battlefield half (tokens cease to exist) and,
// for the graveyard half, every creature card in every graveyard read before
// anything moves and returned to its owner's hand. Symmetric.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b9a477b-f2dd-47b2-8493-eb63f44a2688",
		Name:         "Soulquake",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var cards []uuid.UUID
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Graveyard == nil {
					continue
				}
				for _, c := range p.Graveyard.Cards {
					if c.IsCreature() && !c.IsToken() {
						cards = append(cards, c.InstanceID)
					}
				}
			}
			if err := (effects.BounceAllMatching{Match: effects.Creature()}).Apply(ctx); err != nil {
				return err
			}
			for _, id := range cards {
				if _, err := ctx.Game.ReturnFromAnyZoneToHandForEffect(ctx.Controller(), ctx.Source(), id); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
