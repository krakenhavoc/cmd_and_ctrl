package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// defender_controls_view_test.go — ADR 0107 §2 (#1879) on the wire. A
// creature that "can't attack unless defending player controls an Island"
// carries one attack_target_restrictions row per opponent with no Island,
// with the Island named (the chip's second sentence, decision 3), and the
// viewer's legal_actions digest offers it only the opponents that have
// one. The client's attack rings are drawn from that digest. The test
// asks for a Desert, because busyTable gives every seat basic lands.

const viewSerpentOracle = "test-view-island-serpent"

func withViewSerpentStatic(t *testing.T) {
	t.Helper()
	prev := game.CatalogStaticAbilities
	t.Cleanup(func() { game.CatalogStaticAbilities = prev })
	game.CatalogStaticAbilities = func(key string) []game.StaticAbility {
		if key != viewSerpentOracle {
			if prev == nil {
				return nil
			}
			return prev(key)
		}
		return []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
					Source: source.InstanceID, SourceName: source.Name,
					DefenderMustControl: []game.PermanentQuery{{Subtypes: []string{"Desert"}}},
				})
			},
		}}
	}
}

func TestDefenderMustControlReachesTheCardViewAndTheDigest(t *testing.T) {
	withViewSerpentStatic(t)
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	islands := g.Seats[(g.Turn.ActiveSeat+1)%4]
	serpent := uuid.New()
	island := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: serpent, Name: "Sea Serpent", TypeLine: "Creature — Serpent",
		OracleID: viewSerpentOracle, Power: 5, Toughness: 5, Owner: active.ID, Controller: active.ID,
	})
	g.Battlefield.PushTop(game.Card{
		InstanceID: island, Name: "Desert", TypeLine: "Land — Desert", Owner: islands.ID, Controller: islands.ID,
	})
	knowTheTable(g)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{serpent, island} {
			g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		}
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)

	v := ViewOfGameFor(g, active.ID.String())
	cv := cardInFrame(v, serpent.String())
	if cv == nil {
		t.Fatal("serpent not in the frame")
	}
	var refused []string
	for _, r := range cv.AttackTargetRestrictions {
		if r.Unless != "Desert" || !r.Planeswalkers || r.Source != "Sea Serpent" {
			t.Errorf("row = %+v, want a Desert row from Sea Serpent", r)
		}
		refused = append(refused, r.Player)
	}
	var want []string
	for _, p := range g.Seats {
		if p.ID != active.ID && p.ID != islands.ID && !p.Eliminated {
			want = append(want, p.ID.String())
		}
	}
	slices.Sort(refused)
	slices.Sort(want)
	if !slices.Equal(refused, want) {
		t.Errorf("rows name %v, want the opponents with no Desert %v", refused, want)
	}

	if v.LegalActions == nil || v.LegalActions.Sources[serpent.String()] == nil {
		t.Fatal("the serpent has no digest entry: it should be able to attack the Desert player")
	}
	targets := v.LegalActions.Sources[serpent.String()].AttackTargets
	if !slices.Equal(targets, []string{islands.ID.String()}) {
		t.Errorf("digest attack_targets = %v, want only the Desert player", targets)
	}
}
