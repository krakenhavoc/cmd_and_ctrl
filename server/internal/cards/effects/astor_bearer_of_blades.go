package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Astor, Bearer of Blades — Legendary Creature — Human Warrior {2}{R}{W},
// 4/4 (Dominaria United):
//
//	"When Astor enters, look at the top seven cards of your library. You
//	 may reveal an Equipment or Vehicle card from among them and put it
//	 into your hand. Put the rest on the bottom of your library in a
//	 random order.
//	 Equipment you control have equip {1}.
//	 Vehicles you control have crew 1."
//
// The enter trigger is Contagious Vorrac's look-reveal-and-take over seven
// cards, with the rest to the bottom in a random order. The two statics are
// layer-6 ability grants (Puresteel Paladin's shape): each Equipment you
// control gains an equip {1} ability and each Vehicle you control a crew 1
// ability, in addition to what it prints (CR 702.6d, CR 702.122a).
//
// No simplification.
const (
	astorEquipGrant = "astor-bearer-of-blades/equip-one"
	astorCrewGrant  = "astor-bearer-of-blades/crew-one"
)

func init() {
	Register(Spec{
		OracleID:     "196962f0-6aff-4198-9e0c-eaf42aeaf147",
		Name:         "Astor, Bearer of Blades",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{Key: astorEquipGrant, Activated: []ActivatedAbility{EquipAbility("{1}")}, Text: "Equip {1}"},
			{Key: astorCrewGrant, Activated: []ActivatedAbility{{
				Label:   "Crew 1",
				Cost:    CrewCost(1),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Vehicle"),
			}}, Text: "Crew 1"},
		},
		Static: []game.StaticAbility{
			GrantAbilities(EquipmentYouControl, astorEquipGrant),
			GrantAbilities(AstorVehicleYouControl, astorCrewGrant),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Astor, Bearer of Blades — look at the top seven cards; you may reveal an Equipment or Vehicle card and put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 7),
						Match:    Or(OfSubtype("Equipment"), OfSubtype("Vehicle")),
						Max:      1,
						Optional: true,
						Label:    "Astor, Bearer of Blades — you may reveal an Equipment or Vehicle card and put it into your hand",
						Key:      AstorTake,
					}.Apply(ctx)
				}),
		},
	})
}

// astorTake is the take: the rest on the bottom in a random order. An
// on-disk key.
var AstorTake = RegisterTake("choose-cards/astor-bearer-of-blades-take",
	TakeDef{Reveal: true, Then: TakeRestOnBottomInRandomOrder})

// astorVehicleYouControl is "Vehicles you control": the source's
// controller's permanents with the Vehicle subtype, read post-layer.
func AstorVehicleYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.HasSubtype("Vehicle")
}
