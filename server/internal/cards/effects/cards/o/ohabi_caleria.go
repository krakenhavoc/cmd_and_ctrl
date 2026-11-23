package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ohabi Caleria — Legendary Creature — Elf Archer {1}{G}{W}, 1/3:
//
//	"Reach
//	 Untap all Archers you control during each other player's untap step.
//	 Whenever an Archer you control deals damage to a creature, you may pay
//	 {2}. If you do, draw a card."
//
// The untap clause is not a trigger: it widens the untap step's
// turn-based action (CR 502.3, 502.4), as Seedborn Muse's does, so it sits
// on Spec.UntapStep over the Archers its controller controls, Ohabi
// included, and it is read through the layers (changeling is every creature
// type, CR 702.73a).
//
// The trigger is one per damage event, combat or not (CR 603.2c): the
// dealer is an Archer its controller controlled as it dealt the damage, and
// the recipient is a creature. The payment is asked as it resolves (CR
// 603.5); if it is made, one card is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e403fadb-537c-41ac-924a-450f0eabe2c4",
		Name:            "Ohabi Caleria",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		UntapStep: []game.UntapStepPermission{
			effects.UntapDuringEachOtherPlayersUntapStep("Ohabi Caleria — untap all Archers you control",
				func(c game.Card) bool { return c.IsCreature() && c.HasSubtypeWord("Archer") }),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || ev.Amount <= 0 {
					return false
				}
				controller, ok := effects.B34DamageSourceController(ev, g)
				if !ok || controller != source.Controller {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				if !ok || !dealer.IsCreature() || !dealer.HasSubtypeWord("Archer") {
					return false
				}
				victim, ok := g.LookupCardForEffect(ev.Target)
				return ok && victim.IsCreature() && effects.B15OnBattlefield(g, ev.Target)
			}, "Ohabi Caleria — you may pay {2}; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Ohabi Caleria — pay {2} to draw a card?",
						Then:     effects.MayPayDrawACard,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
