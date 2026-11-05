package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cogwork Wrestler — Artifact Creature — Gnome {U}, 1/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Flash
//	 When this creature enters, target creature an opponent controls gets
//	 -2/-0 until end of turn."
//
// Flash is a printed keyword (CR 702.8a). The enters trigger targets as
// it goes on the stack and is removed with no legal target (CR 603.3d);
// the shrink is locked to the target at resolution (CR 611.2c) and ends
// in the cleanup step (CR 514.2). A target no longer legal is left alone
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "db27e3f0-5750-4c4b-9da9-70c3ab627f9c",
		Name:            "Cogwork Wrestler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Cogwork Wrestler — target creature an opponent controls gets -2/-0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return BoostUntilEOT{
						Target:    id,
						Power:     -2,
						Toughness: 0,
						Label:     "Cogwork Wrestler — -2/-0 until end of turn",
					}.Apply(ctx)
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
				ForTargets(HarmUntilEOT(0, -2, 0))),
		},
	})
}
