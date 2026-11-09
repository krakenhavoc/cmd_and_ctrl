package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hedron Fields of Agadeem — Plane — Zendikar (Commander Masters, #4870):
//
//	"Creatures with power 7 or greater can't attack or block.
//	 Whenever chaos ensues, create a 7/7 colorless Eldrazi creature token
//	 with annihilator 1. (Whenever it attacks, defending player sacrifices
//	 a permanent of their choice.)"
//
// A plane (ADR 0216 §6). The static is a restriction over every creature
// whose FINISHED power is 7 or greater (CR 613.11): the layer-6 static
// records the ceiling and game/power_floor.go folds it into CantAttack |
// CantBlock once every layer has run, so an anthem that lifts a creature to
// 7 takes it out of combat and a -N/-N brings it back, per the ruling (the
// power is checked as attackers or blockers are declared, so a creature
// already declared keeps attacking). The chaos ability (CR 311.7) makes the
// planar controller's 7/7 Eldrazi token with annihilator 1; the token can't
// attack itself while this plane is face up, which is the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f7daaf2-3206-4f52-928a-f5116b782367",
		Name:         "Hedron Fields of Agadeem",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Creatures with power 7 or greater can't attack or block.",
			Layer:     game.Layer6Ability,
			AppliesTo: allCreatures,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if c.AttackBlockPowerCeiling == 0 || c.AttackBlockPowerCeiling > 7 {
					c.AttackBlockPowerCeiling = 7
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Hedron Fields of Agadeem — create a 7/7 colorless Eldrazi creature token with annihilator 1",
				Do(CreateToken{Template: TokenCard("7/7 colorless Eldrazi with annihilator 1"), N: 1})),
		},
	})
}
