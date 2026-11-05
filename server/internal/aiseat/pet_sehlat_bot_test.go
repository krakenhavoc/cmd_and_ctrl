package aiseat_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pet_sehlat_bot_test.go — the bot half of #4973: "This creature can't
// attack alone" (Pet Sehlat, game.AttackLimitNotAlone) must never wedge
// a bot table, and a bot must never get the Sehlat to attack alone. The
// risky policy is one that declares the Sehlat first and then wants to
// pass: the enumerator withholds the pass and offers the companions as
// the always-legal answer, which is what the runner takes on a decline.

const botPetSehlatOracle = "509f255f-c3ee-4836-a583-d018060d0c4d"

// sehlatFirst attacks with the Sehlat whenever it is offered, then
// passes; with the pass withheld it takes the always-legal answer, as
// the runner does for a decline.
type sehlatFirst struct{}

func (sehlatFirst) Name() string { return "sehlat-first" }

func (sehlatFirst) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	for i, m := range in.Moves {
		if strings.HasPrefix(m.Label, "Attack") && strings.HasSuffix(m.Label, "with Pet Sehlat") {
			return aiseat.Decision{Index: i}, nil
		}
	}
	if pi := aiseat.PassIndex(in.Moves); pi >= 0 {
		return aiseat.Decision{Index: pi}, nil
	}
	if si := aiseat.SafeIndex(in.Moves); si >= 0 {
		return aiseat.Decision{Index: si}, nil
	}
	return aiseat.Decision{Index: 0}, nil
}

func sehlatTable(t *testing.T, companions int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newRoom(t, 2, 41).Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	active := g.Seats[g.Turn.ActiveSeat]
	sehlat := limitPush(g, active, "Pet Sehlat", "Creature — Bear", botPetSehlatOracle, 4, 3)
	for i := 0; i < companions; i++ {
		limitPush(g, active, "Ogre", "Creature — Ogre", "", 4, 4)
	}
	for _, p := range g.Seats {
		if p.ID != active.ID {
			limitPush(g, p, "Bear", "Creature — Bear", "", 2, 2)
		}
	}
	return g, sehlat
}

func TestBotsNeverAttackAloneWithPetSehlat(t *testing.T) {
	policies := map[string]aiseat.Policy{
		"heuristic":    heuristic.New(),
		"aggressive":   aggressive(),
		"sehlat-first": sehlatFirst{},
	}
	for _, companions := range []int{0, 1} {
		for name, pol := range policies {
			t.Run(fmt.Sprintf("%s/companions=%d", name, companions), func(t *testing.T) {
				g, sehlat := sehlatTable(t, companions)
				sehlatAttacked := false
				driveOneCombatWatching(t, g, pol, func() {
					if g.Turn.Step == game.StepDeclareAttackers {
						return
					}
					attackers := 0
					for i := range g.Battlefield.Cards {
						c := &g.Battlefield.Cards[i]
						if c.AttackingTarget != uuid.Nil {
							attackers++
							if c.InstanceID == sehlat {
								sehlatAttacked = true
							}
						}
					}
					if sehlatAttacked && attackers < 2 && g.Turn.Step == game.StepDeclareBlockers {
						t.Fatalf("companions=%d: the Sehlat attacks alone", companions)
					}
				})
				if name == "sehlat-first" && companions == 1 && !sehlatAttacked {
					t.Error("the Sehlat with a companion never attacked")
				}
				if companions == 0 && sehlatAttacked {
					t.Error("the Sehlat attacked with nothing beside it")
				}
			})
		}
	}
}
