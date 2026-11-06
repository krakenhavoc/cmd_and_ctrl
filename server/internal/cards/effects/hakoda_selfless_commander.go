package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hakoda, Selfless Commander — Legendary Creature — Human Warrior Ally
// {3}{W}, 3/5:
//
//	"Vigilance
//	 You may look at the top card of your library any time.
//	 You may cast Ally spells from the top of your library.
//	 Sacrifice Hakoda: Creatures you control get +0/+5 and gain
//	 indestructible until end of turn."
//
// The middle two lines are Elven Chorus's (LibraryTopVisible and a
// top-of-library cast permission) narrowed to Ally spells, Conspicuous
// Snoop's filter: costs and timing are paid and followed as usual (the
// ruling). The sacrifice ability is Dauntless Escort's with the +0/+5
// pump beside the indestructible, both locked to the creatures you
// control as it resolves (CR 611.2c); Hakoda is gone by then, so it gets
// neither.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:          "de3a469a-e38b-456b-994d-abec1b010e42",
		Name:              "Hakoda, Selfless Commander",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"vigilance"},
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{NonLandOnly: true, CreatureType: "Ally"},
				"Cast Ally spells from the top of your library (Hakoda, Selfless Commander)"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice Hakoda: Creatures you control get +0/+5 and gain indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive, Saves: &game.Saves{Matches: game.SweepCreatures}},
			Cost:    game.AbilityCost{SacrificeSelf: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (BoostUntilEOT{
					Match:     And(Creature(), YouControl()),
					Toughness: 5,
					Label:     "Hakoda, Selfless Commander — +0/+5",
				}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"indestructible"},
					Label:    "Hakoda, Selfless Commander — indestructible",
				}.Apply(ctx)
			},
		}},
	})
}
