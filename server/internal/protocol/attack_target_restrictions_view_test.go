package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_target_restrictions_view_test.go — ADR 0106 §2 (#1794) on the
// wire. A creature that "can't attack its owner or planeswalkers its
// owner controls" carries the restriction on its card view, with the
// owner resolved to a seat id (the chip's data, owner decision 3), and
// the viewer's legal_actions digest leaves the owner out of that
// creature's attack_targets while the per-seat Turn.AttackTargets
// keeps it (it is about prices, not one creature).

const viewSleeperOracle = "test-view-sleeper-agent"

func withViewSleeperStatic(t *testing.T) {
	t.Helper()
	prev := game.CatalogStaticAbilities
	t.Cleanup(func() { game.CatalogStaticAbilities = prev })
	game.CatalogStaticAbilities = func(key string) []game.StaticAbility {
		if key != viewSleeperOracle {
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
					Source: source.InstanceID, SourceName: source.Name, NotOwner: true, NotOwnersPlaneswalkers: true,
				})
			},
		}}
	}
}

func TestCantAttackOwnerReachesTheCardViewAndTheDigest(t *testing.T) {
	withViewSleeperStatic(t)
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%4]
	sleeper := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: sleeper, Name: "Sleeper Agent", TypeLine: "Legendary Creature — Phyrexian Minion",
		OracleID: viewSleeperOracle, Power: 5, Toughness: 5, Owner: owner.ID, Controller: active.ID,
	})
	// The same restriction on a creature its owner controls says
	// nothing CR 506.2 does not, so it is not sent.
	home := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: home, Name: "Sleeper At Home", TypeLine: "Creature — Minion",
		OracleID: viewSleeperOracle, Power: 1, Toughness: 1, Owner: active.ID, Controller: active.ID,
	})
	knowTheTable(g)
	// The zone moves a real entry emits, so the layer pass reaches the
	// statics; neither has been here only since this turn.
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{sleeper, home} {
			g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		}
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)

	v := ViewOfGameFor(g, active.ID.String())
	cv := cardInFrame(v, sleeper.String())
	if cv == nil {
		t.Fatal("sleeper not in the frame")
	}
	want := []AttackTargetRestrictionView{{Player: owner.ID.String(), Planeswalkers: true, Source: "Sleeper Agent"}}
	if !slices.Equal(cv.AttackTargetRestrictions, want) {
		t.Errorf("attack_target_restrictions = %+v, want %+v", cv.AttackTargetRestrictions, want)
	}
	if hv := cardInFrame(v, home.String()); hv == nil || len(hv.AttackTargetRestrictions) != 0 {
		t.Errorf("a restriction naming the creature's own controller was sent: %+v", hv)
	}
	// Every viewer sees the chip: it is public, like the card.
	if ov := cardInFrame(ViewOfGameFor(g, owner.ID.String()), sleeper.String()); ov == nil || len(ov.AttackTargetRestrictions) != 1 {
		t.Errorf("the owner's frame lacks the restriction: %+v", ov)
	}

	if v.LegalActions == nil || v.LegalActions.Sources[sleeper.String()] == nil {
		t.Fatal("the sleeper has no digest entry: it should be able to attack the other two seats")
	}
	targets := v.LegalActions.Sources[sleeper.String()].AttackTargets
	if slices.Contains(targets, owner.ID.String()) || len(targets) != 2 {
		t.Errorf("digest attack_targets = %v: want the two seats that are not its owner", targets)
	}
	perSeat := false
	for _, row := range v.Turn.AttackTargets {
		if row.ID == owner.ID.String() {
			perSeat = true
		}
	}
	if !perSeat {
		t.Error("Turn.AttackTargets lost the owner: it is per seat, for prices and limits")
	}
}
