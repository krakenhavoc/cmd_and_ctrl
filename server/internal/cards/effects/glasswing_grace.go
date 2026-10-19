package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glasswing Grace // Age-Graced Chapel — Modal DFC. Front: Enchantment —
// Aura {3}{W/B}{W/B}: "Enchant creature. Enchanted creature gets +2/+2 and
// has flying and lifelink." Back: Land (mdfc_lands.go).
//
// An Aura declared as the "enchant creature" clause plus two statics that
// read the host live: the layer 7c bonus and the layer 6 keywords. When
// the host goes, state-based actions put the Aura into the graveyard
// (CR 704.5m). No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a3e8c9b-e458-4661-980d-0a84a4c2452b",
		Name:         "Glasswing Grace",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("flying", "lifelink"),
		},
	})
}
