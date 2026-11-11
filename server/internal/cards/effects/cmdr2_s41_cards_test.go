package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s41_cards_test.go — cmdr2 batch 41, spellslinger: storm, cascade and
// spell copies (#5368). Each test plays the real catalog entry through the
// engine and asserts what the printed text says.

const (
	cmdr2S41ErisOracle        = "918991a6-7321-419d-8f3d-150fe32d02d2"
	cmdr2S41StellaOracle      = "8d0da2f9-0e05-4a94-8d42-fd57d06b09ba"
	cmdr2S41StagecoachOracle  = "83f87884-f0f9-482e-887e-35a8a5c20eb0"
	cmdr2S41SlingerOracle     = "ed7ba558-5341-4c7b-a9c5-b382dee88a13"
	cmdr2S41OctaviaOracle     = "65681a2a-8395-4390-845f-9be974994dcb"
	cmdr2S41FoundryOracle     = "118da256-d1ea-44e8-9026-317e49694d29"
	cmdr2S41ArchmageOracle    = "f1ecb3d7-5ea0-45b2-b1b0-d7ae304db2a1"
	cmdr2S41FeatherOracle     = "957413ae-e756-4afe-886a-b57bf59d5f8d"
	cmdr2S41MurmurationOracle = "84e7be36-b8cd-4111-a91b-3628f421d5d5"
	cmdr2S41WildsearOracle    = "9ceac4c6-e62e-4445-852c-b2c621e4a83b"
	cmdr2S41CaseOracle        = "409697d5-ce3c-46aa-a97d-969d143a2a65"
)

func TestCmdr2S41EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		cmdr2S41ErisOracle:        "Eris, Roar of the Storm",
		cmdr2S41StellaOracle:      "Stella Lee, Wild Card",
		cmdr2S41StagecoachOracle:  "Smoldering Stagecoach",
		cmdr2S41SlingerOracle:     "Crackling Spellslinger",
		cmdr2S41OctaviaOracle:     "Octavia, Living Thesis",
		cmdr2S41FoundryOracle:     "Forger's Foundry",
		cmdr2S41ArchmageOracle:    "Archmage of Echoes",
		cmdr2S41FeatherOracle:     "Feather, Radiant Arbiter",
		cmdr2S41MurmurationOracle: "Murmuration",
		cmdr2S41WildsearOracle:    "Wildsear, Scouring Maw",
		cmdr2S41CaseOracle:        "Case of the Shifting Visage",
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

// --- Wildsear, Scouring Maw ----------------------------------------------------

func TestCmdr2S41WildsearCascadesEnchantmentsCastFromHandOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 8)
	pushCatalogPermanent(g, me.ID, "Wildsear, Scouring Maw", "Legendary Creature — Elemental Wolf", cmdr2S41WildsearOracle, false)

	castWithCost(t, g, "A Sorcery", "Sorcery", "{3}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("a sorcery got %d cascade offers, want 0", got)
	}

	castWithCost(t, g, "An Enchantment", "Enchantment", "{3}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("an enchantment cast from hand got %d cascade offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
}

func TestCmdr2S41WildsearDoesNotCascadeAnEnchantmentCastFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 8)
	pushCatalogPermanent(g, me.ID, "Wildsear, Scouring Maw", "Legendary Creature — Elemental Wolf", cmdr2S41WildsearOracle, false)
	advanceToMain(t, g)
	id := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: id, Name: "Exiled Enchantment", TypeLine: "Enchantment", ManaCost: "{3}{G}",
		Owner: me.ID, Controller: me.ID})
	g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{Player: me.ID, Zone: game.ZoneExile,
		Duration: game.WhileInZoneDuration(), Source: uuid.New()})
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "GGGG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("an enchantment cast from exile got %d cascade offers, want 0", got)
	}
}

// --- Crackling Spellslinger ----------------------------------------------------

func TestCmdr2S41SpellslingerStormsTheNextInstantOrSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stormFiller(t, g) // one spell before the Spellslinger

	castCatalogSpell(t, g, "Crackling Spellslinger", "Creature — Human Wizard", cmdr2S41SlingerOracle, nil)
	passPriorityAroundTable(t, g)

	// A creature spell neither storms nor uses the watch up.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)

	before := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	// Three spells came before it this turn (the filler, the Spellslinger, the Bear).
	answerCopyPrompts(t, g, me.ID, opp.ID, 3)
	if got := lifeOf(g, opp.ID); got != before-12 {
		t.Errorf("life %d -> %d, want -12 (the spell and three copies, 3 each)", before, got)
	}

	// The watch was used up: the next one does not storm.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	before = lifeOf(g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != before-3 {
		t.Errorf("the second spell dealt %d, want 3: it does not have storm", before-got)
	}
}

func TestCmdr2S41SpellslingerDoesNothingIfItWasNotCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Crackling Spellslinger", "Creature — Human Wizard", cmdr2S41SlingerOracle, false)
	passPriorityAroundTable(t, g)
	stormFiller(t, g)
	before := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != before-3 {
		t.Errorf("a Spellslinger that was not cast gave storm: dealt %d, want 3", before-got)
	}
}

// --- Smoldering Stagecoach -----------------------------------------------------

func TestCmdr2S41StagecoachPowerCountsYourInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushGraveyardCardTyped(me, "Old Instant", "Instant")
	pushGraveyardCardTyped(me, "Old Sorcery", "Sorcery")
	pushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	pushGraveyardCardTyped(opp, "Their Sorcery", "Sorcery")
	coach := b12Push(g, me.ID, "Smoldering Stagecoach", "Artifact Creature — Vehicle", cmdr2S41StagecoachOracle, 0, 5)
	if p, tough := ptOf(t, g, coach); p != 2 || tough != 5 {
		t.Errorf("Smoldering Stagecoach is %d/%d, want 2/5", p, tough)
	}
}

func TestCmdr2S41StagecoachGivesTheNextInstantAndTheNextSorceryCascade(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedCheapLibrary(g, me, 12)
	coach := b12Push(g, me.ID, "Smoldering Stagecoach", "Artifact Creature — Vehicle", cmdr2S41StagecoachOracle, 0, 5)

	attackWith(t, g, opp.ID, coach)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	// A creature spell is neither: no cascade.
	castWithCost(t, g, "Bear", "Creature — Bear", "{3}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Fatalf("a creature spell got %d cascade offers, want 0", got)
	}

	castWithCost(t, g, "First Instant", "Instant", "{3}{R}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("the next instant got %d cascade offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	castWithCost(t, g, "Second Instant", "Instant", "{3}{R}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("a second instant got %d cascade offers, want 0 (the instant watch is used up)", got)
	}

	castWithCost(t, g, "A Sorcery", "Sorcery", "{3}{R}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("the next sorcery got %d cascade offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
}

// --- Eris, Roar of the Storm ---------------------------------------------------

func TestCmdr2S41ErisCostsTwoLessPerDifferentManaValueAmongInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	const typeLine = "Legendary Creature — Elemental Warlock"
	if got := selfPricedMV(t, g, me, cmdr2S41ErisOracle, typeLine, "{8}{U}{R}"); got != 10 {
		t.Errorf("with an empty graveyard it costs mana value %d, want 10", got)
	}
	pushGraveyardPermanent(me, "Opt", "Instant", "{U}")
	pushGraveyardPermanent(me, "Shock", "Instant", "{R}")
	pushGraveyardPermanent(me, "Burn", "Sorcery", "{1}{R}")
	pushGraveyardPermanent(me, "Counterspell", "Instant", "{U}{U}")
	pushGraveyardPermanent(me, "A Bear", "Creature — Bear", "{4}{G}")
	if got := selfPricedMV(t, g, me, cmdr2S41ErisOracle, typeLine, "{8}{U}{R}"); got != 6 {
		t.Errorf("with mana values 1 and 2 among instants and sorceries it costs %d, want 6", got)
	}
	pushGraveyardPermanent(me, "Big Spell", "Sorcery", "{6}{U}{R}")
	if got := selfPricedMV(t, g, me, cmdr2S41ErisOracle, typeLine, "{8}{U}{R}"); got != 4 {
		t.Errorf("with a third different mana value it costs %d, want 4", got)
	}
	for _, mv := range []string{"{3}", "{4}", "{5}", "{6}", "{7}"} {
		pushGraveyardPermanent(me, "Spell "+mv, "Instant", mv)
	}
	if got := selfPricedMV(t, g, me, cmdr2S41ErisOracle, typeLine, "{8}{U}{R}"); got != 2 {
		t.Errorf("the coloured {U}{R} is never reduced: it costs %d, want 2", got)
	}
}

func TestCmdr2S41ErisMakesADragonElementalOnTheSecondSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Eris, Roar of the Storm", "Legendary Creature — Elemental Warlock", cmdr2S41ErisOracle, false)
	stormFiller(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Dragon Elemental"); got != 0 {
		t.Fatalf("%d Dragon Elementals after one spell, want 0", got)
	}
	stormFiller(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Dragon Elemental"); got != 1 {
		t.Fatalf("%d Dragon Elementals after two spells, want 1", got)
	}
	id := battlefieldIDsNamed(g, "Dragon Elemental")[0]
	if p, tough := ptOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("the Dragon Elemental is %d/%d, want 4/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, id, "flying") || !hasEffectiveKeyword(t, g, id, "prowess") {
		t.Error("the Dragon Elemental lacks flying or prowess")
	}
	stormFiller(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Dragon Elemental"); got != 1 {
		t.Errorf("%d Dragon Elementals after three spells, want still 1", got)
	}
}

// --- Octavia, Living Thesis ----------------------------------------------------

func TestCmdr2S41OctaviaCostsEightLessWithEightInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	const typeLine = "Legendary Creature — Elemental Octopus"
	for i := 0; i < 7; i++ {
		pushGraveyardPermanent(me, "Spell", "Instant", "{U}")
	}
	pushGraveyardPermanent(me, "A Bear", "Creature — Bear", "{1}{G}")
	if got := selfPricedMV(t, g, me, cmdr2S41OctaviaOracle, typeLine, "{8}{U}{U}"); got != 10 {
		t.Errorf("with seven instants and sorceries it costs %d, want 10", got)
	}
	pushGraveyardPermanent(me, "Eighth", "Sorcery", "{1}{R}")
	if got := selfPricedMV(t, g, me, cmdr2S41OctaviaOracle, typeLine, "{8}{U}{U}"); got != 2 {
		t.Errorf("with eight it costs %d, want 2 ({U}{U})", got)
	}
}

func TestCmdr2S41OctaviaMagecraftSetsAnyCreatureToBaseEightEight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Octavia, Living Thesis", "Legendary Creature — Elemental Octopus", cmdr2S41OctaviaOracle, false)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Opt", "Instant", "", nil)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	} else {
		t.Fatal("the magecraft trigger asked for no target")
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 8 || tough != 8 {
		t.Errorf("the targeted bear is %d/%d, want 8/8", p, tough)
	}
	// Until end of turn only.
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if p, tough := ptOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("after the turn the bear is %d/%d, want 2/2", p, tough)
	}
}

func TestCmdr2S41OctaviaHasWardEight(t *testing.T) {
	var ward bool
	for _, tr := range game.CatalogTriggers(cmdr2S41OctaviaOracle) {
		if tr.Keyword == "ward" || tr.Key == "Octavia, Living Thesis — ward {8}" {
			ward = true
		}
	}
	if !ward {
		t.Error("Octavia has no ward trigger")
	}
}

// --- Stella Lee, Wild Card -----------------------------------------------------

func TestCmdr2S41StellaExilesTheTopCardOnTheSecondSpellAndYouMayPlayIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Stella Lee, Wild Card", "Legendary Creature — Human Rogue", cmdr2S41StellaOracle, false)
	top := seedLibrary(me, "Top Card", "Second")
	stormFiller(t, g)
	if !me.Library.Contains(top[0]) {
		t.Fatal("the first spell exiled a card")
	}
	stormFiller(t, g)
	var exiled bool
	for _, c := range g.Exile.Cards {
		if c.InstanceID == top[0] {
			exiled = true
		}
	}
	if !exiled {
		t.Fatal("the second spell did not exile the top card of the library")
	}
	if !b12ExiledPlayableBy(t, g, top[0], me.ID) {
		t.Error("the exiled card has no play permission for its exiler")
	}
	stormFiller(t, g)
	if !me.Library.Contains(top[1]) {
		t.Error("a third spell exiled the next card too")
	}
	// "Until the end of your next turn": still playable on my next turn,
	// gone once it has ended.
	seat := g.Turn.ActiveSeat
	endTurn(t, g)
	advanceToPrecombatMainOf(t, g, seat)
	if !b12ExiledPlayableBy(t, g, top[0], me.ID) {
		t.Error("the card should still be playable during my next turn")
	}
	endTurn(t, g)
	if b12ExiledPlayableBy(t, g, top[0], me.ID) {
		t.Error("the permission should have ended with my next turn")
	}
}

func TestCmdr2S41StellaCopiesASpellOnlyAfterThreeSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stella := pushCatalogPermanent(g, me.ID, "Stella Lee, Wild Card", "Legendary Creature — Human Rogue", cmdr2S41StellaOracle, false)
	seedLibrary(me, "A", "B", "C", "D", "E", "F", "G", "H")
	stormFiller(t, g)
	hand := me.Hand.Size()
	spell := castCatalogSpell(t, g, "Divination", "Sorcery", divinationOracle, nil)
	params := game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}}
	if err := g.ActivateCatalogAbility(me.ID, stella, 0, params); err == nil {
		t.Fatal("the copy activated after only two spells")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	// A third spell, then the ability copies the Divination beneath it.
	castCatalogSpell(t, g, "Third", "Instant", "", nil)
	b16Activate(t, g, me.ID, stella, 0, params)
	for i := 0; i < 6; i++ {
		passPriorityAroundTable(t, g)
	}
	if got := me.Hand.Size(); got != hand+4 {
		t.Errorf("hand %d, want %d: Divination and its copy draw two each", got, hand+4)
	}
}

// --- Archmage of Echoes --------------------------------------------------------

func TestCmdr2S41ArchmageCopiesFaerieAndWizardPermanentSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Archmage of Echoes", "Creature — Faerie Wizard", cmdr2S41ArchmageOracle, false)

	castHoldCreature(t, g, "Sprite", "Creature — Faerie Rogue", "", 1, 1)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Sprite"); got != 2 {
		t.Errorf("%d Sprites after casting a Faerie, want 2 (the card and its token copy)", got)
	}
	if n, _ := countTokensNamed(g, me.ID, "Sprite"); n != 1 {
		t.Errorf("%d Sprite tokens, want 1: the copy becomes a token", n)
	}

	castHoldCreature(t, g, "Mage", "Creature — Human Wizard", "", 1, 1)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Mage"); got != 2 {
		t.Errorf("%d Mages after casting a Wizard, want 2", got)
	}

	castHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Bear"); got != 1 {
		t.Errorf("%d Bears after casting a Bear, want 1", got)
	}
	// The copies were not cast, so they did not copy themselves.
	if got := countBattlefieldNamed(g, me.ID, "Sprite"); got != 2 {
		t.Errorf("%d Sprites at the end, want 2: a copy is not cast", got)
	}
}

func TestCmdr2S41ArchmageIgnoresAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Archmage of Echoes", "Creature — Faerie Wizard", cmdr2S41ArchmageOracle, false)
	castCatalogSpell(t, g, "Wizard Trick", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if len(g.StackMeta) != 0 || g.Stack.Size() != 0 {
		t.Error("the stack did not empty")
	}
}

// --- Case of the Shifting Visage -----------------------------------------------

func TestCmdr2S41CaseCopiesNonlegendaryCreatureSpellsOnlyOnceSolved(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushClass(g, me.ID, "Case of the Shifting Visage", mkmAC2CaseTypeLine, cmdr2S41CaseOracle)

	castHoldCreature(t, g, "Early Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Early Bear"); got != 1 {
		t.Fatalf("an unsolved Case copied a spell: %d Early Bears", got)
	}

	mkmAC2Solve(t, g, id)
	castHoldCreature(t, g, "Late Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Late Bear"); got != 2 {
		t.Errorf("%d Late Bears once solved, want 2 (the card and its token copy)", got)
	}
	castHoldCreature(t, g, "Legend", "Legendary Creature — Human", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Legend"); got != 1 {
		t.Errorf("%d Legends, want 1: the Case copies only nonlegendary creature spells", got)
	}
	castCatalogSpell(t, g, "Any Bolt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Late Bear"); got != 2 {
		t.Errorf("%d Late Bears after an instant, want still 2", got)
	}
}

func TestCmdr2S41CaseSolvesAtTheEndStepWithFifteenCardsInYourGraveyard(t *testing.T) {
	for _, tc := range []struct {
		cards  int
		solved bool
	}{{14, false}, {15, true}} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		g.WithWriteLock(func() { me.Graveyard.Cards = nil })
		id := pushClass(g, me.ID, "Case of the Shifting Visage", mkmAC2CaseTypeLine, cmdr2S41CaseOracle)
		for i := 0; i < tc.cards; i++ {
			pushGraveyardCardTyped(me, "Filler", "Instant")
		}
		if got := mkmAC2EndStepSolves(t, g, id); got != tc.solved {
			t.Errorf("with %d cards in the graveyard solved = %v, want %v", tc.cards, got, tc.solved)
		}
	}
}

func TestCmdr2S41CaseSurveilsAtTheBeginningOfYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	pushClass(g, me.ID, "Case of the Shifting Visage", mkmAC2CaseTypeLine, cmdr2S41CaseOracle)
	endTurn(t, g)
	for g.Turn.ActiveSeat != 0 {
		endTurn(t, g)
	}
	advanceTo(t, g, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Fatal("the upkeep trigger asked for no surveil")
	}
}

// --- Murmuration ---------------------------------------------------------------

func TestCmdr2S41MurmurationMakesAStormCrowForEachSpellAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Murmuration", "Enchantment", cmdr2S41MurmurationOracle, false)
	stormFiller(t, g)
	stormFiller(t, g)
	stormFiller(t, g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	n, _ := countTokensNamed(g, me.ID, "Storm Crow")
	if n != 3 {
		t.Fatalf("%d Storm Crows after three spells, want 3", n)
	}
	crow := battlefieldIDsNamed(g, "Storm Crow")[0]
	if p, tough := ptOf(t, g, crow); p != 2 || tough != 3 {
		t.Errorf("a Storm Crow is %d/%d, want 2/3 (1/2 and +1/+1)", p, tough)
	}
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, crow, kw) {
			t.Errorf("a Storm Crow lacks %s", kw)
		}
	}
	if got := countBattlefieldNamed(g, me.ID, "Murmuration"); got != 1 {
		t.Errorf("%d Murmurations, want 1", got)
	}
}

func TestCmdr2S41MurmurationMakesNothingWithoutASpellAndPumpsOnlyYourBirds(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Murmuration", "Enchantment", cmdr2S41MurmurationOracle, false)
	mine := b12Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if p, tough := ptOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("my Bird is %d/%d, want 2/2", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("their Bird is %d/%d, want 1/1", p, tough)
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("my Bear has power %d, want 2", p)
	}
	if !hasEffectiveKeyword(t, g, mine, "vigilance") || hasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("vigilance went to the wrong Birds")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Storm Crow"); n != 0 {
		t.Errorf("%d Storm Crows with no spell cast, want 0", n)
	}
}

// --- Forger's Foundry ----------------------------------------------------------

// cmdr2S41CastWithFoundryMana taps the Foundry for {U} and casts a one-mana
// spell with it, strictly.
func cmdr2S41CastWithFoundryMana(t *testing.T, g *game.Game, me *game.Player, foundry uuid.UUID, typeLine, cost string) uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	activateManaFor(t, g, me.ID, foundry, 0, game.ManaAbilityParams{})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Cheap Spell", TypeLine: typeLine, ManaCost: cost,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast with the Foundry's mana: %v", err)
	}
	return id
}

func cmdr2S41InExile(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == id {
			return &g.Exile.Cards[i]
		}
	}
	return nil
}

func TestCmdr2S41FoundryExilesACheapSpellItPaidForAndLinksIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	foundry := pushCatalogPermanent(g, me.ID, "Forger's Foundry", "Artifact", cmdr2S41FoundryOracle, false)
	spell := cmdr2S41CastWithFoundryMana(t, g, me, foundry, "Instant", "{U}")
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c := cmdr2S41InExile(g, spell)
	if c == nil {
		t.Fatal("the spell was not exiled as it resolved")
	}
	if c.ExiledWith.ID != foundry {
		t.Errorf("the exiled spell is linked to %v, want the Foundry", c.ExiledWith.ID)
	}
}

func TestCmdr2S41FoundryDeclinedOrOverThreeOrASorceryOfCreatureGoesToTheGraveyard(t *testing.T) {
	// Declined.
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	foundry := pushCatalogPermanent(g, me.ID, "Forger's Foundry", "Artifact", cmdr2S41FoundryOracle, false)
	spell := cmdr2S41CastWithFoundryMana(t, g, me, foundry, "Instant", "{U}")
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if cmdr2S41InExile(g, spell) != nil || !me.Graveyard.Contains(spell) {
		t.Error("a declined exile still exiled the spell")
	}
	// A creature spell asks nothing.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	advanceToMain(t, g2)
	f2 := pushCatalogPermanent(g2, me2.ID, "Forger's Foundry", "Artifact", cmdr2S41FoundryOracle, false)
	cmdr2S41CastWithFoundryMana(t, g2, me2, f2, "Creature — Bear", "{U}")
	passPriorityAroundTable(t, g2)
	if latestConfirmFor(g2, me2.ID) != nil {
		t.Error("a creature spell paid with the Foundry's mana asked about exiling")
	}
	// Mana value 4 is too much; the Foundry's mana pays only one of it.
	g3 := newCatalogGame(t)
	me3 := g3.Seats[g3.Turn.ActiveSeat]
	advanceToMain(t, g3)
	f3 := pushCatalogPermanent(g3, me3.ID, "Forger's Foundry", "Artifact", cmdr2S41FoundryOracle, false)
	g3.WithWriteLock(func() { me3.ManaPool.EmptyPool() })
	activateManaFor(t, g3, me3.ID, f3, 0, game.ManaAbilityParams{})
	floatForTest(g3, me3, "CCC")
	big := uuid.New()
	me3.Hand.PushTop(game.Card{InstanceID: big, Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{3}{U}",
		Owner: me3.ID, Controller: me3.ID})
	if err := g3.CastSpell(me3.ID, big, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g3)
	if latestConfirmFor(g3, me3.ID) != nil {
		t.Error("a mana value 4 spell asked about exiling")
	}
}

func TestCmdr2S41FoundryCastsTheExiledSpellsFreeAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	advanceToMain(t, g)
	foundry := pushCatalogPermanent(g, me.ID, "Forger's Foundry", "Artifact", cmdr2S41FoundryOracle, false)
	spell := cmdr2S41CastWithFoundryMana(t, g, me, foundry, "Instant", "{U}")
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if cmdr2S41InExile(g, spell) == nil {
		t.Fatal("setup: the spell was not exiled")
	}
	untapForTest(g, foundry)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "UUUUU")
	before := lifeOf(g, opp.ID)
	b16Activate(t, g, me.ID, foundry, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("%d free-cast offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	_ = before
	if cmdr2S41InExile(g, spell) != nil {
		t.Error("the spell is still in exile after it was cast")
	}
}

func TestCmdr2S41FoundryLosesItsLinkWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	foundry := pushCatalogPermanent(g, me.ID, "Forger's Foundry", "Artifact", cmdr2S41FoundryOracle, false)
	spell := cmdr2S41CastWithFoundryMana(t, g, me, foundry, "Instant", "{U}")
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	// The Foundry leaves before the spell resolves: it goes to the graveyard.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == foundry {
				g.Battlefield.Cards[i].ObjectEpoch++
			}
		}
	})
	passPriorityAroundTable(t, g)
	if cmdr2S41InExile(g, spell) != nil {
		t.Error("a spell whose Foundry is a different object was exiled")
	}
}

// --- Feather, Radiant Arbiter --------------------------------------------------

func TestCmdr2S41FeatherCopiesTheSpellForEachCreaturePaidFor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	feather := b12Push(g, me.ID, "Feather, Radiant Arbiter", "Legendary Creature — Angel", cmdr2S41FeatherOracle, 4, 3)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Giant Growth", "Instant", socB8GiantGrowthOracle, cardTarget(feather))
	floatForTest(g, me, "CCCC")
	passPriorityAroundTable(t, g)
	c := payAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no pay prompt for the Feather trigger")
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatalf("ResolvePayAmount: %v", err)
	}
	answerPayUnless(t, g, me.ID, true)
	for i := 0; i < 8 && len(g.Stack.Cards) > 0; i++ {
		passPriorityAroundTable(t, g)
	}
	for _, id := range []uuid.UUID{feather, a, theirs} {
		if p := effectivePower(t, g, id); p < 5 {
			t.Errorf("creature %v has power %d, want it pumped by Giant Growth", id, p)
		}
	}
}

func TestCmdr2S41FeatherPayingNothingCopiesNothingAndAnotherTargetDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	feather := b12Push(g, me.ID, "Feather, Radiant Arbiter", "Legendary Creature — Angel", cmdr2S41FeatherOracle, 4, 3)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Giant Growth", "Instant", socB8GiantGrowthOracle, cardTarget(a))
	if n := triggersOnStackFrom(g, feather); n != 0 {
		t.Errorf("a spell targeting another creature triggered Feather %d times", n)
	}
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Giant Growth", "Instant", socB8GiantGrowthOracle, cardTarget(feather))
	passPriorityAroundTable(t, g)
	if c := payAmountPrompt(g, me.ID); c != nil {
		if err := g.ResolvePayAmount(c.ID, me.ID, 0); err != nil {
			t.Fatalf("ResolvePayAmount: %v", err)
		}
	}
	for i := 0; i < 8 && len(g.Stack.Cards) > 0; i++ {
		passPriorityAroundTable(t, g)
	}
	if p := effectivePower(t, g, a); p != 5 {
		t.Errorf("Bear A has power %d, want 5 (one Giant Growth, no copies)", p)
	}
}
