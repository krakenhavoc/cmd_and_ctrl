package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "0f1b5cea-2035-444c-9f63-87dcb9782c74",
		Name:            "Gyome, Master Chef",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Gyome, Master Chef — create a Food token for each nontoken creature you had enter this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := g.TurnTallyFor(item.Controller).NontokenCreaturesEntered
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: FoodToken(), N: n}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice a Food: Target creature gains indestructible until end of turn. Tap it.",
			Cost:    Plus(ManaCost("{1}"), game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))}),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
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
				if err := (GrantKeywordUntilEOT{
					Target:   target,
					Keywords: []string{"indestructible"},
					Label:    "Gyome, Master Chef — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
				return TapTarget{Target: target}.Apply(ctx)
			},
		}},
	})
}
