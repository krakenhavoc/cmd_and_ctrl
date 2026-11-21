package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Obscuring Haze — Instant {2}{G}:
//
//	"If you control a commander, you may cast this spell without paying
//	 its mana cost.
//	 Prevent all damage that would be dealt this turn by creatures your
//	 opponents control."
//
// The free cast is the Commander 2020 cycle's alternative cost
// (FreeIfYouControlCommander: any commander permanent you control, the
// ruling). The shield is Thwart the Enemy's: #2026's controller test,
// read as each creature would deal damage (CR 609.7b).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0a31db0-69db-405c-99a3-945617900c54",
		Name:         "Obscuring Haze",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.FreeIfYouControlCommander("Cast without paying its mana cost (you control a commander)"),
		},
		OnResolve: effects.SourceShieldSpell(effects.PreventDamageFromSource{Protect: effects.ShieldAnything, Queries: effects.CreatureSources(),
			Filter: game.DamageSourceFilter{Controller: game.SourceControllerOpponents}}),
	})
}
