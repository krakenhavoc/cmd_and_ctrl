package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b126_cards_test.go — Commander 2020 cards, S82 OLD batch 126
// (#5513), played through the real catalog entries.

const (
	oldB126TributeOracle    = "ba0fc2ca-7084-454a-9d4e-5cbca945a7ca"
	oldB126AbsentOracle     = "e8d78a83-c932-4b55-8f75-7094c672c3a9"
	oldB126CharmOracle      = "418bc6c8-972b-4f07-aac0-8d003b9c1b17"
	oldB126FluctuatorOracle = "5ed0c7cc-9969-4387-90f8-3084dd18026b"
	oldB126ResonanceOracle  = "93255c68-2ed1-499b-82c6-672823451ce9"
	oldB126SeacliffOracle   = "a37544b6-0048-4213-8e40-76ba8a0b6d1b"
	oldB126BladeOracle      = "ab0dd461-25bc-4374-bed8-72fc2d51b5e1"
	oldB126FrolickerOracle  = "c0a18a53-d520-4795-9b3b-0f93bcf35848"
	oldB126NiblisOracle     = "ed2f3339-5c3e-473f-b08e-58288c4ad9c2"
	oldB126TuskerOracle     = "0a994401-d37c-4e30-9410-422e7ae5a715"
)

func TestOldB126CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB126TributeOracle:    "Tribute to the Wild",
		oldB126AbsentOracle:     "Unexpectedly Absent",
		oldB126CharmOracle:      "Temur Charm",
		oldB126FluctuatorOracle: "Fluctuator",
		oldB126ResonanceOracle:  "Crystalline Resonance",
		oldB126SeacliffOracle:   "Soaring Seacliff",
		oldB126BladeOracle:      "Sanctuary Blade",
		oldB126FrolickerOracle:  "Eon Frolicker",
		oldB126NiblisOracle:     "Niblis of Frost",
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

func oldB126Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB126Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- Tribute to the Wild -----------------------------------------------------

func TestOldB126TributeMakesEachOpponentSacrificeAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB126Seats(g)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Relic", "Artifact", "", false)
	relic := effectstest.PushCatalogPermanent(g, opp.ID, "Their Relic", "Artifact", "", false)
	aura := effectstest.PushCatalogPermanent(g, opp.ID, "Their Shrine", "Enchantment", "", false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Tribute to the Wild", "Instant", oldB126TributeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, aura)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, aura); ok {
		t.Error("the chosen enchantment was not sacrificed")
	}
	if _, ok := effectstest.BattlefieldCard(g, relic); !ok {
		t.Error("the unchosen artifact was sacrificed too")
	}
	if _, ok := effectstest.BattlefieldCard(g, mine); !ok {
		t.Error("the caster sacrificed their own artifact")
	}
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("a creature was sacrificed")
	}
}

// --- Unexpectedly Absent -----------------------------------------------------

func TestOldB126UnexpectedlyAbsentTucksBeneathTheTopXCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB126Seats(g)
	for i := 0; i < 5; i++ {
		effectstest.TopOfLibrary(opp, "Filler", "Sorcery", "{1}")
	}
	victim := effectstest.PushCatalogPermanent(g, opp.ID, "Their Relic", "Artifact", "", false)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Their Island", "Basic Land — Island", "", false)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Unexpectedly Absent", "Instant", oldB126AbsentOracle, oldB126Card(land)); err == nil {
		t.Fatal("Unexpectedly Absent targeted a land")
	}
	effectstest.Pr8CastX(t, g, "Unexpectedly Absent", "Instant", oldB126AbsentOracle, 2, oldB126Card(victim))
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Fatal("the permanent is still on the battlefield")
	}
	// Top to bottom: two cards, then the permanent.
	var pos = -1
	g.ReadSnapshot(func() {
		n := len(opp.Library.Cards)
		for i, c := range opp.Library.Cards {
			if c.InstanceID == victim {
				pos = n - 1 - i
			}
		}
	})
	if pos != 2 {
		t.Errorf("the permanent is %d cards from the top, want 2 (just beneath the top 2)", pos)
	}
	_ = me
}

// --- Temur Charm -----------------------------------------------------------

func TestOldB126TemurCharmPumpsAndFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB126Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB126CharmOracle, []int{0}, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	// 3/3 after the pump deals 3 to the 3/3 and takes 3.
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("the opposing 3/3 survived a 3-power fight")
	}
	if _, ok := effectstest.BattlefieldCard(g, mine); ok {
		t.Error("the pumped 3/3 survived 3 damage")
	}
}

func TestOldB126TemurCharmStopsSmallCreaturesBlockingEvenLateArrivals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB126Seats(g)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 4, 4)
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB126CharmOracle, []int{2}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.B12Creature(g, opp.ID, "Late", "Creature — Bear", 1, 1)
	restricted := func(id uuid.UUID) (r bool) {
		g.ReadSnapshot(func() { r = effectstest.BattlefieldCardFor(g, id).Effective().Restrictions&game.CantBlock != 0 })
		return r
	}
	if !restricted(small) || !restricted(late) {
		t.Error("a creature with power 3 or less can still block")
	}
	if restricted(big) {
		t.Error("a 4-power creature can't block")
	}
	_ = me
}

// --- Fluctuator ------------------------------------------------------------

func TestOldB126FluctuatorMakesYourCyclingCheaper(t *testing.T) {
	for _, c := range []struct {
		name      string
		fluct     func(g *game.Game, me, opp *game.Player)
		wantCycle bool
	}{
		{"mine", func(g *game.Game, me, _ *game.Player) {
			effectstest.PushCatalogPermanent(g, me.ID, "Fluctuator", "Artifact", oldB126FluctuatorOracle, false)
		}, true},
		{"theirs", func(g *game.Game, _, opp *game.Player) {
			effectstest.PushCatalogPermanent(g, opp.ID, "Fluctuator", "Artifact", oldB126FluctuatorOracle, false)
		}, false},
		{"none", func(*game.Game, *game.Player, *game.Player) {}, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := oldB126Seats(g)
			c.fluct(g, me, opp)
			effectstest.AdvanceToMain(t, g)
			id := effectstest.PushCatalogHandCard(me, "Warped Tusker", "Creature — Dinosaur", oldB126TuskerOracle)
			effectstest.FloatMana(t, g, me, "{G}")
			err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
			if (err == nil) != c.wantCycle {
				t.Errorf("cycling {2}{G} with only {G} floating: err = %v, want success %v", err, c.wantCycle)
			}
		})
	}
}

// --- Soaring Seacliff ------------------------------------------------------

func TestOldB126SoaringSeacliffEntersTappedAndGivesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB126Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.EnterFromHand(t, g, me.ID, "Soaring Seacliff", "Land", oldB126SeacliffOracle)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target creature did not gain flying")
	}
	if !effectstest.TappedOnBattlefield(t, g, land) {
		t.Error("Soaring Seacliff entered untapped")
	}
}

// --- Sanctuary Blade -------------------------------------------------------

func TestOldB126SanctuaryBladePumpsAndGivesProtectionFromTheLastChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB126Seats(g)
	blade := effectstest.PushCatalogPermanent(g, me.ID, "Sanctuary Blade", "Artifact — Equipment", oldB126BladeOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	attach := func(to uuid.UUID, color string) {
		t.Helper()
		effectstest.FloatMana(t, g, me, "{C}{C}{C}")
		if err := g.ActivateCatalogAbility(me.ID, blade, 0, game.ActivateAbilityParams{Targets: oldB126Card(to), Strict: true}); err != nil {
			t.Fatalf("equip: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerColor(t, g, me.ID, color)
		effectstest.PassPriorityAroundTable(t, g)
	}
	attach(bear, "R")
	if p, _ := effectstest.PtOf(t, g, bear); p != 4 {
		t.Errorf("equipped creature has power %d, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "protection from red") {
		t.Error("the equipped creature has no protection from the chosen color")
	}
	attach(other, "U")
	if effectstest.HasEffectiveKeyword(t, g, bear, "protection from red") || effectstest.HasEffectiveKeyword(t, g, bear, "protection from blue") {
		t.Error("the unequipped creature kept a grant")
	}
	if !effectstest.HasEffectiveKeyword(t, g, other, "protection from blue") || effectstest.HasEffectiveKeyword(t, g, other, "protection from red") {
		t.Error("the Blade did not use the newest chosen color")
	}
}

// --- Eon Frolicker ---------------------------------------------------------

func TestOldB126EonFrolickerGivesTheOpponentAnExtraTurnAndYouProtectionFromThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB126Seats(g)
	walker := effectstest.P1g2Planeswalker(g, me.ID, 4)
	theirs := effectstest.P1g2Planeswalker(g, opp.ID, 4)
	effectstest.CastHoldCreature(t, g, "Eon Frolicker", "Creature — Elemental Otter", oldB126FrolickerOracle, 5, 5)
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB126ExtraTurnsFor(g, opp); n != 1 {
		t.Errorf("%d extra turns queued for the opponent, want 1", n)
	}
	tok := game.ProtectionFromPlayer(opp.ID)
	if !effectstest.HasEffectiveKeyword(t, g, walker, tok) {
		t.Error("my planeswalker has no protection from the opponent")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, tok) {
		t.Error("their planeswalker got the protection")
	}
}

func TestOldB126EonFrolickerDoesNothingWhenNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB126Seats(g)
	effectstest.EnterFromHand(t, g, me.ID, "Eon Frolicker", "Creature — Elemental Otter", oldB126FrolickerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an Eon Frolicker that wasn't cast asked for a target")
	}
	if n := oldB126ExtraTurnsFor(g, opp); n != 0 {
		t.Errorf("%d extra turns queued", n)
	}
}

// --- Niblis of Frost -------------------------------------------------------

func TestOldB126NiblisTapsAndFreezesOnInstantsAndSorceriesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB126Seats(g)
	effectstest.B12Push(g, me.ID, "Niblis of Frost", "Creature — Spirit", oldB126NiblisOracle, 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Creature Spell", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a creature spell triggered Niblis of Frost")
	}
	effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.B35SettleTriggerOrders(t, g, me.ID)
	if !effectstest.TappedOnBattlefield(t, g, bear) {
		t.Error("the target creature was not tapped")
	}
	if c, _ := effectstest.BattlefieldCard(g, bear); len(c.NextUntapSkips) != 1 {
		t.Error("the target creature does not skip its next untap")
	}
}

func oldB126ExtraTurnsFor(g *game.Game, p *game.Player) int {
	n := 0
	for _, et := range g.ExtraTurnsQueued() {
		if et.Seat >= 0 && et.Seat < len(g.Seats) && g.Seats[et.Seat].ID == p.ID {
			n++
		}
	}
	return n
}
