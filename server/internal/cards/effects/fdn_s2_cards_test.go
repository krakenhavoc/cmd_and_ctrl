package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_s2_cards_test.go — Foundations, S batch 2 (#4465): the S cards the
// first S batch left as probably seam-free. Each test plays the real
// catalog entry through the engine and asserts what its text says.

const (
	fdnS2CerberusOracle  = "39cae244-96e1-4ea3-899a-3aa871cca31c"
	fdnS2SanctuaryOracle = "3c0f99b8-0222-4fac-932a-eb5d77826564"
	fdnS2SphinxOracle    = "ddb7e868-8f1d-4d57-9ee2-ebadb17b398a"
	fdnS2SpinnerOracle   = "6883e79e-0b4f-4fb8-aee3-a3905246a408"
	fdnS2RallierOracle   = "e0b15374-b4f6-48e0-8477-4eb5d5c2a938"
	fdnS2ThiefOracle     = "c85987b4-090e-472e-b9af-1114cce56ffb"
	fdnS2NobleOracle     = "c55909d3-1f67-4a4a-9d53-8513d6cf96d8"
	fdnS2HealerOracle    = "1729c6f2-2452-4fbc-a1bc-3e625f2ad8e1"
	fdnS2SurrakOracle    = "fa56a5fa-ef96-404c-8fb6-d1f5fcebb52e"
	fdnS2ShamblerOracle  = "ea6ee33f-0cba-4e96-817f-dc89c9064ead"
	fdnS2ScavengeOracle  = "398afdf2-4a12-430e-85c9-c68847230939"
	fdnS2SyrAlinOracle   = "5d9a4905-4a26-410e-936c-56de849e9ca3"
)

func TestFdnS2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnS2CerberusOracle:  "Slumbering Cerberus",
		fdnS2SanctuaryOracle: "Soulstone Sanctuary",
		fdnS2SphinxOracle:    "Sphinx of Forgotten Lore",
		fdnS2SpinnerOracle:   "Spinner of Souls",
		fdnS2RallierOracle:   "Squad Rallier",
		fdnS2ThiefOracle:     "Stromkirk Bloodthief",
		fdnS2NobleOracle:     "Stromkirk Noble",
		fdnS2HealerOracle:    "Sun-Blessed Healer",
		fdnS2SurrakOracle:    "Surrak, the Hunt Caller",
		fdnS2ShamblerOracle:  "Suspicious Shambler",
		fdnS2ScavengeOracle:  "Sylvan Scavenging",
		fdnS2SyrAlinOracle:   "Syr Alin, the Lion's Claw",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestFdnS2PrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Sphinx of Forgotten Lore", "Creature — Sphinx", fdnS2SphinxOracle, 3, 3, []string{"flash", "flying"}},
		{"Spinner of Souls", "Creature — Spider Spirit", fdnS2SpinnerOracle, 4, 3, []string{"reach"}},
		{"Sun-Blessed Healer", "Creature — Human Cleric", fdnS2HealerOracle, 3, 1, []string{"lifelink"}},
		{"Syr Alin, the Lion's Claw", "Legendary Creature — Human Knight", fdnS2SyrAlinOracle, 4, 4, []string{"first strike"}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
		assertKeywords(t, g, id, c.keywords...)
	}
}

// fdnS2Tap taps a permanent on the battlefield.
func fdnS2Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { b16Tap(g, id) })
}

// --- Slumbering Cerberus ---------------------------------------------------

func TestFdnS2CerberusStaysTappedThroughYourUntapStep(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	dog := b12Push(g, me.ID, "Slumbering Cerberus", "Creature — Dog", fdnS2CerberusOracle, 4, 2)
	fdnS2Tap(g, dog)
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	advanceToMainOf(t, g, seat)
	if !b16Tapped(t, g, dog) {
		t.Error("it untapped during its controller's untap step")
	}
}

func TestFdnS2CerberusUntapsAtEachEndStepOnlyAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		dog := b12Push(g, me.ID, "Slumbering Cerberus", "Creature — Dog", fdnS2CerberusOracle, 4, 2)
		bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		fdnS2Tap(g, dog)
		if died {
			killCreature(t, g, opp.ID, bear)
		}
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		if tapped := b16Tapped(t, g, dog); tapped == died {
			t.Errorf("a creature died %v: tapped %v at the end step", died, tapped)
		}
	}
}

// --- Soulstone Sanctuary ---------------------------------------------------

func TestFdnS2SanctuaryBecomesAThreeThreeVigilantCreatureForGood(t *testing.T) {
	g, me, _ := spendTable(t)
	land := pushCatalogPermanent(g, me.ID, "Soulstone Sanctuary", "Land", fdnS2SanctuaryOracle, false)
	apaMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{4} paid with three mana")
	}
	apaMana(me, "C")
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	check := func(when string) {
		c := effectiveCardForTest(t, g, land)
		if !c.IsCreature() || !c.IsLand() {
			t.Errorf("%s: creature %v, land %v, want both", when, c.IsCreature(), c.IsLand())
		}
		for _, sub := range []string{"Elf", "Human", "Vampire"} {
			if !c.HasSubtype(sub) {
				t.Errorf("%s: the animated land is not a %s", when, sub)
			}
		}
		if p, tough := effectivePower(t, g, land), effectiveToughness(t, g, land); p != 3 || tough != 3 {
			t.Errorf("%s: P/T %d/%d, want 3/3", when, p, tough)
		}
		if !hasEffectiveKeyword(t, g, land, "vigilance") {
			t.Errorf("%s: no vigilance", when)
		}
		if len(c.Colors) != 0 {
			t.Errorf("%s: colours %v, want colourless", when, c.Colors)
		}
	}
	check("as it resolves")
	advanceToMainOf(t, g, 1)
	check("a turn later")
}

// --- Sphinx of Forgotten Lore ----------------------------------------------

func TestFdnS2SphinxGrantsFlashbackWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bolt := seedGraveyardCard(t, g, "Filler Bolt", "Instant", "test-fdn-s2-bolt")
	sphinx := b12Push(g, me.ID, "Sphinx of Forgotten Lore", "Creature — Sphinx", fdnS2SphinxOracle, 3, 3)
	advanceToMain(t, g)
	if perm := grantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard); perm.Granted() {
		t.Fatal("the card had flashback before the Sphinx attacked")
	}
	declareAttack(t, g, opp.ID, sphinx)
	pickTriggerTarget(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	perm := grantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard)
	if !perm.Granted() || perm.AltCostKey != "flashback" || !perm.ExileOnResolution || perm.Cost != "" {
		t.Errorf("granted %v key %q exile %v cost %q, want flashback at its mana cost", perm.Granted(), perm.AltCostKey, perm.ExileOnResolution, perm.Cost)
	}
}

// --- Spinner of Souls ------------------------------------------------------

func TestFdnS2SpinnerRevealsUntilACreatureWhenANontokenCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Spinner of Souls", "Creature — Spider Spirit", fdnS2SpinnerOracle, 4, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// Top first: Land, Elf, Other Land.
	land, elf := uuid.New(), uuid.New()
	for _, c := range []game.Card{
		{InstanceID: uuid.New(), Name: "Other Land", TypeLine: "Land"},
		{InstanceID: elf, Name: "Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1},
		{InstanceID: land, Name: "Land", TypeLine: "Land"},
	} {
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
	}
	library := me.Library.Size()

	tok := TokenCard("2/2 black Zombie")
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = me.ID, me.ID
	killCreature(t, g, me.ID, pushBattlefieldCardWithTimestamp(g, tok))
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || me.Hand.Contains(elf) {
		t.Fatal("a token dying triggered the Spinner")
	}

	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(elf) {
		t.Fatal("the revealed creature card is not in hand")
	}
	if me.Library.Size() != library-1 {
		t.Errorf("library %d -> %d, want one card fewer", library, me.Library.Size())
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top == land {
		t.Error("the revealed land is still on top; the rest go to the bottom")
	}
}

// --- Squad Rallier ---------------------------------------------------------

func TestFdnS2RallierTakesACreatureWithPowerTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rallier := pushCatalogPermanent(g, me.ID, "Squad Rallier", "Creature — Human Scout", fdnS2RallierOracle, true)
	// Top first: Giant, Scout, Spell, Land.
	big, small := uuid.New(), uuid.New()
	for _, c := range []game.Card{
		{InstanceID: uuid.New(), Name: "Land", TypeLine: "Land"},
		{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery"},
		{InstanceID: small, Name: "Scout", TypeLine: "Creature — Human", Power: 2, Toughness: 1},
		{InstanceID: big, Name: "Giant", TypeLine: "Creature — Giant", Power: 3, Toughness: 3},
	} {
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
	}
	library := me.Library.Size()
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, rallier, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{2}{W} paid with {C}{C}")
	}
	floatForTest(g, me, "W")
	if err := g.ActivateCatalogAbility(me.ID, rallier, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{2}{W} while summoning sick: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick")
	}
	for _, id := range pick.ChooseCards {
		if id == big {
			t.Error("the power-3 creature was offered")
		}
	}
	answerChooseCards(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(small) {
		t.Error("the power-2 creature is not in hand")
	}
	if me.Library.Size() != library-1 {
		t.Errorf("library %d -> %d, want one card fewer", library, me.Library.Size())
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top == big {
		t.Error("the Giant is still on top; the rest go to the bottom")
	}
}

// --- Stromkirk Bloodthief --------------------------------------------------

func TestFdnS2BloodthiefCountersAVampireOnlyAfterAnOpponentLostLife(t *testing.T) {
	for _, lost := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		thief := b12Push(g, me.ID, "Stromkirk Bloodthief", "Creature — Vampire Rogue", fdnS2ThiefOracle, 2, 2)
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		if lost {
			g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
		}
		advanceTo(t, g, game.StepEnd)
		if !lost {
			passPriorityAroundTable(t, g)
			if latestPickTarget(g, me.ID) != nil || countersOn(g, thief, game.CounterPlusOne) != 0 {
				t.Error("it triggered with no opponent having lost life")
			}
			continue
		}
		pick := latestPickTarget(g, me.ID)
		if pick == nil {
			t.Fatal("no target prompt after an opponent lost life")
		}
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Fatal("a non-Vampire was accepted as the target")
		}
		pickTriggerTarget(t, g, me.ID, thief)
		passPriorityAroundTable(t, g)
		if n := countersOn(g, thief, game.CounterPlusOne); n != 1 {
			t.Errorf("%d counters on the Vampire, want 1", n)
		}
	}
}

// --- Stromkirk Noble -------------------------------------------------------

func TestFdnS2NobleCantBeBlockedByHumans(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	noble := b12Push(g, me.ID, "Stromkirk Noble", "Creature — Vampire Noble", fdnS2NobleOracle, 1, 1)
	human := b12Creature(g, opp.ID, "Their Human", "Creature — Human Soldier", 2, 2)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(noble, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(human, noble); err == nil {
		t.Error("a Human blocked Stromkirk Noble")
	}
	if err := g.DeclareBlocker(bear, noble); err != nil {
		t.Errorf("a non-Human could not block: %v", err)
	}
}

func TestFdnS2NobleGrowsWhenItDealsCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	noble := b12Push(g, me.ID, "Stromkirk Noble", "Creature — Vampire Noble", fdnS2NobleOracle, 1, 1)
	attackWith(t, g, opp.ID, noble)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, noble, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters after combat damage to a player, want 1", n)
	}
}

// --- Sun-Blessed Healer ----------------------------------------------------

func TestFdnS2HealerReturnsASmallPermanentOnlyWhenKicked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	stone := uuid.New()
	big := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: stone, Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big Relic", TypeLine: "Artifact", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})

	if _, err := castWithOptionalCosts(t, g, "Sun-Blessed Healer", "Creature — Human Cleric", fdnS2HealerOracle, nil, nil, nil); err != nil {
		t.Fatalf("unkicked CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || len(g.PendingTriggers) != 0 {
		t.Fatal("an unkicked Healer triggered")
	}

	if _, err := castWithOptionalCosts(t, g, "Sun-Blessed Healer", "Creature — Human Cleric", fdnS2HealerOracle, nil, []int{0}, nil); err != nil {
		t.Fatalf("kicked CastSpell: %v", err)
	}
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Fatal("a mana value 3 card was accepted as the target")
	}
	pickCard(t, g, me.ID, stone)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, stone); !ok || c.Controller != me.ID {
		t.Error("the kicked Healer did not return the mana value 2 permanent card to the battlefield")
	}
}

// --- Surrak, the Hunt Caller -----------------------------------------------

func TestFdnS2SurrakGivesHasteOnlyWithEightPower(t *testing.T) {
	for _, formidable := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		b12Push(g, me.ID, "Surrak, the Hunt Caller", "Legendary Creature — Human Warrior", fdnS2SurrakOracle, 5, 4)
		power := 2
		if formidable {
			power = 3
		}
		fresh := b12Creature(g, me.ID, "Bear", "Creature — Bear", power, power)
		advanceTo(t, g, game.StepBeginCombat)
		if !formidable {
			passPriorityAroundTable(t, g)
			if latestPickTarget(g, me.ID) != nil || hasEffectiveKeyword(t, g, fresh, "haste") {
				t.Error("formidable triggered with total power 7")
			}
			continue
		}
		pickTriggerTarget(t, g, me.ID, fresh)
		passPriorityAroundTable(t, g)
		if !hasEffectiveKeyword(t, g, fresh, "haste") {
			t.Error("the target did not gain haste with total power 8")
		}
	}
}

// --- Suspicious Shambler ---------------------------------------------------

func TestFdnS2ShamblerMakesTwoZombiesFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain {
		t.Fatalf("fixture starts in a main phase (%s): the test wants an instant-speed window", g.Turn.Step)
	}
	id := pushCatalogGraveyardCard(me, "Suspicious Shambler", "Creature — Zombie", fdnS2ShamblerOracle, 4, 2)
	floatForTest(g, me, "CCCCBB")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated outside a main phase: it is sorcery speed")
	}
	advanceToMain(t, g)
	floatForTest(g, me, "CCCCBB")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{4}{B}{B}, exile this card from your graveyard (main phase): %v", err)
	}
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("the card is not exiled as a cost")
	}
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Zombie"); n != 2 {
		t.Errorf("%d Zombies, want 2", n)
	}
}

// --- Sylvan Scavenging -----------------------------------------------------

func TestFdnS2ScavengingCounterMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Sylvan Scavenging", "Enchantment", fdnS2ScavengeOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt at the end step")
	}
	answerMode(t, g, c, me.ID, 0)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
}

func TestFdnS2ScavengingRaccoonNeedsPowerFour(t *testing.T) {
	for _, big := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		pushCatalogPermanent(g, me.ID, "Sylvan Scavenging", "Enchantment", fdnS2ScavengeOracle, false)
		power := 3
		if big {
			power = 4
		}
		b12Creature(g, me.ID, "Beast", "Creature — Beast", power, power)
		advanceTo(t, g, game.StepEnd)
		c := modePickChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no mode prompt at the end step")
		}
		answerMode(t, g, c, me.ID, 1)
		passPriorityAroundTable(t, g)
		want := 0
		if big {
			want = 1
		}
		if n := countBattlefieldByName(g, "Raccoon"); n != want {
			t.Errorf("power %d: %d Raccoons, want %d", power, n, want)
		}
	}
}

// --- Syr Alin, the Lion's Claw ---------------------------------------------

func TestFdnS2SyrAlinPumpsTheOthersWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	alin := b12Push(g, me.ID, "Syr Alin, the Lion's Claw", "Legendary Creature — Human Knight", fdnS2SyrAlinOracle, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, alin)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the other creature is %d/%d, want 3/3", p, tough)
	}
	if p := effectivePower(t, g, alin); p != 4 {
		t.Errorf("Syr Alin has power %d, want 4: it pumps the others", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}
