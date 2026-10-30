package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alex Wilder, Runaway — {1}{R} Legendary Creature — Human Hero Villain 1/3:
//
//	"Whenever Alex Wilder or another creature you control enters, if you
//	 cast it from anywhere other than your hand, it gets +2/+0 and gains
//	 haste until end of turn.
//	 Escape—{2}{R}, Exile three other cards from your graveyard. (You may
//	 cast this card from your graveyard for its escape cost.)"
//
// "If you cast it from anywhere other than your hand" is the entered
// permanent's cast record (CR 400.7d, Card.Provenance): cast by the player
// who controls it now, from a zone that is not a hand. A creature put onto
// the battlefield without a cast (a token, a reanimation) has no record and
// never qualifies; Alex himself escaped from the graveyard does (he is
// "Alex Wilder or another creature"). Both clauses read the creature as the
// object that entered: if it has left by resolution there is nothing to
// pump (a last-known-information object cannot be changed, CR 608.2h).
//
// Escape is the engine's (CR 702.138): Escape("{2}{R}", 3) from the
// graveyard.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	const label = "Alex Wilder — it gets +2/+0 and gains haste until end of turn"
	Register(Spec{
		OracleID:         "f3031397-c2d5-40c6-8eca-88faed00a5ea",
		Name:             "Alex Wilder, Runaway",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Escape("{2}{R}", 3)},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, msc18CreatureYouCastFromOutsideYourHand, label, msc18PumpTheEnteredCreature(label)),
		},
	})
}

// msc18CreatureYouCastFromOutsideYourHand is the trigger and its intervening
// if: a creature under the source's controller's control entered (the
// source itself counts), and its controller cast it from a zone other than a
// hand.
func msc18CreatureYouCastFromOutsideYourHand(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() {
		return false
	}
	cast, known := c.CastByItsController()
	return known && cast && c.Provenance.FromZone != "" && c.Provenance.FromZone != game.ZoneHand
}

// msc18PumpTheEnteredCreature is "it gets +2/+0 and gains haste until end of
// turn", on the creature that entered, if it is still that object.
func msc18PumpTheEnteredCreature(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		info, ok := ctx.TriggeringPermanent()
		if !ok || info.Left {
			return nil
		}
		id := ctx.Trigger().Object.Ref().ID
		if err := (BoostUntilEOT{Target: id, Power: 2, Label: label}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: label}.Apply(ctx)
	}
}
