package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanguine Statuette — Artifact {1}{R} (Innistrad: Crimson Vow, #5495):
//
//	"When this artifact enters, create a Blood token.
//	 Whenever you sacrifice a Blood token, you may have this artifact
//	 become a 3/3 Vampire artifact creature with haste until end of turn."
//
// The animation is Iron Suitcase's: a layer-4 type add (Vampire, on top of
// Artifact and Creature), a layer-7b 3/3 and haste, all until end of turn
// and pinned to this object (CR 400.7). The "may" is asked as the trigger
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35553c2a-0150-49e4-81da-f356c6a17253",
		Name:         "Sanguine Statuette",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sanguine Statuette — create a Blood token",
				effects.Do(effects.CreateToken{Template: effects.BloodToken(), N: 1})),
			effects.Optional(effects.On(game.EventSacrifice, effects.VowB04YouSacrificedBlood,
				"Sanguine Statuette — it becomes a 3/3 Vampire artifact creature with haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if effects.SourceIsNewObject(g, item) {
						return nil
					}
					if err := (effects.BecomeCreatureUntilEOT{
						Subtypes: []string{"Vampire"}, SetPower: 3, SetToughness: 3,
						Label: "Sanguine Statuette — 3/3 Vampire artifact creature",
					}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"haste"}, Label: "Sanguine Statuette — haste"}.Apply(ctx)
				}),
				"Sanguine Statuette — have it become a 3/3 Vampire artifact creature with haste until end of turn?"),
		},
	})
}
