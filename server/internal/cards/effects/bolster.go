package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bolster.go — bolster (CR 701.39), #3277.
//
//	Bolster{N: 1}                        "bolster 1"
//	Bolster{N: 2, Then: func(ctx, id)}   "bolster 2. The chosen creature
//	                                     gains trample until end of turn. …"
//
// CR 701.39a: "Bolster N" means "Choose a creature you control with the
// least toughness or tied for least toughness among creatures you
// control. Put N +1/+1 counters on that creature."
//
// The choice is made as the effect resolves and does not target (so
// hexproof and shroud are beside the point). Toughness is read after the
// layers, so an anthem or a -X/-X effect moves the line. Several
// creatures tied for least: their controller picks, through the ordinary
// choose-cards prompt. A single candidate is taken with no prompt, and no
// creature at all does nothing (Then is not run, since there is no
// chosen creature to refer to).
//
// "The chosen creature" is the result of the choice: Then is handed its
// instance ID once the counters have been put on it (the choice stands
// even if the counters are prevented or replaced), so the rest of the
// same resolution can refer to it. Then runs with g.mu held and may queue
// further effects; anything printed after "bolster N" in the same ability
// goes in it, as with PickOption.
type Bolster struct {
	N    int
	Then func(ctx *Context, chosen uuid.UUID) error

	// ThenTail is Then as data (#4253 slice 10b): a registered tail told
	// the chosen creature as its one ID, once the counters are on it. A
	// bolster with no Then asks a keyed pick carrying it (and the asking
	// item, so a CardTail body reads it), so a table waiting on the choice
	// is a restore point. Set Then or ThenTail, not both; Then wins.
	ThenTail game.KeyedTail
}

func (b Bolster) Apply(ctx *Context) error {
	me := ctx.Controller()
	g := ctx.Game
	g.RecomputeLayersIfStaleLocked()
	var least []uuid.UUID
	lowest := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != me || !c.IsCreature() {
			continue
		}
		t := c.CurrentToughness()
		switch {
		case len(least) == 0 || t < lowest:
			lowest, least = t, []uuid.UUID{c.InstanceID}
		case t == lowest:
			least = append(least, c.InstanceID)
		}
	}
	switch len(least) {
	case 0:
		return nil
	case 1:
		return b.finish(ctx, least[0])
	}
	prompt := game.ChooseCardsPrompt{
		Chooser:  me,
		Source:   ctx.Source(),
		Question: "Bolster — choose a creature with the least toughness among creatures you control",
		Cards:    least,
		Min:      1,
		Max:      1,
	}
	if b.Then == nil {
		// Nothing hangs off the choice but data, so the pick is keyed: the
		// counters' number and the tail ride the prompt and the table stays a
		// restore point (#4253, slice 10b).
		prompt.ThenKey = bolsterPick
		prompt.LiveZone = game.ZoneBattlefield
		prompt.CarryNums = []int{b.N}
		prompt.Tail = b.ThenTail
		prompt.Asked = game.NewAskedItem(ctx.Item)
	} else {
		item := ctx.Item
		prompt.Zone = game.ZoneBattlefield
		prompt.Then = func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) != 1 {
				return nil
			}
			return b.finish(NewContext(g, item), picked[0])
		}
	}
	g.QueueChooseCardsForEffect(prompt)
	return nil
}

// bolsterPick is the chosen creature getting its counters, keyed (#4253)
// for a bolster with no "then": the number of counters is the prompt's one
// carried number.
// Since slice 10b the prompt also carries the bolster's tail (CardsChosen.Tail)
// and the asking item.
var bolsterPick = game.RegisterChooseCardsThen("choose-cards/bolster", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 || len(r.Nums) == 0 {
		return nil
	}
	ctx, _ := askedContext(g, r.Asked, r.Chooser, r.Source, nil)
	return Bolster{N: r.Nums[0], ThenTail: r.Tail}.finish(ctx, r.Picked[0])
})

// finish puts the counters on the chosen creature and runs Then. It
// calls AddCounterForEffect directly rather than AddCounter: the chosen
// creature may be the source itself, and AddCounter's "this permanent
// left" guards would misread a bolster that picked it.
func (b Bolster) finish(ctx *Context, chosen uuid.UUID) error {
	if b.N > 0 {
		if err := nothingIfGone(ctx.Game.AddCounterForEffect(chosen, game.CounterPlusOne, b.N)); err != nil {
			return err
		}
	}
	if b.Then == nil {
		return b.ThenTail.Run(ctx.Game, []uuid.UUID{chosen})
	}
	return b.Then(ctx, chosen)
}
