package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sami Whammy slice sami-a (home tracker #2190): the commander and
// the legends.

const (
	samiOracle        = "c4175a34-70bb-47e5-8406-efdc4d2cd079"
	birgiOracle       = "fb81e4d3-1d8c-4779-be62-87cf49277e51"
	darettiOracle     = "0eb1d539-6301-46b0-903e-487eda34253f"
	tezzeretOracle    = "0eb11b2d-a397-48da-a0f3-9f0f83f42282"
	mightstoneOracle  = "c396db03-bf11-4e20-b630-4f9aa8fd78da"
	echoesOracle      = "23a79523-4be0-4d17-80aa-5ea024cb3463"
	allFatesOracle    = "cef5361a-a189-4e05-a8b8-fc764aed73b9"
	nexusBecomeOracle = "07142ee9-dc2c-4b33-ad20-2f5285225e86"
	guidelightOracle  = "98005890-d566-4a27-906f-625513171e85"
)

func samiArtifact(g *game.Game, owner uuid.UUID, name, cost string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Artifact", ManaCost: cost,
		Owner: owner, Controller: owner,
	})
}

func TestSamiGivesYourSpellsAffinityForArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{3}{U}"); got != 4 {
		t.Fatalf("no Sami: %d, want 4", got)
	}
	pushCatalogPermanent(g, me.ID, "Sami, Wildcat Captain", "Legendary Creature — Human Artificer Rogue", samiOracle, false)
	// Sami alone: no artifacts, no discount.
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{3}{U}"); got != 4 {
		t.Errorf("Sami with no artifacts: %d, want 4", got)
	}
	samiArtifact(g, me.ID, "Rock One", "{2}")
	samiArtifact(g, me.ID, "Rock Two", "{2}")
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{3}{U}"); got != 2 {
		t.Errorf("two artifacts: %d, want 2", got)
	}
	// Any spell, any type, and never below the coloured part.
	if got := priceInHand(t, g, me, "Cheap Bear", "Creature — Bear", "{1}{G}"); got != 1 {
		t.Errorf("{1}{G} with two artifacts: %d, want 1 (generic only)", got)
	}
	// An opponent's spells, and an opponent's artifacts, do not count.
	if got := priceInHand(t, g, opp, "Divination", "Sorcery", "{3}{U}"); got != 4 {
		t.Errorf("an opponent's spell: %d, want 4", got)
	}
	samiArtifact(g, opp.ID, "Their Rock", "{2}")
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{3}{U}"); got != 2 {
		t.Errorf("an opponent's artifact counted: %d, want 2", got)
	}
}

func TestSamiHasDoubleStrikeAndVigilance(t *testing.T) {
	spec, ok := Lookup(samiOracle)
	if !ok || spec.Completeness != CompletenessFull {
		t.Fatal("Sami must be registered complete")
	}
	if len(spec.PrintedKeywords) != 2 {
		t.Errorf("keywords = %v, want double strike and vigilance", spec.PrintedKeywords)
	}
}

func TestBirgiAddsKeptRedManaWhenYouCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	spec, ok := Lookup(birgiOracle)
	if !ok || len(spec.Triggered) != 1 {
		t.Fatal("Birgi is not registered with its cast trigger")
	}
	if spec.Completeness != CompletenessCaveats {
		t.Error("Birgi's boast clause is not implemented; it must say so")
	}
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: me.ID}
		if err := spec.Triggered[0].Effect(g, item); err != nil {
			t.Fatalf("trigger effect: %v", err)
		}
	})
	mkWantPool(t, me, map[string]int{"R": 1})
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"R": 1})
}

func TestHarnfelDiscardExilesTopTwoAndLetsYouPlayThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	spec, ok := Lookup(birgiOracle + "#1")
	if !ok || len(spec.Activated) != 1 {
		t.Fatal("Harnfel's back face is not registered")
	}
	ids := sunbirdLibrary(me,
		game.Card{Name: "A Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "A Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Third", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
	)
	g.WithWriteLock(func() {
		if err := spec.Activated[0].Effect(g, &game.StackItem{Controller: me.ID}); err != nil {
			t.Fatalf("effect: %v", err)
		}
	})
	if !inExile(g, ids[0]) || !inExile(g, ids[1]) {
		t.Error("the top two cards are exiled")
	}
	if inExile(g, ids[2]) {
		t.Error("only two cards are exiled")
	}
	if len(me.CastPermissions) < 1 {
		t.Errorf("expected a play permission for the exiled cards, got %d", len(me.CastPermissions))
	}
}

func TestDarettiPowerIsTheGreatestArtifactManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	daretti := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Daretti, Rocketeer Engineer", OracleID: darettiOracle,
		TypeLine: "Legendary Creature — Goblin Artificer", ManaCost: "{4}{R}", Toughness: 5,
		Owner: me.ID, Controller: me.ID,
	})
	if got := effectivePower(t, g, daretti); got != 0 {
		t.Errorf("no artifacts: power %d, want 0", got)
	}
	samiArtifact(g, me.ID, "Small", "{1}")
	samiArtifact(g, me.ID, "Big", "{6}")
	if got := effectivePower(t, g, daretti); got != 6 {
		t.Errorf("power %d, want 6", got)
	}
	if got := effectiveToughness(t, g, daretti); got != 5 {
		t.Errorf("toughness %d, want 5", got)
	}
}

func darettiSetup(t *testing.T) (g *game.Game, me *game.Player, daretti, dead, fodder uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	daretti = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Daretti, Rocketeer Engineer", OracleID: darettiOracle,
		TypeLine: "Legendary Creature — Goblin Artificer", ManaCost: "{4}{R}", Toughness: 5,
		Owner: me.ID, Controller: me.ID,
	})
	dead = uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
	fodder = samiArtifact(g, me.ID, "Fodder", "{1}")
	return
}

func runDaretti(t *testing.T, g *game.Game, me *game.Player, daretti, dead uuid.UUID) {
	t.Helper()
	spec, _ := Lookup(darettiOracle)
	g.WithWriteLock(func() {
		item := &game.StackItem{
			Controller: me.ID, SourceCardID: daretti,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
		}
		if err := spec.Triggered[0].Effect(g, item); err != nil {
			t.Fatalf("effect: %v", err)
		}
	})
}

func TestDarettiSacrificesAnArtifactToReturnTheChosenCard(t *testing.T) {
	g, me, daretti, dead, fodder := darettiSetup(t)
	runDaretti(t, g, me, daretti, dead)
	answerMayChoice(t, g, me.ID, true)
	answerSacrifice(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Error("the chosen artifact card returns to the battlefield")
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed artifact is gone")
	}
}

func TestDarettiDeclinedOrWithoutAnArtifactReturnsNothing(t *testing.T) {
	g, me, daretti, dead, fodder := darettiSetup(t)
	runDaretti(t, g, me, daretti, dead)
	answerMayChoice(t, g, me.ID, false)
	if g.Battlefield.Contains(dead) || !g.Battlefield.Contains(fodder) {
		t.Error("declining sacrifices nothing and returns nothing")
	}

	g, me, daretti, dead, fodder = darettiSetup(t)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	runDaretti(t, g, me, daretti, dead)
	if len(g.PendingChoices) != 0 {
		t.Error("with no artifact to sacrifice there is no question")
	}
	if g.Battlefield.Contains(dead) {
		t.Error("nothing returns without a sacrifice")
	}
}

func tezzeretIn(g *game.Game, me *game.Player, loyalty int) uuid.UUID {
	return pushCatalogWalker(g, me.ID, "Tezzeret, Cruel Captain", tezzeretOracle, loyalty)
}

func TestTezzeretLoyaltyGrowsWhenAnArtifactEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tez := tezzeretIn(g, me, 4)
	castCatalogSpell(t, g, "Some Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, tez); got != 5 {
		t.Errorf("loyalty %d after an artifact entered, want 5", got)
	}
	castCatalogSpell(t, g, "Some Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, tez); got != 5 {
		t.Errorf("a nonartifact creature changed loyalty to %d", got)
	}
}

func TestTezzeretZeroUntapsAndGrowsAnArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	tez := tezzeretIn(g, me, 4)
	golem := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 2, Tapped: true, Owner: me.ID, Controller: me.ID,
	})
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Tapped: true, Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, tez, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: golem}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(golem); c.Tapped || countersOn(g, golem, game.CounterPlusOne) != 1 {
		t.Errorf("artifact creature: tapped=%v counters=%d, want untapped with one counter", c.Tapped, countersOn(g, golem, game.CounterPlusOne))
	}
	// Planeswalker abilities are once per turn; use a second walker turn state.
	g.WithWriteLock(func() { delete(g.LoyaltyActivatedThisTurn, tez) })
	if err := g.ActivateCatalogAbility(me.ID, tez, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate on the plain creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(bear); c.Tapped || countersOn(g, bear, game.CounterPlusOne) != 0 {
		t.Errorf("plain creature: tapped=%v counters=%d, want untapped and no counter", c.Tapped, countersOn(g, bear, game.CounterPlusOne))
	}
}

func TestTezzeretMinusThreeFindsACheapArtifact(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	tez := tezzeretIn(g, me, 4)
	cheap := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Sol Ring-ish", TypeLine: "Artifact", ManaCost: "{1}", Owner: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, tez, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		answerSearchByID(t, g, me.ID, cheap)
	}
	if !me.Hand.Contains(cheap) {
		t.Error("the mana value 1 artifact is in hand")
	}
	if got := loyaltyCount(g, tez); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

// tezzeretEmblemCombat gives me Tezzeret's emblem through the real -7
// and walks to my beginning of combat, resolving the emblem's trigger.
func tezzeretEmblemCombat(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	tez := tezzeretIn(g, me, 7)
	if err := g.ActivateCatalogAbility(me.ID, tez, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate -7: %v", err)
	}
	passPriorityAroundTable(t, g)
	for g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if len(g.PendingChoices) > 0 {
		answerFirstPickTarget(t, g)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
}

func TestTezzeretEmblemAnimatesANoncreatureArtifactAndGrowsIt(t *testing.T) {
	spec, ok := Lookup(tezzeretOracle)
	if !ok || spec.Emblem == nil || len(spec.Emblem.Triggered) != 1 {
		t.Fatal("Tezzeret's emblem is not declared")
	}
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := samiArtifact(g, me.ID, "Plain Rock", "{2}")
	tezzeretEmblemCombat(t, g, me)
	c := findBattlefieldCardForTest(g, rock)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Robot") {
		t.Errorf("the rock should be a Robot artifact creature, got %q", c.Effective().Types)
	}
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("0/0 with three counters is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestTezzeretEmblemOnAnArtifactCreatureOnlyAddsCounters(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	golem := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	tezzeretEmblemCombat(t, g, me)
	c := findBattlefieldCardForTest(g, golem)
	if c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("a 2/2 artifact creature is %d/%d, want 5/5", c.CurrentPower(), c.CurrentToughness())
	}
	if c.HasSubtype("Robot") {
		t.Error("an artifact that is already a creature is not made a Robot")
	}
}

func castMightstone(t *testing.T, g *game.Game, mode int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "The Mightstone and Weakstone", "Legendary Artifact — Powerstone", mightstoneOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt from the enters trigger: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func TestMightstoneAndWeakstoneDrawsTwo(t *testing.T) {
	spec, ok := Lookup(mightstoneOracle)
	if !ok || len(spec.Triggered) != 1 || spec.Triggered[0].Modes == nil || len(spec.ManaAbilities) != 1 {
		t.Fatal("The Mightstone and Weakstone is not wired with its modal trigger and mana ability")
	}
	if len(spec.ManaAbilities[0].Restrictions) != 1 {
		t.Error("the mana carries the nonartifact-spell restriction")
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	castMightstone(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("hand grew by %d, want 2", got)
	}
}

func TestMightstoneAndWeakstoneShrinksACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Victim", TypeLine: "Creature — Bear",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	castMightstone(t, g, 1)
	answerFirstPickTarget(t, g)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(victim) {
		t.Error("the 4/4 gets -5/-5 and dies")
	}
}

func TestEchoesDoublesColorlessTriggersAndCopiesColorlessSpells(t *testing.T) {
	spec, ok := Lookup(echoesOracle)
	if !ok || len(spec.TriggerDoublers) != 1 || len(spec.Triggered) != 1 {
		t.Fatal("Echoes of Eternity is not wired")
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	echoes := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Echoes of Eternity", OracleID: echoesOracle,
		TypeLine: "Kindred Enchantment — Eldrazi", ManaCost: "{3}{C}{C}{C}",
		Owner: me.ID, Controller: me.ID,
	})
	_ = echoes
	d := spec.TriggerDoublers[0]
	colorless := game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID}
	red := game.Card{InstanceID: uuid.New(), Name: "Bolt Bear", TypeLine: "Creature — Bear", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID}
	mine := game.Characteristic{Controller: me.ID}
	q := func(src game.Card, lki game.Characteristic, fromSpell bool) game.TriggerDoublingQuery {
		return game.TriggerDoublingQuery{
			Doubler: game.Card{InstanceID: echoes}, DoublerLKI: mine,
			Source: src, SourceLKI: lki, FromSpell: fromSpell, Ability: &game.TriggeredAbility{},
		}
	}
	if !d.Applies(g, q(colorless, mine, false)) {
		t.Error("a colorless permanent you control is doubled")
	}
	redLKI := game.Characteristic{Controller: me.ID, Colors: []string{"R"}}
	if d.Applies(g, q(red, redLKI, false)) {
		t.Error("a red permanent is not doubled")
	}
	if !d.Applies(g, q(colorless, mine, true)) {
		t.Error("a colorless spell you control is doubled")
	}
	if d.Applies(g, q(game.Card{InstanceID: echoes}, mine, false)) {
		t.Error("Echoes itself is excluded: it says another colorless permanent")
	}
	opp := game.Characteristic{Controller: g.Seats[1].ID}
	if d.Applies(g, q(colorless, opp, false)) {
		t.Error("an opponent's colorless permanent is not doubled")
	}
}

func TestAllFatesScrollDrawsForDifferentlyNamedLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spec, ok := Lookup(allFatesOracle)
	if !ok || len(spec.Activated) != 1 || len(spec.ManaAbilities) != 1 {
		t.Fatal("All-Fates Scroll is not wired")
	}
	for _, n := range []string{"Forest", "Forest", "Island", "Mountain"} {
		pushCatalogPermanent(g, me.ID, n, "Basic Land — "+n, "", false)
	}
	before := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := spec.Activated[0].Effect(g, &game.StackItem{Controller: me.ID}); err != nil {
			t.Fatalf("effect: %v", err)
		}
	})
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("drew %d, want 3 (Forest, Island, Mountain)", got)
	}
}

func TestNexusOfBecomingExilesACardAndMakesA3x3GolemCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: card, Name: "Fat Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{G}",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID})
	spec, _ := Lookup(nexusBecomeOracle)
	g.WithWriteLock(func() {
		if err := spec.Triggered[0].Effect(g, &game.StackItem{Controller: me.ID}); err != nil {
			t.Fatalf("effect: %v", err)
		}
	})
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no hand prompt")
	}
	for _, id := range c.ChooseCards {
		if id == spell {
			t.Error("an instant is not an artifact or creature card")
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{card}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if !inExile(g, card) {
		t.Error("the chosen card is exiled")
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Fat Bear" {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil {
		t.Fatal("no token copy on the battlefield")
	}
	if token.CurrentPower() != 3 || token.CurrentToughness() != 3 || !token.IsArtifact() || !token.IsCreature() || !token.HasSubtype("Golem") || !token.HasSubtype("Bear") {
		t.Errorf("token = %dx%d %q, want a 3/3 Golem Bear artifact creature", token.CurrentPower(), token.CurrentToughness(), token.TypeLine)
	}
}

func TestGuidelightMatrixDrawsAndAnimatesAVehicle(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	spec, ok := Lookup(guidelightOracle)
	if !ok || spec.Completeness != CompletenessFull {
		t.Fatal("Guidelight Matrix is not registered as complete (its saddle ability is implemented, #2695)")
	}
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Guidelight Matrix", "Artifact", guidelightOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		// the helper seeds the card (+1), casting removes it, the ETB draws one
		t.Errorf("hand %d, want %d (the enter trigger drew one)", me.Hand.Size(), before+1)
	}
	vehicle := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rig", TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	matrix := pushCatalogPermanent(g, me.ID, "Guidelight Matrix", "Artifact", guidelightOracle, false)
	mkAdd(t, g, me, "{C}{C}", game.AddManaOptions{})
	// Index 0 is the saddle ability, 1 the Vehicle animation (printed order).
	if err := g.ActivateCatalogAbility(me.ID, matrix, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: vehicle}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, vehicle); !c.IsCreature() {
		t.Error("the Vehicle becomes a creature")
	}
}
