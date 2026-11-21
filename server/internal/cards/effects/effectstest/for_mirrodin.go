package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castForMirrodin casts the named Equipment from the active seat's hand
// and settles the spell and the For Mirrodin! trigger. Returns the
// Equipment and the Rebel it made.
func CastForMirrodin(t *testing.T, g *game.Game, name, oracle string) (equipment, rebel uuid.UUID) {
	t.Helper()
	equipment = CastCatalogSpell(t, g, name, EquipTypeLine, oracle, nil)
	PassPriorityAroundTable(t, g)
	rebel = FindBattlefieldByName(g, "Rebel")
	if rebel == uuid.Nil {
		t.Fatalf("%s made no Rebel token", name)
	}
	if host := AttachmentHostOf(t, g, equipment); host.Kind != game.TargetCard || host.ID != rebel {
		t.Fatalf("%s AttachedTo = %+v, want the Rebel %s", name, host, rebel)
	}
	return equipment, rebel
}

// forMirrodinKeywords are the keywords any of the nine cards grants; a
// card's test checks that it grants its own and none of the others.
var forMirrodinKeywords = []string{"flying", "haste", "vigilance", "trample", "first strike"}

func ForMirrodinCheck(t *testing.T, g *game.Game, what string, id uuid.UUID, power, toughness int, keywords ...string) {
	t.Helper()
	if p, th := EffectivePTOf(t, g, id); p != power || th != toughness {
		t.Errorf("%s is %d/%d, want %d/%d", what, p, th, power, toughness)
	}
	for _, kw := range forMirrodinKeywords {
		if got, want := HasEffectiveKeyword(t, g, id, kw), ContainsString(keywords, kw); got != want {
			t.Errorf("%s has %s = %v, want %v", what, kw, got, want)
		}
	}
}
