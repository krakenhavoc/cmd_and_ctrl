package cardsp

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Paladin Elizabeth Taggerdy — Legendary Creature — Human Knight {1}{R}{W},
// 3/2 (Fallout, #5948):
//
//	"Battalion — Whenever Paladin Elizabeth Taggerdy and at least two other
//	 creatures attack, draw a card, then you may put a creature card with mana
//	 value X or less from your hand onto the battlefield tapped and attacking,
//	 where X is Paladin Elizabeth Taggerdy's power."
//
// Battalion is her own attack with two other attackers in the declaration
// (battalionAttack). She draws first, so the card drawn can be the one put
// onto the battlefield; the candidates are the creature cards in hand as the
// put is offered, with a mana value of at most her power as it is then (her
// last-known power if she has left; an {X} in a card's cost is 0 in hand,
// CR 202.3e). The card enters tapped and attacking a player, planeswalker or
// battle its controller chooses (Kinscaer Sentry's HandAttackerPutter, CR
// 508.4), never declared, so it triggers no "whenever attacks" ability. It may
// be declined.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3cfea32e-6b52-441b-90aa-d8a137997297",
		Name:         "Paladin Elizabeth Taggerdy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.BattalionAttack,
				"Paladin Elizabeth Taggerdy — draw a card, then you may put a creature card with mana value X or less from your hand onto the battlefield tapped and attacking",
				paladinElizabethTrigger),
		},
	})
}

var paladinElizabethPutter = effects.NewHandAttackerPutter("paladin-elizabeth-taggerdy", "Paladin Elizabeth Taggerdy", paladinElizabethEntered)

// paladinElizabethEntered follows the entry with nothing. An on-disk key.
var paladinElizabethEntered = effects.EnteredTail("paladin-elizabeth-taggerdy/entered", func(_ *effects.Context, _ uuid.UUID, _ game.TailOutcome) error {
	return nil
})

func paladinElizabethTrigger(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	power := 0
	if effects.OnBattlefield(g, item.SourceCardID) {
		power = effects.B31CurrentPowerOf(g, item.SourceCardID)
	} else if info, ok := g.LastKnownPermanentForEffect(item.SourceCardID); ok {
		power = info.Power
	}
	candidates := effects.HandCardsMatching(g, item.Controller, effects.And(effects.Creature(), effects.ManaValueLE(power)))
	paladinElizabethPutter.Offer(g, item, candidates,
		"Paladin Elizabeth Taggerdy — put a creature card with mana value "+strconv.Itoa(power)+" or less from your hand onto the battlefield tapped and attacking?")
	return nil
}
