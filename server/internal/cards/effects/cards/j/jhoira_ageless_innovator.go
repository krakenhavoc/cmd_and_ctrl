package cardsj

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jhoira, Ageless Innovator — Legendary Creature — Human Artificer {U}{R},
// 2/3 (Dominaria United, #5487):
//
//	"{T}: Put two ingenuity counters on Jhoira, then you may put an artifact
//	 card with mana value X or less from your hand onto the battlefield, where
//	 X is the number of ingenuity counters on Jhoira."
//
// The counters go on first and X is read after them, so the first activation
// offers mana value 2 or less. X is Jhoira's counters as it resolves, by
// last-known information if she has left the battlefield (CR 608.2h), in which
// case no counters are placed and X is whatever she last had. The card is put
// onto the battlefield, not cast, so it pays nothing, and the put may be declined.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83b9a734-ee92-44c8-bf47-ca69b6d6405e",
		Name:         "Jhoira, Ageless Innovator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Put two ingenuity counters on Jhoira, then you may put an artifact card with mana value X or less from your hand onto the battlefield, where X is the number of ingenuity counters on Jhoira.",
			// The artifact put onto the battlefield may be a creature that
			// blocks, so the ability can answer a combat (ADR 0142).
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if effects.OnBattlefield(g, ctx.Source()) {
					if err := (effects.AddCounter{Target: ctx.Source(), Kind: effects.LadyOctopusCounter, N: 2}).Apply(ctx); err != nil {
						return err
					}
				}
				x := effects.LadyOctopusCounters(g, ctx.Source())
				return effects.PutFromHandOntoBattlefield{
					Match:    effects.And(effects.Artifact(), effects.ManaValueLE(x)),
					Optional: true,
					Label:    "Jhoira, Ageless Innovator — you may put an artifact card with mana value " + strconv.Itoa(x) + " or less from your hand onto the battlefield",
				}.Apply(ctx)
			},
		}},
	})
}
