package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Buy Your Silence — Sorcery {4}{W} (Streets of New Capenna, #5491):
//
//	"Exile target nonland permanent. Its controller creates a Treasure
//	 token."
//
// The controller is read before the exile moves the card (CR 608.2h), as
// Swords to Plowshares reads its life gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2fd01a30-0c0e-4ac5-b96d-b2a6dad31f0c",
		Name:         "Buy Your Silence",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			victim, ok := controllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (ExileTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Controller: victim, Template: TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
