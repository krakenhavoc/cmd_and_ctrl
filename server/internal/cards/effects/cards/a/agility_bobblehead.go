package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agility Bobblehead — Artifact — Bobblehead {3} (Fallout, #5948):
//
//	"{T}: Add one mana of any color.
//	 {3}, {T}: Up to X target creatures you control each gain haste until
//	 end of turn and can't be blocked this turn except by creatures with
//	 haste, where X is the number of Bobbleheads you control as you
//	 activate this ability."
//
// Coalition Relic's any-color mana ability, and an "up to X" clause whose X
// is read once as the ability is activated (CR 602.2b; TargetSpec.MaxFrom),
// so a Bobblehead that leaves in response does not shrink the picks. Each
// target still legal at resolution gets haste and Gingerbrute's block rule
// until end of turn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83a3897c-c245-4fad-b253-3f8bbe6445fb",
		Name:         "Agility Bobblehead",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{3}, {T}: Up to X target creatures you control each gain haste until end of turn and can't be blocked this turn except by creatures with haste, where X is the number of Bobbleheads you control as you activate this ability.",
			Purpose: game.Purpose{
				Answers: game.AnswerCombatGrant,
				Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.S82GapB60UpToXCreatures(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"}, Label: "Agility Bobblehead — haste"}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.CantBeBlockedThisTurnExceptBy{Target: t.ID, Keywords: []string{"haste"}, Text: "creatures with haste", Label: "Agility Bobblehead"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
