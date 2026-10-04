package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Skullspore Nexus — Legendary Artifact {6}{G}{G}:
//
//	"This spell costs {X} less to cast, where X is the greatest power
//	 among creatures you control.
//	 Whenever one or more nontoken creatures you control die, create a
//	 green Fungus Dinosaur creature token with base power and toughness
//	 each equal to the total power of those creatures.
//	 {2}, {T}: Double target creature's power until end of turn."
//
// THE DISCOUNT is The Great Henge's: X priced at CR 601.2f with the
// spell already on the stack, spent against generic mana only, and
// fixed once mana abilities start (the 2023-11-10 rulings).
//
// THE TRIGGER is a batch trigger (CR 603.2c): one per event that kills
// one or more of your nontoken creatures, so a wrath makes one token,
// not one per creature. As it resolves it walks that batch's events
// and adds up the power of each nontoken creature you controlled that
// died in it, as it last existed on the battlefield (CR 603.10a, the
// ruling): counters and anthems count, and a creature that shrank to a
// negative power subtracts. The total is the token's base power and
// toughness (layer 7b), not a +X/+X on top of something.
//
// THE ACTIVATION is CR 701.10b's double: +X/+0 until end of turn, X
// the target's power as the ability resolves (doubling.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba26ff0a-e714-44f2-95cf-1a5a6088edf9",
		Name:         "The Skullspore Nexus",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(greatestPowerAmongCreaturesYouControl(),
				"This spell costs {X} less to cast, where X is the greatest power among creatures you control."),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventLTB, skullsporeNontokenCreatureYouControlDied,
				"The Skullspore Nexus — create a Fungus Dinosaur with the total power of those creatures",
				skullsporeFungusDinosaur)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Double target creature's power until end of turn.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return DoublePowerUntilEOT(ctx, id, "The Skullspore Nexus — double power")
			},
		}},
	})
}

// skullsporeNontokenCreatureYouControlDied is one death in the batch:
// a nontoken creature the Nexus's controller controlled as it died.
func skullsporeNontokenCreatureYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && !IsToken(dead) && leftUnderControlOf(ev, dead) == source.Controller
}

// skullsporeFungusDinosaur totals the batch's qualifying deaths and
// makes the token.
func skullsporeFungusDinosaur(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	batch := item.Trigger.Event.Batch
	source := game.Card{Controller: item.Controller}
	total := 0
	seen := map[uuid.UUID]bool{}
	for _, ev := range g.EventsThisTurn() {
		// An unstamped event (Batch zero) is a batch of one: itself.
		sameBatch := ev.Batch == batch && (batch != 0 || ev.Seq == item.Trigger.Event.Seq)
		if !sameBatch || seen[ev.CardID] || !skullsporeNontokenCreatureYouControlDied(ev, &source, game.Characteristic{}, g) {
			continue
		}
		seen[ev.CardID] = true
		if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok {
			total += info.Power
		}
	}
	return CreateToken{
		Controller: item.Controller,
		Template: game.Card{
			Name:      "Fungus Dinosaur",
			TypeLine:  "Token Creature — Fungus Dinosaur",
			Colors:    []string{"G"},
			Power:     total,
			Toughness: total,
		},
		N: 1,
	}.Apply(NewContext(g, item))
}
