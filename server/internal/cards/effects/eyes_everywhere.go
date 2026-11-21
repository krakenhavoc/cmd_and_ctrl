package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7d0a1ae5-cd93-4789-9666-fc7bbbc6b613",
		Name:         "Eyes Everywhere",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Eyes Everywhere — scry 1", Do(Scry{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{5}{U}: Exchange control of this enchantment and target nonland permanent. Activate only as a sorcery.",
			Cost:         ManaCost("{5}{U}"),
			SorcerySpeed: true,
			Targets:      TargetPermanent("target nonland permanent", Nonland()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return ExchangeControl{A: item.SourceCardID, B: id, Label: "Eyes Everywhere — exchange control"}.Apply(ctx)
			},
		}},
	})
}
