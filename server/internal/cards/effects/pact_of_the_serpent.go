package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pact of the Serpent — Sorcery {1}{B}{B}:
//
//	"Choose a creature type. Target player draws X cards and loses X life,
//	 where X is the number of creatures they control of the chosen type."
func init() {
	Register(Spec{
		OracleID:     "f275039e-3dba-49b0-92aa-92721b1c7f95",
		Name:         "Pact of the Serpent",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Pact of the Serpent — choose a creature type",
				ctx.ItemTail(pactOfTheSerpentTypeChosen))
			return nil
		},
	})
}

var pactOfTheSerpentTypeChosen = creatureTypeChosen("then/pact-of-the-serpent-type-chosen", func(ctx *Context, t string, _ game.TailOutcome) error {
	for _, target := range ctx.LegalTargets() {
		if target.Kind != game.TargetPlayer {
			continue
		}
		x := cmdr2B03CreaturesYouControlOfType(ctx.Game, target.ID, t)
		if x < 1 {
			return nil
		}
		if err := (DrawCards{Player: target.ID, N: x}).Apply(ctx); err != nil {
			return err
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), target.ID, -x)
	}
	return nil
})
