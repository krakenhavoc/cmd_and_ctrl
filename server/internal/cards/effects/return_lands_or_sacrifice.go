package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// returnLandsOrSacrifice is the shared body of "When this creature
// enters, sacrifice it unless you return N <lands> you control to their
// owner's hand" (Multiverse Gift Box, #5632): Bull Elephant's two Forests
// and Ovinomancer's three basic lands. Quickling's "sacrifice it unless you
// return another creature" (quickling.go) is the one-card form.
//
// The choice is made as the trigger resolves. With fewer than N matching
// lands under your control the creature is sacrificed at once: the
// payment is all N or nothing, so it can't be made in part. Otherwise a
// choose_cards prompt asks for exactly N of them, or none, which declines
// the payment and sacrifices the creature. The N lands are returned
// together, and the creature stays only if all N really reached a hand;
// one sent elsewhere by a replacement is a payment not made. A creature
// that has left the battlefield, or come back as a new object, is not
// sacrificed (CR 400.7), the same guard every SacrificePermanent has.
//
// Each card registers its own pick and tail keys (on-disk identities,
// never renamed or reused); n and match are fixed at init and the
// registered functions capture nothing else.
type returnLandsOrSacrifice struct {
	question string
	n        int
	// match is whether `land` is one the payment can return. `source` is
	// the creature as the table now has it, or nil once it is gone, so a
	// land type read through the creature's text (CR 612.2) can fall back
	// to the printed word.
	match func(source *game.Card, land game.Card) bool
	pick  game.ChooseCardsThen
	tail  game.TailRef
}

// newReturnLandsOrSacrifice registers the pick and the tail under the
// card's keys. Call it once, from a package-level var.
func newReturnLandsOrSacrifice(question string, n int, pickKey, tailKey string, match func(source *game.Card, land game.Card) bool) *returnLandsOrSacrifice {
	r := &returnLandsOrSacrifice{question: question, n: n, match: match}
	r.tail = CardTail(tailKey, r.afterReturn)
	r.pick = askedPickRule(pickKey, r.exactlyN, r.picked)
	return r
}

// exactlyN is the pick's set rule: all N lands (an empty answer, the
// payment declined, never reaches a set rule).
func (r *returnLandsOrSacrifice) exactlyN(picked []game.Card) bool { return len(picked) == r.n }

// picked returns the N lands named, or sacrifices the creature when none
// were named or one of them has gone.
func (r *returnLandsOrSacrifice) picked(ctx *Context, c game.CardsChosen) error {
	if r.allStillThere(ctx, c.Picked) {
		return ctx.Game.BounceCardsToHandKeyedForEffect(c.Picked, ctx.ItemTail(r.tail))
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
}

// allStillThere is whether exactly N lands were named and every one is
// still on the battlefield.
func (r *returnLandsOrSacrifice) allStillThere(ctx *Context, ids []uuid.UUID) bool {
	if len(ids) != r.n {
		return false
	}
	for _, id := range ids {
		if !onBattlefield(ctx.Game, id) {
			return false
		}
	}
	return true
}

// afterReturn keeps the creature only when all N lands reached a hand.
func (r *returnLandsOrSacrifice) afterReturn(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == r.n {
		return nil
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
}

// trigger is the enters trigger, labelled with the card's text.
func (r *returnLandsOrSacrifice) trigger(label string) game.TriggeredAbility {
	return WhenThisEnters(label, r.enters)
}

func (r *returnLandsOrSacrifice) enters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var source *game.Card
	if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
		source = &c
	}
	var lands []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsLand() && r.match(source, c) {
			lands = append(lands, c.InstanceID)
		}
	}
	if len(lands) < r.n {
		return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     item.SourceCardID,
		Question:   r.question,
		Cards:      lands,
		Min:        0,
		Max:        r.n,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    r.pick,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}
