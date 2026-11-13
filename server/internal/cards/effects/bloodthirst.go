package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bloodthirst.go — bloodthirst N (Guildpact, #5599):
//
//	CR 702.54a "Bloodthirst is a static ability. "Bloodthirst N" means
//	"If an opponent was dealt damage this turn, this permanent enters
//	with N +1/+1 counters on it.""
//
// A CR 614.1c self-replacement on the permanent's own entry, the same
// shape as revolt's and raid's "enters with counters if". The condition
// is read as the permanent enters, against the controller it enters
// under: any damage dealt this turn to any opponent, combat or not, from
// any source (anOpponentWasDealtAtLeastThisTurn(1), Spinerock Knoll's
// read). Damage dealt, not life lost: a drain does not count, and a
// prevented point was never dealt.
//
// Several instances each apply (CR 702.54c), one replacement row each.
// "Bloodthirst X" (CR 702.54b) is a different count and is not this.

// Bloodthirst is "Bloodthirst N" on the card `name`.
func Bloodthirst(n int, name string) game.ReplacementEffect {
	dealt := anOpponentWasDealtAtLeastThisTurn(1)
	return game.ReplacementEffect{
		Label:     name + " — bloodthirst " + strconv.Itoa(n),
		RowLabel:  "Bloodthirst " + strconv.Itoa(n) + ": enters with " + counterCountPhrase(n, game.CounterPlusOne) + " on it if an opponent was dealt damage this turn.",
		EntryOnly: true,
		Watches:   []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID &&
				dealt(g, src.Controller)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, n)
			return nil
		},
	}
}
