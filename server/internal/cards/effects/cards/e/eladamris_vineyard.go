package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eladamri's Vineyard — Enchantment {G} (Secret Lair Drop, #5462):
//
//	"At the beginning of each player's first main phase, that player adds
//	 {G}{G}."
//
// Cheering Crowd's trigger over every player's precombat main phase (the
// first main phase, Hulking Raptor's reading): the mana goes to the player
// whose phase it is, read off the event (CR 106.4), and empties at the end of
// the phase as all mana does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "faa085c3-705f-465d-9290-0e22276ad06d",
		Name:         "Eladamri's Vineyard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginPrecombatMain, effects.AnyPlayer,
				"Eladamri's Vineyard — that player adds {G}{G}",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.TriggeringActor(item)
					if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
						return nil
					}
					return g.AddManaForEffect(player, item.SourceCardID, "{G}{G}")
				}),
		},
	})
}
