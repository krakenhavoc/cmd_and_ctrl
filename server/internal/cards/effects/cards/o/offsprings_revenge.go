package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Offspring's Revenge — Enchantment {2}{R}{W}{B} (Ikoria: Lair of Behemoths,
// #5512):
//
//	"At the beginning of combat on your turn, exile target red, white, or
//	 black creature card from your graveyard. Create a token that's a copy of
//	 that card, except it's 1/1. It gains haste until your next turn."
//
// The Scarab God's exile-then-copy (ExileTargetThenTokenCopy): the target is
// chosen as the trigger goes on the stack and only a card that really reached
// exile is copied (CR 608.2b). The 1/1 is a copiable exception on the token
// (CR 707.9b, Coiling Rebirth's), and the haste is a separate effect on that
// token that lasts until your next turn (CR 611.2b, Rootwise Survivor's), not
// part of the copy.
//
// No simplification.
func init() {
	trigger := effects.AtBeginningOfYourCombat("Offspring's Revenge — exile target red, white, or black creature card from your graveyard, create a token copy of it that's 1/1 with haste until your next turn",
		func(g *game.Game, item *game.StackItem) error {
			return effects.ExileTargetThenTokenCopy(g, item, offspringsRevengeCopy)
		})
	trigger.Targets = effects.TargetCardInGraveyard("target red, white, or black creature card from your graveyard",
		effects.YouOwn(), effects.Creature(), effects.Or(effects.OfColor("R"), effects.OfColor("W"), effects.OfColor("B")))
	effects.Register(effects.Spec{
		OracleID:     "e9fa3e3c-4a7e-4a18-b2ef-15d53ff55321",
		Name:         "Offspring's Revenge",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}

// offspringsRevengeCopy is "create a token that's a copy of that card, except
// it's 1/1. It gains haste until your next turn", run once the card has
// landed in exile. An on-disk key.
var offspringsRevengeCopy = effects.CardTail("then/offsprings-revenge-copy", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	tmpl, ok := effects.TokenCopyTemplate(ctx.Game, o.IDs[0])
	if !ok {
		return nil
	}
	tmpl.Power, tmpl.Toughness = 1, 1
	tmpl.VariableToughness = false
	tmpl.PrintedPTKnown = true
	made, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), effects.CopiedToken(tmpl), 1, game.TokenEntryOptions{})
	if err != nil {
		return err
	}
	for _, id := range made {
		if err := (effects.ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
			Label:    "Offspring's Revenge — haste until your next turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
