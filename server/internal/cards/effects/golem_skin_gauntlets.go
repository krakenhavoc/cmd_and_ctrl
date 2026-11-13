package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Golem-Skin Gauntlets — Artifact — Equipment {1} (Mirrodin, #5606):
//
//	"Equipped creature gets +1/+0 for each Equipment attached to it.
//	 Equip {2}"
//
// The per-count pump (PumpAttachedPer), a layer 7c modification
// (CR 613.4c), counting every Equipment attached to the creature the
// Gauntlets are attached to, whoever controls it, the Gauntlets
// themselves included. It is read on every recompute, so a second
// Equipment attached or moved away changes the bonus at once. An
// unattached Gauntlets pumps nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a956224-8bdb-40d8-8405-2aa29eb228e0",
		Name:         "Golem-Skin Gauntlets",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttachedPer(1, 0, equipmentOnTheHost)},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}

// equipmentOnTheHost counts "each Equipment attached to it" for an
// Equipment's own host: every Equipment attached to that permanent, the
// source included. aurasAndEquipmentOnTheHost's twin without the Auras.
func equipmentOnTheHost(g *game.Game, source *game.Card) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	host := g.AttachedHostOf(source)
	if host == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.IsAttachedTo(host.InstanceID) && a.HasSubtype("Equipment") {
			n++
		}
	}
	return n
}
