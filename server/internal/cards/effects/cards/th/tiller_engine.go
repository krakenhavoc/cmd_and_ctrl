package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "62f6baba-da3b-45b8-a3c1-efb75763cca8",
		Name:         "Tiller Engine",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			Key:     effects.B31TillerEngineLabel,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B31LandYouControlEnteredTapped(ev, source, g)
			},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
			Modes: effects.ChooseOne(
				effects.ModeDoing("Untap that land.", nil,
					func(item *game.StackItem, ctx *effects.Context, _ int) error {
						entered := item.Trigger.Event.CardID
						if !effects.OnBattlefield(ctx.Game, entered) {
							return nil
						}
						return effects.UntapTarget{Target: entered}.Apply(ctx)
					}),
				effects.ModeDoing("Tap target nonland permanent an opponent controls.",
					effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls()),
					effects.TapTheModesTarget),
			),
		}},
	})
}
