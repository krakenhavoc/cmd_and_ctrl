package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malboro — Creature — Plant Horror {4}{B}{B}, 4/4 (Final Fantasy,
// #4462):
//
//	"Bad Breath — When this creature enters, each opponent discards a
//	 card, loses 2 life, and exiles the top three cards of their library.
//	 Swampcycling {2}"
//
// Each opponent chooses their own discard (CR 701.9b), loses 2 life
// (a life change, not damage) and exiles the top three cards of their
// library face up (MillToZone to exile, which is not a mill). "Bad
// Breath" is a flavor word with no rules meaning. Swampcycling is Canyon
// Crawler's typecycling (CR 702.29e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6804e329-800b-4aa4-9a00-b9a1f4b9537b",
		Name:         "Malboro",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Malboro — Bad Breath: each opponent discards a card, loses 2 life, and exiles the top three cards of their library",
				malboroBadBreath),
		},
		Activated: []ActivatedAbility{
			Typecycling("Swampcycling", "{2}", "a Swamp card", IsLandWithSubtype("Swamp")),
		},
	})
}

// malboroBadBreath runs the three instructions for each opponent in turn
// order.
func malboroBadBreath(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: opp, Source: item.SourceCardID, N: 1})
		if err := g.ChangePlayerLifeForEffect(ctx.Source(), opp, -2); err != nil {
			return err
		}
		if err := (MillToZone{Player: opp, N: 3, To: game.ZoneExile}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
