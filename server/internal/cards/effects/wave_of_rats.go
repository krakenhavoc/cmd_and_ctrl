package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wave of Rats — Creature — Rat {3}{B}, 4/2:
//
//	"Trample
//	 When this creature dies, if it dealt combat damage to a player this
//	 turn, return it to the battlefield under its owner's control.
//	 Blitz {4}{B}"
//
// Blitz (CR 702.152, #3772). The return is an intervening if (CR 603.4),
// read when the trigger would fire and again as it resolves, from the
// turn tally of which creature OBJECTS dealt damage to which player
// (TurnTally.DamageDealers, #2149). The object that died is the one
// whose epoch is one zone change before the card in the graveyard, the
// same reading sameGraveyardObject makes, so a Wave of Rats that hit
// someone, was flickered, and then died does not come back.
//
// A blitzed Wave of Rats that connected is sacrificed at the end step,
// draws a card, and returns as a new object with no haste and no
// blitz riders, to stay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "ccb0f59b-7947-4563-9e16-5d23246f4232",
		Name:             "Wave of Rats",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{Blitz("{4}{B}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return cardDied(ev, source) && diedObjectDealtCombatDamageToAPlayerThisTurn(g, source.InstanceID)
			}, "Wave of Rats — return it to the battlefield", waveOfRatsReturn),
		},
	})
}

// diedObjectDealtCombatDamageToAPlayerThisTurn reports whether the
// battlefield object the graveyard card `cardID` was, one zone change
// ago, dealt combat damage to a player this turn.
func diedObjectDealtCombatDamageToAPlayerThisTurn(g *game.Game, cardID uuid.UUID) bool {
	z := g.FindCardZoneForEffect(cardID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return false
	}
	epoch := -1
	for i := range z.Cards {
		if z.Cards[i].InstanceID == cardID {
			epoch = z.Cards[i].ObjectEpoch - 1
		}
	}
	for _, r := range g.TurnTally.DamageDealers {
		if r.Source == cardID && r.Epoch == epoch && r.Combat {
			return true
		}
	}
	return false
}

// waveOfRatsReturn re-checks the intervening if and returns the card,
// untapped, under its owner's control.
func waveOfRatsReturn(g *game.Game, item *game.StackItem) error {
	if !diedObjectDealtCombatDamageToAPlayerThisTurn(g, item.SourceCardID) {
		return nil
	}
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || !sameGraveyardObject(z, item) {
		return nil
	}
	_, err := g.ReturnFromGraveyardWithCountersForEffect(item.SourceCardID, uuid.Nil, false, nil)
	return nothingIfGone(err)
}
