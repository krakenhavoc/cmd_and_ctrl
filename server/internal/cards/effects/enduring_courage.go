package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enduring Courage — Enchantment Creature — Dog Glimmer {2}{R}{R}, 3/3
// (Duskmourn, #4466):
//
//	"Whenever another creature you control enters, it gets +2/+0 and gains
//	 haste until end of turn.
//	 When Enduring Courage dies, if it was a creature, return it to the
//	 battlefield under its owner's control. It's an enchantment. (It's not a
//	 creature.)"
//
// The first half is one trigger on every other creature entering under your
// control; it pumps the creature that entered (read off the trigger's event)
// and, if that creature has left by resolution, does nothing (CR 608.2b).
// The dies half is the Enduring cycle's shared trigger
// (WhenThisDiesReturnItAsAnEnchantment).
//
// No simplification.
func init() {
	const label = "Enduring Courage — it gets +2/+0 and gains haste until end of turn"
	Register(Spec{
		OracleID:     "b3de29cb-7d7d-4566-b5bf-37e571bd5d78",
		Name:         "Enduring Courage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEnteredUnderYourControl, label, Do(Instruction(func(ctx *Context) error {
				if ctx.Item.Trigger == nil {
					return nil
				}
				id := ctx.Item.Trigger.Event.CardID
				if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
					return nil
				}
				if err := (BoostUntilEOT{Target: id, Power: 2, Label: "Enduring Courage — +2/+0 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Enduring Courage — haste until end of turn"}.Apply(ctx)
			}))),
			WhenThisDiesReturnItAsAnEnchantment("Enduring Courage"),
		},
	})
}
