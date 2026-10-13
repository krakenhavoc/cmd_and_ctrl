package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tiller Engine — Artifact Creature — Construct {2}, 1/3 (EDHREC
// rank 3309):
//
//	"Whenever a land you control enters tapped, choose one —
//	 • Untap that land.
//	 • Tap target nonland permanent an opponent controls."
//
// Amulet of Vigor with a second mode. "Enters tapped" is read off
// the land at the moment EventETB fires (b31LandYouControlEnteredTapped
// — every entry path stamps Tapped before it announces the entry),
// and each land that enters tapped is its own trigger, as printed.
//
// The modal trigger is a real "choose one" (TriggeredAbility.Modes,
// CR 603.3c), chosen as the trigger goes on the stack. The tap bullet
// is not offered while no opponent controls a nonland permanent,
// which leaves the untap. The land is read off the triggering event
// rather than captured, so a table with this trigger waiting on the
// stack is a restore point.
func init() {
	Register(Spec{
		OracleID:     "62f6baba-da3b-45b8-a3c1-efb75763cca8",
		Name:         "Tiller Engine",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			Key:     b31TillerEngineLabel,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b31LandYouControlEnteredTapped(ev, source, g)
			},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
			Modes: ChooseOne(
				ModeDoing("Untap that land.", nil,
					func(item *game.StackItem, ctx *Context, _ int) error {
						entered := item.Trigger.Event.CardID
						if !onBattlefield(ctx.Game, entered) {
							return nil
						}
						return UntapTarget{Target: entered}.Apply(ctx)
					}),
				ModeDoing("Tap target nonland permanent an opponent controls.",
					TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls()),
					TapTheModesTarget),
			),
		}},
	})
}
