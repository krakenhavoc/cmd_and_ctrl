package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cult Healer — Creature — Human Doctor {2}{W}, 3/3:
//
//	"Eerie — Whenever an enchantment you control enters and whenever you
//	fully unlock a Room, this creature gains lifelink until end of turn."
//
// thisCreatureUntilEOT (eerie_helpers.go).
//
// Eerie is one ability with two conditions (Eerie, rooms.go): an
// enchantment entering under your control, or you fully unlocking a Room.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c51da69b-7fd7-43a3-be29-c44fc0c36130",
		Name:         "Cult Healer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Eerie("Cult Healer — gains lifelink until end of turn (eerie)",
				thisCreatureUntilEOT("Cult Healer — lifelink until end of turn", 0, 0, "lifelink")),
		},
	})
}
