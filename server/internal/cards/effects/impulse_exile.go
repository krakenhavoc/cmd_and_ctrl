package effects

import (
	"fmt"
	"strconv"

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
//
// TotalManaValueAtMost is "you may cast any number of spells with total
// mana value N or less from among them without paying their mana costs"
// (Primeval Spawn, #2017). It is not a window for the turn: the casts are
// the resolving ability's (CR 608.2g), so the grant is a free, flash-speed,
// cast-only permission that closes on the holder's next pass and leaves an
// uncast card in exile. The cards the exile lands share ONE permission
// carrying a running total (game.CastPermission.ManaValueBudget): each cast
// spends its spell's mana value, X at 0 as a free cast locks it, and a card
// whose mana value is more than what is left is not offered, by the one
// gate the cast path, the client's offers and the bot enumerator share. A
// land among the cards is stranded, as for any cast-only grant. It is its
// own shape, so it refuses the options that mean a different window (Free,
// CastOnly, WhileExiled, Window, FaceDown, Condition, Filter) rather than
// quietly overriding them.
//
// EachManaValueAtMost is "you may cast any number of spells with mana value
// N or less from among them without paying their mana costs" (Hazoret's
// Undying Fury, #4021): the same resolving-ability window, but the cap is
// on EACH spell and there is no running total. The one permission carries
// MaxSpellManaValue and AnyNumber, so a cast strikes only its own card and
// the rest stay open until the holder passes. It refuses the same options
// as TotalManaValueAtMost and cannot be combined with it.
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
	Condition   *game.PermissionCondition
	Window      *game.Duration
	Filter      game.PermissionFilter

	// TotalManaValueAtMost is the shared budget described above. Nil is the
	// ordinary one-grant-per-turn impulse exile.
	TotalManaValueAtMost *int

	// EachManaValueAtMost is the per-spell cap described above. Nil is the
	// ordinary impulse exile.
	EachManaValueAtMost *int
}

func (e ExileTopWithPermission) Apply(ctx *Context) error {
	n := e.N
	if n <= 0 {
		n = 1
	}
	if e.TotalManaValueAtMost != nil || e.EachManaValueAtMost != nil {
		return e.applyBudgeted(ctx, n)
	}
	perm := e.permission()
	if e.FaceDown {
		return ctx.Game.ExileTopFaceDownWithPermissionForEffect(e.From, e.GrantTo, n, perm)
	}
	_, err := ctx.Game.ExileTopWithPermissionForEffect(e.From, e.GrantTo, n, perm)
	return err
}

// ApplyThen is Apply for a caller that acts on what was exiled, or that
// chains another exile behind this one (#4266, exileTopEachPlayerInTurn):
// `then` runs once the exile has settled, with the cards that reached exile,
// at once on a clean run and from the answer when a CR 616.1 prompt paused
// it. Face up and unbudgeted only, which is every caller that chains; the
// face-down and budgeted forms have no continuation yet and are refused
// rather than run without one.
func (e ExileTopWithPermission) ApplyThen(ctx *Context, then func(ctx *Context, landed []uuid.UUID) error) error {
	if e.FaceDown || e.TotalManaValueAtMost != nil || e.EachManaValueAtMost != nil {
		return fmt.Errorf("ExileTopWithPermission.ApplyThen: the face-down and budgeted forms have no continuation")
	}
	n := e.N
	if n <= 0 {
		n = 1
	}
	// The context is rebuilt inside the continuation from the live *Game,
	// as MillToZone.Apply does: an undo restores the game in place.
	item := ctx.Item
	return ctx.Game.ExileTopWithPermissionThenForEffect(e.From, e.GrantTo, n, e.permission(),
		func(g *game.Game, landed []uuid.UUID) error {
			if then == nil {
				return nil
			}
			return then(NewContext(g, item), landed)
		})
}

// permission is the grant a plain (unbudgeted) ExileTopWithPermission stamps.
func (e ExileTopWithPermission) permission() game.CastPermission {
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
	if e.Window != nil {
		perm.Duration = *e.Window
	}
	perm.Condition = e.Condition
	perm.Filter = e.Filter
	if e.Free {
		// "{0}", not empty: an empty Cost means "pay the printed
		// cost", which is the opposite of what this clause grants.
		perm.Cost = "{0}"
	}
	return perm
}

// applyBudgeted is the TotalManaValueAtMost and EachManaValueAtMost form:
// one free, flash-speed, cast-only permission over the cards that landed,
// spent down by each cast (a running total) or kept for every other card
// (a cap on each spell).
func (e ExileTopWithPermission) applyBudgeted(ctx *Context, n int) error {
	if e.Free || e.CastOnly || e.WhileExiled || e.Window != nil || e.FaceDown ||
		e.Condition != nil || e.Filter != (game.PermissionFilter{}) {
		return fmt.Errorf("ExileTopWithPermission: TotalManaValueAtMost and EachManaValueAtMost are their own window and take no other option")
	}
	if e.TotalManaValueAtMost != nil && e.EachManaValueAtMost != nil {
		return fmt.Errorf("ExileTopWithPermission: TotalManaValueAtMost and EachManaValueAtMost are two readings of one clause; name one")
	}
	perm := game.CastPermission{
		// "{0}", not empty: an empty Cost means "pay the printed cost".
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		AnyColor:    e.AnyColor || e.AnyType,
		AnyType:     e.AnyType,
		LapseOnPass: game.LapseStaysInExile,
		Source:      ctx.Source(),
	}
	if e.TotalManaValueAtMost != nil {
		budget := *e.TotalManaValueAtMost
		if budget < 0 {
			return fmt.Errorf("ExileTopWithPermission: TotalManaValueAtMost %d is negative", budget)
		}
		perm.ManaValueBudget = &budget
		perm.Label = "Cast any number of spells with total mana value " + strconv.Itoa(budget) + " or less without paying their mana costs"
	} else {
		limit := *e.EachManaValueAtMost
		if limit < 0 {
			return fmt.Errorf("ExileTopWithPermission: EachManaValueAtMost %d is negative", limit)
		}
		perm.MaxSpellManaValue = &limit
		perm.AnyNumber = true
		perm.Label = "Cast any number of spells with mana value " + strconv.Itoa(limit) + " or less without paying their mana costs"
	}
	if src, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok {
		perm.SourceName = src.Name
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

// EachPlayerExilesTopAndMayPlay is "each player exiles the top N cards
// of their library. <For a while>, each player may play the cards they
// exiled this way" (#2559): Memory Vessel's seven until your next turn,
// Rocco, Street Chef's one until your next end step. Every seated
// player gets one permission, over the cards THEY exiled and held by
// them — the owner, unlike Etali's single holder — for the window `d`
// the caller stamps against its own controller ("your next turn").
// Play, not cast: a land among them is played and spends a land drop
// (ADR 0066 §6).
func EachPlayerExilesTopAndMayPlay(g *game.Game, n int, d game.Duration) error {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if _, err := g.ExileTopWithPermissionForEffect(p.ID, p.ID, n, game.CastPermission{Duration: d}); err != nil {
			return err
		}
	}
	return nil
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
