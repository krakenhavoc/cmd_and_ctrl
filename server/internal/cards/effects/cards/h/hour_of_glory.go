package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hour of Glory — {3}{B} Instant (Hour of Devastation):
//
//	"Exile target creature. If that creature was a God, its controller reveals
//	 their hand and exiles all cards from it with the same name as that
//	 creature."
//
// The exile reports what it moved; the clause after it reads the creature as it
// was (CR 608.2h): a God's controller reveals their hand and every card in it
// with the creature's name is exiled in one go.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d97fd43b-146f-4d13-9c45-51bceaff742e",
		Name:         "Hour of Glory",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.ExileAllMatching{
				Match:    func(_ *game.Game, _ uuid.UUID, o game.Card) bool { return o.InstanceID == target },
				ThenTail: ctx.Tail(hourOfGloryThen),
			}.Apply(ctx)
		},
	})
}

var hourOfGloryThen = effects.MassTail("mass/hour-of-glory-god-hand", func(ctx *effects.Context, swept []game.Card, _ int) error {
	if len(swept) == 0 || !swept[0].HasSubtypeWord("God") {
		return nil
	}
	name, victim := swept[0].Name, swept[0].Controller
	p := ctx.Game.PlayerByIDForEffect(victim)
	if p == nil || p.Hand == nil {
		return nil
	}
	ctx.Game.RevealHandForEffect(victim)
	var ids []uuid.UUID
	for _, o := range p.Hand.Cards {
		if game.HasName(o, name) {
			ids = append(ids, o.InstanceID)
		}
	}
	ctx.Game.ExileCardsForEffect(ids)
	return nil
})
