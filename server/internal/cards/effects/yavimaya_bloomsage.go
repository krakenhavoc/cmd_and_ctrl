package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yavimaya Bloomsage // Channel — Creature — Dryad Druid {2}{G}, 2/2 //
// Sorcery {G}{G} (Secrets of Strixhaven Commander, #4885; preparation
// card, CR 722):
//
//	"At the beginning of your end step, put a +1/+1 counter on target
//	 creature you control. Then if that creature has power 7 or greater,
//	 this creature becomes prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Channel — "Until end of turn, any time you could activate a mana
//	 ability, you may pay 1 life. If you do, add {C}."
//
// The target is chosen as the trigger goes on the stack and the trigger
// is removed when you control no creature (CR 603.3d). The power is read
// after the counter lands. A target gone by resolution means the ability
// does not resolve, so nothing becomes prepared. It does not enter
// prepared. The prepare spell is Channel word for word and shares its
// body (channel.go).
//
// No simplification.
const yavimayaBloomsageOracleID = "d9068051-b51c-4700-b492-30b148e3a72c"

func init() {
	counter := On(game.EventBeginEndStep, ByYou,
		"Yavimaya Bloomsage — put a +1/+1 counter on target creature you control",
		yavimayaBloomsageEndStep)
	counter.Targets = TargetCreature("target creature you control", YouControl())
	Register(Spec{
		OracleID:     yavimayaBloomsageOracleID,
		Name:         "Yavimaya Bloomsage",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{counter},
	})
	Register(Spec{
		OracleID:     yavimayaBloomsageOracleID + "#1",
		Name:         "Channel",
		Completeness: CompletenessFull,
		OnResolve:    channelOnResolve,
	})
}

// yavimayaBloomsageEndStep puts the counter on the target; the power
// check is the placement's keyed tail, so it reads the creature after the
// counter has landed even when a CR 616 ordering prompt pauses it.
func yavimayaBloomsageEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		then := ctx.ItemTail(yavimayaBloomsagePrepare)
		then.Carry = []uuid.UUID{t.ID}
		return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, t.ID, game.CounterPlusOne, 1, then)
	}
	return nil
}

// yavimayaBloomsagePrepare is "Then if that creature has power 7 or
// greater, this creature becomes prepared"; Carry is that creature. An
// on-disk key: never renamed, never reused.
var yavimayaBloomsagePrepare = CardTail("yavimaya-bloomsage/prepare", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) < 1 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.Tail.Carry[0])
	if !ok || !c.IsCreature() || c.CurrentPower() < 7 {
		return nil
	}
	return BecomePrepared{Target: ctx.Source()}.Apply(ctx)
})
