package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thought Shucker — Creature — Rat Rogue {1}{U}, 1/3 (Bloomburrow,
// #4467):
//
//	"Threshold — {1}{U}: Put a +1/+1 counter on this creature and draw a
//	 card. Activate only if there are seven or more cards in your
//	 graveyard and only once."
//
// Threshold is an ability word (CR 207.2c); here it is the activation
// condition GraveyardAtLeast(7) over any cards, as Barbarian Ring's is.
// "Only once" is the OnlyOnce bit (CR 602.5b), once per object, so a
// Shucker that leaves and comes back may activate it again (CR 400.7),
// as Mild-Mannered Librarian's is. A Shucker that left in response gets
// no counter, and you still draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33f4fcd5-4eea-4146-a564-da5511a8ee0a",
		Name:         "Thought Shucker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "Threshold — {1}{U}: Put a +1/+1 counter on this creature and draw a card. Activate only if there are seven or more cards in your graveyard and only once.",
			Purpose:   game.Purpose{Answers: game.AnswerPump, Draws: 1},
			Cost:      ManaCost("{1}{U}"),
			Condition: GraveyardAtLeast(7, nil),
			OnlyOnce:  true,
			Effect:    thoughtShuckerCounterAndDraw,
		}},
	})
}

// thoughtShuckerCounterAndDraw is the ability's effect, in printed order.
func thoughtShuckerCounterAndDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if sourceIsStillThisPermanent(g, item) {
		if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
}
