package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lim-Dûl's Vault — Instant {U}{B} (Alliances, S82 GAP2 batch 13, #5948,
// #6101):
//
//	"Look at the top five cards of your library. As many times as you
//	 choose, you may pay 1 life, put those cards on the bottom of your
//	 library in any order, then look at the top five cards of your
//	 library. Then shuffle and put the last cards you looked at this way
//	 on top in any order."
//
// A loop of prompts the engine already has, each a keyed continuation so a
// table waiting at any point is a restore point:
//
//  1. Look at the top five (the look-at-top prompt: only you see them, and
//     you put them back in an order of your choosing).
//  2. "Pay 1 life, put those cards on the bottom and look at the next
//     five?" A yes pays the life (CR 119.4: only with at least 1 life, and
//     never for a player who can't lose life) and puts those cards on the
//     bottom in an order you choose (the put-in-library prompt), then goes
//     back to 1. A player who can't pay is not asked.
//  3. A no ends the loop: the library is shuffled except the cards you last
//     looked at, which stay on top in the order you gave them in step 1
//     (ShuffleLibraryKeepingOnTopForEffect, #6101).
//
// Step 1's order is the final "on top in any order": the order you choose
// with those five in front of you is the order they keep, and putting them
// on the bottom instead discards it. A library shorter than five shows what
// it has, and one of five or fewer cards shows the same cards again after a
// payment, exactly as the printed loop does.
//
// No simplification.
func init() {
	// Registered here rather than as package-level initialisers: the loop
	// refers back to itself (look, ask, bottom, look), which Go's
	// initialisation order refuses.
	limDulsVaultLooked = game.RegisterTail("then/lim-duls-vault-looked", limDulsVaultAsk)
	limDulsVaultAgain = game.RegisterConfirmThen("confirm/lim-duls-vault-again", limDulsVaultAnswer)
	limDulsVaultBottomed = game.RegisterTail("then/lim-duls-vault-bottomed", limDulsVaultLookAgain)
	effects.Register(effects.Spec{
		OracleID:     "3f8e7a45-4c6e-4ee6-93d0-b7de9715ec97",
		Name:         "Lim-Dûl's Vault",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			limDulsVaultLook(ctx.Game, item.Controller, item.SourceCardID)
			return nil
		},
	})
}

// The loop's on-disk keys, registered in init. Never renamed, never reused.
var (
	limDulsVaultLooked   game.TailRef
	limDulsVaultAgain    game.ConfirmThen
	limDulsVaultBottomed game.TailRef
)

// limDulsVaultLook is step 1: look at the top five and put them back in an
// order of your choosing, then ask whether to go again.
func limDulsVaultLook(g *game.Game, you, source uuid.UUID) {
	g.LookAtTopKeyedForEffect(you, source, 5, game.KeyedTail{Ref: limDulsVaultLooked, Player: you, Source: source})
}

// limDulsVaultLooked is step 2's question, asked of a player who can pay 1
// life; one who can't goes straight to step 3.
func limDulsVaultAsk(g *game.Game, o game.TailOutcome) error {
	you, source := o.Tail.Player, o.Tail.Source
	top := g.LookAtTopOfLibraryForEffect(you, 5)
	p := g.PlayerByIDForEffect(you)
	if len(top) == 0 || p == nil || p.Eliminated || p.Life < 1 {
		return g.ShuffleLibraryKeepingOnTopForEffect(you, top)
	}
	_, err := g.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
		Chooser:         you,
		Source:          source,
		Question:        "Lim-Dûl's Vault — pay 1 life, put those cards on the bottom of your library and look at the top five cards?",
		AcceptLabel:     "Pay 1 life and look again",
		DeclineLabel:    "Keep these on top and shuffle",
		LifeCost:        1,
		DeclineEndsLoop: true,
	}, limDulsVaultAgain, top)
	return err
}

// limDulsVaultAgain is the answer: a yes pays and puts the cards on the
// bottom in any order, then looks again; a no, or a payment the player
// cannot make, shuffles and keeps them on top.
func limDulsVaultAnswer(g *game.Game, r game.ConfirmAnswer) error {
	you, top := r.Chooser, r.Carry
	if !r.Accepted {
		return g.ShuffleLibraryKeepingOnTopForEffect(you, top)
	}
	if p := g.PlayerByIDForEffect(you); p == nil || p.Life < 1 || g.PayLifeForEffect(r.Source, you, 1) != nil {
		return g.ShuffleLibraryKeepingOnTopForEffect(you, top)
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   you,
		Source:    r.Source,
		Cards:     top,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    "Lim-Dûl's Vault — put those cards on the bottom of your library in any order",
		ThenTail:  game.KeyedTail{Ref: limDulsVaultBottomed, Player: you, Source: r.Source},
	})
}

// limDulsVaultBottomed looks at the next five once the cards are on the
// bottom.
func limDulsVaultLookAgain(g *game.Game, o game.TailOutcome) error {
	limDulsVaultLook(g, o.Tail.Player, o.Tail.Source)
	return nil
}
