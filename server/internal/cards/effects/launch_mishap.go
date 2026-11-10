package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Launch Mishap — Instant {2}{U} (Foundations Commander):
//
//	"Counter target creature or planeswalker spell. Create a 1/1
//	 colorless Thopter artifact creature token with flying."
//
// The counter, then the Thopter for the CASTER of Launch Mishap (not the
// countered spell's controller). With its one target gone at resolution
// the whole spell does nothing (CR 608.2b), so no Thopter is made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d4ed768-0a37-4604-8121-43920ed9fb4a",
		Name:         "Launch Mishap",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1, CreatureTokens: 1},
		Targets:      TargetSpell("target creature or planeswalker spell", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}.Apply(ctx)
		},
	})
}
