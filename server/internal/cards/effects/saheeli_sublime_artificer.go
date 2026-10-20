package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saheeli, Sublime Artificer — Legendary Planeswalker — Saheeli
// {1}{U/R}{U/R}, loyalty 5:
//
//	"Whenever you cast a noncreature spell, create a 1/1 colorless Servo
//	 artifact creature token.
//	 −2: Target artifact you control becomes a copy of another target
//	     artifact or creature you control until end of turn, except it's
//	     an artifact in addition to its other types."
//
// The static trigger is the ordinary cast watcher over Noncreature(),
// one Servo per spell.
//
// −2 is two target clauses, the second Distinct from the first
// ("another"), both judged at activation and again at resolution
// (CR 608.2b): with either gone or illegal there is nothing to become or
// nothing to copy, and nothing happens. The copy is BecomeCopy's
// until-end-of-turn duration copy (CR 707.2, 611.2), with the "except"
// clause adding the Artifact card type to the copied values (CR 707.9b),
// so a copied creature stays an artifact.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:        "ab655810-f67e-4576-9174-35b16f5e3235",
		Name:            "Saheeli, Sublime Artificer",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Saheeli, Sublime Artificer — create a 1/1 Servo",
				Do(CreateToken{Template: TokenCard("1/1 colorless Servo artifact"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "−2: Target artifact you control becomes a copy of another target artifact or creature you control until end of turn, except it's an artifact in addition to its other types.",
			Cost:  LoyaltyCost(-2),
			Targets: Clauses(
				TargetPermanent("target artifact you control", Artifact(), YouControl()),
				Distinct(TargetPermanent("another target artifact or creature you control", Or(Artifact(), Creature()), YouControl())),
			),
			Effect: saheeliSublimeArtificerCopy,
		}},
	})
}

// saheeliSublimeArtificerCopy is the −2: the first target becomes a copy
// of the second until end of turn, and is an artifact as well.
func saheeliSublimeArtificerCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	artifact, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	of, ok := ctx.ClauseTarget(1)
	if !ok {
		return nil
	}
	return BecomeCopy{
		Targets: []uuid.UUID{artifact.ID},
		Of:      of.ID,
		Except:  func(v *game.PrintedValues) { v.AddCardType("Artifact") },
		Label:   "Saheeli, Sublime Artificer — becomes a copy until end of turn",
	}.Apply(ctx)
}
