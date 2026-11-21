package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kasmina, Enigmatic Mentor — Legendary Planeswalker {3}{U}, loyalty 5
// (Jumpstart 2022, #5948):
//
//	"Spells your opponents cast that target a creature or planeswalker
//	 you control cost {2} more to cast.
//	 −2: Create a 2/2 blue Wizard creature token. Draw a card, then
//	 discard a card."
//
// Syr Elenora's target tax, read from the announced targets (so it is
// priced once the targets are chosen) and widened to any creature or
// planeswalker Kasmina's controller controls; the −2 is a token, then a
// loot.
//
// No simplification.
func init() {
	tax := effects.CostsMore(2, "Spells your opponents cast that target a creature or planeswalker you control cost {2} more to cast.",
		effects.OpponentsSpell(), effects.TargetsACreatureOrPlaneswalkerYouControl())
	tax.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:        "df561b38-fe2a-4db9-a9d4-ec25b1c7572a",
		Name:            "Kasmina, Enigmatic Mentor",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		CostModifiers:   []game.CostModifier{tax},
		Activated: []effects.ActivatedAbility{{
			Label:   "−2: Create a 2/2 blue Wizard creature token. Draw a card, then discard a card.",
			Cost:    effects.LoyaltyCost(-2),
			Purpose: game.Purpose{Tokens: 1, Draws: 1},
			Effect: effects.Do(
				effects.CreateToken{Template: effects.TokenCard("2/2 blue Wizard"), N: 1},
				effects.Instruction(func(ctx *effects.Context) error { return effects.LootOne(ctx.Game, ctx.Item, 1) }),
			),
		}},
	})
}
