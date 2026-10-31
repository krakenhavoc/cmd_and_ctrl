package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Namor the Sub-Mariner — {1}{U}{U} Legendary Creature — Mutant Merfolk
// Villain */4 (Marvel Super Heroes, #4455):
//
//	"Flying
//	 Namor's power is equal to the number of Merfolk you control.
//	 Whenever you cast a noncreature spell with one or more blue mana symbols
//	 in its mana cost, create that many 1/1 blue Merfolk creature tokens."
//
// The power is a characteristic-defining ability (CR 604.3, layer 7a): every
// permanent you control with the Merfolk subtype counts, Namor himself and a
// changeling included. The cast trigger counts the blue symbols of the
// spell's mana cost (CR 202.1: hybrid symbols that offer blue count once;
// {X} and generic costs count none) off the spell named by the cast event,
// and is skipped for a spell with none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dff11626-ad4d-443a-b7db-f508790b5fd2",
		Name:            "Namor the Sub-Mariner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:    "Power is equal to the number of Merfolk you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for i := range g.Battlefield.Cards {
					p := &g.Battlefield.Cards[i]
					if p.Controller == source.Controller && p.HasSubtype(source.TextSubtype("Merfolk")) {
						n++
					}
				}
				c.Power = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, mshB6NoncreatureSpellWithBlueSymbols,
				"Namor the Sub-Mariner — create a 1/1 blue Merfolk for each blue mana symbol in its mana cost",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					spell, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
					if !ok {
						return nil
					}
					n := mshB6ColorSymbols(spell.ManaCost, ctx.TextColor("U"))
					if n == 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 blue Merfolk"), N: n}.Apply(ctx)
				}),
		},
	})
}

// mshB6NoncreatureSpellWithBlueSymbols is "you cast a noncreature spell with
// one or more blue mana symbols in its mana cost".
func mshB6NoncreatureSpellWithBlueSymbols(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsCreature() && mshB6ColorSymbols(c.ManaCost, source.TextColor("U")) > 0
}

// mshB6ColorSymbols counts the mana symbols of `cost` that offer `color`,
// one per symbol however many colors the symbol offers.
func mshB6ColorSymbols(cost, color string) int {
	parsed, err := game.ParseCost(cost)
	if err != nil {
		return 0
	}
	n := 0
	for _, req := range parsed.Required {
		if symbolIsAnyColor(req.Options, []string{color}) {
			n++
		}
	}
	return n
}
