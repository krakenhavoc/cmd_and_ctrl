package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bane's Contingency — Instant {1}{U}{U} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"Counter target spell. If that spell targets a commander you control,
//	 instead counter that spell, scry 2, then draw a card."
//
// The spell's targets are read as the counter resolves, while it is still on
// the stack: a commander is a card the deck importer marked, and "you control"
// is the caster's, so a commander on the battlefield or on the stack qualifies,
// and a player or an opponent's commander does not. A spell that can't be
// countered stays and the scry and draw still happen, as printed ("instead
// counter that spell, scry 2, then draw" is one instruction list).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8089c3eb-99a8-40ca-882a-c1d4e30075ee",
		Name:         "Bane's Contingency",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2, Draws: 1},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			spellID := item.Targets[0].ID
			hitsCommander := false
			if spell := ctx.Game.StackItemForEffect(spellID); spell != nil {
				for _, t := range spell.Targets {
					if t.Kind != game.TargetCard {
						continue
					}
					if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok && c.IsCommander && c.Controller == ctx.Controller() {
						hitsCommander = true
					}
				}
			}
			if err := (effects.CounterTarget{StackID: spellID}).Apply(ctx); err != nil {
				return err
			}
			if !hitsCommander {
				return nil
			}
			you := ctx.Controller()
			return effects.Scry{Player: you, N: 2, ThenTail: game.KeyedTail{Ref: effects.ThenDrawACard, Player: you}}.Apply(ctx)
		},
	})
}
