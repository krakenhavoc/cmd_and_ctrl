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

const alexiosOracleID = "81c6c4f0-81ad-463c-a42b-60a73e6d70f3"

// alexiosIn seats Alexios for p, past summoning sickness.
func alexiosIn(g *game.Game, p *game.Player) uuid.UUID {
	id := pushCatalogPermanent(g, p.ID, "Alexios, Deimos of Kosmos", "Legendary Creature — Human Berserker", alexiosOracleID, false)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	return id
}

// offeredSacrifices lists every permanent the enumerator names in a
// sacrifice payment for `source`'s abilities.
func offeredSacrifices(t *testing.T, g *game.Game, seat, source uuid.UUID) map[uuid.UUID]bool {
	t.Helper()
	out := map[uuid.UUID]bool{}
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != "activate_ability" || m.Source != source {
			continue
		}
		var p struct {
			SacrificeIDs []string `json:"sacrifice_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		for _, s := range p.SacrificeIDs {
			if id, err := uuid.Parse(s); err == nil {
				out[id] = true
			}
		}
	}
	return out
}

// "Can't be sacrificed" (#3275) as a cost: Goblin Bombardment can't eat
// Alexios. The engine refuses the payment, and the enumerator offers the
// other creature only.
func TestAlexiosCantPayASacrificeCost(t *testing.T) {
	g, me, opp := flipMain(t)
	alexios := alexiosIn(g, me)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	bombardment := pushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, false)

	offered := offeredSacrifices(t, g, me.ID, bombardment)
	if offered[alexios] || !offered[bear] {
		t.Fatalf("enumerator offers %v, want the bear and not Alexios", offered)
	}

	err := g.ActivateCatalogAbility(me.ID, bombardment, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{alexios},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if !errors.Is(err, game.ErrCantBeSacrificed) {
		t.Fatalf("Bombardment paid with Alexios: %v, want ErrCantBeSacrificed", err)
	}
	if findBattlefieldCardForTest(g, alexios) == nil {
		t.Fatal("the refused payment sacrificed Alexios")
	}
}

// As an effect: an edict passes over Alexios, so a player whose only
// creature he is sacrifices nothing, and another player still sacrifices.
func TestAlexiosIsNotAnEdictChoice(t *testing.T) {
	g, me, opp := flipMain(t)
	alexios := alexiosIn(g, me)
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		if cands := g.EffectSacrificeCandidatesForEffect(me.ID, nil); slices.Contains(cands, alexios) {
			t.Errorf("edict candidates %v include Alexios", cands)
		}
		g.EachPlayerSacrificesForEffect(uuid.Nil, uuid.Nil, nil, "sacrifice a permanent")
	})
	tfSettle(t, g)
	if findBattlefieldCardForTest(g, alexios) == nil {
		t.Error("the edict took Alexios")
	}
	if findBattlefieldCardForTest(g, bear) != nil {
		t.Error("the opponent's lone bear survived the edict")
	}
}

// "At the beginning of each player's upkeep, that player gains control
// of Alexios, untaps it, and puts a +1/+1 counter on it. It gains haste
// until end of turn."
func TestAlexiosGoesToEachUpkeepsPlayer(t *testing.T) {
	g, me, opp := flipMain(t)
	alexios := alexiosIn(g, me)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == alexios {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	advanceToUpkeepOfSeat(t, g, next)
	tfSettle(t, g)

	c := flipState(t, g, alexios)
	if c.Controller != opp.ID {
		t.Fatalf("controller = %v, want the upkeep's player %v", c.Controller, opp.ID)
	}
	if c.Tapped {
		t.Error("Alexios should be untapped")
	}
	if c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", c.Counters[game.CounterPlusOne])
	}
	if !game.HasKeyword(&c, "haste") || !game.HasKeyword(&c, "trample") {
		t.Errorf("keywords = %v, want haste and trample", c.Keywords)
	}
	var cant bool
	g.WithWriteLock(func() { cant = g.CantBeSacrificedForEffect(alexios) })
	if !cant {
		t.Error("Alexios can be sacrificed under his new controller")
	}
}
