package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eyes Everywhere — Enchantment {2}{U} (Ravnica Remastered):
//
//	"At the beginning of your upkeep, scry 1.
//	 {5}{U}: Exchange control of this enchantment and target nonland
//	 permanent. Activate only as a sorcery."
//
// Avarice Totem's exchange (ExchangeControl) with a sorcery-speed gate. The
// exchange does nothing if either permanent has left (CR 701.12b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d0a1ae5-cd93-4789-9666-fc7bbbc6b613",
		Name:         "Eyes Everywhere",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Eyes Everywhere — scry 1", effects.Do(effects.Scry{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{5}{U}: Exchange control of this enchantment and target nonland permanent. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{5}{U}"),
			SorcerySpeed: true,
			Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.ExchangeControl{A: item.SourceCardID, B: id, Label: "Eyes Everywhere — exchange control"}.Apply(ctx)
			},
		}},
	})
}
