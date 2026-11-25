package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Legion's End — Sorcery {1}{B} (Core Set 2020):
//
//	"Exile target creature an opponent controls with mana value 2 or less and
//	 all other creatures that player controls with the same name as that
//	 creature. Then that player reveals their hand and exiles all cards with
//	 that name from their hand and graveyard."
//
// Declaration in Stone's sweep by name: the target is a real target, the other
// creatures that player controls with its post-layer name are not, and all of
// them leave for exile together (CR 400.7). The tail then reveals that player's
// hand (CR 701.20a) and exiles every card of that name from their hand and
// graveyard as one move. The name and the player are read from the exiled
// permanents as they last existed (CR 608.2h). With the target illegal the
// spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7df8a8aa-1856-4fe6-a543-5de65da4f49f",
		Name:         "Legion's End",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature an opponent controls with mana value 2 or less", effects.OpponentControls(), effects.ManaValueLE(2)),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, err := effects.OldB157ExileTargetAndSameName(ctx, ctx.Tail(legionsEndThen))
			return err
		},
	})
}

// legionsEndThen is "then that player reveals their hand and exiles all cards
// with that name from their hand and graveyard". An on-disk key: never
// renamed, never reused.
var legionsEndThen = effects.MassTail("mass/legions-end-name-sweep", func(ctx *effects.Context, swept []game.Card, _ int) error {
	if len(swept) == 0 {
		return nil
	}
	name, victim := swept[0].Name, swept[0].Controller
	p := ctx.Game.PlayerByIDForEffect(victim)
	if p == nil || p.Hand == nil {
		return nil
	}
	ctx.Game.RevealHandForEffect(victim)
	var ids []uuid.UUID
	for _, z := range []*game.Zone{p.Hand, p.Graveyard} {
		if z == nil {
			continue
		}
		for _, o := range z.Cards {
			if game.HasName(o, name) {
				ids = append(ids, o.InstanceID)
			}
		}
	}
	ctx.Game.ExileCardsForEffect(ids)
	return nil
})
