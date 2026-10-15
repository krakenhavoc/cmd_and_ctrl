package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Caldaia Guardian — Creature — Human Soldier {3}{G}, 4/3:
//
//	"Whenever this creature or another creature you control with mana
//	 value 4 or greater dies, create two 1/1 green and white Citizen
//	 creature tokens.
//	 Blitz {2}{G}"
//
// Blitz (CR 702.152, #3772). The mana-value clause belongs to "another
// creature"; the Guardian itself is mana value 4 anyway. The dead
// creature's controller is the one it had as it died (#1682) and its
// mana value is read off the card in the graveyard, where a nontoken
// creature keeps its printed mana cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "fd55d686-510e-4b6c-b8eb-458ee7c19fff",
		Name:             "Caldaia Guardian",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Blitz("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, caldaiaGuardianDied, "Caldaia Guardian — create two 1/1 Citizens",
				Do(CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 2})),
		},
	})
}

// caldaiaGuardianDied is "this creature or another creature you control
// with mana value 4 or greater dies".
func caldaiaGuardianDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if cardDied(ev, source) {
		return true
	}
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	return ok && leftUnderControlOf(ev, dead) == source.Controller && dead.ManaValue() >= 4
}
