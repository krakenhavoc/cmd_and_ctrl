package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Warchief — Creature — Goblin Warrior {1}{R}{R}, 2/2 (Scourge):
//
//	"Goblin spells you cast cost {1} less to cast.
//	 Goblins you control have haste."
//
// Undead Warchief's discount for Goblins (one cost modifier, CR 601.2f,
// generic mana only; a changeling or kindred Goblin spell counts) and a
// layer-6 keyword grant (CR 613.1f) to every Goblin creature you control,
// the Warchief included. Two Warchiefs discount a Goblin spell by {2}.
//
// No simplification.
func init() {
	haste := TribalKeywordGrant(TribeFilter{Tribes: []string{"Goblin"}, YoursOnly: true}, "haste")
	haste.Label = "Goblins you control have haste."
	Register(Spec{
		OracleID:     "39882df0-c20f-469d-94ba-1617224e71a1",
		Name:         "Goblin Warchief",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Goblin spells you cast cost {1} less to cast.", YourSpell(), SpellOfCreatureType("Goblin")),
		},
		Static: []game.StaticAbility{haste},
	})
}
