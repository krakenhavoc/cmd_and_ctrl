package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cid, Freeflier Pilot — Legendary Creature — Human Warrior Pilot {1}{W},
// 2/2 (#3656):
//
//	"Equipment and Vehicle spells you cast cost {1} less to cast.
//	 Jump — During your turn, Cid has flying.
//	 {2}, {T}: Return target Equipment or Vehicle card from your
//	 graveyard to your hand."
//
// Danitha's cost reducer over Equipment and Vehicle spells (generic mana
// only, floored at zero). Jump is a layer 6 flying grant on Cid himself
// that is on only while it is his controller's turn, re-read every
// recompute. The graveyard return is Archaeomancer's targeted return as
// an activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3bdfb25-dfb2-4df4-b502-ef0909a39f86",
		Name:         "Cid, Freeflier Pilot",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Equipment and Vehicle spells you cast cost {1} less to cast.",
				effects.YourSpell(), equipmentOrVehicleSpell()),
		},
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.IsCreature() && effects.IsActivePlayer(g, source.Controller)
			}, "flying"),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, {T}: Return target Equipment or Vehicle card from your graveyard to your hand.",
			Cost:  effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target Equipment or Vehicle card in your graveyard",
				effects.Or(effects.OfSubtype("Equipment"), effects.OfSubtype("Vehicle")), effects.YouOwn()),
			Effect: effects.ReturnTargetCardToHand,
		}},
	})
}

func equipmentOrVehicleSpell() effects.CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.HasSubtype("Equipment") || q.Card.HasSubtype("Vehicle")
	}
}
