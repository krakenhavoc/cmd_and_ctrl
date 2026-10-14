package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Progenitor Exarch — Creature — Phyrexian Cat Cleric {X}{X}{W}, 1/2
// (March of the Machine):
//
//	"When this creature enters, incubate 3 X times.
//	 {T}: Transform target Incubator token you control."
//
// "Incubate 3 X times" is X separate incubates (Incubate.Times), each
// its own token with three +1/+1 counters: X = 2 is two Incubators of
// three, not one of six. X is the X the creature was cast with
// (CR 107.3m), read off the card as the trigger is built (Card.CastX)
// and stamped on the item, the way Springleaf Parade does; with X = 0
// (or an Exarch put onto the battlefield without being cast) the
// ability resolves and creates nothing.
//
// The {T} ability needs the Exarch to have been under its controller's
// control since the turn began (CR 302.6), which the engine enforces for
// every {T} cost on a creature. Its target is an Incubator token, front
// face up (IsIncubator).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4052f710-6ba7-406f-8fb7-653a684a80e0",
		Name:         "Progenitor Exarch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Progenitor Exarch — incubate 3 X times",
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Progenitor Exarch — incubate 3 X times")
				item.Params.Amount = source.CastX()
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := item.Params.Amount
				if x <= 0 {
					return nil
				}
				return Incubate{N: 3, Times: x}.Apply(NewContext(g, item))
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Transform target Incubator token you control.",
			Cost:    TapCost(),
			Targets: TargetIncubatorYouControl(),
			Effect:  func(g *game.Game, item *game.StackItem) error { return transformTargetIncubator(NewContext(g, item)) },
		}},
	})
}
