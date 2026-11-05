package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Irma, Part-Time Mutant — Legendary Creature — Human Mutant Shapeshifter
// {2}{U}, 1/1 (TMNT Commander, #5011):
//
//	"At the beginning of combat on your turn, Irma becomes a copy of up
//	 to one other target creature you control, except her name is Irma,
//	 Part-Time Mutant and she has this ability. Then put a +1/+1 counter
//	 on her."
//
// A duration copy with no stated duration (CR 611.2a, become_copy.go):
// it lasts until Irma leaves the battlefield or copies again, and the
// engine keeps only the newest record. The values are the target's
// copiable values (CR 707.2), so a target that is itself copying
// something gives the thing it copies, and counters, tapped status and
// attachments are not copied. She does not enter, so the copied
// creature's enters abilities do nothing.
//
// The except clause is two edits made as part of the copy (CR 707.9):
// the name, and this ability granted as a named bundle (CR 707.9a), so
// Irma as a copy of a Bear still has the trigger, and a Clone of her
// does too. The printed trigger and the granted one are the same
// TriggeredAbility value, so they cannot drift.
//
// The rulings: with no target chosen, she still gets the counter. With
// a target chosen that is illegal as the ability resolves, the ability
// does not resolve at all (CR 608.2b): no copy and no counter. The
// counter goes on her only while she is still the same object on the
// battlefield (CounterOnThis).
//
// No simplification.
const irmaPartTimeMutantGrant = "irma-part-time-mutant/this-ability"

const irmaPartTimeMutantText = "At the beginning of combat on your turn, Irma becomes a copy of up to one other target creature you control, except her name is Irma, Part-Time Mutant and she has this ability. Then put a +1/+1 counter on her."

var irmaPartTimeMutantTrigger = Targeting(
	AtBeginningOfYourCombat("Irma, Part-Time Mutant — becomes a copy of another creature you control, then a +1/+1 counter",
		irmaPartTimeMutantCopyThenCounter),
	Another(UpToOneTargetCreature("up to one other target creature you control", YouControl())),
)

func init() {
	Register(Spec{
		OracleID:     "deda54da-4fcc-4da3-9337-7768a2db3283",
		Name:         "Irma, Part-Time Mutant",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       irmaPartTimeMutantGrant,
			Text:      irmaPartTimeMutantText,
			Triggered: []game.TriggeredAbility{irmaPartTimeMutantTrigger},
		}},
		Triggered: []game.TriggeredAbility{irmaPartTimeMutantTrigger},
	})
}

func irmaPartTimeMutantCopyThenCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if targets := ctx.LegalTargets(); len(targets) > 0 {
		if err := (BecomeCopy{
			Targets:  []uuid.UUID{ctx.Source()},
			Of:       targets[0].ID,
			Duration: CopyIndefinite,
			Except: func(v *game.PrintedValues) {
				v.SetName("Irma, Part-Time Mutant")
				v.GrantAbility(irmaPartTimeMutantGrant)
			},
			Label: "Irma, Part-Time Mutant — becomes a copy",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
