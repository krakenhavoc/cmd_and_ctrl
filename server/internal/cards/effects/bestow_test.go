package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bestow_test.go — ADR 0141 (#2862), CR 702.103: the bestow cards on
// the real catalog. The cast, the Aura spell on the stack, the attach,
// CR 702.103e's creature resolution and CR 702.103f's unattached
// creature.

const (
	nighthowlerOracle      = "57b3f7fc-1812-4134-a645-6cef48a8aa71"
	countlessBattlesOracle = "441b51b0-ecd0-466d-9ca8-e754eb563aec"
	boonSatyrOracle        = "aedcd2fb-813a-4915-b7b9-ac7479863462"
	celestialArchonOracle  = "d2b29bb0-9fb5-46cd-a546-fa6c4ff6113c"
	hopefulEidolonOracle   = "13c1ca1e-cd17-49ef-bc31-97bff99ec1e3"
	nyxbornRollickerOracle = "ebf2974a-963e-425a-8f8c-55f0a36984c3"
)

func TestBestowCardsAreRegisteredWithTheirOffer(t *testing.T) {
	for oracle, want := range map[string]struct{ name, cost string }{
		nighthowlerOracle:      {"Nighthowler", "{2}{B}{B}"},
		countlessBattlesOracle: {"Eidolon of Countless Battles", "{2}{W}{W}"},
		boonSatyrOracle:        {"Boon Satyr", "{3}{G}{G}"},
		celestialArchonOracle:  {"Celestial Archon", "{5}{W}{W}"},
		hopefulEidolonOracle:   {"Hopeful Eidolon", "{3}{W}"},
		nyxbornRollickerOracle: {"Nyxborn Rollicker", "{1}{R}"},
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != want.name {
			t.Errorf("%s is not registered under %s", want.name, oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: want Full with no caveat, got %v %q", want.name, spec.Completeness, spec.Caveats)
		}
		alt := game.AlternativeCostByKey(oracle, game.BestowKey)
		if alt == nil || !alt.Bestow || alt.ManaCost != want.cost || alt.Targets == nil {
			t.Errorf("%s: bestow offer %+v, want Bestow %s with an enchant creature clause", want.name, alt, want.cost)
		}
	}
}

// bestowCard puts a bestow card in the active seat's hand and walks to
// a main phase.
func bestowCard(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id
}

// castBestowed casts the card bestowed onto host.
func castBestowed(t *testing.T, g *game.Game, id, host uuid.UUID) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		AlternativeCost: game.BestowKey,
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: host}},
	}); err != nil {
		t.Fatalf("cast bestowed: %v", err)
	}
}

// cardState reads one card's zone flags and type predicates under the
// read lock.
type bestowState struct {
	onStack, onBattlefield   bool
	creature, aura, bestowed bool
	attachedTo               uuid.UUID
	horror                   bool
}

func readBestowState(g *game.Game, id uuid.UUID) bestowState {
	var s bestowState
	g.ReadSnapshot(func() {
		read := func(c game.Card) {
			s.creature = c.IsCreature()
			s.aura = c.IsAura()
			s.bestowed = c.Bestowed
			s.horror = c.HasSubtype("Horror")
			if c.AttachedTo.Kind == game.TargetCard {
				s.attachedTo = c.AttachedTo.ID
			}
		}
		for _, c := range g.Stack.Cards {
			if c.InstanceID == id {
				s.onStack = true
				read(c)
			}
		}
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				s.onBattlefield = true
				read(c)
			}
		}
	})
	return s
}

// Bestowed, Nighthowler is an Aura spell on the stack (CR 702.103b),
// attaches to its target as it resolves (CR 608.3c), is not a creature
// on the battlefield, and gives the host its +X/+X.
func TestNighthowlerBestowedIsAnAuraThatPumpsItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	pushGraveyardPermanent(me, "Dead Elf", "Creature — Elf", "{G}")
	howler := bestowCard(t, g, "Nighthowler", "Enchantment Creature — Horror", nighthowlerOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	castBestowed(t, g, howler, bear)
	s := readBestowState(g, howler)
	if !s.onStack || s.creature || !s.aura || s.horror || !s.bestowed {
		t.Fatalf("on the stack a bestowed spell is an Aura spell with no creature types: %+v", s)
	}

	passPriorityAroundTable(t, g)
	s = readBestowState(g, howler)
	if !s.onBattlefield || s.attachedTo != bear {
		t.Fatalf("resolved bestowed: want on the battlefield attached to the bear, got %+v", s)
	}
	if s.creature || !s.aura {
		t.Errorf("a bestowed Aura is not a creature: %+v", s)
	}
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("host gets +X/+X for two creature cards: power %d, want 4", got)
	}
	if got := effectiveToughness(t, g, bear); got != 4 {
		t.Errorf("host toughness %d, want 4", got)
	}
}

// CR 702.103f: when the host dies, the Aura becomes unattached and
// stays on the battlefield as a creature, and its X now counts the
// host as well.
func TestNighthowlerBecomesACreatureWhenItsHostDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	howler := bestowCard(t, g, "Nighthowler", "Enchantment Creature — Horror", nighthowlerOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castBestowed(t, g, howler, bear)
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy host: %v", err)
		}
	})
	g.RunStateChecksForTest()

	s := readBestowState(g, howler)
	if !s.onBattlefield {
		t.Fatalf("CR 702.103f: the Aura stays on the battlefield, got %+v", s)
	}
	if s.bestowed || s.aura || !s.creature || !s.horror || s.attachedTo != uuid.Nil {
		t.Errorf("unattached, it is an enchantment creature — Horror again: %+v", s)
	}
	if got := effectivePower(t, g, howler); got != 2 {
		t.Errorf("X counts both creature cards, the host included: power %d, want 2", got)
	}
}

// A bestowed 0/0 whose host dies with nothing else in a graveyard is a
// creature again, and a 0/0 creature dies (CR 704.5f) — it is not put
// into the graveyard as an Aura (CR 704.5m).
func TestEidolonOfCountlessBattlesStaysAsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eidolon := bestowCard(t, g, "Eidolon of Countless Battles", "Enchantment Creature — Spirit", countlessBattlesOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	castBestowed(t, g, eidolon, bear)
	passPriorityAroundTable(t, g)

	// Two creatures and one Aura (the Eidolon itself): +3/+3 on the bear.
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("bestowed: bear power %d, want 2 + 3", got)
	}
	if got := effectivePower(t, g, other); got != 1 {
		t.Errorf("only the enchanted creature is pumped: elf power %d", got)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	// Now a creature with the elf: two creatures, no Auras.
	if got := effectivePower(t, g, eidolon); got != 2 {
		t.Errorf("unattached Eidolon: power %d, want 2", got)
	}
}

// CR 702.103e and 608.3b: a bestowed spell whose target is gone when it
// resolves is not countered. It resolves as a creature spell and enters
// unattached.
func TestBestowedSpellWithAnIllegalTargetResolvesAsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	satyr := bestowCard(t, g, "Boon Satyr", "Enchantment Creature — Satyr", boonSatyrOracle, 4, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castBestowed(t, g, satyr, bear)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })

	passPriorityAroundTable(t, g)
	s := readBestowState(g, satyr)
	if !s.onBattlefield {
		t.Fatalf("a bestowed spell with an illegal target still resolves, got %+v", s)
	}
	if s.bestowed || !s.creature || s.aura || s.attachedTo != uuid.Nil {
		t.Errorf("it resolves as an unattached creature: %+v", s)
	}
	if got := effectivePower(t, g, satyr); got != 4 {
		t.Errorf("a 4/2 creature: power %d", got)
	}
}

// The ordinary cast is untouched: a creature spell with no target.
func TestBoonSatyrHardCastIsACreature(t *testing.T) {
	g := newCatalogGame(t)
	satyr := castCatalogSpell(t, g, "Boon Satyr", "Enchantment Creature — Satyr", boonSatyrOracle, nil)
	if s := readBestowState(g, satyr); !s.creature || s.aura || s.bestowed {
		t.Errorf("hard-cast on the stack: a creature spell, got %+v", s)
	}
	passPriorityAroundTable(t, g)
	if s := readBestowState(g, satyr); !s.onBattlefield || !s.creature || s.aura {
		t.Errorf("hard-cast resolves as a creature: %+v", s)
	}
}

// An Aura spell requires a target (CR 303.4a, 702.103b), and enchant
// creature means a creature.
func TestBestowNeedsACreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rollicker := bestowCard(t, g, "Nyxborn Rollicker", "Enchantment Creature — Satyr", nyxbornRollickerOracle, 1, 1)
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain",
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, rollicker, game.CastSpellParams{AlternativeCost: game.BestowKey}); err == nil {
		t.Error("a bestowed cast with no target was accepted")
	}
	err := g.CastSpell(me.ID, rollicker, game.CastSpellParams{
		AlternativeCost: game.BestowKey,
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	})
	if err == nil {
		t.Error("a bestowed cast onto a land was accepted")
	}
	if errors.Is(err, game.ErrCardNotFound) {
		t.Errorf("refused for the wrong reason: %v", err)
	}
}

// Celestial Archon's grant reaches the host, and the host loses it when
// the Archon is gone.
func TestCelestialArchonBestowedGrantsFlyingAndFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	archon := bestowCard(t, g, "Celestial Archon", "Enchantment Creature — Archon", celestialArchonOracle, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castBestowed(t, g, archon, bear)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 6 {
		t.Errorf("bear power %d, want 6", got)
	}
	for _, kw := range []string{"flying", "first strike"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the host has %s", kw)
		}
	}
}

// Hopeful Eidolon as an Aura gives lifelink, and the Eidolon itself is
// not a creature that could attack or block.
func TestHopefulEidolonBestowedGivesLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eidolon := bestowCard(t, g, "Hopeful Eidolon", "Enchantment Creature — Spirit", hopefulEidolonOracle, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castBestowed(t, g, eidolon, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the host has lifelink")
	}
	if s := readBestowState(g, eidolon); s.creature {
		t.Errorf("a bestowed Aura is not a creature: %+v", s)
	}
}
