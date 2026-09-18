package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani, the Greathearted — Legendary Planeswalker — Ajani {2}{G}{W},
// starting loyalty 5 (EDHREC rank 4105):
//
//	"Creatures you control have vigilance.
//	 +1: You gain 3 life.
//	 −2: Put a +1/+1 counter on each creature you control and a
//	     loyalty counter on each other planeswalker you control."
//
// A four-mana walker with five loyalty, a passive that matters, and a
// minus that does not cost him the board. In a Selesnya counters deck
// the −2 is a Cathars' Crusade trigger for every creature at once and
// a free tick on every other walker you have; the vigilance static is
// what lets you cast him, attack with everything, and still hold the
// board back.
//
// THE STATIC IS A REAL LAYER 6 GRANT, not a trigger, so it is live
// the moment Ajani lands and it covers creatures that arrive later —
// including the tokens something else made this turn. It says
// "creatures YOU CONTROL" with no "other" (Ajani is not a creature
// anyway), and it stops the instant he leaves.
//
// THE −2 IS TWO INSTRUCTIONS AND THE SECOND SAYS "EACH OTHER". Ajani
// does not tick himself up — he has just paid two loyalty and gets
// nothing back — but every OTHER planeswalker you control does,
// which is the line that makes a superfriends deck play him. The
// creature half has no "other" and no exclusion: every creature you
// control gets a counter, including one that entered this turn.
//
// Both halves are snapshotted before the first counter lands, because
// adding counters triggers abilities that can move permanents in and
// out of the two sets being walked.
//
// The counters are placed by Ajani's controller, so a Doubling Season
// or a Hardened Scales on your side applies to the creature half and
// (Doubling Season) to the loyalty half — that falls out of going
// through the shared counter primitive rather than writing the field.
//
// No simplification: this walker has no ultimate to omit, which makes
// it one of the few in the catalog that ships complete.
func init() {
	Register(Spec{
		OracleID:     "f5d9be71-91d0-4166-ba58-cbbf5d490c40",
		Name:         "Ajani, the Greathearted",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 5,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = append(c.Abilities, "vigilance")
			},
		}},
		Activated: []ActivatedAbility{
			{
				Label:  "+1: You gain 3 life.",
				Cost:   LoyaltyCost(1),
				Effect: Do(GainLife{Amount: 3}),
			},
			{
				Label:  "−2: Put a +1/+1 counter on each creature you control and a loyalty counter on each other planeswalker you control.",
				Cost:   LoyaltyCost(-2),
				Effect: b39AjaniMinusTwo,
			},
		},
	})
}

// b39AjaniMinusTwo is the −2: a +1/+1 counter on every creature its
// controller controls, then a loyalty counter on every OTHER
// planeswalker they control. Both sets are snapshotted first, because
// a counter-placement trigger can change the battlefield underneath
// the walk.
func b39AjaniMinusTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var creatures, walkers []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller {
			continue
		}
		if c.IsCreature() {
			creatures = append(creatures, c.InstanceID)
		}
		if c.IsPlaneswalker() && c.InstanceID != item.SourceCardID {
			walkers = append(walkers, c.InstanceID)
		}
	}
	for _, id := range creatures {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, id := range walkers {
		if err := (AddCounter{Target: id, Kind: game.CounterLoyalty, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
