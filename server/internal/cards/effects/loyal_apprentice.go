package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loyal Apprentice — Creature — Human Artificer {1}{R}, 2/1:
//
//	"Haste
//	 Lieutenant — At the beginning of combat on your turn, if you
//	 control your commander, create a 1/1 colorless Thopter artifact
//	 creature token with flying. That token gains haste until end of
//	 turn."
//
// "Lieutenant" is an ability word, so the condition is an ordinary
// intervening-if (CR 603.4): b18ControlsYourCommander is the shared
// lieutenant read (a commander you OWN and control — a stolen
// opponent's commander does not count), checked when the combat step
// begins and again when the trigger resolves. The haste is scoped to
// the one token for the one turn, Legion Warboss's shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7f268b15-ac92-4e98-821b-78d15d9285d9",
		Name:            "Loyal Apprentice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) &&
					b18ControlsYourCommander(g, source.Controller)
			}, "Loyal Apprentice — create a 1/1 Thopter with flying that gains haste",
				loyalApprenticeThopter),
		},
	})
}

// loyalApprenticeThopter re-checks the lieutenant condition, creates
// the Thopter and gives that one token haste until end of turn.
func loyalApprenticeThopter(g *game.Game, item *game.StackItem) error {
	if !b18ControlsYourCommander(g, item.Controller) {
		return nil
	}
	ctx := NewContext(g, item)
	cursor := b25LastEventSeq(g)
	if err := (CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}).Apply(ctx); err != nil {
		return err
	}
	for _, id := range b27TokensCreatedByAfter(g, item.Controller, cursor) {
		return GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"haste"},
			Label:    "Loyal Apprentice — the Thopter gains haste",
		}.Apply(ctx)
	}
	return nil
}
