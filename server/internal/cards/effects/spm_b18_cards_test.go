package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b18_cards_test.go — Marvel's Spider-Man, batch 18 (#4460). Each test
// plays the real catalog entry and asserts what its text says.

const (
	spmB18PassageOracle   = "17770eeb-ecbf-4510-af2d-acececfbe146"
	spmB18MysterioOracle  = "40f664d8-4c73-44aa-8fb1-d38771a42520"
	spmB18NormanOracle    = "aa5b06a4-90c3-45e7-ab83-cb5c96ded21d"
	spmB18OriginOracle    = "7fbe9056-190f-40a1-bee9-59ebd6f981f5"
	spmB18ParkerOracle    = "03d1a4e9-3ebc-491f-8741-5d933b500abf"
	spmB18SandmanOracle   = "8b205480-b1fb-490c-b1c9-cf5cc3c5acdb"
	spmB18ScarletOracle   = "13ab2f3f-f65e-42b5-b743-39b30a4c4751"
	spmB18VerseOracle     = "7ef7d902-1f05-40b3-8997-dec197b0b678"
	spmB18WomanOracle     = "be2b9c6d-4ecb-49ec-b276-4aa93c5dfc00"
	spmB18SymbioteOracle  = "bc681ecf-2a52-4782-9a47-90f5a9f435d8"
	spmB18SymbioteType    = "Legendary Creature — Symbiote Spider Hero"
	spmB18SandmanTypeLine = "Legendary Creature — Sand Elemental Villain"
)

func TestSpmB18EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spmB18PassageOracle:       "Multiversal Passage",
		spmB18MysterioOracle:      "Mysterio, Master of Illusion",
		spmB18NormanOracle:        "Norman Osborn",
		spmB18NormanOracle + "#1": "Green Goblin",
		spmB18OriginOracle:        "Origin of Spider-Man",
		spmB18ParkerOracle:        "Parker Luck",
		spmB18SandmanOracle:       "Sandman, Shifting Scoundrel",
		spmB18ScarletOracle:       "Scarlet Spider, Ben Reilly",
		spmB18VerseOracle:         "Spider-Verse",
		spmB18WomanOracle:         "Spider-Woman, Stunning Savior",
		spmB18SymbioteOracle:      "Symbiote Spider-Man",
	}
	for oracle, name := range want {
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

// spmB18Graveyard puts a card into p's graveyard.
func spmB18Graveyard(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Graveyard.PushTop(c) })
	return c.InstanceID
}

// spmB18LibraryTop puts a card on top of p's library.
func spmB18LibraryTop(g *game.Game, p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	})
	return id
}

func spmB18Count(g *game.Game, controller uuid.UUID, name string) (all, tokens int) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			all++
			if c.IsToken() {
				tokens++
			}
		}
	}
	return
}

// --- Multiversal Passage ---

func TestSpmB18PassageChoosesATypeAndPaysTwoLifeToEnterUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	id := playLandFromHand(t, g, "Multiversal Passage", spmB18PassageOracle)
	answerEntryPayLife(t, g, me.ID, true)
	pick := latestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no basic land type prompt")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Label == "Swamp" {
			idx = i
		}
	}
	answerOptionPick(t, g, me.ID, idx)
	card, ok := battlefieldCard(g, id)
	if !ok || card.Tapped {
		t.Fatalf("the Passage is not on the battlefield untapped: %v %v", ok, card.Tapped)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	if got := effectiveSubtypes(t, g, id); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("subtypes %v, want [Swamp]", got)
	}
	if got := ltProduced(t, g, id); !reflect.DeepEqual(got, []string{"{B}"}) {
		t.Errorf("taps for %v, want {B}", got)
	}
}

func TestSpmB18PassageEntersTappedWhenTheLifeIsNotPaid(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	id := playLandFromHand(t, g, "Multiversal Passage", spmB18PassageOracle)
	answerEntryPayLife(t, g, me.ID, false)
	answerOptionPick(t, g, me.ID, 0)
	card, ok := battlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatal("the Passage did not enter tapped")
	}
	if me.Life != life {
		t.Errorf("life changed: %d -> %d", life, me.Life)
	}
	if got := effectiveSubtypes(t, g, id); len(got) != 1 {
		t.Errorf("subtypes %v, want exactly the chosen one", got)
	}
}

// --- Mysterio, Master of Illusion ---

func TestSpmB18MysterioMakesAnIllusionPerNontokenVillainAndExilesThemWhenHeLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Creature(g, me.ID, "Henchman", "Creature — Human Villain", 1, 1)
	b12Creature(g, me.ID, "Token Villain", "Token Creature — Villain", 1, 1)
	b12Creature(g, me.ID, "Hero", "Creature — Human Hero", 1, 1)
	b12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 1, 1)

	mysterio := castCatalogSpell(t, g, "Mysterio, Master of Illusion", "Legendary Creature — Human Villain", spmB18MysterioOracle, nil)
	passPriorityAroundTable(t, g)
	all, tokens := spmB18Count(g, me.ID, "Illusion Villain")
	if all != 2 || tokens != 2 {
		t.Fatalf("%d Illusion Villains (%d tokens), want 2: Mysterio and the Henchman", all, tokens)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Illusion Villain" && (c.Power != 3 || c.Toughness != 3 || !c.HasSubtype("Illusion") || !c.HasSubtype("Villain")) {
			t.Errorf("token %+v, want a 3/3 Illusion Villain", c.TypeLine)
		}
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(mysterio); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if all, _ := spmB18Count(g, me.ID, "Illusion Villain"); all != 0 {
		t.Errorf("%d Illusion Villains remain after Mysterio left", all)
	}
	if all, _ := spmB18Count(g, me.ID, "Henchman"); all != 1 {
		t.Error("the Henchman was exiled too")
	}
}

func TestSpmB18MysterioGoneBeforeHisTriggerLeavesTheTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Henchman", "Creature — Human Villain", 1, 1)
	mysterio := castCatalogSpell(t, g, "Mysterio, Master of Illusion", "Legendary Creature — Human Villain", spmB18MysterioOracle, nil)
	msc17UntilATriggerIsOnTheStack(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(mysterio); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if all, _ := spmB18Count(g, me.ID, "Illusion Villain"); all != 1 {
		t.Fatalf("%d Illusion Villains, want 1 (the Henchman; Mysterio had left)", all)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("a delayed exile was scheduled for a Mysterio that had already left: %+v", g.DelayedTriggers)
	}
}

// --- Norman Osborn // Green Goblin ---

func spmB18NormanCard(owner uuid.UUID, face int) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: spmB18NormanOracle, Layout: game.LayoutModalDFC,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Norman Osborn", TypeLine: "Legendary Creature — Human Scientist Villain", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 1, Toughness: 1},
			{Name: "Green Goblin", TypeLine: "Legendary Creature — Goblin Human Villain", ManaCost: "{1}{U}{B}{R}", Colors: []string{"U", "B", "R"}, Power: 3, Toughness: 3},
		},
	}
	c.SetFace(face)
	return c
}

func TestSpmB18NormanCantBeBlockedConnivesAndTransforms(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	norman := pushBattlefieldCardWithTimestamp(g, spmB18NormanCard(me.ID, 0))
	blocker := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	drawn := spmB18LibraryTop(g, me, "Drawn Rock", "Artifact", "{2}")
	brAttack(t, g, norman)
	if err := g.DeclareBlocker(blocker, norman); err == nil {
		t.Fatal("Norman Osborn was blocked")
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(drawn) {
		t.Fatal("conniving drew no card")
	}
	answerDiscard(t, g, me.ID, drawn)
	passPriorityAroundTable(t, g)
	if plusOnes(g, norman) != 1 {
		t.Errorf("counters %d, want 1 for a nonland discard", plusOnes(g, norman))
	}

	advanceTo(t, g, game.StepPostcombatMain)
	floatForTest(g, me, "CUBR")
	if err := g.ActivateCatalogAbility(me.ID, norman, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("transform: %v", err)
	}
	passPriorityAroundTable(t, g)
	card := battlefieldCardFor(g, norman)
	if card == nil || card.ActiveFace != 1 || card.Name != "Green Goblin" {
		t.Fatalf("after the transform: %+v, want Green Goblin", card)
	}
	if !hasKeywordOnBattlefield(t, g, norman, "flying") || !hasKeywordOnBattlefield(t, g, norman, "menace") {
		t.Error("Green Goblin lacks flying and menace")
	}
}

func TestSpmB18NormanTransformsOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	norman := pushBattlefieldCardWithTimestamp(g, spmB18NormanCard(me.ID, 0))
	advanceTo(t, g, game.StepUpkeep)
	floatForTest(g, me, "CUBR")
	if err := g.ActivateCatalogAbility(me.ID, norman, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed outside a main phase")
	}
}

func TestSpmB18GreenGoblinGivesDiscardedCardsMayhemAtTheirManaCostLessTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	pushBattlefieldCardWithTimestamp(g, spmB18NormanCard(me.ID, 1))
	bear := mayhemHandCard(g, me, "Discarded Bear", "Creature — Bear", "{2}{G}{G}", "")
	mayhemDiscard(t, g, me, bear)
	milled := spmB18Graveyard(g, me, game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}"})
	land := mayhemHandCard(g, me, "Discarded Land", "Land", "", "")
	mayhemDiscard(t, g, me, land)

	floatForTest(g, me, "GGG")
	if err := g.CastSpell(me.ID, milled, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedMayhem}); err == nil {
		t.Fatal("a card that was not discarded this turn was cast with mayhem")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedMayhem}); err == nil {
		t.Fatal("{G} paid {2}{G}{G} less {2}")
	}
	floatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedMayhem}); err != nil {
		t.Fatalf("granted mayhem for {G}{G}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the discarded Bear did not enter")
	}
	if !me.Graveyard.Contains(milled) {
		t.Error("the milled Bear left the graveyard")
	}
}

// --- Origin of Spider-Man ---

func spmB18OriginAtChapter(g *game.Game, me *game.Player, lore int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Origin of Spider-Man", TypeLine: "Enchantment — Saga",
		OracleID: spmB18OriginOracle, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLore: lore},
	})
}

func TestSpmB18OriginChapterOneMakesASpider(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Origin of Spider-Man", "Enchantment — Saga", spmB18OriginOracle, nil)
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Spider" && c.Power == 2 && c.Toughness == 1 && hasKeywordOnBattlefield(t, g, c.InstanceID, "reach") {
			found = true
		}
	}
	if !found {
		t.Error("no 2/1 green Spider with reach")
	}
}

func TestSpmB18OriginChapterTwoMakesYourCreatureALegendarySpiderHero(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	spmB18OriginAtChapter(g, me, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no chapter II target prompt")
	}
	if hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, bear) {
		t.Fatalf("chapter II offered %v; want only your creatures", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 1 {
		t.Errorf("counters %d, want 1", plusOnes(g, bear))
	}
	c := layeredCard(t, g, bear)
	if !c.IsLegendary() || !c.HasSubtype("Spider") || !c.HasSubtype("Hero") || !c.HasSubtype("Bear") || !c.IsCreature() {
		t.Errorf("type line now %v %v %v, want a Legendary Creature — Bear Spider Hero", c.Effective().Supertypes, c.Effective().Types, c.Effective().Subtypes)
	}
	nextTurnOfMine(t, g)
	if c := layeredCard(t, g, bear); !c.IsLegendary() || !c.HasSubtype("Spider") {
		t.Error("the type change ended; it has no duration")
	}
}

func TestSpmB18OriginChapterThreeGivesDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := spmB18OriginAtChapter(g, me, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "double strike") {
		t.Error("the creature lacks double strike")
	}
	if loreCountersOn(g, saga) >= 0 {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

// --- Parker Luck ---

func TestSpmB18ParkerLuckEachLosesTheOthersManaValueThenTakesTheirCard(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	b12Push(g, me.ID, "Parker Luck", "Enchantment", spmB18ParkerOracle, 0, 0)
	mine := spmB18LibraryTop(g, me, "My Three", "Sorcery", "{2}{B}")
	theirs := spmB18LibraryTop(g, opp, "Their Two", "Instant", "{1}{U}")
	myLife, oppLife := me.Life, opp.Life
	advanceTo(t, g, game.StepEnd)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt at the end step")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Fatal("one target player was accepted for two")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID}, {Kind: game.TargetPlayer, ID: opp.ID},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != myLife-2 || opp.Life != oppLife-3 {
		t.Errorf("life %d / %d, want %d / %d", me.Life, opp.Life, myLife-2, oppLife-3)
	}
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("the revealed cards are not in their owners' hands")
	}
}

// --- Sandman, Shifting Scoundrel ---

func spmB18SandmanInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	return spmB18Graveyard(g, me, game.Card{Name: "Sandman, Shifting Scoundrel", TypeLine: spmB18SandmanTypeLine,
		OracleID: spmB18SandmanOracle, ManaCost: "{1}{G}{G}"})
}

func TestSpmB18SandmanReturnsWithALandTappedAndCountsLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	for i := 0; i < 3; i++ {
		b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	sandman := spmB18SandmanInGraveyard(g, me)
	land := spmB18Graveyard(g, me, game.Card{Name: "Dead Forest", TypeLine: "Basic Land — Forest"})
	rock := spmB18Graveyard(g, me, game.Card{Name: "Dead Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	floatForTest(g, me, "CCCGG")
	if err := g.ActivateCatalogAbility(me.ID, sandman, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err == nil {
		t.Fatal("a nonland card was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, sandman, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{sandman, land} {
		c, ok := battlefieldCard(g, id)
		if !ok || !c.Tapped {
			t.Fatalf("%s is not on the battlefield tapped", id)
		}
	}
	c := layeredCard(t, g, sandman)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("Sandman is %d/%d, want 4/4 (four lands)", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSpmB18SandmanStaysWhenTheLandTargetIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sandman := spmB18SandmanInGraveyard(g, me)
	land := spmB18Graveyard(g, me, game.Card{Name: "Dead Forest", TypeLine: "Basic Land — Forest"})
	floatForTest(g, me, "CCCGG")
	if err := g.ActivateCatalogAbility(me.ID, sandman, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(land); err != nil {
			t.Fatalf("exile the land: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(sandman) {
		t.Error("Sandman returned though the ability's only target was gone")
	}
}

func TestSpmB18SandmanCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sandman := b12Push(g, me.ID, "Sandman, Shifting Scoundrel", spmB18SandmanTypeLine, spmB18SandmanOracle, 0, 0)
	for i := 0; i < 2; i++ {
		b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	small := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	brAttack(t, g, sandman)
	brRefusal(t, g.DeclareBlocker(small, sandman), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(big, sandman); err != nil {
		t.Fatalf("a 3-power creature could not block: %v", err)
	}
}

// --- Scarlet Spider, Ben Reilly ---

func TestSpmB18ScarletSpiderWebSlungEntersWithTheReturnedCreaturesManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	scarlet := handCardFull(me, "Scarlet Spider, Ben Reilly", "Legendary Creature — Spider Human Hero", "{1}{R}{G}", spmB18ScarletOracle, []string{"R", "G"})
	tapped := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tapped Ogre", TypeLine: "Creature — Ogre",
		ManaCost: "{2}{R}", Power: 3, Toughness: 3, Tapped: true, Owner: me.ID, Controller: me.ID})
	floatForTest(g, me, "RG")
	if err := g.CastSpell(me.ID, scarlet, game.CastSpellParams{Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{tapped}}); err != nil {
		t.Fatalf("web-slinging: %v", err)
	}
	if !me.Hand.Contains(tapped) {
		t.Fatal("the tapped creature did not return to hand")
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, scarlet) != 3 {
		t.Errorf("counters %d, want 3 (the Ogre's mana value)", plusOnes(g, scarlet))
	}
	if !hasKeywordOnBattlefield(t, g, scarlet, "trample") {
		t.Error("no trample")
	}
}

func TestSpmB18ScarletSpiderCastForItsManaCostGetsNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	scarlet := handCardFull(me, "Scarlet Spider, Ben Reilly", "Legendary Creature — Spider Human Hero", "{1}{R}{G}", spmB18ScarletOracle, []string{"R", "G"})
	floatForTest(g, me, "CRG")
	if err := g.CastSpell(me.ID, scarlet, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scarlet) || plusOnes(g, scarlet) != 0 {
		t.Errorf("on battlefield %v with %d counters, want 0", g.Battlefield.Contains(scarlet), plusOnes(g, scarlet))
	}
}

// --- Spider-Verse ---

func TestSpmB18SpiderVerseLegendRuleSparesSpidersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Spider-Verse", "Enchantment", spmB18VerseOracle, 0, 0)
	a := b12Creature(g, me.ID, "Test Spider", "Legendary Creature — Spider Human", 2, 2)
	b := b12Creature(g, me.ID, "Test Spider", "Legendary Creature — Spider Human", 2, 2)
	g.RunStateChecksForTest()
	if !onBoard(g, a) || !onBoard(g, b) || legendPrompt(g, me.ID) != nil {
		t.Fatal("two legendary Spiders with one name did not both stay")
	}
	b12Creature(g, me.ID, legendName, legendTypeLine, 2, 2)
	b12Creature(g, me.ID, legendName, legendTypeLine, 2, 2)
	b12Creature(g, opp.ID, "Test Spider", "Legendary Creature — Spider Human", 2, 2)
	b12Creature(g, opp.ID, "Test Spider", "Legendary Creature — Spider Human", 2, 2)
	g.RunStateChecksForTest()
	if legendPrompt(g, me.ID) == nil {
		t.Error("a legendary non-Spider pair was exempted")
	}
	if legendPrompt(g, opp.ID) == nil {
		t.Error("the opponent's Spiders were exempted")
	}
}

func TestSpmB18SpiderVerseCopiesASpellFromTheGraveyardOnceAndTheCopyHasHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	b12Push(g, me.ID, "Spider-Verse", "Enchantment", spmB18VerseOracle, 0, 0)
	first := mayhemHandCard(g, me, "Spider-Islanders", "Creature — Spider Horror Citizen", "{3}{R}", spm8IslandersOracle)
	mayhemDiscard(t, g, me, first)
	floatForTest(g, me, "CR")
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	all, tokens := spmB18Count(g, me.ID, "Spider-Islanders")
	if all != 2 || tokens != 1 {
		t.Fatalf("%d Spider-Islanders (%d tokens), want the card and one token copy", all, tokens)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Spider-Islanders" {
			continue
		}
		if haste := hasKeywordOnBattlefield(t, g, c.InstanceID, "haste"); haste != c.IsToken() {
			t.Errorf("token %v has haste %v; only the copy gains haste", c.IsToken(), haste)
		}
	}

	// Only once each turn.
	second := mayhemHandCard(g, me, "Spider-Islanders", "Creature — Spider Horror Citizen", "{3}{R}", spm8IslandersOracle)
	mayhemDiscard(t, g, me, second)
	floatForTest(g, me, "CR")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err != nil {
		t.Fatalf("second mayhem cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if all, tokens := spmB18Count(g, me.ID, "Spider-Islanders"); all != 3 || tokens != 1 {
		t.Errorf("%d Spider-Islanders (%d tokens) after the second cast, want 3 and still 1", all, tokens)
	}
}

func TestSpmB18SpiderVerseIgnoresSpellsFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Spider-Verse", "Enchantment", spmB18VerseOracle, 0, 0)
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if all, tokens := spmB18Count(g, me.ID, "Grizzly Bears"); all != 1 || tokens != 0 {
		t.Errorf("%d Grizzly Bears (%d tokens), want only the cast one", all, tokens)
	}
}

// --- Spider-Woman, Stunning Savior ---

func TestSpmB18SpiderWomanTapsOpposingArtifactsAndCreaturesOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Spider-Woman, Stunning Savior", "Legendary Creature — Spider Human Hero", spmB18WomanOracle, 2, 2)
	if !hasKeywordOnBattlefield(t, g, g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID, "flying") {
		t.Error("Spider-Woman lacks flying")
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, TokenCard("2/2 green Bear"), 1)
		_ = g.CreateTokenForEffect(opp.ID, TreasureToken(), 1)
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 green Bear"), 1)
	})
	for _, c := range g.Battlefield.Cards {
		if !c.IsToken() {
			continue
		}
		if want := c.Controller == opp.ID; c.Tapped != want {
			t.Errorf("%s under %s: tapped %v, want %v", c.Name, c.Controller, c.Tapped, want)
		}
	}
}

// --- Symbiote Spider-Man ---

func TestSpmB18SymbioteLooksAtThatManyAndKeepsOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sym := b12Push(g, me.ID, "Symbiote Spider-Man", spmB18SymbioteType, spmB18SymbioteOracle, 2, 4)
	ids := seedLibrary(me, "A", "B", "C", "D")
	attackWith(t, g, opp.ID, sym)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no look prompt after 2 combat damage")
	}
	if len(pick.ChooseCards) != 2 {
		t.Fatalf("looked at %d cards, want 2", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards[1:]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	g.SettleResolution()
	if !me.Hand.Contains(pick.ChooseCards[1]) || !me.Graveyard.Contains(pick.ChooseCards[0]) {
		t.Error("want one card in hand and the other in the graveyard")
	}
	_ = ids
}

func TestSpmB18SymbioteFindNewHostGivesItsTriggerToYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sym := spmB18Graveyard(g, me, game.Card{Name: "Symbiote Spider-Man", TypeLine: spmB18SymbioteType,
		OracleID: spmB18SymbioteOracle, ManaCost: "{2}{U/B}", Power: 2, Toughness: 4})
	bear := b12Creature(g, me.ID, "Host Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	floatForTest(g, me, "CCB")
	if err := g.ActivateCatalogAbility(me.ID, sym, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Fatal("an opponent's creature was a legal host")
	}
	if err := g.ActivateCatalogAbility(me.ID, sym, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(sym) {
		t.Error("the cost did not exile the card")
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 1 {
		t.Fatalf("counters %d, want 1", plusOnes(g, bear))
	}
	seedLibrary(me, "A", "B", "C", "D")
	nextTurnOfMine(t, g)
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 3 {
		t.Fatalf("the host's look: %+v, want three cards for its 3 damage", pick)
	}
}

func TestSpmB18SymbioteFindNewHostOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepUpkeep)
	sym := spmB18Graveyard(g, me, game.Card{Name: "Symbiote Spider-Man", TypeLine: spmB18SymbioteType,
		OracleID: spmB18SymbioteOracle, ManaCost: "{2}{U/B}", Power: 2, Toughness: 4})
	bear := b12Creature(g, me.ID, "Host Bear", "Creature — Bear", 2, 2)
	floatForTest(g, me, "CCU")
	if err := g.ActivateCatalogAbility(me.ID, sym, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("Find New Host was activated in the upkeep")
	}
}
