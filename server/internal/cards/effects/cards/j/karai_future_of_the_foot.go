package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karai, Future of the Foot — Legendary Creature — Human Ninja {1}{W}{B},
// 3/3 (TMT, #4457):
//
//	"Sneak {2}{W}{B} (… She enters tapped and attacking.)
//	 Whenever Karai deals combat damage to a player, return target
//	 creature card from your graveyard to your hand. If her sneak cost
//	 was paid this turn, instead return that card to the battlefield."
//
// Sneak is Sneak(cost) (CR 702.190). "Her sneak cost was paid this turn"
// reads the cast record the way Adipose Offspring reads emerge
// (ObjectSnapshot.AltCost, CR 400.7d): the trigger's subject is the
// creature that dealt the damage (SubjectIsEventSource), snapshotted as
// the damage is dealt, so the answer holds if she leaves in response.
// "This turn" is that object having entered this turn: a sneak cast
// happens only in its caster's declare blockers step and resolves that
// turn, so a Karai sneaked in on an earlier turn returns to hand.
//
// No simplification.
func init() {
	t := effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer("Karai, Future of the Foot — return target creature card from your graveyard to your hand, or to the battlefield if her sneak cost was paid this turn",
		karaiFutureOfTheFootReturn),
		effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()))
	t.SubjectIsEventSource = true
	effects.Register(effects.Spec{
		OracleID:         "7246638e-0361-40e3-8019-3c10210a593e",
		Name:             "Karai, Future of the Foot",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Sneak("{2}{W}{B}")},
		Triggered:        []game.TriggeredAbility{t},
	})
}

// karaiFutureOfTheFootReturn is the trigger's body: to hand, or, when her
// sneak cost was paid this turn, to the battlefield.
func karaiFutureOfTheFootReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if obj := ctx.Trigger().Object; obj != nil && effects.TmtB11PaidSneak(obj.AltCost) && g.EnteredThisTurn(obj.ID) {
		return effects.ReturnLegalGraveyardTargetsToBattlefield(ctx)
	}
	return effects.ReturnLegalGraveyardTargetsToHand(ctx)
}
