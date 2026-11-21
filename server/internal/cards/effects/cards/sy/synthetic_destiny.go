package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Synthetic Destiny — Instant {4}{U}{U}:
//
//	"Exile all creatures you control. At the beginning of the next end
//	 step, reveal cards from the top of your library until you reveal
//	 that many creature cards, put all creature cards revealed this way
//	 onto the battlefield, then shuffle the rest of the revealed cards
//	 into your library."
//
// "That many" is the number of creatures actually exiled, counted when
// the exile finishes and carried to the delayed trigger as its
// Amount. A token is exiled too (and then ceases to exist), so it
// counts. The delayed trigger goes on the stack at the next end step
// (CR 603.7), so every player can respond.
//
// A library with fewer creature cards than asked for reveals itself
// entirely. A zero count reveals nothing and still shuffles.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "640a7d2c-42c4-4ea7-bd5b-72f12a65785e",
		Name:         "Synthetic Destiny",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileAllMatching{Match: effects.And(effects.Creature(), effects.YouControl()), ThenTail: ctx.ItemTail(syntheticDestinySchedule)}.Apply(ctx)
		},
	})
}

var syntheticDestinyBody = game.DelayedBody("synthetic-destiny/reveal-that-many-creatures", syntheticDestinyReveal)

func syntheticDestinyReveal(g *game.Game, item *game.StackItem, params game.EffectParams) error {
	ctx := effects.NewContext(g, item)
	player := item.Controller
	p := ctx.PlayerByID(player)
	if p == nil || p.Library == nil {
		return nil
	}
	var run []uuid.UUID
	found := 0
	// The top of the library is the last element, as revealUntil reads it.
	for i := len(p.Library.Cards) - 1; i >= 0 && found < params.Amount; i-- {
		c := p.Library.Cards[i]
		run = append(run, c.InstanceID)
		if !c.IsToken() && c.IsCreature() {
			found++
		}
	}
	if len(run) > 0 {
		g.RevealForEffect(game.RevealSpec{
			Player: player,
			Source: ctx.Source(),
			Reason: "Synthetic Destiny — reveal until that many creature cards",
			Cards:  run,
		})
	}
	return effects.PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() },
		All:      true,
		ThenTail: effects.ThenShuffle,
	}.Apply(ctx)
}

// syntheticDestinySchedule is the sweep's delayed trigger, told how many
// were exiled, keyed (#4253 slice 10a).
var syntheticDestinySchedule = effects.MassTail("mass/synthetic-destiny-schedule", func(ctx *effects.Context, _ []game.Card, exiled int) error {
	return effects.ScheduleDelayedTrigger{
		Label:  "Synthetic Destiny — reveal until that many creature cards and put them onto the battlefield",
		Body:   syntheticDestinyBody,
		Params: game.EffectParams{Amount: exiled},
	}.Apply(ctx)
})
