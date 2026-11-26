package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scrapyard Salvo — Sorcery {1}{R}{R} (New Phyrexia, #5575):
//
//	"Scrapyard Salvo deals damage to target player or planeswalker equal to the
//	 number of artifact cards in your graveyard."
//
// The count is the artifact cards (by printed type; a card in a graveyard has no
// layers) in the caster's graveyard as the spell resolves. A target that became
// illegal means the spell does nothing (CR 608.2b); with none the damage is
// zero and nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f63894a8-318f-4cfb-9a17-359fb9153e74",
		Name:         "Scrapyard Salvo",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			p := ctx.Game.PlayerByIDForEffect(item.Controller)
			if p == nil || p.Graveyard == nil {
				return nil
			}
			n := 0
			for _, c := range p.Graveyard.Cards {
				if c.IsArtifact() {
					n++
				}
			}
			for _, t := range ctx.LegalTargets() {
				if t.ID == uuid.Nil || n <= 0 {
					continue
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
			}
			return nil
		},
	})
}
