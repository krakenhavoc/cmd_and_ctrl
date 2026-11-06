package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whirlwing Stormbrood // Dynamic Soar — Tarkir: Dragonstorm (#4463), an
// omen card (CR 720, #5175):
//
//	Whirlwing Stormbrood — Creature — Dragon {4}{U}, 4/3
//	  "Flash
//	   Flying
//	   You may cast sorcery spells and Dragon spells as though they had
//	   flash."
//	Dynamic Soar — Sorcery — Omen {2}{G}
//	  "Put three +1/+1 counters on target creature you control. (Then
//	   shuffle this card into its owner's library.)"
//
// The Dragon under the bare oracle ID, the Omen under "<oracle>#1"; the
// shuffle is the engine's (CR 720.3d, game/omen.go). The timing grant
// is two CastKindAsThoughFlash rules, Gandalf's sorcery spells and Blur
// of Heroism's creature-type filter with Dragon; a spell that is either
// gets flash.
//
// No simplification.

const whirlwingStormbroodOracleID = "4fbfc814-df10-4d61-ada9-c96cc8ee8d8c"

func init() {
	const grant = "You may cast sorcery spells and Dragon spells as though they had flash."
	Register(Spec{
		OracleID:        whirlwingStormbroodOracleID,
		Name:            "Whirlwing Stormbrood",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{SorceryOnly: true}, grant),
			CastKindAsThoughFlash(game.PermissionFilter{CreatureType: "Dragon"}, grant),
		},
	})
	Register(Spec{
		OracleID:     whirlwingStormbroodOracleID + "#1",
		Name:         "Dynamic Soar",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
		},
	})
}
