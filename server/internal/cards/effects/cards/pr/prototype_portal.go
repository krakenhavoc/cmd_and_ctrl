package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prototype Portal — Artifact {4} (Commander 2018, S82 OLD batch 193, #5526):
//
//	"Imprint — When this artifact enters, you may exile an artifact card from
//	 your hand.
//	 {X}, {T}: Create a token that's a copy of the exiled card. X is the mana
//	 value of that card."
//
// Isochron Scepter's imprint: an ETB trigger that exiles the picked card
// LINKED to the Portal object (CR 607.2a), so a second Portal's card, or a card
// a Portal imprinted before it left and came back, is never the one copied.
// The activation's {X} is defined by the text (CR 107.3c, #6240): the engine
// prices it as generic mana equal to the mana value of the card exiled with
// this Portal (XIsTheManaValueOfTheExiledCard), nothing asks for an X, and
// with nothing imprinted it costs only {T} and makes nothing. The token copies
// the card where it sits in exile, by its printed copiable values (CR 707.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f3b6c25-b855-4ea5-a8ad-d954a6cdffb7",
		Name:         "Prototype Portal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Prototype Portal — you may exile an artifact card from your hand", prototypePortalImprint),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: Create a token that's a copy of the exiled card. X is the mana value of that card.",
			Cost:    effects.Plus(effects.XIsTheManaValueOfTheExiledCard(), effects.TapCost()),
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerMakesBlocker},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ref, ok := ctx.SourceRef()
				if !ok {
					return nil
				}
				linked := g.ExiledWithForEffect(game.PermissionCardRef(ref))
				if len(linked) == 0 {
					return nil
				}
				return effects.CreateTokenCopy{Controller: item.Controller, Copy: linked[len(linked)-1].InstanceID, N: 1}.Apply(ctx)
			},
		}},
	})
}

// prototypePortalImprint asks for the artifact card and exiles it linked to
// the Portal.
func prototypePortalImprint(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	if _, ok := g.SourceObjectForEffect(item); !ok {
		return nil
	}
	p := g.PlayerByIDForEffect(controller)
	if p == nil || p.Hand == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.IsArtifact() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   source,
		Question: "Prototype Portal — you may exile an artifact card from your hand",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		ThenKey:  prototypePortalImprintPick,
		LiveZone: game.ZoneHand,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// prototypePortalImprintPick exiles the picked card linked to the Portal object
// that asked, keyed: the asking item (and the object it names) rides the
// prompt as data. An on-disk key.
var prototypePortalImprintPick = effects.AskedPick("choose-cards/old-b193-prototype-portal-imprint", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	ref, ok := ctx.Game.SourceObjectForEffect(ctx.Item)
	if !ok {
		return nil
	}
	return ctx.Game.ExileLinkedForEffect(ref, r.Picked[0])
})
