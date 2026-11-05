package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tempestra, Dame of Games — {1}{R} Legendary Creature — Elemental Illusion
// 1/3 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"{2}{R}, {T}, Sacrifice an artifact: Create a token that's a copy of
//	 another target creature you control, except it isn't legendary. It
//	 gains haste. Sacrifice it at the beginning of the next end step."
//
// Chrome Dome's activation with Loki, the Deceiver's exception: the token
// takes the target's copiable values (CR 707.2, so a creature that is itself
// a copy is copied as what it copies, and a token as its creating effect
// made it) with "isn't legendary" written into them (CR 707.9b). Haste is
// not a copiable value: it is a continuous grant to that token with no
// stated duration (CR 611.2a). The sacrifice is a delayed trigger at the
// beginning of the next end step (CR 603.7). A creature source, so the tap
// waits out summoning sickness (CR 302.6); the artifact is sacrificed as a
// cost (CR 602.2b). The target is re-read as the ability resolves
// (CR 608.2b): a creature that left in response makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12401ad1-b5c9-43dc-9389-04aaf5db8b26",
		Name:         "Tempestra, Dame of Games",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, {T}, Sacrifice an artifact: Create a token that's a copy of another target creature you control, except it isn't legendary. It gains haste. Sacrifice it at the beginning of the next end step.",
			Cost:    Plus(ManaCost("{2}{R}"), TapCost(), SacrificeN(1, "an artifact", Artifact())),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect:  tempestraCopy,
		}},
	})
}

// tempestraCopy makes the nonlegendary token copy of the still-legal target.
func tempestraCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	copyOf := FirstLegalBattlefieldTarget(ctx)
	if copyOf == uuid.Nil {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(g, copyOf)
	if !ok {
		return nil
	}
	tmpl.TypeLine = removeLegendaryFromTypeLine(tmpl.TypeLine)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: copiedToken(tmpl), Count: 1}},
	}, ctx.ItemTail(tempestraCopyRest))
}

// tempestraCopyRest is "It gains haste. Sacrifice it at the beginning of
// the next end step." as a keyed tail, told the tokens the creation
// actually made. An on-disk key: never renamed, never reused.
var tempestraCopyRest = CardTail("tempestra-dame-of-games/haste-and-sacrifice-copy", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
			Duration: game.IndefiniteDuration(), Label: "Tempestra — the token gains haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	if len(o.IDs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Item.Controller,
		Label:      "Tempestra — sacrifice the token",
		Cards:      o.IDs,
		Body:       sacrificeListedCardsBody,
	}.Apply(ctx)
})
