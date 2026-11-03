package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Choreographed Sparks — Instant {R}{R} (Secrets of Strixhaven, #4456):
//
//	"This spell can't be copied.
//	 Choose one or both —
//	 • Copy target instant or sorcery spell you control. You may choose
//	   new targets for the copy.
//	 • Copy target creature spell you control. The copy gains haste and
//	   'At the beginning of the end step, sacrifice this token.'"
//
// "This spell can't be copied" is Spec.CantBeCopied. With both bullets
// chosen each has its own target group and they resolve in printed
// order.
//
// The first bullet is Reverberate's CopySpell narrowed to "you control".
//
// The second is Double Major's copy with an except clause (#4688, ADR
// 0218). CR 707.9a: haste and the quoted trigger are gained as part of
// the copying process, so both are COPIABLE VALUES of the copy — a
// keyword (PrintedValues.AddKeyword) and a named ability bundle
// (PrintedValues.GrantAbility). CR 608.3f turns the resolving copy into
// a token that keeps those values, a Clone of the token copies them, and
// a token that loses all its abilities loses the trigger and stays. That
// is why this is not a delayed "sacrifice it at the next end step",
// which would be an effect of the spell and survive both.
//
// "The end step" is the next end step that comes, whoever's turn it is
// (the 2026-03-20 ruling: "next" was removed). The trigger is the
// token's own, so it sacrifices the token at THAT end step — AtEachStep,
// Lightning Serpent's shape.
const choreographedSparksGrant = "choreographed-sparks/sacrifice-this-token"

func init() {
	Register(Spec{
		OracleID:     "61495793-b0c7-40a5-ad3e-9df8e2b8096b",
		Name:         "Choreographed Sparks",
		Completeness: CompletenessFull,
		CantBeCopied: true,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Copy target instant or sorcery spell you control. You may choose new targets for the copy.",
				instantOrSorcerySpell("target instant or sorcery spell you control", YouControl()),
				copyModeTargetSpellWithNewTargets),
			ModeDoing("Copy target creature spell you control. The copy gains haste and \"At the beginning of the end step, sacrifice this token.\"",
				TargetSpell("target creature spell you control", Creature(), YouControl()),
				choreographedSparksCopyCreature),
		),
		Grants: []AbilityGrant{{
			Key:  choreographedSparksGrant,
			Text: "At the beginning of the end step, sacrifice this token.",
			Triggered: []game.TriggeredAbility{
				AtEachStep(game.StepEnd, "Choreographed Sparks — sacrifice this token", func(g *game.Game, item *game.StackItem) error {
					return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
			},
		}},
	})
}

// choreographedSparksCopyCreature is the second bullet: copy the targeted
// creature spell, gaining haste and the sacrifice trigger as copiable
// values (CR 707.9a).
func choreographedSparksCopyCreature(_ *game.StackItem, ctx *Context, occurrence int) error {
	refs := ctx.ModeTargets(occurrence)
	if len(refs) == 0 {
		return nil
	}
	return CopySpell{
		StackID:    refs[0].ID,
		Controller: ctx.Controller(),
		Except: func(v *game.PrintedValues) {
			v.AddKeyword("haste")
			v.GrantAbility(choreographedSparksGrant)
		},
	}.Apply(ctx)
}
