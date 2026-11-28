package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flailing Drake — Creature — Drake {3}{G}, 2/3 (Tempest, S82 OLD b756, #5628):
//
//	"Flying
//	 Whenever this creature blocks or becomes blocked by a creature, that
//	 creature gets +1/+1 until end of turn."
//
// Witherscale Wurm's per-pair block trigger (CR 509.3b, 509.3d): one event for
// each blocker-attacker pair that names the Drake, so it fires once for each
// creature it blocks or is blocked by, and the other creature of that pair gets
// the bonus. A creature that has left the battlefield by resolution gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "93b6a170-a816-4556-aa27-a729231715e7",
		Name:            "Flailing Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SkewerSlingerFoe(ev, source.InstanceID) != uuid.Nil
			}, "Flailing Drake — that creature gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					foe := effects.SkewerSlingerFoe(item.Trigger.Event, item.SourceCardID)
					if foe == uuid.Nil || !effects.OnBattlefield(g, foe) {
						return nil
					}
					return effects.BoostUntilEOT{Target: foe, Power: 1, Toughness: 1, Label: "Flailing Drake — +1/+1"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
