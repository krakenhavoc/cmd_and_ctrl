package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Innistrad: Crimson Vow batch 14: eight transform double-faced cards, run
// as imported cards through the real triggers, costs and layers.

type vowB14Card struct {
	oracle, front, frontType, cost, frontText, fp, ft string
	back, backType, backText, bp, bt                  string
	colors                                            []string
	keywords                                          []string
}

var (
	vowFarmer = vowB14Card{"60d591f6-6d84-45d2-9d78-a2ef07a3e403", "Desperate Farmer", "Creature — Human Peasant", "{2}{B}",
		"Lifelink\nWhen another creature you control dies, transform this creature.", "2", "2",
		"Depraved Harvester", "Creature — Human Knight", "Lifelink", "4", "3", []string{"B"}, []string{"Lifelink"}}
	vowRecluse = vowB14Card{"9ad999fe-889a-426a-86b3-c0c311b82041", "Ragged Recluse", "Creature — Human Peasant", "{1}{B}",
		"At the beginning of your end step, if you discarded a card this turn, transform this creature.", "2", "1",
		"Odious Witch", "Creature — Human Warlock", "Whenever this creature attacks, defending player loses 1 life and you gain 1 life.", "3", "3", []string{"B"}, nil}
	vowBystander = vowB14Card{"fc58a086-2743-4465-bf16-4a20701cd3b1", "Panicked Bystander", "Creature — Human Peasant", "{1}{W}",
		"Whenever this creature or another creature you control dies, you gain 1 life.\nAt the beginning of your end step, if you gained 3 or more life this turn, transform this creature.", "2", "2",
		"Cackling Culprit", "Creature — Human Rogue", "Whenever this creature or another creature you control dies, you gain 1 life.\n{1}{B}: This creature gains deathtouch until end of turn.", "3", "5", []string{"W"}, nil}
	vowOddity = vowB14Card{"90379825-1894-4b79-bf72-445175e97ffb", "Ulvenwald Oddity", "Creature — Beast", "{2}{G}{G}",
		"Trample, haste\n{5}{G}{G}: Transform this creature.", "4", "4",
		"Ulvenwald Behemoth", "Creature — Beast Horror", "Trample, haste\nOther creatures you control get +1/+1 and have trample and haste.", "8", "8", []string{"G"}, []string{"Trample", "Haste"}}
	vowFodder = vowB14Card{"aa2c0d23-8df8-408b-a8c0-c4cd005f078d", "Catapult Fodder", "Creature — Zombie", "{2}{B}",
		"At the beginning of combat on your turn, if you control three or more creatures that each have toughness greater than their power, transform this creature.", "1", "5",
		"Catapult Captain", "Creature — Zombie", "{2}{B}, {T}, Sacrifice another creature: Target opponent loses life equal to the sacrificed creature's toughness.", "2", "6", []string{"B"}, nil}
	vowSquire = vowB14Card{"a6375c5b-ec78-4ee8-b681-8d92797a5b43", "Bloodsworn Squire", "Creature — Vampire Soldier", "{3}{B}",
		"{1}{B}, Discard a card: This creature gains indestructible until end of turn. Tap it. Then if there are four or more creature cards in your graveyard, transform this creature.", "3", "3",
		"Bloodsworn Knight", "Creature — Vampire Knight", "Bloodsworn Knight's power and toughness are each equal to the number of creature cards in your graveyard.\n{1}{B}, Discard a card: This creature gains indestructible until end of turn. Tap it.", "*", "*", []string{"B"}, nil}
	vowGrove = vowB14Card{"1be58157-186e-48e4-b375-dc5f10d785ac", "Dormant Grove", "Enchantment", "{3}{G}",
		"At the beginning of combat on your turn, put a +1/+1 counter on target creature you control. Then if that creature has toughness 6 or greater, transform this enchantment.", "", "",
		"Gnarled Grovestrider", "Creature — Treefolk", "Vigilance\nOther creatures you control have vigilance.", "3", "6", []string{"G"}, []string{"Vigilance"}}
	vowStatue = vowB14Card{"d8c0f103-4658-4f4f-8429-c06bdd017846", "Foreboding Statue", "Artifact Creature — Construct", "{3}",
		"{T}: Add one mana of any color. Put an omen counter on this creature.\nAt the beginning of your end step, if there are three or more omen counters on this creature, untap it, then transform it.", "1", "2",
		"Forsaken Thresher", "Artifact Creature — Construct", "At the beginning of your first main phase, add one mana of any color.", "5", "5", nil, nil}
)

var vowB14All = []vowB14Card{vowFarmer, vowRecluse, vowBystander, vowOddity, vowFodder, vowSquire, vowGrove, vowStatue}

func (w vowB14Card) row() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     w.front + " // " + w.back,
		Layout:   "transform",
		TypeLine: w.frontType + " // " + w.backType,
		OracleID: uuid.MustParse(w.oracle),
		Keywords: append([]string{"Transform"}, w.keywords...),
		CardFaces: []cards.CardFace{
			{Name: w.front, TypeLine: w.frontType, ManaCost: w.cost, Colors: w.colors, Power: w.fp, Toughness: w.ft, OracleText: w.frontText},
			{Name: w.back, TypeLine: w.backType, Colors: w.colors, Power: w.bp, Toughness: w.bt, OracleText: w.backText},
		},
	}
}

func vowB14Put(t *testing.T, g *game.Game, w vowB14Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToHand(w.row(), p)
	c, ok := cardInZone(p.Hand, id)
	if !ok {
		t.Fatalf("%s did not reach the hand", w.front)
	}
	p.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	return id
}

func vowB14PutBack(t *testing.T, g *game.Game, w vowB14Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := vowB14Put(t, g, w, p)
	iwTransform(t, g, id)
	if iwFace(g, id) != 1 {
		t.Fatalf("%s did not turn over", w.front)
	}
	return id
}

func vowB14Gain(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, n); err != nil {
			t.Fatalf("gain life: %v", err)
		}
	})
}

func TestVowB14CardsAreRegisteredFull(t *testing.T) {
	for _, w := range vowB14All {
		front, ok := Lookup(w.oracle)
		if !ok || front.Name != w.front || front.Completeness != CompletenessFull || len(front.Caveats) != 0 {
			t.Errorf("%s: front registered %v as %+v", w.front, ok, front.Name)
		}
		back, ok := Lookup(w.oracle + "#1")
		if !ok || back.Name != w.back || back.Completeness != CompletenessFull || len(back.Caveats) != 0 {
			t.Errorf("%s: back registered %v as %+v", w.back, ok, back.Name)
		}
	}
}

func TestVowB14DesperateFarmerTransformsWhenAnotherCreatureOfYoursDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	farmer := vowB14Put(t, g, vowFarmer, me)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, opp.ID, theirs)
	monSettle(t, g)
	if iwFace(g, farmer) != 0 {
		t.Fatal("an opponent's creature dying transformed the Farmer")
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, bear)
	monSettle(t, g)
	if iwFace(g, farmer) != 1 {
		t.Fatal("another creature you control dying did not transform the Farmer")
	}
	if !slices.Contains(effectiveAbilities(t, g, farmer), "lifelink") {
		t.Error("Depraved Harvester lacks lifelink")
	}
	if p, tough := effectivePTOf(t, g, farmer); p != 4 || tough != 3 {
		t.Errorf("Depraved Harvester is %d/%d, want 4/3", p, tough)
	}
}

func TestVowB14RaggedRecluseNeedsADiscardThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	quiet := vowB14Put(t, g, vowRecluse, me)
	advanceToEndStepOf(t, g, 0)
	monSettle(t, g)
	if iwFace(g, quiet) != 0 {
		t.Fatal("transformed without a discard")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	recluse := vowB14Put(t, g2, vowRecluse, me2)
	card := pushHandCard(g2, me2)
	g2.WithWriteLock(func() {
		if err := g2.DiscardCardsKeyedForEffect(me2.ID, []uuid.UUID{card}, uuid.Nil, game.KeyedTail{}); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	advanceToEndStepOf(t, g2, 0)
	monSettle(t, g2)
	if iwFace(g2, recluse) != 1 {
		t.Fatal("a discard this turn did not transform the Recluse")
	}
}

func TestVowB14OdiousWitchDrainsTheDefendingPlayerOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	witch := vowB14PutBack(t, g, vowRecluse, me)
	before, mine := opp.Life, me.Life
	attackWith(t, g, opp.ID, witch)
	monSettle(t, g)
	// The drain is 1; the unblocked 3/3 then deals 3 combat damage.
	if opp.Life != before-1-3 || me.Life != mine+1 {
		t.Errorf("life = %d/%d, want %d/%d", me.Life, opp.Life, mine+1, before-4)
	}
}

func TestVowB14PanickedBystanderGainsLifeOnDeathsAndTransformsAtThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bystander := vowB14Put(t, g, vowBystander, me)
	start := me.Life
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, bear)
	monSettle(t, g)
	if me.Life != start+1 {
		t.Fatalf("life = %d, want %d after a creature died", me.Life, start+1)
	}
	advanceToEndStepOf(t, g, 0)
	monSettle(t, g)
	if iwFace(g, bystander) != 0 {
		t.Fatal("transformed having gained only 1 life")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	b2 := vowB14Put(t, g2, vowBystander, me2)
	vowB14Gain(t, g2, me2, 3)
	advanceToEndStepOf(t, g2, 0)
	monSettle(t, g2)
	if iwFace(g2, b2) != 1 {
		t.Fatal("gaining 3 life did not transform the Bystander")
	}
}

func TestVowB14CacklingCulpritGainsLifeOnItsOwnDeathAndDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	culprit := vowB14PutBack(t, g, vowBystander, me)
	start := me.Life
	addPool(t, g, me.ID, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, culprit, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	monSettle(t, g)
	if !slices.Contains(effectiveAbilities(t, g, culprit), "deathtouch") {
		t.Error("no deathtouch after {1}{B}")
	}
	killCreature(t, g, me.ID, culprit)
	monSettle(t, g)
	if me.Life != start+1 {
		t.Errorf("life = %d, want %d: its own death gains 1", me.Life, start+1)
	}
}

func TestVowB14UlvenwaldOddityTransformsAndTheBehemothLordsOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oddity := vowB14Put(t, g, vowOddity, me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	addPool(t, g, me.ID, "{G}{G}{G}{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, oddity, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with six mana, wanting seven")
	}
	addPool(t, g, me.ID, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, oddity, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	monSettle(t, g)
	if iwFace(g, oddity) != 1 {
		t.Fatal("did not transform")
	}
	if p, _ := effectivePTOf(t, g, bear); p != 3 {
		t.Errorf("Bear power = %d, want 3", p)
	}
	if p, _ := effectivePTOf(t, g, oddity); p != 8 {
		t.Errorf("Behemoth power = %d, want 8 (it pumps only others)", p)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("opponent's Bear power = %d, want 2", p)
	}
	ab := effectiveAbilities(t, g, bear)
	if !slices.Contains(ab, "trample") || !slices.Contains(ab, "haste") {
		t.Errorf("Bear abilities = %v, want trample and haste", ab)
	}
	if slices.Contains(effectiveAbilities(t, g, theirs), "haste") {
		t.Error("an opponent's creature got haste")
	}
}

func TestVowB14CatapultFodderCountsCreaturesWithToughnessOverPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := vowB14Put(t, g, vowFodder, me)
	b12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	advanceTo(t, g, game.StepBeginCombat)
	monSettle(t, g)
	if iwFace(g, fodder) != 0 {
		t.Fatal("transformed with only two such creatures")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	f2 := vowB14Put(t, g2, vowFodder, me2)
	b12Creature(g2, me2.ID, "Wall", "Creature — Wall", 0, 4)
	b12Creature(g2, me2.ID, "Wall", "Creature — Wall", 1, 4)
	b12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g2, game.StepBeginCombat)
	monSettle(t, g2)
	if iwFace(g2, f2) != 1 {
		t.Fatal("did not transform with three such creatures")
	}
}

func TestVowB14CatapultCaptainDrainsForTheSacrificedToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	captain := vowB14PutBack(t, g, vowFodder, me)
	wall := b12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	addPool(t, g, me.ID, "{B}{B}{B}")
	start := opp.Life
	params := game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{wall},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	if err := g.ActivateCatalogAbility(me.ID, captain, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	monSettle(t, g)
	if opp.Life != start-4 {
		t.Errorf("opponent life = %d, want %d", opp.Life, start-4)
	}
	// "Another" creature: the Captain can't sacrifice itself.
	addPool(t, g, me.ID, "{B}{B}{B}")
	selfParams := game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{captain},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	if err := g.ActivateCatalogAbility(me.ID, captain, 0, selfParams); err == nil {
		t.Error("the Captain sacrificed itself")
	}
}

func vowB14Activate(t *testing.T, g *game.Game, me *game.Player, id, discard uuid.UUID) {
	t.Helper()
	addPool(t, g, me.ID, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	monSettle(t, g)
}

func TestVowB14BloodswornSquireTransformsOnlyWithFourCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	squire := vowB14Put(t, g, vowSquire, me)
	for range 2 {
		pushGraveyardCardForTest(me, "Dead")
	}
	vowB14Activate(t, g, me, squire, pushHandCard(g, me))
	c, _ := battlefieldCard(g, squire)
	if iwFace(g, squire) != 0 || !c.Tapped {
		t.Fatalf("face %d tapped %v: want front face, tapped", iwFace(g, squire), c.Tapped)
	}
	if !slices.Contains(effectiveAbilities(t, g, squire), "indestructible") {
		t.Error("no indestructible")
	}

	// The discarded creature card is the fourth.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	s2 := vowB14Put(t, g2, vowSquire, me2)
	for range 3 {
		pushGraveyardCardForTest(me2, "Dead")
	}
	hand := pushHandCard(g2, me2)
	g2.WithWriteLock(func() {
		for i := range me2.Hand.Cards {
			if me2.Hand.Cards[i].InstanceID == hand {
				me2.Hand.Cards[i].TypeLine = "Creature — Test"
			}
		}
	})
	vowB14Activate(t, g2, me2, s2, hand)
	if iwFace(g2, s2) != 1 {
		t.Fatal("four creature cards did not transform the Squire")
	}
	if p, tough := effectivePTOf(t, g2, s2); p != 4 || tough != 4 {
		t.Errorf("Bloodsworn Knight is %d/%d, want 4/4", p, tough)
	}
	killCreature(t, g2, me2.ID, b12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2))
	if p, _ := effectivePTOf(t, g2, s2); p != 5 {
		t.Errorf("Bloodsworn Knight power = %d, want 5", p)
	}
}

func TestVowB14DormantGroveCountersThenTransformsAtToughnessSix(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grove := vowB14Put(t, g, vowGrove, me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 4)
	advanceTo(t, g, game.StepBeginCombat)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	monSettle(t, g)
	if tough := effectiveCardForTest(t, g, bear).CurrentToughness(); tough != 5 {
		t.Fatalf("Bear toughness = %d, want 5 (counters %d, stack empty %v)", tough, counterCount(g, bear, game.CounterPlusOne), stackFullyEmpty(g))
	}
	if iwFace(g, grove) != 0 {
		t.Fatal("transformed at toughness 5")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	g2grove := vowB14Put(t, g2, vowGrove, me2)
	big := b12Creature(g2, me2.ID, "Big", "Creature — Beast", 2, 5)
	other := b12Creature(g2, me2.ID, "Other", "Creature — Bear", 2, 2)
	advanceTo(t, g2, game.StepBeginCombat)
	answerPickTarget(t, g2, big)
	passPriorityAroundTable(t, g2)
	if iwFace(g2, g2grove) != 1 {
		t.Fatal("the counter taking the creature to toughness 6 did not transform the Grove")
	}
	if !slices.Contains(effectiveAbilities(t, g2, other), "vigilance") {
		t.Error("Gnarled Grovestrider did not give the other creature vigilance")
	}
	if slices.Contains(effectiveAbilities(t, g2, g2grove), "vigilance") != true {
		t.Error("Gnarled Grovestrider lacks vigilance")
	}
}

func TestVowB14ForebodingStatueMakesManaCountsOmensAndTransforms(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	statue := vowB14Put(t, g, vowStatue, me)
	g.WithWriteLock(func() {
		if _, err := g.AddCounterMustSettleNowForEffect(statue, "omen", 2); err != nil {
			t.Fatalf("counters: %v", err)
		}
	})
	if err := g.ActivateManaAbility(me.ID, statue, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	resolveManaPickAnyColor(t, g, me.ID)
	if len(me.ManaPool) != 1 {
		t.Errorf("mana pool = %d, want 1", len(me.ManaPool))
	}
	if got := counterCount(g, statue, "omen"); got != 3 {
		t.Fatalf("omen counters = %d, want 3", got)
	}
	advanceToEndStepOf(t, g, 0)
	monSettle(t, g)
	c, _ := battlefieldCard(g, statue)
	if iwFace(g, statue) != 1 || c.Tapped {
		t.Errorf("face %d tapped %v: want Forsaken Thresher, untapped", iwFace(g, statue), c.Tapped)
	}

	// Two omens is not enough.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	s2 := vowB14Put(t, g2, vowStatue, me2)
	if err := g2.ActivateManaAbility(me2.ID, s2, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	resolveManaPickAnyColor(t, g2, me2.ID)
	advanceToEndStepOf(t, g2, 0)
	monSettle(t, g2)
	if iwFace(g2, s2) != 0 {
		t.Error("transformed with one omen counter")
	}
}

func TestVowB14ForsakenThresherAddsManaInYourFirstMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vowB14PutBack(t, g, vowStatue, me)
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			if latestChoiceOfKind(g, game.PendingChoiceMana) != nil {
				resolveManaPickAnyColor(t, g, me.ID)
				if len(me.ManaPool) != 1 {
					t.Errorf("mana pool = %d, want 1", len(me.ManaPool))
				}
				return
			}
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("no mana prompt at the next first main phase")
}
