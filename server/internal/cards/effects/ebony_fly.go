package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ebony Fly — Artifact {2} (The List, S82):
//
//	"This artifact enters tapped.
//	 {T}: Add {C}.
//	 {4}: Roll a d6. Until end of turn, you may have this artifact become
//	 an X/X Insect artifact creature with flying, where X is the result.
//	 Whenever this artifact attacks, another target attacking creature
//	 gains flying until end of turn."
//
// The d6 goes through the one dice door; its tail asks "you may" (MayChoice)
// and, on a yes, makes this artifact an X/X Insect artifact creature with
// flying until end of turn (CR 611.2). The result is the natural die, carried
// as the prompt's Word. The animation is pinned to this object (CR 400.7), so
// it does nothing for a Fly that left. The attack trigger is Helitrooper's.
//
// No simplification.
func init() {
	const label = "Ebony Fly — another attacking creature gains flying"
	attack := WheneverThisAttacks(label, firstTargetGainsUntilEndOfTurn(0, label, "flying"))
	attack.Targets = Another(TargetCreature("another target attacking creature", AttackingCreature()))
	Register(Spec{
		OracleID:     "2a2f2d6c-3937-4e4b-9e4a-3480e0c787d0",
		Name:         "Ebony Fly",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}: Roll a d6. Until end of turn, you may have this artifact become an X/X Insect artifact creature with flying.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{4}"),
			Effect:  Do(RollDice{Sides: 6, N: 1, Then: ebonyFlyRolled}),
		}},
		Triggered: []game.TriggeredAbility{attack},
	})
}

// ebonyFlyRolled asks the "you may" once the d6 is kept. An on-disk key.
var ebonyFlyRolled = DiceTail("dice/ebony-fly-rolled", func(ctx *Context, rolls []int) error {
	return MayChoice{
		Question: "Ebony Fly — have it become an " + strconv.Itoa(rolls[0]) + "/" + strconv.Itoa(rolls[0]) + " Insect artifact creature with flying until end of turn?",
		Then:     ebonyFlyBecomes,
		Word:     strconv.Itoa(rolls[0]),
	}.Apply(ctx)
})

// ebonyFlyBecomes is the "yes": X/X Insect artifact creature with flying.
// An on-disk key.
var ebonyFlyBecomes = MayYes("confirm/ebony-fly-becomes", func(ctx *Context, r game.ConfirmAnswer) error {
	x := wordInt(r)
	if x <= 0 {
		return nil
	}
	if err := (BecomeCreatureUntilEOT{
		Types:        []string{"Artifact", "Creature"},
		Subtypes:     []string{"Insect"},
		SetPower:     x,
		SetToughness: x,
		Label:        "Ebony Fly — becomes an X/X Insect artifact creature",
	}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"flying"}, Label: "Ebony Fly — flying"}.Apply(ctx)
})
