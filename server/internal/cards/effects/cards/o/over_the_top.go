package cardso

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Over the Top — Sorcery {5}{R}{R} (The Brothers' War, #5482):
//
//	"Each player reveals a number of cards from the top of their library
//	 equal to the number of nonland permanents they control, puts all
//	 permanent cards they revealed this way onto the battlefield, and puts
//	 the rest into their graveyard."
//
// Genesis Wave's reveal and put, once per player. Every player's number is
// read, and every player's cards revealed, in one first instruction before
// anything enters (CR 608.2c: the numbers are what each player controls as
// the spell begins to resolve, so permanents an earlier player's cards put
// onto the battlefield don't change a later player's count). The puts follow
// one instruction per seat in turn order (ResolveSteps), each waiting for
// the replacement prompts the one before opened. A player's permanent cards
// enter together, then their other revealed cards go to their graveyard (not
// a mill: PutRestIntoGraveyard). A number of 0 reveals nothing.
//
// No simplification.
func init() {
	steps := []effects.Applier{effects.Instruction(overTheTopReveal)}
	for i := 0; i < game.MaxPlayers; i++ {
		steps = append(steps, effects.Instruction(overTheTopPut(i)))
	}
	effects.Register(effects.Spec{
		OracleID:     "b82d8bdf-e292-490a-aee7-d830cc2c29df",
		Name:         "Over the Top",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve:    effects.ResolveSteps(steps...),
	})
}

func overTheTopKey(i int) string { return "over-the-top/" + strconv.Itoa(i) }

// overTheTopReveal reveals each player's cards, in turn order, and files
// the revealed ids under the seat's key.
func overTheTopReveal(ctx *effects.Context) error {
	g := ctx.Game
	players := effects.ApnapPlayers(g)
	counts := make([]int, len(players))
	for i, p := range players {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == p && !c.IsLand() {
				counts[i]++
			}
		}
	}
	for i, p := range players {
		var revealed = g.RevealTopOfLibraryForEffect(p, ctx.Source(), counts[i],
			"Over the Top — revealed from the top of the library")
		ctx.RecordStep(overTheTopKey(i), revealed)
	}
	return nil
}

// overTheTopPut is the i-th player's "puts all permanent cards they
// revealed this way onto the battlefield, and puts the rest into their
// graveyard".
func overTheTopPut(i int) func(ctx *effects.Context) error {
	return func(ctx *effects.Context) error {
		players := effects.ApnapPlayers(ctx.Game)
		if i >= len(players) {
			return nil
		}
		revealed := ctx.StepRecord(overTheTopKey(i)).IDs
		if len(revealed) == 0 {
			return nil
		}
		return effects.PutFromLibraryOntoBattlefield{
			Player:   players[i],
			Cards:    revealed,
			All:      true,
			Label:    "Over the Top — put all permanent cards revealed this way onto the battlefield",
			ThenTail: effects.PutRestIntoGraveyardTail,
		}.Apply(ctx)
	}
}
