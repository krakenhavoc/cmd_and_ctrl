package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Champion of the Weird — Creature — Goblin Berserker {3}{B}, 5/5
// (Lorwyn Eclipsed, #4458):
//
//	"As an additional cost to cast this spell, behold a Goblin and exile
//	 it. (Exile a Goblin you control or a Goblin card from your hand.)
//	 Pay 1 life, Blight 2: Target opponent blights 2. Activate only as a
//	 sorcery.
//	 When this creature leaves the battlefield, return the exiled card to
//	 its owner's hand."
//
// The cost and the leave trigger are Champions of the Shoal's
// (BeholdAndExileCost, ReturnTheExiledCardToHand): a countered Champion
// leaves the beheld card in exile (the ruling). The activation pays 1
// life and blights 2 as its cost (BlightN: both -1/-1 counters on one
// creature you control, chosen at activation, and refused with no
// creature that can take them), then the targeted opponent blights 2 as
// the effect (Blight: they choose one creature they control; with none
// they don't blight, CR 701.68b). An opponent who is no longer a legal
// target blights nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "1726f9cd-6f18-4509-b95b-bca82ba795a4",
		Name:           "Champion of the Weird",
		Completeness:   CompletenessFull,
		AdditionalCost: BeholdAndExileCost("a", "Goblin"),
		Activated: []ActivatedAbility{{
			Label:        "Pay 1 life, Blight 2: Target opponent blights 2. Activate only as a sorcery.",
			Cost:         Plus(PayLife(1), BlightN(2)),
			Targets:      TargetPlayer("target opponent", Opponent()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				opp, ok := firstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return Blight{N: 2, Players: []uuid.UUID{opp}}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			ReturnTheExiledCardToHand("Champion of the Weird"),
		},
	})
}
