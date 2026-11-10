package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Another Round — Sorcery {X}{X}{2}{W} (Outlaws of Thunder Junction):
//
//	"Exile any number of creatures you control, then return them to the
//	 battlefield under their owner's control. Then repeat this process X more
//	 times."
//
// X+1 rounds. Each round asks for the creatures to exile (any number, none is
// legal; chosen again every round, so the set can differ), exiles them in one
// batch and returns what reached exile together as new objects under their
// owners' control (CR 400.7; a token ceases to exist and does not return,
// CR 111.8). The abilities the rounds trigger wait until the spell has finished
// resolving, as the card's ruling says. Every step of the loop is a keyed
// prompt or tail carrying the rounds left, so a table waiting on a pick is a
// restore point; with no creature to choose there is nothing for the
// remaining rounds to do and the spell ends.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8095ca78-db19-4724-a6ff-eacc85fa2274",
		Name:         "Another Round",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return anotherRoundAsk(ctx, ctx.X())
		},
	})
}

// anotherRoundAsk is one round: the pick of creatures you control, `left` more
// rounds to come after it.
func anotherRoundAsk(ctx *Context, left int) error {
	var creatures []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() {
			creatures = append(creatures, c.InstanceID)
		}
	}
	if len(creatures) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   ctx.Controller(),
		Source:    ctx.Source(),
		Question:  "Another Round — exile any number of creatures you control, then return them",
		Cards:     creatures,
		Min:       0,
		Max:       len(creatures),
		LiveZone:  game.ZoneBattlefield,
		ThenKey:   anotherRoundPicked,
		CarryNums: []int{left},
		Asked:     game.NewAskedItem(ctx.Item),
	})
	return nil
}

// anotherRoundNext starts the next round, if one is left.
func anotherRoundNext(ctx *Context, left int) error {
	if left <= 0 {
		return nil
	}
	return anotherRoundAsk(ctx, left-1)
}

// anotherRoundPicked is the creatures chosen: exiled together, then returned
// by anotherRoundReturn. An on-disk key: never renamed, never reused.
var anotherRoundPicked game.ChooseCardsThen

func anotherRoundPickedBody(ctx *Context, r game.CardsChosen) error {
	left := 0
	if len(r.Nums) > 0 {
		left = r.Nums[0]
	}
	if len(r.Picked) == 0 {
		return anotherRoundNext(ctx, left)
	}
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked, withN(ctx.ItemTail(anotherRoundReturn), left))
}

// anotherRoundReturn is the exile settled: what reached exile comes back
// together under its owners' control, then the next round. An on-disk key.
var anotherRoundReturn game.TailRef

func anotherRoundReturnBody(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return anotherRoundNext(ctx, o.Tail.N)
	}
	return ReturnFromExileTogether{Targets: o.IDs, ThenTail: anotherRoundAfter, ThenN: o.Tail.N}.Apply(ctx)
}

// anotherRoundAfter is the creatures back on the battlefield: the next round.
// An on-disk key.
var anotherRoundAfter game.TailRef

func init() {
	anotherRoundPicked = askedPick("choose-cards/another-round-picked", anotherRoundPickedBody)
	anotherRoundReturn = CardTail("then/another-round-return", anotherRoundReturnBody)
	anotherRoundAfter = TogetherTail("another-round/after-return", func(ctx *Context, _ []uuid.UUID, n int) error {
		return anotherRoundNext(ctx, n)
	})
}
