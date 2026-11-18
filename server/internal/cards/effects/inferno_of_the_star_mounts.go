package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inferno of the Star Mounts — Legendary Creature — Dragon {4}{R}{R}, 6/6
// (Adventures in the Forgotten Realms Commander):
//
//	"This spell can't be countered.
//	 Flying, haste
//	 {R}: Inferno of the Star Mounts gets +1/+0 until end of turn. When its
//	 power becomes 20 this way, it deals 20 damage to any target."
//
// Uncounterable is Spec.CantBeCountered. The activation is firebreathing; as
// it resolves its power is read before and after the pump, and it raises the
// CR 603.12 reflexive trigger only when this pump is what took the power to
// exactly 20, so one that is already 20 or more (or that other effects took
// past it) does not. The reflexive trigger chooses its "any target" as it goes
// on the stack and deals the 20 damage as Inferno, as it last existed if it
// has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cd54438d-7f9d-4d80-b15b-0f0bf4becc15",
		Name:            "Inferno of the Star Mounts",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"flying", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "{R}: Inferno of the Star Mounts gets +1/+0 until end of turn. When its power becomes 20 this way, it deals 20 damage to any target.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{R}"),
			Effect:  infernoOfTheStarMountsPump,
		}},
	})
}

// infernoOfTheStarMountsDamageBody is the reflexive trigger's registration:
// "deals 20 damage to any target", the amount on Params.Amount.
var infernoOfTheStarMountsDamageBody = game.ReflexiveBody("inferno-of-the-star-mounts/damage", breechesBlastEffect, constTargets(TargetAny))

func infernoOfTheStarMountsPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	before, ok := afrB18PowerOf(g, ctx.Source())
	if !ok {
		return nil
	}
	if err := (BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Inferno of the Star Mounts — +1/+0"}).Apply(ctx); err != nil {
		return err
	}
	g.RecomputeLayersIfStaleLocked()
	after, ok := afrB18PowerOf(g, ctx.Source())
	if !ok || before >= 20 || after != 20 {
		return nil
	}
	return ReflexiveTrigger{
		Label:  "Inferno of the Star Mounts — 20 damage to any target",
		Body:   infernoOfTheStarMountsDamageBody,
		Params: game.EffectParams{Amount: 20},
	}.Apply(ctx)
}
