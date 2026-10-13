package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blossoming Tortoise — Creature — Turtle {2}{G}{G}, 3/3 (EDHREC rank
// 3278):
//
//	"Whenever this creature enters or attacks, mill three cards, then
//	 return a land card from your graveyard to the battlefield tapped.
//	 Activated abilities of lands you control cost {1} less to
//	 activate.
//	 Land creatures you control get +1/+1."
//
// The lands deck's engine: a land back from the graveyard on entry
// and on every attack. The trigger is ONE printed ability with two
// trigger conditions (b21SelfEnteredOrAttacked) and Teval's body
// (b25MillThreeThenReturnLandTapped).
//
// "Then return a land card from your graveyard" is asked as the
// trigger resolves (CR 608.2c), over the lands in the graveyard AFTER
// the mill, so a land the trigger itself milled can be the one that
// comes back. The printed return is not optional, so the pick has a
// floor of one whenever the graveyard holds a land. Nothing is
// targeted, so the trigger always mills.
//
// The anthem is a layer 7c modify over land creatures the
// controller controls — post-layer types, so an animated land
// counts.
//
// The land enters tapped: "tapped" rides its entry event (#1872), so
// nothing sees an untapped land arrive or a land become tapped.
//
// The discount is the activation cost pass Boom Scholar and Sam,
// Loyal Attendant use (ActivationCostsLess, #1184): the ability's source
// is a land its controller controls, and generic mana is all it takes
// off (CR 601.2f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e1bf23b-7fb0-45ff-8544-fce9fa3eba00",
		Name:         "Blossoming Tortoise",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			ActivationCostsLess(1,
				"Activated abilities of lands you control cost {1} less to activate.",
				ofALandYouControl()),
		},
		Static: []game.StaticAbility{{
			Label:     "Land creatures you control get +1/+1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: b31LandCreaturesYouControl,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventAttack}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b21SelfEnteredOrAttacked(ev, source)
			}, b31BlossomingTortoiseLabel, b25MillThreeThenReturnLandTappedAlways),
		},
	})
}

// ofALandYouControl — "activated abilities of lands you control": the
// ability's source is a land under the modifier's controller.
func ofALandYouControl() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.Controller == q.Source.Controller && q.Card.IsLand()
	}
}
