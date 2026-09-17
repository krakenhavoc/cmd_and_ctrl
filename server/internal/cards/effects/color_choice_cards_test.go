package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// color_choice_cards_test.go — the catalog cards #742's "choose a
// color" prompt and one-pick-N-mana amount unblocked.

// --- helpers --------------------------------------------------------

// pendingOfKind returns the first open choice of `kind`, or nil.
func pendingOfKind(g *game.Game, kind game.PendingChoiceKind) *game.PendingChoice {
	var out *game.PendingChoice
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == kind {
				out = c
				return
			}
		}
	})
	return out
}

// answerColor answers the open choose_color prompt owed by `chooser`.
func answerColor(t *testing.T, g *game.Game, chooser uuid.UUID, color string) *game.PendingChoice {
	t.Helper()
	c := pendingOfKind(g, game.PendingChoiceColor)
	if c == nil {
		t.Fatalf("no choose_color prompt is open (answering %s)", color)
	}
	if c.Chooser != chooser {
		t.Fatalf("the open choose_color prompt belongs to %s, not %s", c.Chooser, chooser)
	}
	if err := g.ResolveColorChoice(c.ID, chooser, color); err != nil {
		t.Fatalf("ResolveColorChoice(%s): %v", color, err)
	}
	return c
}

// pushChosenColorPermanent seeds a catalog permanent and answers its
// "as this enters, choose a color" prompt, the way
// pushNamedTribePermanent does for creature types.
func pushChosenColorPermanent(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracleID, color string) uuid.UUID {
	t.Helper()
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() { g.QueueColorChoiceForEffect(owner, id, name, nil) })
	answerColor(t, g, owner, color)
	return id
}

func untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
}

// --- stored colour: lands --------------------------------------------

// Thriving Isle end to end: played from hand it enters tapped and asks
// for a colour other than blue; afterwards it taps for blue or that
// colour, as one two-option pick.
func TestThrivingIsleEntersTappedAndTapsForBlueOrTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	isle := b13PlayAs(t, g, 0, "Thriving Isle", "Land", "69fc70b8-b143-4662-ac95-e2743037239d")

	if !b13Tapped(t, g, isle) {
		t.Error("Thriving Isle entered untapped")
	}
	prompt := pendingOfKind(g, game.PendingChoiceColor)
	if prompt == nil {
		t.Fatal("no colour prompt as the land entered")
	}
	if want := []string{"W", "B", "R", "G"}; !reflect.DeepEqual(prompt.ColorOptions, want) {
		t.Errorf("options = %v, want %v (never blue)", prompt.ColorOptions, want)
	}
	if err := g.ResolveColorChoice(prompt.ID, me.ID, "U"); err == nil {
		t.Error("blue was accepted for a color other than blue")
	}
	answerColor(t, g, me.ID, "R")

	untap(g, isle)
	if err := g.ActivateManaAbility(me.ID, isle, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || !reflect.DeepEqual(pick.ColorOptions, []string{"U", "R"}) {
		t.Fatalf("mana pick = %+v, want {U} or {R}", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(g.Seats[0]); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want [R]", got)
	}
}

// Every land in the two tables: a tapped entry, the right prompt, and
// the right mana once answered.
func TestChosenColorLandsTapForTheirColors(t *testing.T) {
	type land struct {
		name, oracle, printed string // printed "" = one-colour land
	}
	lands := []land{
		{"Thriving Heath", "d1946630-e224-40db-8f0d-388b09622288", "W"},
		{"Thriving Grove", "a8052556-8962-4130-86a8-6fb7b6a324f7", "G"},
		{"Thriving Bluff", "91fceb34-0f2d-4392-be27-00dcd765637f", "R"},
		{"Thriving Moor", "bff416bb-d193-4c45-b2c1-7c297dbfad08", "B"},
		{"Sea Gate", "b574c540-9f8a-4fd4-8809-d02c9b099ddc", "U"},
		{"Citadel Gate", "15f1fe23-5af4-4fc4-8cde-2e0bf9f9be0c", "W"},
		{"Manor Gate", "dd6e67c0-66a1-49b7-8a86-3cf4b209fd07", "G"},
		{"Cliffgate", "1999b5ac-21fb-4d99-ad72-58bf507f9a59", "R"},
		{"Black Dragon Gate", "dde6bce5-8bbe-4866-b5aa-2c05c7d37241", "B"},
		{"Uncharted Haven", "d23c3613-bc5e-4fc5-939c-62a090c53a79", ""},
		{"Crossroads Village", "b26cfeb0-7bbe-4d93-8eed-e832f175a80c", ""},
		{"Mirage Mesa", "e103f422-85c0-43f8-8a2f-8b7863e503fa", ""},
		{"Valgavoth's Lair", "660d44a2-391a-416c-b46c-ddcc3739f527", ""},
	}
	for _, l := range lands {
		t.Run(l.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := b13PlayAs(t, g, 0, l.name, "Land", l.oracle)
			if !b13Tapped(t, g, id) {
				t.Error("entered untapped")
			}
			prompt := pendingOfKind(g, game.PendingChoiceColor)
			if prompt == nil {
				t.Fatal("no colour prompt")
			}
			wantOptions := game.AllColors
			if l.printed != "" {
				wantOptions = game.ColorsOtherThan(l.printed)
			}
			if !reflect.DeepEqual(prompt.ColorOptions, wantOptions) {
				t.Errorf("options = %v, want %v", prompt.ColorOptions, wantOptions)
			}
			chosen := wantOptions[len(wantOptions)-1]
			answerColor(t, g, me.ID, chosen)
			untap(g, id)
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			if l.printed == "" {
				if got := poolColors(g.Seats[0]); !reflect.DeepEqual(got, []string{chosen}) {
					t.Errorf("pool = %v, want [%s]", got, chosen)
				}
				return
			}
			pick := pendingOfKind(g, game.PendingChoiceMana)
			if pick == nil || !reflect.DeepEqual(pick.ColorOptions, []string{l.printed, chosen}) {
				t.Fatalf("mana pick = %+v, want {%s} or {%s}", pick, l.printed, chosen)
			}
		})
	}
}

// Before the colour is chosen a one-colour land adds nothing — never
// "any colour".
func TestChosenColorManaIsNothingBeforeTheAnswer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	haven := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Uncharted Haven", TypeLine: "Land",
		OracleID: "d23c3613-bc5e-4fc5-939c-62a090c53a79", Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateManaAbility(me.ID, haven, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(g.Seats[0]); len(got) != 0 {
		t.Errorf("pool = %v, want nothing while no colour is chosen", got)
	}
}

// --- stored colour: artifacts --------------------------------------

func TestColdsteelHeartEntersTappedAndTapsForTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	heart := castCatalogSpell(t, g, "Coldsteel Heart", "Snow Artifact", "21d700e8-0255-418e-96a0-6fe05b3f836d", nil)
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, heart) {
		t.Error("Coldsteel Heart entered untapped")
	}
	answerColor(t, g, me.ID, "G")
	untap(g, heart)
	if err := g.ActivateManaAbility(me.ID, heart, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(g.Seats[0]); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool = %v, want [G] with no pick", got)
	}
	if pendingOfKind(g, game.PendingChoiceMana) != nil {
		t.Error("one mana of the chosen color queued a pick")
	}
}

func TestHeraldicBannerPumpsYourCreaturesOfTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	redMine := b31Push(g, me, "Goblin Guide", "Creature — Goblin Scout", "", "{R}", 2, 2, "R")
	rakdos := b31Push(g, me, "Rakdos Cackler", "Creature — Devil", "", "{B/R}", 2, 2, "B", "R")
	greenMine := b31Push(g, me, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	redTheirs := b31Push(g, them, "Raging Goblin", "Creature — Goblin Berserker", "", "{R}", 1, 1, "R")

	banner := pushChosenColorPermanent(t, g, me, "Heraldic Banner", "Artifact", "3525e263-e29a-49bf-a29f-fb3ce43bbd33", "R")

	for id, want := range map[uuid.UUID]int{redMine: 3, rakdos: 3, greenMine: 2, redTheirs: 1} {
		if got := effectivePower(t, g, id); got != want {
			t.Errorf("power of %s = %d, want %d", id, got, want)
		}
	}
	if got := effectiveToughness(t, g, redMine); got != 2 {
		t.Errorf("toughness = %d, want 2 (+1/+0)", got)
	}
	if err := g.ActivateManaAbility(me, banner, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(g.Seats[0]); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want [R]", got)
	}
}

// --- one pick, N tokens ---------------------------------------------

func TestGildedLotusAddsThreeOfOneColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	lotus := b31Push(g, me, "Gilded Lotus", "Artifact", "9a02a9a7-39d9-4763-85d3-747a0540b60b", "{5}", 0, 0)
	if err := g.ActivateManaAbility(me, lotus, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || len(pick.ColorOptions) != 5 || pick.ManaAmounts["W"] != 3 {
		t.Fatalf("pick = %+v, want one five-colour pick of three", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me, "B"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(g.Seats[0]); !reflect.DeepEqual(got, []string{"B", "B", "B"}) {
		t.Errorf("pool = %v, want three {B}", got)
	}
}

func TestLotusFieldEntersTappedAndSacrificesTwoLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	b31Push(g, me.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	field := b13PlayAs(t, g, 0, "Lotus Field", "Land", "134d5b82-7940-4b33-a922-7f9d1f403e50")
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, field) {
		t.Error("Lotus Field entered untapped")
	}
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceSacrifice && c.Chooser == me.ID {
				n++
			}
		}
	})
	if n != 2 {
		t.Errorf("sacrifice prompts = %d, want 2", n)
	}
}

func TestLotusFieldTapsForThreeOfOneColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	field := b31Push(g, me, "Lotus Field", "Land", "134d5b82-7940-4b33-a922-7f9d1f403e50", "", 0, 0)
	if err := g.ActivateManaAbility(me, field, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no pick")
	}
	if err := g.ResolveManaChoice(pick.ID, me, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(g.Seats[0]); !reflect.DeepEqual(got, []string{"G", "G", "G"}) {
		t.Errorf("pool = %v, want three {G}", got)
	}
}

func TestNyxLotusAddsDevotionToTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	b31Push(g, me, "Nyxbloom", "Creature — Elf", "", "{2}{G}{G}", 2, 2, "G")
	b31Push(g, me, "Kitchen Finks", "Creature — Ouphe", "", "{1}{G/W}{G/W}", 3, 2, "G", "W")
	b31Push(g, me, "Ponder Imp", "Creature — Imp", "", "{U}", 1, 1, "U")
	lotus := b31Push(g, me, "Nyx Lotus", "Legendary Artifact", "cfb8cdde-11f6-441a-942b-32d8c191fc90", "{4}", 0, 0)

	if err := g.ActivateManaAbility(me, lotus, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no pick")
	}
	// Devotion: green 4 (two {G} plus two hybrid), white 2, blue 1;
	// black and red 0 are not offered.
	if want := []string{"W", "U", "G"}; !reflect.DeepEqual(pick.ColorOptions, want) {
		t.Errorf("options = %v, want %v", pick.ColorOptions, want)
	}
	if want := map[string]int{"W": 2, "U": 1, "G": 4}; !reflect.DeepEqual(pick.ManaAmounts, want) {
		t.Errorf("amounts = %v, want %v", pick.ManaAmounts, want)
	}
	if err := g.ResolveManaChoice(pick.ID, me, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(g.Seats[0]); !reflect.DeepEqual(got, []string{"G", "G", "G", "G"}) {
		t.Errorf("pool = %v, want four {G}", got)
	}
}

func TestNyxLotusEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	lotus := castCatalogSpell(t, g, "Nyx Lotus", "Legendary Artifact", "cfb8cdde-11f6-441a-942b-32d8c191fc90", nil)
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, lotus) {
		t.Error("Nyx Lotus entered untapped")
	}
}

func TestMonaLisaAddsHerPowerInOneColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	mona := b31Push(g, me, "Mona Lisa, Science Geek", "Legendary Creature — Lizard Mutant",
		"e6490e98-37a9-4a8b-a04c-cceeb22b0c35", "{2}{G}", 3, 3, "G")
	if err := g.ActivateManaAbility(me, mona, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || pick.ManaAmounts["U"] != 3 {
		t.Fatalf("pick = %+v, want amounts of three", pick)
	}
}

func TestWhiteLotusTileCountsTheLargestSharedType(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	b31Push(g, me, "Llanowar Elves", "Creature — Elf Druid", "", "{G}", 1, 1, "G")
	b31Push(g, me, "Elvish Visionary", "Creature — Elf Shaman", "", "{1}{G}", 1, 1, "G")
	b31Push(g, me, "Fauna Shaman", "Creature — Elf Shaman", "", "{1}{G}", 2, 2, "G")
	b31Push(g, me, "Goblin Guide", "Creature — Goblin Scout", "", "{R}", 2, 2, "R")
	b31Push(g, them, "Elvish Mystic", "Creature — Elf Druid", "", "{G}", 1, 1, "G")
	shapeshifter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Universal Automaton", TypeLine: "Artifact Creature — Shapeshifter",
		Power: 1, Toughness: 1, Keywords: []string{"changeling"}, Owner: me, Controller: me,
	})
	_ = shapeshifter
	tile := b31Push(g, me, "White Lotus Tile", "Artifact", "f8a5e009-45b7-4e62-a494-1579f5fc0ba6", "{4}", 0, 0)

	if err := g.ActivateManaAbility(me, tile, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	// Three Elves plus the changeling; the opponent's Elf does not count.
	if pick == nil || pick.ManaAmounts["G"] != 4 {
		t.Fatalf("pick = %+v, want amounts of four", pick)
	}
}

func TestIlysianCaryatidUpgradesWithAFourPowerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	caryatid := b31Push(g, me, "Ilysian Caryatid", "Creature — Plant", "5f2be3c2-060a-43e1-b63b-9cd3c78ffcb0", "{1}{G}", 1, 1, "G")
	if err := g.ActivateManaAbility(me, caryatid, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || len(pick.ManaAmounts) != 0 || len(pick.ColorOptions) != 5 {
		t.Fatalf("without a 4-power creature: pick = %+v, want one mana of any color", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me, "W"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}

	b31Push(g, me, "Craterhoof", "Creature — Beast", "", "{5}{G}{G}{G}", 5, 5, "G")
	untap(g, caryatid)
	if err := g.ActivateManaAbility(me, caryatid, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick = pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || pick.ManaAmounts["U"] != 2 {
		t.Fatalf("with a 4-power creature: pick = %+v, want two of any one color", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(g.Seats[0]); !reflect.DeepEqual(got, []string{"W", "U", "U"}) {
		t.Errorf("pool = %v, want [W U U]", got)
	}
}

func TestSanctumOfFruitfulHarvestAddsOneColorPerShrine(t *testing.T) {
	g := newCatalogGame(t) // cursor at seat 0's draw step
	me := g.Seats[0].ID
	b31Push(g, me, "Sanctum of Fruitful Harvest", "Legendary Enchantment — Shrine", "132859dd-de66-45c6-8af4-ab5e202a17b0", "{2}{G}", 0, 0)
	b31Push(g, me, "Honden of Life's Web", "Legendary Enchantment — Shrine", "", "{4}{G}", 0, 0)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if g.Turn.Step != game.StepPrecombatMain {
		t.Fatalf("step = %v, want precombat main", g.Turn.Step)
	}
	passPriorityAroundTable(t, g)
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || pick.ManaAmounts["G"] != 2 {
		t.Fatalf("pick = %+v, want two of any one color", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(g.Seats[0]); !reflect.DeepEqual(got, []string{"G", "G"}) {
		t.Errorf("pool = %v, want two {G}", got)
	}
}

// --- colour chosen at resolution ------------------------------------

func TestWashOutReturnsPermanentsOfTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	red := b31Push(g, them.ID, "Goblin Guide", "Creature — Goblin Scout", "", "{R}", 2, 2, "R")
	gruul := b31Push(g, me.ID, "Burning-Tree Emissary", "Creature — Elemental", "", "{R/G}{R/G}", 2, 2, "R", "G")
	green := b31Push(g, them.ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	colorless := b31Push(g, me.ID, "Sol Ring", "Artifact", "", "{1}", 0, 0)

	castCatalogSpell(t, g, "Wash Out", "Sorcery", "54748cb1-d92a-4212-ad76-417ee79b5ef1", nil)
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "R")

	for id, wantGone := range map[uuid.UUID]bool{red: true, gruul: true, green: false, colorless: false} {
		_, onField := battlefieldCard(g, id)
		if onField == wantGone {
			t.Errorf("%s on the battlefield = %v, want %v", id, onField, !wantGone)
		}
	}
	if !them.Hand.Contains(red) || !me.Hand.Contains(gruul) {
		t.Error("the returned permanents did not go to their owners' hands")
	}
}

func TestSelectiveObliterationAsksEachPlayerThenExiles(t *testing.T) {
	g := newCatalogGame(t)
	seats := g.Seats
	// Seat 0: a mono-white and a white-blue permanent, chooses white.
	whiteMine := b31Push(g, seats[0].ID, "Savannah Lions", "Creature — Cat", "", "{W}", 2, 1, "W")
	azorius := b31Push(g, seats[0].ID, "Azorius Guildmage", "Creature — Vedalken Wizard", "", "{W/U}{W/U}", 2, 2, "W", "U")
	// Seat 1: a green creature, chooses blue.
	greenTheirs := b31Push(g, seats[1].ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	// Seat 2: a colourless rock survives whatever they choose.
	rock := b31Push(g, seats[2].ID, "Sol Ring", "Artifact", "", "{1}", 0, 0)
	// Seat 3: a red creature, chooses red.
	redTheirs := b31Push(g, seats[3].ID, "Goblin Guide", "Creature — Goblin Scout", "", "{R}", 2, 2, "R")

	castCatalogSpell(t, g, "Selective Obliteration", "Sorcery", "9f95027a-5f04-48b4-99a0-8913be3a520a", nil)
	passPriorityAroundTable(t, g)

	// APNAP: the active player (seat 0) first, then turn order.
	answerColor(t, g, seats[0].ID, "W")
	for id := range map[uuid.UUID]bool{whiteMine: true, azorius: true, greenTheirs: true} {
		if _, ok := battlefieldCard(g, id); !ok {
			t.Fatal("something was exiled before every player had chosen")
		}
	}
	answerColor(t, g, seats[1].ID, "U")
	answerColor(t, g, seats[2].ID, "B")
	answerColor(t, g, seats[3].ID, "R")

	for id, wantKept := range map[uuid.UUID]bool{
		whiteMine: true, azorius: false, greenTheirs: false, rock: true, redTheirs: true,
	} {
		if _, ok := battlefieldCard(g, id); ok != wantKept {
			t.Errorf("%s kept = %v, want %v", id, ok, wantKept)
		}
	}
	if pendingOfKind(g, game.PendingChoiceColor) != nil {
		t.Error("a colour prompt is still open after every player answered")
	}
}

func TestOonaExilesAndMakesFaeriesForTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	aangAdvanceToMain(t, g, 0)
	oona := b31Push(g, me.ID, "Oona, Queen of the Fae", "Legendary Creature — Faerie Wizard",
		"6052822d-47a2-4d69-a32d-40cdd600d7a9", "{3}{U/B}{U/B}{U/B}", 5, 5, "U", "B")
	// The top of a library is the LAST element.
	them.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: them.ID, Controller: them.ID})
	them.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Counterspell", TypeLine: "Instant", ManaCost: "{U}{U}", Owner: them.ID, Controller: them.ID})
	them.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Brainstorm", TypeLine: "Instant", ManaCost: "{U}", Owner: them.ID, Controller: them.ID})
	them.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: them.ID, Controller: them.ID})
	b31AddMana(me, "U", "U", "U", "U")

	if err := g.ActivateCatalogAbility(me.ID, oona, 0, game.ActivateAbilityParams{
		XValue:  3,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	libBefore := them.Library.Size()
	answerColor(t, g, me.ID, "U")

	if got := libBefore - them.Library.Size(); got != 3 {
		t.Errorf("exiled %d cards, want X=3", got)
	}
	faeries := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Faerie Rogue" && c.Controller == me.ID {
			faeries++
			if !reflect.DeepEqual(c.Colors, []string{"U", "B"}) {
				t.Errorf("token colours = %v, want blue and black", c.Colors)
			}
		}
	}
	// Lightning Bolt (red), Brainstorm (blue), Counterspell (blue).
	if faeries != 2 {
		t.Errorf("Faerie Rogues = %d, want 2", faeries)
	}
}
