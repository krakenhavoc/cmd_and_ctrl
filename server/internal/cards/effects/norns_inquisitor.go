package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Norn's Inquisitor — Creature — Phyrexian Knight {1}{W}, 1/1 (March of
// the Machine):
//
//	"When this creature enters, incubate 2.
//	 Whenever a permanent you control transforms into a Phyrexian, put a
//	 +1/+1 counter on it."
//
// The second ability is a CR 701.27e "transforms into" trigger: the
// permanent turned over (EventTransform, ADR 0079) and the face now up
// is a Phyrexian. The engine emits the event with the permanent already
// on its new face, so the check reads its type line as it is NOW (the
// Incubator's back face is the Phyrexian Token; a real double-faced
// card's back reads the same way). It fires for any permanent the
// controller has, the Inquisitor itself included, and for a convert as
// well as a transform (CR 701.28 follows the rules for transforming).
//
// The counter goes on the permanent as it is at resolution, and only if
// it is still on the battlefield: a creature that left in response
// takes no counter (last-known information is read, never changed).
//
// A permanent that was ALREADY a Phyrexian and turns over to another
// Phyrexian face triggers it too, because the printed condition is about
// what it is after the turn-over. No printed card makes the two cases
// differ.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89a59a6d-e9d4-405c-8518-4da86dae30d1",
		Name:         "Norn's Inquisitor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Norn's Inquisitor — incubate 2", Do(Incubate{N: 2})),
			On(game.EventTransform, aPermanentYouControlTransformedIntoAPhyrexian,
				"Norn's Inquisitor — put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok || info.Left {
						return nil
					}
					return AddCounter{Target: ctx.Trigger().Object.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
		},
	})
}

// aPermanentYouControlTransformedIntoAPhyrexian is "whenever a
// permanent you control transforms into a Phyrexian": the turned-over
// permanent's controller is the source's, and the face it now shows has
// the Phyrexian creature type (read through HasSubtype, so a changeling
// counts, CR 702.73a).
func aPermanentYouControlTransformedIntoAPhyrexian(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller != source.Controller {
		return false
	}
	return c.HasSubtype("Phyrexian")
}
