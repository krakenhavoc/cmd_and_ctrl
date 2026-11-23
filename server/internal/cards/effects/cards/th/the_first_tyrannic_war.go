package cardsth

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The First Tyrannic War — Enchantment — Saga {2}{G}{U}{R} (Warhammer 40,000
// Commander):
//
//	I      — You may put a creature card from your hand onto the
//	         battlefield. If its mana cost contains {X}, it enters with a
//	         number of +1/+1 counters on it equal to the number of lands
//	         you control.
//	II, III — Double the number of each kind of counter on target creature
//	         you control.
//
// Chapter I is Eureka Moment's put (a "may", so declining is a real answer),
// with its own registered continuation so the counters are the entry's own
// (CR 614.1c, "enters with"): the picked card is read for {X} in its printed
// mana cost, and the lands are counted as it is put onto the battlefield.
// Chapters II and III double every kind of counter on the target as one
// placement each, reading the counts first (B17DoubleCountersOn); a target
// that left or became illegal does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85fd2f65-957e-446b-990a-8711a1e1c4f6",
		Name:         "The First Tyrannic War",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel("The First Tyrannic War", 1, "you may put a creature card from your hand onto the battlefield"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					candidates := effects.HandCardsMatching(g, item.Controller, effects.Creature())
					if len(candidates) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Destination: game.ChooseOntoBattlefield,
						Chooser:     item.Controller,
						Source:      ctx.Source(),
						Question:    "The First Tyrannic War — you may put a creature card from your hand onto the battlefield",
						Cards:       candidates,
						Min:         0,
						Max:         1,
						ThenKey:     firstTyrannicWarPick,
						LiveZone:    game.ZoneHand,
					})
					return nil
				}),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel("The First Tyrannic War", 2, "double the number of each kind of counter on target creature you control"),
				effects.TargetCreature("target creature you control", effects.YouControl()), firstTyrannicWarDouble),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel("The First Tyrannic War", 3, "double the number of each kind of counter on target creature you control"),
				effects.TargetCreature("target creature you control", effects.YouControl()), firstTyrannicWarDouble),
		},
	})
}

// firstTyrannicWarPick puts the picked creature card onto the battlefield,
// with a +1/+1 counter for each land its controller controls when its
// printed mana cost contains {X}.
var firstTyrannicWarPick = game.RegisterChooseCardsThen("choose-cards/the-first-tyrannic-war-put-creature",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return nil
		}
		opts := game.HandEntryOptions{Controller: r.Chooser}
		if c, ok := g.LookupCardForEffect(r.Picked[0]); ok && strings.Contains(c.ManaCost, "{X}") {
			if lands := effects.CountControlled(g, r.Chooser, func(c game.Card) bool { return c.IsLand() }); lands > 0 {
				opts.EntersWithCounters = map[string]int{game.CounterPlusOne: lands}
			}
		}
		return g.PutFromHandOntoBattlefieldKeyedForEffect(r.Picked[0], opts, game.KeyedTail{})
	})

func firstTyrannicWarDouble(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return effects.B17DoubleCountersOn(ctx, t.ID)
		}
	}
	return nil
}
