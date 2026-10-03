package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Outrider en-Kor — Creature — Kor Rebel Knight {2}{W}, 2/2:
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 {0}: The next 1 damage that would be dealt to this creature this turn
//	 is dealt to target creature you control instead."
//
// ADR 0108 §9 (#1905): the en-Kor row (enKorRow) and Flanking
// (flanking.go). The rulings: blocked by several creatures at once, you
// choose which source's 1 damage each activation redirects (divide_shield);
// the damage is still the original source's; a target that has left
// leaves the damage where it was.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "644b7aba-a7b8-4861-9e83-73329cf85be2",
		Name:         "Outrider en-Kor",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Flanking("Outrider en-Kor")},
		Activated:    []ActivatedAbility{enKorRow()},
	})
}
