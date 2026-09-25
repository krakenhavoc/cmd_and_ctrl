package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gatekeeper of Malakir — {B}{B} 2/2 Vampire Warrior:
//
//	"Kicker {B}
//	 When this creature enters, if it was kicked, target player
//	 sacrifices a creature of their choice."
//
// The "if it was kicked" ETB (CR 702.33c) on a PERMANENT, and the
// second consumer of ADR 0073 §5's carried record after Wolfbriar
// Elemental. Two things it pins that Wolfbriar does not:
//
//   - The clause is an INTERVENING IF (CR 603.4), so it gates whether
//     the ability triggers at all rather than what it does. An
//     unkicked Gatekeeper puts nothing on the stack, and nobody gets
//     a window to respond to a trigger that is not happening.
//   - The trigger TARGETS, and a target is chosen at CR 603.3d — after
//     the harvester has already decided the ability triggered. So the
//     kicked read has to happen in AppliesTo, off the permanent,
//     before any of the targeting machinery runs.
//
// "Of their choice" is the rules content of the sacrifice: the target
// player picks, from their own creatures, through the same
// SacrificeChoice prompt Fleshbag Marauder's fan-out uses. It does
// not target the creature, so hexproof and shroud are irrelevant to
// it — only the PLAYER is targeted.
func init() {
	Register(Spec{
		OracleID:     "6781f8ae-2a86-4e3d-bc43-48809c9d6c26",
		Name:         "Gatekeeper of Malakir",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			Kicker("{B}"),
		},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				// CR 603.4: "if it was kicked" is checked here, when
				// the ability would trigger, and it reads the record
				// the resolution path carried onto this permanent.
				AppliesTo: AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return game.CardKickedTimes(*source) > 0
				}),
				Key:     "Gatekeeper of Malakir — kicked, target player sacrifices",
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					victim := uuid.Nil
					for _, t := range item.Targets {
						if t.Kind == game.TargetPlayer {
							victim = t.ID
							break
						}
					}
					if victim == uuid.Nil {
						return nil
					}
					var creatures []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == victim && c.IsCreature() {
							creatures = append(creatures, c.InstanceID)
						}
					}
					return SacrificeChoice{
						Player:     victim,
						Candidates: creatures,
						Question:   "Gatekeeper of Malakir — sacrifice a creature",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
