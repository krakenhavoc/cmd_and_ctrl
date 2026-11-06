package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Messenger Hawk — Creature — Bird Scout {2}{U/B}, 1/2 (Avatar: The Last
// Airbender, #4459):
//
//	"Flying
//	 When this creature enters, create a Clue token.
//	 This creature gets +2/+0 as long as you've drawn two or more cards
//	 this turn."
//
// The Clue is created, not investigated (the ruling). The pump is a
// layer 7c modify on itself read off its controller's draw tally for the
// turn (Foggy Swamp Hunters' count), so draws before the Hawk entered
// count too; the tally is not on the battlefield, so the static declares
// DependsOnHandSize, which a draw drops the cached layers for.
//
// No simplification.
func init() {
	pump := game.StaticAbility{
		Label:    "Gets +2/+0 as long as you've drawn two or more cards this turn.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && g.TurnTallyFor(source.Controller).CardsDrawn >= 2
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += 2
		},
		DependsOnHandSize: true,
	}
	Register(Spec{
		OracleID:        "4029e884-8154-42b3-ae29-3cb98374a294",
		Name:            "Messenger Hawk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{pump},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Messenger Hawk — create a Clue token", Do(CreateToken{Template: ClueToken(), N: 1})),
		},
	})
}
