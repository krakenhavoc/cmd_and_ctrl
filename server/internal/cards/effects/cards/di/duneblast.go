package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Duneblast — Sorcery {4}{W}{B}{G} (Commander deck reprint):
//
//	"Choose up to one creature. Destroy the rest."
//
// The creature is CHOSEN as the spell resolves, not targeted (it may be an
// opponent's, hexproof or not), by the caster, from every creature on the
// battlefield, with a floor of none. Everything else is destroyed as one
// simultaneous event; regeneration is allowed, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "14104a52-c4e8-4755-8196-7fa5c5a35f36",
		Name:     "Duneblast",
		// ADR 0126 §6: the one creature the caster chooses stays.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChoosePermanents{
				Question:   "Duneblast — choose up to one creature; destroy the rest",
				Candidates: duneblastCandidates,
				Keyed:      duneblastDestroyTheRest,
			}.Apply(ctx)
		},
	})
}

var duneblastDestroyTheRest = effects.RegisterPermanentsPick("prompt-run/duneblast-destroy-the-rest", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		return effects.DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.IsCreature() && !picked.Contains(c.InstanceID)
		}}.Apply(ctx)
	},
})

// duneblastCandidates is every creature on the battlefield, floor 0, ceiling 1.
//
// Caller holds g.mu.
func duneblastCandidates(g *game.Game, _ uuid.UUID) ([]uuid.UUID, int, int) {
	var ids []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids, 0, 1
}
