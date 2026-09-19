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
// DECLARED SIMPLIFICATION (weaker than printed): the {1}, {T} filter
// ability is NOT implemented. "Add two mana of DIFFERENT COLORS" is a
// constraint ACROSS two produced symbols, and the mana grammar has no
// way to say it — a pipe slot ("{W|U|B|R|G}{W|U|B|R|G}") picks each
// symbol independently and would happily produce {U}{U}, which is
// STRONGER than the printed card and is the direction #259 forbids.
// Rather than ship a filter that is better than the one on the card,
// the Beacon ships without it: a colourless land with a lifegain
// trigger. That is strictly less than printed.
//
// The rest is complete. "Whenever you CAST a planeswalker spell" is a
// cast trigger, so it fires with the spell still on the stack and
// pays even if the planeswalker is countered. The type is read off
// the spell, so a creature that is also a planeswalker (Gideon) and
// an artifact planeswalker both count.
//
// The restricted-mana machinery the missing ability would need does
// exist since #352 (ManaRestrictCast + ManaRestrictType), so when the
// grammar grows a "different colors" slot this file is a two-line
// change and the caveat goes away.
func init() {
	Register(Spec{
		OracleID:     "073169f2-da3a-4a93-8c01-b3fd8558d225",
		Name:         "Interplanar Beacon",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The {1}, {T} ability that adds two mana of different colors for planeswalker spells isn't available — the Beacon taps for {C} only.",
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsPlaneswalker()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Interplanar Beacon — you gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
