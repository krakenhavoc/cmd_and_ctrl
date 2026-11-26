package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strength from the Fallen — Enchantment {1}{G} (Journey into Nyx, S82 OLD,
// #5558):
//
//	"Constellation — Whenever this enchantment or another enchantment you
//	 control enters, target creature gets +X/+X until end of turn, where X is
//	 the number of creature cards in your graveyard."
//
// Agent of Erebos's constellation condition over Moodmark Painter's count of
// creature cards in your graveyard, read as the trigger resolves.
//
// No simplification.
func init() {
	creatureCards := effects.G2CardsInYourGraveyard(game.Card.IsCreature)
	effects.Register(effects.Spec{
		OracleID:     "23643af9-00b7-45e1-8471-7031d08bbe30",
		Name:         "Strength from the Fallen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.ConstellationTriggers, "Strength from the Fallen — target creature gets +X/+X until end of turn, where X is the number of creature cards in your graveyard (constellation)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					x := creatureCards(item, ctx)
					return effects.BoostUntilEOT{Target: t.ID, Power: x, Toughness: x, Label: "Strength from the Fallen — +X/+X"}.Apply(ctx)
				}), effects.TargetCreature("target creature")),
		},
	})
}
