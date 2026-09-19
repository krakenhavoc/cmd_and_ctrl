package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vehicles.go — the card-facing half of S27's Vehicle support
// (CR 301.7, CR 702.122). Two pieces, and they are two because they
// sit on opposite sides of the stack:
//
//	CrewCost(n)              the COST, paid at announce
//	BecomeCreatureUntilEOT   the EFFECT, applied at resolution
//
// Crew is an ordinary CR 602 activated ability. It uses the stack, so
// it can be responded to, and the creatures tap when the ability is
// ANNOUNCED — paying the cost — not when it resolves. Both fall out
// of putting the crew number in game.AbilityCost rather than doing
// anything special; the engine half lives in game/activated.go.
//
// A Vehicle is written as:
//
//	Activated: []ActivatedAbility{{
//	    Label:  "Crew 1",
//	    Cost:   CrewCost(1),
//	    Effect: CrewEffect("Smuggler's Copter"),
//	}},
//
// and that is the whole card, plus whatever triggers it prints.

// CrewCost is "Crew N" (CR 702.122a) — tap any number of untapped
// creatures you control with total power N or more.
//
// Deliberately NOT composed with TapCost(): the creatures that tap
// are not the source. A Vehicle that tapped itself to crew could
// never attack, which is the entire point of the mechanic.
func CrewCost(n int) game.AbilityCost { return game.AbilityCost{Crew: n} }

// BecomeCreatureUntilEOT is "this permanent becomes an artifact
// creature until end of turn" — the effect every crew ability has,
// and the one a handful of non-Vehicle cards share.
//
// Layer 4 (type-changing), pinned to one instance and registered in
// the turn-scoped registry (ADR 0035), so it expires at cleanup like
// any other until-end-of-turn effect.
//
// WHAT IT DOES NOT SET, AND WHY. A Vehicle prints its own power and
// toughness — Smuggler's Copter is a 3/3 on the card — and the deck
// importer stamps them onto game.Card like any creature's, so
// printedCharacteristic already reports 3/3 and the only thing
// missing is the CREATURE type. Setting P/T here as well would be a
// second, redundant source of truth that silently wins over the
// printed one the day a Vehicle gets a +1/+1 counter.
//
// Power / Toughness are therefore OPTIONAL and only for the cards
// that genuinely say a number the printed line doesn't — "becomes a
// 4/4 artifact creature" on a permanent that prints no P/T at all.
// Leave them zero for every Vehicle.
type BecomeCreatureUntilEOT struct {
	// Target is the permanent that becomes a creature. Zero means
	// "the source of the ability", which is what every crew ability
	// wants.
	Target uuid.UUID

	// Types are the card types ADDED, lowercase-insensitive. Empty
	// defaults to {"Artifact", "Creature"}, the crew wording.
	Types []string

	// Subtypes are creature types added alongside — "becomes a 4/4
	// Angel artifact creature". Empty adds none, which is right for
	// crew: a crewed Vehicle keeps the Vehicle subtype and gains no
	// creature type at all.
	Subtypes []string

	// SetPower / SetToughness override the printed P/T at layer 7b.
	// Both zero means "leave the printed values alone" — see the
	// type comment. A card that really wants a 0/0 is not
	// expressible and does not exist.
	SetPower     int
	SetToughness int

	Label string
}

func (b BecomeCreatureUntilEOT) Apply(ctx *Context) error {
	target := b.Target
	if target == uuid.Nil {
		target = ctx.Source()
	}
	if target == uuid.Nil {
		return nil
	}
	// CR 400.7: pin to the instance AND the battlefield-entry stamp,
	// so a Vehicle that is flickered in response stops being a
	// creature rather than coming back animated. Same contract the
	// other until-EOT primitives use.
	stamp := int64(0)
	found := false
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.InstanceID == target {
			stamp = c.EnteredBattlefieldAt
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	applies := func(t *game.Card, _ *game.Game, _ *game.Card) bool {
		return t.InstanceID == target && t.EnteredBattlefieldAt == stamp
	}

	types := b.Types
	if len(types) == 0 {
		types = []string{"Artifact", "Creature"}
	}
	addedTypes := append([]string(nil), types...)
	addedSubtypes := append([]string(nil), b.Subtypes...)
	label := eotLabel(b.Label, "becomes a creature until end of turn")

	ctx.Game.RegisterScopedStaticForEffect(game.StaticAbility{
		Layer:     game.Layer4Type,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, t := range addedTypes {
				if !eotHasType(c.Types, t) {
					c.Types = append(c.Types, t)
				}
			}
			for _, s := range addedSubtypes {
				if !eotHasType(c.Subtypes, s) {
					c.Subtypes = append(c.Subtypes, s)
				}
			}
		},
	}, ctx.Source(), label, ctx.Game.UntilEndOfTurnDuration())

	if b.SetPower != 0 || b.SetToughness != 0 {
		power, toughness := b.SetPower, b.SetToughness
		ctx.Game.RegisterScopedStaticForEffect(game.StaticAbility{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: applies,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power = power
				c.Toughness = toughness
			},
		}, ctx.Source(), label+" (base P/T)", ctx.Game.UntilEndOfTurnDuration())
	}
	return nil
}

// CrewEffect is the resolution half of every printed crew ability:
// the source becomes an artifact creature until end of turn.
// Returned as a closure because ActivatedAbility.Effect is a
// function, and because the card's name belongs in the label.
func CrewEffect(cardName string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return BecomeCreatureUntilEOT{
			Label: cardName + " — crewed",
		}.Apply(NewContext(g, item))
	}
}

// eotHasType reports whether a type / subtype token is already in the
// list, case-insensitively. Keeps the grant idempotent when a
// Vehicle is crewed twice in a turn, which is legal and common — the
// second crew is how you re-animate a Vehicle that a removal spell
// turned back into an artifact.
func eotHasType(list []string, want string) bool {
	for _, t := range list {
		if len(t) == len(want) && equalFoldASCIIEffects(t, want) {
			return true
		}
	}
	return false
}

// equalFoldASCIIEffects is strings.EqualFold restricted to ASCII.
// Type and subtype names are ASCII in every Scryfall type line.
func equalFoldASCIIEffects(a, b string) bool {
	for i := 0; i < len(a); i++ {
		x, y := a[i], b[i]
		if x >= 'A' && x <= 'Z' {
			x += 'a' - 'A'
		}
		if y >= 'A' && y <= 'Z' {
			y += 'a' - 'A'
		}
		if x != y {
			return false
		}
	}
	return true
}
