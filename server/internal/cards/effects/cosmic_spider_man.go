package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cosmic Spider-Man — Legendary Creature — Spider Human Hero
// {W}{U}{B}{R}{G}, 5/5 (Marvel's Spider-Man):
//
//	"Flying, first strike, trample, lifelink, haste
//	 At the beginning of combat on your turn, other Spiders you control
//	 gain flying, first strike, trample, lifelink, and haste until end of
//	 turn."
//
// Iron Man's keyword grant to the other permanents the controller has at
// resolution (CR 611.2c: the set is fixed as the trigger resolves, so a
// Spider that arrives later gets nothing). "Other" is by identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0334a2f8-a74a-44c9-ba64-d9b7fabe160c",
		Name:            "Cosmic Spider-Man",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "trample", "lifelink", "haste"},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Cosmic Spider-Man — other Spiders you control gain flying, first strike, trample, lifelink, and haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					self, me := item.SourceCardID, item.Controller
					return GrantKeywordUntilEOT{
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.InstanceID != self && c.Controller == me && c.HasSubtypeWord("Spider")
						},
						Keywords: []string{"flying", "first strike", "trample", "lifelink", "haste"},
						Label:    "Cosmic Spider-Man — keywords until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
