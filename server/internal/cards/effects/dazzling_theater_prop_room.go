package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Room(RoomSpec{
		OracleID:     "c46a02db-13d6-477f-9da0-822599470168",
		Name:         "Dazzling Theater // Prop Room",
		Completeness: CompletenessFull,
		Left: Door{GrantedTapCosts: []game.GrantedTapCost{
			SpellsYouCastHaveConvoke("Dazzling Theater — creature spells you cast have convoke.", SpellIs(Creature())),
		}},
		Right: Door{UntapStep: []game.UntapStepPermission{
			untapDuringEachOtherPlayersUntapStep("Prop Room — untap each creature you control",
				func(c game.Card) bool { return c.IsCreature() }),
		}},
	}))
}
