package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rionya, Fire Dancer — Legendary Creature — Human Wizard {3}{R}{R}, 3/4
// (Secrets of Strixhaven Commander, #4834):
//
//	"At the beginning of combat on your turn, create X tokens that are
//	 copies of another target creature you control, where X is one plus
//	 the number of instant and sorcery spells you've cast this turn. They
//	 gain haste. Exile them at the beginning of the next end step."
//
// The target is another creature (the clause excludes Rionya as an
// object, so a second Rionya or a token copy of her stays eligible). X is
// read as the trigger resolves, so an instant cast in response counts (the
// ruling); CastTally.InstantSorcery counts the instants and sorceries
// cast, not copies, as printed.
//
// The copies are the target's copiable values (CR 707.2; the ruling: an
// X in the mana cost is 0, and nothing but copiable values is copied),
// each given haste as the token is made (TokenCopyGainsHaste); the tokens
// are exiled this turn anyway, so "gain haste" has no duration to get wrong.
// The tokens that actually landed are handed to the end step exile through
// the creation's keyed tail (socB8ExileTokensAtNextEndStep, CR 603.7), the
// shape Manaform Hellkite uses: a table waiting on a replacement prompt for
// the creation is a restore point, and a token that a replacement kept off
// the battlefield is simply not exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be4f47e9-1bd2-4aa9-bc85-3f203ad21639",
		Name:         "Rionya, Fire Dancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfYourCombat("Rionya, Fire Dancer — create X tokens that are copies of another target creature you control; they gain haste; exile them at the next end step",
					rionyaCopies),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}

func rionyaCopies(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	tmpl, ok := effects.TokenCopyTemplate(g, target)
	if !ok {
		return nil
	}
	effects.TokenCopyGainsHaste(&tmpl)
	x := 1 + g.CastTallyFor(ctx.Controller()).InstantSorcery
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: ctx.Controller(),
		Source:     ctx.Source(),
		Groups:     []game.TokenGroup{{Template: effects.CopiedToken(tmpl), Count: x}},
	}, ctx.ItemTail(rionyaExileTheCopies))
}

// rionyaExileTheCopies is "exile them at the beginning of the next end
// step". An on-disk key.
var rionyaExileTheCopies = effects.SocB8ExileTokensAtNextEndStep("rionya-fire-dancer/exile-the-copies",
	"Rionya, Fire Dancer — exile the tokens")
