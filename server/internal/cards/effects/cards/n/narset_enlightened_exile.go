package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Narset, Enlightened Exile — Legendary Creature — Human Monk {1}{U}{R}{W},
// 3/4 (March of the Machine: The Aftermath, #5476):
//
//	"Creatures you control have prowess.
//	 Whenever Narset attacks, exile target noncreature, nonland card with
//	 mana value less than Narset's power from a graveyard and copy it. You
//	 may cast the copy without paying its mana cost."
//
// The grant is Bria, Riptide Rogue's (prowess is derived from the effective
// ability list, so a creature with its own prowess triggers twice,
// CR 702.108b). The trigger is Kaervek, the Punisher's exile-and-copy with
// a free cast (CopyAndCast, CR 707.12): the card comes from any graveyard,
// and "less than Narset's power" is data on the clause, judged as the
// target is chosen and as the trigger resolves (CR 608.2b), against
// last-known power if she has left (CR 608.2h). Only a card that really
// reached exile is copied; the copy is cast during the resolution
// (CR 608.2g) and a permanent copy would become a token, though a
// noncreature, nonland permanent card (an artifact or enchantment) is
// possible here.
//
// No simplification.
var narsetCopyExiled = effects.CardTail("narset-enlightened-exile/copy-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.CopyAndCast{Cards: o.IDs, Free: true}.Apply(effects.NewContext(ctx.Game, ctx.Item))
})

func init() {
	effects.Register(effects.Spec{
		OracleID:     "d92725de-ead0-4a6c-83db-55ec3acc7184",
		Name:         "Narset, Enlightened Exile",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.KeywordGrant(effects.CreaturesYouControl, "prowess")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventAttack, effects.ThisAttacked,
					"Narset, Enlightened Exile — exile target noncreature, nonland card with mana value less than her power from a graveyard and copy it; you may cast the copy without paying its mana cost",
					effects.MatB04ExileTargetsAndCopy(narsetCopyExiled)),
				effects.RelativeToSource(
					effects.TargetCardInGraveyard("target noncreature, nonland card with mana value less than Narset's power from a graveyard",
						effects.Noncreature(), effects.Nonland()),
					effects.ManaValueLessThanPower())),
		},
	})
}
