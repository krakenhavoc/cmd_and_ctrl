package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bitter Chill — Enchantment — Aura {1}{U} (Wilds of Eldraine, #4471):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 When this Aura is put into a graveyard from the battlefield, you may
//	 pay {1}. If you do, scry 1, then draw a card."
//
// Dramatic Accusation's tap and untap lock. The last ability is Chromatic
// Star's leaves-for-the-graveyard trigger, fired however the Aura left (the
// enchanted creature dying included). The "may pay {1}" is asked as it
// resolves (CR 603.5, 118.12); a yes is Preordain's scry 1 finished before
// the draw, so the card left on top is the one drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "be0dcad2-8e84-4e9f-b69c-837b40b8fb83",
		Name:                  "Bitter Chill",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm, Scry: 1, Draws: 1},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bitter Chill — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
			effects.WhenThisDies("Bitter Chill — you may pay {1}; if you do, scry 1, then draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Bitter Chill — pay {1} to scry 1, then draw a card?",
						Then:     bitterChillPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// bitterChillPaid is the paid branch: scry 1, then draw. An on-disk key:
// never renamed, never reused.
var bitterChillPaid = effects.PayUnlessBranches("pay-unless/bitter-chill-scry-draw",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		controller := ctx.Controller()
		return effects.Scry{
			Player:   controller,
			N:        1,
			ThenTail: game.KeyedTail{Ref: effects.ThenDrawACard, Player: controller},
		}.Apply(ctx)
	}, nil)
