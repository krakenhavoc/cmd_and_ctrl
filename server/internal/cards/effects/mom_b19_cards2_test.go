package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Sword of Once and Future ---

func TestSwordOfOnceAndFutureGivesPlusTwoAndProtection(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := seedEquipment(g, me.ID, "Sword of Once and Future", momB19SwordOracle)
	equipTo(t, g, me.ID, sword, bear)
	assertKeywords(t, g, bear, "protection from blue", "protection from black")
	c := battlefieldCardPtr(t, g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("%d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSwordOfOnceAndFutureSurveilsThenCastsACheapSpellFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := seedEquipment(g, me.ID, "Sword of Once and Future", momB19SwordOracle)
	equipTo(t, g, me.ID, sword, bear)
	lib := seedLibrary(me, "Binned Spell", "Keep Me", "Third")
	// The surveil will put this one in the graveyard first.
	g.WithWriteLock(func() {
		me.Library.Remove(lib[0])
		me.Library.PushTop(game.Card{InstanceID: lib[0], Name: "Binned Spell", TypeLine: "Instant", ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID})
	})
	cheap := gyCard(me, "Old Bolt", "Instant", "{R}", "")
	big := gyCard(me, "Old Wrath", "Sorcery", "{2}{W}{W}", "")
	creature := gyCard(me, "Old Bear", "Creature — Bear", "{1}", "")
	theirs := gyCard(opp, "Their Bolt", "Instant", "{R}", "")

	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no surveil prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{lib[0]}, []uuid.UUID{lib[1]}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{big, creature, theirs} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true,
			AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
			t.Errorf("cast %s, which the clause does not allow", id)
		}
	}
	// The card the surveil binned is a legal pick, being mana value 2.
	if err := g.CastSpell(me.ID, lib[0], game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast the surveilled card free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(lib[0]); z == nil || z.Kind != game.ZoneExile {
		t.Error("the spell was not exiled instead of going to the graveyard")
	}
	// "A spell": the one cast spent the permission.
	if err := g.CastSpell(me.ID, cheap, game.CastSpellParams{FromZone: "graveyard", Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Error("cast a second spell")
	}
}

// --- Invasion of Vryn // Overloaded Mage-Ring ---

func TestInvasionOfVrynDrawsThreeThenDiscardsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C", "D", "E")
	advanceToMain(t, g)
	hand := me.Hand.Size()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Invasion of Vryn", TypeLine: "Battle — Siege", ManaCost: "{3}{U}",
		OracleID: momB19VrynOracle, Owner: me.ID, Controller: me.ID})
	floatForTest(g, me, "UUUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerProtectorPrompt(t, g, me.ID, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Fatalf("hand %d, want %d (three drawn, discard pending)", me.Hand.Size(), hand+3)
	}
	if c := discardChoiceFor(g, me.ID); c == nil {
		t.Fatalf("no discard prompt: %+v", g.PendingChoices)
	}
}

func TestOverloadedMageRingCopiesASpellYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ring := b12Push(g, me.ID, "Overloaded Mage-Ring", "Artifact", momB19VrynOracle+"#1", 0, 0)
	spell := castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: spell}}
	// Not free: a pool that cannot pay the {1} refuses.
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, ring, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, ring) {
		t.Error("the Mage-Ring was not sacrificed")
	}
	if got := countOnBattlefield(g, "Grizzly Bears", me.ID); got != 2 {
		t.Errorf("Grizzly Bears on the battlefield = %d, want 2 (the spell and its token copy)", got)
	}
}

// --- Chandra, Hope's Beacon ---

func TestChandraHopesBeaconMinusXDamagesUpToTwoTargets(t *testing.T) {
	g, me, w := momB19Walker(t, "Chandra, Hope's Beacon", momB19ChandraOracle, 5)
	a, b := g.Seats[1], g.Seats[2]
	la, lb := a.Life, b.Life
	b16Activate(t, g, me.ID, w, 2, game.ActivateAbilityParams{XValue: 3, Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetPlayer, ID: b.ID}}})
	passPriorityAroundTable(t, g)
	if a.Life != la-3 || b.Life != lb-3 {
		t.Errorf("lives %d, %d; want %d, %d", a.Life, b.Life, la-3, lb-3)
	}
	if got := countersOn(g, w, game.CounterLoyalty); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestChandraHopesBeaconPlusTwoAddsTwoMana(t *testing.T) {
	g, me, w := momB19Walker(t, "Chandra, Hope's Beacon", momB19ChandraOracle, 5)
	b16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	for _, c := range []string{"R", "G"} {
		pick := riderLatestManaPick(g, me.ID)
		if pick == nil {
			t.Fatalf("missing a colour pick for %s", c)
		}
		if err := g.ResolveManaChoice(pick.ID, me.ID, c); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	var got int
	g.ReadSnapshot(func() { got = me.ManaPool.Total() })
	if got != 2 {
		t.Errorf("pool %d, want 2", got)
	}
	if loyalty := countersOn(g, w, game.CounterLoyalty); loyalty != 7 {
		t.Errorf("loyalty %d, want 7", loyalty)
	}
}

func TestChandraHopesBeaconPlusOneExilesFiveAndAllowsOneInstantOrSorcery(t *testing.T) {
	g, me, w := momB19Walker(t, "Chandra, Hope's Beacon", momB19ChandraOracle, 5)
	ids := seedLibrary(me, "Top Instant", "Second Sorcery", "Third Creature", "Fourth Land", "Fifth Instant", "Sixth")
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		types := []string{"Instant", "Sorcery", "Creature — Bear", "Land", "Instant", "Instant"}
		for i := len(types) - 1; i >= 0; i-- {
			me.Library.PushTop(game.Card{InstanceID: ids[i], Name: "Card", TypeLine: types[i], Owner: me.ID, Controller: me.ID})
		}
	})
	b16Activate(t, g, me.ID, w, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	for _, id := range ids[:5] {
		if !g.Exile.Contains(id) {
			t.Fatalf("%s was not exiled", id)
		}
	}
	if g.Exile.Contains(ids[5]) {
		t.Fatal("exiled a sixth card")
	}
	if err := g.CastSpell(me.ID, ids[2], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast the creature from exile")
	}
	if err := g.CastSpell(me.ID, ids[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, ids[1], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast a second spell: it says an instant or sorcery spell")
	}
}

func TestChandraHopesBeaconCopiesTheFirstInstantOrSorceryEachTurn(t *testing.T) {
	g, me, _ := momB19Walker(t, "Chandra, Hope's Beacon", momB19ChandraOracle, 5)
	count := func() int {
		n := 0
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemSpell && it.Controller == me.ID {
				n++
			}
		}
		return n
	}
	castCatalogSpell(t, g, "First Sorcery", "Sorcery", "", nil)
	g.RunStateChecksForTest()
	passOnceForTrigger(t, g)
	if got := count(); got != 2 {
		t.Fatalf("%d spells on the stack after the first cast, want 2 (the spell and its copy)", got)
	}
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Second Sorcery", "Sorcery", "", nil)
	g.RunStateChecksForTest()
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemTriggered {
			t.Error("a second trigger on the same turn")
		}
	}
}

func passOnceForTrigger(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 4; i++ {
		top := false
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemTriggered {
				top = true
			}
		}
		if !top {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

// --- Mirror-Style Master ---

func TestMirrorStyleMasterCopiesEachAttackingModifiedCreatureAndExilesThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	master := b12Push(g, me.ID, "Mirror-Style Master", "Creature — Human Warrior", momB19MirrorOracle, 3, 3)
	modified := b12Creature(g, me.ID, "Modified Bear", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Wolf", "Creature — Wolf", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(modified, game.CounterPlusOne, 1) })
	declareAttack(t, g, opp.ID, master, modified, plain)
	passAndSettleAttackSplit(t, g, me.ID)
	var tokens []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() {
			tokens = append(tokens, c.InstanceID)
			if c.Name != "Modified Bear" {
				t.Errorf("a token copy of %q: only modified creatures are copied", c.Name)
			}
			if !c.Tapped || c.AttackingTarget == uuid.Nil {
				t.Errorf("the token is tapped=%v attacking=%v; want tapped and attacking", c.Tapped, c.AttackingTarget != uuid.Nil)
			}
		}
	}
	if len(tokens) != 1 {
		t.Fatalf("%d tokens, want 1", len(tokens))
	}
	if countersOn(g, tokens[0], game.CounterPlusOne) != 0 {
		t.Error("the copy carries the counter")
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tokens[0]) {
		t.Error("the token was not exiled at end of combat")
	}
}

func TestMirrorStyleMasterBackupGivesAnotherCreatureTheAttackAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Mirror-Style Master", "Creature — Human Warrior", momB19MirrorOracle, 3, 3)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if countersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Fatal("no counter on the Bear from backup")
	}
	// The Bear is modified now and carries the ability until end of turn.
	declareAttack(t, g, opp.ID, bear)
	passAndSettleAttackSplit(t, g, me.ID)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Bear" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d Bear token copies, want 1", n)
	}
}
