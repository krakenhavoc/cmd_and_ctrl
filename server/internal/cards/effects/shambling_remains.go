package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shambling Remains — Creature — Zombie Horror {1}{B}{R}, 4/3 (Conflux,
// #5587):
//
//	"This creature can't block.
//	 Unearth {B}{R} ({B}{R}: Return this card from your graveyard to the
//	 battlefield. It gains haste. Exile it at the beginning of the next
//	 end step or if it would leave the battlefield. Unearth only as a
//	 sorcery.)"
//
// Craven Giant's self restriction, checked as blockers are declared (CR
// 509.1b), and the shared unearth ability (unearth.go, CR 702.84a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a29fe9f-89bf-405e-acf6-1b56435cb9ee",
		Name:         "Shambling Remains",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Activated:    []ActivatedAbility{Unearth("{B}{R}")},
	})
}
