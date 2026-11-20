package testsu

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// usg_batch4_cards_test.go — Urza's Saga, batch 4 (#5624). Each test
// plays the real catalog entry and asserts what its text says.

const (
	usgB4DespondencyOracle       = "a68ac1ba-f51e-446c-abf0-7a2bd4d65a8c"
	usgB4DestructiveUrgeOracle   = "5ab62b09-1135-4d22-866b-3a0d64228bf9"
	usgB4DiscipleOfGraceOracle   = "7f4ceca9-0f55-4fa6-8e64-befbc4303d4e"
	usgB4DiscipleOfLawOracle     = "a4689767-46c3-42ab-af27-5877e226bc05"
	usgB4DisorderOracle          = "586da84b-1384-4da2-983d-340c22b73eb8"
	usgB4DisruptiveStudentOracle = "74315869-46e3-43a2-abc0-d86d0a70d343"
	usgB4DouseOracle             = "3bdc96d9-7272-4e38-b81a-ec928dd49eaa"
	usgB4DragonBloodOracle       = "b751c1c6-195d-4023-9d0b-3c91d6b834d4"
	usgB4DriftingDjinnOracle     = "fdea1945-0bb6-4a83-8cb1-5ba6a4dc39bd"
	usgB4DriftingMeadowOracle    = "c6eb2814-0021-4308-ad44-6c8cc59b0d1c"
	usgB4DromosaurOracle         = "4576cc5b-7b3a-41a7-9b84-acadf2d4adf5"
	usgB4EasternPaladinOracle    = "b445482a-9422-41c7-bedb-3b9f500d2f27"
	usgB4ElectryteOracle         = "5ff0c25f-4593-460b-9d7f-d9c37f135a18"
	usgB4EliteArchersOracle      = "2230ec18-4dcf-435e-b495-1905cb94e8b8"
)

func TestUsgB4EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		usgB4DespondencyOracle:       "Despondency",
		usgB4DestructiveUrgeOracle:   "Destructive Urge",
		usgB4DiscipleOfGraceOracle:   "Disciple of Grace",
		usgB4DiscipleOfLawOracle:     "Disciple of Law",
		usgB4DisorderOracle:          "Disorder",
		usgB4DisruptiveStudentOracle: "Disruptive Student",
		usgB4DouseOracle:             "Douse",
		usgB4DragonBloodOracle:       "Dragon Blood",
		usgB4DriftingDjinnOracle:     "Drifting Djinn",
		usgB4DriftingMeadowOracle:    "Drifting Meadow",
		usgB4DromosaurOracle:         "Dromosaur",
		usgB4EasternPaladinOracle:    "Eastern Paladin",
		usgB4ElectryteOracle:         "Electryte",
		usgB4EliteArchersOracle:      "Elite Archers",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// usgB4Cycles cycles a card from the active seat's hand for {2} and checks
// it was replaced by one draw.
func usgB4Cycles(t *testing.T, name, typeLine, oracle string) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	id, _ := effectstest.CycleFromHand(t, g, name, typeLine, oracle, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Errorf("%s: the cycled card is not in the graveyard", name)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("%s: hand %d → %d, want the cycled card replaced by one draw", name, hand, me.Hand.Size())
	}
}

// usgB4Attacking seeds a creature of owner's that is attacking `defender`.
func usgB4Attacking(g *game.Game, owner, defender uuid.UUID, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Attacker", TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
		AttackingTarget: defender,
	})
}

// usgB4Blocking seeds a creature of owner's that is blocking `attacker`.
func usgB4Blocking(g *game.Game, owner, attacker uuid.UUID, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blocker", TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
		BlockingTarget: attacker,
	})
}

// --- Despondency ---------------------------------------------------------------------

func TestUsgB4DespondencyShrinksAndReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	aura := effectstest.AuraCast(t, g, "Despondency", usgB4DespondencyOracle, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 1 || tough != 3 {
		t.Errorf("the enchanted Bear is %d/%d, want 1/3", p, tough)
	}
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID},
		bear,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || me.Graveyard.Contains(aura) {
		t.Fatal("Despondency is still on the battlefield or in the graveyard")
	}
	if !me.Hand.Contains(aura) {
		t.Error("Despondency did not return to its owner's hand")
	}
}

// --- Destructive Urge ----------------------------------------------------------------

func TestUsgB4DestructiveUrgeMakesTheDamagedPlayerSacrificeALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Permanent(g, me.ID, "My Forest", "Basic Land — Forest")
	a := effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	b := effectstest.B12Permanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	effectstest.AuraCast(t, g, "Destructive Urge", usgB4DestructiveUrgeOracle, bear)

	// Damage that is not combat damage does nothing.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: bear, Target: opp.ID, Amount: 2})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil {
		t.Fatal("noncombat damage asked for a sacrifice")
	}

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the damaged player was not asked to sacrifice a land")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, b)
	if g.Battlefield.Contains(b) || !g.Battlefield.Contains(a) || !g.Battlefield.Contains(mine) {
		t.Error("want only the chosen land of the damaged player sacrificed")
	}
}

// --- Disciple of Grace, Disciple of Law ------------------------------------------------

func TestUsgB4DisciplesHaveProtectionAndCycle(t *testing.T) {
	for _, tc := range []struct{ name, oracle, keyword string }{
		{"Disciple of Grace", usgB4DiscipleOfGraceOracle, "protection from black"},
		{"Disciple of Law", usgB4DiscipleOfLawOracle, "protection from red"},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, "Creature — Human Cleric", tc.oracle, 1, 2)
		if !effectstest.HasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s lacks %s", tc.name, tc.keyword)
		}
		usgB4Cycles(t, tc.name, "Creature — Human Cleric", tc.oracle)
	}
}

// --- Disorder -----------------------------------------------------------------------

func TestUsgB4DisorderHitsWhiteCreaturesAndTheirControllers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	dying := effectstest.UsgB3Colored(g, opp.ID, "Their White Bear", "W")
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Third's White Giant", TypeLine: "Creature — Giant",
		Colors: []string{"W"}, Power: 4, Toughness: 4, Owner: third.ID, Controller: third.ID,
	})
	red := effectstest.UsgB3Colored(g, me.ID, "My Red Bear", "R")
	lives := map[uuid.UUID]int{me.ID: me.Life, opp.ID: opp.Life, third.ID: third.Life}
	effectstest.CastCatalogSpell(t, g, "Disorder", "Sorcery", usgB4DisorderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dying) || !opp.Graveyard.Contains(dying) {
		t.Error("the 2/2 white creature did not die")
	}
	if got := effectstest.DamageMarkedOn(g, big); got != 2 {
		t.Errorf("the white 4/4 has %d damage, want 2", got)
	}
	if got := effectstest.DamageMarkedOn(g, red); got != 0 {
		t.Errorf("the red creature has %d damage, want 0", got)
	}
	// The opponent's only white creature died to this damage, and they
	// were still dealt 2: who controls one is read before the damage.
	for _, tc := range []struct {
		who  *game.Player
		want int
	}{{me, 0}, {opp, 2}, {third, 2}} {
		if got := lives[tc.who.ID] - tc.who.Life; got != tc.want {
			t.Errorf("seat %s lost %d life, want %d", tc.who.ID, got, tc.want)
		}
	}
}

// --- Disruptive Student ------------------------------------------------------------------

func TestUsgB4DisruptiveStudentCountersUnlessOnePaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		student := effectstest.B12Push(g, me.ID, "Disruptive Student", "Creature — Human Wizard", usgB4DisruptiveStudentOracle, 1, 1)
		spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
		if err := g.ActivateCatalogAbility(me.ID, student, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(spell)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if !effectstest.B13Tapped(t, g, student) {
			t.Error("the Student did not tap as the cost")
		}
		effectstest.PassUntilTaxed(t, g, me.ID)
		if pay {
			effectstest.FillPool(me, 1)
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if g.Stack.Contains(spell) != pay {
			t.Errorf("paid=%v: the spell is still on the stack = %v", pay, g.Stack.Contains(spell))
		}
	}
}

// --- Douse ---------------------------------------------------------------------------

func TestUsgB4DouseCountersOnlyARedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	douse := effectstest.PushCatalogPermanent(g, me.ID, "Douse", "Enchantment", usgB4DouseOracle, false)
	blue := effectstest.Batch01OpponentCasts(t, g, opp, "Blue Instant", "", "{U}", nil)
	if err := effectstest.UsgB2Activate(t, g, me, douse, "CU", effectstest.CardRefs(blue)); err == nil {
		t.Error("Douse targeted a blue spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })

	red := effectstest.Batch01OpponentCasts(t, g, opp, "Red Instant", "", "{R}", nil)
	if err := effectstest.UsgB2Activate(t, g, me, douse, "U", effectstest.CardRefs(red)); err == nil {
		t.Fatal("Douse activated for {U}; it costs {1}{U}")
	}
	if err := effectstest.UsgB2Activate(t, g, me, douse, "C", effectstest.CardRefs(red)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(red) || !opp.Graveyard.Contains(red) {
		t.Error("the red spell was not countered")
	}
}

// --- Dragon Blood --------------------------------------------------------------------

func TestUsgB4DragonBloodPutsACounterOnTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	blood := effectstest.PushCatalogPermanent(g, me.ID, "Dragon Blood", "Artifact", usgB4DragonBloodOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := effectstest.UsgB2Activate(t, g, me, blood, "CC", effectstest.CardRefs(bear)); err == nil {
		t.Fatal("activated for {2}; it costs {3}")
	}
	if err := effectstest.UsgB2Activate(t, g, me, blood, "C", effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr8Counters(g, bear, "+1/+1"); got != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", got)
	}
	if !effectstest.B13Tapped(t, g, blood) {
		t.Error("Dragon Blood did not tap")
	}
}

// --- Drifting Djinn ----------------------------------------------------------------

func TestUsgB4DriftingDjinnFliesPaysItsUpkeepAndCycles(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		djinn := effectstest.B12Push(g, me.ID, "Drifting Djinn", "Creature — Djinn", usgB4DriftingDjinnOracle, 5, 5)
		if !effectstest.HasKeywordOnBattlefield(t, g, djinn, "flying") {
			t.Error("Drifting Djinn lacks flying")
		}
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.DcMana(t, g, me.ID, "{C}{U}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if g.Battlefield.Contains(djinn) != pay {
			t.Errorf("pay=%v: on the battlefield = %v", pay, g.Battlefield.Contains(djinn))
		}
	}
	usgB4Cycles(t, "Drifting Djinn", "Creature — Djinn", usgB4DriftingDjinnOracle)
}

// --- Drifting Meadow -----------------------------------------------------------------

func TestUsgB4DriftingMeadowEntersTappedTapsForWhiteAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.B12PlayFromHand(t, g, "Drifting Meadow", "Land", usgB4DriftingMeadowOracle, game.CastSpellParams{})
	if !effectstest.B13Tapped(t, g, land) {
		t.Error("Drifting Meadow entered untapped")
	}
	effectstest.UntapForTest(g, land)
	if got := effectstest.LtProduced(t, g, land); !slices.Equal(got, []string{"{W}"}) {
		t.Errorf("Drifting Meadow adds %v, want {W}", got)
	}
	usgB4Cycles(t, "Drifting Meadow", "Land", usgB4DriftingMeadowOracle)
}

// --- Dromosaur ----------------------------------------------------------------------

func TestUsgB4DromosaurGetsPlusTwoMinusTwoOncePerBlockOrBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dino := effectstest.B12Push(g, me.ID, "Dromosaur", "Creature — Dinosaur", usgB4DromosaurOracle, 2, 3)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: dino, Target: a, Amount: 1})
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: dino, Target: b, Amount: 2})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, dino); p != 4 || tough != 1 {
		t.Fatalf("after blocking two attackers Dromosaur is %d/%d, want 4/1: blocks triggers once", p, tough)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	dino = effectstest.B12Push(g, me.ID, "Dromosaur", "Creature — Dinosaur", usgB4DromosaurOracle, 2, 3)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: dino, CardID: dino, Target: dino})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, dino); p != 4 || tough != 1 {
		t.Errorf("after becoming blocked Dromosaur is %d/%d, want 4/1", p, tough)
	}
}

// --- Eastern Paladin ----------------------------------------------------------------

func TestUsgB4EasternPaladinDestroysOnlyAGreenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	paladin := effectstest.B12Push(g, me.ID, "Eastern Paladin", "Creature — Phyrexian Zombie Knight", usgB4EasternPaladinOracle, 3, 3)
	green := effectstest.UsgB3Colored(g, opp.ID, "Their Elf", "G")
	white := effectstest.UsgB3Colored(g, opp.ID, "Their Knight", "W")
	if err := effectstest.UsgB2Activate(t, g, me, paladin, "BB", effectstest.CardRefs(white)); err == nil {
		t.Fatal("Eastern Paladin targeted a white creature")
	}
	if err := effectstest.UsgB2Activate(t, g, me, paladin, "", effectstest.CardRefs(green)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) || !opp.Graveyard.Contains(green) {
		t.Error("the green creature was not destroyed")
	}
	if !g.Battlefield.Contains(white) {
		t.Error("the white creature was destroyed")
	}
}

// --- Electryte ------------------------------------------------------------------------

func TestUsgB4ElectryteHitsEveryBlockingCreatureWithItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	electryte := effectstest.B12Push(g, me.ID, "Electryte", "Creature — Trilobite Beast", usgB4ElectryteOracle, 3, 3)
	other := usgB4Attacking(g, me.ID, opp.ID, 2, 2)
	blocker := usgB4Blocking(g, opp.ID, other, 1, 5)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 1, 5)

	// Combat damage to a creature is not to defending player.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: electryte, Target: idle, Amount: 3, Combat: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, blocker); got != 0 {
		t.Fatalf("damage to a creature triggered it: the blocker has %d damage", got)
	}

	effectstest.DealCombatDamageToPlayer(g, electryte, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, blocker); got != 3 {
		t.Errorf("the blocking creature has %d damage, want 3 (Electryte's power)", got)
	}
	if got := effectstest.DamageMarkedOn(g, idle); got != 0 {
		t.Errorf("a creature that is not blocking has %d damage, want 0", got)
	}
}

// --- Elite Archers ------------------------------------------------------------------

func TestUsgB4EliteArchersShootAnAttackingOrBlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	archers := effectstest.B12Push(g, me.ID, "Elite Archers", "Creature — Human Soldier Archer", usgB4EliteArchersOracle, 3, 3)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 5)
	attacker := usgB4Attacking(g, opp.ID, me.ID, 2, 5)
	if err := effectstest.UsgB2Activate(t, g, me, archers, "", effectstest.CardRefs(idle)); err == nil {
		t.Fatal("Elite Archers targeted a creature that is neither attacking nor blocking")
	}
	if err := effectstest.UsgB2Activate(t, g, me, archers, "", effectstest.CardRefs(attacker)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, attacker); got != 3 {
		t.Errorf("the attacker has %d damage, want 3", got)
	}
}
