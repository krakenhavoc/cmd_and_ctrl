package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// text_box_test.go — exchanging text boxes (CR 612.5, ADR 0160):
// Exchange of Words and Deadpool, Trading Card, and the engine shape
// under them.

const (
	exchangeOfWordsOracle = "a55ef412-257b-4ecd-9440-4545521be938"
	tbSerraOracle         = "4b7ac066-e5c7-43e6-9e7e-2739b24a905d"
	tbLordOracle          = "cc7f290f-ca00-4285-9bdb-4b4402444f30"
)

// tbPut puts a creature with printed stats onto the battlefield with an
// entry stamp.
func tbPut(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Keywords: keywords,
		Owner: owner, Controller: owner,
	})
}

func tbPutDeadpool(g *game.Game, owner uuid.UUID) uuid.UUID {
	return tbPut(g, owner, "Deadpool, Trading Card", "Legendary Creature — Mutant Mercenary Hero", deadpoolOracle, 5, 3)
}

func tbPutSerra(g *game.Game, owner uuid.UUID) uuid.UUID {
	return tbPut(g, owner, "Serra Angel", "Creature — Angel", tbSerraOracle, 4, 4)
}

// tbExchange makes the exchange directly, as a resolving effect would.
func tbExchange(t *testing.T, g *game.Game, a, b uuid.UUID, d game.Duration) {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() { ok = g.ExchangeTextBoxesForEffect(uuid.Nil, a, b, d, "test exchange") })
	if !ok {
		t.Fatal("the exchange did not happen")
	}
}

// tbCastExchangeOfWords casts Exchange of Words and answers its trigger
// with the two creatures.
func tbCastExchangeOfWords(t *testing.T, g *game.Game, a, b uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	eow := castCatalogSpell(t, g, "Exchange of Words", "Enchantment", exchangeOfWordsOracle, nil)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	return eow
}

func tbActivatedLabels(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	var out []string
	for _, a := range game.ActivatedAbilitiesForCard(gaLayered(t, g, id)) {
		out = append(out, a.Label)
	}
	return out
}

const deadpoolSacLabel = "{3}, Sacrifice this creature: Each other player draws a card."

// The exchange both ways: keywords, activated rows, triggers — and the
// printed characteristics stay put (CR 612.5).
func TestExchangeOfWordsSwapsTheTextBothWays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serra := tbPutSerra(g, me.ID)
	dp := tbPutDeadpool(g, opp.ID)
	tbCastExchangeOfWords(t, g, serra, dp)

	if hasEffectiveKeyword(t, g, serra, "flying") || hasEffectiveKeyword(t, g, serra, "vigilance") {
		t.Error("Serra Angel gave her text away, so she has no flying or vigilance")
	}
	if !hasEffectiveKeyword(t, g, dp, "flying") || !hasEffectiveKeyword(t, g, dp, "vigilance") {
		t.Error("Deadpool has Serra Angel's text, so he has flying and vigilance")
	}
	if got := tbActivatedLabels(t, g, serra); !slices.Equal(got, []string{deadpoolSacLabel}) {
		t.Errorf("Serra Angel's activated rows = %v, want Deadpool's", got)
	}
	if got := tbActivatedLabels(t, g, dp); len(got) != 0 {
		t.Errorf("Deadpool's activated rows = %v, want none", got)
	}
	s, d := gaLayered(t, g, serra), gaLayered(t, g, dp)
	if s.Name != "Serra Angel" || s.CurrentPower() != 4 || s.CurrentToughness() != 4 {
		t.Errorf("Serra Angel keeps her name and 4/4; got %q %d/%d", s.Name, s.CurrentPower(), s.CurrentToughness())
	}
	if d.Name != "Deadpool, Trading Card" || d.CurrentPower() != 5 || d.CurrentToughness() != 3 {
		t.Errorf("Deadpool keeps his name and 5/3; got %q %d/%d", d.Name, d.CurrentPower(), d.CurrentToughness())
	}
	if s.TextBoxFrom() != "Deadpool, Trading Card" || d.TextBoxFrom() != "Serra Angel" {
		t.Errorf("text_box_from = %q / %q", s.TextBoxFrom(), d.TextBoxFrom())
	}

	// The upkeep trigger is Serra's now, so it costs ME the life, and
	// Deadpool's controller nothing.
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if opp.Life != 40 {
		t.Errorf("Deadpool's controller lost life on their upkeep (%d); his text is gone", opp.Life)
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != 37 {
		t.Errorf("Serra's controller life = %d after their upkeep, want 37", me.Life)
	}
}

// The activated row is offered to the HOST's controller, the bot
// enumerates it, and it resolves as Deadpool's ability with Serra as
// "this creature".
func TestExchangedActivatedAbilityIsOfferedAndResolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serra := tbPutSerra(g, me.ID)
	dp := tbPutDeadpool(g, opp.ID)
	tbExchange(t, g, serra, dp, game.IndefiniteDuration())
	advanceToMain(t, g)
	for i := 0; i < 3; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	found := false
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == "activate_ability" && m.Source == serra {
			found = true
		}
		if m.Type == "activate_ability" && m.Source == dp {
			t.Error("Deadpool's controller must not be offered an ability his text no longer has")
		}
	}
	if !found {
		t.Fatal("the enumerator should offer Serra Angel's exchanged sacrifice ability")
	}
	hands := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i] = p.Hand.Size()
	}
	if err := g.ActivateCatalogAbility(me.ID, serra, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(serra) {
		t.Error("Serra Angel was the sacrifice")
	}
	if me.Hand.Size() != hands[0] {
		t.Error("the activator is not an 'other player'")
	}
	for i := 1; i < len(g.Seats); i++ {
		if g.Seats[i].Hand.Size() != hands[i]+1 {
			t.Errorf("seat %d should have drawn a card", i)
		}
	}
}

// A static in the exchanged text applies from its NEW holder: Lord of
// Atlantis's "other Merfolk" now pumps the Lord itself.
func TestExchangedStaticAppliesFromItsNewHolder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lord := tbPut(g, me.ID, "Lord of Atlantis", "Creature — Merfolk", tbLordOracle, 2, 2)
	bear := tbPut(g, me.ID, "Grizzly Bears", "Creature — Bear", "test-2779-bears", 2, 2)
	if p := effectivePower(t, g, lord); p != 2 {
		t.Fatalf("Lord of Atlantis before = %d power, want 2", p)
	}
	tbExchange(t, g, lord, bear, game.IndefiniteDuration())
	if p := effectivePower(t, g, lord); p != 3 {
		t.Errorf("Lord of Atlantis after = %d power, want 3 (the Bears' new text pumps other Merfolk)", p)
	}
	if !hasEffectiveKeyword(t, g, lord, "islandwalk") {
		t.Error("the Lord is an 'other Merfolk' to the Bears now and has islandwalk")
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("the Bears are not Merfolk and get nothing; power = %d", p)
	}
}

// One creature leaves: it takes its half with it, and the other keeps
// its new text until Exchange of Words goes; then it has its own again.
func TestExchangeOfWordsWhenOneLeavesAndWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serra := tbPutSerra(g, me.ID)
	dp := tbPutDeadpool(g, opp.ID)
	eow := tbCastExchangeOfWords(t, g, serra, dp)
	if !g.Battlefield.Contains(eow) {
		t.Fatal("Exchange of Words should be on the battlefield")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(serra); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, dp, "flying") {
		t.Error("Deadpool keeps Serra's text after she leaves, while the enchantment remains")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(eow); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, dp, "flying") {
		t.Error("the exchange ends with Exchange of Words")
	}
	if got := tbActivatedLabels(t, g, dp); !slices.Equal(got, []string{deadpoolSacLabel}) {
		t.Errorf("Deadpool has his own text again; rows = %v", got)
	}
	if d := gaLayered(t, g, dp); d.TextBoxFrom() != "" {
		t.Error("no text_box_from once the exchange ends")
	}
}

// CR 701.12a: a target gone by resolution means no part of the exchange.
func TestExchangeOfWordsDoesNothingIfATargetLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serra := tbPutSerra(g, me.ID)
	dp := tbPutDeadpool(g, opp.ID)
	castCatalogSpell(t, g, "Exchange of Words", "Enchantment", exchangeOfWordsOracle, nil)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, serra, dp)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(serra); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, dp, "flying") {
		t.Error("half an exchange happened")
	}
	if got := tbActivatedLabels(t, g, dp); !slices.Equal(got, []string{deadpoolSacLabel}) {
		t.Errorf("Deadpool keeps his own text; rows = %v", got)
	}
}

// The same when a target is still there but no longer a creature: it is
// an illegal target (CR 608.2b), and half an exchange is still none.
func TestExchangeOfWordsDoesNothingIfATargetStoppedBeingACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serra := tbPutSerra(g, me.ID)
	dp := tbPutDeadpool(g, opp.ID)
	castCatalogSpell(t, g, "Exchange of Words", "Enchantment", exchangeOfWordsOracle, nil)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, serra, dp)
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(serra), []game.Mod{game.RemoveTypesMod("Creature")},
			game.IndefiniteDuration(), "test not a creature") {
			t.Fatal("type change not registered")
		}
		// As every mutator does before it lets go of the lock.
		g.RecomputeLayersIfStaleLocked()
	})
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, dp, "flying") {
		t.Error("half an exchange happened with an illegal target")
	}
}

// A copy of an exchanged creature copies its PRINTED text (CR 707.2:
// text-changing effects are not copied).
func TestCloneOfAnExchangedCreatureCopiesThePrintedText(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	serra := tbPutSerra(g, me.ID)
	bear := tbPut(g, me.ID, "Grizzly Bears", "Creature — Bear", "test-2779-bears", 2, 2)
	tbExchange(t, g, serra, bear, game.IndefiniteDuration())
	cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, serra)
	if !hasEffectiveKeyword(t, g, cloneID, "flying") {
		t.Error("the Clone copies Serra Angel's printed text, flying included")
	}
	if hasEffectiveKeyword(t, g, serra, "flying") {
		t.Error("Serra herself still has the Bears' text")
	}
}

// Layer 1 then layer 3: a creature that becomes a copy of something
// AFTER its text was exchanged keeps the exchanged text.
func TestDurationCopyAfterTheExchangeKeepsTheExchangedText(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	serra := tbPutSerra(g, me.ID)
	bear := tbPut(g, me.ID, "Grizzly Bears", "Creature — Bear", "test-2779-bears", 2, 2)
	dp := tbPutDeadpool(g, me.ID)
	tbExchange(t, g, serra, bear, game.IndefiniteDuration())
	g.WithWriteLock(func() {
		v, _ := g.CopiableValuesForEffect(dp)
		if !g.BecomeCopyForEffect(uuid.Nil, dp, []uuid.UUID{bear}, v, game.IndefiniteDuration(), "test copy") {
			t.Fatal("the copy did not happen")
		}
	})
	b := gaLayered(t, g, bear)
	if b.Name != "Deadpool, Trading Card" {
		t.Fatalf("the Bears became a copy of Deadpool; name = %q", b.Name)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("layer 3 follows layer 1: the copy still has Serra Angel's exchanged text")
	}
	if got := tbActivatedLabels(t, g, bear); len(got) != 0 {
		t.Errorf("the copied Deadpool text is overridden by the exchange; rows = %v", got)
	}
}

// Layer 6 after layer 3: an ability-removing effect takes the exchanged
// abilities too.
func TestAbilityLossAfterTheExchangeRemovesTheExchangedText(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	serra := tbPutSerra(g, me.ID)
	dp := tbPutDeadpool(g, me.ID)
	tbExchange(t, g, serra, dp, game.IndefiniteDuration())
	g.WithWriteLock(func() {
		if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(serra), []game.Mod{game.LoseAllAbilitiesMod()},
			game.IndefiniteDuration(), "test loses all abilities") {
			t.Fatal("removal not registered")
		}
	})
	if got := tbActivatedLabels(t, g, serra); len(got) != 0 {
		t.Errorf("Serra lost all abilities, the exchanged ones included; rows = %v", got)
	}
	if !hasEffectiveKeyword(t, g, dp, "flying") {
		t.Error("the removal applies to Serra only")
	}
}

// CR 603.10a: an exchanged creature that dies with a dies trigger in its
// GAINED text triggers; the one that gave it away does not.
func TestExchangedDiesTriggerFiresFromLastKnownText(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fam := tbPut(g, me.ID, "Filigree Familiar", "Artifact Creature — Fox", filigreeFamiliarOracle, 2, 2)
	bear := tbPut(g, me.ID, "Grizzly Bears", "Creature — Bear", "test-2779-bears", 2, 2)
	tbExchange(t, g, fam, bear, game.IndefiniteDuration())
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(fam); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("the Familiar gave its dies trigger away and draws nothing")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the Bears died with the Familiar's text and draw a card; hand %d -> %d", hand, me.Hand.Size())
	}
}

// The importer's keywords (Card.Keywords) are rules text too.
func TestImportedKeywordsTravelWithTheText(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trampler := tbPut(g, me.ID, "Trampling Beast", "Creature — Beast", "test-2779-beast", 3, 3, "trample")
	bear := tbPut(g, me.ID, "Grizzly Bears", "Creature — Bear", "test-2779-bears", 2, 2)
	tbExchange(t, g, trampler, bear, game.IndefiniteDuration())
	if hasEffectiveKeyword(t, g, trampler, "trample") || !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("trample moves with the text box")
	}
}

// A token's text is its template's, reached through its token key; the
// ability list it carries on the instance is that same text, so it
// stops answering once the text has gone elsewhere.
func TestExchangedTokenCarriedAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// The carried slice, as a restore re-derives it from the token key.
	tmpl := FoodToken()
	tmpl.ActivatedAbilities = game.ActivatedAbilitiesForCard(tmpl)
	food := pushToken(g, me.ID, tmpl)
	bear := tbPut(g, me.ID, "Grizzly Bears", "Creature — Bear", "test-2779-bears", 2, 2)
	want := tbActivatedLabels(t, g, food)
	if len(want) != 1 {
		t.Fatalf("a Food has one activated ability; got %v", want)
	}
	tbExchange(t, g, food, bear, game.IndefiniteDuration())
	if got := tbActivatedLabels(t, g, food); len(got) != 0 {
		t.Errorf("the Food gave its text away; rows = %v", got)
	}
	if got := tbActivatedLabels(t, g, bear); !slices.Equal(got, want) {
		t.Errorf("the Bears have the Food's text; rows = %v, want %v", got, want)
	}
}

// Undo and restore: the record is data and the result is rebuilt.
func TestTextBoxExchangeSurvivesCloneAndSnapshot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	serra := tbPutSerra(g, me.ID)
	dp := tbPutDeadpool(g, me.ID)
	tbExchange(t, g, serra, dp, game.IndefiniteDuration())
	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	for name, h := range map[string]*game.Game{"clone": g.Clone(), "restore": restored} {
		if !hasEffectiveKeyword(t, h, dp, "flying") || hasEffectiveKeyword(t, h, serra, "flying") {
			t.Errorf("%s: the exchanged keywords did not survive", name)
		}
		if got := tbActivatedLabels(t, h, serra); !slices.Equal(got, []string{deadpoolSacLabel}) {
			t.Errorf("%s: Serra's rows = %v", name, got)
		}
	}
}

// Deadpool end to end: the entry prompt is a text exchange, his enters
// trigger is the gained text's (CR 603.6a), and the other creature gets
// his abilities.
func TestDeadpoolExchangesAsHeEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fam := tbPut(g, me.ID, "Filigree Familiar", "Artifact Creature — Fox", filigreeFamiliarOracle, 2, 2)
	dp := castCreatureSpellForCopy(t, g, "Deadpool, Trading Card", "Legendary Creature — Mutant Mercenary Hero", deadpoolOracle, 5, 3)
	life := me.Life
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	p := copyPrompt(g)
	if p == nil {
		t.Fatal("Deadpool asks as he enters")
	}
	if !p.ExchangesText || !slices.Contains(p.CopyOptions, fam) {
		t.Fatalf("the prompt should be a text exchange offering the Familiar: %+v", p)
	}
	resolveWithCopyChoice(t, g, fam)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("Deadpool entered with the Familiar's text and gains 2 life; life %d -> %d", life, me.Life)
	}
	if got := tbActivatedLabels(t, g, fam); !slices.Equal(got, []string{deadpoolSacLabel}) {
		t.Errorf("the Familiar has Deadpool's sacrifice ability; rows = %v", got)
	}
	if got := len(game.ManaAbilitiesForCard(gaLayered(t, g, dp))); got != 1 {
		t.Errorf("Deadpool has the Familiar's mana ability; %d rows", got)
	}
	if d := gaLayered(t, g, dp); d.CurrentPower() != 5 || d.TextBoxFrom() != "Filigree Familiar" {
		t.Errorf("Deadpool is still a 5/3 with the Familiar's text; got %d power, from %q", d.CurrentPower(), d.TextBoxFrom())
	}
}

// Declining is a legal answer: he keeps his own text.
func TestDeadpoolMayDeclineTheExchange(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fam := tbPut(g, me.ID, "Filigree Familiar", "Artifact Creature — Fox", filigreeFamiliarOracle, 2, 2)
	dp := castCreatureSpellForCopy(t, g, "Deadpool, Trading Card", "Legendary Creature — Mutant Mercenary Hero", deadpoolOracle, 5, 3)
	resolveWithCopyChoice(t, g, uuid.Nil)
	if got := tbActivatedLabels(t, g, dp); !slices.Equal(got, []string{deadpoolSacLabel}) {
		t.Errorf("Deadpool keeps his own text; rows = %v", got)
	}
	if got := tbActivatedLabels(t, g, fam); len(got) != 0 {
		t.Errorf("the Familiar keeps its own; rows = %v", got)
	}
}
