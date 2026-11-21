package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summon: Fenrir — Enchantment Creature — Saga Wolf {2}{G}, 3/2 (Final
// Fantasy, #4462):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Crescent Fang — Search your library for a basic land card, put
//	 it onto the battlefield tapped, then shuffle.
//	 II — Heavenward Howl — When you next cast a creature spell this turn,
//	 that creature enters with an additional +1/+1 counter on it.
//	 III — Ecliptic Growl — Draw a card if you control the creature with
//	 the greatest power or tied for the greatest power."
//
// A Saga creature. Chapter I is a Rampant Growth chapter. Chapter II is
// Savage Summoning's promise on the next creature spell its controller
// casts this turn (game.NextSpellPromise with a +1/+1 counter), spent as
// that spell is cast and lapsing at end of turn. Chapter III compares the
// power of every creature on the battlefield as the layers now stand; the
// Saga itself is one of them.
//
// No simplification.
func init() {
	const name = "Summon: Fenrir"
	effects.Register(effects.Spec{
		OracleID:     "3509171e-8d7d-4e9e-97a1-58e6c7d225ed",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "Crescent Fang — search for a basic land, put it onto the battlefield tapped"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        name + " — a basic land card",
						Source:        item.SourceCardID,
					}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "Heavenward Howl — your next creature spell this turn enters with an additional +1/+1 counter"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantNextSpellPromise{From: name, Promise: game.NextSpellPromise{
						Filter:      game.PermissionFilter{CreatureOnly: true},
						CounterKind: game.CounterPlusOne,
						Counters:    1,
						Text:        "When you next cast a creature spell this turn, that creature enters with an additional +1/+1 counter on it.",
					}}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "Ecliptic Growl — draw a card if you control the creature with the greatest power"),
				func(g *game.Game, item *game.StackItem) error {
					if !effects.YouControlTheCreatureWithTheGreatestPower(g, item.Controller) {
						return nil
					}
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
