package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_gl_cards_test.go — Outlaws of Thunder Junction (OTJ), the G–L
// batch (#4468): the cards that need no new engine seam. Each test plays
// the real catalog entry through the engine and asserts what the printed
// text says.

const (
	otjGLGeyserDrakeOracle        = "411b9158-e427-4099-bd30-3fbb22edb553"
	otjGLGoldPanOracle            = "91415d12-b13d-4b5c-bb60-6ec6f4fdde6e"
	otjGLGoldRushOracle           = "45687a24-e8cb-4864-8721-410eb8320aab"
	otjGLHardbristleBanditOracle  = "116eea43-4235-4608-8c6a-599e860ae2f1"
	otjGLHarrierStrixOracle       = "eb154ebb-d02f-405a-a557-3d5ec7648c42"
	otjGLHellspurPosseBossOracle  = "15260fa6-6062-423a-a248-81292ea46178"
	otjGLHolyCowOracle            = "748003fd-84c5-4c99-b146-d7d382eef64d"
	otjGLHonestRutsteinOracle     = "439bf6d1-bd65-474b-a70c-7fdd446de4bb"
	otjGLIronFistPulverizerOracle = "e5cea282-9e61-47da-8c5d-f5fca6a508a4"
	otjGLKraumOracle              = "1d764f70-985e-4cf0-9482-f30f73b9baf1"
	otjGLLonelyArroyoOracle       = "36508a8a-d1a4-400e-bfda-09436bc4d5d4"
	otjGLLushOasisOracle          = "b6a965eb-cffb-41c1-925a-7cf3e8e2f248"
)

func TestOtjGLEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		otjGLGeyserDrakeOracle:        "Geyser Drake",
		otjGLGoldPanOracle:            "Gold Pan",
		otjGLGoldRushOracle:           "Gold Rush",
		otjGLHardbristleBanditOracle:  "Hardbristle Bandit",
		otjGLHarrierStrixOracle:       "Harrier Strix",
		otjGLHellspurPosseBossOracle:  "Hellspur Posse Boss",
		otjGLHolyCowOracle:            "Holy Cow",
		otjGLHonestRutsteinOracle:     "Honest Rutstein",
		otjGLIronFistPulverizerOracle: "Iron-Fist Pulverizer",
		otjGLKraumOracle:              "Kraum, Violent Cacophony",
		otjGLLonelyArroyoOracle:       "Lonely Arroyo",
		otjGLLushOasisOracle:          "Lush Oasis",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// otjGLAnswerScry answers the open scry prompt by keeping every card on
// top, failing if there is none or it looks at other than `n` cards.
func otjGLAnswerScry(t *testing.T, g *game.Game, chooser uuid.UUID, n int) {
	t.Helper()
	c := scryChoiceFor(g, chooser)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != n {
		t.Fatalf("scry looks at %d cards, want %d", len(c.ScryCards), n)
	}
	if err := g.ResolveScry(c.ID, chooser, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
}

// --- the Desert duals ----------------------------------------------------------

func TestOtjGLDesertDualsEnterTappedPingAnOpponentAndTapForTheirColours(t *testing.T) {
	for _, row := range []struct {
		name, oracle string
		colours      []string
	}{
		{"Lonely Arroyo", otjGLLonelyArroyoOracle, []string{"W", "U"}},
		{"Lush Oasis", otjGLLushOasisOracle, []string{"G", "U"}},
	} {
		for _, colour := range row.colours {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[2]
			before := opp.Life
			land := b12PlayFromHand(t, g, row.name, "Land — Desert", row.oracle, game.CastSpellParams{})
			if !b16Tapped(t, g, land) {
				t.Fatalf("%s: the Desert enters tapped", row.name)
			}
			if p := latestPickTarget(g, me.ID); p == nil || hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
				t.Fatalf("%s: the trigger asks for an opponent, never you: %+v", row.name, p)
			}
			b16PickPlayer(t, g, me.ID, opp.ID)
			passPriorityAroundTable(t, g)
			if opp.Life != before-1 {
				t.Errorf("%s: target opponent takes 1: %d → %d", row.name, before, opp.Life)
			}
			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
			b28TapForMana(t, g, me.ID, land, colour)
			if got := poolColors(me); len(got) != 1 || got[0] != colour {
				t.Errorf("%s: tapped for {%s}: pool %v", row.name, colour, got)
			}
		}
	}
}

// --- Geyser Drake ----------------------------------------------------------------

func TestOtjGLGeyserDrakeDiscountsYourSpellsOnlyOnOtherPlayersTurns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drake := b12Push(g, me.ID, "Geyser Drake", "Creature — Drake", otjGLGeyserDrakeOracle, 2, 3)
	if !hasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("flying is printed")
	}
	advanceToMain(t, g)
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("on your own turn: %d, want 3 (undiscounted)", got)
	}
	advanceToMainOf(t, g, 1)
	if got := priceInHand(t, g, me, "Opt", "Instant", "{1}{U}"); got != 1 {
		t.Errorf("on an opponent's turn: %d, want 1 ({1} less)", got)
	}
	if got := priceInHand(t, g, opp, "Opt", "Instant", "{1}{U}"); got != 2 {
		t.Errorf("an opponent's spell: %d, want 2 (undiscounted)", got)
	}
}

// --- Gold Pan --------------------------------------------------------------------

func TestOtjGLGoldPanMakesATreasureAndPumpsTheEquippedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	pan := castCatalogSpell(t, g, "Gold Pan", "Artifact — Equipment", otjGLGoldPanOracle, nil)
	passPriorityAroundTable(t, g)
	if got := b16CountNamed(g, "Treasure"); got != 1 {
		t.Fatalf("%d Treasures when the Pan entered, want 1", got)
	}
	b06AddMana(me, "C")
	equipTo(t, g, me.ID, pan, bear)
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 3/3", p, tt)
	}
}

// --- Gold Rush -------------------------------------------------------------------

func TestOtjGLGoldRushCountsTheNewTreasureAndLocksTheBonus(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TreasureToken())
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Gold Rush", "Instant", otjGLGoldRushOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if got := b16CountNamed(g, "Treasure"); got != 2 {
		t.Fatalf("%d Treasures, want 2 (one made by the spell)", got)
	}
	if p, tt := ptOf(t, g, bear); p != 6 || tt != 6 {
		t.Fatalf("two Treasures: the Bear is %d/%d, want 6/6", p, tt)
	}
	pushToken(g, me.ID, TreasureToken())
	if p, tt := ptOf(t, g, bear); p != 6 || tt != 6 {
		t.Errorf("the count is taken once: a later Treasure made the Bear %d/%d", p, tt)
	}
	advanceToMainOf(t, g, 1)
	if p, tt := ptOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("until end of turn only: the Bear is %d/%d next turn", p, tt)
	}
}

func TestOtjGLGoldRushWithNoTargetStillMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Gold Rush", "Instant", otjGLGoldRushOracle, nil)
	passPriorityAroundTable(t, g)
	if got := b16CountNamed(g, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

// --- Hardbristle Bandit ----------------------------------------------------------

func TestOtjGLHardbristleBanditUntapsOnTheFirstCrimeEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bandit := b12Push(g, me.ID, "Hardbristle Bandit", "Creature — Plant Rogue", otjGLHardbristleBanditOracle, 1, 1)
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	b28TapForMana(t, g, me.ID, bandit, "R")
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Fatalf("tapped for {R}: pool %v", got)
	}

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(mine))
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, bandit) {
		t.Fatal("targeting your own creature is no crime: the Bandit stays tapped")
	}

	otjB1BoltOpponent(t, g, opp.ID)
	if b16Tapped(t, g, bandit) {
		t.Fatal("a crime untaps the Bandit")
	}
	b28TapForMana(t, g, me.ID, bandit, "G")
	otjB1BoltOpponent(t, g, opp.ID)
	if !b16Tapped(t, g, bandit) {
		t.Error("only once each turn: a second crime untapped it again")
	}
}

// --- Harrier Strix ---------------------------------------------------------------

func TestOtjGLHarrierStrixTapsATargetPermanentAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Drawn")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	strix := msc17Play(t, g, me, game.Card{Name: "Harrier Strix", TypeLine: "Creature — Bird",
		ManaCost: "{U}", OracleID: otjGLHarrierStrixOracle, Power: 1, Toughness: 1})
	if !hasEffectiveKeyword(t, g, strix, "flying") {
		t.Error("flying is printed")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if !hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, strix) {
		t.Fatalf("any permanent may be targeted: %v", p.PickTargetCards)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
		t.Fatalf("pick their Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirs) {
		t.Fatal("the target permanent is tapped")
	}

	b06AddMana(me, "U", "C", "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, strix, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1 before the discard", got-hand)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", discardOwed(g, me.ID))
	}
	if b16Tapped(t, g, strix) {
		t.Error("the loot has no tap cost")
	}
}

// --- Hellspur Posse Boss ---------------------------------------------------------

func TestOtjGLHellspurPosseBossMakesTwoMercenariesAndGivesOtherOutlawsHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rogue := b12Creature(g, me.ID, "My Rogue", "Creature — Human Rogue", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Rogue", "Creature — Human Rogue", 2, 2)
	advanceToMain(t, g)
	boss := msc17Play(t, g, me, game.Card{Name: "Hellspur Posse Boss", TypeLine: "Creature — Lizard Rogue",
		ManaCost: "{2}{R}{R}", OracleID: otjGLHellspurPosseBossOracle, Power: 2, Toughness: 4})
	mercs := otjMPMercenaries(g, me.ID)
	if len(mercs) != 2 {
		t.Fatalf("%d Mercenary tokens, want 2", len(mercs))
	}
	for _, id := range append(mercs, rogue) {
		if !hasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s: another outlaw you control has haste", id)
		}
	}
	if hasEffectiveKeyword(t, g, boss, "haste") {
		t.Error("\"other\" outlaws: the Boss does not give itself haste")
	}
	if hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("a Bear is no outlaw")
	}
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's Rogue is not an outlaw you control")
	}
	// The new Mercenaries have haste, so their tap ability is ready at once.
	b16Activate(t, g, me.ID, mercs[0], 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 2 {
		t.Errorf("the Mercenary pumped the Bear to %d/%d, want 3/2", p, tt)
	}
}

// --- Holy Cow --------------------------------------------------------------------

func TestOtjGLHolyCowHasFlashGainsTwoAndScriesOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Top", "Second")
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	before := me.Life
	cow, err := msc17Cast(t, g, me, game.Card{Name: "Holy Cow", TypeLine: "Creature — Ox Angel",
		ManaCost: "{2}{W}", OracleID: otjGLHolyCowOracle, Power: 2, Toughness: 2}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("flash: cast with a spell on the stack: %v", err)
	}
	passPriorityAroundTable(t, g)
	otjGLAnswerScry(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("you gain 2 life: %d → %d", before, me.Life)
	}
	if !hasEffectiveKeyword(t, g, cow, "flying") {
		t.Error("flying is printed")
	}
}

// --- Honest Rutstein -------------------------------------------------------------

func TestOtjGLHonestRutsteinReturnsACreatureCardAndDiscountsCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	spell := batch01GraveyardCard(me, "Old Shock", "Instant")
	theirs := batch01GraveyardCard(opp, "Their Dead Bear", "Creature — Bear")
	advanceToMain(t, g)
	msc17Play(t, g, me, game.Card{Name: "Honest Rutstein", TypeLine: "Legendary Creature — Human Warlock",
		ManaCost: "{1}{B}{G}", OracleID: otjGLHonestRutsteinOracle, Power: 3, Toughness: 2})
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if !hasID(p.PickTargetCards, dead) || hasID(p.PickTargetCards, spell) || hasID(p.PickTargetCards, theirs) {
		t.Fatalf("only a creature card from your graveyard: %v", p.PickTargetCards)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dead}); err != nil {
		t.Fatalf("pick the Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !handHasID(me, dead) {
		t.Error("the creature card returns to your hand")
	}

	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 1 {
		t.Errorf("your creature spell: %d, want 1 ({1} less)", got)
	}
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("a noncreature spell: %d, want 3 (undiscounted)", got)
	}
	if got := priceInHand(t, g, opp, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("an opponent's creature spell: %d, want 2 (undiscounted)", got)
	}
}

// --- Iron-Fist Pulverizer --------------------------------------------------------

func TestOtjGLIronFistPulverizerHitsAnOpponentAndScriesOnTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Top", "Second")
	giant := pushCatalogPermanent(g, me.ID, "Iron-Fist Pulverizer", "Creature — Giant Warrior", otjGLIronFistPulverizerOracle, false)
	if !hasEffectiveKeyword(t, g, giant, "reach") {
		t.Error("reach is printed")
	}
	b24CastNoncreature(t, g, "First", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the first spell triggered it")
	}
	before := opp.Life
	b24CastNoncreature(t, g, "Second", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the second spell: no target prompt")
	}
	if hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
		t.Fatalf("target opponent, never you: %v", p.PickTargetPlayers)
	}
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	otjGLAnswerScry(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("target opponent takes 2: %d → %d", before, opp.Life)
	}
	b24CastNoncreature(t, g, "Third", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the third spell triggered it")
	}
}

// --- Kraum, Violent Cacophony ----------------------------------------------------

func TestOtjGLKraumGrowsAndDrawsOnTheSecondSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two")
	kraum := pushCatalogPermanent(g, me.ID, "Kraum, Violent Cacophony", "Legendary Creature — Zombie Horror", otjGLKraumOracle, false)
	if !hasEffectiveKeyword(t, g, kraum, "flying") {
		t.Error("flying is printed")
	}
	hand := me.Hand.Size()
	b24CastNoncreature(t, g, "First", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || countersOn(g, kraum, game.CounterPlusOne) != 0 {
		t.Fatal("the first spell triggered it")
	}
	b24CastNoncreature(t, g, "Second", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the second spell draws one: hand %d → %d", hand, me.Hand.Size())
	}
	if got := countersOn(g, kraum, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on Kraum, want 1", got)
	}
	b24CastNoncreature(t, g, "Third", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || countersOn(g, kraum, game.CounterPlusOne) != 1 {
		t.Error("the third spell triggered it")
	}
}
