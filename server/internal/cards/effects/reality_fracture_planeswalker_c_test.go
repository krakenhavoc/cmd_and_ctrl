package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_planeswalker_c_test.go — slice fra-planeswalker-c.

const (
	pwcMindMeandererOracle   = "18d1765b-3b60-4dda-93a6-d02a4e53f185"
	pwcRefuteDestinyOracle   = "2a18b3f0-8a78-46d8-b60f-3cd634062f1b"
	pwcSilenceTheEchoOracle  = "43654381-c55c-4d6d-abfd-12932ceb5a4a"
	pwcStingerquillOracle    = "26e20ae5-4059-4d63-9db5-6e12421d1aba"
	pwcTamOracle             = "c92c7744-0c92-4a36-9ada-4ffca812009e"
	pwcTerminalCritOracle    = "213b0814-f768-4408-9cb7-6f5960bcbb7a"
	pwcTeyoDiamondOracle     = "a3bfbe1a-7831-45bd-9e42-790c28f6054c"
	pwcTeyoLightshieldOracle = "77a9fac6-3c26-4456-9e0b-f5839a3dcae2"
	pwcTheoristOracle        = "89a6e876-0b00-4671-9652-766fd7ef9bf2"
	pwcTheorixAnnexOracle    = "00ad5200-5179-4d1b-8bb3-a09e1b31be58"
	pwcTomikOracle           = "7a501f7e-eec8-45f7-9ac3-483fd8f5ca5e"
	pwcTransformativeOracle  = "8ec56e2a-380f-4de7-9fb7-52de803008fc"
	pwcVigorbloomOracle      = "fcb35d42-bcd2-427c-9640-ff3602b79912"
	pwcVindictiveOracle      = "d6b478d1-5015-49ba-b1aa-cbc2b08107a6"
	pwcWinterOracle          = "84eaac2e-74ce-4a16-9faa-409df7c58eb3"
	pwcWrathOracle           = "c1994d6a-984d-4b75-9198-1f1822eab6eb"
	pwcYourFateOracle        = "28a362c2-0c99-48e0-b9cb-485a6c1f350b"
)

// pwcCreature seats a creature with colours and a printed mana cost.
func pwcCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int, manaCost string, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", Power: power, Toughness: toughness,
		ManaCost: manaCost, Colors: colors, Owner: owner, Controller: owner,
	})
}

// pwcWalker seats a planeswalker with the given loyalty.
func pwcWalker(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, loyalty int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, Colors: colors,
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

// pwcCastFromHand casts a card with explicit P/T from the active seat.
func pwcCastFromHand(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
	me.Hand.PushTop(c)
	toMainForCost(t, g)
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, params)
}

// pwcAnswerOwnPermanents answers a "choose from your own permanents" prompt.
func pwcAnswerOwnPermanents(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, chooser)
	if c == nil {
		t.Fatal("no own-permanents prompt is open")
	}
	if err := g.ResolveOwnPermanents(c.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
}

func pwcLife(g *game.Game, i int) int { return g.Seats[i].Life }

// --- Terminal Criticism ------------------------------------------------

func TestTerminalCriticismDestroysBlueOrRedAndGainsOne(t *testing.T) {
	for _, color := range []string{"U", "R"} {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		victim := pwcCreature(g, opp.ID, "Victim", 2, 2, "", color)
		life := pwcLife(g, 0)
		castCatalogSpell(t, g, "Terminal Criticism", "Instant", pwcTerminalCritOracle, rfCardTarget(victim))
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(victim) {
			t.Errorf("a %s creature should be destroyed", color)
		}
		if got := pwcLife(g, 0); got != life+1 {
			t.Errorf("life = %d, want %d", got, life+1)
		}
	}
}

func TestTerminalCriticismHitsABluePlaneswalkerAndRefusesGreen(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	walker := pwcWalker(g, opp.ID, "Blue Walker", "Planeswalker — Test", "", 3, "U")
	castCatalogSpell(t, g, "Terminal Criticism", "Instant", pwcTerminalCritOracle, rfCardTarget(walker))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("a blue planeswalker should be destroyed")
	}

	green := pwcCreature(g, opp.ID, "Green", 2, 2, "", "G")
	if err := castCatalogSpellErr(t, g, "Terminal Criticism", "Instant", pwcTerminalCritOracle, rfCardTarget(green)); err == nil {
		t.Error("a green creature is not a legal target")
	}
}

// --- Refute Destiny ----------------------------------------------------

func TestRefuteDestinyExilesGreenOrBlueAndSurveils(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	me := g.Seats[0]
	seedLibrary(me, "Top Card")
	victim := pwcCreature(g, opp.ID, "Victim", 2, 2, "", "G")
	castCatalogSpell(t, g, "Refute Destiny", "Sorcery", pwcRefuteDestinyOracle, rfCardTarget(victim))
	passPriorityAroundTable(t, g)
	if !inExile(g, victim) {
		t.Error("a green creature should be exiled")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID) == nil {
		t.Error("Refute Destiny should ask a surveil question")
	}
}

func TestRefuteDestinyRefusesARedCreature(t *testing.T) {
	g := newCatalogGame(t)
	red := pwcCreature(g, g.Seats[1].ID, "Red", 2, 2, "", "R")
	if err := castCatalogSpellErr(t, g, "Refute Destiny", "Sorcery", pwcRefuteDestinyOracle, rfCardTarget(red)); err == nil {
		t.Error("a red creature is not a legal target")
	}
}

// --- Your Fate Ends Here -----------------------------------------------

func TestYourFateEndsHereNeedsManaValueThree(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := pwcCreature(g, opp.ID, "Small", 2, 2, "{1}{G}")
	big := pwcCreature(g, opp.ID, "Big", 3, 3, "{2}{G}")
	if err := castCatalogSpellErr(t, g, "Your Fate Ends Here", "Instant", pwcYourFateOracle, rfCardTarget(small)); err == nil {
		t.Error("a mana value 2 creature is not a legal target")
	}
	seedLibrary(g.Seats[0], "Top Card")
	castCatalogSpell(t, g, "Your Fate Ends Here", "Instant", pwcYourFateOracle, rfCardTarget(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a mana value 3 creature should be destroyed")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceSurveil, g.Seats[0].ID) == nil {
		t.Error("the spell should ask a surveil question")
	}
}

// --- Silence the Echo --------------------------------------------------

func TestSilenceTheEchoPaysEitherBranch(t *testing.T) {
	for i, key := range []string{"sacrifice", "mana"} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := pwcCreature(g, opp.ID, "Victim", 2, 2, "")
		params := game.CastSpellParams{CostBranch: branch(i), Targets: rfCardTarget(victim)}
		var fodder uuid.UUID
		if key == "sacrifice" {
			fodder = pwcCreature(g, me.ID, "Fodder", 1, 1, "")
			params.SacrificeIDs = []uuid.UUID{fodder}
		}
		if _, err := castWithTapParams(t, g, "Silence the Echo", "Sorcery", "", pwcSilenceTheEchoOracle, params); err != nil {
			t.Fatalf("%s branch: %v", key, err)
		}
		if key == "sacrifice" && g.Battlefield.Contains(fodder) {
			t.Error("the sacrificed creature should be gone")
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(victim) {
			t.Errorf("%s branch: the target should be destroyed", key)
		}
	}
}

func TestSilenceTheEchoCanSacrificeAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pwcCreature(g, opp.ID, "Victim", 2, 2, "")
	walker := pwcWalker(g, me.ID, "Fodder Walker", "Planeswalker — Test", "", 3)
	params := game.CastSpellParams{CostBranch: branch(0), SacrificeIDs: []uuid.UUID{walker}, Targets: rfCardTarget(victim)}
	if _, err := castWithTapParams(t, g, "Silence the Echo", "Sorcery", "", pwcSilenceTheEchoOracle, params); err != nil {
		t.Fatalf("sacrificing a planeswalker: %v", err)
	}
	if g.Battlefield.Contains(walker) {
		t.Error("the planeswalker should be sacrificed")
	}
}

// --- Wrath of the Bloodmane -------------------------------------------

func TestWrathOfTheBloodmaneCostsOneLessWithALegend(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, pwcWrathOracle, "Instant", "{2}{R}"); got != 3 {
		t.Errorf("no legend: cost %d, want 3", got)
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Test", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	if got := selfPricedMV(t, g, me, pwcWrathOracle, "Instant", "{2}{R}"); got != 2 {
		t.Errorf("with a legend: cost %d, want 2", got)
	}
}

func TestWrathOfTheBloodmaneDealsFourToACreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pwcCreature(g, opp.ID, "Bear", 4, 4, "")
	walker := pwcWalker(g, opp.ID, "Walker", "Planeswalker — Test", "", 5)
	castCatalogSpell(t, g, "Wrath of the Bloodmane", "Instant", pwcWrathOracle, rfCardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("4 damage should kill a 4/4")
	}
	castCatalogSpell(t, g, "Wrath of the Bloodmane", "Instant", pwcWrathOracle, rfCardTarget(walker))
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, walker); got != 1 {
		t.Errorf("walker loyalty = %d, want 1", got)
	}
	if err := castCatalogSpellErr(t, g, "Wrath of the Bloodmane", "Instant", pwcWrathOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Error("a player is not a legal target")
	}
}

// --- Vindictive Triumph ------------------------------------------------

func TestVindictiveTriumphReturnsASmallPermanentTappedThenExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pwcCreature(g, opp.ID, "Small", 2, 2, "{1}{G}")
	castCatalogSpell(t, g, "Vindictive Triumph", "Instant", pwcVindictiveOracle, rfCardTarget(victim))
	passPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Small" {
			back = c
		}
	}
	if back == nil {
		t.Fatal("a mana value 2 creature should return to the battlefield")
	}
	if back.Controller != me.ID || !back.Tapped {
		t.Errorf("returned creature controller %v tapped %v, want me and tapped", back.Controller == me.ID, back.Tapped)
	}
	returned := back.InstanceID
	if returned == victim {
		t.Error("it should return as a new object")
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(returned) {
		t.Error("the returned creature should be exiled at the next end step")
	}
}

func TestVindictiveTriumphLeavesABigPermanentExiled(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := pwcCreature(g, opp.ID, "Big", 4, 4, "{3}{G}")
	castCatalogSpell(t, g, "Vindictive Triumph", "Instant", pwcVindictiveOracle, rfCardTarget(big))
	passPriorityAroundTable(t, g)
	if !inExile(g, big) {
		t.Error("a mana value 4 creature should stay exiled")
	}
	if onBattlefieldByName(g, "Big") {
		t.Error("a mana value 4 creature must not return")
	}
}

func TestVindictiveTriumphTokenIsGone(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Vindictive Triumph", "Instant", pwcVindictiveOracle, rfCardTarget(tok))
	passPriorityAroundTable(t, g)
	if onBattlefieldByName(g, "Spirit") {
		t.Error("an exiled token ceases to exist and must not return")
	}
}

// --- Teyo --------------------------------------------------------------

func TestTeyosGrantTheKeywordAndCounters(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, keyword string
		cost                  string
		power, toughness      int
	}{
		{"Teyo, Diamondblade Mage", pwcTeyoDiamondOracle, "deathtouch", "{3}{B}", 3, 1},
		{"Teyo, Lightshield Expert", pwcTeyoLightshieldOracle, "hexproof", "{1}{W}", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			bear := pwcCreature(g, me.ID, "Bear", 2, 2, "")
			walker := pwcWalker(g, me.ID, "Walker", "Planeswalker — Test", "", 3)
			// A creature target gets the keyword and a +1/+1 counter.
			if _, err := pwcCastFromHand(t, g, game.Card{
				Name: tc.name, TypeLine: "Legendary Creature — Human", OracleID: tc.oracle,
				Power: tc.power, Toughness: tc.toughness, Keywords: []string{"flash"},
			}, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			passPriorityAroundTable(t, g)
			b04WaitForPick(t, g, me.ID)
			pickCard(t, g, me.ID, bear)
			passPriorityAroundTable(t, g)
			if !hasAbility(effectiveAbilities(t, g, bear), tc.keyword) {
				t.Errorf("the bear should have %s", tc.keyword)
			}
			if got := rfCounters(t, g, bear); got != 1 {
				t.Errorf("+1/+1 counters on the bear = %d, want 1", got)
			}
			if got := loyaltyCount(g, walker); got != 3 {
				t.Errorf("an untargeted walker's loyalty = %d, want 3", got)
			}
		})
	}
}

func TestTeyoPutsALoyaltyCounterOnAPlaneswalkerAndNoPlusCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := pwcWalker(g, me.ID, "Walker", "Planeswalker — Test", "", 3)
	if _, err := pwcCastFromHand(t, g, game.Card{
		Name: "Teyo, Lightshield Expert", TypeLine: "Legendary Creature — Human Cleric", OracleID: pwcTeyoLightshieldOracle,
		Power: 1, Toughness: 1,
	}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, walker)
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, walker); got != 4 {
		t.Errorf("loyalty = %d, want 4", got)
	}
	if got := rfCounters(t, g, walker); got != 0 {
		t.Errorf("a planeswalker that is not a creature got %d +1/+1 counters", got)
	}
}

// --- Tam, the Possibility ---------------------------------------------

func TestTamDiscountsOnlyYourPlaneswalkerSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Tam, the Possibility", "Legendary Creature — Gorgon Wizard", pwcTamOracle, false)
	if got := priceInHand(t, g, me, "Walker", "Legendary Planeswalker — Test", "{2}{U}{U}"); got != 3 {
		t.Errorf("my planeswalker spell: %d, want 3", got)
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{2}{G}"); got != 3 {
		t.Errorf("my creature spell: %d, want 3 (untouched)", got)
	}
	if got := priceInHand(t, g, opp, "Walker", "Legendary Planeswalker — Test", "{2}{U}{U}"); got != 4 {
		t.Errorf("their planeswalker spell: %d, want 4 (undiscounted)", got)
	}
}

func TestTamProliferatesOncePerPlaneswalkerType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	tam := pushCatalogPermanent(g, me.ID, "Tam, the Possibility", "Legendary Creature — Gorgon Wizard", pwcTamOracle, false)
	// Two types (Jace, Gideon) across three walkers.
	pwcWalker(g, me.ID, "Jace A", "Legendary Planeswalker — Jace", "", 3)
	pwcWalker(g, me.ID, "Jace B", "Legendary Planeswalker — Jace", "", 3)
	gideon := pwcWalker(g, me.ID, "Gideon", "Legendary Planeswalker — Gideon", "", 3)
	g.Seats[0].ManaPool = append(g.Seats[0].ManaPool,
		game.ManaToken{Color: "W", Source: uuid.New()}, game.ManaToken{Color: "U", Source: uuid.New()}, game.ManaToken{Color: "B", Source: uuid.New()}, game.ManaToken{Color: "R", Source: uuid.New()}, game.ManaToken{Color: "G", Source: uuid.New()})
	if err := g.ActivateCatalogAbility(me.ID, tam, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i := 0; i < 2; i++ {
		c := proliferatePrompt(g)
		if c == nil {
			t.Fatalf("proliferate prompt %d never opened", i+1)
		}
		answerProliferate(t, g, gideon)
		passPriorityAroundTable(t, g)
	}
	if c := proliferatePrompt(g); c != nil {
		t.Error("a third proliferate opened; two planeswalker types means exactly two")
	}
	if got := loyaltyCount(g, gideon); got != 5 {
		t.Errorf("Gideon loyalty = %d, want 5 after two proliferates", got)
	}
}

// --- Tomik, Orzhov Lawmage --------------------------------------------

func TestTomikGrantsFlyingToACreatureWithAPlusCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	tomik := pushCatalogPermanent(g, me.ID, "Tomik, Orzhov Lawmage", "Legendary Creature — Human Advisor", pwcTomikOracle, false)
	plain := pwcCreature(g, me.ID, "Plain", 2, 2, "")
	counted := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	if err := g.ActivateCatalogAbility(me.ID, tomik, 0, game.ActivateAbilityParams{Targets: cardRefs(plain)}); err == nil {
		t.Error("a creature with no +1/+1 counter is not a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, tomik, 0, game.ActivateAbilityParams{Targets: cardRefs(counted)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasAbility(effectiveAbilities(t, g, counted), "flying") {
		t.Error("the countered creature should gain flying")
	}
}

// --- Winter, Tormented Loner ------------------------------------------

func castWinter(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id, err := pwcCastFromHand(t, g, game.Card{
		Name: "Winter, Tormented Loner", TypeLine: "Legendary Creature — Human Warlock", OracleID: pwcWinterOracle,
		Power: 0, Toughness: 3,
	}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast Winter: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestWinterSacrificeMakesEachOpponentSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := pwcCreature(g, me.ID, "Fodder", 1, 1, "")
	opp := map[int]uuid.UUID{}
	for i := 1; i < len(g.Seats); i++ {
		opp[i] = pwcCreature(g, g.Seats[i].ID, "Theirs", 2, 2, "")
	}
	castWinter(t, g)
	pwcAnswerOwnPermanents(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the chosen creature should be sacrificed")
	}
	for i := 1; i < len(g.Seats); i++ {
		answerSacrifice(t, g, g.Seats[i].ID, opp[i])
		if g.Battlefield.Contains(opp[i]) {
			t.Errorf("seat %d should have sacrificed its creature", i)
		}
	}
}

func TestWinterDeclineDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := pwcCreature(g, me.ID, "Fodder", 1, 1, "")
	theirs := pwcCreature(g, g.Seats[1].ID, "Theirs", 2, 2, "")
	castWinter(t, g)
	pwcAnswerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(theirs) {
		t.Error("declining must sacrifice nothing, mine or theirs")
	}
	if sacrificeChoiceFor(g, g.Seats[1].ID) != nil {
		t.Error("no opponent should be asked to sacrifice")
	}
}

func TestWinterGetsPowerFromCreatureAndPlaneswalkerCardsInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	winter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Winter, Tormented Loner", TypeLine: "Legendary Creature — Human Warlock",
		OracleID: pwcWinterOracle, Power: 0, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	pushGraveyardTyped(me, "Dead Bear", "Creature — Bear")
	pushGraveyardTyped(me, "Dead Walker", "Planeswalker — Test")
	pushGraveyardTyped(me, "Dead Spell", "Sorcery")
	pwcCreature(g, me.ID, "Bump", 1, 1, "") // a battlefield entry invalidates the cached layers
	if got := effectivePower(t, g, winter); got != 2 {
		t.Errorf("power = %d, want 2 (one creature card, one planeswalker card)", got)
	}
}

// --- Mind Meanderer ----------------------------------------------------

func TestMindMeandererFightsAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	theirs := pwcCreature(g, g.Seats[1].ID, "Theirs", 3, 3, "")
	meanderer, err := pwcCastFromHand(t, g, game.Card{
		Name: "Mind Meanderer", TypeLine: "Creature — Bird Fish Illusion", OracleID: pwcMindMeandererOracle,
		Power: 4, Toughness: 4,
	}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("4 damage should kill the 3/3 it fought")
	}
	if got := damageMarkedOn(g, meanderer); got != 3 {
		t.Errorf("Mind Meanderer took %d damage, want 3", got)
	}
}

func TestMindMeandererHasVigilanceOnlyWithAJace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	meanderer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mind Meanderer", TypeLine: "Creature — Bird Fish Illusion",
		OracleID: pwcMindMeandererOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	if hasAbility(effectiveAbilities(t, g, meanderer), "vigilance") {
		t.Error("no Jace: no vigilance")
	}
	pwcWalker(g, me.ID, "Jace", "Legendary Planeswalker — Jace", "", 3)
	if !hasAbility(effectiveAbilities(t, g, meanderer), "vigilance") {
		t.Error("with a Jace planeswalker: vigilance")
	}
}

// --- The Theorist, Jace Beleren ---------------------------------------

func TestTheoristPlusOneMakesAnIllusion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	jace := pwcWalker(g, me.ID, "The Theorist, Jace Beleren", "Legendary Planeswalker — Jace", pwcTheoristOracle, 3)
	b16Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{})
	if onBattlefieldNamed(g, "Illusion") != 1 {
		t.Error("the +1 should make one Illusion token")
	}
	if got := loyaltyCount(g, jace); got != 4 {
		t.Errorf("loyalty = %d, want 4", got)
	}
}

func TestTheoristMinusTwoBouncesOnePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	jace := pwcWalker(g, me.ID, "The Theorist, Jace Beleren", "Legendary Planeswalker — Jace", pwcTheoristOracle, 3)
	a1 := pwcCreature(g, g.Seats[1].ID, "A1", 2, 2, "")
	a2 := pwcCreature(g, g.Seats[1].ID, "A2", 2, 2, "")
	b1 := pwcCreature(g, g.Seats[2].ID, "B1", 2, 2, "")
	mine := pwcCreature(g, me.ID, "Mine", 2, 2, "")
	if err := g.ActivateCatalogAbility(me.ID, jace, 1, game.ActivateAbilityParams{Targets: cardRefs(a1, a2)}); err == nil {
		t.Error("two targets under the same opponent must be refused")
	}
	if err := g.ActivateCatalogAbility(me.ID, jace, 1, game.ActivateAbilityParams{Targets: cardRefs(mine)}); err == nil {
		t.Error("my own creature is not a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, jace, 1, game.ActivateAbilityParams{Targets: cardRefs(a1, b1)}); err != nil {
		t.Fatalf("one per opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a1) || g.Battlefield.Contains(b1) {
		t.Error("both chosen permanents should be bounced")
	}
	if !g.Battlefield.Contains(a2) || !g.Battlefield.Contains(mine) {
		t.Error("the others must stay")
	}
	if got := loyaltyCount(g, jace); got != 1 {
		t.Errorf("loyalty = %d, want 1", got)
	}
}

func TestTheoristMinusSixDrawsThreeThenCountersEqualToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	seedLibrary(me, "A", "B", "C", "D", "E")
	jace := pwcWalker(g, me.ID, "The Theorist, Jace Beleren", "Legendary Planeswalker — Jace", pwcTheoristOracle, 6)
	c1 := pwcCreature(g, me.ID, "C1", 1, 1, "")
	c2 := pwcCreature(g, me.ID, "C2", 1, 1, "")
	theirs := pwcCreature(g, g.Seats[1].ID, "Theirs", 1, 1, "")
	before := handSize(me)
	b16Activate(t, g, me.ID, jace, 2, game.ActivateAbilityParams{})
	x := handSize(me)
	if x != before+3 {
		t.Fatalf("hand = %d, want %d", x, before+3)
	}
	for _, id := range []uuid.UUID{c1, c2} {
		if got := rfCounters(t, g, id); got != x {
			t.Errorf("counters on my creature = %d, want %d", got, x)
		}
	}
	if got := rfCounters(t, g, theirs); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

func TestTheoristDrawsOnEachOpponentsDrawStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pwcWalker(g, me.ID, "The Theorist, Jace Beleren", "Legendary Planeswalker — Jace", pwcTheoristOracle, 3)
	seedLibrary(me, "A", "B", "C")
	before := handSize(me)
	advanceToDrawStepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != before+1 {
		t.Errorf("hand = %d, want %d (one card in an opponent's draw step)", got, before+1)
	}
}

// --- the Annex lands ----------------------------------------------------

func TestAnnexLandsEnterTappedUnlessYouControlAPlaneswalker(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Stingerquill Annex", pwcStingerquillOracle},
		{"Theorix Annex", pwcTheorixAnnexOracle},
		{"Transformative Commons", pwcTransformativeOracle},
		{"Vigorbloom Annex", pwcVigorbloomOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := playLandFromHand(t, g, tc.name, tc.oracle)
			if c, ok := battlefieldCard(g, id); !ok || !c.Tapped {
				t.Error("without a planeswalker the land enters tapped")
			}
			if n := tapEventsFor(g, id); n != 0 {
				t.Errorf("%d tap events; it must ENTER tapped", n)
			}

			g2 := newCatalogGame(t)
			me2 := g2.Seats[g2.Turn.ActiveSeat]
			pwcWalker(g2, me2.ID, "Walker", "Planeswalker — Test", "", 3)
			id2 := playLandFromHand(t, g2, tc.name, tc.oracle)
			if c, ok := battlefieldCard(g2, id2); !ok || c.Tapped {
				t.Error("with a planeswalker the land enters untapped")
			}
			_ = me
		})
	}
}

func TestAnnexLandsAnOpponentsPlaneswalkerDoesNotCount(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pwcWalker(g, opp.ID, "Their Walker", "Planeswalker — Test", "", 3)
	id := playLandFromHand(t, g, "Theorix Annex", pwcTheorixAnnexOracle)
	if c, ok := battlefieldCard(g, id); !ok || !c.Tapped {
		t.Error("an opponent's planeswalker does not untap the land")
	}
}
