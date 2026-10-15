package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d3bdfb25-dfb2-4df4-b502-ef0909a39f86",
		Name:         "Cid, Freeflier Pilot",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Equipment and Vehicle spells you cast cost {1} less to cast.",
				YourSpell(), equipmentOrVehicleSpell()),
		},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.IsCreature() && isActivePlayer(g, source.Controller)
			}, "flying"),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Return target Equipment or Vehicle card from your graveyard to your hand.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCardInGraveyard("target Equipment or Vehicle card in your graveyard",
				Or(OfSubtype("Equipment"), OfSubtype("Vehicle")), YouOwn()),
			Effect: returnTargetCardToHand,
		}},
	})
}

func equipmentOrVehicleSpell() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.HasSubtype("Equipment") || q.Card.HasSubtype("Vehicle")
	}
}
