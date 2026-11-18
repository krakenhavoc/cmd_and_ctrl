package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b08_cards_test.go — S82 Adventures in the Forgotten Realms batch 8:
// the heavier Equipment and the Equipment legends. Every test drives the real
// equip, trigger, combat and layer paths.

const (
	afr8HolyAvengerOracle   = "86df1de7-967a-4420-847f-7a77d0217a15"
	afr8BeltOracle          = "eb70939f-c4ce-4224-ab74-acc24771d7f0"
	afr8FiendlashOracle     = "2d21dc80-2900-4dc3-9657-f17b50716481"
	afr8WandOracle          = "9fd521d5-baa3-4054-a990-d23265863c71"
	afr8BruenorOracle       = "ea7de96f-db69-421d-a951-ffaeac140379"
	afr8CattiBrieOracle     = "f44ffa35-bd88-4996-8b20-46329565df9b"
	afr8VorpalOracle        = "61bfab3f-d8f0-4ff5-9ddf-3fd605070c8f"
	afr8DancingSwordOracle  = "da3b12d1-8867-4063-b50d-3b2f5ce0549f"
	afr8IcingdeathOracle    = "1dc2c213-72f6-4112-96a8-bf6ad553c707"
	afr8GaleaOracle         = "30c0df75-9822-4016-a52b-d2e69dd58124"
	afr8LegendaryEquipLine  = "Legendary Artifact — Equipment"
	afr8IcingdeathTongue    = "Icingdeath, Frost Tongue"
	afr8BruenorTypeLine     = "Legendary Creature — Dwarf Warrior"
	afr8CattiBrieTypeLine   = "Legendary Creature — Human Archer"
	afr8IcingdeathTypeLine  = "Legendary Creature — Dragon"
	afr8GaleaTypeLine       = "Legendary Creature — Elf Knight"
	afr8VanillaCreatureLine = "Creature — Bear"
)

func TestAfr8CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		afr8HolyAvengerOracle:  "Holy Avenger",
		afr8BeltOracle:         "Belt of Giant Strength",
		afr8FiendlashOracle:    "Fiendlash",
		afr8WandOracle:         "Wand of Orcus",
		afr8BruenorOracle:      "Bruenor Battlehammer",
		afr8CattiBrieOracle:    "Catti-brie of Mithral Hall",
		afr8VorpalOracle:       "Vorpal Sword",
		afr8DancingSwordOracle: "Dancing Sword",
		afr8IcingdeathOracle:   "Icingdeath, Frost Tyrant",
		afr8GaleaOracle:        "Galea, Kindler of Hope",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
		if len(spec.Caveats) != 0 {
			t.Errorf("%s carries caveats %v", name, spec.Caveats)
		}
	}
}

// afr8Table is seat 0's precombat main with a 2/2 Bear.
func afr8Table(t *testing.T) (g *game.Game, me, opp *game.Player, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 0)
	bear = pushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: afr8VanillaCreatureLine, Power: 2, Toughness: 2})
	return g, me, opp, bear
}

func afr8Equipment(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return pushTestPermanent(g, owner, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle})
}

func afr8Has(t *testing.T, g *game.Game, id uuid.UUID, keyword string) bool {
	t.Helper()
	c := gaLayered(t, g, id)
	return game.HasKeyword(&c, keyword)
}

// --- Fiendlash ----------------------------------------------------------------

func TestAfr8FiendlashPumpsGivesReachAndReflectsDamageAsPower(t *testing.T) {
	g, _, opp, bear := afr8Table(t)
	me := g.Seats[0]
	fl := afr8Equipment(g, me.ID, "Fiendlash", "Artifact — Equipment", afr8FiendlashOracle)
	equipTo(t, g, me.ID, fl, bear)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Fatalf("power %d, want 4 (+2/+0)", got)
	}
	if !afr8Has(t, g, bear, "reach") {
		t.Fatal("the equipped creature has no reach")
	}
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, bear, 1); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	monSettle(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	monSettle(t, g)
	if opp.Life != life-4 {
		t.Fatalf("opponent life %d -> %d, want -4 (the creature's power)", life, opp.Life)
	}
	found := false
	for _, ev := range dealDamageEventsTo(g, opp.ID) {
		if ev.Source == bear && ev.Amount == 4 {
			found = true
		}
	}
	if !found {
		t.Error("the equipped creature is not the source of the damage")
	}
}

func TestAfr8FiendlashTriggersOnceForSimultaneousDamageAndIgnoresAnUnequippedCreature(t *testing.T) {
	g, me, opp, bear := afr8Table(t)
	fl := afr8Equipment(g, me.ID, "Fiendlash", "Artifact — Equipment", afr8FiendlashOracle)
	equipTo(t, g, me.ID, fl, bear)
	other := pushTestPermanent(g, me.ID, game.Card{Name: "Other", TypeLine: afr8VanillaCreatureLine, Power: 3, Toughness: 3})
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, other, 1)
	})
	g.RunStateChecksForTest()
	if got := sub3TriggersFrom(g, fl); got != 0 {
		t.Fatalf("%d triggers for an unequipped creature", got)
	}
	life := opp.Life
	srcA := pushTestPermanent(g, opp.ID, game.Card{Name: "A", TypeLine: afr8VanillaCreatureLine, Power: 1, Toughness: 1})
	srcB := pushTestPermanent(g, opp.ID, game.Card{Name: "B", TypeLine: afr8VanillaCreatureLine, Power: 1, Toughness: 1})
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToCreatureForEffect(srcA, bear, 1)
			return g.DealDamageToCreatureForEffect(srcB, bear, 1)
		})
	})
	g.RunStateChecksForTest()
	if got := sub3TriggersFrom(g, fl); got > 1 {
		t.Fatalf("%d triggers for one damage instance, want 1", got)
	}
	monSettle(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	monSettle(t, g)
	if opp.Life != life-4 {
		t.Fatalf("opponent life %d -> %d, want -4", life, opp.Life)
	}
}

// --- Bruenor Battlehammer -------------------------------------------------------

func TestAfr8BruenorPumpsEachCreatureForEachEquipmentAttachedToIt(t *testing.T) {
	g, me, _, bear := afr8Table(t)
	pushTestPermanent(g, me.ID, game.Card{Name: "Bruenor Battlehammer", TypeLine: afr8BruenorTypeLine, OracleID: afr8BruenorOracle, Power: 5, Toughness: 3})
	bare := pushTestPermanent(g, me.ID, game.Card{Name: "Bare", TypeLine: afr8VanillaCreatureLine, Power: 2, Toughness: 2})
	bs := afr8Equipment(g, me.ID, "Bonesplitter", "Artifact — Equipment", kiliBonesplitterOracle)
	equipTo(t, g, me.ID, bs, bear)
	if got := effectivePower(t, g, bear); got != 2+2+2 {
		t.Errorf("one Equipment: power %d, want 6 (Bonesplitter +2, Bruenor +2)", got)
	}
	bs2 := afr8Equipment(g, me.ID, "Bonesplitter", "Artifact — Equipment", kiliBonesplitterOracle)
	equipTo(t, g, me.ID, bs2, bear)
	if got := effectivePower(t, g, bear); got != 2+4+4 {
		t.Errorf("two Equipment: power %d, want 10", got)
	}
	if got := effectivePower(t, g, bare); got != 2 {
		t.Errorf("an unequipped creature has power %d, want 2", got)
	}
	// Moving an Equipment away takes its bonus with it.
	equipTo(t, g, me.ID, bs2, bare)
	if got := effectivePower(t, g, bear); got != 6 {
		t.Errorf("after moving one Equipment, power %d, want 6", got)
	}
	if got := effectivePower(t, g, bare); got != 6 {
		t.Errorf("the creature it moved to has power %d, want 6", got)
	}
}

func TestAfr8BruenorPaysZeroForTheFirstEquipEachTurnOnly(t *testing.T) {
	g, me, _, bear := afr8Table(t)
	pushTestPermanent(g, me.ID, game.Card{Name: "Bruenor Battlehammer", TypeLine: afr8BruenorTypeLine, OracleID: afr8BruenorOracle, Power: 5, Toughness: 3})
	bs := afr8Equipment(g, me.ID, "Bonesplitter", "Artifact — Equipment", kiliBonesplitterOracle)
	bs2 := afr8Equipment(g, me.ID, "Bonesplitter", "Artifact — Equipment", kiliBonesplitterOracle)
	row := equipRow(t, g, bs)
	tw, ok := twinOf(g, me.ID, bs, row)
	if !ok {
		t.Fatal("no {0} twin before any equip this turn")
	}
	if err := g.ActivateCatalogAbility(me.ID, bs, tw.Index, game.ActivateAbilityParams{
		Ref: tw.Ref, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("the {0} equip with an empty pool: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !attachedTo(t, g, bs, bear) {
		t.Fatal("Bonesplitter is not attached to the Bear")
	}
	if _, ok := twinOf(g, me.ID, bs2, row); ok {
		t.Error("a {0} twin is offered after the first equip of the turn")
	}
	if err := strictEquip(t, g, me.ID, bs2, bear, ""); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("the second equip for nothing: %v, want insufficient mana", err)
	}
	if err := strictEquip(t, g, me.ID, bs2, bear, "{C}"); err != nil {
		t.Fatalf("the second equip for {1}: %v", err)
	}
}

func TestAfr8BruenorOffersNoZeroEquipWithoutHim(t *testing.T) {
	g, me, _, bear := afr8Table(t)
	bs := afr8Equipment(g, me.ID, "Bonesplitter", "Artifact — Equipment", kiliBonesplitterOracle)
	if _, ok := twinOf(g, me.ID, bs, equipRow(t, g, bs)); ok {
		t.Fatal("a {0} twin with no Bruenor")
	}
	if err := strictEquip(t, g, me.ID, bs, bear, ""); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("equip for nothing: %v, want insufficient mana", err)
	}
}

// --- Icingdeath, Frost Tyrant ---------------------------------------------------

// afr8EquipTokenTo activates the equip row of an Equipment token, whose
// abilities live under its token key rather than an oracle id.
func afr8EquipTokenTo(t *testing.T, g *game.Game, controller, equipment, creature uuid.UUID) {
	t.Helper()
	card := gaLayered(t, g, equipment)
	rows := game.ActivatedAbilitiesForCard(card)
	for i, r := range rows {
		if r.Equip {
			if err := g.ActivateCatalogAbility(controller, equipment, i, game.ActivateAbilityParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}},
			}); err != nil {
				t.Fatalf("equip: %v", err)
			}
			passPriorityAroundTable(t, g)
			return
		}
	}
	t.Fatal("the token has no equip row")
}

func TestAfr8IcingdeathLeavesAnEquipmentTokenThatTapsADefender(t *testing.T) {
	g, me, opp, bear := afr8Table(t)
	dragon := pushTestPermanent(g, me.ID, game.Card{Name: "Icingdeath, Frost Tyrant", TypeLine: afr8IcingdeathTypeLine, OracleID: afr8IcingdeathOracle, Power: 4, Toughness: 3, Keywords: []string{"flying", "vigilance"}})
	sub3Destroy(t, g, dragon)
	passPriorityAroundTable(t, g)
	var tongue uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == afr8IcingdeathTongue {
			tongue = c.InstanceID
			if !c.IsToken() || !c.IsArtifact() || !c.HasSubtype("Equipment") || !c.IsLegendary() || !c.HasColor("W") || c.Controller != me.ID {
				t.Fatalf("the token is %+v", c)
			}
		}
	}
	if tongue == uuid.Nil {
		t.Fatal("no Icingdeath, Frost Tongue after the dragon died")
	}
	afr8EquipTokenTo(t, g, me.ID, tongue, bear)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Fatalf("power %d, want 4 (+2/+0)", got)
	}
	blocker := pushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4})
	declareAttack(t, g, opp.ID, bear)
	lockInAttacks(t, g)
	answerPickTarget(t, g, blocker)
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, blocker); !tapped {
		t.Fatal("the defender's creature was not tapped")
	}
}

func TestAfr8IcingdeathTongueEquipCostsTwo(t *testing.T) {
	g, me, _, bear := afr8Table(t)
	dragon := pushTestPermanent(g, me.ID, game.Card{Name: "Icingdeath, Frost Tyrant", TypeLine: afr8IcingdeathTypeLine, OracleID: afr8IcingdeathOracle, Power: 4, Toughness: 3})
	sub3Destroy(t, g, dragon)
	passPriorityAroundTable(t, g)
	var tongue uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == afr8IcingdeathTongue {
			tongue = c.InstanceID
		}
	}
	if tongue == uuid.Nil {
		t.Fatal("no token")
	}
	if got := equipPrice(t, g, me.ID, tongue, bear); got != "{2}" {
		t.Fatalf("equip: %s, want {2}", got)
	}
}

// --- Wand of Orcus --------------------------------------------------------------

func TestAfr8WandGivesDeathtouchToTheCreatureAndYourZombiesWhenItAttacks(t *testing.T) {
	g, me, opp, bear := afr8Table(t)
	zombie := pushTestPermanent(g, me.ID, game.Card{Name: "Zombie", TypeLine: "Creature — Zombie", Power: 2, Toughness: 2})
	oppZombie := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Zombie", TypeLine: "Creature — Zombie", Power: 2, Toughness: 2})
	other := pushTestPermanent(g, me.ID, game.Card{Name: "Other", TypeLine: afr8VanillaCreatureLine, Power: 2, Toughness: 2})
	wand := afr8Equipment(g, me.ID, "Wand of Orcus", afr8LegendaryEquipLine, afr8WandOracle)
	equipTo(t, g, me.ID, wand, bear)
	declareAttack(t, g, opp.ID, bear)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if !afr8Has(t, g, bear, "deathtouch") {
		t.Error("the attacker has no deathtouch")
	}
	if !afr8Has(t, g, zombie, "deathtouch") {
		t.Error("a Zombie you control has no deathtouch")
	}
	if afr8Has(t, g, oppZombie, "deathtouch") {
		t.Error("an opponent's Zombie has deathtouch")
	}
	if afr8Has(t, g, other, "deathtouch") {
		t.Error("a non-Zombie has deathtouch")
	}
}

func TestAfr8WandTriggersWhenTheEquippedCreatureBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 1)
	blocker := pushTestPermanent(g, opp.ID, game.Card{Name: "Blocker", TypeLine: afr8VanillaCreatureLine, Power: 1, Toughness: 4})
	wand := afr8Equipment(g, opp.ID, "Wand of Orcus", afr8LegendaryEquipLine, afr8WandOracle)
	equipTo(t, g, opp.ID, wand, blocker)
	aroundTheTableToMainOf(t, g, 0)
	atk := pushTestPermanent(g, me.ID, game.Card{Name: "Attacker", TypeLine: afr8VanillaCreatureLine, Power: 1, Toughness: 4})
	declareAttack(t, g, opp.ID, atk)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, atk); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if !afr8Has(t, g, blocker, "deathtouch") {
		t.Error("the equipped blocker has no deathtouch")
	}
}

func TestAfr8WandMakesAsManyZombiesAsTheCombatDamageDealt(t *testing.T) {
	g, me, opp, bear := afr8Table(t)
	wand := afr8Equipment(g, me.ID, "Wand of Orcus", afr8LegendaryEquipLine, afr8WandOracle)
	equipTo(t, g, me.ID, wand, bear)
	life := opp.Life
	declareAttack(t, g, opp.ID, bear)
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if lost := life - opp.Life; lost != 2 {
		t.Fatalf("the defender lost %d, want 2", lost)
	}
	zombies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Zombie") {
			zombies++
			if c.Power != 2 || c.Toughness != 2 || !c.HasColor("B") {
				t.Errorf("Zombie token is %+v, want a 2/2 black", c)
			}
		}
	}
	if zombies != 2 {
		t.Fatalf("%d Zombie tokens, want 2 (the combat damage dealt)", zombies)
	}
}

func TestAfr8WandMakesNoZombiesWhenTheCreatureIsBlocked(t *testing.T) {
	g, me, opp, bear := afr8Table(t)
	wand := afr8Equipment(g, me.ID, "Wand of Orcus", afr8LegendaryEquipLine, afr8WandOracle)
	equipTo(t, g, me.ID, wand, bear)
	wall := pushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4})
	onceStepBlock(t, g, opp.ID, bear, game.StepCombatDamage, wall)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Zombie") {
			t.Fatal("a Zombie was made though no combat damage was dealt to a player")
		}
	}
}

// --- Holy Avenger ---------------------------------------------------------------

func afr8HolyAvengerTable(t *testing.T) (g *game.Game, me, opp *game.Player, bear, ha uuid.UUID) {
	t.Helper()
	g, me, opp, bear = afr8Table(t)
	ha = afr8Equipment(g, me.ID, "Holy Avenger", "Artifact — Equipment", afr8HolyAvengerOracle)
	equipTo(t, g, me.ID, ha, bear)
	return g, me, opp, bear, ha
}

func TestAfr8HolyAvengerGivesDoubleStrikeAndPutsAnAuraFromHandOntoTheCreature(t *testing.T) {
	g, me, opp, bear, _ := afr8HolyAvengerTable(t)
	if !afr8Has(t, g, bear, "double strike") {
		t.Fatal("the equipped creature has no double strike")
	}
	aura := handCardForTest(me, "Test Aura", "Enchantment — Aura", "")
	other := handCardForTest(me, "Other Aura", "Enchantment — Aura", "")
	declareAttack(t, g, opp.ID, bear)
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepFirstStrikeDamage)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, aura)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, aura)
	if !ok {
		t.Fatal("the Aura did not enter the battlefield")
	}
	if !c.IsAttachedTo(bear) {
		t.Errorf("the Aura is attached to %+v, want the Bear", c.AttachedTo)
	}
	if !me.Hand.Contains(other) {
		t.Error("the other Aura left the hand")
	}
}

func TestAfr8HolyAvengerMayDeclineAndAsksNothingWithoutAnAura(t *testing.T) {
	g, me, opp, bear, _ := afr8HolyAvengerTable(t)
	creature := handCardForTest(me, "Not An Aura", "Creature — Bear", "")
	declareAttack(t, g, opp.ID, bear)
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepFirstStrikeDamage)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("asked for an Aura with none in hand")
	}
	if !me.Hand.Contains(creature) {
		t.Fatal("a non-Aura left the hand")
	}
}

func TestAfr8HolyAvengerDeclineKeepsTheAuraInHand(t *testing.T) {
	g, me, opp, bear, _ := afr8HolyAvengerTable(t)
	aura := handCardForTest(me, "Test Aura", "Enchantment — Aura", "")
	declareAttack(t, g, opp.ID, bear)
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepFirstStrikeDamage)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Fatal("the declined Aura left the hand")
	}
}

// --- Catti-brie of Mithral Hall ---------------------------------------------------

func TestAfr8CattiBriePutsACounterPerEquipmentWhenSheAttacks(t *testing.T) {
	g, me, opp, _ := afr8Table(t)
	catti := pushTestPermanent(g, me.ID, game.Card{Name: "Catti-brie of Mithral Hall", TypeLine: afr8CattiBrieTypeLine, OracleID: afr8CattiBrieOracle, Power: 2, Toughness: 2, Keywords: []string{"reach", "first strike"}})
	for i := 0; i < 2; i++ {
		bs := afr8Equipment(g, me.ID, "Bonesplitter", "Artifact — Equipment", kiliBonesplitterOracle)
		equipTo(t, g, me.ID, bs, catti)
	}
	declareAttack(t, g, opp.ID, catti)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := counterOnBattlefieldCard(g, catti, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d +1/+1 counters, want 2 (one per Equipment)", got)
	}
}

func TestAfr8CattiBrieWithNoEquipmentGetsNoCounter(t *testing.T) {
	g, me, opp, _ := afr8Table(t)
	catti := pushTestPermanent(g, me.ID, game.Card{Name: "Catti-brie of Mithral Hall", TypeLine: afr8CattiBrieTypeLine, OracleID: afr8CattiBrieOracle, Power: 2, Toughness: 2})
	declareAttack(t, g, opp.ID, catti)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := counterOnBattlefieldCard(g, catti, game.CounterPlusOne); got != 0 {
		t.Fatalf("%d counters with no Equipment, want 0", got)
	}
}

// afr8CattiBrieDefends is seat 1's Catti-brie holding `counters` +1/+1
// counters, facing seat 0's attacking 5/5.
func afr8CattiBrieDefends(t *testing.T, counters int) (g *game.Game, def *game.Player, catti, attacker uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 0)
	attacker = pushTestPermanent(g, atk.ID, game.Card{Name: "Brute", TypeLine: afr8VanillaCreatureLine, Power: 5, Toughness: 5})
	catti = pushTestPermanent(g, def.ID, game.Card{Name: "Catti-brie of Mithral Hall", TypeLine: afr8CattiBrieTypeLine, OracleID: afr8CattiBrieOracle, Power: 2, Toughness: 2, Counters: map[string]int{game.CounterPlusOne: counters}})
	declareAttack(t, g, def.ID, attacker)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	return g, def, catti, attacker
}

func TestAfr8CattiBrieRemovesEveryCounterAndDealsThatMuchDamage(t *testing.T) {
	g, def, catti, attacker := afr8CattiBrieDefends(t, 3)
	fillPool(def, 1)
	if err := g.ActivateCatalogAbility(def.ID, catti, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: attacker}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterOnBattlefieldCard(g, catti, game.CounterPlusOne); got != 0 {
		t.Fatalf("%d counters left, want all removed as a cost", got)
	}
	passPriorityAroundTable(t, g)
	found := false
	for _, ev := range dealDamageEventsTo(g, attacker) {
		if ev.Source == catti && ev.Amount == 3 {
			found = true
		}
	}
	if !found {
		t.Error("Catti-brie did not deal 3 damage to the attacker")
	}
}

func TestAfr8CattiBrieRefusesAPartialRemoval(t *testing.T) {
	g, def, catti, attacker := afr8CattiBrieDefends(t, 3)
	fillPool(def, 1)
	err := g.ActivateCatalogAbility(def.ID, catti, 0, game.ActivateAbilityParams{
		Strict:        true,
		CounterCounts: []int{2},
		Targets:       []game.TargetRef{{Kind: game.TargetCard, ID: attacker}},
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("removing 2 of 3: %v, want ErrInvalidParam", err)
	}
	if got := counterOnBattlefieldCard(g, catti, game.CounterPlusOne); got != 3 {
		t.Fatalf("%d counters, want the refused activation to cost nothing", got)
	}
}

func TestAfr8CattiBrieWithNoCountersIsPayableAndDealsNothing(t *testing.T) {
	g, def, catti, attacker := afr8CattiBrieDefends(t, 0)
	fillPool(def, 1)
	if err := g.ActivateCatalogAbility(def.ID, catti, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: attacker}},
	}); err != nil {
		t.Fatalf("activate with no counters: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, ev := range dealDamageEventsTo(g, attacker) {
		if ev.Source == catti && ev.Amount > 0 {
			t.Fatalf("Catti-brie dealt %d damage with no counters removed", ev.Amount)
		}
	}
}

func TestAfr8CattiBrieCantTargetYourOwnOrANonCombatant(t *testing.T) {
	g, def, catti, _ := afr8CattiBrieDefends(t, 2)
	mine := pushTestPermanent(g, def.ID, game.Card{Name: "Mine", TypeLine: afr8VanillaCreatureLine, Power: 1, Toughness: 1})
	fillPool(def, 1)
	err := g.ActivateCatalogAbility(def.ID, catti, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting your own creature: %v, want ErrIllegalTarget", err)
	}
}

// --- Belt of Giant Strength --------------------------------------------------------

func TestAfr8BeltSetsBasePowerAndToughnessToTenAndKeepsModifiers(t *testing.T) {
	g, me, _, bear := afr8Table(t)
	pushTestPermanent(g, me.ID, game.Card{Name: "Bruenor Battlehammer", TypeLine: afr8BruenorTypeLine, OracleID: afr8BruenorOracle, Power: 5, Toughness: 3})
	belt := afr8Equipment(g, me.ID, "Belt of Giant Strength", "Artifact — Equipment", afr8BeltOracle)
	equipTo(t, g, me.ID, belt, bear)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 12 || tough != 10 {
		t.Fatalf("equipped Bear is %d/%d, want 12/10 (base 10/10, Bruenor's +2/+0)", p, tough)
	}
}

func TestAfr8BeltEquipCostsTenLessThePowerOfItsTarget(t *testing.T) {
	g, me, _, bear := afr8Table(t)
	belt := afr8Equipment(g, me.ID, "Belt of Giant Strength", "Artifact — Equipment", afr8BeltOracle)
	big := pushTestPermanent(g, me.ID, game.Card{Name: "Brute", TypeLine: afr8VanillaCreatureLine, Power: 4, Toughness: 4})
	huge := pushTestPermanent(g, me.ID, game.Card{Name: "Huge", TypeLine: afr8VanillaCreatureLine, Power: 12, Toughness: 12})
	if got := equipPrice(t, g, me.ID, belt, bear); got != "{8}" {
		t.Errorf("equip onto a 2-power creature: %s, want {8}", got)
	}
	if got := equipPrice(t, g, me.ID, belt, big); got != "{6}" {
		t.Errorf("equip onto a 4-power creature: %s, want {6}", got)
	}
	if got := equipPrice(t, g, me.ID, belt, huge); got != "{0}" {
		t.Errorf("equip onto a 12-power creature: %s, want {0} (never below zero)", got)
	}
	// The engine charges what it shows.
	if err := strictEquip(t, g, me.ID, belt, bear, "{C}{C}{C}{C}{C}{C}{C}"); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("seven mana for an {8} equip: %v, want insufficient mana", err)
	}
	if err := strictEquip(t, g, me.ID, belt, huge, ""); err != nil {
		t.Fatalf("equip onto a 12-power creature for nothing: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !attachedTo(t, g, belt, huge) {
		t.Fatal("Belt is not attached to the huge creature")
	}
}

// --- Vorpal Sword --------------------------------------------------------------------

func afr8VorpalTable(t *testing.T) (g *game.Game, me, opp *game.Player, bear, sword uuid.UUID) {
	t.Helper()
	g, me, opp, bear = afr8Table(t)
	sword = afr8Equipment(g, me.ID, "Vorpal Sword", "Artifact — Equipment", afr8VorpalOracle)
	equipTo(t, g, me.ID, sword, bear)
	return g, me, opp, bear, sword
}

func TestAfr8VorpalSwordPumpsAndGivesDeathtouch(t *testing.T) {
	g, _, _, bear, _ := afr8VorpalTable(t)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	if !afr8Has(t, g, bear, "deathtouch") {
		t.Error("no deathtouch")
	}
}

func TestAfr8VorpalSwordEquipCostsTwoBlack(t *testing.T) {
	g, me, _, bear := afr8Table(t)
	sword := afr8Equipment(g, me.ID, "Vorpal Sword", "Artifact — Equipment", afr8VorpalOracle)
	p, err := g.PriceActivation(me.ID, sword, equipRow(t, g, sword), []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	if got := p.Total.String(); got != "{B}{B}" {
		t.Errorf("equip: %s, want {B}{B}", got)
	}
}

func TestAfr8VorpalSwordWithoutTheAbilityJustDealsDamage(t *testing.T) {
	g, _, opp, bear, _ := afr8VorpalTable(t)
	declareAttack(t, g, opp.ID, bear)
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Eliminated {
		t.Fatal("the defender lost without the ability")
	}
}

func TestAfr8VorpalSwordActivatedMakesTheCombatDamagedPlayerLose(t *testing.T) {
	g, me, opp, bear, sword := afr8VorpalTable(t)
	fillPool(me, 5)
	fillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, bear)
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if !opp.Eliminated {
		t.Fatalf("the defender (life %d) did not lose the game", opp.Life)
	}
}

func TestAfr8VorpalSwordGrantEndsAtEndOfTurn(t *testing.T) {
	g, me, opp, bear, sword := afr8VorpalTable(t)
	fillPool(me, 5)
	fillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCleanup)
	aroundTheTableToMain(t, g)
	advanceToMainOf(t, g, 0)
	declareAttack(t, g, opp.ID, bear)
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Eliminated {
		t.Fatal("the until-end-of-turn grant outlived its turn")
	}
}

// --- Dancing Sword ----------------------------------------------------------------------

func afr8DancingTable(t *testing.T) (g *game.Game, me, opp *game.Player, bear, sword uuid.UUID) {
	t.Helper()
	g, me, opp, bear = afr8Table(t)
	sword = afr8Equipment(g, me.ID, "Dancing Sword", "Artifact — Equipment", afr8DancingSwordOracle)
	equipTo(t, g, me.ID, sword, bear)
	return g, me, opp, bear, sword
}

func TestAfr8DancingSwordPumps(t *testing.T) {
	g, _, _, bear, _ := afr8DancingTable(t)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 3 {
		t.Fatalf("equipped Bear is %d/%d, want 4/3", p, tough)
	}
}

func TestAfr8DancingSwordBecomesAConstructWhenTheCreatureDiesAndYouSayYes(t *testing.T) {
	g, me, opp, bear, sword := afr8DancingTable(t)
	sub3Destroy(t, g, bear)
	ask := awaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, sword)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Construct") {
		t.Fatalf("the Sword is %v, want an artifact creature — Construct", c.TypeLine)
	}
	if c.HasSubtype("Equipment") {
		t.Error("it is still an Equipment")
	}
	if p, tough := effectivePower(t, g, sword), effectiveToughness(t, g, sword); p != 2 || tough != 1 {
		t.Errorf("it is %d/%d, want 2/1", p, tough)
	}
	if !afr8Has(t, g, sword, "flying") {
		t.Error("it has no flying")
	}
	if !afr8WardCountersABolt(t, g, opp, sword) {
		t.Error("a Lightning Bolt from an opponent who can't pay resolved: it has no ward {1}")
	}
	// The change has no end: it is still a creature at the next turn.
	advanceTo(t, g, game.StepCleanup)
	aroundTheTableToMain(t, g)
	if c := gaLayered(t, g, sword); !c.IsCreature() || c.HasSubtype("Equipment") {
		t.Error("the change ended with the turn")
	}
}

func TestAfr8DancingSwordStaysAnEquipmentWhenYouSayNo(t *testing.T) {
	g, me, _, bear, sword := afr8DancingTable(t)
	sub3Destroy(t, g, bear)
	ask := awaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := gaLayered(t, g, sword)
	if c.IsCreature() || !c.HasSubtype("Equipment") {
		t.Fatalf("the Sword is %v after declining", c.TypeLine)
	}
}

func TestAfr8DancingSwordNoLongerEquipsOnceItIsAConstruct(t *testing.T) {
	g, me, _, bear, sword := afr8DancingTable(t)
	sub3Destroy(t, g, bear)
	ask := awaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	other := pushTestPermanent(g, me.ID, game.Card{Name: "Other", TypeLine: afr8VanillaCreatureLine, Power: 2, Toughness: 2})
	_ = g.ActivateCatalogAbility(me.ID, sword, len(game.ActivatedAbilitiesForCard(game.Card{OracleID: afr8DancingSwordOracle}))-1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}},
	})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if attachedTo(t, g, sword, other) {
		t.Fatal("a Construct creature stayed attached to a creature")
	}
	if got := effectivePower(t, g, other); got != 2 {
		t.Errorf("the other creature has power %d, want 2 (no Equipment bonus)", got)
	}
}

// afr8LibraryTop puts a card on top of a player's library and returns its ID.
func afr8LibraryTop(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.PushTop(c)
	return c.InstanceID
}

// afr8WardCountersABolt has the opponent cast a Lightning Bolt at the
// permanent with an empty pool and reports whether ward {1} countered it.
func afr8WardCountersABolt(t *testing.T, g *game.Game, opp *game.Player, id uuid.UUID) bool {
	t.Helper()
	boltWardedCreature(t, g, opp, id)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		return false
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	return !boltResolved(g, id)
}

// --- Galea, Kindler of Hope ----------------------------------------------------------------

func afr8GaleaTable(t *testing.T) (g *game.Game, me *game.Player, bear uuid.UUID) {
	t.Helper()
	g, me, _, bear = afr8Table(t)
	pushTestPermanent(g, me.ID, game.Card{Name: "Galea, Kindler of Hope", TypeLine: afr8GaleaTypeLine, OracleID: afr8GaleaOracle, Power: 4, Toughness: 4})
	return g, me, bear
}

func TestAfr8GaleaCastsAnEquipmentFromTheTopAndItAttachesAsItEnters(t *testing.T) {
	g, me, bear := afr8GaleaTable(t)
	top := afr8LibraryTop(me, game.Card{Name: "Bonesplitter", TypeLine: "Artifact — Equipment", ManaCost: "{1}", OracleID: kiliBonesplitterOracle})
	fillPool(me, 1)
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: string(game.ZoneLibrary), Strict: true}); err != nil {
		t.Fatalf("cast from the top of the library: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if !attachedTo(t, g, top, bear) {
		t.Fatal("the Equipment is not attached to the creature")
	}
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
}

func TestAfr8GaleaCastsAnAuraFromTheTopButNotACreature(t *testing.T) {
	g, me, _ := afr8GaleaTable(t)
	creature := afr8LibraryTop(me, game.Card{Name: "Bear Cub", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 1, Toughness: 1})
	fillPool(me, 1)
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{FromZone: string(game.ZoneLibrary), Strict: true}); err == nil {
		t.Fatal("a creature spell was cast from the top of the library")
	}
	aura := afr8LibraryTop(me, game.Card{Name: "Test Aura", TypeLine: "Enchantment — Aura", ManaCost: "{1}"})
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{FromZone: string(game.ZoneLibrary), Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: g.Battlefield.Cards[0].InstanceID}}}); err != nil {
		t.Fatalf("an Aura from the top: %v", err)
	}
}

func TestAfr8GaleaGivesNothingToAnEquipmentCastFromHand(t *testing.T) {
	g, me, bear := afr8GaleaTable(t)
	_ = bear
	hand := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: hand, Name: "Bonesplitter", TypeLine: "Artifact — Equipment", ManaCost: "{1}", OracleID: kiliBonesplitterOracle, Owner: me.ID, Controller: me.ID})
	fillPool(me, 1)
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.PendingChoices {
		if p != nil && p.Kind == game.PendingChoicePickTarget {
			t.Fatal("an Equipment cast from hand asked to attach as it entered")
		}
	}
	if attachedTo(t, g, hand, bear) {
		t.Fatal("an Equipment cast from hand attached itself")
	}
}
