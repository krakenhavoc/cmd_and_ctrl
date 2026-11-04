package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const amazingAllianceLabel = "Amazing Alliance — you gain 1 life for each legendary creature you attacked with"

// Amazing Alliance — Enchantment {1}{W}{W} (Marvel's Spider-Man Eternal):
//
//	"Creatures you control get +1/+1.
//	 Whenever you attack with one or more legendary creatures, you gain
//	 that much life."
//
// The anthem is a layer 7c modification over creatures you control, the
// same grant as Glorious Anthem, live for creatures that arrive later.
//
// "That much" has no amount in the printed text; the owner's reading
// (2026-11-02) is one life for each legendary creature you attacked
// with in that declaration. The engine emits one EventAttack per
// attacker, so the ability waits for the declaration to finish
// (AtBatchEnd, ThatManyEvents) and gains the number of attack events its
// AppliesTo accepted: two legendary attackers are one trigger and 2
// life, a non-legendary attacker adds nothing, and an attack with no
// legendary creature does not trigger it. "Legendary" is read through the
// layered view (attackedWithALegendaryCreature), so a creature that is
// legendary only because of an effect counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a824aff-52fe-4402-b074-884afb5286d4",
		Name:         "Amazing Alliance",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "Creatures you control get +1/+1.",
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			withAtBatchEnd(On(game.EventAttack, attackedWithALegendaryCreature, amazingAllianceLabel,
				func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: settledEventAmount(item)}.Apply(NewContext(g, item))
				}), ThatManyEvents()),
		},
	})
}
