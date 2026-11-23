package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b01_cards_test.go — S82 reprint pass, batch 1 (#5908): Aegar, the
// Freezing Flame; Anafenza, Kin-Tree Spirit; Atraxa, Praetors' Voice; Baral,
// Chief of Compliance; Ezuri, Claw of Progress; Fairgrounds Trumpeter;
// Firesong and Sunspeaker; Firja, Judge of Valor; Grimgrin, Corpse-Born;
// Gyruda, Doom of Depths; Horobi, Death's Wail.

const (
	s82rAegarOracle     = "592d847e-e3f3-425d-89f9-01ab533ad581"
	s82rAnafenzaOracle  = "f7fa6ac0-70db-4586-b881-5a19db644dfb"
	s82rAtraxaOracle    = "7e6b9b59-cd68-4e3c-827b-38833c92d6eb"
	s82rBaralOracle     = "baf80db7-2ca8-4bfc-b5b6-25ed51878dd7"
	s82rEzuriOracle     = "40eacdae-fae5-46b9-a9e5-9211469f74fc"
	s82rTrumpeterOracle = "3dfc6a52-d58c-4455-96fe-ec193743d67a"
	s82rFiresongOracle  = "fffaa634-42a0-4038-b027-24f28754fec7"
	s82rFirjaOracle     = "7d5dcf8c-ecda-49ae-980c-c8c9016252d9"
	s82rGrimgrinOracle  = "a2aa2c8f-7058-4020-9881-663b104ea051"
	s82rGyrudaOracle    = "25eaa977-ca85-4f8d-8be5-9297a3edb14f"
	s82rHorobiOracle    = "44e8d51e-0885-49df-a77c-cf4a88e41bc1"
)

func TestS82RprB01CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82rAegarOracle:     "Aegar, the Freezing Flame",
		s82rAnafenzaOracle:  "Anafenza, Kin-Tree Spirit",
		s82rAtraxaOracle:    "Atraxa, Praetors' Voice",
		s82rBaralOracle:     "Baral, Chief of Compliance",
		s82rEzuriOracle:     "Ezuri, Claw of Progress",
		s82rTrumpeterOracle: "Fairgrounds Trumpeter",
		s82rFiresongOracle:  "Firesong and Sunspeaker",
		s82rFirjaOracle:     "Firja, Judge of Valor",
		s82rGrimgrinOracle:  "Grimgrin, Corpse-Born",
		s82rGyrudaOracle:    "Gyruda, Doom of Depths",
		s82rHorobiOracle:    "Horobi, Death's Wail",
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

// s82rCounters is how many +1/+1 counters a card carries.
func s82rCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.B12Card(t, g, id).Counters[game.CounterPlusOne]
}

// s82rEnter casts a vanilla creature for p from hand, resolving it and any
// trigger it causes.
func s82rEnter(t *testing.T, g *game.Game, p *game.Player, name, typeLine string, power, toughness int) uuid.UUID {
	t.Helper()
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID}
	effectstest.ToMain(t, g)
	p.Hand.PushTop(c)
	if err := g.CastSpell(p.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return c.InstanceID
}

// --- Anafenza, Kin-Tree Spirit ------------------------------------------------

func TestS82RprAnafenzaBolstersWhenAnotherNontokenCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	anafenza := effectstest.B12Push(g, me.ID, "Anafenza, Kin-Tree Spirit", "Legendary Creature — Spirit Soldier", s82rAnafenzaOracle, 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	small := effectstest.B12Creature(g, me.ID, "Small Bear", "Creature — Bear", 1, 1)
	_ = anafenza
	_ = big
	s82rEnter(t, g, me, "Mid Bear", "Creature — Bear", 3, 3)
	// The least toughness among creatures you control is the 1/1 Small Bear.
	if got := s82rCounters(t, g, small); got != 1 {
		t.Errorf("Small Bear has %d +1/+1 counters, want 1 (bolster 1 goes to the least toughness)", got)
	}
}

func TestS82RprAnafenzaIgnoresTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	anafenza := effectstest.B12Push(g, me.ID, "Anafenza, Kin-Tree Spirit", "Legendary Creature — Spirit Soldier", s82rAnafenzaOracle, 2, 2)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, game.Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1}, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.StackFullyEmpty(g) {
		t.Error("a token entering triggered Anafenza")
	}
	if got := s82rCounters(t, g, anafenza); got != 0 {
		t.Errorf("Anafenza has %d counters", got)
	}
}

func TestS82RprAnafenzaDoesNotTriggerForItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Lone Bear", "Creature — Bear", 2, 2)
	effectstest.ToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Anafenza, Kin-Tree Spirit", "Legendary Creature — Spirit Soldier", s82rAnafenzaOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82rCounters(t, g, bear) + s82rCounters(t, g, id); got != 0 {
		t.Errorf("Anafenza entering bolstered (%d counters)", got)
	}
}

// --- Atraxa, Praetors' Voice -----------------------------------------------

func TestS82RprAtraxaProliferatesAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	atraxa := effectstest.B12Push(g, me.ID, "Atraxa, Praetors' Voice", "Legendary Creature — Phyrexian Angel Horror", s82rAtraxaOracle, 4, 4)
	for _, kw := range []string{"flying", "vigilance", "deathtouch", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, atraxa, kw) {
			t.Errorf("Atraxa lacks %s", kw)
		}
	}
	bear := effectstest.B12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(bear, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ProliferatePrompt(g) == nil {
		t.Fatal("Atraxa's end step trigger asked nothing")
	}
	effectstest.AnswerProliferate(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82rCounters(t, g, bear); got != 2 {
		t.Errorf("the Bear has %d +1/+1 counters after Atraxa's proliferate, want 2", got)
	}
}

// --- Fairgrounds Trumpeter ----------------------------------------------------

func TestS82RprTrumpeterGrowsAtAnEndStepAfterACounterWasPutOnYourPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trumpeter := effectstest.B12Push(g, me.ID, "Fairgrounds Trumpeter", "Creature — Elephant", s82rTrumpeterOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AddCountersForTest(g, bear, game.CounterPlusOne, 1)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82rCounters(t, g, trumpeter); got != 1 {
		t.Errorf("Trumpeter has %d +1/+1 counters, want 1", got)
	}
}

func TestS82RprTrumpeterDoesNothingWithoutACounterThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%2]
	trumpeter := effectstest.B12Push(g, me.ID, "Fairgrounds Trumpeter", "Creature — Elephant", s82rTrumpeterOracle, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// A counter on an opponent's permanent, and one on yours from a past
	// turn's worth of state, do not count: only this turn, only your own.
	effectstest.AddCountersForTest(g, theirs, game.CounterPlusOne, 1)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82rCounters(t, g, trumpeter); got != 0 {
		t.Errorf("Trumpeter has %d counters, want 0", got)
	}
}

// --- Firja, Judge of Valor -----------------------------------------------------

func TestS82RprFirjaLooksAtThreeOnTheSecondSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Firja, Judge of Valor", "Legendary Creature — Angel Cleric", s82rFirjaOracle, 2, 4)
	effectstest.ToMain(t, g)
	me.Library.Cards = nil
	var ids []uuid.UUID
	for _, n := range []string{"Deep", "Third", "Second", "First"} {
		ids = append(ids, effectstest.SeedLibraryTop(me, n, "Sorcery", ""))
	}
	// ids[3] "First" is on top, then "Second", "Third", "Deep".
	effectstest.CastFromHandForTest(t, g, me, "Spell One", "Sorcery", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Fatal("the first spell triggered Firja")
	}
	effectstest.CastFromHandForTest(t, g, me, "Spell Two", "Sorcery", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pc := effectstest.LatestChooseCards(g, me.ID)
	if pc == nil {
		t.Fatal("the second spell raised no pick")
	}
	if len(pc.ChooseCards) != 3 || pc.ChooseMin != 1 || pc.ChooseMax != 1 {
		t.Fatalf("offered %v [%d,%d], want the top three, exactly one", pc.ChooseCards, pc.ChooseMin, pc.ChooseMax)
	}
	if err := g.ResolveChooseCards(pc.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[2]) {
		t.Error("the pick is not in hand")
	}
	for _, rest := range []uuid.UUID{ids[3], ids[1]} {
		if !me.Graveyard.Contains(rest) {
			t.Errorf("a card not picked is not in the graveyard")
		}
	}
	if !me.Library.Contains(ids[0]) || me.Graveyard.Contains(ids[0]) {
		t.Error("the fourth card was touched")
	}
	effectstest.CastFromHandForTest(t, g, me, "Spell Three", "Sorcery", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("the third spell triggered Firja again")
	}
}

// --- Horobi, Death's Wail ------------------------------------------------------

func TestS82RprHorobiDestroysAnyCreatureThatBecomesATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%2]
	effectstest.B12Push(g, me.ID, "Horobi, Death's Wail", "Legendary Creature — Spirit", s82rHorobiOracle, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	effectstest.ToMain(t, g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Fatalf("a spell that targets a player was affected: life %d", opp.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("an opponent's creature that became a target survived Horobi (a 5/5 Bolted for 3)")
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("your own targeted creature survived Horobi")
	}
}

func TestS82RprHorobiDestroysItselfWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	horobi := effectstest.B12Push(g, me.ID, "Horobi, Death's Wail", "Legendary Creature — Spirit", s82rHorobiOracle, 4, 4)
	effectstest.ToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(horobi))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(horobi) {
		t.Error("Horobi survived being targeted")
	}
}

// --- Baral, Chief of Compliance --------------------------------------------------

func TestS82RprBaralReducesInstantsAndSorceriesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Baral, Chief of Compliance", "Legendary Creature — Human Wizard", s82rBaralOracle, 1, 3)
	effectstest.ToMain(t, g)
	bolt := game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{1}{R}", Colors: []string{"R"},
		OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(bolt)
	effectstest.PayMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, bolt.InstanceID, game.CastSpellParams{Strict: true, Targets: effectstest.B16TargetPlayer(g.Seats[1].ID)}); err != nil {
		t.Fatalf("a {1}{R} instant with one {R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	bear := game.Card{InstanceID: uuid.New(), Name: "Costly Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(bear)
	effectstest.PayMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature spell got the discount")
	}
}

func TestS82RprBaralLootsWhenYourCounterspellCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Baral, Chief of Compliance", "Legendary Creature — Human Wizard", s82rBaralOracle, 1, 3)
	effectstest.Msc18Library(me, 8)
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetPlayer(me.ID))
	effectstest.CastCatalogSpell(t, g, "Counterspell", "Instant", "cc187110-1148-4090-bbb8-e205694a39f5", effectstest.CardRefs(bolt))
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 40 {
		t.Fatalf("the Bolt resolved: life %d", me.Life)
	}
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no \"draw a card?\" prompt after the counter")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after drawing, want %d", me.Hand.Size(), hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
}

func TestS82RprBaralDoesNotLootWhenAnOpponentCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Baral, Chief of Compliance", "Legendary Creature — Human Wizard", s82rBaralOracle, 1, 3)
	effectstest.ToMain(t, g)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	oppCounter := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: oppCounter, Name: "Counterspell", TypeLine: "Instant", OracleID: "cc187110-1148-4090-bbb8-e205694a39f5", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, oppCounter, game.CastSpellParams{Targets: effectstest.CardRefs(bolt)}); err != nil {
		t.Fatalf("opponent's Counterspell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm); ask != nil {
		t.Error("Baral offered a loot for an opponent's counterspell")
	}
}

// --- Ezuri, Claw of Progress -------------------------------------------------

func TestS82RprEzuriCountsPowerTwoOrLessAndGrowsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Ezuri, Claw of Progress", "Legendary Creature — Phyrexian Elf Warrior", s82rEzuriOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Ally Bear", "Creature — Bear", 2, 2)
	s82rEnter(t, g, me, "Small One", "Creature — Bear", 2, 2)
	s82rEnter(t, g, me, "Big One", "Creature — Bear", 3, 3)
	if got := effects.ExperienceCounters(g, me.ID); got != 1 {
		t.Fatalf("experience counters %d, want 1 (only the power 2 creature counted)", got)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	total := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name != "Ezuri, Claw of Progress" {
			total += c.Counters[game.CounterPlusOne]
		}
	}
	if total != 1 {
		t.Errorf("%d +1/+1 counters were put on other creatures, want X = 1", total)
	}
	ezuri := effectstest.FindBattlefieldByName(g, "Ezuri, Claw of Progress")
	if s82rCounters(t, g, ezuri) != 0 {
		t.Error("Ezuri put the counters on itself")
	}
}

// --- Aegar, the Freezing Flame -------------------------------------------------

func s82rAegarTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%2]
	aegar := effectstest.B12Push(g, me.ID, "Aegar, the Freezing Flame", "Legendary Creature — Giant Wizard", s82rAegarOracle, 3, 3)
	effectstest.ToMain(t, g)
	return g, me, opp, aegar
}

func TestS82RprAegarDrawsWhenYourSpellDealsExcessDamage(t *testing.T) {
	g, me, opp, _ := s82rAegarTable(t)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(wall))
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Fatal("the Wall survived")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1", got)
	}
}

func TestS82RprAegarDrawsWhenAGiantOrWizardDealsExcessDamage(t *testing.T) {
	g, me, opp, aegar := s82rAegarTable(t)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 1)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(aegar, wall, 3); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1 (Aegar is a Giant Wizard)", got)
	}
}

func TestS82RprAegarNeedsExcessDamage(t *testing.T) {
	g, me, opp, _ := s82rAegarTable(t)
	exact := effectstest.B12Creature(g, opp.ID, "Exact", "Creature — Wall", 0, 3)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(exact))
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(exact) {
		t.Fatal("the creature survived exactly lethal damage")
	}
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("hand grew by %d on exactly lethal damage, want 0", got)
	}
}

func TestS82RprAegarIgnoresDamageFromOtherSources(t *testing.T) {
	g, me, opp, _ := s82rAegarTable(t)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mineBear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		// An opponent's creature (not a Giant or Wizard) and my own Bear.
		if err := g.DealDamageToCreatureForEffect(mineBear, theirs, 5); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(theirs, wall, 5); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("hand grew by %d, want 0: a Bear is neither a Giant, a Wizard nor a spell", got)
	}
	// An opponent's spell on an opponent's creature is not a spell you controlled.
	other := effectstest.B12Creature(g, opp.ID, "Other Wall", "Creature — Wall", 0, 1)
	effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetCard(other))
	hand = me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("hand grew by %d on an opponent's spell, want 0", got)
	}
}

func TestS82RprAegarIgnoresYourOwnCreatures(t *testing.T) {
	g, me, _, aegar := s82rAegarTable(t)
	mine := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 1)
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(aegar, mine, 3); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("hand grew by %d for excess damage to your own creature, want 0", got)
	}
}

// --- Firesong and Sunspeaker -----------------------------------------------------

func s82rFiresongTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%2]
	effectstest.B12Push(g, me.ID, "Firesong and Sunspeaker", "Legendary Creature — Minotaur Cleric", s82rFiresongOracle, 4, 6)
	effectstest.ToMain(t, g)
	return g, me, opp
}

func TestS82RprFiresongGivesRedInstantsLifelink(t *testing.T) {
	g, me, opp := s82rFiresongTable(t)
	bolt := game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", Colors: []string{"R"},
		OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(bolt)
	life, oppLife := me.Life, opp.Life
	if err := g.CastSpell(me.ID, bolt.InstanceID, game.CastSpellParams{Targets: effectstest.B16TargetPlayer(opp.ID)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 || me.Life != life+3 {
		t.Errorf("life me %d opp %d, want %d and %d", me.Life, opp.Life, life+3, oppLife-3)
	}
	if !effectstest.StackFullyEmpty(g) {
		t.Error("a red spell's lifelink triggered the white clause")
	}
}

func TestS82RprFiresongDealsThreeWhenAWhiteSpellGainsYouLife(t *testing.T) {
	g, me, opp := s82rFiresongTable(t)
	salve := game.Card{InstanceID: uuid.New(), Name: "Revitalize", TypeLine: "Instant", Colors: []string{"W"},
		OracleID: "b1385b03-cb4b-4812-857f-7421f1df39af", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(salve)
	oppLife := opp.Life
	if err := g.CastSpell(me.ID, salve.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 {
		t.Errorf("opponent's life %d, want %d", opp.Life, oppLife-3)
	}
}

func TestS82RprFiresongIgnoresANonwhiteSpellsLifeGain(t *testing.T) {
	g, me, opp := s82rFiresongTable(t)
	salve := game.Card{InstanceID: uuid.New(), Name: "Green Revitalize", TypeLine: "Instant", Colors: []string{"G"},
		OracleID: "b1385b03-cb4b-4812-857f-7421f1df39af", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(salve)
	oppLife := opp.Life
	if err := g.CastSpell(me.ID, salve.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a nonwhite spell's life gain triggered Firesong and Sunspeaker")
	}
}

// --- Gyruda, Doom of Depths -------------------------------------------------------

func TestS82RprGyrudaPutsAnEvenCreatureFromTheMilledCardsOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%2]
	effectstest.ToMain(t, g)
	me.Library.Cards, opp.Library.Cards = nil, nil
	deepMine := effectstest.SeedLibraryTop(me, "Deep Even Mine", "Creature — Bear", "{2}")
	mineOdd := effectstest.SeedLibraryTop(me, "Odd Bear", "Creature — Bear", "{1}")
	mineSpell := effectstest.SeedLibraryTop(me, "Even Sorcery", "Sorcery", "{2}")
	mineLand := effectstest.SeedLibraryTop(me, "Land", "Land", "")
	mineEven := effectstest.SeedLibraryTop(me, "Even Bear", "Creature — Bear", "{2}")
	mineTop := effectstest.SeedLibraryTop(me, "Top Odd", "Creature — Bear", "{3}")
	deepTheirs := effectstest.SeedLibraryTop(opp, "Deep Even Theirs", "Creature — Bear", "{4}")
	theirEven := effectstest.SeedLibraryTop(opp, "Their Even Bear", "Creature — Bear", "{4}")
	theirX := effectstest.SeedLibraryTop(opp, "Their X Bear", "Creature — Bear", "{X}{X}")
	theirOdd := effectstest.SeedLibraryTop(opp, "Their Odd Bear", "Creature — Bear", "{5}")
	theirTop := effectstest.SeedLibraryTop(opp, "Their Top Spell", "Instant", "{2}")
	_ = []uuid.UUID{deepMine, deepTheirs, mineLand, mineSpell, mineOdd, mineTop, theirTop, theirOdd}
	id := effectstest.PushCatalogHandCard(me, "Gyruda, Doom of Depths", "Legendary Creature — Demon Kraken", s82rGyrudaOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pc := effectstest.LatestChooseCards(g, me.ID)
	if pc == nil {
		t.Fatal("no pick for the milled creature")
	}
	want := map[uuid.UUID]bool{mineEven: true, theirEven: true, theirX: true}
	if len(pc.ChooseCards) != len(want) || pc.ChooseMin != 1 || pc.ChooseMax != 1 {
		t.Fatalf("offered %v [%d,%d], want the even creatures among the milled cards", pc.ChooseCards, pc.ChooseMin, pc.ChooseMax)
	}
	for _, c := range pc.ChooseCards {
		if !want[c] {
			t.Errorf("offered %s, which is not an even creature among the milled cards", effectstest.B12Card(t, g, c).Name)
		}
	}
	if err := g.ResolveChooseCards(pc.ID, me.ID, []uuid.UUID{theirEven}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got, ok := effectstest.BattlefieldCard(g, theirEven)
	if !ok {
		// A returned card is a new object; find it by name.
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Their Even Bear" {
				got, ok = c, true
			}
		}
	}
	if !ok || got.Controller != me.ID {
		t.Fatalf("the creature is not on the battlefield under my control: %+v", got)
	}
	if !opp.Graveyard.Contains(theirX) || !me.Graveyard.Contains(mineEven) {
		t.Error("the cards not chosen did not stay in the graveyards")
	}
	if opp.Library.Contains(deepTheirs) == false || me.Library.Contains(deepMine) == false {
		t.Error("more than four cards were milled")
	}
}

func TestS82RprGyrudaMillsAndPutsNothingWithoutAnEvenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%2]
	effectstest.ToMain(t, g)
	for _, p := range []*game.Player{me, opp} {
		p.Library.Cards = nil
		for i := 0; i < 5; i++ {
			effectstest.SeedLibraryTop(p, "Odd Bear", "Creature — Bear", "{1}")
		}
	}
	id := effectstest.PushCatalogHandCard(me, "Gyruda, Doom of Depths", "Legendary Creature — Demon Kraken", s82rGyrudaOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("a pick was offered with no even creature among the milled cards")
	}
	if me.Graveyard.Size() != 4 || opp.Graveyard.Size() != 4 {
		t.Errorf("graveyards %d and %d, want 4 each", me.Graveyard.Size(), opp.Graveyard.Size())
	}
}

// --- Grimgrin, Corpse-Born ---------------------------------------------------------

func TestS82RprGrimginEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Grimgrin, Corpse-Born", "Legendary Creature — Zombie Warrior", s82rGrimgrinOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, id); !ok || !c.Tapped {
		t.Errorf("Grimgrin on the battlefield %v, tapped %v: it enters tapped", ok, c.Tapped)
	}
}

func TestS82RprGrimginSacrificeUntapsItAndGrowsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	grim := effectstest.B12Push(g, me.ID, "Grimgrin, Corpse-Born", "Legendary Creature — Zombie Warrior", s82rGrimgrinOracle, 5, 5)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == grim {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, grim, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{grim}}); err == nil {
		t.Fatal("Grimgrin sacrificed itself: the cost says another creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, grim, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, grim); c.Tapped || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("Grimgrin tapped %v with %d counters, want untapped with 1", c.Tapped, c.Counters[game.CounterPlusOne])
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
}

func TestS82RprGrimginAttackDestroysADefendersCreatureThenGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%2]
	grim := effectstest.B12Push(g, me.ID, "Grimgrin, Corpse-Born", "Legendary Creature — Zombie Warrior", s82rGrimgrinOracle, 5, 5)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.FinABAttack(t, g, opp.ID, grim)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, victim)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the defending player's creature survived")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was destroyed: the target is the defender's")
	}
	if got := s82rCounters(t, g, grim); got != 1 {
		t.Errorf("Grimgrin has %d counters, want 1", got)
	}
}
