package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// add_mana.go — the AddMana primitive: mana produced by a SPELL or a
// non-mana ability rather than by a CR 605 mana ability. Its own
// file, like flicker.go and token_copy.go, rather than a new entry
// in primitives.go — that file is the one every batch reaches for.
//
// Dark Ritual is the whole reason this exists: "Add {B}{B}{B}" is a
// spell that resolves off the stack, so it cannot be a ManaAbility
// (which never uses the stack), and until this primitive every mana
// in the engine came from ActivateManaAbility. Mana Drain's delayed
// "add an amount of {C}" is the same shape one turn later.

// AddMana puts the mana described by Produced into Player's pool.
// Produced uses the brace grammar ManaAbility.Produced does, pipe
// syntax included — a "{W|U|B|R|G}" slot queues the same colour
// pick a Birds of Paradise activation would: every printed colour,
// with the controller's commander identity listed first, unless
// NarrowToCommanderIdentity is set.
//
// The mana is attributed to the resolving item's source card, and it
// empties with the pool at the end of the step (CR 106.4): mana from
// a spell is spent in the step the spell resolved in, which is why
// Dark Ritual is a main-phase card. Player left zero means the
// item's controller.
type AddMana struct {
	Player   uuid.UUID
	Produced string

	// Amount is the count form (#3112) for "Add {R} for each ...",
	// "add N mana of any one color": a game.ManaAmount instead of
	// Produced, used when Produced is empty. The engine builds the slots
	// from the count with the MaxProducedMana ceiling applied to N before
	// anything is allocated, and announces a cut — nothing builds a
	// string whose length is a game quantity. Same pick, restriction,
	// rider and replacement behaviour as Produced. Use ManaN to build one.
	Amount game.ManaAmount

	// NarrowToCommanderIdentity intersects a pipe pick with the
	// controller's commander colour identity, exactly as
	// ManaAbility.NarrowToCommanderIdentity does for a mana ability.
	// Set it only when the printed text says "any color in your
	// commander's color identity"; "any color" (Lotus Cobra, Deathrite
	// Shaman) leaves it off and offers all five, identity first.
	NarrowToCommanderIdentity bool

	// Restrictions is "spend this mana only to …" (game.ManaRestrict*),
	// for fixed-colour output. Smoky Lounge.
	Restrictions []string

	// Riders are the spend riders the mana carries (ADR 0109 §11):
	// Domri, Chaos Bringer's "If that mana is spent on a creature spell,
	// it gains riot". Build them with the mana_spend_rider.go
	// constructors.
	Riders []game.ManaSpendRider
}

// KeepManaUntilEndOfTurn is the mark for "until end of turn, you don't
// lose this mana as steps and phases end" (Karn, Legacy Reforged; Savage
// Ventmaw; #2166). Put it in AddMana.Riders (or a mana ability's
// SpendRiders); the step-boundary sweep reads it (game/mana_keep.go) and
// it expires when the cleanup step begins.
func KeepManaUntilEndOfTurn() game.ManaSpendRider {
	return game.ManaSpendRider{Kind: game.ManaRiderKeepUntilEndOfTurn}
}

func (a AddMana) Apply(ctx *Context) error {
	if a.Produced == "" && len(a.Amount.Options) == 0 {
		return nil
	}
	player := a.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.AddManaWithOptionsForEffect(player, ctx.Source(), a.Produced,
		game.AddManaOptions{Amount: a.Amount, NarrowToCommanderIdentity: a.NarrowToCommanderIdentity, Restrictions: a.Restrictions, Riders: a.Riders})
}

// ManaN is the count form of "Add N mana", each from the colour options
// given: one option ("{R}") is N of that colour, several ("B", "R") are N
// independent picks (Burnt Offering's "{B} or {R}" each). A count of zero
// or less adds nothing. See AddMana.Amount.
func ManaN(n int, options ...string) game.ManaAmount {
	return game.ManaAmount{Options: options, N: n}
}
