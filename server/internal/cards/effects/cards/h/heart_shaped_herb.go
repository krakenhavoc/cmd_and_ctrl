package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heart-Shaped Herb — Artifact {4} (Marvel Super Heroes Commander, #4455):
//
//	"If a source an opponent controls would deal damage to you, prevent 1 of
//	 that damage.
//	 {2}, {T}, Sacrifice this artifact: You may sacrifice a creature. If you
//	 do, return that card to the battlefield under its owner's control with
//	 three +1/+1 counters on it and you become the monarch."
//
// The prevention is Ajani Steadfast's emblem shape with the amount taken off
// by one (CR 615): damage of 1 is prevented entirely, larger damage loses 1.
// It reads the source's last-known controller, so a spell's or a dead
// creature's damage counts as the opponent's.
//
// The ability's "you may sacrifice a creature" is a resolution-time optional
// pick (Chain of Vapor's choose-cards shape with a minimum of zero). Once the
// sacrifice has happened ("if you do"), the card returns from the graveyard
// with the counters as its own entry (CR 614.1c) under its owner's control (a
// token has nothing to return, a commander may go to the command zone
// instead, and neither is returned), and you become the monarch whether or
// not a card was there to return.
//
// No simplification.
func init() {
	ab := effects.ActivatedAbility{
		Label:   "{2}, {T}, Sacrifice this artifact: You may sacrifice a creature. If you do, return that card to the battlefield under its owner's control with three +1/+1 counters on it and you become the monarch.",
		Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
		Purpose: game.Purpose{Answers: game.AnswerValue},
		Effect:  heartShapedHerbAsk,
	}
	effects.Register(effects.Spec{
		OracleID:     "20418984-bc6d-42e5-ada6-30b5281925c0",
		Name:         "Heart-Shaped Herb",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{heartShapedHerbPrevention()},
		Activated:    []effects.ActivatedAbility{ab},
	})
}

// heartShapedHerbPrevention is "if a source an opponent controls would deal
// damage to you, prevent 1 of that damage".
func heartShapedHerbPrevention() game.ReplacementEffect {
	return effects.OldB499PreventOneOfTheDamageToYou("Heart-Shaped Herb")
}

// heartShapedHerbAsk is the resolution: the optional sacrifice prompt.
func heartShapedHerbAsk(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && !g.CantBeSacrificedForEffect(c.InstanceID) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     ctx.Source(),
		Question:   "Heart-Shaped Herb — you may sacrifice a creature to return it with three +1/+1 counters and become the monarch",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		ThenKey:    heartShapedHerbPicked,
		LiveZone:   game.ZoneBattlefield,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// heartShapedHerbPicked sacrifices the pick, if there is one. An on-disk key:
// never renamed, never reused.
var heartShapedHerbPicked = game.RegisterChooseCardsThen("choose-cards/heart-shaped-herb-sacrifice", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || r.Asked == nil {
		return nil
	}
	return g.SacrificeAllKeyedForEffect(uuid.Nil, r.Picked, game.KeyedTail{
		Ref: heartShapedHerbSacrificed, Player: r.Chooser, Source: r.Source,
		Carry: []uuid.UUID{r.Picked[0]}, Asked: game.NewAskedItem(r.Asked),
	})
})

// heartShapedHerbSacrificed is the creature gone: its card returns with the
// counters, and the controller becomes the monarch. An on-disk key: never
// renamed, never reused.
var heartShapedHerbSacrificed = game.RegisterTail("then/heart-shaped-herb-return-and-monarch", func(g *game.Game, o game.TailOutcome) error {
	item := o.Tail.Asked.StackItem()
	if item == nil || len(o.Tail.Carry) == 0 {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if err := effects.PutCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, o.Tail.Carry[:1],
		game.ZoneEntryOptions{EntersWithCounters: map[string]int{game.CounterPlusOne: 3}}, game.TailRef{}, 0); err != nil {
		return err
	}
	return effects.BecomeTheMonarch{Player: o.Tail.Player}.Apply(ctx)
})
