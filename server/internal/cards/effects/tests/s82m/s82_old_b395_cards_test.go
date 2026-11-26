package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB395Baleful   = "0930a976-f69e-44a8-8b19-72095f20ed13"
	oldB395Heliod    = "182c9af9-d867-4848-ab36-ac7b1914f41a"
	oldB395Lampad    = "6ead91e1-4062-4080-bbc7-a4b62f8c6111"
	oldB395Erebos    = "dc39c52b-9bee-4cec-a970-1867314944ef"
	oldB395Thassa    = "fcb97d06-e0b8-4a13-ab7c-1304ccb7ace2"
	oldB395Reaper    = "10494ce1-7a38-44ae-9881-a83e44daaf7b"
	oldB395Steed     = "5acf1424-6886-4777-be9c-a845502e08bd"
	oldB395Marauder  = "7a8bfc6b-0a24-42ad-a2c5-ac8d29d4c1b9"
	oldB395Catoblep  = "c6f68e4b-af2b-43d5-8052-104defe7f3ec"
	oldB395Disciple  = "d7cbe54b-f217-45f5-8583-c254a1c80bb8"
	oldB395EnchType  = "Enchantment Creature — Test"
	oldB395BearType  = "Creature — Bear"
	oldB395BlackCost = "{B}{B}"
)

var oldB395Cards = map[string]string{
	oldB395Baleful: "Baleful Eidolon", oldB395Heliod: "Heliod's Emissary", oldB395Lampad: "Cavern Lampad",
	oldB395Erebos: "Erebos's Emissary", oldB395Thassa: "Thassa's Emissary", oldB395Reaper: "Reaper of the Wilds",
	oldB395Steed: "Fleshmad Steed", oldB395Marauder: "Mogis's Marauder", oldB395Catoblep: "Loathsome Catoblepas",
	oldB395Disciple: "Disciple of Phenax",
}

func TestOldB395AreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range oldB395Cards {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB395BestowedBonusesAndGrantsOnTheHost(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		wantP, wantT int
		keyword      string
	}{
		{"Baleful Eidolon", oldB395Baleful, 1, 1, 3, 3, "deathtouch"},
		{"Cavern Lampad", oldB395Lampad, 2, 2, 4, 4, "intimidate"},
		{"Heliod's Emissary", oldB395Heliod, 3, 3, 5, 5, ""},
		{"Erebos's Emissary", oldB395Erebos, 3, 3, 5, 5, ""},
		{"Thassa's Emissary", oldB395Thassa, 3, 3, 5, 5, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			card := effectstest.BestowCard(t, g, tc.name, oldB395EnchType, tc.oracle, tc.power, tc.tough)
			bear := effectstest.B12Creature(g, me.ID, "Bear", oldB395BearType, 2, 2)
			effectstest.CastBestowed(t, g, card, bear)
			effectstest.PassPriorityAroundTable(t, g)
			if s := effectstest.ReadBestowState(g, card); !s.OnBattlefield || s.AttachedTo != bear || s.Creature {
				t.Fatalf("bestowed %s is not an Aura on the bear: %+v", tc.name, s)
			}
			if p, tough := effectstest.PtOf(t, g, bear); p != tc.wantP || tough != tc.wantT {
				t.Errorf("the bear is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
			if tc.keyword != "" && !effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the bear has no %s", tc.keyword)
			}
		})
	}
}

func TestOldB395EmissariesAreCreaturesWithTheirOwnKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	baleful := effectstest.B12Push(g, me.ID, "Baleful Eidolon", oldB395EnchType, oldB395Baleful, 1, 1)
	lampad := effectstest.B12Push(g, me.ID, "Cavern Lampad", oldB395EnchType, oldB395Lampad, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, baleful, "deathtouch") || !effectstest.HasEffectiveKeyword(t, g, lampad, "intimidate") {
		t.Error("the creatures should keep their own deathtouch and intimidate")
	}
}

func TestOldB395HeliodsEmissaryTapsACreatureWhenItOrItsHostAttacks(t *testing.T) {
	// The creature itself attacks.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elk := effectstest.B12Push(g, me.ID, "Heliod's Emissary", oldB395EnchType, oldB395Heliod, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB395BearType, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB395BearType, 2, 2)
	effectstest.AttackWithTrigger(t, g, elk, opp.ID, victim)
	if !effectstest.B16Tapped(t, g, victim) {
		t.Error("the creature attacked but the target was not tapped")
	}
	if effectstest.B16Tapped(t, g, mine) {
		t.Error("an unrelated creature of mine was tapped")
	}

	// Bestowed, its host attacks.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	card := effectstest.BestowCard(t, g, "Heliod's Emissary", oldB395EnchType, oldB395Heliod, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB395BearType, 2, 2)
	victim = effectstest.B12Creature(g, opp.ID, "Victim", oldB395BearType, 2, 2)
	effectstest.CastBestowed(t, g, card, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AttackWithTrigger(t, g, bear, opp.ID, victim)
	if !effectstest.B16Tapped(t, g, victim) {
		t.Error("the host attacked but the target was not tapped")
	}
}

func TestOldB395ErebossEmissaryPumpsItselfOrItsHostForADiscardedCreature(t *testing.T) {
	// As a creature it pumps itself.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Erebos's Emissary", oldB395EnchType, oldB395Erebos, 3, 3)
	fodder := effectstest.HandCardOf(g, me, "Fodder", oldB395BearType, "", "")
	spell := effectstest.HandCardOf(g, me, "Spell", "Sorcery", "", "")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spell}, Strict: true}); err == nil {
		t.Fatal("discarded a noncreature card")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, id); p != 5 || tough != 5 {
		t.Errorf("the Emissary is %d/%d, want 5/5", p, tough)
	}
	if !me.Graveyard.Contains(fodder) {
		t.Error("the creature card was not discarded")
	}

	// Bestowed it pumps the enchanted creature instead.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	card := effectstest.BestowCard(t, g, "Erebos's Emissary", oldB395EnchType, oldB395Erebos, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB395BearType, 2, 2)
	effectstest.CastBestowed(t, g, card, bear)
	effectstest.PassPriorityAroundTable(t, g)
	fodder = effectstest.HandCardOf(g, me, "Fodder", oldB395BearType, "", "")
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate bestowed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("the bear is %d/%d, want 7/7 (+3/+3 and +2/+2)", p, tough)
	}
}

func TestOldB395ThassasEmissaryDrawsOnCombatDamageByItOrItsHost(t *testing.T) {
	// The creature itself connects.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	crab := effectstest.B12Push(g, me.ID, "Thassa's Emissary", oldB395EnchType, oldB395Thassa, 3, 3)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, crab)
	effectstest.InitSettle(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}

	// Bestowed, its host connects.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	card := effectstest.BestowCard(t, g, "Thassa's Emissary", oldB395EnchType, oldB395Thassa, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB395BearType, 2, 2)
	effectstest.CastBestowed(t, g, card, bear)
	effectstest.PassPriorityAroundTable(t, g)
	hand = me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.InitSettle(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("host connected: hand %d → %d, want +1", hand, me.Hand.Size())
	}
	// A creature that is not the host does not draw.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Thassa's Emissary", oldB395EnchType, oldB395Thassa, 3, 3)
	other := effectstest.B12Creature(g, me.ID, "Other", oldB395BearType, 2, 2)
	hand = me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, other)
	effectstest.InitSettle(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("another creature connected: hand %d → %d, want no draw", hand, me.Hand.Size())
	}
}

func TestOldB395ReaperOfTheWildsScriesWhenAnotherCreatureDiesAndGainsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reaper := effectstest.B12Push(g, me.ID, "Reaper of the Wilds", "Creature — Gorgon", oldB395Reaper, 4, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB395BearType, 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("another creature died but there is no scry prompt")
	}
	effectstest.B26AnswerScryAllTop(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, reaper, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, reaper, "deathtouch") {
		t.Error("no deathtouch after {B}")
	}
	effectstest.FillPoolColored(me, "G", 2)
	if err := g.ActivateCatalogAbility(me.ID, reaper, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{1}{G}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, reaper, "hexproof") {
		t.Error("no hexproof after {1}{G}")
	}
}

func TestOldB395ReaperOfTheWildsDoesNotScryForItsOwnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	reaper := effectstest.B12Push(g, me.ID, "Reaper of the Wilds", "Creature — Gorgon", oldB395Reaper, 4, 5)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(reaper) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Error("the Reaper scried for its own death")
	}
}

func TestOldB395FleshmadSteedTapsWhenAnotherCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	steed := effectstest.B12Push(g, me.ID, "Fleshmad Steed", "Creature — Horse", oldB395Steed, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB395BearType, 2, 2)
	if effectstest.B16Tapped(t, g, steed) {
		t.Fatal("the Steed starts tapped")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, steed) {
		t.Error("the Steed did not tap when another creature died")
	}
}

// oldB395Marauder casts Mogis's Marauder with its printed {2}{B} cost so its
// own {B} counts toward devotion.
func oldB395CastMarauder(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mogis's Marauder", TypeLine: "Creature — Human Berserker", OracleID: oldB395Marauder,
		ManaCost: "{2}{B}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 3)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB395MogissMarauderGrantsIntimidateAndHasteToUpToDevotionTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	// {B}{B} of devotion already on the board, plus the Marauder's own {B}: X is 3.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dark Rite", TypeLine: "Enchantment", ManaCost: oldB395BlackCost, Owner: me.ID, Controller: me.ID})
	a := effectstest.B12Creature(g, me.ID, "A", oldB395BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", oldB395BearType, 2, 2)
	oldB395CastMarauder(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for no targets")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
		t.Fatalf("two targets with devotion 3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "intimidate") || !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("creature %s lacks intimidate and haste", id)
		}
	}
}

func TestOldB395MogissMarauderTargetCountIsBoundedByDevotion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "A", oldB395BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", oldB395BearType, 2, 2)
	oldB395CastMarauder(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for no targets")
	}
	// Devotion is 1 (the Marauder's own {B}): two targets are too many.
	if err := g.ResolvePickTargets(p.ID, me.ID,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err == nil {
		t.Fatal("two targets were accepted with devotion 1")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: a}); err != nil {
		t.Fatalf("one target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, a, "intimidate") || effectstest.HasEffectiveKeyword(t, g, b, "intimidate") {
		t.Error("only the chosen creature should gain intimidate")
	}
}

func TestOldB395LoathsomeCatoblepasMustBeBlockedAndShrinksOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	beast := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loathsome Catoblepas", TypeLine: "Creature — Beast", OracleID: oldB395Catoblep,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID, Keywords: []string{"haste"},
	})
	effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "G", 3)
	if err := g.ActivateCatalogAbility(me.ID, beast, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ReqToBlockers(t, g, beast)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got == "" {
		t.Error("the refusal has no sentence")
	}
}

func TestOldB395LoathsomeCatoblepasDeathGivesAnOpponentsCreatureMinusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := effectstest.B12Push(g, me.ID, "Loathsome Catoblepas", "Creature — Beast", oldB395Catoblep, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB395BearType, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB395BearType, 4, 4)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(beast) })
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the death trigger asks for no target")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("targeted my own creature")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
		t.Fatalf("target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tough := effectstest.PtOf(t, g, theirs); pw != 1 || tough != 1 {
		t.Errorf("their creature is %d/%d, want 1/1", pw, tough)
	}
}

func TestOldB395DiscipleOfPhenaxRevealsDevotionManyAndDiscardsTheChosenOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp.Hand.Cards = nil
	var hand []uuid.UUID
	for i := 0; i < 4; i++ {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Their Card", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
		hand = append(hand, id)
	}
	disciple := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: disciple, Name: "Disciple of Phenax", TypeLine: "Creature — Human Cleric", OracleID: oldB395Disciple,
		ManaCost: "{2}{B}{B}", Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, disciple, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// Devotion to black is 2, the Disciple's own {B}{B}.
	ask := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if ask == nil || ask.ChooseMin != 2 || ask.ChooseMax != 2 {
		t.Fatalf("the target player chooses exactly 2 cards to reveal, got %+v", ask)
	}
	reveal := hand[:2]
	effectstest.AnswerChooseCards(t, g, opp.ID, reveal...)
	rp := effectstest.RevealPickFor(g, me.ID)
	if rp == nil || len(rp.ChooseCards) != 2 {
		t.Fatalf("want 2 revealed cards for the Disciple's controller, got %+v", rp)
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{hand[3]}); err == nil {
		t.Fatal("picked a card that was not revealed")
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{reveal[1]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if !opp.Graveyard.Contains(reveal[1]) || opp.Hand.Size() != 3 {
		t.Errorf("the chosen card was not discarded (hand %d)", opp.Hand.Size())
	}
}
