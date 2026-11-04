package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vulture, Scheming Scavenger — Legendary Creature — Human Artificer Villain
// {5}{U/B}, 4/6 (Marvel's Spider-Man, #4460):
//
//	"Flying
//	 Whenever Vulture attacks, other Villains you control gain flying until
//	 end of turn."
//
// The set of Villains is fixed as the trigger resolves (CR 611.2c): every
// other Villain its controller then controls, attacking or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0409765a-bbb4-465e-99b3-3687bffc3cf9",
		Name:            "Vulture, Scheming Scavenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Vulture, Scheming Scavenger — other Villains you control gain flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					self, me := item.SourceCardID, item.Controller
					return GrantKeywordUntilEOT{
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.InstanceID != self && c.Controller == me && c.IsCreature() && c.HasSubtypeWord("Villain")
						},
						Keywords: []string{"flying"},
						Label:    "Vulture, Scheming Scavenger — flying",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
