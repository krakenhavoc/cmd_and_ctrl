package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curfew — Instant {U} (Urza's Saga, S82 OLD batch 729, #5624):
//
//	"Each player returns a creature they control to its owner's hand."
//
// Untargeted, so nothing can fizzle it and hexproof is irrelevant. Each
// player chooses one creature they control as it resolves, in APNAP order
// (CR 101.4), each seeing the earlier choices; a player with one creature is
// not asked and one with none returns nothing. The chosen creatures return
// together after the last choice (CR 608.2e).
//
// No simplification.
func init() {
	curfewPicked = effects.RegisterPermanentsPick("prompt-run/curfew-return-a-creature", effects.PermanentsDef{
		ThenRun: func(ctx *effects.Context, picked game.PromptedPicks, r game.RunDone) error {
			rest, chosen := effects.UnpackIDPair(r.Carry)
			return curfewAskNext(ctx, rest, append(chosen, picked.Cards()...))
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "6ff1e793-1569-4887-b6bb-d9855c5e8a26",
		Name:         "Curfew",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return curfewAskNext(ctx, effects.ApnapPlayers(ctx.Game), nil)
		},
	})
}

// curfewPicked is one player's pick answered: the next player is asked, or
// the chosen creatures return together. Set in init because the answer asks
// the next player, which names the key again. An on-disk key.
var curfewPicked effects.PermanentsThen

// curfewAskNext picks for the first player in `order` who has a real choice,
// carrying the players still to ask and the creatures picked so far.
func curfewAskNext(ctx *effects.Context, order, chosen []uuid.UUID) error {
	g := ctx.Game
	for i, p := range order {
		mine, _, _ := effects.OldB729OwnCreatureCandidates(g, p)
		switch len(mine) {
		case 0:
			continue
		case 1:
			chosen = append(chosen, mine[0])
			continue
		}
		return effects.ChoosePermanents{
			Player:     p,
			Of:         []uuid.UUID{p},
			Question:   "Curfew — return a creature you control to its owner's hand",
			Candidates: effects.OldB729OwnCreatureCandidates,
			Keyed:      curfewPicked,
			Carry:      effects.PackIDLists(order[i+1:], chosen),
		}.Apply(ctx)
	}
	return effects.ReturnAllToHand{Cards: chosen}.Apply(ctx)
}
