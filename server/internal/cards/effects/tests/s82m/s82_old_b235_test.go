package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b235_test.go — Hour of Devastation, batch 235 (#5534): each test plays
// the real catalog entry through the real cast, activation and trigger paths.

const (
	oldB235VisageOracle      = "21bb3c03-c1c0-40ec-8df4-8d0054c3742b"
	oldB235VileOracle        = "3d747c43-a09b-4209-b042-2b89156612d8"
	oldB235EncouragementOID  = "55e897e9-5c70-4e32-8f24-3e0fefa319cd"
	oldB235DjeruOracle       = "5d520dd4-7986-41fa-ab94-9226da751668"
	oldB235NissasDefeatOID   = "5f808e3b-6b69-4b46-8a92-36d960527acc"
	oldB235ChandrasDefeatOID = "a57ad23b-cff3-45e5-98b3-1b20224aaa19"
	oldB235ThreatsOracle     = "b1bc4498-43c0-461c-afba-ef81d743e5aa"
	oldB235VizierOracle      = "cb5aa38e-6a76-4c0c-be14-3c546acd353d"
	oldB235UncageOracle      = "d2043685-f091-444f-9c35-d238afa24d7b"
	oldB235SacredCatOracle   = "d85ea576-a794-44bf-b405-1f1c49477409"
)

func TestOldB235CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB235VisageOracle:      "Visage of Bolas",
		oldB235VileOracle:        "Vile Manifestation",
		oldB235EncouragementOID:  "Nissa's Encouragement",
		oldB235DjeruOracle:       "Djeru, With Eyes Open",
		oldB235NissasDefeatOID:   "Nissa's Defeat",
		oldB235ChandrasDefeatOID: "Chandra's Defeat",
		oldB235ThreatsOracle:     "Imaginary Threats",
		oldB235VizierOracle:      "Vizier of the Anointed",
		oldB235UncageOracle:      "Uncage the Menagerie",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func oldB235Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- Visage of Bolas ---

func TestOldB235VisageOfBolasFetchesNicolBolasThenTapsForThreeColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Nicol Bolas, the Deceiver", TypeLine: "Legendary Creature — Elder Dragon"})
	visage := effectstest.CastCatalogSpell(t, g, "Visage of Bolas", "Artifact", oldB235VisageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("Nicol Bolas, the Deceiver did not reach the hand")
	}
	for _, want := range []string{"U", "B", "R"} {
		if err := g.ActivateManaAbility(me.ID, visage, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		pick := effectstest.RiderLatestManaPick(g, me.ID)
		if pick == nil || len(pick.ColorOptions) != 3 {
			t.Fatalf("colour options %+v, want U, B and R", pick)
		}
		effectstest.RiderAnswerManaPicks(t, g, me.ID, want)
		if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != want {
			t.Errorf("pool %v, want [%s]", got, want)
		}
		me.ManaPool.EmptyPool()
		effectstest.B08Untap(g, visage)
	}
}

func TestOldB235VisageOfBolasMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me, game.Card{Name: "Nicol Bolas, the Deceiver", TypeLine: "Legendary Creature — Elder Dragon"})
	effectstest.CastCatalogSpell(t, g, "Visage of Bolas", "Artifact", oldB235VisageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(ids[0]) {
		t.Error("declining still took the card")
	}
}

// --- Vile Manifestation ---

func TestOldB235VileManifestationGrowsPerCyclingCardInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vile := effectstest.PushCatalogPermanent(g, me.ID, "Vile Manifestation", "Creature — Horror", oldB235VileOracle, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, vile)
		c.Power, c.Toughness = 0, 4
	})
	if p := effectstest.EffectivePower(t, g, vile); p != 0 {
		t.Fatalf("power with no cycling cards = %d, want 0", p)
	}
	// Two cycling cards and one without cycling.
	for _, id := range []uuid.UUID{
		effectstest.GyCard(me, "Vile Manifestation", "Creature — Horror", "{1}{B}", oldB235VileOracle),
		effectstest.GyCard(me, "Vile Manifestation", "Creature — Horror", "{1}{B}", oldB235VileOracle),
		effectstest.GyCard(me, "Plain Bear", "Creature — Bear", "{1}{G}", ""),
	} {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneGraveyard})
		})
	}
	g.RunStateChecksForTest()
	if p := effectstest.EffectivePower(t, g, vile); p != 2 {
		t.Errorf("power with two cycling cards in the graveyard = %d, want 2", p)
	}
	if tt := effectstest.EffectiveToughness(t, g, vile); tt != 4 {
		t.Errorf("toughness = %d, want 4", tt)
	}
}

func TestOldB235VileManifestationCyclesForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, "Vile Manifestation", "Creature — Horror", "{1}{B}", oldB235VileOracle, []string{"B"})
	effectstest.PushLibraryCard(me, game.Card{Name: "Drawn", TypeLine: "Creature — Bear"})
	before := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled card is not in the graveyard")
	}
	if me.Hand.Size() != before {
		t.Errorf("hand size %d, want %d (one out, one drawn)", me.Hand.Size(), before)
	}
}

// --- Nissa's Encouragement ---

func TestOldB235NissasEncouragementTakesOneOfEachNameFromLibraryAndGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	libForest := effectstest.PushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	behemoth := effectstest.PushLibraryCard(me, game.Card{Name: "Brambleweft Behemoth", TypeLine: "Creature — Treefolk"})
	nissa := effectstest.GyCard(me, "Nissa, Genesis Mage", "Legendary Planeswalker — Nissa", "", "")
	effectstest.PushLibraryCard(me, game.Card{Name: "Other", TypeLine: "Creature — Bear"})
	effectstest.CastCatalogSpell(t, g, "Nissa's Encouragement", "Sorcery", oldB235EncouragementOID, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, libForest, behemoth, nissa)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Forest": libForest, "Brambleweft Behemoth": behemoth, "Nissa, Genesis Mage": nissa} {
		if !me.Hand.Contains(id) {
			t.Errorf("%s did not reach the hand", name)
		}
	}
}

func TestOldB235NissasEncouragementRefusesTwoForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	b := effectstest.PushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Nissa's Encouragement", "Sorcery", oldB235EncouragementOID, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{a, b}); err == nil {
		t.Error("two cards named Forest were accepted")
	}
}

// --- Djeru, With Eyes Open ---

func TestOldB235DjeruFetchesAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "A Walker", TypeLine: "Legendary Planeswalker — Test"})
	effectstest.CastCatalogSpell(t, g, "Djeru, With Eyes Open", "Legendary Creature — Human Warrior", oldB235DjeruOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the planeswalker did not reach the hand")
	}
}

func oldB235Walker(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := effectstest.PushCatalogPermanent(g, owner, "Test Walker", "Legendary Planeswalker — Test", "", false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, id)
		c.Counters = map[string]int{game.CounterLoyalty: 5}
	})
	return id
}

func TestOldB235DjeruPreventsOneDamageToYourPlaneswalkersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Djeru, With Eyes Open", "Legendary Creature — Human Warrior", oldB235DjeruOracle, false)
	mine := oldB235Walker(g, me.ID)
	theirs := oldB235Walker(g, opp.ID)
	striker := effectstest.B12Creature(g, opp.ID, "Striker", "Creature — Bear", 3, 3)
	effectstest.P2Ping(t, g, striker, mine, 3)
	effectstest.P2Ping(t, g, striker, theirs, 3)
	if got := effectstest.CountersOn(g, mine, game.CounterLoyalty); got != 3 {
		t.Errorf("your planeswalker has %d loyalty after 3 damage, want 3 (one prevented)", got)
	}
	if got := effectstest.CountersOn(g, theirs, game.CounterLoyalty); got != 2 {
		t.Errorf("the opponent's planeswalker has %d loyalty, want 2 (no prevention)", got)
	}
	effectstest.P2Ping(t, g, striker, mine, 1)
	if got := effectstest.CountersOn(g, mine, game.CounterLoyalty); got != 3 {
		t.Errorf("a single point of damage was not prevented entirely: loyalty %d", got)
	}
	before := me.Life
	effectstest.P2Ping(t, g, striker, me.ID, 3)
	if me.Life != before-3 {
		t.Errorf("damage to you was changed: %d -> %d", before, me.Life)
	}
}

// --- Nissa's Defeat ---

func oldB235Colored(g *game.Game, owner uuid.UUID, name, typeLine string, colors ...string) uuid.UUID {
	c := game.Card{Name: name, TypeLine: typeLine, Colors: colors, Power: 1, Toughness: 1}
	if strings.Contains(typeLine, "Planeswalker") {
		c.Counters = map[string]int{game.CounterLoyalty: 5}
	}
	return effectstest.ApaPush(g, owner, owner, c)
}

func TestOldB235NissasDefeatDestroysANissaAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nissa := oldB235Colored(g, opp.ID, "Nissa, Test", "Legendary Planeswalker — Nissa", "G")
	effectstest.PushLibraryCard(me, game.Card{Name: "Drawn", TypeLine: "Creature — Bear"})
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Nissa's Defeat", "Sorcery", oldB235NissasDefeatOID, oldB235Card(nissa))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(nissa) {
		t.Error("the Nissa survived")
	}
	if me.Hand.Size() != before+1 {
		t.Errorf("hand size %d, want %d (one card drawn)", me.Hand.Size(), before+1)
	}
}

func TestOldB235NissasDefeatOtherLegalTargetsDrawNothing(t *testing.T) {
	for name, build := range map[string]func(*game.Game, uuid.UUID) uuid.UUID{
		"Forest": func(g *game.Game, o uuid.UUID) uuid.UUID {
			return oldB235Colored(g, o, "Forest", "Basic Land — Forest")
		},
		"enchantment": func(g *game.Game, o uuid.UUID) uuid.UUID { return oldB235Colored(g, o, "Vines", "Enchantment", "G") },
		"walker": func(g *game.Game, o uuid.UUID) uuid.UUID {
			return oldB235Colored(g, o, "Garruk", "Legendary Planeswalker — Garruk", "G")
		},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		target := build(g, opp.ID)
		effectstest.PushLibraryCard(me, game.Card{Name: "Drawn", TypeLine: "Creature — Bear"})
		before := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Nissa's Defeat", "Sorcery", oldB235NissasDefeatOID, oldB235Card(target))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(target) {
			t.Errorf("%s: the target survived", name)
		}
		if me.Hand.Size() != before {
			t.Errorf("%s: hand size %d, want %d (no draw)", name, me.Hand.Size(), before)
		}
	}
}

func TestOldB235NissasDefeatRefusesAnIllegalTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	for name, id := range map[string]uuid.UUID{
		"Island":              oldB235Colored(g, opp.ID, "Island", "Basic Land — Island"),
		"blue enchantment":    oldB235Colored(g, opp.ID, "Pool", "Enchantment", "U"),
		"green creature":      oldB235Colored(g, opp.ID, "Bear", "Creature — Bear", "G"),
		"nongreen planeswalk": oldB235Colored(g, opp.ID, "Jace", "Legendary Planeswalker — Jace", "U"),
	} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Nissa's Defeat", "Sorcery", oldB235NissasDefeatOID, oldB235Card(id)); err == nil {
			t.Errorf("%s was a legal target", name)
		}
	}
}

// --- Chandra's Defeat ---

func TestOldB235ChandrasDefeatKillsARedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	imp := oldB235Colored(g, opp.ID, "Imp", "Creature — Imp", "R")
	effectstest.CastCatalogSpell(t, g, "Chandra's Defeat", "Instant", oldB235ChandrasDefeatOID, oldB235Card(imp))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(imp) {
		t.Error("5 damage did not kill the red creature")
	}
}

func TestOldB235ChandrasDefeatOnAChandraMayLoot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chandra := oldB235Colored(g, opp.ID, "Chandra, Test", "Legendary Planeswalker — Chandra", "R")
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, chandra).Counters = map[string]int{game.CounterLoyalty: 7}
	})
	pitch := effectstest.HandCard(me, "Pitch", "Creature — Bear")
	effectstest.PushLibraryCard(me, game.Card{Name: "Drawn", TypeLine: "Creature — Bear"})
	effectstest.CastCatalogSpell(t, g, "Chandra's Defeat", "Instant", oldB235ChandrasDefeatOID, oldB235Card(chandra))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, chandra, game.CounterLoyalty); got != 2 {
		t.Errorf("the Chandra has %d loyalty after 5 damage, want 2", got)
	}
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the chosen card was not discarded")
	}
	if effectstest.ZoneOf(g, pitch) != game.ZoneGraveyard {
		t.Errorf("pitched card in %v", effectstest.ZoneOf(g, pitch))
	}
}

func TestOldB235ChandrasDefeatOnANonChandraRedWalkerDoesNotLoot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := oldB235Colored(g, opp.ID, "Test Walker", "Legendary Planeswalker — Test", "R")
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, walker).Counters = map[string]int{game.CounterLoyalty: 7}
	})
	effectstest.HandCard(me, "Pitch", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Chandra's Defeat", "Instant", oldB235ChandrasDefeatOID, oldB235Card(walker))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("a non-Chandra planeswalker offered the loot")
	}
}

func TestOldB235ChandrasDefeatRefusesANonRedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := oldB235Colored(g, g.Seats[1].ID, "Bear", "Creature — Bear", "G")
	if err := effectstest.CastCatalogSpellErr(t, g, "Chandra's Defeat", "Instant", oldB235ChandrasDefeatOID, oldB235Card(bear)); err == nil {
		t.Error("a green creature was a legal target")
	}
}

// --- Imaginary Threats ---

func TestOldB235ImaginaryThreatsForcesAttacksAndFreezesTheCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Imaginary Threats", "Instant", oldB235ThreatsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	_ = me
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear || g.Battlefield.Cards[i].InstanceID == rock {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if !effectstest.AlliTapped(t, g, bear) {
		t.Error("the creature untapped during its controller's next untap step")
	}
	if effectstest.AlliTapped(t, g, rock) {
		t.Error("a noncreature permanent was frozen")
	}
}

func TestOldB235ImaginaryThreatsCyclesForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, "Imaginary Threats", "Instant", "{2}{U}{U}", oldB235ThreatsOracle, []string{"U"})
	effectstest.PushLibraryCard(me, game.Card{Name: "Drawn", TypeLine: "Creature — Bear"})
	before := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Errorf("hand size %d, want %d", me.Hand.Size(), before)
	}
}

// --- Vizier of the Anointed ---

func TestOldB235VizierSearchesAnEmbalmCreatureIntoTheGraveyardAndDrawsOnActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryCard(me, game.Card{Name: "Drawn", TypeLine: "Creature — Bear"})
	plain := effectstest.PushLibraryCard(me, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear"})
	cat := effectstest.PushLibraryCard(me, game.Card{Name: "Sacred Cat", TypeLine: "Creature — Cat", OracleID: oldB235SacredCatOracle,
		ManaCost: "{W}", Power: 1, Toughness: 1})
	effectstest.CastCatalogSpell(t, g, "Vizier of the Anointed", "Creature — Human Cleric", oldB235VizierOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Plain Bear") != uuid.Nil {
		t.Error("a creature without embalm or eternalize was offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, cat)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(cat) {
		t.Fatal("Sacred Cat did not reach the graveyard")
	}
	_ = plain
	before := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, cat, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("embalm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand size %d after embalming, want %d (the Vizier draws)", me.Hand.Size(), before+1)
	}
}

// --- Uncage the Menagerie ---

func oldB235Creature(name, cost string) game.Card {
	return game.Card{Name: name, TypeLine: "Creature — Beast", ManaCost: cost, Power: 1, Toughness: 1}
}

func TestOldB235UncageTakesUpToXDifferentlyNamedCreaturesOfManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		oldB235Creature("Alpha", "{1}{G}"), oldB235Creature("Beta", "{1}{G}"), oldB235Creature("Beta", "{1}{G}"),
		oldB235Creature("Gamma", "{2}{G}"))
	effectstest.Pr8CastX(t, g, "Uncage the Menagerie", "Sorcery", oldB235UncageOracle, 2, nil)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[1], ids[2]}); err == nil {
		t.Fatal("two creatures with the same name were accepted")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], ids[3]}); err == nil {
		t.Fatal("a creature of mana value 3 was accepted for X = 2")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], ids[1]}); err != nil {
		t.Fatalf("a legal pair was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[1]) {
		t.Error("the chosen creatures did not reach the hand")
	}
}

func TestOldB235UncageWithXZeroTakesNothingAndShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me, oldB235Creature("Alpha", ""))
	effectstest.Pr8CastX(t, g, "Uncage the Menagerie", "Sorcery", oldB235UncageOracle, 0, nil)
	if me.Hand.Contains(ids[0]) {
		t.Error("X = 0 took a card")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("X = 0 offered a search")
	}
}
