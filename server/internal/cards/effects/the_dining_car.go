package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Dining Car — Plane — Spacecraft (Doctor Who, #4870):
//
//	"When you planeswalk to The Dining Car, each player creates a Food
//	 token.
//	 Sixty-Six Seconds — At the beginning of your upkeep, sacrifice a
//	 creature with the least toughness among creatures you control. Then
//	 investigate.
//	 Whenever chaos ensues, activated abilities of artifact tokens you
//	 control cost {2} less to activate this turn."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). Each
// player creates their own Food. The upkeep ability has the planar controller
// choose among the creatures tied for the least toughness (CR 701.21a: with
// none, nothing is sacrificed), and investigates afterwards either way, once
// the sacrifice has settled. The chaos ability (CR 311.7) is a duration effect
// on the planar controller, not on the plane, so it lasts the turn even if the
// table planeswalks (game.GrantActivationDiscountForEffect): one more CR 601.2f
// reduction of generic mana on each activation of an artifact token they
// control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7099883d-0a4b-47ae-bff8-a96efa7a3a9a",
		Name:         "The Dining Car",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThis("The Dining Car — each player creates a Food token", avB30cEachPlayerCreatesAFood),
			AtYourUpkeep("The Dining Car — Sixty-Six Seconds: sacrifice a creature with the least toughness among creatures you control, then investigate",
				theDiningCarSacrifice),
			WheneverChaosEnsues("The Dining Car — activated abilities of artifact tokens you control cost {2} less to activate this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					g.GrantActivationDiscountForEffect(item.Controller, game.ActivationDiscountGrant{
						Amount: 2, ArtifactTokensOnly: true,
						Label: "The Dining Car — activated abilities of artifact tokens you control cost {2} less to activate this turn",
					}, item.SourceCardID, DurationUntilEndOfTurn(ctx))
					return nil
				}),
		},
	})
}

func theDiningCarSacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var least []uuid.UUID
	lowest := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsCreature() {
			continue
		}
		t := c.CurrentToughness()
		switch {
		case len(least) == 0 || t < lowest:
			lowest, least = t, []uuid.UUID{c.InstanceID}
		case t == lowest:
			least = append(least, c.InstanceID)
		}
	}
	return SacrificeChoice{
		Player:     item.Controller,
		Candidates: least,
		Question:   "The Dining Car — sacrifice a creature with the least toughness",
		ThenTail:   ctx.ItemTail(theDiningCarInvestigate),
	}.Apply(ctx)
}

// theDiningCarInvestigate is "Then investigate", once the sacrifice has
// settled. An on-disk key: never renamed, never reused.
var theDiningCarInvestigate = CardTail("then/the-dining-car-investigate", func(ctx *Context, _ game.TailOutcome) error {
	return CreateToken{Template: ClueToken(), N: 1}.Apply(ctx)
})
