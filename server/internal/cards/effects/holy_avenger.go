package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Holy Avenger — Artifact — Equipment {2}{W} (Adventures in the Forgotten
// Realms Commander):
//
//	"Equipped creature has double strike.
//	 Whenever equipped creature deals combat damage, you may put an Aura card
//	 from your hand onto the battlefield attached to it.
//	 Equip {2}{W}"
//
// The trigger is Banshee's Blade's: one per combat damage step however many
// objects the creature damaged, so a double striker triggers in each step
// (CR 510.4, CR 603.2c). The offer is made as the ability resolves, from the
// hand as it is then, and the Aura goes on attached to the creature the
// trigger names (CR 303.4f). An Aura that can't enchant it stays in hand
// (CR 303.4i); a creature that has left gets no Aura.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86df1de7-967a-4420-847f-7a77d0217a15",
		Name:         "Holy Avenger",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("double strike")},
		Triggered: []game.TriggeredAbility{
			WheneverEquippedCreatureDealsCombatDamage(
				"Holy Avenger — you may put an Aura card from your hand onto the battlefield attached to the creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return s82AfrB08OfferAuraFromHand(g, item.Controller, item.SourceCardID, item.Trigger.Event.Source)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}{W}")},
	})
}
