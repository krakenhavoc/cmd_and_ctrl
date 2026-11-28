package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amulet of Unmaking — Artifact {5} (Mirage):
//
//	"{5}, {T}, Exile this artifact: Exile target artifact, creature, or land.
//	 Activate only as a sorcery."
//
// Brittle Effigy's cost shape (mana, tap and exiling the artifact itself, all
// paid as the ability is activated, CR 602.2b) at sorcery speed (CR 602.5d),
// with one target that is an artifact, a creature or a land, exiled as the
// ability resolves if still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "945d5351-ce7d-488c-bc63-269ad82b74cb",
		Name:         "Amulet of Unmaking",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{5}, {T}, Exile this artifact: Exile target artifact, creature, or land. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.ExileThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
			Effect:       effects.ExileFirstLegalCardTarget,
		}},
	})
}
