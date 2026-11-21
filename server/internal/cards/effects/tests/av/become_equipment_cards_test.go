package testsav

import (
	"testing"

	"github.com/google/uuid"

	cardsth "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/th"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// seedIrencrag puts The Irencrag on the battlefield.
func seedIrencrag(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.GaPut(g, owner, "The Irencrag", "Legendary Artifact", cardsth.IrencragOracleID)
}

// legendEnters puts a legendary 2/2 onto the battlefield under `owner`
// and fires its entry, answering The Irencrag's "you may" with `yes`.
func legendEnters(t *testing.T, g *game.Game, owner uuid.UUID, name string, yes bool) uuid.UUID {
	t.Helper()
	legend := effectstest.SeedLegendaryCreature(g, owner, name)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: owner, CardID: legend})
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, owner, yes)
	effectstest.PassPriorityAroundTable(t, g)
	return legend
}

func TestTheIrencragBecomesEverflameAndEquips(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	crag := seedIrencrag(g, me.ID)
	if n := len(game.ManaAbilitiesForCard(effectstest.GaLayered(t, g, crag))); n != 1 {
		t.Fatalf("The Irencrag has %d mana abilities, want its {T}: Add {C}", n)
	}

	legend := legendEnters(t, g, me.ID, "Hero", true)

	c := effectstest.GaLayered(t, g, crag)
	ch := c.Effective()
	if ch.Name != cardsth.IrencragEverflameName {
		t.Errorf("name = %q, want %q", ch.Name, cardsth.IrencragEverflameName)
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
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.BeEquip(t, g, me.ID, crag, legend, "Equip {3}")
	if host := effectstest.AttachmentHostOf(t, g, crag); host.ID != legend {
		t.Fatalf("Everflame is attached to %+v, want the legend", host)
	}
	if p, tough := effectstest.EffectivePower(t, g, legend), effectstest.EffectiveToughness(t, g, legend); p != 5 || tough != 5 {
		t.Errorf("equipped legend is %d/%d, want 5/5", p, tough)
	}

	// No duration: it is still Everflame, still pumping, turns later.
	effectstest.PassTurnsTo(t, g, 0)
	effectstest.AdvanceToMain(t, g)
	if got := effectstest.GaLayered(t, g, crag).Effective().Name; got != cardsth.IrencragEverflameName {
		t.Errorf("after a turn the name is %q; the change has no duration", got)
	}
	if p := effectstest.EffectivePower(t, g, legend); p != 5 {
		t.Errorf("after a turn the legend's power is %d, want 5", p)
	}
}

func TestTheIrencragDeclinedStaysTheIrencrag(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	crag := seedIrencrag(g, me.ID)
	legendEnters(t, g, me.ID, "Hero", false)

	c := effectstest.GaLayered(t, g, crag)
	if c.Effective().Name != "The Irencrag" || c.HasSubtype("Equipment") {
		t.Errorf("declined, yet it is %q (%v)", c.Effective().Name, c.Effective().Subtypes)
	}
	if n := len(game.ManaAbilitiesForCard(c)); n != 1 {
		t.Errorf("declined, yet it has %d mana abilities", n)
	}
	if idx, _ := effectstest.BeGrantedRow(t, g, crag, "Equip {3}"); idx >= 0 {
		t.Error("declined, yet it has equip")
	}
}

// A non-legendary creature never offers the change.
func TestTheIrencragIgnoresAnOrdinaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	seedIrencrag(g, me.ID)
	bear := effectstest.SeedBear(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	first := seedIrencrag(g, me.ID)
	legendEnters(t, g, me.ID, "Hero", true)

	second := seedIrencrag(g, me.ID)
	g.RunStateChecksForTest()
	if effectstest.LegendPrompt(g, me.ID) != nil {
		t.Fatal("The Irencrag and Everflame, Heroes' Legacy were treated as the same legend")
	}

	// A second legend: both Irencrag-born permanents trigger; Everflame
	// has lost its trigger, so only the new Irencrag asks.
	legend := effectstest.SeedLegendaryCreature(g, me.ID, "Another Hero")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: legend})
	})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	ch := effectstest.LegendPrompt(g, me.ID)
	if ch == nil {
		t.Fatal("two Everflames did not trigger the legend rule")
	}
	if len(ch.PickTargetCards) != 2 || !effectstest.OnBoard(g, first) || !effectstest.OnBoard(g, second) {
		t.Errorf("legend-rule prompt over %v, want both Everflames", ch.PickTargetCards)
	}
}

// The change is data: a table holding it is a restore point, and the
// restored Everflame is still named, typed, equipped and pumping.
func TestEverflameSurvivesASnapshotRoundTrip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	crag := seedIrencrag(g, me.ID)
	legend := legendEnters(t, g, me.ID, "Hero", true)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.BeEquip(t, g, me.ID, crag, legend, "Equip {3}")

	restored := effectstest.RestoreRoundTrip(t, g, true)
	c := effectstest.GaLayered(t, restored, crag)
	if c.Effective().Name != cardsth.IrencragEverflameName || !c.HasSubtype("Equipment") {
		t.Errorf("restored: %q %v, want Everflame the Equipment", c.Effective().Name, c.Effective().Subtypes)
	}
	if p := effectstest.EffectivePower(t, restored, legend); p != 5 {
		t.Errorf("restored: the legend's power is %d, want 5", p)
	}
	if idx, _ := effectstest.BeGrantedRow(t, restored, crag, "Equip {3}"); idx < 0 {
		t.Error("restored: Everflame lost equip")
	}
}

// Gemcutter Buccaneer: its Treasures are Equipment with both equips and
// "+2/+0"; when it leaves, they stop being Equipment and fall off.
func TestGemcutterBuccaneerTreasuresAreEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	pirate := effectstest.B12Push(g, me.ID, "Gemcutter Buccaneer", "Creature — Orc Pirate Artificer", effectstest.GemcutterOracle, 1, 3)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: pirate})
	})
	effectstest.PassPriorityAroundTable(t, g)

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
	tr := effectstest.GaLayered(t, g, treasure)
	if !tr.HasSubtype("Equipment") || !tr.HasSubtype("Treasure") {
		t.Fatalf("Treasure subtypes = %v, want Treasure and Equipment", tr.Effective().Subtypes)
	}
	if n := len(game.ManaAbilitiesForCard(tr)); n == 0 {
		t.Error("the Treasure lost its own mana ability")
	}

	// Equip Pirate {1} reaches the Pirate and not a bear.
	bear := effectstest.SeedBear(g, me.ID)
	// Settling the trigger can walk the turn on; equip is sorcery speed.
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	idx, ref := effectstest.BeGrantedRow(t, g, treasure, "Equip Pirate {1}")
	if idx < 0 {
		t.Fatal("no Equip Pirate {1} row")
	}
	if err := g.ActivateCatalogAbility(me.ID, treasure, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("Equip Pirate targeted a bear")
	}
	effectstest.BeEquip(t, g, me.ID, treasure, pirate, "Equip Pirate {1}")
	if p, tough := effectstest.EffectivePower(t, g, pirate), effectstest.EffectiveToughness(t, g, pirate); p != 3 || tough != 3 {
		t.Errorf("equipped Buccaneer is %d/%d, want 3/3", p, tough)
	}
	// Equip {3} reaches any creature you control.
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.BeEquip(t, g, me.ID, treasure, bear, "Equip {3}")
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("equipped bear's power is %d, want 4", p)
	}
	if p := effectstest.EffectivePower(t, g, pirate); p != 1 {
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
	if tr := effectstest.GaLayered(t, g, treasure); tr.HasSubtype("Equipment") {
		t.Error("the Treasure is still an Equipment without the Buccaneer")
	}
	if host := effectstest.AttachmentHostOf(t, g, treasure); host.Kind != "" {
		t.Errorf("the Treasure is still attached to %+v", host)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the bear's power is %d, want 2", p)
	}
}

// Puresteel Paladin: equip {0} on your Equipment only with metalcraft,
// beside the Equipment's own equip.
func TestPuresteelPaladinGrantsEquipZeroWithMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.GaPut(g, me.ID, "Puresteel Paladin", "Creature — Human Knight", effectstest.G3PuresteelOracle)
	bear := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Bonesplitter", "452e3f5f-ce17-4682-966b-5cc100210aee")
	effectstest.SeedPermanent(g, me.ID, "Ornithopter Husk", "Artifact")

	if idx, _ := effectstest.BeGrantedRow(t, g, blade, "Equip {0}"); idx >= 0 {
		t.Fatal("equip {0} with two artifacts")
	}
	effectstest.SeedPermanent(g, me.ID, "Another Husk", "Artifact")
	effectstest.BeEquip(t, g, me.ID, blade, bear, "Equip {0}")
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("bear's power is %d, want 4 after equip {0}", p)
	}
	// The printed equip is still there (CR 702.6d).
	rows := game.ActivatedAbilitiesForCard(effectstest.GaLayered(t, g, blade))
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
