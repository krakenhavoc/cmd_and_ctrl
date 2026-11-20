package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skull Storm — Sorcery {7}{B}{B} (Commander 2018, #5526):
//
//	"When you cast this spell, copy it for each time you've cast your
//	 commander from the command zone this game.
//	 Each opponent sacrifices a creature of their choice. Each opponent who
//	 can't loses half their life, rounded up."
//
// Echo Storm's cast trigger (echoStormCopies) over Momentum Breaker's edict:
// each opponent picks the creature, and the keyed continuation runs once
// every asked opponent has answered. An opponent who was never asked (no
// creature to sacrifice) is exactly "an opponent who can't", and loses half
// their life rounded up, as Betor's last clause reads it; an opponent who
// sacrificed loses nothing. Each copy is its own edict and its own check.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f34d70e-7a3a-43f7-b2c9-f89519804d3c",
		Name:         "Skull Storm",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Skull Storm — copy it for each time you've cast your commander from the command zone", echoStormCopies),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return EachPlayerSacrifices{
				ExceptController: true,
				Match:            Creature(),
				Label:            "a creature",
				Keyed:            skullStormEdicted,
			}.Apply(ctx)
		},
	})
}

// skullStormEdicted is the life loss for each opponent who could not
// sacrifice, keyed (#4290). An on-disk key: never renamed, never reused.
var skullStormEdicted = RegisterEdictThen("prompt-run/skull-storm-half-life", func(ctx *Context, sacrificed game.PromptedSacrifices) error {
	var cant []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if !sacrificed.Sacrificed(opp) {
			cant = append(cant, opp)
		}
	}
	for _, opp := range cant {
		p := ctx.PlayerByID(opp)
		if p == nil || p.Life <= 0 {
			continue
		}
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), opp, -((p.Life + 1) / 2)); err != nil {
			return err
		}
	}
	return nil
})
