package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burnout — Instant {1}{R} (Alliances, #5634):
//
//	"Counter target instant spell if it's blue.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Any instant spell is a legal target; whether it is blue is asked as
// Burnout resolves, with "blue" read as Burnout's text now says it
// (CR 612.2), and only a blue one is countered (CR 701.6a). A legal
// target that is not blue still lets Burnout resolve, so the delayed draw
// at the next upkeep at the table is scheduled either way (CR 603.7a,
// next_upkeep_draw.go). A target that has left the stack leaves Burnout
// with no legal target, and it does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90865918-2f9a-4ab6-a881-e86f61a527cc",
		Name:         "Burnout",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target instant spell", Instant()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 {
				target := item.Targets[0].ID
				if card, ok := ctx.Game.LookupCardForEffect(target); ok && card.HasColor(ctx.TextColor("U")) {
					if err := (CounterTarget{StackID: target}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return drawACardAtTheNextUpkeep("Burnout — draw a card").Apply(ctx)
		},
	})
}
