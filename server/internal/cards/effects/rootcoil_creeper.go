package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rootcoil Creeper — Creature — Plant Horror {G}{U}, 2/2 (Midnight Hunt):
//
//	"{T}: Add one mana of any color.
//	 {T}: Add two mana of any one color. Spend this mana only to cast spells
//	 from your graveyard.
//	 {G}{U}, {T}, Exile this creature: Return target card with flashback you
//	 own from exile to your hand."
//
// Two mana abilities, the second carrying the graveyard-cast restriction
// (ManaRestrictCastFromGraveyard, Lord of the Forsaken's), and an activated
// ability whose cost exiles the Creeper. "Card with flashback" is the card's
// own printed flashback (Sorceress's Schemes), so a card granted flashback
// does not qualify; the target is re-judged as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "101f06f8-7980-41d1-98fa-3cb959380737",
		Name:         "Rootcoil Creeper",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
			{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     OneColorOfAmount(2),
				Label:        "Add two mana of any one color. Spend this mana only to cast spells from your graveyard",
				Restrictions: []string{game.ManaRestrictCastFromGraveyard},
			},
		},
		Activated: []ActivatedAbility{{
			Label:   "{G}{U}, {T}, Exile this creature: Return target card with flashback you own from exile to your hand.",
			Cost:    Plus(ManaCost("{G}{U}"), TapCost(), ExileThis()),
			Targets: TargetCardInExile("target card with flashback you own from exile", s82MidB11FlashbackYouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					_, err := g.ReturnFromAnyZoneToHandForEffect(item.Controller, item.SourceCardID, t.ID)
					return err
				}
				return nil
			},
		}},
	})
}
