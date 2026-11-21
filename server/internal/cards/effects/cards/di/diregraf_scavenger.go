package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diregraf Scavenger — Creature — Zombie Bear {3}{B}, 2/3 (Innistrad:
// Crimson Vow):
//
//	"Deathtouch
//	 When this creature enters, exile up to one target card from a
//	 graveyard. If a creature card was exiled this way, each opponent
//	 loses 2 life and you gain 2 life."
//
// The clause waits for the exile and reads the card's type as it was
// before it moved (ExileThenIfItWas, as Scavenging Ooze does), so a card
// the exile did not take, or a noncreature one, drains nobody. With no
// target the trigger still goes on the stack and does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "54731e3f-a84a-4fc9-8c02-2f769133c71c",
		Name:            "Diregraf Scavenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Diregraf Scavenger — exile up to one target card from a graveyard", diregrafScavengerExile),
				effects.TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		},
	})
}

func diregrafScavengerExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.ExileThenIfItWas{Target: t.ID, Was: effects.WasCreatureCard, Keyed: diregrafScavengerIfCreature}.Apply(ctx)
	}
	return nil
}

// diregrafScavengerIfCreature is "each opponent loses 2 life and you gain
// 2 life". An on-disk key: never renamed, never reused.
var diregrafScavengerIfCreature = effects.IfItWasTail("if-it-was/diregraf-scavenger-drain", func(ctx *effects.Context) error {
	if err := (effects.EachOpponentLosesLife{N: 2}).Apply(ctx); err != nil {
		return err
	}
	return effects.GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
}, nil)
