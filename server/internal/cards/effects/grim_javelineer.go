package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Javelineer — Creature — Human Warrior {2}{B}, 3/2 (Aetherdrift,
// #4464):
//
//	"Whenever you attack, target attacking creature gets +1/+0 until end
//	 of turn. When that creature dies this turn, surveil 1."
//
// Boggart Prankster's trigger (once per attack declaration of yours, aimed
// at any attacking creature as it goes on the stack, CR 603.3d) with the
// pump at +1/+0, then a delayed triggered ability the resolution creates
// about that one object (CR 603.7a, 603.7c): a creature that left and came
// back is a new object and surveils nothing. The delayed trigger's
// controller surveils (CR 603.7d). A target gone by resolution gets no
// pump and no delayed trigger (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd31487e-e95a-4d90-8c85-742aa83c3e90",
		Name:         "Grim Javelineer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Grim Javelineer — target attacking creature gets +1/+0 until end of turn; when it dies this turn, surveil 1",
					grimJavelineerPump),
				TargetCreature("target attacking creature", AttackingCreature()))),
		},
	})
}

// grimJavelineerSurveilBody is the delayed trigger's body: its controller
// surveils 1. An on-disk key.
var grimJavelineerSurveilBody = game.SimpleDelayedBody("grim-javelineer/surveil-1", func(g *game.Game, item *game.StackItem) error {
	return Surveil{N: 1}.Apply(NewContext(g, item))
})

// grimJavelineerPump gives the still-legal attacking creature +1/+0 until
// end of turn and sets the death watch on it.
func grimJavelineerPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (BoostUntilEOT{Target: id, Power: 1, Label: "Grim Javelineer — +1/+0"}).Apply(ctx); err != nil {
		return err
	}
	return whenThatCreatureDiesThisTurn(ctx, id, "Grim Javelineer — when it dies this turn, surveil 1", grimJavelineerSurveilBody)
}
