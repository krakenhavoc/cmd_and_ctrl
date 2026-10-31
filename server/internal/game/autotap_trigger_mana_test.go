package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// autotap_trigger_mana_test.go — ADR 0074's amendment of 2026-10-30
// (#4716), slice 2: the auto-tap planner counts the mana triggered mana
// abilities add when it taps a source, judged by the trigger's own
// condition through the judge the executor fires with. The catalog cards
// and the reported game are in cards/effects; these are the rules, on
// stubbed triggers.

const (
	plannerGhastOracle      = "test-planner-crypt-ghast"
	plannerWakeOracle       = "test-planner-miraris-wake"
	plannerFertileOracle    = "test-planner-fertile-ground"
	plannerMonumentOracle   = "test-planner-forsaken-monument"
	plannerEveryTapOracle   = "test-planner-every-tap"
	plannerTreasureOracle   = "test-planner-ghast-treasure"
	plannerFlareOracle      = "test-planner-mana-flare"
	plannerWildGrowthOracle = wildGrowthTestOracle
)

// plannerTriggers is one stub per shape the planner has to judge. Each is
// the printed card's condition as the catalog writes it.
func plannerTriggers(key string) []ManaTrigger {
	yours := func(prod ManaProduced, source *Card) bool { return prod.Controller == source.Controller }
	fixed := func(s string) func(ManaProduced, *Card, *Game) string {
		return func(ManaProduced, *Card, *Game) string { return s }
	}
	anyType := func(prod ManaProduced, _ *Card, _ *Game) string { return prod.ProducedColorPipe() }
	ghast := ManaTrigger{
		Label: "Crypt Ghast — add an additional {B}",
		AppliesTo: func(prod ManaProduced, source *Card, _ *Game) bool {
			return yours(prod, source) && prod.Source.IsLand() && prod.Source.HasSubtype("Swamp")
		},
		Produced: fixed("{B}"),
	}
	switch key {
	case plannerWildGrowthOracle:
		return []ManaTrigger{wildGrowthTrigger("{G}")}
	case plannerFertileOracle:
		return []ManaTrigger{wildGrowthTrigger("{W|U|B|R|G}")}
	case plannerGhastOracle, plannerTreasureOracle:
		return []ManaTrigger{ghast}
	case plannerWakeOracle:
		return []ManaTrigger{{
			Label: "Mirari's Wake — add one mana of any type that land produced",
			AppliesTo: func(prod ManaProduced, source *Card, _ *Game) bool {
				return yours(prod, source) && prod.Source.IsLand()
			},
			Produced: anyType,
		}}
	case plannerFlareOracle:
		return []ManaTrigger{{
			Label:     "Mana Flare — that player adds one mana of any type that land produced",
			AppliesTo: func(prod ManaProduced, _ *Card, _ *Game) bool { return prod.Source.IsLand() },
			Produced:  anyType,
		}}
	case plannerMonumentOracle:
		return []ManaTrigger{{
			Label: "Forsaken Monument — add an additional {C}",
			AppliesTo: func(prod ManaProduced, source *Card, _ *Game) bool {
				if !yours(prod, source) {
					return false
				}
				for _, c := range prod.Colors {
					if c == "C" {
						return true
					}
				}
				return false
			},
			Produced: fixed("{C}"),
		}}
	case plannerEveryTapOracle:
		return []ManaTrigger{{
			Label:     "Every tap — add {G}",
			AppliesTo: func(prod ManaProduced, source *Card, _ *Game) bool { return yours(prod, source) },
			Produced:  fixed("{G}"),
		}}
	}
	return nil
}

// triggerTable is a game with plannerTriggers installed.
func triggerTable(t *testing.T) (*Game, *Player, *Player) {
	t.Helper()
	g := newActiveGame(t)
	withCatalogManaTriggers(t, plannerTriggers)
	return g, g.Seats[0], g.Seats[1]
}

// pushLandFor seeds a land whose mana is its basic land types.
func pushLandFor(g *Game, owner *Player, name, typeLine string) uuid.UUID {
	return pushBattlefieldForTest(g, owner.ID, name, typeLine, "")
}

// pushPlannerSource seeds a permanent with its own mana abilities and an
// oracle key for the triggers it carries ("" for none).
func pushPlannerSource(g *Game, owner *Player, name, typeLine, oracle string, abilities ...ManaAbilityShape) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = typeLine
	c.OracleID = oracle
	c.ManaAbilities = abilities
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// tapAbility is "{T}: Add <produced>".
func tapAbility(produced string) ManaAbilityShape {
	return ManaAbilityShape{TapCost: true, Produced: produced, Label: "Add " + produced}
}

// planTopUp is the move list's planner, under a read lock.
func planTopUp(t *testing.T, g *Game, who *Player, cost string) (tapPlan, bool) {
	t.Helper()
	var (
		plan tapPlan
		ok   bool
	)
	g.ReadSnapshot(func() {
		plan, _, ok = g.autoTapTopUpLocked(who.ID, costFor(t, cost), 0, ManaSpendContext{}, nil, 0)
	})
	return plan, ok
}

// payAutoTapped is the auto-tap payer's path (applyAutoTapLocked): the
// plan the clone check passes, with its fallback, carried out and spent.
// It reports whether the cost was paid; the pool keeps what floated.
func payAutoTapped(t *testing.T, g *Game, who *Player, cost string, prefer ManaSourceKinds) bool {
	t.Helper()
	parsed := costFor(t, cost)
	paid := false
	g.WithWriteLock(func() {
		plan, short, ok := g.autoTapFundedLocked(who, parsed, 0, ManaSpendContext{}, nil, prefer)
		if !ok {
			return
		}
		g.materializePlanInLocked(who, plan, short, ManaSpendContext{})
		_, paid = who.ManaPool.SpendManaFor(parsed, 0, ManaSpendContext{})
	})
	return paid
}

// sourcesFor is the gather's candidates for one card.
func sourcesFor(g *Game, who *Player, card uuid.UUID) []tapSource {
	var out []tapSource
	g.ReadSnapshot(func() {
		for _, s := range gatherSources(g, who.ID, nil, &autoTapWish{}, false) {
			if s.CardID == card {
				out = append(out, s)
			}
		}
	})
	return out
}

func isTapped(t *testing.T, g *Game, id uuid.UUID) bool {
	t.Helper()
	c := findCard(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// A Forest under a Wild Growth-style {G} is a two-mana source, so {G}{G}
// taps that Forest and leaves the plain one alone.
func TestPlannerCountsAFixedTrigger(t *testing.T) {
	g, me, _ := triggerTable(t)
	plain := pushForest(g, me)
	grown := pushForest(g, me)
	pushAuraOn(g, me, "Wild Growth", plannerWildGrowthOracle, grown)

	plan, ok := planTopUp(t, g, me, "{G}{G}")
	if !ok || len(plan) != 1 || plan[0].CardID != grown {
		t.Fatalf("plan for {G}{G} = %+v (ok %v), want the enchanted Forest alone", plan, ok)
	}
	if !reflect.DeepEqual(plan[0].TriggerColors, []string{"G"}) {
		t.Errorf("TriggerColors = %v, want [G]: the trigger's {G} pays the second pip", plan[0].TriggerColors)
	}
	if !payAutoTapped(t, g, me, "{G}{G}", 0) {
		t.Fatal("{G}{G} was not paid")
	}
	if !isTapped(t, g, grown) || isTapped(t, g, plain) {
		t.Errorf("tapped: enchanted %v, plain %v; want only the enchanted Forest", isTapped(t, g, grown), isTapped(t, g, plain))
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana floating, want none: %v", n, me.ManaPool)
	}
}

// The trigger reads the tapped land's subtype off prod.Source: an Island
// gets nothing from a Swamp-only trigger, and a Swamp-typed dual gets the
// extra {B} whichever colour it makes, so it needs no expansion.
func TestPlannerJudgesTheTappedLandsSubtype(t *testing.T) {
	g, me, _ := triggerTable(t)
	pushBattlefieldForTest(g, me.ID, "Crypt Ghast", "Creature — Spirit", plannerGhastOracle)
	pushLandFor(g, me, "Island", "Basic Land — Island")
	swamp := pushLandFor(g, me, "Swamp", "Basic Land — Swamp")

	if plan, ok := planTopUp(t, g, me, "{U}{U}"); ok {
		t.Errorf("{U}{U} planned off one Island beside a Swamp-only trigger: %+v", plan)
	}
	plan, ok := planTopUp(t, g, me, "{B}{B}")
	if !ok || len(plan) != 1 || plan[0].CardID != swamp {
		t.Errorf("plan for {B}{B} = %+v (ok %v), want the Swamp alone", plan, ok)
	}

	grave := pushPlannerSource(g, me, "Watery Grave", "Land — Island Swamp", "", tapAbility("{U|B}"))
	cands := sourcesFor(g, me, grave)
	var pick *tapSource
	for i := range cands {
		c := &cands[i]
		if c.TriggerSlots != 1 || c.OneColor != "" || c.PinnedSlots != nil {
			t.Errorf("candidate %+v: want one trigger slot and no expansion", *c)
		}
		if own := c.Slots[:c.ownSlots()]; len(own) == 1 && len(own[0].Options) == 2 {
			pick = c
		}
	}
	if pick == nil {
		t.Fatalf("no {U|B} pick among the dual's candidates: %+v", cands)
	}
	if got := pick.Slots[pick.ownSlots()].Options; !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("the dual's trigger slot = %v, want [B]", got)
	}
}

// CR 603.3a and the card's own text: "whenever a player taps a land for
// mana, that player adds …" sees the payer's land, so an opponent's Mana
// Flare is counted for the payer.
func TestPlannerCountsAnOpponentsThatPlayerTrigger(t *testing.T) {
	g, me, opp := triggerTable(t)
	pushBattlefieldForTest(g, opp.ID, "Mana Flare", "Enchantment", plannerFlareOracle)
	pushLandFor(g, me, "Swamp", "Basic Land — Swamp")
	pushLandFor(g, me, "Swamp", "Basic Land — Swamp")

	plan, ok := planTopUp(t, g, me, "{B}{B}")
	if !ok || len(plan) != 1 {
		t.Fatalf("plan for {B}{B} = %+v (ok %v), want one Swamp under the opponent's Mana Flare", plan, ok)
	}
	if !payAutoTapped(t, g, me, "{B}{B}", 0) || me.ManaPool.Total() != 0 {
		t.Errorf("{B}{B} off one flared Swamp: pool %v", me.ManaPool)
	}
}

// …and "whenever YOU tap a Swamp" is the trigger's controller's taps, so
// an opponent's Crypt Ghast adds nothing to the payer's plan.
func TestPlannerIgnoresAnOpponentsYouTrigger(t *testing.T) {
	g, me, opp := triggerTable(t)
	pushBattlefieldForTest(g, opp.ID, "Crypt Ghast", "Creature — Spirit", plannerGhastOracle)
	pushLandFor(g, me, "Swamp", "Basic Land — Swamp")

	if plan, ok := planTopUp(t, g, me, "{B}{B}"); ok {
		t.Errorf("{B}{B} planned off one Swamp and an opponent's Crypt Ghast: %+v", plan)
	}
	if !payAutoTapped(t, g, me, "{B}", 0) || me.ManaPool.Total() != 0 {
		t.Errorf("{B} off one Swamp: pool %v, want exactly the Swamp's {B} spent", me.ManaPool)
	}
}

// "One mana of any type that land produced" depends on the pick, so a
// {W|U} land under Mirari's Wake is two candidates, the pick fixed in
// each: it pays {W}{W} alone, and never {W}{U}, which the card cannot.
func TestPlannerExpandsAPickWhenTheTriggerReadsTheType(t *testing.T) {
	g, me, _ := triggerTable(t)
	pushBattlefieldForTest(g, me.ID, "Mirari's Wake", "Enchantment", plannerWakeOracle)
	dual := pushPlannerSource(g, me, "Azorius Land", "Land", "", tapAbility("{W|U}"))

	cands := sourcesFor(g, me, dual)
	var picks []string
	for _, c := range cands {
		picks = append(picks, c.OneColor)
		if c.TriggerSlots != 1 || !reflect.DeepEqual(c.Slots[1].Options, []string{c.OneColor}) {
			t.Errorf("candidate %+v: want the pick fixed and the trigger making the same colour", c)
		}
	}
	if !reflect.DeepEqual(picks, []string{"W", "U"}) {
		t.Fatalf("candidates fixed %v, want one per colour [W U]", picks)
	}
	if plan, ok := planTopUp(t, g, me, "{W}{U}"); ok {
		t.Errorf("{W}{U} planned off one dual and Mirari's Wake: %+v (stronger than printed)", plan)
	}
	if !payAutoTapped(t, g, me, "{W}{W}", 0) {
		t.Fatal("{W}{W} was not paid off the dual and Mirari's Wake")
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana floating after {W}{W}, want none: %v", n, me.ManaPool)
	}
}

// Item 5: a trigger's own pick makes the colour the plan booked. {1}{U/R}{B}
// off a Forest under a Fertile Ground-style "any color" and one Swamp: the
// plan books the trigger for {U/R} and the Swamp for {B}. A greedy pick
// would take {B}, the narrower pip, and leave {U/R} unpaid.
func TestTriggerPickMakesTheBookedColour(t *testing.T) {
	g, me, _ := triggerTable(t)
	forest := pushForest(g, me)
	pushAuraOn(g, me, "Fertile Ground", plannerFertileOracle, forest)
	swamp := pushLandFor(g, me, "Swamp", "Basic Land — Swamp")

	plan, ok := planTopUp(t, g, me, "{1}{U/R}{B}")
	if !ok || len(plan) != 2 {
		t.Fatalf("plan = %+v (ok %v), want the Forest and the Swamp", plan, ok)
	}
	for _, e := range plan {
		if e.CardID == forest && (len(e.TriggerColors) != 1 || (e.TriggerColors[0] != "U" && e.TriggerColors[0] != "R")) {
			t.Errorf("the Forest's trigger booked %v, want the {U/R}", e.TriggerColors)
		}
	}
	if !payAutoTapped(t, g, me, "{1}{U/R}{B}", 0) {
		t.Fatal("{1}{U/R}{B} was not paid: the trigger did not make the colour it was booked for")
	}
	if !isTapped(t, g, forest) || !isTapped(t, g, swamp) {
		t.Error("both lands should have been tapped")
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana floating, want none: %v", n, me.ManaPool)
	}
}

// CR 106.12a, 106.12b: a doubler doubles the Swamp's own mana, and the
// trigger is judged once on that doubled production and adds one {B} that
// is not itself doubled (Mana Reflection's ruling). Three, not four.
func TestDoublerThenTriggerCountsOnce(t *testing.T) {
	g, me, _ := triggerTable(t)
	g.testReplacements = append(g.testReplacements, manaProducedReplacement(me.ID, 2, "Test Reflection"))
	pushBattlefieldForTest(g, me.ID, "Crypt Ghast", "Creature — Spirit", plannerGhastOracle)
	swamp := pushLandFor(g, me, "Swamp", "Basic Land — Swamp")

	if plan, ok := planTopUp(t, g, me, "{B}{B}{B}{B}"); ok {
		t.Errorf("{B}{B}{B}{B} planned off one Swamp: %+v, want three mana, not four", plan)
	}
	plan, ok := planTopUp(t, g, me, "{B}{B}{B}")
	if !ok || len(plan) != 1 || plan[0].CardID != swamp {
		t.Fatalf("plan for {B}{B}{B} = %+v (ok %v), want the Swamp alone", plan, ok)
	}
	if !payAutoTapped(t, g, me, "{B}{B}{B}", 0) || me.ManaPool.Total() != 0 {
		t.Errorf("{B}{B}{B} off one doubled Swamp: pool %v", me.ManaPool)
	}
}

// One production, one trigger: a "{C} was produced" trigger beside a
// {C}{C} source adds one {C}, not one per {C} (Forsaken Monument's ruling).
func TestTriggerCountsOncePerProduction(t *testing.T) {
	g, me, _ := triggerTable(t)
	pushBattlefieldForTest(g, me.ID, "Forsaken Monument", "Artifact", plannerMonumentOracle)
	pushPlannerSource(g, me, "Sol Ring", "Artifact", "", tapAbility("{C}{C}"))

	if plan, ok := planTopUp(t, g, me, "{4}"); ok {
		t.Errorf("{4} planned off one Sol Ring: %+v, want three mana", plan)
	}
	if plan, ok := planTopUp(t, g, me, "{3}"); !ok || len(plan) != 1 {
		t.Fatalf("plan for {3} = %+v (ok %v), want the Sol Ring alone", plan, ok)
	}
	if !payAutoTapped(t, g, me, "{3}", 0) || me.ManaPool.Total() != 0 {
		t.Errorf("{3} off one Sol Ring under Forsaken Monument: pool %v", me.ManaPool)
	}
}

// CR 106.12: only an ability with {T} in its cost taps the permanent for
// mana. A sacrifice-only source and a card in hand get no trigger slots,
// as no trigger fires for them; a land beside them does.
func TestNoTriggerManaWithoutATap(t *testing.T) {
	g, me, _ := triggerTable(t)
	pushBattlefieldForTest(g, me.ID, "Every Tap", "Enchantment", plannerEveryTapOracle)
	gold := pushPlannerSource(g, me, "Gold", "Token Artifact — Gold", "",
		ManaAbilityShape{SacrificeCost: true, Produced: "{W|U|B|R|G}", Label: "Sacrifice: Add one mana of any color"})
	guide := seedSpiritGuide(me, "Simian Spirit Guide", "{R}")
	forest := pushForest(g, me)

	var all []tapSource
	g.ReadSnapshot(func() { all = gatherSources(g, me.ID, nil, &autoTapWish{}, false) })
	seen := map[uuid.UUID]int{}
	for _, s := range all {
		seen[s.CardID] = s.TriggerSlots
	}
	for name, id := range map[string]uuid.UUID{"Gold": gold, "Spirit Guide": guide} {
		if n, ok := seen[id]; !ok || n != 0 {
			t.Errorf("%s: trigger slots %d (gathered %v), want a source with none", name, n, ok)
		}
	}
	if seen[forest] != 1 {
		t.Errorf("Forest: trigger slots %d, want 1 — the tap fires the trigger", seen[forest])
	}
}

// Item 8: the plan judges every trigger against the board before the
// payment, and the executor at each tap. Here they differ: the Crypt
// Ghast-style trigger rides a Treasure the cast wants cracked, so the plan
// cracks it first and the Swamp's extra {B} never arrives. The clone check
// refuses that plan and the payer pays with the plan it made before the
// amendment: the Treasure and both Swamps.
func TestPlannerFallsBackWhenTheTriggerManaWillNotArrive(t *testing.T) {
	g, me, _ := triggerTable(t)
	treasure := pushPlannerSource(g, me, "Treasure", "Token Artifact — Treasure", plannerTreasureOracle,
		ManaAbilityShape{TapCost: true, SacrificeCost: true, Produced: "{W|U|B|R|G}", Label: "{T}, Sacrifice: Add one mana of any color"})
	s1 := pushLandFor(g, me, "Swamp", "Basic Land — Swamp")
	s2 := pushLandFor(g, me, "Swamp", "Basic Land — Swamp")

	var (
		first tapPlan
		funds bool
	)
	g.WithWriteLock(func() {
		var short ParsedCost
		var ok bool
		first, short, ok = g.autoTapTopUpLocked(me.ID, costFor(t, "{B}{B}{B}"), 0, ManaSpendContext{}, nil, ManaSourceTreasure)
		if !ok {
			t.Fatal("no plan counting the trigger mana")
		}
		funds = g.planFundsLocked(me, first, short, costFor(t, "{B}{B}{B}"), 0, ManaSpendContext{})
	})
	if len(first) != 2 || first[0].CardID != treasure || !first.countsTriggerMana() {
		t.Fatalf("first plan = %+v, want the wished Treasure, then one Swamp counted for {B}{B}", first)
	}
	if funds {
		t.Fatal("the clone check passed a plan whose trigger leaves with the Treasure")
	}

	if !payAutoTapped(t, g, me, "{B}{B}{B}", ManaSourceTreasure) {
		t.Fatal("{B}{B}{B} was refused, but the plan before the amendment pays it")
	}
	if findCard(g, treasure) != nil || !isTapped(t, g, s1) || !isTapped(t, g, s2) {
		t.Error("want the Treasure cracked and both Swamps tapped, the fallback plan")
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana floating, want none: %v", n, me.ManaPool)
	}
}

// The gate: with no triggered mana ability on the board the planner is
// the one it was. Counting trigger mana or not, every plan is the same.
func TestPlanUnchangedWithoutManaTriggers(t *testing.T) {
	g, me, _ := triggerTable(t)
	pushForest(g, me)
	pushLandFor(g, me, "Island", "Basic Land — Island")
	pushPlannerSource(g, me, "Azorius Land", "Land", "", tapAbility("{W|U}"))
	pushPlannerSource(g, me, "Sol Ring", "Artifact", "", tapAbility("{C}{C}"))
	pushPlannerSource(g, me, "Treasure", "Token Artifact — Treasure", "",
		ManaAbilityShape{TapCost: true, SacrificeCost: true, Produced: "{W|U|B|R|G}", Label: "{T}, Sacrifice: Add one mana of any color"})

	for _, cost := range []string{"{G}", "{1}{U}", "{1}{W}{U}", "{3}", "{2}{G}{W}", "{W}{U}{G}{1}"} {
		var (
			with, without tapPlan
			okW, okWO     bool
		)
		g.ReadSnapshot(func() {
			with, _, okW = g.autoTapTopUpWithLocked(me.ID, costFor(t, cost), 0, ManaSpendContext{}, nil, 0, true)
			without, _, okWO = g.autoTapTopUpWithLocked(me.ID, costFor(t, cost), 0, ManaSpendContext{}, nil, 0, false)
		})
		if okW != okWO || !reflect.DeepEqual(with, without) {
			t.Errorf("%s: counting %+v (%v), not counting %+v (%v)", cost, with, okW, without, okWO)
		}
		if with.countsTriggerMana() {
			t.Errorf("%s: a plan counted trigger mana on a board with none", cost)
		}
	}
}

// CR 106.7, Decision 5 and the amendment's item 9: what a permanent
// "could produce" counts only its own abilities. Wild Growth's {G} is the
// Aura's, so an Island under it could produce {U} and nothing else.
func TestCouldProduceIgnoresTriggeredMana(t *testing.T) {
	g, me, _ := triggerTable(t)
	island := pushLandFor(g, me, "Island", "Basic Land — Island")
	pushAuraOn(g, me, "Wild Growth", plannerWildGrowthOracle, island)

	var got []string
	g.ReadSnapshot(func() { got = g.ProducibleManaLocked(*findCard(g, island)) })
	if !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("the enchanted Island could produce %v, want [U]", got)
	}
}

// Owner decision 1: every combination of a source's picks is judged, up
// to 36. A two-pick land under Mirari's Wake (25) is one candidate per
// combination, its picks pinned, and pays {W}{W}{W} alone; a three-pick
// land (125) gets no trigger mana in the plan.
func TestTriggerPickExpansionIsCappedAt36(t *testing.T) {
	g, me, _ := triggerTable(t)
	pushBattlefieldForTest(g, me.ID, "Mirari's Wake", "Enchantment", plannerWakeOracle)
	two := pushPlannerSource(g, me, "Two Picks", "Land", "", tapAbility("{W|U|B|R|G}{W|U|B|R|G}"))
	three := pushPlannerSource(g, me, "Three Picks", "Land", "", tapAbility("{W|U|B|R|G}{W|U|B|R|G}{W|U|B|R|G}"))

	twos := sourcesFor(g, me, two)
	if len(twos) != 25 {
		t.Fatalf("two-pick land: %d candidates, want 25", len(twos))
	}
	for _, c := range twos {
		if c.TriggerSlots != 1 || len(c.PinnedSlots) != 2 || c.PinnedSlots[0] == "" || c.PinnedSlots[1] == "" {
			t.Errorf("two-pick candidate %+v: want both picks pinned and one trigger slot", c)
			break
		}
	}
	threes := sourcesFor(g, me, three)
	if len(threes) != 1 || threes[0].TriggerSlots != 0 {
		t.Fatalf("three-pick land: %+v, want one candidate with no trigger mana (above the cap)", threes)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == three {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if !payAutoTapped(t, g, me, "{W}{W}{W}", 0) {
		t.Fatal("{W}{W}{W} was not paid off the two-pick land and Mirari's Wake")
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana floating, want none: %v", n, me.ManaPool)
	}
}

// #2455's costed plan counts trigger mana too, and names the trigger's
// permanent as the source of the mana that funds a Signet. One Swamp
// under Crypt Ghast makes {B}{B}; one of them pays an Orzhov Signet's {1},
// and {W}{B}{B} is paid with nothing left over.
func TestCostedPlanCountsTriggerMana(t *testing.T) {
	g, me, _ := triggerTable(t)
	pushBattlefieldForTest(g, me.ID, "Crypt Ghast", "Creature — Spirit", plannerGhastOracle)
	swamp := pushLandFor(g, me, "Swamp", "Basic Land — Swamp")
	signet := pushPlannerSource(g, me, "Orzhov Signet", "Artifact", "",
		ManaAbilityShape{TapCost: true, ManaCost: "{1}", Produced: "{W}{B}", Label: "{1}, {T}: Add {W}{B}"})

	plan, ok := planTopUp(t, g, me, "{W}{B}{B}")
	if !ok || len(plan) != 2 {
		t.Fatalf("plan = %+v (ok %v), want the Swamp and the Signet", plan, ok)
	}
	if plan[0].CardID != swamp || len(plan[0].TriggerColors) != 1 || plan[1].CardID != signet || !plan[1].Costed {
		t.Errorf("plan = %+v, want the Swamp with its trigger booked, then the costed Signet", plan)
	}
	if !payAutoTapped(t, g, me, "{W}{B}{B}", 0) {
		t.Fatal("{W}{B}{B} was not paid off a Ghast Swamp and a Signet")
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana floating, want none: %v", n, me.ManaPool)
	}
}
