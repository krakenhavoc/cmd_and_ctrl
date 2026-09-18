package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grazilaxx, Illithid Scholar — Legendary Creature — Horror {1}{U}{U},
// 3/2 (EDHREC rank 1950):
//
//	"Whenever a creature you control becomes blocked, you may return
//	 it to its owner's hand.
//	 Whenever one or more creatures you control deal combat damage
//	 to a player, draw a card."
//
// The attack-with-ETB-creatures commander. Two triggers:
//
//   - "Becomes blocked" is EventBecomesBlocked (#830), emitted once
//     per blocked attacker when the block declaration is locked in
//     (CR 506.4), and naming the attacker in Target. So a double
//     block is one trigger without any dedupe, and a blocker
//     re-pointed away before the lock-in never blocked this creature
//     at all. The "may" is a real prompt; the bounce reads the
//     attacker off the event and returns it if it is still on the
//     battlefield.
//   - "One or more … deal combat damage" is the Professional
//     Face-Breaker dedup by label: the engine emits one damage event
//     per creature, and the second is declined as a later event of
//     the same batch (OncePerBatch; see AGENTS.md §7).
//
// No simplification.
const b18GrazilaxxDrawLabel = "Grazilaxx, Illithid Scholar — draw a card"

func init() {
	Register(Spec{
		OracleID:     "d22ff377-d282-4a28-9dce-96f25913dc96",
		Name:         "Grazilaxx, Illithid Scholar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBecomesBlocked},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventBecomesBlocked {
						return false
					}
					attacker, ok := g.LookupCardForEffect(ev.Target)
					return ok && attacker.IsCreature() && attacker.Controller == source.Controller
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Grazilaxx — return the blocked creature to its owner's hand?"},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					attacker := ev.Target
					return game.NewTriggeredItem(source, "Grazilaxx, Illithid Scholar — return the blocked creature to hand",
						func(g *game.Game, item *game.StackItem) error {
							if z := g.FindCardZoneForEffect(attacker); z == nil || z.Kind != game.ZoneBattlefield {
								return nil
							}
							return BounceToHand{Target: attacker}.Apply(NewContext(g, item))
						})
				},
			},
			OncePerBatch(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, b18GrazilaxxDrawLabel, Do(DrawCards{N: 1}))),
		},
	})
}
