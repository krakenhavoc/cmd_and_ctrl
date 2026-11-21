package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// living_weapon_test.go pins the keyword itself rather than the three
// cards that carry it. What is worth asserting is exactly the part
// that would break silently:
//
//   - the Germ is made AND the Equipment lands on it, in one
//     resolution, so the 0/0 is never a 0/0 with nothing on it when
//     state-based actions next run;
//   - the pump the host gets is the Equipment's, so the same helper
//     really is serving three different cards;
//   - the Germ dies the moment the Equipment leaves, which is the
//     other half of the same rule and is what Batterskull's {3}
//     bounce does on purpose.

const (
	BatterskullOracle    = "d12e5ce0-5705-4c35-9a93-b883db52c80c"
	BatterboneOracle     = "389d459d-446a-4b84-82b2-a30fe6ced11f"
	KaldraCompleatOracle = "7359e82b-db79-488d-a1d4-75a00f12a4cf"
)

// castLivingWeapon casts the named Equipment from the active seat's
// hand and settles the spell and the living weapon trigger it queues.
// Returns the Equipment and the Germ it made.
func CastLivingWeapon(t *testing.T, g *game.Game, name, oracle string) (equipment, germ uuid.UUID) {
	t.Helper()
	equipment = CastCatalogSpell(t, g, name, EquipTypeLine, oracle, nil)
	PassPriorityAroundTable(t, g)
	germ = FindBattlefieldByName(g, "Phyrexian Germ")
	if germ == uuid.Nil {
		t.Fatalf("%s made no Germ token", name)
	}
	return equipment, germ
}
