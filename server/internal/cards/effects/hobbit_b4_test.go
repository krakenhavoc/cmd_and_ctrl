package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB / HOC) cards, half B, batch 4. Every test drives
// the registered catalog row through the engine.

// hobB4Activate floats `mana`, then activates ability `idx` of `id` with
// strict payment (the cost must really be paid) at `targets`.
func hobB4Activate(t *testing.T, g *game.Game, who, id uuid.UUID, idx int, mana string, targets []game.TargetRef) error {
	t.Helper()
	if mana != "" {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	return g.ActivateCatalogAbility(who, id, idx, game.ActivateAbilityParams{Targets: targets, Strict: true})
}

// hobB4Creature puts a creature with the given type line onto the
// battlefield, untapped and ready to attack.
func hobB4Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// hobB4NoPickPrompt reports that no pick_target prompt is open.
func hobB4NoPickPrompt(g *game.Game) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return false
		}
	}
	return true
}

// --- Elvenking's Harper -----------------------------------------------

// hobB4HarperBlock makes `bear` unblockable with the Harper at `step`,
// then attacks with it and a creature that was not chosen, and reports
// whether the opposing creature may block each.
func hobB4HarperBlock(t *testing.T, step game.Step) (blocksChosen, blocksOther bool) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harper := hobPush(t, g, me.ID, "Elvenking's Harper", "Creature — Elf Bard", 2, 2)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	other := pushVanillaCreature(g, me.ID, "Other", 2, 2)
	blocker := pushVanillaCreature(g, opp.ID, "Blocker", 2, 2)
	advanceTo(t, g, step)
	if err := hobB4Activate(t, g, me.ID, harper, 0, "{C}{C}{C}{C}{C}", cardRefs(bear)); err == nil {
		t.Fatal("activated without paying {U}")
	}
	if err := hobB4Activate(t, g, me.ID, harper, 0, "{U}", cardRefs(bear)); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{bear, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	blocksChosen = g.DeclareBlocker(blocker, bear) == nil
	blocksOther = g.DeclareBlocker(blocker, other) == nil
	return blocksChosen, blocksOther
}

func TestHobElvenkingsHarperMakesACreatureUnblockable(t *testing.T) {
	chosen, other := hobB4HarperBlock(t, game.StepPrecombatMain)
	if chosen {
		t.Error("the creature the Harper chose was blocked")
	}
	if !other {
		t.Error("a creature the Harper did not choose cannot be blocked")
	}
}

func TestHobElvenkingsHarperActivatesAtInstantSpeed(t *testing.T) {
	chosen, other := hobB4HarperBlock(t, game.StepBeginCombat)
	if chosen {
		t.Error("the creature the Harper chose was blocked")
	}
	if !other {
		t.Error("a creature the Harper did not choose cannot be blocked")
	}
}

func TestHobElvenkingsHarperMayTargetAnyCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harper := hobPush(t, g, me.ID, "Elvenking's Harper", "Creature — Elf Bard", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := hobB4Activate(t, g, me.ID, harper, 0, "{C}{C}{C}{C}{U}", cardRefs(theirs)); err != nil {
		t.Fatalf("targeting an opponent's creature: %v", err)
	}
}

// --- Dúnedain Blade ---------------------------------------------------

func TestHobDunedainBladeEquipsAHumanForOneAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	human := hobB4Creature(g, me.ID, "Human Guard", "Creature — Human Soldier", 2, 2)
	bear := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Dúnedain Blade", hobOracle(t, "Dúnedain Blade"))

	if err := hobB4Activate(t, g, me.ID, blade, 0, "", cardRefs(bear)); err == nil {
		t.Fatal("Equip Human {1} accepted a non-Human creature")
	}
	if err := hobB4Activate(t, g, me.ID, blade, 0, "", cardRefs(human)); err == nil {
		t.Fatal("Equip Human {1} was activated without paying {1}")
	}
	if err := hobB4Activate(t, g, me.ID, blade, 0, "{C}", cardRefs(human)); err != nil {
		t.Fatalf("Equip Human {1}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, blade); host.ID != human {
		t.Fatalf("the Blade is on %v, want the Human", host)
	}
	if p, tg := effectivePower(t, g, human), effectiveToughness(t, g, human); p != 4 || tg != 3 {
		t.Errorf("the equipped Human is %d/%d, want 4/3", p, tg)
	}
}

func TestHobDunedainBladeEquipsAnyCreatureForThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	human := hobB4Creature(g, me.ID, "Human Guard", "Creature — Human Soldier", 2, 2)
	bear := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Dúnedain Blade", hobOracle(t, "Dúnedain Blade"))

	if err := hobB4Activate(t, g, me.ID, blade, 1, "{C}{C}", cardRefs(bear)); err == nil {
		t.Fatal("Equip {3} was activated for two mana")
	}
	if err := hobB4Activate(t, g, me.ID, blade, 1, "{C}", cardRefs(bear)); err != nil {
		t.Fatalf("Equip {3}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 4/3", p, tg)
	}

	// Re-equipping moves the Blade (CR 701.3a): the Human gets it for {1}.
	if err := hobB4Activate(t, g, me.ID, blade, 0, "{C}", cardRefs(human)); err != nil {
		t.Fatalf("Equip Human {1}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("the Bear kept the Blade's bonus, power %d", p)
	}
	if p := effectivePower(t, g, human); p != 4 {
		t.Errorf("the Human has power %d, want 4", p)
	}
}

func TestHobDunedainBladeEquipIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	human := hobB4Creature(g, me.ID, "Human Guard", "Creature — Human Soldier", 2, 2)
	blade := seedEquipment(g, me.ID, "Dúnedain Blade", hobOracle(t, "Dúnedain Blade"))
	advanceTo(t, g, game.StepBeginCombat)
	if err := hobB4Activate(t, g, me.ID, blade, 0, "{C}", cardRefs(human)); err == nil {
		t.Error("equipped outside a main phase")
	}
}

// --- Pelargir Survivor ------------------------------------------------

func TestHobPelargirSurvivorManaPaysOnlyForInstantsAndSorceries(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, color string
		ok                    bool
	}{
		{"instant", "Instant", "U", true},
		{"sorcery", "Sorcery", "R", true},
		{"creature", "Creature — Bear", "G", false},
		{"artifact", "Artifact", "W", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := spendTable(t)
			survivor := pushCatalogPermanent(g, me.ID, "Pelargir Survivor", "Creature — Human Peasant", hobOracle(t, "Pelargir Survivor"), false)
			if err := g.ActivateManaAbility(me.ID, survivor, 0, game.ManaAbilityParams{Colors: []string{tc.color}}); err != nil {
				t.Fatalf("tap for {%s}: %v", tc.color, err)
			}
			if units := me.ManaPool.Units(); len(units) != 1 || units[0].Color != tc.color {
				t.Fatalf("pool = %+v, want one %s", me.ManaPool, tc.color)
			}
			spell := handSpell(me, "Test "+tc.name, tc.typeLine, "{"+tc.color+"}")
			err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true})
			if tc.ok && err != nil {
				t.Errorf("the Survivor's mana could not pay for a %s: %v", tc.name, err)
			}
			if !tc.ok {
				refusedForMana(t, err, "the Survivor's mana for a "+tc.name)
			}
		})
	}
}

func TestHobPelargirSurvivorManaCannotPayForAnAbility(t *testing.T) {
	g, me, _ := spendTable(t)
	survivor := pushCatalogPermanent(g, me.ID, "Pelargir Survivor", "Creature — Human Peasant", hobOracle(t, "Pelargir Survivor"), false)
	other := pushCatalogPermanent(g, me.ID, "Pelargir Survivor", "Creature — Human Peasant", hobOracle(t, "Pelargir Survivor"), false)
	if err := g.ActivateManaAbility(me.ID, survivor, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatal(err)
	}
	otherCard, _ := battlefieldCard(g, other)
	cost, _ := game.ParseCost("{U}")
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(otherCard)) {
		t.Error("the Survivor's mana pays an activation cost, stronger than printed")
	}
}

func TestHobPelargirSurvivorIsSubjectToSummoningSickness(t *testing.T) {
	g, me, _ := spendTable(t)
	survivor := pushCatalogPermanent(g, me.ID, "Pelargir Survivor", "Creature — Human Peasant", hobOracle(t, "Pelargir Survivor"), true)
	if err := g.ActivateManaAbility(me.ID, survivor, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Error("a summoning-sick Survivor tapped for mana")
	}
}

func TestHobPelargirSurvivorMillsThreeAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	survivor := hobPush(t, g, me.ID, "Pelargir Survivor", "Creature — Human Peasant", 1, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	library, grave := opp.Library.Size(), opp.Graveyard.Size()
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := hobB4Activate(t, g, me.ID, survivor, 0, "{C}{C}{C}{C}{C}", target); err == nil {
		t.Fatal("activated without paying {U}")
	}
	if err := hobB4Activate(t, g, me.ID, survivor, 0, "{U}", target); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != library-3 {
		t.Errorf("library %d -> %d, want three cards milled", library, got)
	}
	if got := opp.Graveyard.Size(); got != grave+3 {
		t.Errorf("graveyard %d -> %d, want three more", grave, got)
	}
	if c, _ := battlefieldCard(g, survivor); !c.Tapped {
		t.Error("the Survivor did not tap for its own cost")
	}
	if err := hobB4Activate(t, g, me.ID, survivor, 0, "{C}{C}{C}{C}{C}{U}", target); err == nil {
		t.Error("activated again while tapped")
	}
}

// --- Minas Tirith Garrison --------------------------------------------

func TestHobMinasTirithGarrisonPowerIsTheCardsInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	garrison := hobPush(t, g, me.ID, "Minas Tirith Garrison", "Creature — Human Soldier", 0, 5)
	c := gaLayered(t, g, garrison)
	if want := me.Hand.Size(); c.CurrentPower() != want || c.CurrentToughness() != 5 {
		t.Fatalf("Garrison is %d/%d with %d cards in hand, want %d/5", c.CurrentPower(), c.CurrentToughness(), want, want)
	}
	g.WithWriteLock(func() {
		if err := (DrawCards{Player: me.ID, N: 2}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("DrawCards: %v", err)
		}
	})
	c = gaLayered(t, g, garrison)
	if want := me.Hand.Size(); c.CurrentPower() != want {
		t.Errorf("after drawing two the Garrison's power is %d, want %d", c.CurrentPower(), want)
	}
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(me.ID, me.Hand.Size()); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	if c = gaLayered(t, g, garrison); c.CurrentPower() != 0 {
		t.Errorf("with an empty hand the Garrison's power is %d, want 0", c.CurrentPower())
	}
}

// hobB4GarrisonAttack attacks with a Garrison beside two untapped Humans,
// a tapped Human and an untapped non-Human, and returns the table with
// the "tap any number" prompt open.
func hobB4GarrisonAttack(t *testing.T) (g *game.Game, me *game.Player, humanA, humanB, tappedHuman, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	garrison := hobPush(t, g, me.ID, "Minas Tirith Garrison", "Creature — Human Soldier", 0, 5)
	humanA = hobB4Creature(g, me.ID, "Human A", "Creature — Human Peasant", 1, 1)
	humanB = hobB4Creature(g, me.ID, "Human B", "Creature — Human Soldier", 1, 1)
	tappedHuman = pushTappedVanillaCreature(g, me.ID, "Tired Human", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tappedHuman {
				g.Battlefield.Cards[i].TypeLine = "Creature — Human Peasant"
			}
		}
	})
	bear = pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	declareAttack(t, g, g.Seats[1].ID, garrison)
	passPriorityAroundTable(t, g)
	return g, me, humanA, humanB, tappedHuman, bear
}

func TestHobMinasTirithGarrisonOffersOnlyUntappedHumans(t *testing.T) {
	g, me, humanA, humanB, _, _ := hobB4GarrisonAttack(t)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to tap Humans")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[humanA] || !offered[humanB] {
		t.Errorf("offered %d cards %v, want exactly the two untapped Humans", len(offered), pick.ChooseCards)
	}
}

func TestHobMinasTirithGarrisonDrawsACardPerHumanTapped(t *testing.T) {
	g, me, humanA, humanB, tappedHuman, bear := hobB4GarrisonAttack(t)
	hand := me.Hand.Size()
	answerChooseCards(t, g, me.ID, humanA, humanB)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{humanA, humanB} {
		if c, _ := battlefieldCard(g, id); !c.Tapped {
			t.Errorf("%s was not tapped", c.Name)
		}
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want two cards for two Humans", hand, got)
	}
	if c, _ := battlefieldCard(g, bear); c.Tapped {
		t.Error("the non-Human was tapped")
	}
	if c, _ := battlefieldCard(g, tappedHuman); !c.Tapped {
		t.Error("the already-tapped Human was untapped")
	}
}

func TestHobMinasTirithGarrisonTapsOnlyTheHumansYouChoose(t *testing.T) {
	g, me, humanA, humanB, _, _ := hobB4GarrisonAttack(t)
	hand := me.Hand.Size()
	answerChooseCards(t, g, me.ID, humanB)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, humanA); c.Tapped {
		t.Error("the Human that was not chosen was tapped")
	}
	if c, _ := battlefieldCard(g, humanB); !c.Tapped {
		t.Error("the chosen Human was not tapped")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card", hand, got)
	}
}

func TestHobMinasTirithGarrisonMayTapNone(t *testing.T) {
	g, me, humanA, humanB, _, _ := hobB4GarrisonAttack(t)
	hand := me.Hand.Size()
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{humanA, humanB} {
		if c, _ := battlefieldCard(g, id); c.Tapped {
			t.Errorf("%s was tapped though none was chosen", c.Name)
		}
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want no draw", hand, got)
	}
}

func TestHobMinasTirithGarrisonAsksNothingWithoutHumans(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	garrison := hobPush(t, g, me.ID, "Minas Tirith Garrison", "Creature — Human Soldier", 0, 5)
	pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hand := me.Hand.Size()
	declareAttack(t, g, g.Seats[1].ID, garrison)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked to tap Humans with none untapped")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want no draw", hand, got)
	}
}

// --- Andúril, Flame of the West ---------------------------------------

func hobB4Anduril(t *testing.T, legendary bool) (g *game.Game, me, opp *game.Player, creature, sword uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	if legendary {
		creature = seedLegendaryCreature(g, me.ID, "Commander")
	} else {
		creature = seedBear(g, me.ID)
	}
	sword = seedEquipment(g, me.ID, "Andúril, Flame of the West", hobOracle(t, "Andúril, Flame of the West"))
	equipTo(t, g, me.ID, sword, creature)
	return g, me, opp, creature, sword
}

func TestHobAndurilGivesPlusThreePlusOne(t *testing.T) {
	g, _, _, creature, _ := hobB4Anduril(t, false)
	if p, tg := effectivePower(t, g, creature), effectiveToughness(t, g, creature); p != 5 || tg != 3 {
		t.Errorf("the equipped creature is %d/%d, want 5/3", p, tg)
	}
}

// hobB4Spirits returns the Spirit tokens a seat controls.
func hobB4Spirits(g *game.Game, who uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == who && IsToken(c) && c.Name == "Spirit" {
			out = append(out, c)
		}
	}
	return out
}

func TestHobAndurilNonlegendaryAttackerMakesTwoTappedSpiritsNotAttacking(t *testing.T) {
	g, me, opp, creature, _ := hobB4Anduril(t, false)
	declareAttack(t, g, opp.ID, creature)
	passAndSettleAttackSplit(t, g, me.ID)
	spirits := hobB4Spirits(g, me.ID)
	if len(spirits) != 2 {
		t.Fatalf("%d Spirits, want 2", len(spirits))
	}
	for _, s := range spirits {
		if !s.Tapped {
			t.Error("a Spirit entered untapped")
		}
		if s.AttackingTarget != uuid.Nil {
			t.Error("a Spirit entered attacking though the creature is not legendary")
		}
		if s.CurrentPower() != 1 || s.CurrentToughness() != 1 {
			t.Errorf("a Spirit is %d/%d, want 1/1", s.CurrentPower(), s.CurrentToughness())
		}
		if !hasEffectiveKeyword(t, g, s.InstanceID, "flying") {
			t.Error("a Spirit lacks flying")
		}
		if !s.HasColor("W") {
			t.Error("a Spirit is not white")
		}
	}
}

func TestHobAndurilLegendaryAttackerMakesTwoTappedAttackingSpirits(t *testing.T) {
	g, me, opp, creature, _ := hobB4Anduril(t, true)
	declareAttack(t, g, opp.ID, creature)
	passAndSettleAttackSplit(t, g, me.ID)
	spirits := hobB4Spirits(g, me.ID)
	if len(spirits) != 2 {
		t.Fatalf("%d Spirits, want 2 (not 4: it is instead)", len(spirits))
	}
	for _, s := range spirits {
		if !s.Tapped || s.AttackingTarget != opp.ID {
			t.Errorf("a Spirit is tapped=%v attacking %s, want tapped and attacking %s", s.Tapped, s.AttackingTarget, opp.ID)
		}
	}
}

func TestHobAndurilReadsLegendaryFromLastKnownInformation(t *testing.T) {
	g, me, opp, creature, _ := hobB4Anduril(t, true)
	declareAttack(t, g, opp.ID, creature)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(creature); err != nil {
			t.Fatal(err)
		}
	})
	passAndSettleAttackSplit(t, g, me.ID)
	spirits := hobB4Spirits(g, me.ID)
	if len(spirits) != 2 {
		t.Fatalf("%d Spirits, want 2 from the departed creature's trigger", len(spirits))
	}
	for _, s := range spirits {
		if !s.Tapped || s.AttackingTarget != opp.ID {
			t.Errorf("a Spirit is tapped=%v attacking %s, want tapped and attacking: the creature was legendary", s.Tapped, s.AttackingTarget)
		}
	}
}

func TestHobAndurilIgnoresOtherAttackers(t *testing.T) {
	g, me, opp, _, _ := hobB4Anduril(t, false)
	other := pushVanillaCreature(g, me.ID, "Other", 2, 2)
	declareAttack(t, g, opp.ID, other)
	passAndSettleAttackSplit(t, g, me.ID)
	if n := len(hobB4Spirits(g, me.ID)); n != 0 {
		t.Errorf("%d Spirits for a creature that is not equipped, want none", n)
	}
}

// --- Dragon-Cursed Halls ----------------------------------------------

func TestHobDragonCursedHallsMakesATreasureWhenTheCreatureHitsAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	halls := pushCatalogPermanent(g, me.ID, "Dragon-Cursed Halls", "Land", hobOracle(t, "Dragon-Cursed Halls"), false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	other := pushVanillaCreature(g, me.ID, "Other", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := hobB4Activate(t, g, me.ID, halls, 0, "", cardRefs(bear)); err == nil {
		t.Fatal("activated without paying {1}")
	}
	if err := hobB4Activate(t, g, me.ID, halls, 0, "{C}", cardRefs(bear)); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, halls); !c.Tapped {
		t.Error("the Halls did not tap for their own cost")
	}
	life := opp.Life
	attackWith(t, g, opp.ID, bear, other)
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Fatalf("opponent life %d -> %d, want both creatures to connect", life, opp.Life)
	}
	if n := battlefieldCountNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want exactly one (only the chosen creature has the ability)", n)
	}
}

func TestHobDragonCursedHallsAddsColorless(t *testing.T) {
	g, me, _ := spendTable(t)
	halls := pushCatalogPermanent(g, me.ID, "Dragon-Cursed Halls", "Land", hobOracle(t, "Dragon-Cursed Halls"), false)
	if err := g.ActivateManaAbility(me.ID, halls, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if units := me.ManaPool.Units(); len(units) != 1 || units[0].Color != "C" {
		t.Errorf("pool = %+v, want one {C}", me.ManaPool)
	}
}

func TestHobDragonCursedHallsMakesNoTreasureWhenBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	halls := pushCatalogPermanent(g, me.ID, "Dragon-Cursed Halls", "Land", hobOracle(t, "Dragon-Cursed Halls"), false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	blocker := pushVanillaCreature(g, opp.ID, "Blocker", 3, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := hobB4Activate(t, g, me.ID, halls, 0, "{C}", cardRefs(bear)); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if n := battlefieldCountNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("%d Treasures for combat damage dealt to a creature, want none", n)
	}
}

// --- Bolg's Company ---------------------------------------------------

func TestHobBolgsCompanyHasHasteOnlyWhileYouControlAnotherGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolg := hobPush(t, g, me.ID, "Bolg's Company", "Creature — Goblin Soldier", 2, 2)
	if hasEffectiveKeyword(t, g, bolg, "haste") {
		t.Error("Bolg's Company has haste alone: it is a Goblin, but not another one")
	}
	theirs := hobB4Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	if hasEffectiveKeyword(t, g, bolg, "haste") {
		t.Error("an opponent's Goblin gave Bolg's Company haste")
	}
	nonGoblin := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if hasEffectiveKeyword(t, g, bolg, "haste") {
		t.Error("a non-Goblin gave Bolg's Company haste")
	}
	goblin := hobB4Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	if !hasEffectiveKeyword(t, g, bolg, "haste") {
		t.Error("Bolg's Company lacks haste with another Goblin")
	}
	if hasEffectiveKeyword(t, g, goblin, "haste") {
		t.Error("the other Goblin got Bolg's haste: it is a self-only grant")
	}
	hobDestroy(t, g, goblin)
	if hasEffectiveKeyword(t, g, bolg, "haste") {
		t.Error("Bolg's Company kept haste after the other Goblin died")
	}
	_, _ = theirs, nonGoblin
}

func TestHobBolgsCompanySacrificesAnotherGoblinForBlackAndRed(t *testing.T) {
	g, me, _ := spendTable(t)
	oracle := hobOracle(t, "Bolg's Company")
	// Summoning sick: the haste from the other Goblin is what lets it tap.
	bolg := pushCatalogPermanent(g, me.ID, "Bolg's Company", "Creature — Goblin Soldier", oracle, true)
	goblin := hobB4Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	for _, bad := range []uuid.UUID{bolg, bear} {
		if err := g.ActivateManaAbility(me.ID, bolg, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bad}}); err == nil {
			t.Fatalf("sacrificed %s to Bolg's Company", bad)
		}
	}
	if !g.Battlefield.Contains(bolg) || !g.Battlefield.Contains(bear) || me.ManaPool.Total() != 0 {
		t.Fatal("a refused activation changed the game")
	}
	if err := g.ActivateManaAbility(me.ID, bolg, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{goblin}}); err != nil {
		t.Fatalf("sacrificing another Goblin: %v", err)
	}
	if g.Battlefield.Contains(goblin) {
		t.Error("the Goblin is still on the battlefield")
	}
	units := me.ManaPool.Units()
	if len(units) != 2 || units[0].Color != "B" || units[1].Color != "R" {
		t.Errorf("pool = %+v, want {B}{R}", me.ManaPool)
	}
	if c, _ := battlefieldCard(g, bolg); !c.Tapped {
		t.Error("Bolg's Company did not tap")
	}
}

func TestHobBolgsCompanyCanSacrificeASecondBolgsCompany(t *testing.T) {
	g, me, _ := spendTable(t)
	oracle := hobOracle(t, "Bolg's Company")
	first := pushCatalogPermanent(g, me.ID, "Bolg's Company", "Creature — Goblin Soldier", oracle, false)
	second := pushCatalogPermanent(g, me.ID, "Bolg's Company", "Creature — Goblin Soldier", oracle, false)
	if err := g.ActivateManaAbility(me.ID, first, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{second}}); err != nil {
		t.Fatalf("sacrificing a second Bolg's Company: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("pool = %+v, want {B}{R}", me.ManaPool)
	}
}

// --- Haunt of the Dead Marshes ----------------------------------------

func TestHobHauntOfTheDeadMarshesScriesOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hobCast(t, g, "Haunt of the Dead Marshes", "Creature — Nightmare Elf", 1, 1)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry 1 prompt on entering")
	}
	answerScryKeepAll(t, g, me.ID)
}

// hobB4HauntReturn puts the Haunt in the graveyard of the active seat
// and tries to bring it back with {2}{B}.
func hobB4HauntReturn(t *testing.T, g *game.Game) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	id := pushCatalogGraveyardCard(me, "Haunt of the Dead Marshes", "Creature — Nightmare Elf", hobOracle(t, "Haunt of the Dead Marshes"), 1, 1)
	return id, hobB4Activate(t, g, me.ID, id, 0, "{C}{C}{B}", nil)
}

func TestHobHauntOfTheDeadMarshesReturnsTappedWithALegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLegendaryCreature(g, me.ID, "Commander")
	id, err := hobB4HauntReturn(t, g)
	if err != nil {
		t.Fatalf("activation with a legendary creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the Haunt did not return to the battlefield")
	}
	if !c.Tapped {
		t.Error("the Haunt returned untapped")
	}
	if me.Graveyard.Contains(id) {
		t.Error("the Haunt is still in the graveyard")
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("the returned Haunt did not scry on entering")
	}
}

func TestHobHauntOfTheDeadMarshesNeedsALegendaryCreatureYouControl(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(g *game.Game, me, opp *game.Player)
	}{
		{"nothing", func(*game.Game, *game.Player, *game.Player) {}},
		{"nonlegendary creature", func(g *game.Game, me, _ *game.Player) { seedBear(g, me.ID) }},
		{"legendary noncreature", func(g *game.Game, me, _ *game.Player) {
			seedPermanent(g, me.ID, "Legendary Relic", "Legendary Artifact")
		}},
		{"an opponent's legend", func(g *game.Game, _, opp *game.Player) { seedLegendaryCreature(g, opp.ID, "Their Commander") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			tc.setup(g, g.Seats[0], g.Seats[1])
			id, err := hobB4HauntReturn(t, g)
			if err == nil {
				t.Fatal("the Haunt returned without a legendary creature you control")
			}
			if !g.Seats[0].Graveyard.Contains(id) {
				t.Error("a refused activation moved the Haunt")
			}
		})
	}
}

func TestHobHauntOfTheDeadMarshesStillReturnsIfTheLegendLeavesInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	legend := seedLegendaryCreature(g, me.ID, "Commander")
	id, err := hobB4HauntReturn(t, g)
	if err != nil {
		t.Fatalf("activation: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(legend); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("the activation condition was re-checked on resolution: it is checked only when activating")
	}
}

// --- Azog, Moria's Ruin -----------------------------------------------

func hobB4CastAzog(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return hobCast(t, g, "Azog, Moria's Ruin", "Legendary Creature — Goblin Soldier", 1, 3)
}

func TestHobAzogDestroysAnOpponentsCreatureAndTheyAmassItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Ogre", 4, 4)
	azog := hobB4CastAzog(t, g)
	hand := me.Hand.Size()
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: azog}); err == nil {
		t.Fatal("Azog may target himself, but the trigger says \"other\"")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the target is still on the battlefield")
	}
	if got := hobArmyCounters(t, g, opp.ID); got != 4 {
		t.Errorf("the opponent's Army has %d counters, want 4 (the Ogre's power)", got)
	}
	if army := armyOf(t, g, opp.ID); !army.HasSubtype("Goblin") {
		t.Error("the opponent's Army is not a Goblin")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want no draw for an opponent's creature", hand, got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Army") {
			t.Error("Azog's controller amassed, but the target's controller does")
		}
	}
}

func TestHobAzogDestroyingYourOwnCreatureAmassesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := hobB4Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	hobB4CastAzog(t, g)
	hand := me.Hand.Size()
	answerPickTarget(t, g, mine)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the target is still on the battlefield")
	}
	if got := hobArmyCounters(t, g, me.ID); got != 3 {
		t.Errorf("Army has %d counters, want 3", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card for destroying your own creature", hand, got)
	}
}

func TestHobAzogAmassesEvenIfTheCreatureIsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Golem", TypeLine: "Artifact Creature — Golem",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"indestructible"},
	})
	hobB4CastAzog(t, g)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("an indestructible creature was destroyed")
	}
	if got := hobArmyCounters(t, g, opp.ID); got != 3 {
		t.Errorf("Army has %d counters, want 3: the amass does not depend on the destruction", got)
	}
}

func TestHobAzogMayChooseNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Ogre", 4, 4)
	hobB4CastAzog(t, g)
	b26DeclinePickTarget(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("a creature was destroyed though no target was chosen")
	}
	for _, c := range g.Battlefield.Cards {
		if c.HasSubtype("Army") {
			t.Error("an Army was made though no target was chosen")
		}
	}
}

func TestHobAzogAloneHasNothingToTarget(t *testing.T) {
	g := newCatalogGame(t)
	azog := hobB4CastAzog(t, g)
	passPriorityAroundTable(t, g)
	if !hobB4NoPickPrompt(g) {
		t.Error("asked for a target with no other creature on the battlefield")
	}
	if !g.Battlefield.Contains(azog) {
		t.Error("Azog destroyed himself")
	}
}

// --- Sauron, the Lidless Eye ------------------------------------------

func TestHobSauronStealsUntapsAndHastesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	victim := pushTappedVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	hobCast(t, g, "Sauron, the Lidless Eye", "Legendary Creature — Avatar Horror", 4, 4)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("the trigger accepted a creature its controller controls")
	}
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, victim); got != me.ID {
		t.Errorf("controller %s, want Sauron's controller", got)
	}
	if c, _ := battlefieldCard(g, victim); c.Tapped {
		t.Error("the stolen creature is still tapped")
	}
	assertKeywords(t, g, victim, "haste")
}

func TestHobSauronPumpsYourCreaturesAndDrainsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sauron := hobPush(t, g, me.ID, "Sauron, the Lidless Eye", "Legendary Creature — Avatar Horror", 4, 4)
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, g.Seats[1].ID, "Their Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	life := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		life[i] = p.Life
	}
	if err := hobB4Activate(t, g, me.ID, sauron, 0, "{C}{B}", nil); err == nil {
		t.Fatal("activated without paying {R}")
	}
	if err := hobB4Activate(t, g, me.ID, sauron, 0, "{R}", nil); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{sauron: 6, bear: 4, theirs: 2} {
		if c := gaLayered(t, g, id); c.CurrentPower() != want {
			t.Errorf("%s has power %d, want %d", c.Name, c.CurrentPower(), want)
		}
	}
	if c := gaLayered(t, g, bear); c.CurrentToughness() != 2 {
		t.Errorf("toughness %d, want it unchanged at 2", c.CurrentToughness())
	}
	for i, p := range g.Seats {
		want := life[i] - 2
		if i == 0 {
			want = life[i]
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
	late := pushVanillaCreature(g, me.ID, "Late Bear", 2, 2)
	if c := gaLayered(t, g, late); c.CurrentPower() != 2 {
		t.Errorf("a creature that arrived afterwards has power %d, want 2 (CR 611.2c)", c.CurrentPower())
	}
}

func TestHobSauronPumpsTheCreatureItStole(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	sauron := hobCast(t, g, "Sauron, the Lidless Eye", "Legendary Creature — Avatar Horror", 4, 4)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if err := hobB4Activate(t, g, me.ID, sauron, 0, "{C}{B}{R}", nil); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, victim); c.CurrentPower() != 4 {
		t.Errorf("the stolen creature has power %d, want 4", c.CurrentPower())
	}
}
