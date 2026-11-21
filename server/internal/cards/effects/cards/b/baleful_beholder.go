package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baleful Beholder — Creature — Beholder {4}{B}{B}, 6/5 (AFR):
//
//	"When this creature enters, choose one —
//	 • Antimagic Cone — Each opponent sacrifices an enchantment of their choice.
//	 • Fear Ray — Creatures you control gain menace until end of turn."
func init() {
	beholder := effects.WhenThisEnters("Baleful Beholder — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	beholder.Modes = effects.ChooseOne(
		effects.ModeDoing("Antimagic Cone — Each opponent sacrifices an enchantment of their choice.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Enchantment(), Label: "an enchantment"}.Apply(ctx)
			}),
		effects.ModeDoing("Fear Ray — Creatures you control gain menace until end of turn.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"menace"},
					Label: "Baleful Beholder — Fear Ray: menace until end of turn"}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "e7645ac4-6e05-426a-a62d-a33b30c33102",
		Name:         "Baleful Beholder",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{beholder},
	})
}
