package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b565_test.go — Coldsnap, batch 565 (#5597): each test plays the
// real catalog entry through the real engine paths.

const (
	b565GarzaZol  = "5b48e926-fc0e-403c-a47b-1f1d96e2eaf7"
	b565Dragon    = "3f8182e5-4df5-4beb-a3b7-387ed1e86c99"
	b565Fylgja    = "4d244787-66d2-4a32-914d-a3e97d393370"
	b565Spider    = "42176320-ee7f-44c9-916e-5ee4a6108be3"
	b565Owl       = "3c0f7c4d-ac93-4a58-a87f-bf694c709624"
	b565Razorback = "c5d27469-96e1-44ec-a887-2fc26f9e10d3"
)

func TestS82OldB565CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b565GarzaZol: "Garza Zol, Plague Queen", b565Dragon: "Rimescale Dragon", b565Fylgja: "Ursine Fylgja",
		b565Spider: "Frostweb Spider", b565Owl: "Rimefeather Owl", b565Razorback: "Woolly Razorback",
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

func b565SnowLand(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	return effectstest.BlbB1Push(g, owner, "Snow-Covered "+subtype, "Basic Snow Land — "+subtype, "", 0, 0)
}

func b565Counters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	return effectstest.LayeredCard(t, g, id).Counters[kind]
}

// b565Enter puts the card onto the battlefield from the hand, so its entry
// replacements apply, with the printed power and toughness.
func b565Enter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.OracleID, c.Power, c.Toughness = typeLine, oracle, power, toughness
	owner.Hand.PushTop(c)
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		id, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, game.HandEntryOptions{Controller: owner.ID})
		if err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, id).SummonedThisTurn = false })
	return id
}

func TestS82OldB565GarzaZolGrowsWhenACreatureItDamagedDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	garza := effectstest.B12Push(g, me.ID, "Garza Zol, Plague Queen", "Legendary Creature — Vampire Noble", b565GarzaZol, 5, 5)
	victim := effectstest.B16Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2, "G")
	bystander := effectstest.B16Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, garza, kw) {
			t.Errorf("Garza Zol has no %s", kw)
		}
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(garza, victim, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	effectstest.KillCreature(t, g, opp.ID, bystander)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b565Counters(t, g, garza, game.CounterPlusOne); got != 0 {
		t.Errorf("a creature Garza never damaged died: counters = %d, want 0", got)
	}
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b565Counters(t, g, garza, game.CounterPlusOne); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
}

func TestS82OldB565GarzaZolDrawIsOptional(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		garza := effectstest.B12Push(g, me.ID, "Garza Zol, Plague Queen", "Legendary Creature — Vampire Noble", b565GarzaZol, 5, 5)
		hand := me.Hand.Size()
		effectstest.DealCombatDamageToPlayer(g, garza, opp.ID, 5)
		effectstest.AssistCardSettle(t, g, me.ID, nil, yes)
		want := 0
		if yes {
			want = 1
		}
		if got := me.Hand.Size() - hand; got != want {
			t.Errorf("answered %v: drew %d, want %d", yes, got, want)
		}
	}
}

func TestS82OldB565RimescaleDragonTapsAndIcesAndKeepsItDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.B12Push(g, me.ID, "Rimescale Dragon", "Snow Creature — Dragon", b565Dragon, 5, 5)
	frozen := effectstest.B16Creature(g, me.ID, "Frozen", "Creature — Bear", 2, 2, "G")
	free := effectstest.B16Creature(g, me.ID, "Free", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	// {S} needs snow mana: three nonsnow Forests do not pay it.
	for i := 0; i < 3; i++ {
		effectstest.BlbB1Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	}
	params := game.ActivateAbilityParams{Strict: true, AutoTap: true, Targets: effectstest.CardRefs(frozen)}
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, params); err == nil {
		t.Fatal("{2}{S} paid with nonsnow Forests: accepted")
	}
	for i := 0; i < 3; i++ {
		b565SnowLand(g, me.ID, "Island")
	}
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, params); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, frozen); c == nil || !c.Tapped {
		t.Fatal("the target was not tapped")
	}
	if got := b565Counters(t, g, frozen, "ice"); got != 1 {
		t.Fatalf("ice counters = %d, want 1", got)
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, free).Tapped = true })
	effectstest.AdvanceToDrawStepOfSeat(t, g, 0)
	if c := effectstest.FindBattlefieldCardByID(g, frozen); c == nil || !c.Tapped {
		t.Error("a creature with an ice counter untapped")
	}
	if c := effectstest.FindBattlefieldCardByID(g, free); c == nil || c.Tapped {
		t.Error("a creature without one stayed tapped")
	}
}

func TestS82OldB565UrsineFylgjaHealsAndRefills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := b565Enter(t, g, me, "Ursine Fylgja", "Creature — Spirit Bear", b565Fylgja, 3, 3)
	if got := b565Counters(t, g, bear, "healing"); got != 4 {
		t.Fatalf("entered with %d healing counters, want 4", got)
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("remove a counter: %v", err)
	}
	if got := b565Counters(t, g, bear, "healing"); got != 3 {
		t.Errorf("after the cost %d counters, want 3 (paid on activation)", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.New(), bear, 3); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(bear) {
		t.Error("3 damage with one prevented should leave a 3/3 alive")
	}
	effectstest.B06AddMana(me, "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, bear, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{2}{W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := b565Counters(t, g, bear, "healing"); got != 4 {
		t.Errorf("after the refill %d counters, want 4", got)
	}
}

func TestS82OldB565FrostwebSpiderGrowsOnlyFromBlockingAFlyer(t *testing.T) {
	for _, flyer := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		var attacker uuid.UUID
		if flyer {
			attacker = effectstest.Pr7bCreature(g, me.ID, "Bird", 1, "{1}", []string{"U"}, "flying")
		} else {
			attacker = effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1, "G")
		}
		spider := b565Enter(t, g, opp, "Frostweb Spider", "Snow Creature — Spider", b565Spider, 1, 3)
		if !effectstest.HasKeywordOnBattlefield(t, g, spider, "reach") {
			t.Fatal("Frostweb Spider has no reach")
		}
		effectstest.DeclareAttack(t, g, opp.ID, attacker)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(spider, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
		effectstest.LockInBlocks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AdvanceTo(t, g, game.StepCombatDamage)
		effectstest.PassPriorityAroundTable(t, g)
		if got := b565Counters(t, g, spider, game.CounterPlusOne); got != 0 {
			t.Errorf("flyer %v: %d counters before end of combat, want 0", flyer, got)
		}
		effectstest.AdvanceTo(t, g, game.StepEndCombat)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if flyer {
			want = 1
		}
		if got := b565Counters(t, g, spider, game.CounterPlusOne); got != want {
			t.Errorf("flyer %v: %d counters after end of combat, want %d", flyer, got, want)
		}
	}
}

func TestS82OldB565RimefeatherOwlCountsSnowPermanentsAndIcesMakeSnow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	owl := effectstest.B12Push(g, me.ID, "Rimefeather Owl", "Snow Creature — Bird", b565Owl, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if p, tough := effectstest.PtOf(t, g, owl); p != 1 || tough != 1 {
		t.Fatalf("alone %d/%d, want 1/1", p, tough)
	}
	b565SnowLand(g, me.ID, "Island")
	b565SnowLand(g, opp.ID, "Swamp")
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	if p, tough := effectstest.PtOf(t, g, owl); p != 3 || tough != 3 {
		t.Fatalf("with two snow lands %d/%d, want 3/3", p, tough)
	}
	b565SnowLand(g, me.ID, "Forest")
	if err := g.ActivateCatalogAbility(me.ID, owl, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("{1}{S}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := b565Counters(t, g, bear, "ice"); got != 1 {
		t.Fatalf("ice counters on the Bear = %d, want 1", got)
	}
	if !effectstest.LayeredCard(t, g, bear).HasSupertype("snow") {
		t.Error("a permanent with an ice counter is not snow")
	}
	if p, tough := effectstest.PtOf(t, g, owl); p != 5 || tough != 5 {
		t.Errorf("with three snow lands and an iced Bear %d/%d, want 5/5", p, tough)
	}
}

func b565BlockedBy(t *testing.T, g *game.Game, attacker, blocker uuid.UUID, defender uuid.UUID) {
	t.Helper()
	effectstest.DeclareAttack(t, g, defender, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82OldB565WoollyRazorbackIsAnIcedDefenderThatThawsAsItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B16Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 3, "G")
	boar := b565Enter(t, g, opp, "Woolly Razorback", "Creature — Boar Beast", b565Razorback, 7, 7)
	if got := b565Counters(t, g, boar, "ice"); got != 3 {
		t.Fatalf("entered with %d ice counters, want 3", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, boar, "defender") {
		t.Error("an iced Razorback has no defender")
	}
	b565BlockedBy(t, g, attacker, boar, opp.ID)
	if got := b565Counters(t, g, boar, "ice"); got != 2 {
		t.Errorf("after blocking %d ice counters, want 2", got)
	}
	if !g.Battlefield.Contains(attacker) {
		t.Error("the 7/7 with ice counters on it should deal no combat damage")
	}
}

func TestS82OldB565WoollyRazorbackThawedDealsDamageAndLosesDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B16Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 3, "G")
	boar := b565Enter(t, g, opp, "Woolly Razorback", "Creature — Boar Beast", b565Razorback, 7, 7)
	g.WithWriteLock(func() { delete(effectstest.FindBattlefieldCardByID(g, boar).Counters, "ice") })
	g.BumpLayerVersionForTest()
	if effectstest.HasKeywordOnBattlefield(t, g, boar, "defender") {
		t.Error("a thawed Razorback still has defender")
	}
	b565BlockedBy(t, g, attacker, boar, opp.ID)
	if g.Battlefield.Contains(attacker) {
		t.Error("a Razorback with no ice counter should have killed the attacker")
	}
}
