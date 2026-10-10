package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// become_equipment_cards_test.go — #2562: a permanent that becomes an
// Equipment, or is given equip, by another effect (ADR 0093 amendment
// 2026-10-10). The Irencrag through a resolved effect that renames it,
// Gemcutter Buccaneer and Puresteel Paladin through statics. Every test
// drives the real trigger, activation and layer paths and asserts what a
// player sees: the name, the type line, the rows, the equipped creature's
// size.

// beGrantedRow is the index and ref of the granted activated row with
// this label on a permanent, or -1.
func beGrantedRow(t *testing.T, g *game.Game, id uuid.UUID, label string) (int, string) {
	t.Helper()
	rows, origins := game.ActivatedAbilitiesWithOrigins(gaLayered(t, g, id))
	for i, o := range origins {
		if o.Granted() && rows[i].Label == label {
			return i, o.Ref
		}
	}
	return -1, ""
}

// beEquip activates a granted equip row onto `host` and lets it resolve.
func beEquip(t *testing.T, g *game.Game, controller, equipment, host uuid.UUID, label string) {
	t.Helper()
	idx, ref := beGrantedRow(t, g, equipment, label)
	if idx < 0 {
		t.Fatalf("%s has no granted %q row", equipment, label)
	}
	if err := g.ActivateCatalogAbility(controller, equipment, idx, game.ActivateAbilityParams{
		Ref:     ref,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}},
	}); err != nil {
		t.Fatalf("%s: %v", label, err)
	}
	passPriorityAroundTable(t, g)
}

// seedIrencrag puts The Irencrag on the battlefield.
func seedIrencrag(g *game.Game, owner uuid.UUID) uuid.UUID {
	return gaPut(g, owner, "The Irencrag", "Legendary Artifact", irencragOracleID)
}

// legendEnters puts a legendary 2/2 onto the battlefield under `owner`
// and fires its entry, answering The Irencrag's "you may" with `yes`.
func legendEnters(t *testing.T, g *game.Game, owner uuid.UUID, name string, yes bool) uuid.UUID {
	t.Helper()
	legend := seedLegendaryCreature(g, owner, name)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: owner, CardID: legend})
	})
	answerLatestTriggerPrompt(t, g, owner, yes)
	passPriorityAroundTable(t, g)
	return legend
}

func TestTheIrencragBecomesEverflameAndEquips(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	crag := seedIrencrag(g, me.ID)
	if n := len(game.ManaAbilitiesForCard(gaLayered(t, g, crag))); n != 1 {
		t.Fatalf("The Irencrag has %d mana abilities, want its {T}: Add {C}", n)
	}

	legend := legendEnters(t, g, me.ID, "Hero", true)

	c := gaLayered(t, g, crag)
	ch := c.Effective()
	if ch.Name != irencragEverflameName {
		t.Errorf("name = %q, want %q", ch.Name, irencragEverflameName)
	}
	if !c.IsLegendary() || !c.IsArtifact() || !c.HasSubtype("Equipment") || c.IsCreature() {
		t.Errorf("type line = %v %v — %v, want a legendary Equipment artifact", ch.Supertypes, ch.Types, ch.Subtypes)
	}
	// "Loses all other abilities": the {C} and the trigger are gone.
	if n := len(game.ManaAbilitiesForCard(c)); n != 0 {
		t.Errorf("Everflame has %d mana abilities, want none", n)
	}
	if n := len(game.TriggersForCard(c)); n != 0 {
		t.Errorf("Everflame has %d triggered abilities, want none", n)
	}

	// Equip {3} is activatable once granted, and the +3/+3 follows it.
	floatMana(t, g, me, "{C}{C}{C}")
	beEquip(t, g, me.ID, crag, legend, "Equip {3}")
	if host := attachmentHostOf(t, g, crag); host.ID != legend {
		t.Fatalf("Everflame is attached to %+v, want the legend", host)
	}
	if p, tough := effectivePower(t, g, legend), effectiveToughness(t, g, legend); p != 5 || tough != 5 {
		t.Errorf("equipped legend is %d/%d, want 5/5", p, tough)
	}

	// No duration: it is still Everflame, still pumping, turns later.
	passTurnsTo(t, g, 0)
	advanceToMain(t, g)
	if got := gaLayered(t, g, crag).Effective().Name; got != irencragEverflameName {
		t.Errorf("after a turn the name is %q; the change has no duration", got)
	}
	if p := effectivePower(t, g, legend); p != 5 {
		t.Errorf("after a turn the legend's power is %d, want 5", p)
	}
}

func TestTheIrencragDeclinedStaysTheIrencrag(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	crag := seedIrencrag(g, me.ID)
	legendEnters(t, g, me.ID, "Hero", false)

	c := gaLayered(t, g, crag)
	if c.Effective().Name != "The Irencrag" || c.HasSubtype("Equipment") {
		t.Errorf("declined, yet it is %q (%v)", c.Effective().Name, c.Effective().Subtypes)
	}
	if n := len(game.ManaAbilitiesForCard(c)); n != 1 {
		t.Errorf("declined, yet it has %d mana abilities", n)
	}
	if idx, _ := beGrantedRow(t, g, crag, "Equip {3}"); idx >= 0 {
		t.Error("declined, yet it has equip")
	}
}

// A non-legendary creature never offers the change.
func TestTheIrencragIgnoresAnOrdinaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedIrencrag(g, me.ID)
	bear := seedBear(g, me.ID)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: bear})
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
			t.Fatal("The Irencrag offered to change for a non-legendary creature")
		}
	}
}

// The legend rule reads the new name (CR 612.8, 704.5j): The Irencrag
// beside Everflame is no pair, and two Everflames are.
func TestTheLegendRuleSeesEverflamesName(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	first := seedIrencrag(g, me.ID)
	legendEnters(t, g, me.ID, "Hero", true)

	second := seedIrencrag(g, me.ID)
	g.RunStateChecksForTest()
	if legendPrompt(g, me.ID) != nil {
		t.Fatal("The Irencrag and Everflame, Heroes' Legacy were treated as the same legend")
	}

	// A second legend: both Irencrag-born permanents trigger; Everflame
	// has lost its trigger, so only the new Irencrag asks.
	legend := seedLegendaryCreature(g, me.ID, "Another Hero")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: legend})
	})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	ch := legendPrompt(g, me.ID)
	if ch == nil {
		t.Fatal("two Everflames did not trigger the legend rule")
	}
	if len(ch.PickTargetCards) != 2 || !onBoard(g, first) || !onBoard(g, second) {
		t.Errorf("legend-rule prompt over %v, want both Everflames", ch.PickTargetCards)
	}
}

// The change is data: a table holding it is a restore point, and the
// restored Everflame is still named, typed, equipped and pumping.
func TestEverflameSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	crag := seedIrencrag(g, me.ID)
	legend := legendEnters(t, g, me.ID, "Hero", true)
	floatMana(t, g, me, "{C}{C}{C}")
	beEquip(t, g, me.ID, crag, legend, "Equip {3}")

	restored := restoreRoundTrip(t, g, true)
	c := gaLayered(t, restored, crag)
	if c.Effective().Name != irencragEverflameName || !c.HasSubtype("Equipment") {
		t.Errorf("restored: %q %v, want Everflame the Equipment", c.Effective().Name, c.Effective().Subtypes)
	}
	if p := effectivePower(t, restored, legend); p != 5 {
		t.Errorf("restored: the legend's power is %d, want 5", p)
	}
	if idx, _ := beGrantedRow(t, restored, crag, "Equip {3}"); idx < 0 {
		t.Error("restored: Everflame lost equip")
	}
}

// Gemcutter Buccaneer: its Treasures are Equipment with both equips and
// "+2/+0"; when it leaves, they stop being Equipment and fall off.
func TestGemcutterBuccaneerTreasuresAreEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pirate := b12Push(g, me.ID, "Gemcutter Buccaneer", "Creature — Orc Pirate Artificer", gemcutterOracle, 1, 3)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: pirate})
	})
	passPriorityAroundTable(t, g)

	var treasure uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.HasSubtype("Treasure") {
				treasure = c.InstanceID
			}
		}
	})
	if treasure == uuid.Nil {
		t.Fatal("no Treasure was made")
	}
	tr := gaLayered(t, g, treasure)
	if !tr.HasSubtype("Equipment") || !tr.HasSubtype("Treasure") {
		t.Fatalf("Treasure subtypes = %v, want Treasure and Equipment", tr.Effective().Subtypes)
	}
	if n := len(game.ManaAbilitiesForCard(tr)); n == 0 {
		t.Error("the Treasure lost its own mana ability")
	}

	// Equip Pirate {1} reaches the Pirate and not a bear.
	bear := seedBear(g, me.ID)
	// Settling the trigger can walk the turn on; equip is sorcery speed.
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	idx, ref := beGrantedRow(t, g, treasure, "Equip Pirate {1}")
	if idx < 0 {
		t.Fatal("no Equip Pirate {1} row")
	}
	if err := g.ActivateCatalogAbility(me.ID, treasure, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("Equip Pirate targeted a bear")
	}
	beEquip(t, g, me.ID, treasure, pirate, "Equip Pirate {1}")
	if p, tough := effectivePower(t, g, pirate), effectiveToughness(t, g, pirate); p != 3 || tough != 3 {
		t.Errorf("equipped Buccaneer is %d/%d, want 3/3", p, tough)
	}
	// Equip {3} reaches any creature you control.
	floatMana(t, g, me, "{C}{C}{C}")
	beEquip(t, g, me.ID, treasure, bear, "Equip {3}")
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("equipped bear's power is %d, want 4", p)
	}
	if p := effectivePower(t, g, pirate); p != 1 {
		t.Errorf("the Buccaneer kept the bonus: power %d", p)
	}

	// The Buccaneer leaves: the Treasure is not an Equipment any more,
	// unattaches (CR 704.5n) and the bonus is gone.
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(pirate); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if tr := gaLayered(t, g, treasure); tr.HasSubtype("Equipment") {
		t.Error("the Treasure is still an Equipment without the Buccaneer")
	}
	if host := attachmentHostOf(t, g, treasure); host.Kind != "" {
		t.Errorf("the Treasure is still attached to %+v", host)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("the bear's power is %d, want 2", p)
	}
}

// Puresteel Paladin: equip {0} on your Equipment only with metalcraft,
// beside the Equipment's own equip.
func TestPuresteelPaladinGrantsEquipZeroWithMetalcraft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gaPut(g, me.ID, "Puresteel Paladin", "Creature — Human Knight", g3PuresteelOracle)
	bear := seedBear(g, me.ID)
	blade := seedEquipment(g, me.ID, "Bonesplitter", "452e3f5f-ce17-4682-966b-5cc100210aee")
	seedPermanent(g, me.ID, "Ornithopter Husk", "Artifact")

	if idx, _ := beGrantedRow(t, g, blade, "Equip {0}"); idx >= 0 {
		t.Fatal("equip {0} with two artifacts")
	}
	seedPermanent(g, me.ID, "Another Husk", "Artifact")
	beEquip(t, g, me.ID, blade, bear, "Equip {0}")
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("bear's power is %d, want 4 after equip {0}", p)
	}
	// The printed equip is still there (CR 702.6d).
	rows := game.ActivatedAbilitiesForCard(gaLayered(t, g, blade))
	own := 0
	for _, r := range rows {
		if r.Label == "Equip {1}" {
			own++
		}
	}
	if own != 1 {
		t.Errorf("Bonesplitter rows = %d own Equip {1}, want 1", own)
	}
}
