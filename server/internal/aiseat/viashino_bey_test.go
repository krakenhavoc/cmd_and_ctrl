package aiseat_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// viashino_bey_test.go — the bot half of #5840. "If this creature
// attacks, all creatures you control attack if able" makes an attack
// set that includes the Bey owe every other creature that can attack,
// and leaves a set without it free. The enumerator is the bot's whole
// world, so the bot can only end a declaration the engine accepts: once
// the Bey is declared the pass is withheld and the others' attacks are
// the always-legal answers. Each policy below must finish the combat
// (driveOneCombat fails on a refused move, a decline by the priority
// holder, or a combat that never ends), and whenever the Bey attacked,
// so did everything else.

const botViashinoBeyOracle = "d9c9b933-74f4-4898-bbb8-2da76b86ccda"

// beyFirst attacks with the Bey and nothing else it is free to leave
// home: the Bey's attack when offered, then the pass, and while the pass
// is withheld the always-legal answer the runner falls back to on a
// decline (FallbackDeclineAlwaysLegal).
type beyFirst struct{ bey uuid.UUID }

func (beyFirst) Name() string { return "bey-first" }

func (p beyFirst) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	for i, m := range in.Moves {
		if m.Kind == legal.KindAttack && m.Source == p.bey {
			return aiseat.Decision{Index: i}, nil
		}
	}
	if pi := aiseat.PassIndex(in.Moves); pi >= 0 {
		return aiseat.Decision{Index: pi}, nil
	}
	if si := aiseat.SafeIndex(in.Moves); si >= 0 {
		return aiseat.Decision{Index: si}, nil
	}
	return aiseat.Decision{Index: aiseat.Decline}, nil
}

func TestBotsAttackWithEverythingOrWithoutViashinoBey(t *testing.T) {
	for _, name := range []string{"heuristic", "aggressive", "bey-first"} {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 4)
			active := g.Seats[g.Turn.ActiveSeat]
			bey := limitPush(g, active, "Viashino Bey", "Creature — Lizard", botViashinoBeyOracle, 4, 3)
			// A raw push leaves the layers as they were; the zone move
			// is what a real arrival would announce, and it marks them
			// stale so the Bey's static reaches the Ogres.
			g.WithWriteLock(func() {
				g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: bey, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
				for i := range g.Battlefield.Cards {
					if c := &g.Battlefield.Cards[i]; c.InstanceID == bey {
						c.SummonedThisTurn = false
					}
				}
			})
			var pol aiseat.Policy
			switch name {
			case "heuristic":
				pol = heuristic.New()
			case "aggressive":
				pol = aggressive()
			default:
				pol = beyFirst{bey: bey}
			}
			var others []uuid.UUID
			for i := range g.Battlefield.Cards {
				c := &g.Battlefield.Cards[i]
				if c.Controller == active.ID && c.IsCreature() && c.InstanceID != bey {
					others = append(others, c.InstanceID)
				}
			}
			attacked := map[uuid.UUID]bool{}
			driveOneCombatWatching(t, g, pol, func() {
				for i := range g.Battlefield.Cards {
					c := &g.Battlefield.Cards[i]
					if c.AttackingTarget != uuid.Nil {
						attacked[c.InstanceID] = true
					}
				}
			})
			if name == "bey-first" && !attacked[bey] {
				t.Fatal("the bey-first policy never attacked with the Bey")
			}
			if !attacked[bey] {
				return
			}
			for _, id := range others {
				if !attacked[id] {
					t.Errorf("the Bey attacked but %s stayed home", id)
				}
			}
		})
	}
}
