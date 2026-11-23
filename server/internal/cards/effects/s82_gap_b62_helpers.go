package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b62_helpers.go — GAP batch 62 (#5948): the shared pieces of
// the batch's cards.

// anotherHumanDiedUnderYourControlThisTurn reports whether a creature
// other than `self` died this turn under `controller`'s control and was a
// Human as it last existed (CR 603.10a). `human` is the word "Human" as the
// source reads it (source.TextSubtype, CR 612.2). The turn's zone moves are the
// record, so a token that has ceased to exist still counts.
func AnotherHumanDiedUnderYourControlThisTurn(g *game.Game, controller, self uuid.UUID, human string) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == self {
			continue
		}
		// A token that has since ceased to exist is no longer found, but its
		// departure event carries what it last was (CR 603.10a).
		dead, _ := g.LookupCardForEffect(ev.CardID)
		if LeftAsType(ev, dead, "creature") && LeftUnderControlOf(ev, dead) == controller && LeftAsSubtype(ev, dead, human) {
			return true
		}
	}
	return false
}

// gapB62AttachedTo reports, for the permanent `id` as it last existed on
// the battlefield (the departure record, CR 603.10a: Equipment falls off
// and an Aura is put away before a "dies" trigger is looked at), whether
// an Aura (aura true) or an Equipment (aura false) was attached to it.
func gapB62AttachedTo(g *game.Game, id uuid.UUID, aura bool) bool {
	info, ok := g.LastKnownPermanentForEffect(id)
	if !ok {
		return false
	}
	want := "Equipment"
	if aura {
		want = "Aura"
	}
	for _, ref := range info.Attachments {
		if c, live := g.LookupCardForEffect(ref.ID); live && c.ObjectEpoch == ref.Epoch && OnBattlefield(g, ref.ID) {
			if c.HasSubtype(want) {
				return true
			}
			continue
		}
		if rec, found := g.LastKnownPermanentForEffect(ref.ID); found && rec.Epoch == ref.Epoch &&
			KeywordSliceContains(rec.Characteristic.Subtypes, want) {
			return true
		}
	}
	return false
}

// gapB62DiesAttached is a "when this creature dies, if it was enchanted /
// equipped, ..." row: the source's own death, with the intervening "if"
// (CR 603.4) asked as the creature dies and again as the ability resolves,
// both against the departure record (so it stays true once the Aura is in
// the graveyard and the Equipment is unattached).
func GapB62DiesAttached(label string, aura bool, effect Effect) game.TriggeredAbility {
	name := "it was equipped"
	if aura {
		name = "it was enchanted"
	}
	cond := CondBoard{
		Name: name,
		Test: func(ctx *Context) bool { return gapB62AttachedTo(ctx.Game, ctx.Item.SourceCardID, aura) },
	}
	return InterveningIf(cond, On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return CardDied(ev, source) && gapB62AttachedTo(g, ev.CardID, aura)
	}, label, effect))
}

// gapB62RollTotal is the result of the roll instruction whose dice carry
// `batch`: the total of its kept dice, each its natural result plus the
// instruction's modifier (CR 706.2). A planar die has no number.
func GapB62RollTotal(g *game.Game, batch uint64) int {
	total := 0
	for _, rolled := range g.Events {
		if rolled.Kind == game.EventRollDie && rolled.BatchSeq == batch && rolled.PlanarFace == "" {
			total += rolled.RollResult()
		}
	}
	return total
}
