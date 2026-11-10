package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b18_test.go — Commander deck cards, batch 18 (#5368).

const (
	cmdrB18KeigaOracle     = "521b1598-58c6-47a0-a7b3-021d70bf2b05"
	cmdrB18BehindOracle    = "e84837bd-4f94-43e1-876e-d82cae6ee247"
	cmdrB18WitchOracle     = "8b1aa532-4f19-407e-83a5-13b3160ad8c2"
	cmdrB18JunjiOracle     = "78a48f4e-ba4c-4d8d-938a-9ddb342a885b"
	cmdrB18LordOracle      = "3a25d1e2-63d8-4a93-b0ce-f9307736b3f9"
	cmdrB18RenderOracle    = "028dcef2-3de9-4dde-ba17-f9e077b3190d"
	cmdrB18SelectionOracle = "b48a149c-aead-43b2-a99e-c964d6d28aa0"
	cmdrB18FiendOracle     = "bb714bc6-9f32-42f6-9064-f28a383675b5"
	cmdrB18TasigurOracle   = "837e3fde-241f-4826-ad04-f146fdcf2278"
)

func TestCmdrB18KeigaDiesAndStealsACreatureForGood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keiga := pushTestPermanent(g, me.ID, game.Card{Name: "Keiga, the Tide Star", OracleID: cmdrB18KeigaOracle,
		TypeLine: "Legendary Creature — Dragon Spirit", Power: 5, Toughness: 5})
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := g.DestroyPermanentForEffect(keiga); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		answerPickTarget(t, g, theirs)
		passPriorityAroundTable(t, g)
	}
	c, ok := battlefieldCard(g, theirs)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the creature was not stolen: %+v", c)
	}
	// Indefinitely: still mine after the turn ends.
	advanceToDeclareAttackersOf(t, g, 1)
	if c, _ := battlefieldCard(g, theirs); c.Controller != me.ID {
		t.Error("the control change wore off")
	}
}

func TestCmdrB18BehindTheScenesGivesSkulkAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scenes := pushTestPermanent(g, me.ID, game.Card{Name: "Behind the Scenes", OracleID: cmdrB18BehindOracle, TypeLine: "Enchantment"})
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	if !hasEffectiveKeyword(t, g, mine, "skulk") {
		t.Error("my creatures have skulk")
	}
	if hasEffectiveKeyword(t, g, theirs, "skulk") {
		t.Error("their creatures don't")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"},
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, scenes, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, mine); p != 3 {
		t.Errorf("my creature power %d, want 3", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("their creature power %d, want 2", p)
	}
}

func TestCmdrB18ChitteringWitchMakesARatPerOpponentAndShrinks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Chittering Witch", "Creature — Human Warlock", cmdrB18WitchOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Rat"); got != 3 {
		t.Fatalf("%d Rats, want one per opponent (3)", got)
	}
	witch := findBattlefieldByName(g, "Chittering Witch")
	victim := pushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	rat := findBattlefieldByName(g, "Rat")
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rat},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, victim); ok {
		t.Error("the 2/2 should have died to -2/-2")
	}
	if _, ok := battlefieldCard(g, rat); ok {
		t.Error("the sacrificed Rat is still there")
	}
}

func TestCmdrB18JunjiDrainsAndDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	junji := pushTestPermanent(g, me.ID, game.Card{Name: "Junji, the Midnight Sky", OracleID: cmdrB18JunjiOracle,
		TypeLine: "Legendary Creature — Dragon Spirit", Power: 5, Toughness: 5})
	before := opp.Life
	if err := g.DestroyPermanentForEffect(junji); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Junji's death asks for a mode")
	}
	answerMode(t, g, c, me.ID, 0)
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("opponent lost %d life, want 2", got)
	}
	if discardChoiceFor(g, opp.ID) == nil {
		t.Fatal("the opponent should be asked to discard two")
	}
	if got := discardChoiceFor(g, opp.ID).ChooseMax; got != 2 {
		t.Errorf("discard prompt asks for %d, want 2", got)
	}
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("the controller discards nothing")
	}
}

func TestCmdrB18JunjiReanimatesANonDragonAndYouLoseTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	junji := pushTestPermanent(g, me.ID, game.Card{Name: "Junji, the Midnight Sky", OracleID: cmdrB18JunjiOracle,
		TypeLine: "Legendary Creature — Dragon Spirit", Power: 5, Toughness: 5})
	dragon := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dragon, Name: "Dead Dragon", TypeLine: "Creature — Dragon", Owner: opp.ID, Controller: opp.ID})
	bear := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	before := me.Life
	if err := g.DestroyPermanentForEffect(junji); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Junji's death asks for a mode")
	}
	answerMode(t, g, c, me.ID, 1)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target asked")
	}
	if err := g.ResolvePickTarget(pick.ID, pick.Chooser, game.TargetRef{Kind: game.TargetCard, ID: dragon}); err == nil {
		t.Fatal("a Dragon is not a legal target")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	got, ok := battlefieldCard(g, bear)
	if !ok || got.Controller != me.ID {
		t.Fatalf("the Bear did not return under my control: %+v", got)
	}
	if me.Life != before-2 {
		t.Errorf("life %d, want %d", me.Life, before-2)
	}
}

func TestCmdrB18LordOfTheForsakenMillsAndHasGraveyardOnlyMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := pushTestPermanent(g, me.ID, game.Card{Name: "Lord of the Forsaken", OracleID: cmdrB18LordOracle,
		TypeLine: "Creature — Demon", Power: 6, Toughness: 6})
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	libBefore := opp.Library.Size()
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ActivateCatalogAbility(me.ID, lord, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := libBefore - opp.Library.Size(); got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
	// The Lord itself can't be the sacrifice.
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ActivateCatalogAbility(me.ID, lord, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{lord},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Error("the Lord can't sacrifice itself: it says another creature")
	}
	// The mana ability: pay 1 life, get {C} that only casts from a graveyard.
	life := me.Life
	me.ManaPool = nil
	if err := g.ActivateManaAbility(me.ID, lord, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "C" {
		t.Fatalf("pool = %+v, want one {C}", me.ManaPool)
	}
	cost, _ := game.ParseCost("{C}")
	spend := game.ManaSpendContext{Purpose: game.SpendPurposeCast, Types: []string{"Sorcery"}}
	spend.CastFrom = game.ZoneHand
	if me.ManaPool.CanPayFor(cost, 0, spend) {
		t.Error("the mana must not pay for a spell cast from hand")
	}
	spend.CastFrom = game.ZoneGraveyard
	if !me.ManaPool.CanPayFor(cost, 0, spend) {
		t.Error("the mana pays for a spell cast from a graveyard")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendContext{Purpose: game.SpendPurposeActivate}) {
		t.Error("the mana must not pay for an activation")
	}
}

func TestCmdrB18MindbladeRenderDrawsOncePerStepWhenAWarriorConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	pushTestPermanent(g, me.ID, game.Card{Name: "Mindblade Render", OracleID: cmdrB18RenderOracle,
		TypeLine: "Creature — Azra Warrior", Power: 1, Toughness: 3})
	w1 := pushTestPermanent(g, me.ID, game.Card{Name: "Warrior One", TypeLine: "Creature — Human Warrior", Power: 2, Toughness: 2})
	w2 := pushTestPermanent(g, me.ID, game.Card{Name: "Warrior Two", TypeLine: "Creature — Human Warrior", Power: 2, Toughness: 2})
	bear := pushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	handBefore, lifeBefore := me.Hand.Size(), me.Life
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(w1, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(w2, third.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("drew %d cards, want exactly 1", got)
	}
	if me.Life != lifeBefore-1 {
		t.Errorf("life %d, want %d", me.Life, lifeBefore-1)
	}
}

func TestCmdrB18MindbladeRenderIgnoresNonWarriorDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTestPermanent(g, me.ID, game.Card{Name: "Mindblade Render", OracleID: cmdrB18RenderOracle,
		TypeLine: "Creature — Azra Warrior", Power: 1, Toughness: 3})
	bear := pushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	handBefore := me.Hand.Size()
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Errorf("a Bear drew %d cards", got)
	}
}

func TestCmdrB18NecromanticSelectionReturnsOneAsABlackZombieAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 3, 3)
	spell := castCatalogSpell(t, g, "Necromantic Selection", "Sorcery", cmdrB18SelectionOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 2 {
		t.Fatalf("want a pick between the two destroyed creature cards, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	got, ok := battlefieldCard(g, theirs)
	if !ok || got.Controller != me.ID {
		t.Fatalf("their Bear did not return under my control: %+v", got)
	}
	if _, ok := battlefieldCard(g, mine); ok {
		t.Error("my Bear should be in the graveyard")
	}
	if c := layeredCard(t, g, theirs); !c.HasSubtype("Zombie") || !c.HasSubtype("Test") || !c.HasColor("B") {
		t.Errorf("it should be a black Zombie with its other types: %v %v", c.TypeLine, c.Colors)
	}
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("Necromantic Selection should be exiled, not in the graveyard")
	}
}

func TestCmdrB18NecromanticSelectionWithNoCreatureCardJustExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := castCatalogSpell(t, g, "Necromantic Selection", "Sorcery", cmdrB18SelectionOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) {
		t.Error("the spell should still exile itself")
	}
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("nothing to choose")
	}
}

func TestCmdrB18NecropolisFiendShrinksByTheCardsExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fiend := pushTestPermanent(g, me.ID, game.Card{Name: "Necropolis Fiend", OracleID: cmdrB18FiendOracle,
		TypeLine: "Creature — Demon", Power: 4, Toughness: 5})
	big := pushVanillaCreature(g, opp.ID, "Big", 3, 3)
	var fuel []uuid.UUID
	for i := 0; i < 3; i++ {
		fuel = append(fuel, pushGraveyardCardForTest(me, "Fuel"))
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, fiend, 0, game.ActivateAbilityParams{
		Strict:   true,
		XValue:   3,
		ExileIDs: fuel,
		Targets:  []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range fuel {
		if !g.Exile.Contains(id) {
			t.Error("a named card was not exiled")
		}
	}
	if _, ok := battlefieldCard(g, big); ok {
		t.Error("a 3/3 under -3/-3 should have died")
	}
}

func TestCmdrB18NecropolisFiendXIsTheNumberOfCardsNamed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fiend := pushTestPermanent(g, me.ID, game.Card{Name: "Necropolis Fiend", OracleID: cmdrB18FiendOracle,
		TypeLine: "Creature — Demon", Power: 4, Toughness: 5})
	big := pushVanillaCreature(g, opp.ID, "Big", 3, 3)
	fuel := []uuid.UUID{pushGraveyardCardForTest(me, "Fuel"), pushGraveyardCardForTest(me, "Fuel")}
	// Two cards named, X sent as 5: X is what the cards say, and only two mana is owed.
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, fiend, 0, game.ActivateAbilityParams{
		Strict:   true,
		XValue:   5,
		ExileIDs: fuel,
		Targets:  []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, big); p != 1 {
		t.Errorf("power %d, want 1 (3 - 2)", p)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool %v: two mana were owed for two cards", me.ManaPool)
	}
	// With no mana left, naming a card still costs {1}.
	more := pushGraveyardCardForTest(me, "More Fuel")
	advanceToDeclareAttackersOf(t, g, 1)
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.ActivateCatalogAbility(me.ID, fiend, 0, game.ActivateAbilityParams{
		Strict: true, XValue: 1, ExileIDs: []uuid.UUID{more},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	}); err == nil {
		t.Error("an activation with no mana to pay {1} should be refused")
	}
}

func TestCmdrB18TasigurHasDelve(t *testing.T) {
	for _, oracle := range []string{cmdrB18TasigurOracle, cmdrB18FiendOracle} {
		spec, ok := Lookup(oracle)
		if !ok || !spec.Delve {
			t.Errorf("%s should have delve", oracle)
		}
	}
}

func TestCmdrB18TasigurMillsThenAnOpponentChoosesANonlandCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tas := pushTestPermanent(g, me.ID, game.Card{Name: "Tasigur, the Golden Fang", OracleID: cmdrB18TasigurOracle,
		TypeLine: "Legendary Creature — Human Shaman", Power: 4, Toughness: 5})
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	libBefore := me.Library.Size()
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "U"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, tas, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := libBefore - me.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
	// Three opponents: the controller names one, who then chooses.
	opp := g.Seats[1]
	if latestOptionPickFor(g, me.ID) != nil {
		answerChoosePlayer(t, g, me.ID, opp)
	}
	pick := latestChooseCardsFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the chosen opponent was not asked: %+v", g.PendingChoices)
	}
	for _, id := range pick.ChooseCards {
		if id == land {
			t.Error("a land was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{spell}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(spell) {
		t.Error("the chosen card did not go to my hand")
	}
}
