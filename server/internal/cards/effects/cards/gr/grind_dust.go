package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grind // Dust — Sorcery // Sorcery with aftermath (Hour of Devastation):
//
//	Grind {1}{B}: "Put a -1/-1 counter on each of up to two target creatures."
//	Dust {3}{W}: "Aftermath (Cast this spell only from your graveyard. Then
//	 exile it.) Exile any number of target creatures that have -1/-1 counters
//	 on them."
//
// Dust's clause is judged on the creature's counters at announce and again as
// it resolves (CR 608.2b): a creature that lost its counters is not exiled.
func init() {
	const oracle = "ab5b697b-ab62-4ee7-a10e-cc937fef3095"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Grind",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Dust",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("any number of target creatures that have -1/-1 counters on them",
			func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Counters[game.CounterMinusOne] > 0 }).WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error { return effects.ExileEachLegalCardTarget(ctx) },
	})
}
