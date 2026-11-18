package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "35553c2a-0150-49e4-81da-f356c6a17253",
		Name:         "Sanguine Statuette",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sanguine Statuette — create a Blood token",
				Do(CreateToken{Template: BloodToken(), N: 1})),
			Optional(On(game.EventSacrifice, vowB04YouSacrificedBlood,
				"Sanguine Statuette — it becomes a 3/3 Vampire artifact creature with haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if sourceIsNewObject(g, item) {
						return nil
					}
					if err := (BecomeCreatureUntilEOT{
						Subtypes: []string{"Vampire"}, SetPower: 3, SetToughness: 3,
						Label: "Sanguine Statuette — 3/3 Vampire artifact creature",
					}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"haste"}, Label: "Sanguine Statuette — haste"}.Apply(ctx)
				}),
				"Sanguine Statuette — have it become a 3/3 Vampire artifact creature with haste until end of turn?"),
		},
	})
}
