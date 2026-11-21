package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kondas_banner_test.go — #5846: Konda's Banner's "can be attached only
// to a legendary creature" at the three places an attach is decided (an
// equip, an effect that attaches it, the CR 704.5n state-based action),
// and its two "share a …" anthems.

const kondasBannerOracle = "a355908b-4359-4bfe-9bc4-ccf721ab3a13"

// kbCreature puts a creature with the given type line and colors on the
// battlefield under `owner`.
func kbCreature(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost, Colors: colors,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

func kbBanner(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Konda's Banner", TypeLine: "Legendary Artifact — Equipment",
		ManaCost: "{2}", OracleID: kondasBannerOracle, Owner: owner, Controller: owner,
	})
}

// kbActivateEquip activates the Banner's equip ability at `creature` and
// lets it resolve, failing the test if the activation is refused.
func kbActivateEquip(t *testing.T, g *game.Game, controller, banner, creature uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(controller, banner, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}},
	}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// The equip ability is printed "Equip {2}": it may target any creature
// you control (CR 702.6a), and the rulings on the same clause (Gate
// Smasher, O-Naginata) say an equip aimed at a creature the Equipment
// can't be attached to resolves and does nothing. CR 301.5b / 701.3b:
// the Banner doesn't move, unattached or already on a legend.
func TestKondasBannerEquipDoesNotAttachToANonlegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	legend := effectstest.SeedLegendaryCreature(g, me.ID, "Commander")
	banner := kbBanner(g, me.ID)

	before := len(g.Events)
	kbActivateEquip(t, g, me.ID, banner, bear)
	if host := effectstest.AttachmentHostOf(t, g, banner); host.ID != uuid.Nil {
		t.Fatalf("Banner attached to %+v; it can be attached only to a legendary creature", host)
	}
	if n := effectstest.CountCatalogEvents(g, game.EventAttach, before); n != 0 {
		t.Errorf("EventAttach count = %d, want 0: the Banner never became attached", n)
	}
	if n := effectstest.CountCatalogEvents(g, game.EventUnattach, before); n != 0 {
		t.Errorf("EventUnattach count = %d, want 0: nothing was attached to come off", n)
	}
	if n := effectstest.CountCatalogEvents(g, game.EventAttachSkipped, before); n != 1 {
		t.Errorf("EventAttachSkipped count = %d, want 1", n)
	}
	if n := effectstest.CountCatalogEvents(g, game.EventEffectError, before); n != 0 {
		t.Errorf("EventEffectError count = %d, want 0", n)
	}

	kbActivateEquip(t, g, me.ID, banner, legend)
	if host := effectstest.AttachmentHostOf(t, g, banner); host.Kind != game.TargetCard || host.ID != legend {
		t.Fatalf("Banner on %+v, want the legendary creature %s", host, legend)
	}

	// Moving it onto the nonlegendary creature leaves it on the legend.
	kbActivateEquip(t, g, me.ID, banner, bear)
	if host := effectstest.AttachmentHostOf(t, g, banner); host.Kind != game.TargetCard || host.ID != legend {
		t.Errorf("Banner moved to %+v; an equip at a nonlegendary creature must leave it on the legend", host)
	}
}

// An effect that attaches the Banner (Brass Squire's "Attach target
// Equipment you control to target creature you control") is held to the
// same clause: CR 301.5b, the Equipment doesn't move.
func TestKondasBannerAttachEffectHonoursTheRestriction(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	squire := effectstest.PushCatalogPermanent(g, me.ID, "Brass Squire", "Artifact Creature — Myr", effectstest.BrassSquireOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	legend := effectstest.SeedLegendaryCreature(g, me.ID, "Commander")
	banner := kbBanner(g, me.ID)

	squireAttach := func(host uuid.UUID) {
		t.Helper()
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == squire {
					g.Battlefield.Cards[i].Tapped = false
				}
			}
		})
		if err := g.ActivateCatalogAbility(me.ID, squire, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: banner}, {Kind: game.TargetCard, ID: host}},
		}); err != nil {
			t.Fatalf("Brass Squire: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}

	squireAttach(bear)
	if host := effectstest.AttachmentHostOf(t, g, banner); host.ID == bear {
		t.Fatal("Brass Squire attached the Banner to a nonlegendary creature")
	}

	squireAttach(legend)
	if host := effectstest.AttachmentHostOf(t, g, banner); host.Kind != game.TargetCard || host.ID != legend {
		t.Fatalf("Banner on %+v, want the legendary creature", host)
	}

	squireAttach(bear)
	if host := effectstest.AttachmentHostOf(t, g, banner); host.ID != legend {
		t.Errorf("Banner moved to %+v; it can't be attached to the nonlegendary creature", host)
	}

	// The engine's own question, which "move an Equipment" choices list
	// their hosts with (AttachLegalForEffect), says the same.
	g.ReadSnapshot(func() {
		if g.AttachLegalForEffect(banner, bear) {
			t.Error("AttachLegalForEffect admits a nonlegendary creature")
		}
		if !g.AttachLegalForEffect(banner, legend) {
			t.Error("AttachLegalForEffect refuses the legendary creature")
		}
	})
}

// CR 301.5c / 704.5n: an Equipment attached to an illegal permanent
// becomes unattached and stays on the battlefield. A legend that stops
// being legendary is such a permanent for the Banner.
func TestKondasBannerFallsOffACreatureThatStopsBeingLegendary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	legend := effectstest.SeedLegendaryCreature(g, me.ID, "Commander")
	banner := kbBanner(g, me.ID)
	kbActivateEquip(t, g, me.ID, banner, legend)
	if host := effectstest.AttachmentHostOf(t, g, banner); host.ID != legend {
		t.Fatalf("Banner on %+v, want the legend", host)
	}

	// It stops being legendary (the type line stands in for a layer-4
	// effect removing the supertype).
	before := len(g.Events)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == legend {
				g.Battlefield.Cards[i].TypeLine = "Creature — Human Soldier"
			}
		}
		g.BumpLayerVersionForTest()
	})
	g.RunStateChecksForTest()

	if host := effectstest.AttachmentHostOf(t, g, banner); host.ID != uuid.Nil {
		t.Errorf("Banner still attached to %+v after its creature stopped being legendary", host)
	}
	if !g.Battlefield.Contains(banner) {
		t.Error("the Banner left the battlefield; CR 704.5n leaves an Equipment where it is")
	}
	if n := effectstest.CountCatalogEvents(g, game.EventUnattach, before); n != 1 {
		t.Errorf("EventUnattach count = %d, want 1", n)
	}
}

// The two anthems: +1/+1 for sharing a color, +1/+1 for sharing a
// creature type, so +2/+2 for both and never more (the 2004-12-01
// rulings); the equipped creature counts for both; whose creature it is
// doesn't matter; "Artifact" is not a creature type; and an unattached
// Banner pumps nothing.
func TestKondasBannerPumpsCreaturesThatShareAColorOrACreatureType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	legend := kbCreature(g, me.ID, "Isamaru", "Legendary Creature — Human Samurai", "{W}{U}", "W", "U")
	both := kbCreature(g, me.ID, "Twin", "Creature — Human Samurai", "{W}{U}", "W", "U")
	colorOnly := kbCreature(g, me.ID, "Elf", "Creature — Elf Druid", "{U}", "U")
	typeOnly := kbCreature(g, me.ID, "Warrior", "Creature — Human Warrior", "{R}", "R")
	neither := kbCreature(g, me.ID, "Goblin", "Creature — Goblin", "{R}", "R")
	golem := kbCreature(g, me.ID, "Golem", "Artifact Creature — Golem", "{3}")
	theirs := kbCreature(g, opp.ID, "Their Monk", "Creature — Human Monk", "{W}", "W")
	banner := kbBanner(g, me.ID)

	for _, id := range []uuid.UUID{legend, both, colorOnly, typeOnly, neither, golem, theirs} {
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 2 || tough != 2 {
			t.Fatalf("an unattached Banner pumps nothing: %s is %d/%d", id, p, tough)
		}
	}

	kbActivateEquip(t, g, me.ID, banner, legend)

	want := map[uuid.UUID]int{
		legend:    4, // itself: a color and a creature type
		both:      4, // two colors and two types still +2/+2
		colorOnly: 3,
		typeOnly:  3,
		neither:   2,
		golem:     2, // colorless, and Artifact is not a creature type
		theirs:    4, // an opponent's creature too
	}
	names := map[uuid.UUID]string{legend: "legend", both: "both", colorOnly: "color only",
		typeOnly: "type only", neither: "neither", golem: "golem", theirs: "opponent's"}
	for id, n := range want {
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != n || tough != n {
			t.Errorf("%s: %d/%d, want %d/%d", names[id], p, tough, n, n)
		}
	}

	// The printed restriction shows on the tile as a static row.
	var rows []game.AbilityRow
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == banner {
				rows = game.AbilityRowsOfPtr(&g.Battlefield.Cards[i])
			}
		}
	})
	found := false
	for _, r := range rows {
		if r.Kind == game.AbilityRowStatic && r.Label == "Can be attached only to a legendary creature." {
			found = true
		}
	}
	if !found {
		t.Errorf("ability rows %+v lack the attach restriction", rows)
	}
}
