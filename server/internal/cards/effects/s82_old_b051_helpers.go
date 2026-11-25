package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b051_helpers.go — shared seams of Strixhaven: School of Mages,
// batch 51 (#5502).

// oldB51CopyIfPaid is the paid branch of "you may pay {2}. If you do, copy
// that spell. You may choose new targets for the copy.": the spell rides the
// prompt as its Carry, and is copied from last-known information if it has
// left the stack by now (CR 707.10, CR 608.2h). An on-disk key: never renamed,
// never reused.
var oldB51CopyIfPaid = PayUnlessBranches("pay-unless/old-b051-copy-that-spell",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return CopySpell{
			StackID:          r.Carry[0],
			Controller:       ctx.Controller(),
			FromLastKnown:    true,
			ChooseNewTargets: true,
		}.Apply(ctx)
	}, nil)

// MayPayTwoToCopyThatSpell is the effect of a cast trigger that reads "you may
// pay {2}. If you do, copy that spell. You may choose new targets for the
// copy." (Rowan, Scholar of Sparks' emblem, Reflective Golem). The spell is
// held so the answer can arrive after it resolves (a pay-unless does not stop
// the table, #4276).
func MayPayTwoToCopyThatSpell(question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		spellID := item.Trigger.Event.CardID
		g.HoldSpellForEffect(spellID)
		return MayPay{
			Chooser:  item.Controller,
			Cost:     "{2}",
			Question: question,
			Then:     oldB51CopyIfPaid,
			Carry:    []uuid.UUID{spellID},
		}.Apply(NewContext(g, item))
	}
}

// EntersTappedSacrificeUnlessPaySpec is the whole spec of a land that "enters
// tapped. When this land enters, sacrifice it unless you pay {1}. {T}: Add one
// mana of any color." (Rupture Spire, Archway Commons): the sacrifice is asked
// as the trigger resolves and only if the land is still the object that
// entered (CR 400.7).
func EntersTappedSacrificeUnlessPaySpec(oracleID, name string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(name+" — sacrifice it unless you pay {1}",
				func(g *game.Game, item *game.StackItem) error {
					if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
						return nil
					}
					return UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: name + " — pay {1} or sacrifice " + name + "?",
						Else:     UpkeepSacrificeSource,
					}.Apply(NewContext(g, item))
				}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	}
}
