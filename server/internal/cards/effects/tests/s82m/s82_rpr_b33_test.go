package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestRprB33AshenRiderExilesOnEnterAndOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	land := effectstest.BlbB1Push(g, opp.ID, "Their Land", "Land", "", 0, 0)
	if !effectstest.P1g2HasKeyword(g, effectstest.BlbB1Push(g, me.ID, "Probe", "Creature — Archon", rprAshenRider, 5, 5), "flying") {
		t.Error("Ashen Rider lacks flying")
	}
	// Enters: cast it.
	if _, err := effectstest.BlbHCastWithMana(t, g, "Ashen Rider", "Creature — Archon", rprAshenRider, "{4}{W}{W}{B}{B}", 5, 5,
		"W", "W", "B", "B", "W", "W", "W", "W"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: land}, true)
	if !g.Exile.Contains(land) {
		t.Fatal("the enter trigger did not exile its target")
	}

	// Dies.
	rider := effectstest.BlbB1Push(g, me.ID, "Ashen Rider", "Creature — Archon", rprAshenRider, 5, 5)
	other := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rider) })
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: other}, true)
	if !g.Exile.Contains(other) {
		t.Error("the dies trigger did not exile its target")
	}
}

func TestRprB33BituminousBlastDealsFourAndCascades(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	big := effectstest.BlbB1Push(g, opp.ID, "Big", "Creature — Beast", "", 4, 4)
	small := effectstest.BlbB1Push(g, opp.ID, "Small", "Creature — Beast", "", 4, 5)
	spec, _ := effects.Lookup(rprBituminousBlast)
	if len(spec.Triggered) != 1 || spec.Triggered[0].Keyword != effects.KeywordCascade {
		t.Fatalf("want exactly one cascade trigger, got %+v", spec.Triggered)
	}
	effectstest.CastCatalogSpell(t, g, "Bituminous Blast", "Instant", rprBituminousBlast, effectstest.CardTarget(big))
	_ = me
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("4 damage did not kill the 4/4")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the untargeted creature died")
	}
}

func TestRprB33ChameleonColossusPumpsByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	id := effectstest.BlbB1Push(g, me.ID, "Chameleon Colossus", "Creature — Shapeshifter", rprChameleon, 4, 4)
	for _, kw := range []string{game.KeywordChangeling, "protection from black"} {
		if !effectstest.P1g2HasKeyword(g, id, kw) {
			t.Errorf("lacks %s", kw)
		}
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G", "G", "G", "G", "G", "G", "G", "G")
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	top := effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R", "R", "R", "R")
	effectstest.PaCast(t, g, "Commune with Lava", "Instant", "{X}{R}{R}", rprCommune, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		if !g.Exile.Contains(top[i]) {
			t.Errorf("card %d was not exiled", i)
		}
		if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top[i]), me.ID) {
			t.Errorf("card %d is not playable", i)
		}
		if effectstest.ExiledPermission(g, top[i]).CastOnly {
			t.Errorf("card %d is cast-only, but the clause says play", i)
		}
	}
	if g.Exile.Contains(top[3]) {
		t.Error("a fourth card was exiled")
	}
}

func TestRprB33CrucibleStoresCountersAndPaysRestrictedMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Crucible of the Spirit Dragon", "Land", rprCrucible, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R")
	// {1},{T}: a storage counter.
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("store: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, land, game.CounterStorage); got != 1 {
		t.Fatalf("storage counters %d, want 1", got)
	}
	effectstest.AddCountersForTest(g, land, game.CounterStorage, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, land).Tapped = false })
	effectstest.ActivateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{CounterCounts: []int{2}})
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
	if got := effectstest.CounterCount(g, land, game.CounterStorage); got != 0 {
		t.Errorf("storage counters left %d, want 0", got)
	}
}

func TestRprB33CuratorScriesOnceForACycleAndOnceForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Curator of Mysteries", "Creature — Sphinx", rprCurator, 4, 4)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	effectstest.S58HandCards(me, 1)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("a discard did not scry")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

func TestRprB33CuratorCyclingIsOneScry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Curator of Mysteries", "Creature — Sphinx", rprCurator, 4, 4)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.CycleFromHand(t, g, "Curator of Mysteries", "Creature — Sphinx", rprCurator, "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("cycling another card did not scry")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Error("cycling scried twice")
	}
}

func TestRprB33CurseOfVerbosityDrawsForYouAndTheAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	curseCtl := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	for _, p := range []*game.Player{active, curseCtl} {
		effectstest.SeedLibrary(p, "A", "B", "C")
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Curse of Verbosity", TypeLine: "Enchantment — Aura Curse",
		OracleID: rprVerbosity, Owner: curseCtl.ID, Controller: curseCtl.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: victim.ID},
	})
	a1 := effectstest.BlbB1Push(g, active.ID, "Bear", "Creature — Bear", "", 2, 2)
	a2 := effectstest.BlbB1Push(g, active.ID, "Bear 2", "Creature — Bear", "", 2, 2)
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
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := curseCtl.Hand.Size() - ctlHand; got != 1 {
		t.Errorf("the Curse's controller drew %d, want 1", got)
	}
	if got := active.Hand.Size() - hand; got != 1 {
		t.Errorf("the attacking opponent drew %d, want 1 (two attackers, one declaration)", got)
	}
}

func TestRprB33DarkDwellerOracleSacrificesForAnImpulse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	oracle := effectstest.BlbB1Push(g, me.ID, "Dark-Dweller Oracle", "Creature — Goblin Shaman", rprDarkDweller, 2, 2)
	fodder := effectstest.BlbB1Push(g, me.ID, "Fodder", "Creature — Bear", "", 1, 1)
	top := effectstest.SeedLibrary(me, "Top", "Next")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, oracle, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if !g.Exile.Contains(top[0]) || g.Exile.Contains(top[1]) {
		t.Fatal("exactly the top card should be exiled")
	}
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, top[0]), me.ID) {
		t.Error("the exiled card is not playable")
	}
	// Without the mana the activation is refused.
	if err := g.ActivateCatalogAbility(me.ID, oracle, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{oracle}}); err == nil {
		t.Error("activated with an empty pool")
	}
}

func TestRprB33DeadMansChestExilesItsPowerFromTheOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	victim := effectstest.BlbB1Push(g, opp.ID, "Their Brute", "Creature — Beast", "", 3, 3)
	top := effectstest.SeedLibrary(opp, "A", "B", "C", "D")
	mine := effectstest.SeedLibrary(me, "M1", "M2")
	effectstest.CastCatalogSpell(t, g, "Dead Man's Chest", "Enchantment — Aura", rprDeadMansChest, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ {
		if !g.Exile.Contains(top[i]) {
			t.Errorf("card %d of the owner's library was not exiled", i)
		}
		perm := effectstest.ExiledPermission(g, top[i])
		if !effectstest.PermissionLive(g, perm, me.ID) || !perm.CastOnly || !perm.AnyType {
			t.Errorf("card %d: live=%v castOnly=%v anyType=%v", i, effectstest.PermissionLive(g, perm, me.ID), perm.CastOnly, perm.AnyType)
		}
	}
	if g.Exile.Contains(top[3]) || g.Exile.Contains(mine[0]) || g.Exile.Contains(mine[1]) {
		t.Error("exiled a card it should not have")
	}
}

func TestRprB33DeadMansChestRefusesYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.BlbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Dead Man's Chest", "Enchantment — Aura", rprDeadMansChest, effectstest.CardTarget(mine)); err == nil {
		t.Error("enchanted a creature its controller controls")
	}
}

func TestRprB33DecreeOfSavageryCountersAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	a := effectstest.BlbB1Push(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	b := effectstest.BlbB1Push(g, me.ID, "B", "Creature — Bear", "", 1, 1)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Decree of Savagery", "Instant", rprDecree, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, a, game.CounterPlusOne) != 4 || effectstest.CounterCount(g, b, game.CounterPlusOne) != 4 {
		t.Error("each of your creatures should get four counters")
	}
	if effectstest.CounterCount(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("an opponent's creature got counters")
	}
}

func TestRprB33DecreeOfSavageryCycleTriggerIsOptional(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, _, _ := effectstest.BlbB1Seats(g)
		bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		effectstest.SeedLibrary(me, "A", "B")
		hand := me.Hand.Size()
		effectstest.CycleFromHand(t, g, "Decree of Savagery", "Instant", rprDecree, "{C}{C}{C}{C}{C}{C}")
		effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, yes)
		want := 0
		if yes {
			want = 4
		}
		if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != want {
			t.Errorf("answer %v: %d counters, want %d", yes, got, want)
		}
		if got := me.Hand.Size() - hand; got != 1 {
			t.Errorf("cycling drew %d, want 1", got)
		}
	}
}

func demandingDragonTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	if _, err := effectstest.BlbHCastWithMana(t, g, "Demanding Dragon", "Creature — Dragon", rprDemandingDragon, "{3}{R}{R}", 5, 5,
		"R", "R", "R", "R", "R"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	return g, me, opp
}

func TestRprB33DemandingDragonDamagesAnOpponentWithNothingToSacrifice(t *testing.T) {
	g, me, opp := demandingDragonTable(t)
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	if got := effectstest.LifeOf(g, opp.ID); got != life-5 {
		t.Errorf("life %d -> %d, want -5", life, got)
	}
}

func TestRprB33DemandingDragonLetsTheOpponentSacrificeInstead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	if _, err := effectstest.BlbHCastWithMana(t, g, "Demanding Dragon", "Creature — Dragon", rprDemandingDragon, "{3}{R}{R}", 5, 5,
		"R", "R", "R", "R", "R"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	effectstest.AnswerChooseCards(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent's creature was not sacrificed")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life {
		t.Errorf("life %d -> %d, want unchanged", life, got)
	}
}

func TestRprB33DemandingDragonTakeTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	if _, err := effectstest.BlbHCastWithMana(t, g, "Demanding Dragon", "Creature — Dragon", rprDemandingDragon, "{3}{R}{R}", 5, 5,
		"R", "R", "R", "R", "R"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was sacrificed though the damage was taken")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life-5 {
		t.Errorf("life %d -> %d, want -5", life, got)
	}
}
