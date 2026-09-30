package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gremlin Tamer — Creature — Human Scout {W}{U}, 2/2:
//
//	"Eerie — Whenever an enchantment you control enters and whenever you
//	fully unlock a Room, create a 1/1 red Gremlin creature token."
//
// The token is the table row "1/1 red Gremlin" (tokens_table.go).
//
// Eerie is one ability with two conditions (Eerie, rooms.go): an
// enchantment entering under your control, or you fully unlocking a Room.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4e34165-4e53-4c31-bf61-eb772bc3c054",
		Name:         "Gremlin Tamer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Eerie("Gremlin Tamer — create a 1/1 red Gremlin (eerie)", Do(CreateToken{Template: TokenCard("1/1 red Gremlin"), N: 1})),
		},
	})
}
