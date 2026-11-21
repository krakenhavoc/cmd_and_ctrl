package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// theVisionLabel is the trigger's stack label, and with it the key its
// "hasn't been chosen this turn" memory is kept under.
const theVisionLabel = "The Vision — you cast a noncreature spell"

// The Vision — Legendary Artifact Creature — Robot Hero {4}, 2/5:
//
//	"Flying, vigilance
//	 Whenever you cast a noncreature spell, choose one that hasn't been
//	 chosen this turn —
//	 • Solar Beam — The Vision gains double strike until end of turn.
//	 • Density Control — The Vision gains indestructible until end of
//	   turn.
//	 • Technopathy — Draw a card."
//
// A cast trigger (YouCast(Noncreature())) with ChooseOneNotChosenThisTurn
// (ADR 0097): the first three noncreature spells a turn each pay a
// different bullet, and a fourth triggers and is removed with no
// effect. The bullet names are flavour words and have no rules
// meaning; they stay in the labels because the picker shows the
// bullet verbatim.
//
// The two keyword bullets are until-end-of-turn grants to the Vision
// itself, pinned to this object (CR 611.2c), so a Vision that is
// flickered in response does not keep them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "32aee76a-e738-4898-985a-801d1cce15ea",
		Name:            "The Vision",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			theVisionTrigger(),
		},
	})
}

func theVisionTrigger() game.TriggeredAbility {
	t := effects.WheneverYouCast(effects.Noncreature(), theVisionLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOneNotChosenThisTurn(
		effects.ModeDoing("Solar Beam — The Vision gains double strike until end of turn.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return theVisionGains(item, ctx, "double strike")
			}),
		effects.ModeDoing("Density Control — The Vision gains indestructible until end of turn.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return theVisionGains(item, ctx, "indestructible")
			}),
		effects.ModeDoing("Technopathy — Draw a card.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
	)
	return t
}

// theVisionGains is the two keyword bullets: "The Vision gains <kw>
// until end of turn". A Vision that left and came back before the
// trigger resolved is a new object (CR 400.7) and gains nothing.
func theVisionGains(item *game.StackItem, ctx *effects.Context, kw string) error {
	if effects.SourceIsNewObject(ctx.Game, item) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{kw},
		Label: "The Vision — " + kw + " until end of turn"}.Apply(ctx)
}
