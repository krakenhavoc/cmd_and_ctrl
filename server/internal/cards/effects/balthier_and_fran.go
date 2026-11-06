package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Balthier and Fran — Legendary Creature — Human Rabbit {1}{R}{G}, 4/3
// (Final Fantasy, #4462):
//
//	"Reach
//	 Vehicles you control get +1/+1 and have reach and vigilance.
//	 Whenever a Vehicle crewed by Balthier and Fran this turn attacks, if
//	 it's the first combat phase of the turn, you may pay {1}{R}{G}. If
//	 you do, after this phase, there is an additional combat phase."
//
// The anthem is Cloudspire Captain's layer 7c modify plus two layer 6
// keyword grants, over every Vehicle its controller controls, crewed or
// not (Vehicle is an artifact type, which no text change reaches).
//
// "Crewed by Balthier and Fran this turn" reads the turn's crew records
// (ADR 0186): the attacking Vehicle, as the object it is now, has a crew
// payment this turn that tapped THIS Balthier and Fran object. A Vehicle
// it saddled, or one another creature crewed, does not count, nor does
// one crewed by an earlier Balthier and Fran that left and came back
// (CR 400.7). "If it's the first combat phase of the turn" is an
// intervening if (CR 603.4), asked as the Vehicle attacks and again as
// the trigger resolves. The payment is Hellkite Charger's MayPay, which
// holds the table in the declare attackers step until it is answered, and
// the extra combat is the shared ExtraCombatAfterThisPhase.
//
// No simplification.
func init() {
	vehiclesYouControl := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.HasSubtype("Vehicle")
	}
	Register(Spec{
		OracleID:        "570cb560-c246-4652-b230-ab34f966c78a",
		Name:            "Balthier and Fran",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{
			{
				Label:     "Vehicles you control get +1/+1.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: vehiclesYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
			KeywordGrant(vehiclesYouControl, "reach"),
			KeywordGrant(vehiclesYouControl, "vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, aVehicleCrewedByThisAttackedInTheFirstCombat,
				"Balthier and Fran — you may pay {1}{R}{G} for an additional combat phase",
				balthierAndFranMayPay),
		},
	})
}

// aVehicleCrewedByThisAttackedInTheFirstCombat is the trigger condition:
// the attacker is a Vehicle whose crew records this turn name the source
// object, and this is the turn's first combat phase.
func aVehicleCrewedByThisAttackedInTheFirstCombat(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || !IsFirstCombatPhase(g) {
		return false
	}
	vehicle, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !vehicle.HasSubtype("Vehicle") {
		return false
	}
	ref, ok := g.PermanentRefForEffect(ev.CardID)
	if !ok {
		return false
	}
	me := game.ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}
	for _, r := range g.CrewRecordsForEffect(ref, false) {
		if !r.Saddle && r.Creature == me {
			return true
		}
	}
	return false
}

// balthierAndFranMayPay re-checks the intervening if and offers the
// payment.
func balthierAndFranMayPay(g *game.Game, item *game.StackItem) error {
	if !IsFirstCombatPhase(g) {
		return nil
	}
	return MayPay{
		Chooser:    item.Controller,
		Cost:       "{1}{R}{G}",
		Question:   "Balthier and Fran — pay {1}{R}{G}? If you do, after this phase, there is an additional combat phase.",
		Then:       balthierAndFranExtraCombat,
		InThisStep: true,
	}.Apply(NewContext(g, item))
}

// balthierAndFranExtraCombat is the paid branch. An on-disk key.
var balthierAndFranExtraCombat = PayUnlessBranches("pay-unless/balthier-and-fran-extra-combat",
	func(ctx *Context, _ game.PayUnlessAnswer) error { return ExtraCombatAfterThisPhase().Apply(ctx) }, nil)
