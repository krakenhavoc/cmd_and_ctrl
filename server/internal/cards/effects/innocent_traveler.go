package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// innocentTravelerOracle is shared by both faces: the front registers
// under it, the back under innocentTravelerOracle + "#1".
const innocentTravelerOracle = "7950a7ff-7c2e-40fe-b64a-324f4fbcf528"

// Innocent Traveler // Malicious Invader — a transforming creature:
//
//	Innocent Traveler {2}{B}{B}, Creature — Human, 1/3
//	  "At the beginning of your upkeep, any opponent may sacrifice a
//	   creature of their choice. If no one does, transform this
//	   creature."
//
//	Malicious Invader, Creature — Vampire, 3/3
//	  "Flying
//	   This creature gets +2/+0 as long as an opponent controls a
//	   Human."
//
// ADR 0150 (#1903): "any opponent" skips the Traveler's controller, so
// the sacrifice is offered to each opponent in APNAP order from the
// active player (the controller, whose upkeep it is) until one makes
// it. The payer picks the creature, and an opponent with no creature
// can't pay (CR 118.3). The offer holds the upkeep until it is
// answered. If nobody pays, the Traveler turns over in place
// (TransformThis); one that has already transformed since the trigger
// went on the stack does not (CR 701.27f).
//
// No simplification.
func init() {
	// "Human" is a word of the Invader's text; OpponentControlsAtLeast
	// lends it (CR 612.2, #3984).
	human := func(c game.Card) bool { return c.HasSubtypeWord("Human") }
	opponentHasHuman := OpponentControlsAtLeast(1, human)
	Register(Spec{
		OracleID:     innocentTravelerOracle,
		Name:         "Innocent Traveler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Innocent Traveler — any opponent may sacrifice a creature, or it transforms",
				func(g *game.Game, item *game.StackItem) error {
					return AnyPlayerPays{
						Pay:       SacrificePayment(1, "creature", "creatures", QueryType("creature")),
						Opponents: true,
						Question:  "Innocent Traveler — sacrifice a creature to stop it transforming?",
						OnNobody: func(ctx *Context) error {
							return TransformThis{}.Apply(ctx)
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:        innocentTravelerOracle + "#1",
		Name:            "Malicious Invader",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:    "This creature gets +2/+0 as long as an opponent controls a Human.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && opponentHasHuman(g, source.Controller, source.InstanceID)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
			},
		}},
	})
}
