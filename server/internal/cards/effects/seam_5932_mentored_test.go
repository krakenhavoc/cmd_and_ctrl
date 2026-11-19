package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #5932: "mentors a creature" (CR 702.134c) is game.EventMentored, recorded
// as a mentor ability resolves; Aegis of the Legion watches it for the
// creature it equips.

const aegisOfTheLegionOracle = "99eb05eb-11ef-456f-9132-166a0691dd04"

// aegisScene is a 2/2 bearer (3/3 with the Aegis) on the active seat, the
// Aegis equipped to it through its own Equip {3}, and an allied 1/2.
func aegisScene(t *testing.T) (g *game.Game, me *game.Player, aegis, bearer, small uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	bearer = ally(g, "Bearer", 2)
	small = ally(g, "Small", 1)
	aegis = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Aegis of the Legion", TypeLine: equipTypeLine,
		OracleID: aegisOfTheLegionOracle, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	fillPool(me, 3)
	equipTo(t, g, me.ID, aegis, bearer)
	if !findBattlefieldCardByID(g, aegis).IsAttachedTo(bearer) {
		t.Fatal("the Aegis is not attached")
	}
	return g, me, aegis, bearer, small
}

func TestAegisOfTheLegionEquipCostsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bearer := ally(g, "Bearer", 2)
	aegis := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Aegis of the Legion", TypeLine: equipTypeLine,
		OracleID: aegisOfTheLegionOracle, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, aegis, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bearer}}, Strict: true,
	}); err == nil {
		t.Fatal("equip paid with two mana, want a refusal: Equip costs {3}")
	}
}

func TestAegisOfTheLegionPumpsAndGrantsMentor(t *testing.T) {
	g, me, _, bearer, small := aegisScene(t)
	if c, _ := g.LookupCardForEffect(bearer); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Fatalf("bearer is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	prompt := mentorAttack(t, g, bearer, small)
	if offered := prompt.PickTargetCards; len(offered) != 1 || offered[0] != small {
		t.Fatalf("mentor offered %v, want only the 1-power attacker", offered)
	}
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)

	got := allCountersOn(t, g, small)
	if got[game.CounterPlusOne] != 1 {
		t.Errorf("+1/+1 counters = %d, want 1 (mentor)", got[game.CounterPlusOne])
	}
	if got[game.CounterShield] != 1 {
		t.Errorf("shield counters = %d, want 1 (equipped creature mentored it)", got[game.CounterShield])
	}
	evs := eventsOfKind(g, game.EventMentored)
	if len(evs) != 1 || evs[0].Source != bearer || evs[0].CardID != small || evs[0].Actor != me.ID {
		t.Fatalf("EventMentored = %+v, want one with Source the bearer and CardID the mentored creature", evs)
	}
	if n := allCountersOn(t, g, bearer)[game.CounterShield]; n != 0 {
		t.Errorf("bearer has %d shield counters, want 0: the counter goes on the mentored creature", n)
	}
}

// Another creature's mentor is not the equipped creature mentoring.
func TestAegisOfTheLegionIgnoresAnotherCreaturesMentor(t *testing.T) {
	g, me, _, _, small := aegisScene(t)
	other := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hammer Dropper", TypeLine: "Creature — Soldier",
		OracleID: hammerDropperOracle, Power: 5, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	mentorAttack(t, g, other, small)
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)

	got := allCountersOn(t, g, small)
	if got[game.CounterPlusOne] != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got[game.CounterPlusOne])
	}
	if got[game.CounterShield] != 0 {
		t.Errorf("shield counters = %d, want 0: the Hammer Dropper is not the equipped creature", got[game.CounterShield])
	}
	if evs := eventsOfKind(g, game.EventMentored); len(evs) != 1 || evs[0].Source != other {
		t.Errorf("EventMentored = %+v, want one from the Hammer Dropper", evs)
	}
}

// A mentor ability whose target became illegal does not resolve (CR
// 608.2b), so nothing was mentored and the Aegis does not trigger.
func TestAegisOfTheLegionNoTriggerWhenMentorTargetBecameIllegal(t *testing.T) {
	g, me, _, bearer, small := aegisScene(t)
	mentorAttack(t, g, bearer, small)
	pickCard(t, g, me.ID, small)
	findBattlefieldCardByID(g, small).Counters = map[string]int{game.CounterPlusOne: 2}
	passPriorityAroundTable(t, g)

	if n := len(eventsOfKind(g, game.EventMentored)); n != 0 {
		t.Errorf("EventMentored emitted %d times, want 0", n)
	}
	if n := allCountersOn(t, g, small)[game.CounterShield]; n != 0 {
		t.Errorf("shield counters = %d, want 0", n)
	}
}

// The mentored creature flickered before the Aegis trigger resolves is a
// new object (CR 400.7): "that creature" is gone and gets no shield.
func TestAegisOfTheLegionShieldMissesAFlickeredCreature(t *testing.T) {
	g, me, _, bearer, small := aegisScene(t)
	mentorAttack(t, g, bearer, small)
	pickCard(t, g, me.ID, small)
	for i := 0; i < 32 && len(eventsOfKind(g, game.EventMentored)) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if len(eventsOfKind(g, game.EventMentored)) != 1 {
		t.Fatal("the mentor ability never resolved")
	}
	flickerForTest(t, g, me.ID, small)
	passPriorityAroundTable(t, g)

	for _, c := range g.Battlefield.Cards {
		if c.Name == "Small" && c.Counters[game.CounterShield] != 0 {
			t.Errorf("the flickered creature got %d shield counters, want 0", c.Counters[game.CounterShield])
		}
	}
}
