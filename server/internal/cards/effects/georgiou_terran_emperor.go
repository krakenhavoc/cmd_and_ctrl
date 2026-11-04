package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Georgiou, Terran Emperor — Legendary Creature — Human Rogue Noble
// {4}{B}{B}, 4/6:
//
//	"You don't lose the game for having 0 or less life.
//	 At the beginning of your end step, you lose 3 life and draw a card.
//	 Then if your life total is 0 or less, create ten 1/1 red Officer
//	 creature tokens."
//
// Phyrexian Unlife's "can't lose" gate, narrowed to CR 704.5a's 0 or
// less life (ADR 0057): poison and drawing from an empty library still
// end the game, and the next state-based check takes its controller out
// once Georgiou leaves while they are at 0 or less. The end-step trigger
// loses the 3 life, draws, and only then reads the life total; "then if"
// is not an intervening "if", so the trigger always goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31932df4-0be8-45c4-8b22-84119d9c20ba",
		Name:         "Georgiou, Terran Emperor",
		Completeness: CompletenessFull,
		GameEndGates: []game.GameEndGate{{Scope: game.GateYou, CantLose: true, Causes: []game.LossCause{game.LossLife}}},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Georgiou, Terran Emperor — lose 3 life and draw a card, then maybe ten Officers",
				func(g *game.Game, item *game.StackItem) error {
					if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -3); err != nil {
						return err
					}
					ctx := NewContext(g, item)
					if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
						return err
					}
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Life > 0 {
						return nil
					}
					return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 red Officer"), N: 10}.Apply(ctx)
				}),
		},
	})
}
