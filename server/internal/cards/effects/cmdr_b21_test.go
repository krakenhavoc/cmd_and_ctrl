package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b21_test.go — Commander deck cards, batch 21 (#5368).

const (
	cmdrB21MigrationPathOracle = "db4907f3-4071-47c4-9329-96b4813435be"
	cmdrB21BattlementOracle    = "585f62dc-4461-42f1-a3a4-b19a1e550d2d"
	cmdrB21CaryatidOracle      = "13d4c46b-c2d6-44cc-a252-4a991d471854"
	cmdrB21WitnessOracle       = "a52a06aa-56fc-4845-bbcd-a54f16369b34"
	cmdrB21TowerDefenseOracle  = "58853200-efbe-41cc-87f1-720841308f4d"
	cmdrB21TitanOracle         = "a95309ac-f920-474d-81b5-0f68c8a54f34"
	cmdrB21AronOracle          = "d8b06606-9b1f-4e72-a280-12a6e912bb50"
	cmdrB21AtarkaOracle        = "c0b3bbce-977c-42a9-afcb-dabdfc717c97"
	cmdrB21BaralOracle         = "386a6fa7-db5b-4e82-bd25-fd86d910ed0c"
	cmdrB21ExpansionOracle     = "af6f771f-5154-4fb3-8ed4-768d71eea568"
)

func cmdrB21AssertFull(t *testing.T, oracle, name string) {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", name)
	}
	if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
	}
}

func TestCmdrB21MigrationPathFetchesTwoBasicsTappedAndCycles(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21MigrationPathOracle, "Migration Path")
	g := newCatalogGame(t)
	me := g.Seats[0]
	stapleLibraryCard(me, "Basic Forest", "Basic Land — Forest")
	stapleLibraryCard(me, "Basic Island", "Basic Land — Island")
	stapleLibraryCard(me, "Basic Swamp", "Basic Land — Swamp")
	castCatalogSpell(t, g, "Migration Path", "Sorcery", cmdrB21MigrationPathOracle, nil)
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c == nil {
		t.Fatal("no search prompt")
	}
	answerSearchNamed(t, g, me.ID, "Basic Forest", "Basic Island")
	passPriorityAroundTable(t, g)
	n := 0
	for _, bc := range g.Battlefield.Cards {
		if bc.Controller == me.ID && bc.IsLand() {
			n++
			if !bc.Tapped {
				t.Errorf("%s entered untapped", bc.Name)
			}
		}
	}
	if n != 2 {
		t.Errorf("%d lands fetched, want 2", n)
	}

	// Cycling {2}.
	advanceTo(t, g, game.StepPrecombatMain)
	id := pushCatalogHandCard(me, "Migration Path", "Sorcery", cmdrB21MigrationPathOracle)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("cycled with an empty pool")
	}
	apaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled card should be in the graveyard")
	}
}

func TestCmdrB21OvergrownBattlementCountsCreaturesWithDefender(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21BattlementOracle, "Overgrown Battlement")
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushCatalogPermanent(g, me.ID, "Overgrown Battlement", "Creature — Wall", cmdrB21BattlementOracle, false)
	tapForMana(t, g, me.ID, a)
	if got := sortedPool(me); len(got) != 1 || got[0] != "G" {
		t.Fatalf("alone: pool = %v, want {G}", got)
	}
	me.ManaPool.EmptyPool()
	b := pushCatalogPermanent(g, me.ID, "Overgrown Battlement", "Creature — Wall", cmdrB21BattlementOracle, false)
	_ = b
	pushCatalogPermanent(g, me.ID, "Sylvan Caryatid", "Creature — Plant", cmdrB21CaryatidOracle, false)
	// An opponent's defender and a creature without defender don't count.
	pushCatalogPermanent(g, g.Seats[1].ID, "Overgrown Battlement", "Creature — Wall", cmdrB21BattlementOracle, false)
	pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	// Tapping the first one: three defenders (two Battlements, one Caryatid).
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == a {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	tapForMana(t, g, me.ID, a)
	if got := sortedPool(me); len(got) != 3 {
		t.Errorf("pool = %v, want three {G}", got)
	}
}

func TestCmdrB21SylvanCaryatidTapsForAnyColourAndHasHexproof(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21CaryatidOracle, "Sylvan Caryatid")
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Sylvan Caryatid", "Creature — Plant", cmdrB21CaryatidOracle, false)
	if !hasEffectiveKeyword(t, g, id, "hexproof") || !hasEffectiveKeyword(t, g, id, "defender") {
		t.Error("Sylvan Caryatid should have defender and hexproof")
	}
	tapForMana(t, g, me.ID, id)
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("pick = %+v, want one mana of any of five colours", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool = %v, want {U}", got)
	}
}

func TestCmdrB21TimelessWitnessReturnsACardMandatorilyAndEternalizes(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21WitnessOracle, "Timeless Witness")
	g := newCatalogGame(t)
	me := g.Seats[0]
	body := pushGraveyardCardForTest(me, "Old Body")
	castAndResolveCreature(t, g, "Timeless Witness", "Creature — Human Shaman", cmdrB21WitnessOracle)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("the trigger asked \"you may\"; Timeless Witness's return is mandatory")
	}
	pickCard(t, g, me.ID, body)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(body) {
		t.Error("the targeted card did not return to hand")
	}

	// Eternalize {5}{G}{G}: a 4/4 black Zombie copy, whose own trigger runs.
	other := pushGraveyardCardForTest(me, "Second Body")
	wit := seedGraveyardCard(t, g, "Timeless Witness", "Creature — Human Shaman", cmdrB21WitnessOracle)
	if err := g.ActivateCatalogAbility(me.ID, wit, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("eternalized with an empty pool")
	}
	apaMana(me, "C", "C", "C", "C", "C", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, wit, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("eternalize: %v", err)
	}
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(wit) {
		t.Error("the eternalized card should be exiled")
	}
	if !me.Hand.Contains(other) {
		t.Error("the token's enters trigger did not return the second card")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Timeless Witness" && c.Controller == me.ID && c.InstanceID != wit {
			if !c.IsToken() {
				continue
			}
			found = true
			if c.Power != 4 || c.Toughness != 4 {
				t.Errorf("token %d/%d, want a 4/4 token", c.Power, c.Toughness)
			}
		}
	}
	if !found {
		t.Error("no eternalized token")
	}
}

func TestCmdrB21TowerDefenseBoostsAndGrantsReachToYourCreaturesOnly(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21TowerDefenseOracle, "Tower Defense")
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Tower Defense", "Instant", cmdrB21TowerDefenseOracle, nil)
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, mine) != 2 || effectiveToughness(t, g, mine) != 7 {
		t.Errorf("my bear %d/%d, want 2/7", effectivePower(t, g, mine), effectiveToughness(t, g, mine))
	}
	if !hasEffectiveKeyword(t, g, mine, "reach") {
		t.Error("my bear should have reach")
	}
	if effectiveToughness(t, g, theirs) != 2 || hasEffectiveKeyword(t, g, theirs, "reach") {
		t.Error("their bear was affected")
	}
}

func TestCmdrB21ToweringTitanCountsToughnessAndSacrificesForTrample(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21TitanOracle, "Towering Titan")
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := pushCatalogPermanent(g, me.ID, "Sylvan Caryatid", "Creature — Plant", cmdrB21CaryatidOracle, false) // 1/1 seeded, defender
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 9, 9)
	toughness := effectiveToughness(t, g, wall) + 3
	titan := castAndResolveCreature(t, g, "Towering Titan", "Creature — Giant", cmdrB21TitanOracle)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, titan); got != toughness {
		t.Errorf("Titan entered with %d counters, want %d (other creatures you control)", got, toughness)
	}

	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, titan, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("sacrificed a creature without defender")
	}
	if err := g.ActivateCatalogAbility(me.ID, titan, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{wall}}); err != nil {
		t.Fatalf("sacrifice a defender: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, still := battlefieldCard(g, wall); still {
		t.Error("the defender was not sacrificed")
	}
	if !hasEffectiveKeyword(t, g, titan, "trample") || !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("all creatures should have trample")
	}
	their := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == opp.ID && hasEffectiveKeyword(t, g, c.InstanceID, "trample") {
			their++
		}
	}
	if their != 1 {
		t.Errorf("%d of the opponent's creatures have trample, want 1 (all creatures)", their)
	}
}

func TestCmdrB21AronSacrificesAnotherCreatureForACounterOnEach(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21AronOracle, "Aron, Benalia's Ruin")
	g := newCatalogGame(t)
	me := g.Seats[0]
	aron := pushCatalogPermanent(g, me.ID, "Aron, Benalia's Ruin", "Legendary Creature — Phyrexian Human", cmdrB21AronOracle, false)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, aron, "menace") {
		t.Error("Aron should have menace")
	}
	// The cost: {W}{B}. An empty pool is refused, and Aron can't pay with himself.
	if err := g.ActivateCatalogAbility(me.ID, aron, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	apaMana(me, "W", "B")
	if err := g.ActivateCatalogAbility(me.ID, aron, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{aron}}); err == nil {
		t.Fatal("Aron sacrificed himself; the cost says another creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, aron, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{a}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, still := battlefieldCard(g, a); still {
		t.Error("Bear A was not sacrificed")
	}
	if c := plusOneCounters(g, aron); c != 1 {
		t.Errorf("Aron has %d counters, want 1", c)
	}
	if c := plusOneCounters(g, b); c != 1 {
		t.Errorf("Bear B has %d counters, want 1", c)
	}
	if c := plusOneCounters(g, theirs); c != 0 {
		t.Errorf("their Bear has %d counters, want 0", c)
	}
}

func TestCmdrB21AtarkaGivesEveryAttackingDragonDoubleStrike(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21AtarkaOracle, "Atarka, World Render")
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atarka := pushCatalogPermanent(g, me.ID, "Atarka, World Render", "Legendary Creature — Dragon", cmdrB21AtarkaOracle, false)
	whelp := b12Creature(g, me.ID, "Whelp", "Creature — Dragon", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, atarka, whelp, bear)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{atarka, whelp} {
		if !hasEffectiveKeyword(t, g, id, "double strike") {
			t.Errorf("attacking Dragon %s lacks double strike", id)
		}
	}
	if hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("the attacking Bear gained double strike")
	}
}

// cmdrB21SplitInHand puts Expansion // Explosion into me's hand.
func cmdrB21SplitInHand(me *game.Player) uuid.UUID {
	card := game.Card{
		InstanceID: uuid.New(), OracleID: cmdrB21ExpansionOracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Expansion", TypeLine: "Instant", ManaCost: "{U/R}{U/R}", Colors: []string{"U", "R"}},
			{Name: "Explosion", TypeLine: "Instant", ManaCost: "{X}{U}{U}{R}{R}", Colors: []string{"U", "R"}},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	return card.InstanceID
}

func TestCmdrB21ExpansionCopiesASmallSpellButNotABigOne(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21ExpansionOracle, "Expansion")
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	big := handCardFull(me, "Big Spell", "Instant", "{3}{R}{R}", "", []string{"R"})
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the big spell: %v", err)
	}
	exp := cmdrB21SplitInHand(me)
	apaMana(me, "U", "R")
	if err := g.CastSpell(me.ID, exp, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}}); err == nil {
		t.Fatal("Expansion targeted a mana value 5 spell")
	}
	// A mana value 4 spell is a legal target and gets copied.
	small := handCardFull(me, "Small Bolt", "Instant", "{2}{R}{R}", boltOracleCombat, []string{"R"})
	if err := g.CastSpell(me.ID, small, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast the small spell: %v", err)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	apaMana(me, "U", "R")
	lifeBefore := opp.Life
	if err := g.CastSpell(me.ID, exp, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}}}); err != nil {
		t.Fatalf("Expansion on a mana value 4 spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the copy did not offer new targets")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("re-pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != lifeBefore-6 {
		t.Errorf("opponent life %d -> %d, want the Bolt and its copy (6)", lifeBefore, opp.Life)
	}
}

func TestCmdrB21ExplosionDealsXAndTheOtherTargetDraws(t *testing.T) {
	spec, ok := Lookup(cmdrB21ExpansionOracle + "#1")
	if !ok || spec.Name != "Explosion" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Explosion registered wrongly: %+v", spec)
	}
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	exp := cmdrB21SplitInHand(me)
	handBefore := third.Hand.Size()
	lifeBefore := opp.Life
	apaMana(me, "U", "U", "R", "R", "C", "C", "C")
	if err := g.CastSpell(me.ID, exp, game.CastSpellParams{
		Strict: true, Face: 1, XValue: 3,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: third.ID}},
	}); err != nil {
		t.Fatalf("cast Explosion: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != lifeBefore-3 {
		t.Errorf("opponent life %d -> %d, want 3 damage", lifeBefore, opp.Life)
	}
	if third.Hand.Size() != handBefore+3 {
		t.Errorf("third player's hand %d -> %d, want 3 cards drawn", handBefore, third.Hand.Size())
	}
}

func cmdrB21BaralTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Baral and Kari Zev", "Legendary Creature — Human", cmdrB21BaralOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return g, me
}

func cmdrB21Ragavans(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "First Mate Ragavan" && c.Controller == owner {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestCmdrB21BaralOffersALesserSpellSharingATypeAndCastingItMakesNoMonkey(t *testing.T) {
	cmdrB21AssertFull(t, cmdrB21BaralOracle, "Baral and Kari Zev")
	g, me := cmdrB21BaralTable(t)
	cheapInstant := handCardFull(me, "Cheap Instant", "Instant", "{1}", "", nil)
	cheapSorcery := handCardFull(me, "Cheap Sorcery", "Sorcery", "{1}", "", nil)
	equalInstant := handCardFull(me, "Equal Instant", "Instant", "{2}{U}", "", []string{"U"})
	cheapCreature := handCardFull(me, "Cheap Creature", "Creature — Bear", "{1}", "", nil)
	big := handCardFull(me, "Big Instant", "Instant", "{2}{U}", "", []string{"U"})
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheapInstant)
	for _, no := range []uuid.UUID{cheapSorcery, equalInstant, cheapCreature} {
		if hasID(offer.DiscardOptions, no) {
			t.Errorf("card %s was offered though it is not a lesser spell sharing a type", no)
		}
	}
	if err := castThroughResolutionCast(t, g, me.ID, cheapInstant, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(cmdrB21Ragavans(g, me.ID)); n != 0 {
		t.Errorf("%d Ragavan tokens after casting the spell, want 0", n)
	}
	// A second instant this turn doesn't trigger again.
	second := handCardFull(me, "Second Instant", "Instant", "{3}", "", nil)
	handCardFull(me, "Tiny Instant", "Instant", "{1}", "", nil)
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast second: %v", err)
	}
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil || len(cmdrB21Ragavans(g, me.ID)) != 0 {
		t.Error("the second instant or sorcery this turn triggered Baral and Kari Zev")
	}
}

func TestCmdrB21BaralDecliningMakesAHastyRagavan(t *testing.T) {
	g, me := cmdrB21BaralTable(t)
	cheapSorcery := handCardFull(me, "Cheap Sorcery", "Sorcery", "{1}", "", nil)
	big := handCardFull(me, "Big Sorcery", "Sorcery", "{3}", "", nil)
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheapSorcery)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	tokens := cmdrB21Ragavans(g, me.ID)
	if len(tokens) != 1 {
		t.Fatalf("%d Ragavan tokens, want 1", len(tokens))
	}
	c, _ := battlefieldCard(g, tokens[0])
	if c.Power != 2 || c.Toughness != 1 || !c.IsLegendary() || !c.HasSubtype("Monkey") || !c.HasSubtype("Pirate") {
		t.Errorf("token = %s %d/%d %q", c.Name, c.Power, c.Toughness, c.TypeLine)
	}
	if !hasEffectiveKeyword(t, g, tokens[0], "haste") {
		t.Error("the token should have haste this turn")
	}
}

func TestCmdrB21BaralWithNothingToCastMakesRagavanAtOnce(t *testing.T) {
	g, me := cmdrB21BaralTable(t)
	handCardFull(me, "Cheap Creature", "Creature — Bear", "{1}", "", nil)
	big := handCardFull(me, "Big Instant", "Instant", "{1}", "", nil)
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(cmdrB21Ragavans(g, me.ID)); n != 1 {
		t.Errorf("%d Ragavan tokens, want 1 (no lesser spell to cast)", n)
	}
}
