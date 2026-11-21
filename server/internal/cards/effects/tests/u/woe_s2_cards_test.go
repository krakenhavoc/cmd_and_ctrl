package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardss "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/s"
	cardsso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/so"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_s2_cards_test.go — Wilds of Eldraine, the second S batch (#4471).
// Each test plays the real catalog entry and asserts what its text says.

const (
	woeS2ViperOracle      = cardss.ScaldingViperOracleID
	woeS2ScarecrowOracle  = "cf7bb5d4-c33f-4879-a793-6d6f2a4d94be"
	woeS2SlingerOracle    = "2e8433a2-ad71-4d30-84fe-12b709197824"
	woeS2FaerieOracle     = "a17633aa-b1a7-4ae5-b051-6747975d8a2f"
	woeS2KeepguardOracle  = "89078485-c438-4035-95dc-dbe37fb13e50"
	woeS2SpriteOracle     = "b0632f1b-14e5-4534-b2c5-af9d557fdcc0"
	woeS2CovenOracle      = cardsso.SpellscornCovenOracleID
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// woeS2Enters puts a permanent onto the battlefield under owner's control
// and announces its entry.
func woeS2Enters(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	t.Helper()
	id := effectstest.B12Permanent(g, owner, name, typeLine)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: owner})
	})
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// woeS2Activate activates a catalog ability of card with the given mana
// in p's pool and settles it.
func woeS2Activate(t *testing.T, g *game.Game, p *game.Player, card uuid.UUID, index int, colors ...string) {
	t.Helper()
	g.WithWriteLock(func() { p.ManaPool.EmptyPool() })
	effectstest.B06AddMana(p, colors...)
	if err := g.ActivateCatalogAbility(p.ID, card, index, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func woeS2Tapped(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && c.Tapped
}

// --- Scalding Viper // Steam Clean ---------------------------------------------

func TestWoeS2ScaldingViperBurnsAnOpponentForACheapSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, owner := effectstest.WoeB1Seats(g)
	effectstest.B12Push(g, owner.ID, "Scalding Viper", "Creature — Elemental Snake", woeS2ViperOracle, 2, 1)
	effectstest.ToMain(t, g)
	life := caster.Life
	effectstest.WoeS2CastInstant(t, g, caster, "{2}{R}")
	if caster.Life != life-1 {
		t.Fatalf("life %d after a mana value 3 spell, want %d", caster.Life, life-1)
	}
	effectstest.WoeS2CastInstant(t, g, caster, "{3}{R}")
	if caster.Life != life-1 {
		t.Errorf("life %d after a mana value 4 spell, want it unchanged at %d", caster.Life, life-1)
	}
}

func TestWoeS2ScaldingViperIgnoresItsControllersSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.B12Push(g, me.ID, "Scalding Viper", "Creature — Elemental Snake", woeS2ViperOracle, 2, 1)
	effectstest.ToMain(t, g)
	life := me.Life
	effectstest.WoeS2CastInstant(t, g, me, "{R}")
	if me.Life != life {
		t.Errorf("life %d after your own spell, want %d", me.Life, life)
	}
}

func TestWoeS2SteamCleanBouncesANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	card := woeB1Adventure(me, woeS2ViperOracle,
		game.Face{Name: "Scalding Viper", TypeLine: "Creature — Elemental Snake", ManaCost: "{1}{R}", Colors: []string{"R"}, Power: 2, Toughness: 1},
		game.Face{Name: "Steam Clean", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{U}", Colors: []string{"U"}})
	woeB1CastAdventure(t, g, me, card, effectstest.CardRefs(relic))
	if g.Battlefield.Contains(relic) || !opp.Hand.Contains(relic) {
		t.Error("the artifact did not return to its owner's hand")
	}
}

// --- Scarecrow Guide -------------------------------------------------------------

func TestWoeS2ScarecrowGuideFiltersOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := effectstest.WoeB1Seats(g)
	guide := effectstest.B12Push(g, me.ID, "Scarecrow Guide", "Artifact Creature — Scarecrow", woeS2ScarecrowOracle, 2, 1)
	effectstest.AssertKeywords(t, g, guide, "reach")
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "C", 2)
	if err := g.ActivateManaAbility(me.ID, guide, 0, game.ManaAbilityParams{Colors: []string{"W"}}); err != nil {
		t.Fatalf("{1}: add {W}: %v", err)
	}
	effectstest.MkWantPool(t, me, map[string]int{"C": 1, "W": 1})
	if err := g.ActivateManaAbility(me.ID, guide, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Error("a second activation this turn was allowed")
	}
}

// --- Skewer Slinger --------------------------------------------------------------

func TestWoeS2SkewerSlingerPokesWhatItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	slinger := effectstest.B12Push(g, opp.ID, "Skewer Slinger", "Creature — Dwarf Knight", woeS2SlingerOracle, 1, 3)
	effectstest.AssertKeywords(t, g, slinger, "reach")
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", woeS2BearType, 2, 4)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", woeS2BearType, 2, 4)
	effectstest.WoeS2Block(g, opp.ID, slinger, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, attacker); got != 1 {
		t.Errorf("the blocked creature has %d damage, want 1", got)
	}
	if got := effectstest.Pr6Marked(g, bystander); got != 0 {
		t.Errorf("a creature outside the block has %d damage, want 0", got)
	}
}

func TestWoeS2SkewerSlingerPokesEachCreatureBlockingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	slinger := effectstest.B12Push(g, me.ID, "Skewer Slinger", "Creature — Dwarf Knight", woeS2SlingerOracle, 1, 3)
	first := effectstest.B12Creature(g, opp.ID, "First Blocker", woeS2BearType, 2, 4)
	second := effectstest.B12Creature(g, opp.ID, "Second Blocker", woeS2BearType, 2, 4)
	effectstest.WoeS2Block(g, opp.ID, first, slinger)
	effectstest.WoeS2Block(g, opp.ID, second, slinger)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{first, second} {
		if got := effectstest.Pr6Marked(g, id); got != 1 {
			t.Errorf("a blocker has %d damage, want 1", got)
		}
	}
	if got := effectstest.Pr6Marked(g, slinger); got != 0 {
		t.Errorf("the Slinger dealt itself %d damage", got)
	}
}

// --- Sleep-Cursed Faerie ---------------------------------------------------------

func TestWoeS2SleepCursedFaerieEntersStunnedAndUntapsOneCounterAtATime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	faerie := effectstest.CastRingCreature(t, g, "Sleep-Cursed Faerie", "Creature — Faerie Wizard", woeS2FaerieOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(faerie) {
		t.Fatal("the Faerie did not resolve")
	}
	effectstest.AssertKeywords(t, g, faerie, "flying")
	if !woeS2Tapped(g, faerie) {
		t.Error("the Faerie entered untapped")
	}
	if n := effectstest.CountersOn(g, faerie, game.CounterStun); n != 3 {
		t.Fatalf("%d stun counters, want 3", n)
	}
	woeS2Activate(t, g, me, faerie, 0, "C", "U")
	if n := effectstest.CountersOn(g, faerie, game.CounterStun); n != 2 || !woeS2Tapped(g, faerie) {
		t.Errorf("after one untap: %d stun counters and tapped=%v, want 2 and tapped", n, woeS2Tapped(g, faerie))
	}
	woeS2Activate(t, g, me, faerie, 0, "C", "U")
	woeS2Activate(t, g, me, faerie, 0, "C", "U")
	woeS2Activate(t, g, me, faerie, 0, "C", "U")
	if n := effectstest.CountersOn(g, faerie, game.CounterStun); n != 0 || woeS2Tapped(g, faerie) {
		t.Errorf("after four untaps: %d stun counters and tapped=%v, want 0 and untapped", n, woeS2Tapped(g, faerie))
	}
}

func TestWoeS2SleepCursedFaerieHasWardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	faerie := effectstest.PushCatalogPermanent(g, me.ID, "Sleep-Cursed Faerie", "Creature — Faerie Wizard", woeS2FaerieOracle, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, faerie)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {2} payment for the opponent targeting the Faerie")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(faerie) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}
}

// --- Slumbering Keepguard --------------------------------------------------------

func TestWoeS2SlumberingKeepguardScriesWhenYourEnchantmentEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.PushCatalogPermanent(g, me.ID, "Slumbering Keepguard", "Creature — Human Knight", woeS2KeepguardOracle, false)
	woeS2Enters(t, g, opp.ID, "Their Charm", "Enchantment")
	if c := effectstest.ScryChoiceFor(g, me.ID); c != nil {
		t.Fatal("an opponent's enchantment made you scry")
	}
	woeS2Enters(t, g, me.ID, "My Charm", "Enchantment")
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("scry prompt %v, want one looking at one card", c)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

func TestWoeS2SlumberingKeepguardGrowsPerEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.WoeB1Seats(g)
	keep := effectstest.B12Push(g, me.ID, "Slumbering Keepguard", "Creature — Human Knight", woeS2KeepguardOracle, 1, 1)
	effectstest.B12Permanent(g, me.ID, "My Charm", "Enchantment")
	effectstest.B12Permanent(g, me.ID, "My Other Charm", "Enchantment")
	effectstest.B12Permanent(g, opp.ID, "Their Charm", "Enchantment")
	woeS2Activate(t, g, me, keep, 0, "C", "C", "W")
	if p, tough := effectstest.EffectivePTOf(t, g, keep); p != 3 || tough != 3 {
		t.Errorf("the Keepguard is %d/%d, want 3/3 with two enchantments", p, tough)
	}
}

// --- Snaremaster Sprite ----------------------------------------------------------

func TestWoeS2SnaremasterSpriteStunsWhenYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", woeS2BearType, 2, 2)
	sprite := effectstest.CastRingCreature(t, g, "Snaremaster Sprite", woeS2SpriteTypeLine, woeS2SpriteOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, sprite, "flying")
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked for before the payment was made")
	}
	effectstest.B06AddMana(me, "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !woeS2Tapped(g, theirs) {
		t.Error("their creature is not tapped")
	}
	if n := effectstest.CountersOn(g, theirs, game.CounterStun); n != 1 {
		t.Errorf("stun counters %d, want 1", n)
	}
}

func TestWoeS2SnaremasterSpriteDecliningDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", woeS2BearType, 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Snaremaster Sprite", woeS2SpriteTypeLine, woeS2SpriteOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a target prompt appeared after declining")
	}
	if woeS2Tapped(g, theirs) || effectstest.CountersOn(g, theirs, game.CounterStun) != 0 {
		t.Error("declining still tapped or stunned their creature")
	}
}

// --- Spellscorn Coven // Take It Back --------------------------------------------

func TestWoeS2SpellscornCovenMakesEachOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	for _, p := range g.Seats {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
	coven := effectstest.CastRingCreature(t, g, "Spellscorn Coven", "Creature — Faerie Warlock", woeS2CovenOracle, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, coven, "flying")
	for _, p := range g.Seats {
		owed := effectstest.DiscardOwed(g, p.ID)
		if p.ID == me.ID && owed != 0 {
			t.Errorf("you owe %d discards, want none", owed)
		}
		if p.ID != me.ID && owed != 1 {
			t.Errorf("an opponent owes %d discards, want 1", owed)
		}
	}
}

func TestWoeS2TakeItBackReturnsASpellToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shock := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: shock, Name: "Shock", TypeLine: "Instant", OracleID: effectstest.B5ShockOracle, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("cast Shock: %v", err)
	}
	card := woeB1Adventure(me, woeS2CovenOracle,
		game.Face{Name: "Spellscorn Coven", TypeLine: "Creature — Faerie Warlock", ManaCost: "{3}{B}", Colors: []string{"B"}, Power: 2, Toughness: 3},
		game.Face{Name: "Take It Back", TypeLine: "Instant — Adventure", ManaCost: "{2}{U}", Colors: []string{"U"}})
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.CardRefs(shock)}); err != nil {
		t.Fatalf("cast Take It Back: %v", err)
	}
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	relic := effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact")
	effectstest.CastAndResolveCreature(t, g, "Stockpiling Celebrant", woeS2CelebrantTypeStr, woeS2CelebrantOracle)
	effectstest.PickTriggerTarget(t, g, me.ID, relic)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || !me.Hand.Contains(relic) {
		t.Error("the artifact did not return to your hand")
	}
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %v, want one looking at two cards", c)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

func TestWoeS2StockpilingCelebrantDecliningKeepsThePermanentAndSkipsTheScry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	relic := effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact")
	effectstest.CastAndResolveCreature(t, g, "Stockpiling Celebrant", woeS2CelebrantTypeStr, woeS2CelebrantOracle)
	effectstest.PickTriggerTarget(t, g, me.ID, relic)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(relic) {
		t.Error("declining returned the artifact")
	}
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Error("declining still scried")
	}
}

func TestWoeS2StockpilingCelebrantCannotTargetALandOrAnOpponentsPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	land := effectstest.B12Permanent(g, me.ID, "My Land", "Land")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	mine := effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact")
	effectstest.CastAndResolveCreature(t, g, "Stockpiling Celebrant", woeS2CelebrantTypeStr, woeS2CelebrantOracle)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range []uuid.UUID{land, theirs} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
			t.Error("an illegal permanent was accepted as the target")
		}
	}
	effectstest.PickTriggerTarget(t, g, me.ID, mine)
}

// --- Syr Ginger, the Meal Ender --------------------------------------------------

func TestWoeS2SyrGingerHasItsKeywordsOnlyWhileAnOpponentHasAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	ginger := effectstest.B12Push(g, me.ID, "Syr Ginger, the Meal Ender", woeS2GingerTypeLine, woeS2GingerOracle, 3, 1)
	effectstest.B12Permanent(g, me.ID, "My Walker", "Legendary Planeswalker — Jace")
	for _, kw := range []string{"trample", "hexproof", "haste"} {
		if effectstest.HasEffectiveKeyword(t, g, ginger, kw) {
			t.Errorf("Syr Ginger has %s with only your own planeswalker out", kw)
		}
	}
	effectstest.B12Permanent(g, opp.ID, "Their Walker", "Legendary Planeswalker — Chandra")
	for _, kw := range []string{"trample", "hexproof", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, ginger, kw) {
			t.Errorf("Syr Ginger lacks %s while an opponent controls a planeswalker", kw)
		}
	}
}

func TestWoeS2SyrGingerGrowsAndScriesWhenAnotherArtifactOfYoursGoesToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	ginger := effectstest.B12Push(g, me.ID, "Syr Ginger, the Meal Ender", woeS2GingerTypeLine, woeS2GingerOracle, 3, 1)
	effectstest.WoeB1Destroy(t, g, effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact"))
	effectstest.WoeB1Destroy(t, g, effectstest.B12Permanent(g, me.ID, "My Charm", "Enchantment"))
	if n := effectstest.CountersOn(g, ginger, game.CounterPlusOne); n != 0 || effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Fatalf("%d counters after an opponent's artifact and your enchantment died, want 0 and no scry", n)
	}
	effectstest.WoeB1Destroy(t, g, effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact"))
	if n := effectstest.CountersOn(g, ginger, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after your artifact died, want 1", n)
	}
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("scry prompt %v, want one looking at one card", c)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

func TestWoeS2SyrGingerSacrificesForLifeEqualToItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := effectstest.WoeB1Seats(g)
	ginger := effectstest.B12Push(g, me.ID, "Syr Ginger, the Meal Ender", woeS2GingerTypeLine, woeS2GingerOracle, 3, 1)
	effectstest.AddCountersForTest(g, ginger, game.CounterPlusOne, 1)
	life := me.Life
	woeS2Activate(t, g, me, ginger, 0, "C", "C")
	if g.Battlefield.Contains(ginger) {
		t.Fatal("Syr Ginger was not sacrificed")
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d (its power as it was sacrificed)", me.Life, life+4)
	}
}
