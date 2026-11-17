package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zephyr Singer — Creature — Siren Pirate {2}{U}{U} 3/4 (March of the
// Machine):
//
//	"Convoke
//	 Flying, vigilance
//	 When this creature enters, put a flying counter on each creature that
//	 convoked it."
//
// The creatures that convoked it ride the permanent (CR 400.7d,
// CastProvenance.Convoked) and are read as the trigger is built, so the
// counters still land if the Singer has left. One that has left the
// battlefield since is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1800ae74-0b54-4547-ab8f-ea61418d73e8",
		Name:            "Zephyr Singer",
		Completeness:    CompletenessFull,
		TapCost:         Convoke(),
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       zephyrSingerLabel,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, zephyrSingerLabel)
				item.Payload = convokersPayload(source)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, id := range ctx.PayloadCards() {
					if !onBattlefield(g, id) {
						continue
					}
					if err := (AddCounter{Target: id, Kind: game.CounterFlying, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

const zephyrSingerLabel = "Zephyr Singer — put a flying counter on each creature that convoked it"
