package effects

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// mkm_collect_evidence_cards_test.go — #5193: the Murders at Karlov Manor
// cards built on collect evidence (CR 701.59), each played through its
// real catalog entry: the optional additional cost to cast, the activated
// ability's cost, and the effect form.

const (
	mkmCEPollenOracle     = "ec7d550b-9e13-4249-89c7-6ad43f71ea11"
	mkmCEMaskOracle       = "85ac3efc-5e19-4661-b52f-116f57e39581"
	mkmCEBiteOracle       = "ff387264-31da-4f6b-88c6-2843f56acbda"
	mkmCESpriteOracle     = "5e238e67-00f2-4656-bf69-0115e47321a5"
	mkmCEConfessOracle    = "96480e86-70a8-401d-abf8-a7d509311b73"
	mkmCEResearcherOracle = "75e7f7c7-fa98-497b-bf5e-8f74571e190c"
	mkmCEWhispererOracle  = "923b2a1b-d7b6-439d-9681-c7ba88165c31"
	mkmCEIzoniOracle      = "8074afbf-190b-4a24-b939-cc559bdffe38"
	mkmCEOrbOracle        = "bb039957-9e88-48cf-b41c-38d42eb87251"
	mkmCECollectorOracle  = "566b65bb-fa5d-4da6-a766-b08128ee0049"
	mkmCEInspectorOracle  = "22a33139-a730-4792-9dd5-2c0f99f96855"
)

func TestMkmCollectEvidenceCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mkmCEPollenOracle:     "Analyze the Pollen",
		mkmCEMaskOracle:       "Behind the Mask",
		mkmCEBiteOracle:       "Bite Down on Crime",
		mkmCESpriteOracle:     "Crimestopper Sprite",
		mkmCEConfessOracle:    "Extract a Confession",
		mkmCEResearcherOracle: "Forensic Researcher",
		mkmCEWhispererOracle:  "Hedge Whisperer",
		mkmCEIzoniOracle:      "Izoni, Center of the Web",
		mkmCEOrbOracle:        "Polygraph Orb",
		mkmCECollectorOracle:  "Sample Collector",
		mkmCEInspectorOracle:  "Vitu-Ghazi Inspector",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// The optional cost's shape: one offer keyed collect-evidence, an
// any-number graveyard exile whose floor is the printed N.
func TestMkmCollectEvidenceOffersAreTheSeam(t *testing.T) {
	for oracle, n := range map[string]int{
		mkmCEPollenOracle: 8, mkmCEMaskOracle: 6, mkmCEBiteOracle: 6,
		mkmCESpriteOracle: 6, mkmCEConfessOracle: 6, mkmCEInspectorOracle: 6,
	} {
		costs := game.OptionalCostsFor(oracle)
		if len(costs) != 1 || costs[0].Key != game.CollectEvidenceKey {
			t.Errorf("%s: optional costs %+v, want one collect evidence offer", oracle, costs)
			continue
		}
		if ec := costs[0].ExileCards; !ec.IsCollectEvidence() || ec.ManaValueAtLeast != n {
			t.Errorf("%s: clause %+v, want collect evidence %d", oracle, ec, n)
		}
	}
}

// Register refuses the shapes that would charge something unprinted.
func TestRegisterRefusesAMalformedCollectEvidence(t *testing.T) {
	for name, spec := range map[string]Spec{
		"a floor that is not collect evidence": {OracleID: "test-ce-bad-floor", Name: "Bad Floor",
			Activated: []ActivatedAbility{{Label: "Exile: draw.", Cost: game.AbilityCost{ExileCards: &game.ExileCost{
				AnyNumber: true, From: game.ZoneHand, ManaValueAtLeast: 3}}, Effect: Do()}}},
		"two floors": {OracleID: "test-ce-two-floors", Name: "Two Floors",
			Activated: []ActivatedAbility{{Label: "Exile: draw.", Cost: game.AbilityCost{ExileCards: &game.ExileCost{
				AnyNumber: true, From: game.ZoneGraveyard, ManaValueAtLeast: 3,
				SymbolTotal: &game.ManaSymbolTotal{Color: "B", AtLeast: 2}}}, Effect: Do()}}},
		"two optional collect evidence costs": {OracleID: "test-ce-two", Name: "Two Offers",
			OptionalCosts: []game.AdditionalCost{OptionalCollectEvidence(2), OptionalCollectEvidence(3)}},
		"an optional collect evidence beside a mandatory graveyard exile": {OracleID: "test-ce-beside", Name: "Beside",
			AdditionalCost: ExileFromGraveyardCost(2, "two cards", nil),
			OptionalCosts:  []game.AdditionalCost{OptionalCollectEvidence(3)}},
		"collect evidence with mana beside it": {OracleID: "test-ce-mana", Name: "With Mana",
			OptionalCosts: []game.AdditionalCost{func() game.AdditionalCost {
				c := OptionalCollectEvidence(3)
				c.ManaCost = "{1}"
				return c
			}()}},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("Register accepted %s", name)
				}
			}()
			Register(spec)
		})
	}
}

// --- helpers --------------------------------------------------------------

// mkmCECast puts a catalog card in the active seat's hand, walks to a main
// phase and casts it with params (no mana is charged).
func mkmCECast(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	return id, g.CastSpell(me.ID, id, params)
}

// mkmCECollecting is the cast params that take the card's one collect
// evidence offer with `ids`, beside the given targets.
func mkmCECollecting(targets []game.TargetRef, ids ...uuid.UUID) game.CastSpellParams {
	return game.CastSpellParams{Targets: targets, OptionalCosts: []int{0}, ExileIDs: ids}
}

// mkmCEEvidence fills p's graveyard with one sorcery per mana cost.
func mkmCEEvidence(p *game.Player, costs ...string) []uuid.UUID {
	out := make([]uuid.UUID, len(costs))
	for i, cost := range costs {
		out[i] = gyCard(p, "Evidence", "Sorcery", cost, "")
	}
	return out
}

// --- Analyze the Pollen ---------------------------------------------------

func TestAnalyzeThePollenSearchesForALandOrWithEvidenceACreature(t *testing.T) {
	for _, collected := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		creature := pushLibraryCardForTest(me, game.Card{Name: "Library Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})
		forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		params := game.CastSpellParams{}
		if collected {
			params = mkmCECollecting(nil, mkmCEEvidence(me, "{4}", "{4}")...)
		}
		if _, err := mkmCECast(t, g, "Analyze the Pollen", "Sorcery", mkmCEPollenOracle, params); err != nil {
			t.Fatalf("collected=%v: cast: %v", collected, err)
		}
		passPriorityAroundTable(t, g)
		c := searchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("collected=%v: no search prompt", collected)
		}
		offered := map[uuid.UUID]bool{}
		for _, id := range c.SearchCards {
			offered[id] = true
		}
		if !offered[forest] {
			t.Errorf("collected=%v: the basic land is not offered", collected)
		}
		if offered[creature] != collected {
			t.Errorf("collected=%v: creature offered = %v, want %v", collected, offered[creature], collected)
		}
		pick := forest
		if collected {
			pick = creature
		}
		answerSearchByID(t, g, me.ID, pick)
		if !me.Hand.Contains(pick) {
			t.Errorf("collected=%v: the found card is not in hand", collected)
		}
	}
}

// CR 701.59b: with less than 8 in the graveyard the offer is refused.
func TestAnalyzeThePollenRefusesEvidenceBelowEight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gy := mkmCEEvidence(me, "{3}", "{4}")
	if _, err := mkmCECast(t, g, "Analyze the Pollen", "Sorcery", mkmCEPollenOracle, mkmCECollecting(nil, gy...)); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("cast with 7 = %v, want ErrInvalidParam", err)
	}
	if me.Graveyard.Size() != 2 {
		t.Error("a refused collect evidence exiled cards")
	}
}

// The bot reads the same gate on a cast: a move that takes the offer names
// cards worth 8 or more, and none takes it while the graveyard is short.
func TestAnalyzeThePollenEnumeratorOffersEvidenceOnlyWhenPayable(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	pollen := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pollen, Name: "Analyze the Pollen", TypeLine: "Sorcery", ManaCost: "{G}",
		OracleID: mkmCEPollenOracle, Owner: me.ID, Controller: me.ID})
	fillPoolColored(me, "G", 1)
	values := map[uuid.UUID]int{
		gyCard(me, "Evidence", "Sorcery", "{3}", ""): 3,
		gyCard(me, "Evidence", "Sorcery", "{4}", ""): 4,
	}
	collecting := func() (moves int, short bool) {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != "cast_spell" || m.Source != pollen {
				continue
			}
			var p struct {
				OptionalCosts []int    `json:"optional_costs"`
				ExileIDs      []string `json:"exile_ids"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			if len(p.OptionalCosts) == 0 {
				continue
			}
			moves++
			total := 0
			for _, s := range p.ExileIDs {
				total += values[uuid.MustParse(s)]
			}
			if total < 8 {
				short = true
			}
		}
		return moves, short
	}
	if n, _ := collecting(); n != 0 {
		t.Errorf("%d moves collect evidence 8 over a graveyard of 7", n)
	}
	id := gyCard(me, "More Evidence", "Sorcery", "{1}", "")
	values[id] = 1
	n, short := collecting()
	if n == 0 {
		t.Fatal("no move collects evidence 8 over a graveyard of 8")
	}
	if short {
		t.Error("a move names cards worth less than 8")
	}
}

// --- Behind the Mask ------------------------------------------------------

func TestBehindTheMaskSetsFourThreeOrWithEvidenceOneOne(t *testing.T) {
	for _, tc := range []struct {
		collected bool
		p, tough  int
	}{{false, 4, 3}, {true, 1, 1}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rock := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
		params := game.CastSpellParams{Targets: []game.TargetRef{cardRef(rock)}}
		if tc.collected {
			params = mkmCECollecting(params.Targets, mkmCEEvidence(me, "{6}")...)
		}
		if _, err := mkmCECast(t, g, "Behind the Mask", "Instant", mkmCEMaskOracle, params); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		ch := effectiveOf(t, g, rock)
		if !slices.Contains(ch.Types, "Creature") || !slices.Contains(ch.Types, "Artifact") {
			t.Errorf("collected=%v: the target is not an artifact creature: %+v", tc.collected, ch.Types)
		}
		if p, tough := effectivePTOf(t, g, rock); p != tc.p || tough != tc.tough {
			t.Errorf("collected=%v: %d/%d, want %d/%d", tc.collected, p, tough, tc.p, tc.tough)
		}
	}
}

// --- Bite Down on Crime ---------------------------------------------------

// "This spell costs {2} less to cast if evidence was collected": the
// discount reads the announcement.
func TestBiteDownOnCrimeCostsTwoLessWithEvidence(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{InstanceID: uuid.New(), Name: "Bite Down on Crime", TypeLine: "Sorcery", ManaCost: "{3}{G}",
		OracleID: mkmCEBiteOracle, Owner: me.ID, Controller: me.ID}
	base, err := game.ParseCost(card.ManaCost)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		optional []int
		want     int
	}{{nil, 4}, {[]int{0}, 2}} {
		out, err := g.ApplyCostModifiers(base, game.CostQuery{Card: card, Controller: me.ID, FromZone: game.ZoneHand, OptionalCosts: tc.optional})
		if err != nil {
			t.Fatal(err)
		}
		if out.ManaValue() != tc.want {
			t.Errorf("optional %v: price %d, want %d", tc.optional, out.ManaValue(), tc.want)
		}
	}
}

func TestBiteDownOnCrimePumpsThenBites(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Giant", 4, 4)
	if _, err := mkmCECast(t, g, "Bite Down on Crime", "Sorcery", mkmCEBiteOracle,
		mkmCECollecting([]game.TargetRef{cardRef(mine), {Kind: game.TargetCard, ID: theirs, Slot: 1}}, mkmCEEvidence(me, "{3}", "{3}")...)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("a 4/2 bite did not kill the 4/4")
	}
	if p, _ := effectivePTOf(t, g, mine); p != 4 {
		t.Errorf("power %d, want 4 (+2/+0)", p)
	}
}

// --- Crimestopper Sprite --------------------------------------------------

func TestCrimestopperSpriteTapsAndWithEvidenceStuns(t *testing.T) {
	for _, collected := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
		params := game.CastSpellParams{}
		if collected {
			params = mkmCECollecting(nil, mkmCEEvidence(me, "{2}", "{4}")...)
		}
		if _, err := mkmCECast(t, g, "Crimestopper Sprite", "Creature — Faerie Detective", mkmCESpriteOracle, params); err != nil {
			t.Fatalf("collected=%v: cast: %v", collected, err)
		}
		passPriorityAroundTable(t, g)
		answerPickTargetIfOpen(t, g, victim)
		passPriorityAroundTable(t, g)
		if !isTapped(g, victim) {
			t.Errorf("collected=%v: the target is not tapped", collected)
		}
		want := 0
		if collected {
			want = 1
		}
		if got := countersOn(g, victim, game.CounterStun); got != want {
			t.Errorf("collected=%v: %d stun counters, want %d", collected, got, want)
		}
	}
}

// --- Extract a Confession -------------------------------------------------

func TestExtractAConfessionWithEvidenceTakesTheGreatestPower(t *testing.T) {
	for _, collected := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		small := pushVanillaCreature(g, opp.ID, "Their Elf", 1, 1)
		big := pushVanillaCreature(g, opp.ID, "Their Giant", 5, 5)
		params := game.CastSpellParams{}
		if collected {
			params = mkmCECollecting(nil, mkmCEEvidence(me, "{6}")...)
		}
		if _, err := mkmCECast(t, g, "Extract a Confession", "Sorcery", mkmCEConfessOracle, params); err != nil {
			t.Fatalf("collected=%v: cast: %v", collected, err)
		}
		passPriorityAroundTable(t, g)
		// Without evidence the opponent picks between the two; with it the
		// 5/5 is the only creature they may sacrifice, so nothing is asked.
		pick := chooseCardsChoiceFor(g, opp.ID)
		if !collected {
			if pick == nil || len(pick.ChooseCards) != 2 {
				t.Fatalf("collected=false: the opponent is not offered both creatures: %+v", pick)
			}
			answerChooseCards(t, g, opp.ID, small)
		} else if pick != nil {
			if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != big {
				t.Fatalf("collected=true: offered %v, want the 5/5 only", pick.ChooseCards)
			}
			answerChooseCards(t, g, opp.ID, big)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(big) != !collected || g.Battlefield.Contains(small) != collected {
			t.Errorf("collected=%v: the 5/5 survives %v, the 1/1 survives %v", collected, g.Battlefield.Contains(big), g.Battlefield.Contains(small))
		}
	}
}

// --- Vitu-Ghazi Inspector -------------------------------------------------

func TestVituGhaziInspectorTriggersOnlyIfEvidenceWasCollected(t *testing.T) {
	for _, collected := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
		life := me.Life
		params := game.CastSpellParams{}
		if collected {
			params = mkmCECollecting(nil, mkmCEEvidence(me, "{1}", "{5}")...)
		}
		if _, err := mkmCECast(t, g, "Vitu-Ghazi Inspector", "Creature — Elf Detective", mkmCEInspectorOracle, params); err != nil {
			t.Fatalf("collected=%v: cast: %v", collected, err)
		}
		passPriorityAroundTable(t, g)
		answerPickTargetIfOpen(t, g, bear)
		passPriorityAroundTable(t, g)
		wantCounters, wantLife := 0, life
		if collected {
			wantCounters, wantLife = 1, life+2
		}
		if got := countersOn(g, bear, game.CounterPlusOne); got != wantCounters {
			t.Errorf("collected=%v: %d +1/+1 counters, want %d", collected, got, wantCounters)
		}
		if me.Life != wantLife {
			t.Errorf("collected=%v: life %d, want %d", collected, me.Life, wantLife)
		}
	}
}

// --- Forensic Researcher --------------------------------------------------

func TestForensicResearcherTapsWithEvidenceThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	researcher := pushCatalogPermanent(g, me.ID, "Forensic Researcher", "Creature — Merfolk Detective", mkmCEResearcherOracle, false)
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	gy := mkmCEEvidence(me, "{1}", "{1}", "{2}")
	activate := func(ids ...uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, researcher, 1, game.ActivateAbilityParams{
			Targets: []game.TargetRef{cardRef(victim)}, ExileIDs: ids})
	}
	if err := activate(gy[0], gy[1]); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("collect evidence 3 with a total of 2: %v, want ErrInvalidParam", err)
	}
	if isTapped(g, researcher) || me.Graveyard.Size() != 3 {
		t.Fatal("a refused activation paid part of its cost")
	}
	if err := activate(gy[0], gy[2]); err != nil {
		t.Fatalf("collect evidence 3 with exactly 3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !isTapped(g, victim) {
		t.Error("the target is not tapped")
	}
	if !inExile(g, gy[0]) || !inExile(g, gy[2]) || !me.Graveyard.Contains(gy[1]) {
		t.Error("only the two named cards are exiled")
	}
}

// The view and the bot read the same gate: the row ships the floor, and
// the enumerator offers the ability only with enough in the graveyard.
func TestForensicResearcherViewAndEnumeratorReadTheGate(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	researcher := pushCatalogPermanent(g, me.ID, "Forensic Researcher", "Creature — Merfolk Detective", mkmCEResearcherOracle, false)
	pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	gy := mkmCEEvidence(me, "{2}")

	row := activatedRowOf(t, g, me.ID, researcher, 1)
	if !row.ExileCostAnyNumber || row.ExileCostManaValueMin != 3 || row.ExileCostZone != "graveyard" {
		t.Errorf("row %+v, want an any-number graveyard exile with a mana value floor of 3", row)
	}
	if row.CantActivate == "" {
		t.Error("a graveyard totalling 2 is not greyed")
	}
	offered := func() []uuid.UUID {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != "activate_ability" || m.Source != researcher {
				continue
			}
			var p struct {
				Index    int      `json:"ability_index"`
				ExileIDs []string `json:"exile_ids"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			if p.Index != 1 {
				continue
			}
			var ids []uuid.UUID
			for _, s := range p.ExileIDs {
				ids = append(ids, uuid.MustParse(s))
			}
			return ids
		}
		return nil
	}
	if ids := offered(); ids != nil {
		t.Errorf("the enumerator offered collect evidence 3 over a total of 2: %v", ids)
	}
	gy = append(gy, gyCard(me, "More Evidence", "Sorcery", "{1}", ""))
	ids := offered()
	if len(ids) != 2 || !slices.Contains(ids, gy[0]) || !slices.Contains(ids, gy[1]) {
		t.Fatalf("the enumerator offered %v, want both cards (total 3)", ids)
	}
	if row := activatedRowOf(t, g, me.ID, researcher, 1); row.CantActivate != "" {
		t.Errorf("a graveyard totalling 3 is greyed: %q", row.CantActivate)
	}
}

// --- Hedge Whisperer ------------------------------------------------------

func TestHedgeWhispererAnimatesALandWhileItStaysTapped(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	whisperer := pushCatalogPermanent(g, me.ID, "Hedge Whisperer", "Creature — Elf Druid Detective", mkmCEWhispererOracle, false)
	land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	gy := mkmCEEvidence(me, "{4}")
	if err := g.ActivateCatalogAbility(me.ID, whisperer, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{cardRef(land)}, ExileIDs: gy}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	ch := effectiveOf(t, g, land)
	if !slices.Contains(ch.Types, "Creature") || !slices.Contains(ch.Types, "Land") ||
		!slices.Contains(ch.Subtypes, "Plant") || !slices.Contains(ch.Subtypes, "Boar") || !slices.Equal(ch.Colors, []string{"G"}) {
		t.Errorf("the land is not a Plant Boar creature land: %+v %+v", ch.Types, ch.Subtypes)
	}
	if p, tough := effectivePTOf(t, g, land); p != 5 || tough != 5 {
		t.Errorf("%d/%d, want 5/5", p, tough)
	}
	if !hasEffectiveKeyword(t, g, land, "haste") {
		t.Error("the land has no haste")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(whisperer) })
	if ch := effectiveOf(t, g, land); slices.Contains(ch.Types, "Creature") {
		t.Error("the land is still a creature after the Whisperer untapped")
	}
}

// --- Polygraph Orb --------------------------------------------------------

func TestPolygraphOrbPunishesWithEvidenceThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	orb := pushCatalogPermanent(g, me.ID, "Polygraph Orb", "Artifact", mkmCEOrbOracle, false)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	gy := mkmCEEvidence(me, "{1}", "{1}")
	if err := g.ActivateCatalogAbility(me.ID, orb, 0, game.ActivateAbilityParams{ExileIDs: gy}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("collect evidence 3 with 2: %v, want ErrInvalidParam", err)
	}
	gy = append(gy, gyCard(me, "More Evidence", "Sorcery", "{1}", ""))
	if err := g.ActivateCatalogAbility(me.ID, orb, 0, game.ActivateAbilityParams{ExileIDs: gy}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the opponent is not asked: %+v", g.PendingChoices)
	}
	labels := map[string]bool{}
	for _, o := range pick.PickOptions {
		labels[o.Label] = true
	}
	if !labels["Lose 3 life"] || !labels["Sacrifice a creature"] || !labels["Discard a card"] {
		t.Fatalf("options %+v, want lose 3 / sacrifice a creature / discard", pick.PickOptions)
	}
	for i, o := range pick.PickOptions {
		if o.Label == "Sacrifice a creature" {
			answerOptionPick(t, g, opp.ID, i)
		}
	}
	sac := latestChooseCardsFor(g, opp.ID)
	if sac == nil {
		t.Fatal("the sacrifice branch asks nothing")
	}
	if err := g.ResolveChooseCards(sac.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
}

func TestPolygraphOrbEntersTakingTwoAndLosingTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life, hand, grave := me.Life, me.Hand.Size(), me.Graveyard.Size()
	if _, err := mkmCECast(t, g, "Polygraph Orb", "Artifact", mkmCEOrbOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick from the top four")
	}
	answerChooseCards(t, g, me.ID, pick.ChooseCards[:2]...)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || me.Graveyard.Size() != grave+2 {
		t.Errorf("hand %d (want %d), graveyard %d (want %d)", me.Hand.Size(), hand+2, me.Graveyard.Size(), grave+2)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

// --- the effect form: Izoni and Sample Collector --------------------------

// mkmCEAttack declares `attacker` and passes until its attack trigger has
// resolved.
func mkmCEAttack(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

func spidersOf(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Spider" && c.IsToken() {
			n++
		}
	}
	return n
}

func TestIzoniCollectsEvidenceForTwoSpiders(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	izoni := pushCatalogPermanent(g, me.ID, "Izoni, Center of the Web", "Legendary Creature — Elf Detective", mkmCEIzoniOracle, false)
	gy := mkmCEEvidence(me, "{1}", "{2}", "{3}")
	mkmCEAttack(t, g, izoni, opp.ID)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 0 || pick.ChooseMax != 3 {
		t.Fatalf("prompt %+v, want any number of the three, none included", pick)
	}
	// Short of 4 is refused, and nothing moves.
	if err := g.ResolveChooseCards(pick.ID, me.ID, gy[:2]); err == nil {
		t.Fatal("collect evidence 4 with a total of 3 was accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{gy[0], gy[2]}); err != nil {
		t.Fatalf("collect evidence 4 with exactly 4: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inExile(g, gy[0]) || !inExile(g, gy[2]) || !me.Graveyard.Contains(gy[1]) {
		t.Error("only the two named cards are exiled")
	}
	if n := spidersOf(g, me.ID); n != 2 {
		t.Errorf("%d Spiders, want 2", n)
	}
}

func TestIzoniDeclinedOrShortCreatesNothing(t *testing.T) {
	t.Run("declined", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		izoni := pushCatalogPermanent(g, me.ID, "Izoni, Center of the Web", "Legendary Creature — Elf Detective", mkmCEIzoniOracle, false)
		mkmCEEvidence(me, "{5}")
		mkmCEAttack(t, g, izoni, opp.ID)
		answerChooseCards(t, g, me.ID)
		passPriorityAroundTable(t, g)
		if n := spidersOf(g, me.ID); n != 0 || me.Graveyard.Size() != 1 {
			t.Errorf("declining made %d Spiders and left %d in the graveyard", n, me.Graveyard.Size())
		}
	})
	t.Run("the graveyard is short", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		izoni := pushCatalogPermanent(g, me.ID, "Izoni, Center of the Web", "Legendary Creature — Elf Detective", mkmCEIzoniOracle, false)
		mkmCEEvidence(me, "{1}", "{2}")
		mkmCEAttack(t, g, izoni, opp.ID)
		if pick := chooseCardsChoiceFor(g, me.ID); pick != nil {
			t.Errorf("asked with a total of 3: %+v", pick)
		}
	})
}

func TestIzoniSacrificesFourTokensToSurveilDrawAndGain(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	izoni := pushCatalogPermanent(g, me.ID, "Izoni, Center of the Web", "Legendary Creature — Elf Detective", mkmCEIzoniOracle, false)
	var tokens []uuid.UUID
	for range 4 {
		tokens = append(tokens, pushToken(g, me.ID, ClueToken()))
	}
	life, hand := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, izoni, 0, game.ActivateAbilityParams{SacrificeIDs: tokens}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || me.Life != life+2 {
		t.Errorf("hand %d → %d, life %d → %d; want +2 cards and +2 life", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestSampleCollectorCollectsEvidenceThenCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	collector := pushCatalogPermanent(g, me.ID, "Sample Collector", "Creature — Troll Detective", mkmCECollectorOracle, false)
	gy := mkmCEEvidence(me, "{3}")
	mkmCEAttack(t, g, collector, opp.ID)
	answerChooseCards(t, g, me.ID, gy...)
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, collector)
	passPriorityAroundTable(t, g)
	if !inExile(g, gy[0]) {
		t.Error("the evidence was not exiled")
	}
	if n := countersOn(g, collector, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
}
