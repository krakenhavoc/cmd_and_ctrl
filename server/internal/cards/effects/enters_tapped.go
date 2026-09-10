package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// enters_tapped.go — the CR 614 self-replacement for "~ enters
// tapped". Its own file rather than helpers.go, per the convention
// #231 set: concurrent card batches collide on shared helper files.

// SelfEntersTapped is the CR 614 self-replacement behind "This land
// enters tapped" / "~ enters tapped" — every Temple, Guildgate and
// tri-land, and the honest version of what Worn Powerstone previously
// approximated with an OnETB tap.
//
// The distinction is observable: an OnETB tap means the permanent
// enters UNTAPPED and is tapped a beat later, so anything watching for
// an untapped permanent entering, or for a tap event, sees the wrong
// thing. A replacement means it was never untapped on the battlefield
// at all.
//
// Declared on Spec.Replacements. The engine consults an entering
// card's own replacements specifically so this works — the pipeline
// runs pre-push, so the card is not on the battlefield to be found by
// the ordinary walk.
func SelfEntersTapped() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = true
			return nil
		},
	}
}
