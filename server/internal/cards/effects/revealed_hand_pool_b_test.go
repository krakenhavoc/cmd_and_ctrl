package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0116 pool B (#2078): the revealed-hand pick on modal spells,
// alternative and variable costs, and spells with text after the
// discard. One test for each card's own clause; the pick itself is
// pinned in revealed_hand_discard_test.go.

const (
	auntiesSentenceOracle      = "406e6db7-6e31-44b9-a793-57ee10688f71"
	cerebralConfiscationOracle = "1ea71ea2-65ca-4907-b857-dbd1ba8efacd"
	splittingHeadacheOracle    = "3a5937de-5957-47a9-9bd4-0dc7f8b833ef"
	poisonTheWatersOracle      = "2c9bf40b-ddcb-4f46-a81e-56b91dfc784f"
	marduCharmOracle           = "147eabab-9d93-4cf4-811b-0efa3b84c5b5"
	daiLiIndoctrinationOracle  = "5591efbc-1b2c-489a-b4f3-935c9b6bd71a"
	drillBitOracle             = "0dbd6e47-a8b4-4268-ba44-8924cd4963a6"
	dreadFugueOracle           = "efe4194a-b3c5-4308-abfe-553e035f5e64"
	venarianGlimmerOracle      = "336cc3ab-c1c7-4bf1-b1d7-dfc9f3c556ab"
	gruesomeDiscoveryOracle    = "ecdae60b-c594-4e70-909b-83483104a42c"
	riversGraspOracle          = "f671583f-bd22-4d3e-bf19-fb740c607f8f"
	devourIntellectOracle      = "abf8264e-d020-4939-8092-83469b1a2244"
	humiliateOracle            = "cf18a1ca-1fb5-4c90-b20a-6ae2fb54a9bb"
	renderSpeechlessOracle     = "6fdddab4-fcc9-4622-9c17-c69f9d4bca08"
	egoDrainOracle             = "b4fc198f-3595-4880-ad12-96c18fbf9c33"
	memoryTheftOracle          = "1d60e2a7-2059-48cb-a6ab-36ba25b80b3a"
)

func pbBear() game.Card {
	return game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2}
}

func pbRing() game.Card {
	return game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"}
}

func pbPacifism() game.Card {
	return game.Card{Name: "Pacifism", TypeLine: "Enchantment — Aura", ManaCost: "{1}{W}", Colors: []string{"W"}}
}

func pbDivination() game.Card {
	return game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}", Colors: []string{"U"}}
}

// pbCast puts a card in the active seat's hand, moves to a main phase
// and casts it with params. Returns the caster.
func pbCast(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, params game.CastSpellParams) *game.Player {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	id := handCardFull(me, name, typeLine, manaCost, oracle, nil)
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return me
}

// pbModal casts a "choose one" spell for mode `mode` with targets.
func pbModal(t *testing.T, g *game.Game, name, typeLine, oracle string, mode int, targets ...game.TargetRef) {
	t.Helper()
	pbCast(t, g, name, typeLine, oracle, "", game.CastSpellParams{Modes: []int{mode}, Targets: targets})
	passPriorityAroundTable(t, g)
}

func pbPlayer(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetPlayer, ID: id} }
func pbCard(id uuid.UUID) game.TargetRef   { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// pbChooseCards is the one open choose-cards prompt for chooser, or nil.
func pbChooseCards(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func pbNoPick(t *testing.T, g *game.Game) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDiscardFromHand {
			t.Fatalf("a revealed-hand pick is open: %+v", c)
		}
	}
}

func pbLand(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: owner, Controller: owner,
	})
}

// Auntie's Sentence and Dai Li Indoctrination: "nonland permanent card"
// takes the creature, artifact and enchantment, not the land or the
// instant or sorcery.
func TestNonlandPermanentCardPicks(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Auntie's Sentence", "Sorcery", auntiesSentenceOracle},
		{"Dai Li Indoctrination", "Sorcery — Lesson", daiLiIndoctrinationOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			victim := g.Seats[1]
			ids := revealHand(victim, rhForest(), rhBolt(), pbBear(), pbRing(), pbPacifism(), pbDivination())
			pbModal(t, g, tc.name, tc.typeLine, tc.oracle, 0, pbPlayer(victim.ID))
			pick := openPick(t, g)
			if want := ids[2:5]; !sameIDs(pick.DiscardOptions, want) {
				t.Errorf("options = %v, want the creature, the artifact and the enchantment %v", pick.DiscardOptions, want)
			}
			if pick.DiscardLabel != "nonland permanent card" {
				t.Errorf("label = %q", pick.DiscardLabel)
			}
		})
	}
}

// Auntie's Sentence's second bullet shrinks its own target.
func TestAuntiesSentenceShrinksTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	bear := pr7Creature(g, g.Seats[1].ID, "Big Bear", 4)
	pbModal(t, g, "Auntie's Sentence", "Sorcery", auntiesSentenceOracle, 1, pbCard(bear))
	pbNoPick(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2 || tg != 2 {
		t.Errorf("Big Bear = %d/%d, want 2/2", p, tg)
	}
}

// Dai Li Indoctrination's second bullet earthbends its own target land.
func TestDaiLiIndoctrinationEarthbendsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pbLand(g, me.ID)
	pbModal(t, g, "Dai Li Indoctrination", "Sorcery — Lesson", daiLiIndoctrinationOracle, 1, pbCard(land))
	if got := countersOn(g, land, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if !layeredCard(t, g, land).IsCreature() {
		t.Error("the earthbent land is not a creature")
	}
}

// Cerebral Confiscation and Splitting Headache's "discards two cards" is
// the player's own choice; Splitting Headache's pick takes any card,
// lands included, and Cerebral Confiscation's only nonland cards.
func TestCerebralConfiscationAndSplittingHeadache(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		wantAll      bool
	}{
		{"Cerebral Confiscation", cerebralConfiscationOracle, false},
		{"Splitting Headache", splittingHeadacheOracle, true},
	} {
		t.Run(tc.name+" discards two", func(t *testing.T) {
			g := newCatalogGame(t)
			victim := g.Seats[1]
			revealHand(victim, rhForest(), rhBolt(), pbBear())
			pbModal(t, g, tc.name, "Sorcery", tc.oracle, 0, pbPlayer(victim.ID))
			pbNoPick(t, g)
			c := pbChooseCards(g, victim.ID)
			if c == nil {
				t.Fatal("the victim was not asked to discard")
			}
			if c.ChooseMin != 2 || c.ChooseMax != 2 {
				t.Errorf("discard bounds = %d..%d, want 2", c.ChooseMin, c.ChooseMax)
			}
		})
		t.Run(tc.name+" reveals", func(t *testing.T) {
			g := newCatalogGame(t)
			victim := g.Seats[1]
			ids := revealHand(victim, rhForest(), rhBolt(), pbBear())
			pbModal(t, g, tc.name, "Sorcery", tc.oracle, 1, pbPlayer(victim.ID))
			want := ids[1:]
			if tc.wantAll {
				want = ids
			}
			if pick := openPick(t, g); !sameIDs(pick.DiscardOptions, want) {
				t.Errorf("options = %v, want %v", pick.DiscardOptions, want)
			}
		})
	}
}

// Poison the Waters: artifact or creature cards only, and the first
// bullet shrinks every creature.
func TestPoisonTheWaters(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	ids := revealHand(victim, rhForest(), rhBolt(), pbBear(), pbRing(), pbPacifism())
	pbModal(t, g, "Poison the Waters", "Sorcery", poisonTheWatersOracle, 1, pbPlayer(victim.ID))
	if pick := openPick(t, g); !sameIDs(pick.DiscardOptions, ids[2:4]) {
		t.Errorf("options = %v, want the Bears and the Sol Ring %v", pick.DiscardOptions, ids[2:4])
	}

	g = newCatalogGame(t)
	mine := pr7Creature(g, g.Seats[0].ID, "Mine", 3)
	theirs := pr7Creature(g, g.Seats[1].ID, "Theirs", 3)
	pbModal(t, g, "Poison the Waters", "Sorcery", poisonTheWatersOracle, 0)
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tg := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 2 || tg != 3 {
			t.Errorf("creature = %d/%d, want 2/3", p, tg)
		}
	}
}

// Mardu Charm: noncreature, nonland cards only; the Warriors have first
// strike.
func TestMarduCharm(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	ids := revealHand(victim, rhForest(), rhBolt(), pbBear(), pbRing())
	pbModal(t, g, "Mardu Charm", "Instant", marduCharmOracle, 2, pbPlayer(victim.ID))
	want := []uuid.UUID{ids[1], ids[3]}
	if pick := openPick(t, g); !sameIDs(pick.DiscardOptions, want) {
		t.Errorf("options = %v, want the Bolt and the Sol Ring %v", pick.DiscardOptions, want)
	}

	g = newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pbModal(t, g, "Mardu Charm", "Instant", marduCharmOracle, 1)
	var warriors []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Warrior" {
			warriors = append(warriors, c.InstanceID)
		}
	}
	if len(warriors) != 2 {
		t.Fatalf("warriors = %d, want 2", len(warriors))
	}
	for _, id := range warriors {
		if !hasEffectiveKeyword(t, g, id, "first strike") {
			t.Error("a Warrior has no first strike")
		}
	}
}

// Drill Bit: spectacle is offered only once an opponent has lost life
// this turn.
func TestDrillBitSpectacle(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	toMainForCost(t, g)
	revealHand(victim, rhBolt())
	bit := handCardFull(me, "Drill Bit", "Sorcery", "{2}{B}", drillBitOracle, []string{"B"})
	params := game.CastSpellParams{Strict: true, AlternativeCost: "spectacle", Targets: []game.TargetRef{pbPlayer(victim.ID)}}

	floatForTest(g, me, "B")
	if err := g.CastSpell(me.ID, bit, params); err == nil {
		t.Fatal("spectacle was accepted with no life lost this turn")
	}
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, victim.ID, -1); err != nil {
			t.Fatalf("life loss: %v", err)
		}
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, victim.ID, 2)
	})
	if err := g.CastSpell(me.ID, bit, params); err != nil {
		t.Fatalf("spectacle after an opponent lost life: %v", err)
	}
	passPriorityAroundTable(t, g)
	openPick(t, g)
}

// Dread Fugue: mana value 2 or less for {B}, any nonland card for its
// cleave cost.
func TestDreadFugueCleaveRemovesTheManaValueLimit(t *testing.T) {
	for _, tc := range []struct {
		alt  string
		pool string
		all  bool
	}{{"", "B", false}, {"cleave", "CCB", true}} {
		t.Run("alt="+tc.alt, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			toMainForCost(t, g)
			ids := revealHand(victim, rhForest(), rhBolt(), pbDivination(), rhFireball())
			floatForTest(g, me, tc.pool)
			fugue := handCardFull(me, "Dread Fugue", "Sorcery", "{B}", dreadFugueOracle, []string{"B"})
			if err := g.CastSpell(me.ID, fugue, game.CastSpellParams{
				Strict: true, AlternativeCost: tc.alt, Targets: []game.TargetRef{pbPlayer(victim.ID)},
			}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			passPriorityAroundTable(t, g)
			want := []uuid.UUID{ids[1], ids[3]}
			if tc.all {
				want = ids[1:]
			}
			if pick := openPick(t, g); !sameIDs(pick.DiscardOptions, want) {
				t.Errorf("options = %v, want %v", pick.DiscardOptions, want)
			}
		})
	}
}

// Venarian Glimmer: the announced X caps the mana value.
func TestVenarianGlimmerReadsX(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	ids := revealHand(victim, rhForest(), rhBolt(), pbBear(), pbDivination())
	castXSpell(t, g, "Venarian Glimmer", "Instant", venarianGlimmerOracle, "{X}{U}", 2,
		[]game.TargetRef{pbPlayer(victim.ID)})
	passPriorityAroundTable(t, g)
	pick := openPick(t, g)
	if !sameIDs(pick.DiscardOptions, ids[1:3]) {
		t.Errorf("options = %v, want the Bolt and the Bears %v", pick.DiscardOptions, ids[1:3])
	}
	if pick.DiscardLabel != "nonland card with mana value 2 or less" {
		t.Errorf("label = %q", pick.DiscardLabel)
	}
}

// Gruesome Discovery: the player discards two of their choice, or with
// morbid the caster picks two of any kind from the revealed hand.
func TestGruesomeDiscoveryMorbid(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	revealHand(victim, rhForest(), rhBolt(), pbBear())
	castCatalogSpell(t, g, "Gruesome Discovery", "Sorcery", gruesomeDiscoveryOracle, []game.TargetRef{pbPlayer(victim.ID)})
	passPriorityAroundTable(t, g)
	pbNoPick(t, g)
	if c := pbChooseCards(g, victim.ID); c == nil || c.ChooseMin != 2 {
		t.Fatalf("victim's own discard = %+v, want a choice of two", c)
	}

	g = newCatalogGame(t)
	victim = g.Seats[1]
	ids := revealHand(victim, rhForest(), rhBolt(), pbBear())
	g.TurnTally.CreaturesDied = 1
	castCatalogSpell(t, g, "Gruesome Discovery", "Sorcery", gruesomeDiscoveryOracle, []game.TargetRef{pbPlayer(victim.ID)})
	passPriorityAroundTable(t, g)
	pick := openPick(t, g)
	if pick.Chooser != g.Seats[0].ID || pick.Count != 2 || !sameIDs(pick.DiscardOptions, ids) {
		t.Errorf("morbid pick = chooser %v count %d options %v", pick.Chooser, pick.Count, pick.DiscardOptions)
	}
	if pbChooseCards(g, victim.ID) != nil {
		t.Error("morbid still asked the victim to choose")
	}
}

// River's Grasp: {U} bounces, {B} picks, both does both.
func TestRiversGraspReadsTheColoursSpent(t *testing.T) {
	for _, tc := range []struct {
		pool         string
		bounce, pick bool
	}{{"UCCC", true, false}, {"BCCC", false, true}, {"UBCC", true, true}} {
		t.Run(tc.pool, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			toMainForCost(t, g)
			bear := pr7Creature(g, victim.ID, "Theirs", 2)
			revealHand(victim, rhBolt())
			floatForTest(g, me, tc.pool)
			grasp := handCardFull(me, "River's Grasp", "Sorcery", "{3}{U/B}", riversGraspOracle, []string{"U", "B"})
			if err := g.CastSpell(me.ID, grasp, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
				{Kind: game.TargetCard, ID: bear, Slot: 0},
				{Kind: game.TargetPlayer, ID: victim.ID, Slot: 1},
			}}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			passPriorityAroundTable(t, g)
			bounced := victim.Hand.Contains(bear)
			if bounced != tc.bounce {
				t.Errorf("bounced = %v, want %v", bounced, tc.bounce)
			}
			picked := false
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceDiscardFromHand {
					picked = true
					if tc.bounce && !containsID(c.DiscardOptions, bear) {
						t.Error("the bounced creature is not in the revealed hand")
					}
				}
			}
			if picked != tc.pick {
				t.Errorf("pick open = %v, want %v", picked, tc.pick)
			}
		})
	}
}

func containsID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// Devour Intellect: an ordinary discard, unless Treasure mana paid for
// it.
func TestDevourIntellectTreasure(t *testing.T) {
	for _, treasure := range []bool{false, true} {
		g := newCatalogGame(t)
		me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		toMainForCost(t, g)
		revealHand(victim, rhForest(), rhBolt())
		if treasure {
			crackATreasureFor(t, g, me, "B")
		} else {
			floatForTest(g, me, "B")
		}
		devour := handCardFull(me, "Devour Intellect", "Sorcery", "{B}", devourIntellectOracle, []string{"B"})
		if err := g.CastSpell(me.ID, devour, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{pbPlayer(victim.ID)}}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if treasure {
			if pick := openPick(t, g); pick.DiscardLabel != "nonland card" {
				t.Errorf("label = %q", pick.DiscardLabel)
			}
			continue
		}
		pbNoPick(t, g)
		if pbChooseCards(g, victim.ID) == nil {
			t.Error("without Treasure mana the victim was not asked to discard")
		}
	}
}

// Humiliate: after the pick, a +1/+1 counter on a creature you choose.
func TestHumiliatePutsACounterOnYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := pr7Creature(g, me.ID, "Mine", 2)
	revealHand(victim, rhBolt())
	castCatalogSpell(t, g, "Humiliate", "Sorcery", humiliateOracle, []game.TargetRef{pbPlayer(victim.ID)})
	passPriorityAroundTable(t, g)
	pick := openPick(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, pick.DiscardOptions); err != nil {
		t.Fatalf("pick: %v", err)
	}
	var own *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOwnPermanents {
			own = c
		}
	}
	if own == nil {
		t.Fatal("no creature choice")
	}
	if err := g.ResolveOwnPermanents(own.ID, me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("choose creature: %v", err)
	}
	if got := countersOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// Humiliate with no creature: the pick still happens and nothing else
// is asked.
func TestHumiliateWithNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	revealHand(victim, rhBolt())
	castCatalogSpell(t, g, "Humiliate", "Sorcery", humiliateOracle, []game.TargetRef{pbPlayer(victim.ID)})
	passPriorityAroundTable(t, g)
	openPick(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOwnPermanents {
			t.Error("asked for a creature with none on the battlefield")
		}
	}
}

// Render Speechless: two +1/+1 counters on the second target.
func TestRenderSpeechlessCountersTheSecondTarget(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	theirs := pr7Creature(g, victim.ID, "Theirs", 2)
	revealHand(victim, rhBolt())
	castCatalogSpell(t, g, "Render Speechless", "Sorcery", renderSpeechlessOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: victim.ID, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	openPick(t, g)
	if got := countersOn(g, theirs, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

// Ego Drain: without a Faerie you exile a card from your own hand; with
// one you don't.
func TestEgoDrainExilesWithoutAFaerie(t *testing.T) {
	for _, faerie := range []bool{false, true} {
		g := newCatalogGame(t)
		me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		revealHand(victim, rhBolt())
		mine := handCardFull(me, "Divination", "Sorcery", "{2}{U}", "", []string{"U"})
		if faerie {
			pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Faerie Seer", TypeLine: "Creature — Faerie", Power: 1, Toughness: 1,
				Owner: me.ID, Controller: me.ID,
			})
		}
		castCatalogSpell(t, g, "Ego Drain", "Sorcery", egoDrainOracle, []game.TargetRef{pbPlayer(victim.ID)})
		passPriorityAroundTable(t, g)
		openPick(t, g)
		c := pbChooseCards(g, me.ID)
		if faerie {
			if c != nil {
				t.Error("asked to exile while controlling a Faerie")
			}
			continue
		}
		if c == nil {
			t.Fatal("not asked to exile a card")
		}
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{mine}); err != nil {
			t.Fatalf("exile: %v", err)
		}
		if !g.Exile.Contains(mine) {
			t.Error("the chosen card is not in exile")
		}
	}
}

// Memory Theft: the victim's face-up adventurer card in exile may go to
// their graveyard; another player's, or a card with no Adventure, is
// never offered.
func TestMemoryTheftOffersTheVictimsAdventurers(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	revealHand(victim, rhBolt())
	theirs := adventurerInExile(g, victim.ID)
	adventurerInExile(g, g.Seats[2].ID)
	g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Owner: victim.ID, Controller: victim.ID})

	castCatalogSpell(t, g, "Memory Theft", "Sorcery", memoryTheftOracle, []game.TargetRef{pbPlayer(victim.ID)})
	passPriorityAroundTable(t, g)
	openPick(t, g)
	c := pbChooseCards(g, me.ID)
	if c == nil {
		t.Fatal("no adventurer choice")
	}
	if c.ChooseMin != 0 || !sameIDs(c.ChooseCards, []uuid.UUID{theirs}) {
		t.Errorf("offer = %v (min %d), want only the victim's adventurer", c.ChooseCards, c.ChooseMin)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	if !victim.Graveyard.Contains(theirs) {
		t.Error("the adventurer is not in its owner's graveyard")
	}
}

func adventurerInExile(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Exile.PushTop(game.Card{
		InstanceID: id, Name: "Bonecrusher Giant", TypeLine: "Creature — Giant", Layout: game.LayoutAdventure,
		Owner: owner, Controller: owner,
	})
	return id
}
