package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcade Gannon — Legendary Creature — Human Doctor {2}{W}{U}, 2/3
// (Fallout):
//
//	"{T}: Draw a card, then discard a card. Put a quest counter on Arcade
//	 Gannon.
//	 For Auld Lang Syne — Once during each of your turns, you may cast an
//	 artifact or Human spell from your graveyard with mana value less than
//	 or equal to the number of quest counters on Arcade Gannon."
//
// The loot is the draw's continuation (the discard is chosen after the draw);
// the counter goes on in the same resolution, and nothing can act between
// the two. The permission is Lurrus of the Dream-Den's: one use a turn
// (a one-type budget), my turn only, from my graveyard, with the mana value
// cap read from the quest counters on Arcade as the spell is cast
// (MaxSpellManaValueCountersOnSource, derived like Maralen's count) and the
// "artifact or Human" filter (PermissionFilter.ArtifactOrSubtype). It lives
// only while Arcade is on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e001d7e6-71c9-42fa-9d5c-9455da66e4ec",
		Name:         "Arcade Gannon",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Discards: 1},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Draw a card, then discard a card. Put a quest counter on Arcade Gannon.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := drawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, ""); err != nil {
					return err
				}
				return AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}.Apply(NewContext(g, item))
			},
		}},
		CastPermissions: []game.CastPermission{{
			Zone:                              game.ZoneGraveyard,
			CastOnly:                          true,
			Filter:                            game.PermissionFilter{ArtifactOrSubtype: "Human"},
			MaxSpellManaValueCountersOnSource: "quest",
			PerType:                           []string{game.PermissionTypePermanent},
			Timing:                            game.TimingYourTurnOnly,
			Label:                             "Once during each of your turns, cast an artifact or Human spell from your graveyard with mana value less than or equal to the number of quest counters on Arcade Gannon (For Auld Lang Syne)",
		}},
	})
}
