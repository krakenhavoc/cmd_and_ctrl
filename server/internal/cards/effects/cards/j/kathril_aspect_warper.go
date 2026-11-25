package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kathril, Aspect Warper — {2}{W}{B}{G} Legendary Creature — Nightmare Insect
// 3/3 (Commander 2020):
//
//	"When Kathril enters, put a flying counter on any creature you control if
//	 a creature card in your graveyard has flying. Repeat this process for
//	 first strike, double strike, deathtouch, hexproof, indestructible,
//	 lifelink, menace, reach, trample, and vigilance. Then put a +1/+1
//	 counter on Kathril for each counter put on a creature this way."
//
// The keywords are walked in the printed order. Each one whose creature card
// is in the controller's graveyard (read through the card's own abilities,
// HasKeyword) asks for a creature of theirs as it comes up, and its counter
// is a CR 122.1b keyword counter. The placement is keyed (AddCounterByKeyed)
// so that the count is what the engine settled: a counter that was prevented
// or doubled is counted as it was put. The graveyard is read at each step,
// as the instruction reads, and the +1/+1 counters go on last.
func init() {
	kathrilTail = effects.CardTail("tail/kathril-next-keyword", func(ctx *effects.Context, o game.TailOutcome) error {
		if len(o.Tail.Nums) < 2 {
			return nil
		}
		return kathrilStep(ctx, o.Tail.Nums[0], o.Tail.Nums[1]+o.Amount)
	})
	kathrilPick = effects.AskedPick("choose-cards/kathril-keyword-counter-target", func(ctx *effects.Context, r game.CardsChosen) error {
		if len(r.Nums) < 2 {
			return nil
		}
		next, total := r.Nums[0]+1, r.Nums[1]
		if len(r.Picked) == 0 {
			return kathrilStep(ctx, next, total)
		}
		return kathrilPut(ctx, r.Picked[0], r.Nums[0], total)
	})
	effects.Register(effects.Spec{
		OracleID:     "730897a9-3ead-4990-9492-1a2040d8ccac",
		Name:         "Kathril, Aspect Warper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kathril — put a keyword counter on a creature you control for each keyword among creature cards in your graveyard, then a +1/+1 counter on Kathril for each",
				func(g *game.Game, item *game.StackItem) error {
					return kathrilStep(effects.NewContext(g, item), 0, 0)
				}),
		},
	})
}

var (
	kathrilTail game.TailRef
	kathrilPick game.ChooseCardsThen
)

// kathrilKeywords is the printed order of the walk.
var kathrilKeywords = []string{
	game.CounterFlying, game.CounterFirstStrike, game.CounterDoubleStrike, game.CounterDeathtouch,
	game.CounterHexproof, game.CounterIndestructible, game.CounterLifelink, game.CounterMenace,
	game.CounterReach, game.CounterTrample, game.CounterVigilance,
}

func kathrilGraveyardHas(g *game.Game, player uuid.UUID, kw string) bool {
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return false
	}
	for i := range p.Graveyard.Cards {
		c := p.Graveyard.Cards[i]
		if c.IsCreature() && game.HasKeyword(&c, kw) {
			return true
		}
	}
	return false
}

// kathrilStep runs the walk from keyword index `from`, with `total` counters
// put so far, until it reaches a keyword that needs a placement (which hands
// on through the tail) or the end (the +1/+1 counters).
func kathrilStep(ctx *effects.Context, from, total int) error {
	g := ctx.Game
	for i := from; i < len(kathrilKeywords); i++ {
		if !kathrilGraveyardHas(g, ctx.Controller(), kathrilKeywords[i]) {
			continue
		}
		var cands []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == ctx.Controller() && c.IsCreature() {
				cands = append(cands, c.InstanceID)
			}
		}
		switch len(cands) {
		case 0:
			continue
		case 1:
			return kathrilPut(ctx, cands[0], i, total)
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:   ctx.Controller(),
			Source:    ctx.Source(),
			Question:  "Kathril — put a " + kathrilKeywords[i] + " counter on any creature you control",
			Cards:     cands,
			Min:       1,
			Max:       1,
			ThenKey:   kathrilPick,
			CarryNums: []int{i, total},
			LiveZone:  game.ZoneBattlefield,
			Asked:     game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	if total <= 0 {
		return nil
	}
	return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: total}.Apply(ctx)
}

// kathrilPut puts the counter for keyword index i on the creature and hands
// the settled count to the next step.
func kathrilPut(ctx *effects.Context, creature uuid.UUID, i, total int) error {
	t := ctx.ItemTail(kathrilTail)
	t.Nums = []int{i + 1, total}
	return ctx.Game.AddCounterByKeyedForEffect(ctx.Controller(), creature, kathrilKeywords[i], 1, t)
}
