package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsd "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/d"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Earthshape --------------------------------------------------------

// TestEarthshapeShieldsCreaturesUpToTheLandsPower is the card on a
// quiet board: a 3/3 land, so the 3-power creature is shielded and the
// 4-power one is not, and the caster gains hexproof.
func TestEarthshapeShieldsCreaturesUpToTheLandsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushEarthbendLand(g, me.ID, "Plains", "Basic Land — Plains")
	three := effectstest.B39Creature(g, me.ID, "Three", "Creature — Bear", 3, 3)
	four := effectstest.B39Creature(g, me.ID, "Four", "Creature — Bear", 4, 4)

	effectstest.CastCatalogSpell(t, g, "Earthshape", "Instant", effectstest.EarthshapeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	effectstest.PassPriorityAroundTable(t, g)

	if p, _, _, _ := effectstest.EarthbentBody(t, g, land); p != 3 {
		t.Fatalf("that land's power = %d, want 3", p)
	}
	for _, id := range []uuid.UUID{land, three} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "hexproof") || !effectstest.HasKeywordOnBattlefield(t, g, id, "indestructible") {
			t.Errorf("%s (power ≤ 3) is not hexproof and indestructible", effectstest.FindBattlefieldCardByID(g, id).Name)
		}
	}
	if effectstest.HasKeywordOnBattlefield(t, g, four, "hexproof") {
		t.Error("the 4-power creature gained hexproof; the land is a 3/3")
	}
	if !effectstest.PlayerHasHexproof(g, me) {
		t.Error("you gain hexproof until end of turn")
	}
}

// TestEarthshapeReadsTheLandAfterAPausedPlacement is #1282 on the
// card ADR 0081 deferred for it, on the board #1289 was filed for.
// Doubling Season and Hardened Scales both apply to the earthbend
// counters, so the placement waits on the CR 616 prompt; read on the
// next line, "that land's power" was the bare 0/0 body and the 5-power
// creature was left unshielded.
//
// The land has no counters, which is the real case: it is a 0/0
// creature for the whole of the pause. Before #1289 the state-based
// sweep ran while the prompt was open, killed it to CR 704.5f, and the
// earthbend return brought it back as a new object with nothing on it.
// CR 704.3 holds the sweep until the resolution has finished.
func TestEarthshapeReadsTheLandAfterAPausedPlacement(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	land := effectstest.PushEarthbendLand(g, me.ID, "Plains", "Basic Land — Plains")
	five := effectstest.B39Creature(g, me.ID, "Five", "Creature — Beast", 5, 5)

	effectstest.CastCatalogSpell(t, g, "Earthshape", "Instant", effectstest.EarthshapeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.FindBattlefieldCardByID(g, land) == nil {
		t.Fatal("the 0/0 land died to CR 704.5f while its counters were still owed (#1289)")
	}
	if !g.ResolutionPaused() {
		t.Fatal("Earthshape's resolution is not paused on its CR 616 prompt")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, five, "hexproof") {
		t.Fatal("the sweep ran with the earthbend counters still owed to the CR 616 prompt")
	}
	if effectstest.PlayerHasHexproof(g, me) {
		t.Fatal("the rest of the sentence ran before the counters landed")
	}
	// Doubling Season first: 3 → 6 → 7, onto the same land.
	effectstest.AnswerOrderBySource(t, g, ds, hs)

	if p, _, _, _ := effectstest.EarthbentBody(t, g, land); p != 7 {
		t.Fatalf("that land's power = %d, want 7", p)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, five, "hexproof") || !effectstest.HasKeywordOnBattlefield(t, g, five, "indestructible") {
		t.Error("the 5-power creature is not shielded by a 7-power land")
	}
	if !effectstest.PlayerHasHexproof(g, me) {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	army := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie Army", TypeLine: "Token Creature — Zombie Army",
		Power: 0, Toughness: 0, PrintedPTKnown: true,
		Counters: map[string]int{game.CounterPlusOne: 1},
		Owner:    me.ID, Controller: me.ID,
	})
	wall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 6, PrintedPTKnown: true, Owner: opp.ID, Controller: opp.ID,
	})

	effectstest.CastCatalogSpell(t, g, "Widespread Brutality", "Sorcery", effectstest.WidespreadBrutalityOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if w := effectstest.FindBattlefieldCardByID(g, wall); w == nil || w.DamageMarked != 0 {
		t.Fatal("the Army dealt damage with its counters still owed to the CR 616 prompt")
	}

	// Doubling Season first: amass 2 → 4 → 5, onto the one it had.
	effectstest.AnswerOrderBySource(t, g, ds, hs)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountersOn(g, army, game.CounterPlusOne); got != 6 {
		t.Fatalf("Army counters = %d, want 6", got)
	}
	if effectstest.FindBattlefieldCardByID(g, wall) != nil {
		t.Error("the 0/6 wall survived a 6-power Army — the sweep read the pre-amass power")
	}
}

// TestWidespreadBrutalityWithNoArmyKeepsTheFreshArmyThroughThePause is
// #1289's amass half on the real card. With no Army the amass creates
// 0/0 tokens (two: Doubling Season doubles the creation too), the
// choose-an-Army prompt pauses, and then the counter placement pauses
// on the CR 616 prompt. Before #1289 the sweep ran during the first
// pause and killed both 0/0s before either could be chosen, so CR
// 701.47c's "the Army you amassed" was nobody and nothing was dealt.
func TestWidespreadBrutalityWithNoArmyKeepsTheFreshArmyThroughThePause(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	wall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 5, PrintedPTKnown: true, Owner: opp.ID, Controller: opp.ID,
	})

	effectstest.CastCatalogSpell(t, g, "Widespread Brutality", "Sorcery", effectstest.WidespreadBrutalityOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)

	var armies []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype(game.ArmySubtype) {
			armies = append(armies, c.InstanceID)
		}
	}
	if len(armies) != 2 {
		t.Fatalf("%d Armies during the pause, want the two fresh 0/0s (#1289)", len(armies))
	}
	pick := g.PendingChoices[0]
	if pick.Kind != game.PendingChoiceChooseCards {
		t.Fatalf("first prompt = %q, want the choose-an-Army prompt", pick.Kind)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{armies[0]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if effectstest.FindBattlefieldCardByID(g, armies[0]) == nil {
		t.Fatal("the chosen 0/0 Army died while its counters were still owed (#1289)")
	}

	// Doubling Season first: amass 2 → 4 → 5.
	effectstest.AnswerOrderBySource(t, g, ds, hs)

	if got := effectstest.CountersOn(g, armies[0], game.CounterPlusOne); got != 5 {
		t.Fatalf("Army counters = %d, want 5", got)
	}
	if effectstest.FindBattlefieldCardByID(g, armies[1]) != nil {
		t.Error("the Army that got no counters is still a 0/0 on the battlefield after the resolution")
	}
	if effectstest.FindBattlefieldCardByID(g, wall) != nil {
		t.Error("the 0/5 wall survived a 5-power Army: the Army was read before its counters landed")
	}
}

// --- Dawn of a New Age -----------------------------------------------

// TestDawnOfANewAgeRemovalNeverPausesWithDoublingSeasonInPlay is
// #1291's real-card exit criterion, superseding the test-only removal
// probes this used before the fix. Two Doubling Seasons — the only
// real catalog counter-doubler that isn't gated to "+1/+1 on a
// creature" — are exactly the board that would have queued a CR 616
// ordering prompt on the hope-counter removal before #1291: both
// AppliesTo predicates lacked the ev.CounterDelta > 0 check, so a
// removal (a negative delta) looked like a placement to double. With
// the check in place, neither replacement is even ACTIVE for a
// removal, so the counter comes off in one uninterrupted step and the
// end step's later clauses ("if you do, draw"; "then if no hope
// counters, cash in") read the right count without ever seeing a
// pause.
func TestDawnOfANewAgeRemovalNeverPausesWithDoublingSeasonInPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dawn := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dawn of a New Age", TypeLine: "Enchantment",
		OracleID: effectstest.B39DawnOfANewAgeOracle, Counters: map[string]int{"hope": 2},
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season II", me.ID)
	hand, life := me.Hand.Size(), me.Life

	g.WithWriteLock(func() {
		item := &game.StackItem{SourceCardID: dawn, Controller: me.ID}
		if err := cardsd.B39DawnOfANewAgeEndStep(g, item); err != nil {
			t.Fatalf("end step: %v", err)
		}
	})

	if len(g.PendingChoices) != 0 {
		t.Fatalf(`%d pending choice(s) after the removal with two Doubling Seasons in play.

A removal is not a placement (CR 122.6, CR 614.1) — Doubling Season's
"if AN EFFECT WOULD PUT" never applies to it, so even with two of them
on the board there is nothing to order and no prompt to pause on.`, len(g.PendingChoices))
	}
	if got := effectstest.CountersOn(g, dawn, cardsd.B39HopeCounter); got != 1 {
		t.Errorf("hope counters = %d, want 1 (removed exactly one — Doubling Season does not double a removal)", got)
	}
	if !g.Battlefield.Contains(dawn) {
		t.Fatal("the enchantment was sacrificed with a hope counter still on it")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one draw", hand, me.Hand.Size())
	}
	if me.Life != life {
		t.Errorf("life %d → %d, want unchanged — a hope counter is still on the enchantment, so it has not cashed in", life, me.Life)
	}
}

// --- Gemstone Mine -----------------------------------------------------

// TestGemstoneMineRemovalNeverPausesWithDoublingSeasonInPlay is
// #1291's real-card counterpart for the Mine's rider. Two Doubling
// Seasons in play would have paused the mining-counter removal on a
// CR 616 ordering prompt before the fix; with it, neither replacement
// is active for a removal, so the rider's "if none left, sacrifice"
// check runs against the true post-removal count in one step.
func TestGemstoneMineRemovalNeverPausesWithDoublingSeasonInPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B31Push(g, me.ID, "Gemstone Mine", "Land", effectstest.B31GemstoneMineOracle, "", 0, 0)
	effectstest.FindBattlefieldCardByID(g, mine).Counters = map[string]int{"mining": 2}
	effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season II", me.ID)

	g.WithWriteLock(func() {
		if err := effects.B31RemoveMiningCounterOrSacrifice(g, me.ID, mine); err != nil {
			t.Fatalf("rider: %v", err)
		}
	})

	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d pending choice(s) after the removal with two Doubling Seasons in play — a removal must never pause", len(g.PendingChoices))
	}
	if got := effectstest.CountersOn(g, mine, "mining"); got != 1 {
		t.Errorf("mining counters = %d, want 1 (removed exactly one — not doubled)", got)
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the Mine was sacrificed with a mining counter still on it")
	}
}
