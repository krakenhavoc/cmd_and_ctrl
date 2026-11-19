package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agent Frank Horrigan — Legendary Creature — Mutant Warrior {5}{B}{G},
// 8/6 (Fallout, #5948):
//
//	"Trample
//	 Agent Frank Horrigan has indestructible as long as it attacked
//	 this turn.
//	 Whenever Agent Frank Horrigan enters or attacks, proliferate
//	 twice."
//
// The indestructible is a layer 6 self-grant read from the turn's
// attack records (Horrigan's own AttackedThisTurn; DependsOnEvents wakes
// the layer cache as the attack is announced), the shape of War
// Historian's. "Proliferate twice" is two proliferates in a row, the
// second asked once the first is answered (Contagion Engine).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aedcde3b-3ed5-443c-a81c-03d785153e55",
		Name:            "Agent Frank Horrigan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{agentFrankHorriganIndestructible()},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Agent Frank Horrigan — proliferate twice", func(g *game.Game, item *game.StackItem) error {
				return proliferateTimesKeyed(g, item.Controller, item.SourceCardID, 2)
			}),
		},
	})
}

func agentFrankHorriganIndestructible() game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfCreature(target, g, source) && g.AttackedThisTurn(source.InstanceID)
	}, "indestructible")
	ab.Label = "Has indestructible as long as it attacked this turn."
	ab.DependsOnEvents = []game.EventKind{game.EventAttack}
	return ab
}
