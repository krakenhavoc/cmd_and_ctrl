package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// retarget.go — the card-facing half of CR 115.7 (#1196): "change the
// target of target spell or ability with a single target" and "you
// may choose new targets for target spell or ability".
//
// Two pieces and no third: a PREDICATE for the printed clause ("with
// a single target") and a PRIMITIVE that opens the engine's offer.
// Everything else — which new targets are legal, protection and
// hexproof, how many slots may move, what happens when there is
// nowhere else to point — lives in game/retarget.go, because it is
// the announce gate run again rather than anything a card knows.
//
// See docs/decisions/0019-structured-targeting.md, the 2026-09-22
// amendment.

// HasASingleTarget matches a spell on the stack with exactly one
// target — the "with a single target" half of Bolt Bend's,
// Misdirection's, Ricochet Trap's and Imp's Mischief's printed
// clause.
//
// It counts SLOTS, not distinct objects: a spell that chose the same
// creature for two different instances of the word "target" (CR
// 115.3's AllowSame) has two targets and is not a single-target
// spell, which is the ruling.
func HasASingleTarget() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return g.StackItemTargetCountForEffect(c.InstanceID) == 1
	}
}

// ChangeTargets is CR 115.7 from a card file: offer `Chooser` the
// chance to redirect the spell or ability `StackID`.
//
// "Nothing to redirect" is a normal outcome, not an error, and this
// is where that is decided: a spell that is no longer on the stack
// (countered in response), or one with no targets at all (a Swat
// pointed at a Wrath of God, which is a legal target and a pointless
// one), leaves the card doing nothing. The engine's entry point is
// strict about both so a card that meant something else fails loudly;
// this primitive is where the printed "and then nothing happens" is
// absorbed.
type ChangeTargets struct {
	// StackID is the item being redirected — item.Targets[0].ID for
	// every card that prints this.
	StackID uuid.UUID

	// Chooser defaults to the resolving spell's controller, which is
	// every printed case.
	Chooser uuid.UUID

	// Policy is RetargetChangeOne for "change the target of…" and
	// RetargetChooseNew for "choose new targets for…".
	Policy game.RetargetPolicy

	// Optional marks a printed "you may".
	Optional bool

	// Reason is the prompt header. Empty falls back to the clause
	// label the target was announced under.
	Reason string
}

func (c ChangeTargets) Apply(ctx *Context) error {
	if c.StackID == uuid.Nil {
		return nil
	}
	chooser := c.Chooser
	if chooser == uuid.Nil {
		chooser = ctx.Controller()
	}
	if !ctx.Game.RetargetableForEffect(c.StackID) {
		// Countered in response, already resolved, or a spell with
		// no targets. All three are "the card did nothing".
		return nil
	}
	err := ctx.Game.OfferRetargetForEffect(game.RetargetOffer{
		ItemID:   c.StackID,
		Chooser:  chooser,
		Policy:   c.Policy,
		Optional: c.Optional,
		Source:   ctx.Source(),
		Reason:   c.Reason,
	})
	if err == game.ErrCardNotFound {
		return nil
	}
	return err
}
