package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dazzling Theater // Prop Room — Enchantment — Room (ADR 0103):
//
//	Dazzling Theater {3}{W}: "Creature spells you cast have convoke."
//	Prop Room {2}{W}: "Untap each creature you control during each other
//	 player's untap step."
//
// Prop Room is the untap-step permission Seedborn Muse uses, narrowed to
// creatures. Dazzling Theater is a granted tap cost (convoke_grant.go,
// the "grant-convoke" seam, #2137), gated on its door like every other
// Room ability.
//
// No simplification.
func init() {
	effects.Register(effects.Room(effects.RoomSpec{
		OracleID:     "c46a02db-13d6-477f-9da0-822599470168",
		Name:         "Dazzling Theater // Prop Room",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{GrantedTapCosts: []game.GrantedTapCost{
			effects.SpellsYouCastHaveConvoke("Dazzling Theater — creature spells you cast have convoke.", effects.SpellIs(effects.Creature())),
		}},
		Right: effects.Door{UntapStep: []game.UntapStepPermission{
			effects.UntapDuringEachOtherPlayersUntapStep("Prop Room — untap each creature you control",
				func(c game.Card) bool { return c.IsCreature() }),
		}},
	}))
}
