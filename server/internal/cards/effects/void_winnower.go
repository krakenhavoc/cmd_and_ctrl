package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Void Winnower — Creature — Eldrazi {9}, 11/9 (slice 296-m):
//
//	"Your opponents can't cast spells with even mana values. (Zero is
//	 even.)
//	 Your opponents can't block with creatures with even mana values."
//
// The cast half is an ordinary game.CastRestriction (S42, ADR 0073
// §7): OpponentsCantCast narrowed by mana value parity, read off the
// spell's own ManaValue() — which is already zero for a spell with X
// in its cost while it is not on the stack (CR 107.3b), matching the
// printed parenthetical.
//
// The block half has no catalog constructor of its own — every card
// in the "restrictions" family (restrictions.go) is a fixed bit on
// a fixed set (self, attached, until-end-of-turn), and this is a
// permanent, board-wide, PARITY-conditioned restriction with no
// existing shape to reuse. It is a plain Layer 6 StaticAbility
// setting Characteristic.Restrictions |= CantBlock on any opponent's
// creature with an even mana value, read fresh on every layer
// recompute (game.BlockPairRefusalLocked already reads that bit for
// every other "can't block" card).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70ac902e-1eb5-4f83-a5a6-00ac89cfe50f",
		Name:         "Void Winnower",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			OpponentsCantCast("Your opponents can't cast spells with even mana values.", evenManaValueCard),
		},
		Static: []game.StaticAbility{
			{
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller != source.Controller && target.ManaValue()%2 == 0
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Restrictions |= game.CantBlock
				},
			},
		},
	})
}

func evenManaValueCard(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.ManaValue()%2 == 0
}
