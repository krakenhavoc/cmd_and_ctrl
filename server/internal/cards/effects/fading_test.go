package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fading_test.go — the fading seam (CR 702.32a, #5802) and the Nemesis
// cards that ship on it, on the real engine.

const (
	fadingAncientHydraOracle       = "8e1d3722-9bd1-4ab7-b4a5-9265fca3aba9"
	fadingBlastodermOracle         = "d6ee8bd6-0207-46ee-9892-c1d1b23b934e"
	fadingCloudskateOracle         = "648e4072-b991-46fd-b057-1e807091dcdf"
	fadingDefenderEnVecOracle      = "bb8a55a5-596e-427c-b745-4de8b0dfb095"
	fadingJoltingMerfolkOracle     = "3b71aeee-a4a8-46d3-9973-06db14d8bf6c"
	fadingPhyrexianProwlerOracle   = "128ec451-97f6-4f01-8929-689fd4c45f40"
	fadingRejuvenationOracle       = "c1ebe189-db71-47cd-a763-8cb8c305383a"
	fadingRustingGolemOracle       = "688c232f-385b-4bd5-aed0-e55059111fb3"
	fadingSaprolingBurstOracle     = "7b31ecf8-5266-4376-ace0-8b873706c4ea"
	fadingSkyshroudBehemothOracle  = "77bee041-6f96-44e3-a18c-b52e508b9bdf"
	fadingSkyshroudRidgebackOracle = "34a160bc-5d7e-483c-b24c-ab30b7d313b0"
	fadingWoodripperOracle         = "d0653734-9eb7-48c6-9042-65ffdd50fd2a"
	fadingTangleWireOracle         = "c4ea0931-43e0-4214-9e0b-213eee3a7e69"

	fadingDoublingSeasonOracle = "01546b7d-a233-4176-8843-d732074dc5b6"
	fadingGloriousAnthemOracle = "e3886fe8-9b76-4613-8891-4ec74657c087"
)

func TestFadingCardsAreRegistered(t *testing.T) {
	want := map[string]string{
		fadingAncientHydraOracle:       "Ancient Hydra",
		fadingBlastodermOracle:         "Blastoderm",
		fadingCloudskateOracle:         "Cloudskate",
		fadingDefenderEnVecOracle:      "Defender en-Vec",
		fadingJoltingMerfolkOracle:     "Jolting Merfolk",
		fadingPhyrexianProwlerOracle:   "Phyrexian Prowler",
		fadingRejuvenationOracle:       "Rejuvenation Chamber",
		fadingRustingGolemOracle:       "Rusting Golem",
		fadingSaprolingBurstOracle:     "Saproling Burst",
		fadingSkyshroudBehemothOracle:  "Skyshroud Behemoth",
		fadingSkyshroudRidgebackOracle: "Skyshroud Ridgeback",
		fadingWoodripperOracle:         "Woodripper",
		fadingTangleWireOracle:         "Tangle Wire",
	}
	for id, name := range want {
		s, ok := Lookup(id)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, id)
			continue
		}
		if s.Name != name {
			t.Errorf("%s is registered as %q, want %q", id, s.Name, name)
		}
		if s.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", name, s.Completeness)
		}
	}
}

// Fading is two abilities (CR 702.32a), so a card with one half and not
// the other is a card that either never fades or fades from nothing.
// Every catalog card with a fade-counter entry has exactly one fading
// upkeep trigger, and the other way round.
func TestEveryFadingCardHasBothHalves(t *testing.T) {
	for _, s := range All() {
		entries := 0
		for _, r := range s.Replacements {
			if r.SelfEntryCounters != nil && r.SelfEntryCounters.Kind == game.CounterFade {
				entries++
			}
		}
		triggers := 0
		for _, tr := range s.Triggered {
			if tr.Keyword == KeywordFading {
				triggers++
			}
		}
		if entries != triggers || entries > 1 {
			t.Errorf("%s: %d fade-counter entries and %d fading upkeep triggers, want one of each", s.Name, entries, triggers)
		}
	}
}

// fadingToUpkeepOfSeat0 walks to the next beginning of seat 0's upkeep
// and stops there, with whatever that put on the stack still waiting.
func fadingToUpkeepOfSeat0(t *testing.T, g *game.Game) {
	t.Helper()
	fadingToUpkeepOf(t, g, 0)
}

// fadingToUpkeepOf is fadingToUpkeepOfSeat0 for any seat.
func fadingToUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == seat && g.Turn.Step == game.StepUpkeep); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// fadingTriggerWaiting reports whether a fading upkeep trigger from id
// is waiting, on the stack or about to be put there.
func fadingTriggerWaiting(g *game.Game, id uuid.UUID) bool {
	if item := triggerOnStack(g, id); item != nil {
		return true
	}
	for _, p := range g.PendingTriggers {
		if p != nil && p.SourceCardID == id {
			return true
		}
	}
	return false
}

func fadingNoChoiceFor(t *testing.T, g *game.Game, who uuid.UUID, when string) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == who {
			t.Fatalf("%s: the fading permanent's controller was asked something (%s)", when, c.Kind)
		}
	}
}

// Fading 2: enters with two counters, loses one at each of the next two
// upkeeps, and is sacrificed at the third, when there is none to
// remove. The trigger fires at that third upkeep with no counter on the
// permanent (no intervening "if", unlike vanishing), and nobody is asked
// anything on the way.
func TestFadingRemovesOneCounterPerUpkeepThenSacrificesWhenItCant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b38CastCreature(t, g, "Skyshroud Ridgeback", "Creature — Beast", fadingSkyshroudRidgebackOracle, 2, 3)
	if got := counterCount(g, id, game.CounterFade); got != 2 {
		t.Fatalf("entered with %d fade counters, want 2", got)
	}
	for _, want := range []int{1, 0} {
		fadingToUpkeepOfSeat0(t, g)
		if !fadingTriggerWaiting(g, id) {
			t.Fatalf("no fading trigger at the upkeep (counters %d)", counterCount(g, id, game.CounterFade))
		}
		passPriorityAroundTable(t, g)
		fadingNoChoiceFor(t, g, me.ID, "after removing a counter")
		if got := counterCount(g, id, game.CounterFade); got != want {
			t.Fatalf("after an upkeep: %d fade counters, want %d", got, want)
		}
	}
	// Zero counters: still on the battlefield until the trigger resolves.
	fadingToUpkeepOfSeat0(t, g)
	if _, here := battlefieldCard(g, id); !here {
		t.Fatal("the Ridgeback left before its upkeep trigger resolved")
	}
	if !fadingTriggerWaiting(g, id) {
		t.Fatal("fading did not trigger with no fade counter on it")
	}
	passPriorityAroundTable(t, g)
	fadingNoChoiceFor(t, g, me.ID, "after the sacrifice")
	if _, here := battlefieldCard(g, id); here {
		t.Fatal("the Ridgeback survived an upkeep with no fade counter to remove")
	}
	if !inGraveyard(me, id) {
		t.Error("the sacrificed Ridgeback is not in its owner's graveyard")
	}
}

// The entry counters go through the counter window, so Doubling Season
// doubles them (Doubling Season's ruling: a permanent that would enter
// with counters enters with twice as many).
func TestFadingEntryCountersAreDoubledByDoublingSeason(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Doubling Season", "Enchantment", fadingDoublingSeasonOracle, false)
	id := b38CastCreature(t, g, "Blastoderm", "Creature — Beast", fadingBlastodermOracle, 5, 5)
	if got := counterCount(g, id, game.CounterFade); got != 6 {
		t.Errorf("Blastoderm under Doubling Season entered with %d fade counters, want 6", got)
	}
	if !hasEffectiveKeyword(t, g, id, "shroud") {
		t.Error("Blastoderm has no shroud")
	}
}

// Skyshroud Behemoth's counters and "enters tapped" are one replacement:
// it arrives tapped with two counters and no CR 616 ordering question.
func TestSkyshroudBehemothEntersTappedWithTwoFadeCountersAndNoQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b38CastCreature(t, g, "Skyshroud Behemoth", "Creature — Beast", fadingSkyshroudBehemothOracle, 10, 10)
	fadingNoChoiceFor(t, g, me.ID, "as it entered")
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the Behemoth did not enter")
	}
	if !c.Tapped {
		t.Error("the Behemoth entered untapped")
	}
	if got := c.Counters[game.CounterFade]; got != 2 {
		t.Errorf("%d fade counters, want 2", got)
	}
}

// "Remove a fade counter" is a cost paid as the ability is activated: it
// can be paid once per counter and not after, and a permanent that spent
// them all is sacrificed at the next upkeep.
func TestFadeCounterCostSpendsTheCountersTheUpkeepWouldRemove(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b38CastCreature(t, g, "Phyrexian Prowler", "Creature — Phyrexian Zombie Mercenary", fadingPhyrexianProwlerOracle, 3, 3)
	for i := 0; i < 3; i++ {
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		if got := counterCount(g, id, game.CounterFade); got != 2-i {
			t.Fatalf("after activation %d: %d fade counters, want %d (paid at announcement)", i+1, got, 2-i)
		}
		passPriorityAroundTable(t, g)
	}
	if p, tg := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 6 || tg != 6 {
		t.Errorf("Prowler is %d/%d after three pumps, want 6/6", p, tg)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated with no fade counter to remove")
	}
	fadingToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	if _, here := battlefieldCard(g, id); here {
		t.Error("the Prowler survived the upkeep after spending every fade counter")
	}
}

func TestJoltingMerfolkTapsATargetCreatureForAFadeCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	id := b38CastCreature(t, g, "Jolting Merfolk", "Creature — Merfolk", fadingJoltingMerfolkOracle, 2, 2)
	if got := counterCount(g, id, game.CounterFade); got != 4 {
		t.Fatalf("entered with %d fade counters, want 4", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tapped, ok := battlefieldCardTapped(g, bear); !ok || !tapped {
		t.Error("the target creature is not tapped")
	}
	if got := counterCount(g, id, game.CounterFade); got != 3 {
		t.Errorf("%d fade counters after one activation, want 3", got)
	}
}

// Ancient Hydra's ping costs {1} and a fade counter; Defender en-Vec's
// shield costs a fade counter and prevents the next 2 damage to its
// target this turn, so a 2/2 it shields survives two pings.
func TestAncientHydraPingsAndDefenderEnVecShields(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hydra := b38CastCreature(t, g, "Ancient Hydra", "Creature — Hydra", fadingAncientHydraOracle, 5, 1)
	defender := b38CastCreature(t, g, "Defender en-Vec", "Creature — Human Cleric", fadingDefenderEnVecOracle, 2, 4)
	if got := counterCount(g, hydra, game.CounterFade); got != 5 {
		t.Fatalf("Hydra entered with %d fade counters, want 5", got)
	}

	// The {1} is part of the cost: with no mana the ping is refused and
	// no fade counter is spent. Only a Strict activation pays mana.
	drainPool(t, g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, hydra, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Strict:  true,
	}); err == nil {
		t.Fatal("pinged with no mana for its {1}")
	}
	if got := counterCount(g, hydra, game.CounterFade); got != 5 {
		t.Fatalf("Hydra has %d fade counters after a refused ping, want 5", got)
	}

	before := lifeOf(g, opp.ID)
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, hydra, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Strict:  true,
	}); err != nil {
		t.Fatalf("ping: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != before-1 {
		t.Errorf("opponent life %d, want %d", got, before-1)
	}
	if got := counterCount(g, hydra, game.CounterFade); got != 4 {
		t.Errorf("Hydra has %d fade counters after one ping, want 4", got)
	}

	if err := g.ActivateCatalogAbility(me.ID, defender, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("shield: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, defender, game.CounterFade); got != 3 {
		t.Errorf("Defender has %d fade counters after one shield, want 3", got)
	}
	for i := 0; i < 2; i++ {
		fillPool(me, 1)
		if err := g.ActivateCatalogAbility(me.ID, hydra, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		}); err != nil {
			t.Fatalf("ping %d at the Bear: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	if _, here := battlefieldCard(g, bear); !here {
		t.Error("the shielded Bear died to two pings; 2 damage should have been prevented")
	}
}

func TestWoodripperDestroysATargetArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := pushCatalogPermanent(g, opp.ID, "Relic", "Artifact", "", false)
	id := b38CastCreature(t, g, "Woodripper", "Creature — Beast", fadingWoodripperOracle, 4, 6)
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, here := battlefieldCard(g, relic); here {
		t.Error("the artifact survived")
	}
	if got := counterCount(g, id, game.CounterFade); got != 2 {
		t.Errorf("%d fade counters after one activation, want 2", got)
	}
}

func TestRejuvenationChamberGainsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b38CastCreature(t, g, "Rejuvenation Chamber", "Artifact", fadingRejuvenationOracle, 0, 0)
	if got := counterCount(g, id, game.CounterFade); got != 2 {
		t.Fatalf("entered with %d fade counters, want 2", got)
	}
	before := lifeOf(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != before+2 {
		t.Errorf("life %d, want %d", got, before+2)
	}
}

// Rusting Golem is as big as its fade counters: 5/5, then 4/4 after an
// upkeep, and a 0/0 the state-based actions take once the fifth upkeep
// removes the last.
func TestRustingGolemIsTheSizeOfItsFadeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b38CastCreature(t, g, "Rusting Golem", "Artifact Creature — Golem", fadingRustingGolemOracle, 0, 0)
	if p, tg := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 5 || tg != 5 {
		t.Fatalf("Rusting Golem is %d/%d on entry, want 5/5", p, tg)
	}
	for want := 4; want >= 1; want-- {
		fadingToUpkeepOfSeat0(t, g)
		passPriorityAroundTable(t, g)
		if p, tg := effectivePower(t, g, id), effectiveToughness(t, g, id); p != want || tg != want {
			t.Fatalf("Rusting Golem is %d/%d, want %d/%d", p, tg, want, want)
		}
	}
	fadingToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	if _, here := battlefieldCard(g, id); here {
		t.Error("Rusting Golem survived as a 0/0")
	}
	if !inGraveyard(me, id) {
		t.Error("Rusting Golem is not in its owner's graveyard")
	}
}

func fadingSaprolingsOf(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == "Saproling" {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// Each Saproling is as big as the fade counters on the Saproling Burst
// that made it, shrinking as the Burst spends them, and the Burst's
// leaves trigger destroys them even when something else keeps them
// alive (a Glorious Anthem's +1/+1).
func TestSaprolingBurstsSaprolingsTrackItsFadeCountersAndDieWithIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	burst := b38CastCreature(t, g, "Saproling Burst", "Enchantment", fadingSaprolingBurstOracle, 0, 0)
	if got := counterCount(g, burst, game.CounterFade); got != 7 {
		t.Fatalf("entered with %d fade counters, want 7", got)
	}
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, burst, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	saps := fadingSaprolingsOf(g, me.ID)
	if len(saps) != 2 {
		t.Fatalf("%d Saprolings, want 2", len(saps))
	}
	for _, s := range saps {
		if p, tg := effectivePower(t, g, s), effectiveToughness(t, g, s); p != 5 || tg != 5 {
			t.Errorf("Saproling is %d/%d with five fade counters on the Burst, want 5/5", p, tg)
		}
	}
	fadingToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	for _, s := range saps {
		if p := effectivePower(t, g, s); p != 4 {
			t.Errorf("Saproling power %d after the upkeep, want 4", p)
		}
	}

	pushCatalogPermanent(g, me.ID, "Glorious Anthem", "Enchantment", fadingGloriousAnthemOracle, false)
	killCreature(t, g, me.ID, burst)
	passPriorityAroundTable(t, g)
	if left := fadingSaprolingsOf(g, me.ID); len(left) != 0 {
		t.Errorf("%d Saprolings survived the Burst leaving, want 0", len(left))
	}
}

// A Saproling Burst that spends its last counter makes 0/0s of every
// Saproling it made, including the one that last counter paid for.
func TestSaprolingBurstsLastCounterLeavesNoSaproling(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	burst := b38CastCreature(t, g, "Saproling Burst", "Enchantment", fadingSaprolingBurstOracle, 0, 0)
	for i := 0; i < 7; i++ {
		if err := g.ActivateCatalogAbility(me.ID, burst, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	if left := fadingSaprolingsOf(g, me.ID); len(left) != 0 {
		t.Errorf("%d Saprolings with no fade counter on the Burst, want 0", len(left))
	}
	if _, here := battlefieldCard(g, burst); !here {
		t.Error("the Burst left before its upkeep")
	}
}

func fadingTappedCount(g *game.Game, ids []uuid.UUID) int {
	n := 0
	for _, id := range ids {
		if tapped, ok := battlefieldCardTapped(g, id); ok && tapped {
			n++
		}
	}
	return n
}

// Tangle Wire at another player's upkeep: that player chooses as many of
// their untapped artifacts, creatures and lands as the Wire has fade
// counters, and a player with no more than that taps them all unasked.
func TestTangleWireTapsOnePermanentPerFadeCounterAtEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	opp, third := g.Seats[1], g.Seats[2]
	wire := b38CastCreature(t, g, "Tangle Wire", "Artifact", fadingTangleWireOracle, 0, 0)
	if got := counterCount(g, wire, game.CounterFade); got != 4 {
		t.Fatalf("entered with %d fade counters, want 4", got)
	}
	var oppLands, thirdLands []uuid.UUID
	for i := 0; i < 5; i++ {
		oppLands = append(oppLands, b12Permanent(g, opp.ID, "Wastes", "Basic Land"))
	}
	for i := 0; i < 2; i++ {
		thirdLands = append(thirdLands, b12Permanent(g, third.ID, "Wastes", "Basic Land"))
	}

	fadingToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, opp.ID, oppLands[:4]...)
	passPriorityAroundTable(t, g)
	if got := fadingTappedCount(g, oppLands); got != 4 {
		t.Errorf("the upkeep's player tapped %d lands, want 4", got)
	}
	if tapped, _ := battlefieldCardTapped(g, oppLands[4]); tapped {
		t.Error("the land not chosen was tapped")
	}
	if got := counterCount(g, wire, game.CounterFade); got != 4 {
		t.Errorf("another player's upkeep removed a fade counter: %d, want 4", got)
	}

	fadingToUpkeepOf(t, g, 2)
	passPriorityAroundTable(t, g)
	fadingNoChoiceFor(t, g, third.ID, "with fewer permanents than counters")
	if got := fadingTappedCount(g, thirdLands); got != 2 {
		t.Errorf("a player with two lands and four counters tapped %d, want both", got)
	}
}
