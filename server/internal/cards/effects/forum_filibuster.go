package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forum Filibuster — Enchantment {3}{W}{W} (Secrets of Strixhaven
// Commander, #4834):
//
//	"At the beginning of your upkeep, create a 2/1 white and black
//	 Inkling creature token with flying. When you do, return up to one
//	 target Aura or Equipment card from your graveyard to the battlefield
//	 attached to that token."
//
// Dáin Ironfoot's shape: the Inkling is created through the keyed token
// door, and "when you do" is a CR 603.12 reflexive trigger raised only
// for an Inkling that really arrived, with the token as its payload. Its
// target is chosen as it goes on the stack, not when the upkeep trigger
// does (the ruling), so the table can respond to it.
//
// The return names the host (ReturnFromGraveyardTogether.AttachedTo): an
// Aura is not asked what to enchant, and one that can't legally enchant
// the token, or whose token has left, stays in the graveyard (CR 303.4i).
// An Equipment is attached as it enters and left to CR 301.5c.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a9cc11b-a223-46f0-bf58-c1e08ad1253d",
		Name:         "Forum Filibuster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Forum Filibuster — create a 2/1 Inkling with flying, then return an Aura or Equipment attached to it", forumFilibusterCreateInkling),
		},
	})
}

// forumFilibusterCreateInkling is the upkeep trigger's first sentence.
func forumFilibusterCreateInkling(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(TokenCard("2/1 white and black Inkling with flying")),
			Count:    1,
		}},
	}, ctx.ItemTail(forumFilibusterInklingCreated))
}

// forumFilibusterInklingCreated is "when you do". An on-disk key.
var forumFilibusterInklingCreated = CardTail("forum-filibuster/inkling-created", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ReflexiveTrigger{
		Label: "Forum Filibuster — return up to one Aura or Equipment card attached to the Inkling",
		Cards: o.IDs[:1],
		Body:  forumFilibusterReturnBody,
	}.Apply(ctx)
})

// forumFilibusterReturnBody is the reflexive trigger's registration: the
// Inkling is the payload, the card the trigger's own target. An on-disk
// key.
var forumFilibusterReturnBody = game.ReflexiveBody("forum-filibuster/return-attached", simpleBody(forumFilibusterReturn),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("up to one target Aura or Equipment card from your graveyard",
			YouOwn(), Or(Subtype("Aura"), Subtype("Equipment"))).WithCount(0, 1)
	}))

// forumFilibusterReturn returns the target attached to the Inkling.
func forumFilibusterReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	inklings := ctx.PayloadCards()
	targets := ctx.LegalTargets()
	if len(inklings) == 0 || len(targets) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets:    []uuid.UUID{targets[0].ID},
		Controller: item.Controller,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: inklings[0]},
	}.Apply(ctx)
}
