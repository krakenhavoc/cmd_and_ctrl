package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func pbRing() game.Card {
	return game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"}
}

func pbPacifism() game.Card {
	return game.Card{Name: "Pacifism", TypeLine: "Enchantment — Aura", ManaCost: "{1}{W}", Colors: []string{"W"}}
}

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
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: owner, Controller: owner,
	})
}

// Auntie's Sentence and Dai Li Indoctrination: "nonland permanent card"
// takes the creature, artifact and enchantment, not the land or the
// instant or sorcery.
func TestNonlandPermanentCardPicks(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Auntie's Sentence", "Sorcery", effectstest.AuntiesSentenceOracle},
		{"Dai Li Indoctrination", "Sorcery — Lesson", effectstest.DaiLiIndoctrinationOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			victim := g.Seats[1]
			ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbBear(), pbRing(), pbPacifism(), effectstest.PbDivination())
			effectstest.PbModal(t, g, tc.name, tc.typeLine, tc.oracle, 0, effectstest.PbPlayer(victim.ID))
			pick := effectstest.OpenPick(t, g)
			if want := ids[2:5]; !effectstest.SameIDs(pick.DiscardOptions, want) {
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
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.Pr7Creature(g, g.Seats[1].ID, "Big Bear", 4)
	effectstest.PbModal(t, g, "Auntie's Sentence", "Sorcery", effectstest.AuntiesSentenceOracle, 1, effectstest.PbCard(bear))
	pbNoPick(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 2 || tg != 2 {
		t.Errorf("Big Bear = %d/%d, want 2/2", p, tg)
	}
}

// Dai Li Indoctrination's second bullet earthbends its own target land.
func TestDaiLiIndoctrinationEarthbendsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pbLand(g, me.ID)
	effectstest.PbModal(t, g, "Dai Li Indoctrination", "Sorcery — Lesson", effectstest.DaiLiIndoctrinationOracle, 1, effectstest.PbCard(land))
	if got := effectstest.CountersOn(g, land, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if !effectstest.LayeredCard(t, g, land).IsCreature() {
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
		{"Cerebral Confiscation", effectstest.CerebralConfiscationOracle, false},
		{"Splitting Headache", effectstest.SplittingHeadacheOracle, true},
	} {
		t.Run(tc.name+" discards two", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			victim := g.Seats[1]
			effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbBear())
			effectstest.PbModal(t, g, tc.name, "Sorcery", tc.oracle, 0, effectstest.PbPlayer(victim.ID))
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
			g := effectstest.NewCatalogGame(t)
			victim := g.Seats[1]
			ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbBear())
			effectstest.PbModal(t, g, tc.name, "Sorcery", tc.oracle, 1, effectstest.PbPlayer(victim.ID))
			want := ids[1:]
			if tc.wantAll {
				want = ids
			}
			if pick := effectstest.OpenPick(t, g); !effectstest.SameIDs(pick.DiscardOptions, want) {
				t.Errorf("options = %v, want %v", pick.DiscardOptions, want)
			}
		})
	}
}

// Poison the Waters: artifact or creature cards only, and the first
// bullet shrinks every creature.
func TestPoisonTheWaters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbBear(), pbRing(), pbPacifism())
	effectstest.PbModal(t, g, "Poison the Waters", "Sorcery", effectstest.PoisonTheWatersOracle, 1, effectstest.PbPlayer(victim.ID))
	if pick := effectstest.OpenPick(t, g); !effectstest.SameIDs(pick.DiscardOptions, ids[2:4]) {
		t.Errorf("options = %v, want the Bears and the Sol Ring %v", pick.DiscardOptions, ids[2:4])
	}

	g = effectstest.NewCatalogGame(t)
	mine := effectstest.Pr7Creature(g, g.Seats[0].ID, "Mine", 3)
	theirs := effectstest.Pr7Creature(g, g.Seats[1].ID, "Theirs", 3)
	effectstest.PbModal(t, g, "Poison the Waters", "Sorcery", effectstest.PoisonTheWatersOracle, 0)
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 2 || tg != 3 {
			t.Errorf("creature = %d/%d, want 2/3", p, tg)
		}
	}
}

// Mardu Charm: noncreature, nonland cards only; the Warriors have first
// strike.
func TestMarduCharm(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbBear(), pbRing())
	effectstest.PbModal(t, g, "Mardu Charm", "Instant", effectstest.MarduCharmOracle, 2, effectstest.PbPlayer(victim.ID))
	want := []uuid.UUID{ids[1], ids[3]}
	if pick := effectstest.OpenPick(t, g); !effectstest.SameIDs(pick.DiscardOptions, want) {
		t.Errorf("options = %v, want the Bolt and the Sol Ring %v", pick.DiscardOptions, want)
	}

	g = effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PbModal(t, g, "Mardu Charm", "Instant", effectstest.MarduCharmOracle, 1)
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
		if !effectstest.HasEffectiveKeyword(t, g, id, "first strike") {
			t.Error("a Warrior has no first strike")
		}
	}
}

// Drill Bit: spectacle is offered only once an opponent has lost life
// this turn.
func TestDrillBitSpectacle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	effectstest.RevealHand(victim, effectstest.RhBolt())
	bit := effectstest.HandCardFull(me, "Drill Bit", "Sorcery", "{2}{B}", effectstest.DrillBitOracle, []string{"B"})
	params := game.CastSpellParams{Strict: true, AlternativeCost: "spectacle", Targets: []game.TargetRef{effectstest.PbPlayer(victim.ID)}}

	effectstest.FloatForTest(g, me, "B")
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
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.OpenPick(t, g)
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
			g := effectstest.NewCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			effectstest.ToMainForCost(t, g)
			ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbDivination(), effectstest.RhFireball())
			effectstest.FloatForTest(g, me, tc.pool)
			fugue := effectstest.HandCardFull(me, "Dread Fugue", "Sorcery", "{B}", effectstest.DreadFugueOracle, []string{"B"})
			if err := g.CastSpell(me.ID, fugue, game.CastSpellParams{
				Strict: true, AlternativeCost: tc.alt, Targets: []game.TargetRef{effectstest.PbPlayer(victim.ID)},
			}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want := []uuid.UUID{ids[1], ids[3]}
			if tc.all {
				want = ids[1:]
			}
			if pick := effectstest.OpenPick(t, g); !effectstest.SameIDs(pick.DiscardOptions, want) {
				t.Errorf("options = %v, want %v", pick.DiscardOptions, want)
			}
		})
	}
}

// Venarian Glimmer: the announced X caps the mana value.
func TestVenarianGlimmerReadsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbBear(), effectstest.PbDivination())
	effectstest.CastXSpell(t, g, "Venarian Glimmer", "Instant", effectstest.VenarianGlimmerOracle, "{X}{U}", 2,
		[]game.TargetRef{effectstest.PbPlayer(victim.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenPick(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, ids[1:3]) {
		t.Errorf("options = %v, want the Bolt and the Bears %v", pick.DiscardOptions, ids[1:3])
	}
	if pick.DiscardLabel != "nonland card with mana value 2 or less" {
		t.Errorf("label = %q", pick.DiscardLabel)
	}
}

// Gruesome Discovery: the player discards two of their choice, or with
// morbid the caster picks two of any kind from the revealed hand.
func TestGruesomeDiscoveryMorbid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]
	effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbBear())
	effectstest.CastCatalogSpell(t, g, "Gruesome Discovery", "Sorcery", effectstest.GruesomeDiscoveryOracle, []game.TargetRef{effectstest.PbPlayer(victim.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	pbNoPick(t, g)
	if c := pbChooseCards(g, victim.ID); c == nil || c.ChooseMin != 2 {
		t.Fatalf("victim's own discard = %+v, want a choice of two", c)
	}

	g = effectstest.NewCatalogGame(t)
	victim = g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbBear())
	g.TurnTally.CreaturesDied = 1
	effectstest.CastCatalogSpell(t, g, "Gruesome Discovery", "Sorcery", effectstest.GruesomeDiscoveryOracle, []game.TargetRef{effectstest.PbPlayer(victim.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenPick(t, g)
	if pick.Chooser != g.Seats[0].ID || pick.Count != 2 || !effectstest.SameIDs(pick.DiscardOptions, ids) {
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
			g := effectstest.NewCatalogGame(t)
			me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			effectstest.ToMainForCost(t, g)
			bear := effectstest.Pr7Creature(g, victim.ID, "Theirs", 2)
			effectstest.RevealHand(victim, effectstest.RhBolt())
			effectstest.FloatForTest(g, me, tc.pool)
			grasp := effectstest.HandCardFull(me, "River's Grasp", "Sorcery", "{3}{U/B}", effectstest.RiversGraspOracle, []string{"U", "B"})
			if err := g.CastSpell(me.ID, grasp, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
				{Kind: game.TargetCard, ID: bear, Slot: 0},
				{Kind: game.TargetPlayer, ID: victim.ID, Slot: 1},
			}}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			bounced := victim.Hand.Contains(bear)
			if bounced != tc.bounce {
				t.Errorf("bounced = %v, want %v", bounced, tc.bounce)
			}
			picked := false
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceDiscardFromHand {
					picked = true
					if tc.bounce && !effectstest.ContainsID(c.DiscardOptions, bear) {
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

// Devour Intellect: an ordinary discard, unless Treasure mana paid for
// it.
func TestDevourIntellectTreasure(t *testing.T) {
	for _, treasure := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt())
		if treasure {
			effectstest.CrackATreasureFor(t, g, me, "B")
		} else {
			effectstest.FloatForTest(g, me, "B")
		}
		devour := effectstest.HandCardFull(me, "Devour Intellect", "Sorcery", "{B}", effectstest.DevourIntellectOracle, []string{"B"})
		if err := g.CastSpell(me.ID, devour, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.PbPlayer(victim.ID)}}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if treasure {
			if pick := effectstest.OpenPick(t, g); pick.DiscardLabel != "nonland card" {
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
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := effectstest.Pr7Creature(g, me.ID, "Mine", 2)
	effectstest.RevealHand(victim, effectstest.RhBolt())
	effectstest.CastCatalogSpell(t, g, "Humiliate", "Sorcery", effectstest.HumiliateOracle, []game.TargetRef{effectstest.PbPlayer(victim.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenPick(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOwnPermanents {
			t.Fatal("the creature choice is open with the pick: it ran before the answer")
		}
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, pick.DiscardOptions); err != nil {
		t.Fatalf("pick: %v", err)
	}
	g.SettleResolution()
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
	if got := effectstest.CountersOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// Humiliate with no creature: the pick still happens and nothing else
// is asked.
func TestHumiliateWithNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]
	effectstest.RevealHand(victim, effectstest.RhBolt())
	effectstest.CastCatalogSpell(t, g, "Humiliate", "Sorcery", effectstest.HumiliateOracle, []game.TargetRef{effectstest.PbPlayer(victim.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.OpenPick(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOwnPermanents {
			t.Error("asked for a creature with none on the battlefield")
		}
	}
}

// Render Speechless: two +1/+1 counters on the second target.
func TestRenderSpeechlessCountersTheSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := effectstest.Pr7Creature(g, victim.ID, "Theirs", 2)
	effectstest.RevealHand(victim, effectstest.RhBolt())
	effectstest.CastCatalogSpell(t, g, "Render Speechless", "Sorcery", effectstest.RenderSpeechlessOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: victim.ID, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenPick(t, g)
	if got := effectstest.CountersOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Fatalf("+1/+1 counters = %d with the pick open, want 0: they came before the answer", got)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, pick.DiscardOptions); err != nil {
		t.Fatalf("pick: %v", err)
	}
	g.SettleResolution()
	if got := effectstest.CountersOn(g, theirs, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

// Ego Drain: without a Faerie you exile a card from your own hand; with
// one you don't.
func TestEgoDrainExilesWithoutAFaerie(t *testing.T) {
	for _, faerie := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		effectstest.RevealHand(victim, effectstest.RhBolt())
		mine := effectstest.HandCardFull(me, "Divination", "Sorcery", "{2}{U}", "", []string{"U"})
		if faerie {
			effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Faerie Seer", TypeLine: "Creature — Faerie", Power: 1, Toughness: 1,
				Owner: me.ID, Controller: me.ID,
			})
		}
		effectstest.CastCatalogSpell(t, g, "Ego Drain", "Sorcery", effectstest.EgoDrainOracle, []game.TargetRef{effectstest.PbPlayer(victim.ID)})
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.OpenPick(t, g)
		if pbChooseCards(g, me.ID) != nil {
			t.Fatal("asked to exile from your hand with the pick open: it ran before the answer")
		}
		if err := g.ResolvePendingChoice(pick.ID, me.ID, pick.DiscardOptions); err != nil {
			t.Fatalf("pick: %v", err)
		}
		g.SettleResolution()
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
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	hand := effectstest.RevealHand(victim, effectstest.RhBolt())
	theirs := adventurerInExile(g, victim.ID)
	adventurerInExile(g, g.Seats[2].ID)
	g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Owner: victim.ID, Controller: victim.ID})

	effectstest.CastCatalogSpell(t, g, "Memory Theft", "Sorcery", effectstest.MemoryTheftOracle, []game.TargetRef{effectstest.PbPlayer(victim.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenPick(t, g)
	// The Adventure offer is the sentence after the discard, so it waits for the
	// pick (CR 608.2c, #4155).
	if pbChooseCards(g, me.ID) != nil {
		t.Fatal("the Adventure offer is open with the pick open")
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, hand); err != nil {
		t.Fatalf("taking the Bolt: %v", err)
	}
	g.SettleResolution()
	c := pbChooseCards(g, me.ID)
	if c == nil {
		t.Fatal("no adventurer choice")
	}
	if c.ChooseMin != 0 || !effectstest.SameIDs(c.ChooseCards, []uuid.UUID{theirs}) {
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
