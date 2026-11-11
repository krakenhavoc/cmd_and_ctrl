package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rot Hulk — Creature — Zombie {5}{B}{B}, 5/5 (Doctor Who Commander):
//
//	"Menace
//	 When this creature enters, return up to X target Zombie cards from your
//	 graveyard to the battlefield, where X is the number of opponents you
//	 have."
//
// X is counted as the trigger is put on the stack: the players still in the
// game other than the Hulk's controller as it entered (the trigger's own
// context, as Worf's clauses read it). Each still-legal pick returns in one
// simultaneous entry (CR 614.12). A Zombie card is a card whose printed type
// line has the subtype, so a Zombie Human counts.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Rot Hulk — return up to X target Zombie cards from your graveyard to the battlefield",
		func(g *game.Game, item *game.StackItem) error {
			return returnLegalGraveyardTargetsToBattlefield(NewContext(g, item))
		})
	enters.TargetsFrom = func(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
		controller := source.Controller
		if obj := tc.Object; obj != nil && obj.ID == source.InstanceID && obj.Controller != uuid.Nil {
			controller = obj.Controller
		}
		n := 0
		for _, p := range g.Seats {
			if p != nil && !p.Eliminated && p.ID != controller {
				n++
			}
		}
		if n == 0 {
			return nil
		}
		return TargetCardInGraveyard("up to X target Zombie cards from your graveyard, where X is the number of opponents you have",
			YouOwn(), Subtype("Zombie")).WithCount(0, n)
	}
	Register(Spec{
		OracleID:        "36117e99-a2bc-46f9-8e17-bd2e2a9c31f4",
		Name:            "Rot Hulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
