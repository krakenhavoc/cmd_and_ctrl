package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// everquill_phoenix_test.go — ADR 0223 (#5910), CR 702.140 and CR 730:
// mutate on the real catalog, with Everquill Phoenix. Over and under,
// abilities from the cards underneath, "whenever this creature
// mutates", an illegal target, the exits, a copy and a restore.

const (
	everquillOracle = "56ba78c1-d1eb-4cc1-b70d-98c35d2972ba"
	llanowarOracle  = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
)

func TestEverquillPhoenixIsRegisteredWithItsMutateOffer(t *testing.T) {
	spec, ok := Lookup(everquillOracle)
	if !ok || spec.Name != "Everquill Phoenix" {
		t.Fatalf("Everquill Phoenix is not registered: %+v", spec)
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("want Full with no caveat, got %v %q", spec.Completeness, spec.Caveats)
	}
	alt := game.AlternativeCostByKey(everquillOracle, game.MutateKey)
	if alt == nil || alt.ManaCost != "{3}{R}" || alt.Targets == nil || !alt.Targets.OwnedBySourceOwner {
		t.Fatalf("mutate offer %+v, want {3}{R} targeting a creature with the spell's owner", alt)
	}
}

// phoenixInHand puts an Everquill Phoenix in `owner`'s hand and walks
// to a main phase.
func phoenixInHand(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.PlayerByID(owner).Hand.PushTop(game.Card{
		InstanceID: id, Name: "Everquill Phoenix", TypeLine: "Creature — Phoenix",
		OracleID: everquillOracle, ManaCost: "{2}{R}{R}", Colors: []string{"R"},
		Keywords: []string{"flying", "mutate"}, Power: 4, Toughness: 4,
		Owner: owner, Controller: owner,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id
}

// llanowarOn puts a catalog Llanowar Elves ("{T}: Add {G}.") on the
// battlefield under its owner, since an earlier turn.
func llanowarOn(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		OracleID: llanowarOracle, ManaCost: "{G}", Colors: []string{"G"},
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

func castMutating(g *game.Game, caster, spell, host uuid.UUID, strict bool) error {
	return g.CastSpell(caster, spell, game.CastSpellParams{
		AlternativeCost: game.MutateKey,
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: host}},
		Strict:          strict,
	})
}

// resolveMutate passes priority until the over/under question, answers
// it (0 over, 1 under), and resolves whatever is left.
func resolveMutate(t *testing.T, g *game.Game, controller uuid.UUID, under bool) {
	t.Helper()
	passPriorityAroundTable(t, g)
	idx := 0
	if under {
		idx = 1
	}
	answerOptionPick(t, g, controller, idx)
	passPriorityAroundTable(t, g)
}

func featherCount(g *game.Game, owner uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Feather" && c.Controller == owner && c.IsToken() && c.IsArtifact() && !c.IsCreature() {
				n++
			}
		}
	})
	return n
}

// mergedView reads a battlefield permanent's merged facts under the
// write lock, with the layer cache caught up.
type mergedFacts struct {
	onBattlefield    bool
	name             string
	power, toughness int
	flying           bool
	tapsForMana      bool
	components       []uuid.UUID
	componentNames   []string
}

func readMerged(t *testing.T, g *game.Game, id uuid.UUID) mergedFacts {
	t.Helper()
	var f mergedFacts
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		return f
	}
	f.onBattlefield = true
	f.name = c.Name
	f.power, f.toughness = c.PowerForComparison(), c.CurrentToughness()
	f.flying = game.HasKeyword(c, "flying")
	f.tapsForMana = len(game.ManaAbilitiesForCard(*c)) > 0
	for _, part := range game.MergedComponentsForEffect(*c) {
		f.components = append(f.components, part.InstanceID)
		f.componentNames = append(f.componentNames, part.Name)
	}
	return f
}

func zoneHolds(z *game.Zone, id uuid.UUID) bool { return z != nil && z.Contains(id) }

// Over: the Elves become Everquill Phoenix — the same object, its
// counter kept (CR 730.2c) — with the Elves' mana ability from under it
// (CR 702.140e), and the Phoenix's trigger makes a Feather (CR
// 702.140d). The Phoenix card left the stack and entered nowhere.
func TestEverquillPhoenixMutatesOverAndMakesAFeather(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := llanowarOn(g, me.ID)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(elves, "+1/+1", 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	phoenix := phoenixInHand(t, g, me.ID)
	if err := castMutating(g, me.ID, phoenix, elves, false); err != nil {
		t.Fatalf("cast mutating: %v", err)
	}
	resolveMutate(t, g, me.ID, false)

	f := readMerged(t, g, elves)
	if !f.onBattlefield || f.name != "Everquill Phoenix" {
		t.Fatalf("over: the Elves' permanent is now Everquill Phoenix, got %+v", f)
	}
	if f.power != 5 || f.toughness != 5 {
		t.Errorf("over: a 4/4 with the Elves' +1/+1 counter is 5/5, got %d/%d", f.power, f.toughness)
	}
	if !f.flying || !f.tapsForMana {
		t.Errorf("over: flying from the top and the Elves' mana ability from under it, got %+v", f)
	}
	if len(f.components) != 2 || f.components[0] != phoenix || f.components[1] != elves {
		t.Errorf("components topmost first: %v %v", f.components, f.componentNames)
	}
	if findBattlefieldCardByID(g, phoenix) != nil || zoneHolds(g.Stack, phoenix) {
		t.Errorf("the Phoenix card is part of the permanent, not an object of its own")
	}
	if n := featherCount(g, me.ID); n != 1 {
		t.Errorf("whenever this creature mutates: %d Feathers, want 1", n)
	}
}

// Under: the creature stays Llanowar Elves, a 1/1, and has flying and the
// mutate trigger of the Phoenix underneath it.
func TestEverquillPhoenixMutatesUnderAndLendsItsAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := llanowarOn(g, me.ID)
	phoenix := phoenixInHand(t, g, me.ID)
	if err := castMutating(g, me.ID, phoenix, elves, false); err != nil {
		t.Fatalf("cast mutating: %v", err)
	}
	resolveMutate(t, g, me.ID, true)

	f := readMerged(t, g, elves)
	if f.name != "Llanowar Elves" || f.power != 1 || f.toughness != 1 {
		t.Fatalf("under: the creature keeps the top card's characteristics, got %+v", f)
	}
	if !f.flying || !f.tapsForMana {
		t.Errorf("under: the Elves have the Phoenix's flying and their own mana ability, got %+v", f)
	}
	if len(f.components) != 2 || f.components[0] != elves || f.components[1] != phoenix {
		t.Errorf("components topmost first: %v", f.componentNames)
	}
	if n := featherCount(g, me.ID); n != 1 {
		t.Errorf("the trigger of the card underneath fires: %d Feathers, want 1", n)
	}
}

// A second Phoenix mutating onto a mutated Phoenix: both components and
// the new one print the trigger, and the permanent has all of them, so
// "whenever this creature mutates" triggers twice — the host's own
// included.
func TestASecondMutateTriggersEveryComponentsAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := llanowarOn(g, me.ID)
	first := phoenixInHand(t, g, me.ID)
	if err := castMutating(g, me.ID, first, elves, false); err != nil {
		t.Fatalf("cast first: %v", err)
	}
	resolveMutate(t, g, me.ID, false)
	second := phoenixInHand(t, g, me.ID)
	if err := castMutating(g, me.ID, second, elves, false); err != nil {
		t.Fatalf("cast second: %v", err)
	}
	resolveMutate(t, g, me.ID, true)

	f := readMerged(t, g, elves)
	if len(f.components) != 3 || f.components[0] != first || f.components[2] != second {
		t.Fatalf("three components, the new one on the bottom: %v", f.componentNames)
	}
	if n := featherCount(g, me.ID); n != 3 {
		t.Errorf("one Feather from the first mutate and two from the second: %d, want 3", n)
	}
}

// CR 702.140b: the target left in response, so the spell resolves as an
// ordinary creature spell: the Phoenix enters as its own 4/4, nobody is
// asked over or under, and nothing mutated.
func TestEverquillPhoenixWithAnIllegalTargetResolvesAsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := llanowarOn(g, me.ID)
	phoenix := phoenixInHand(t, g, me.ID)
	if err := castMutating(g, me.ID, phoenix, elves, false); err != nil {
		t.Fatalf("cast mutating: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(elves); err != nil {
			t.Fatalf("destroy target: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if c := latestOptionPickFor(g, me.ID); c != nil {
		t.Fatalf("no over/under question without a legal target: %+v", c)
	}
	f := readMerged(t, g, phoenix)
	if !f.onBattlefield || f.name != "Everquill Phoenix" || f.power != 4 || len(f.components) != 0 {
		t.Fatalf("the Phoenix enters as a creature of its own: %+v", f)
	}
	if n := featherCount(g, me.ID); n != 0 {
		t.Errorf("nothing mutated: %d Feathers", n)
	}
}

// CR 702.140a: a non-Human creature with the same owner as the spell.
func TestMutateTargetMustBeANonHumanCreatureWithTheSpellsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	human := b12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 1, 1)
	theirs := llanowarOn(g, opp.ID)
	stolen := llanowarOn(g, opp.ID)
	g.WithWriteLock(func() {
		if c := findBattlefieldCardByID(g, stolen); c != nil {
			c.Controller = me.ID
		}
	})
	mine := llanowarOn(g, me.ID)
	phoenix := phoenixInHand(t, g, me.ID)
	for name, host := range map[string]uuid.UUID{"a Human": human, "an opponent's creature": theirs, "a creature you control but don't own": stolen} {
		err := castMutating(g, me.ID, phoenix, host, true)
		if !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("%s is not a legal mutate target: err %v", name, err)
		}
	}
	if err := castMutating(g, me.ID, phoenix, mine, true); !errors.Is(err, game.ErrInsufficientMana) {
		t.Errorf("mutate costs {3}{R}: with an empty pool, err %v", err)
	}
	for _, color := range []string{"R", "R", "R", "R"} {
		me.ManaPool.AddMana(game.ManaToken{Color: color})
	}
	if err := castMutating(g, me.ID, phoenix, mine, true); err != nil {
		t.Fatalf("a non-Human creature you own is a legal target: %v", err)
	}
}

// mutatedElves builds a Phoenix-over-Elves permanent and returns both
// IDs.
func mutatedElves(t *testing.T, g *game.Game, under bool) (elves, phoenix uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	elves = llanowarOn(g, me.ID)
	phoenix = phoenixInHand(t, g, me.ID)
	if err := castMutating(g, me.ID, phoenix, elves, false); err != nil {
		t.Fatalf("cast mutating: %v", err)
	}
	resolveMutate(t, g, me.ID, under)
	return elves, phoenix
}

// CR 730.3: one creature dies, and both cards go to the graveyard, each
// as itself (the Elves card is Llanowar Elves again).
func TestAMutatedCreatureDiesAsBothCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves, phoenix := mutatedElves(t, g, false)
	deathsBefore := countEvents(g, game.EventLTB)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(elves); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if !zoneHolds(me.Graveyard, elves) || !zoneHolds(me.Graveyard, phoenix) {
		t.Fatalf("both cards are in the graveyard: %+v", me.Graveyard.Cards)
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == elves && c.Name != "Llanowar Elves" {
			t.Errorf("the Elves card is itself again in the graveyard, got %q", c.Name)
		}
		if c.InstanceID == phoenix && c.Name != "Everquill Phoenix" {
			t.Errorf("the Phoenix card is itself in the graveyard, got %q", c.Name)
		}
	}
	if n := countEvents(g, game.EventLTB) - deathsBefore; n != 1 {
		t.Errorf("one permanent left the battlefield: %d leave events", n)
	}
}

// Bounce and exile move every card too (CR 730.3).
func TestAMutatedCreatureBouncesAndIsExiledAsBothCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves, phoenix := mutatedElves(t, g, true)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(elves); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	if !zoneHolds(me.Hand, elves) || !zoneHolds(me.Hand, phoenix) {
		t.Fatalf("both cards return to hand: %+v", me.Hand.Cards)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	elves2, phoenix2 := mutatedElves(t, g2, false)
	g2.WithWriteLock(func() {
		if err := g2.ExileCardForEffect(elves2); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	if !zoneHolds(g2.Exile, elves2) || !zoneHolds(g2.Exile, phoenix2) {
		t.Fatalf("both cards are exiled: %+v", g2.Exile.Cards)
	}
	if findBattlefieldCardByID(g2, elves2) != nil || zoneHolds(me2.Hand, phoenix2) {
		t.Errorf("nothing is left behind")
	}
}

// CR 707.2, 730.2a: a copy of a mutated permanent copies the top card's
// characteristics and every component's abilities, and is not itself
// merged.
func TestACopyOfAMutatedCreatureHasEveryAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves, _ := mutatedElves(t, g, false)
	var copyID uuid.UUID
	g.WithWriteLock(func() {
		before := map[uuid.UUID]bool{}
		for _, c := range g.Battlefield.Cards {
			before[c.InstanceID] = true
		}
		item := &game.StackItem{Controller: me.ID, SourceCardID: elves}
		if err := (CreateTokenCopy{Controller: me.ID, Copy: elves, N: 1}).Apply(ctxFor(g, item)); err != nil {
			t.Fatalf("token copy: %v", err)
		}
		for _, c := range g.Battlefield.Cards {
			if !before[c.InstanceID] && c.Name == "Everquill Phoenix" {
				copyID = c.InstanceID
			}
		}
	})
	if copyID == uuid.Nil {
		t.Fatalf("no token copy of the mutated creature")
	}
	f := readMerged(t, g, copyID)
	if f.power != 4 || !f.flying || !f.tapsForMana || len(f.components) != 0 {
		t.Errorf("the copy is a 4/4 Phoenix with flying and the Elves' mana ability, unmerged: %+v", f)
	}
}

// The over/under question is a restore point, and the answer given
// after a restore still merges; a restored mutated permanent keeps its
// components and still leaves as both cards.
func TestMutateSurvivesASaveAndRestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := llanowarOn(g, me.ID)
	phoenix := phoenixInHand(t, g, me.ID)
	if err := castMutating(g, me.ID, phoenix, elves, false); err != nil {
		t.Fatalf("cast mutating: %v", err)
	}
	passPriorityAroundTable(t, g)
	g = restoreAtKeyedPrompt(t, g, "mutate/over-under")
	me = g.Seats[g.Turn.ActiveSeat]
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	f := readMerged(t, g, elves)
	if f.name != "Llanowar Elves" || !f.flying || len(f.components) != 2 {
		t.Fatalf("the restored answer put the Phoenix under: %+v", f)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	f = readMerged(t, g, elves)
	if f.name != "Llanowar Elves" || !f.flying || !f.tapsForMana || len(f.components) != 2 || f.components[1] != phoenix {
		t.Fatalf("the restored permanent is still merged: %+v", f)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(elves); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !zoneHolds(me.Graveyard, elves) || !zoneHolds(me.Graveyard, phoenix) {
		t.Fatalf("the restored permanent dies as both cards: %+v", me.Graveyard.Cards)
	}
}

// The Feather: "{1}, Sacrifice this token: Return target Phoenix card
// from your graveyard to the battlefield tapped."
func TestFeatherReturnsAPhoenixTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves, phoenix := mutatedElves(t, g, false)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(elves); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	for latestOptionPickFor(g, me.ID) != nil {
		answerOptionPick(t, g, me.ID, 0)
	}
	var feather uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Feather" {
				feather = c.InstanceID
			}
		}
	})
	if feather == uuid.Nil {
		t.Fatalf("no Feather")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, feather, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: elves}},
	}); err == nil {
		t.Fatalf("Llanowar Elves is not a Phoenix card")
	}
	if err := g.ActivateCatalogAbility(me.ID, feather, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: phoenix}},
	}); err != nil {
		t.Fatalf("activate Feather: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardByID(g, phoenix)
	if c == nil || !c.Tapped || c.Name != "Everquill Phoenix" {
		t.Fatalf("the Phoenix returns tapped: %+v", c)
	}
	if findBattlefieldCardByID(g, feather) != nil {
		t.Errorf("the Feather was sacrificed")
	}
}
