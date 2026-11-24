package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b017_cards_test.go — Secret Lair Drop: Lands (S82 older sets,
// batch 017, #5462). Each test plays the real catalog entry.

const (
	oldB17ScryingSheets = "0f98e055-ab61-4314-aae8-9d3c19f66acf"
	oldB17Howltooth     = "463fc699-f4fc-4112-a6b3-6dcb642203e6"
	oldB17Stardew       = "6a4ee425-b3b8-487d-866c-9e2d73682466"
	oldB17Havengul      = "e71ac446-02a4-4468-8d29-f28b21617665"
	oldB17Shelldock     = "f748b2fb-6c2a-400a-8e96-fa4e4a1dfe80"
	oldB17Azcanta       = "f74c4d96-bc4a-4d32-9519-a753d192144e"
)

func TestOldB17EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB17ScryingSheets:                    "Scrying Sheets",
		oldB17Howltooth:                        "Howltooth Hollow",
		"5b2460a5-6ae5-4cad-ba94-1a9e98e6e4c0": "Snow-Covered Island",
		"5f0d3be8-e63e-4ade-ae58-6b0c14f2ce6d": "Snow-Covered Forest",
		oldB17Stardew:                          "Stardew Valley",
		"ac8cc74d-e43b-4118-bba0-dfa8b9c04d45": "Snow-Covered Plains",
		"ca9f660b-e07d-4f42-a46e-abd0ca72510c": "Snow-Covered Mountain",
		"d8239a86-7184-4005-ba1e-2dddcd756c47": "Snow-Covered Swamp",
		oldB17Havengul:                         "Havengul Laboratory",
		oldB17Havengul + "#1":                  "Havengul Mystery",
		oldB17Shelldock:                        "Shelldock Isle",
		oldB17Azcanta:                          "Search for Azcanta",
		oldB17Azcanta + "#1":                   "Azcanta, the Sunken Ruin",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// A Snow-Covered basic taps for its colour and the mana pays {S}; the
// nonsnow Island's mana does not (CR 107.4h).
func TestOldB17SnowBasicsPayForSnowCosts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sheets := effectstest.BlbB1Push(g, me.ID, "Scrying Sheets", "Snow Land", oldB17ScryingSheets, 0, 0)
	plain := effectstest.BlbB1Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
	plain2 := effectstest.BlbB1Push(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, sheets, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("{1}{S} auto-tapped from nonsnow Islands: accepted")
	}
	if effectstest.FindBattlefieldCardByID(g, plain).Tapped || effectstest.FindBattlefieldCardByID(g, plain2).Tapped {
		t.Fatal("a refused activation tapped a land")
	}
	for _, tc := range []struct{ name, oracle, sub string }{
		{"Snow-Covered Island", "5b2460a5-6ae5-4cad-ba94-1a9e98e6e4c0", "Island"},
		{"Snow-Covered Forest", "5f0d3be8-e63e-4ade-ae58-6b0c14f2ce6d", "Forest"},
		{"Snow-Covered Plains", "ac8cc74d-e43b-4118-bba0-dfa8b9c04d45", "Plains"},
		{"Snow-Covered Mountain", "ca9f660b-e07d-4f42-a46e-abd0ca72510c", "Mountain"},
		{"Snow-Covered Swamp", "d8239a86-7184-4005-ba1e-2dddcd756c47", "Swamp"},
	} {
		id := effectstest.BlbB1Push(g, me.ID, tc.name, "Basic Snow Land — "+tc.sub, tc.oracle, 0, 0)
		c := effectstest.FindBattlefieldCardByID(g, id)
		if !c.HasSupertype("snow") || !c.HasSupertype("basic") {
			t.Errorf("%s is not a basic snow land", tc.name)
		}
	}
}

func oldB17Sheets(t *testing.T, g *game.Game, me *game.Player, top struct{ Name, TypeLine string }) (uuid.UUID, uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	sheets := effectstest.BlbB1Push(g, me.ID, "Scrying Sheets", "Snow Land", oldB17ScryingSheets, 0, 0)
	effectstest.BlbB1Push(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island", "5b2460a5-6ae5-4cad-ba94-1a9e98e6e4c0", 0, 0)
	effectstest.BlbB1Push(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "5f0d3be8-e63e-4ade-ae58-6b0c14f2ce6d", 0, 0)
	ids := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{top})
	if err := g.ActivateCatalogAbility(me.ID, sheets, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("activate Scrying Sheets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return sheets, ids[0]
}

func TestOldB17ScryingSheetsMayTakeASnowCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, top := oldB17Sheets(t, g, me, struct{ Name, TypeLine string }{"Snow Bear", "Snow Creature — Bear"})
	effectstest.B35AnswerOption(t, g, me.ID, "Reveal it and put it into your hand")
	if !me.Hand.Contains(top) {
		t.Fatal("the snow card is not in hand")
	}
}

func TestOldB17ScryingSheetsMayLeaveASnowCardOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, top := oldB17Sheets(t, g, me, struct{ Name, TypeLine string }{"Snow Bear", "Snow Creature — Bear"})
	effectstest.B35AnswerOption(t, g, me.ID, "Leave it on top")
	if me.Hand.Contains(top) || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != top {
		t.Fatal("the declined card moved")
	}
}

func TestOldB17ScryingSheetsOffersNothingForANonsnowCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, top := oldB17Sheets(t, g, me, struct{ Name, TypeLine string }{"Bear", "Creature — Bear"})
	if c := effectstest.LatestOptionPickFor(g, me.ID); c != nil {
		t.Fatalf("a nonsnow card was offered: %+v", c.PickOptions)
	}
	if me.Hand.Contains(top) {
		t.Fatal("a nonsnow card went to hand")
	}
}

func oldB17Hideaway(t *testing.T, g *game.Game, name, oracle string) (*game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	top := effectstest.StackTopLibrary(me, "A", "B", "C", "D")
	land := game.NewCard(name, me.ID)
	land.OracleID = oracle
	land.TypeLine = "Land"
	effectstest.EnterWithHideaway(t, g, land, top[0])
	return me, land.InstanceID, top[0]
}

func oldB17ActivateHidden(t *testing.T, g *game.Game, me *game.Player, land uuid.UUID, color string) {
	t.Helper()
	effectstest.UntapForTest(g, land)
	effectstest.FloatForTest(g, me, color)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB17HowltoothHollowNeedsEveryHandEmpty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, land, hidden := oldB17Hideaway(t, g, "Howltooth Hollow", oldB17Howltooth)
	if c, _ := g.LookupCardForEffect(land); !c.Tapped {
		t.Error("Howltooth Hollow did not enter tapped")
	}
	g.Seats[1].Hand.PushTop(game.NewCard("Held", g.Seats[1].ID))
	oldB17ActivateHidden(t, g, me, land, "B")
	if effectstest.PlayHiddenGrant(g, hidden) != nil {
		t.Fatal("granted while an opponent held a card")
	}
	for _, p := range g.Seats {
		p.Hand.Cards = nil
	}
	oldB17ActivateHidden(t, g, me, land, "B")
	if effectstest.PlayHiddenGrant(g, hidden) == nil {
		t.Fatal("no grant with every hand empty")
	}
}

func TestOldB17ShelldockIsleNeedsALibraryOfTwentyOrFewer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, land, hidden := oldB17Hideaway(t, g, "Shelldock Isle", oldB17Shelldock)
	for _, p := range g.Seats {
		for len(p.Library.Cards) < 30 {
			p.Library.PushTop(game.NewCard("Filler", p.ID))
		}
		p.Library.Cards = p.Library.Cards[:30]
	}
	oldB17ActivateHidden(t, g, me, land, "U")
	if effectstest.PlayHiddenGrant(g, hidden) != nil {
		t.Fatal("granted with every library above twenty")
	}
	g.Seats[2].Library.Cards = g.Seats[2].Library.Cards[:20]
	oldB17ActivateHidden(t, g, me, land, "U")
	if effectstest.PlayHiddenGrant(g, hidden) == nil {
		t.Fatal("no grant with one library at twenty")
	}
}

func TestOldB17StardewValleyMakesFoodByTappingACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	valley := effectstest.BlbB1Push(g, me.ID, "Stardew Valley", "Land", oldB17Stardew, 0, 0)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, valley, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("made Food with no creature to tap")
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, valley, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{bear}})
	effectstest.PassPriorityAroundTable(t, g)
	food := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Food" && c.Controller == me.ID {
			food++
		}
	}
	if food != 1 {
		t.Fatalf("Food tokens = %d, want 1", food)
	}
	if !effectstest.FindBattlefieldCardByID(g, bear).Tapped {
		t.Error("the creature was not tapped")
	}
}

func oldB17StardewGift(t *testing.T, g *game.Game) (me, opp *game.Player, gift uuid.UUID) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	valley := effectstest.BlbB1Push(g, me.ID, "Stardew Valley", "Land", oldB17Stardew, 0, 0)
	gift = effectstest.PushVanillaCreature(g, me.ID, "Gift", 3, 3)
	effectstest.B06AddMana(me, "C", "C", "C")
	handBefore := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, valley, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: gift}}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand %d, want %d: no card drawn", me.Hand.Size(), handBefore+1)
	}
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	return me, opp, gift
}

func TestOldB17StardewValleyGiftIsTakenWhenTheOtherPlayerAccepts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, gift := oldB17StardewGift(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Fatal("the activator was asked; the chosen player decides")
	}
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	if c, _ := g.LookupCardForEffect(gift); c.Controller != opp.ID {
		t.Fatalf("controller = %s, want the opponent", c.Controller)
	}
}

func TestOldB17StardewValleyGiftStaysWhenTheOtherPlayerDeclines(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, gift := oldB17StardewGift(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, false)
	if c, _ := g.LookupCardForEffect(gift); c.Controller != me.ID {
		t.Fatalf("controller = %s, want it to stay", c.Controller)
	}
}

func TestOldB17StardewValleyGiftIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	valley := effectstest.BlbB1Push(g, me.ID, "Stardew Valley", "Land", oldB17Stardew, 0, 0)
	gift := effectstest.PushVanillaCreature(g, me.ID, "Gift", 3, 3)
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, valley, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: gift}}}); err == nil {
		t.Fatal("activated the gift ability in the upkeep")
	}
}

func oldB17Clues(t *testing.T, g *game.Game, me *game.Player, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		id := effectstest.BlbB1Push(g, me.ID, "Clue", "Token Artifact — Clue", "", 0, 0)
		if err := g.SacrificePermanent(me.ID, id); err != nil {
			t.Fatalf("sacrifice Clue: %v", err)
		}
	}
}

func oldB17Lab(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Havengul Laboratory", OracleID: oldB17Havengul, Layout: game.LayoutTransform,
		Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Havengul Laboratory", TypeLine: "Legendary Land"},
			{Name: "Havengul Mystery", TypeLine: "Legendary Land"},
		},
	})
}

func TestOldB17HavengulLaboratoryInvestigatesAndTransformsAfterThreeClues(t *testing.T) {
	for _, tc := range []struct {
		clues int
		want  string
	}{{2, "Havengul Laboratory"}, {3, "Havengul Mystery"}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		lab := oldB17Lab(g, me)
		effectstest.B06AddMana(me, "C", "C", "C", "C")
		effectstest.B16Activate(t, g, me.ID, lab, 0, game.ActivateAbilityParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
		if len(effectstest.MkmMP2Clues(g, me.ID)) != 1 {
			t.Fatal("the laboratory did not investigate")
		}
		oldB17Clues(t, g, me, tc.clues)
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if c, _ := g.LookupCardForEffect(lab); c.Name != tc.want {
			t.Errorf("%d Clues sacrificed: the land is %q, want %q", tc.clues, c.Name, tc.want)
		}
	}
}

func TestOldB17HavengulMysteryReturnsACreatureAndTransformsWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	lab := oldB17Lab(g, me)
	dead := effectstest.MkmMP2Grave(me, "Fallen Bear", "Creature — Bear")
	oldB17Clues(t, g, me, 3)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(lab); c.Name != "Havengul Mystery" {
		t.Fatalf("the land is %q, want Havengul Mystery", c.Name)
	}
	effectstest.CloudDeckPickTargets(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(dead); !ok || !effects.OnBattlefield(g, dead) || c.Controller != me.ID {
		t.Fatal("the creature card was not returned to the battlefield")
	}
	if c, _ := g.LookupCardForEffect(lab); c.Name != "Havengul Mystery" {
		t.Fatalf("transformed early: %q", c.Name)
	}
	if err := g.SacrificePermanent(me.ID, dead); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(lab); c.Name != "Havengul Laboratory" {
		t.Errorf("after the creature left the land is %q, want Havengul Laboratory", c.Name)
	}
}

func TestOldB17HavengulMysteryTapsForBlackAtALifeCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mystery := effectstest.PushCatalogPermanent(g, me.ID, "Havengul Mystery", "Legendary Land", oldB17Havengul+"#1", false)
	life := me.Life
	effectstest.B28TapForMana(t, g, me.ID, mystery, "B")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Fatalf("pool %v, want {B}", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

func TestOldB17SearchForAzcantaSurveilsThenMayTransformAtSevenCards(t *testing.T) {
	for _, tc := range []struct {
		name       string
		graveyard  int
		answer     bool
		wantAsked  bool
		wantAzcant bool
	}{
		{"six cards: no question", 6, false, false, false},
		{"seven cards, declined", 7, false, true, false},
		{"seven cards, accepted", 7, true, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			search := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Search for Azcanta", OracleID: oldB17Azcanta, Layout: game.LayoutTransform,
				Owner: me.ID, Controller: me.ID,
				Faces: []game.Face{
					{Name: "Search for Azcanta", TypeLine: "Legendary Enchantment", ManaCost: "{1}{U}"},
					{Name: "Azcanta, the Sunken Ruin", TypeLine: "Legendary Land"},
				},
			})
			for i := 0; i < tc.graveyard-1; i++ {
				effectstest.MkmMP2Grave(me, "Filler", "Sorcery")
			}
			effectstest.StackTopLibrary(me, "Top")
			effectstest.BeginUpkeep(t, g, me.ID)
			sv := effectstest.EoeQZSurveil(t, g, me.ID, 1)
			// Put the surveilled card into the graveyard: that is the seventh.
			if err := g.ResolveSurveil(sv.ID, me.ID, sv.ScryCards, nil); err != nil {
				t.Fatalf("ResolveSurveil: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			asked := effectstest.LatestConfirmFor(g, me.ID) != nil
			if asked != tc.wantAsked {
				t.Fatalf("asked = %v, want %v (graveyard %d)", asked, tc.wantAsked, me.Graveyard.Size())
			}
			if asked {
				effectstest.AnswerMayChoice(t, g, me.ID, tc.answer)
			}
			c, _ := g.LookupCardForEffect(search)
			if got := c.Name == "Azcanta, the Sunken Ruin"; got != tc.wantAzcant {
				t.Errorf("land face up = %v, want %v (%s)", got, tc.wantAzcant, c.Name)
			}
		})
	}
}

func TestOldB17AzcantaTheSunkenRuinDigsForANoncreatureNonlandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ruin := effectstest.PushCatalogPermanent(g, me.ID, "Azcanta, the Sunken Ruin", "Legendary Land", oldB17Azcanta+"#1", false)
	top := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{
		{"Bear", "Creature — Bear"}, {"Forest", "Basic Land — Forest"}, {"Bolt", "Instant"}, {"Ring", "Artifact"},
	})
	effectstest.B06AddMana(me, "U", "U", "U")
	effectstest.B16Activate(t, g, me.ID, ruin, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no card choice")
	}
	for _, bad := range []uuid.UUID{top[0], top[1]} {
		for _, id := range c.ChooseCards {
			if id == bad {
				t.Fatalf("a creature or land card was offered")
			}
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top[2])
	if !me.Hand.Contains(top[2]) {
		t.Error("the instant is not in hand")
	}
}
