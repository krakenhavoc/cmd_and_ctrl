package testsr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_order_batch6_test.go — #4155 (CR 608.2c): the sixth batch of
// hand-written bodies written as steps. It adds a real card for each shape
// batch 6 converted, each with the prompt open:
//
//   - a modal spell's ModeSteps on a CR 616 counter order (Konstrari Charm);
//   - a modal spell's ModeSteps on a dredge offer (Vigorbloom Charm), one that
//     files the number it drew for the life loss (Rise of the Deathbringer);
//   - a spell's ResolveSteps on a dredge offer, three sentences (Live Fast);
//   - a spell's ResolveSteps on Khârn's question, live and across a restore
//     (Narset's Rebuke);
//   - a loyalty ability's Do that files that it had a target, on Khârn's
//     question, live and across a restore (Sorin, Vampire Lord's −2);
//   - a vote's continuation steps on a CR 616 counter order (Orchard
//     Elemental).
//
// A dredge offer and a CR 616 order resume through closures, so a table that
// holds one is not a restore point; an undo (a clone) stands in for the restore
// there, as in resolution_order_batch1_test.go. Khârn's question is data.

const (
	b6KonstrariCharmOracle   = "40961b23-b153-434a-aeb9-b9c2ff613dcb"
	b6VigorbloomCharmOracle  = "826f17d1-3806-4579-8b1b-2082b7b03d69"
	b6RiseDeathbringerOracle = "57ea9f63-8267-4303-a3dd-9ded220294dc"
	b6NarsetsRebukeOracle    = "247075c5-62f8-41a1-91b3-562ff0aabb30"
	b6SorinVampireLordOracle = "753dcc18-3697-49c5-80c6-2d8dc8ee8c9a"
)

// b6DeclineDredge declines every dredge offer `p` has open on `table`.
func b6DeclineDredge(t *testing.T, table *game.Game, p *game.Player) {
	t.Helper()
	for effectstest.OptionalOffer(table, p) != nil {
		effectstest.AnswerOptional(t, table, p, false)
	}
	table.SettleResolution()
}

// Konstrari Charm's second mode — "Put two +1/+1 counters on target creature.
// It gains trample until end of turn." With Doubling Season and Hardened Scales
// out the counters wait on the CR 616 order, and the trample waits behind it.
func TestKonstrariCharmTrampleWaitsForTheCounterOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	effectstest.CastModal(t, g, "Konstrari Charm", "Instant", b6KonstrariCharmOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Fatal("the creature has trample with the order open: the keyword was granted before the counters")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		effectstest.AnswerOrderBySource(t, table, ds, hs)
		if got := effectstest.B12Counter(t, table, bear, "+1/+1"); got != 5 {
			t.Fatalf("+1/+1 counters %d, want 5 ((2×2)+1, Doubling Season first)", got)
		}
		if !effectstest.HasKeywordOnBattlefield(t, table, bear, "trample") {
			t.Fatal("the creature did not gain trample after the order")
		}
	}
}

// Vigorbloom Charm's second mode — "You draw a card and gain 3 life." The life
// gain waits for the dredge offer the draw makes.
func TestVigorbloomCharmLifeGainWaitsForDredge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AwbPadLibrary(me)
	effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	life := me.Life
	effectstest.CastModal(t, g, "Vigorbloom Charm", "Instant", b6VigorbloomCharmOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the gain ran before the answer", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		b6DeclineDredge(t, table, p)
		if p.Life != life+3 {
			t.Fatalf("life %d after the answer, want %d", p.Life, life+3)
		}
	}
}

// Rise of the Deathbringer's first mode — "Draw cards equal to the greatest
// power among creatures you control. You lose life equal to the number of cards
// drawn this way." The loss waits for the dredge offers and reads the number
// the first step filed.
func TestRiseOfTheDeathbringerLossWaitsForDredgeAndReadsTheNumber(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AwbPadLibrary(me)
	effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	effectstest.B12Creature(g, me.ID, "Brute", "Creature — Ogre", 3, 3)
	life := me.Life
	effectstest.CastModal(t, g, "Rise of the Deathbringer", "Sorcery", b6RiseDeathbringerOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the loss ran before the answer", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		b6DeclineDredge(t, table, p)
		if p.Life != life-3 {
			t.Fatalf("life %d after the draws, want %d (the greatest power is 3)", p.Life, life-3)
		}
	}
}

// Live Fast — "You draw two cards, lose 2 life, and get {E}{E}." The life loss
// and the energy wait for the dredge offers the draws make.
func TestLiveFastLossAndEnergyWaitForDredge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AwbPadLibrary(me)
	effectstest.GyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Live Fast", "Sorcery", effectstest.EaLiveFast, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OptionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life || effectstest.EnergyOf(me) != 0 {
		t.Fatalf("life %d energy %d with dredge open, want %d and 0: they ran before the answer", me.Life, effectstest.EnergyOf(me), life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		b6DeclineDredge(t, table, p)
		if p.Life != life-2 || effectstest.EnergyOf(p) != 2 {
			t.Fatalf("life %d energy %d after the draws, want %d and 2", p.Life, effectstest.EnergyOf(p), life-2)
		}
	}
}

// Narset's Rebuke — "Narset's Rebuke deals 5 damage to target creature. Add
// {U}{R}{W}. Exile that creature if it would die this turn." Aimed at Khârn the
// Betrayer, whose prevention asks its controller for an opponent, the mana
// waits for the answer, live and on a table restored while the question is open.
func TestNarsetsRebukeManaWaitsForKharnsQuestion(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		kharn := effectstest.KharnPush(g, me.ID, me.ID)
		effectstest.CastCatalogSpell(t, g, "Narset's Rebuke", "Instant", b6NarsetsRebukeOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: kharn}})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatalf("restored=%v: Khârn's question is not open", restored)
		}
		if n := len(me.ManaPool); n != 0 {
			t.Fatalf("restored=%v: %d mana in the pool with the question open, want 0: it was added before the answer", restored, n)
		}
		if restored {
			g = effectstest.RestoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		effectstest.KharnAnswer(t, g, me.ID, third.ID)
		g.SettleResolution()
		if n := len(me.ManaPool); n != 3 {
			t.Fatalf("restored=%v: %d mana in the pool after the answer, want 3", restored, n)
		}
	}
}

// Sorin, Vampire Lord's −2 — "Sorin deals 4 damage to any target. You gain 4
// life." Aimed at Khârn the Betrayer, the life gain waits for the answer, live
// and on a table restored while the question is open, and still happens: the
// first step filed that the ability had a target before the damage.
func TestSorinVampireLordGainWaitsForKharnsQuestion(t *testing.T) {
	for _, restored := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.ToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		kharn := effectstest.KharnPush(g, me.ID, me.ID)
		sorin := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Sorin, Vampire Lord", "Legendary Planeswalker — Sorin", b6SorinVampireLordOracle, 4)
		life := me.Life
		if err := g.ActivateCatalogAbility(me.ID, sorin, 1, game.ActivateAbilityParams{
			Targets: effectstest.CardRefs(kharn),
		}); err != nil {
			t.Fatalf("restored=%v: activate: %v", restored, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
			t.Fatalf("restored=%v: Khârn's question is not open", restored)
		}
		if me.Life != life {
			t.Fatalf("restored=%v: life %d with the question open, want %d: the gain ran before the answer", restored, me.Life, life)
		}
		if restored {
			g = effectstest.RestoreRoundTrip(t, g, true)
			me = g.Seats[g.Turn.ActiveSeat]
		}
		effectstest.KharnAnswer(t, g, me.ID, third.ID)
		if me.Life != life+4 {
			t.Fatalf("restored=%v: life %d after the answer, want %d", restored, me.Life, life+4)
		}
	}
}

// Sorin, Vampire Lord's −2 on a creature the damage kills — the life is still
// gained: whether the ability had a target is read before the damage, not after
// it, when the creature is already dead (#4082).
func TestSorinVampireLordGainsLifeWhenTheDamageKillsItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	sorin := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Sorin, Vampire Lord", "Legendary Planeswalker — Sorin", b6SorinVampireLordOracle, 4)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, sorin, 1, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(bear),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear survived 4 damage")
	}
	if me.Life != life+4 {
		t.Fatalf("life %d, want %d", me.Life, life+4)
	}
}

// Orchard Elemental — "Put two +1/+1 counters on this creature for each sprout
// vote. You gain 3 life for each harvest vote." With Doubling Season and
// Hardened Scales out the counters wait on the CR 616 order, and the life gain
// waits behind it, reading the tallies the vote's continuation filed.
func TestOrchardElementalLifeWaitsForTheCounterOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.VoteOrder(g)[0]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	effectstest.CastCatalogSpell(t, g, "Orchard Elemental", "Creature — Elemental", effectstest.OrchardElementalOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "sprout", "sprout", "sprout", "harvest")
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if me.Life != game.StartingLife {
		t.Fatalf("life %d with the order open, want %d: the gain ran before the answer", me.Life, game.StartingLife)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		effectstest.AnswerOrderBySource(t, table, ds, hs)
		table.SettleResolution()
		id, ok := effectstest.B43BattlefieldByName(table, "Orchard Elemental")
		if !ok {
			t.Fatal("Orchard Elemental did not enter")
		}
		if got := effectstest.AangCounterCount(table, id, effectstest.VotingATestPlusOneCounterKind); got != 13 {
			t.Fatalf("counters %d, want 13 ((6×2)+1, Doubling Season first)", got)
		}
		if p := table.PlayerByIDForEffect(me.ID); p.Life != game.StartingLife+3 {
			t.Fatalf("life %d after the order, want %d (3 for one harvest vote)", p.Life, game.StartingLife+3)
		}
	}
}
