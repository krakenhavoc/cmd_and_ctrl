package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hexproof_from_cards_test.go — #3172, the proof cards for "hexproof from
// [quality]" (ADR 0038's amendment of 2026-10-13): Eradicator Valkyrie,
// Breaker of Creation, Dragonfire Blade and Veil of Summer. The engine
// half (the grammar, the choke point, the player half) is in
// game/hexproof_from_test.go; these prove each card reaches it.

var (
	SrcBlack      = &game.Characteristic{Colors: []string{"B"}, Types: []string{"Instant"}}
	SrcRed        = &game.Characteristic{Colors: []string{"R"}, Types: []string{"Instant"}}
	SrcBlue       = &game.Characteristic{Colors: []string{"U"}, Types: []string{"Instant"}}
	SrcGold       = &game.Characteristic{Colors: []string{"U", "B"}, Types: []string{"Instant"}}
	SrcColourless = &game.Characteristic{Types: []string{"Artifact"}}
	SrcWalker     = &game.Characteristic{Colors: []string{"W"}, Types: []string{"Planeswalker"}}
)

// creatureTargetableFrom is creatureTargetableBy for a source with
// known characteristics: hexproof from [quality] asks the source, not
// just whose it is.
func CreatureTargetableFrom(g *game.Game, by, id uuid.UUID, src *game.Characteristic) bool {
	var ok bool
	g.WithWriteLock(func() {
		ts := game.TargetSource{Controller: by, Snapshot: src}
		for _, c := range g.LegalTargetsForEffect(ts, effects.TargetCreature("target creature")).Cards {
			if c == id {
				ok = true
			}
		}
	})
	return ok
}
