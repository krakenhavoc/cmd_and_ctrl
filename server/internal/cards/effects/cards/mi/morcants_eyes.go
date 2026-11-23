package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morcant's Eyes — Kindred Enchantment — Elf {1}{G} (Lorwyn Eclipsed, #4458):
//
//	"At the beginning of your upkeep, surveil 1.
//	 {4}{G}{G}, Sacrifice this enchantment: Create X 2/2 black and green Elf
//	 creature tokens, where X is the number of Elf cards in your graveyard.
//	 Activate only as a sorcery."
//
// Sacrificing is a cost, so the Eyes (an Elf card) are in the graveyard when
// X is counted as the ability resolves and count themselves.
func init() {
	const label = "{4}{G}{G}, Sacrifice this enchantment: Create X 2/2 black and green Elf creature tokens, where X is the number of Elf cards in your graveyard. Activate only as a sorcery."
	effects.Register(effects.Spec{
		OracleID:     "c55a27cc-8e25-4714-b1da-86e5c1c6661a",
		Name:         "Morcant's Eyes",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourStep(game.StepUpkeep, "Morcant's Eyes — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        label,
			Cost:         effects.Plus(effects.ManaCost("{4}{G}{G}"), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				elf := ctx.TextSubtype("Elf")
				n := len(effects.B42CardsInGraveyardOf(g, item.Controller, func(c game.Card) bool { return c.HasSubtype(elf) }))
				if n == 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("2/2 black and green Elf"), N: n}.Apply(ctx)
			},
		}},
	})
}
