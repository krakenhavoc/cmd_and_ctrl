package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// werewolf_cards_test.go — the daybound // nightbound werewolves in the
// catalogue (#2561, ADR 0132), on real imported cards.
//
// They go through deck.ToGameCard for transform_cards_test.go's reason,
// sharper here: daybound and nightbound reach a face only through the
// importer's per-face keyword stamping (the dump's `keywords` array
// narrowed by each face's own text), and a fixture built by hand would
// skip exactly the path a player's deck takes.

const (
	fearfulVillagerOracle   = "5fd09dbc-8bcd-4fe0-91b5-b00e721fa7eb"
	infestationExpertOracle = "bf69dc2a-9aec-4181-bbe9-70875055ec03"
	villageWatchOracle      = "b2ecaae4-41ee-4c61-b5ce-db4364b307fc"
)

// werewolfRow builds a Scryfall `transform` record the way the bulk dump
// has every werewolf: the top-level `keywords` array names both
// keywords, and each face's oracle text has its own keyword lines.
func werewolfRow(oracleID, front, frontType, frontCost, frontText, back, backType, backText string, fp, ft, bp, bt string, colors []string, extraKeywords ...string) cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     front + " // " + back,
		Layout:   "transform",
		TypeLine: frontType + " // " + backType,
		OracleID: uuid.MustParse(oracleID),
		Keywords: append([]string{"Daybound", "Nightbound"}, extraKeywords...),
		CardFaces: []cards.CardFace{
			{Name: front, TypeLine: frontType, ManaCost: frontCost, Colors: colors, Power: fp, Toughness: ft, OracleText: frontText},
			{Name: back, TypeLine: backType, Colors: colors, Power: bp, Toughness: bt, OracleText: backText},
		},
	}
}

func fearfulVillagerRow() cards.Card {
	return werewolfRow(fearfulVillagerOracle,
		"Fearful Villager", "Creature — Human Werewolf", "{2}{R}",
		"Menace (This creature can't be blocked except by two or more creatures.)\nDaybound (If a player casts no spells during their own turn, it becomes night next turn.)",
		"Fearsome Werewolf", "Creature — Werewolf",
		"Menace (This creature can't be blocked except by two or more creatures.)\nNightbound (If a player casts at least two spells during their own turn, it becomes day next turn.)",
		"2", "3", "4", "3", []string{"R"}, "Menace")
}

func infestationExpertRow() cards.Card {
	return werewolfRow(infestationExpertOracle,
		"Infestation Expert", "Creature — Human Werewolf", "{4}{G}",
		"Whenever this creature enters or attacks, create a 1/1 green Insect creature token.\nDaybound (If a player casts no spells during their own turn, it becomes night next turn.)",
		"Infested Werewolf", "Creature — Werewolf",
		"Whenever this creature enters or attacks, create two 1/1 green Insect creature tokens.\nNightbound (If a player casts at least two spells during their own turn, it becomes day next turn.)",
		"3", "4", "4", "5", []string{"G"})
}

func villageWatchRow() cards.Card {
	return werewolfRow(villageWatchOracle,
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
	g := newCatalogGame(t)
	p := g.Seats[0]
	id := importToHand(fearfulVillagerRow(), p)
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
	g := newCatalogGame(t)
	p := g.Seats[0]
	id := importAndCast(t, g, fearfulVillagerRow(), p)
	monSettle(t, g)
	if !g.IsDay() {
		t.Fatalf("designation = %q, want day", dnDesignation(g))
	}
	c, ok := battlefieldCard(g, id)
	if !ok || c.ActiveFace != 0 {
		t.Fatalf("Fearful Villager on the battlefield = %v, face %d, want the front", ok, c.ActiveFace)
	}
}

// Night turns it over, day turns it back, and menace rides both faces.
func TestWerewolfTurnsOverWithTheDesignationAndKeepsItsKeywords(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[0]
	id := importAndCast(t, g, fearfulVillagerRow(), p)
	monSettle(t, g)

	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	c, _ := battlefieldCard(g, id)
	if c.ActiveFace != 1 || c.Name != "Fearsome Werewolf" {
		t.Fatalf("night: %s face %d, want Fearsome Werewolf", c.Name, c.ActiveFace)
	}
	kws := battlefieldKeywords(g, id)
	if !slices.Contains(kws, "menace") || !slices.Contains(kws, "nightbound") {
		t.Errorf("night abilities = %v, want menace + nightbound", kws)
	}
	if got := effectivePower(t, g, id); got != 4 {
		t.Errorf("night power = %d, want 4", got)
	}

	g.WithWriteLock(func() { g.BecomeDayForEffect() })
	c, _ = battlefieldCard(g, id)
	if c.ActiveFace != 0 || c.Name != "Fearful Villager" {
		t.Fatalf("day: %s face %d, want Fearful Villager", c.Name, c.ActiveFace)
	}
	if got := effectivePower(t, g, id); got != 2 {
		t.Errorf("day power = %d, want 2", got)
	}
}

// Nothing else turns a werewolf over: Moonmist-style "transform" is a
// no-op on it (CR 702.145b / 702.145e).
func TestAnEffectCannotTransformAWerewolf(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[0]
	id := importAndCast(t, g, fearfulVillagerRow(), p)
	monSettle(t, g)
	g.WithWriteLock(func() {
		if err := (Transform{Target: id}).Apply(NewContext(g, &game.StackItem{Controller: p.ID})); err != nil {
			t.Fatalf("Transform: %v", err)
		}
	})
	if c, _ := battlefieldCard(g, id); c.ActiveFace != 0 {
		t.Fatalf("an effect transformed a daybound permanent (face %d)", c.ActiveFace)
	}
}

// The back face's "enters" ability is the one that sees a werewolf cast
// at night enter (CR 702.145b): two Insects, not one, and no transform.
func TestInfestationExpertCastAtNightEntersAsInfestedWerewolf(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationNight })
	before := countKind(g, game.EventTransform)
	id := importAndCast(t, g, infestationExpertRow(), p)
	monSettle(t, g)
	c, ok := battlefieldCard(g, id)
	if !ok || c.Name != "Infested Werewolf" || c.ActiveFace != 1 {
		t.Fatalf("battlefield: %v %q face %d, want Infested Werewolf on its back face", ok, c.Name, c.ActiveFace)
	}
	if got := countKind(g, game.EventTransform) - before; got != 0 {
		t.Errorf("a card cast at night 'transformed' %d times; it entered that way", got)
	}
	if n, _ := countTokensNamed(g, p.ID, "Insect"); n != 2 {
		t.Errorf("Insect tokens = %d, want 2 (the back face's enters trigger)", n)
	}
}

// By day it is the front face's single Insect, and turning over at night
// does not trigger "enters" again.
func TestInfestationExpertByDayMakesOneInsectAndTurningOverMakesNone(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[0]
	importAndCast(t, g, infestationExpertRow(), p)
	monSettle(t, g)
	if n, _ := countTokensNamed(g, p.ID, "Insect"); n != 1 {
		t.Fatalf("Insect tokens = %d, want 1", n)
	}
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	monSettle(t, g)
	if n, _ := countTokensNamed(g, p.ID, "Insect"); n != 1 {
		t.Errorf("transforming at night made more Insects: %d, want 1 (nothing entered, CR 712.18)", n)
	}
}

// Village Reavers' grant covers Wolves and Werewolves you control,
// itself included, and nothing else.
func TestVillageReaversGivesHasteToYourWolvesAndWerewolves(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[0]
	id := importAndCast(t, g, villageWatchRow(), p)
	monSettle(t, g)
	wolf := pushCatalogPermanent(g, p.ID, "Wolf", "Creature — Wolf", "", false)
	bear := pushCatalogPermanent(g, p.ID, "Bear", "Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, g.Seats[1].ID, "Their Wolf", "Creature — Wolf", "", false)

	if slices.Contains(effectiveAbilities(t, g, wolf), "haste") {
		t.Fatal("a Wolf had haste by day, before Village Reavers")
	}
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	if !slices.Contains(effectiveAbilities(t, g, id), "haste") {
		t.Error("Village Reavers does not have haste itself (it is a Werewolf)")
	}
	if !slices.Contains(effectiveAbilities(t, g, wolf), "haste") {
		t.Error("your Wolf did not get haste")
	}
	if slices.Contains(effectiveAbilities(t, g, bear), "haste") {
		t.Error("your Bear got haste")
	}
	if slices.Contains(effectiveAbilities(t, g, theirs), "haste") {
		t.Error("an opponent's Wolf got haste: the grant says 'you control'")
	}
}
