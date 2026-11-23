package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const mirrormindCrownOracle = "353db389-b829-4e66-85d5-6018bc87c3e0"

func TestMirrormindCrownIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(mirrormindCrownOracle)
	if !ok || spec.Name != "Mirrormind Crown" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// mmcSetup parks the Crown and a tapped Golem with two +1/+1 counters,
// and equips the Golem through the Crown's own Equip {2} (strict, so an
// unpaid activation is refused first).
func mmcSetup(t *testing.T, g *game.Game, me *game.Player, equip bool) (crown, host uuid.UUID) {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	crown = effectstest.PushCatalogPermanent(g, me.ID, "Mirrormind Crown", "Artifact — Equipment", mirrormindCrownOracle, false)
	host = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Host Golem", TypeLine: "Artifact Creature — Golem",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	if !equip {
		return crown, host
	}
	params := game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(host)}
	if err := g.ActivateCatalogAbility(me.ID, crown, 0, params); err == nil {
		t.Fatal("equipped without paying {2}")
	}
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, crown, 0, params); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, crown); !ok || !c.IsAttachedTo(host) {
		t.Fatal("the Crown is not attached to the Golem")
	}
	effectstest.AddCountersForTest(g, host, game.CounterPlusOne, 2)
	effectstest.TapForTest(t, g, host)
	return crown, host
}

// The first creation this turn may become that many copies of the
// equipped creature: plain copies, untapped and without its counters.
// The second creation this turn is made as printed.
func TestMirrormindCrownTurnsTheFirstCreationIntoCopiesOfTheEquippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	_, host := mmcSetup(t, g, me, true)

	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 2)
	mm2AnswerOptional(t, g, me.ID, true)
	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 0 {
		t.Errorf("Treasures = %d, want none", got)
	}
	var copies []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Host Golem" && c.InstanceID != host {
				copies = append(copies, c)
			}
		}
	})
	if len(copies) != 2 {
		t.Fatalf("%d copies of the Golem, want 2", len(copies))
	}
	for _, c := range copies {
		if !effects.IsToken(c) || c.Tapped || c.Counters[game.CounterPlusOne] != 0 {
			t.Errorf("a copy is token=%v tapped=%v with %d counters; want an untapped token without counters",
				effects.IsToken(c), c.Tapped, c.Counters[game.CounterPlusOne])
		}
		if lc := effectstest.LayeredCard(t, g, c.InstanceID); lc.CurrentPower() != 3 || lc.CurrentToughness() != 3 {
			t.Errorf("a copy is %d/%d, want the printed 3/3", lc.CurrentPower(), lc.CurrentToughness())
		}
	}

	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 1)
	if p := effectstest.PendingOfKind(g, game.PendingChoiceOptionalReplacement); p != nil {
		t.Fatal("the second creation of the turn was offered the swap")
	}
	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
}

// Unattached, the Crown does nothing.
func TestMirrormindCrownDoesNothingUnattached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mmcSetup(t, g, me, false)
	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 1)
	if p := effectstest.PendingOfKind(g, game.PendingChoiceOptionalReplacement); p != nil {
		t.Fatal("an unattached Crown offered the swap")
	}
	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 1 {
		t.Errorf("Treasures = %d, want 1", got)
	}
}

// Declining makes the tokens as printed, and the turn's first time is
// used up; another player's creation isn't yours.
func TestMirrormindCrownDecliningAndOtherPlayersTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mmcSetup(t, g, me, true)

	effectstest.MakeTokens(t, g, opp.ID, effects.TreasureToken(), 1)
	if p := effectstest.PendingOfKind(g, game.PendingChoiceOptionalReplacement); p != nil {
		t.Fatal("an opponent's creation was offered the swap")
	}
	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 1)
	mm2AnswerOptional(t, g, me.ID, false)
	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 1)
	if p := effectstest.PendingOfKind(g, game.PendingChoiceOptionalReplacement); p != nil {
		t.Fatal("declining did not use up the turn's first time")
	}
	if got := effectstest.OnBattlefieldNamed(g, "Treasure"); got != 3 {
		t.Errorf("Treasures = %d, want 3", got)
	}
	if got := effectstest.OnBattlefieldNamed(g, "Host Golem"); got != 1 {
		t.Errorf("Host Golems = %d, want only the original", got)
	}
}
