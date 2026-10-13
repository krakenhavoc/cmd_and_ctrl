package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hexproof_from_cards_test.go — #3172, the proof cards for "hexproof from
// [quality]" (ADR 0038's amendment of 2026-10-13): Eradicator Valkyrie,
// Breaker of Creation, Dragonfire Blade and Veil of Summer. The engine
// half (the grammar, the choke point, the player half) is in
// game/hexproof_from_test.go; these prove each card reaches it.

var (
	srcBlack      = &game.Characteristic{Colors: []string{"B"}, Types: []string{"Instant"}}
	srcRed        = &game.Characteristic{Colors: []string{"R"}, Types: []string{"Instant"}}
	srcBlue       = &game.Characteristic{Colors: []string{"U"}, Types: []string{"Instant"}}
	srcGold       = &game.Characteristic{Colors: []string{"U", "B"}, Types: []string{"Instant"}}
	srcColourless = &game.Characteristic{Types: []string{"Artifact"}}
	srcWalker     = &game.Characteristic{Colors: []string{"W"}, Types: []string{"Planeswalker"}}
)

// creatureTargetableFrom is creatureTargetableBy for a source with
// known characteristics: hexproof from [quality] asks the source, not
// just whose it is.
func creatureTargetableFrom(g *game.Game, by, id uuid.UUID, src *game.Characteristic) bool {
	var ok bool
	g.WithWriteLock(func() {
		ts := game.TargetSource{Controller: by, Snapshot: src}
		for _, c := range g.LegalTargetsForEffect(ts, TargetCreature("target creature")).Cards {
			if c == id {
				ok = true
			}
		}
	})
	return ok
}

func playerTargetableFrom(g *game.Game, by, id uuid.UUID, src *game.Characteristic) bool {
	var ok bool
	g.WithWriteLock(func() {
		ts := game.TargetSource{Controller: by, Snapshot: src}
		for _, p := range g.LegalTargetsForEffect(ts, TargetPlayer("target player")).Players {
			if p == id {
				ok = true
			}
		}
	})
	return ok
}

func TestHexproofFromCardsAreComplete(t *testing.T) {
	for name, oracle := range map[string]string{
		"Eradicator Valkyrie": eradicatorValkyrieOracle,
		"Breaker of Creation": breakerOfCreationOracle,
		"Dragonfire Blade":    dragonfireBladeOracle,
		"Veil of Summer":      cgVeilOfSummerOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: want CompletenessFull with no caveats, got %v %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestEradicatorValkyrieHasHexproofFromPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	valkyrie := pushCatalogPermanent(g, me.ID, "Eradicator Valkyrie", "Creature — Angel Berserker", eradicatorValkyrieOracle, false)

	if !effectiveAbilitiesContain(t, g, valkyrie, game.HexproofFromPlaneswalkers) {
		t.Fatalf("abilities %v lack %q", effectiveAbilities(t, g, valkyrie), game.HexproofFromPlaneswalkers)
	}
	if effectiveAbilitiesContain(t, g, valkyrie, "hexproof") {
		t.Error("it is hexproof from planeswalkers only, not plain hexproof (#259)")
	}
	if creatureTargetableFrom(g, opp.ID, valkyrie, srcWalker) {
		t.Error("an opponent's planeswalker ability must not target it")
	}
	if !creatureTargetableFrom(g, opp.ID, valkyrie, srcBlack) {
		t.Error("an opponent's non-planeswalker source still can")
	}
	if !creatureTargetableFrom(g, me.ID, valkyrie, srcWalker) {
		t.Error("its own controller's planeswalkers still can")
	}
}

func TestBreakerOfCreationHasHexproofFromEachColor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	breaker := pushCatalogPermanent(g, me.ID, "Breaker of Creation", "Creature — Eldrazi", breakerOfCreationOracle, false)

	for name, src := range map[string]*game.Characteristic{"red": srcRed, "blue": srcBlue, "gold": srcGold, "black": srcBlack} {
		if creatureTargetableFrom(g, opp.ID, breaker, src) {
			t.Errorf("an opponent's %s source must not target it", name)
		}
	}
	if !creatureTargetableFrom(g, opp.ID, breaker, srcColourless) {
		t.Error("a colourless source still can")
	}
	if !creatureTargetableFrom(g, me.ID, breaker, srcRed) {
		t.Error("its own controller's coloured spells still can")
	}
}

func TestDragonfireBladeGrantsHexproofFromMonocolored(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Dragonfire Blade", dragonfireBladeOracle)

	if !creatureTargetableFrom(g, opp.ID, bear, srcRed) {
		t.Fatal("baseline: an unequipped bear is targetable")
	}
	equipTo(t, g, me.ID, blade, bear)

	if creatureTargetableFrom(g, opp.ID, bear, srcRed) {
		t.Error("an opponent's monocolored source must not target the equipped creature")
	}
	if !creatureTargetableFrom(g, opp.ID, bear, srcGold) {
		t.Error("a multicolored source still can")
	}
	if !creatureTargetableFrom(g, opp.ID, bear, srcColourless) {
		t.Error("a colourless source still can")
	}
	if !creatureTargetableFrom(g, me.ID, bear, srcRed) {
		t.Error("the controller's own spells still can")
	}
}

// Veil of Summer: you and permanents you control gain hexproof from
// blue and from black until end of turn, and the grant ends with it.
func TestVeilOfSummerGrantsHexproofFromBlueAndBlackToYouAndYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	theirs := seedBear(g, opp.ID)

	castCatalogSpell(t, g, "Veil of Summer", "Instant", cgVeilOfSummerOracle, nil)
	passPriorityAroundTable(t, g)

	for name, src := range map[string]*game.Characteristic{"blue": srcBlue, "black": srcBlack, "blue-black": srcGold} {
		if creatureTargetableFrom(g, opp.ID, bear, src) {
			t.Errorf("an opponent's %s source must not target your permanent", name)
		}
		if playerTargetableFrom(g, opp.ID, me.ID, src) {
			t.Errorf("an opponent's %s source must not target you", name)
		}
	}
	if !creatureTargetableFrom(g, opp.ID, bear, srcRed) || !playerTargetableFrom(g, opp.ID, me.ID, srcRed) {
		t.Error("a red source still can: it is hexproof from blue and black only")
	}
	if !creatureTargetableFrom(g, me.ID, bear, srcBlue) || !playerTargetableFrom(g, me.ID, me.ID, srcBlue) {
		t.Error("your own blue spells still can")
	}
	if !creatureTargetableFrom(g, me.ID, theirs, srcBlue) {
		t.Error("only your permanents are covered, not theirs")
	}
	if !creatureTargetableFrom(g, opp.ID, theirs, srcBlue) {
		t.Error("their own permanent is untouched")
	}
	// CR 611.2c: the permanents are the ones you controlled as it resolved.
	late := seedBear(g, me.ID)
	if !creatureTargetableFrom(g, opp.ID, late, srcBlue) {
		t.Error("a permanent that arrives after the spell resolves is not covered")
	}

	// And it ends with the turn.
	startSeat := g.Turn.ActiveSeat
	for i := 0; i < 30 && g.Turn.ActiveSeat == startSeat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if !creatureTargetableFrom(g, opp.ID, bear, srcBlue) || !playerTargetableFrom(g, opp.ID, me.ID, srcBlue) {
		t.Error("the grant must end at cleanup")
	}
}
