package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gateway Sneak — Creature — Vedalken Rogue {2}{U}, 1/3 (Foundations,
// #4465):
//
//	"Whenever a Gate you control enters, this creature can't be blocked
//	 this turn.
//	 Whenever this creature deals combat damage to a player, draw a card."
//
// The first trigger watches every permanent that enters under its
// controller's control and fires for one with the Gate subtype, read off
// the entered permanent (Archway Angel's "Gate"). It makes the Sneak
// itself unblockable until end of turn (Kappa Cannoneer's restriction).
// The second is the common combat-damage draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66c227f2-0e74-43e2-ab24-3866d15c5eef",
		Name:         "Gateway Sneak",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtypeWord("Gate")
			}, "Gateway Sneak — can't be blocked this turn", Do(Instruction(func(ctx *Context) error {
				return RestrictUntilEOT{
					Target:       ctx.Source(),
					Restrictions: game.CantBeBlocked,
					Label:        "Gateway Sneak — can't be blocked this turn",
				}.Apply(ctx)
			}))),
			WheneverThisDealsCombatDamageToAPlayer("Gateway Sneak — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
