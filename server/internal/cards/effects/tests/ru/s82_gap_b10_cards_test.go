package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b10_cards_test.go — Innistrad Remastered creatures, S82 GAP
// batch 10 (#5948), played through the real catalog entries.

const (
	gapB10BoarOracle      = "be1fc2cf-00b8-424f-b683-3743fba285c9"
	gapB10GhoultreeOracle = "5f2628e3-e29a-4d77-bc27-6e1ce240d192"
	gapB10CaptainOracle   = "1a60dc0f-e387-4d28-8ff0-c3f1883d0c5d"
	gapB10LoggerOracle    = "29b22d41-4c6e-46de-8553-983d05f12f0f"
	gapB10ProvOracle      = "7d84873a-c079-4a0f-8878-c62e4bade20a"
	gapB10LumberOracle    = "6a9d6f32-c4a6-41ea-8188-235ea064dc53"
	gapB10NooseOracle     = "fb1d521f-cb64-4b87-b3f4-a74e91a60349"
	gapB10GuardianOracle  = "61ad6bd2-c079-44e7-a1d9-8089f086ec69"
	gapB10HowlerOracle    = "30c3f88f-11d3-4b5f-8936-bab1c65bd693"
	gapB10VillagersOracle = "dcf692f6-baec-4b60-a6e3-2a8b731d7816"
	gapB10WolfOracle      = "8b492764-10b6-4506-be11-22daa9220a91"
)

var (
	gapB10Logger = effectstest.IwCard{Oracle: gapB10LoggerOracle, Front: "Hinterland Logger", FrontType: "Creature — Human Werewolf", Cost: "{1}{G}", FrontText: effectstest.IwFrontText, Fp: "2", Ft: "1",
		Back: "Timber Shredder", BackText: "Trample\n" + effectstest.IwBackText, Bp: "4", Bt: "2", Colors: []string{"G"}, Keywords: []string{"Trample"}}
	gapB10Villagers = effectstest.IwCard{Oracle: gapB10VillagersOracle, Front: "Villagers of Estwald", FrontType: "Creature — Human Werewolf", Cost: "{2}{G}", FrontText: effectstest.IwFrontText, Fp: "2", Ft: "3",
		Back: "Howlpack of Estwald", BackText: effectstest.IwBackText, Bp: "4", Bt: "6", Colors: []string{"G"}, Keywords: nil}
	gapB10Howler = effectstest.IwCard{Oracle: gapB10HowlerOracle, Front: "Shrill Howler", FrontType: "Creature — Werewolf Horror", Cost: "{2}{G}",
		FrontText: "Creatures with power less than this creature's power can't block it.\n{5}{G}: Transform this creature.", Fp: "3", Ft: "1",
		Back: "Howling Chorus", BackText: "Creatures with power less than this creature's power can't block it.\nWhenever this creature deals combat damage to a player, create a 3/2 colorless Eldrazi Horror creature token.", Bp: "3", Bt: "5",
		Colors: []string{"G"}, Keywords: nil}
)

func TestGapB10EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		gapB10BoarOracle: "Festerhide Boar", gapB10GhoultreeOracle: "Ghoultree",
		gapB10CaptainOracle: "Hamlet Captain", gapB10LoggerOracle: "Hinterland Logger",
		gapB10LoggerOracle + "#1": "Timber Shredder", gapB10ProvOracle: "Intrepid Provisioner",
		gapB10LumberOracle: "Lumberknot", gapB10NooseOracle: "Noose Constrictor",
		gapB10GuardianOracle: "Pack Guardian", gapB10HowlerOracle: "Shrill Howler",
		gapB10HowlerOracle + "#1": "Howling Chorus", gapB10VillagersOracle: "Villagers of Estwald",
		gapB10VillagersOracle + "#1": "Howlpack of Estwald", gapB10WolfOracle: "Young Wolf",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// --- Festerhide Boar ---

func TestGapB10BoarEntersWithCountersOnlyAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		if died {
			bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 1)
			effectstest.KillCreature(t, g, opp.ID, bear)
			g.RunStateChecksForTest()
		}
		effectstest.CastCatalogSpell(t, g, "Festerhide Boar", "Creature — Boar", gapB10BoarOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		boar := effectstest.BattlefieldIDNamed(g, me.ID, "Festerhide Boar")
		want := 0
		if died {
			want = 2
		}
		if got := effectstest.LayeredCard(t, g, boar).Counters[game.CounterPlusOne]; got != want {
			t.Errorf("died=%v: %d +1/+1 counters, want %d", died, got, want)
		}
	}
}

// --- Ghoultree ---

func TestGapB10GhoultreeCostsLessPerCreatureCardInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	broB06FillGraveyard(g, me, 3)
	effectstest.B06AddMana(me, "G", "C", "C", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Ghoultree", "Creature — Zombie Treefolk", "{7}{G}", gapB10GhoultreeOracle,
		game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("four mana cast a spell that costs {4}{G} with three creature cards in the graveyard")
	}
	effectstest.B06AddMana(me, "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Ghoultree", "Creature — Zombie Treefolk", "{7}{G}", gapB10GhoultreeOracle,
		game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("five mana should pay {4}{G}: %v", err)
	}
}

// --- Hamlet Captain ---

func TestGapB10HamletCaptainPumpsOtherHumansWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	captain := effectstest.B12Push(g, me.ID, "Hamlet Captain", "Creature — Human Warrior", gapB10CaptainOracle, 2, 2)
	human := effectstest.B12Creature(g, me.ID, "Villager", "Creature — Human Peasant", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Human", "Creature — Human Peasant", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(captain, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"the Captain", captain, 2}, {"your other Human", human, 2}, {"your Bear", bear, 2}, {"their Human", theirs, 1}} {
		if got := effectstest.LayeredCard(t, g, tc.id).CurrentPower(); got != tc.want {
			t.Errorf("%s has power %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestGapB10HamletCaptainPumpsOtherHumansWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me := g.Seats[g.Turn.ActiveSeat]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	captain := effectstest.B12Push(g, opp.ID, "Hamlet Captain", "Creature — Human Warrior", gapB10CaptainOracle, 2, 2)
	human := effectstest.B12Creature(g, opp.ID, "Villager", "Creature — Human Peasant", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(captain, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, human).CurrentPower(); got != 2 {
		t.Errorf("the other Human has power %d, want 2", got)
	}
}

// --- Intrepid Provisioner ---

func TestGapB10ProvisionerPumpsAnotherHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	human := effectstest.B12Creature(g, me.ID, "Villager", "Creature — Human Peasant", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Intrepid Provisioner", "Creature — Human Scout", gapB10ProvOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, bear) {
			t.Error("a Bear is offered as the Human target")
		}
		if effectstest.HasID(p.PickTargetCards, effectstest.BattlefieldIDNamed(g, me.ID, "Intrepid Provisioner")) {
			t.Error("the Provisioner itself is offered as 'another' Human")
		}
		effectstest.PickTriggerTarget(t, g, me.ID, human)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.LayeredCard(t, g, human).CurrentPower(); got != 3 {
		t.Errorf("the Human has power %d, want 3", got)
	}
}

// --- Lumberknot ---

func TestGapB10LumberknotGrowsWhenAnyCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knot := effectstest.B12Push(g, me.ID, "Lumberknot", "Creature — Treefolk", gapB10LumberOracle, 1, 1)
	effectstest.AssertKeywords(t, g, knot, "hexproof")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.KillCreature(t, g, opp.ID, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, knot).CurrentPower(); got != 2 {
		t.Errorf("power %d after one death, want 2", got)
	}
}

// --- Noose Constrictor ---

func TestGapB10NooseConstrictorDiscardsToGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	noose := effectstest.B12Push(g, me.ID, "Noose Constrictor", "Creature — Snake", gapB10NooseOracle, 2, 2)
	effectstest.AssertKeywords(t, g, noose, "reach")
	effectstest.AdvanceToMain(t, g)
	card := effectstest.PushCatalogHandCard(me, "Fodder", "Sorcery", "")
	if err := g.ActivateCatalogAbility(me.ID, noose, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.LayeredCard(t, g, noose).CurrentPower(), effectstest.LayeredCard(t, g, noose).CurrentToughness(); p != 3 || tough != 3 {
		t.Errorf("Noose Constrictor is %d/%d, want 3/3", p, tough)
	}
	if me.Hand.Contains(card) {
		t.Error("the card was not discarded")
	}
}

// --- Pack Guardian ---

func TestGapB10PackGuardianMakesAWolfForADiscardedLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushCatalogHandCard(me, "Forest", "Basic Land — Forest", "")
	spell := effectstest.PushCatalogHandCard(me, "Sorcery", "Sorcery", "")
	effectstest.CastHoldCreature(t, g, "Pack Guardian", "Creature — Wolf Spirit", gapB10GuardianOracle, 4, 3)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no pick of a land card")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != land {
		t.Fatalf("offered %v, want only the Forest %v (not the Sorcery %v)", pick.ChooseCards, land, spell)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(land) {
		t.Error("the land was not discarded")
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Wolf"); got != 1 {
		t.Errorf("%d Wolf tokens, want 1", got)
	}
}

func TestGapB10PackGuardianMakesNoWolfWhenYouDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushCatalogHandCard(me, "Forest", "Basic Land — Forest", "")
	effectstest.CastHoldCreature(t, g, "Pack Guardian", "Creature — Wolf Spirit", gapB10GuardianOracle, 4, 3)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no pick of a land card")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) || effectstest.CountBattlefieldNamed(g, me.ID, "Wolf") != 0 {
		t.Error("declining kept neither the land nor the no-Wolf rule")
	}
}

func TestGapB10PackGuardianAsksNothingWithoutALandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.PushCatalogHandCard(me, "Sorcery", "Sorcery", "")
	effectstest.CastHoldCreature(t, g, "Pack Guardian", "Creature — Wolf Spirit", gapB10GuardianOracle, 4, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Error("asked to discard a land with none in hand")
	}
	effectstest.AssertKeywords(t, g, effectstest.BattlefieldIDNamed(g, me.ID, "Pack Guardian"), "flash")
}

// --- Young Wolf ---

func TestGapB10YoungWolfReturnsOnceWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wolf := effectstest.B12Push(g, me.ID, "Young Wolf", "Creature — Wolf", gapB10WolfOracle, 1, 1)
	effectstest.AssertKeywords(t, g, wolf, "undying")
	effectstest.KillCreature(t, g, me.ID, wolf)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.BattlefieldIDNamed(g, me.ID, "Young Wolf")
	if back == uuid.Nil {
		t.Fatal("Young Wolf did not return")
	}
	if got := effectstest.LayeredCard(t, g, back).CurrentPower(); got != 2 {
		t.Errorf("returned power %d, want 2 (a +1/+1 counter)", got)
	}
	effectstest.KillCreature(t, g, me.ID, back)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldIDNamed(g, me.ID, "Young Wolf") != uuid.Nil {
		t.Error("Young Wolf returned a second time with a counter on it")
	}
}

// --- the werewolves ---

func TestGapB10WerewolvesTransformAfterAQuietTurnAndBack(t *testing.T) {
	for _, w := range []effectstest.IwCard{gapB10Logger, gapB10Villagers, gapB10Howler} {
		if w.Front == "Shrill Howler" {
			continue // its own turn-over is the {5}{G} ability
		}
		g, me := effectstest.IwTable(t)
		id := effectstest.IwPut(t, g, w, me)
		effectstest.IwNextUpkeep(t, g, nil)
		if effectstest.IwFace(g, id) != 1 {
			t.Fatalf("%s did not transform after a turn with no spells (face %d)", w.Front, effectstest.IwFace(g, id))
		}
		if w.Front == "Hinterland Logger" {
			effectstest.AssertKeywords(t, g, id, "trample")
		}
		effectstest.IwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 2})
		if effectstest.IwFace(g, id) != 0 {
			t.Errorf("%s did not turn back after a player cast two spells", w.Back)
		}
	}
}

func TestGapB10WerewolvesStayFrontAfterASpell(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, gapB10Logger, me)
	effectstest.IwNextUpkeep(t, g, map[int]int{(g.Turn.ActiveSeat + 1) % len(g.Seats): 1})
	if effectstest.IwFace(g, id) != 0 {
		t.Errorf("one spell last turn still turned the Logger over (face %d)", effectstest.IwFace(g, id))
	}
}

func TestGapB10ShrillHowlerTransformsForSixAndChorusMakesHorrors(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.IwPut(t, g, gapB10Howler, me)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, effectstest.BroB04Pool("{5}{G}"), nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatalf("Shrill Howler is on face %d, want Howling Chorus", effectstest.IwFace(g, id))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Eldrazi Horror"); got != 1 {
		t.Errorf("%d Eldrazi Horror tokens, want 1", got)
	}
}

func TestGapB10ShrillHowlerCantBeBlockedByLowerPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	howler := effectstest.B12Push(g, me.ID, "Shrill Howler", "Creature — Werewolf Horror", gapB10HowlerOracle, 3, 1)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(howler, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, howler); err == nil {
		t.Error("a creature with power 2 blocked a power 3 Howler")
	}
	if err := g.DeclareBlocker(big, howler); err != nil {
		t.Errorf("a creature with power 3 can block: %v", err)
	}
}
