package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zero Point Ballad — Sorcery {X}{B} (Edge of Eternities, #4461):
//
//	"Destroy all creatures with toughness X or less. You lose X life. If
//	 X is 6 or more, return a creature card put into a graveyard this way
//	 to the battlefield under your control."
//
// Ceaseless Conflict's sweep: DestroyAllMatching hands its tail the
// pre-move copies of the creatures that were really destroyed, so one that
// was indestructible, regenerated or replaced away is not among them.
// Toughness is judged as the spell resolves. The life is lost first, which
// is unobservable beside the sweep and keeps the return last, as printed.
// With X of 6 or more the caster names one of the creature cards that
// reached a graveyard that way (a token is no card; the card must still be
// in a graveyard, and may be anyone's) and it returns under the caster's
// control (Cemetery Prowler's keyed pick over Reanimate's return).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "835abd35-27ea-45fe-b55d-25081c57bd4a",
		Name:         "Zero Point Ballad",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if err := (GainLife{Player: item.Controller, Amount: -x}).Apply(ctx); err != nil {
				return err
			}
			tail := ctx.ItemTail(zeroPointBalladReturn)
			tail.N = x
			return DestroyAllMatching{Match: And(Creature(), ToughnessLE(x)), ThenTail: tail}.Apply(ctx)
		},
	})
}

// zeroPointBalladReturn is the sweep's tail: with X of 6 or more, the
// caster picks one creature card that went to a graveyard this way. An
// on-disk key.
var zeroPointBalladReturn = MassTailOf("mass/zero-point-ballad-return", func(ctx *Context, swept []game.Card, _ int, t game.KeyedTail) error {
	if t.N < 6 {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range swept {
		if IsToken(c) || !c.IsCreature() {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(c.InstanceID); z != nil && z.Kind == game.ZoneGraveyard {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Zero Point Ballad — return a creature card put into a graveyard this way to the battlefield under your control",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  zeroPointBalladReturned,
		Asked:    t.Asked,
	})
	return nil
})

// zeroPointBalladReturned puts the picked card onto the battlefield under
// the caster's control.
var zeroPointBalladReturned = game.RegisterChooseCardsThen("choose-cards/zero-point-ballad-return", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || r.Asked == nil {
		return nil
	}
	return ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneBattlefield, Controller: r.Chooser}.Apply(NewContext(g, r.Asked))
})
