package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aid from the Cowl — Enchantment {3}{G}{G}:
//
//	"Revolt — At the beginning of your end step, if a permanent left the
//	 battlefield under your control this turn, reveal the top card of your
//	 library. If it's a permanent card, you may put it onto the battlefield.
//	 Otherwise, you may put it on the bottom of your library."
//
// The trigger is the shared revolt end-step shape (AtYourEndStepIfRevolt),
// an intervening "if" checked as the step begins and again on resolution
// (CR 603.4). A revealed token is not a permanent card (CR 108.2) and falls
// on the "otherwise" side, where it can't be moved (CR 111.8). Each
// "you may" is a yes/no put to the controller as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9597c974-6120-43a7-9ad3-7e4fddc8c9f5",
		Name:         "Aid from the Cowl",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfRevolt("Aid from the Cowl — reveal the top card of your library; a permanent card may go onto the battlefield, anything else may go on the bottom", aidFromTheCowlReveal),
		},
	})
}

func aidFromTheCowlReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 1,
		"Aid from the Cowl — reveal the top card of your library")
	if len(revealed) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(revealed[0])
	if !ok {
		return nil
	}
	if isPermanentCard(c) {
		return MayChoice{
			Player:   ctx.Controller(),
			Question: "Aid from the Cowl — put " + c.Name + " onto the battlefield?",
			YesLabel: "Put it onto the battlefield",
			NoLabel:  "Leave it on top",
			Then:     aidFromTheCowlPut,
			Carry:    []uuid.UUID{revealed[0]},
		}.Apply(ctx)
	}
	if IsToken(c) {
		return nil
	}
	return MayChoice{
		Player:   ctx.Controller(),
		Question: "Aid from the Cowl — put " + c.Name + " on the bottom of your library?",
		YesLabel: "Put it on the bottom",
		NoLabel:  "Leave it on top",
		Then:     aidFromTheCowlBottom,
		Carry:    []uuid.UUID{revealed[0]},
	}.Apply(ctx)
}

var aidFromTheCowlPut = MayYes("s82-rpr-b04/aid-from-the-cowl-put", func(ctx *Context, r game.ConfirmAnswer) error {
	id := carryAt(r, 0)
	if !bisonWhistleStillInLibrary(ctx.Game, id) {
		return nil
	}
	_, err := ctx.Game.PutFromLibraryOntoBattlefieldForEffect(id, game.LibraryEntryOptions{Controller: r.Chooser})
	return err
})

var aidFromTheCowlBottom = MayYes("s82-rpr-b04/aid-from-the-cowl-bottom", func(ctx *Context, r game.ConfirmAnswer) error {
	id := carryAt(r, 0)
	if !bisonWhistleStillInLibrary(ctx.Game, id) {
		return nil
	}
	return ctx.Game.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, []uuid.UUID{id})
})
