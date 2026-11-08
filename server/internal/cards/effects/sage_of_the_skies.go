package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sage of the Skies — Creature — Human Monk {2}{W}, 2/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"When you cast this spell, if you've cast another spell this turn,
//	 copy this spell. (The copy becomes a token.)
//	 Flying, lifelink"
//
// Hatut Zeraze Strike Force's cast trigger (WhenYouCastThisSpell, CR
// 601.2i) with Eshki Dragonclaw's intervening "if" over the turn's cast
// tally (CR 603.4): the tally already counts the Sage as it is cast, so
// "another spell" is a total of two or more. The tally never falls within
// a turn, so a condition true as the ability triggers is still true as it
// resolves. The copy is a spell copy (CR 707.10) that becomes a token as it
// resolves (CR 111.13); it was not cast, so it triggers nothing that reads
// a cast, its own cast trigger included. The spell is named, not targeted,
// so a Sage countered in response is still copied from last-known
// information (CR 608.2h). Flying and lifelink are printed keywords.
//
// No simplification.
func init() {
	const label = "Sage of the Skies — copy this spell"
	trigger := InterveningIf(CondBoard{
		Name: "you've cast another spell this turn",
		Test: func(ctx *Context) bool { return sageOfTheSkiesCastAnother(ctx.Game, ctx.Controller()) },
	}, WhenYouCastThisSpell(label, func(g *game.Game, item *game.StackItem) error {
		return CopySpell{StackID: item.SourceCardID, Controller: item.Controller, Count: 1,
			FromLastKnown: true}.Apply(NewContext(g, item))
	}))
	trigger.AppliesTo = AllOf(Self, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		return sageOfTheSkiesCastAnother(g, ev.Actor)
	})
	Register(Spec{
		OracleID:        "cfc7cef7-0a95-49f8-970f-abff2cacd82d",
		Name:            "Sage of the Skies",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}

// sageOfTheSkiesCastAnother is "you've cast another spell this turn": the
// tally counts the Sage itself, so another is a second.
func sageOfTheSkiesCastAnother(g *game.Game, player uuid.UUID) bool {
	return g.CastTallyFor(player).Total >= 2
}
