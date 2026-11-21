package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b7_cards_test.go — Avatar: The Last Airbender batch 7 (#4459).

const (
	avB7ScrollOracle     = "882d032b-db36-401a-8075-a9fabc90c53d"
	avB7GraspOracle      = "520b2799-66bd-40bf-9bc0-9cd3aed321f5"
	avB7LiesOracle       = "f56c6c35-d09b-4bee-a685-98e9afed13d1"
	avB7AzulaOracle      = "89f96032-5cba-404d-82b0-38d2be963e4b"
	avB7MerchantsOracle  = "0e6609b4-f359-4bf1-a06e-f4317a0e7b3c"
	avB7RioterOracle     = "5193172c-9024-409e-bd9e-0387971d65fe"
	avB7ColonyOracle     = "29fd050c-dd1d-4e1b-b3cc-f7d8d18e4dcb"
	avB7InspectorOracle  = "5d950fd4-8956-4b1a-ab0f-ab31bcffce4e"
	avB7CrawlerOracle    = "bac94fc9-808b-4dc9-b405-408dd2aa6194"
	avB7CatGatorOracle   = "349e6b9e-40a8-49bc-9cd3-4a22f5363fa5"
	avB7RioterTypeLine   = "Creature — Human Rogue Ally"
	avB7ColonyTypeLine   = "Creature — Bird Insect"
	avB7CrawlerTypeLine  = "Creature — Spider Beast"
	avB7CatGatorTypeLine = "Creature — Fish Crocodile"
)

func TestAvatarB7CardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		avB7ScrollOracle:    "Waterbending Scroll",
		avB7GraspOracle:     "Watery Grasp",
		avB7LiesOracle:      "Azula Always Lies",
		avB7AzulaOracle:     "Azula, On the Hunt",
		avB7MerchantsOracle: "Beetle-Headed Merchants",
		avB7RioterOracle:    "Boiling Rock Rioter",
		avB7ColonyOracle:    "Buzzard-Wasp Colony",
		avB7InspectorOracle: "Callous Inspector",
		avB7CrawlerOracle:   "Canyon Crawler",
		avB7CatGatorOracle:  "Cat-Gator",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func avB7Named(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

func avB7Pool(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
}

// --- Waterbending Scroll --------------------------------------------------------

func TestAvatarB7WaterbendingScrollCostsOneLessPerIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	scroll := effectstest.PushCatalogPermanent(g, me.ID, "Waterbending Scroll", "Artifact", avB7ScrollOracle, false)
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")

	// Two Islands: {4}. Three mana is not enough.
	avB7Pool(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for {3} with two Islands")
	}
	if effectstest.IsTapped(g, scroll) {
		t.Fatal("a refused activation tapped the Scroll")
	}
	avB7Pool(t, g, me, "{C}")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate for {4} with two Islands: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if !effectstest.IsTapped(g, scroll) {
		t.Error("the {T} was not paid")
	}
}

func TestAvatarB7WaterbendingScrollIsFreeWithSixIslandsButStillTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	scroll := effectstest.PushCatalogPermanent(g, me.ID, "Waterbending Scroll", "Artifact", avB7ScrollOracle, false)
	for i := 0; i < 7; i++ {
		effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	}
	if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with seven Islands and no mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a second time while tapped")
	}
}

// --- Watery Grasp ---------------------------------------------------------------

func TestAvatarB7WateryGraspLocksAndWaterbendsTheCreatureIntoItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.TapForTest(t, g, theirs)

	grasp := effectstest.CastCatalogSpell(t, g, "Watery Grasp", "Enchantment — Aura", avB7GraspOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(grasp); !ok || c.AttachedTo.ID != theirs {
		t.Fatal("Watery Grasp is not attached to the creature")
	}

	// Its controller's untap step leaves it tapped.
	effectstest.AdvanceToPrecombatMainOf(t, g, (seat+1)%len(g.Seats))
	if !effectstest.IsTapped(g, theirs) {
		t.Fatal("the enchanted creature untapped during its controller's untap step")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	helpers := []uuid.UUID{
		effectstest.PushVanillaCreature(g, me.ID, "Helper A", 1, 1),
		effectstest.PushVanillaCreature(g, me.ID, "Helper B", 1, 1),
	}
	avB7Pool(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, grasp, 0, game.ActivateAbilityParams{
		WaterbendIDs: helpers, Strict: true,
	}); err != nil {
		t.Fatalf("Waterbend {5} with three mana and two taps: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the enchanted creature is still on the battlefield")
	}
	if !effectstest.LibraryContains(opp, theirs) {
		t.Error("the creature is not in its owner's library")
	}
	if g.Battlefield.Contains(grasp) {
		t.Error("the Aura stayed on the battlefield enchanting nothing")
	}
}

func TestAvatarB7WateryGraspWaterbendCannotBePaidShort(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	grasp := effectstest.CastCatalogSpell(t, g, "Watery Grasp", "Enchantment — Aura", avB7GraspOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	helpers := []uuid.UUID{effectstest.PushVanillaCreature(g, me.ID, "Helper", 1, 1)}
	avB7Pool(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, grasp, 0, game.ActivateAbilityParams{
		WaterbendIDs: helpers, Strict: true,
	}); err == nil {
		t.Fatal("three mana and one tap paid Waterbend {5}")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("a refused activation moved the creature")
	}
}

// --- Azula Always Lies ----------------------------------------------------------

func TestAvatarB7AzulaAlwaysLiesBothModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	effectstest.P1g3CastModal(t, g, "Azula Always Lies", "Instant — Lesson", avB7LiesOracle, []int{0, 1},
		effectstest.P1g3ModeCard(theirs, 0), effectstest.P1g3ModeCard(mine, 1))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("-1/-1 did not kill the 1/1")
	}
	if got := effectstest.CountersOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the second target, want 1", got)
	}
}

// Both modes may name the same creature (CR 115.3): a 1/1 that gets -1/-1
// and a +1/+1 counter survives as a 1/1.
func TestAvatarB7AzulaAlwaysLiesBothModesOnOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	effectstest.P1g3CastModal(t, g, "Azula Always Lies", "Instant — Lesson", avB7LiesOracle, []int{0, 1},
		effectstest.P1g3ModeCard(mine, 0), effectstest.P1g3ModeCard(mine, 1))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Fatal("the creature died; the counter should keep it alive")
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, mine); got != 1 {
		t.Errorf("power %d, want 1", got)
	}
}

func TestAvatarB7AzulaAlwaysLiesCounterModeAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	effectstest.P1g3CastModal(t, g, "Azula Always Lies", "Instant — Lesson", avB7LiesOracle, []int{1}, effectstest.P1g3ModeCard(mine, 0))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, mine); got != 3 {
		t.Errorf("power %d, want 3 (no -1/-1 from the mode not chosen)", got)
	}
}

// --- Azula, On the Hunt ---------------------------------------------------------

func TestAvatarB7AzulaOnTheHuntFirebendsLosesLifeAndMakesAClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	azula := effectstest.B27Push(g, me.ID, "Azula, On the Hunt", "Legendary Creature — Human Noble", avB7AzulaOracle, "{3}{B}", 4, 3, "B")
	life := effectstest.LifeOf(g, me.ID)
	effectstest.DeclareAttack(t, g, opp.ID, azula)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID) {
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.MkWantPool(t, me, map[string]int{"R": 2})
	if got := effectstest.LifeOf(g, me.ID); got != life-1 {
		t.Errorf("life %d, want %d", got, life-1)
	}
	if got := len(avB5CluesB7(g, me.ID)); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
	effectstest.MkAdvanceTo(t, g, game.StepEndCombat)
	effectstest.MkWantPool(t, me, map[string]int{"R": 2})
	effectstest.MkAdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.MkWantPool(t, me, map[string]int{})
}

func avB5CluesB7(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Clue" && effects.IsToken(c) && c.Controller == controller {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// --- Beetle-Headed Merchants ----------------------------------------------------

func TestAvatarB7BeetleHeadedMerchantsSacrificesToDrawAndGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	merchants := effectstest.B27Push(g, me.ID, "Beetle-Headed Merchants", "Creature — Human Citizen", avB7MerchantsOracle, "{4}{B}", 5, 4, "B")
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")

	effectstest.DeclareAttack(t, g, opp.ID, merchants)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !effectstest.HasID(prompt.ChooseCards, relic) || !effectstest.HasID(prompt.ChooseCards, bear) ||
		effectstest.HasID(prompt.ChooseCards, merchants) || effectstest.HasID(prompt.ChooseCards, land) {
		t.Fatalf("want a may-sacrifice over another creature or artifact: %+v", prompt)
	}
	hand := me.Hand.Size()
	effectstest.AnswerOwnPermanents(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if got := effectstest.CountersOn(g, merchants, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestAvatarB7BeetleHeadedMerchantsDecliningDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	merchants := effectstest.B27Push(g, me.ID, "Beetle-Headed Merchants", "Creature — Human Citizen", avB7MerchantsOracle, "{4}{B}", 5, 4, "B")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, merchants)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a declined sacrifice sacrificed the creature")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d", got, hand)
	}
	if got := effectstest.CountersOn(g, merchants, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters after declining, want 0", got)
	}
}

// --- Boiling Rock Rioter --------------------------------------------------------

func TestAvatarB7BoilingRockRioterTapsAnAllyToExileAndCastsTheAllyOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	rioter := effectstest.B27Push(g, me.ID, "Boiling Rock Rioter", avB7RioterTypeLine, avB7RioterOracle, "{2}{B}", 3, 3, "B")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == rioter {
			g.Battlefield.Cards[i].SummonedThisTurn = true
		}
	}
	myAlly := effectstest.ValigarmandaGraveCard(me, "My Ally", "Creature — Human Ally", "{R}")
	theirAlly := effectstest.ValigarmandaGraveCard(opp, "Their Ally", "Creature — Human Ally", "{R}")
	myBear := effectstest.ValigarmandaGraveCard(me, "My Bear", "Creature — Bear", "{R}")

	// The cost is not {T}: the summoning-sick Rioter taps itself.
	if err := g.ActivateCatalogAbility(me.ID, rioter, 0, game.ActivateAbilityParams{
		TapIDs:  []uuid.UUID{rioter},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: myAlly}},
		Strict:  true,
	}); err != nil {
		t.Fatalf("tap the Rioter itself to exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(myAlly) {
		t.Fatal("the targeted card was not exiled")
	}
	// A non-Ally cannot pay the cost.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, rioter, 0, game.ActivateAbilityParams{
		TapIDs:  []uuid.UUID{bear},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirAlly}},
		Strict:  true,
	}); err == nil {
		t.Fatal("a non-Ally paid the tap cost")
	}
	other := effectstest.B12Creature(g, me.ID, "Other Ally", "Creature — Human Ally", 1, 1)
	for _, id := range []uuid.UUID{theirAlly, myBear} {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(other) })
		if err := g.ActivateCatalogAbility(me.ID, rioter, 0, game.ActivateAbilityParams{
			TapIDs:  []uuid.UUID{other},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
			Strict:  true,
		}); err != nil {
			t.Fatalf("tap another Ally to exile: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}

	// Next turn: attack. Firebending adds {R}; only my own Ally card is
	// on offer, paid for with that {R}.
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.DeclareAttack(t, g, opp.ID, rioter)
	effectstest.PassPriorityAroundTable(t, g)
	fire := effectstest.PendingItemLabelled(g, "Boiling Rock Rioter — firebending 1", rioter)
	if fire == nil {
		t.Fatal("no firebending trigger queued")
	}
	effectstest.AnswerTriggerOrderResolvingFirst(t, g, fire.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 1})
	for _, id := range []uuid.UUID{theirAlly, myBear} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
			t.Errorf("cast %v, which is not an Ally card I own", id)
		}
	}
	if err := g.CastSpell(me.ID, myAlly, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast my exiled Ally with the firebending mana: %v", err)
	}
	if g.Exile.Contains(myAlly) {
		t.Error("the Ally is still in exile")
	}
}

func TestAvatarB7BoilingRockRioterOffersNothingWithoutAPermission(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Boiling Rock Rioter", avB7RioterTypeLine, avB7RioterOracle, false)
	ally := effectstest.ValigarmandaGraveCard(me, "My Ally", "Creature — Human Ally", "{R}")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, ally); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	avB7Pool(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, ally, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast an exiled Ally the Rioter never exiled")
	}
}

// --- Buzzard-Wasp Colony --------------------------------------------------------

func TestAvatarB7BuzzardWaspColonyMaySacrificeItselfToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	colony := effectstest.TdCastFromHand(t, g, game.Card{Name: "Buzzard-Wasp Colony", TypeLine: avB7ColonyTypeLine,
		OracleID: avB7ColonyOracle, ManaCost: "{3}{B}", Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, colony, "flying") {
		t.Error("no flying")
	}
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !effectstest.HasID(prompt.ChooseCards, relic) || !effectstest.HasID(prompt.ChooseCards, colony) {
		t.Fatalf("want a may-sacrifice over an artifact or creature, the Colony included: %+v", prompt)
	}
	hand := me.Hand.Size()
	effectstest.AnswerOwnPermanents(t, g, me.ID, colony)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(colony) {
		t.Error("the Colony was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestAvatarB7BuzzardWaspColonyTakesEveryKindOfCounterFromTheDead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colony := effectstest.B27Push(g, me.ID, "Buzzard-Wasp Colony", avB7ColonyTypeLine, avB7ColonyOracle, "{3}{B}", 2, 2, "B")
	mine := effectstest.B27Push(g, me.ID, "Mine", "Creature — Bear", "", "", 2, 2)
	effectstest.SetCounters(g, mine, map[string]int{game.CounterPlusOne: 2, "shield": 1})
	bare := effectstest.B27Push(g, me.ID, "Bare", "Creature — Bear", "", "", 2, 2)
	theirs := effectstest.B27Push(g, opp.ID, "Theirs", "Creature — Bear", "", "", 2, 2)
	effectstest.SetCounters(g, theirs, map[string]int{game.CounterPlusOne: 3})

	// No counters, or not mine: no trigger.
	effectstest.B27Kill(g, bare)
	effectstest.B27Kill(g, theirs)
	g.RunStateChecksForTest()
	if got := effectstest.TriggersOnStackFrom(g, colony); got != 0 {
		t.Fatalf("%d triggers from a bare creature and an opponent's, want 0", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(mine) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, colony, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if got := effectstest.CountersOn(g, colony, "shield"); got != 1 {
		t.Errorf("%d shield counters, want 1", got)
	}
}

// -1/-1 counters come along too (the ruling), and may kill the Colony.
func TestAvatarB7BuzzardWaspColonyTakesMinusCountersAndMayDieOfThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	colony := effectstest.B27Push(g, me.ID, "Buzzard-Wasp Colony", avB7ColonyTypeLine, avB7ColonyOracle, "{3}{B}", 2, 2, "B")
	weak := effectstest.B27Push(g, me.ID, "Weakened", "Creature — Bear", "", "", 5, 5)
	effectstest.SetCounters(g, weak, map[string]int{game.CounterMinusOne: 2})
	effectstest.B27Kill(g, weak)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(colony) {
		t.Errorf("the Colony survived two -1/-1 counters (%d on it)", effectstest.CountersOn(g, colony, game.CounterMinusOne))
	}
}

func TestAvatarB7BuzzardWaspColonyDoesNotTriggerOnItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colony := effectstest.B27Push(g, me.ID, "Buzzard-Wasp Colony", avB7ColonyTypeLine, avB7ColonyOracle, "{3}{B}", 2, 2, "B")
	other := effectstest.B27Push(g, me.ID, "Buzzard-Wasp Colony", avB7ColonyTypeLine, avB7ColonyOracle, "{3}{B}", 2, 2, "B")
	effectstest.SetCounters(g, colony, map[string]int{game.CounterPlusOne: 1})
	_ = opp
	effectstest.B27Kill(g, colony)
	g.RunStateChecksForTest()
	if got := effectstest.TriggersOnStackFrom(g, colony); got != 0 {
		t.Errorf("the dying Colony triggered for itself %d times", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, other, game.CounterPlusOne); got != 1 {
		t.Errorf("the other Colony has %d counters, want 1", got)
	}
}

// --- Callous Inspector ----------------------------------------------------------

func TestAvatarB7CallousInspectorDiesDamagesYouAndMakesAClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	inspector := effectstest.B27Push(g, me.ID, "Callous Inspector", "Creature — Human Soldier", avB7InspectorOracle, "{B}", 1, 1, "B")
	if !effectstest.HasEffectiveKeyword(t, g, inspector, "menace") {
		t.Error("no menace")
	}
	life := effectstest.LifeOf(g, me.ID)
	effectstest.B27Kill(g, inspector)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life-1 {
		t.Errorf("life %d, want %d", got, life-1)
	}
	if got := len(avB5CluesB7(g, me.ID)); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Canyon Crawler -------------------------------------------------------------

func TestAvatarB7CanyonCrawlerMakesFoodOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crawler := effectstest.TdCastFromHand(t, g, game.Card{Name: "Canyon Crawler", TypeLine: avB7CrawlerTypeLine,
		OracleID: avB7CrawlerOracle, ManaCost: "{4}{B}{B}", Power: 6, Toughness: 6}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, crawler, "deathtouch") {
		t.Error("no deathtouch")
	}
	if got := avB7Named(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food tokens, want 1", got)
	}
}

func TestAvatarB7CanyonCrawlerSwampcyclesForASwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Watery Grave", "Land — Island Swamp"),
	)
	swamp, island, grave := ids[0], ids[1], ids[2]
	id, _ := effectstest.CycleFromHand(t, g, "Canyon Crawler", avB7CrawlerTypeLine, avB7CrawlerOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Crawler is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 2 || effectstest.HasID(c.SearchCards, island) || !effectstest.HasID(c.SearchCards, swamp) || !effectstest.HasID(c.SearchCards, grave) {
		t.Fatalf("search offers %v, want the two Swamp cards", c.SearchCards)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{grave}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(grave) {
		t.Error("the Swamp card did not reach the hand")
	}
}

// --- Cat-Gator ------------------------------------------------------------------

func TestAvatarB7CatGatorDealsDamageEqualToSwampsWithLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	}
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	gator := effectstest.TdCastFromHand(t, g, game.Card{Name: "Cat-Gator", TypeLine: avB7CatGatorTypeLine,
		OracleID: avB7CatGatorOracle, ManaCost: "{6}{B}", Power: 3, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, gator, "lifelink") {
		t.Error("no lifelink")
	}
	myLife, theirLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != theirLife-3 {
		t.Errorf("opponent at %d, want %d (three Swamps)", got, theirLife-3)
	}
	if got := effectstest.LifeOf(g, me.ID); got != myLife+3 {
		t.Errorf("me at %d, want %d (lifelink)", got, myLife+3)
	}
}
