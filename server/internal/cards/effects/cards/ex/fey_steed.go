package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fey Steed — Creature — Elk {2}{W}{W}, 3/3 (Adventures in the Forgotten
// Realms Commander):
//
//	"Whenever this creature attacks, another target attacking creature
//	 you control gains indestructible until end of turn.
//	 Whenever a creature or planeswalker you control becomes the target of
//	 a spell or ability an opponent controls, you may draw a card."
//
// Two printed triggers, two registered rows. The attack trigger targets
// another attacking creature you control and grants indestructible
// (CR 702.12) as it resolves, re-checking the target (CR 608.2b). The
// second is Surrak's EventBecomesTarget watch, fired once per target slot
// (CR 115.3, so a spell with two of your permanents as targets triggers
// twice), narrowed to a creature or planeswalker on the battlefield that
// you control and a spell or ability an opponent of yours controls. It
// can be Fey Steed itself. The draw is a "you may" (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18cced8b-9836-4e11-bb4e-69a971a33a50",
		Name:         "Fey Steed",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(effects.WheneverThisAttacks("Fey Steed — another target attacking creature you control gains indestructible until end of turn",
					feySteedGrant),
					effects.Another(effects.TargetCreature("another target attacking creature you control", effects.AttackingCreature(), effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible"))),
			effects.OptionalDraw(effects.On(game.EventBecomesTarget, feySteedYourPermanentTargetedByAnOpponent,
				"Fey Steed — you may draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Fey Steed — draw a card?", 1),
		},
	})
}

func feySteedGrant(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"},
		Label: "Fey Steed — indestructible until end of turn"}.Apply(ctx)
}

// feySteedYourPermanentTargetedByAnOpponent is "a creature or planeswalker
// you control becomes the target of a spell or ability an opponent
// controls": the event's object is on the battlefield under your control
// and its Actor, the controller of the targeting spell or ability, is an
// opponent.
func feySteedYourPermanentTargetedByAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil || ev.CardID == uuid.Nil {
		return false
	}
	you := source.Controller
	if you == uuid.Nil || ev.Actor == uuid.Nil || ev.Actor == you || g.PlayerByIDForEffect(ev.Actor) == nil {
		return false
	}
	if !effects.OnBattlefield(g, ev.CardID) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == you && (c.IsCreature() || c.IsPlaneswalker())
}
