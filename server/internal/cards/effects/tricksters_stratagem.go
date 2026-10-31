package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trickster's Stratagem — Sorcery {3}{U} (Marvel Super Heroes, #4455):
//
//	"The owner of target creature an opponent controls puts it into their
//	 library second from the top or on the bottom. Then up to one target
//	 creature you control connives. (Draw a card, then discard a card. If
//	 you discarded a nonland card, put a +1/+1 counter on that creature.)"
//
// Temporal Cleansing's tuck (the creature's OWNER picks second from the
// top or the bottom, on the tuck route so a commander is offered the
// command zone), with the connive as the move's registered tail so it
// follows the owner's choice rather than racing it. The second target is
// "up to one": an unchosen or illegal one connives nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "170ee399-d3cd-40ef-a340-193c0f0f6002",
		Name:         "Trickster's Stratagem",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature an opponent controls", OpponentControls()),
			TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var carry []uuid.UUID
			if mine, ok := ctx.ClauseTarget(1); ok {
				carry = append(carry, mine.ID)
			}
			then := game.KeyedTail{Ref: mshA8StratagemConnive, Player: ctx.Controller(), Source: ctx.Source(), Carry: carry}
			theirs, ok := ctx.ClauseTarget(0)
			if !ok {
				return then.Run(ctx.Game, nil)
			}
			owner, found := ctx.Game.LookupCardForEffect(theirs.ID)
			if !found {
				return then.Run(ctx.Game, nil)
			}
			return PutInLibraryInAnyOrder{
				Chooser:   owner.Owner,
				Cards:     []uuid.UUID{theirs.ID},
				From:      game.ZoneBattlefield,
				Placement: game.LibraryPlaceTopOrBottom,
				TopDepth:  2,
				Label:     "Trickster's Stratagem — put it into your library second from the top or on the bottom",
				ThenTail:  then,
			}.Apply(ctx)
		},
	})
}

// mshA8StratagemConnive is the "then up to one target creature you control
// connives" half of Trickster's Stratagem. An on-disk key: never renamed,
// never reused.
var mshA8StratagemConnive = game.RegisterTail("msh-a8/tricksters-stratagem-connive", func(g *game.Game, o game.TailOutcome) error {
	return conniveNext(g, o.Tail.Player, o.Tail.Carry)
})
