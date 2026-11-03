package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brudiclad, Telchor Engineer — Legendary Artifact Creature — Phyrexian
// Artificer 4/4 for {4}{U}{R} (Secrets of Strixhaven Commander, #4834):
//
//	"Creature tokens you control have haste.
//	 At the beginning of combat on your turn, create a 2/1 blue Phyrexian
//	 Myr artifact creature token. Then you may choose a token you
//	 control. If you do, each other token you control becomes a copy of
//	 that token."
//
// The haste grant is Teysa Karlov's layer 6 shape over
// creatureTokensYouControl.
//
// The trigger is one ability in three steps, each waiting on the last.
// The Myr is created through CreateTokensKeyedForEffect, whose keyed tail
// runs once the token exists (a token doubler's ordering prompt can pause
// the creation), so the new Myr is on the battlefield to be chosen and to
// be copied over. The choice is a ChoosePermanents pick of zero or one of
// the tokens you control, with no target (CR 115.10a's "choose" does not
// target): a token with hexproof can be chosen, and choosing none is the
// printed "you may". The copy is BecomeCopy with no duration (CR 611.2a,
// CopyIndefinite) over every OTHER token you control as the pick is
// answered, so a token that arrives later in the turn is left alone
// (CR 611.2c). The copiable values are the chosen token's (CR 707.2),
// read as the copy begins.
func init() {
	Register(Spec{
		OracleID:     "a7ac5d98-805e-4212-be94-333a0c0c44a7",
		Name:         "Brudiclad, Telchor Engineer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(creatureTokensYouControl, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Brudiclad, Telchor Engineer — create a 2/1 Myr, then tokens may become copies of one token", brudicladBeginCombat),
		},
	})
}

func brudicladBeginCombat(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("2/1 blue Phyrexian Myr artifact")), Count: 1}},
	}, ctx.ItemTail(brudicladChooseToken))
}

// brudicladChooseToken is "Then you may choose a token you control."
var brudicladChooseToken = CardTail("brudiclad/choose-a-token", func(ctx *Context, _ game.TailOutcome) error {
	return ChoosePermanents{
		Question: "Brudiclad, Telchor Engineer — you may choose a token you control; each other token you control becomes a copy of it",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			ids := tokensYouControlExcept(g, of, uuid.Nil)
			return ids, 0, min(1, len(ids))
		},
		Keyed: brudicladChosen,
	}.Apply(ctx)
})

// brudicladChosen is "If you do, each other token you control becomes a
// copy of that token."
var brudicladChosen = RegisterPermanentsPick("prompt-run/brudiclad-tokens-become-copies", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		chosen := picked.Cards()
		if len(chosen) == 0 || !onBattlefield(ctx.Game, chosen[0]) {
			return nil
		}
		others := tokensYouControlExcept(ctx.Game, ctx.Controller(), chosen[0])
		if len(others) == 0 {
			return nil
		}
		return BecomeCopy{
			Targets:  others,
			Of:       chosen[0],
			Duration: CopyIndefinite,
			Label:    "Brudiclad, Telchor Engineer — becomes a copy of the chosen token",
		}.Apply(ctx)
	},
})
