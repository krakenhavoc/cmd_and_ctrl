package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primaris Chaplain — Creature — Astartes Cleric {2}{W}{B}, 3/3:
//
//	"Battle cry
//	 Rosarius — Whenever this creature attacks, it gains indestructible until
//	 end of turn."
//
// Two separate attack triggers, as printed: the shared battle cry row, and
// Rosarius, which resolves independently of it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd024849-3005-4509-ad83-65a9bcf956ac",
		Name:         "Primaris Chaplain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.BattleCry("Primaris Chaplain"),
			effects.WheneverThisAttacks("Primaris Chaplain — it gains indestructible until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"indestructible"},
						Label:    "Primaris Chaplain — indestructible",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
