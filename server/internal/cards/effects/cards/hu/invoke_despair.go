package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invoke Despair — Sorcery {1}{B}{B}{B}{B} (Kamigawa: Neon Dynasty, S82
// #5493):
//
//	"Target opponent sacrifices a creature of their choice. If they can't,
//	 they lose 2 life and you draw a card. Then repeat this process for an
//	 enchantment and a planeswalker."
//
// Three rounds in order (CR 608.2c): a creature, then an enchantment, then a
// planeswalker. Each round is Consuming Vapors' keyed sacrifice run, whose
// continuation (the run always finishes, even when nobody could be asked)
// reads what the opponent really sacrificed: nothing means they couldn't,
// so they lose 2 life and the caster draws, and then the next round starts.
// The round rides the run as a number; the caster and the opponent as its
// carry. A permanent they could sacrifice but whose sacrifice was replaced is
// not sacrificed, so it counts as "can't".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9ee4c80-20fb-4322-be02-1c78d37d7a99",
		Name:         "Invoke Despair",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return neoB15InvokeDespairRound(ctx.Game, ctx.Controller(), victim, ctx.Source(), 0)
		},
	})
}

// invokeDespairRounds are the three sacrifices, in printed order.
var invokeDespairRounds = []struct {
	label string
	pred  effects.CardPredicate
}{
	{"a creature", effects.Creature()},
	{"an enchantment", effects.Enchantment()},
	{"a planeswalker", effects.Planeswalker()},
}

// neoB15InvokeDespairRound asks round `round` (0-based) of the opponent's
// sacrifice. Past the last round there is nothing more to do.
func neoB15InvokeDespairRound(g *game.Game, me, victim, source uuid.UUID, round int) error {
	if round >= len(invokeDespairRounds) {
		return nil
	}
	r := invokeDespairRounds[round]
	return g.PlayerSacrificesThenKeyedDataForEffect(source, victim,
		effects.SacrificeSpec(r.label, r.pred),
		"Invoke Despair — sacrifice "+r.label, 1,
		invokeDespairAfter, game.RunData{Carry: []uuid.UUID{me, victim}, Nums: []int{round}})
}

// invokeDespairAfter is "if they can't, they lose 2 life and you draw a
// card", then the next round. An on-disk key: never renamed, never reused.
var invokeDespairAfter game.PromptRunThen

// The registration is made in init rather than as the var's initializer:
// the continuation starts the next round, which names the continuation.
func init() {
	invokeDespairAfter = game.RegisterPromptRunThen("prompt-run/neo-invoke-despair-after", func(g *game.Game, r game.RunDone) error {
		if len(r.Carry) < 2 || len(r.Nums) < 1 {
			return nil
		}
		me, victim, round := r.Carry[0], r.Carry[1], r.Nums[0]
		if len(r.Sacrifices().By(victim)) == 0 {
			ctx := effects.KeyedContext(g, me, r.Source)
			if p := g.PlayerByIDForEffect(victim); p != nil && !p.Eliminated {
				if err := g.ChangePlayerLifeForEffect(r.Source, victim, -2); err != nil {
					return err
				}
			}
			if err := (effects.DrawCards{Player: me, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return neoB15InvokeDespairRound(g, me, victim, r.Source, round+1)
	})
}
