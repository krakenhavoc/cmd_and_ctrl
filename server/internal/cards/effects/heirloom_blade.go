package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heirloom Blade — Artifact — Equipment {3}:
//
//	"Equipped creature gets +3/+1.
//	 Whenever equipped creature dies, you may reveal cards from the top of
//	 your library until you reveal a creature card that shares a creature
//	 type with it. Put that card into your hand and the rest on the bottom
//	 of your library in a random order.
//	 Equip {1}"
func init() {
	Register(Spec{
		OracleID:     "7c3a8766-e440-48c5-a561-197490efdba5",
		Name:         "Heirloom Blade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(3, 1)},
		Triggered: []game.TriggeredAbility{
			Optional(
				On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return equippedCreatureDied(ev, source)
				}, "Heirloom Blade — reveal cards until a creature card that shares a creature type with it and put it into your hand",
					heirloomBladeReveal),
				"Heirloom Blade — reveal cards until a creature card that shares a creature type with the dead creature?"),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}

// heirloomBladeReveal reads the creature types the equipped creature had as
// it died (CR 603.10: the trigger's last-known object), and a changeling's
// printed all-types from its card in the graveyard.
func heirloomBladeReveal(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Object == nil {
		return nil
	}
	every := false
	if c, ok := g.LookupCardForEffect(item.Trigger.Object.ID); ok {
		every = game.HasAllCreatureTypes(&c)
	}
	word := cmdr2B03TypesWord(item.Trigger.Object.Subtypes, every)
	return cmdr2B03RevealUntilSharingTypeToHand(NewContext(g, item), word,
		"Heirloom Blade — reveal until a creature card that shares a creature type")
}
