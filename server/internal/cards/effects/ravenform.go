package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ravenform — Sorcery {2}{U} (Lost Caverns of Ixalan Commander):
//
//	"Exile target artifact or creature. Its controller creates a 1/1 blue
//	 Bird creature token with flying.
//	 Foretell {U} (...)"
//
// The controller is read before the exile, as Beast Within does, so the
// Bird goes to whoever controlled the permanent when the spell resolved.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "d3d61a6a-7870-44df-92e5-513dac2c2c3e",
		Name:           "Ravenform",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Foretell("{U}")},
		Targets:        TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := controllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (ExileTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Controller: controller, Template: TokenCard("1/1 blue Bird with flying"), N: 1}.Apply(ctx)
		},
	})
}
