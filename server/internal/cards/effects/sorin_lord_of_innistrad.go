package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sorin, Lord of Innistrad — Legendary Planeswalker — Sorin
// {2}{W}{B}, starting loyalty 3 (#4077, deck request Sorin's Birthday
// Party):
//
//	"+1: Create a 1/1 black Vampire creature token with lifelink.
//	 −2: You get an emblem with "Creatures you control get +1/+0."
//	 −6: Destroy up to three target creatures and/or other
//	     planeswalkers. Return each card put into a graveyard this way
//	     to the battlefield under your control."
//
// THE EMBLEM (CR 114, ADR 0064): one layer-7c +1/+0 on the creatures its
// owner controls, applied from the command zone, so it outlives Sorin
// and counts creatures that arrive later (an emblem is not a
// resolution-time snapshot the way an "until end of turn" pump is). It
// is the same predicate Elspeth, Sun's Champion's emblem uses.
//
// THE −6 destroys the (up to three) targets as one simultaneous event
// (DestroyPermanentsKeyedForEffect), then returns every one that really
// reached a graveyard — read off the landed outcome, so an indestructible
// target (never destroyed), a token (it ceases to exist, CR 111.7) and a
// permanent a replacement sent elsewhere are not returned, as printed.
// The cards enter together under the activator's control (CR 614.12, one
// simultaneous entry), planeswalkers with their printed loyalty and
// creatures as new objects. "Other planeswalkers" is the clause's
// ExcludeSource (CR 109.1): Sorin himself is not a legal pick, and
// another copy of him is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c780dda-ee76-482a-a610-53c321c4fdec",
		Name:         "Sorin, Lord of Innistrad",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 3,
		Emblem: &EmblemSpec{
			Label: "Sorin, Lord of Innistrad emblem",
			Text:  "Creatures you control get +1/+0.",
			Static: []game.StaticAbility{{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: creaturesTheEmblemsOwnerControls,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
				},
			}},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Create a 1/1 black Vampire creature token with lifelink.",
				Cost:  LoyaltyCost(1),
				Effect: Do(CreateToken{
					Template: TokenCard("1/1 black Vampire with lifelink"),
					N:        1,
				}),
			},
			{
				Label: "−2: You get an emblem with \"Creatures you control get +1/+0.\"",
				Cost:  LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−6: Destroy up to three target creatures and/or other planeswalkers. Return each card put into a graveyard this way to the battlefield under your control.",
				Cost:  LoyaltyCost(-6),
				Targets: Another(TargetPermanent("up to three target creatures and/or other planeswalkers",
					Or(Creature(), Planeswalker()))).WithCount(0, 3),
				Effect: sorinLordDestroyAndTake,
			},
		},
	})
}

// sorinLordDestroyAndTake is the −6: destroy every still-legal target
// together, then return, simultaneously and under the activator's
// control, the ones that were put into a graveyard.
func sorinLordDestroyAndTake(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var doomed []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			doomed = append(doomed, t.ID)
		}
	}
	if len(doomed) == 0 {
		return nil
	}
	return g.DestroyPermanentsKeyedForEffect(doomed, ctx.ItemTail(sorinLordReturn))
}

// sorinLordReturn returns the permanents destroyed this way, together and
// under the activator's control, keyed (#4253 slice 10a).
var sorinLordReturn = CardTail("then/sorin-lord-of-innistrad-return", func(ctx *Context, o game.TailOutcome) error {
	return ReturnFromGraveyardTogether{Targets: o.IDs, Controller: ctx.Controller()}.Apply(ctx)
})
