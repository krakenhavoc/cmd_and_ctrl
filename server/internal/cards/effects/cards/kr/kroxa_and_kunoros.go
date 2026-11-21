package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kroxa and Kunoros — Legendary Creature — Elder Giant Dog {3}{R}{W}{B}, 6/6
// (March of the Machine Commander):
//
//	"Vigilance, menace, lifelink
//	 Whenever Kroxa and Kunoros enters or attacks, you may exile five cards
//	 from your graveyard. When you do, return target creature card from your
//	 graveyard to the battlefield."
//
// The "may exile five" is a pick made as the trigger resolves (CR 608.2c):
// exactly five graveyard cards or none, offered only with five to give. The
// cards go to exile first; a reflexive trigger (CR 603.12) then chooses its
// own "target creature card from your graveyard" as it goes on the stack, so
// a creature card that was among the five is no longer a candidate, and with
// none left the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "56679f06-982b-4f33-9e48-ecc64c802393",
		Name:            "Kroxa and Kunoros",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Kroxa and Kunoros — you may exile five cards from your graveyard; when you do, return target creature card from your graveyard to the battlefield",
				kroxaOffer),
		},
	})
}

var kroxaExileFive = effects.ExileCardsFromGraveyard(5, "five cards from your graveyard", nil)

// kroxaOffer asks for the five cards, when the graveyard holds that many.
func kroxaOffer(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !g.ExileCostPayableForEffect(item.Controller, uuid.Nil, kroxaExileFive) {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Kroxa and Kunoros — exile five cards from your graveyard to return a creature card? Choose none to decline.",
		Cards:    g.ExileCostOptionsForEffect(item.Controller, uuid.Nil, kroxaExileFive),
		Min:      0,
		Max:      5,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  kroxaExiled,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// kroxaExiled is the answer: exactly five cards, exiled, then the reflexive
// trigger. An on-disk key: never renamed, never reused.
var kroxaExiled = effects.AskedPickRule("choose-cards/kroxa-and-kunoros-exile-five",
	func(picked []game.Card) bool { return len(picked) == 5 },
	func(ctx *effects.Context, r game.CardsChosen) error {
		if len(r.Picked) != 5 {
			return nil
		}
		for _, id := range r.Picked {
			if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return effects.ReflexiveTrigger{
			Label: "Kroxa and Kunoros — return target creature card from your graveyard to the battlefield",
			Body:  kroxaReturnBody,
		}.Apply(ctx)
	})

// kroxaReturnBody is the reflexive trigger: return the target to the
// battlefield. An on-disk key.
var kroxaReturnBody = game.ReflexiveBody("kroxa-and-kunoros/return", effects.SimpleBody(kroxaReturn),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature())
	}))

func kroxaReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
	}
	return nil
}
