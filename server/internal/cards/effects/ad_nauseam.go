package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ad Nauseam — Instant {3}{B}{B} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Reveal the top card of your library and put that card into your
//	 hand. You lose life equal to its mana value. You may repeat this
//	 process any number of times."
//
// One pass is Dark Confidant's flip: the reveal is a real reveal and not
// a draw (CR 701.20a), so nothing that watches draws sees it. The mana
// value is read off the revealed card before it moves, then it goes to
// your hand, then you lose that much life. Each pass is three steps, so
// the life loss waits for any prompt the move asked and the next question
// waits for any prompt the loss asked (CR 608.2c, #4155).
//
// "You may repeat this process any number of times" is asked after every
// pass, and a yes runs the same three steps again (Dance with
// Calamity's loop). Nothing stops the loop but the player: state-based
// actions are not checked mid-resolution, so you can reveal past zero
// life and lose when the spell is done (CR 704.5a), as printed. With an
// empty library a pass reveals nothing, loses nothing and a repeat could
// do nothing more, so it asks no more.
//
// No simplification.
func init() {
	// Registered here, not in the var below: the answer runs the pass,
	// which names this key (an initialization cycle otherwise).
	adNauseamRepeat = MayYes("confirm/ad-nauseam-repeat", func(ctx *Context, _ game.ConfirmAnswer) error {
		return adNauseamPass.Run(ctx)
	})
	Register(Spec{
		OracleID:     "981b0e21-e5e6-4a1e-bfde-679d56623f7f",
		Name:         "Ad Nauseam",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(adNauseamPass),
	})
}

// adNauseamRecord files the pass's revealed card and its mana value.
const adNauseamRecord = "ad-nauseam/pass"

// adNauseamRepeat is the "yes" to another pass. An on-disk key, assigned
// in init.
var adNauseamRepeat game.ConfirmThen

// adNauseamPass is one pass of the process: reveal and take, lose, ask.
var adNauseamPass = RegisterThenSteps("steps/ad-nauseam-pass",
	Instruction(func(ctx *Context) error {
		var revealed []uuid.UUID
		if err := (RevealTopOfLibrary{
			Player:   ctx.Controller(),
			N:        1,
			Reason:   "Ad Nauseam — reveal the top card of your library",
			Revealed: &revealed,
		}).Apply(ctx); err != nil {
			return err
		}
		if len(revealed) == 0 {
			ctx.RecordStep(adNauseamRecord, nil, 0)
			return nil
		}
		value := 0
		if c, ok := ctx.Game.LookupCardForEffect(revealed[0]); ok {
			value = c.ManaValue()
		}
		ctx.RecordStep(adNauseamRecord, revealed, value)
		return BounceToHand{Target: revealed[0]}.Apply(ctx)
	}),
	Instruction(func(ctx *Context) error {
		if n := ctx.StepRecord(adNauseamRecord).N(); n > 0 {
			return GainLife{Amount: -n}.Apply(ctx)
		}
		return nil
	}),
	Instruction(func(ctx *Context) error {
		if len(ctx.StepRecord(adNauseamRecord).IDs) == 0 {
			return nil
		}
		return MayChoice{
			Question: "Ad Nauseam — repeat the process?",
			YesLabel: "Reveal another card",
			NoLabel:  "Stop",
			Then:     adNauseamRepeat,
		}.Apply(ctx)
	}),
)
