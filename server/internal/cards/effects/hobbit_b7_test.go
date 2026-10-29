package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB) cards, half B, batch 7. Every test drives the
// registered catalog row through the engine.

const (
	hobB7Dwarf    = "Creature — Dwarf Warrior"
	hobB7Artifact = "Artifact"
)

func hobB7Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[game.CounterPlusOne]
}

func hobB7Player(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func hobB7Seed(g *game.Game, owner uuid.UUID, name, typeLine string, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, seedPermanent(g, owner, name, typeLine))
	}
	return ids
}

// --- Iron Hills / Goblin-town ------------------------------------------

func TestHobIronHillsEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := b13PlayAs(t, g, 0, "Iron Hills", "Land", hobOracle(t, "Iron Hills"))
	if !hobB5Tapped(t, g, id) {
		t.Error("Iron Hills entered untapped")
	}
}

func TestHobIronHillsTapsForRedOrWhiteOnly(t *testing.T) {
	for _, tc := range []struct {
		color string
		ok    bool
	}{{"R", true}, {"W", true}, {"B", false}, {"G", false}} {
		t.Run(tc.color, func(t *testing.T) {
			g, me, _ := spendTable(t)
			hills := pushCatalogPermanent(g, me.ID, "Iron Hills", "Land", hobOracle(t, "Iron Hills"), false)
			err := g.ActivateManaAbility(me.ID, hills, 0, game.ManaAbilityParams{Colors: []string{tc.color}})
			if tc.ok != (err == nil) {
				t.Fatalf("tap for %s: err = %v, want ok = %v", tc.color, err, tc.ok)
			}
			if tc.ok {
				if units := me.ManaPool.Units(); len(units) != 1 || units[0].Color != tc.color {
					t.Errorf("pool = %+v, want one %s", me.ManaPool, tc.color)
				}
			}
		})
	}
}

func TestHobIronHillsSacrificesForTwoCountersOnADwarf(t *testing.T) {
	g, me, _ := spendTable(t)
	hills := pushCatalogPermanent(g, me.ID, "Iron Hills", "Land", hobOracle(t, "Iron Hills"), false)
	dwarf := hobB4Creature(g, me.ID, "Test Dwarf", hobB7Dwarf, 2, 2)
	if err := hobB4Activate(t, g, me.ID, hills, 0, "{C}{C}{R}", cardRefs(dwarf)); err == nil {
		t.Fatal("activated without {W}")
	}
	if err := hobB4Activate(t, g, me.ID, hills, 0, "{W}", cardRefs(dwarf)); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, hills) {
		t.Error("Iron Hills was not sacrificed")
	}
	if !me.Graveyard.Contains(hills) {
		t.Error("Iron Hills is not in the graveyard")
	}
	if got := hobB7Counters(t, g, dwarf); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

func TestHobIronHillsOnlyTargetsYourDwarves(t *testing.T) {
	g, me, opp := spendTable(t)
	hills := pushCatalogPermanent(g, me.ID, "Iron Hills", "Land", hobOracle(t, "Iron Hills"), false)
	bear := seedBear(g, me.ID)
	theirs := hobB4Creature(g, opp.ID, "Their Dwarf", hobB7Dwarf, 2, 2)
	for name, target := range map[string]uuid.UUID{"a non-Dwarf": bear, "an opponent's Dwarf": theirs} {
		if err := hobB4Activate(t, g, me.ID, hills, 0, "{C}{C}{R}{W}", cardRefs(target)); err == nil {
			t.Errorf("Iron Hills targeted %s", name)
		}
		if !battlefieldHas(g, hills) {
			t.Fatalf("a refused activation sacrificed the land (%s)", name)
		}
	}
}

func TestHobIronHillsIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hills := pushCatalogPermanent(g, me.ID, "Iron Hills", "Land", hobOracle(t, "Iron Hills"), false)
	dwarf := hobB4Creature(g, me.ID, "Test Dwarf", hobB7Dwarf, 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := hobB4Activate(t, g, me.ID, hills, 0, "{C}{C}{R}{W}", cardRefs(dwarf)); err == nil {
		t.Fatal("Iron Hills activated outside a main phase")
	}
	if !battlefieldHas(g, hills) {
		t.Error("a refused activation sacrificed the land")
	}
}

func TestHobIronHillsDoesNothingIfTheDwarfIsGone(t *testing.T) {
	g, me, _ := spendTable(t)
	hills := pushCatalogPermanent(g, me.ID, "Iron Hills", "Land", hobOracle(t, "Iron Hills"), false)
	dwarf := hobB4Creature(g, me.ID, "Test Dwarf", hobB7Dwarf, 2, 2)
	if err := hobB4Activate(t, g, me.ID, hills, 0, "{C}{C}{R}{W}", cardRefs(dwarf)); err != nil {
		t.Fatalf("activation: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(dwarf); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, hills) {
		t.Error("the land is a cost: it stays sacrificed")
	}
}

func TestHobGoblinTownEntersTappedAndTapsForBlackOrRed(t *testing.T) {
	g := newCatalogGame(t)
	id := b13PlayAs(t, g, 0, "Goblin-town", "Land", hobOracle(t, "Goblin-town"))
	if !hobB5Tapped(t, g, id) {
		t.Error("Goblin-town entered untapped")
	}
	for _, tc := range []struct {
		color string
		ok    bool
	}{{"B", true}, {"R", true}, {"W", false}, {"U", false}} {
		t.Run(tc.color, func(t *testing.T) {
			g, me, _ := spendTable(t)
			town := pushCatalogPermanent(g, me.ID, "Goblin-town", "Land", hobOracle(t, "Goblin-town"), false)
			err := g.ActivateManaAbility(me.ID, town, 0, game.ManaAbilityParams{Colors: []string{tc.color}})
			if tc.ok != (err == nil) {
				t.Fatalf("tap for %s: err = %v, want ok = %v", tc.color, err, tc.ok)
			}
		})
	}
}

func TestHobGoblinTownCountersOnAGoblinOrOrcOnly(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		ok             bool
	}{
		{"goblin", "Creature — Goblin Warrior", true},
		{"orc", "Creature — Orc Soldier", true},
		{"goblin orc", "Creature — Goblin Orc", true},
		{"elf", "Creature — Elf Druid", false},
		{"dwarf", hobB7Dwarf, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := spendTable(t)
			town := pushCatalogPermanent(g, me.ID, "Goblin-town", "Land", hobOracle(t, "Goblin-town"), false)
			target := hobB4Creature(g, me.ID, "Test "+tc.name, tc.typeLine, 2, 2)
			err := hobB4Activate(t, g, me.ID, town, 0, "{C}{C}{B}{R}", cardRefs(target))
			if tc.ok != (err == nil) {
				t.Fatalf("activation: err = %v, want ok = %v", err, tc.ok)
			}
			if !tc.ok {
				if !battlefieldHas(g, town) {
					t.Error("a refused activation sacrificed the land")
				}
				return
			}
			passPriorityAroundTable(t, g)
			if battlefieldHas(g, town) {
				t.Error("Goblin-town was not sacrificed")
			}
			if got := hobB7Counters(t, g, target); got != 2 {
				t.Errorf("+1/+1 counters = %d, want 2", got)
			}
		})
	}
}

func TestHobGoblinTownNeedsAllOfItsMana(t *testing.T) {
	g, me, _ := spendTable(t)
	town := pushCatalogPermanent(g, me.ID, "Goblin-town", "Land", hobOracle(t, "Goblin-town"), false)
	goblin := hobB4Creature(g, me.ID, "Test Goblin", "Creature — Goblin Warrior", 2, 2)
	if err := hobB4Activate(t, g, me.ID, town, 0, "{C}{C}{B}", cardRefs(goblin)); err == nil {
		t.Fatal("activated without {R}")
	}
}

// --- Eagle's Rescue ----------------------------------------------------

func TestHobEaglesRescueGivesPlusTwoAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedBear(g, me.ID)
	castCatalogSpell(t, g, "Eagle's Rescue", "Enchantment — Aura", hobOracle(t, "Eagle's Rescue"), cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("enchanted bear = %d/%d, want 4/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature does not fly")
	}
}

// hobB7Rescue puts Eagle's Rescue in the active seat's graveyard and tries
// to bring it back at `target` with {2}{W/U}{W/U}.
func hobB7Rescue(t *testing.T, g *game.Game, target uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[0]
	id := pushCatalogGraveyardCard(me, "Eagle's Rescue", "Enchantment — Aura", hobOracle(t, "Eagle's Rescue"), 0, 0)
	return id, hobB4Activate(t, g, me.ID, id, 0, "{C}{C}{W}{U}", cardRefs(target))
}

func TestHobEaglesRescueReturnsFromTheGraveyardAttachedToASmallCreature(t *testing.T) {
	g, me, _ := spendTable(t)
	small := hobB4Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	id, err := hobB7Rescue(t, g, small)
	if err != nil {
		t.Fatalf("activation from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, id) || me.Graveyard.Contains(id) {
		t.Fatal("Eagle's Rescue did not return to the battlefield")
	}
	if host := attachmentHostOf(t, g, id); host.ID != small {
		t.Errorf("attached to %v, want the Mouse", host.ID)
	}
	if p, tough := effectivePTOf(t, g, small); p != 3 || tough != 3 {
		t.Errorf("the Mouse = %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, small, "flying") {
		t.Error("the Mouse does not fly")
	}
}

func TestHobEaglesRescueTargetsPowerOneOrLessOnly(t *testing.T) {
	g, me, _ := spendTable(t)
	zero := hobB4Creature(g, me.ID, "Egg", "Creature — Egg", 0, 3)
	big := hobB4Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if id, err := hobB7Rescue(t, g, big); err == nil {
		t.Error("returned to a power-2 creature")
	} else if !me.Graveyard.Contains(id) {
		t.Error("a refused activation moved the card")
	}
	if _, err := hobB7Rescue(t, g, zero); err != nil {
		t.Errorf("a power-0 creature is a legal target: %v", err)
	}
}

func TestHobEaglesRescueCannotTargetAnOpponentsCreature(t *testing.T) {
	g, _, opp := spendTable(t)
	theirs := hobB4Creature(g, opp.ID, "Their Mouse", "Creature — Mouse", 1, 1)
	if _, err := hobB7Rescue(t, g, theirs); err == nil {
		t.Error("returned attached to an opponent's creature")
	}
}

func TestHobEaglesRescueReturnIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	small := hobB4Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	id, err := hobB7Rescue(t, g, small)
	if err == nil {
		t.Fatal("returned outside a main phase")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("a refused activation moved the card")
	}
}

func TestHobEaglesRescueCannotBeActivatedFromTheBattlefield(t *testing.T) {
	g, me, _ := spendTable(t)
	small := hobB4Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	aura := pushCatalogPermanent(g, me.ID, "Eagle's Rescue", "Enchantment — Aura", hobOracle(t, "Eagle's Rescue"), false)
	if err := hobB4Activate(t, g, me.ID, aura, 0, "{C}{C}{W}{U}", cardRefs(small)); err == nil {
		t.Error("the graveyard ability was activated from the battlefield")
	}
}

// --- Vow to Erebor -----------------------------------------------------

// hobB7Vow casts Vow to Erebor at `target` and resolves it.
func hobB7Vow(t *testing.T, g *game.Game, target uuid.UUID) error {
	t.Helper()
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	_, err := hobB3Cast(t, g, me, "Vow to Erebor", "Instant", "{1}{W}", "{C}{W}", game.CastSpellParams{Targets: cardRefs(target)})
	if err != nil {
		return err
	}
	passPriorityAroundTable(t, g)
	return nil
}

func TestHobVowToEreborUntapsAndPumpsAndAsksNothingOfANonDwarf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true})
	seedEquipment(g, me.ID, "Test Sword", "")
	if err := hobB7Vow(t, g, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if hobB5Tapped(t, g, bear) {
		t.Error("the Bear is still tapped")
	}
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear = %d/%d, want 4/4", p, tough)
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a non-Dwarf was offered an Equipment")
	}
}

func TestHobVowToEreborAttachesAnEquipmentToADwarf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dwarf := pushPermanent(g, me.ID, game.Card{Name: "Dwarf", TypeLine: hobB7Dwarf, Power: 2, Toughness: 2, Tapped: true})
	sword := seedEquipment(g, me.ID, "Test Sword", "")
	if err := hobB7Vow(t, g, dwarf); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if hobB5Tapped(t, g, dwarf) {
		t.Error("the Dwarf is still tapped")
	}
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("the Dwarf's controller was not offered an Equipment")
	}
	answerChooseCards(t, g, me.ID, sword)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, sword); host.ID != dwarf {
		t.Errorf("the Equipment is attached to %v, want the Dwarf", host.ID)
	}
}

func TestHobVowToEreborMayDeclineTheAttach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dwarf := pushPermanent(g, me.ID, game.Card{Name: "Dwarf", TypeLine: hobB7Dwarf, Power: 2, Toughness: 2})
	sword := seedEquipment(g, me.ID, "Test Sword", "")
	if err := hobB7Vow(t, g, dwarf); err != nil {
		t.Fatalf("cast: %v", err)
	}
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, sword); host.ID != uuid.Nil {
		t.Errorf("the declined Equipment is attached to %v", host.ID)
	}
	if p, _ := effectivePTOf(t, g, dwarf); p != 4 {
		t.Errorf("declining the attach lost the pump: power %d", p)
	}
}

func TestHobVowToEreborMovesAnEquipmentFromAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dwarf := pushPermanent(g, me.ID, game.Card{Name: "Dwarf", TypeLine: hobB7Dwarf, Power: 2, Toughness: 2})
	bear := seedBear(g, me.ID)
	sword := seedEquipment(g, me.ID, "Test Sword", "")
	if err := g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatal(err)
	}
	if err := hobB7Vow(t, g, dwarf); err != nil {
		t.Fatalf("cast: %v", err)
	}
	answerChooseCards(t, g, me.ID, sword)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, sword); host.ID != dwarf {
		t.Errorf("the Equipment is on %v, want it moved to the Dwarf", host.ID)
	}
}

func TestHobVowToEreborOffersOnlyYourEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwarf := pushPermanent(g, me.ID, game.Card{Name: "Dwarf", TypeLine: hobB7Dwarf, Power: 2, Toughness: 2})
	mine := seedEquipment(g, me.ID, "My Sword", "")
	theirs := seedEquipment(g, opp.ID, "Their Sword", "")
	if err := hobB7Vow(t, g, dwarf); err != nil {
		t.Fatalf("cast: %v", err)
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no Equipment offered")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[mine] || offered[theirs] {
		t.Errorf("offered %v, want only my Equipment", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID)
}

func TestHobVowToEreborDwarfWithoutEquipmentAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dwarf := pushPermanent(g, me.ID, game.Card{Name: "Dwarf", TypeLine: hobB7Dwarf, Power: 2, Toughness: 2})
	if err := hobB7Vow(t, g, dwarf); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked about an Equipment with none to attach")
	}
}

func TestHobVowToEreborCannotTargetAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	theirs := pushPermanent(g, g.Seats[1].ID, game.Card{Name: "Dwarf", TypeLine: hobB7Dwarf, Power: 2, Toughness: 2})
	if err := hobB7Vow(t, g, theirs); err == nil {
		t.Error("Vow to Erebor targeted an opponent's creature")
	}
}

// --- Dragon's Desire ---------------------------------------------------

func TestHobDragonsDesireAddsRedForEachOpponentArtifact(t *testing.T) {
	g, me, opp := spendTable(t)
	hobB7Seed(g, opp.ID, "Relic", hobB7Artifact, 2)
	hobB7Seed(g, g.Seats[2].ID, "Relic", hobB7Artifact, 1)
	hobB7Seed(g, me.ID, "My Relic", hobB7Artifact, 3)
	if _, err := hobB3Cast(t, g, me, "Dragon's Desire", "Sorcery", "{2}{R}{R}", "{C}{C}{R}{R}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	units := me.ManaPool.Units()
	if len(units) != 3 {
		t.Fatalf("pool = %+v, want {R}{R}{R}", me.ManaPool)
	}
	for _, u := range units {
		if u.Color != "R" {
			t.Errorf("pool = %+v, want only red", me.ManaPool)
		}
	}
}

func TestHobDragonsDesireAddsNothingWithoutOpponentArtifacts(t *testing.T) {
	g, me, _ := spendTable(t)
	hobB7Seed(g, me.ID, "My Relic", hobB7Artifact, 2)
	if _, err := hobB3Cast(t, g, me, "Dragon's Desire", "Sorcery", "{2}{R}{R}", "{C}{C}{R}{R}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if units := me.ManaPool.Units(); len(units) != 0 {
		t.Errorf("pool = %+v, want empty", me.ManaPool)
	}
}

func TestHobDragonsDesireCountsOpponentArtifactCreaturesAndTokens(t *testing.T) {
	g, me, opp := spendTable(t)
	hobB7Seed(g, opp.ID, "Golem", "Artifact Creature — Golem", 1)
	hobB7Seed(g, opp.ID, "Treasure", "Token Artifact — Treasure", 1)
	hobB7Seed(g, opp.ID, "Bear", "Creature — Bear", 2)
	if _, err := hobB3Cast(t, g, me, "Dragon's Desire", "Sorcery", "{2}{R}{R}", "{C}{C}{R}{R}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if units := me.ManaPool.Units(); len(units) != 2 {
		t.Errorf("pool = %+v, want two {R}", me.ManaPool)
	}
}

// --- Colossal Whale ----------------------------------------------------

func TestHobColossalWhaleHasIslandwalk(t *testing.T) {
	g := newCatalogGame(t)
	whale := hobPush(t, g, g.Seats[0].ID, "Colossal Whale", "Creature — Whale", 5, 5)
	if !hasEffectiveKeyword(t, g, whale, "islandwalk") {
		t.Error("Colossal Whale lacks islandwalk")
	}
}

// hobB7WhaleAttack attacks opp with the Whale and chooses `victim` for the
// trigger; the "you may" is left for the caller.
func hobB7WhaleAttack(t *testing.T, g *game.Game, whale, victim uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(whale, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	pickCard(t, g, me.ID, victim)
}

func TestHobColossalWhaleExilesADefendingCreatureUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	whale := hobPush(t, g, me.ID, "Colossal Whale", "Creature — Whale", 5, 5)
	victim := hobB4Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hobB7WhaleAttack(t, g, whale, victim)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, victim) {
		t.Fatal("the Bear was not exiled")
	}
	if !g.Exile.Contains(victim) {
		t.Fatal("the Bear is not in exile")
	}
	hobDestroy(t, g, whale)
	g.RunStateChecksForTest()
	// It returns as a new object (CR 400.7) under its owner's control.
	if got := battlefieldCountNamed(g, opp.ID, "Their Bear"); got != 1 {
		t.Errorf("the Bear did not return under its owner's control when the Whale left (%d)", got)
	}
	if got := battlefieldCountNamed(g, me.ID, "Their Bear"); got != 0 {
		t.Errorf("the Bear returned under the Whale's controller (%d)", got)
	}
}

func TestHobColossalWhaleMayDeclineToExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	whale := hobPush(t, g, me.ID, "Colossal Whale", "Creature — Whale", 5, 5)
	victim := hobB4Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hobB7WhaleAttack(t, g, whale, victim)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, victim) {
		t.Error("the Bear was exiled after the controller said no")
	}
}

func TestHobColossalWhaleOnlyTargetsTheDefendingPlayersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	whale := hobPush(t, g, me.ID, "Colossal Whale", "Creature — Whale", 5, 5)
	hobB4Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bystander := hobB4Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(whale, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bystander}); err == nil {
		t.Error("the Whale targeted a creature the defending player does not control")
	}
}

func TestHobColossalWhaleExilesNothingIfItLeavesFirst(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	whale := hobPush(t, g, me.ID, "Colossal Whale", "Creature — Whale", 5, 5)
	victim := hobB4Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hobB7WhaleAttack(t, g, whale, victim)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(whale); err != nil {
			t.Fatal(err)
		}
	})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, victim) {
		t.Error("the Bear was exiled by a Whale that had already left (CR 610.3c)")
	}
}

// --- Stone-Giant of High Pass -----------------------------------------

func hobB7Boulders(g *game.Game, who uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == who && c.Name == "Stone Boulder" {
			out = append(out, c)
		}
	}
	return out
}

func TestHobStoneGiantMakesAStoneBoulderOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobCast(t, g, "Stone-Giant of High Pass", "Creature — Giant", 7, 7)
	boulders := hobB7Boulders(g, me.ID)
	if len(boulders) != 1 {
		t.Fatalf("%d Stone Boulders, want 1", len(boulders))
	}
	b := boulders[0]
	if !b.IsArtifact() || !b.IsCreature() || !b.HasSubtype("Wall") {
		t.Errorf("the token is %q, want an artifact creature Wall", b.TypeLine)
	}
	if p, tough := effectivePTOf(t, g, b.InstanceID); p != 3 || tough != 1 {
		t.Errorf("Stone Boulder = %d/%d, want 3/1", p, tough)
	}
	if !hasEffectiveKeyword(t, g, b.InstanceID, "defender") {
		t.Error("Stone Boulder lacks defender")
	}
	if len(b.Colors) != 0 {
		t.Errorf("Stone Boulder colours = %v, want colorless", b.Colors)
	}
}

func TestHobStoneGiantMakesAnotherStoneBoulderWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := hobPush(t, g, me.ID, "Stone-Giant of High Pass", "Creature — Giant", 7, 7)
	declareAttack(t, g, opp.ID, giant)
	passPriorityAroundTable(t, g)
	if got := len(hobB7Boulders(g, me.ID)); got != 1 {
		t.Fatalf("%d Stone Boulders after attacking, want 1", got)
	}
}

func TestHobStoneGiantSacrificesAnArtifactForFourDamage(t *testing.T) {
	g, me, opp := spendTable(t)
	giant := hobPush(t, g, me.ID, "Stone-Giant of High Pass", "Creature — Giant", 7, 7)
	relic := seedPermanent(g, me.ID, "Relic", hobB7Artifact)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{R}"); err != nil {
		t.Fatal(err)
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{
		Targets: hobB7Player(opp.ID), SacrificeIDs: []uuid.UUID{relic}, Strict: true,
	}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, relic) {
		t.Error("the artifact was not sacrificed")
	}
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want 4 damage", life, opp.Life)
	}
}

func TestHobStoneGiantDamageKillsACreature(t *testing.T) {
	g, me, opp := spendTable(t)
	giant := hobPush(t, g, me.ID, "Stone-Giant of High Pass", "Creature — Giant", 7, 7)
	relic := seedPermanent(g, me.ID, "Relic", hobB7Artifact)
	victim := hobB4Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 4)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{
		Targets: cardRefs(victim), SacrificeIDs: []uuid.UUID{relic}, Strict: true,
	}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, victim) {
		t.Error("4 damage did not kill the 2/4")
	}
}

func TestHobStoneGiantNeedsAnArtifactToSacrifice(t *testing.T) {
	g, me, opp := spendTable(t)
	giant := hobPush(t, g, me.ID, "Stone-Giant of High Pass", "Creature — Giant", 7, 7)
	bear := seedBear(g, me.ID)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{
		Targets: hobB7Player(opp.ID), SacrificeIDs: []uuid.UUID{bear}, Strict: true,
	}); err == nil {
		t.Error("sacrificed a nonartifact creature")
	}
	if !battlefieldHas(g, bear) {
		t.Error("a refused activation sacrificed the Bear")
	}
}

// --- Cavern-Hoard Dragon ----------------------------------------------

func TestHobCavernHoardDragonKeywords(t *testing.T) {
	g := newCatalogGame(t)
	dragon := hobPush(t, g, g.Seats[0].ID, "Cavern-Hoard Dragon", "Creature — Dragon", 6, 6)
	for _, kw := range []string{"flying", "trample", "haste"} {
		if !hasEffectiveKeyword(t, g, dragon, kw) {
			t.Errorf("Cavern-Hoard Dragon lacks %s", kw)
		}
	}
}

func TestHobCavernHoardDragonCostsLessByTheGreatestOpponentArtifactCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	price := func() int {
		return hobB3Price(t, g, me, "Cavern-Hoard Dragon", "Creature — Dragon", "{7}{R}{R}")
	}
	if got := price(); got != 9 {
		t.Fatalf("with no opponent artifacts: %d, want 9", got)
	}
	hobB7Seed(g, me.ID, "My Relic", hobB7Artifact, 5)
	if got := price(); got != 9 {
		t.Errorf("your own artifacts reduced the cost: %d, want 9", got)
	}
	hobB7Seed(g, opp.ID, "Relic", hobB7Artifact, 2)
	hobB7Seed(g, g.Seats[2].ID, "Relic", hobB7Artifact, 3)
	if got := price(); got != 6 {
		t.Errorf("with opponents holding 2 and 3: %d, want 6 (the greatest, not the sum)", got)
	}
}

func TestHobCavernHoardDragonNeverCostsLessThanItsColoredMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobB7Seed(g, opp.ID, "Relic", hobB7Artifact, 12)
	if got := hobB3Price(t, g, me, "Cavern-Hoard Dragon", "Creature — Dragon", "{7}{R}{R}"); got != 2 {
		t.Errorf("with 12 artifacts: %d, want {R}{R}", got)
	}
}

func TestHobCavernHoardDragonMakesATreasureForEachArtifactThePlayerControls(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := hobPush(t, g, me.ID, "Cavern-Hoard Dragon", "Creature — Dragon", 6, 6)
	hobB7Seed(g, opp.ID, "Relic", hobB7Artifact, 3)
	hobB7Seed(g, g.Seats[2].ID, "Other Relic", hobB7Artifact, 5)
	hobB7Seed(g, opp.ID, "Bear", "Creature — Bear", 2)
	declareAttack(t, g, opp.ID, dragon)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := hobB5TreasuresOf(g, me.ID); got != 3 {
		t.Errorf("%d Treasures, want 3 (the artifacts the damaged player controls)", got)
	}
}

func TestHobCavernHoardDragonMakesNoTreasureIfThePlayerHasNoArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := hobPush(t, g, me.ID, "Cavern-Hoard Dragon", "Creature — Dragon", 6, 6)
	hobB7Seed(g, g.Seats[2].ID, "Other Relic", hobB7Artifact, 3)
	life := opp.Life
	declareAttack(t, g, opp.ID, dragon)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-6 {
		t.Fatalf("opponent life %d -> %d, want the Dragon to connect", life, opp.Life)
	}
	if got := hobB5TreasuresOf(g, me.ID); got != 0 {
		t.Errorf("%d Treasures, want none", got)
	}
}

// --- Desolation of Smaug ------------------------------------------------

func TestHobDesolationOfSmaugDamagesEveryNonDragonCreature(t *testing.T) {
	g, me, opp := spendTable(t)
	small := hobB4Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
	big := hobB4Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 4)
	dragon := hobB4Creature(g, opp.ID, "Whelp", "Creature — Dragon", 5, 5)
	changeling := hobB4Creature(g, opp.ID, "Dragon Zombie", "Creature — Dragon Zombie", 3, 3)
	hobB7Desolate(t, g, me, "R", "R", "R", "R")
	if battlefieldHas(g, small) {
		t.Error("the 2/3 survived 3 damage")
	}
	if !battlefieldHas(g, big) || damageMarkedOn(g, big) != 3 {
		t.Errorf("the 2/4 should survive with 3 damage (marked %d)", damageMarkedOn(g, big))
	}
	for name, id := range map[string]uuid.UUID{"Dragon": dragon, "Dragon Zombie": changeling} {
		if !battlefieldHas(g, id) || damageMarkedOn(g, id) != 0 {
			t.Errorf("the %s was damaged (marked %d)", name, damageMarkedOn(g, id))
		}
	}
	for _, p := range g.Seats {
		if p.Life != 40 {
			t.Errorf("player life %d: the spell is creatures only", p.Life)
		}
	}
}

// hobB7ManaPicks answers the open colour picks of `who`, one colour each in
// the order given, and fails if any is left open.
func hobB7ManaPicks(t *testing.T, g *game.Game, who uuid.UUID, colors ...string) {
	t.Helper()
	for _, color := range colors {
		pick := riderLatestManaPick(g, who)
		if pick == nil {
			t.Fatalf("no colour pick open for %s", color)
		}
		if len(pick.ManaRestrictions) != 2 {
			t.Fatalf("the pick carries %v, want the cast and Dragon restrictions", pick.ManaRestrictions)
		}
		if err := g.ResolveManaChoice(pick.ID, who, color); err != nil {
			t.Fatalf("ResolveManaChoice(%s): %v", color, err)
		}
	}
	if pick := riderLatestManaPick(g, who); pick != nil {
		t.Fatal("more colour picks than the four mana printed")
	}
	passPriorityAroundTable(t, g)
}

// hobB7Desolate casts Desolation of Smaug and answers its four colour picks.
func hobB7Desolate(t *testing.T, g *game.Game, me *game.Player, colors ...string) {
	t.Helper()
	if _, err := hobB3Cast(t, g, me, "Desolation of Smaug", "Sorcery", "{2}{R}{R}", "{C}{C}{R}{R}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	hobB7ManaPicks(t, g, me.ID, colors...)
}

func TestHobDesolationOfSmaugAddsFourManaForDragonSpellsOnly(t *testing.T) {
	g, me, _ := spendTable(t)
	hobB7Desolate(t, g, me, "R", "R", "R", "R")
	if units := me.ManaPool.Units(); len(units) != 4 {
		t.Fatalf("pool = %+v, want four mana", me.ManaPool)
	}
	bear := handSpell(me, "Test Bear", "Creature — Bear", "{R}")
	refusedForMana(t, g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}), "the mana on a Bear")
	dragon := handSpell(me, "Test Dragon", "Creature — Dragon", "{R}{R}{R}{R}")
	if err := g.CastSpell(me.ID, dragon, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the mana could not pay for a Dragon spell: %v", err)
	}
}

func TestHobDesolationOfSmaugManaMayBeAnyColors(t *testing.T) {
	g, me, _ := spendTable(t)
	hobB7Desolate(t, g, me, "W", "U", "B", "G")
	dragon := handSpell(me, "Test Dragon", "Creature — Dragon", "{W}{U}{B}{G}")
	if err := g.CastSpell(me.ID, dragon, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("four different colours could not pay {W}{U}{B}{G}: %v", err)
	}
}

func TestHobDesolationOfSmaugManaCannotPayForAnAbility(t *testing.T) {
	g, me, _ := spendTable(t)
	hobB7Desolate(t, g, me, "R", "R", "R", "R")
	other := pushCatalogPermanent(g, me.ID, "Colossal Whale", "Creature — Whale", hobOracle(t, "Colossal Whale"), false)
	otherCard, _ := battlefieldCard(g, other)
	cost, _ := game.ParseCost("{R}")
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(otherCard)) {
		t.Error("the Dragon-only mana pays an activation cost")
	}
}

// --- Smaug the Magnificent ---------------------------------------------

func TestHobSmaugTheMagnificentKeywords(t *testing.T) {
	g := newCatalogGame(t)
	smaug := hobPush(t, g, g.Seats[0].ID, "Smaug the Magnificent", "Legendary Creature — Dragon", 4, 3)
	for _, kw := range []string{"flying", "haste"} {
		if !hasEffectiveKeyword(t, g, smaug, kw) {
			t.Errorf("Smaug the Magnificent lacks %s", kw)
		}
	}
}

func TestHobSmaugTheMagnificentHitsAnyTargetForItsTreasureCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	smaug := hobPush(t, g, me.ID, "Smaug the Magnificent", "Legendary Creature — Dragon", 4, 3)
	hobB7Seed(g, me.ID, "Treasure", "Token Artifact — Treasure", 2)
	hobB7Seed(g, me.ID, "Relic", hobB7Artifact, 3)
	hobB7Seed(g, opp.ID, "Treasure", "Token Artifact — Treasure", 4)
	declareAttack(t, g, opp.ID, smaug)
	answerPickTargetPlayer(t, g, third.ID)
	passPriorityAroundTable(t, g)
	if third.Life != 38 {
		t.Errorf("third player life %d, want 38 (two Treasures)", third.Life)
	}
}

func TestHobSmaugTheMagnificentCanHitACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	smaug := hobPush(t, g, me.ID, "Smaug the Magnificent", "Legendary Creature — Dragon", 4, 3)
	hobB7Seed(g, me.ID, "Treasure", "Token Artifact — Treasure", 3)
	victim := hobB4Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	declareAttack(t, g, opp.ID, smaug)
	pickTriggerTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, victim) {
		t.Error("three damage did not kill the 2/3")
	}
}

func TestHobSmaugTheMagnificentWithoutTreasuresDealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	smaug := hobPush(t, g, me.ID, "Smaug the Magnificent", "Legendary Creature — Dragon", 4, 3)
	declareAttack(t, g, opp.ID, smaug)
	answerPickTargetPlayer(t, g, third.ID)
	passPriorityAroundTable(t, g)
	if third.Life != 40 {
		t.Errorf("third player life %d, want 40", third.Life)
	}
}

func TestHobSmaugTheMagnificentMakesATreasureEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobPush(t, g, me.ID, "Smaug the Magnificent", "Legendary Creature — Dragon", 4, 3)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := hobB5TreasuresOf(g, me.ID); got != 1 {
		t.Errorf("%d Treasures after my upkeep, want 1", got)
	}
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := hobB5TreasuresOf(g, me.ID); got != 1 {
		t.Errorf("%d Treasures after an opponent's upkeep, want still 1", got)
	}
}

// --- Fíli and Kíli, Joyous ----------------------------------------------

func TestHobFiliAndKiliTapForRedRedAndHaveHaste(t *testing.T) {
	g, me, _ := spendTable(t)
	pair := pushCatalogPermanent(g, me.ID, "Fíli and Kíli, Joyous", "Legendary Creature — Dwarf Bard", hobOracle(t, "Fíli and Kíli, Joyous"), true)
	if !hasEffectiveKeyword(t, g, pair, "haste") {
		t.Error("Fíli and Kíli lack haste")
	}
	if err := g.ActivateManaAbility(me.ID, pair, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the turn it arrived: %v", err)
	}
	units := me.ManaPool.Units()
	if len(units) != 2 || units[0].Color != "R" || units[1].Color != "R" {
		t.Errorf("pool = %+v, want {R}{R}", me.ManaPool)
	}
}

func TestHobFiliAndKiliManaPaysOnlyForDwarfEquipmentAndSagaSpells(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		ok             bool
	}{
		{"dwarf", "Creature — Dwarf Warrior", true},
		{"dwarf artifact creature", "Artifact Creature — Dwarf", true},
		{"equipment", "Artifact — Equipment", true},
		{"saga", "Enchantment — Saga", true},
		{"bear", "Creature — Bear", false},
		{"plain artifact", "Artifact", false},
		{"instant", "Instant", false},
		{"enchantment", "Enchantment", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := spendTable(t)
			pair := pushCatalogPermanent(g, me.ID, "Fíli and Kíli, Joyous", "Legendary Creature — Dwarf Bard", hobOracle(t, "Fíli and Kíli, Joyous"), false)
			if err := g.ActivateManaAbility(me.ID, pair, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			spell := handSpell(me, "Test "+tc.name, tc.typeLine, "{R}{R}")
			err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true})
			if tc.ok && err != nil {
				t.Errorf("the pair's mana could not pay for a %s spell: %v", tc.name, err)
			}
			if !tc.ok {
				refusedForMana(t, err, "the pair's mana for a "+tc.name)
			}
		})
	}
}

func TestHobFiliAndKiliManaCannotPayForAnAbility(t *testing.T) {
	g, me, _ := spendTable(t)
	pair := pushCatalogPermanent(g, me.ID, "Fíli and Kíli, Joyous", "Legendary Creature — Dwarf Bard", hobOracle(t, "Fíli and Kíli, Joyous"), false)
	other := pushCatalogPermanent(g, me.ID, "Fíli and Kíli, Joyous", "Legendary Creature — Dwarf Bard", hobOracle(t, "Fíli and Kíli, Joyous"), false)
	if err := g.ActivateManaAbility(me.ID, pair, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	otherCard, _ := battlefieldCard(g, other)
	cost, _ := game.ParseCost("{R}")
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(otherCard)) {
		t.Error("the pair's mana pays an activation cost, stronger than printed")
	}
}

// --- Ori, Plate Stacker -------------------------------------------------

func TestHobOriDestroysOpponentsArtifactsAndEnchantmentsAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	relic := hobB7Seed(g, opp.ID, "Relic", hobB7Artifact, 2)
	ench := hobB7Seed(g, third.ID, "Ward", "Enchantment", 1)
	both := hobB7Seed(g, opp.ID, "Artifact Enchantment", "Enchantment Artifact", 1)
	bear := hobB7Seed(g, opp.ID, "Bear", "Creature — Bear", 1)
	mine := hobB7Seed(g, me.ID, "My Relic", hobB7Artifact, 1)
	life := me.Life
	hobCast(t, g, "Ori, Plate Stacker", "Legendary Creature — Dwarf Bard", 3, 3)
	for _, id := range append(append(relic, ench...), both...) {
		if battlefieldHas(g, id) {
			t.Errorf("%s survived", id)
		}
	}
	if !battlefieldHas(g, bear[0]) || !battlefieldHas(g, mine[0]) {
		t.Error("a creature or my own artifact was destroyed")
	}
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 (an artifact enchantment counts once)", life, me.Life)
	}
}

func TestHobOriDoesNotCountWhatSurvives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	survivor := pushPermanent(g, opp.ID, game.Card{Name: "Hard Relic", TypeLine: hobB7Artifact, Keywords: []string{"indestructible"}})
	hobB7Seed(g, opp.ID, "Relic", hobB7Artifact, 1)
	life := me.Life
	hobCast(t, g, "Ori, Plate Stacker", "Legendary Creature — Dwarf Bard", 3, 3)
	if !battlefieldHas(g, survivor) {
		t.Fatal("the indestructible artifact was destroyed")
	}
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1 (only the destroyed one)", life, me.Life)
	}
}

func TestHobOriWithNothingToDestroyGainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobB7Seed(g, me.ID, "My Relic", hobB7Artifact, 2)
	life := me.Life
	hobCast(t, g, "Ori, Plate Stacker", "Legendary Creature — Dwarf Bard", 3, 3)
	if me.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, me.Life)
	}
}

// --- Bard, Heir of Girion ------------------------------------------------

func TestHobBardIsAReachVigilantAnthemForOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bard := hobPush(t, g, me.ID, "Bard, Heir of Girion", "Legendary Creature — Human Archer", 4, 4)
	mine := seedBear(g, me.ID)
	theirs := seedBear(g, opp.ID)
	for _, kw := range []string{"reach", "vigilance"} {
		if !hasEffectiveKeyword(t, g, bard, kw) {
			t.Errorf("Bard lacks %s", kw)
		}
	}
	if p, tough := effectivePTOf(t, g, bard); p != 4 || tough != 4 {
		t.Errorf("Bard = %d/%d, want his own 4/4", p, tough)
	}
	if p, tough := effectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("my Bear = %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their Bear = %d/%d, want 2/2", p, tough)
	}
}

func TestHobBardDrawsOnceWhenYouAttackWithSeveralCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bard := hobPush(t, g, me.ID, "Bard, Heir of Girion", "Legendary Creature — Human Archer", 4, 4)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, bard, a, b)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1 (one attack, one trigger)", got)
	}
}

func TestHobBardDrawsWhenAnyCreatureAttacksNotJustBard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobPush(t, g, me.ID, "Bard, Heir of Girion", "Legendary Creature — Human Archer", 4, 4)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1", got)
	}
}

func TestHobBardDoesNotDrawWhenNobodyAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobPush(t, g, me.ID, "Bard, Heir of Girion", "Legendary Creature — Human Archer", 4, 4)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("hand grew by %d without an attack", got)
	}
}
