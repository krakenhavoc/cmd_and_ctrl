package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arwen, Mortal Queen — Legendary Creature — Elf Noble {1}{G}{W}, 2/2:
//
//	"Arwen enters with an indestructible counter on her.
//	 {1}, Remove an indestructible counter from Arwen: Another target
//	 creature gains indestructible until end of turn. Put a +1/+1 counter
//	 and a lifelink counter on that creature and a +1/+1 counter and a
//	 lifelink counter on Arwen."
//
// An indestructible counter is a keyword counter (CR 122.1b): while it is
// there Arwen is indestructible, and removing it as the cost (paid as the
// ability is announced, so a response cannot spend it twice) leaves her
// destructible. The ability targets another creature, so it fizzles whole
// if that creature is gone (CR 608.2b); Arwen's half is skipped if she has
// left, or come back as a new object, by the time it resolves.
//
// No simplification. The Hobbit (#4454).
func init() {
	Register(Spec{
		OracleID:     "9519a4ea-c4a0-4127-b06a-aac9ffaf7a32",
		Name:         "Arwen, Mortal Queen",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterIndestructible, 1, "Arwen, Mortal Queen: enters with an indestructible counter on her"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Remove an indestructible counter from Arwen: Another target creature gains indestructible until end of turn. Put a +1/+1 counter and a lifelink counter on that creature and a +1/+1 counter and a lifelink counter on Arwen.",
			Cost:    Plus(ManaCost("{1}"), RemoveCountersFromThis(game.CounterIndestructible, 1)),
			Targets: Another(TargetCreature("another target creature")),
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "indestructible")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Arwen, Mortal Queen"}).Apply(ctx); err != nil {
					return err
				}
				if err := arwenPutCounters(ctx, id); err != nil {
					return err
				}
				if onBattlefield(g, item.SourceCardID) && !sourceIsNewObject(g, item) {
					return arwenPutCounters(ctx, item.SourceCardID)
				}
				return nil
			},
		}},
	})
}

// arwenPutCounters is "a +1/+1 counter and a lifelink counter" on one
// creature.
func arwenPutCounters(ctx *Context, id uuid.UUID) error {
	for _, kind := range []string{game.CounterPlusOne, game.CounterLifelink} {
		if err := (AddCounter{Target: id, Kind: kind, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
