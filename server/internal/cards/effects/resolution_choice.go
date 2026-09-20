package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_choice.go — the two card-facing shapes of "an OPPONENT
// chooses while this resolves" (#568, CR 608.2).
//
// MayChoice (may_choice.go) already addresses a yes/no to any seat,
// which covers Combustible Gearhulk, Charismatic Conqueror and Painful
// Quandary. What it cannot say is a choice among three consequences,
// or a choice made in two steps by two different players. Those are
// the two shapes here, and both are built on game.PendingChoiceOptionPick
// plus #552's chaining rather than on kinds of their own:
//
//   - PickOption — "choose one of the following", addressed to an
//     opponent. Torment of Hailfire's "loses 3 life unless that player
//     sacrifices a nonland permanent of their choice or discards a
//     card".
//   - PileSplit — "an opponent separates those cards into two piles.
//     Put one pile into your hand and the other into your graveyard."
//     Fact or Fiction.
//
// # The list is built from what the chooser can DO
//
// CR 608.2's "as much as possible" is enforced at queue time, not at
// answer time: an option the chooser cannot take is never offered. A
// player with no nonland permanent is not shown "sacrifice a nonland
// permanent"; a player with an empty hand is not shown "discard a
// card". That is the rule the engine's own contract depends on — the
// enumerator marks the FIRST option always-legal, so the first option
// must be one that always works.

// PickOption asks Player to choose one of Options while an effect
// resolves, and runs Then with the index they picked.
//
// Anything printed after the choice goes in Then, not after this
// primitive returns: Apply only queues the prompt.
//
// A prompt with no options queues nothing and calls Then with -1, so a
// card whose whole option list turned out to be impossible still gets
// to finish. A chooser who has left the game (CR 800.4a) is the same
// case.
type PickOption struct {
	// Player is asked. Zero means the resolving effect's controller.
	Player uuid.UUID

	// From owns the material the options are about, when that is not
	// the chooser — the redaction pass reads it to decide what the
	// chooser may see of a pile that is not theirs. Zero means the
	// chooser's own.
	From uuid.UUID

	// Question is the prompt's header, written the way the card is.
	Question string

	// Options are the branches in the card's printed order. The FIRST
	// must be one the chooser can always take.
	Options []game.ChoiceOption

	// Then receives the index of the chosen option, or -1 when no
	// question could be asked. Runs with g.mu held; may queue
	// further choices.
	Then func(ctx *Context, index int) error
}

func (p PickOption) Apply(ctx *Context) error {
	chooser := p.Player
	if chooser == uuid.Nil {
		chooser = ctx.Controller()
	}
	item := ctx.Item
	then := p.Then
	queued := ctx.Game.QueueOptionPickForEffect(game.OptionPickPrompt{
		Chooser:    chooser,
		FromPlayer: p.From,
		Source:     ctx.Source(),
		Question:   p.Question,
		Options:    p.Options,
		Then: func(g *game.Game, index int) error {
			if then == nil {
				return nil
			}
			return then(NewContext(g, item), index)
		},
	})
	if queued != uuid.Nil || then == nil {
		return nil
	}
	return then(ctx, -1)
}

// PileSplit is "an opponent separates those cards into two piles; you
// take one of them" — two chained prompts to two different seats.
//
// The cards must already be REVEALED (RevealTopOfLibrary, or
// RevealCards) before this is called. The splitter is being asked
// about cards they do not own, and protocol's redaction pass shows
// them only what they are entitled to see: an unrevealed pool would
// reach them as an empty prompt. That is deliberate — the alternative
// is an engine that hands one seat a handle on another's library.
type PileSplit struct {
	// Splitter separates the cards. Zero is not meaningful; every
	// printed card of this family names an opponent.
	Splitter uuid.UUID

	// Chooser takes one of the two piles. Zero means the resolving
	// effect's controller, which is every printed card of this family.
	Chooser uuid.UUID

	// Owner owns the cards. Zero means the Chooser — Fact or Fiction
	// splits the controller's own library.
	Owner uuid.UUID

	// SplitQuestion / PickQuestion are the two prompts' headers.
	SplitQuestion, PickQuestion string

	// Cards are the revealed cards being separated, in the order the
	// table saw them.
	Cards []uuid.UUID

	// Then receives the pile the chooser TOOK and the pile they left,
	// in that order. Runs with g.mu held.
	Then func(ctx *Context, taken, left []uuid.UUID) error
}

func (p PileSplit) Apply(ctx *Context) error {
	chooser := p.Chooser
	if chooser == uuid.Nil {
		chooser = ctx.Controller()
	}
	item := ctx.Item
	then := p.Then
	ctx.Game.QueuePileSplitForEffect(game.PileSplitPrompt{
		Splitter:      p.Splitter,
		Chooser:       chooser,
		Owner:         p.Owner,
		Source:        ctx.Source(),
		SplitQuestion: p.SplitQuestion,
		PickQuestion:  p.PickQuestion,
		Cards:         p.Cards,
		Then: func(g *game.Game, taken, left []uuid.UUID) error {
			if then == nil {
				return nil
			}
			return then(NewContext(g, item), taken, left)
		},
	})
	return nil
}

// NonlandPermanentsControlledBy lists the nonland permanents a player
// controls, in battlefield order — the candidate set behind "sacrifices
// a nonland permanent of their choice".
//
// Here rather than in a card file because it is the shape of the
// clause, not of the card: an option list has to be built from what the
// chooser can actually do (see the file comment), and every card that
// offers a sacrifice branch needs the same answer.
//
// Caller must hold g.mu — it is an effect-time read.
func NonlandPermanentsControlledBy(g *game.Game, playerID uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller != playerID || c.IsLand() {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// SacrificeChoice queues "sacrifice one of these permanents" as a
// chained card pick, and runs Then once the permanent has actually
// gone.
//
// game.PlayerSacrificesForEffect is the prompt for a sacrifice that is
// the END of an effect (Grave Pact fans one out per player and nothing
// follows). This one is for a sacrifice in the MIDDLE of one — Torment
// of Hailfire repeats X times, and the next repetition must not be
// asked until this one has finished — so it needs a continuation, and
// it gets one the way QueueDiscardChoiceForEffect does: by riding
// PendingChoiceChooseCards, whose frame already carries a `Then` and
// re-checks the live zone on submit.
//
// It is not a second sacrifice system. The permanent leaves through
// SacrificePermanentForEffect, the one sacrifice path, so EventSacrifice
// and the dies-triggers are identical either way; what differs is only
// which prompt kind asked, and the CHOOSE-CARDS kind is the one with a
// continuation.
//
// A player with no candidate permanent is not prompted and Then runs
// immediately — a mandatory sacrifice with nothing to sacrifice does
// nothing (CR 701.21a), and the rest of the effect still happens.
type SacrificeChoice struct {
	// Player sacrifices. Required.
	Player uuid.UUID

	// Candidates are the permanents they may choose from, computed by
	// the caller (NonlandPermanentsControlledBy and friends) so the
	// clause's predicate stays in the card file.
	Candidates []uuid.UUID

	// Question is the prompt's header.
	Question string

	// Then runs once the permanent has gone, or immediately when
	// there was nothing to sacrifice.
	Then func(ctx *Context) error
}

func (s SacrificeChoice) Apply(ctx *Context) error {
	then := s.Then
	if len(s.Candidates) == 0 {
		if then == nil {
			return nil
		}
		return then(ctx)
	}
	item := ctx.Item
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    s.Player,
		FromPlayer: s.Player,
		Source:     ctx.Source(),
		Question:   s.Question,
		Cards:      s.Candidates,
		Min:        1,
		Max:        1,
		// Re-checked against the live battlefield on submit: the
		// prompt is asynchronous and a permanent can leave between
		// the question and the answer.
		Zone: game.ZoneBattlefield,
		// #993: the picked permanent goes through the sacrifice's
		// CONTINUATION, so `Then` really does run "once the permanent
		// has gone" the way this type promises. The fire-and-forget
		// call returns nil while a sacrificed commander's owner is
		// still answering CR 903.9, and the clause behind it then ran
		// with the permanent on the battlefield and the question open
		// — Chain of Vapor offered its copy while the land's owner was
		// mid-prompt. The answer is not read: the prompt was mandatory
		// (Min 1), so a clause hanging off it is "then", not "if you
		// do".
		Then: func(g *game.Game, picked []uuid.UUID) error {
			return g.SacrificeAllThenForEffect(uuid.Nil, picked, func(g *game.Game, _ []uuid.UUID) error {
				return resumeClause(g, item, then)
			})
		},
	})
	return nil
}

// PermanentsControlledBy lists every permanent a player controls, in
// battlefield order — the candidate set behind a bare "sacrifice a
// permanent" (Read the Runes), which unlike the nonland clause above
// includes lands.
//
// Caller must hold g.mu — it is an effect-time read.
func PermanentsControlledBy(g *game.Game, playerID uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == playerID {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
