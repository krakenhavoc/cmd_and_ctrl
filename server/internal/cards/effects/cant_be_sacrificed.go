package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cant_be_sacrificed.go — the card-facing half of "can't be sacrificed"
// (#3275). The engine half, and every route that reads it, is
// game/cant_be_sacrificed.go. A card file says which form it prints:
//
//	Static: []game.StaticAbility{CantBeSacrificed()},           // Alexios, Deimos of Kosmos
//	CantBeSacrificedThisTurn{Target: id}.Apply(ctx)             // Slicer, Hired Muscle
//
// The static is the permanent's own ability, so it goes with its
// abilities (CR 613.1f). The "this turn" form is a resolved effect's
// rule about one object (CR 611.2), pinned to it (CR 400.7), so a
// permanent that loses all abilities keeps it and a flickered one does
// not.

// CantBeSacrificed is "~ can't be sacrificed" (Alexios, Deimos of
// Kosmos; Hithlain Rope's "This artifact can't be sacrificed").
func CantBeSacrificed() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Can't be sacrificed.",
		Layer:     game.Layer6Ability,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.CantBeSacrificed = true
		},
	}
}

// CantBeSacrificedThisTurn is "<Target> can't be sacrificed this turn"
// (Slicer, Hired Muscle): until cleanup, a cost can't be paid with it,
// an effect can't choose it, and an instruction to sacrifice it does
// nothing (CR 101.2, 118.3).
type CantBeSacrificedThisTurn struct {
	Target uuid.UUID
	Label  string
}

func (c CantBeSacrificedThisTurn) Apply(ctx *Context) error {
	if c.Target == uuid.Nil || ctx.isNewSourceObject(c.Target) { // #1432
		return nil
	}
	label := c.Label
	if label == "" {
		label = "Can't be sacrificed this turn"
	}
	ctx.Game.CantBeSacrificedThisTurnForEffect(ctx.Source(), c.Target, label)
	return nil
}
