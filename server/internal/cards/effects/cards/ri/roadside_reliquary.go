package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2fb13687-0518-4ba0-a5ae-dd609464b026",
		Name:         "Roadside Reliquary",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice Roadside Reliquary: Draw a card if you control an artifact. Draw a card if you control an enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  roadsideReliquaryDraws,
		}},
	})
}

// roadsideReliquaryDraws is the two conditional draws, each asked of
// the board as it stands when the ability resolves.
func roadsideReliquaryDraws(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.IsArtifact() }) > 0 {
		if err := (effects.DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.IsEnchantment() }) > 0 {
		return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
	}
	return nil
}
