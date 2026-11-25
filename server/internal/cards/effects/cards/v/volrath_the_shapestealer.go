package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volrath, the Shapestealer — Legendary Creature — Phyrexian Shapeshifter
// {2}{B}{G}{U}, 7/5 (Commander 2019):
//
//	"At the beginning of combat on your turn, put a -1/-1 counter on up to one
//	 target creature.
//	 {1}: Until your next turn, Volrath becomes a copy of target creature with
//	 a counter on it, except it's 7/5 and it has this ability."
//
// The counter is a targeted beginning-of-combat trigger with an "up to one"
// clause, so no target is fine. The ability is Mimeoplasm, Revered One's
// copy of a creature with Shapesharer's duration (until your next turn, CR
// 611.2b): the exception sets the base size 7/5 (PrintedValues.SetPowerToughness)
// and grants this very ability as a bundle, so a Volrath that has become another
// creature can copy again. The target is any creature with a counter of any kind
// on it, judged as the ability is put on the stack and as it resolves
// (CR 608.2b).
//
// No simplification.
const volrathGrant = "volrath-the-shapestealer/this-ability"

var volrathCopyAbility = effects.ActivatedAbility{
	Label:   "{1}: Until your next turn, Volrath becomes a copy of target creature with a counter on it, except it's 7/5 and it has this ability.",
	Cost:    effects.ManaCost("{1}"),
	Targets: effects.TargetCreature("target creature with a counter on it", func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return effects.HasCounterOfKind(c, "") }),
	Effect: func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		targets := ctx.LegalTargets()
		if len(targets) == 0 {
			return nil
		}
		return effects.BecomeCopy{
			Targets:  []uuid.UUID{ctx.Source()},
			Of:       targets[0].ID,
			Duration: effects.CopyUntilYourNextTurn,
			Except: func(v *game.PrintedValues) {
				v.SetPowerToughness(7, 5)
				v.GrantAbility(volrathGrant)
			},
			Label: "Volrath, the Shapestealer — becomes a copy until your next turn",
		}.Apply(ctx)
	},
}

func init() {
	counter := effects.AtBeginningOfYourCombat("Volrath, the Shapestealer — put a -1/-1 counter on up to one target creature",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.AddCounter{Target: id, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
		})
	counter.Targets = effects.TargetCreature("up to one target creature").WithCount(0, 1)
	effects.Register(effects.Spec{
		OracleID:     "a70d8e3d-a62d-4c4c-890b-e502e48c4475",
		Name:         "Volrath, the Shapestealer",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       volrathGrant,
			Text:      "{1}: Until your next turn, Volrath becomes a copy of target creature with a counter on it, except it's 7/5 and it has this ability.",
			Activated: []effects.ActivatedAbility{volrathCopyAbility},
		}},
		Triggered: []game.TriggeredAbility{counter},
		Activated: []effects.ActivatedAbility{volrathCopyAbility},
	})
}
