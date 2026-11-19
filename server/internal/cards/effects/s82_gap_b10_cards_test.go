package effects

import (
	"testing"

	"github.com/google/uuid"

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
	gapB10Logger = iwCard{gapB10LoggerOracle, "Hinterland Logger", "Creature — Human Werewolf", "{1}{G}", iwFrontText, "2", "1",
		"Timber Shredder", "Trample\n" + iwBackText, "4", "2", []string{"G"}, []string{"Trample"}}
	gapB10Villagers = iwCard{gapB10VillagersOracle, "Villagers of Estwald", "Creature — Human Werewolf", "{2}{G}", iwFrontText, "2", "3",
		"Howlpack of Estwald", iwBackText, "4", "6", []string{"G"}, nil}
	gapB10Howler = iwCard{gapB10HowlerOracle, "Shrill Howler", "Creature — Werewolf Horror", "{2}{G}",
		"Creatures with power less than this creature's power can't block it.\n{5}{G}: Transform this creature.", "3", "1",
		"Howling Chorus", "Creatures with power less than this creature's power can't block it.\nWhenever this creature deals combat damage to a player, create a 3/2 colorless Eldrazi Horror creature token.", "3", "5",
		[]string{"G"}, nil}
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
		spec, ok := Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// --- Festerhide Boar ---

func TestGapB10BoarEntersWithCountersOnlyAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		if died {
			bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 1)
			killCreature(t, g, opp.ID, bear)
			g.RunStateChecksForTest()
		}
		castCatalogSpell(t, g, "Festerhide Boar", "Creature — Boar", gapB10BoarOracle, nil)
		passPriorityAroundTable(t, g)
		boar := battlefieldIDNamed(g, me.ID, "Festerhide Boar")
		want := 0
		if died {
			want = 2
		}
		if got := layeredCard(t, g, boar).Counters[game.CounterPlusOne]; got != want {
			t.Errorf("died=%v: %d +1/+1 counters, want %d", died, got, want)
		}
	}
}

// --- Ghoultree ---

func TestGapB10GhoultreeCostsLessPerCreatureCardInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	broB06FillGraveyard(g, me, 3)
	b06AddMana(me, "G", "C", "C", "C")
	if _, err := castFromHandErr(g, me, "Ghoultree", "Creature — Zombie Treefolk", "{7}{G}", gapB10GhoultreeOracle,
		game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("four mana cast a spell that costs {4}{G} with three creature cards in the graveyard")
	}
	b06AddMana(me, "C")
	if _, err := castFromHandErr(g, me, "Ghoultree", "Creature — Zombie Treefolk", "{7}{G}", gapB10GhoultreeOracle,
		game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("five mana should pay {4}{G}: %v", err)
	}
}

// --- Hamlet Captain ---

func TestGapB10HamletCaptainPumpsOtherHumansWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	captain := b12Push(g, me.ID, "Hamlet Captain", "Creature — Human Warrior", gapB10CaptainOracle, 2, 2)
	human := b12Creature(g, me.ID, "Villager", "Creature — Human Peasant", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Human", "Creature — Human Peasant", 1, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(captain, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"the Captain", captain, 2}, {"your other Human", human, 2}, {"your Bear", bear, 2}, {"their Human", theirs, 1}} {
		if got := layeredCard(t, g, tc.id).CurrentPower(); got != tc.want {
			t.Errorf("%s has power %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestGapB10HamletCaptainPumpsOtherHumansWhenItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me := g.Seats[g.Turn.ActiveSeat]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	captain := b12Push(g, opp.ID, "Hamlet Captain", "Creature — Human Warrior", gapB10CaptainOracle, 2, 2)
	human := b12Creature(g, opp.ID, "Villager", "Creature — Human Peasant", 1, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(captain, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, human).CurrentPower(); got != 2 {
		t.Errorf("the other Human has power %d, want 2", got)
	}
}

// --- Intrepid Provisioner ---

func TestGapB10ProvisionerPumpsAnotherHuman(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	human := b12Creature(g, me.ID, "Villager", "Creature — Human Peasant", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Intrepid Provisioner", "Creature — Human Scout", gapB10ProvOracle, nil)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, bear) {
			t.Error("a Bear is offered as the Human target")
		}
		if hasID(p.PickTargetCards, battlefieldIDNamed(g, me.ID, "Intrepid Provisioner")) {
			t.Error("the Provisioner itself is offered as 'another' Human")
		}
		pickTriggerTarget(t, g, me.ID, human)
		passPriorityAroundTable(t, g)
	}
	if got := layeredCard(t, g, human).CurrentPower(); got != 3 {
		t.Errorf("the Human has power %d, want 3", got)
	}
}

// --- Lumberknot ---

func TestGapB10LumberknotGrowsWhenAnyCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knot := b12Push(g, me.ID, "Lumberknot", "Creature — Treefolk", gapB10LumberOracle, 1, 1)
	assertKeywords(t, g, knot, "hexproof")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 1)
	killCreature(t, g, opp.ID, bear)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, knot).CurrentPower(); got != 2 {
		t.Errorf("power %d after one death, want 2", got)
	}
}

// --- Noose Constrictor ---

func TestGapB10NooseConstrictorDiscardsToGrow(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	noose := b12Push(g, me.ID, "Noose Constrictor", "Creature — Snake", gapB10NooseOracle, 2, 2)
	assertKeywords(t, g, noose, "reach")
	advanceToMain(t, g)
	card := pushCatalogHandCard(me, "Fodder", "Sorcery", "")
	if err := g.ActivateCatalogAbility(me.ID, noose, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := layeredCard(t, g, noose).CurrentPower(), layeredCard(t, g, noose).CurrentToughness(); p != 3 || tough != 3 {
		t.Errorf("Noose Constrictor is %d/%d, want 3/3", p, tough)
	}
	if me.Hand.Contains(card) {
		t.Error("the card was not discarded")
	}
}

// --- Pack Guardian ---

func TestGapB10PackGuardianMakesAWolfForADiscardedLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogHandCard(me, "Forest", "Basic Land — Forest", "")
	spell := pushCatalogHandCard(me, "Sorcery", "Sorcery", "")
	castHoldCreature(t, g, "Pack Guardian", "Creature — Wolf Spirit", gapB10GuardianOracle, 4, 3)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no pick of a land card")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != land {
		t.Fatalf("offered %v, want only the Forest %v (not the Sorcery %v)", pick.ChooseCards, land, spell)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(land) {
		t.Error("the land was not discarded")
	}
	if got := countBattlefieldNamed(g, me.ID, "Wolf"); got != 1 {
		t.Errorf("%d Wolf tokens, want 1", got)
	}
}

func TestGapB10PackGuardianMakesNoWolfWhenYouDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogHandCard(me, "Forest", "Basic Land — Forest", "")
	castHoldCreature(t, g, "Pack Guardian", "Creature — Wolf Spirit", gapB10GuardianOracle, 4, 3)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no pick of a land card")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) || countBattlefieldNamed(g, me.ID, "Wolf") != 0 {
		t.Error("declining kept neither the land nor the no-Wolf rule")
	}
}

func TestGapB10PackGuardianAsksNothingWithoutALandCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	pushCatalogHandCard(me, "Sorcery", "Sorcery", "")
	castHoldCreature(t, g, "Pack Guardian", "Creature — Wolf Spirit", gapB10GuardianOracle, 4, 3)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Error("asked to discard a land with none in hand")
	}
	assertKeywords(t, g, battlefieldIDNamed(g, me.ID, "Pack Guardian"), "flash")
}

// --- Young Wolf ---

func TestGapB10YoungWolfReturnsOnceWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wolf := b12Push(g, me.ID, "Young Wolf", "Creature — Wolf", gapB10WolfOracle, 1, 1)
	assertKeywords(t, g, wolf, "undying")
	killCreature(t, g, me.ID, wolf)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	back := battlefieldIDNamed(g, me.ID, "Young Wolf")
	if back == uuid.Nil {
		t.Fatal("Young Wolf did not return")
	}
	if got := layeredCard(t, g, back).CurrentPower(); got != 2 {
		t.Errorf("returned power %d, want 2 (a +1/+1 counter)", got)
	}
	killCreature(t, g, me.ID, back)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if battlefieldIDNamed(g, me.ID, "Young Wolf") != uuid.Nil {
		t.Error("Young Wolf returned a second time with a counter on it")
	}
}

// --- the werewolves ---

func TestGapB10WerewolvesTransformAfterAQuietTurnAndBack(t *testing.T) {
	for _, w := range []iwCard{gapB10Logger, gapB10Villagers, gapB10Howler} {
		if w.front == "Shrill Howler" {
			continue // its own turn-over is the {5}{G} ability
		}
		g, me := iwTable(t)
		id := iwPut(t, g, w, me)
		iwNextUpkeep(t, g, nil)
		if iwFace(g, id) != 1 {
			t.Fatalf("%s did not transform after a turn with no spells (face %d)", w.front, iwFace(g, id))
		}
		if w.front == "Hinterland Logger" {
			assertKeywords(t, g, id, "trample")
		}
		iwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 2})
		if iwFace(g, id) != 0 {
			t.Errorf("%s did not turn back after a player cast two spells", w.back)
		}
	}
}

func TestGapB10WerewolvesStayFrontAfterASpell(t *testing.T) {
	g, me := iwTable(t)
	id := iwPut(t, g, gapB10Logger, me)
	iwNextUpkeep(t, g, map[int]int{(g.Turn.ActiveSeat + 1) % len(g.Seats): 1})
	if iwFace(g, id) != 0 {
		t.Errorf("one spell last turn still turned the Logger over (face %d)", iwFace(g, id))
	}
}

func TestGapB10ShrillHowlerTransformsForSixAndChorusMakesHorrors(t *testing.T) {
	g, me := iwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := iwPut(t, g, gapB10Howler, me)
	if err := hobB4Activate(t, g, me.ID, id, 0, broB04Pool("{5}{G}"), nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if iwFace(g, id) != 1 {
		t.Fatalf("Shrill Howler is on face %d, want Howling Chorus", iwFace(g, id))
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Eldrazi Horror"); got != 1 {
		t.Errorf("%d Eldrazi Horror tokens, want 1", got)
	}
}

func TestGapB10ShrillHowlerCantBeBlockedByLowerPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	howler := b12Push(g, me.ID, "Shrill Howler", "Creature — Werewolf Horror", gapB10HowlerOracle, 3, 1)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(howler, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, howler); err == nil {
		t.Error("a creature with power 2 blocked a power 3 Howler")
	}
	if err := g.DeclareBlocker(big, howler); err != nil {
		t.Errorf("a creature with power 3 can block: %v", err)
	}
}
