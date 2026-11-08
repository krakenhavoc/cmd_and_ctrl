package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_t_cards_test.go — Wilds of Eldraine, the T batch (#4471). Each test
// plays the real catalog entry and asserts what its text says.

const (
	woeTNightmaresOracle   = "e5557c9f-e4c2-41af-9fcc-371409e1dec6"
	woeTLookoutOracle      = "d6a767c2-8bc8-4686-899a-0ba3a525977a"
	woeTRatterOracle       = "ba36f27a-08fc-46a5-bc55-2a71091cebcc"
	woeTHartOracle         = tempestHartOracleID
	woeTTomeseekerOracle   = "b07bf08b-d9ec-429f-be6d-a7afa27c60c4"
	woeTWitchstalkerOracle = "97b9e3f1-7569-45e0-aef9-0d8d2845dbf1"
	woeTCliqueOracle       = threadbindCliqueOracleID
	woeTAdmirerOracle      = "77b48b3c-c780-4c5b-90e6-33c5fd940c16"
	woeTTotentanzOracle    = "cfa46f8f-b3ec-4ee0-acdd-622f3177e02a"
	woeTCookieOracle       = "8a1a5862-c027-4c15-99a7-1795d0da96da"
	woeTOupheOracle        = "16cdba97-73cf-44ba-9d64-f27009b829f8"
	woeTGuideOracle        = "dab16c93-9788-46e0-9345-838ebeda8d54"
	woeTTwinsOracle        = twiningTwinsOracleID
	woeTHunterOracle       = twoHeadedHunterOracleID
)

func TestWoeTEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		woeTNightmaresOracle:    "Taken by Nightmares",
		woeTLookoutOracle:       "Tanglespan Lookout",
		woeTRatterOracle:        "Tattered Ratter",
		woeTHartOracle:          "Tempest Hart",
		woeTHartOracle + "#1":   "Scan the Clouds",
		woeTTomeseekerOracle:    "Tenacious Tomeseeker",
		woeTWitchstalkerOracle:  "Territorial Witchstalker",
		woeTCliqueOracle:        "Threadbind Clique",
		woeTCliqueOracle + "#1": "Rip the Seams",
		woeTAdmirerOracle:       "Toadstool Admirer",
		woeTTotentanzOracle:     "Totentanz, Swarm Piper",
		woeTCookieOracle:        "Tough Cookie",
		woeTOupheOracle:         "Troublemaker Ouphe",
		woeTGuideOracle:         "Tuinvale Guide",
		woeTTwinsOracle:         "Twining Twins",
		woeTTwinsOracle + "#1":  "Swift Spiral",
		woeTHunterOracle:        "Two-Headed Hunter",
		woeTHunterOracle + "#1": "Twice the Rage",
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

// --- Taken by Nightmares ---------------------------------------------------------

func TestWoeTTakenByNightmaresExilesWithoutScryingWithNoEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	castCatalogSpell(t, g, "Taken by Nightmares", "Instant", woeTNightmaresOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Error("the creature was not exiled")
	}
	if c := scryChoiceFor(g, me.ID); c != nil {
		t.Errorf("a scry prompt with no enchantment: %v", c)
	}
}

func TestWoeTTakenByNightmaresScriesTwoWhenYouControlAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	b12Permanent(g, me.ID, "My Anthem", "Enchantment")
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	castCatalogSpell(t, g, "Taken by Nightmares", "Instant", woeTNightmaresOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Error("the creature was not exiled")
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %v, want one looking at two cards", c)
	}
	answerScryKeepAll(t, g, me.ID)
}

// --- Tanglespan Lookout ----------------------------------------------------------

// woeTAuraEnters puts an Aura onto the battlefield under owner's control,
// attached to host, and announces its entry.
func woeTAuraEnters(t *testing.T, g *game.Game, owner, host uuid.UUID) {
	t.Helper()
	aura := b12Permanent(g, owner, "Some Aura", "Enchantment — Aura")
	woeB1Attach(t, g, aura, host)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: aura, Actor: owner})
	})
	passPriorityAroundTable(t, g)
}

func TestWoeTTanglespanLookoutDrawsWhenYourAuraEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	pushCatalogPermanent(g, me.ID, "Tanglespan Lookout", "Creature — Satyr", woeTLookoutOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	theirBear := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)

	hand := handSize(me)
	woeTAuraEnters(t, g, opp.ID, theirBear)
	if got := handSize(me); got != hand {
		t.Fatalf("hand %d after an opponent's Aura entered, want %d", got, hand)
	}
	enchantment := b12Permanent(g, me.ID, "My Anthem", "Enchantment")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: enchantment, Actor: me.ID})
	})
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand {
		t.Fatalf("hand %d after a non-Aura enchantment entered, want %d", got, hand)
	}
	woeTAuraEnters(t, g, me.ID, bear)
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand %d after your Aura entered, want %d", got, hand+1)
	}
}

// --- Tattered Ratter -------------------------------------------------------------

func TestWoeTTatteredRatterPumpsYourBlockedRatOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Tattered Ratter", "Creature — Human Peasant", woeTRatterOracle, false)
	rat := b12Creature(g, me.ID, "My Rat", "Creature — Rat", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	wallA := b12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 5)
	wallB := b12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 5)
	declareAttack(t, g, opp.ID, rat, bear)
	pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wallA: rat, wallB: bear})
	if p, tough := effectivePTOf(t, g, rat); p != 3 || tough != 1 {
		t.Errorf("the blocked Rat is %d/%d, want 3/1", p, tough)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the blocked non-Rat is %d/%d, want an untouched 2/2", p, tough)
	}
}

// --- Tempest Hart // Scan the Clouds ---------------------------------------------

func TestWoeTTempestHartGrowsOnlyForSpellsOfManaValueFiveOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	hart := b12Push(g, me.ID, "Tempest Hart", "Creature — Elemental Elk", woeTHartOracle, 3, 4)
	assertKeywords(t, g, hart, "trample")
	msc11Cast(t, g, "Four Spell", "Sorcery", "{2}{G}{G}", nil)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, hart); got != 0 {
		t.Fatalf("a mana value 4 spell gave %d counters, want 0", got)
	}
	msc11Cast(t, g, "Five Spell", "Sorcery", "{3}{G}{G}", nil)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, hart); got != 1 {
		t.Errorf("a mana value 5 spell gave %d counters, want 1", got)
	}
}

func TestWoeTScanTheCloudsDrawsTwoThenDiscardsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	card := woeB1Adventure(me, woeTHartOracle,
		game.Face{Name: "Tempest Hart", TypeLine: "Creature — Elemental Elk", ManaCost: "{3}{G}", Colors: []string{"G"}, Power: 3, Toughness: 4},
		game.Face{Name: "Scan the Clouds", TypeLine: "Instant — Adventure", ManaCost: "{1}{U}", Colors: []string{"U"}})
	hand := handSize(me)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand+1 {
		t.Fatalf("hand %d after the draws, want %d (the Adventure left, two came in)", got, hand+1)
	}
	if got := discardOwed(g, me.ID); got != 2 {
		t.Fatalf("%d discards owed, want 2", got)
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand-1 {
		t.Errorf("hand %d after the discards, want %d", got, hand-1)
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("the Adventure card is not in exile after it resolved")
	}
}

// --- Tenacious Tomeseeker --------------------------------------------------------

func TestWoeTTenaciousTomeseekerBargainedReturnsAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	token := woeILToken(g, me.ID)
	spell := pushGraveyardCardTyped(me, "Old Bolt", "Instant")
	creature := pushGraveyardCardTyped(me, "Old Bear", woeMPBearType)
	p1g2Cast(t, g, "Tenacious Tomeseeker", "Creature — Human Knight", woeTTomeseekerOracle, "", game.CastSpellParams{
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the bargained Tomeseeker asked for no target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: creature}); err == nil {
		t.Error("a creature card was a legal target")
	}
	pickTriggerTarget(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(spell) {
		t.Error("the instant was not returned to your hand")
	}
}

func TestWoeTTenaciousTomeseekerUnbargainedDoesNothingOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	spell := pushGraveyardCardTyped(me, "Old Bolt", "Instant")
	castAndResolveCreature(t, g, "Tenacious Tomeseeker", "Creature — Human Knight", woeTTomeseekerOracle)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an unbargained Tomeseeker asked for a target")
	}
	if me.Hand.Contains(spell) {
		t.Error("an unbargained Tomeseeker returned a card")
	}
}

// --- Territorial Witchstalker ----------------------------------------------------

func TestWoeTTerritorialWitchstalkerAttacksWithAPowerFourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	ws := b12Push(g, me.ID, "Territorial Witchstalker", "Creature — Wolf", woeTWitchstalkerOracle, 2, 3)
	b12Creature(g, me.ID, "My Giant", "Creature — Giant", 4, 4)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, ws); p != 3 || tough != 3 {
		t.Errorf("the Witchstalker is %d/%d, want 3/3", p, tough)
	}
	assertKeywords(t, g, ws, "defender")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ws, opp.ID); err != nil {
		t.Errorf("the Witchstalker could not attack: %v", err)
	}
}

func TestWoeTTerritorialWitchstalkerStaysHomeWithoutAPowerFourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	ws := b12Push(g, me.ID, "Territorial Witchstalker", "Creature — Wolf", woeTWitchstalkerOracle, 2, 3)
	b12Creature(g, me.ID, "My Bear", woeMPBearType, 3, 3)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if triggerOnStackFrom(g, "Territorial Witchstalker") {
		t.Fatal("the trigger fired with no power 4 creature")
	}
	if p, _ := effectivePTOf(t, g, ws); p != 2 {
		t.Errorf("the Witchstalker has power %d, want 2", p)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ws, opp.ID); err == nil {
		t.Error("a creature with defender attacked")
	}
}

// --- Threadbind Clique // Rip the Seams ------------------------------------------

func woeTClique(p *game.Player) game.Card {
	return woeB1Adventure(p, woeTCliqueOracle,
		game.Face{Name: "Threadbind Clique", TypeLine: "Creature — Faerie", ManaCost: "{3}{U}", Colors: []string{"U"}, Power: 3, Toughness: 3},
		game.Face{Name: "Rip the Seams", TypeLine: "Instant — Adventure", ManaCost: "{2}{W}", Colors: []string{"W"}})
}

func TestWoeTRipTheSeamsDestroysATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	tapForTest(t, g, theirs)
	woeB1CastAdventure(t, g, me, woeTClique(me), cardRefs(theirs))
	if g.Battlefield.Contains(theirs) {
		t.Error("the tapped creature was not destroyed")
	}
}

func TestWoeTRipTheSeamsRefusesAnUntappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	card := woeTClique(me)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(theirs)}); err == nil {
		t.Error("an untapped creature was a legal target")
	}
}

func TestWoeTThreadbindCliqueFlies(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	clique := pushCatalogPermanent(g, me.ID, "Threadbind Clique", "Creature — Faerie", woeTCliqueOracle, false)
	assertKeywords(t, g, clique, "flying")
}

// --- Toadstool Admirer -----------------------------------------------------------

func TestWoeTToadstoolAdmirerGrowsWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	admirer := pushCatalogPermanent(g, me.ID, "Toadstool Admirer", "Creature — Ouphe", woeTAdmirerOracle, false)
	b16Activate(t, g, me.ID, admirer, 0, game.ActivateAbilityParams{})
	if got := plusOnes(g, admirer); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	b16Activate(t, g, me.ID, admirer, 0, game.ActivateAbilityParams{})
	if got := plusOnes(g, admirer); got != 2 {
		t.Errorf("%d +1/+1 counters after a second activation, want 2", got)
	}
}

// --- Totentanz, Swarm Piper ------------------------------------------------------

func TestWoeTTotentanzMakesARatWhenANontokenCreatureOfYoursDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	piper := pushCatalogPermanent(g, me.ID, "Totentanz, Swarm Piper", "Legendary Creature — Human Warlock Bard", woeTTotentanzOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	woeB1Destroy(t, g, theirs)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 0 {
		t.Fatalf("%d Rats after an opponent's creature died, want 0", got)
	}
	woeB1Destroy(t, g, bear)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Fatalf("%d Rats after your nontoken creature died, want 1", got)
	}
	woeB1Destroy(t, g, findBattlefieldByName(g, "Rat"))
	if got := b27CountNamed(g, me.ID, "Rat"); got != 0 {
		t.Fatalf("%d Rats after a Rat token died, want 0 (a token makes nothing)", got)
	}
	woeB1Destroy(t, g, piper)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Errorf("%d Rats after Totentanz died, want 1", got)
	}
}

func TestWoeTTotentanzGivesAnAttackingRatDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	piper := pushCatalogPermanent(g, me.ID, "Totentanz, Swarm Piper", "Legendary Creature — Human Warlock Bard", woeTTotentanzOracle, false)
	attacker := b12Creature(g, me.ID, "Attacking Rat", "Creature — Rat", 1, 1)
	home := b12Creature(g, me.ID, "Home Rat", "Creature — Rat", 1, 1)
	bear := b12Creature(g, me.ID, "Attacking Bear", woeMPBearType, 2, 2)
	declareAttack(t, g, opp.ID, attacker, bear)
	for name, id := range map[string]uuid.UUID{"a Rat that is not attacking": home, "an attacking non-Rat": bear} {
		if err := g.ActivateCatalogAbility(me.ID, piper, 0, game.ActivateAbilityParams{Targets: cardRefs(id)}); err == nil {
			t.Errorf("%s was a legal target", name)
		}
	}
	b16Activate(t, g, me.ID, piper, 0, game.ActivateAbilityParams{Targets: cardRefs(attacker)})
	if !hasEffectiveKeyword(t, g, attacker, "deathtouch") {
		t.Error("the attacking Rat has no deathtouch")
	}
}

// --- Tough Cookie ----------------------------------------------------------------

func TestWoeTToughCookieMakesAFoodAndAnimatesIt(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	cookie := castAndResolveCreature(t, g, "Tough Cookie", "Artifact Creature — Food Golem", woeTCookieOracle)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Fatalf("%d Food tokens on entry, want 1", got)
	}
	food := findBattlefieldByName(g, "Food")
	if err := g.ActivateCatalogAbility(me.ID, cookie, 0, game.ActivateAbilityParams{Targets: cardRefs(cookie)}); err == nil {
		t.Error("an artifact creature was a legal target")
	}
	b16Activate(t, g, me.ID, cookie, 0, game.ActivateAbilityParams{Targets: cardRefs(food)})
	if !isCreatureNow(g, food) {
		t.Fatal("the Food is not a creature")
	}
	if p, tough := effectivePTOf(t, g, food); p != 4 || tough != 4 {
		t.Errorf("the Food is %d/%d, want 4/4", p, tough)
	}
	endTurn(t, g)
	if isCreatureNow(g, food) {
		t.Error("the Food is still a creature after the turn")
	}
}

func TestWoeTToughCookieSacrificesItselfForThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	cookie := pushCatalogPermanent(g, me.ID, "Tough Cookie", "Artifact Creature — Food Golem", woeTCookieOracle, false)
	life := lifeOf(g, me.ID)
	b16Activate(t, g, me.ID, cookie, 1, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(cookie) {
		t.Error("Tough Cookie was not sacrificed")
	}
	if got := lifeOf(g, me.ID); got != life+3 {
		t.Errorf("you are at %d, want %d", got, life+3)
	}
}

// --- Troublemaker Ouphe ----------------------------------------------------------

func TestWoeTTroublemakerOupheBargainedExilesAnOpponentsArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	token := woeILToken(g, me.ID)
	theirs := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	p1g2Cast(t, g, "Troublemaker Ouphe", "Creature — Ouphe", woeTOupheOracle, "", game.CastSpellParams{
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Error("the opponent's artifact was not exiled")
	}
}

func TestWoeTTroublemakerOupheUnbargainedDoesNothingOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	theirs := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	castAndResolveCreature(t, g, "Troublemaker Ouphe", "Creature — Ouphe", woeTOupheOracle)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an unbargained Ouphe asked for a target")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("an unbargained Ouphe exiled a permanent")
	}
}

// --- Tuinvale Guide --------------------------------------------------------------

func TestWoeTTuinvaleGuideGetsPowerAndLifelinkOnlyWhileCelebrating(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	guide := b12Push(g, me.ID, "Tuinvale Guide", "Creature — Faerie Scout", woeTGuideOracle, 2, 3)
	assertKeywords(t, g, guide, "flying")
	if p, tough := effectivePTOf(t, g, guide); p != 2 || tough != 3 {
		t.Errorf("without celebration the Guide is %d/%d, want 2/3", p, tough)
	}
	if hasEffectiveKeyword(t, g, guide, "lifelink") {
		t.Error("the Guide has lifelink without celebration")
	}
	celebrationTokens(t, g, me.ID, 2)
	if p, tough := effectivePTOf(t, g, guide); p != 3 || tough != 3 {
		t.Errorf("celebrating, the Guide is %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, guide, "lifelink") {
		t.Error("the Guide has no lifelink while celebrating")
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, guide, "lifelink") {
		t.Error("lifelink outlasted the turn")
	}
}

// --- Twining Twins // Swift Spiral -----------------------------------------------

func TestWoeTTwiningTwinsHasFlyingAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	twins := pushCatalogPermanent(g, me.ID, "Twining Twins", "Creature — Faerie Wizard", woeTTwinsOracle, false)
	assertKeywords(t, g, twins, "flying", "vigilance")
}

func woeTTwins(p *game.Player) game.Card {
	return woeB1Adventure(p, woeTTwinsOracle,
		game.Face{Name: "Twining Twins", TypeLine: "Creature — Faerie Wizard", ManaCost: "{2}{U}{U}", Colors: []string{"U"}, Power: 4, Toughness: 4},
		game.Face{Name: "Swift Spiral", TypeLine: "Instant — Adventure", ManaCost: "{1}{W}", Colors: []string{"W"}})
}

func TestWoeTSwiftSpiralBlinksANontokenCreatureAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	woeB1CastAdventure(t, g, me, woeTTwins(me), cardRefs(theirs))
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the creature was not exiled")
	}
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	back := findBattlefieldByName(g, "Their Bear")
	if back == uuid.Nil {
		t.Fatal("the creature did not return at the end step")
	}
	if c, _ := battlefieldCard(g, back); c.Controller != opp.ID {
		t.Error("the creature did not return under its owner's control")
	}
}

func TestWoeTSwiftSpiralRefusesAToken(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	card := woeTTwins(me)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(token)}); err == nil {
		t.Error("a token was a legal target")
	}
}

// --- Two-Headed Hunter // Twice the Rage -----------------------------------------

func TestWoeTTwiceTheRageGrantsDoubleStrikeUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	card := woeB1Adventure(me, woeTHunterOracle,
		game.Face{Name: "Two-Headed Hunter", TypeLine: "Creature — Giant", ManaCost: "{4}{R}", Colors: []string{"R"}, Power: 5, Toughness: 4},
		game.Face{Name: "Twice the Rage", TypeLine: "Instant — Adventure", ManaCost: "{1}{R}", Colors: []string{"R"}})
	woeB1CastAdventure(t, g, me, card, cardRefs(bear))
	if !hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("the target has no double strike")
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("double strike outlasted the turn")
	}
}

func TestWoeTTwoHeadedHunterHasMenace(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	hunter := pushCatalogPermanent(g, me.ID, "Two-Headed Hunter", "Creature — Giant", woeTHunterOracle, false)
	assertKeywords(t, g, hunter, "menace")
}
