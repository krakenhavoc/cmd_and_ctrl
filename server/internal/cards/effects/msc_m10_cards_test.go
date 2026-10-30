package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m10_cards_test.go — Marvel Super Heroes Commander, M–Z batch 10
// (#4455). Each test plays the real catalog entry and asserts what its text
// says.

const (
	mscM10TombstoneOracle  = "6ad2cbb6-2291-420c-bfc8-ed3f079b59bf"
	mscM10UltimoOracle     = "43d79ee6-8ba1-4ec9-9bb2-907ed8401a2f"
	mscM10VenomOracle      = "6731d861-cc52-4a2e-8c33-154681fd50e4"
	mscM10GuardsmanOracle  = "0068f8c1-d1a2-4f7b-b39f-963acb2c023b"
	mscM10SkyCycleOracle   = "599402d9-2ea6-4709-b010-b35636be8c8c"
	mscM10SpiderMobileOrcl = "3c322d99-2bc4-4618-aeb3-5583fce54f1b"
	mscM10RoyalTalonOracle = "d6f36172-d3ef-49e9-a55f-ce2983bb2439"
	mscM10FantasticarOrcl  = "9b0abb58-4e2c-45dc-99b1-61a4cb79354b"
	mscM10ThanosCopterOrcl = "e58e94dc-9602-48db-9dd9-9d3051c33a76"
	mscM10NYamiOracle      = "5dec1b09-4df0-4e9b-9dca-b80134be3343"
	mscM10NamorKingOracle  = "171a0a09-4aee-466c-aebd-3b0d4c1f51b4"
	mscM10NamorSeasOracle  = "5278728e-9203-4d18-97a5-1c4f751a5acf"
	mscM10ProwlerOracle    = "76742bfc-ae99-4b06-b29d-e0e41d92ae14"
	mscM10TigerSharkOracle = "286be139-1ddf-4b04-86b2-cf17a06c262a"
	mscM10ThingSwingOracle = "b8306385-ec89-49f5-ac83-7522e4f0546c"
	mscM10TeamUpOracle     = "431c5b04-7387-4724-adfa-d1eb7874e79e"
	mscM10TippyToeOracle   = "6a1bce89-0c11-4d8d-aacb-c0a5a5effffc"
	mscM10ExperimentOracle = "5b7bbc76-22cc-482e-8d65-34cff34c93c8"
)

func TestMscM10EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mscM10TombstoneOracle:  "Tombstone, Career Criminal",
		mscM10UltimoOracle:     "Ultimo, Civilization's End",
		mscM10VenomOracle:      "Venom, Evil Unleashed",
		mscM10GuardsmanOracle:  "Vault Guardsman",
		mscM10SkyCycleOracle:   "Sky Cycle",
		mscM10SpiderMobileOrcl: "Spider-Mobile",
		mscM10RoyalTalonOracle: "Royal Talon Fighter Jet",
		mscM10FantasticarOrcl:  "The Fantasticar",
		mscM10ThanosCopterOrcl: "The Thanos-Copter",
		mscM10NYamiOracle:      "N'Yami-Class Mother Ship",
		mscM10NamorKingOracle:  "Namor, Atlantean King",
		mscM10NamorSeasOracle:  "Namor, Scourge of the Seas",
		mscM10ProwlerOracle:    "Prowler, Clawed Thief",
		mscM10TigerSharkOracle: "Tiger Shark, Abyssal Hunter",
		mscM10ThingSwingOracle: "Thing Swing",
		mscM10TeamUpOracle:     "Terrific Team-Up",
		mscM10TippyToeOracle:   "Tippy-Toe, Terrific Partner",
		mscM10ExperimentOracle: "Unstable Experiment",
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

// mscM10Hand puts a fully specified card into a player's hand.
func mscM10Hand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

// mscM10Connive answers the connive discard prompt with `discard`.
func mscM10Connive(t *testing.T, g *game.Game, who uuid.UUID, discard uuid.UUID) {
	t.Helper()
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, who)
	if pick == nil {
		t.Fatal("the creature did not connive: no discard prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, who, []uuid.UUID{discard}); err != nil {
		t.Fatalf("connive discard: %v", err)
	}
}

// mscM10EmptyHand clears a player's hand and gives them a known library.
func mscM10EmptyHand(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() { p.Hand.Cards = nil })
	seedLibrary(p, "Drawn A", "Drawn B", "Drawn C")
}

// --- Tombstone, Career Criminal ---

func TestMscM10TombstoneReturnsOnlyAVillainCardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	villain := pushCatalogGraveyardCard(me, "Dead Villain", "Creature — Human Villain", "", 2, 2)
	hero := pushCatalogGraveyardCard(me, "Dead Hero", "Creature — Human Hero", "", 2, 2)
	theirs := pushCatalogGraveyardCard(opp, "Their Villain", "Creature — Human Villain", "", 2, 2)

	castHoldCreature(t, g, "Tombstone, Career Criminal", "Legendary Creature — Human Villain", mscM10TombstoneOracle, 2, 2)
	ch := pickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{hero, theirs} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("a card that is not a Villain in your graveyard was accepted (%s)", bad)
		}
	}
	answerPickTarget(t, g, villain)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(villain) || me.Graveyard.Contains(villain) {
		t.Error("the Villain card was not returned to hand")
	}
	if !me.Graveyard.Contains(hero) || !opp.Graveyard.Contains(theirs) {
		t.Error("a card that was not the target left its graveyard")
	}
}

func TestMscM10TombstoneNeverTriggersWithoutAVillainInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogGraveyardCard(me, "Dead Hero", "Creature — Human Hero", "", 2, 2)
	castHoldCreature(t, g, "Tombstone, Career Criminal", "Legendary Creature — Human Villain", mscM10TombstoneOracle, 2, 2)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("the trigger asked for a target with no Villain card to take")
		}
	}
}

func TestMscM10TombstoneMakesOnlyYourVillainSpellsCostOneLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	pushCatalogPermanent(g, me.ID, "Tombstone, Career Criminal", "Legendary Creature — Human Villain", mscM10TombstoneOracle, false)
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
		t.Errorf("my Hero spell costs %s, want {2}{B}", got)
	}
	if got := price(opp, card(opp.ID, "Creature — Human Villain")); got != "{2}{B}" {
		t.Errorf("an opponent's Villain spell costs %s, want {2}{B}", got)
	}
}

// --- Ultimo, Civilization's End ---

func TestMscM10UltimoEntersAndEveryOpponentSacrificesACreatureOfTheirChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	powerUpMain(t, g)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	a := b12Creature(g, o1.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, o1.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, o2.ID, "Bear C", "Creature — Bear", 2, 2)

	ultimo := castPowerUpCreature(t, g, me, "Ultimo, Civilization's End", "Legendary Artifact Creature — Robot Villain", "{1}", "{C}", mscM10UltimoOracle, 6)
	assertKeywords(t, g, ultimo, "menace")
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, o1.ID, b)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(b) || !g.Battlefield.Contains(a) {
		t.Error("the opponent with two creatures did not sacrifice the one they chose")
	}
	if g.Battlefield.Contains(c) {
		t.Error("the opponent with one creature did not sacrifice it")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(ultimo) {
		t.Error("the controller sacrificed a creature to their own edict")
	}
	if o3.Life != 40 {
		t.Errorf("an opponent with no creature lost life: %d", o3.Life)
	}
}

func TestMscM10UltimoDiscardsFromHandForTwoBlackAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	spare := b12Creature(g, g.Seats[2].ID, "Third Bear", "Creature — Bear", 2, 2)
	id := mscM10Hand(me, game.Card{Name: "Ultimo, Civilization's End", TypeLine: "Legendary Artifact Creature — Robot Villain",
		ManaCost: "{4}{B}{B}", OracleID: mscM10UltimoOracle, Power: 6, Toughness: 5})

	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ability was activated with {C}{C} instead of {2}{B}")
	}
	if !me.Hand.Contains(id) {
		t.Fatal("a refused activation discarded the card")
	}
	floatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if me.Hand.Contains(id) || !me.Graveyard.Contains(id) {
		t.Error("the card was not discarded as a cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || g.Battlefield.Contains(spare) {
		t.Error("an opponent with a single creature kept it")
	}
}

// --- Venom, Evil Unleashed ---

func TestMscM10VenomExilesItselfForTwoCountersAndDeathtouchAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	id, _ := activateFromGraveyard(t, g, "Venom, Evil Unleashed", "Legendary Creature — Symbiote Villain",
		mscM10VenomOracle, 4, 5, "{C}{C}{B}", game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		})
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("Venom was not exiled from the graveyard as a cost")
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	assertKeywords(t, g, bear, "deathtouch")
	advanceToUpkeepOf(t, g, 1)
	if containsString(effectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("deathtouch lasted past the end of the turn")
	}
	if got := plusOnes(g, bear); got != 2 {
		t.Errorf("the counters did not stay: %d", got)
	}
}

func TestMscM10VenomCannotBeActivatedOutsideAMainPhaseOrWithoutThreeMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	id := pushCatalogGraveyardCard(me, "Venom, Evil Unleashed", "Legendary Creature — Symbiote Villain", mscM10VenomOracle, 4, 5)
	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{C}{B}")
	params := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, params); err == nil {
		t.Fatal("activated at the beginning of combat; it is sorcery speed")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	bear2 := b12Creature(g2, me2.ID, "My Bear", "Creature — Bear", 2, 2)
	id2 := pushCatalogGraveyardCard(me2, "Venom, Evil Unleashed", "Legendary Creature — Symbiote Villain", mscM10VenomOracle, 4, 5)
	toMain(t, g2)
	floatMana(t, g2, me2, "{C}{B}")
	params2 := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear2}}}
	if err := g2.ActivateCatalogAbility(me2.ID, id2, 0, params2); err == nil {
		t.Fatal("activated for two mana")
	}
	if !me2.Graveyard.Contains(id2) {
		t.Error("a refused activation exiled Venom")
	}
}

// --- Vault Guardsman ---

func mscM10CastGuardsman(t *testing.T, g *game.Game, me *game.Player, tap []uuid.UUID, mana string) (uuid.UUID, error) {
	t.Helper()
	toMainForCost(t, g)
	id := mscM10Hand(me, game.Card{Name: "Vault Guardsman", TypeLine: "Artifact Creature — Human Advisor",
		ManaCost: "{5}{W}", OracleID: mscM10GuardsmanOracle, Colors: []string{"W"}, Power: 3, Toughness: 4})
	floatMana(t, g, me, mana)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: tap})
}

func TestMscM10VaultGuardsmanRefusesAFiveManaCastWithoutConvoke(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Convoker One", "Creature — Bear", 1, 1)
	if _, err := mscM10CastGuardsman(t, g, me, nil, "{C}{C}{C}{C}{W}"); err == nil {
		t.Fatal("cast for five mana with no convoke; it costs {5}{W}")
	}
}

func TestMscM10VaultGuardsmanConvokesAndExilesAnArtifactOrCreatureUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d1 := b12Creature(g, me.ID, "Convoker One", "Creature — Bear", 1, 1)
	d2 := b12Creature(g, me.ID, "Convoker Two", "Creature — Bear", 1, 1)
	d3 := b12Creature(g, me.ID, "Convoker Three", "Creature — Bear", 1, 1)
	mine := b12Creature(g, me.ID, "Untapped Friend", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	guard, err := mscM10CastGuardsman(t, g, me, []uuid.UUID{d1, d2, d3}, "{C}{C}{W}")
	if err != nil {
		t.Fatalf("cast with three convokers: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{d1, d2, d3} {
		if c, _ := g.LookupCardForEffect(id); !c.Tapped {
			t.Error("a convoker was not tapped")
		}
	}
	ch := pickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{mine, land} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as a target", bad)
		}
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !msc10InExile(g, theirs) {
		t.Fatal("the targeted creature was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(guard) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("%d copies of the exiled creature under its owner's control after the Guardsman left, want 1", n)
	}
}

func TestMscM10VaultGuardsmanTakesAnOpponentsArtifactToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	_, err := mscM10CastGuardsman(t, g, me, nil, "{C}{C}{C}{C}{C}{W}")
	if err != nil {
		t.Fatalf("cast for six mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	if !msc10InExile(g, rock) {
		t.Error("the opponent's artifact was not exiled")
	}
}

func TestMscM10VaultGuardsmanGoneBeforeTheTriggerResolvesExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	guard, err := mscM10CastGuardsman(t, g, me, nil, "{C}{C}{C}{C}{C}{W}")
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTopOfStack(t, g)
	answerPickTarget(t, g, theirs)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(guard) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) || msc10InExile(g, theirs) {
		t.Error("a Guardsman gone before its trigger resolved still exiled the creature")
	}
}

// --- Prowler, Clawed Thief ---

func TestMscM10ProwlerConnivesWhenAnotherVillainYouControlEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	prowler := pushCatalogPermanent(g, me.ID, "Prowler, Clawed Thief", "Legendary Creature — Human Rogue Villain", mscM10ProwlerOracle, false)
	assertKeywords(t, g, prowler, "menace")
	mscM10EmptyHand(g, me)
	spare := handCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)

	mshB2CastVillain(t, g, me, "Test Villain")
	passPriorityAroundTable(t, g)
	mscM10Connive(t, g, me.ID, spare)
	if got := plusOnes(g, prowler); got != 1 {
		t.Errorf("Prowler has %d +1/+1 counters, want 1 (it discarded a nonland card)", got)
	}
	if me.Hand.Contains(spare) {
		t.Error("the discarded card is still in hand")
	}
}

func TestMscM10ProwlerGetsNoCounterForALandDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	prowler := pushCatalogPermanent(g, me.ID, "Prowler, Clawed Thief", "Legendary Creature — Human Rogue Villain", mscM10ProwlerOracle, false)
	mscM10EmptyHand(g, me)
	land := handCardFull(me, "Forest", "Basic Land — Forest", "", "", nil)
	mshB2CastVillain(t, g, me, "Test Villain")
	passPriorityAroundTable(t, g)
	mscM10Connive(t, g, me.ID, land)
	if got := plusOnes(g, prowler); got != 0 {
		t.Errorf("Prowler has %d counters after discarding a land, want 0", got)
	}
}

func TestMscM10ProwlerIgnoresItselfNonVillainsAndOpposingVillains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	mscM10EmptyHand(g, me)
	// Prowler entering is not "another" Villain.
	castPowerUpCreature(t, g, me, "Prowler, Clawed Thief", "Legendary Creature — Human Rogue Villain", "{1}", "{C}", mscM10ProwlerOracle, 2)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Fatal("Prowler connived when he himself entered")
	}
	castPowerUpCreature(t, g, me, "Plain Hero", "Creature — Human Hero", "{1}", "{C}", "", 2)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Fatal("Prowler connived for a Hero")
	}
	theirs := b12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirs}) })
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Fatal("Prowler connived for an opponent's Villain")
	}
}

// --- Tiger Shark, Abyssal Hunter ---

func TestMscM10TigerSharkConnivesWhenItEntersAndWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	mscM10EmptyHand(g, me)
	spare := handCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	shark := castPowerUpCreature(t, g, me, "Tiger Shark, Abyssal Hunter", "Legendary Creature — Shark Human Villain", "{1}", "{C}", mscM10TigerSharkOracle, 2)
	passPriorityAroundTable(t, g)
	mscM10Connive(t, g, me.ID, spare)
	if got := plusOnes(g, shark); got != 1 {
		t.Fatalf("after entering: %d counters, want 1", got)
	}

	// A Shark that is not summoning sick attacks and connives again.
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	mscM10EmptyHand(g2, me2)
	spare2 := handCardFull(me2, "Spell B", "Sorcery", "{1}", "", nil)
	shark2 := pushCatalogPermanent(g2, me2.ID, "Tiger Shark, Abyssal Hunter", "Legendary Creature — Shark Human Villain", mscM10TigerSharkOracle, false)
	declareAttack(t, g2, opp2.ID, shark2)
	passPriorityAroundTable(t, g2)
	mscM10Connive(t, g2, me2.ID, spare2)
	if got := plusOnes(g2, shark2); got != 1 {
		t.Errorf("after attacking: %d counters, want 1", got)
	}
	_ = opp
}

func TestMscM10TigerSharkCannotBeBlockedForFourAndAHybridMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shark := pushCatalogPermanent(g, me.ID, "Tiger Shark, Abyssal Hunter", "Legendary Creature — Shark Human Villain", mscM10TigerSharkOracle, false)
	toMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, shark, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for four mana")
	}
	assertRestrictions(t, g, shark, 0)
	floatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, shark, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, shark, game.CantBeBlocked)
}

// --- Unstable Experiment ---

func mscM10Experiment(t *testing.T, g *game.Game, player, creature uuid.UUID) error {
	t.Helper()
	refs := []game.TargetRef{{Kind: game.TargetPlayer, ID: player, Slot: 0}}
	if creature != uuid.Nil {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: creature, Slot: 1})
	}
	return castCatalogSpellErr(t, g, "Unstable Experiment", "Instant", mscM10ExperimentOracle, refs)
}

func TestMscM10UnstableExperimentDrawsThePlayerThenYourCreatureConnives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	mscM10EmptyHand(g, me)
	seedLibrary(opp, "Opp Draw")
	spare := handCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	oppHand := opp.Hand.Size()
	if err := mscM10Experiment(t, g, opp.ID, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != oppHand+1 {
		t.Errorf("the target player's hand %d -> %d, want +1", oppHand, opp.Hand.Size())
	}
	mscM10Connive(t, g, me.ID, spare)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("the creature has %d counters, want 1", got)
	}
}

func TestMscM10UnstableExperimentCreatureIsOptionalAndMustBeYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mscM10EmptyHand(g, me)
	if err := mscM10Experiment(t, g, me.ID, theirs); err == nil {
		t.Fatal("cast at a creature an opponent controls")
	}
	handBefore := me.Hand.Size()
	if err := mscM10Experiment(t, g, me.ID, uuid.Nil); err != nil {
		t.Fatalf("cast at a player alone: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Error("something connived with no creature chosen")
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand size %d, want %d (the spell left the hand, the player drew one)", me.Hand.Size(), handBefore+1)
	}
}

// --- shared vehicle helpers ---

const mscM10VehicleType = "Artifact — Vehicle"

// mscM10Vehicle puts a catalog Vehicle onto the battlefield with its printed
// stats, past summoning sickness.
func mscM10Vehicle(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	return pushPermanent(g, owner, game.Card{Name: name, TypeLine: mscM10VehicleType, OracleID: oracle, Power: power, Toughness: toughness})
}

// mscM10Crew crews ability 0 of `vehicle` with the given creatures and lets it
// resolve.
func mscM10Crew(t *testing.T, g *game.Game, me *game.Player, vehicle uuid.UUID, crewers ...uuid.UUID) {
	t.Helper()
	toMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, vehicle, 0, game.ActivateAbilityParams{CrewIDs: crewers}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCardByID(g, vehicle); !ok || !c.IsCreature() {
		t.Fatal("the crewed Vehicle is not a creature")
	}
}

func TestMscM10EveryVehicleCrewsForItsPrintedNumber(t *testing.T) {
	for oracle, crew := range map[string]int{
		mscM10SkyCycleOracle: 2, mscM10SpiderMobileOrcl: 2, mscM10RoyalTalonOracle: 2,
		mscM10ThanosCopterOrcl: 2, mscM10NYamiOracle: 3,
	} {
		abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: oracle})
		if len(abilities) != 1 || abilities[0].Cost.Crew != crew {
			t.Errorf("%s: activated abilities %+v, want the one Crew %d", oracle, abilities, crew)
		}
	}
	if abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: mscM10FantasticarOrcl}); len(abilities) != 0 {
		t.Errorf("The Fantasticar has no crew ability, got %d abilities", len(abilities))
	}
}

// --- Sky Cycle ---

func TestMscM10SkyCycleDealsTwiceTheVehiclesYouControlToACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mscM10Vehicle(g, me.ID, "Other Vehicle", "", 2, 2)
	mscM10Vehicle(g, opp.ID, "Their Vehicle", "", 2, 2) // not yours
	big := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	cycle := castHoldCreature(t, g, "Sky Cycle", mscM10VehicleType, mscM10SkyCycleOracle, 4, 3)
	assertKeywords(t, g, cycle, "flying")
	answerPickTarget(t, g, big)
	passPriorityAroundTable(t, g)
	// Sky Cycle and the other Vehicle: X = 4, exactly the Ogre's toughness.
	if g.Battlefield.Contains(big) {
		t.Error("4 damage (twice two Vehicles) did not destroy the 4/4")
	}
}

func TestMscM10SkyCycleAloneDealsTwoAndTheCreatureIsOptional(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	castHoldCreature(t, g, "Sky Cycle", mscM10VehicleType, mscM10SkyCycleOracle, 4, 3)
	answerPickTarget(t, g, ogre)
	passPriorityAroundTable(t, g)
	dmg := damageMarkedOn(g, ogre)
	if dmg != 2 {
		t.Errorf("the Ogre has %d damage, want 2 (twice the one Vehicle)", dmg)
	}

	g2 := newCatalogGame(t)
	opp2 := g2.Seats[1]
	bear := b12Creature(g2, opp2.ID, "Their Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g2, "Sky Cycle", mscM10VehicleType, mscM10SkyCycleOracle, 4, 3)
	ch := pickTargetPrompt(t, g2)
	if err := g2.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
		t.Fatalf("choosing no target: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(bear) {
		t.Error("no target was chosen but the Bear died")
	}
}

// --- Spider-Mobile ---

func TestMscM10SpiderMobileGetsPlusOnePlusOneForEachSpiderWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mobile := mscM10Vehicle(g, me.ID, "Spider-Mobile", mscM10SpiderMobileOrcl, 3, 3)
	assertKeywords(t, g, mobile, "trample")
	s1 := b12Creature(g, me.ID, "Spider One", "Creature — Spider Hero", 2, 2)
	b12Creature(g, me.ID, "Spider Two", "Creature — Spider Hero", 1, 1)
	b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 1, 1)
	b12Creature(g, opp.ID, "Their Spider", "Creature — Spider", 1, 1)
	mscM10Crew(t, g, me, mobile, s1)
	declareAttack(t, g, opp.ID, mobile)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, mobile); p != 5 || tough != 5 {
		t.Errorf("Spider-Mobile is %d/%d, want 5/5 (two Spiders of mine, not the opponent's)", p, tough)
	}
}

func TestMscM10SpiderMobileGrowsOnceWhenItBlocksNoMatterHowManyAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mobile := mscM10Vehicle(g, me.ID, "Spider-Mobile", mscM10SpiderMobileOrcl, 3, 3)
	b12Creature(g, me.ID, "Spider One", "Creature — Spider Hero", 2, 2)
	a := b12Creature(g, opp.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Attacker B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, CardID: mobile, Target: a, Actor: me.ID, Amount: 1})
		g.EmitEvent(game.Event{Kind: game.EventBlock, CardID: mobile, Target: b, Actor: me.ID, Amount: 2})
	})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, mobile); p != 4 || tough != 4 {
		t.Errorf("Spider-Mobile is %d/%d after blocking two attackers, want 4/4 (one trigger, one Spider)", p, tough)
	}
}

// --- Royal Talon Fighter Jet ---

func mscM10CastJet(t *testing.T, g *game.Game, me *game.Player, x int) uuid.UUID {
	t.Helper()
	toMainForCost(t, g)
	id := mscM10Hand(me, game.Card{Name: "Royal Talon Fighter Jet", TypeLine: mscM10VehicleType, ManaCost: "{X}{W}{W}",
		OracleID: mscM10RoyalTalonOracle, Colors: []string{"W"}, Power: 1, Toughness: 1})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}"); err != nil {
		t.Fatal(err)
	}
	if x > 0 {
		floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}"[:3*x])
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x}); err != nil {
		t.Fatalf("cast with X=%d: %v", x, err)
	}
	return id
}

func mscM10Soldiers(g *game.Game, owner uuid.UUID) int {
	return namedOnBattlefield(g, "Soldier", owner)
}

func TestMscM10RoyalTalonEntersWithXCountersAndMakesThatManySoldiersThenAgainOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jet := mscM10CastJet(t, g, me, 3)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, jet); got != 3 {
		t.Fatalf("%d +1/+1 counters, want 3", got)
	}
	assertKeywords(t, g, jet, "flying")
	if got := mscM10Soldiers(g, me.ID); got != 3 {
		t.Fatalf("%d Soldier tokens on entering, want 3", got)
	}
	// Crew it with the three Soldiers (power 1 each), then attack.
	var crew []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" && c.Controller == me.ID {
			crew = append(crew, c.InstanceID)
		}
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	mscM10Crew(t, g, me, jet, crew[:2]...)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == jet {
				g.Battlefield.Cards[i].SummonedThisTurn = false
				g.Battlefield.Cards[i].EnteredBattlefieldAt = 0
			}
		}
	})
	declareAttack(t, g, opp.ID, jet)
	passPriorityAroundTable(t, g)
	if got := mscM10Soldiers(g, me.ID); got != 6 {
		t.Errorf("%d Soldiers after attacking, want 6 (three more)", got)
	}
}

func TestMscM10RoyalTalonForZeroMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jet := mscM10CastJet(t, g, me, 0)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, jet); got != 0 {
		t.Errorf("%d counters for X=0", got)
	}
	if got := mscM10Soldiers(g, me.ID); got != 0 {
		t.Errorf("%d Soldiers for X=0, want none", got)
	}
}

func TestMscM10RoyalTalonCountsItsLastKnownCountersWhenItLeavesFirst(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jet := mscM10CastJet(t, g, me, 2)
	resolveTopOfStack(t, g)
	if got := plusOnes(g, jet); got != 2 {
		t.Fatalf("%d counters, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(jet) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := mscM10Soldiers(g, me.ID); got != 2 {
		t.Errorf("%d Soldiers from a Jet that left with two counters, want 2 (last-known information)", got)
	}
}

// --- The Fantasticar ---

func TestMscM10FantasticarMayBecomeACreatureWhenYouCastANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	car := mscM10Vehicle(g, me.ID, "The Fantasticar", mscM10FantasticarOrcl, 4, 4)
	assertKeywords(t, g, car, "flying")
	advanceToMain(t, g)

	castFree(t, g, me, "Bear Cub", "Creature — Bear")
	if latestTriggerPromptFor(g, me.ID) != nil {
		t.Fatal("a creature spell triggered The Fantasticar")
	}
	castFree(t, g, me, "Shock", "Instant")
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCardByID(g, car); c.IsCreature() {
		t.Fatal("declining still animated The Fantasticar")
	}
	castFree(t, g, me, "Shock Two", "Instant")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCardByID(g, car); !c.IsCreature() {
		t.Error("The Fantasticar did not become an artifact creature")
	}
}

func TestMscM10FantasticarSacrificesOnTheFourthNoncreatureSpellForFourConstructs(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	car := mscM10Vehicle(g, me.ID, "The Fantasticar", mscM10FantasticarOrcl, 4, 4)
	advanceToMain(t, g)
	mscM10SetNoncreatureTally(g, me.ID, 3)
	castFree(t, g, me, "Fourth Spell", "Instant")
	mscM10AnswerPrompts(t, g, me.ID, map[string]bool{"sacrifice": true})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(car) {
		t.Fatal("The Fantasticar is still on the battlefield")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Construct" && c.Controller == me.ID {
			n++
			if c.Power != 4 || c.Toughness != 4 || !game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "haste") || !c.IsToken() {
				t.Errorf("Construct is %d/%d %v, want a 4/4 flying haste artifact token", c.Power, c.Toughness, c.Keywords)
			}
		}
	}
	if n != 4 {
		t.Errorf("%d Constructs, want 4", n)
	}
}

func TestMscM10FantasticarKeepsItselfWhenYouDeclineOrItIsNotTheFourthNoncreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	car := mscM10Vehicle(g, me.ID, "The Fantasticar", mscM10FantasticarOrcl, 4, 4)
	advanceToMain(t, g)
	mscM10SetNoncreatureTally(g, me.ID, 3)
	// A creature spell is not the fourth noncreature spell.
	castFree(t, g, me, "Bear Cub", "Creature — Bear")
	if p := latestTriggerPromptFor(g, me.ID); p != nil {
		t.Fatalf("a creature spell raised a prompt: %q", p.Reason)
	}
	// The fourth, declined everywhere.
	castFree(t, g, me, "Fourth Spell", "Instant")
	if asked := mscM10AnswerPrompts(t, g, me.ID, map[string]bool{}); !mscM10Asked(asked, "sacrifice") {
		t.Errorf("the fourth noncreature spell did not offer the sacrifice; asked %q", asked)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(car) {
		t.Error("The Fantasticar was sacrificed though the player declined")
	}
	if n := namedOnBattlefield(g, "Construct", me.ID); n != 0 {
		t.Errorf("%d Constructs without the sacrifice", n)
	}
	// The fifth noncreature spell does not offer the sacrifice either.
	castFree(t, g, me, "Fifth Spell", "Instant")
	if asked := mscM10AnswerPrompts(t, g, me.ID, map[string]bool{}); mscM10Asked(asked, "sacrifice") {
		t.Errorf("the fifth noncreature spell offered the sacrifice: %q", asked)
	}
}

// mscM10AnswerPrompts settles every trigger-order and optional-trigger prompt
// `who` is asked until the stack is empty: yes when the question mentions a
// word set to true, no otherwise.
func mscM10AnswerPrompts(t *testing.T, g *game.Game, who uuid.UUID, yesWords map[string]bool) (asked []string) {
	t.Helper()
	for i := 0; i < 16; i++ {
		if triggerOrderPrompt(g) != nil {
			answerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if p := latestTriggerPromptFor(g, who); p != nil {
			asked = append(asked, p.Reason)
			yes := false
			for w := range yesWords {
				if strings.Contains(strings.ToLower(p.Reason), w) {
					yes = true
				}
			}
			if err := g.ResolveTriggerPrompt(p.ID, who, yes); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return asked
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return asked
}

func mscM10Asked(asked []string, word string) bool {
	for _, q := range asked {
		if strings.Contains(strings.ToLower(q), word) {
			return true
		}
	}
	return false
}

// mscM10SetNoncreatureTally sets how many noncreature spells `who` has cast this
// turn, so the next one is the (n+1)th.
func mscM10SetNoncreatureTally(g *game.Game, who uuid.UUID, n int) {
	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{}
		}
		g.SpellsCastThisTurn[who] = game.CastTally{Total: n, Noncreature: n}
	})
}

// --- The Thanos-Copter ---

func TestMscM10ThanosCopterAnimatesYourVehiclesWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	powerUpMain(t, g)
	mine := mscM10Vehicle(g, me.ID, "My Vehicle", "", 3, 3)
	theirs := mscM10Vehicle(g, opp.ID, "Their Vehicle", "", 3, 3)
	copter := castPowerUpCreature(t, g, me, "The Thanos-Copter", "Legendary "+mscM10VehicleType, "{1}", "{C}", mscM10ThanosCopterOrcl, 5)
	assertKeywords(t, g, copter, "flying", "haste")
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, copter} {
		if c, _ := battlefieldCardByID(g, id); !c.IsCreature() {
			t.Errorf("%s did not become an artifact creature", c.Name)
		}
	}
	if c, _ := battlefieldCardByID(g, theirs); c.IsCreature() {
		t.Error("an opponent's Vehicle became a creature")
	}
	// Until end of turn only.
	advanceToUpkeepOf(t, g, 1)
	if c, _ := battlefieldCardByID(g, mine); c.IsCreature() {
		t.Error("the Vehicle was still a creature on the next turn")
	}
}

func TestMscM10ThanosCopterDrawsForEachVehicleOfYoursThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	copter := mscM10Vehicle(g, me.ID, "The Thanos-Copter", mscM10ThanosCopterOrcl, 5, 5)
	b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	crew := b12Creature(g, me.ID, "Crew Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "Draw A", "Draw B", "Draw C")
	mscM10Crew(t, g, me, copter, crew)
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, copter, mustFindNamed(g, me.ID, "Plain Bear"))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1 (only the Vehicle's damage counts, not the Bear's)", got)
	}
}

func mustFindNamed(g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			return c.InstanceID
		}
	}
	panic("no permanent named " + name)
}

// --- N'Yami-Class Mother Ship ---

func mscM10NYamiHit(t *testing.T, g *game.Game, me, opp *game.Player) {
	t.Helper()
	ship := mscM10Vehicle(g, me.ID, "N'Yami-Class Mother Ship", mscM10NYamiOracle, 5, 7)
	assertKeywords(t, g, ship, "flying", "vigilance", "haste")
	crew := b12Creature(g, me.ID, "Crew Giant", "Creature — Giant", 3, 3)
	mscM10Crew(t, g, me, ship, crew)
	attackWith(t, g, opp.ID, ship)
	passPriorityAroundTable(t, g)
}

func TestMscM10NYamiPutsAPermanentCardOntoTheBattlefieldOrElseIntoHand(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		top := plTop(me, "Top Bear", "Creature — Bear", "{1}{G}")
		mscM10NYamiHit(t, g, me, opp)
		if chooseCardsChoiceFor(g, me.ID) == nil {
			t.Fatal("a permanent card on top offered no choice")
		}
		if accept {
			answerChooseCards(t, g, me.ID, top)
			if !g.Battlefield.Contains(top) || me.Hand.Contains(top) {
				t.Error("accepted: the permanent card did not enter the battlefield")
			}
		} else {
			answerChooseCards(t, g, me.ID)
			if g.Battlefield.Contains(top) || !me.Hand.Contains(top) {
				t.Error("declined: the card must go into your hand")
			}
		}
	}
}

func TestMscM10NYamiPutsANonpermanentCardIntoYourHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := plTop(me, "Lightning Bolt", "Instant", "{R}")
	mscM10NYamiHit(t, g, me, opp)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("an instant offered to be put onto the battlefield")
	}
	if !me.Hand.Contains(bolt) || g.Battlefield.Contains(bolt) {
		t.Error("the instant did not go into your hand")
	}
}

// --- Namor, Atlantean King ---

func TestMscM10NamorKingMakesAMerfolkForEachNoncreatureSpellYouCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	namor := b27Push(g, me.ID, "Namor, Atlantean King", "Legendary Creature — Mutant Merfolk Noble", mscM10NamorKingOracle, "{1}{W}{U}", 2, 2, "W", "U")
	assertKeywords(t, g, namor, "flying")
	advanceToMain(t, g)
	castFree(t, g, me, "Bear Cub", "Creature — Bear")
	if n := namedOnBattlefield(g, "Merfolk", me.ID); n != 0 {
		t.Fatalf("a creature spell made %d Merfolk", n)
	}
	castFree(t, g, me, "Shock", "Instant")
	castFree(t, g, opp, "Their Shock", "Instant")
	if n := namedOnBattlefield(g, "Merfolk", me.ID); n != 1 {
		t.Errorf("%d Merfolk tokens, want 1 (an opponent's spell makes none)", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Merfolk" && (c.Power != 1 || c.Toughness != 1 || !c.IsToken()) {
			t.Errorf("Merfolk token is %d/%d", c.Power, c.Toughness)
		}
	}
}

func TestMscM10NamorKingPumpsOthersAttackingAPlayerWithMoreLife(t *testing.T) {
	g := newCatalogGame(t)
	me, rich, poor := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	namor := b27Push(g, me.ID, "Namor, Atlantean King", "Legendary Creature — Mutant Merfolk Noble", mscM10NamorKingOracle, "{1}{W}{U}", 2, 2, "W", "U")
	toRich := b12Creature(g, me.ID, "Attacker One", "Creature — Bear", 2, 2)
	toPoor := b12Creature(g, me.ID, "Attacker Two", "Creature — Bear", 2, 2)
	me.Life, rich.Life, poor.Life = 30, 40, 20
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []struct {
		id, at uuid.UUID
	}{{namor, rich.ID}, {toRich, rich.ID}, {toPoor, poor.ID}} {
		if err := g.DeclareAttacker(a.id, a.at); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, toRich); p != 4 {
		t.Errorf("the other attacker of the richer player has power %d, want 4", p)
	}
	if p, _ := effectivePTOf(t, g, toPoor); p != 2 {
		t.Errorf("an attacker of someone else has power %d, want 2", p)
	}
	if p, _ := effectivePTOf(t, g, namor); p != 2 {
		t.Errorf("Namor himself has power %d, want 2 (other creatures only)", p)
	}
}

func TestMscM10NamorKingDoesNotTriggerAgainstAPlayerWithNoMoreLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	namor := b27Push(g, me.ID, "Namor, Atlantean King", "Legendary Creature — Mutant Merfolk Noble", mscM10NamorKingOracle, "{1}{W}{U}", 2, 2, "W", "U")
	other := b12Creature(g, me.ID, "Other Attacker", "Creature — Bear", 2, 2)
	me.Life, opp.Life = 30, 30
	declareAttack(t, g, opp.ID, namor, other)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, other); p != 2 {
		t.Errorf("an equal-life opponent still gave +2/+0: power %d", p)
	}
}

// --- Namor, Scourge of the Seas ---

func TestMscM10NamorSeasGivesFlyingToOtherMerfolkWithAPlusOneCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	namor := b27Push(g, me.ID, "Namor, Scourge of the Seas", "Legendary Creature — Mutant Merfolk Villain", mscM10NamorSeasOracle, "{3}{U}{U}", 4, 4, "U")
	assertKeywords(t, g, namor, "flying")
	withCounter := b12Creature(g, me.ID, "Counter Merfolk", "Creature — Merfolk", 1, 1)
	without := b12Creature(g, me.ID, "Bare Merfolk", "Creature — Merfolk", 1, 1)
	bear := b12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 1, 1)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{withCounter, bear, theirs} {
			_ = g.AddCounterForEffect(id, game.CounterPlusOne, 1)
		}
	})
	assertKeywords(t, g, withCounter, "flying")
	for name, id := range map[string]uuid.UUID{"a Merfolk without a counter": without, "a non-Merfolk with a counter": bear, "an opponent's Merfolk": theirs} {
		if containsString(effectiveAbilities(t, g, id), "flying") {
			t.Errorf("%s has flying", name)
		}
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == withCounter {
				delete(g.Battlefield.Cards[i].Counters, game.CounterPlusOne)
			}
		}
		g.BumpLayerVersionForTest()
	})
	g.RunStateChecksForTest()
	if containsString(effectiveAbilities(t, g, withCounter), "flying") {
		t.Error("the Merfolk kept flying after its counter left")
	}
}

func TestMscM10NamorSeasConnivesATargetCreatureYouControlAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b27Push(g, me.ID, "Namor, Scourge of the Seas", "Legendary Creature — Mutant Merfolk Villain", mscM10NamorSeasOracle, "{3}{U}{U}", 4, 4, "U")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mscM10EmptyHand(g, me)
	spare := handCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	advanceTo(t, g, game.StepBeginCombat)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("an opponent's creature was a legal target")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	mscM10Connive(t, g, me.ID, spare)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("the targeted creature has %d counters, want 1", got)
	}
}

// --- Thing Swing ---

func mscM10Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle, mana string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	toMainForCost(t, g)
	id := handCardFull(me, name, typeLine, cost, oracle, nil)
	if mana != "" {
		floatMana(t, g, me, mana)
	}
	params.Strict = true
	return id, g.CastSpell(me.ID, id, params)
}

func mscM10TwoClauses(source, victim uuid.UUID) []game.TargetRef {
	return []game.TargetRef{
		{Kind: game.TargetCard, ID: source, Slot: 0},
		{Kind: game.TargetCard, ID: victim, Slot: 1},
	}
}

func TestMscM10ThingSwingCostsTwoLessOnlyWhenItTargetsAHeroYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	hero := b12Creature(g, me.ID, "My Hero", "Creature — Human Hero", 3, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	theirHero := b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 3, 3)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
	price := func(refs []game.TargetRef) int {
		return selfPriced(t, g, me, mscM10ThingSwingOracle, "Instant", "{3}{G}", game.ZoneHand, refs).ManaValue()
	}
	if got := price(mscM10TwoClauses(hero, ogre)); got != 2 {
		t.Errorf("with my Hero as the source: %d, want 2", got)
	}
	if got := price(mscM10TwoClauses(bear, ogre)); got != 4 {
		t.Errorf("with a non-Hero source: %d, want 4", got)
	}
	if got := price(mscM10TwoClauses(bear, theirHero)); got != 4 {
		t.Errorf("at an opposing Hero: %d, want 4 (it must be a Hero you control)", got)
	}
}

func TestMscM10ThingSwingDealsTwiceTheHeroesPowerForTwoMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hero := b12Creature(g, me.ID, "My Hero", "Creature — Human Hero", 3, 3)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 6, 6)
	if _, err := mscM10Cast(t, g, me, "Thing Swing", "Instant", "{3}{G}", mscM10ThingSwingOracle, "{C}{G}",
		game.CastSpellParams{Targets: mscM10TwoClauses(hero, ogre)}); err != nil {
		t.Fatalf("cast for {1}{G} at a Hero: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("6 damage (twice a 3-power Hero) did not destroy the 6/6")
	}
}

func TestMscM10ThingSwingRefusesTwoManaAtANonHeroAndDealsNothingIfTheSourceLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 6, 6)
	if _, err := mscM10Cast(t, g, me, "Thing Swing", "Instant", "{3}{G}", mscM10ThingSwingOracle, "{C}{G}",
		game.CastSpellParams{Targets: mscM10TwoClauses(bear, ogre)}); err == nil {
		t.Fatal("cast for {1}{G} at a non-Hero source")
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	hero := b12Creature(g2, me2.ID, "My Hero", "Creature — Human Hero", 3, 3)
	ogre2 := b12Creature(g2, opp2.ID, "Their Ogre", "Creature — Ogre", 6, 6)
	if _, err := mscM10Cast(t, g2, me2, "Thing Swing", "Instant", "{3}{G}", mscM10ThingSwingOracle, "{C}{G}",
		game.CastSpellParams{Targets: mscM10TwoClauses(hero, ogre2)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g2.WithWriteLock(func() { _ = g2.DestroyPermanentForEffect(hero) })
	g2.RunStateChecksForTest()
	passPriorityAroundTable(t, g2)
	if damageMarkedOn(g2, ogre2) != 0 {
		t.Error("a Hero that left before resolution still dealt damage")
	}
}

// --- Terrific Team-Up ---

func TestMscM10TerrificTeamUpCostsTwoLessWithAPermanentOfManaValueFourOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	if got := selfPricedMV(t, g, me, mscM10TeamUpOracle, "Instant", "{3}{G}"); got != 4 {
		t.Errorf("with no big permanent: %d, want 4", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 3, Toughness: 3})
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if got := selfPricedMV(t, g, me, mscM10TeamUpOracle, "Instant", "{3}{G}"); got != 4 {
		t.Errorf("with only a mana value 3 permanent and a land: %d, want 4", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Four Drop", TypeLine: "Artifact", ManaCost: "{4}"})
	if got := selfPricedMV(t, g, me, mscM10TeamUpOracle, "Instant", "{3}{G}"); got != 2 {
		t.Errorf("with a mana value 4 permanent: %d, want 2", got)
	}
}

func TestMscM10TerrificTeamUpPumpsOneOrTwoCreaturesAndEachDealsItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 3, 3)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 7, 7)
	pushPermanent(g, me.ID, game.Card{Name: "Four Drop", TypeLine: "Artifact", ManaCost: "{4}"})
	refs := []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 0},
		{Kind: game.TargetCard, ID: ogre, Slot: 1},
	}
	if _, err := mscM10Cast(t, g, me, "Terrific Team-Up", "Instant", "{3}{G}", mscM10TeamUpOracle, "{C}{G}", game.CastSpellParams{Targets: refs}); err != nil {
		t.Fatalf("cast for {1}{G} with a mana value 4 permanent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, a); p != 3 {
		t.Errorf("Bear A has power %d, want 3", p)
	}
	if p, _ := effectivePTOf(t, g, b); p != 4 {
		t.Errorf("Bear B has power %d, want 4", p)
	}
	// 3 + 4 = 7 damage destroys the 7/7.
	if g.Battlefield.Contains(ogre) {
		t.Error("the combined 7 damage did not destroy the 7/7")
	}
}

func TestMscM10TerrificTeamUpRefusesTwoManaWithoutABigPermanentAndAThirdTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 7, 7)
	refs := []game.TargetRef{{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: ogre, Slot: 1}}
	if _, err := mscM10Cast(t, g, me, "Terrific Team-Up", "Instant", "{3}{G}", mscM10TeamUpOracle, "{C}{G}", game.CastSpellParams{Targets: refs}); err == nil {
		t.Fatal("cast for {1}{G} with no permanent of mana value 4 or more")
	}
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	x := b12Creature(g2, me2.ID, "Bear X", "Creature — Bear", 2, 2)
	y := b12Creature(g2, me2.ID, "Bear Y", "Creature — Bear", 2, 2)
	z := b12Creature(g2, me2.ID, "Bear Z", "Creature — Bear", 2, 2)
	ogre2 := b12Creature(g2, g2.Seats[1].ID, "Their Ogre", "Creature — Ogre", 7, 7)
	three := []game.TargetRef{
		{Kind: game.TargetCard, ID: x, Slot: 0}, {Kind: game.TargetCard, ID: y, Slot: 0}, {Kind: game.TargetCard, ID: z, Slot: 0},
		{Kind: game.TargetCard, ID: ogre2, Slot: 1},
	}
	if err := castCatalogSpellErr(t, g2, "Terrific Team-Up", "Instant", mscM10TeamUpOracle, three); err == nil {
		t.Fatal("cast with three creatures of mine as the first clause")
	}
}

func TestMscM10TerrificTeamUpStillPumpsWhenTheRecipientIsGoneAndSkipsASourceThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 3, 3)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 7, 7)
	refs := []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 0},
		{Kind: game.TargetCard, ID: ogre, Slot: 1},
	}
	castCatalogSpell(t, g, "Terrific Team-Up", "Instant", mscM10TeamUpOracle, refs)
	// In response: Bear B dies.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(b) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, a); p != 3 {
		t.Errorf("the surviving source has power %d, want 3", p)
	}
	if got := damageMarkedOn(g, ogre); got != 3 {
		t.Errorf("the Ogre took %d, want 3 (only Bear A dealt damage)", got)
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	c := b12Creature(g2, me2.ID, "Bear C", "Creature — Bear", 2, 2)
	ogre2 := b12Creature(g2, opp2.ID, "Their Ogre", "Creature — Ogre", 7, 7)
	castCatalogSpell(t, g2, "Terrific Team-Up", "Instant", mscM10TeamUpOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: c, Slot: 0}, {Kind: game.TargetCard, ID: ogre2, Slot: 1},
	})
	g2.WithWriteLock(func() { _ = g2.DestroyPermanentForEffect(ogre2) })
	g2.RunStateChecksForTest()
	passPriorityAroundTable(t, g2)
	if p, _ := effectivePTOf(t, g2, c); p != 3 {
		t.Errorf("with the recipient gone the source has power %d, want 3 (the pump is independent)", p)
	}
}

// --- Tippy-Toe, Terrific Partner ---

func mscM10Foods(g *game.Game, owner uuid.UUID) int { return namedOnBattlefield(g, "Food", owner) }

func TestMscM10TippyToeAddsOneFoodToEachTokenCreationOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Tippy-Toe, Terrific Partner", "Legendary Creature — Squirrel Hero", mscM10TippyToeOracle, false)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 2) })
	if n := namedOnBattlefield(g, "Treasure", me.ID); n != 2 {
		t.Errorf("%d Treasures, want the 2 asked for", n)
	}
	if n := mscM10Foods(g, me.ID); n != 1 {
		t.Fatalf("%d Foods after one creation, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 3) })
	if n := mscM10Foods(g, me.ID); n != 2 {
		t.Errorf("%d Foods after a second creation, want 2 (one per instruction, not per token)", n)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TreasureToken(), 1) })
	if n := mscM10Foods(g, opp.ID); n != 0 {
		t.Errorf("an opponent's token creation gave %d Foods, want 0", n)
	}
}

func TestMscM10TwoTippyToesEachAddAFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Tippy-Toe, Terrific Partner", "Legendary Creature — Squirrel Hero", mscM10TippyToeOracle, false)
	pushCatalogPermanent(g, me.ID, "Tippy-Toe, Terrific Partner", "Legendary Creature — Squirrel Hero", mscM10TippyToeOracle, false)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	if n := mscM10Foods(g, me.ID); n != 2 {
		t.Errorf("%d Foods with two Tippy-Toes, want 2", n)
	}
}

func TestMscM10TippyToeDrawsAtYourEndStepOnlyIfYouGainedLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Tippy-Toe, Terrific Partner", "Legendary Creature — Squirrel Hero", mscM10TippyToeOracle, false)
	seedLibrary(me, "Draw A", "Draw B")
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards after gaining life, want 1", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	advanceToMain(t, g2)
	pushCatalogPermanent(g2, me2.ID, "Tippy-Toe, Terrific Partner", "Legendary Creature — Squirrel Hero", mscM10TippyToeOracle, false)
	seedLibrary(me2, "Draw A", "Draw B")
	hand2 := me2.Hand.Size()
	g2.WithWriteLock(func() { _ = g2.ChangePlayerLifeForEffect(uuid.Nil, me2.ID, -3) })
	advanceToEndStepOf(t, g2, 0)
	passPriorityAroundTable(t, g2)
	if got := me2.Hand.Size() - hand2; got != 0 {
		t.Errorf("drew %d cards after only losing life, want 0", got)
	}
}
