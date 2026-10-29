package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fíli the Pathfinder — Legendary Creature — Dwarf Scout {3}{W}, 2/2:
//
//	"Storied (If you control three or more artifacts, legendaries,
//	 and/or Sagas, you have an enduring story for the rest of the game.)
//	 As long as you have an enduring story, creatures you control get
//	 +1/+1.
//	 Whenever Fíli or another nontoken Dwarf you control enters, create a
//	 2/2 red Dwarf creature token."
//
// Storied is the engine's (CR 702.195a). The anthem is a layer 7c
// static read off the designation. The entry trigger is Fíli himself,
// token or not, or another Dwarf you control that is not a token, so the
// Dwarf tokens it makes don't make more.
//
// No simplification. The Hobbit (#4475).
func init() {
	Register(Spec{
		OracleID:        "c91c7d6b-a502-4b66-9509-3c1eda00e763",
		Name:            "Fíli the Pathfinder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordStoried},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			Label:     "As long as you have an enduring story, creatures you control get +1/+1.",
			AppliesTo: WhileEnduringStory(CreaturesYouControl),
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, filiOrAnotherNontokenDwarfEntered, "Fíli the Pathfinder — create a 2/2 red Dwarf",
				Do(CreateToken{Template: TokenCard("2/2 red Dwarf"), N: 1})),
		},
	})
}

// filiOrAnotherNontokenDwarfEntered is "Fíli or another nontoken Dwarf you
// control enters".
func filiOrAnotherNontokenDwarfEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && !IsToken(c) && c.HasSubtype(source.TextSubtype("Dwarf"))
}
