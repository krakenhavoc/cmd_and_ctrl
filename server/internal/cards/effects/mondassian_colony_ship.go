package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mondassian Colony Ship — Plane — Spacecraft (Doctor Who, #4870):
//
//	"Whenever a creature attacks, it gets +1/+1 until end of turn for
//	 each other creature its controller controls that shares a creature
//	 type with it.
//	 Whenever chaos ensues, turn target creature face down. It becomes a
//	 2/2 Cyberman artifact creature."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first ability triggers for every attacking creature, anyone's; the
// count is taken as it resolves, over the other creatures the attacker's
// controller controls that share a creature type with it, changelings and
// "every creature type" included (game.SharesCreatureType).
// The chaos ability (CR 311.7) is Cyber Conversion's: the creature is
// turned face down as a 2/2 Cyberman artifact creature with no name and
// no color (CR 708.2, the rulings), and stops being one if it is turned
// face up.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cc34a5b-550d-4d3e-b3e5-4d4630649882",
		Name:         "Mondassian Colony Ship",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, aCreatureAttacked, "Mondassian Colony Ship — it gets +1/+1 until end of turn for each other creature its controller controls that shares a creature type with it",
				mondassianColonyShipPump),
			Targeting(WheneverChaosEnsues("Mondassian Colony Ship — turn target creature face down; it becomes a 2/2 Cyberman artifact creature",
				turnSingleTargetFaceDownAs(CybermanBody)), TargetCreature("target creature")),
		},
	})
}

// aCreatureAttacked is "whenever a creature attacks": one trigger per
// attacking creature, whoever controls it.
func aCreatureAttacked(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventAttack && ev.CardID != uuid.Nil
}

func mondassianColonyShipPump(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	attacker, ok := battlefieldCardForEffect(g, item.Trigger.Event.CardID)
	if !ok || !attacker.IsCreature() {
		return nil
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == attacker.InstanceID || c.Controller != attacker.Controller || !c.IsCreature() {
			continue
		}
		if game.SharesCreatureType(&attacker, &c) {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return BoostUntilEOT{Target: attacker.InstanceID, Power: n, Toughness: n,
		Label: "Mondassian Colony Ship — +1/+1 for each creature sharing a type"}.Apply(NewContext(g, item))
}
