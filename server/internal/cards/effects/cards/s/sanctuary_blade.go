package cardss

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sanctuary Blade — Artifact — Equipment {2} (Commander 2020, S82 OLD, #5513):
//
//	"As this Equipment becomes attached to a creature, choose a color.
//	 Equipped creature gets +2/+0 and has protection from the last chosen
//	 color.
//	 Equip {3}"
//
// Dinosaur Headdress's "as it becomes attached" hook (CR 614.12 shape, run by
// the engine's one attach verb every time the Blade becomes attached, by an
// equip or any other effect) asks its controller for a color and stores it on
// the Blade, where it survives the Blade being unattached or moved: the next
// attach asks again and the newest answer wins ("the last chosen color"). The
// static reads the stored color live, so an Equipment attached before any
// answer has the pump and no protection. The Blade is colorless, so protection
// from the chosen color never makes it fall off (CR 301.5c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab0dd461-25bc-4374-bed8-72fc2d51b5e1",
		Name:         "Sanctuary Blade",
		Completeness: effects.CompletenessFull,
		AsAttached:   effects.ChooseColorAsEnters(game.ColorForProtection, "Sanctuary Blade"),
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			{
				Layer:     game.Layer6Ability,
				Label:     "Equipped creature has protection from the last chosen color.",
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					if source.ChosenColor == "" {
						return
					}
					tok := game.ProtectionFromColor(source.ChosenColor)
					if !slices.Contains(c.Abilities, tok) {
						c.Abilities = game.AppendKeywordAbility(c.Abilities, tok)
					}
				},
			},
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
