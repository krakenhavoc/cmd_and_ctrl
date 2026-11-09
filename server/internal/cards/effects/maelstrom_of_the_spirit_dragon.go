package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maelstrom of the Spirit Dragon — Land (Tarkir: Dragonstorm, #4463):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a Dragon
//	 spell or an Omen spell.
//	 {4}, {T}, Sacrifice this land: Search your library for a Dragon card,
//	 reveal it, put it into your hand, then shuffle."
//
// The restriction is one cast-only tag over two subtypes (the `|`
// alternation of game.ManaRestrictSubtype's keyed tags): a Dragon spell is
// a spell with the Dragon subtype, and an Omen spell is an instant or
// sorcery spell with the Omen spell type, which a card cast as its Omen
// has on the stack (CR 720.3a) and a permanent half never does. The mana
// can't pay an activation cost. "Any color" is flat: the text does not
// mention a commander's identity. The search is Abzan Monument's with a
// Dragon card in place of a basic land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49e9fba7-8465-4bbb-95db-73a7e149f494",
		Name:         "Maelstrom of the Spirit Dragon",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     "{W|U|B|R|G}",
				Label:        "Add one mana of any color. Spend this mana only to cast a Dragon spell or an Omen spell.",
				Restrictions: []string{game.ManaRestrictCast, "subtype:Dragon|Omen"},
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{4}, {T}, Sacrifice this land: Search your library for a Dragon card, reveal it, put it into your hand, then shuffle.",
			Cost:  Plus(ManaCost("{4}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtype("Dragon") },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Maelstrom of the Spirit Dragon — a Dragon card",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
