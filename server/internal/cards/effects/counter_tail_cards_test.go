package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_tail_cards_test.go — #1282 on real cards. Every card here
// sequences a clause behind a counter placement that READS the
// counters, and every pausing test builds a real pausing board: two
// different counter replacements on the same placement, which queues
// the CR 616 ordering prompt and returns with nothing placed.

const earthshapeOracle = "300f17a2-1594-40cb-9187-952898a83622"

// answerOrderBySource answers the one open CR 616 prompt in the order
// of the permanents that contributed each replacement.
func answerOrderBySource(t *testing.T, g *game.Game, sources ...uuid.UUID) {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the one CR 616 prompt", len(g.PendingChoices))
	}
	p := g.PendingChoices[0]
	if p.Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %q, want %q", p.Kind, game.PendingChoiceReplacementOrder)
	}
	order := make([]game.ReplacementEffectID, 0, len(sources))
	for _, src := range sources {
		order = append(order, replacementIDForSource(t, g, p.ReplacementEffectIDs, src))
	}
	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, order); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

// answerOrderByLabel is answerOrderBySource for sourceless probes.
func answerOrderByLabel(t *testing.T, g *game.Game, labels ...string) {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the one CR 616 prompt", len(g.PendingChoices))
	}
	p := g.PendingChoices[0]
	byLabel := map[string]game.ReplacementEffectID{}
	for _, id := range p.ReplacementEffectIDs {
		label, _ := g.ReplacementOptionMetaForEffect(id)
		byLabel[label] = id
	}
	order := make([]game.ReplacementEffectID, 0, len(labels))
	for _, l := range labels {
		id, ok := byLabel[l]
		if !ok {
			t.Fatalf("no replacement labelled %q in the prompt", l)
		}
		order = append(order, id)
	}
	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, order); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

// removalProbe is a sourceless replacement on REMOVALS of one counter
// kind. Two of them is a pausing board for a card that removes a
// counter and then checks what is left.
func removalProbe(label, name string, rewrite func(int) int) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventCounterPlaced},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
			return ev.Kind == game.RepEventCounter && ev.CounterName == name && ev.CounterDelta < 0
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.CounterDelta = rewrite(ev.CounterDelta)
			return nil
		},
		Label: label,
	}
}

func registerPausingRemovalBoard(g *game.Game, name string) {
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(removalProbe("remove twice as many", name, func(n int) int { return n * 2 }))
		g.RegisterReplacementForTest(removalProbe("remove one more", name, func(n int) int { return n - 1 }))
	})
}

func hasKeywordOnBattlefield(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return game.HasKeyword(c, kw)
}

func playerHasHexproof(g *game.Game, p *game.Player) bool {
	var has bool
	g.WithWriteLock(func() { has = g.PlayerHasKeywordLocked(p, game.KeywordHexproof) })
	return has
}

// --- Earthshape --------------------------------------------------------

// TestEarthshapeShieldsCreaturesUpToTheLandsPower is the card on a
// quiet board: a 3/3 land, so the 3-power creature is shielded and the
// 4-power one is not, and the caster gains hexproof.
func TestEarthshapeShieldsCreaturesUpToTheLandsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushEarthbendLand(g, me.ID, "Plains", "Basic Land — Plains")
	three := b39Creature(g, me.ID, "Three", "Creature — Bear", 3, 3)
	four := b39Creature(g, me.ID, "Four", "Creature — Bear", 4, 4)

	castCatalogSpell(t, g, "Earthshape", "Instant", earthshapeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)

	if p, _, _, _ := earthbentBody(t, g, land); p != 3 {
		t.Fatalf("that land's power = %d, want 3", p)
	}
	for _, id := range []uuid.UUID{land, three} {
		if !hasKeywordOnBattlefield(t, g, id, "hexproof") || !hasKeywordOnBattlefield(t, g, id, "indestructible") {
			t.Errorf("%s (power ≤ 3) is not hexproof and indestructible", findBattlefieldCardByID(g, id).Name)
		}
	}
	if hasKeywordOnBattlefield(t, g, four, "hexproof") {
		t.Error("the 4-power creature gained hexproof; the land is a 3/3")
	}
	if !playerHasHexproof(g, me) {
		t.Error("you gain hexproof until end of turn")
	}
}

// TestEarthshapeReadsTheLandAfterAPausedPlacement is #1282 on the
// card ADR 0081 deferred for it. Doubling Season and Hardened Scales
// both apply to the earthbend counters, so the placement waits on the
// CR 616 prompt; read on the next line, "that land's power" was the
// 0/0 body with only the counter it already had, and the 5-power creature
// was left unshielded.
//
// The land starts with one +1/+1 counter (an earlier earthbend's). On a
// bare land the pause has a second, separate problem: the resolution
// ends while the prompt is open, the state-based sweep runs, and the
// fresh 0/0 dies before its counters land — #1289, not this one.
func TestEarthshapeReadsTheLandAfterAPausedPlacement(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	land := pushEarthbendLand(g, me.ID, "Plains", "Basic Land — Plains")
	findBattlefieldCardByID(g, land).Counters = map[string]int{game.CounterPlusOne: 1}
	five := b39Creature(g, me.ID, "Five", "Creature — Beast", 5, 5)

	castCatalogSpell(t, g, "Earthshape", "Instant", earthshapeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)

	if hasKeywordOnBattlefield(t, g, five, "hexproof") {
		t.Fatal("the sweep ran with the earthbend counters still owed to the CR 616 prompt")
	}
	if playerHasHexproof(g, me) {
		t.Fatal("the rest of the sentence ran before the counters landed")
	}
	// Doubling Season first: 3 → 6 → 7, onto the one it had.
	answerOrderBySource(t, g, ds, hs)

	if p, _, _, _ := earthbentBody(t, g, land); p != 8 {
		t.Fatalf("that land's power = %d, want 8", p)
	}
	if !hasKeywordOnBattlefield(t, g, five, "hexproof") || !hasKeywordOnBattlefield(t, g, five, "indestructible") {
		t.Error("the 5-power creature is not shielded by an 8-power land")
	}
	if !playerHasHexproof(g, me) {
		t.Error("you gain hexproof once the sentence runs")
	}
}

// --- Widespread Brutality ------------------------------------------------

// TestWidespreadBrutalityDealsThePowerTheCountersLandedAt: CR 701.47c's
// "the Army you amassed deals damage equal to its power" on a Doubling
// Season + Hardened Scales board. The amass onto a 1/1 Army pauses on
// the counter-order prompt; before #1282 the sweep ran at once and
// dealt 1, so the 0/6 wall lived.
func TestWidespreadBrutalityDealsThePowerTheCountersLandedAt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	army := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie Army", TypeLine: "Token Creature — Zombie Army",
		Power: 0, Toughness: 0, PrintedPTKnown: true,
		Counters: map[string]int{game.CounterPlusOne: 1},
		Owner:    me.ID, Controller: me.ID,
	})
	wall := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 6, PrintedPTKnown: true, Owner: opp.ID, Controller: opp.ID,
	})

	castCatalogSpell(t, g, "Widespread Brutality", "Sorcery", widespreadBrutalityOracl, nil)
	passPriorityAroundTable(t, g)
	if w := findBattlefieldCardByID(g, wall); w == nil || w.DamageMarked != 0 {
		t.Fatal("the Army dealt damage with its counters still owed to the CR 616 prompt")
	}

	// Doubling Season first: amass 2 → 4 → 5, onto the one it had.
	answerOrderBySource(t, g, ds, hs)
	passPriorityAroundTable(t, g)

	if got := countersOn(g, army, game.CounterPlusOne); got != 6 {
		t.Fatalf("Army counters = %d, want 6", got)
	}
	if findBattlefieldCardByID(g, wall) != nil {
		t.Error("the 0/6 wall survived a 6-power Army — the sweep read the pre-amass power")
	}
}

// --- Dawn of a New Age -----------------------------------------------

// TestDawnOfANewAgeCashesInAfterAPausedRemoval: the end-step removal
// pauses on a CR 616 prompt, so the draw and the "if there are no hope
// counters" check wait with it. Read on the next line, the check saw
// the counters still there and the enchantment was never cashed in.
func TestDawnOfANewAgeCashesInAfterAPausedRemoval(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dawn := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dawn of a New Age", TypeLine: "Enchantment",
		OracleID: b39DawnOfANewAgeOracle, Counters: map[string]int{"hope": 2},
		Owner: me.ID, Controller: me.ID,
	})
	registerPausingRemovalBoard(g, "hope")
	hand, life := me.Hand.Size(), me.Life

	g.WithWriteLock(func() {
		item := &game.StackItem{SourceCardID: dawn, Controller: me.ID}
		if err := b39DawnOfANewAgeEndStep(g, item); err != nil {
			t.Fatalf("end step: %v", err)
		}
	})
	if me.Hand.Size() != hand || !g.Battlefield.Contains(dawn) {
		t.Fatal("the rest of the end step ran with the removal still owed to the CR 616 prompt")
	}
	// ×2 then one more: −1 → −2 → −3, every hope counter gone.
	answerOrderByLabel(t, g, "remove twice as many", "remove one more")

	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one draw", hand, me.Hand.Size())
	}
	if g.Battlefield.Contains(dawn) {
		t.Error("no hope counters left and the enchantment was not sacrificed")
	}
	if me.Life != life+4 {
		t.Errorf("life %d → %d, want +4", life, me.Life)
	}
}

// --- Gemstone Mine -----------------------------------------------------

// TestGemstoneMineIsSacrificedAfterAPausedRemoval: the rider removes a
// mining counter and sacrifices the Mine when none is left. With the
// removal paused, the check has to wait for it.
func TestGemstoneMineIsSacrificedAfterAPausedRemoval(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b31Push(g, me.ID, "Gemstone Mine", "Land", b31GemstoneMineOracle, "", 0, 0)
	findBattlefieldCardByID(g, mine).Counters = map[string]int{"mining": 2}
	registerPausingRemovalBoard(g, "mining")

	g.WithWriteLock(func() {
		if err := b31RemoveMiningCounterOrSacrifice(g, me.ID, mine); err != nil {
			t.Fatalf("rider: %v", err)
		}
	})
	if !g.Battlefield.Contains(mine) {
		t.Fatal("the Mine was sacrificed before the removal it depends on")
	}
	answerOrderByLabel(t, g, "remove twice as many", "remove one more")
	if g.Battlefield.Contains(mine) {
		t.Error("no mining counters left and the Mine was not sacrificed — the check ran before the removal landed")
	}
}
