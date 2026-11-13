package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dimir Cutpurse — Creature — Spirit {1}{U}{B}, 2/2 (Ravnica: City of
// Guilds, #5600):
//
//	"Whenever this creature deals combat damage to a player, that player
//	 discards a card and you draw a card."
//
// Blazing Specter's discard, then the draw as the next step, so it waits
// for the damaged player's pick (CR 608.2c). The player was dealt the
// damage, read off the trigger's event; a player with an empty hand
// discards nothing, and you still draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "56c1a548-c80c-432b-bf92-d4ef6117b298",
		Name:         "Dimir Cutpurse",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Dimir Cutpurse — that player discards a card and you draw a card",
				Do(
					Instruction(func(ctx *Context) error { return discardOne(ctx.Game, ctx.Item, ctx.Item.Trigger.Event.Target) }),
					DrawCards{N: 1},
				)),
		},
	})
}
