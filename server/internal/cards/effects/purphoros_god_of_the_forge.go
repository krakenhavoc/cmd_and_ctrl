package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Purphoros, God of the Forge — Legendary Enchantment Creature — God
// {3}{R}, 6/5:
//
//	"Indestructible
//	 As long as your devotion to red is less than five, Purphoros isn't
//	 a creature.
//	 Whenever another creature you control enters, Purphoros deals 2
//	 damage to each opponent.
//	 {2}{R}: Creatures you control get +1/+0 until end of turn."
//
// The Theros God shape (Thassa, Erebos): indestructible rides
// PrintedKeywords and the devotion gate is the shared godUnlessDevotion.
// The enters trigger is the "pings the table" body every such card
// shares, and the pump locks its set as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4fdbbec2-e921-4b63-958d-f9ba1e417197",
		Name:            "Purphoros, God of the Forge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{godUnlessDevotion("R")},
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl(
				"Purphoros, God of the Forge — 2 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					return damageToEachOpponent(g, item, 2)
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{R}: Creatures you control get +1/+0 until end of turn.",
			Cost:  ManaCost("{2}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Match: And(Creature(), YouControl()),
					Power: 1,
					Label: "Purphoros, God of the Forge — creatures you control get +1/+0",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
