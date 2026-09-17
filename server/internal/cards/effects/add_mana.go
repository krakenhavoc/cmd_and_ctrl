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
// pick a Birds of Paradise activation would, narrowed to the
// controller's commander identity like Treasure is unless
// IgnoreCommanderIdentity is set.
//
// The mana is attributed to the resolving item's source card, and it
// empties with the pool at the end of the step (CR 106.4): mana from
// a spell is spent in the step the spell resolved in, which is why
// Dark Ritual is a main-phase card. Player left zero means the
// item's controller.
type AddMana struct {
	Player   uuid.UUID
	Produced string

	// IgnoreCommanderIdentity keeps a pipe pick at its printed width,
	// exactly as ManaAbility.IgnoreCommanderIdentity does for a mana
	// ability. Set it whenever the printed text says "any color" with
	// no commander-identity clause (Lotus Cobra, Deathrite Shaman).
	IgnoreCommanderIdentity bool
}

func (a AddMana) Apply(ctx *Context) error {
	if a.Produced == "" {
		return nil
	}
	player := a.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.AddManaWithOptionsForEffect(player, ctx.Source(), a.Produced,
		game.AddManaOptions{IgnoreCommanderIdentity: a.IgnoreCommanderIdentity})
}
