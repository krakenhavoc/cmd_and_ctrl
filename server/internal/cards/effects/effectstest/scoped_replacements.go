package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scoped_replacements_test.go pins ADR 0041 phase 3 tier 3b for the
// replacement effects (#1497): Fog, a prevention shield, the Whip's and
// unearth's redirect and Cosmic Intervention are ScopedEffect records,
// so a table holding one is a restore point, the effect still works in
// the restored game, an undo rewinds a spent charge, and the redirect
// lasts as long as its object does (#1591).

// scopedReplacementKinds are the tier 3b replacement kinds.
var scopedReplacementKinds = map[game.ModKind]bool{
	game.ModPreventCombatDamage:     true,
	game.ModPreventDamage:           true,
	game.ModExileInsteadOfLeaving:   true,
	game.ModExileInsteadOfGraveyard: true,
}

// scopedReplacementCount is how many live records hold a replacement
// mod — what `len(g.TurnScopedReplacements)` used to answer.
func ScopedReplacementCount(g *game.Game) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, e := range g.ScopedEffects {
			for _, m := range e.Mods {
				if scopedReplacementKinds[m.Kind] {
					n++
					break
				}
			}
		}
	})
	return n
}
