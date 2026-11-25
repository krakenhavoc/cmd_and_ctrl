package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// akroma_vision_of_ixidor_test.go — Akroma, Vision of Ixidor (#6162): the
// combat pump counts partner as an ability a permanent has.

const akromaVisionOracle = "b9be5ce9-08bb-4361-940c-9031e429ba86"

func TestAkromaVisionIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(akromaVisionOracle)
	if !ok || spec.Name != "Akroma, Vision of Ixidor" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

func akromaCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int, keywords ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Human", Owner: owner, Controller: owner,
		Power: power, Toughness: toughness, Keywords: keywords,
	})
}

// A partner commander with flying gets +2/+2 (flying and partner); a
// hexproof-from creature with protection gets +2/+2; a vanilla creature
// and Akroma itself get nothing; an opponent's creature gets nothing.
func TestAkromaVisionCountsPartnerAmongTheListedAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	akroma := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Akroma, Vision of Ixidor", TypeLine: "Legendary Creature — Angel",
		OracleID: akromaVisionOracle, Owner: me.ID, Controller: me.ID, Power: 6, Toughness: 6,
	})
	partner := akromaCreature(g, me.ID, "Partner Commander", 2, 2, "flying", game.KeywordPartner)
	warded := akromaCreature(g, me.ID, "Warded", 1, 1, "hexproof from black", "protection from red")
	vanilla := akromaCreature(g, me.ID, "Vanilla", 3, 3)
	theirs := akromaCreature(g, opp.ID, "Their Partner", 2, 2, "flying", game.KeywordPartner)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []struct {
		id         uuid.UUID
		name       string
		power, tgh int
	}{
		{partner, "the partner commander", 4, 4},
		{warded, "the hexproof-from, protection creature", 3, 3},
		{vanilla, "the vanilla creature", 3, 3},
		{akroma, "Akroma", 6, 6},
		{theirs, "the opponent's creature", 2, 2},
	} {
		if p, tough := effectstest.PtOf(t, g, c.id); p != c.power || tough != c.tgh {
			t.Errorf("%s is %d/%d, want %d/%d", c.name, p, tough, c.power, c.tgh)
		}
	}
	// The bonus lasts until end of turn.
	effectstest.RunCleanup(t, g)
	if p, _ := effectstest.PtOf(t, g, partner); p != 2 {
		t.Errorf("after cleanup the partner commander has power %d, want 2", p)
	}
}

// "Partner with" from Scryfall's keyword array is the partner ability
// (CR 702.124n).
func TestPartnerWithIsThePartnerKeyword(t *testing.T) {
	for _, raw := range []string{"Partner", "partner", "Partner with"} {
		kws, ok := game.CanonicalKeywords(raw)
		if !ok || len(kws) != 1 || kws[0] != game.KeywordPartner {
			t.Errorf("CanonicalKeywords(%q) = %v, %v, want [partner]", raw, kws, ok)
		}
	}
	for _, raw := range []string{"Partner with Pir", "Choose a background", "Doctor's companion", "Partner—Friends forever"} {
		if kws, ok := game.CanonicalKeywords(raw); ok {
			t.Errorf("CanonicalKeywords(%q) = %v; not a keyword-only line", raw, kws)
		}
	}
}
