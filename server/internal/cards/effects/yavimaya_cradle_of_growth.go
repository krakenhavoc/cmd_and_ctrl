package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yavimaya, Cradle of Growth — Legendary Land:
//
//	"Each land is a Forest in addition to its other land types."
//
// Urborg, Tomb of Yawgmoth with a different land type, and named by
// AGENTS.md §7 as one of the cards the layer-dependency work (#668 /
// #669, ADR 0067) released. Everything Urborg's file says applies
// here: the sentence is one layer-4 static, the intrinsic mana
// ability is DERIVED from effective subtypes (CR 305.6) rather than
// declared, and "each land" means every land on the battlefield under
// every player's control — Yavimaya itself included, so Yavimaya
// alone taps for {G}.
//
// "In addition to its other land types" means nothing is lost, so the
// subtype is APPENDED rather than set: an Island still makes {U} and
// keeps its printed ability in slot 0, which is what stops the
// auto-tapper's first-ability choice changing under an untouched
// land.
//
// CR 613.8a: "each land" reads a type other layer-4 effects write, so
// this depends on any effect that MAKES something a land and applies
// after it whichever entered first. Against a Magus of the Moon, which
// is a set rather than an add, the two are independent and settle by
// timestamp — a Magus that arrived later silences this in layer 4 and
// every nonbasic is a Mountain; one that arrived first is itself made
// a Mountain Forest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8dd5f5af-d2d8-4356-8617-8381081b930c",
		Name:         "Yavimaya, Cradle of Growth",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{EachLandIsAlso("Forest")},
	})
}
