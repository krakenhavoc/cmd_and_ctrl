package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sensational Spider-Man — Legendary Creature — Spider Human Hero
// {1}{W}{U}, 3/3 (Marvel's Spider-Man, #4460):
//
//	"Whenever Sensational Spider-Man attacks, tap target creature
//	 defending player controls and put a stun counter on it. Then you may
//	 remove up to three stun counters from among all permanents. Draw
//	 cards equal to the number of stun counters removed this way."
//
// The target is read off this attack's defending player (Aetherstorm
// Roc's TargetsFrom). The stun counter is placed through the settling form
// of the counter placement, so the removal that follows (which can take
// that very counter back) waits out a CR 616 ordering prompt. "Remove up to
// three stun counters from among all permanents" is up to three rounds of
// one pick each, over every permanent that holds a stun counter on the
// battlefield: the controller names a permanent, one stun counter comes
// off it, and the next round is offered the board as it now stands, so
// two counters can come off one permanent. Declining a round ends the
// removal. The cards drawn are the number actually removed. Should the
// target have become illegal, the removal and the draw still happen
// (CR 608.2b); with no target at all the trigger never goes on the stack.
//
// No simplification.
func init() {
	attack := WheneverThisAttacks("Sensational Spider-Man — tap target creature defending player controls and put a stun counter on it; then remove up to three stun counters from among all permanents and draw that many cards", sensationalAttack)
	attack.TargetsFrom = TargetCreatureDefendingPlayerControls
	Register(Spec{
		OracleID:     "8c489e39-b9ff-400f-a855-3cddf7e112e6",
		Name:         "Sensational Spider-Man",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{attack},
	})
}

// sensationalMaxRemoved is the printed "up to three".
const sensationalMaxRemoved = 3

func sensationalAttack(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return g.AddCounterByKeyedForEffect(ctx.Controller(), t.ID, game.CounterStun, 1, ctx.ItemTail(sensationalAfterStun))
	}
	return sensationalAsk(ctx, 0)
}

// sensationalAfterStun runs once the stun counter has settled (or been
// replaced away): the removal starts with none removed yet.
var sensationalAfterStun = CardTail("tail/sensational-spider-man-after-stun", func(ctx *Context, _ game.TailOutcome) error {
	return sensationalAsk(ctx, 0)
})

// sensationalStunned is every permanent on the battlefield holding a stun
// counter.
func sensationalStunned(g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Counters[game.CounterStun] > 0 {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// sensationalAsk offers the next round, or draws when the three are used
// up or nothing holds a stun counter any more.
func sensationalAsk(ctx *Context, removed int) error {
	stunned := sensationalStunned(ctx.Game)
	if removed >= sensationalMaxRemoved || len(stunned) == 0 {
		return sensationalDraw(ctx, removed)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser: ctx.Controller(),
		Source:  ctx.Item.SourceCardID,
		Question: fmt.Sprintf("Sensational Spider-Man — remove a stun counter from a permanent? (%d of %d removed; none to stop)",
			removed, sensationalMaxRemoved),
		Cards:     stunned,
		Min:       0,
		Max:       1,
		ThenKey:   sensationalPick,
		LiveZone:  game.ZoneBattlefield,
		CarryNums: []int{removed},
		Asked:     game.NewAskedItem(ctx.Item),
	})
	return nil
}

// sensationalPick is one round's answer: no pick stops the removal; a pick
// takes a stun counter off that permanent if it still holds one.
var sensationalPick game.ChooseCardsThen

func init() {
	sensationalPick = askedPick("choose-cards/sensational-spider-man-remove-stun", sensationalPicked)
}

func sensationalPicked(ctx *Context, r game.CardsChosen) error {
	removed := 0
	if len(r.Nums) > 0 {
		removed = r.Nums[0]
	}
	if len(r.Picked) == 0 {
		return sensationalDraw(ctx, removed)
	}
	if c, ok := ctx.Game.LookupCardForEffect(r.Picked[0]); ok && c.Counters[game.CounterStun] > 0 {
		if err := (AddCounter{Target: r.Picked[0], Kind: game.CounterStun, N: -1}).Apply(ctx); err != nil {
			return err
		}
		removed++
	}
	return sensationalAsk(ctx, removed)
}

func sensationalDraw(ctx *Context, removed int) error {
	if removed <= 0 {
		return nil
	}
	return DrawCards{Player: ctx.Controller(), N: removed}.Apply(ctx)
}
