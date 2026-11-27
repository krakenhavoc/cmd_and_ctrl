package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strip Bare — Instant {W} (Shadowmoor, #5591):
//
//	"Destroy all Auras and Equipment attached to target creature."
//
// Everything attached to the creature as the spell resolves that is an Aura
// or an Equipment is destroyed together, whoever controls it. Other
// permanents attached to it (a Fortification cannot be) stay. A target that
// has become illegal makes the spell fizzle (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "957e8ded-90e7-4bc6-98c1-eb6af1fd0ce5",
		Name:         "Strip Bare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			host, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			g := ctx.Game
			var ids []uuid.UUID
			for _, id := range g.AttachmentsOf(host) {
				if c, found := g.LookupCardForEffect(id); found && (c.HasSubtype("Aura") || c.HasSubtype("Equipment")) {
					ids = append(ids, id)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			g.DestroyPermanentsForEffect(ids, game.DestroyOptions{})
			return nil
		},
	})
}
