package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// By Invitation Only — Sorcery {3}{W}{W} (Innistrad: Crimson Vow, S82):
//
//	"Choose a number between 0 and 13. Each player sacrifices that many
//	 creatures of their choice."
//
// The caster picks the number as it resolves (Expel the Interlopers'
// ChooseNumber); then every player, in turn order starting with the
// active player, sacrifices that many creatures of their own choice
// (Blasphemous Edict's PlayerSacrificesNForEffect). A player with fewer
// creatures sacrifices all they have. A sacrifice is not destruction, so
// indestructible does not help (CR 701.21a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d64e96bf-6815-462c-93b4-c9c22dc43eda",
		Name:         "By Invitation Only",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: only as many creatures as the chosen number.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseNumber{
				Min:      0,
				Max:      13,
				Question: "By Invitation Only — choose a number between 0 and 13",
				Unit:     game.PayAmountOther,
				Goal:     byInvitationOnlyGoal,
				Then:     byInvitationOnlySacrifices,
			}.Apply(ctx)
		},
	})
}

var byInvitationOnlySacrifices = effects.NumberThen("number/by-invitation-only-sacrifice",
	func(ctx *effects.Context, r game.NumberAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		for _, p := range effects.SeatsFromActive(ctx.Game) {
			ctx.Game.PlayerSacrificesNForEffect(ctx.Source(), p,
				effects.SacrificeSpec("a creature", effects.Creature()),
				"By Invitation Only — sacrifice creatures", r.Number)
		}
		return nil
	})

// byInvitationOnlyGoal is the bot's pick: the number that costs the
// opponents the most creatures net of what it costs the caster.
func byInvitationOnlyGoal(ctx *effects.Context) int {
	me := ctx.Controller()
	mine, theirs := 0, []int{}
	counts := map[uuid.UUID]int{}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if !c.IsCreature() {
			continue
		}
		if c.Controller == me {
			mine++
		} else {
			counts[c.Controller]++
		}
	}
	for _, n := range counts {
		theirs = append(theirs, n)
	}
	best, bestScore := 0, 0
	for n := 0; n <= 13; n++ {
		score := -min(n, mine)
		for _, t := range theirs {
			score += min(n, t)
		}
		if score > bestScore {
			best, bestScore = n, score
		}
	}
	return best
}
