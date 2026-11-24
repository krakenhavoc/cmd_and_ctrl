package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scuttletide — Enchantment {1}{U} (Modern Horizons 2, #5501):
//
//	"{1}, Discard a card: Create a 0/3 blue Crab creature token.
//	 Delirium — Crabs you control get +1/+1 as long as there are four or
//	 more card types among cards in your graveyard."
//
// The lord is read through the layers (Crab is a creature type of the
// permanent, a changeling counts) and the delirium test counts the card
// types of the graveyard as it is now (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2be9c1e6-0226-497a-b548-21bf1cfbc286",
		Name:         "Scuttletide",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Delirium — Crabs you control get +1/+1 as long as there are four or more card types among cards in your graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: effects.WhileYouHave(effects.DrcDeliriumMet, func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Crab"))
			}),
			Apply: effects.PumpBy(1, 1),
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Discard a card: Create a 0/3 blue Crab creature token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.DiscardACard()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("0/3 blue Crab"), N: 1}),
		}},
	})
}
