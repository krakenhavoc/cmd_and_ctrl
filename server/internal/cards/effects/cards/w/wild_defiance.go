package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wild Defiance — {2}{G} Enchantment: "Whenever a creature you control becomes
// the target of an instant or sorcery spell, that creature gets +3/+3 until end
// of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76703951-f42f-4e1d-af51-e1e06a89bd5e",
		Name:         "Wild Defiance",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.CreatureYouControlTargetedByAnInstantOrSorcerySpell,
				"Wild Defiance — that creature gets +3/+3 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{Target: item.Trigger.Event.CardID, Power: 3, Toughness: 3, Label: "Wild Defiance — +3/+3"}.
						Apply(effects.NewContext(g, item))
				}),
		},
	})
}
