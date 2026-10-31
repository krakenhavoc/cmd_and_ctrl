package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Thanos-Copter — {6} Legendary Artifact — Vehicle 5/5 (Marvel Super
// Heroes Commander, #4455):
//
//	"Flying, haste
//	 When The Thanos-Copter enters, Vehicles you control become artifact
//	 creatures until end of turn.
//	 Whenever a Vehicle you control deals combat damage to a player, draw a
//	 card.
//	 Crew 2"
//
// The enter trigger animates every Vehicle its controller controls as it
// resolves (itself included, the crew effect's BecomeCreatureUntilEOT aimed
// at each one); a Vehicle that arrives later this turn is not one of them
// (CR 611.2c). The draw is Setzer's per-Vehicle combat-damage condition: one
// draw for each Vehicle that connects, whoever crewed it. Crew is Dependable
// Quinjet's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e58e94dc-9602-48db-9dd9-9d3051c33a76",
		Name:            "The Thanos-Copter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("The Thanos-Copter — Vehicles you control become artifact creatures until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					vehicle := ctx.TextSubtype("Vehicle")
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != item.Controller || !c.HasSubtype(vehicle) {
							continue
						}
						if err := (BecomeCreatureUntilEOT{Target: c.InstanceID, Label: "The Thanos-Copter — becomes an artifact creature"}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
			On(game.EventDealDamage, setzerVehicleDealtCombatDamageToAPlayer,
				"The Thanos-Copter — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("The Thanos-Copter"),
		}},
	})
}
