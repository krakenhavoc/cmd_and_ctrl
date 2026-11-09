package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "d209b948-9afb-4fd1-a961-72c87282878c",
		Name:         "Kenrith, the Returned King",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{R}: All creatures gain trample and haste until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{R}"),
				Effect: Do(GrantKeywordUntilEOT{
					Match:    Creature(),
					Keywords: []string{"trample", "haste"},
					Label:    "Kenrith, the Returned King — trample and haste",
				}),
			},
			{
				Label:   "{1}{G}: Put a +1/+1 counter on target creature.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{1}{G}"),
				Targets: TargetCreature("target creature"),
				Effect:  PlusOneCounterOnTheTarget,
			},
			{
				Label:   "{2}{W}: Target player gains 5 life.",
				Purpose: ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 5}),
				Cost:    ManaCost("{2}{W}"),
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p, ok := firstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return GainLife{Player: p, Amount: 5}.Apply(ctx)
				},
			},
			{
				Label:   "{3}{U}: Target player draws a card.",
				Purpose: ForTargets(game.TargetPurpose{Slot: 0, Draws: 1}),
				Cost:    ManaCost("{3}{U}"),
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p, ok := firstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					return DrawCards{Player: p, N: 1}.Apply(ctx)
				},
			},
			{
				Label:   "{4}{B}: Put target creature card from a graveyard onto the battlefield under its owner's control.",
				Purpose: ForTargets(Reanimates(0)),
				Cost:    ManaCost("{4}{B}"),
				Targets: targetCreatureInAnyGraveyard(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					reanimateSingleTarget(NewContext(g, item), uuid.Nil)
					return nil
				},
			},
		},
	})
}
