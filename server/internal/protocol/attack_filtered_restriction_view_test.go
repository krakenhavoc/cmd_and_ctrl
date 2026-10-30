package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const viewFlierWardOracle = "test-view-flier-ward"

// #4601: a restriction with an attacker filter ("creatures with flying
// can't attack you") is a chip only on the creatures it binds right now,
// it names the player alone (no planeswalkers), and the digest leaves that
// player out of the flier's attack_targets and not the ground creature's.
func TestAttackerFilteredRestrictionIsAChipOnlyOnMatchingCreatures(t *testing.T) {
	prev := game.CatalogStaticAbilities
	t.Cleanup(func() { game.CatalogStaticAbilities = prev })
	game.CatalogStaticAbilities = func(key string) []game.StaticAbility {
		if key != viewFlierWardOracle {
			if prev == nil {
				return nil
			}
			return prev(key)
		}
		return []game.StaticAbility{{
			Layer:     game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool { return target.IsCreature() },
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
					Source: source.InstanceID, SourceName: source.Name, NotPlayer: source.Controller, NotPlayerOnly: true,
					OnlyAttackersMatching: []game.PermanentQuery{{Types: []string{"creature"}, Keyword: "flying"}},
				})
			},
		}}
	}
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]
	warded := g.Seats[(g.Turn.ActiveSeat+1)%4]
	ward, flier, bear := uuid.New(), uuid.New(), uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: ward, Name: "Test Ward", TypeLine: "Enchantment",
		OracleID: viewFlierWardOracle, Owner: warded.ID, Controller: warded.ID})
	g.Battlefield.PushTop(game.Card{InstanceID: flier, Name: "Hawk", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: active.ID, Controller: active.ID, Keywords: []string{"flying"}})
	g.Battlefield.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID})
	knowTheTable(g)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{ward, flier, bear} {
			g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		}
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)

	v := ViewOfGameFor(g, active.ID.String())
	fv, bv := cardInFrame(v, flier.String()), cardInFrame(v, bear.String())
	if fv == nil || bv == nil {
		t.Fatal("the creatures are not in the frame")
	}
	want := []AttackTargetRestrictionView{{
		Player: warded.ID.String(), Source: "Test Ward", Reason: "it's a creature with flying",
	}}
	if !slices.Equal(fv.AttackTargetRestrictions, want) {
		t.Errorf("flier chip = %+v, want %+v", fv.AttackTargetRestrictions, want)
	}
	if len(bv.AttackTargetRestrictions) != 0 {
		t.Errorf("a creature without flying carries a chip: %+v", bv.AttackTargetRestrictions)
	}
	if v.LegalActions == nil || v.LegalActions.Sources[flier.String()] == nil || v.LegalActions.Sources[bear.String()] == nil {
		t.Fatal("the creatures have no digest entries")
	}
	if slices.Contains(v.LegalActions.Sources[flier.String()].AttackTargets, warded.ID.String()) {
		t.Error("the flier's digest still offers the protected player")
	}
	if !slices.Contains(v.LegalActions.Sources[bear.String()].AttackTargets, warded.ID.String()) {
		t.Error("the ground creature's digest lost the protected player")
	}
}
