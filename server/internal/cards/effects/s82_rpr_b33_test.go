package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b33_test.go — S82 reprint pass, batch 33: eleven cards through
// the real catalog.

const (
	rprAshenRider      = "455f66eb-3e33-4b51-9ec1-32288bdd3a68"
	rprBituminousBlast = "d291d8f7-bb67-470c-8794-cd553c91e5cc"
	rprChameleon       = "7b8f1458-c638-4907-9363-f7cc22d04605"
	rprCommune         = "f4aa0f64-d202-42a1-af62-0e53bbe87782"
	rprCrucible        = "ecfbebc9-7fc7-474e-8c59-8ede800e082e"
	rprCurator         = "04858b93-0a2a-44d0-942d-7fadf589cbb7"
	rprVerbosity       = "c6f76fa7-095e-4bfe-a38c-5c4531880880"
	rprDarkDweller     = "17917b7a-2b2b-4a96-a21e-10cd91375660"
	rprDeadMansChest   = "e5fdd6fd-6492-47ec-8b74-9ff22ff9849a"
	rprDecree          = "96a6a8d4-6077-470f-984e-b576a24557a2"
	rprDemandingDragon = "c8f1e3bd-bbf0-4750-b054-03553fc61850"
)

func TestRprB33Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		rprAshenRider: "Ashen Rider", rprBituminousBlast: "Bituminous Blast",
		rprChameleon: "Chameleon Colossus", rprCommune: "Commune with Lava",
		rprCrucible: "Crucible of the Spirit Dragon", rprCurator: "Curator of Mysteries",
		rprVerbosity: "Curse of Verbosity", rprDarkDweller: "Dark-Dweller Oracle",
		rprDeadMansChest: "Dead Man's Chest", rprDecree: "Decree of Savagery",
		rprDemandingDragon: "Demanding Dragon",
	} {
		stdRegisteredFull(t, oracle, name)
	}
}

func TestRprB33AshenRiderExilesOnEnterAndOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	land := blbB1Push(g, opp.ID, "Their Land", "Land", "", 0, 0)
	if !p1g2HasKeyword(g, blbB1Push(g, me.ID, "Probe", "Creature — Archon", rprAshenRider, 5, 5), "flying") {
		t.Error("Ashen Rider lacks flying")
	}
	// Enters: cast it.
	if _, err := blbHCastWithMana(t, g, "Ashen Rider", "Creature — Archon", rprAshenRider, "{4}{W}{W}{B}{B}", 5, 5,
		"W", "W", "B", "B", "W", "W", "W", "W"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: land}, true)
	if !g.Exile.Contains(land) {
		t.Fatal("the enter trigger did not exile its target")
	}

	// Dies.
	rider := blbB1Push(g, me.ID, "Ashen Rider", "Creature — Archon", rprAshenRider, 5, 5)
	other := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rider) })
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: other}, true)
	if !g.Exile.Contains(other) {
		t.Error("the dies trigger did not exile its target")
	}
}

func TestRprB33BituminousBlastDealsFourAndCascades(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	big := blbB1Push(g, opp.ID, "Big", "Creature — Beast", "", 4, 4)
	small := blbB1Push(g, opp.ID, "Small", "Creature — Beast", "", 4, 5)
	spec, _ := Lookup(rprBituminousBlast)
	if len(spec.Triggered) != 1 || spec.Triggered[0].Keyword != KeywordCascade {
		t.Fatalf("want exactly one cascade trigger, got %+v", spec.Triggered)
	}
	castCatalogSpell(t, g, "Bituminous Blast", "Instant", rprBituminousBlast, cardTarget(big))
	_ = me
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("4 damage did not kill the 4/4")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the untargeted creature died")
	}
}

func TestRprB33ChameleonColossusPumpsByItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	id := blbB1Push(g, me.ID, "Chameleon Colossus", "Creature — Shapeshifter", rprChameleon, 4, 4)
	for _, kw := range []string{game.KeywordChangeling, "protection from black"} {
		if !p1g2HasKeyword(g, id, kw) {
			t.Errorf("lacks %s", kw)
		}
	}
	advanceToMain(t, g)
	b06AddMana(me, "G", "G", "G", "G", "G", "G", "G", "G")
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
	}
	var p, tough int
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(id)
		p, tough = c.PowerForComparison(), c.CurrentToughness()
	})
	// 4 -> 8 -> 16.
	if p != 16 || tough != 16 {
		t.Errorf("got %d/%d, want 16/16 after two activations", p, tough)
	}
}

func TestRprB33CommuneWithLavaExilesXAndGrantsPlay(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	top := seedLibrary(me, "A", "B", "C", "D")
	advanceToMain(t, g)
	b06AddMana(me, "R", "R", "R", "R", "R")
	paCast(t, g, "Commune with Lava", "Instant", "{X}{R}{R}", rprCommune, game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		if !g.Exile.Contains(top[i]) {
			t.Errorf("card %d was not exiled", i)
		}
		if !permissionLive(g, exiledPermission(g, top[i]), me.ID) {
			t.Errorf("card %d is not playable", i)
		}
		if exiledPermission(g, top[i]).CastOnly {
			t.Errorf("card %d is cast-only, but the clause says play", i)
		}
	}
	if g.Exile.Contains(top[3]) {
		t.Error("a fourth card was exiled")
	}
}

func TestRprB33CrucibleStoresCountersAndPaysRestrictedMana(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	land := pushCatalogPermanent(g, me.ID, "Crucible of the Spirit Dragon", "Land", rprCrucible, false)
	advanceToMain(t, g)
	b06AddMana(me, "R")
	// {1},{T}: a storage counter.
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("store: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, land, game.CounterStorage); got != 1 {
		t.Fatalf("storage counters %d, want 1", got)
	}
	addCountersForTest(g, land, game.CounterStorage, 1)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, land).Tapped = false })
	activateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{CounterCounts: []int{2}})
	// Each of the two mana is its own pick of any colour.
	for i := 0; i < 2; i++ {
		var pick *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
				pick = c
			}
		}
		if pick == nil {
			break
		}
		if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool %d, want 2 mana for 2 counters removed", me.ManaPool.Total())
	}
	for _, tok := range me.ManaPool.Units() {
		if len(tok.Restrictions) == 0 {
			t.Error("the mana carries no spending restriction")
		}
	}
	if got := counterCount(g, land, game.CounterStorage); got != 0 {
		t.Errorf("storage counters left %d, want 0", got)
	}
}

func TestRprB33CuratorScriesOnceForACycleAndOnceForADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Curator of Mysteries", "Creature — Sphinx", rprCurator, 4, 4)
	seedLibrary(me, "A", "B", "C", "D")
	advanceToMain(t, g)
	s58HandCards(me, 1)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("a discard did not scry")
	}
	answerScryKeepAll(t, g, me.ID)
}

func TestRprB33CuratorCyclingIsOneScry(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Curator of Mysteries", "Creature — Sphinx", rprCurator, 4, 4)
	seedLibrary(me, "A", "B", "C", "D")
	cycleFromHand(t, g, "Curator of Mysteries", "Creature — Sphinx", rprCurator, "{U}")
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("cycling another card did not scry")
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("cycling scried twice")
	}
}

func TestRprB33CurseOfVerbosityDrawsForYouAndTheAttacker(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	curseCtl := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	for _, p := range []*game.Player{active, curseCtl} {
		seedLibrary(p, "A", "B", "C")
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Curse of Verbosity", TypeLine: "Enchantment — Aura Curse",
		OracleID: rprVerbosity, Owner: curseCtl.ID, Controller: curseCtl.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: victim.ID},
	})
	a1 := blbB1Push(g, active.ID, "Bear", "Creature — Bear", "", 2, 2)
	a2 := blbB1Push(g, active.ID, "Bear 2", "Creature — Bear", "", 2, 2)
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	hand, ctlHand := active.Hand.Size(), curseCtl.Hand.Size()
	for _, a := range []uuid.UUID{a1, a2} {
		if err := g.DeclareAttacker(a, victim.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := curseCtl.Hand.Size() - ctlHand; got != 1 {
		t.Errorf("the Curse's controller drew %d, want 1", got)
	}
	if got := active.Hand.Size() - hand; got != 1 {
		t.Errorf("the attacking opponent drew %d, want 1 (two attackers, one declaration)", got)
	}
}

func TestRprB33DarkDwellerOracleSacrificesForAnImpulse(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	oracle := blbB1Push(g, me.ID, "Dark-Dweller Oracle", "Creature — Goblin Shaman", rprDarkDweller, 2, 2)
	fodder := blbB1Push(g, me.ID, "Fodder", "Creature — Bear", "", 1, 1)
	top := seedLibrary(me, "Top", "Next")
	advanceToMain(t, g)
	b06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, oracle, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if !g.Exile.Contains(top[0]) || g.Exile.Contains(top[1]) {
		t.Fatal("exactly the top card should be exiled")
	}
	if !permissionLive(g, exiledPermission(g, top[0]), me.ID) {
		t.Error("the exiled card is not playable")
	}
	// Without the mana the activation is refused.
	if err := g.ActivateCatalogAbility(me.ID, oracle, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{oracle}}); err == nil {
		t.Error("activated with an empty pool")
	}
}

func TestRprB33DeadMansChestExilesItsPowerFromTheOwnersLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	victim := blbB1Push(g, opp.ID, "Their Brute", "Creature — Beast", "", 3, 3)
	top := seedLibrary(opp, "A", "B", "C", "D")
	mine := seedLibrary(me, "M1", "M2")
	castCatalogSpell(t, g, "Dead Man's Chest", "Enchantment — Aura", rprDeadMansChest, cardTarget(victim))
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	passPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		if !g.Exile.Contains(top[i]) {
			t.Errorf("card %d of the owner's library was not exiled", i)
		}
		perm := exiledPermission(g, top[i])
		if !permissionLive(g, perm, me.ID) || !perm.CastOnly || !perm.AnyType {
			t.Errorf("card %d: live=%v castOnly=%v anyType=%v", i, permissionLive(g, perm, me.ID), perm.CastOnly, perm.AnyType)
		}
	}
	if g.Exile.Contains(top[3]) || g.Exile.Contains(mine[0]) || g.Exile.Contains(mine[1]) {
		t.Error("exiled a card it should not have")
	}
}

func TestRprB33DeadMansChestRefusesYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Dead Man's Chest", "Enchantment — Aura", rprDeadMansChest, cardTarget(mine)); err == nil {
		t.Error("enchanted a creature its controller controls")
	}
}

func TestRprB33DecreeOfSavageryCountersAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	a := blbB1Push(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	b := blbB1Push(g, me.ID, "B", "Creature — Bear", "", 1, 1)
	theirs := blbB1Push(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Decree of Savagery", "Instant", rprDecree, nil)
	passPriorityAroundTable(t, g)
	if counterCount(g, a, game.CounterPlusOne) != 4 || counterCount(g, b, game.CounterPlusOne) != 4 {
		t.Error("each of your creatures should get four counters")
	}
	if counterCount(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("an opponent's creature got counters")
	}
}

func TestRprB33DecreeOfSavageryCycleTriggerIsOptional(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me, _, _ := blbB1Seats(g)
		bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		seedLibrary(me, "A", "B")
		hand := me.Hand.Size()
		cycleFromHand(t, g, "Decree of Savagery", "Instant", rprDecree, "{C}{C}{C}{C}{C}{C}")
		assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, yes)
		want := 0
		if yes {
			want = 4
		}
		if got := counterCount(g, bear, game.CounterPlusOne); got != want {
			t.Errorf("answer %v: %d counters, want %d", yes, got, want)
		}
		if got := me.Hand.Size() - hand; got != 1 {
			t.Errorf("cycling drew %d, want 1", got)
		}
	}
}

func demandingDragonTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	if _, err := blbHCastWithMana(t, g, "Demanding Dragon", "Creature — Dragon", rprDemandingDragon, "{3}{R}{R}", 5, 5,
		"R", "R", "R", "R", "R"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	return g, me, opp
}

func TestRprB33DemandingDragonDamagesAnOpponentWithNothingToSacrifice(t *testing.T) {
	g, me, opp := demandingDragonTable(t)
	life := lifeOf(g, opp.ID)
	passPriorityAroundTable(t, g)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	if got := lifeOf(g, opp.ID); got != life-5 {
		t.Errorf("life %d -> %d, want -5", life, got)
	}
}

func TestRprB33DemandingDragonLetsTheOpponentSacrificeInstead(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	if _, err := blbHCastWithMana(t, g, "Demanding Dragon", "Creature — Dragon", rprDemandingDragon, "{3}{R}{R}", 5, 5,
		"R", "R", "R", "R", "R"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	life := lifeOf(g, opp.ID)
	passPriorityAroundTable(t, g)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	answerOptionPick(t, g, opp.ID, 1)
	answerChooseCards(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent's creature was not sacrificed")
	}
	if got := lifeOf(g, opp.ID); got != life {
		t.Errorf("life %d -> %d, want unchanged", life, got)
	}
}

func TestRprB33DemandingDragonTakeTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	if _, err := blbHCastWithMana(t, g, "Demanding Dragon", "Creature — Dragon", rprDemandingDragon, "{3}{R}{R}", 5, 5,
		"R", "R", "R", "R", "R"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	life := lifeOf(g, opp.ID)
	passPriorityAroundTable(t, g)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	answerOptionPick(t, g, opp.ID, 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was sacrificed though the damage was taken")
	}
	if got := lifeOf(g, opp.ID); got != life-5 {
		t.Errorf("life %d -> %d, want -5", life, got)
	}
}
