package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cursed Courtier — Creature — Human Noble {1}{W}:
//
//	"Lifelink
//	 When this creature enters, create a Cursed Role token attached to
//	 it. (Enchanted creature is 1/1.)"
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "17916cf9-1e6c-41bd-96ce-2b050d828838",
		Name:            "Cursed Courtier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cursed Courtier — create a Cursed Role token attached to it",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.B09SourceStillOnBattlefield(g, item) {
						return nil
					}
					return effects.CreateRoleToken{Role: effects.RoleCursed, Host: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
