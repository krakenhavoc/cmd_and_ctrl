package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Death of a Thousand Stings — Instant — Arcane {4}{B} (Saviors of
// Kamigawa, #5948): "Target player loses 1 life and you gain 1 life.
// At the beginning of your upkeep, if you have more cards in hand than
// each opponent, you may return this card from your graveyard to your
// hand." One of the cards Pinchy McStingbutt conjures (ADR 0099
// amendment 2026-11-24).
//
// The upkeep trigger works from the graveyard (CR 113.6) and its "if" is
// an intervening if (CR 603.4): asked as the upkeep begins and again as
// it resolves, before the "you may".
//
// No simplification.
func init() {
	const label = "Death of a Thousand Stings — return this card from your graveyard to your hand"
	effects.Register(effects.Spec{
		OracleID:     "25895c28-23c1-439e-858b-5cb31fdc31de",
		Name:         "Death of a Thousand Stings",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.ResolveSteps(effects.TargetPlayerLosesLife{N: 1}, effects.GainLife{Amount: 1}),
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.Optional(effects.InterveningIfYou("you have more cards in hand than each opponent", MoreCardsInHandThanEachOpponent,
				effects.AtYourUpkeep(label, effects.ReturnThisCardFromYourGraveyardToHand)), label+"?")),
		},
	})
}

// MoreCardsInHandThanEachOpponent is "you have more cards in hand than
// each opponent": strictly more than every opponent still in the game.
func MoreCardsInHandThanEachOpponent(g *game.Game, you uuid.UUID) bool {
	me := g.PlayerByIDForEffect(you)
	if me == nil || me.Hand == nil {
		return false
	}
	mine := len(me.Hand.Cards)
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you {
			continue
		}
		if p.Hand != nil && len(p.Hand.Cards) >= mine {
			return false
		}
	}
	return true
}
