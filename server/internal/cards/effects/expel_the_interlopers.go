package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expel the Interlopers — Sorcery {3}{W}{W} (EDHREC rank 2772):
//
//	"Choose a number between 0 and 10. Destroy all creatures with
//	 power greater than or equal to the chosen number."
//
// The adjustable wrath: 0 is Wrath of God, 4 spares the tokens and
// the mana dorks, 11 does not exist. The number is chosen as the spell
// RESOLVES (ChooseNumber, CR 608.2d), with the board in front of the
// caster and nothing for opponents to react to; the sweep is the mass
// destroy over "creature with power >= N", power read once the number
// is chosen, so the whole set leaves as one event and a Blood Artist
// caught in it sees every death.
//
// The bot's suggested number is the one that kills the most of the
// table's creatures and the fewest of its own; a person's stepper opens
// on it.
//
// No simplification. (The indestructible gap this file once declared
// was an engine gap, and S30 (#470) closed it for every "destroy all"
// in the catalog at once.)
func init() {
	Register(Spec{
		OracleID: "744b7e93-c893-4770-a3bb-6a45b7f8e4f7",
		Name:     "Expel the Interlopers",
		// ADR 0126 §6: only power at least the chosen number.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChooseNumber{
				Min:      0,
				Max:      10,
				Question: "Expel the Interlopers — choose a number between 0 and 10",
				Unit:     game.PayAmountOther,
				Goal:     expelTheInterlopersGoal,
				Then: func(ctx *Context, n int) error {
					return DestroyAllMatching{Match: And(Creature(), PowerGE(n))}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}

// expelTheInterlopersGoal is the number that destroys the most
// creatures the caster does not control for the fewest they do, ties
// to the higher number (it spares more of everything).
func expelTheInterlopersGoal(ctx *Context) int {
	me := ctx.Controller()
	best, bestScore := 10, -1<<30
	for n := 0; n <= 10; n++ {
		score := 0
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if !c.IsCreature() || c.CurrentPower() < n {
				continue
			}
			if c.Controller == me {
				score--
			} else {
				score++
			}
		}
		if score >= bestScore {
			best, bestScore = n, score
		}
	}
	return best
}
