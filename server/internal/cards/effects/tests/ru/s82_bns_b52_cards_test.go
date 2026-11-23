package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b52_cards_test.go — Foundations Jumpstart legends, batch 52
// (#5929). Each test plays the real catalog entry through the engine.

const (
	b52PolOracle      = "2f200a14-7e5b-4b7c-a283-5cf9abb1f95d"
	b52PsemillaOracle = "12712419-a2a9-483d-a65a-2b727ae404ae"
	b52QalaOracle     = "1fc1b0d7-7f15-4d00-bc8f-e06489afd425"
	b52RevOracle      = "8345eb05-c04e-4c1d-9776-5ab7d91a7842"
	b52ShroofusOracle = "e0cebb09-d10d-4b70-96be-82df0e5d52d3"
	b52SlinzaOracle   = "8bed095d-1538-4c80-b147-3ec308a8acb1"
	b52SutinaOracle   = "91f410c5-4313-462f-8d21-da63988d919f"
	b52TaboraxOracle  = "955abada-aba2-41c2-96ff-58e3679ad7e5"
	b52TaekoOracle    = "3d3eb043-8ce6-461a-b8bc-67f0ba6cd580"
	b52ThuridOracle   = "0b96cd7a-2132-4ff2-9666-6e2fbd0a2160"
	b52UrdnanOracle   = "adc1f850-9217-4a90-8e6a-02ed226ef8df"
)

func TestBNSB52EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b52PolOracle:      "Pol Jamaar, Illusionist",
		b52PsemillaOracle: "Psemilla, Meletian Poet",
		b52QalaOracle:     "Qala, Ajani's Pridemate",
		b52RevOracle:      "Rev, Tithe Extractor",
		b52ShroofusOracle: "Shroofus Sproutsire",
		b52SlinzaOracle:   "Slinza, the Spiked Stampede",
		b52SutinaOracle:   "Sutina, Speaker of the Tajuru",
		b52TaboraxOracle:  "Taborax, Hope's Demise",
		b52TaekoOracle:    "Taeko, the Patient Avalanche",
		b52ThuridOracle:   "Thurid, Mare of Destiny",
		b52UrdnanOracle:   "Urdnan, Dromoka Warrior",
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

// --- Urdnan, Dromoka Warrior ------------------------------------------------

func b52Urdnan(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Urdnan, Dromoka Warrior", "Legendary Creature — Human Warrior", b52UrdnanOracle, 1, 1)
}

func TestBNSB52UrdnanEntersAndCountersATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	u := effectstest.CastAndResolveCreature(t, g, "Urdnan, Dromoka Warrior", "Legendary Creature — Human Warrior", b52UrdnanOracle)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", n)
	}
	if n := effectstest.CountersOn(g, u, game.CounterPlusOne); n != 0 {
		t.Errorf("Urdnan has %d counters, want 0", n)
	}
}

func TestBNSB52UrdnanGivesFirstStrikeToOneCounterAndDoubleToTwo(t *testing.T) {
	for _, counters := range []int{1, 2, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		u := b52Urdnan(g, me.ID)
		bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
		g.WithWriteLock(func() {
			c := effectstest.FindBattlefieldCardByID(g, bear)
			c.Counters = map[string]int{game.CounterPlusOne: counters}
		})
		effectstest.AdvanceToMain(t, g)
		effectstest.DeclareAttack(t, g, opp.ID, u, bear)
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		wantFS, wantDS := counters == 1, counters >= 2
		if got := effectstest.HasKeywordOnBattlefield(t, g, bear, "first strike"); got != wantFS {
			t.Errorf("%d counters: first strike = %v, want %v", counters, got, wantFS)
		}
		if got := effectstest.HasKeywordOnBattlefield(t, g, bear, "double strike"); got != wantDS {
			t.Errorf("%d counters: double strike = %v, want %v", counters, got, wantDS)
		}
	}
}

func TestBNSB52UrdnanNeedsAnAttackerWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	u := b52Urdnan(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, u)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TriggersOnStackFrom(g, u) != 0 {
		t.Error("the attack trigger went on the stack with no creature holding a +1/+1 counter")
	}
}

// --- Shroofus Sproutsire ----------------------------------------------------

func TestBNSB52ShroofusMakesAsManySaprolingsAsASaprolingDealt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shroofus := effectstest.B12Push(g, me.ID, "Shroofus Sproutsire", "Legendary Creature — Saproling", b52ShroofusOracle, 1, 1)
	effectstest.AssertKeywords(t, g, shroofus, "trample")
	sap := effectstest.B12Creature(g, me.ID, "Saproling Scout", "Creature — Saproling", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Saproling"); n != 0 {
		t.Fatalf("a Bear connected and made %d Saprolings", n)
	}
	effectstest.DealCombatDamageToPlayer(g, sap, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Saproling"); n != 3 {
		t.Errorf("a 3-damage Saproling made %d tokens, want 3", n)
	}
}

func TestBNSB52ShroofusIgnoresAnOpponentsSaprolingAndNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Shroofus Sproutsire", "Legendary Creature — Saproling", b52ShroofusOracle, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Saproling", "Creature — Saproling", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Saproling", "Creature — Saproling", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, theirs, me.ID, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: mine, Target: opp.ID, Amount: 2})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Saproling"); n != 0 {
		t.Errorf("made %d Saprolings from an opponent's creature or noncombat damage", n)
	}
}

// --- Thurid, Mare of Destiny ------------------------------------------------

func TestBNSB52ThuridAnthemsOtherPegasiUnicornsAndHorsesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thurid := effectstest.B12Push(g, me.ID, "Thurid, Mare of Destiny", "Legendary Creature — Pegasus", b52ThuridOracle, 2, 4)
	effectstest.AssertKeywords(t, g, thurid, "flying", "lifelink")
	horse := effectstest.B12Creature(g, me.ID, "Horse", "Creature — Horse", 2, 2)
	unicorn := effectstest.B12Creature(g, me.ID, "Unicorn", "Creature — Unicorn", 2, 2)
	pegasus := effectstest.B12Creature(g, me.ID, "Pegasus", "Creature — Pegasus", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Horse", "Creature — Horse", 2, 2)
	for id, want := range map[uuid.UUID]int{thurid: 2, horse: 3, unicorn: 3, pegasus: 3, bear: 2, theirs: 2} {
		if p := effectstest.EffectivePower(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
}

func TestBNSB52ThuridCopiesAHorseCreatureSpellButNotABear(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     int
	}{{"Creature — Bear", 1}, {"Creature — Horse", 2}, {"Creature — Unicorn", 2}, {"Creature — Pegasus", 2}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Thurid, Mare of Destiny", "Legendary Creature — Pegasus", b52ThuridOracle, 2, 4)
		effectstest.CastCatalogSpell(t, g, "Steed", tc.typeLine, "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CountBattlefieldNamed(g, me.ID, "Steed"); got != tc.want {
			t.Errorf("%s: %d Steeds on the battlefield, want %d", tc.typeLine, got, tc.want)
		}
	}
}

func TestBNSB52ThuridDoesNotCopyANoncreatureSpellOfTheType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Thurid, Mare of Destiny", "Legendary Creature — Pegasus", b52ThuridOracle, 2, 4)
	effectstest.CastCatalogSpell(t, g, "Horse Tribal", "Kindred Sorcery — Horse", "", nil)
	if n := g.Stack.Size(); n != 1 {
		t.Errorf("stack holds %d items after a Horse-typed noncreature spell, want just the spell", n)
	}
}

// --- Psemilla, Meletian Poet ------------------------------------------------

func b52Psemilla(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Psemilla, Meletian Poet", "Legendary Creature — Human Bard", b52PsemillaOracle, 1, 1)
}

func TestBNSB52PsemillaMakesANymphOnlyForTheFirstEnchantmentSpellEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	b52Psemilla(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Creature Spell", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Nymph"); n != 0 {
		t.Fatalf("a non-enchantment spell made %d Nymphs", n)
	}
	effectstest.CastCatalogSpell(t, g, "Aura One", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Nymph"); n != 1 {
		t.Fatalf("the first enchantment spell made %d Nymphs, want 1", n)
	}
	effectstest.CastCatalogSpell(t, g, "Aura Two", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Nymph"); n != 1 {
		t.Errorf("the second enchantment spell made another Nymph: %d tokens", n)
	}
}

func TestBNSB52PsemillaGrowsAtCombatOnlyWithFiveEnchantments(t *testing.T) {
	for _, enchantments := range []int{4, 5} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		p := b52Psemilla(g, me.ID)
		for i := 0; i < enchantments; i++ {
			effectstest.B12Permanent(g, me.ID, "Enchantment", "Enchantment")
		}
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		wantPower := 1
		if enchantments >= 5 {
			wantPower = 5
		}
		if got := effectstest.EffectivePower(t, g, p); got != wantPower {
			t.Errorf("%d enchantments: power %d, want %d", enchantments, got, wantPower)
		}
		if got := effectstest.HasKeywordOnBattlefield(t, g, p, "lifelink"); got != (enchantments >= 5) {
			t.Errorf("%d enchantments: lifelink = %v", enchantments, got)
		}
	}
}

// --- Qala, Ajani's Pridemate ------------------------------------------------

func b52Qala(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Qala, Ajani's Pridemate", "Legendary Creature — Cat Warrior", b52QalaOracle, 3, 3)
}

func TestBNSB52QalaGrowsOnEachLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	q := b52Qala(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(q, me.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, q, game.CounterPlusOne); n != 1 {
		t.Errorf("one 3-life gain put %d counters, want 1", n)
	}
}

func TestBNSB52QalaPumpsOtherAttackersByItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	q := b52Qala(g, me.ID)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardByID(g, q).Counters = map[string]int{game.CounterPlusOne: 2, "charge": 1}
	})
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	idle := effectstest.PushVanillaCreature(g, me.ID, "Idle", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, q, a)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, a); p != 5 {
		t.Errorf("the other attacker has power %d, want 2+3", p)
	}
	if p := effectstest.EffectivePower(t, g, idle); p != 2 {
		t.Errorf("a creature that did not attack has power %d, want 2", p)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, q); p != 5 {
		t.Errorf("Qala has power %d, want 3 + two +1/+1 counters (no boost for herself)", p)
	}
}

// --- Sutina, Speaker of the Tajuru ------------------------------------------

// answerPermanentsPick answers the seat's own-permanents prompt with ids.
func answerPermanentsPick(t *testing.T, g *game.Game, seat uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, seat)
	if pick == nil {
		t.Fatalf("no own-permanents prompt for %s", seat)
	}
	if err := g.ResolveOwnPermanents(pick.ID, seat, ids); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
}

func TestBNSB52SutinaFetchesATappedBasicOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	before := effectstest.CountBattlefieldNamed(g, me.ID, "Forest")
	effectstest.CastAndResolveCreature(t, g, "Sutina, Speaker of the Tajuru", "Legendary Creature — Elf Scout", b52SutinaOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Forest"); got != before+1 {
		t.Fatalf("Forests on the battlefield %d → %d, want one more", before, got)
	}
}

func TestBNSB52SutinaBouncesALandThenCountersATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s := effectstest.B12Push(g, me.ID, "Sutina, Speaker of the Tajuru", "Legendary Creature — Elf Scout", b52SutinaOracle, 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, s)
	effectstest.PassPriorityAroundTable(t, g)
	answerPermanentsPick(t, g, me.ID, land)
	effectstest.PickTriggerTarget(t, g, me.ID, s)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land stayed on the battlefield")
	}
	if n := effectstest.CountersOn(g, s, game.CounterPlusOne); n != 1 {
		t.Errorf("Sutina has %d +1/+1 counters, want 1", n)
	}
}

func TestBNSB52SutinaDecliningTheBounceMakesNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s := effectstest.B12Push(g, me.ID, "Sutina, Speaker of the Tajuru", "Legendary Creature — Elf Scout", b52SutinaOracle, 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, s)
	effectstest.PassPriorityAroundTable(t, g)
	answerPermanentsPick(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("the land left the battlefield though none was picked")
	}
	if n := effectstest.CountersOn(g, s, game.CounterPlusOne); n != 0 {
		t.Errorf("Sutina has %d counters after declining, want 0", n)
	}
}

// --- Taborax, Hope's Demise -------------------------------------------------

func b52Taborax(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Taborax, Hope's Demise", "Legendary Creature — Demon Cleric", b52TaboraxOracle, 2, 2)
}

func TestBNSB52TaboraxGetsLifelinkAtFiveCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tab := b52Taborax(g, me.ID)
	effectstest.AssertKeywords(t, g, tab, "flying")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(tab, game.CounterPlusOne, 4) })
	if effectstest.HasKeywordOnBattlefield(t, g, tab, "lifelink") {
		t.Error("lifelink at four counters")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(tab, game.CounterPlusOne, 1) })
	if !effectstest.HasKeywordOnBattlefield(t, g, tab, "lifelink") {
		t.Error("no lifelink at five counters")
	}
}

func b52Kill(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
}

func TestBNSB52TaboraxCountersForANontokenButOnlyDrawsForAClericOnYes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tab := b52Taborax(g, me.ID)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b52Kill(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, tab, game.CounterPlusOne); n != 1 {
		t.Fatalf("a Bear died and Taborax has %d counters, want 1", n)
	}
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Fatal("a non-Cleric offered a draw")
	}
	cleric := effectstest.B12Creature(g, me.ID, "Acolyte", "Creature — Human Cleric", 1, 1)
	hand, life := me.Hand.Size(), me.Life
	b52Kill(g, cleric)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, tab, game.CounterPlusOne); n != 2 {
		t.Errorf("Taborax has %d counters, want 2", n)
	}
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d→%d, life %d→%d: want one card drawn and 1 life lost", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestBNSB52TaboraxDecliningTheDrawCostsNoLifeAndTokensAndTaboraxItselfDontCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tab := b52Taborax(g, me.ID)
	cleric := effectstest.B12Creature(g, me.ID, "Acolyte", "Creature — Human Cleric", 1, 1)
	life := me.Life
	b52Kill(g, cleric)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("declining cost life: %d → %d", life, me.Life)
	}
	tok := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 green Saproling"))
	b52Kill(g, tok)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, tab, game.CounterPlusOne); n != 1 {
		t.Errorf("a token's death changed the counters to %d, want 1", n)
	}
}

// --- Pol Jamaar, Illusionist ------------------------------------------------

func TestBNSB52PolJamaarDrawsForEachCreatureOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Elf One", "Creature — Elf Warrior", 1, 1)
	effectstest.B12Creature(g, me.ID, "Elf Two", "Creature — Elf Druid", 1, 1)
	effectstest.B12Creature(g, me.ID, "Elf Three", "Creature — Elf", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	hand := me.Hand.Size()
	effectstest.CastAndPause(t, g, "Pol Jamaar, Illusionist", "Legendary Creature — Human Illusion Wizard", b52PolOracle)
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards for three Elves you control, want 3", got)
	}
}

func TestBNSB52PolJamaarCountsItselfAndIgnoresOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, opp.ID, "Their Wizard", "Creature — Wizard", 1, 1)
	hand := me.Hand.Size()
	effectstest.CastAndPause(t, g, "Pol Jamaar, Illusionist", "Legendary Creature — Human Illusion Wizard", b52PolOracle)
	effectstest.AnswerCreatureType(t, g, me.ID, "Wizard")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1 (Pol Jamaar is the only Wizard you control)", got)
	}
}

// --- Rev, Tithe Extractor ---------------------------------------------------

func b52Rev(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Rev, Tithe Extractor", "Legendary Creature — Human Rogue", b52RevOracle, 3, 3)
}

func TestBNSB52RevGivesATargetDeathtouchWhenYouAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rev := b52Rev(g, me.ID)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, rev, bear)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "deathtouch") {
		t.Error("the targeted creature has no deathtouch")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, rev, "deathtouch") {
		t.Error("Rev has deathtouch though he was not the target")
	}
}

func TestBNSB52RevTreasureAndFaceDownCastableExileOncePerPlayerHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rev := b52Rev(g, me.ID)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	top := effectstest.SeedTopOfLibrary(opp, game.Card{Name: "Their Top Card", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: opp.ID, Controller: opp.ID})
	libBefore := opp.Library.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, rev, opp.ID, 3)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures for two creatures hitting one player, want 1", n)
	}
	if opp.Library.Size() != libBefore-1 {
		t.Errorf("library %d → %d, want exactly one card exiled", libBefore, opp.Library.Size())
	}
	if effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneExile) == nil {
		t.Error("the exiled card cannot be cast by Rev's controller")
	}
	if effectstest.GrantedPermissionOn(g, opp.ID, top, game.ZoneExile) != nil {
		t.Error("the card's owner was granted a permission too")
	}
}

// --- Slinza, the Spiked Stampede --------------------------------------------

func b52Slinza(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Slinza, the Spiked Stampede", "Legendary Creature — Beast", b52SlinzaOracle, 5, 5)
}

// b52Enter casts a free creature card of the given printing and resolves
// just the spell, leaving what it triggered pending.
func b52Enter(t *testing.T, g *game.Game, name, typeLine string, power int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: power, Toughness: power,
		Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestBNSB52SlinzaPutsACounterOnOtherBeastsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slinza := b52Slinza(g, me.ID)
	beast := b52Enter(t, g, "Boar", "Creature — Beast", 1)
	if n := effectstest.CountersOn(g, beast, game.CounterPlusOne); n != 1 {
		t.Errorf("an entering Beast has %d counters, want 1", n)
	}
	bear := b52Enter(t, g, "Cub", "Creature — Bear", 1)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("a Bear entered with %d counters, want 0", n)
	}
	if n := effectstest.CountersOn(g, slinza, game.CounterPlusOne); n != 0 {
		t.Errorf("Slinza has %d counters, want 0", n)
	}
}

func TestBNSB52SlinzaTakesTwoOffBeastSpellsOnly(t *testing.T) {
	for _, tc := range []struct {
		typeLine, cost string
		lands          int
		ok             bool
	}{
		{"Creature — Beast", "{2}{R}", 1, true},
		{"Creature — Beast", "{3}{R}", 1, false},
		{"Creature — Beast", "{1}{R}", 1, true}, // generic only: {R} is still owed
		{"Creature — Bear", "{2}{R}", 1, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.B27Push(g, me.ID, "Slinza, the Spiked Stampede", "Legendary Creature — Beast", b52SlinzaOracle, "{4}{G}", 5, 5, "G")
		_, err := effectstest.TrekB37CastFromHand(t, g, game.Card{Name: "Spell", TypeLine: tc.typeLine, ManaCost: tc.cost, Power: 1, Toughness: 1}, tc.lands)
		if (err == nil) != tc.ok {
			t.Errorf("%s %s with %d Mountains: cast error %v, want ok=%v", tc.typeLine, tc.cost, tc.lands, err, tc.ok)
		}
	}
}

func TestBNSB52SlinzaFightsWhenAPowerFourCreatureEntersAndYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slinza := b52Slinza(g, me.ID)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.StdB23Mana(t, g, me, "{G}{G}")
	b52Enter(t, g, "Giant", "Creature — Giant", 4)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt when a 4-power creature entered")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the opposing creature survived a 5-damage fight")
	}
	if !g.Battlefield.Contains(slinza) {
		t.Error("Slinza died fighting a 2/3")
	}
}

func TestBNSB52SlinzaDoesNotTriggerForPowerThreeOrDeclinedPayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b52Slinza(g, me.ID)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.StdB23Mana(t, g, me, "{G}{G}")
	b52Enter(t, g, "Small", "Creature — Giant", 3)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a power-3 creature offered the payment")
	}
	b52Enter(t, g, "Giant", "Creature — Giant", 4)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the opposing creature died though the payment was declined")
	}
}

func TestBNSB52SlinzaEnteringOffersTheFightEvenAtLowPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, g.Seats[1].ID, "Victim", "Creature — Bear", 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.StdB23Mana(t, g, me, "{G}{G}")
	effectstest.CastAndResolveCreature(t, g, "Slinza, the Spiked Stampede", "Legendary Creature — Beast", b52SlinzaOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("Slinza's own entry offered no payment")
	}
}

// --- Taeko, the Patient Avalanche -------------------------------------------

func b52Unblockable(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return effectstest.EffectiveOf(t, g, id).Restrictions&game.CantBeBlocked != 0
}

func b52Taeko(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Taeko, the Patient Avalanche", "Legendary Creature — Turtle Ninja", b52TaekoOracle, 4, 5)
}

func TestBNSB52TaekoEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CastAndResolveCreature(t, g, "Taeko, the Patient Avalanche", "Legendary Creature — Turtle Ninja", b52TaekoOracle)
	c := effectstest.FindBattlefieldCardByID(g, id)
	if c == nil || !c.Tapped {
		t.Errorf("Taeko did not enter tapped: %+v", c)
	}
	_ = me
}

func TestBNSB52TaekoGrowsOnlyWhenAnotherCreatureLeavesWithoutDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	taeko := b52Taeko(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	bounced := effectstest.B12Creature(g, me.ID, "Bounced", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bounced) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, taeko, game.CounterPlusOne); n != 1 {
		t.Fatalf("a bounced creature gave %d counters, want 1", n)
	}
	exiled := effectstest.B12Creature(g, me.ID, "Exiled", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(exiled) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, taeko, game.CounterPlusOne); n != 2 {
		t.Fatalf("an exiled creature: %d counters, want 2", n)
	}
	died := effectstest.B12Creature(g, me.ID, "Died", "Creature — Bear", 1, 1)
	effectstest.B27Kill(g, died)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, taeko, game.CounterPlusOne); n != 2 {
		t.Errorf("a creature that died changed the counters to %d, want 2", n)
	}
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, taeko, game.CounterPlusOne); n != 2 {
		t.Errorf("an opponent's creature leaving changed the counters to %d, want 2", n)
	}
}

func TestBNSB52TaekoPaidAttackMakesAnAttackerUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	taeko := b52Taeko(g, me.ID)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, taeko, bear)
	effectstest.StdB23Mana(t, g, me, "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !b52Unblockable(t, g, bear) {
		t.Error("the targeted attacker can still be blocked")
	}
	if b52Unblockable(t, g, taeko) {
		t.Error("Taeko became unblockable though she was not the target")
	}
}

func TestBNSB52TaekoDecliningTheAttackPaymentChangesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	taeko := b52Taeko(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, taeko)
	effectstest.StdB23Mana(t, g, me, "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if b52Unblockable(t, g, taeko) {
		t.Error("Taeko became unblockable though the payment was declined")
	}
}
