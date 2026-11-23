package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b207_test.go — cards missing from sets since AFR, batch 207 (#5948):
// Double Masters 2022 enchantments and instants. Each test plays the real
// catalog entry and asserts what its text says.

const (
	gap207AgonyWarp  = "377c217b-ddad-4e1a-9122-f347113d3dfa"
	gap207Myth       = "3a726e40-829c-47ec-81a7-58a04164fc3f"
	gap207Aria       = "6cb9ca93-e0bb-4758-bf0d-c9a8fce5f11f"
	gap207Legion     = "7b23bceb-1b82-4c9c-b7c1-8e8ef4643fb0"
	gap207Domestic   = "86da33ac-f8ac-48c8-a50b-22fd761dff8f"
	gap207Web        = "8b25f353-0798-4a03-a928-6a701a919468"
	gap207Krosa      = "b50a6db0-16ce-4e98-afbd-ec37e1653e7c"
	gap207Spree      = "c0c96c7a-02d8-40cb-a08b-4a2a3b631eef"
	gap207Ascendancy = "c3c802ba-dfbb-4f63-ae9a-e29c2e30013d"
	gap207Bounty     = "e6b20c30-88c2-4cb0-a5e4-46f1df924a7c"
)

func TestGap207EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap207AgonyWarp:  "Agony Warp",
		gap207Myth:       "Myth Realized",
		gap207Aria:       "Mayael's Aria",
		gap207Legion:     "Legion's Initiative",
		gap207Domestic:   "Domestication",
		gap207Web:        "Arachnus Web",
		gap207Krosa:      "Might of Old Krosa",
		gap207Spree:      "Auger Spree",
		gap207Ascendancy: "Jeskai Ascendancy",
		gap207Bounty:     "Bounty of the Luxa",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gap207Count(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

// --- Agony Warp --------------------------------------------------------------

func TestGap207AgonyWarpShrinksOnePowerAndAnotherToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	a := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	b := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.Ecl6Cast(t, g, me, "Agony Warp", "Instant", gap207AgonyWarp, 0, 0, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, a); p != 2 || tt != 5 {
		t.Errorf("the first target is %d/%d, want 2/5", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, b); p != 2 || tt != 1 {
		t.Errorf("the second target is %d/%d, want 2/1", p, tt)
	}
}

// --- Auger Spree -------------------------------------------------------------

func TestGap207AugerSpreeGivesPlusFourMinusFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	foe := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 2, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.Ecl6Cast(t, g, me, "Auger Spree", "Instant", gap207Spree, 0, 0, game.CastSpellParams{Targets: effectstest.B16TargetCard(foe)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, foe); p != 6 || tt != 1 {
		t.Errorf("the target is %d/%d, want 6/1", p, tt)
	}
}

// --- Might of Old Krosa ------------------------------------------------------

func TestGap207MightIsPlusFourInYourMainPhaseAndPlusTwoOtherwise(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.Ecl6Cast(t, g, me, "Might of Old Krosa", "Instant", gap207Krosa, 0, 0, game.CastSpellParams{Targets: effectstest.B16TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 6 || tt != 6 {
		t.Errorf("in the main phase the Bear is %d/%d, want 6/6", p, tt)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, _ := effectstest.Ecl6Seats(g2)
	bear2 := effectstest.B12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	effectstest.Ecl6Cast(t, g2, me2, "Might of Old Krosa", "Instant", gap207Krosa, 0, 0, game.CastSpellParams{Targets: effectstest.B16TargetCard(bear2)})
	effectstest.PassPriorityAroundTable(t, g2)
	if p, tt := effectstest.PtOf(t, g2, bear2); p != 4 || tt != 4 {
		t.Errorf("in combat the Bear is %d/%d, want 4/4", p, tt)
	}
}

// --- Domestication -----------------------------------------------------------

func gap207GiveCounters(g *game.Game, id uuid.UUID, n int) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: n}
			}
		}
	})
}

func TestGap207DomesticationStealsAndIsSacrificedAtMyEndStepOnceItsPowerIsFour(t *testing.T) {
	for _, counters := range []int{0, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.Ecl6Seats(g)
		foe := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
		effectstest.AdvanceToMain(t, g)
		aura := effectstest.Ecl6Cast(t, g, me, "Domestication", "Enchantment — Aura", gap207Domestic, 0, 0, game.CastSpellParams{Targets: effectstest.B16TargetCard(foe)})
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.ControllerOf(t, g, foe); got != me.ID {
			t.Fatalf("the Wolf is controlled by %s, want me", got)
		}
		gap207GiveCounters(g, foe, counters)
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		_, auraHere := effectstest.BattlefieldCard(g, aura)
		if counters == 0 && !auraHere {
			t.Error("the Aura was sacrificed with the creature's power under 4")
		}
		if counters == 2 {
			if auraHere {
				t.Error("the Aura stayed with the creature's power at 4")
			}
			if got := effectstest.ControllerOf(t, g, foe); got != opp.ID {
				t.Errorf("the Wolf is controlled by %s after the Aura left, want its owner", got)
			}
		}
	}
}

// --- Arachnus Web ------------------------------------------------------------

func TestGap207ArachnusWebRestrictsAndIsDestroyedAtAnyEndStepOnceTheHostHasFourPower(t *testing.T) {
	for _, counters := range []int{0, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.Ecl6Seats(g)
		foe := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
		effectstest.AdvanceToMain(t, g)
		aura := effectstest.Ecl6Cast(t, g, me, "Arachnus Web", "Enchantment — Aura", gap207Web, 0, 0, game.CastSpellParams{Targets: effectstest.B16TargetCard(foe)})
		effectstest.PassPriorityAroundTable(t, g)
		var r game.Restriction
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			r = effectstest.FindBattlefieldCardByID(g, foe).Effective().Restrictions
		})
		for _, want := range []game.Restriction{game.CantAttackOrBlock, game.CantActivate} {
			if r&want != want {
				t.Fatalf("restrictions %v lack %v", r, want)
			}
		}
		gap207GiveCounters(g, foe, counters)
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		_, auraHere := effectstest.BattlefieldCard(g, aura)
		if counters == 0 && !auraHere {
			t.Error("the Aura was destroyed with the creature's power under 4")
		}
		if counters == 2 && auraHere {
			t.Error("the Aura stayed with the creature's power at 4")
		}
	}
}

// --- Mayael's Aria -----------------------------------------------------------

func TestGap207AriaCountersLifeAndTheWinAreEachGatedOnPower(t *testing.T) {
	cases := []struct {
		power    int
		counters bool
		life     int
		wins     bool
		name     string
	}{
		{4, false, 0, false, "power 4 does nothing"},
		{5, true, 0, false, "power 5 gets counters"},
		{9, true, 10, false, "power 9 grows to 10 and gains life"},
		{19, true, 10, true, "power 19 grows to 20 and wins"},
	}
	for _, tc := range cases {
		g := effectstest.NewCatalogGame(t)
		me, _ := effectstest.Ecl6Seats(g)
		fat := effectstest.B12Creature(g, me.ID, "Fatty", "Creature — Giant", tc.power, tc.power)
		effectstest.PushCatalogPermanent(g, me.ID, "Mayael's Aria", "Enchantment", gap207Aria, false)
		life := me.Life
		effectstest.AdvanceToMain(t, g)
		effectstest.AdvanceToUpkeepOf(t, g, indexOf(g, me))
		effectstest.PassPriorityAroundTable(t, g)
		c, _ := effectstest.BattlefieldCard(g, fat)
		if got := c.Counters[game.CounterPlusOne] == 1; got != tc.counters {
			t.Errorf("%s: counters placed %v, want %v", tc.name, got, tc.counters)
		}
		if me.Life != life+tc.life {
			t.Errorf("%s: life %d, want %d", tc.name, me.Life, life+tc.life)
		}
		if won := g.State == game.StateEnded && g.Outcome != nil && g.Outcome.Kind == game.OutcomeWin && g.Outcome.Winner == me.ID; won != tc.wins {
			t.Errorf("%s: won %v, want %v", tc.name, won, tc.wins)
		}
	}
}

func indexOf(g *game.Game, p *game.Player) int {
	for i, s := range g.Seats {
		if s == p {
			return i
		}
	}
	return -1
}

// --- Jeskai Ascendancy -------------------------------------------------------

func TestGap207AscendancyPumpsAndUntapsThenMayLootOnANoncreatureCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Jeskai Ascendancy", "Enchantment", gap207Ascendancy, false)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	effectstest.B16Tap(g, mine)
	base := effectstest.EffectivePower(t, g, mine)
	hand := effectstest.HandCount(g, me)
	effectstest.CastCatalogSpell(t, g, "Filler", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4 && effectstest.TriggerOrderPrompt(g) != nil; i++ {
		effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerDiscard(t, g, me.ID, g.Seats[g.Turn.ActiveSeat].Hand.Cards[0].InstanceID)
	if got := effectstest.HandCount(g, me); got != hand {
		t.Errorf("hand %d after the loot, want %d (drew one, discarded one)", got, hand)
	}
	if p := effectstest.EffectivePower(t, g, mine); p != base+1 {
		t.Errorf("my creature has power %d, want %d", p, base+1)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != base {
		t.Errorf("an opponent's creature has power %d, want unchanged", p)
	}
	if effectstest.B16Tapped(t, g, mine) {
		t.Error("my creature stayed tapped")
	}
}

// --- Bounty of the Luxa ------------------------------------------------------

func TestGap207BountyAlternatesAFloodCounterAndADrawWithThreeMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	b := effectstest.PushCatalogPermanent(g, me.ID, "Bounty of the Luxa", "Enchantment", gap207Bounty, false)
	hand := effectstest.HandCount(g, me)
	effectstest.AdvanceToPrecombatMainOf(t, g, indexOf(g, me))
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, b); c.Counters["flood"] != 1 {
		t.Fatalf("%d flood counters after the first main phase, want 1", c.Counters["flood"])
	}
	if got := effectstest.HandCount(g, me); got != hand+2 {
		t.Errorf("hand %d, want %d (draw step plus one drawn)", got, hand+2)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, indexOf(g, me))
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, b); c.Counters["flood"] != 0 {
		t.Errorf("%d flood counters after the second, want them all removed", c.Counters["flood"])
	}
	if pool := len(me.ManaPool); pool < 3 {
		t.Errorf("the pool holds %d mana, want {C}{G}{U}", pool)
	}
}

// --- Myth Realized -----------------------------------------------------------

func TestGap207MythRealizedGrowsLoreAndAnimatesToItsCounterCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	m := effectstest.PushCatalogPermanent(g, me.ID, "Myth Realized", "Enchantment", gap207Myth, false)
	effectstest.CastCatalogSpell(t, g, "Filler", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "C", "C", "W")
	effectstest.B16Activate(t, g, me.ID, m, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, m); c.Counters[game.CounterLore] != 2 {
		t.Fatalf("%d lore counters, want 2 (one cast, one paid)", c.Counters[game.CounterLore])
	}
	if c, _ := effectstest.BattlefieldCard(g, m); c.IsCreature() {
		t.Fatal("a creature before it was animated")
	}
	effectstest.B06AddMana(me, "W")
	effectstest.B16Activate(t, g, me.ID, m, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, m)
	if p, tt := effectstest.PtOf(t, g, m); p != 2 || tt != 2 || !c.HasSubtype("Monk") || !c.HasSubtype("Avatar") {
		t.Fatalf("animated: %d/%d %v, want a 2/2 Monk Avatar", p, tt, c.Effective().Subtypes)
	}
	effectstest.B06AddMana(me, "C", "C", "W")
	effectstest.B16Activate(t, g, me.ID, m, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, m); p != 3 {
		t.Errorf("power %d after another lore counter, want 3", p)
	}
}

// --- Legion's Initiative -----------------------------------------------------

func TestGap207LegionsInitiativeAnthemsThenBlinksYourCreaturesBackAtCombatWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	redOne := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Thing", TypeLine: "Creature — Goblin",
		Colors: []string{"R"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	whiteOne := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "White Thing", TypeLine: "Creature — Human",
		Colors: []string{"W"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Red", TypeLine: "Creature — Goblin",
		Colors: []string{"R"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	leg := effectstest.PushCatalogPermanent(g, me.ID, "Legion's Initiative", "Enchantment", gap207Legion, false)
	if p, tt := effectstest.PtOf(t, g, redOne); p != 3 || tt != 2 {
		t.Errorf("my red creature is %d/%d, want 3/2", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, whiteOne); p != 2 || tt != 3 {
		t.Errorf("my white creature is %d/%d, want 2/3", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("their red creature is %d/%d, want 2/2", p, tt)
	}
	effectstest.B06AddMana(me, "R", "W")
	effectstest.B16Activate(t, g, me.ID, leg, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, redOne); ok {
		t.Fatal("my red creature was not exiled")
	}
	if _, ok := effectstest.BattlefieldCard(g, theirs); !ok {
		t.Fatal("an opponent's creature was exiled")
	}
	if n := gap207Count(g, "Legion's Initiative"); n != 0 {
		t.Fatalf("%d Legion's Initiative left (it exiles itself as a cost)", n)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"Red Thing", "White Thing"} {
		var id uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				id = c.InstanceID
			}
		}
		if id == uuid.Nil {
			t.Errorf("%s did not return at the beginning of combat", name)
			continue
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s has no haste", name)
		}
	}
	if gap207Count(g, "Legion's Initiative") != 0 {
		t.Error("the enchantment came back")
	}
}
