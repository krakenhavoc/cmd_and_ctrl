package testsm

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

const (
	mavindaOracle = "1c248187-0be3-4a03-a817-14f68f638ae7"
	shockOracle   = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
)

// mavindaGrantsShock activates Mavinda at a Shock in the active seat's
// graveyard and resolves it, with a 3/3 of the seat's own on the
// battlefield.
func mavindaGrantsShock(t *testing.T) (g *game.Game, me, opp *game.Player, mavinda, shock, mine uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	opp = g.Seats[1]
	mavinda = effectstest.Cmdr2B05Card(g, me.ID, "Mavinda, Students' Advocate", "Legendary Creature — Bird Advisor", mavindaOracle, 2, 3)
	mine = effectstest.AcPermanent(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	shock = effectstest.GraveyardCardOf(g, me, "Shock", "Instant", "{R}", shockOracle)
	effectstest.B16Activate(t, g, me.ID, mavinda, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: shock}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, opp, mavinda, shock, mine
}

func shockParams(target game.TargetRef) game.CastSpellParams {
	return game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true,
		Targets: []game.TargetRef{target},
	}
}

// CR 601.2f with the targets announced: aimed at your own creature, the
// Shock costs {R}; it resolves and is exiled instead of going to the
// graveyard.
func TestMavindaCastAtYourCreatureCostsTheManaCostAndIsExiled(t *testing.T) {
	g, me, _, _, shock, mine := mavindaGrantsShock(t)
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, shock, shockParams(game.TargetRef{Kind: game.TargetCard, ID: mine})); err != nil {
		t.Fatalf("Shock at my own creature for {R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if zone, _ := effectstest.CardWhere(g, me, shock); zone != "exile" {
		t.Errorf("the Shock is in %s, want exile", zone)
	}
}

// Aimed at a player it costs {8}{R}: {R} alone is refused, {8}{R} pays.
func TestMavindaCastAtAPlayerCostsEightMore(t *testing.T) {
	g, me, opp, _, shock, _ := mavindaGrantsShock(t)
	at := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	price, err := g.PriceCast(me.ID, effectstest.CardByID(g, shock), shockParams(at))
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if got := price.Total.String(); got != "{8}{R}" {
		t.Errorf("Shock at a player costs %s, want {8}{R}", got)
	}
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, shock, shockParams(at)); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("Shock at a player for {R}: %v, want ErrInsufficientMana", err)
	}
	effectstest.FloatForTest(g, me, "CCCCCCCC")
	life := opp.Life
	if err := g.CastSpell(me.ID, shock, shockParams(at)); err != nil {
		t.Fatalf("Shock at a player for {8}{R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-2)
	}
}

// An opponent's creature is not "a creature you control": still {8} more.
// The enumerator prices each target set the same way (#544).
func TestMavindaSurchargeReadsWhoControlsTheTarget(t *testing.T) {
	g, me, opp, mavinda, shock, mine := mavindaGrantsShock(t)
	theirs := effectstest.AcPermanent(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	price, err := g.PriceCast(me.ID, effectstest.CardByID(g, shock), shockParams(game.TargetRef{Kind: game.TargetCard, ID: theirs}))
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if got := price.Total.String(); got != "{8}{R}" {
		t.Errorf("Shock at an opponent's creature costs %s, want {8}{R}", got)
	}

	// With only {R}, the enumerator offers the Shock at my creatures and at
	// nothing else.
	effectstest.FloatForTest(g, me, "R")
	var offered []string
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind != legal.KindCast || m.Source != shock {
			continue
		}
		var p struct {
			Targets []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode: %v", err)
		}
		for _, tg := range p.Targets {
			offered = append(offered, tg.ID)
		}
	}
	// My creatures are the Bear and Mavinda herself.
	sawMine := false
	for _, id := range offered {
		switch id {
		case mine.String():
			sawMine = true
		case mavinda.String():
		default:
			t.Errorf("with {R} the Shock is offered at %s, which is not a creature I control", id)
		}
	}
	if !sawMine {
		t.Errorf("with {R} the Shock is not offered at my creature: %v", offered)
	}
}

// Activate only once each turn.
func TestMavindaActivatesOnlyOnceEachTurn(t *testing.T) {
	g, me, _, mavinda, _, _ := mavindaGrantsShock(t)
	other := effectstest.GraveyardCardOf(g, me, "Shock", "Instant", "{R}", shockOracle)
	err := g.ActivateCatalogAbility(me.ID, mavinda, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}},
	})
	if err == nil {
		t.Fatal("a second activation the same turn was accepted")
	}
}
