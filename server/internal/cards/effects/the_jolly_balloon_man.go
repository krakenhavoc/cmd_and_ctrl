package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Jolly Balloon Man — Legendary Creature — Human Clown {1}{R}{W}, 1/4
// (Duskmourn: House of Horror, #4466):
//
//	"Haste
//	 {1}, {T}: Create a token that's a copy of another target creature you
//	 control, except it's a 1/1 red Balloon creature in addition to its
//	 other colors and types and it has flying and haste. Sacrifice it at
//	 the beginning of the next end step. Activate only as a sorcery."
//
// Reflection of Kiki-Jiki's ability with Ember Island Production's
// exceptions (CR 707.9): the token copies the target's copiable values as
// the ability resolves, then its base power and toughness are 1/1, red is
// added to its colours, Balloon to its creature types, and flying and haste
// to its keywords, all as copiable values of the token (CR 707.9a). The
// colour and the subtype are words of this card's text, read through it
// (CR 612.2). The token's sacrifice is a delayed trigger at the next end
// step (CR 603.7), scheduled for the token the copy made. A target that
// left in response makes nothing (CR 608.2b). Haste is printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "82ea519f-0e14-43d3-bf46-a1c5352e9600",
		Name:            "The Jolly Balloon Man",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:        "{1}, {T}: Create a token that's a copy of another target creature you control, except it's a 1/1 red Balloon creature in addition to its other colors and types and it has flying and haste. Sacrifice it at the beginning of the next end step. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}"), TapCost()),
			Targets:      Another(TargetCreature("another target creature you control", YouControl())),
			SorcerySpeed: true,
			Effect:       jollyBalloonManCopy,
		}},
	})
}

// jollyBalloonManCopy makes the Balloon copy and schedules its sacrifice.
func jollyBalloonManCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		red := ctx.TextColors("R")
		balloon := ctx.TextSubtype("Balloon")
		cursor := b25LastEventSeq(g)
		if err := (CreateTokenCopy{
			Controller: item.Controller,
			Copy:       t.ID,
			N:          1,
			Except: func(tok *game.Card) {
				colors := append([]string(nil), tok.EffectiveColors()...)
				for _, c := range red {
					if !slices.Contains(colors, c) {
						colors = append(colors, c)
					}
				}
				tok.Colors = colors
				tok.TypeLine = withSubtypeInAddition(tok.TypeLine, balloon)
				tok.Power, tok.Toughness = 1, 1
				tok.VariableToughness = false
				tok.PrintedPTKnown = true
				for _, kw := range []string{"flying", "haste"} {
					if !slices.Contains(tok.Keywords, kw) {
						tok.Keywords = append(tok.Keywords, kw)
					}
				}
			},
		}).Apply(ctx); err != nil {
			return err
		}
		tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
		if len(tokens) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label: "The Jolly Balloon Man — sacrifice the token",
			Cards: tokens,
			Body:  sacrificeListedCardsBody,
		}.Apply(ctx)
	}
	return nil
}
