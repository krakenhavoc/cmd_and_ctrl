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
//   - "Becomes blocked" is EventBlock (S31), which names the blocker
//     in CardID and the attacker in Target, and is emitted once per
//     BLOCKER — so a double block would fire the printed ability
//     twice. CR 509.1h says once per attacker, so the second and
//     later blocks of the same attacker are declined by walking the
//     log back to the attacker's own EventAttack
//     (b18AttackerAlreadyBlocked). The "may" is a real prompt; the
//     bounce reads the attacker off the event and returns it if it
//     is still on the battlefield.
//   - "One or more … deal combat damage" is the Professional
//     Face-Breaker dedup by label: the engine emits one damage event
//     per creature, and the second is declined while the first
//     trigger is pending or on the stack.
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
				Watches: []game.EventKind{game.EventBlock},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventBlock {
						return false
					}
					attacker, ok := g.LookupCardForEffect(ev.Target)
					if !ok || !attacker.IsCreature() || attacker.Controller != source.Controller {
						return false
					}
					return !b18AttackerAlreadyBlocked(g, ev)
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
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return combatDamageToPlayerBy(ev, source.Controller, g) &&
						!b12TriggerPendingOrOnStack(g, source, b18GrazilaxxDrawLabel)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, b18GrazilaxxDrawLabel,
						func(g *game.Game, item *game.StackItem) error {
							return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
