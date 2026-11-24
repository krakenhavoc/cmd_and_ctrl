package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rout — Sorcery {3}{W}{W} (Commander 2021 / The List, S82 OLD b058, #5504):
//
//	"You may cast this spell as though it had flash if you pay {2} more to
//	 cast it. (You may cast it any time you could cast an instant.)
//	 Destroy all creatures. They can't be regenerated."
//
// Wrath of God's sweep (regeneration refused, CR 701.19c) with the optional
// {2} that opens the instant-speed window for that cast (FlashForMore, CR
// 601.3c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "d69aa624-23d1-406a-b7e0-9140a6f57929",
		Name:          "Rout",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OptionalCosts: []game.AdditionalCost{effects.FlashForMore("{2}")},
		OnResolve:     effects.WrathDestroyAllCreaturesNoRegen,
	})
}
