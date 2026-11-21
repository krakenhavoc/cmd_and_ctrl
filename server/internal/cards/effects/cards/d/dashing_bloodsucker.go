package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dashing Bloodsucker — Creature — Vampire Warrior {3}{B}, 2/5:
//
//	"Eerie — Whenever an enchantment you control enters and whenever you
//	fully unlock a Room, this creature gets +2/+0 and gains lifelink until
//	end of turn."
//
// thisCreatureUntilEOT (eerie_helpers.go).
//
// Eerie is one ability with two conditions (Eerie, rooms.go): an
// enchantment entering under your control, or you fully unlocking a Room.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "211dca11-b633-4deb-9d51-210fd3843995",
		Name:         "Dashing Bloodsucker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Eerie("Dashing Bloodsucker — +2/+0 and lifelink until end of turn (eerie)",
				effects.ThisCreatureUntilEOT("Dashing Bloodsucker — +2/+0 and lifelink until end of turn", 2, 0, "lifelink")),
		},
	})
}
