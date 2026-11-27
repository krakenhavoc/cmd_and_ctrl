package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b748_helpers.go — shared pieces of S82 OLD batch 748 (#5627).

// PayOnePerCounterOnThis is the mana component "Pay {1} for each <kind>
// counter on this creature" (Skeleton Scavengers): generic mana equal to the
// counters of that kind on the source as the activation is announced
// (game.AbilityCost.ManaPerCounter), which the activator doesn't choose.
func PayOnePerCounterOnThis(kind string) game.AbilityCost {
	return game.AbilityCost{ManaPerCounter: kind}
}

// CrovaxUpkeepOffer is Crovax the Cursed's upkeep ability: "you may sacrifice
// a creature. If you do, put a +1/+1 counter on Crovax. If you don't, remove a
// +1/+1 counter from Crovax." Zoyowa Lava-Tongue's pay-or-else shape with a
// registered branch for each answer. A controller with no creature to
// sacrifice is not asked (CR 118.3) and "doesn't".
func CrovaxUpkeepOffer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cost, action := crovaxSacrifice.Prompt()
	if !g.PayActionCanPayForEffect(ctx.Controller(), action) {
		return crovaxShrink(ctx)
	}
	return UpkeepPayUnless{
		Chooser:  ctx.Controller(),
		Cost:     cost,
		Action:   action,
		Question: "Crovax the Cursed — sacrifice a creature to put a +1/+1 counter on Crovax? (If you don't, remove one.)",
		Else:     crovaxBranches,
	}.Apply(ctx)
}

var crovaxSacrifice = SacrificePayment(1, "creature", "creatures", game.PermanentQuery{Types: []string{"creature"}})

// crovaxGrow and crovaxShrink are the two answers; each acts only while Crovax
// is still the permanent the trigger came from (CR 400.7).
func crovaxGrow(ctx *Context) error {
	return PutCounterOnSelf(ctx.Game, ctx.Item)
}

func crovaxShrink(ctx *Context) error {
	if !SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return RemoveACounterFromThis(game.CounterPlusOne)(ctx.Game, ctx.Item)
}

// crovaxBranches registers both answers under one key. An on-disk key: never
// renamed, never reused.
var crovaxBranches = PayUnlessBranches("pay-unless/s82-old-b748-crovax-the-cursed",
	func(ctx *Context, _ game.PayUnlessAnswer) error { return crovaxGrow(ctx) },
	func(ctx *Context, _ game.PayUnlessAnswer) error { return crovaxShrink(ctx) })

// S82OldB748Cannibalize is Cannibalize's resolution: the spell's controller
// chooses which of the still-legal target creatures is exiled, and the other
// gets two +1/+1 counters. A lone legal target is exiled with nothing to put
// counters on.
func S82OldB748Cannibalize(_ *game.StackItem, ctx *Context) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	switch len(ids) {
	case 0:
		return nil
	case 1:
		return (ExileTarget{Target: ids[0]}).Apply(ctx)
	}
	_, err := ctx.Game.ChooseCardsRunThenKeyedForEffect([]game.ChooseCardsPrompt{{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Cannibalize — choose the creature to exile; the other gets two +1/+1 counters",
		Cards:    ids,
		Min:      1,
		Max:      1,
		Zone:     game.ZoneBattlefield,
	}}, s82OldB748CannibalizeChosen, ctx.Source(), ctx.RunData([]uuid.UUID{ctx.Controller(), ids[0], ids[1]}))
	return err
}

// s82OldB748CannibalizeChosen exiles the chosen creature and puts the counters
// on the other if it is still on the battlefield. An on-disk key: never
// renamed, never reused.
var s82OldB748CannibalizeChosen = RunThen("prompt-run/s82-old-b748-cannibalize-exile", func(ctx *Context, r game.RunDone) error {
	if len(r.Carry) < 3 {
		return nil
	}
	chooser, pair := r.Carry[0], r.Carry[1:3]
	picked := r.Picks().By(chooser)
	if len(picked) == 0 {
		return nil
	}
	other := pair[0]
	if picked[0] == pair[0] {
		other = pair[1]
	}
	// The counters go on first: the two instructions touch different objects, so
	// the order changes nothing, and the "is the other still here" test then
	// reads the board before nothing but a counter has been put on it.
	if OnBattlefield(ctx.Game, other) {
		if err := (AddCounter{Target: other, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
			return err
		}
	}
	return (ExileTarget{Target: picked[0]}).Apply(ctx)
})
