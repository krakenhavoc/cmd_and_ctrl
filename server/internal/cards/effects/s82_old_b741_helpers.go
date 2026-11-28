package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b741_helpers.go — Exodus counters (S82 OLD batch 741, #5626).

// S82OldB741MoveAllPlusOneOntoThis is Spike Cannibal's enter trigger: "move
// all +1/+1 counters from all creatures onto it". Every other creature loses
// its +1/+1 counters and the source gains the total. The source's own
// counters stay where they are. Nothing moves if the source has left the
// battlefield or come back as a new object (CR 400.7).
func S82OldB741MoveAllPlusOneOntoThis(g *game.Game, item *game.StackItem) error {
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	type held struct {
		id uuid.UUID
		n  int
	}
	var from []held
	total := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == item.SourceCardID || !c.IsCreature() {
			continue
		}
		if n := c.Counters[game.CounterPlusOne]; n > 0 {
			from = append(from, held{c.InstanceID, n})
		}
	}
	for _, h := range from {
		if err := g.AddCounterForEffect(h.id, game.CounterPlusOne, -h.n); err != nil {
			return NothingIfGone(err)
		}
		total += h.n
	}
	if total == 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: total}.Apply(NewContext(g, item))
}

// S82OldB741MindMaggotsDiscard is Mind Maggots' enter trigger: "discard any
// number of creature cards. For each card discarded this way, put two +1/+1
// counters on this creature." The controller picks any number of the creature
// cards in their hand; none in hand asks nothing.
func S82OldB741MindMaggotsDiscard(g *game.Game, item *game.StackItem) error {
	var creatures []uuid.UUID
	if p := NewContext(g, item).PlayerByID(item.Controller); p != nil && p.Hand != nil {
		for _, c := range p.Hand.Cards {
			if c.IsCreature() {
				creatures = append(creatures, c.InstanceID)
			}
		}
	}
	if len(creatures) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Mind Maggots — discard any number of creature cards",
		Cards:    creatures,
		Min:      0,
		Max:      len(creatures),
		Zone:     game.ZoneHand,
		LiveZone: game.ZoneHand,
		ThenKey:  s82OldB741MindMaggotsPicked,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// s82OldB741MindMaggotsPicked discards the chosen creature cards. An on-disk
// key: never renamed, never reused.
var s82OldB741MindMaggotsPicked = AskedPick("choose-cards/mind-maggots-discard", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ctx.Game.DiscardCardsKeyedForEffect(r.Chooser, r.Picked, r.Source, ctx.ItemTail(s82OldB741MindMaggotsDiscarded))
})

// s82OldB741MindMaggotsDiscarded puts two +1/+1 counters on the Maggots for
// each card that really was discarded, if it is still the same permanent. An
// on-disk key.
var s82OldB741MindMaggotsDiscarded = CardTail("then/mind-maggots-discarded", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || !SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: 2 * len(o.IDs)}.Apply(ctx)
})
