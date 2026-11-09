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

// toby_beast_bot_test.go — the bot half of #5396: Toby, Beastie
// Befriender's Beast token "can't block alone" (game.BlockRule.NotAlone)
// must never wedge a bot table, and a bot must never get the Beast to
// block alone. The enumerator never offers the Beast's block alone; with
// a companion it offers the pair as one move, and once a companion is
// staged, the Beast's singles.

// beastFirst attacks with everything it is offered, blocks with the
// Beast whenever a block naming it is offered, then finishes.
type beastFirst struct{}

func (beastFirst) Name() string { return "beast-first" }

func (beastFirst) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	for i, m := range in.Moves {
		if strings.HasPrefix(m.Label, "Attack") {
			return aiseat.Decision{Index: i}, nil
		}
	}
	for i, m := range in.Moves {
		if strings.HasPrefix(m.Label, "Block") && strings.Contains(m.Label, "Beast") {
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

// tobyBeastTable: the active seat attacks with Ogres; each other seat
// holds a Beast and `companions` Bears.
func tobyBeastTable(t *testing.T, companions int) (*game.Game, map[uuid.UUID]bool) {
	t.Helper()
	g := newRoom(t, 2, 41).Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	active := g.Seats[g.Turn.ActiveSeat]
	limitPush(g, active, "Ogre", "Creature — Ogre", "", 3, 3)
	limitPush(g, active, "Ogre", "Creature — Ogre", "", 3, 3)
	beasts := map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		if p.ID == active.ID {
			continue
		}
		c := game.Card{
			InstanceID: uuid.New(),
			Name:       "Beast", TypeLine: "Token Creature — Beast",
			TokenKey: game.TokenKey("beast-not-alone"),
			Power:    4, Toughness: 4, Colors: []string{"W"},
			Owner: p.ID, Controller: p.ID,
		}
		c.AddKnowersAll(seatIDs(g))
		g.Battlefield.PushTop(c)
		beasts[c.InstanceID] = true
		for i := 0; i < companions; i++ {
			limitPush(g, p, "Bear", "Creature — Bear", "", 2, 2)
		}
	}
	return g, beasts
}

func TestBotsNeverBlockAloneWithTobysBeast(t *testing.T) {
	policies := map[string]aiseat.Policy{
		"heuristic":   heuristic.New(),
		"aggressive":  aggressive(),
		"beast-first": beastFirst{},
	}
	for _, companions := range []int{0, 1} {
		for name, pol := range policies {
			t.Run(fmt.Sprintf("%s/companions=%d", name, companions), func(t *testing.T) {
				g, beasts := tobyBeastTable(t, companions)
				for i := range g.Battlefield.Cards {
					if c := &g.Battlefield.Cards[i]; beasts[c.InstanceID] && !game.CantBlockAlone(c) {
						t.Fatal("the Beast token does not read as can't-block-alone")
					}
				}
				beastBlocked := false
				driveOneCombatWatching(t, g, pol, func() {
					// Once the defender has declared, what stands is
					// what was announced.
					// Judged in declare blockers only: a companion that
					// dies in combat damage leaves the Beast blocking,
					// as the rulings say it should.
					if g.Turn.Step != game.StepDeclareBlockers {
						return
					}
					blockers := map[uuid.UUID]int{}
					var beastBlocking uuid.UUID
					for i := range g.Battlefield.Cards {
						c := &g.Battlefield.Cards[i]
						if len(c.BlockedAttackers()) == 0 {
							continue
						}
						if g.BlockDeclarationStatusOf(c.Controller) != game.BlockDeclarationDeclared {
							continue
						}
						blockers[c.Controller]++
						if beasts[c.InstanceID] {
							beastBlocked = true
							beastBlocking = c.Controller
						}
					}
					if beastBlocking != uuid.Nil && blockers[beastBlocking] < 2 {
						t.Fatalf("companions=%d: the Beast blocks alone", companions)
					}
				})
				if companions == 0 && beastBlocked {
					t.Error("the Beast blocked with nothing beside it")
				}
				if name == "beast-first" && companions == 1 && !beastBlocked {
					t.Error("the Beast with a companion never blocked")
				}
			})
		}
	}
}
