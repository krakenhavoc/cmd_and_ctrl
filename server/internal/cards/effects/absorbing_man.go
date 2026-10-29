package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Absorbing Man — Legendary Creature — Human Villain {1}{G}{U}, 4/4
// (Marvel Super Heroes):
//
//	"Vigilance
//	 At the beginning of your first main phase, until your next turn,
//	 Absorbing Man becomes a copy of up to one target artifact, non-Aura
//	 enchantment, or land, except his name is Absorbing Man, he's a
//	 legendary 4/4 Human Villain creature in addition to his other types,
//	 and he has vigilance."
//
// Shapesharer's BecomeCopy with CopyUntilYourNextTurn (CR 611.2b): the copy
// has a timestamp of its own and ends as the controller's next turn
// begins, so the trigger he printed is back before that turn's first main
// phase. The "except" clause edits the copiable values: the name, the
// Legendary supertype, the Creature type and the Human and Villain
// subtypes on top of whatever the copied object is (a Forest stays a
// Forest), the 4/4 body and vigilance. While the copy lasts he has the
// copied object's abilities and none of his own, as the rules say.
//
// The target is "up to one", so a trigger with no target changes nothing.
// A target that left in response gives the copy nothing to take and the
// trigger does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ec8e310f-e871-498f-8d80-c3617c5e1433",
		Name:            "Absorbing Man",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourPrecombatMain("Absorbing Man — becomes a copy of up to one target artifact, non-Aura enchantment, or land until your next turn", absorbingManCopy),
				TargetPermanent("up to one target artifact, non-Aura enchantment, or land",
					Or(Artifact(), And(Enchantment(), Not(Subtype("Aura"))), Land())).WithCount(0, 1)),
		},
	})
}

// absorbingManCopy is the trigger's resolution.
func absorbingManCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return BecomeCopy{
		Targets:  []uuid.UUID{ctx.Source()},
		Of:       targets[0].ID,
		Duration: CopyUntilYourNextTurn,
		Except:   absorbingManExcept,
		Label:    "Absorbing Man — becomes a copy until your next turn",
	}.Apply(ctx)
}

// absorbingManExcept is the "except" clause: his name, a legendary 4/4 Human
// Villain creature in addition to the copied types, and vigilance.
func absorbingManExcept(v *game.PrintedValues) {
	v.SetName("Absorbing Man")
	v.AddSupertype("Legendary")
	v.AddCardType("Creature")
	v.AddSubtype("Human")
	v.AddSubtype("Villain")
	v.Power, v.Toughness = 4, 4
	v.VariableToughness = false
	v.PrintedPTKnown = true
	v.AddKeyword("vigilance")
}
