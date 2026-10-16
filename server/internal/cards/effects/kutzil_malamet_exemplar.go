package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kutzil, Malamet Exemplar — Legendary Creature — Cat Warrior {1}{G}{W},
// 3/3 (EDHREC rank 396):
//
//	"Your opponents can't cast spells during your turn.
//	 Whenever one or more creatures you control each with power greater
//	 than its base power deals combat damage to a player, draw a card."
//
// The first line is Grand Abolisher's cast half,
// OpponentsCantCastDuringYourTurn with no narrowing filter.
//
// The trigger is the "one or more creatures you control deal combat
// damage to a player" batch (once per player dealt damage in a combat
// damage step, CR 603.2c), counting only a creature whose power is
// greater than its base power as it deals the damage. Base power is
// CR 208.4b's: the power after characteristic-defining abilities and
// effects that set power, and before pumps, anthems and counters
// (Card.BasePower, recorded by the layer pass after layer 7b). So a
// creature with a +1/+1 counter or under an anthem counts, a creature
// whose power was only SET higher does not, and a 0/0 that grows from
// its own ability counts (the 2023-11-10 rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3233af43-7826-4895-b67e-03c6102c2cd5",
		Name:         "Kutzil, Malamet Exemplar",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			OpponentsCantCastDuringYourTurn("Your opponents can't cast spells during your turn.", nil),
		},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(
				func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.IsCreature() && c.PowerForComparison() > c.BasePower()
				},
				"Kutzil, Malamet Exemplar — draw a card",
				Do(DrawCards{N: 1})),
		},
	})
}
