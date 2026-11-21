package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func playerTargetableFrom(g *game.Game, by, id uuid.UUID, src *game.Characteristic) bool {
	var ok bool
	g.WithWriteLock(func() {
		ts := game.TargetSource{Controller: by, Snapshot: src}
		for _, p := range g.LegalTargetsForEffect(ts, effects.TargetPlayer("target player")).Players {
			if p == id {
				ok = true
			}
		}
	})
	return ok
}

func TestHexproofFromCardsAreComplete(t *testing.T) {
	for name, oracle := range map[string]string{
		"Eradicator Valkyrie": effectstest.EradicatorValkyrieOracle,
		"Breaker of Creation": effectstest.BreakerOfCreationOracle,
		"Dragonfire Blade":    effectstest.DragonfireBladeOracle,
		"Veil of Summer":      effectstest.CgVeilOfSummerOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: want CompletenessFull with no caveats, got %v %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestEradicatorValkyrieHasHexproofFromPlaneswalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	valkyrie := effectstest.PushCatalogPermanent(g, me.ID, "Eradicator Valkyrie", "Creature — Angel Berserker", effectstest.EradicatorValkyrieOracle, false)

	if !effectstest.EffectiveAbilitiesContain(t, g, valkyrie, game.HexproofFromPlaneswalkers) {
		t.Fatalf("abilities %v lack %q", effectstest.EffectiveAbilities(t, g, valkyrie), game.HexproofFromPlaneswalkers)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, valkyrie, "hexproof") {
		t.Error("it is hexproof from planeswalkers only, not plain hexproof (#259)")
	}
	if effectstest.CreatureTargetableFrom(g, opp.ID, valkyrie, effectstest.SrcWalker) {
		t.Error("an opponent's planeswalker ability must not target it")
	}
	if !effectstest.CreatureTargetableFrom(g, opp.ID, valkyrie, effectstest.SrcBlack) {
		t.Error("an opponent's non-planeswalker source still can")
	}
	if !effectstest.CreatureTargetableFrom(g, me.ID, valkyrie, effectstest.SrcWalker) {
		t.Error("its own controller's planeswalkers still can")
	}
}

func TestBreakerOfCreationHasHexproofFromEachColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	breaker := effectstest.PushCatalogPermanent(g, me.ID, "Breaker of Creation", "Creature — Eldrazi", effectstest.BreakerOfCreationOracle, false)

	for name, src := range map[string]*game.Characteristic{"red": effectstest.SrcRed, "blue": effectstest.SrcBlue, "gold": effectstest.SrcGold, "black": effectstest.SrcBlack} {
		if effectstest.CreatureTargetableFrom(g, opp.ID, breaker, src) {
			t.Errorf("an opponent's %s source must not target it", name)
		}
	}
	if !effectstest.CreatureTargetableFrom(g, opp.ID, breaker, effectstest.SrcColourless) {
		t.Error("a colourless source still can")
	}
	if !effectstest.CreatureTargetableFrom(g, me.ID, breaker, effectstest.SrcRed) {
		t.Error("its own controller's coloured spells still can")
	}
}

func TestDragonfireBladeGrantsHexproofFromMonocolored(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Dragonfire Blade", effectstest.DragonfireBladeOracle)

	if !effectstest.CreatureTargetableFrom(g, opp.ID, bear, effectstest.SrcRed) {
		t.Fatal("baseline: an unequipped bear is targetable")
	}
	effectstest.EquipTo(t, g, me.ID, blade, bear)

	if effectstest.CreatureTargetableFrom(g, opp.ID, bear, effectstest.SrcRed) {
		t.Error("an opponent's monocolored source must not target the equipped creature")
	}
	if !effectstest.CreatureTargetableFrom(g, opp.ID, bear, effectstest.SrcGold) {
		t.Error("a multicolored source still can")
	}
	if !effectstest.CreatureTargetableFrom(g, opp.ID, bear, effectstest.SrcColourless) {
		t.Error("a colourless source still can")
	}
	if !effectstest.CreatureTargetableFrom(g, me.ID, bear, effectstest.SrcRed) {
		t.Error("the controller's own spells still can")
	}
}

// Veil of Summer: you and permanents you control gain hexproof from
// blue and from black until end of turn, and the grant ends with it.
func TestVeilOfSummerGrantsHexproofFromBlueAndBlackToYouAndYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)

	effectstest.CastCatalogSpell(t, g, "Veil of Summer", "Instant", effectstest.CgVeilOfSummerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for name, src := range map[string]*game.Characteristic{"blue": effectstest.SrcBlue, "black": effectstest.SrcBlack, "blue-black": effectstest.SrcGold} {
		if effectstest.CreatureTargetableFrom(g, opp.ID, bear, src) {
			t.Errorf("an opponent's %s source must not target your permanent", name)
		}
		if playerTargetableFrom(g, opp.ID, me.ID, src) {
			t.Errorf("an opponent's %s source must not target you", name)
		}
	}
	if !effectstest.CreatureTargetableFrom(g, opp.ID, bear, effectstest.SrcRed) || !playerTargetableFrom(g, opp.ID, me.ID, effectstest.SrcRed) {
		t.Error("a red source still can: it is hexproof from blue and black only")
	}
	if !effectstest.CreatureTargetableFrom(g, me.ID, bear, effectstest.SrcBlue) || !playerTargetableFrom(g, me.ID, me.ID, effectstest.SrcBlue) {
		t.Error("your own blue spells still can")
	}
	if !effectstest.CreatureTargetableFrom(g, me.ID, theirs, effectstest.SrcBlue) {
		t.Error("only your permanents are covered, not theirs")
	}
	if !effectstest.CreatureTargetableFrom(g, opp.ID, theirs, effectstest.SrcBlue) {
		t.Error("their own permanent is untouched")
	}
	// CR 611.2c: the permanents are the ones you controlled as it resolved.
	late := effectstest.SeedBear(g, me.ID)
	if !effectstest.CreatureTargetableFrom(g, opp.ID, late, effectstest.SrcBlue) {
		t.Error("a permanent that arrives after the spell resolves is not covered")
	}

	// And it ends with the turn.
	startSeat := g.Turn.ActiveSeat
	for i := 0; i < 30 && g.Turn.ActiveSeat == startSeat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if !effectstest.CreatureTargetableFrom(g, opp.ID, bear, effectstest.SrcBlue) || !playerTargetableFrom(g, opp.ID, me.ID, effectstest.SrcBlue) {
		t.Error("the grant must end at cleanup")
	}
}
