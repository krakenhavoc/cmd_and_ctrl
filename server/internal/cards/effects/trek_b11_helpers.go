package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b11_helpers.go — shared pieces for the eleventh Star Trek batch
// (#4973).

// trekB11OtherThanSourceObject is "other" in a permanent's own ability
// read as object identity (CR 400.7): every creature but the source
// OBJECT the ability came from. A source that left and came back is a
// new object, and that one is affected like any other creature.
func trekB11OtherThanSourceObject(item *game.StackItem) CardPredicate {
	src := item.SourceObject
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.InstanceID != src.ID || c.ObjectEpoch != src.Epoch
	}
}

// trekB11PutCounterOnEachOfTypeYouControl is "Put a +1/+1 counter on each
// <subtype> you control": post-layer subtypes, so a changeling counts, and
// the word read through the source's text (CR 612.2).
func trekB11PutCounterOnEachOfTypeYouControl(subtype string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		want := g.TextReaderForEffect(item).TextSubtype(subtype)
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == item.Controller && c.IsCreature() && c.HasSubtype(want) {
				ids = append(ids, c.InstanceID)
			}
		}
		return b13PutCounterOnEach(NewContext(g, item), ids)
	}
}

// trekB11Assimilate is the Star Trek keyword action assimilate on one
// card in a graveyard: "Put it onto the battlefield under your control
// with a +1/+1 counter. It's a Borg artifact creature and loses all other
// creature types." CR 205.1b: a "[creature type] artifact creature"
// keeps its other card types and its subtypes other than creature types,
// so the put adds Artifact and replaces the creature types with Borg, in
// one record registered as the permanent lands (ADR 0210). The counter is
// an entry counter, so the creature enters with it.
func trekB11Assimilate(ctx *Context, card uuid.UUID, label string) error {
	return putCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, []uuid.UUID{card},
		game.ZoneEntryOptions{
			Controller:         ctx.Controller(),
			EntersWithCounters: map[string]int{game.CounterPlusOne: 1},
			EntersAs: &game.EntersAs{
				AddTypes:      []string{"Artifact"},
				CreatureTypes: ctx.TextSubtypes("Borg"),
				Label:         label,
			},
		}, game.TailRef{}, 0)
}
