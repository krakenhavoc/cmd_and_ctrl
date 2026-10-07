package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Robobrain War Mind — Artifact Creature — Robot {3}{U}, */5:
//
//	"Robobrain War Mind's power is equal to the number of cards in your
//	 hand.
//	 When this creature enters, you get an amount of {E} (energy
//	 counters) equal to the number of artifact creatures you control.
//	 Whenever this creature attacks, you may pay {E}{E}{E}. If you do,
//	 draw a card."
//
// ADR 0129 §3 (#1995). The power is a characteristic-defining ability
// (CR 604.3, layer 7a), Psychosis Crawler's shape with the toughness
// left printed. The enters count is read as the trigger resolves, the
// War Mind included when it is still there. The attack payment is made
// as the trigger resolves (CR 118.12).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f9390ec-08b5-4761-8cb1-5d60190b9661",
		Name:         "Robobrain War Mind",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:             game.Layer7PT,
			SubLayer:          game.SubLayer7A_CDA,
			DependsOnHandSize: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Hand != nil {
					n = p.Hand.Size()
				}
				c.Power = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Robobrain War Mind — you get {E} for each artifact creature you control",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.Battlefield.Cards {
						if c.Controller == item.Controller && c.IsCreature() && c.IsArtifact() {
							n++
						}
					}
					return GetEnergy{N: n}.Apply(NewContext(g, item))
				}),
			whenThisAttacksMayPayEnergy("Robobrain War Mind", 3, "draw a card", Do(DrawCards{N: 1})),
		},
	})
}
