package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// top_of_library_cards_test.go — #1600: Crystal Skull, Isu Spyglass
// and the play-from-the-top family cards that ride the same
// CastPermission + LibraryTopVisible shape (Magus of the Future,
// Garruk's Horde, Vizier of the Menagerie, Radha, Heart of Keld).

const (
	crystalSkullOracle = "b1f8eb7b-7e01-4bf8-a557-64e16a6052af"
	magusFutureOracle  = "10758aa7-0f64-499f-8010-514e8fc31b7e"
	garruksHordeOracle = "b9da511e-a57f-43c2-aa97-902f3b4c55fb"
	vizierMenOracle    = "4f890d42-e4f1-4eed-aa6c-718865de8927"
	radhaKeldOracle    = "927fb139-0493-48be-8cc4-2c7d854d6a55"
)

// --- Crystal Skull, Isu Spyglass -----------------------------------

func TestCrystalSkullTopCardIsPrivateToItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushTopSource(t, g, me.ID, "Crystal Skull, Isu Spyglass", "Legendary Artifact", crystalSkullOracle)
	if !effectstest.TopVisibleTo(g, me.ID, me.ID) {
		t.Errorf("the Skull's controller cannot look at the top card")
	}
	if effectstest.TopVisibleTo(g, me.ID, opp.ID) {
		t.Errorf("an opponent can see the top card — \"you may look\" is private")
	}
}

func TestCrystalSkullOpensHistoricCardsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	effectstest.PushTopSource(t, g, me.ID, "Crystal Skull, Isu Spyglass", "Legendary Artifact", crystalSkullOracle)

	for _, tc := range []struct {
		name string
		card game.Card
		want bool
	}{
		{"artifact", game.Card{Name: "Top Rock", TypeLine: "Artifact", ManaCost: "{0}"}, true},
		{"legendary creature", game.Card{Name: "Top Legend", TypeLine: "Legendary Creature — Elf", ManaCost: "{0}"}, true},
		{"saga", game.Card{Name: "Top Saga", TypeLine: "Enchantment — Saga", ManaCost: "{0}"}, true},
		{"legendary land", game.Card{Name: "Top Legendary Land", TypeLine: "Legendary Land"}, true},
		{"artifact land", game.Card{Name: "Top Artifact Land", TypeLine: "Artifact Land"}, true},
		{"plain creature", game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"}, false},
		{"plain sorcery", game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"}, false},
		{"basic land", game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			top := effectstest.SeedTopOfLibrary(me, tc.card)
			if got := effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneLibrary).Granted(); got != tc.want {
				t.Fatalf("permission granted = %v, want %v", got, tc.want)
			}
			if !tc.want {
				if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library"}); err == nil {
					t.Errorf("a non-historic card was played off the top")
				}
				if !me.Library.Contains(top) {
					t.Errorf("a refused play moved the card")
				}
			}
		})
	}
}

func TestCrystalSkullPlaysAHistoricLandAndCastsAHistoricSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	effectstest.PushTopSource(t, g, me.ID, "Crystal Skull, Isu Spyglass", "Legendary Artifact", crystalSkullOracle)

	land := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Legendary Land", TypeLine: "Legendary Land"})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play the historic land off the top: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Errorf("the land never reached the battlefield")
	}

	rock := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Rock", TypeLine: "Artifact", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast the historic spell off the top: %v", err)
	}
	if me.Library.Contains(rock) {
		t.Errorf("the cast card is still in the library")
	}
}

func TestCrystalSkullPermissionEndsWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	skull := effectstest.PushTopSource(t, g, me.ID, "Crystal Skull, Isu Spyglass", "Legendary Artifact", crystalSkullOracle)
	top := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Rock", TypeLine: "Artifact", ManaCost: "{0}"})
	if !effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneLibrary).Granted() {
		t.Fatalf("no permission with the Skull out")
	}
	effectstest.RemoveFromBattlefield(g, skull)
	if effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneLibrary).Granted() {
		t.Errorf("the permission outlived the Skull")
	}
	if effectstest.TopVisibleTo(g, me.ID, me.ID) {
		t.Errorf("the look outlived the Skull")
	}
}

func TestCrystalSkullTapsForBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	skull := effectstest.PushTopSource(t, g, me.ID, "Crystal Skull, Isu Spyglass", "Legendary Artifact", crystalSkullOracle)
	if err := g.ActivateManaAbility(me.ID, skull, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := len(me.ManaPool); got != 1 {
		t.Errorf("pool = %d mana, want 1", got)
	}
}

// --- Magus of the Future -------------------------------------------

func TestMagusOfTheFuturePlaysLandsAndCastsSpellsFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ToMain(t, g)
	magus := effectstest.PushTopSource(t, g, me.ID, "Magus of the Future", "Creature — Human Wizard", magusFutureOracle)
	if !effectstest.TopVisibleTo(g, me.ID, opp.ID) {
		t.Errorf("an opponent cannot see a revealed top card")
	}

	land := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play a land off the top: %v", err)
	}
	spell := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"})
	if !effectstest.GrantedPermissionOn(g, me.ID, spell, game.ZoneLibrary).Granted() {
		t.Fatalf("a spell on top was not castable")
	}
	effectstest.RemoveFromBattlefield(g, magus)
	if effectstest.GrantedPermissionOn(g, me.ID, spell, game.ZoneLibrary).Granted() {
		t.Errorf("the permission outlived the Magus")
	}
}

// --- Garruk's Horde ------------------------------------------------

func TestGarruksHordeCastsCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ToMain(t, g)
	horde := effectstest.PushTopSource(t, g, me.ID, "Garruk's Horde", "Creature — Beast", garruksHordeOracle)
	if !effectstest.TopVisibleTo(g, me.ID, opp.ID) {
		t.Errorf("the Horde's top card should be revealed to the table")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, horde, "trample") {
		t.Errorf("the Horde lost trample")
	}

	bear := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast a creature off the top: %v", err)
	}
	for _, c := range []game.Card{
		{Name: "Top Forest", TypeLine: "Basic Land — Forest"},
		{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"},
	} {
		top := effectstest.SeedTopOfLibrary(me, c)
		if effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneLibrary).Granted() {
			t.Errorf("%s was opened by a creature-only permission", c.Name)
		}
		if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library"}); err == nil {
			t.Errorf("%s was played off the top", c.Name)
		}
	}
	effectstest.RemoveFromBattlefield(g, horde)
	top := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneLibrary).Granted() {
		t.Errorf("the permission outlived the Horde")
	}
}

// --- Vizier of the Menagerie ---------------------------------------

func TestVizierOfTheMenagerieCastsCreaturesFromAPrivateTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ToMain(t, g)
	vizier := effectstest.PushTopSource(t, g, me.ID, "Vizier of the Menagerie", "Creature — Snake Cleric", vizierMenOracle)
	if !effectstest.TopVisibleTo(g, me.ID, me.ID) || effectstest.TopVisibleTo(g, me.ID, opp.ID) {
		t.Errorf("Vizier's look should be private to its controller")
	}

	bear := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast a creature off the top: %v", err)
	}
	land := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Errorf("a land was played off the top under a creature-only permission")
	}
	effectstest.RemoveFromBattlefield(g, vizier)
	top := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneLibrary).Granted() {
		t.Errorf("the permission outlived the Vizier")
	}
}

// "You can spend mana of any type to cast creature spells": a red
// mana pays a creature's {U}, and does not pay a noncreature spell's.
func TestVizierOfTheMenagerieLetsAnyManaPayForCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	effectstest.PushTopSource(t, g, me.ID, "Vizier of the Menagerie", "Creature — Snake Cleric", vizierMenOracle)

	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	merfolk := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Merfolk", TypeLine: "Creature — Merfolk", ManaCost: "{U}", Colors: []string{"U"}})
	if err := g.CastSpell(me.ID, merfolk, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("pay a creature's {U} with red mana: %v", err)
	}
}

// --- Radha, Heart of Keld ------------------------------------------

func TestRadhaPlaysLandsFromAPrivateTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ToMain(t, g)
	radha := effectstest.PushTopSource(t, g, me.ID, "Radha, Heart of Keld", "Legendary Creature — Elf Warrior", radhaKeldOracle)
	if !effectstest.TopVisibleTo(g, me.ID, me.ID) || effectstest.TopVisibleTo(g, me.ID, opp.ID) {
		t.Errorf("Radha's look should be private to her controller")
	}
	land := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play a land off the top: %v", err)
	}
	spell := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if effectstest.GrantedPermissionOn(g, me.ID, spell, game.ZoneLibrary).Granted() {
		t.Errorf("a spell on top was opened by a lands-only permission")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Errorf("a creature was cast off the top under a lands-only permission")
	}
	effectstest.RemoveFromBattlefield(g, radha)
	land2 := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Island", TypeLine: "Basic Land — Island"})
	if effectstest.GrantedPermissionOn(g, me.ID, land2, game.ZoneLibrary).Granted() {
		t.Errorf("the permission outlived Radha")
	}
}

func TestRadhaHasFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	radha := effectstest.PushTopSource(t, g, me.ID, "Radha, Heart of Keld", "Legendary Creature — Elf Warrior", radhaKeldOracle)
	if !effectstest.EffectiveAbilitiesContain(t, g, radha, "first strike") {
		t.Errorf("no first strike on its controller's turn")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, radha, "first strike") {
		t.Errorf("first strike on another player's turn")
	}
}

func TestRadhaPumpsByLandsControlled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	radha := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Radha, Heart of Keld", TypeLine: "Legendary Creature — Elf Warrior",
		OracleID: radhaKeldOracle, Owner: me.ID, Controller: me.ID, Power: 3, Toughness: 3,
	})
	for i := 0; i < 3; i++ {
		effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	}
	me.ManaPool.AddMana(
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"},
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "R"}, game.ManaToken{Color: "G"})
	if err := g.ActivateCatalogAbility(me.ID, radha, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, radha); got != 6 {
		t.Errorf("power after the pump = %d, want 3 + 3 lands", got)
	}
}
