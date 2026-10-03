package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lightning, Army of One — Legendary Creature — Human Soldier {1}{R}{W},
// 3/2:
//
//	"First strike, trample, lifelink
//	 Stagger — Whenever Lightning deals combat damage to a player, until
//	 your next turn, if a source would deal damage to that player or a
//	 permanent that player controls, it deals double that damage
//	 instead."
//
// Stagger is an ability word (CR 207.2c). The trigger makes ADR 0108
// §3's multiplier with no source named — ANY source, the player's own
// included — over that player and the permanents they control as the
// damage would be dealt (CR 611.2c), until your next turn begins
// (CR 611.2b). Lightning's first-strike damage triggers it before the
// regular combat damage step, so the rest of the table's combat damage
// that turn is doubled too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "585eb5bc-5a3d-44d8-b593-1ff0d67f96a7",
		Name:            "Lightning, Army of One",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "trample", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Lightning, Army of One — Stagger", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				player := ctx.Trigger().Event.Target
				if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
					return nil
				}
				return MultiplyDamage{Factor: 2, Recipients: game.DamageRecipientsPlayerAndTheirPermanents,
					Player: player, UntilYourNextTurn: true,
					Label: "Lightning, Army of One — Stagger"}.Apply(ctx)
			}),
		},
	})
}
