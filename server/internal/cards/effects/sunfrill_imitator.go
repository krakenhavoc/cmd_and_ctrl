package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunfrill Imitator — Creature — Dinosaur {2}{G}, 3/3 (Lost Caverns of
// Ixalan Commander, #5368):
//
//	"Whenever this creature attacks, you may have it become a copy of
//	 another target Dinosaur you control, except its name is Sunfrill
//	 Imitator and it has this ability."
//
// A copy with no stated duration (CR 611.2a): it lasts while the creature
// stays on the battlefield. "Its name is Sunfrill Imitator" and "has this
// ability" are part of the copiable values (CR 707.9a, 707.9b): the name is
// set and the trigger is granted as a named bundle, so the copy attacks
// again as an Imitator and a Clone of it carries the trigger too. The
// trigger is a "may": asked as it resolves. The target is chosen as the
// trigger goes on the stack and checked again as it resolves (CR 608.2b).
//
// No simplification.
const sunfrillImitatorGrant = "sunfrill-imitator/this-ability"

var sunfrillImitatorTrigger = Optional(Targeting(
	WheneverThisAttacks("Sunfrill Imitator — become a copy of another target Dinosaur you control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			targets := ctx.LegalTargets()
			if len(targets) == 0 || targets[0].Kind != game.TargetCard {
				return nil
			}
			return BecomeCopy{
				Targets:  []uuid.UUID{ctx.Source()},
				Of:       targets[0].ID,
				Duration: CopyIndefinite,
				Except: func(v *game.PrintedValues) {
					v.SetName("Sunfrill Imitator")
					v.GrantAbility(sunfrillImitatorGrant)
				},
				Label: "Sunfrill Imitator — becomes a copy",
			}.Apply(ctx)
		}),
	Another(TargetCreature("another target Dinosaur you control", YouControl(), OfSubtype("Dinosaur")))),
	"Sunfrill Imitator — have it become a copy of that Dinosaur?")

func init() {
	Register(Spec{
		OracleID:     "2f618a4a-8b76-4c87-baa7-fb6a17f6f51c",
		Name:         "Sunfrill Imitator",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       sunfrillImitatorGrant,
			Text:      "Whenever this creature attacks, you may have it become a copy of another target Dinosaur you control, except its name is Sunfrill Imitator and it has this ability.",
			Triggered: []game.TriggeredAbility{sunfrillImitatorTrigger},
		}},
		Triggered: []game.TriggeredAbility{sunfrillImitatorTrigger},
	})
}
