package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bitter Chill — Enchantment — Aura {1}{U} (Wilds of Eldraine, #4471):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 When this Aura is put into a graveyard from the battlefield, you may
//	 pay {1}. If you do, scry 1, then draw a card."
//
// Claustrophobia, whole: the entry trigger taps the creature the Aura
// enchants as it resolves (or last enchanted, CR 608.2h), and the untap
// restriction ends with the Aura. The graveyard trigger is Reach for the
// Sky's (EventLTB fires only from the battlefield, so "landed in a
// graveyard" is the whole condition), and its "you may pay {1}. If you do"
// is Mentor of the Meek's MayPay. The paid branch is Lembas' scry 1 with
// the draw waiting on the scry's answer (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "be0dcad2-8e84-4e9f-b69c-837b40b8fb83",
		Name:                  "Bitter Chill",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bitter Chill — tap enchanted creature", tapEnchantedCreatureOnEntry),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			}, "Bitter Chill — you may pay {1}; if you do, scry 1, then draw a card", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Bitter Chill — pay {1} to scry 1, then draw a card?",
					Then:     bitterChillScryThenDraw,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// bitterChillScryThenDraw is Bitter Chill's paid branch, "scry 1, then
// draw a card". An on-disk key: never renamed, never reused.
var bitterChillScryThenDraw = PayUnlessBranches("pay-unless/bitter-chill-scry-then-draw",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return Scry{Player: ctx.Controller(), N: 1, ThenTail: ctx.TailN(thenDrawCards, 1)}.Apply(ctx)
	}, nil)
