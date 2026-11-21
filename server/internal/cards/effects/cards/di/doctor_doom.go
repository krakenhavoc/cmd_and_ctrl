package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doctor Doom — {4}{B}{B} Legendary Creature — Human Scientist Villain 3/3:
//
//	"When Doctor Doom enters, create two 3/3 colorless Robot Villain
//	 artifact creature tokens named Doombot.
//	 As long as you control an artifact creature or a Plan, Doctor Doom has
//	 indestructible.
//	 At the beginning of your end step, you draw a card and lose 1 life."
//
// The tokens are a table row (tokens_table.go) whose name is Doombot. The
// static grants indestructible to Doom alone while the check holds; it reads
// post-layer types (BattlefieldCardsForEffect), so an animated Vehicle counts
// as an artifact creature, and a Plan is any permanent with the Plan subtype.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e4fa592-71f8-4c71-8c91-78e6326cd736",
		Name:         "Doctor Doom",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && doctorDoomHasCover(g, source.Controller)
			}, "indestructible"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Doctor Doom — create two Doombot tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 colorless Robot Villain artifact named Doombot"), N: 2})),
			effects.AtYourEndStep("Doctor Doom — you draw a card and lose 1 life", effects.B36DrawAndLoseOne),
		},
	})
}

// doctorDoomHasCover is "you control an artifact creature or a Plan".
func doctorDoomHasCover(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && ((c.IsCreature() && c.IsArtifact()) || c.HasSubtype("Plan")) {
			return true
		}
	}
	return false
}
