package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nadaar, Selfless Paladin — Legendary Creature — Dragon Knight {2}{W},
// 3/3:
//
//	"Vigilance
//	 Whenever Nadaar enters or attacks, venture into the dungeon.
//	 Other creatures you control get +1/+1 as long as you've completed a
//	 dungeon."
//
// One ability with two trigger conditions (Sun Titan's shape), and a
// layer 7c anthem gated on Player.CompletedDungeons.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0e3f9345-d986-4d60-a343-45598ab59aff",
		Name:            "Nadaar, Selfless Paladin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks(ventureLabel("Nadaar, Selfless Paladin"), Venture),
		},
		Static: []game.StaticAbility{
			PumpWhileCompletedADungeon(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.InstanceID != source.InstanceID
			}, 1, 1),
		},
	})
}
