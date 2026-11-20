package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Manaforce Mace — Artifact — Equipment {4} (Conflux, #5587):
//
//	"Domain — Equipped creature gets +1/+1 for each basic land type among
//	 lands you control.
//	 Equip {3}"
//
// Blackblade Reforged's per-count pump with the domain count of the
// Mace's controller (domain.go), read on every layer pass, so a new land
// type grows the creature at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e92be29-ccd1-4556-bb2a-ac7902bc77c2",
		Name:         "Manaforce Mace",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				return DomainCount(g, source.Controller)
			}),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
