package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82g21SpikeWeaver   = "7776b2dd-5795-458d-81da-54fed22d5ee2"
	s82g21TinderWall    = "dd2f698c-b3b4-4dc8-9950-9be03ca7dc9e"
	s82g21Seeker        = "704723f3-0490-4d96-930a-aacc6d19ba8d"
	s82g21Explorer      = "fa9677d2-1535-4c57-baa8-c0408e15da47"
	s82g21TeferiDruid   = "f2c9957e-d673-4b32-a201-811849c5de31"
	s82g21AnaxKynaios   = "68de3f03-2ae2-4644-acc8-299a57480e97"
	s82g21AnaxKynaiosNm = "Anax and Cymede & Kynaios and Tiro"
)

func TestS82GapB21EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82g21SpikeWeaver: "Spike Weaver", s82g21TinderWall: "Tinder Wall", s82g21Seeker: "Seeker of Skybreak",
		s82g21Explorer: "Veteran Explorer", s82g21TeferiDruid: "Teferi, Druid of Argoth", s82g21AnaxKynaios: s82g21AnaxKynaiosNm,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82GapB21SpikeWeaverEntersWithThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastCatalogSpell(t, g, "Spike Weaver", "Creature — Spike", s82g21SpikeWeaver, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters, want 3", got)
	}
	if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 3 {
		t.Errorf("Weaver is %d/%d, want 3/3", p, tough)
	}
}

func TestS82GapB21SpikeWeaverMovesACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, spike := effectstest.SthB2Spike(t, g, "Spike Weaver", "Creature — Spike", s82g21SpikeWeaver, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DcMana(t, g, me.ID, "{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, spike, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("target has %d counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, spike, game.CounterPlusOne); got != 2 {
		t.Errorf("Weaver has %d counters, want 2", got)
	}
}

func TestS82GapB21SpikeWeaverFogsCombatDamageForOneAndACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, spike := effectstest.SthB2Spike(t, g, "Spike Weaver", "Creature — Spike", s82g21SpikeWeaver, 3)
	if err := g.ActivateCatalogAbility(me.ID, spike, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the Fog ability was activated with no mana")
	}
	effectstest.DcMana(t, g, me.ID, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, spike, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, spike, game.CounterPlusOne); got != 2 {
		t.Errorf("Weaver has %d counters, want 2", got)
	}
	if effectstest.ScopedReplacementCount(g) != 1 {
		t.Errorf("scoped replacements = %d, want the one combat-damage shield", effectstest.ScopedReplacementCount(g))
	}
}

func TestS82GapB21TinderWallSacrificesForRR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wall := effectstest.PushCatalogPermanent(g, me.ID, "Tinder Wall", "Creature — Plant Wall", s82g21TinderWall, true)
	if err := g.ActivateManaAbility(me.ID, wall, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("pool %v, want R R", got)
	}
	if g.Battlefield.Contains(wall) {
		t.Error("the Wall is sacrificed to its own ability")
	}
}

func TestS82GapB21TinderWallDealsTwoToTheCreatureItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.PushCatalogPermanent(g, opp.ID, "Tinder Wall", "Creature — Plant Wall", s82g21TinderWall, false)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 4, 4)
	other := effectstest.CtrlPushCreature(g, me.ID, "Other")
	effectstest.CrtBlock(g, wall, attacker, opp.ID)
	effectstest.B06AddMana(opp, "R")
	if err := effectstest.CrtActivateRaw(t, g, opp.ID, wall, 0, effectstest.LtCardTarget(other)); err == nil {
		t.Fatal("targeting a creature it is not blocking was accepted")
	}
	if err := effectstest.CrtActivateRaw(t, g, opp.ID, wall, 0, effectstest.LtCardTarget(attacker)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(wall) {
		t.Error("the Wall was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, attacker).DamageMarked; got != 2 {
		t.Errorf("damage on the blocked creature = %d, want 2 (the Wall's last-known self, CR 608.2h)", got)
	}
	if got := effectstest.LayeredCard(t, g, other).DamageMarked; got != 0 {
		t.Errorf("damage on the other creature = %d, want 0", got)
	}
}

func TestS82GapB21SeekerOfSkybreakUntapsTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seeker := effectstest.PushCatalogPermanent(g, me.ID, "Seeker of Skybreak", "Creature — Elf", s82g21Seeker, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, bear).Tapped {
		t.Error("the target is still tapped")
	}
	if !effectstest.FindBattlefieldCardForTest(g, seeker).Tapped {
		t.Error("the Seeker's {T} cost was not paid")
	}
	if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Error("a tapped Seeker activated again")
	}
}

func TestS82GapB21SeekerIsSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seeker := effectstest.PushCatalogPermanent(g, me.ID, "Seeker of Skybreak", "Creature — Elf", s82g21Seeker, true)
	if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(seeker), Strict: true}); err == nil {
		t.Error("a summoning-sick Seeker activated its {T} ability")
	}
}

func TestS82GapB21VeteranExplorerLetsEachPlayerFetchTwoBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	explorer := effectstest.PushDiesCreatureForTest(g, me.ID, "Veteran Explorer", s82g21Explorer, "Creature — Human Soldier Scout", 1, 1)
	want := map[uuid.UUID][]uuid.UUID{}
	for _, p := range g.Seats {
		a := effectstest.B07LibraryLand(p, "Forest", "Basic Land — Forest")
		b := effectstest.B07LibraryLand(p, "Forest", "Basic Land — Forest")
		want[p.ID] = []uuid.UUID{a, b}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(explorer) })
	effectstest.PassPriorityAroundTable(t, g)
	// Seat 1 declines; everyone else takes both.
	for i, p := range g.Seats {
		if i == 1 {
			effectstest.AnswerSearchFailToFind(t, g, p.ID)
			continue
		}
		effectstest.AnswerSearchByID(t, g, p.ID, want[p.ID]...)
	}
	for i, p := range g.Seats {
		for _, id := range want[p.ID] {
			card, on := effectstest.BattlefieldCard(g, id)
			if i == 1 {
				if on {
					t.Errorf("seat 1 declined but a basic entered")
				}
				continue
			}
			if !on || card.Tapped {
				t.Errorf("seat %d: basic %s should be on the battlefield untapped", i, id)
			}
		}
	}
}

func TestS82GapB21TeferiDruidGivesOwnedCreaturesFlashAndBarsOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Teferi, Druid of Argoth", "Legendary Creature — Human Druid", s82g21TeferiDruid, 3, 4)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	mine := effectstest.NsHandCard(me, "My Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, mine.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("an owned creature card at instant speed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	sorcery := effectstest.NsHandCard(me, "My Sorcery", "Sorcery", "")
	if err := g.CastSpell(me.ID, sorcery.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a sorcery card was cast at instant speed")
	}
	instant := effectstest.NsHandCard(opp, "Opp Instant", "Instant", "")
	if err := g.CastSpell(opp.ID, instant.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("an opponent cast an instant under Teferi")
	}
}

func TestS82GapB21AnaxHeroicDrawsThenEachPlayerMayPutALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	anax := effectstest.B12Push(g, me.ID, s82g21AnaxKynaiosNm, "Legendary Creature — Human Soldier", s82g21AnaxKynaios, 3, 8)
	for _, p := range g.Seats {
		effectstest.PushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
		effectstest.PushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	// The controller and seat order: me, then the others after me.
	order := []*game.Player{me}
	for i := 1; i < len(g.Seats); i++ {
		order = append(order, g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)])
	}
	lands := map[uuid.UUID]uuid.UUID{}
	for _, i := range []int{0, 2} { // the controller and the second opponent hold a land
		lands[order[i].ID] = effectstest.HandCardForTest(order[i], "Forest", "Basic Land — Forest", "")
	}
	handBefore := map[uuid.UUID]int{}
	for _, p := range order {
		handBefore[p.ID] = len(p.Hand.Cards)
	}
	effectstest.BnsB25Heroic(t, g, anax)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range order {
		if land, has := lands[p.ID]; has {
			effectstest.AnswerChooseCards(t, g, p.ID, land)
			if !g.Battlefield.Contains(land) {
				t.Errorf("seat %d's land did not enter", i)
			}
		}
	}
	// Hearts on Fire is cast from a hand it was added to, so it nets out: the
	// controller drew once and put a land; the second opponent put a land; the
	// others had none and drew.
	for i, p := range order {
		delta := len(p.Hand.Cards) - handBefore[p.ID]
		want := 1
		switch i {
		case 0:
			want = 0
		case 2:
			want = -1
		}
		if delta != want {
			t.Errorf("seat %d hand changed by %d, want %d", i, delta, want)
		}
	}
}
