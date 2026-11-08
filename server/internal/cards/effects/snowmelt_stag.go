package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snowmelt Stag — Creature — Elemental Elk {3}{U}, 2/5 (Tarkir:
// Dragonstorm, #4463):
//
//	"Vigilance
//	 During your turn, this creature has base power and toughness 5/2.
//	 {5}{U}{U}: This creature can't be blocked this turn."
//
// Vigilance is a printed keyword. The middle line is Doc Ock, Sinister
// Scientist's layer 7b set on the source alone, gated on its controller
// being the active player (Alpharael, Dreaming Acolyte's "during your
// turn"), so counters and pumps still land on top of the 5/2 (CR
// 613.4b, 613.4c). The activated ability is Gearseeker Serpent's: the
// Stag can't be blocked only while it is still the permanent that
// activated, since one that left and came back in response is a new
// object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b25ccb64-e114-4d54-a07a-c8cd4560b511",
		Name:            "Snowmelt Stag",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Label:    "During your turn, this creature has base power and toughness 5/2.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7B_Set,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && isActivePlayer(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power, c.Toughness = 5, 2
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{5}{U}{U}: This creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{5}{U}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Snowmelt Stag — can't be blocked"}.Apply(NewContext(g, item))
			},
		}},
	})
}
