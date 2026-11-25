package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// artifacts_become_equipment_mv_test.go — #6266: Bludgeon Brawl's "Each
// noncreature, non-Equipment artifact is an Equipment with equip {X} and
// 'Equipped creature gets +X/+0,' where X is that artifact's mana value"
// (CR 702.6a, CR 202.3), on a test oracle declared as the card would be.

const testBrawlOracle = "test-6266-bludgeon-brawl"

func registerBrawl(t *testing.T) {
	t.Helper()
	const (
		pumpGrant  = "test-6266/artifact-pump"
		equipGrant = "test-6266/artifact-equip"
	)
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: testBrawlOracle, Name: "Brawl Test",
		Grants: []effects.AbilityGrant{
			{Key: pumpGrant, Static: []game.StaticAbility{effects.PumpAttachedBySourceManaValue("Equipped creature gets +X/+0, where X is this artifact's mana value.")},
				Text: "Equipped creature gets +X/+0."},
			{Key: equipGrant, Activated: []effects.ActivatedAbility{effects.EquipForItsManaValue()}, Text: "Equip {X}"},
		},
		Static: []game.StaticAbility{
			effects.AreAlsoEquipment("Each noncreature, non-Equipment artifact is an Equipment.", effects.NoncreatureNonEquipmentArtifacts),
			effects.Labelled(effects.GrantAbilities(effects.NoncreatureNonEquipmentArtifactsGrantSet, pumpGrant, equipGrant),
				"Each noncreature, non-Equipment artifact has equip {X} and \"Equipped creature gets +X/+0.\""),
		},
	})
}

func brawlArtifact(g *game.Game, owner uuid.UUID, name, typeLine, cost string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Power: 1, Toughness: 1,
		Owner: owner, Controller: owner,
	})
}

func isEquipment(g *game.Game, id uuid.UUID) bool {
	var eff game.Card
	g.ReadSnapshot(func() { eff, _ = g.LookupCardForEffect(id) })
	return eff.HasSubtype("Equipment")
}

// TestBludgeonBrawlEquipAndPumpAreTheArtifactsManaValue is the
// acceptance: a 3-mana-value artifact has equip {3} and gives +3/+0; a
// 0-mana-value one has equip {0} and gives +0/+0; an artifact creature is
// not affected. An opponent's artifact is an Equipment too ("each").
func TestBludgeonBrawlEquipAndPumpAreTheArtifactsManaValue(t *testing.T) {
	registerBrawl(t)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Brawl Test", "Enchantment", testBrawlOracle, false)
	relic := brawlArtifact(g, me.ID, "Coalition Relic", "Artifact", "{3}")
	darksteel := brawlArtifact(g, me.ID, "Darksteel Relic", "Artifact", "")
	golem := brawlArtifact(g, me.ID, "Golem", "Artifact Creature — Golem", "{4}")
	theirs := brawlArtifact(g, opp.ID, "Mind Stone", "Artifact", "{2}")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bear2 := effectstest.B12Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	if !isEquipment(g, relic) || !isEquipment(g, darksteel) || !isEquipment(g, theirs) {
		t.Fatal("a noncreature artifact is not an Equipment")
	}
	if isEquipment(g, golem) {
		t.Error("an artifact creature became an Equipment")
	}
	if idx, _ := effectstest.BeGrantedRow(t, g, golem, "Equip {X}"); idx >= 0 {
		t.Error("an artifact creature has an equip row")
	}

	g.WithWriteLock(func() { effectstest.FillPool(me, 3) })
	effectstest.BeEquip(t, g, me.ID, relic, bear, "Equip {X}")
	if left := me.ManaPool.Total(); left != 0 {
		t.Errorf("equip on a 3-mana-value artifact left %d mana of 3, want it to cost {3}", left)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 5 {
		t.Errorf("creature equipped with the 3-mana-value artifact has power %d, want 5 (+3/+0)", p)
	}
	effectstest.BeEquip(t, g, me.ID, darksteel, bear2, "Equip {X}")
	if p := effectstest.EffectivePower(t, g, bear2); p != 2 {
		t.Errorf("creature equipped with the 0-mana-value artifact has power %d, want 2 (+0/+0)", p)
	}
}
