package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// impulse_exile.go — S21 sub-PR 6: the ExileTopWithPermission
// primitive and its shared helpers.
//
// "Impulse exile" is the community name for "exile the top card of
// a library; you may play it this turn". It is card advantage with
// a shot clock, and it is the reason a red deck can play a long
// game — which is why it's the most-used unimplemented mechanic in
// the Pirates decklist.

// ExileTopWithPermission exiles the top `N` cards of `From`'s
// library and lets `GrantTo` play them until end of turn.
//
// CastOnly mirrors "you may CAST that card" (Ragavan) as against
// "you may PLAY those cards" (Breeches): the former strands a land,
// because playing a land is not casting.
//
// AnyColor mirrors "you may spend mana as though it were mana of
// any color to cast those spells"; AnyType the wider "mana of any
// TYPE", which pays {C} too (#1573).
//
// FaceDown is "exiles it face down. They may look at and play it"
// (Gonti, Night Minister; Outrageous Robbery, #1573): the cards land
// as game.FaceDownPermitted, and only GrantTo may look at them.
// WhileExiled makes the grant last for as long as each card stays in
// exile rather than until end of turn.
// Free is "without paying its mana cost" (#2527, Gix, Yawgmoth
// Praetor): the play costs {0} in place of the printed mana cost, the
// grant Urza, Lord High Artificer's free play and cascade make.
// Additional costs are still owed (CR 118.9a), and a land is simply
// played.
type ExileTopWithPermission struct {
	// From is whose library is exiled off. Often an opponent's.
	From uuid.UUID
	// GrantTo is who may play the cards. Usually the controller of
	// the effect, and usually NOT the owner.
	GrantTo     uuid.UUID
	N           int
	CastOnly    bool
	AnyColor    bool
	AnyType     bool
	FaceDown    bool
	WhileExiled bool
	Free        bool
}

func (e ExileTopWithPermission) Apply(ctx *Context) error {
	n := e.N
	if n <= 0 {
		n = 1
	}
	perm := game.CastPermission{
		CastOnly: e.CastOnly,
		// AnyType implies AnyColor, and saying both keeps the grant
		// any-colour for a binary that predates AnyType (#1573).
		AnyColor: e.AnyColor || e.AnyType,
		AnyType:  e.AnyType,
	}
	if e.WhileExiled {
		perm.Duration = game.WhileInZoneDuration()
	}
	if e.Free {
		// "{0}", not empty: an empty Cost means "pay the printed
		// cost", which is the opposite of what this clause grants.
		perm.Cost = "{0}"
	}
	if e.FaceDown {
		return ctx.Game.ExileTopFaceDownWithPermissionForEffect(e.From, e.GrantTo, n, perm)
	}
	_, err := ctx.Game.ExileTopWithPermissionForEffect(e.From, e.GrantTo, n, perm)
	return err
}

// ExileTopNUntilYourNextTurn exiles the top n cards of the resolving
// effect's controller and lets them play it until the end of their
// next turn — b19ExileTopTwoUntilEndOfNextTurn's shape (Reckless
// Impulse, Wrenn's Resolve) generalized to a card whose printed count
// isn't two (The Legend of Roku's chapter I, top three). ADR 0063's
// seat-turn counter is what makes "your next turn" mean the same
// thing from every seat, so no round-number backstop is needed.
func ExileTopNUntilYourNextTurn(ctx *Context, n int) error {
	controller := ctx.Controller()
	_, err := ctx.Game.ExileTopWithPermissionForEffect(controller, controller, n, game.CastPermission{
		Duration: ctx.Game.UntilEndOfYourNextTurnDuration(controller),
	})
	return err
}

// ExileTopNUntilYourNextEndStep is "exile the top N cards of your
// library. Until your next end step, you may play them" (#2373) — Ob
// Nixilis, Captive Kingpin, Haste Magic, Wiccan, Opera Love Song. The
// window closes as that end step BEGINS (game.UntilYourNextEndStep),
// so it is shorter than "until end of turn" when cast on your own
// turn and ends at the end step of your next turn when made on an
// opponent's turn or in your own end step.
func ExileTopNUntilYourNextEndStep(ctx *Context, n int) error {
	controller := ctx.Controller()
	_, err := ctx.Game.ExileTopWithPermissionForEffect(controller, controller, n, game.CastPermission{
		Duration: ctx.Game.UntilYourNextEndStepDuration(controller),
	})
	return err
}

// damagedOpponent returns the opponent a combat-damage event hit,
// or uuid.Nil — the trigger condition shared by Ragavan ("deals
// combat damage to a player") and the Pirate batch triggers
// ("whenever one or more Pirates you control deal damage to your
// opponents").
//
// Batching caveat (the same one the rest of the catalog
// carries): the engine emits one damage event per source, so a
// two-Pirate strike fires the trigger twice with one player each
// rather than once with two. Every card here does per-player work,
// so the totals match; a card that read the batch size nonlinearly
// would not.
func damagedOpponent(ev game.Event, controller uuid.UUID, g *game.Game) uuid.UUID {
	if !combatDamageToPlayerBy(ev, controller, g) {
		return uuid.Nil
	}
	if ev.Target == controller {
		return uuid.Nil
	}
	return ev.Target
}

// damagedOpponentByAnyDamage is damagedOpponent without CR 603's
// combat-damage restriction — the reading Breeches, Brazen Plunderer
// and Malcolm, Keen-Eyed Navigator need, since both print "deal
// damage" with no "combat" qualifier. Ragavan's own trigger prints
// "deals combat damage to a player" and keeps using damagedOpponent
// above; this is an addition beside it, not a replacement.
func damagedOpponentByAnyDamage(ev game.Event, controller uuid.UUID, g *game.Game) uuid.UUID {
	if !damageToPlayerBy(ev, controller, g) {
		return uuid.Nil
	}
	if ev.Target == controller {
		return uuid.Nil
	}
	return ev.Target
}

// exileTopCardYouMayPlayThisTurn is the impulse bullet Breeches, Eager
// Pillager and Parapet Thrasher share: "Exile the top card of your
// library. You may play it this turn." Play, not cast, so a land may be
// played from exile; the grant's zero duration is "this turn".
func exileTopCardYouMayPlayThisTurn(item *game.StackItem, ctx *Context, _ int) error {
	return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1}.Apply(ctx)
}
