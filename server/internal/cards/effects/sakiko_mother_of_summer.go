package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sakiko, Mother of Summer — Legendary Creature — Snake Shaman {4}{G}{G},
// 3/3 (Commander Masters):
//
//	"Whenever a creature you control deals combat damage to a player, add
//	 that much {G}. Until end of turn, you don't lose this mana as steps and
//	 phases end."
//
// One trigger per creature that connects (each deals its own damage, CR
// 510.2), adding as much {G} as that creature dealt. The amount is read
// from the damage event as the trigger goes on the stack, so a creature
// that has left, or damage that was larger than what is left of the
// creature, changes nothing. The mana carries the keep-until-end-of-turn
// mark (Karn, Legacy Reforged's), so it survives the step boundaries until
// the cleanup step.
//
// No simplification.
func init() {
	const label = "Sakiko, Mother of Summer — add that much {G}; you don't lose this mana as steps and phases end"
	t := On(game.EventDealDamage, gapB103CreatureYouControlDealtCombatDamageToAPlayer, label,
		func(g *game.Game, item *game.StackItem) error {
			if item.Params.Amount <= 0 {
				return nil
			}
			return AddMana{Amount: ManaN(item.Params.Amount, "G"), Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()}}.Apply(NewContext(g, item))
		})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = ev.Amount
		return item
	}
	Register(Spec{
		OracleID:     "e4021e30-16bf-4178-9a3b-565bfdce4384",
		Name:         "Sakiko, Mother of Summer",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
