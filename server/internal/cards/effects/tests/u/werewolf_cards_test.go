package testsu

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func villageWatchRow() cards.Card {
	return effectstest.WerewolfRow(effectstest.VillageWatchOracle,
		"Village Watch", "Creature — Human Werewolf", "{4}{R}",
		"Haste\nDaybound (If a player casts no spells during their own turn, it becomes night next turn.)",
		"Village Reavers", "Creature — Werewolf",
		"Wolves and Werewolves you control have haste.\nNightbound (If a player casts at least two spells during their own turn, it becomes day next turn.)",
		"4", "3", "5", "4", []string{"R"}, "Haste")
}

func battlefieldKeywords(g *game.Game, id uuid.UUID) []string {
	var out []string
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out = append([]string(nil), c.Effective().Abilities...)
			}
		}
	})
	return out
}

// Each face of an imported werewolf carries its own keyword: the front
// daybound, the back nightbound, neither the other's.
func TestImportedWerewolfFacesCarryTheirOwnKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[0]
	id := effectstest.ImportToHand(effectstest.FearfulVillagerRow(), p)
	c := p.Hand.Cards[len(p.Hand.Cards)-1]
	if c.InstanceID != id {
		t.Fatalf("imported card is not on top of the hand")
	}
	if !slices.Contains(c.Keywords, "daybound") || slices.Contains(c.Keywords, "nightbound") || !slices.Contains(c.Keywords, "menace") {
		t.Errorf("front keywords = %v, want daybound + menace only", c.Keywords)
	}
	c.SetFace(1)
	if !slices.Contains(c.Keywords, "nightbound") || slices.Contains(c.Keywords, "daybound") || !slices.Contains(c.Keywords, "menace") {
		t.Errorf("back keywords = %v, want nightbound + menace only", c.Keywords)
	}
}

// Cast at neither day nor night, the werewolf starts the clock (CR
// 702.145d) and stays on its front face.
func TestDayboundWerewolfCastAtNeitherMakesItDay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[0]
	id := effectstest.ImportAndCast(t, g, effectstest.FearfulVillagerRow(), p)
	effectstest.MonSettle(t, g)
	if !g.IsDay() {
		t.Fatalf("designation = %q, want day", effectstest.DnDesignation(g))
	}
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok || c.ActiveFace != 0 {
		t.Fatalf("Fearful Villager on the battlefield = %v, face %d, want the front", ok, c.ActiveFace)
	}
}

// Night turns it over, day turns it back, and menace rides both faces.
func TestWerewolfTurnsOverWithTheDesignationAndKeepsItsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[0]
	id := effectstest.ImportAndCast(t, g, effectstest.FearfulVillagerRow(), p)
	effectstest.MonSettle(t, g)

	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	c, _ := effectstest.BattlefieldCard(g, id)
	if c.ActiveFace != 1 || c.Name != "Fearsome Werewolf" {
		t.Fatalf("night: %s face %d, want Fearsome Werewolf", c.Name, c.ActiveFace)
	}
	kws := battlefieldKeywords(g, id)
	if !slices.Contains(kws, "menace") || !slices.Contains(kws, "nightbound") {
		t.Errorf("night abilities = %v, want menace + nightbound", kws)
	}
	if got := effectstest.EffectivePower(t, g, id); got != 4 {
		t.Errorf("night power = %d, want 4", got)
	}

	g.WithWriteLock(func() { g.BecomeDayForEffect() })
	c, _ = effectstest.BattlefieldCard(g, id)
	if c.ActiveFace != 0 || c.Name != "Fearful Villager" {
		t.Fatalf("day: %s face %d, want Fearful Villager", c.Name, c.ActiveFace)
	}
	if got := effectstest.EffectivePower(t, g, id); got != 2 {
		t.Errorf("day power = %d, want 2", got)
	}
}

// Nothing else turns a werewolf over: Moonmist-style "transform" is a
// no-op on it (CR 702.145b / 702.145e).
func TestAnEffectCannotTransformAWerewolf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[0]
	id := effectstest.ImportAndCast(t, g, effectstest.FearfulVillagerRow(), p)
	effectstest.MonSettle(t, g)
	g.WithWriteLock(func() {
		if err := (effects.Transform{Target: id}).Apply(effects.NewContext(g, &game.StackItem{Controller: p.ID})); err != nil {
			t.Fatalf("Transform: %v", err)
		}
	})
	if c, _ := effectstest.BattlefieldCard(g, id); c.ActiveFace != 0 {
		t.Fatalf("an effect transformed a daybound permanent (face %d)", c.ActiveFace)
	}
}

// The back face's "enters" ability is the one that sees a werewolf cast
// at night enter (CR 702.145b): two Insects, not one, and no transform.
func TestInfestationExpertCastAtNightEntersAsInfestedWerewolf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationNight })
	before := effectstest.CountKind(g, game.EventTransform)
	id := effectstest.ImportAndCast(t, g, effectstest.InfestationExpertRow(), p)
	effectstest.MonSettle(t, g)
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok || c.Name != "Infested Werewolf" || c.ActiveFace != 1 {
		t.Fatalf("battlefield: %v %q face %d, want Infested Werewolf on its back face", ok, c.Name, c.ActiveFace)
	}
	if got := effectstest.CountKind(g, game.EventTransform) - before; got != 0 {
		t.Errorf("a card cast at night 'transformed' %d times; it entered that way", got)
	}
	if n, _ := effectstest.CountTokensNamed(g, p.ID, "Insect"); n != 2 {
		t.Errorf("Insect tokens = %d, want 2 (the back face's enters trigger)", n)
	}
}

// By day it is the front face's single Insect, and turning over at night
// does not trigger "enters" again.
func TestInfestationExpertByDayMakesOneInsectAndTurningOverMakesNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[0]
	effectstest.ImportAndCast(t, g, effectstest.InfestationExpertRow(), p)
	effectstest.MonSettle(t, g)
	if n, _ := effectstest.CountTokensNamed(g, p.ID, "Insect"); n != 1 {
		t.Fatalf("Insect tokens = %d, want 1", n)
	}
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	effectstest.MonSettle(t, g)
	if n, _ := effectstest.CountTokensNamed(g, p.ID, "Insect"); n != 1 {
		t.Errorf("transforming at night made more Insects: %d, want 1 (nothing entered, CR 712.18)", n)
	}
}

// Village Reavers' grant covers Wolves and Werewolves you control,
// itself included, and nothing else.
func TestVillageReaversGivesHasteToYourWolvesAndWerewolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[0]
	id := effectstest.ImportAndCast(t, g, villageWatchRow(), p)
	effectstest.MonSettle(t, g)
	wolf := effectstest.PushCatalogPermanent(g, p.ID, "Wolf", "Creature — Wolf", "", false)
	bear := effectstest.PushCatalogPermanent(g, p.ID, "Bear", "Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Their Wolf", "Creature — Wolf", "", false)

	if slices.Contains(effectstest.EffectiveAbilities(t, g, wolf), "haste") {
		t.Fatal("a Wolf had haste by day, before Village Reavers")
	}
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, id), "haste") {
		t.Error("Village Reavers does not have haste itself (it is a Werewolf)")
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, wolf), "haste") {
		t.Error("your Wolf did not get haste")
	}
	if slices.Contains(effectstest.EffectiveAbilities(t, g, bear), "haste") {
		t.Error("your Bear got haste")
	}
	if slices.Contains(effectstest.EffectiveAbilities(t, g, theirs), "haste") {
		t.Error("an opponent's Wolf got haste: the grant says 'you control'")
	}
}
