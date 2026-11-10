package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Naar Isle — Plane — Wildfire (#4870):
//
//	"At the beginning of your upkeep, put a flame counter on Naar Isle, then
//	 Naar Isle deals damage to you equal to the number of flame counters on it.
//	 Whenever chaos ensues, Naar Isle deals 3 damage to target player or
//	 planeswalker."
//
// The counters sit on the face-up planar card and every flame counter counts,
// whoever's upkeep put it there; a planeswalk away makes a new object with
// none (CR 901.7a). The plane is the damage source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e0bcf65-1eaf-440f-a59b-0b11e1106135",
		Name:         "Naar Isle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Naar Isle — put a flame counter on it, then it deals damage to you equal to the number of flame counters on it",
				func(g *game.Game, item *game.StackItem) error {
					n, ok := g.PutCountersOnFaceUpPlanarCardForEffect(item.SourceCardID, "flame", 1)
					if !ok {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: n}.Apply(NewContext(g, item))
				}),
			TriggerWithPurpose(Targeting(WheneverChaosEnsues("Naar Isle — 3 damage to target player or planeswalker",
				sourceDealsDamageToEachLegalTarget(3)), targetPlayerOrPlaneswalker()), ForTargets(DamageToTarget(0, 3))),
		},
	})
}
