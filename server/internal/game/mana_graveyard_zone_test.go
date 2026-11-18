package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// mana_graveyard_zone_test.go — #5896: a mana ability that functions
// from the GRAVEYARD (CR 113.6b) and owes mana besides its exile:
// Jack-o'-Lantern's "{1}, Exile this card from your graveyard: Add one
// mana of any color."

// seedGraveyardLantern puts a card with that ability into p's graveyard.
func seedGraveyardLantern(p *Player) uuid.UUID {
	c := NewCard("Jack-o'-Lantern", p.ID)
	c.TypeLine = "Artifact"
	c.ManaCost = "{1}"
	c.ManaAbilities = []ManaAbilityShape{{
		Zones:     []ZoneKind{ZoneGraveyard},
		ManaCost:  "{1}",
		ExileSelf: true,
		Produced:  "{W|U|B|R|G}",
		Label:     "{1}, Exile this card from your graveyard: Add one mana of any color.",
	}}
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

// The ordinary pass never takes it (it owes mana); the costed pass does,
// and the plan's description says it is exiled from the graveyard while
// the Mountain taps to pay for it.
func TestAutoTapPlanDescribesAGraveyardCostedSource(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mountain := pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	jack := seedGraveyardLantern(me)

	var (
		entries []AutoTapPlanEntry
		ok      bool
	)
	g.ReadSnapshot(func() {
		if _, plain := g.autoTapPreferringLocked(me.ID, costFor(t, "{U}"), 0, nil, 0); plain {
			t.Error("the ordinary planner paid {U} with a source that owes {1}")
		}
		var plan tapPlan
		plan, ok = g.autoTapCostedLocked(me, costFor(t, "{U}"), 0, nil, &autoTapWish{})
		if ok {
			entries = g.describePlanLocked(me.ID, plan, ManaSpendContext{})
		}
	})
	if !ok || len(entries) != 2 {
		t.Fatalf("Mountain + Jack-o'-Lantern for {U}: ok=%v entries=%+v", ok, entries)
	}
	by := map[uuid.UUID]AutoTapPlanEntry{}
	for _, e := range entries {
		by[e.CardID] = e
	}
	if e := by[mountain]; e.Zone != ZoneBattlefield || !e.Taps {
		t.Errorf("Mountain described as %+v", e)
	}
	if e := by[jack]; e.Zone != ZoneGraveyard || e.Taps || e.Sacrifices || !e.Exiles {
		t.Errorf("Jack-o'-Lantern described as %+v, want exiled from the graveyard", e)
	}
}

// It is never its own fund: alone, it cannot pay for anything.
func TestGraveyardCostedSourceCannotFundItself(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	jack := seedGraveyardLantern(me)

	g.ReadSnapshot(func() {
		if _, ok := g.autoTapCostedLocked(me, costFor(t, "{U}"), 0, nil, &autoTapWish{}); ok {
			t.Error("a costed plan for {U} out of Jack-o'-Lantern alone")
		}
	})
	var short *InsufficientManaError
	if err := g.ActivateManaAbility(me.ID, jack, 0, ManaAbilityParams{AutoTap: true}); !errors.As(err, &short) {
		t.Fatalf("activate with nothing to pay: %v, want insufficient mana", err)
	}
	if !me.Graveyard.Contains(jack) {
		t.Error("the refused activation moved the card")
	}
}
