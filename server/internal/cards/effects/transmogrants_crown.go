package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Transmogrant's Crown — Artifact — Equipment {2} (The Brothers' War, S82):
//
//	"Equipped creature gets +2/+0.
//	 Whenever equipped creature dies, draw a card.
//	 Equip {2} or {B}"
//
// Lead Pipe's shape (PumpAttached, equippedCreatureDied). "Equip {2} or
// {B}" is one equip ability the controller pays either cost of, written as
// two equip abilities that do the same thing, so either cost attaches it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "767c82a7-ed42-4ef3-9d79-dd95e48f4326",
		Name:         "Transmogrant's Crown",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return equippedCreatureDied(ev, source)
			}, "Transmogrant's Crown — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{broCrownEquip("{2}"), broCrownEquip("{B}")},
	})
}

// broCrownEquip is one of the Crown's two equip costs, labelled as the one
// printed line "Equip {2} or {B}".
func broCrownEquip(cost string) ActivatedAbility {
	ab := EquipAbility(cost)
	ab.Label = "Equip {2} or {B}"
	return ab
}
