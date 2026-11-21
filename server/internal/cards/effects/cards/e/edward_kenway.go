package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Edward Kenway — Legendary Creature — Human Assassin Pirate {2}{U}{B}{R},
// 5/5 (Assassin's Creed, #5948):
//
//	"At the beginning of your end step, create a Treasure token for each
//	 tapped Assassin, Pirate, and/or Vehicle you control.
//	 Whenever a Vehicle you control deals combat damage to a player, look
//	 at the top card of that player's library, then exile it face down. You
//	 may play that card for as long as it remains exiled."
//
// The count is read as the end-step trigger resolves: every tapped permanent
// you control that is an Assassin, a Pirate or a Vehicle, each once however
// many of the three it is (Edward himself, if tapped, counts once). The
// second trigger is Gonti, Night Minister's face-down exile with
// WhileExiled, without the any-type mana rider (Edward prints none), granted
// to the Vehicle's controller, who is Edward's controller; one trigger per
// Vehicle that connects. Whether the dealer was a Vehicle is read as it
// dealt the damage, so a crewed Vehicle still counts.
//
// No simplification.
func init() {
	const exile = "Edward Kenway — look at the top card of that player's library, then exile it face down; you may play it for as long as it remains exiled"
	effects.Register(effects.Spec{
		OracleID:     "b2fe6d17-d045-45c6-9ca2-21e73d2e38d2",
		Name:         "Edward Kenway",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Edward Kenway — create a Treasure token for each tapped Assassin, Pirate, and/or Vehicle you control",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					kinds := effects.Or(effects.HasSubtype("Assassin"), effects.HasSubtype("Pirate"), effects.HasSubtype("Vehicle"))
					for i := range g.Battlefield.Cards {
						c := g.Battlefield.Cards[i]
						if c.Controller == item.Controller && c.Tapped && kinds(g, item.Controller, c) {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TreasureToken(), N: n}.Apply(effects.NewContext(g, item))
				}),
			{
				Watches:   []game.EventKind{game.EventDealDamage},
				AppliesTo: effects.GapB37VehicleYouControlDealtCombatDamageToAPlayer,
				Key:       exile,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopWithPermission{
						From:        item.Trigger.Event.Target,
						GrantTo:     item.Controller,
						N:           1,
						FaceDown:    true,
						WhileExiled: true,
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
