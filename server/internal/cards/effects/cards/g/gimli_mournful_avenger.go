package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gimliMournfulLabel = "Gimli, Mournful Avenger — put a +1/+1 counter on Gimli"
	gimliMournfulDone  = "Gimli, Mournful Avenger — counter ability resolved this turn"
)

// Gimli, Mournful Avenger — Legendary Creature — Dwarf Warrior {1}{R}{G},
// 3/2 (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Gimli has indestructible as long as two or more creatures died under
//	 your control this turn.
//	 Whenever another creature you control dies, put a +1/+1 counter on
//	 Gimli. When this ability resolves for the third time this turn, Gimli
//	 fights up to one target creature you don't control."
//
// The first line is a layer 6 grant on while the turn's creature-death
// tally for Gimli's controller is two or more (Celebration's shape over a
// different tally). The counter trigger counts its own resolutions this
// turn on Gimli's per-object mark (Calix's "done this turn" note, so a
// Gimli that left and came back starts from none); the third one creates
// a CR 603.12 reflexive trigger whose target ("up to one target creature
// you don't control") is chosen as it goes on the stack. A Gimli that has
// left the battlefield puts no counter, counts nothing and fights
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f3941d8-b4d2-4faf-881a-f15973b4368d",
		Name:         "Gimli, Mournful Avenger",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.WhileYouHave(effects.GapB125TwoOrMoreCreaturesDied, effects.SelfCreature), "indestructible"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B39AnotherCreatureYouControlDied(ev, source, g)
			}, gimliMournfulLabel, gimliMournfulCounter),
		},
	})
}

func gimliMournfulCounter(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	if err := effects.PlusOneCountersOnThis(1)(g, item); err != nil {
		return err
	}
	effects.B28NoteDoneThisTurn(g, item.SourceCardID, gimliMournfulDone)
	if g.ResolvedThisTurn(item.SourceCardID, gimliMournfulDone) != 3 {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label: "Gimli, Mournful Avenger — Gimli fights up to one target creature you don't control",
		Cards: []uuid.UUID{item.SourceCardID},
		Body:  gimliMournfulFightBody,
	}.Apply(ctx)
}

var gimliMournfulFightBody = game.ReflexiveBody("gimli-mournful-avenger/fight", effects.SimpleBody(gimliMournfulFight),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("up to one target creature you don't control", effects.Not(effects.YouControl())).WithCount(0, 1)
	}))

func gimliMournfulFight(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	gimli := ctx.PayloadCards()
	if len(gimli) == 0 || !effects.OnBattlefield(g, gimli[0]) {
		return nil
	}
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.B10Fight(ctx, gimli[0], id)
}
