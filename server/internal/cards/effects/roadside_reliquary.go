package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roadside Reliquary — Land:
//
//	"{T}: Add {C}."
//	"{2}, {T}, Sacrifice this land: Draw a card if you control an
//	 artifact. Draw a card if you control an enchantment."
//
// A colourless utility land that cashes itself in for up to two
// cards. Both halves are ordinary: a plain colourless tap ability,
// and a CR 602 activated ability whose cost composes mana, the tap
// and the sacrifice-this component.
//
// The two draws are INDEPENDENT sentences, not one "draw two if you
// have both": a board with an artifact and no enchantment draws one,
// a board with both draws two, and an empty board draws none. They
// are written as two conditions and two draws for that reason.
//
// The land is gone by the time the ability resolves — a sacrifice is
// a cost, paid at announce (CR 601.2h) — so the board it asks about
// is the one that is left. It never asks about itself: a land is
// neither an artifact nor an enchantment.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "2fb13687-0518-4ba0-a5ae-dd609464b026",
		Name:         "Roadside Reliquary",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:  "{2}, {T}, Sacrifice Roadside Reliquary: Draw a card if you control an artifact. Draw a card if you control an enchantment.",
			Cost:   Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Effect: roadsideReliquaryDraws,
		}},
	})
}

// roadsideReliquaryDraws is the two conditional draws, each asked of
// the board as it stands when the ability resolves.
func roadsideReliquaryDraws(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if countControlled(g, item.Controller, func(c game.Card) bool { return c.IsArtifact() }) > 0 {
		if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if countControlled(g, item.Controller, func(c game.Card) bool { return c.IsEnchantment() }) > 0 {
		return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
	}
	return nil
}
