package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b13_helpers.go — shared pieces for the Aetherdrift leftovers of
// Standard batches 13 and 14 (#4464, #4452).

// plusOneCounterPlacementOnYourCreatureOrVehicle is
// plusOneCounterPlacementOnYourCreature for "a creature or Vehicle you
// control" (Caradora, Heart of Alacria): an uncrewed Vehicle is not a
// creature but is named by the printed text.
func plusOneCounterPlacementOnYourCreatureOrVehicle(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 || ev.CounterName != "+1/+1" {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.CounterTarget)
	if !ok || target.Controller != src.Controller {
		return false
	}
	return target.IsCreature() || target.HasSubtype("Vehicle")
}

// b13MountOrVehicleCard is "a Mount or Vehicle card" as a search
// predicate. "Mount" is a word of the searching card's text (CR 612.2),
// read through the lens SearchLibrary lends.
func b13MountOrVehicleCard(c game.Card) bool {
	return containsFoldASCII(c.TypeLine, strings.ToLower(c.LentSubtype("Mount"))) || c.HasSubtype("Vehicle")
}

// b13CastFromGraveyardPayingLifeAndSacrificing is "You may cast this card
// from your graveyard by paying <life> life and sacrificing <label> in
// addition to paying its other costs" (Wickerfolk Indomitable). It is not
// an alternative cost (CR 118.8): the printed mana cost is still paid. It
// is built, as retrace is (ADR 0066, 2026-10-07 amendment), as an offer
// that carries an additional cost's rule: ManaCost is the card's printed
// cost, never empty, and claiming it is what opens the graveyard, so the
// life and the sacrifice are owed exactly on a cast from the graveyard and
// never on the hand cast of the same card.
func b13CastFromGraveyardPayingLifeAndSacrificing(mana string, life int, label string, preds ...CardPredicate) game.AlternativeCost {
	ac := SacrificeInstead(1, label, 0, preds...)
	ac.Key = "graveyard_life_sacrifice"
	ac.Label = "Cast this card from your graveyard by paying " + mana + ", paying 2 life and sacrificing " + label
	ac.ManaCost = mana
	ac.Life = life
	ac.FromZone = game.ZoneGraveyard
	return ac
}

// artifactOrCreatureCardLeftYourGraveyard is "an artifact and/or creature
// card leaves your graveyard" (Dredger's Insight), asked of the card as it
// was in the graveyard (CR 603.10a, wasCardInAGraveyard).
func artifactOrCreatureCardLeftYourGraveyard(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := cardThatLeftYourGraveyard(ev, source, g)
	return ok && wasCardInAGraveyard(c, func(c game.Card) bool { return c.IsArtifact() || c.IsCreature() })
}

// b13CreatureCardWithNoAbilities is "a creature card with no abilities" as
// a search predicate: a creature card whose printed text is empty
// (game.PrintsNoAbilities, stamped from the oracle text on import).
func b13CreatureCardWithNoAbilities(c game.Card) bool {
	return c.IsCreature() && game.PrintsNoAbilities(c)
}

// b13NoAbilities is b13CreatureCardWithNoAbilities as a target predicate
// ("target creature card with no abilities", Rise from the Wreck).
func b13NoAbilities() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return b13CreatureCardWithNoAbilities(c) }
}

// pirateAndVehicleAttacked is "a Pirate and a Vehicle attacked this
// combat" (Fearless Swashbuckler): among the objects declared as attackers
// in the combat phase in progress there is a Pirate and there is a Vehicle,
// which may be one object that is both. Caller must hold g.mu.
func pirateAndVehicleAttacked(g *game.Game) bool {
	pirate, vehicle := false, false
	for _, a := range g.AttackersThisCombatForEffect() {
		c, ok := g.LookupCardForEffect(a.ID)
		if !ok {
			continue
		}
		pirate = pirate || isPirate(c)
		vehicle = vehicle || c.HasSubtype("Vehicle")
	}
	return pirate && vehicle
}

// CondPirateAndVehicleAttacked is pirateAndVehicleAttacked as an
// intervening-if Condition. It reads this combat's attack records, which
// are not one of the footprint conditions, so it declares none.
type CondPirateAndVehicleAttacked struct{}

func (CondPirateAndVehicleAttacked) Holds(ctx *Context) bool {
	return pirateAndVehicleAttacked(ctx.Game)
}

func (CondPirateAndVehicleAttacked) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}

// thenPutFoundCheapArtifactsOntoBattlefield is Guidelight Pathmaker's "put it
// onto the battlefield if its mana value is 2 or less": over the cards a
// search put into the hand, each with mana value 2 or less (those still in
// the hand) enters under Tail.Player's control; a more expensive one stays
// in the hand. Archdruid's Charm's shape: a card that passes through the
// hand on the way is not observable, since nothing happens between the two
// moves. An on-disk key.
var thenPutFoundCheapArtifactsOntoBattlefield = game.RegisterTail("then/put-found-artifacts-with-mana-value-2-or-less-onto-battlefield",
	func(g *game.Game, o game.TailOutcome) error {
		for _, id := range o.IDs {
			c, ok := g.LookupCardForEffect(id)
			if !ok {
				continue
			}
			if mv, ok := g.ManaValueForEffect(c); !ok || mv > 2 {
				continue
			}
			if _, err := g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{Controller: o.Tail.Player}); err != nil {
				return err
			}
		}
		return nil
	})

// vehicleCreaturesAreAlsoTheChosenType is "Vehicle creatures you control are
// the chosen creature type in addition to their other types" (Lifecraft
// Engine): IsAlsoTheChosenType's layer 4 append, over the Vehicles its
// controller controls that are creatures. "Vehicle creatures" reads the
// layered view, so a Vehicle that has been crewed counts and one that
// hasn't doesn't; crew is a layer 4 effect too, and CR 613.8 dependency
// ordering applies the crew effect first, whatever the timestamps.
func vehicleCreaturesAreAlsoTheChosenType() game.StaticAbility {
	return game.StaticAbility{
		Label: "Vehicle creatures you control are the chosen creature type in addition to their other types.",
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return source.NamedTribe != "" && target.Controller == source.Controller &&
				target.IsCreature() && target.HasSubtype("Vehicle")
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			addChosenTribe(c, source)
		},
	}
}

// addChosenTribe appends the creature type the source named to c's subtypes
// unless it already has it (a changeling, or a permanent that is already of
// that type), because the wire type line is rebuilt from the slice and would
// print the type twice. Shared by IsAlsoTheChosenType and
// vehicleCreaturesAreAlsoTheChosenType.
func addChosenTribe(c *game.Characteristic, source *game.Card) {
	for _, t := range c.Subtypes {
		if t == source.NamedTribe {
			return
		}
	}
	c.Subtypes = append(c.Subtypes, source.NamedTribe)
}
