package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flayer Drone — {1}{B}{R} Creature — Eldrazi Drone 3/1 (Oath of the
// Gatewatch, S82 OLD b296):
//
//	"Devoid (This card has no color.)
//	 First strike
//	 Whenever another colorless creature you control enters, target opponent
//	 loses 1 life."
//
// Thalia's Lieutenant's "another … you control enters" watcher with the
// colour test read on the entering creature as it entered (CR 603.6a); the
// opponent is aimed as the trigger goes on the stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9e275428-3562-4ed6-8a49-3bb202a5130a",
		Name:            "Flayer Drone",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "first strike"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.IsColorless()
			}, "Flayer Drone — target opponent loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id := effects.TargetedPlayerOf(ctx)
					if id == uuid.Nil {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, id, -1)
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
