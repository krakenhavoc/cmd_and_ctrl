package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestBnsB41VexingGullHasFlashAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gull := effectstest.B12Push(g, me.ID, "Vexing Gull", "Creature — Bird", bns41GullOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, gull, "flying") || !effectstest.HasEffectiveKeyword(t, g, gull, "flash") {
		t.Error("Vexing Gull lacks flash or flying")
	}
}

func TestBnsB41ViciousConquistadorDrainsEachOpponentOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	conq := effectstest.B12Push(g, me.ID, "Vicious Conquistador", "Creature — Vampire Soldier", bns41ConquistadorOrcl, 1, 2)
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Life
	}
	effectstest.AttackWith(t, g, opp.ID, conq)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	voice := effectstest.B12Push(g, me.ID, "Voice of the Woods", "Creature — Elf", bns41VoiceOracle, 2, 2)
	elves := []uuid.UUID{voice}
	for i := 0; i < 4; i++ {
		elves = append(elves, effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1))
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", bns41CreatureTypeBears, 2, 2)
	short := append(append([]uuid.UUID{}, elves[:4]...), bear)
	if err := g.ActivateCatalogAbility(me.ID, voice, 0, game.ActivateAbilityParams{TapIDs: short, Strict: true}); err == nil {
		t.Fatal("a non-Elf paid the tap-five cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, voice, 0, game.ActivateAbilityParams{TapIDs: elves, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.Controller == me.ID {
			found = true
			if effectstest.EffectivePower(t, g, c.InstanceID) != 7 || effectstest.EffectiveToughness(t, g, c.InstanceID) != 7 ||
				!effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "trample") {
				t.Error("the Elemental is not a 7/7 with trample")
			}
		}
	}
	if !found {
		t.Fatal("no Elemental token")
	}
}

func TestBnsB41VoldarenDuelistStopsATargetFromBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	atk := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	wall := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 4)
	other := effectstest.PushVanillaCreature(g, opp.ID, "Other Wall", 0, 4)
	effectstest.CastCatalogSpell(t, g, "Voldaren Duelist", "Creature — Vampire Warrior", bns41DuelistOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, wall)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, atk); err == nil {
		t.Error("the targeted Wall can't block this turn")
	}
	if err := g.DeclareBlocker(other, atk); err != nil {
		t.Errorf("another creature still blocks: %v", err)
	}
}

func TestBnsB41VraskasFinisherDestroysOnlyADamagedOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hurt := effectstest.B12Creature(g, opp.ID, "Hurt Bear", bns41CreatureTypeBears, 3, 3)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh Bear", bns41CreatureTypeBears, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Hurt Bear", bns41CreatureTypeBears, 3, 3)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1)
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, mine, 1)
	})
	effectstest.CastCatalogSpell(t, g, "Vraska's Finisher", "Creature — Gorgon Assassin", bns41FinisherOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		// an undamaged or own creature is not a legal target
		effectstest.PickTriggerTarget(t, g, me.ID, hurt)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if zone, _ := effectstest.CardWhere(g, opp, hurt); zone != "graveyard" {
		t.Errorf("the damaged opposing creature is in %s, want the graveyard", zone)
	}
	if zone, _ := effectstest.CardWhere(g, opp, fresh); zone != "battlefield" {
		t.Errorf("the undamaged creature is in %s, want the battlefield", zone)
	}
	if zone, _ := effectstest.CardWhere(g, me, mine); zone != "battlefield" {
		t.Errorf("your own damaged creature is in %s, want the battlefield", zone)
	}
}

func TestBnsB41WailingGhoulMillsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Mm2Library(me, effectstest.Mm2Spell("A"), effectstest.Mm2Spell("B"), effectstest.Mm2Spell("C"))
	before := len(me.Graveyard.Cards)
	effectstest.CastCatalogSpell(t, g, "Wailing Ghoul", "Creature — Zombie", bns41GhoulOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Graveyard.Cards) - before; got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
	if len(me.Library.Cards) != 1 {
		t.Errorf("library %d, want 1", len(me.Library.Cards))
	}
}

func TestBnsB41WakedancerMakesAZombieOnlyAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", bns41CreatureTypeBears, 2, 2)
		if died {
			effectstest.KillCreature(t, g, opp.ID, bear)
			g.RunStateChecksForTest()
		}
		effectstest.CastCatalogSpell(t, g, "Wakedancer", "Creature — Human Shaman", bns41WakedancerOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if died {
			want = 1
		}
		if got := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); got != want {
			t.Errorf("died=%v: %d Zombies, want %d", died, got, want)
		}
	}
}

func TestBnsB41WakerOfWavesShrinksOnlyOpposingPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Waker of Waves", "Creature — Whale", bns41WakerOracle, 7, 7)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", bns41CreatureTypeBears, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", bns41CreatureTypeBears, 2, 2)
	if p, tough := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 1 || tough != 2 {
		t.Errorf("their Bear is %d/%d, want 1/2", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, mine); p != 2 {
		t.Errorf("my Bear power %d, want 2", p)
	}
}

func TestBnsB41WakerOfWavesChannelsFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	lib := effectstest.Mm2Library(me, effectstest.Mm2Spell("Top"), effectstest.Mm2Spell("Next"), effectstest.Mm2Spell("Deep"))
	waker := effectstest.PushCatalogHandCard(me, "Waker of Waves", "Creature — Whale", bns41WakerOracle)
	if err := effectstest.HobB4Activate(t, g, me.ID, waker, 0, "{U}{C}", nil); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if got := effectstest.ZoneKindOf(g, waker); got != game.ZoneGraveyard {
		t.Fatalf("the discarded card is in %v, want the graveyard", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shaman := effectstest.B12Push(g, me.ID, "Waterkin Shaman", "Creature — Elemental Shaman", bns41WaterkinOracle, 2, 1)
	effectstest.Mm2Library(me, effectstest.Mm2Spell("A"), effectstest.Mm2Spell("B"), effectstest.Mm2Spell("C"))
	effectstest.CastCatalogSpell(t, g, "Wailing Ghoul", "Creature — Zombie", bns41GhoulOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, shaman); p != 2 {
		t.Fatalf("a creature without flying pumped the Shaman: power %d", p)
	}
	effectstest.B12Push(g, me.ID, "Vexing Gull", "Creature — Bird", bns41GullOracle, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Vexing Gull", "Creature — Bird", bns41GullOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, shaman), effectstest.EffectiveToughness(t, g, shaman); p != 3 || tough != 2 {
		t.Errorf("the Shaman is %d/%d after a flier entered, want 3/2", p, tough)
	}
}

func TestBnsB41WeaselbackRedcapFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cap := effectstest.B12Push(g, me.ID, "Weaselback Redcap", "Creature — Goblin Knight", bns41WeaselbackOracle, 1, 1)
	if err := effectstest.HobB4Activate(t, g, me.ID, cap, 0, "{R}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, cap), effectstest.EffectiveToughness(t, g, cap); p != 3 || tough != 1 {
		t.Errorf("the Redcap is %d/%d, want 3/1", p, tough)
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, cap, 0, "", nil); err == nil {
		t.Error("activated with no mana")
	}
}
