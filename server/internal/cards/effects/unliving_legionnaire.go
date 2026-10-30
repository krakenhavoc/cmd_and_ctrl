package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unliving Legionnaire — {3}{B} Creature — Vampire Villain 3/2 (Marvel
// Super Heroes, #4455):
//
//	"Flying
//	 Power-up — {5}{B}{B}: Return up to one target creature card from your
//	 graveyard to your hand. Put two +1/+1 counters on this creature.
//	 (Activate each power-up ability only once. Reduce the cost by its mana
//	 cost if it entered this turn.)"
//
// Sword of Light and Shadow's "up to one target creature card from your
// graveyard", then the counters: choosing no target still puts the counters
// on. A target that has left the graveyard returns nothing and the counters
// still go on. Power-up is the engine's (CR 702.193).
//
// No simplification.
func init() {
	ab := PowerUp("{5}{B}{B}: Return up to one target creature card from your graveyard to your hand. Put two +1/+1 counters on this creature.",
		ManaCost("{5}{B}{B}"), unlivingLegionnaireBody)
	ab.Targets = TargetCardInGraveyard("up to one target creature card from your graveyard", YouOwn(), Creature()).WithCount(0, 1)
	ab.Purpose.Answers = game.AnswerValue
	Register(Spec{
		OracleID:        "0edaf433-4ded-4841-bc86-136780c0767f",
		Name:            "Unliving Legionnaire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{ab},
	})
}

// unlivingLegionnaireBody returns the chosen card, then adds the counters.
func unlivingLegionnaireBody(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if id, ok := b16FirstLegalTargetCard(ctx); ok {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return plusOneCountersOnThis(2)(g, item)
}
