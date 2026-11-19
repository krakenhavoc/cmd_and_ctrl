package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b41_cards_test.go — Foundations Jumpstart batch 41 (S82, #5929),
// played through the real catalog entries.

const (
	bns41GullOracle        = "1a36d2a8-6117-4828-bbe2-bc63212d29d8"
	bns41ConquistadorOrcl  = "26b077df-7a05-459a-b5ff-bf27767bb519"
	bns41VoiceOracle       = "12da8ba3-ca60-44c8-8627-1b4af2a2c4a3"
	bns41DuelistOracle     = "a96d0496-babc-4774-9b72-bea9d2b22743"
	bns41FinisherOracle    = "5df16529-6f1d-4978-b364-97d91abac441"
	bns41GhoulOracle       = "6ce24851-70d2-416c-8733-78928b4eee74"
	bns41WakedancerOracle  = "e850743e-e4c4-4a09-b797-d5f688b51451"
	bns41WakerOracle       = "3a19a959-ea83-404b-89f7-e5dbf4cb7995"
	bns41WaterkinOracle    = "64e0c580-0e55-4f42-8594-11ccb584d633"
	bns41WeaselbackOracle  = "b14bd895-1d4e-4bd5-9ac1-fed56afb6c90"
	bns41CreatureTypeBears = "Creature — Bear"
)

func TestBnsB41AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		bns41GullOracle: "Vexing Gull", bns41ConquistadorOrcl: "Vicious Conquistador",
		bns41VoiceOracle: "Voice of the Woods", bns41DuelistOracle: "Voldaren Duelist",
		bns41FinisherOracle: "Vraska's Finisher", bns41GhoulOracle: "Wailing Ghoul",
		bns41WakedancerOracle: "Wakedancer", bns41WakerOracle: "Waker of Waves",
		bns41WaterkinOracle: "Waterkin Shaman", bns41WeaselbackOracle: "Weaselback Redcap",
	} {
		spec, ok := Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestBnsB41VexingGullHasFlashAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gull := b12Push(g, me.ID, "Vexing Gull", "Creature — Bird", bns41GullOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, gull, "flying") || !hasEffectiveKeyword(t, g, gull, "flash") {
		t.Error("Vexing Gull lacks flash or flying")
	}
}

func TestBnsB41ViciousConquistadorDrainsEachOpponentOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	conq := b12Push(g, me.ID, "Vicious Conquistador", "Creature — Vampire Soldier", bns41ConquistadorOrcl, 1, 2)
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Life
	}
	attackWith(t, g, opp.ID, conq)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		switch s.ID {
		case me.ID:
			if s.Life != before[s.ID] {
				t.Errorf("the attacker lost life: %d -> %d", before[s.ID], s.Life)
			}
		case opp.ID:
			// 1 from the trigger and possibly 1 from combat damage, never less.
			if s.Life > before[s.ID]-1 {
				t.Errorf("%s did not lose life: %d -> %d", s.ID, before[s.ID], s.Life)
			}
		default:
			if s.Life != before[s.ID]-1 {
				t.Errorf("an uninvolved opponent life %d -> %d, want -1", before[s.ID], s.Life)
			}
		}
	}
}

func TestBnsB41VoiceOfTheWoodsTapsFiveElvesForASevenSevenTrampler(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	voice := b12Push(g, me.ID, "Voice of the Woods", "Creature — Elf", bns41VoiceOracle, 2, 2)
	elves := []uuid.UUID{voice}
	for i := 0; i < 4; i++ {
		elves = append(elves, b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1))
	}
	bear := b12Creature(g, me.ID, "Bear", bns41CreatureTypeBears, 2, 2)
	short := append(append([]uuid.UUID{}, elves[:4]...), bear)
	if err := g.ActivateCatalogAbility(me.ID, voice, 0, game.ActivateAbilityParams{TapIDs: short, Strict: true}); err == nil {
		t.Fatal("a non-Elf paid the tap-five cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, voice, 0, game.ActivateAbilityParams{TapIDs: elves, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.Controller == me.ID {
			found = true
			if effectivePower(t, g, c.InstanceID) != 7 || effectiveToughness(t, g, c.InstanceID) != 7 ||
				!hasEffectiveKeyword(t, g, c.InstanceID, "trample") {
				t.Error("the Elemental is not a 7/7 with trample")
			}
		}
	}
	if !found {
		t.Fatal("no Elemental token")
	}
}

func TestBnsB41VoldarenDuelistStopsATargetFromBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	atk := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 4)
	other := pushVanillaCreature(g, opp.ID, "Other Wall", 0, 4)
	castCatalogSpell(t, g, "Voldaren Duelist", "Creature — Vampire Warrior", bns41DuelistOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, wall)
		passPriorityAroundTable(t, g)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, atk); err == nil {
		t.Error("the targeted Wall can't block this turn")
	}
	if err := g.DeclareBlocker(other, atk); err != nil {
		t.Errorf("another creature still blocks: %v", err)
	}
}

func TestBnsB41VraskasFinisherDestroysOnlyADamagedOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hurt := b12Creature(g, opp.ID, "Hurt Bear", bns41CreatureTypeBears, 3, 3)
	fresh := b12Creature(g, opp.ID, "Fresh Bear", bns41CreatureTypeBears, 3, 3)
	mine := b12Creature(g, me.ID, "My Hurt Bear", bns41CreatureTypeBears, 3, 3)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1)
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, mine, 1)
	})
	castCatalogSpell(t, g, "Vraska's Finisher", "Creature — Gorgon Assassin", bns41FinisherOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		// an undamaged or own creature is not a legal target
		pickTriggerTarget(t, g, me.ID, hurt)
		passPriorityAroundTable(t, g)
	}
	if zone, _ := cardWhere(g, opp, hurt); zone != "graveyard" {
		t.Errorf("the damaged opposing creature is in %s, want the graveyard", zone)
	}
	if zone, _ := cardWhere(g, opp, fresh); zone != "battlefield" {
		t.Errorf("the undamaged creature is in %s, want the battlefield", zone)
	}
	if zone, _ := cardWhere(g, me, mine); zone != "battlefield" {
		t.Errorf("your own damaged creature is in %s, want the battlefield", zone)
	}
}

func TestBnsB41WailingGhoulMillsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mm2Library(me, mm2Spell("A"), mm2Spell("B"), mm2Spell("C"))
	before := len(me.Graveyard.Cards)
	castCatalogSpell(t, g, "Wailing Ghoul", "Creature — Zombie", bns41GhoulOracle, nil)
	passPriorityAroundTable(t, g)
	if got := len(me.Graveyard.Cards) - before; got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
	if len(me.Library.Cards) != 1 {
		t.Errorf("library %d, want 1", len(me.Library.Cards))
	}
}

func TestBnsB41WakedancerMakesAZombieOnlyAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		bear := b12Creature(g, opp.ID, "Their Bear", bns41CreatureTypeBears, 2, 2)
		if died {
			killCreature(t, g, opp.ID, bear)
			g.RunStateChecksForTest()
		}
		castCatalogSpell(t, g, "Wakedancer", "Creature — Human Shaman", bns41WakedancerOracle, nil)
		passPriorityAroundTable(t, g)
		want := 0
		if died {
			want = 1
		}
		if got := countBattlefieldNamed(g, me.ID, "Zombie"); got != want {
			t.Errorf("died=%v: %d Zombies, want %d", died, got, want)
		}
	}
}

func TestBnsB41WakerOfWavesShrinksOnlyOpposingPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Waker of Waves", "Creature — Whale", bns41WakerOracle, 7, 7)
	mine := b12Creature(g, me.ID, "My Bear", bns41CreatureTypeBears, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", bns41CreatureTypeBears, 2, 2)
	if p, tough := effectivePower(t, g, theirs), effectiveToughness(t, g, theirs); p != 1 || tough != 2 {
		t.Errorf("their Bear is %d/%d, want 1/2", p, tough)
	}
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("my Bear power %d, want 2", p)
	}
}

func TestBnsB41WakerOfWavesChannelsFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	lib := mm2Library(me, mm2Spell("Top"), mm2Spell("Next"), mm2Spell("Deep"))
	waker := pushCatalogHandCard(me, "Waker of Waves", "Creature — Whale", bns41WakerOracle)
	if err := hobB4Activate(t, g, me.ID, waker, 0, "{U}{C}", nil); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if got := zoneKindOf(g, waker); got != game.ZoneGraveyard {
		t.Fatalf("the discarded card is in %v, want the graveyard", got)
	}
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 2 {
		t.Fatal("no pick between the top two cards")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lib[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	// One of the top two reached the hand and the other the graveyard.
	inHand, inYard := 0, 0
	for _, id := range lib[:2] {
		if me.Hand.Contains(id) {
			inHand++
		}
		if me.Graveyard.Contains(id) {
			inYard++
		}
	}
	if inHand != 1 || inYard != 1 {
		t.Errorf("of the top two: %d in hand, %d in graveyard, want 1 and 1", inHand, inYard)
	}
}

func TestBnsB41WaterkinShamanGrowsWhenAFlierEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shaman := b12Push(g, me.ID, "Waterkin Shaman", "Creature — Elemental Shaman", bns41WaterkinOracle, 2, 1)
	mm2Library(me, mm2Spell("A"), mm2Spell("B"), mm2Spell("C"))
	castCatalogSpell(t, g, "Wailing Ghoul", "Creature — Zombie", bns41GhoulOracle, nil)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, shaman); p != 2 {
		t.Fatalf("a creature without flying pumped the Shaman: power %d", p)
	}
	b12Push(g, me.ID, "Vexing Gull", "Creature — Bird", bns41GullOracle, 2, 2)
	castCatalogSpell(t, g, "Vexing Gull", "Creature — Bird", bns41GullOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, shaman), effectiveToughness(t, g, shaman); p != 3 || tough != 2 {
		t.Errorf("the Shaman is %d/%d after a flier entered, want 3/2", p, tough)
	}
}

func TestBnsB41WeaselbackRedcapFirebreathes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cap := b12Push(g, me.ID, "Weaselback Redcap", "Creature — Goblin Knight", bns41WeaselbackOracle, 1, 1)
	if err := hobB4Activate(t, g, me.ID, cap, 0, "{R}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, cap), effectiveToughness(t, g, cap); p != 3 || tough != 1 {
		t.Errorf("the Redcap is %d/%d, want 3/1", p, tough)
	}
	if err := hobB4Activate(t, g, me.ID, cap, 0, "", nil); err == nil {
		t.Error("activated with no mana")
	}
}
