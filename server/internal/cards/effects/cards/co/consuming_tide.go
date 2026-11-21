package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Consuming Tide — Sorcery {2}{U}{U} (Innistrad: Crimson Vow, S82):
//
//	"Each player chooses a nonland permanent they control. Return all
//	 nonland permanents not chosen this way to their owners' hands. Then
//	 you draw a card for each opponent who has more cards in their hand
//	 than you."
//
// Every player with two or more nonland permanents is asked, in APNAP
// order (CR 101.4); a player with exactly one has nothing to choose and
// keeps it, and one with none chooses nothing. Choices are made before
// anything returns, then every other nonland permanent goes back in one
// simultaneous move (the Evacuation shape). The draw counts hands AFTER
// the return, since it follows "Then".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24cd8f54-c1d4-4040-93bf-a7b9b719ac30",
		Name:         "Consuming Tide",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: each player keeps one nonland permanent.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepBounce, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			var legs []game.ChooseCardsPrompt
			var asked []uuid.UUID
			for _, seat := range effects.ApnapPlayers(g) {
				perms := effects.PermanentsControlledByMatching(g, seat, effects.Nonland())
				if len(perms) < 2 {
					continue
				}
				asked = append(asked, seat)
				legs = append(legs, game.ChooseCardsPrompt{
					Chooser:    seat,
					FromPlayer: seat,
					Source:     item.SourceCardID,
					Question:   "Consuming Tide — choose a nonland permanent you control; the rest return to their owners' hands",
					Cards:      perms,
					Min:        1,
					Max:        1,
					Zone:       game.ZoneBattlefield,
					ThenKey:    game.ChooseCardsRunLeg,
				})
			}
			if len(legs) == 0 {
				return consumingTideFinish(ctx, nil, nil)
			}
			_, err := g.ChooseCardsRunThenKeyedForEffect(legs, consumingTideReturn, item.SourceCardID,
				game.RunData{Carry: asked, Asked: game.NewAskedItem(item)})
			return err
		},
	})
}

var consumingTideReturn = effects.RunThen("run/consuming-tide-return", func(ctx *effects.Context, r game.RunDone) error {
	return consumingTideFinish(ctx, r.Carry, r.Picks().Cards())
})

// consumingTideFinish returns every nonland permanent the asked seats did
// not keep (the other seats had at most one, which they keep), then draws
// for each opponent with more cards in hand than the caster.
func consumingTideFinish(ctx *effects.Context, asked, kept []uuid.UUID) error {
	keep := map[uuid.UUID]bool{}
	for _, id := range kept {
		keep[id] = true
	}
	var rest []uuid.UUID
	for _, seat := range asked {
		for _, id := range effects.PermanentsControlledByMatching(ctx.Game, seat, effects.Nonland()) {
			if !keep[id] {
				rest = append(rest, id)
			}
		}
	}
	if len(rest) > 0 {
		if err := (effects.ReturnAllToHand{Cards: rest}).Apply(ctx); err != nil {
			return err
		}
	}
	me := ctx.Controller()
	mine := effects.B14HandSize(ctx.Game, me)
	n := 0
	effects.EachOpponent(ctx.Game, me, func(opp uuid.UUID) bool {
		if effects.B14HandSize(ctx.Game, opp) > mine {
			n++
		}
		return false
	})
	return effects.DrawCards{Player: me, N: n}.Apply(ctx)
}
