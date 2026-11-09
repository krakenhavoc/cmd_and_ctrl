package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brilliant Wings — Enchantment — Aura {1}{W} (Final Fantasy, #4462):
//
//	"Flash
//	 Enchant creature you control
//	 Enchanted creature has flying and hexproof.
//	 Whenever a creature you control enters, you may pay {1}. If you do,
//	 attach this Aura to that creature."
//
// Shielded by Faith's move at Ant-Man's price: the trigger watches a
// creature entering under your control (CR 603.2), the "you may pay {1}"
// is the pay prompt on resolution, and the move is AttachForEffect, which
// is CR 701.3b's whole rule for an attach that cannot happen: an Aura
// that can't legally enchant that creature stays where it is (CR 303.4i),
// and a creature that has left is not there to be enchanted. An Aura that
// left and came back is a new object this ability knows nothing about
// (CR 400.7), so it moves nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ed06b85a-e85f-47c4-ba90-1664e69a338f",
		Name:            "Brilliant Wings",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(YouControl()),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			GrantToAttached("flying", "hexproof"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl,
				"Brilliant Wings — you may pay {1} to attach this Aura to that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || sourceIsNewObject(g, item) {
						return nil
					}
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Brilliant Wings — pay {1} to attach this Aura to that creature?",
						Then:     brilliantWingsPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// brilliantWingsPaid attaches the Aura to the creature that entered. An
// on-disk key: never renamed, never reused.
var brilliantWingsPaid = PayUnlessBranches("pay-unless/brilliant-wings-attach",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		item := ctx.Item
		if item == nil || item.Trigger == nil || sourceIsNewObject(ctx.Game, item) {
			return nil
		}
		return ctx.Game.AttachForEffect(item.SourceCardID, game.TargetRef{Kind: game.TargetCard, ID: item.Trigger.Event.CardID})
	}, nil)
