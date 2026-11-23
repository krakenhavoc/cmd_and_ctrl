package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kenrith, the Returned King — Legendary Creature — Human Noble {4}{W},
// 5/5 (Final Fantasy, #4462):
//
//	"{R}: All creatures gain trample and haste until end of turn.
//	 {1}{G}: Put a +1/+1 counter on target creature.
//	 {2}{W}: Target player gains 5 life.
//	 {3}{U}: Target player draws a card.
//	 {4}{B}: Put target creature card from a graveyard onto the battlefield
//	 under its owner's control."
//
// Five activated abilities, one per colour, each a plain mana cost. The
// first affects only the creatures on the battlefield as it resolves
// (CR 611.2c, the ruling): GrantKeywordUntilEOT's match is read once, then.
// The last targets a creature card in any graveyard and brings it back under
// its OWNER's control (CR 108.3), and it stays if you leave the game.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d209b948-9afb-4fd1-a961-72c87282878c",
		Name:         "Kenrith, the Returned King",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}: All creatures gain trample and haste until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{R}"),
				Effect: effects.Do(effects.GrantKeywordUntilEOT{
					Match:    effects.Creature(),
					Keywords: []string{"trample", "haste"},
					Label:    "Kenrith, the Returned King — trample and haste",
				}),
			},
			{
				Label:   "{1}{G}: Put a +1/+1 counter on target creature.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}{G}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.PlusOneCounterOnTheTarget,
			},
			{
				Label:   "{2}{W}: Target player gains 5 life.",
				Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 5}),
				Cost:    effects.ManaCost("{2}{W}"),
				Targets: effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					p, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return effects.GainLife{Player: p, Amount: 5}.Apply(ctx)
				},
			},
			{
				Label:   "{3}{U}: Target player draws a card.",
				Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1}),
				Cost:    effects.ManaCost("{3}{U}"),
				Targets: effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					p, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return effects.DrawCards{Player: p, N: 1}.Apply(ctx)
				},
			},
			{
				Label:   "{4}{B}: Put target creature card from a graveyard onto the battlefield under its owner's control.",
				Purpose: effects.ForTargets(effects.Reanimates(0)),
				Cost:    effects.ManaCost("{4}{B}"),
				Targets: effects.TargetCreatureInAnyGraveyard(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					effects.ReanimateSingleTarget(effects.NewContext(g, item), uuid.Nil)
					return nil
				},
			},
		},
	})
}
