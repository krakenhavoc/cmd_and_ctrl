package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b22_test.go — reprint pass, batch 22 (#5908): Jund Panorama,
// Kess, Kresh, Leafkin Druid, Littjara Mirrorlake, Mask of Riddles,
// Mitotic Slime, Oracle's Vault, Power Conduit, Primal Empathy and
// Profane Command, each driven through a real cast, activation or event.

const (
	rprB22JundPanoramaOracle = "f39f33ac-074d-442d-ae4c-1d694ee315f3"
	rprB22KessOracle         = "f5092c14-eec4-472c-999c-ba96c36b2fbb"
	rprB22KreshOracle        = "6343cf6f-b1e3-49ce-8933-8fe6c9dc83f3"
	rprB22LeafkinOracle      = "55eb9310-f24b-410b-850b-a7d0d5117946"
	rprB22MirrorlakeOracle   = "3b577179-10d0-43f1-ac17-9dd2d12c965d"
	rprB22MaskOracle         = "a50c55a7-4e7e-40f6-b76c-1e32598dc9b6"
	rprB22SlimeOracle        = "1162c7ff-c0cb-469a-ad3c-322b0981ae1d"
	rprB22VaultOracle        = "220e3298-1663-4731-b50f-247076e1328b"
	rprB22ConduitOracle      = "2e358f7e-c282-4b92-8255-1436c99cda49"
	rprB22EmpathyOracle      = "24645d9e-5e6d-411d-be41-c2bd1659937f"
	rprB22ProfaneOracle      = "6f9ecfc4-85bd-4171-9031-dbe0e39e0135"
)

func TestRprB22CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB22JundPanoramaOracle: "Jund Panorama",
		rprB22KessOracle:         "Kess, Dissident Mage",
		rprB22KreshOracle:        "Kresh the Bloodbraided",
		rprB22LeafkinOracle:      "Leafkin Druid",
		rprB22MirrorlakeOracle:   "Littjara Mirrorlake",
		rprB22MaskOracle:         "Mask of Riddles",
		rprB22SlimeOracle:        "Mitotic Slime",
		rprB22VaultOracle:        "Oracle's Vault",
		rprB22ConduitOracle:      "Power Conduit",
		rprB22EmpathyOracle:      "Primal Empathy",
		rprB22ProfaneOracle:      "Profane Command",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q completeness %v", name, spec.Name, spec.Completeness)
		}
	}
}

// --- Jund Panorama -------------------------------------------------

func TestRprB22JundPanoramaFetchesASwampMountainOrForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	panorama := b43Catalog(g, me.ID, "Jund Panorama", "Land", rprB22JundPanoramaOracle, 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, panorama, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the fetch opens a search")
	}
	if searchOptionNamed(g, c, "Plains") != uuid.Nil {
		t.Error("a Plains is none of Swamp, Mountain or Forest")
	}
	swamp := searchOptionNamed(g, c, "Swamp")
	if swamp == uuid.Nil || searchOptionNamed(g, c, "Forest") == uuid.Nil {
		t.Fatal("the Swamp and the Forest are offered")
	}
	answerSearchByID(t, g, me.ID, swamp)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(panorama) {
		t.Error("the Panorama sacrificed itself to pay")
	}
	if !g.Battlefield.Contains(swamp) || !b16Tapped(t, g, swamp) {
		t.Error("the basic arrives on the battlefield tapped")
	}
}

// --- Kess, Dissident Mage -----------------------------------------

func rprB22KessTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Kess, Dissident Mage", "Legendary Creature — Human Wizard", rprB22KessOracle, false)
	return g, me
}

// rprB22GraveSpell puts a {0} catalog spell into a graveyard.
func rprB22GraveSpell(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{0}", OracleID: oracle,
		Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

func TestRprB22KessCastsOneInstantOrSorceryFromTheGraveyardAndExilesIt(t *testing.T) {
	g, me := rprB22KessTable(t)
	seedCheapLibrary(g, me, 6)
	burn := rprB22GraveSpell(me, "Divination", "Sorcery", "273b339c-964b-4a18-8eb5-ceb8abcdfd9e")
	counter := ptGraveCard(me, "Counter", "Instant")
	bear := ptGraveCard(me, "Bear", "Creature — Bear")
	if enumeratedFor(g, me.ID, bear) {
		t.Error("a creature card is offered; Kess opens instants and sorceries")
	}
	if err := muldrothaCast(g, me, bear, ""); err == nil {
		t.Error("a creature spell was cast from the graveyard")
	}
	// Cast without claiming Kess's permission, the card has no way out of the
	// graveyard at all: the exile clause cannot be skipped.
	if err := g.CastSpell(me.ID, burn, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("the sorcery was cast from the graveyard without Kess's permission")
	}
	if err := g.CastSpell(me.ID, burn, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "kess_dissident_mage"}); err != nil {
		t.Fatalf("the sorcery from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(burn) || me.Graveyard.Contains(burn) {
		t.Error("the spell cast this way is exiled instead of going to the graveyard")
	}
	// One cast of either kind uses the turn's single permission.
	if err := g.CastSpell(me.ID, counter, game.CastSpellParams{FromZone: "graveyard", Strict: true, AlternativeCost: "kess_dissident_mage"}); err == nil {
		t.Error("a second spell was cast from the graveyard this turn")
	}
	if enumeratedFor(g, me.ID, counter) {
		t.Error("the enumerator still offers a second spell this turn")
	}
}

// --- Kresh the Bloodbraided ----------------------------------------

func TestRprB22KreshTakesCountersEqualToTheDeadCreaturesPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	kresh := b12Push(g, me.ID, "Kresh the Bloodbraided", "Legendary Creature — Human Warrior", rprB22KreshOracle, 3, 3)
	victim := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 4, 4)
	addCountersForTest(g, victim, game.CounterPlusOne, 1) // last-known power 5
	killCreature(t, g, opp.ID, victim)
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(kresh)
	if got := c.Counters[game.CounterPlusOne]; got != 5 {
		t.Errorf("Kresh has %d +1/+1 counters, want 5 (the dead creature's last-known power)", got)
	}
}

func TestRprB22KreshMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	kresh := b12Push(g, me.ID, "Kresh the Bloodbraided", "Legendary Creature — Human Warrior", rprB22KreshOracle, 3, 3)
	victim := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 4, 4)
	killCreature(t, g, opp.ID, victim)
	b06AnswerAllTriggerPrompts(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(kresh)
	if got := c.Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("declined, yet Kresh has %d counters", got)
	}
}

func TestRprB22KreshIgnoresItsOwnDeathAndNonCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	kresh := b12Push(g, me.ID, "Kresh the Bloodbraided", "Legendary Creature — Human Warrior", rprB22KreshOracle, 3, 3)
	art := b12Push(g, me.ID, "Trinket", "Artifact", "", 0, 0)
	killCreature(t, g, me.ID, art)
	if owedMays(g, me.ID) != 0 {
		t.Error("a dying artifact triggers Kresh")
	}
	killCreature(t, g, me.ID, kresh)
	if owedMays(g, me.ID) != 0 {
		t.Error("Kresh's own death triggers it; it says another creature")
	}
}

// --- Leafkin Druid ---------------------------------------------------

func TestRprB22LeafkinDruidAddsOneGUntilYouControlFourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	druid := b12Push(g, me.ID, "Leafkin Druid", "Creature — Elemental Druid", rprB22LeafkinOracle, 0, 3)
	b28TapForMana(t, g, me.ID, druid, "")
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Fatalf("with one creature the pool is %v, want [G]", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(druid) })
	me.ManaPool = nil
	for i := 0; i < 3; i++ {
		b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	}
	b28TapForMana(t, g, me.ID, druid, "")
	if got := poolColors(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("with four creatures the pool is %v, want [G G]", got)
	}
}

// --- Littjara Mirrorlake --------------------------------------------

func TestRprB22MirrorlakeCopiesACreatureWithAnExtraCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	lake := b43Catalog(g, me.ID, "Littjara Mirrorlake", "Land", rprB22MirrorlakeOracle, 0, 0)
	bear := b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	b06AddMana(me, "G", "G", "U", "C", "C")
	// A sorcery-speed ability: refused with a nonempty stack is covered by the engine;
	// here the illegal target is the thing to refuse.
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, lake, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs), Strict: true}); err == nil {
		t.Fatal("an opponent's creature is not a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, lake, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lake) {
		t.Error("the land is sacrificed to pay")
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.Name == "Grizzly Bears" && c.InstanceID != bear {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil || !token.IsToken() {
		t.Fatal("no token copy of the Bears was created")
	}
	if got := token.Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("the copy has %d +1/+1 counters, want 1", got)
	}
	if orig, _ := g.LookupCardForEffect(bear); orig.Counters[game.CounterPlusOne] != 0 {
		t.Error("the original gained a counter")
	}
}

// --- Mask of Riddles -------------------------------------------------

func TestRprB22MaskGrantsFearAndDrawsOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	mask := seedEquipment(g, me.ID, "Mask of Riddles", rprB22MaskOracle)
	if hasKeywordOnBattlefield(t, g, bear, "fear") {
		t.Fatal("fear before equipping")
	}
	equipTo(t, g, me.ID, mask, bear)
	if !hasKeywordOnBattlefield(t, g, bear, "fear") {
		t.Error("the equipped creature has fear")
	}
	hand := me.Hand.Size()
	b27Damage(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("noncombat damage draws nothing")
	}
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("combat damage to a player draws one: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- Mitotic Slime ---------------------------------------------------

func TestRprB22MitoticSlimeSplitsTwiceOver(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	slime := b12Push(g, me.ID, "Mitotic Slime", "Creature — Ooze", rprB22SlimeOracle, 4, 4)
	killCreature(t, g, me.ID, slime)
	passPriorityAroundTable(t, g)
	twos := 0
	var first uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ooze" && c.IsToken() && c.Power == 2 && c.Toughness == 2 {
			twos++
			if first == uuid.Nil {
				first = c.InstanceID
			}
		}
	}
	if twos != 2 {
		t.Fatalf("%d 2/2 Ooze tokens, want 2", twos)
	}
	killCreature(t, g, me.ID, first)
	passPriorityAroundTable(t, g)
	ones := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ooze" && c.IsToken() && c.Power == 1 && c.Toughness == 1 {
			ones++
		}
	}
	if ones != 2 {
		t.Errorf("a dying 2/2 Ooze left %d 1/1 Oozes, want 2", ones)
	}
}

// --- Oracle's Vault --------------------------------------------------

func TestRprB22OraclesVaultBricksUpAndThenPlaysForFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	vault := b43Catalog(g, me.ID, "Oracle's Vault", "Artifact", rprB22VaultOracle, 0, 0)
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top" + string(rune('0'+i)), TypeLine: "Sorcery", ManaCost: "{5}", Owner: me.ID, Controller: me.ID})
	}
	// The free ability needs three brick counters.
	if err := g.ActivateCatalogAbility(me.ID, vault, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the free ability is open with no brick counters")
	}
	for i := 0; i < 3; i++ {
		b06AddMana(me, "C", "C")
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vault) })
		if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		me.ManaPool = nil
	}
	if c, _ := g.LookupCardForEffect(vault); c.Counters["brick"] != 3 {
		t.Fatalf("brick counters = %d, want 3", c.Counters["brick"])
	}
	if g.Exile.Size() != 3 {
		t.Fatalf("exile holds %d cards, want 3", g.Exile.Size())
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vault) })
	if err := g.ActivateCatalogAbility(me.ID, vault, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the free ability with three bricks: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Exile.Size() != 4 {
		t.Fatalf("exile holds %d cards, want 4", g.Exile.Size())
	}
	// The last card pushed (Top3) is exiled first by the paid ability and is
	// castable only for its {5}; the one exiled by the free ability (Top0)
	// costs nothing.
	var paid, free uuid.UUID
	for _, c := range g.Exile.Cards {
		switch c.Name {
		case "Top3":
			paid = c.InstanceID
		case "Top0":
			free = c.InstanceID
		}
	}
	if err := g.CastSpell(me.ID, paid, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("a card exiled by the paid ability is cast without paying its mana cost")
	}
	if err := g.CastSpell(me.ID, free, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting the card the free ability exiled, for free: %v", err)
	}
}

// --- Power Conduit ---------------------------------------------------

func TestRprB22PowerConduitMovesACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	conduit := b43Catalog(g, me.ID, "Power Conduit", "Artifact", rprB22ConduitOracle, 0, 0)
	bear := seedBear(g, me.ID)
	rock := b43Catalog(g, me.ID, "Rock", "Artifact", "", 0, 0)
	addCountersForTest(g, rock, game.CounterCharge, 1)
	// The counter is removed from a permanent you control as a cost; the +1/+1
	// goes on the bear.
	if err := g.ActivateCatalogAbility(me.ID, conduit, 0, game.ActivateAbilityParams{
		Modes:            []int{1},
		Targets:          []game.TargetRef{modeRef(game.TargetCard, bear, 0, 0)},
		CounterSourceIDs: []uuid.UUID{rock},
		CounterCounts:    []int{1},
		CounterKinds:     []string{game.CounterCharge},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(rock); c.Counters[game.CounterCharge] != 0 {
		t.Errorf("the Rock still has %d charge counters", c.Counters[game.CounterCharge])
	}
	if c, _ := g.LookupCardForEffect(bear); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
	}
}

func TestRprB22PowerConduitCannotActivateWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	conduit := b43Catalog(g, me.ID, "Power Conduit", "Artifact", rprB22ConduitOracle, 0, 0)
	bear := seedBear(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, conduit, 0, game.ActivateAbilityParams{
		Modes:   []int{1},
		Targets: []game.TargetRef{modeRef(game.TargetCard, bear, 0, 0)},
		Strict:  true,
	}); err == nil {
		t.Error("activated with no counter anywhere to remove")
	}
}

// --- Primal Empathy --------------------------------------------------

func TestRprB22PrimalEmpathyDrawsWhenYouHoldTheGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Primal Empathy", "Enchantment", rprB22EmpathyOracle, false)
	mine := b12Creature(g, me.ID, "Big", "Creature — Giant", 5, 5)
	b12Creature(g, opp.ID, "Small", "Creature — Bear", 5, 5) // a tie for greatest counts
	hand := me.Hand.Size()
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want a draw", hand, me.Hand.Size())
	}
	if c, _ := g.LookupCardForEffect(mine); c.Counters[game.CounterPlusOne] != 0 {
		t.Error("it drew, so no counter")
	}
}

func TestRprB22PrimalEmpathyPutsACounterOtherwise(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Primal Empathy", "Enchantment", rprB22EmpathyOracle, false)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	hand := me.Hand.Size()
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want no draw", hand, me.Hand.Size())
	}
	if c, _ := g.LookupCardForEffect(small); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the lone creature has %d counters, want 1", c.Counters[game.CounterPlusOne])
	}
}

func TestRprB22PrimalEmpathyAsksWhichCreatureGetsTheCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Primal Empathy", "Enchantment", rprB22EmpathyOracle, false)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	b12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, b)
	if c, _ := g.LookupCardForEffect(b); c.Counters[game.CounterPlusOne] != 1 {
		t.Error("the chosen creature gets the counter")
	}
	if c, _ := g.LookupCardForEffect(a); c.Counters[game.CounterPlusOne] != 0 {
		t.Error("the other creature gets none")
	}
}

// --- Profane Command -------------------------------------------------

func TestRprB22ProfaneCommandDrainsAndShrinks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	life := opp.Life
	b19CastXModal(t, g, "Profane Command", "Sorcery", rprB22ProfaneOracle, "{X}{B}{B}", 3,
		[]int{0, 2},
		[]game.TargetRef{
			modeRef(game.TargetPlayer, opp.ID, 0, 0),
			modeRef(game.TargetCard, victim, 1, 0),
		})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want 3 lost", life, opp.Life)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("-3/-3 kills the 3/3")
	}
	_ = me
}

func TestRprB22ProfaneCommandReanimatesWithinXAndGivesFear(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	small := graveCreature(me, "Small", "{2}")
	big := graveCreature(me, "Big", "{5}")
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	// Mana value 5 is out of reach with X=2.
	hand := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: hand, Name: "Profane Command", TypeLine: "Sorcery", OracleID: rprB22ProfaneOracle, ManaCost: "{X}{B}{B}", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Modes: []int{1, 3}, XValue: 2, Strict: true, Targets: []game.TargetRef{
		modeRef(game.TargetCard, big, 0, 0),
		modeRef(game.TargetCard, a, 1, 0),
	}}); err == nil {
		t.Error("a mana value 5 creature card is a legal target with X=2")
	}
	b19CastXModal(t, g, "Profane Command", "Sorcery", rprB22ProfaneOracle, "{X}{B}{B}", 2,
		[]int{1, 3},
		[]game.TargetRef{
			modeRef(game.TargetCard, small, 0, 0),
			modeRef(game.TargetCard, a, 1, 0),
			modeRef(game.TargetCard, b, 1, 0),
		})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the mana-value-2 creature returned")
	}
	if g.Battlefield.Contains(big) {
		t.Error("the mana-value-5 creature stayed")
	}
	if !hasKeywordOnBattlefield(t, g, a, "fear") || !hasKeywordOnBattlefield(t, g, b, "fear") {
		t.Error("both targeted creatures gain fear")
	}
}
