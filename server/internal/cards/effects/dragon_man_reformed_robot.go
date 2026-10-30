package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragon Man, Reformed Robot — Legendary Artifact Creature — Dragon Robot
// {2}{W}{U}, */5 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Dragon Man's power is equal to the greatest mana value among
//	 noncreature permanents you control and noncreature cards in your
//	 graveyard.
//	 You may cast this card from your graveyard by discarding a card in
//	 addition to paying its other costs."
//
// The power is a layer 7a characteristic-defining ability (CR 604.3,
// 613.4a) read live from the battlefield and graveyard; Dragon Man is a
// creature, so he never counts himself. The graveyard cast is Alien
// Symbiosis's Retrace-model offer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c03ca2e6-fcce-4018-952f-1c451ab92cfd",
		Name:             "Dragon Man, Reformed Robot",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{msc17CastFromGraveyardByDiscarding("{2}{W}{U}")},
		Static: []game.StaticAbility{{
			Label:     "Power is equal to the greatest mana value among noncreature permanents you control and noncreature cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.Controller == source.Controller && !p.IsCreature() && p.ManaValue() > n {
						n = p.ManaValue()
					}
				}
				if me := g.PlayerByIDForEffect(source.Controller); me != nil && me.Graveyard != nil {
					for _, gc := range me.Graveyard.Cards {
						if !gc.IsCreature() && gc.ManaValue() > n {
							n = gc.ManaValue()
						}
					}
				}
				c.Power = n
			},
		}},
	})
}
