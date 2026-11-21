package effectstest

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 — The Brothers' War, batch 4: artifact creatures with unearth.

type BroB04Card struct {
	Name, TypeLine, Oracle string
	Power, Toughness       int
	Unearth                string
}

func BroB04Unearth(t *testing.T, g *game.Game, tc BroB04Card) (uuid.UUID, *game.Player) {
	t.Helper()
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.Oracle})
	idx := -1
	for i, ab := range abilities {
		if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("%s: no unearth", tc.Name)
	}
	me := g.Seats[g.Turn.ActiveSeat]
	ToMain(t, g)
	id := PushCatalogGraveyardCard(me, tc.Name, tc.TypeLine, tc.Oracle, tc.Power, tc.Toughness)
	RfCrAAddMana(t, g, me, BroB04Pool(tc.Unearth))
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("%s: unearth: %v", tc.Name, err)
	}
	PassPriorityAroundTable(t, g)
	return id, me
}

// broB04Pool turns a printed cost into mana to float: generic {N} as N {C}.
func BroB04Pool(cost string) string {
	return regexp.MustCompile(`\{(\d+)\}`).ReplaceAllStringFunc(cost, func(m string) string {
		n, _ := strconv.Atoi(m[1 : len(m)-1])
		return strings.Repeat("{C}", n)
	})
}
