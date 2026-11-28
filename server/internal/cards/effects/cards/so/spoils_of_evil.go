package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spoils of Evil — Instant {2}{B} (Ice Age, S82 OLD batch 851, #5636):
//
//	"For each artifact or creature card in target opponent's graveyard, add {C}
//	 and you gain 1 life."
//
// Songs of the Damned's mana from a count, taken from the targeted opponent's
// graveyard as the spell resolves, with the same number of life gained. An
// opponent who is an illegal target by resolution makes the spell do nothing
// (CR 608.2b); a card that is both an artifact and a creature counts once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a40c52ee-4b45-4b14-8fbc-6529b42bfc9d",
		Name:         "Spoils of Evil",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim := effects.FirstLegalTargetPlayer(ctx)
			p := ctx.Game.PlayerByIDForEffect(victim)
			if p == nil || p.Graveyard == nil {
				return nil
			}
			n := 0
			for _, c := range p.Graveyard.Cards {
				if c.IsCreature() || c.IsArtifact() {
					n++
				}
			}
			if n == 0 {
				return nil
			}
			if err := (effects.AddMana{Amount: effects.ManaN(n, "C")}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: n}.Apply(ctx)
		},
	})
}
