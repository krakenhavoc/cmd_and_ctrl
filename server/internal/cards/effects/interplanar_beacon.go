package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Interplanar Beacon — Land (EDHREC rank 4460):
//
//	"Whenever you cast a planeswalker spell, you gain 1 life.
//	 {T}: Add {C}.
//	 {1}, {T}: Add two mana of different colors. Spend this mana only
//	 to cast planeswalker spells."
//
// The superfriends deck's land: it filters colourless into two
// coloured pips for a planeswalker, and it pays a life every time one
// resolves. In a five-colour Atraxa list the filter is the whole
// reason to run it over a basic.
//
// The {1}, {T} filter is "two mana of different colors" (#2558,
// DifferentColors(2)): one pick of two DIFFERENT colours, never {U}{U},
// and both carry the planeswalker-only spend restriction (#352's
// ManaRestrictCast + ManaRestrictType). It shipped without that ability
// until the produced-mana grammar could say "different"; the caveat
// went with #2558. The auto-tapper never plans it, as it never plans
// any restricted mana; the player activates it from the card.
//
// "Whenever you CAST a planeswalker spell" is a
// cast trigger, so it fires with the spell still on the stack and
// pays even if the planeswalker is countered. The type is read off
// the spell, so a creature that is also a planeswalker (Gideon) and
// an artifact planeswalker both count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "073169f2-da3a-4a93-8c01-b3fd8558d225",
		Name:         "Interplanar Beacon",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         ManaAbilityCost{Mana: "{1}", Tap: true},
				Produced:     DifferentColors(2),
				Label:        "{1}, {T}: Add two mana of different colors. Spend this mana only to cast planeswalker spells.",
				Restrictions: []string{ManaRestrictCast, ManaRestrictType("Planeswalker")},
			},
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsPlaneswalker()
			},
			Key: "Interplanar Beacon — you gain 1 life",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
