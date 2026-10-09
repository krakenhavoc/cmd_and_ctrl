package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_helpers.go — small helpers shared by the Reality
// Fracture (FRA / FRC) cards. Append-only.

// rfCardsMilledThisTurn is the number of cards that were put into
// `owner`'s graveyard from their library this turn (Cruel
// Calculations). It reads this turn's events, which are bounded at the
// real turn boundary, and counts one per move: a card that left the
// graveyard and was milled again counts twice.
//
// A mill announces itself as an EventMill; other library-to-graveyard
// routes announce a plain EventZoneMove. A route that emitted both for
// one card would count it twice, so an event naming the same card as
// the one counted immediately before it is the same move.
//
// Caller holds g.mu.
func rfCardsMilledThisTurn(g *game.Game, owner uuid.UUID) int {
	n := 0
	var lastCounted uuid.UUID
	var lastIdx = -2
	for i, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventMill && ev.Kind != game.EventZoneMove {
			continue
		}
		if ev.OldZone != game.ZoneLibrary || ev.NewZone != game.ZoneGraveyard {
			continue
		}
		if ev.CardID == lastCounted && i-lastIdx <= 2 {
			lastIdx = i
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if ok && c.Owner == owner {
			n++
			lastCounted, lastIdx = ev.CardID, i
		}
	}
	return n
}

// AttackedThisTurn is the target predicate "creature that attacked this
// turn" (Hexhaven Dueling Arena): the per-object attack tally the turn
// keeps, so a creature that attacked and then left combat still counts
// and a creature that left and came back (a new object) does not.
func AttackedThisTurn() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return g.TimesAttackedThisTurn(c.InstanceID) > 0
	}
}

// BecomeUnprepared is "[it] becomes unprepared" (CR 722.3b): the
// permanent loses the prepared designation and the copy of its prepare
// spell kept in exile. Not an error when nothing happens — a permanent
// that isn't prepared, or has left the battlefield, simply stays as it is.
type BecomeUnprepared struct {
	Target uuid.UUID
}

func (b BecomeUnprepared) Apply(ctx *Context) error {
	if ctx.isNewSourceObject(b.Target) { // #1432
		return nil
	}
	return nothingIfGone(ctx.Game.UnprepareForEffect(b.Target))
}

// seedSutureResolve is Seed Suture, the prepare spell of both
// Blossom-Blessed Angel and Emergency Phytomedic: "Put a +1/+1 counter
// on target creature. You gain 1 life."
func seedSutureResolve(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
}

// soulTetherResolve is Soul Tether, the prepare spell of both Heartwood
// Crafter and Konstrari Improviser: "Create a Heartwood token."
func soulTetherResolve(_ *game.StackItem, ctx *Context) error {
	return CreateToken{Template: HeartwoodToken(), N: 1}.Apply(ctx)
}

// castAPreparedSpell reports whether the spell the cast event named is a
// "prepared spell": the copy of a prepare spell its controller cast out
// of exile (CR 722.3c). Those are the only copies the cast path ever
// puts on the stack, so a cast spell that is a copy is a prepared spell.
func castAPreparedSpell(g *game.Game, spellID uuid.UUID) bool {
	if c, ok := g.LookupCardForEffect(spellID); ok && c.PrepareCopy {
		return true
	}
	item := g.StackItemForEffect(spellID)
	return item != nil && item.IsCopy
}

// returnTargetGraveyardCardToHand is the body of "return target <card>
// from your graveyard to your hand": the first legal card target the
// trigger or spell announced goes to its owner's hand. Evolution
// Witness and Carnivorous Cultivator both end this way.
func returnTargetGraveyardCardToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
		}
	}
	return nil
}
