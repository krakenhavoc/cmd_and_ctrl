package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b2_cards_test.go — Marvel Super Heroes, half B, second batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshB2BaronOracle     = "c91e64d4-e687-4ccd-a56e-796cff6e052c"
	mshB2BeastOracle     = "191ba5f0-45f4-4b7f-910d-4bd4d2d28b33"
	mshB2WidowOracle     = "e6c181f1-f4f7-4a96-b50c-717cf2a02098"
	mshB2BiochemOracle   = "45ddfa35-7b3c-4916-b231-dc721b7b56d9"
	mshB2ShieldOracle    = "8255a797-ddd3-4db0-a3f5-dd8f38c3b908"
	mshB2CapWingsOracle  = "942ac2ac-7528-4768-8398-8520a8ee15e5"
	mshB2ClaimOracle     = "3cd15225-e9b6-4998-b545-76536c72e948"
	mshB2ColleenOracle   = "685181dd-0124-4c76-9270-f9dd17e6a221"
	mshB2CrossbonesOracl = "ac950d10-e373-42d3-907b-ee56e45a2084"
	mshB2DaredevilOracle = "88aadf60-91f9-4932-a2bc-63365a139d36"
)

func TestMshB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mshB2BaronOracle:     "Baron Strucker, HYDRA Overlord",
		mshB2BeastOracle:     "Beast, Erudite Aerialist",
		mshB2WidowOracle:     "Black Widow, Double Agent",
		mshB2BiochemOracle:   "Bold Biochemist",
		mshB2ShieldOracle:    "Captain America's Shield",
		mshB2CapWingsOracle:  "Captain America, Wings of Freedom",
		mshB2ClaimOracle:     "Claim the Kingdom",
		mshB2ColleenOracle:   "Colleen Wing, Street Samurai",
		mshB2CrossbonesOracl: "Crossbones, Malicious Mercenary",
		mshB2DaredevilOracle: "Daredevil, Man Without Fear",
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

// mshB2Size is a permanent's power and toughness after every layer.
type mshB2Size struct{ Power, Toughness int }

func mshB2PT(t *testing.T, g *game.Game, id uuid.UUID) mshB2Size {
	t.Helper()
	p, tough := effectivePTOf(t, g, id)
	return mshB2Size{Power: p, Toughness: tough}
}

// mshB2PickIfAsked answers a target pick the engine opened, if it opened
// one: a lone legal target may be chosen for the player.
func mshB2PickIfAsked(t *testing.T, g *game.Game, chooser, id uuid.UUID) {
	t.Helper()
	if latestPickTarget(g, chooser) != nil {
		pickCard(t, g, chooser, id)
	}
}

// mshB2CastVillain casts a plain Villain creature spell from `p`'s hand
// and resolves it, so it enters the battlefield.
func mshB2CastVillain(t *testing.T, g *game.Game, p *game.Player, name string) uuid.UUID {
	t.Helper()
	return castPowerUpCreature(t, g, p, name, "Creature — Human Villain", "{1}", "{C}", "", 2)
}

// --- Baron Strucker, HYDRA Overlord -------------------------------------------

func TestMshB2BaronMakesYourVillainSpellsCostOneLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	pushCatalogPermanent(g, me.ID, "Baron Strucker, HYDRA Overlord", "Legendary Creature — Human Villain", mshB2BaronOracle, false)

	card := func(owner uuid.UUID, typeLine string) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: "Test", TypeLine: typeLine, ManaCost: "{2}{B}", Owner: owner, Controller: owner}
	}
	price := func(p *game.Player, c game.Card) string {
		pr, err := g.PriceCast(p.ID, c, game.CastSpellParams{})
		if err != nil {
			t.Fatalf("PriceCast: %v", err)
		}
		return pr.Total.String()
	}
	if got := price(me, card(me.ID, "Creature — Human Villain")); got != "{1}{B}" {
		t.Errorf("my Villain spell costs %s, want {1}{B}", got)
	}
	if got := price(me, card(me.ID, "Creature — Human Hero")); got != "{2}{B}" {
		t.Errorf("my Hero spell costs %s, want the full {2}{B}", got)
	}
	if got := price(opp, card(opp.ID, "Creature — Human Villain")); got != "{2}{B}" {
		t.Errorf("an opponent's Villain spell costs %s, want the full {2}{B}", got)
	}
}

func TestMshB2BaronOffersToConniveAnotherVillainThatEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	pushCatalogPermanent(g, me.ID, "Baron Strucker, HYDRA Overlord", "Legendary Creature — Human Villain", mshB2BaronOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	spell := handCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	seedLibrary(me, "Drawn")

	villain := mshB2CastVillain(t, g, me, "Test Villain")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("the Villain did not connive: no discard prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	if got := plusOnes(g, villain); got != 1 {
		t.Errorf("Villain has %d +1/+1 counters, want 1 (it discarded a nonland card)", got)
	}
	if got := graveyardNames(me); len(got) != 1 || got[0] != "Spell A" {
		t.Errorf("graveyard = %v, want [Spell A]", got)
	}
}

func TestMshB2BaronConnivesOnlyOncePerTurnButDecliningDoesNotUseItUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	pushCatalogPermanent(g, me.ID, "Baron Strucker, HYDRA Overlord", "Legendary Creature — Human Villain", mshB2BaronOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	spell := handCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	seedLibrary(me, "Drawn A", "Drawn B")

	// Declined: the turn's connive is still available.
	mshB2CastVillain(t, g, me, "Villain One")
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Fatal("declining still conniving")
	}

	// Accepted for the second Villain.
	second := mshB2CastVillain(t, g, me, "Villain Two")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("the second Villain was not offered a connive after the first was declined")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	if got := plusOnes(g, second); got != 1 {
		t.Errorf("second Villain has %d counters, want 1", got)
	}

	// A third Villain the same turn is not offered another.
	library := me.Library.Size()
	mshB2CastVillain(t, g, me, "Villain Three")
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil || latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Error("a third Villain the same turn was offered a connive")
	}
	if me.Library.Size() != library {
		t.Errorf("the library shrank from %d to %d: a card was drawn", library, me.Library.Size())
	}
}

func TestMshB2BaronIgnoresNonVillainsAndOpponentsVillains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	pushCatalogPermanent(g, me.ID, "Baron Strucker, HYDRA Overlord", "Legendary Creature — Human Villain", mshB2BaronOracle, false)

	castPowerUpCreature(t, g, me, "Hero", "Creature — Human Hero", "{1}", "{C}", "", 2)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("a Hero entering offered a connive")
	}
	b12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("an opponent's Villain offered a connive")
	}
}

// --- Beast, Erudite Aerialist -------------------------------------------------

func TestMshB2BeastFliesOnlyAfterYouPutAPlusOneCounterOnHimThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	beast := b12Push(g, me.ID, "Beast, Erudite Aerialist", "Legendary Creature — Mutant Scientist Hero", mshB2BeastOracle, 3, 3)

	if hasKeywordOnBattlefield(t, g, beast, "flying") {
		t.Fatal("Beast has flying with no counter put on him")
	}
	if err := g.AddCounterBy(opp.ID, beast, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	if hasKeywordOnBattlefield(t, g, beast, "flying") {
		t.Error("an opponent's counter gave Beast flying: it must be you who put it")
	}
	if err := g.AddCounterBy(me.ID, beast, game.CounterMinusOne, 1); err != nil {
		t.Fatal(err)
	}
	if hasKeywordOnBattlefield(t, g, beast, "flying") {
		t.Error("a -1/-1 counter gave Beast flying")
	}
	if err := g.AddCounterBy(me.ID, beast, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	if !hasKeywordOnBattlefield(t, g, beast, "flying") {
		t.Error("Beast lacks flying after I put a +1/+1 counter on him")
	}

	advanceToPrecombatMainOf(t, g, 0)
	if hasKeywordOnBattlefield(t, g, beast, "flying") {
		t.Error("Beast kept flying into the next turn")
	}
}

func TestMshB2BeastDrawsWhenItDealsCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := b12Push(g, me.ID, "Beast, Erudite Aerialist", "Legendary Creature — Mutant Scientist Hero", mshB2BeastOracle, 3, 3)
	handBefore := me.Hand.Size()

	attackWith(t, g, opp.ID, beast)
	passPriorityAroundTable(t, g)

	if opp.Life != 37 {
		t.Fatalf("opponent life %d, want 37", opp.Life)
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want one card drawn", handBefore, got)
	}
}

// --- Black Widow, Double Agent ------------------------------------------------

func TestMshB2WidowGivesTheLoneAttackerFirstStrikeAndMenace(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Black Widow, Double Agent", "Legendary Creature — Human Hero Villain", mshB2WidowOracle, 3, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)

	if !hasKeywordOnBattlefield(t, g, bear, "first strike") || !hasKeywordOnBattlefield(t, g, bear, "menace") {
		t.Error("the lone attacker did not gain first strike and menace")
	}
}

func TestMshB2WidowHerselfHasDeathtouchAndTriggersForHerOwnLoneAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	widow := b12Push(g, me.ID, "Black Widow, Double Agent", "Legendary Creature — Human Hero Villain", mshB2WidowOracle, 3, 2)

	if !hasKeywordOnBattlefield(t, g, widow, "deathtouch") {
		t.Error("Black Widow lacks deathtouch")
	}
	declareAttack(t, g, opp.ID, widow)
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, widow, "first strike") || !hasKeywordOnBattlefield(t, g, widow, "menace") {
		t.Error("Black Widow did not give herself first strike and menace")
	}
}

func TestMshB2WidowDoesNothingWhenTwoCreaturesAttackOrAnOpponentAttacksAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Black Widow, Double Agent", "Legendary Creature — Human Hero Villain", mshB2WidowOracle, 3, 2)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if hasKeywordOnBattlefield(t, g, id, "first strike") || hasKeywordOnBattlefield(t, g, id, "menace") {
			t.Error("a creature that did not attack alone gained keywords")
		}
	}
}

// --- Bold Biochemist ----------------------------------------------------------

func TestMshB2BiochemistPowerUpPutsACounterAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	chem := castPowerUpCreature(t, g, me, "Bold Biochemist", "Creature — Human Scientist", "{1}{U}", "{C}{U}", mshB2BiochemOracle, 1)

	if got := chargedPowerUp(t, g, me.ID, chem); got != "{4}" {
		t.Fatalf("the turn it entered: %q, want {4} ({5}{U} less {1}{U})", got)
	}
	handBefore := me.Hand.Size()
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, chem, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, chem); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d -> %d, want two cards drawn", handBefore, got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, chem, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

func TestMshB2BiochemistPowerUpCostsFullPriceTheNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	chem := castPowerUpCreature(t, g, me, "Bold Biochemist", "Creature — Human Scientist", "{1}{U}", "{C}{U}", mshB2BiochemOracle, 1)
	powerUpPassTo(t, g, 0)
	if got := chargedPowerUp(t, g, me.ID, chem); got != "{5}{U}" {
		t.Errorf("the next turn: %q, want {5}{U}", got)
	}
}

// --- Captain America's Shield -------------------------------------------------

func mshB2Shield(g *game.Game, me uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, me, "Captain America's Shield", "Legendary Artifact — Equipment", mshB2ShieldOracle, false)
}

func TestMshB2ShieldIsIndestructibleAndGivesPlusEightToughnessAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	shield := mshB2Shield(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if !hasKeywordOnBattlefield(t, g, shield, "indestructible") {
		t.Error("the Shield is not indestructible")
	}
	equipTo(t, g, me.ID, shield, bear)
	if card := mshB2PT(t, g, bear); card.Power != 2 || card.Toughness != 10 {
		t.Errorf("equipped Bear = %d/%d, want 2/10", card.Power, card.Toughness)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "vigilance") {
		t.Error("the equipped creature lacks vigilance")
	}
}

func TestMshB2ShieldTapsATargetCreatureTheDefendingPlayerControls(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	shield := mshB2Shield(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bystander := b12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	equipTo(t, g, me.ID, shield, bear)

	declareAttack(t, g, opp.ID, bear)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, bystander) || hasID(p.PickTargetCards, bear) {
			t.Errorf("offered a creature the defending player does not control: %v", p.PickTargetCards)
		}
		pickCard(t, g, me.ID, theirs)
	}
	passPriorityAroundTable(t, g)

	if !findBattlefieldCardForTest(g, theirs).Tapped {
		t.Error("the defending player's creature was not tapped")
	}
	if findBattlefieldCardForTest(g, bystander).Tapped {
		t.Error("another player's creature was tapped")
	}
	if findBattlefieldCardForTest(g, bear).Tapped {
		t.Error("the equipped attacker tapped despite vigilance")
	}
}

func TestMshB2ShieldDoesNotTriggerForAnotherCreatureAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	shield := mshB2Shield(g, me.ID)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	equipTo(t, g, me.ID, shield, bear)

	declareAttack(t, g, opp.ID, other)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, theirs).Tapped {
		t.Error("an unequipped creature attacking tapped a creature")
	}
}

// --- Captain America, Wings of Freedom ----------------------------------------

func TestMshB2CapWingsPumpsEachOtherHeroByHisToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	captain := b12Push(g, me.ID, "Captain America, Wings of Freedom", "Legendary Creature — Human Soldier Hero", mshB2CapWingsOracle, 3, 1)
	hero := b12Creature(g, me.ID, "Other Hero", "Creature — Human Hero", 2, 2)
	villain := b12Creature(g, me.ID, "Villain", "Creature — Human Villain", 2, 2)
	theirHero := b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)
	if err := g.AddCounterBy(me.ID, captain, game.CounterPlusOne, 2); err != nil {
		t.Fatal(err)
	}

	selfBefore := mshB2PT(t, g, captain)
	declareAttack(t, g, opp.ID, captain)
	passPriorityAroundTable(t, g)

	// His toughness is 1 + 2 counters = 3, so X = 3.
	if c := mshB2PT(t, g, hero); c.Power != 5 || c.Toughness != 5 {
		t.Errorf("other Hero = %d/%d, want 5/5", c.Power, c.Toughness)
	}
	if c := mshB2PT(t, g, captain); c != selfBefore {
		t.Errorf("Captain America = %d/%d, was %d/%d: he does not pump himself", c.Power, c.Toughness, selfBefore.Power, selfBefore.Toughness)
	}
	if c := mshB2PT(t, g, villain); c.Power != 2 || c.Toughness != 2 {
		t.Errorf("a non-Hero = %d/%d, want 2/2", c.Power, c.Toughness)
	}
	if c := mshB2PT(t, g, theirHero); c.Power != 2 || c.Toughness != 2 {
		t.Errorf("an opponent's Hero = %d/%d, want 2/2", c.Power, c.Toughness)
	}
}

func TestMshB2CapWingsHasFlyingAndFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	captain := b12Push(g, me.ID, "Captain America, Wings of Freedom", "Legendary Creature — Human Soldier Hero", mshB2CapWingsOracle, 3, 1)
	for _, kw := range []string{"flying", "first strike"} {
		if !hasKeywordOnBattlefield(t, g, captain, kw) {
			t.Errorf("Captain America lacks %s", kw)
		}
	}
}

// --- Claim the Kingdom --------------------------------------------------------

func TestMshB2ClaimLandfallPutsACounterOnTheTargetAndAPlanCounterOnItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	claim := pushCatalogPermanent(g, me.ID, "Claim the Kingdom", "Enchantment — Plan", mshB2ClaimOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	playLandFromHand(t, g, "Forest", "")
	mshB2PickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("Bear has %d +1/+1 counters, want 1", got)
	}
	if got := counterCount(g, claim, "plan"); got != 1 {
		t.Errorf("%d plan counters, want 1", got)
	}
}

func TestMshB2ClaimFourthPlanCounterSacrificesItAndGivesAnIndestructibleCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	claim := pushCatalogPermanent(g, me.ID, "Claim the Kingdom", "Enchantment — Plan", mshB2ClaimOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounterBy(me.ID, claim, "plan", 2); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)

	// The third land: three plan counters, still here.
	playLandFromHand(t, g, "Forest", "")
	mshB2PickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(claim) || counterCount(g, claim, "plan") != 3 {
		t.Fatalf("Claim the Kingdom after three plan counters: on battlefield %v, %d counters", g.Battlefield.Contains(claim), counterCount(g, claim, "plan"))
	}
	if counterCount(g, bear, game.CounterIndestructible) != 0 {
		t.Fatal("the indestructible counter came early")
	}

	// The fourth.
	playLandFromHand(t, g, "Forest", "")
	mshB2PickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	mshB2PickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(claim) {
		t.Error("Claim the Kingdom was not sacrificed on its fourth plan counter")
	}
	if got := counterCount(g, bear, game.CounterIndestructible); got != 1 {
		t.Errorf("Bear has %d indestructible counters, want 1", got)
	}
	if got := plusOnes(g, bear); got != 2 {
		t.Errorf("Bear has %d +1/+1 counters, want 2 (the third and fourth landfalls)", got)
	}
}

// --- Colleen Wing, Street Samurai ---------------------------------------------

func TestMshB2ColleenGrowsAndScriesWhenYouTargetACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	colleen := b12Push(g, me.ID, "Colleen Wing, Street Samurai", "Legendary Creature — Human Samurai Hero", mshB2ColleenOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "Top", "Second")

	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(bear)})
	passPriorityAroundTable(t, g)

	if got := plusOnes(g, colleen); got != 1 {
		t.Errorf("Colleen has %d counters, want 1", got)
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("no scry 1 prompt: %+v", c)
	}
}

func TestMshB2ColleenTriggersForASpellTargetingHerself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	colleen := b12Push(g, me.ID, "Colleen Wing, Street Samurai", "Legendary Creature — Human Samurai Hero", mshB2ColleenOracle, 2, 2)

	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(colleen)})
	passPriorityAroundTable(t, g)

	if got := plusOnes(g, colleen); got != 1 {
		t.Errorf("Colleen has %d counters after targeting herself, want 1", got)
	}
}

func TestMshB2ColleenIgnoresASpellTargetingAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	colleen := b12Push(g, me.ID, "Colleen Wing, Street Samurai", "Legendary Creature — Human Samurai Hero", mshB2ColleenOracle, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Djeru's Resolve", "Instant", pr7aDjerusResolve, []game.TargetRef{pr7aCard(theirs)})
	passPriorityAroundTable(t, g)

	if got := plusOnes(g, colleen); got != 0 {
		t.Errorf("Colleen has %d counters from a spell on an opponent's creature, want 0", got)
	}
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("a scry was offered")
	}
}

// --- Crossbones, Malicious Mercenary ------------------------------------------

func TestMshB2CrossbonesGrowsAndPingsEachOpponentWhenAnotherVillainEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	cb := b12Push(g, me.ID, "Crossbones, Malicious Mercenary", "Legendary Creature — Human Mercenary Villain", mshB2CrossbonesOracl, 3, 3)
	if !hasKeywordOnBattlefield(t, g, cb, "deathtouch") {
		t.Error("Crossbones lacks deathtouch")
	}

	mshB2CastVillain(t, g, me, "Villain One")
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, cb); got != 1 {
		t.Errorf("Crossbones has %d counters, want 1", got)
	}
	for _, opp := range g.Seats[1:] {
		if opp.Life != 38 {
			t.Errorf("opponent life %d, want 38", opp.Life)
		}
	}
	if me.Life != 40 {
		t.Errorf("my life %d, want 40", me.Life)
	}

	// Once each turn.
	mshB2CastVillain(t, g, me, "Villain Two")
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, cb); got != 1 {
		t.Errorf("a second Villain the same turn: %d counters, want 1", got)
	}
	if g.Seats[1].Life != 38 {
		t.Errorf("a second Villain the same turn: opponent life %d, want 38", g.Seats[1].Life)
	}

	// And again the next turn.
	advanceToPrecombatMainOf(t, g, 0)
	mshB2CastVillain(t, g, me, "Villain Three")
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, cb); got != 2 {
		t.Errorf("next turn: %d counters, want 2", got)
	}
	if g.Seats[1].Life != 36 {
		t.Errorf("next turn: opponent life %d, want 36", g.Seats[1].Life)
	}
}

func TestMshB2CrossbonesIgnoresNonVillainsAndOpponentsVillains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	cb := b12Push(g, me.ID, "Crossbones, Malicious Mercenary", "Legendary Creature — Human Mercenary Villain", mshB2CrossbonesOracl, 3, 3)

	castPowerUpCreature(t, g, me, "Hero", "Creature — Human Hero", "{1}", "{C}", "", 2)
	b12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, cb); got != 0 {
		t.Errorf("Crossbones has %d counters from creatures that are not my Villains", got)
	}
	if opp.Life != 40 {
		t.Errorf("opponent life %d, want 40", opp.Life)
	}
}

// --- Daredevil, Man Without Fear ----------------------------------------------

func TestMshB2DaredevilHasRadarSense(t *testing.T) {
	spec, ok := Lookup(mshB2DaredevilOracle)
	if !ok || spec.LibraryTopVisible != game.LibraryTopOwner {
		t.Errorf("Daredevil's library-top visibility = %v, want the owner's", spec.LibraryTopVisible)
	}
}

func TestMshB2DaredevilExilesAHeroCardPumpsAndLetsYouPlayIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	dd := b12Push(g, me.ID, "Daredevil, Man Without Fear", "Legendary Creature — Human Hero", mshB2DaredevilOracle, 3, 4)
	hero := b10LibraryTop(me, "Some Hero", "Creature — Human Hero", "{1}{W}", 2, 2)
	library := me.Library.Size()

	declareAttack(t, g, opp.ID, dd)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if !b12ExiledPlayableBy(t, g, hero, me.ID) {
		t.Error("the Hero card is not exiled with a permission to play it")
	}
	if me.Library.Size() != library-1 {
		t.Errorf("library %d -> %d, want one card exiled", library, me.Library.Size())
	}
	if c := mshB2PT(t, g, dd); c.Power != 5 || c.Toughness != 5 {
		t.Errorf("Daredevil = %d/%d, want 5/5 after a Hero card", c.Power, c.Toughness)
	}
}

func TestMshB2DaredevilGivesNoPumpForANonHeroCardButStillLetsYouPlayIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	dd := b12Push(g, me.ID, "Daredevil, Man Without Fear", "Legendary Creature — Human Hero", mshB2DaredevilOracle, 3, 4)
	spell := b10LibraryTop(me, "Some Spell", "Sorcery", "{1}{R}", 0, 0)

	declareAttack(t, g, opp.ID, dd)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if !b12ExiledPlayableBy(t, g, spell, me.ID) {
		t.Error("the card is not exiled with a permission to play it")
	}
	if c := mshB2PT(t, g, dd); c.Power != 3 || c.Toughness != 4 {
		t.Errorf("Daredevil = %d/%d, want 3/4 (not a Hero card)", c.Power, c.Toughness)
	}
}

func TestMshB2DaredevilMayDeclineAndTriggersOncePerAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	dd := b12Push(g, me.ID, "Daredevil, Man Without Fear", "Legendary Creature — Human Hero", mshB2DaredevilOracle, 3, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hero := b10LibraryTop(me, "Some Hero", "Creature — Human Hero", "{1}{W}", 2, 2)
	library := me.Library.Size()

	declareAttack(t, g, opp.ID, dd, bear)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("two attackers asked twice")
	}
	if me.Library.Size() != library || g.Exile.Contains(hero) {
		t.Error("declining still exiled the top card")
	}
	if c := mshB2PT(t, g, dd); c.Power != 3 || c.Toughness != 4 {
		t.Errorf("Daredevil = %d/%d after declining, want 3/4", c.Power, c.Toughness)
	}
}
