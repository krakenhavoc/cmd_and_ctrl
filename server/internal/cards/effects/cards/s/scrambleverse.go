package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scrambleverse — Sorcery {6}{R}{R} (Magic 2012, S82 OLD batch 452, #5573):
//
//	"For each nonland permanent, choose a player at random. Then each
//	 player gains control of each permanent for which they were chosen.
//	 Untap those permanents."
//
// Every nonland permanent on the battlefield gets one draw from the keyed
// pick stream (ChooseAtRandomForEffect, so an undo rewinds it) among the
// players still in the game, all of them before any control changes. Then
// each permanent whose chosen player is not its controller changes hands
// with no stated duration (CR 611.2a), and every one of them is untapped,
// including those that stayed with their controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35fe3c61-8eaa-465d-ac34-63efda72fef2",
		Name:         "Scrambleverse",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			var players []uuid.UUID
			for _, p := range g.Seats {
				if p != nil && !p.Eliminated {
					players = append(players, p.ID)
				}
			}
			if len(players) == 0 {
				return nil
			}
			type pick struct{ perm, player uuid.UUID }
			var picks []pick
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.IsLand() {
					continue
				}
				chosen := g.ChooseAtRandomForEffect(effects.RandomDraw(ctx), players, 1)
				if len(chosen) == 0 {
					continue
				}
				picks = append(picks, pick{perm: c.InstanceID, player: chosen[0]})
			}
			for _, p := range picks {
				if c, ok := g.LookupCardForEffect(p.perm); ok && c.Controller != p.player {
					if err := (effects.GainControl{
						Target:     p.perm,
						Controller: p.player,
						Duration:   game.IndefiniteDuration(),
						Label:      "Scrambleverse — gain control (no stated duration)",
					}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
			}
			for _, p := range picks {
				if err := (effects.UntapTarget{Target: p.perm}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
