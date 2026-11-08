package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_s2_cards_test.go — Wilds of Eldraine, the second S batch (#4471).
// Each test plays the real catalog entry and asserts what its text says.

const (
	woeS2ViperOracle      = scaldingViperOracleID
	woeS2ScarecrowOracle  = "cf7bb5d4-c33f-4879-a793-6d6f2a4d94be"
	woeS2SlingerOracle    = "2e8433a2-ad71-4d30-84fe-12b709197824"
	woeS2FaerieOracle     = "a17633aa-b1a7-4ae5-b051-6747975d8a2f"
	woeS2KeepguardOracle  = "89078485-c438-4035-95dc-dbe37fb13e50"
	woeS2SpriteOracle     = "b0632f1b-14e5-4534-b2c5-af9d557fdcc0"
	woeS2CovenOracle      = spellscornCovenOracleID
	woeS2CelebrantOracle  = "171150fb-dbe1-4a07-a7ac-5f45762a26b0"
	woeS2GingerOracle     = "2eb16535-718f-4b88-92bc-96429d796327"
	woeS2BearType         = "Creature — Bear"
	woeS2GingerTypeLine   = "Legendary Artifact Creature — Food Knight"
	woeS2SpriteTypeLine   = "Creature — Faerie Wizard"
	woeS2CelebrantTypeStr = "Creature — Dwarf Knight"
)

func TestWoeS2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		woeS2ViperOracle:        "Scalding Viper",
		woeS2ViperOracle + "#1": "Steam Clean",
		woeS2ScarecrowOracle:    "Scarecrow Guide",
		woeS2SlingerOracle:      "Skewer Slinger",
		woeS2FaerieOracle:       "Sleep-Cursed Faerie",
		woeS2KeepguardOracle:    "Slumbering Keepguard",
		woeS2SpriteOracle:       "Snaremaster Sprite",
		woeS2CovenOracle:        "Spellscorn Coven",
		woeS2CovenOracle + "#1": "Take It Back",
		woeS2CelebrantOracle:    "Stockpiling Celebrant",
		woeS2GingerOracle:       "Syr Ginger, the Meal Ender",
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

// --- fixtures ------------------------------------------------------------------

// woeS2CastInstant has p cast a vanilla instant with the given cost and
// settles it.
func woeS2CastInstant(t *testing.T, g *game.Game, p *game.Player, cost string) {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: "Some Instant", TypeLine: "Instant", ManaCost: cost, Owner: p.ID, Controller: p.ID})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast a %s instant: %v", cost, err)
	}
	passPriorityAroundTable(t, g)
}

// woeS2Block announces a block of attacker by blocker, as the lock-in does.
func woeS2Block(g *game.Game, actor, blocker, attacker uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: actor, CardID: blocker, Target: attacker, Amount: 1})
	})
}

// woeS2Enters puts a permanent onto the battlefield under owner's control
// and announces its entry.
func woeS2Enters(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	t.Helper()
	id := b12Permanent(g, owner, name, typeLine)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: owner})
	})
	passPriorityAroundTable(t, g)
	return id
}

// woeS2Activate activates a catalog ability of card with the given mana
// in p's pool and settles it.
func woeS2Activate(t *testing.T, g *game.Game, p *game.Player, card uuid.UUID, index int, colors ...string) {
	t.Helper()
	g.WithWriteLock(func() { p.ManaPool.EmptyPool() })
	b06AddMana(p, colors...)
	if err := g.ActivateCatalogAbility(p.ID, card, index, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func woeS2Tapped(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && c.Tapped
}

// --- Scalding Viper // Steam Clean ---------------------------------------------

func TestWoeS2ScaldingViperBurnsAnOpponentForACheapSpell(t *testing.T) {
	g := newCatalogGame(t)
	caster, owner := woeB1Seats(g)
	b12Push(g, owner.ID, "Scalding Viper", "Creature — Elemental Snake", woeS2ViperOracle, 2, 1)
	toMain(t, g)
	life := caster.Life
	woeS2CastInstant(t, g, caster, "{2}{R}")
	if caster.Life != life-1 {
		t.Fatalf("life %d after a mana value 3 spell, want %d", caster.Life, life-1)
	}
	woeS2CastInstant(t, g, caster, "{3}{R}")
	if caster.Life != life-1 {
		t.Errorf("life %d after a mana value 4 spell, want it unchanged at %d", caster.Life, life-1)
	}
}

func TestWoeS2ScaldingViperIgnoresItsControllersSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	b12Push(g, me.ID, "Scalding Viper", "Creature — Elemental Snake", woeS2ViperOracle, 2, 1)
	toMain(t, g)
	life := me.Life
	woeS2CastInstant(t, g, me, "{R}")
	if me.Life != life {
		t.Errorf("life %d after your own spell, want %d", me.Life, life)
	}
}

func TestWoeS2SteamCleanBouncesANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	card := woeB1Adventure(me, woeS2ViperOracle,
		game.Face{Name: "Scalding Viper", TypeLine: "Creature — Elemental Snake", ManaCost: "{1}{R}", Colors: []string{"R"}, Power: 2, Toughness: 1},
		game.Face{Name: "Steam Clean", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{U}", Colors: []string{"U"}})
	woeB1CastAdventure(t, g, me, card, cardRefs(relic))
	if g.Battlefield.Contains(relic) || !opp.Hand.Contains(relic) {
		t.Error("the artifact did not return to its owner's hand")
	}
}

// --- Scarecrow Guide -------------------------------------------------------------

func TestWoeS2ScarecrowGuideFiltersOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, _ := woeB1Seats(g)
	guide := b12Push(g, me.ID, "Scarecrow Guide", "Artifact Creature — Scarecrow", woeS2ScarecrowOracle, 2, 1)
	assertKeywords(t, g, guide, "reach")
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "C", 2)
	if err := g.ActivateManaAbility(me.ID, guide, 0, game.ManaAbilityParams{Colors: []string{"W"}}); err != nil {
		t.Fatalf("{1}: add {W}: %v", err)
	}
	mkWantPool(t, me, map[string]int{"C": 1, "W": 1})
	if err := g.ActivateManaAbility(me.ID, guide, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Error("a second activation this turn was allowed")
	}
}

// --- Skewer Slinger --------------------------------------------------------------

func TestWoeS2SkewerSlingerPokesWhatItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	slinger := b12Push(g, opp.ID, "Skewer Slinger", "Creature — Dwarf Knight", woeS2SlingerOracle, 1, 3)
	assertKeywords(t, g, slinger, "reach")
	attacker := b12Creature(g, me.ID, "Attacker", woeS2BearType, 2, 4)
	bystander := b12Creature(g, me.ID, "Bystander", woeS2BearType, 2, 4)
	woeS2Block(g, opp.ID, slinger, attacker)
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, attacker); got != 1 {
		t.Errorf("the blocked creature has %d damage, want 1", got)
	}
	if got := pr6Marked(g, bystander); got != 0 {
		t.Errorf("a creature outside the block has %d damage, want 0", got)
	}
}

func TestWoeS2SkewerSlingerPokesEachCreatureBlockingIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	slinger := b12Push(g, me.ID, "Skewer Slinger", "Creature — Dwarf Knight", woeS2SlingerOracle, 1, 3)
	first := b12Creature(g, opp.ID, "First Blocker", woeS2BearType, 2, 4)
	second := b12Creature(g, opp.ID, "Second Blocker", woeS2BearType, 2, 4)
	woeS2Block(g, opp.ID, first, slinger)
	woeS2Block(g, opp.ID, second, slinger)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{first, second} {
		if got := pr6Marked(g, id); got != 1 {
			t.Errorf("a blocker has %d damage, want 1", got)
		}
	}
	if got := pr6Marked(g, slinger); got != 0 {
		t.Errorf("the Slinger dealt itself %d damage", got)
	}
}

// --- Sleep-Cursed Faerie ---------------------------------------------------------

func TestWoeS2SleepCursedFaerieEntersStunnedAndUntapsOneCounterAtATime(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	faerie := castRingCreature(t, g, "Sleep-Cursed Faerie", "Creature — Faerie Wizard", woeS2FaerieOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(faerie) {
		t.Fatal("the Faerie did not resolve")
	}
	assertKeywords(t, g, faerie, "flying")
	if !woeS2Tapped(g, faerie) {
		t.Error("the Faerie entered untapped")
	}
	if n := countersOn(g, faerie, game.CounterStun); n != 3 {
		t.Fatalf("%d stun counters, want 3", n)
	}
	woeS2Activate(t, g, me, faerie, 0, "C", "U")
	if n := countersOn(g, faerie, game.CounterStun); n != 2 || !woeS2Tapped(g, faerie) {
		t.Errorf("after one untap: %d stun counters and tapped=%v, want 2 and tapped", n, woeS2Tapped(g, faerie))
	}
	woeS2Activate(t, g, me, faerie, 0, "C", "U")
	woeS2Activate(t, g, me, faerie, 0, "C", "U")
	woeS2Activate(t, g, me, faerie, 0, "C", "U")
	if n := countersOn(g, faerie, game.CounterStun); n != 0 || woeS2Tapped(g, faerie) {
		t.Errorf("after four untaps: %d stun counters and tapped=%v, want 0 and untapped", n, woeS2Tapped(g, faerie))
	}
}

func TestWoeS2SleepCursedFaerieHasWardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	faerie := pushCatalogPermanent(g, me.ID, "Sleep-Cursed Faerie", "Creature — Faerie Wizard", woeS2FaerieOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, faerie)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {2} payment for the opponent targeting the Faerie")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(faerie) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}
}

// --- Slumbering Keepguard --------------------------------------------------------

func TestWoeS2SlumberingKeepguardScriesWhenYourEnchantmentEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	pushCatalogPermanent(g, me.ID, "Slumbering Keepguard", "Creature — Human Knight", woeS2KeepguardOracle, false)
	woeS2Enters(t, g, opp.ID, "Their Charm", "Enchantment")
	if c := scryChoiceFor(g, me.ID); c != nil {
		t.Fatal("an opponent's enchantment made you scry")
	}
	woeS2Enters(t, g, me.ID, "My Charm", "Enchantment")
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("scry prompt %v, want one looking at one card", c)
	}
	answerScryKeepAll(t, g, me.ID)
}

func TestWoeS2SlumberingKeepguardGrowsPerEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := woeB1Seats(g)
	keep := b12Push(g, me.ID, "Slumbering Keepguard", "Creature — Human Knight", woeS2KeepguardOracle, 1, 1)
	b12Permanent(g, me.ID, "My Charm", "Enchantment")
	b12Permanent(g, me.ID, "My Other Charm", "Enchantment")
	b12Permanent(g, opp.ID, "Their Charm", "Enchantment")
	woeS2Activate(t, g, me, keep, 0, "C", "C", "W")
	if p, tough := effectivePTOf(t, g, keep); p != 3 || tough != 3 {
		t.Errorf("the Keepguard is %d/%d, want 3/3 with two enchantments", p, tough)
	}
}

// --- Snaremaster Sprite ----------------------------------------------------------

func TestWoeS2SnaremasterSpriteStunsWhenYouPay(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeS2BearType, 2, 2)
	sprite := castRingCreature(t, g, "Snaremaster Sprite", woeS2SpriteTypeLine, woeS2SpriteOracle, 1, 1)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, sprite, "flying")
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked for before the payment was made")
	}
	b06AddMana(me, "C", "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !woeS2Tapped(g, theirs) {
		t.Error("their creature is not tapped")
	}
	if n := countersOn(g, theirs, game.CounterStun); n != 1 {
		t.Errorf("stun counters %d, want 1", n)
	}
}

func TestWoeS2SnaremasterSpriteDecliningDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeS2BearType, 2, 2)
	castAndResolveCreature(t, g, "Snaremaster Sprite", woeS2SpriteTypeLine, woeS2SpriteOracle)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a target prompt appeared after declining")
	}
	if woeS2Tapped(g, theirs) || countersOn(g, theirs, game.CounterStun) != 0 {
		t.Error("declining still tapped or stunned their creature")
	}
}

// --- Spellscorn Coven // Take It Back --------------------------------------------

func TestWoeS2SpellscornCovenMakesEachOpponentDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	for _, p := range g.Seats {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
	coven := castRingCreature(t, g, "Spellscorn Coven", "Creature — Faerie Warlock", woeS2CovenOracle, 2, 3)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, coven, "flying")
	for _, p := range g.Seats {
		owed := discardOwed(g, p.ID)
		if p.ID == me.ID && owed != 0 {
			t.Errorf("you owe %d discards, want none", owed)
		}
		if p.ID != me.ID && owed != 1 {
			t.Errorf("an opponent owes %d discards, want 1", owed)
		}
	}
}

func TestWoeS2TakeItBackReturnsASpellToItsOwnersHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shock := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: shock, Name: "Shock", TypeLine: "Instant", OracleID: b5ShockOracle, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("cast Shock: %v", err)
	}
	card := woeB1Adventure(me, woeS2CovenOracle,
		game.Face{Name: "Spellscorn Coven", TypeLine: "Creature — Faerie Warlock", ManaCost: "{3}{B}", Colors: []string{"B"}, Power: 2, Toughness: 3},
		game.Face{Name: "Take It Back", TypeLine: "Instant — Adventure", ManaCost: "{2}{U}", Colors: []string{"U"}})
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(shock)}); err != nil {
		t.Fatalf("cast Take It Back: %v", err)
	}
	life := me.Life
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(shock) || !opp.Hand.Contains(shock) {
		t.Error("Shock is not back in its owner's hand")
	}
	if opp.Graveyard.Contains(shock) {
		t.Error("Shock was countered into the graveyard, not returned")
	}
	if me.Life != life {
		t.Errorf("life %d, want %d: Shock never resolved", me.Life, life)
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("the Adventure card is not in exile after it resolved")
	}
}

// --- Stockpiling Celebrant -------------------------------------------------------

func TestWoeS2StockpilingCelebrantReturnsAPermanentAndScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	relic := b12Permanent(g, me.ID, "My Relic", "Artifact")
	castAndResolveCreature(t, g, "Stockpiling Celebrant", woeS2CelebrantTypeStr, woeS2CelebrantOracle)
	pickTriggerTarget(t, g, me.ID, relic)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || !me.Hand.Contains(relic) {
		t.Error("the artifact did not return to your hand")
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %v, want one looking at two cards", c)
	}
	answerScryKeepAll(t, g, me.ID)
}

func TestWoeS2StockpilingCelebrantDecliningKeepsThePermanentAndSkipsTheScry(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	relic := b12Permanent(g, me.ID, "My Relic", "Artifact")
	castAndResolveCreature(t, g, "Stockpiling Celebrant", woeS2CelebrantTypeStr, woeS2CelebrantOracle)
	pickTriggerTarget(t, g, me.ID, relic)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(relic) {
		t.Error("declining returned the artifact")
	}
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("declining still scried")
	}
}

func TestWoeS2StockpilingCelebrantCannotTargetALandOrAnOpponentsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	land := b12Permanent(g, me.ID, "My Land", "Land")
	theirs := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	mine := b12Permanent(g, me.ID, "My Relic", "Artifact")
	castAndResolveCreature(t, g, "Stockpiling Celebrant", woeS2CelebrantTypeStr, woeS2CelebrantOracle)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range []uuid.UUID{land, theirs} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
			t.Error("an illegal permanent was accepted as the target")
		}
	}
	pickTriggerTarget(t, g, me.ID, mine)
}

// --- Syr Ginger, the Meal Ender --------------------------------------------------

func TestWoeS2SyrGingerHasItsKeywordsOnlyWhileAnOpponentHasAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	ginger := b12Push(g, me.ID, "Syr Ginger, the Meal Ender", woeS2GingerTypeLine, woeS2GingerOracle, 3, 1)
	b12Permanent(g, me.ID, "My Walker", "Legendary Planeswalker — Jace")
	for _, kw := range []string{"trample", "hexproof", "haste"} {
		if hasEffectiveKeyword(t, g, ginger, kw) {
			t.Errorf("Syr Ginger has %s with only your own planeswalker out", kw)
		}
	}
	b12Permanent(g, opp.ID, "Their Walker", "Legendary Planeswalker — Chandra")
	for _, kw := range []string{"trample", "hexproof", "haste"} {
		if !hasEffectiveKeyword(t, g, ginger, kw) {
			t.Errorf("Syr Ginger lacks %s while an opponent controls a planeswalker", kw)
		}
	}
}

func TestWoeS2SyrGingerGrowsAndScriesWhenAnotherArtifactOfYoursGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	ginger := b12Push(g, me.ID, "Syr Ginger, the Meal Ender", woeS2GingerTypeLine, woeS2GingerOracle, 3, 1)
	woeB1Destroy(t, g, b12Permanent(g, opp.ID, "Their Relic", "Artifact"))
	woeB1Destroy(t, g, b12Permanent(g, me.ID, "My Charm", "Enchantment"))
	if n := countersOn(g, ginger, game.CounterPlusOne); n != 0 || scryChoiceFor(g, me.ID) != nil {
		t.Fatalf("%d counters after an opponent's artifact and your enchantment died, want 0 and no scry", n)
	}
	woeB1Destroy(t, g, b12Permanent(g, me.ID, "My Relic", "Artifact"))
	if n := countersOn(g, ginger, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after your artifact died, want 1", n)
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("scry prompt %v, want one looking at one card", c)
	}
	answerScryKeepAll(t, g, me.ID)
}

func TestWoeS2SyrGingerSacrificesForLifeEqualToItsPower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, _ := woeB1Seats(g)
	ginger := b12Push(g, me.ID, "Syr Ginger, the Meal Ender", woeS2GingerTypeLine, woeS2GingerOracle, 3, 1)
	addCountersForTest(g, ginger, game.CounterPlusOne, 1)
	life := me.Life
	woeS2Activate(t, g, me, ginger, 0, "C", "C")
	if g.Battlefield.Contains(ginger) {
		t.Fatal("Syr Ginger was not sacrificed")
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d (its power as it was sacrificed)", me.Life, life+4)
	}
}
