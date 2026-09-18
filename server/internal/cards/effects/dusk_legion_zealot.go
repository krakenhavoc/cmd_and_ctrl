package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dusk Legion Zealot — 1/1 Creature — Vampire Soldier for {1}{B}
// (EDHREC rank 4434):
//
//	"When this creature enters, you draw a card and you lose 1
//	 life."
//
// Elvish Visionary's black cousin, and the reason it is played is the
// body rather than the card: a two-mana 1/1 that replaces itself is
// sacrifice fodder that costs nothing, which is what an aristocrats
// deck is short of. Roadmap batch 42 (#449), "no new machinery".
//
// "When this creature enters" is a TRIGGER, not an as-enters clause:
// it uses the stack, every player gets priority, and it can be
// responded to. That is the #578 distinction and it is why this is a
// Triggered entry rather than an AsEnters hook.
//
// Draw first, then lose, as printed. It matters at 1 life: you draw
// the card and then die to the loss, rather than dying before you
// draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea2dce18-195e-4681-a687-f9819edaf9fc",
		Name:         "Dusk Legion Zealot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dusk Legion Zealot — draw a card, lose 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Player: item.Controller, Amount: -1}.Apply(ctx)
				}),
		},
	})
}
