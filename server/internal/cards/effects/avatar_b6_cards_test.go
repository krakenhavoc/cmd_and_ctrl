package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b6_cards_test.go — Avatar: The Last Airbender batch 6 (#4459).

const (
	avB6OtterPenguinOracle  = "e2171c48-a816-41b2-aa55-e0ba2503fd6a"
	avB6SnowballersOracle   = "74224cce-c2ce-4456-8a97-3abbbb199385"
	avB6SerpentOracle       = "98e36c6e-a868-4679-a5b5-6f02097f560b"
	avB6HaikuOracle         = "87a97c72-143d-48a8-a1c5-3c44aa8e5a86"
	avB6OasisOracle         = "b677a73c-f179-46bf-83ee-c2b05f9fca67"
	avB6RevivalOracle       = "68979160-b5ce-4787-8a1e-1f40e614c3b0"
	avB6TeoOracle           = "ea74940f-01de-4804-92ed-b273fddf6d62"
	avB6TigerSealOracle     = "68f84643-c0f9-4d31-ba2e-38b65ab45d04"
	avB6AscensionOracle     = "aeddfdc2-149b-4703-ba57-8d844b25ef08"
	avB6WaterLessonOracle   = "3d28886c-46bf-4d2e-a257-cb9016cc3573"
	avB6ShrineTypeLine      = "Legendary Enchantment — Shrine"
	avB6TeoTypeLine         = "Legendary Creature — Human Pilot Ally"
	avB6SerpentTypeLine     = "Creature — Serpent"
	avB6SerpentManaCost     = "{5}{U}{U}"
	avB6SerpentFlashOffer   = "flash-if-three-lessons"
	avB6AscensionTypeLine   = "Enchantment"
	avB6WaterLessonTypeLine = "Sorcery — Lesson"
)

func TestAvatarB6CardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		avB6OtterPenguinOracle: "Otter-Penguin",
		avB6SnowballersOracle:  "Rowdy Snowballers",
		avB6SerpentOracle:      "Serpent of the Pass",
		avB6HaikuOracle:        "Sokka's Haiku",
		avB6OasisOracle:        "The Spirit Oasis",
		avB6RevivalOracle:      "Spirit Water Revival",
		avB6TeoOracle:          "Teo, Spirited Glider",
		avB6TigerSealOracle:    "Tiger-Seal",
		avB6AscensionOracle:    "Waterbender Ascension",
		avB6WaterLessonOracle:  "Waterbending Lesson",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s is not Full with no caveats", name)
		}
	}
}

// avB6Draw has p draw one card and lets the stack settle.
func avB6Draw(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	hobA5Draw(t, g, p, nil)
	passPriorityAroundTable(t, g)
}

// avB6Tap taps a battlefield permanent directly.
func avB6Tap(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.TapTargetForEffect(id); err != nil {
			t.Fatalf("tap: %v", err)
		}
	})
}

// avB6HandSpell puts a spell with a real mana cost into p's hand.
func avB6HandSpell(p *game.Player, name, typeLine, oracle, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		ManaCost: manaCost, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Otter-Penguin -------------------------------------------------------------

func TestAvatarB6OtterPenguinPumpsAndSlipsOnTheSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	otter := pushCatalogPermanent(g, me.ID, "Otter-Penguin", "Creature — Otter Bird", avB6OtterPenguinOracle, false)
	base, baseT := effectivePower(t, g, otter), effectiveToughness(t, g, otter)

	avB6Draw(t, g, g.Seats[2])
	avB6Draw(t, g, g.Seats[2])
	avB6Draw(t, g, me)
	if p := effectivePower(t, g, otter); p != base {
		t.Fatalf("power %d after the first draw (and another player's two), want %d", p, base)
	}
	if restrictionsOf(t, g, otter)&game.CantBeBlocked != 0 {
		t.Fatal("unblockable before the second draw")
	}
	avB6Draw(t, g, me)
	if p, tough := effectivePower(t, g, otter), effectiveToughness(t, g, otter); p != base+1 || tough != baseT+2 {
		t.Fatalf("%d/%d after the second draw, want %d/%d", p, tough, base+1, baseT+2)
	}
	if restrictionsOf(t, g, otter)&game.CantBeBlocked == 0 {
		t.Error("not unblockable after the second draw")
	}
	avB6Draw(t, g, me)
	if p := effectivePower(t, g, otter); p != base+1 {
		t.Errorf("power %d after the third draw, want it to stay %d", p, base+1)
	}
}

// --- Rowdy Snowballers ---------------------------------------------------------

func TestAvatarB6RowdySnowballersTapsAndStunsAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	castCatalogSpell(t, g, "Rowdy Snowballers", "Creature — Human Peasant Ally", avB6SnowballersOracle, nil)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !isTapped(g, theirs) {
		t.Error("the target was not tapped")
	}
	if n := countersOn(g, theirs, game.CounterStun); n != 1 {
		t.Errorf("%d stun counters, want 1", n)
	}
}

func TestAvatarB6RowdySnowballersCannotTargetYourOwnCreature(t *testing.T) {
	spec, _ := Lookup(avB6SnowballersOracle)
	if len(spec.Triggered) != 1 || spec.Triggered[0].Targets == nil {
		t.Fatalf("want one targeted trigger, got %+v", spec.Triggered)
	}
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	castCatalogSpell(t, g, "Rowdy Snowballers", "Creature — Human Peasant Ally", avB6SnowballersOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if isTapped(g, mine) || countersOn(g, mine, game.CounterStun) != 0 {
		t.Error("your own creature was tapped or stunned")
	}
}

// --- Serpent of the Pass -------------------------------------------------------

func avB6Graveyard(p *game.Player, n int, typeLine string) {
	for i := 0; i < n; i++ {
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old " + typeLine, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	}
}

func TestAvatarB6SerpentOfThePassCostsLessPerNoncreatureNonlandCard(t *testing.T) {
	for _, c := range []struct {
		name string
		pool string
		ok   bool
	}{
		{"discounted by three", "{C}{C}{U}{U}", true},
		{"one short", "{C}{U}{U}", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			avB6Graveyard(me, 3, "Instant")
			avB6Graveyard(me, 2, "Creature — Bear")
			avB6Graveyard(me, 2, "Land")
			id := avB6HandSpell(me, "Serpent of the Pass", avB6SerpentTypeLine, avB6SerpentOracle, avB6SerpentManaCost)
			addPool(t, g, me.ID, c.pool)
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
			if c.ok && err != nil {
				t.Fatalf("cast for %s with three instants in the graveyard: %v", c.pool, err)
			}
			if !c.ok && err == nil {
				t.Fatal("creature and land cards counted toward the discount")
			}
		})
	}
}

func TestAvatarB6SerpentOfThePassNeverDiscountsTheBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	avB6Graveyard(me, 9, "Sorcery")
	id := avB6HandSpell(me, "Serpent of the Pass", avB6SerpentTypeLine, avB6SerpentOracle, avB6SerpentManaCost)
	addPool(t, g, me.ID, "{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the discount reduced {U}{U}")
	}
	addPool(t, g, me.ID, "{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {U}{U} with nine sorceries in the graveyard: %v", err)
	}
}

func TestAvatarB6SerpentOfThePassHasFlashWithThreeLessons(t *testing.T) {
	for _, lessons := range []int{2, 3} {
		t.Run(map[int]string{2: "two lessons", 3: "three lessons"}[lessons], func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceTo(t, g, game.StepBeginCombat)
			avB6Graveyard(me, lessons, "Sorcery — Lesson")
			id := avB6HandSpell(me, "Serpent of the Pass", avB6SerpentTypeLine, avB6SerpentOracle, avB6SerpentManaCost)
			// The lessons are noncreature, nonland cards: the discount applies too.
			addPool(t, g, me.ID, "{C}{C}{C}{C}{U}{U}")
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
				t.Fatal("cast at sorcery speed outside a main phase without the flash offer")
			}
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: avB6SerpentFlashOffer})
			if lessons >= 3 && err != nil {
				t.Fatalf("flash cast with three Lessons in the graveyard: %v", err)
			}
			if lessons < 3 && err == nil {
				t.Fatal("flash cast with only two Lessons in the graveyard")
			}
		})
	}
}

// --- Sokka's Haiku -------------------------------------------------------------

func TestAvatarB6SokkasHaikuCountersDrawsMillsAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := b12Permanent(g, me.ID, "My Island", "Basic Land — Island")
	avB6Tap(t, g, land)
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetPlayer(me.ID))
	hand, lib, yard := me.Hand.Size(), me.Library.Size(), me.Graveyard.Size()
	haiku := castCatalogSpell(t, g, "Sokka's Haiku", "Instant — Lesson", avB6HaikuOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}, {Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != 40 {
		t.Fatal("the Bolt was not countered")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (one card drawn)", me.Hand.Size(), hand+1)
	}
	if me.Library.Size() != lib-4 {
		t.Errorf("library %d, want %d (one drawn, three milled)", me.Library.Size(), lib-4)
	}
	if me.Graveyard.Size() != yard+4 || !me.Graveyard.Contains(haiku) {
		t.Errorf("graveyard %d, want %d (three milled and the Haiku)", me.Graveyard.Size(), yard+4)
	}
	if isTapped(g, land) {
		t.Error("the land was not untapped")
	}
}

func TestAvatarB6SokkasHaikuStillUntapsWhenTheSpellIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := b12Permanent(g, me.ID, "My Island", "Basic Land — Island")
	avB6Tap(t, g, land)
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetPlayer(me.ID))
	lib := me.Library.Size()
	castCatalogSpell(t, g, "Sokka's Haiku", "Instant — Lesson", avB6HaikuOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}, {Kind: game.TargetCard, ID: land}})
	// A Counterspell on top takes the Bolt off the stack before the Haiku
	// resolves.
	castCatalogSpell(t, g, "Counterspell", "Instant", "cc187110-1148-4090-bbb8-e205694a39f5", cardRefs(bolt))
	passPriorityAroundTable(t, g)
	if me.Life != 40 {
		t.Fatal("the Bolt resolved")
	}
	if isTapped(g, land) {
		t.Error("the land was not untapped though it was still a legal target")
	}
	if me.Library.Size() != lib-4 {
		t.Errorf("library %d, want %d: the draw and mill still happen", me.Library.Size(), lib-4)
	}
}

// --- The Spirit Oasis ----------------------------------------------------------

func TestAvatarB6TheSpiritOasisDrawsPerShrineThenOnEachOtherShrine(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Permanent(g, me.ID, "Old Shrine", avB6ShrineTypeLine)
	b12Permanent(g, opp.ID, "Their Shrine", avB6ShrineTypeLine)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "The Spirit Oasis", avB6ShrineTypeLine, avB6OasisOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("drew %d on entering, want 2 (the Oasis and your other Shrine)", got)
	}

	hand = me.Hand.Size()
	castCatalogSpell(t, g, "New Shrine", avB6ShrineTypeLine, "", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d when another Shrine entered, want 1", got)
	}

	hand = me.Hand.Size()
	castCatalogSpell(t, g, "Plain Enchantment", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d when a non-Shrine entered, want 0", got)
	}
	b12Permanent(g, opp.ID, "Their New Shrine", avB6ShrineTypeLine)
	footprinted := 0
	for _, row := range game.CatalogTriggers(avB6OasisOracle) {
		if len(row.Footprint) > 0 {
			footprinted++
		}
	}
	if footprinted < 1 {
		t.Error("the one-sentence draw trigger has no footprint")
	}
}

// --- Spirit Water Revival ------------------------------------------------------

func TestAvatarB6SpiritWaterRevivalUnpaidDrawsTwoAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	avB6Graveyard(me, 3, "Instant")
	yard := me.Graveyard.Size()
	id := avB6HandSpell(me, "Spirit Water Revival", "Sorcery", avB6RevivalOracle, "{1}{U}{U}")
	hand := me.Hand.Size()
	addPool(t, g, me.ID, "{C}{U}{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - (hand - 1); got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if !inExile(g, id) {
		t.Error("the spell did not exile itself")
	}
	if me.Graveyard.Size() != yard {
		t.Errorf("graveyard %d, want it untouched at %d", me.Graveyard.Size(), yard)
	}
	if me.MaxHandSize == game.NoMaxHandSize {
		t.Error("the hand-size grant came without the waterbend")
	}
}

func TestAvatarB6SpiritWaterRevivalWaterbentShufflesDrawsSevenAndLiftsTheHandSize(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	avB6Graveyard(me, 4, "Instant")
	taps := pushTapCostSoldiers(g, me.ID, 6)
	id := avB6HandSpell(me, "Spirit Water Revival", "Sorcery", avB6RevivalOracle, "{1}{U}{U}")
	hand, lib := me.Hand.Size(), me.Library.Size()
	addPool(t, g, me.ID, "{C}{U}{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TapIDs: taps}); err != nil {
		t.Fatalf("cast with waterbend {6} paid by six taps: %v", err)
	}
	for _, s := range taps {
		if !isTapped(g, s) {
			t.Errorf("waterbender %v was not tapped", s)
		}
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - (hand - 1); got != 7 {
		t.Errorf("drew %d, want 7", got)
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("graveyard %d after the shuffle, want 0", me.Graveyard.Size())
	}
	if me.Library.Size() != lib+4-7 {
		t.Errorf("library %d, want %d (four shuffled in, seven drawn)", me.Library.Size(), lib+4-7)
	}
	if me.MaxHandSize != game.NoMaxHandSize {
		t.Errorf("max hand size %d, want none", me.MaxHandSize)
	}
	if !inExile(g, id) {
		t.Error("the spell did not exile itself")
	}
}

func TestAvatarB6SpiritWaterRevivalWaterbendCannotBePaidShort(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	taps := pushTapCostSoldiers(g, me.ID, 5)
	id := avB6HandSpell(me, "Spirit Water Revival", "Sorcery", avB6RevivalOracle, "{1}{U}{U}")
	addPool(t, g, me.ID, "{C}{U}{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TapIDs: taps}); err == nil {
		t.Fatal("five taps and no extra mana paid waterbend {6}")
	}
}

// --- Teo, Spirited Glider ------------------------------------------------------

func TestAvatarB6TeoLootsOnceForFlyersAndCountersOnANonlandDiscard(t *testing.T) {
	for _, c := range []struct {
		name    string
		discard string
		counter bool
	}{
		{"a nonland card", "Instant", true},
		{"a land", "Basic Land — Island", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			trekB35Library(g, me, 3)
			teo := pushCatalogPermanent(g, me.ID, "Teo, Spirited Glider", avB6TeoTypeLine, avB6TeoOracle, false)
			bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
			pitch := handCardOf(g, me, "Pitch", c.discard, "", "")
			hand := me.Hand.Size()
			advanceTo(t, g, game.StepDeclareAttackers)
			for _, id := range []uuid.UUID{teo, bear} {
				if err := g.DeclareAttacker(id, opp.ID); err != nil {
					t.Fatalf("DeclareAttacker: %v", err)
				}
			}
			lockInAttacks(t, g)
			g.RunStateChecksForTest()
			passPriorityAroundTable(t, g)
			if me.Hand.Size() != hand+1 {
				t.Fatalf("hand %d after the trigger, want %d (one draw for one batch)", me.Hand.Size(), hand+1)
			}
			answerDiscard(t, g, me.ID, pitch)
			passPriorityAroundTable(t, g)
			if p := latestPickTarget(g, me.ID); p != nil {
				if !c.counter {
					t.Fatal("a land discard made the reflexive trigger")
				}
				pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
				passPriorityAroundTable(t, g)
			} else if c.counter {
				t.Fatal("a nonland discard made no reflexive trigger")
			}
			want := 0
			if c.counter {
				want = 1
			}
			if got := countersOn(g, bear, game.CounterPlusOne); got != want {
				t.Errorf("%d +1/+1 counters, want %d", got, want)
			}
			if me.Hand.Size() != hand {
				t.Errorf("hand %d after the loot, want %d", me.Hand.Size(), hand)
			}
		})
	}
}

func TestAvatarB6TeoIgnoresAttacksWithoutFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB35Library(g, me, 3)
	pushCatalogPermanent(g, me.ID, "Teo, Spirited Glider", avB6TeoTypeLine, avB6TeoOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || discardOwed(g, me.ID) != 0 {
		t.Error("a ground attack triggered Teo")
	}
}

// --- Tiger-Seal ----------------------------------------------------------------

func TestAvatarB6TigerSealTapsInYourUpkeepAndUntapsOnTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	seat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	me := g.Seats[seat]
	seal := pushCatalogPermanent(g, me.ID, "Tiger-Seal", "Creature — Cat Seal", avB6TigerSealOracle, false)
	if !hasEffectiveKeyword(t, g, seal, "vigilance") {
		t.Error("Tiger-Seal has no vigilance")
	}
	stepToSeat(t, g, seat, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	if !isTapped(g, seal) {
		t.Fatal("the Seal did not tap in your upkeep")
	}
	// The draw step's draw is the first card this turn.
	stepToSeat(t, g, seat, game.StepPrecombatMain)
	passPriorityAroundTable(t, g)
	if !isTapped(g, seal) {
		t.Fatal("the first draw untapped the Seal")
	}
	avB6Draw(t, g, me)
	if isTapped(g, seal) {
		t.Fatal("the second draw did not untap the Seal")
	}
	avB6Tap(t, g, seal)
	avB6Draw(t, g, me)
	if !isTapped(g, seal) {
		t.Error("the third draw untapped the Seal")
	}
}

func TestAvatarB6TigerSealDoesNotTapInAnOpponentsUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	// The active seat's upkeep is behind it; the next upkeep is the
	// next player's.
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	seal := pushCatalogPermanent(g, me.ID, "Tiger-Seal", "Creature — Cat Seal", avB6TigerSealOracle, false)
	stepToSeat(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	passPriorityAroundTable(t, g)
	if isTapped(g, seal) {
		t.Error("the Seal tapped in another player's upkeep")
	}
}

// --- Waterbender Ascension -----------------------------------------------------

func TestAvatarB6WaterbenderAscensionCountsHitsAndDrawsFromTheFourth(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	asc := pushCatalogPermanent(g, me.ID, "Waterbender Ascension", avB6AscensionTypeLine, avB6AscensionOracle, false)
	trekB35SetCounters(g, asc, "quest", 2)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, a, b)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := countersOn(g, asc, "quest"); got != 4 {
		t.Fatalf("%d quest counters, want 4", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1 (the third counter draws nothing, the fourth draws)", got)
	}
}

func TestAvatarB6WaterbenderAscensionWaterbendsACreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	asc := pushCatalogPermanent(g, me.ID, "Waterbender Ascension", avB6AscensionTypeLine, avB6AscensionOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	short := pushTapCostSoldiers(g, me.ID, 3)
	if err := g.ActivateCatalogAbility(me.ID, asc, 0, game.ActivateAbilityParams{
		WaterbendIDs: short, Targets: cardRefs(bear), Strict: true,
	}); err == nil {
		t.Fatal("three taps and an empty pool paid Waterbend {4}")
	}
	helpers := append(short, pushTapCostSoldiers(g, me.ID, 1)...)
	if err := g.ActivateCatalogAbility(me.ID, asc, 0, game.ActivateAbilityParams{
		WaterbendIDs: helpers, Targets: cardRefs(bear), Strict: true,
	}); err != nil {
		t.Fatalf("Waterbend {4} with four taps: %v", err)
	}
	passPriorityAroundTable(t, g)
	if restrictionsOf(t, g, bear)&game.CantBeBlocked == 0 {
		t.Error("the target can still be blocked")
	}
}

// --- Waterbending Lesson -------------------------------------------------------

func avB6CastWaterLesson(t *testing.T, g *game.Game, me *game.Player) (hand int) {
	t.Helper()
	advanceToMain(t, g)
	trekB35Library(g, me, 5)
	id := avB6HandSpell(me, "Waterbending Lesson", avB6WaterLessonTypeLine, avB6WaterLessonOracle, "{3}{U}")
	hand = me.Hand.Size() - 1
	addPool(t, g, me.ID, "{C}{C}{C}{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	return hand
}

func TestAvatarB6WaterbendingLessonWaterbentKeepsAllThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	helpers := pushTapCostSoldiers(g, me.ID, 2)
	hand := avB6CastWaterLesson(t, g, me)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Fatalf("drew %d before the waterbend question, want 3", got)
	}
	prompt := payUnlessFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no waterbend {2} question for the caster")
	}
	if tc := prompt.PayTapCost(); tc == nil || tc.Extra != "{2}" {
		t.Fatalf("the prompt carries no waterbend {2} clause: %+v", tc)
	}
	if err := g.ResolvePayUnlessWithTaps(prompt.ID, me.ID, true, helpers); err != nil {
		t.Fatalf("waterbend with two taps: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range helpers {
		if !isTapped(g, id) {
			t.Errorf("%v was not tapped to waterbend", id)
		}
	}
	if discardOwed(g, me.ID) != 0 {
		t.Error("a discard is owed after the waterbend was paid")
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("hand +%d, want +3", got)
	}
}

func TestAvatarB6WaterbendingLessonDeclinedDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := avB6CastWaterLesson(t, g, me)
	prompt := payUnlessFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no waterbend {2} question for the caster")
	}
	if err := g.ResolvePayUnless(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if got := discardOwed(g, me.ID); got != 1 {
		t.Fatalf("discard owed %d, want 1", got)
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("hand +%d before the discard, want +3", got)
	}
}
