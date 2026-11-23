package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sonic Screwdriver — Artifact {3} (Doctor Who):
//
//	"{T}: Add one mana of any color.
//	 {1}, {T}: Untap another target artifact.
//	 {2}, {T}: Scry 1.
//	 {3}, {T}: Target creature can't be blocked this turn."
//
// Four abilities sharing one {T}, so only one can be activated a
// turn. The mana ability is a pipe-syntax any-color slot with a
// picker (Birds of Paradise's shape); the three CR 602 abilities are
// Voltaic Key's untap, Sensei's Divining Top-style scry, and Rogue's
// Passage's can't-be-blocked, each with its own mana-plus-tap cost.
//
// "Another" on the untap ability is object identity (effects.Another,
// CR 109.1): Sonic Screwdriver can untap a second Sonic Screwdriver.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd68cc31-12fd-48ff-b37b-bccfe4172974",
		Name:         "Sonic Screwdriver",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: Untap another target artifact.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Targets: effects.Another(effects.TargetPermanent("another target artifact", effects.Artifact())),
				Effect:  effects.UntapFirstLegalTarget,
			},
			{
				Label:   "{2}, {T}: Scry 1.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Scry{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{3}, {T}: Target creature can't be blocked this turn.",
				Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.RestrictUntilEOT{
						Target:       id,
						Restrictions: game.CantBeBlocked,
						Label:        "Sonic Screwdriver — can't be blocked",
					}.Apply(ctx)
				},
			},
		},
	})
}
