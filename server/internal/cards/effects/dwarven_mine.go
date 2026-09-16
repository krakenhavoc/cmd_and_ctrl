package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dwarven Mine — Land — Mountain (EDHREC rank 2794):
//
//	"({T}: Add {R}.)
//	 This land enters tapped unless you control three or more other
//	 Mountains.
//	 When this land enters untapped, create a 1/1 red Dwarf creature
//	 token."
//
// The Throne of Eldraine "castle-lite" Mountain. The mana ability is
// the intrinsic one CR 305.6 gives every Mountain, derived by the
// engine from the type line. The enters-tapped clause is a
// self-replacement gated on the Mountain count — read while the Mine
// is still off the battlefield, so "other" needs no exclusion, and
// effective subtypes, so an Urborg-style type grant counts. The
// Dwarf is a real ETB trigger, with a response window, that reads
// the Mine's tapped state as it sits on the battlefield after the
// replacement decided it: entered untapped, one Dwarf; entered
// tapped, nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "74ed0bd3-ac31-41a4-8220-d8e7c8c1c437",
		Name:         "Dwarven Mine",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(func(g *game.Game, controller uuid.UUID) bool {
			return b12OtherMountainsControlled(g, controller, uuid.Nil) >= 3
		})},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b26SelfEnteredUntapped(ev, source)
			}, "Dwarven Mine — create a 1/1 red Dwarf", Do(CreateToken{Template: TokenCard("1/1 red Dwarf"), N: 1})),
		},
	})
}
