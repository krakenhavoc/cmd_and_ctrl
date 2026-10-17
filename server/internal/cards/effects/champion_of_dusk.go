package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Champion of Dusk — Creature — Vampire Knight {3}{B}{B}, 4/4:
//
//	"When this creature enters, you draw X cards and you lose X life,
//	 where X is the number of Vampires you control."
//
// X is counted once, as the trigger resolves, and the Champion counts
// itself: it is a Vampire on the battlefield by then, so a lone Champion
// draws one card and loses one life. The count is every PERMANENT of the
// type, not only creatures ("Vampires you control", not "Vampire
// creatures"), and the word is read through the source's text (CR 612.2,
// #3984) so a Champion under a text-changing effect counts what its text
// now says. The cards come first and the life second, in the printed
// order, from the one X.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f7f0a38-fb7e-4ee8-a712-9aa911f17dae",
		Name:         "Champion of Dusk",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Champion of Dusk — draw a card and lose 1 life for each Vampire you control", championOfDuskDrawAndLose),
		},
	})
}

// championOfDuskDrawAndLose is the printed body. Caller holds g.mu.
func championOfDuskDrawAndLose(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	vampire := g.TextReaderForEffect(item).TextSubtype("Vampire")
	x := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.HasSubtype(vampire) {
			x++
		}
	}
	if x <= 0 {
		return nil
	}
	if err := (DrawCards{Player: item.Controller, N: x}).Apply(ctx); err != nil {
		return err
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -x)
}
