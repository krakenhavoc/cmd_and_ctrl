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
// a single target") and a PRIMITIVE that opens the engine's offer —
// including, since #1743, the offer whose destination is fixed ("…
// to this creature", ChangeTargets.ToSource).
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

	// To pins the destination (#1743): the target changes to this
	// object or not at all, and the chooser picks at most WHICH
	// target changes. See game.RetargetOffer.To. Leave it zero for a
	// free choice.
	To game.TargetRef

	// ToSource is To for the printed "to this creature" (Spellskite,
	// Mizzium Meddler, Hydroelectric Specimen): the destination is
	// the ability's own source permanent, read at resolution. A
	// source that has left the battlefield — or come back as a new
	// object (CR 400.7) — is not "this creature" any more, and
	// nothing changes: "If Spellskite leaves the battlefield before
	// its ability resolves …, no targets are changed" (2020-08-07
	// ruling). Overrides To.
	ToSource bool
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
	to := c.To
	if c.ToSource {
		info, ok := ctx.SourcePermanent()
		if !ok || info.Left {
			return nil
		}
		to = game.TargetRef{Kind: game.TargetCard, ID: ctx.Source()}
	}
	err := ctx.Game.OfferRetargetForEffect(game.RetargetOffer{
		ItemID:   c.StackID,
		Chooser:  chooser,
		Policy:   c.Policy,
		Optional: c.Optional,
		Source:   ctx.Source(),
		Reason:   c.Reason,
		To:       to,
	})
	if err == game.ErrCardNotFound {
		return nil
	}
	return err
}

// changeATargetToThisCreature is the resolution body every "change a
// target of target spell or ability to this creature" card shares
// (#1743): the stack item it targeted, if that is still a legal
// target (CR 608.2b — a spell that resolved or was countered in
// response is gone, and nothing happens), offered to ChangeTargets
// with the destination pinned to the ability's own source.
//
// `optional` is the printed "you may" (Mizzium Meddler, Hydroelectric
// Specimen); Spellskite's activation is mandatory. `reason` is the
// prompt header, shown only when there is a choice to make: which
// target, or — for a "you may" — whether.
func changeATargetToThisCreature(reason string, optional bool) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return ChangeTargets{
				StackID:  t.ID,
				Policy:   game.RetargetChangeOne,
				Optional: optional,
				ToSource: true,
				Reason:   reason,
			}.Apply(ctx)
		}
		return nil
	}
}
