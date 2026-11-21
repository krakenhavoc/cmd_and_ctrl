package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b08_cards_test.go — Phyrexia: All Will Be One, batch 8 (#5479,
// S82): proliferate spells and simple instants and sorceries. Each test
// plays the real catalog entry.

const (
	s82b8RejectImperfectionOracle = "25781dc5-d6ea-4273-9fe3-06a8808b3718"
	s82b8SerumSnareOracle         = "fbc7350c-d392-44bf-92f7-8dd4743a73b4"
	s82b8WhisperOracle            = "23c2cdc9-2b16-423a-b932-2daadeb1021b"
	s82b8CanopyOracle             = "b3360259-19df-43c6-95a9-e76770f2fb77"
	s82b8DrownOracle              = "e4912bc3-bee9-4a2f-a13e-3a99018f8a65"
	s82b8ExpandOracle             = "0fac3187-9b86-4854-854b-a4c203526a64"
	s82b8ThirstingRootsOracle     = "c0ef21d5-f4bf-4417-b701-d98c2dbbaa43"
	s82b8RestorationOracle        = "05d1e797-bda2-4411-a546-919493768650"
	s82b8VatEmergenceOracle       = "03e59169-d8f8-40da-b5f7-645d16365e7b"
	s82b8InsightOracle            = "cf6ff089-63d7-4a98-b95d-2837c017a866"
	s82b8CuriosityOracle          = "6ebbf025-2996-46f2-ac23-992909251525"
	s82b8FeedOracle               = "405a2cf9-62a0-460c-bf2a-ce4c351b61ce"
	s82b8MinorMisstepOracle       = "d91b3c6f-f8a1-4859-90ca-22cd79517552"
)

func TestS82OneB08CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82b8RejectImperfectionOracle: "Reject Imperfection",
		s82b8SerumSnareOracle:         "Serum Snare",
		s82b8WhisperOracle:            "Whisper of the Dross",
		s82b8CanopyOracle:             "Carnivorous Canopy",
		s82b8DrownOracle:              "Drown in Ichor",
		s82b8ExpandOracle:             "Expand the Sphere",
		s82b8ThirstingRootsOracle:     "Thirsting Roots",
		s82b8RestorationOracle:        "Unnatural Restoration",
		s82b8VatEmergenceOracle:       "Vat Emergence",
		s82b8InsightOracle:            "Vivisurgeon's Insight",
		s82b8CuriosityOracle:          "Distorted Curiosity",
		s82b8FeedOracle:               "Feed the Infection",
		s82b8MinorMisstepOracle:       "Minor Misstep",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// s82b8Counted is a creature on the battlefield with one +1/+1 counter, so a
// proliferate has something to grow.
func s82b8Counted(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.B12Creature(g, owner, "Counted Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(id, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	return id
}

func s82b8Prolif(t *testing.T, g *game.Game, id uuid.UUID, want int) {
	t.Helper()
	if effectstest.ProliferatePrompt(g) == nil {
		t.Fatal("no proliferate prompt")
	}
	effectstest.AnswerProliferate(t, g, id)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != want {
		t.Errorf("%d +1/+1 counters after the proliferate, want %d", n, want)
	}
}

func TestS82OneB08RejectImperfectionProliferatesOnlyForSmallSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := s82b8Counted(t, g, me.ID)
	small := effectstest.CastSpellWithCost(t, g, "Cheap Bear", "Creature — Bear", "", "{2}{G}", nil)
	effectstest.CastCatalogSpell(t, g, "Reject Imperfection", "Instant", s82b8RejectImperfectionOracle, effectstest.B16TargetCard(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !me.Graveyard.Contains(small) {
		t.Fatal("the spell was not countered")
	}
	s82b8Prolif(t, g, mine, 2)

	big := effectstest.CastWithCost(t, g, "Wurm", "Creature — Wurm", "{4}{G}{G}", "")
	effectstest.CastCatalogSpell(t, g, "Reject Imperfection", "Instant", s82b8RejectImperfectionOracle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(big) {
		t.Fatal("the big spell was not countered")
	}
	if effectstest.ProliferatePrompt(g) != nil {
		t.Error("a mana value 6 spell must not proliferate")
	}
}

func TestS82OneB08SerumSnareBouncesThenProliferatesForSmallPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := s82b8Counted(t, g, me.ID)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	if err := effectstest.CastCatalogSpellErr(t, g, "Serum Snare", "Instant", s82b8SerumSnareOracle, effectstest.CardRefs(land)); err == nil {
		t.Fatal("Serum Snare was cast at a land")
	}
	small := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Trinket",
		TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Serum Snare", "Instant", s82b8SerumSnareOracle, effectstest.CardRefs(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !opp.Hand.Contains(small) {
		t.Fatal("the permanent was not returned to its owner's hand")
	}
	s82b8Prolif(t, g, mine, 2)

	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Colossus",
		TypeLine: "Artifact", ManaCost: "{4}{G}", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Serum Snare", "Instant", s82b8SerumSnareOracle, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(big) {
		t.Fatal("the big permanent was not returned")
	}
	if effectstest.ProliferatePrompt(g) != nil {
		t.Error("a mana value 5 permanent must not proliferate")
	}
}

func TestS82OneB08WhisperAndDrownShrinkThenProliferate(t *testing.T) {
	for _, c := range []struct {
		name, typ, oracle string
		pt                int
	}{
		{"Whisper of the Dross", "Instant", s82b8WhisperOracle, 1},
		{"Drown in Ichor", "Sorcery", s82b8DrownOracle, 4},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := s82b8Counted(t, g, me.ID)
		victim := effectstest.B12Creature(g, opp.ID, "Big Victim", "Creature — Ogre", 6, 6)
		effectstest.CastCatalogSpell(t, g, c.name, c.typ, c.oracle, effectstest.CardRefs(victim))
		effectstest.PassPriorityAroundTable(t, g)
		cs := effectstest.EffectiveOf(t, g, victim)
		if cs.Power != 6-c.pt || cs.Toughness != 6-c.pt {
			t.Errorf("%s: victim is %d/%d, want %d/%d", c.name, cs.Power, cs.Toughness, 6-c.pt, 6-c.pt)
		}
		s82b8Prolif(t, g, mine, 2)
	}
}

func TestS82OneB08CarnivorousCanopyDestroysThenProliferatesForSmallPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := s82b8Counted(t, g, me.ID)
	ground := effectstest.B12Creature(g, opp.ID, "Ground Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Carnivorous Canopy", "Sorcery", s82b8CanopyOracle, effectstest.CardRefs(ground)); err == nil {
		t.Fatal("Carnivorous Canopy was cast at a creature without flying")
	}
	small := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Trinket",
		TypeLine: "Artifact", ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Carnivorous Canopy", "Sorcery", s82b8CanopyOracle, effectstest.CardRefs(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Fatal("the artifact was not destroyed")
	}
	s82b8Prolif(t, g, mine, 2)

	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Old Charm",
		TypeLine: "Enchantment", ManaCost: "{3}{W}", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Carnivorous Canopy", "Sorcery", s82b8CanopyOracle, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Fatal("the enchantment was not destroyed")
	}
	if effectstest.ProliferatePrompt(g) != nil {
		t.Error("a mana value 4 permanent must not proliferate")
	}
}

func TestS82OneB08ThirstingRootsModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := s82b8Counted(t, g, me.ID)
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.CastCatalogSpellWithModes(t, g, "Thirsting Roots", "Sorcery", s82b8ThirstingRootsOracle, []int{1})
	effectstest.SettleAnsweringProliferate(t, g)
	if n := effectstest.CountersOn(g, mine, game.CounterPlusOne); n != 2 {
		t.Errorf("proliferate mode: %d counters, want 2", n)
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpellWithModes(t, g, "Thirsting Roots", "Sorcery", s82b8ThirstingRootsOracle, []int{0})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1 from the search mode", hand, me.Hand.Size())
	}
	if effectstest.ProliferatePrompt(g) != nil {
		t.Error("the search mode proliferated")
	}
}

func TestS82OneB08ExpandTheSphereProliferatesForEachMissingLand(t *testing.T) {
	for _, lands := range []int{2, 1, 0} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		mine := s82b8Counted(t, g, me.ID)
		effectstest.AdvanceToMain(t, g)
		var cards []game.Card
		for i := 0; i < 6; i++ {
			c := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
			if i < lands {
				c.Name, c.TypeLine = "Forest", "Basic Land — Forest"
			}
			cards = append(cards, c)
		}
		ids := effectstest.DftMPLibraryTop(me, cards...)
		effectstest.CastCatalogSpell(t, g, "Expand the Sphere", "Sorcery", s82b8ExpandOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if ch := effectstest.ChooseCardsChoiceFor(g, me.ID); ch != nil {
			effectstest.AnswerChooseCards(t, g, me.ID, ids[:lands]...)
		}
		for i := 0; i < lands; i++ {
			if !g.Battlefield.Contains(ids[i]) || !effectstest.C2b16IsTapped(g, ids[i]) {
				t.Errorf("lands=%d: land %d is not on the battlefield tapped", lands, i)
			}
		}
		for i := lands; i < 6; i++ {
			if !me.Library.Contains(ids[i]) {
				t.Errorf("lands=%d: card %d is not back in the library", lands, i)
			}
		}
		want := 1 + (2 - lands)
		for i := 0; i < 2-lands; i++ {
			if effectstest.ProliferatePrompt(g) == nil {
				t.Fatalf("lands=%d: proliferate %d of %d did not prompt", lands, i+1, 2-lands)
			}
			effectstest.AnswerProliferate(t, g, mine)
			effectstest.PassPriorityAroundTable(t, g)
		}
		if effectstest.ProliferatePrompt(g) != nil {
			t.Errorf("lands=%d: proliferated too many times", lands)
		}
		if n := effectstest.CountersOn(g, mine, game.CounterPlusOne); n != want {
			t.Errorf("lands=%d: %d counters, want %d", lands, n, want)
		}
	}
}

func TestS82OneB08UnnaturalRestorationReturnsAPermanentCardThenProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := s82b8Counted(t, g, me.ID)
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := effectstest.CastCatalogSpellErr(t, g, "Unnatural Restoration", "Sorcery", s82b8RestorationOracle, effectstest.CardRefs(spell)); err == nil {
		t.Fatal("Unnatural Restoration was cast at a sorcery card")
	}
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Old Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Unnatural Restoration", "Sorcery", s82b8RestorationOracle, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Fatal("the permanent card was not returned to hand")
	}
	s82b8Prolif(t, g, mine, 2)
}

func TestS82OneB08VatEmergenceReanimatesFromAnyGraveyardThenProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := s82b8Counted(t, g, me.ID)
	corpse := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: corpse, Name: "Dead Ogre", TypeLine: "Creature — Ogre",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Vat Emergence", "Sorcery", s82b8VatEmergenceOracle, effectstest.CardRefs(corpse))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(corpse) {
		t.Fatal("the creature did not enter")
	}
	if c, ok := g.LookupCardForEffect(corpse); !ok || c.Controller != me.ID {
		t.Errorf("the creature is not under my control: %+v", c)
	}
	s82b8Prolif(t, g, mine, 2)
}

func TestS82OneB08VivisurgeonsInsightDrawsThreeThenProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := s82b8Counted(t, g, me.ID)
	effectstest.AdvanceToMain(t, g)
	before := effectstest.HandSize(me)
	effectstest.CastCatalogSpell(t, g, "Vivisurgeon's Insight", "Sorcery", s82b8InsightOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != before+3 {
		t.Errorf("hand %d → %d, want +3", before, got)
	}
	s82b8Prolif(t, g, mine, 2)
}

func TestS82OneB08DistortedCuriosityCostsTwoLessWhenCorrupted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Distorted Curiosity", TypeLine: "Sorcery",
		ManaCost: "{2}{U}", OracleID: s82b8CuriosityOracle, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for {U} without corrupted")
	}
	opp.Poison = 2
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for {U} at two poison counters")
	}
	opp.Poison = 3
	before := effectstest.HandSize(me)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("corrupted cast for {U}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != before-1+2 {
		t.Errorf("hand %d → %d, want the spell gone and two drawn", before, got)
	}
}

func TestS82OneB08FeedTheInfectionDrainsOnlyCorruptedOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	g.Seats[1].Poison = 3
	g.Seats[2].Poison = 2
	g.Seats[3].Poison = 5
	life := effectstest.LifeOfOpponents(g)
	mine, hand := me.Life, effectstest.HandSize(me)
	effectstest.CastCatalogSpell(t, g, "Feed the Infection", "Sorcery", s82b8FeedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine-3 || effectstest.HandSize(me) != hand+3 {
		t.Errorf("caster: life %d → %d, hand %d → %d; want -3 life, +3 cards", mine, me.Life, hand, effectstest.HandSize(me))
	}
	got := effectstest.LifeOfOpponents(g)
	if got[0] != life[0]-3 || got[1] != life[1] || got[2] != life[2]-3 {
		t.Errorf("opponent life %v → %v, want seats 1 and 3 to lose 3", life, got)
	}
}

func TestS82OneB08MinorMisstepCountersOnlyMV1OrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	big := effectstest.CastWithCost(t, g, "Wurm", "Creature — Wurm", "{4}{G}{G}", "")
	cheap := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cheap, Name: "Minor Misstep", TypeLine: "Instant",
		OracleID: s82b8MinorMisstepOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, cheap, game.CastSpellParams{Targets: effectstest.B16TargetCard(big), Strict: true}); err == nil {
		t.Fatal("Minor Misstep was cast at a mana value 6 spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	one := effectstest.CastWithCost(t, g, "Elf", "Creature — Elf", "{G}", "")
	effectstest.CastCatalogSpell(t, g, "Minor Misstep", "Instant", s82b8MinorMisstepOracle, effectstest.B16TargetCard(one))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(one) || !me.Graveyard.Contains(one) {
		t.Error("the mana value 1 spell was not countered")
	}
}
