package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gyome, Master Chef — Legendary Creature — Troll Warlock {2}{B}{G}, 5/3
// (Secrets of Strixhaven Commander, #4834):
//
//	"Trample
//	 At the beginning of your end step, create a number of Food tokens
//	 equal to the number of nontoken creatures you had enter the
//	 battlefield under your control this turn.
//	 {1}, Sacrifice a Food: Target creature gains indestructible until
//	 end of turn. Tap it."
//
// The count is the controller's per-turn tally of nontoken creatures that
// entered under their control (PlayerTurnTally.NontokenCreaturesEntered),
// recorded as each entered, so Gyome counts itself on the turn it enters
// and a creature that has since left, stopped being a creature or changed
// hands still counts, as the rulings say. It is read as the trigger
// resolves.
//
// The activation is Underrealm Lich's body on a target: indestructible
// until end of turn, then tap it. It may target a creature that is
// already tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0f1b5cea-2035-444c-9f63-87dcb9782c74",
		Name:            "Gyome, Master Chef",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Gyome, Master Chef — create a Food token for each nontoken creature you had enter this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := g.TurnTallyFor(item.Controller).NontokenCreaturesEntered
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.FoodToken(), N: n}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a Food: Target creature gains indestructible until end of turn. Tap it.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
			Cost:    effects.Plus(effects.ManaCost("{1}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))}),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target := uuid.Nil
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						target = t.ID
						break
					}
				}
				if target == uuid.Nil {
					return nil
				}
				if err := (effects.GrantKeywordUntilEOT{
					Target:   target,
					Keywords: []string{"indestructible"},
					Label:    "Gyome, Master Chef — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.TapTarget{Target: target}.Apply(ctx)
			},
		}},
	})
}
