package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// revolt.go — "if a permanent you controlled left the battlefield this
// turn" (revolt, CR 702.136; #2148).
//
// One fact, read off the per-player turn tally
// (game.PlayerTurnTally.PermanentsLeft): every battlefield exit counts,
// whatever the route or the destination, tokens included, under the
// controller the permanent had as it left. It is a fact about the TURN,
// not about the source: a Renegade that enters after the Treasure was
// sacrificed still sees it.
//
// Three readers, one per place the printed text puts the condition:
//
//   - Revolt(g, you) for a spell's or ability's own resolution
//     (Fatal Push's "instead if").
//   - WhenThisEntersIfRevolt / AtYourEndStepIfRevolt for an intervening
//     "if" (CR 603.4), checked as the trigger fires and again as it
//     resolves.
//   - SelfEntersWithCountersIfRevolt for "enters with N counters if",
//     a CR 614.1c self-replacement.

// Revolt reports whether a permanent `controller` controlled left the
// battlefield this turn. Caller must hold g.mu (every effect body does).
func Revolt(g *game.Game, controller uuid.UUID) bool {
	return g.PermanentLeftThisTurn(controller)
}

// IfRevolt wraps an Effect so it does nothing unless the item's
// controller has revolt as it resolves — the CR 603.4 recheck half of an
// intervening "if".
func IfRevolt(effect Effect) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !Revolt(g, item.Controller) {
			return nil
		}
		return effect(g, item)
	}
}

// YouHaveRevolt is the When for "if a permanent you controlled left the
// battlefield this turn", for composing with AllOf.
func YouHaveRevolt(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return Revolt(g, source.Controller)
}

// WhenThisEntersIfRevolt is "When ~ enters, if a permanent left the
// battlefield under your control this turn, …". Set Targets on the
// returned ability for a targeted body.
func WhenThisEntersIfRevolt(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventETB, AllOf(Self, YouHaveRevolt), label, IfRevolt(effect))
}

// AtYourEndStepIfRevolt is "At the beginning of your end step, if a
// permanent you controlled left the battlefield this turn, …". It
// triggers once however many left. Set Targets on the returned ability
// for a targeted body.
func AtYourEndStepIfRevolt(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventBeginEndStep, AllOf(ByYou, YouHaveRevolt), label, IfRevolt(effect))
}

// SelfEntersWithCountersIfRevolt is "this creature enters with <n>
// <kind> counters on it if a permanent left the battlefield under your
// control this turn". The condition is read as the creature enters; the
// entering creature has not left anything.
func SelfEntersWithCountersIfRevolt(kind string, n int) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID &&
				Revolt(g, src.Controller)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(kind, n)
			return nil
		},
	}
}
