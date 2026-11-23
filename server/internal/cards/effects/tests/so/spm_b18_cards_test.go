package testsso

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSpmB18EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		effectstest.SpmB18PassageOracle:       "Multiversal Passage",
		effectstest.SpmB18MysterioOracle:      "Mysterio, Master of Illusion",
		effectstest.SpmB18NormanOracle:        "Norman Osborn",
		effectstest.SpmB18NormanOracle + "#1": "Green Goblin",
		effectstest.SpmB18OriginOracle:        "Origin of Spider-Man",
		effectstest.SpmB18ParkerOracle:        "Parker Luck",
		effectstest.SpmB18SandmanOracle:       "Sandman, Shifting Scoundrel",
		effectstest.SpmB18ScarletOracle:       "Scarlet Spider, Ben Reilly",
		effectstest.SpmB18VerseOracle:         "Spider-Verse",
		effectstest.SpmB18WomanOracle:         "Spider-Woman, Stunning Savior",
		effectstest.SpmB18SymbioteOracle:      "Symbiote Spider-Man",
	}
	for oracle, name := range want {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	id := effectstest.PlayLandFromHand(t, g, "Multiversal Passage", effectstest.SpmB18PassageOracle)
	effectstest.AnswerEntryPayLife(t, g, me.ID, true)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no basic land type prompt")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Label == "Swamp" {
			idx = i
		}
	}
	effectstest.AnswerOptionPick(t, g, me.ID, idx)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || card.Tapped {
		t.Fatalf("the Passage is not on the battlefield untapped: %v %v", ok, card.Tapped)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	if got := effectstest.EffectiveSubtypes(t, g, id); !reflect.DeepEqual(got, []string{"Swamp"}) {
		t.Errorf("subtypes %v, want [Swamp]", got)
	}
	if got := effectstest.LtProduced(t, g, id); !reflect.DeepEqual(got, []string{"{B}"}) {
		t.Errorf("taps for %v, want {B}", got)
	}
}

func TestSpmB18PassageEntersTappedWhenTheLifeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	id := effectstest.PlayLandFromHand(t, g, "Multiversal Passage", effectstest.SpmB18PassageOracle)
	effectstest.AnswerEntryPayLife(t, g, me.ID, false)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatal("the Passage did not enter tapped")
	}
	if me.Life != life {
		t.Errorf("life changed: %d -> %d", life, me.Life)
	}
	if got := effectstest.EffectiveSubtypes(t, g, id); len(got) != 1 {
		t.Errorf("subtypes %v, want exactly the chosen one", got)
	}
}

// --- Mysterio, Master of Illusion ---

func TestSpmB18MysterioMakesAnIllusionPerNontokenVillainAndExilesThemWhenHeLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Creature(g, me.ID, "Henchman", "Creature — Human Villain", 1, 1)
	effectstest.B12Creature(g, me.ID, "Token Villain", "Token Creature — Villain", 1, 1)
	effectstest.B12Creature(g, me.ID, "Hero", "Creature — Human Hero", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 1, 1)

	mysterio := effectstest.CastCatalogSpell(t, g, "Mysterio, Master of Illusion", "Legendary Creature — Human Villain", effectstest.SpmB18MysterioOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
	if all, _ := spmB18Count(g, me.ID, "Illusion Villain"); all != 0 {
		t.Errorf("%d Illusion Villains remain after Mysterio left", all)
	}
	if all, _ := spmB18Count(g, me.ID, "Henchman"); all != 1 {
		t.Error("the Henchman was exiled too")
	}
}

func TestSpmB18MysterioGoneBeforeHisTriggerLeavesTheTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Henchman", "Creature — Human Villain", 1, 1)
	mysterio := effectstest.CastCatalogSpell(t, g, "Mysterio, Master of Illusion", "Legendary Creature — Human Villain", effectstest.SpmB18MysterioOracle, nil)
	effectstest.Msc17UntilATriggerIsOnTheStack(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(mysterio); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if all, _ := spmB18Count(g, me.ID, "Illusion Villain"); all != 1 {
		t.Fatalf("%d Illusion Villains, want 1 (the Henchman; Mysterio had left)", all)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("a delayed exile was scheduled for a Mysterio that had already left: %+v", g.DelayedTriggers)
	}
}

func TestSpmB18NormanCantBeBlockedConnivesAndTransforms(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	norman := effectstest.PushBattlefieldCardWithTimestamp(g, effectstest.SpmB18NormanCard(me.ID, 0))
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	drawn := spmB18LibraryTop(g, me, "Drawn Rock", "Artifact", "{2}")
	effectstest.BrAttack(t, g, norman)
	if err := g.DeclareBlocker(blocker, norman); err == nil {
		t.Fatal("Norman Osborn was blocked")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(drawn) {
		t.Fatal("conniving drew no card")
	}
	effectstest.AnswerDiscard(t, g, me.ID, drawn)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, norman) != 1 {
		t.Errorf("counters %d, want 1 for a nonland discard", effectstest.PlusOnes(g, norman))
	}

	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.FloatForTest(g, me, "CUBR")
	if err := g.ActivateCatalogAbility(me.ID, norman, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("transform: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	card := effectstest.BattlefieldCardFor(g, norman)
	if card == nil || card.ActiveFace != 1 || card.Name != "Green Goblin" {
		t.Fatalf("after the transform: %+v, want Green Goblin", card)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, norman, "flying") || !effectstest.HasKeywordOnBattlefield(t, g, norman, "menace") {
		t.Error("Green Goblin lacks flying and menace")
	}
}

func TestSpmB18NormanTransformsOnlyAsASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	norman := effectstest.PushBattlefieldCardWithTimestamp(g, effectstest.SpmB18NormanCard(me.ID, 0))
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	effectstest.FloatForTest(g, me, "CUBR")
	if err := g.ActivateCatalogAbility(me.ID, norman, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("transformed outside a main phase")
	}
}

func TestSpmB18GreenGoblinGivesDiscardedCardsMayhemAtTheirManaCostLessTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	effectstest.PushBattlefieldCardWithTimestamp(g, effectstest.SpmB18NormanCard(me.ID, 1))
	bear := effectstest.MayhemHandCard(g, me, "Discarded Bear", "Creature — Bear", "{2}{G}{G}", "")
	effectstest.MayhemDiscard(t, g, me, bear)
	milled := effectstest.SpmB18Graveyard(g, me, game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}"})
	land := effectstest.MayhemHandCard(g, me, "Discarded Land", "Land", "", "")
	effectstest.MayhemDiscard(t, g, me, land)

	effectstest.FloatForTest(g, me, "GGG")
	if err := g.CastSpell(me.ID, milled, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedMayhem}); err == nil {
		t.Fatal("a card that was not discarded this turn was cast with mayhem")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedMayhem}); err == nil {
		t.Fatal("{G} paid {2}{G}{G} less {2}")
	}
	effectstest.FloatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedMayhem}); err != nil {
		t.Fatalf("granted mayhem for {G}{G}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the discarded Bear did not enter")
	}
	if !me.Graveyard.Contains(milled) {
		t.Error("the milled Bear left the graveyard")
	}
}

// --- Origin of Spider-Man ---

func spmB18OriginAtChapter(g *game.Game, me *game.Player, lore int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Origin of Spider-Man", TypeLine: "Enchantment — Saga",
		OracleID: effectstest.SpmB18OriginOracle, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLore: lore},
	})
}

func TestSpmB18OriginChapterOneMakesASpider(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Origin of Spider-Man", "Enchantment — Saga", effectstest.SpmB18OriginOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Spider" && c.Power == 2 && c.Toughness == 1 && effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "reach") {
			found = true
		}
	}
	if !found {
		t.Error("no 2/1 green Spider with reach")
	}
}

func TestSpmB18OriginChapterTwoMakesYourCreatureALegendarySpiderHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	spmB18OriginAtChapter(g, me, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no chapter II target prompt")
	}
	if effectstest.HasID(p.PickTargetCards, theirs) || !effectstest.HasID(p.PickTargetCards, bear) {
		t.Fatalf("chapter II offered %v; want only your creatures", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, bear) != 1 {
		t.Errorf("counters %d, want 1", effectstest.PlusOnes(g, bear))
	}
	c := effectstest.LayeredCard(t, g, bear)
	if !c.IsLegendary() || !c.HasSubtype("Spider") || !c.HasSubtype("Hero") || !c.HasSubtype("Bear") || !c.IsCreature() {
		t.Errorf("type line now %v %v %v, want a Legendary Creature — Bear Spider Hero", c.Effective().Supertypes, c.Effective().Types, c.Effective().Subtypes)
	}
	effectstest.NextTurnOfMine(t, g)
	if c := effectstest.LayeredCard(t, g, bear); !c.IsLegendary() || !c.HasSubtype("Spider") {
		t.Error("the type change ended; it has no duration")
	}
}

func TestSpmB18OriginChapterThreeGivesDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := spmB18OriginAtChapter(g, me, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "double strike") {
		t.Error("the creature lacks double strike")
	}
	if effectstest.LoreCountersOn(g, saga) >= 0 {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

// --- Parker Luck ---

func TestSpmB18ParkerLuckEachLosesTheOthersManaValueThenTakesTheirCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Parker Luck", "Enchantment", effectstest.SpmB18ParkerOracle, 0, 0)
	mine := spmB18LibraryTop(g, me, "My Three", "Sorcery", "{2}{B}")
	theirs := spmB18LibraryTop(g, opp, "Their Two", "Instant", "{1}{U}")
	myLife, oppLife := me.Life, opp.Life
	effectstest.AdvanceTo(t, g, game.StepEnd)
	p := effectstest.LatestPickTarget(g, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-2 || opp.Life != oppLife-3 {
		t.Errorf("life %d / %d, want %d / %d", me.Life, opp.Life, myLife-2, oppLife-3)
	}
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("the revealed cards are not in their owners' hands")
	}
}

// --- Sandman, Shifting Scoundrel ---

func spmB18SandmanInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.SpmB18Graveyard(g, me, game.Card{Name: "Sandman, Shifting Scoundrel", TypeLine: effectstest.SpmB18SandmanTypeLine,
		OracleID: effectstest.SpmB18SandmanOracle, ManaCost: "{1}{G}{G}"})
}

func TestSpmB18SandmanReturnsWithALandTappedAndCountsLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	sandman := spmB18SandmanInGraveyard(g, me)
	land := effectstest.SpmB18Graveyard(g, me, game.Card{Name: "Dead Forest", TypeLine: "Basic Land — Forest"})
	rock := effectstest.SpmB18Graveyard(g, me, game.Card{Name: "Dead Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	effectstest.FloatForTest(g, me, "CCCGG")
	if err := g.ActivateCatalogAbility(me.ID, sandman, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err == nil {
		t.Fatal("a nonland card was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, sandman, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{sandman, land} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok || !c.Tapped {
			t.Fatalf("%s is not on the battlefield tapped", id)
		}
	}
	c := effectstest.LayeredCard(t, g, sandman)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("Sandman is %d/%d, want 4/4 (four lands)", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSpmB18SandmanStaysWhenTheLandTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sandman := spmB18SandmanInGraveyard(g, me)
	land := effectstest.SpmB18Graveyard(g, me, game.Card{Name: "Dead Forest", TypeLine: "Basic Land — Forest"})
	effectstest.FloatForTest(g, me, "CCCGG")
	if err := g.ActivateCatalogAbility(me.ID, sandman, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(land); err != nil {
			t.Fatalf("exile the land: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(sandman) {
		t.Error("Sandman returned though the ability's only target was gone")
	}
}

func TestSpmB18SandmanCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sandman := effectstest.B12Push(g, me.ID, "Sandman, Shifting Scoundrel", effectstest.SpmB18SandmanTypeLine, effectstest.SpmB18SandmanOracle, 0, 0)
	for i := 0; i < 2; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	small := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.BrAttack(t, g, sandman)
	effectstest.BrRefusal(t, g.DeclareBlocker(small, sandman), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(big, sandman); err != nil {
		t.Fatalf("a 3-power creature could not block: %v", err)
	}
}

// --- Scarlet Spider, Ben Reilly ---

func TestSpmB18ScarletSpiderWebSlungEntersWithTheReturnedCreaturesManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	scarlet := effectstest.HandCardFull(me, "Scarlet Spider, Ben Reilly", "Legendary Creature — Spider Human Hero", "{1}{R}{G}", effectstest.SpmB18ScarletOracle, []string{"R", "G"})
	tapped := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tapped Ogre", TypeLine: "Creature — Ogre",
		ManaCost: "{2}{R}", Power: 3, Toughness: 3, Tapped: true, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "RG")
	if err := g.CastSpell(me.ID, scarlet, game.CastSpellParams{Strict: true, AlternativeCost: "web_slinging", AltCostIDs: []uuid.UUID{tapped}}); err != nil {
		t.Fatalf("web-slinging: %v", err)
	}
	if !me.Hand.Contains(tapped) {
		t.Fatal("the tapped creature did not return to hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, scarlet) != 3 {
		t.Errorf("counters %d, want 3 (the Ogre's mana value)", effectstest.PlusOnes(g, scarlet))
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, scarlet, "trample") {
		t.Error("no trample")
	}
}

func TestSpmB18ScarletSpiderCastForItsManaCostGetsNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	scarlet := effectstest.HandCardFull(me, "Scarlet Spider, Ben Reilly", "Legendary Creature — Spider Human Hero", "{1}{R}{G}", effectstest.SpmB18ScarletOracle, []string{"R", "G"})
	effectstest.FloatForTest(g, me, "CRG")
	if err := g.CastSpell(me.ID, scarlet, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scarlet) || effectstest.PlusOnes(g, scarlet) != 0 {
		t.Errorf("on battlefield %v with %d counters, want 0", g.Battlefield.Contains(scarlet), effectstest.PlusOnes(g, scarlet))
	}
}

// --- Spider-Verse ---

func TestSpmB18SpiderVerseLegendRuleSparesSpidersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Spider-Verse", "Enchantment", effectstest.SpmB18VerseOracle, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "Test Spider", "Legendary Creature — Spider Human", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Test Spider", "Legendary Creature — Spider Human", 2, 2)
	g.RunStateChecksForTest()
	if !effectstest.OnBoard(g, a) || !effectstest.OnBoard(g, b) || effectstest.LegendPrompt(g, me.ID) != nil {
		t.Fatal("two legendary Spiders with one name did not both stay")
	}
	effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	effectstest.B12Creature(g, opp.ID, "Test Spider", "Legendary Creature — Spider Human", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Test Spider", "Legendary Creature — Spider Human", 2, 2)
	g.RunStateChecksForTest()
	if effectstest.LegendPrompt(g, me.ID) == nil {
		t.Error("a legendary non-Spider pair was exempted")
	}
	if effectstest.LegendPrompt(g, opp.ID) == nil {
		t.Error("the opponent's Spiders were exempted")
	}
}

func TestSpmB18SpiderVerseCopiesASpellFromTheGraveyardOnceAndTheCopyHasHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	effectstest.B12Push(g, me.ID, "Spider-Verse", "Enchantment", effectstest.SpmB18VerseOracle, 0, 0)
	first := effectstest.MayhemHandCard(g, me, "Spider-Islanders", "Creature — Spider Horror Citizen", "{3}{R}", spm8IslandersOracle)
	effectstest.MayhemDiscard(t, g, me, first)
	effectstest.FloatForTest(g, me, "CR")
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	all, tokens := spmB18Count(g, me.ID, "Spider-Islanders")
	if all != 2 || tokens != 1 {
		t.Fatalf("%d Spider-Islanders (%d tokens), want the card and one token copy", all, tokens)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Spider-Islanders" {
			continue
		}
		if haste := effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "haste"); haste != c.IsToken() {
			t.Errorf("token %v has haste %v; only the copy gains haste", c.IsToken(), haste)
		}
	}

	// Only once each turn.
	second := effectstest.MayhemHandCard(g, me, "Spider-Islanders", "Creature — Spider Horror Citizen", "{3}{R}", spm8IslandersOracle)
	effectstest.MayhemDiscard(t, g, me, second)
	effectstest.FloatForTest(g, me, "CR")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err != nil {
		t.Fatalf("second mayhem cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if all, tokens := spmB18Count(g, me.ID, "Spider-Islanders"); all != 3 || tokens != 1 {
		t.Errorf("%d Spider-Islanders (%d tokens) after the second cast, want 3 and still 1", all, tokens)
	}
}

func TestSpmB18SpiderVerseIgnoresSpellsFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Spider-Verse", "Enchantment", effectstest.SpmB18VerseOracle, 0, 0)
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if all, tokens := spmB18Count(g, me.ID, "Grizzly Bears"); all != 1 || tokens != 0 {
		t.Errorf("%d Grizzly Bears (%d tokens), want only the cast one", all, tokens)
	}
}

// --- Spider-Woman, Stunning Savior ---

func TestSpmB18SpiderWomanTapsOpposingArtifactsAndCreaturesOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Spider-Woman, Stunning Savior", "Legendary Creature — Spider Human Hero", effectstest.SpmB18WomanOracle, 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID, "flying") {
		t.Error("Spider-Woman lacks flying")
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("2/2 green Bear"), 1)
		_ = g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 1)
		_ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 green Bear"), 1)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sym := effectstest.B12Push(g, me.ID, "Symbiote Spider-Man", effectstest.SpmB18SymbioteType, effectstest.SpmB18SymbioteOracle, 2, 4)
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AttackWith(t, g, opp.ID, sym)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sym := effectstest.SpmB18Graveyard(g, me, game.Card{Name: "Symbiote Spider-Man", TypeLine: effectstest.SpmB18SymbioteType,
		OracleID: effectstest.SpmB18SymbioteOracle, ManaCost: "{2}{U/B}", Power: 2, Toughness: 4})
	bear := effectstest.B12Creature(g, me.ID, "Host Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CCB")
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
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, bear) != 1 {
		t.Fatalf("counters %d, want 1", effectstest.PlusOnes(g, bear))
	}
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.NextTurnOfMine(t, g)
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 3 {
		t.Fatalf("the host's look: %+v, want three cards for its 3 damage", pick)
	}
}

func TestSpmB18SymbioteFindNewHostOnlyAsASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	sym := effectstest.SpmB18Graveyard(g, me, game.Card{Name: "Symbiote Spider-Man", TypeLine: effectstest.SpmB18SymbioteType,
		OracleID: effectstest.SpmB18SymbioteOracle, ManaCost: "{2}{U/B}", Power: 2, Toughness: 4})
	bear := effectstest.B12Creature(g, me.ID, "Host Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CCU")
	if err := g.ActivateCatalogAbility(me.ID, sym, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("Find New Host was activated in the upkeep")
	}
}
