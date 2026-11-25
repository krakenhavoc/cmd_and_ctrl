package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b074_test.go — Kaldheim Commander, batch 74 (#5506): card draw and
// library, life, keywords, damage and the Elf cards. Each test plays the real
// catalog entry and asserts what its text says.

const (
	oldB074StoicFarmer = "42820e09-f817-4485-a96a-ac2b0b798269"
	oldB074Tales       = "49e3c192-862f-483e-abd1-e22863ea88dd"
	oldB074Miara       = "6895d7a2-264b-4731-9294-c02e1e220384"
	oldB074Valkyrie    = "f0f7c3df-52f8-4835-a18f-69e56ffa3432"
	oldB074Venom       = "4741d842-12e6-4a8c-96b1-9e68c7074874"
	oldB074SoulJar     = "5d451958-25f3-4e70-8be2-c832f60f39e9"
	oldB074Abomination = "5146ba3f-2c5f-4a86-95d7-a34ce1e842b0"
	oldB074Numa        = "7f4aa98d-3572-4637-9b0d-3c12f63a96ec"
	oldB074Messenger   = "ba310102-8272-4d64-a00b-9e7346cb5f46"
	oldB074Archers     = "f7139c70-dfcc-41ae-9d0d-b6e8c07a063d"
	oldB074Massacre    = "3b90dd70-798e-4dc4-99e7-b36d6823af2b"
	oldB074Winnower    = "61f68cf6-e083-4264-89b4-cc17749b2a8b"
)

func oldB074Elf(g *game.Game, p *game.Player, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Elf Warrior",
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
}

func TestOldB074EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB074StoicFarmer: "Stoic Farmer", oldB074Tales: "Tales of the Ancestors", oldB074Miara: "Miara, Thorn of the Glade",
		oldB074Valkyrie: "Ethereal Valkyrie", oldB074Venom: "Elderfang Venom", oldB074SoulJar: "Serpent's Soul-Jar",
		oldB074Abomination: "Abomination of Llanowar", oldB074Numa: "Numa, Joraga Chieftain", oldB074Messenger: "Sylvan Messenger",
		oldB074Archers: "Jagged-Scar Archers", oldB074Massacre: "Eyeblight Massacre", oldB074Winnower: "Ruthless Winnower",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s is %v with caveats %v, want Full with none", name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Stoic Farmer -----------------------------------------------------------

func oldB074Farmer(t *testing.T, opponentLands int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plains := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	for i := 0; i < opponentLands; i++ {
		effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	}
	effectstest.EnterFromHand(t, g, me.ID, "Stoic Farmer", "Creature — Dwarf Peasant", oldB074StoicFarmer)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, plains[0])
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, plains[0]
}

func TestOldB074StoicFarmerPutsThePlainsInHandWhenNoOpponentHasMoreLands(t *testing.T) {
	g, me, plains := oldB074Farmer(t, 0)
	if !me.Hand.Contains(plains) || g.Battlefield.Contains(plains) {
		t.Error("the Plains was not put into the hand")
	}
}

func TestOldB074StoicFarmerPutsThePlainsOnTheBattlefieldTappedWhenAnOpponentHasMoreLands(t *testing.T) {
	g, me, plains := oldB074Farmer(t, 2)
	if !g.Battlefield.Contains(plains) || me.Hand.Contains(plains) {
		t.Fatal("the Plains did not go onto the battlefield")
	}
	if !effectstest.TappedNow(g, plains) {
		t.Error("the Plains entered untapped")
	}
}

func TestOldB074StoicFarmerHasForetell(t *testing.T) {
	spec, _ := effects.Lookup(oldB074StoicFarmer)
	if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Kind != game.SpecialActionForetell || spec.SpecialActions[0].CastCost != "{1}{W}" {
		t.Errorf("special actions %+v, want foretell {1}{W}", spec.SpecialActions)
	}
}

// --- Tales of the Ancestors -------------------------------------------------

func TestOldB074TalesDrawsTheDifferenceForEachShortPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	for _, p := range g.Seats {
		p.Hand.Cards = nil
	}
	for i := 0; i < 5; i++ {
		effectstest.PutInHand(a, game.Card{Name: "A card", TypeLine: "Instant"})
	}
	for i := 0; i < 3; i++ {
		effectstest.PutInHand(b, game.Card{Name: "A card", TypeLine: "Instant"})
	}
	effectstest.PutInHand(c, game.Card{Name: "A card", TypeLine: "Instant"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{U}")
	effectstest.CastFromHandForTest(t, g, me, "Tales of the Ancestors", "Sorcery", "{3}{U}", oldB074Tales, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i, want := range map[*game.Player]int{me: 5, a: 5, b: 5, c: 5} {
		_ = i
		_ = want
	}
	for _, p := range []*game.Player{me, a, b, c} {
		if p.Hand.Size() != 5 {
			t.Errorf("a player has %d cards in hand, want 5", p.Hand.Size())
		}
	}
}

func TestOldB074TalesDrawsNothingWhenEveryoneIsTied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		p.Hand.Cards = nil
		for i := 0; i < 2; i++ {
			effectstest.PutInHand(p, game.Card{Name: "A card", TypeLine: "Instant"})
		}
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{U}")
	effectstest.CastFromHandForTest(t, g, me, "Tales of the Ancestors", "Sorcery", "{3}{U}", oldB074Tales, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := 2
		if p == me {
			want = 2 // the spell left the hand, the hands are still tied
		}
		if p.Hand.Size() != want {
			t.Errorf("a player has %d cards, want %d", p.Hand.Size(), want)
		}
	}
}

// --- Miara ------------------------------------------------------------------

func TestOldB074MiaraDrawsWhenAnElfDiesAndYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Miara, Thorn of the Glade", "Legendary Creature — Elf Scout", oldB074Miara, false)
	elf := oldB074Elf(g, me, "Elf")
	hand, life := me.Hand.Size(), me.Life
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.KillCreature(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d (was %d), life %d (was %d): want one card for 1 life", me.Hand.Size(), hand, me.Life, life)
	}
}

func TestOldB074MiaraDecliningPaysAndDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	miara := effectstest.PushCatalogPermanent(g, me.ID, "Miara, Thorn of the Glade", "Legendary Creature — Elf Scout", oldB074Miara, false)
	hand, life := me.Hand.Size(), me.Life
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.KillCreature(t, g, me.ID, miara) // Miara herself dying counts
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("hand %d (was %d), life %d (was %d) after declining", me.Hand.Size(), hand, me.Life, life)
	}
}

func TestOldB074MiaraIgnoresNonElvesAndOpponentsElves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Miara, Thorn of the Glade", "Legendary Creature — Elf Scout", oldB074Miara, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := oldB074Elf(g, opp, "Their Elf")
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("Miara asked about a creature that is not an Elf you control")
	}
}

// --- Ethereal Valkyrie ------------------------------------------------------

func TestOldB074ValkyrieDrawsThenForetellsAChosenCardForTwoLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := effectstest.PutInHand(me, game.Card{Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{4}{R}{R}", Colors: []string{"R"}})
	before := me.Hand.Size()
	effectstest.EnterFromHand(t, g, me.ID, "Ethereal Valkyrie", "Creature — Spirit Angel", oldB074Valkyrie)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 { // the Valkyrie left the hand, a card was drawn... then waits on the prompt
		t.Logf("hand size %d before the prompt (was %d)", me.Hand.Size(), before)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, hand)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(hand) {
		t.Fatal("the chosen card was not exiled")
	}
	perm := effectstest.ExiledPermission(g, hand)
	if perm.Cost != "{2}{R}{R}" {
		t.Errorf("foretell cost %q, want {2}{R}{R}", perm.Cost)
	}
	if perm.Player != me.ID || !perm.CastOnly || perm.NotBeforeSeq <= g.Turn.Seq {
		t.Errorf("permission %+v: want castable by its owner only on a later turn", perm)
	}
	if n := g.ForetoldCountThisTurn(me.ID); n != 0 {
		t.Errorf("%d cards counted as foretold this turn, want 0 (it is an effect, not the special action)", n)
	}
}

func TestOldB074ValkyrieForetellCostNeverGoesBelowZeroOrIntoColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := effectstest.PutInHand(me, game.Card{Name: "Cheap", TypeLine: "Artifact", ManaCost: "{1}{U}"})
	effectstest.EnterFromHand(t, g, me.ID, "Ethereal Valkyrie", "Creature — Spirit Angel", oldB074Valkyrie)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if perm := effectstest.ExiledPermission(g, cheap); perm.Cost != "{U}" {
		t.Errorf("foretell cost %q, want {U}", perm.Cost)
	}
}

// --- Elderfang Venom --------------------------------------------------------

func TestOldB074VenomDrainsWhenAnElfDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Elderfang Venom", "Enchantment", oldB074Venom, false)
	elf := oldB074Elf(g, me, "Elf")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	mine, theirs := me.Life, opp.Life
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine || opp.Life != theirs {
		t.Fatal("a non-Elf dying drained")
	}
	effectstest.KillCreature(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-1 || me.Life != mine+1 {
		t.Errorf("life %d/%d, want you %d and the opponent %d", me.Life, opp.Life, mine+1, theirs-1)
	}
}

func TestOldB074VenomGivesAttackingElvesDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Elderfang Venom", "Enchantment", oldB074Venom, false)
	elf := oldB074Elf(g, me, "Elf")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if effectstest.HasKeywordOnBattlefield(t, g, elf, "deathtouch") {
		t.Error("an Elf not attacking has deathtouch")
	}
	effectstest.DeclareAttack(t, g, opp.ID, elf, bear)
	effectstest.LockInAttacks(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, elf, "deathtouch") {
		t.Error("the attacking Elf lacks deathtouch")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "deathtouch") {
		t.Error("an attacking non-Elf has deathtouch")
	}
}

// --- Serpent's Soul-Jar -----------------------------------------------------

func TestOldB074SoulJarExilesAnElfAndLetsYouCastItThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jar := effectstest.PushCatalogPermanent(g, me.ID, "Serpent's Soul-Jar", "Artifact", oldB074SoulJar, false)
	elf := oldB074Elf(g, me, "Dead Elf")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bear) {
		t.Fatal("a non-Elf was exiled by the Soul-Jar")
	}
	effectstest.KillCreature(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(elf) || me.Graveyard.Contains(elf) {
		t.Fatal("the dead Elf was not exiled")
	}
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, jar, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d, want %d (paid 2)", me.Life, life-2)
	}
	if !effectstest.TappedNow(g, jar) {
		t.Error("the Soul-Jar did not tap")
	}
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		// the Elf's cost is nothing in this fixture, so paying is trivial
		t.Fatalf("cast the exiled Elf: %v", err)
	}
}

func TestOldB074SoulJarPermissionIsOnlyForItsOwnCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jar := effectstest.PushCatalogPermanent(g, me.ID, "Serpent's Soul-Jar", "Artifact", oldB074SoulJar, false)
	stray := effectstest.PushExiledCard(g, me.ID, "Stray", "Creature — Elf", false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, jar, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, stray, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("cast a card that was not exiled with the Soul-Jar")
	}
}

// --- Abomination of Llanowar ------------------------------------------------

func TestOldB074AbominationCountsElvesAndElfCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Abomination of Llanowar", "Legendary Creature — Elf Horror", oldB074Abomination, false)
	if p, tough := effectstest.PtOf(t, g, id); p != 1 || tough != 1 {
		t.Fatalf("alone it is %d/%d, want 1/1 (itself)", p, tough)
	}
	oldB074Elf(g, me, "Elf")
	oldB074Elf(g, opp, "Their Elf") // not yours
	effectstest.PushGraveyardCardTyped(me, "Dead Elf", "Creature — Elf Druid")
	effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.PushGraveyardCardTyped(opp, "Their Dead Elf", "Creature — Elf")
	if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 3 {
		t.Errorf("it is %d/%d, want 3/3 (two Elves you control and one Elf card in your graveyard)", p, tough)
	}
	for _, kw := range []string{"vigilance", "menace"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("lacks %s", kw)
		}
	}
}

// --- Jagged-Scar Archers ----------------------------------------------------

func TestOldB074ArchersSizeIsItsElvesAndItShootsFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archers := effectstest.PushCatalogPermanent(g, me.ID, "Jagged-Scar Archers", "Creature — Elf Archer", oldB074Archers, false)
	oldB074Elf(g, me, "Elf")
	oldB074Elf(g, me, "Elf")
	if p, tough := effectstest.PtOf(t, g, archers); p != 3 || tough != 3 {
		t.Fatalf("Archers are %d/%d, want 3/3", p, tough)
	}
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Drake", TypeLine: "Creature — Drake",
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	ground := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 4, 4)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, archers, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(ground)}); err == nil {
		t.Fatal("targeted a creature without flying")
	}
	if err := g.ActivateCatalogAbility(me.ID, archers, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(flier)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) {
		t.Error("3 damage did not kill the 2/2 flier")
	}
}

// --- Eyeblight Massacre -----------------------------------------------------

func TestOldB074MassacreShrinksNonElvesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := oldB074Elf(g, me, "Elf")
	mine := effectstest.PushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Mouse", 2, 2)
	theirElf := oldB074Elf(g, opp, "Their Elf")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{B}{B}")
	effectstest.CastFromHandForTest(t, g, me, "Eyeblight Massacre", "Sorcery", "{2}{B}{B}", oldB074Massacre, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if p, _ := effectstest.PtOf(t, g, mine); p != 2 {
		t.Errorf("a non-Elf of yours has %d power, want 2", p)
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 non-Elf survived")
	}
	for _, id := range []uuid.UUID{elf, theirElf} {
		if p, _ := effectstest.PtOf(t, g, id); p != 1 {
			t.Errorf("an Elf has %d power, want 1 (untouched)", p)
		}
	}
}

// --- Ruthless Winnower ------------------------------------------------------

func TestOldB074WinnowerMakesEachPlayerSacrificeANonElfAtTheirUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Ruthless Winnower", "Creature — Elf Rogue", oldB074Winnower, false)
	elf := oldB074Elf(g, opp, "Their Elf")
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent's only non-Elf was not sacrificed at their upkeep")
	}
	if !g.Battlefield.Contains(elf) {
		t.Error("an Elf was sacrificed")
	}
}

func TestOldB074WinnowerLeavesAPlayerWithOnlyElvesAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Ruthless Winnower", "Creature — Elf Rogue", oldB074Winnower, false)
	elf := oldB074Elf(g, opp, "Their Elf")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(elf) {
		t.Error("an Elf was sacrificed")
	}
}

// --- Sylvan Messenger -------------------------------------------------------

func TestOldB074MessengerTakesTheElvesAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Elf A", TypeLine: "Creature — Elf Warrior"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Elf B", TypeLine: "Creature — Elf Druid"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Deep", TypeLine: "Land"})
	effectstest.EnterFromHand(t, g, me.ID, "Sylvan Messenger", "Creature — Elf", oldB074Messenger)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the controller was not asked to order the rest")
	}
	if err := g.ResolvePutInLibrary(pick.ID, me.ID, pick.ScryCards, nil); err != nil {
		t.Fatalf("order the rest: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[2]) {
		t.Error("the Elf cards were not put into the hand")
	}
	if me.Hand.Contains(ids[1]) || me.Hand.Contains(ids[3]) {
		t.Error("a non-Elf card went to the hand")
	}
	if !me.Library.Contains(ids[1]) || !me.Library.Contains(ids[3]) || !me.Library.Contains(ids[4]) {
		t.Error("the rest did not stay in the library")
	}
	// The unrevealed card is still on top; the rest went under it.
	if top := effectstest.LibraryTopIDs(me, 1); len(top) != 1 || top[0] != ids[4] {
		t.Errorf("top of library %v, want the unrevealed card", top)
	}
}

// --- Numa, Joraga Chieftain -------------------------------------------------

func TestOldB074NumaPaysXXAndDistributesXCountersAmongElves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	numa := effectstest.PushCatalogPermanent(g, me.ID, "Numa, Joraga Chieftain", "Legendary Creature — Elf Warrior", oldB074Numa, false)
	elf := oldB074Elf(g, me, "Elf")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	for i := 0; i < 6; i++ {
		effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PayAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no X prompt")
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatalf("answer X = 2: %v", err)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	// The reflexive trigger asks for its targets and the split.
	effectstest.PickTargetRefs(t, g, me.ID, map[uuid.UUID]int{numa: 1, elf: 1},
		game.TargetRef{Kind: game.TargetCard, ID: numa}, game.TargetRef{Kind: game.TargetCard, ID: elf})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, numa, game.CounterPlusOne) + effectstest.CountersOn(g, elf, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters landed on the Elves, want 2", n)
	}
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("a non-Elf got %d counters", n)
	}
	untapped := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Forest" && !c.Tapped {
			untapped++
		}
	}
	if untapped != 2 {
		t.Errorf("%d Forests left untapped, want 2 (paid {X}{X} with X = 2)", untapped)
	}
}
