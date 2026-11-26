package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vectis Dominator — Artifact Creature — Human Wizard {1}{W}{B}, 0/2 (Alara
// Reborn, #5586):
//
//	"{T}: Tap target creature unless its controller pays 2 life."
//
// The target's controller is asked as the ability resolves, reading the
// controller then (CR 608.2h): pay 2 life (offered only with 2 or more life,
// CR 119.4) and the creature stays as it is, or decline and it is tapped. A
// target that is no longer legal does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ed63322-f877-4f1c-90bd-e1adb3668d6d",
		Name:         "Vectis Dominator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Tap target creature unless its controller pays 2 life.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  S82OldB502TapUnlessPaysLife("Vectis Dominator", 2),
		}},
	})
}

var s82OldB502TapUnlessLifeAnswer = effects.OptionPickThen("option-pick/tap-unless-pays-life", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index == 1 || len(r.Carry) == 0 {
		return nil
	}
	return effects.TapTarget{Target: r.Carry[0]}.Apply(ctx)
})

// S82OldB502TapUnlessPaysLife is "Tap target creature unless its controller
// pays N life": the controller of the target (as it resolves) chooses between
// being tapped and paying the life.
func S82OldB502TapUnlessPaysLife(name string, life int) effects.Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			c, ok := g.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			options := []game.ChoiceOption{{Label: "Let it be tapped"}}
			if p := g.PlayerByIDForEffect(c.Controller); p != nil && p.Life >= life {
				options = append(options, game.ChoiceOption{Label: "Pay life", LifeCost: life})
			}
			if len(options) == 1 {
				return effects.TapTarget{Target: t.ID}.Apply(ctx)
			}
			return effects.PickOption{
				Player:   c.Controller,
				Question: name + " — pay " + lifeWord(life) + ", or " + c.Name + " becomes tapped?",
				Options:  options,
				ThenKey:  s82OldB502TapUnlessLifeAnswer,
				Carry:    []uuid.UUID{t.ID},
			}.Apply(ctx)
		}
		return nil
	}
}

func lifeWord(n int) string {
	if n == 1 {
		return "1 life"
	}
	return string(rune('0'+n)) + " life"
}
