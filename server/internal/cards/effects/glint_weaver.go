package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glint Weaver — Creature — Spider {5}{G}{G}, 3/3 (#1658, unblocked
// by #1656):
//
//	"Reach
//	 When this creature enters, distribute three +1/+1 counters
//	 among one, two, or three target creatures, then you gain life
//	 equal to the greatest toughness among creatures you control."
//
// Armament Dragon's division WITHOUT "you control" — any one, two, or
// three creatures on the battlefield are legal targets, printed as is.
// The life gain reads the board AFTER the counters land (the trigger's
// own "then"), so a Glint Weaver that just grew itself counts its own
// new toughness.
func init() {
	Register(Spec{
		OracleID:        "eebe3fd6-afd0-4bb3-9e8f-c737497dae33",
		Name:            "Glint Weaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets: TargetCreature("one, two, or three target creatures").
				WithCount(1, 3).Dividing(Divide(3)),
			Key: "Glint Weaver — distribute three +1/+1 counters, then gain life",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := PutDividedCounters(ctx, game.CounterPlusOne); err != nil {
					return err
				}
				n := glintWeaverGreatestToughnessControlledBy(g, item.Controller)
				return GainLife{Player: item.Controller, Amount: n}.Apply(ctx)
			},
		}},
	})
}

// glintWeaverGreatestToughnessControlledBy is "the greatest toughness
// among creatures you control", read after the counters have already
// landed so a Glint Weaver that grew itself counts its new toughness.
// Zero when the controller has no creatures.
func glintWeaverGreatestToughnessControlledBy(g *game.Game, controller uuid.UUID) int {
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.IsCreature() {
			continue
		}
		if t := c.CurrentToughness(); t > best {
			best = t
		}
	}
	return best
}
