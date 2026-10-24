package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gogo, Mysterious Mime — Legendary Creature — Wizard {3}{R}, 2/2
// (deck #3655, seam #3774):
//
//	"At the beginning of combat on your turn, you may have Gogo become
//	 a copy of another target creature you control until end of turn,
//	 except its name is Gogo, Mysterious Mime. If you do, Gogo and that
//	 creature each get +2/+0 and gain haste until end of turn and
//	 attack this turn if able."
//
// Assembled from three existing shapes, with one small addition:
//
//   - The target is chosen as the trigger goes on the stack (CR 603.3d)
//     and the "you may" is asked as it resolves (MayChoice, Kimahri's
//     order) — not before, so declining costs nothing and the target is
//     already public when the choice is made. "Another" is the clause's
//     ExcludeSource, an object comparison, so a token copy of Gogo is
//     still a legal pick. With no other creature you control there is no
//     legal target and the trigger is removed without a prompt.
//   - The copy is BecomeCopy with the default until-end-of-turn
//     duration (CR 707.2, 611.2) and an except clause that sets the
//     name. Gogo is the same object throughout, so a copy is a layer-1
//     record on it and the pump below is a layer-6 / layer-7c record
//     that applies on top whatever order they were made in.
//   - "If you do" is the copy's own report (BecomeCopy.ApplyReports,
//     added for this card). It is false when Gogo left the battlefield
//     or came back as a new object while the trigger waited (CR 400.7),
//     and then nobody gets the +2/+0, the haste or the requirement. The
//     other creature leaving does not stop Gogo's half: each "gets" is
//     its own instruction (CR 608.2c).
//   - Each creature gets ONE record pinned to that object: +2/+0 (layer
//     7c), haste (layer 6) and "attacks this turn if able"
//     (AddAttackRequirementMod, Legion Warboss's shape), all ending at
//     cleanup. The requirement is enforced by the attack validator, and
//     the legal-move enumerator and the bots read the same record, so a
//     bot seat attacks with both or has no legal declaration without
//     them. A creature that cannot attack (tapped, defender, the
//     requirement impossible) is simply not forced (CR 508.1d).
//
// No simplification.
func init() {
	const label = "Gogo, Mysterious Mime — you may have Gogo become a copy of another target creature you control until end of turn"
	Register(Spec{
		OracleID:     "6fdcb563-9279-4365-8a78-26e75fbfd917",
		Name:         "Gogo, Mysterious Mime",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat(label, gogoMimic),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}

const gogoName = "Gogo, Mysterious Mime"

// gogoMimic is the trigger body: ask, copy, and only if the copy began,
// pump and force both creatures.
func gogoMimic(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	other, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return MayChoice{
		Question: "Gogo, Mysterious Mime — have Gogo become a copy of that creature until end of turn?",
		Then:     gogoAnswer,
		Carry:    []uuid.UUID{other},
	}.Apply(ctx)
}

// gogoAnswer is the "yes": Gogo becomes a copy of the creature (the carried
// ID) until end of turn, and only if the copy began, both are pumped and
// forced to attack. Keyed (#4291).
var gogoAnswer = MayYes("confirm/gogo-copy", func(ctx *Context, r game.ConfirmAnswer) error {
	self, other := ctx.Item.SourceCardID, carryAt(r, 0)
	copied, err := BecomeCopy{
		Targets: []uuid.UUID{self},
		Of:      other,
		Except:  func(v *game.PrintedValues) { v.SetName(gogoName) },
		Label:   "Gogo, Mysterious Mime — becomes a copy",
	}.ApplyReports(ctx)
	if err != nil || !copied {
		return err
	}
	// "If you do": Gogo and that creature each.
	for _, id := range []uuid.UUID{self, other} {
		if err := (ScopedEffectFor{
			Target: id,
			Mods: []game.Mod{
				game.ModifyPTMod(2, 0),
				game.AddKeywordsMod("haste"),
				game.AddAttackRequirementMod(uuid.Nil),
			},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "Gogo, Mysterious Mime — +2/+0, haste, attacks this turn if able",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
