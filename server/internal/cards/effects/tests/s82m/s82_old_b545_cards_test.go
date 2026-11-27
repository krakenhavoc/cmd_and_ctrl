package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b545_cards_test.go — S82 OLD batch 545 (#5593, Lorwyn: pump and
// combat tricks): Inner-Flame Igniter, Cenn's Heir, Facevaulter, Kithkin
// Greatheart, Dreamspoiler Witches, Thorntooth Witch, Final Revels, Flamekin
// Brawler, Kithkin Daggerdare, Ceaseless Searblades, Lowland Oaf and Boggart
// Sprite-Chaser. Each test plays the real catalog entry through the engine.

const (
	oldB545Igniter    = "059788eb-2c55-4da2-bf54-aa2fa7954169"
	oldB545Heir       = "42adcc63-cfb0-4b62-b8e1-dda01189652a"
	oldB545Facevaulte = "602b30a0-ecf3-4639-a512-ee020675e432"
	oldB545Greatheart = "6034e7c1-b3ec-4927-b517-f8e3ca87252e"
	oldB545Witches    = "822eb600-cfaf-4ea1-a016-0c1f0cfa6436"
	oldB545Thorntooth = "942eefa3-6775-428b-84aa-48c3bd78d3b8"
	oldB545Revels     = "978aef8e-6af8-4838-97ce-afb84baa9edc"
	oldB545Brawler    = "9e3dd85c-ba19-4d22-9f90-252676f3cffa"
	oldB545Daggerdare = "b4fd5a3f-3aec-4336-9dcd-6971c117515a"
	oldB545Searblades = "bbed284c-987a-4765-b3a8-4f18a228ad0e"
	oldB545Oaf        = "c14f973e-b2b0-4d22-87f6-c37f7eb92864"
	oldB545Chaser     = "e52d8e5c-59cf-48bb-80f7-f7996c8a81bc"
)

func TestOldB545EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB545Igniter: "Inner-Flame Igniter", oldB545Heir: "Cenn's Heir", oldB545Facevaulte: "Facevaulter",
		oldB545Greatheart: "Kithkin Greatheart", oldB545Witches: "Dreamspoiler Witches",
		oldB545Thorntooth: "Thorntooth Witch", oldB545Revels: "Final Revels", oldB545Brawler: "Flamekin Brawler",
		oldB545Daggerdare: "Kithkin Daggerdare", oldB545Searblades: "Ceaseless Searblades",
		oldB545Oaf: "Lowland Oaf", oldB545Chaser: "Boggart Sprite-Chaser",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB545Activate(t *testing.T, g *game.Game, p *game.Player, src uuid.UUID, params game.ActivateAbilityParams) {
	t.Helper()
	params.Strict = true
	if err := g.ActivateCatalogAbility(p.ID, src, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func oldB545Stats(t *testing.T, g *game.Game, id uuid.UUID, wantP, wantT int) {
	t.Helper()
	if p, tough := effectstest.PtOf(t, g, id); p != wantP || tough != wantT {
		t.Errorf("is %d/%d, want %d/%d", p, tough, wantP, wantT)
	}
}

// --- Flamekin Brawler ------------------------------------------------------------

func TestOldB545BrawlerFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	b := effectstest.B12Push(g, me.ID, "Flamekin Brawler", "Creature — Elemental Warrior", oldB545Brawler, 0, 2)
	if err := g.ActivateCatalogAbility(me.ID, b, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	effectstest.B06AddMana(me, "R", "R")
	oldB545Activate(t, g, me, b, game.ActivateAbilityParams{})
	oldB545Activate(t, g, me, b, game.ActivateAbilityParams{})
	oldB545Stats(t, g, b, 2, 2)
}

// --- Inner-Flame Igniter ---------------------------------------------------------

func TestOldB545IgniterGivesFirstStrikeOnlyOnTheThirdResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ig := effectstest.B12Push(g, me.ID, "Inner-Flame Igniter", "Creature — Elemental Warrior", oldB545Igniter, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	for i := 1; i <= 3; i++ {
		effectstest.B06AddMana(me, "R", "R", "R")
		oldB545Activate(t, g, me, ig, game.ActivateAbilityParams{})
		wantFS := i == 3
		if got := effectstest.HasKeywordOf(g, other, "first strike"); got != wantFS {
			t.Errorf("after resolution %d: first strike = %v, want %v", i, got, wantFS)
		}
		oldB545Stats(t, g, other, 2+i, 2)
	}
	oldB545Stats(t, g, theirs, 2, 2)
	if effectstest.HasKeywordOf(g, theirs, "first strike") {
		t.Error("an opponent's creature got first strike")
	}
	effectstest.B06AddMana(me, "R", "R", "R")
	oldB545Activate(t, g, me, ig, game.ActivateAbilityParams{})
	oldB545Stats(t, g, other, 6, 2)
}

// --- Cenn's Heir -----------------------------------------------------------------

func TestOldB545HeirGrowsForEachOtherAttackingKithkin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heir := effectstest.B12Push(g, me.ID, "Cenn's Heir", "Creature — Kithkin Soldier", oldB545Heir, 1, 1)
	k1 := effectstest.B12Creature(g, me.ID, "Kithkin A", "Creature — Kithkin Soldier", 1, 1)
	k2 := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Changeling", TypeLine: "Creature — Shapeshifter", Power: 1, Toughness: 1,
		Keywords: []string{game.KeywordChangeling}, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{heir, k1, k2, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oldB545Stats(t, g, heir, 3, 3)
}

func TestOldB545HeirAloneStaysOneOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heir := effectstest.B12Push(g, me.ID, "Cenn's Heir", "Creature — Kithkin Soldier", oldB545Heir, 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(heir, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oldB545Stats(t, g, heir, 1, 1)
}

// --- Facevaulter -----------------------------------------------------------------

func TestOldB545FacevaulterSacrificesAGoblinToGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	f := effectstest.B12Push(g, me.ID, "Facevaulter", "Creature — Goblin Warrior", oldB545Facevaulte, 1, 1)
	gob := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, f, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{gob}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oldB545Stats(t, g, f, 3, 3)
	if g.Battlefield.Contains(gob) {
		t.Error("the Goblin was not sacrificed")
	}
}

func TestOldB545FacevaulterCannotSacrificeANonGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	f := effectstest.B12Push(g, me.ID, "Facevaulter", "Creature — Goblin Warrior", oldB545Facevaulte, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, f, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("sacrificed a Bear for a Goblin cost")
	}
}

// --- Kithkin Greatheart / Boggart Sprite-Chaser ------------------------------------

func TestOldB545GreatheartNeedsAGiant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	k := effectstest.B12Push(g, me.ID, "Kithkin Greatheart", "Creature — Kithkin Soldier", oldB545Greatheart, 2, 1)
	g.RunStateChecksForTest()
	oldB545Stats(t, g, k, 2, 1)
	if effectstest.HasKeywordOf(g, k, "first strike") {
		t.Error("first strike without a Giant")
	}
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Giant", "Creature — Giant", 3, 3)
	g.RunStateChecksForTest()
	oldB545Stats(t, g, k, 2, 1)
	giant := effectstest.B12Creature(g, me.ID, "Giant", "Creature — Giant", 3, 3)
	g.RunStateChecksForTest()
	oldB545Stats(t, g, k, 3, 2)
	if !effectstest.HasKeywordOf(g, k, "first strike") {
		t.Error("no first strike with a Giant")
	}
	_ = g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, giant)
	g.RunStateChecksForTest()
	oldB545Stats(t, g, k, 2, 1)
}

func TestOldB545ChaserNeedsAFaerie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	k := effectstest.B12Push(g, me.ID, "Boggart Sprite-Chaser", "Creature — Goblin Warrior", oldB545Chaser, 1, 2)
	g.RunStateChecksForTest()
	oldB545Stats(t, g, k, 1, 2)
	if effectstest.HasKeywordOf(g, k, "flying") {
		t.Error("flying without a Faerie")
	}
	effectstest.B12Creature(g, me.ID, "Faerie", "Creature — Faerie Rogue", 1, 1)
	g.RunStateChecksForTest()
	oldB545Stats(t, g, k, 2, 3)
	if !effectstest.HasKeywordOf(g, k, "flying") {
		t.Error("no flying with a Faerie")
	}
}

// --- Dreamspoiler Witches / Thorntooth Witch ---------------------------------------

func TestOldB545WitchesShrinkOnlyOnAnOpponentsTurn(t *testing.T) {
	for _, onOpponentsTurn := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Dreamspoiler Witches", "Creature — Faerie Wizard", oldB545Witches, 2, 2)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
		if onOpponentsTurn {
			g.Turn.ActiveSeat = 1
		}
		effectstest.AvB1Cast(t, g, me, game.Card{Name: "Trick", TypeLine: "Instant"}, game.CastSpellParams{})
		if onOpponentsTurn {
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickCard(t, g, me.ID, bear)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := 3
		if onOpponentsTurn {
			want = 2
		}
		oldB545Stats(t, g, bear, want, want)
	}
}

func TestOldB545WitchesMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Dreamspoiler Witches", "Creature — Faerie Wizard", oldB545Witches, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	g.Turn.ActiveSeat = 1
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Trick", TypeLine: "Instant"}, game.CastSpellParams{})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	oldB545Stats(t, g, bear, 3, 3)
}

func TestOldB545ThorntoothTriggersOnATreefolkSpellOnly(t *testing.T) {
	for _, treefolk := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Thorntooth Witch", "Creature — Treefolk Shaman", oldB545Thorntooth, 3, 4)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 4, 4)
		line := "Creature — Elf"
		if treefolk {
			line = "Creature — Treefolk"
		}
		effectstest.AvB1Cast(t, g, me, game.Card{Name: "Spell", TypeLine: line, Power: 1, Toughness: 1}, game.CastSpellParams{})
		if treefolk {
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickCard(t, g, me.ID, bear)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if treefolk {
			oldB545Stats(t, g, bear, 7, 1)
		} else {
			oldB545Stats(t, g, bear, 4, 4)
		}
	}
}

// --- Final Revels ----------------------------------------------------------------

func TestOldB545FinalRevelsBothModes(t *testing.T) {
	for mode, want := range [][2]int{{4, 2}, {2, 0}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 3)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 3)
		effectstest.CastCatalogSpellWithModes(t, g, "Final Revels", "Sorcery", oldB545Revels, []int{mode})
		effectstest.PassPriorityAroundTable(t, g)
		for _, id := range []uuid.UUID{mine, theirs} {
			if p, tough := effectstest.PtOf(t, g, id); p != want[0] || tough != want[1]+1 {
				t.Errorf("mode %d: %d/%d, want %d/%d", mode, p, tough, want[0], want[1]+1)
			}
		}
	}
}

// --- Kithkin Daggerdare ----------------------------------------------------------

func TestOldB545DaggerdarePumpsAnAttackerAndRefusesANonAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d := effectstest.B12Push(g, me.ID, "Kithkin Daggerdare", "Creature — Kithkin Soldier", oldB545Daggerdare, 1, 1)
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, d, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}}}); err == nil {
		t.Fatal("targeted a creature that is not attacking")
	}
	effectstest.ApaActivate(t, g, me, d, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: atk}}})
	oldB545Stats(t, g, atk, 4, 4)
	if !effectstest.HmlTapped(t, g, d) {
		t.Error("the Daggerdare did not tap")
	}
}

// --- Ceaseless Searblades --------------------------------------------------------

func TestOldB545SearbladesTriggersOnAnElementalsAbilityOnly(t *testing.T) {
	for _, elemental := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		s := effectstest.B12Push(g, me.ID, "Ceaseless Searblades", "Creature — Elemental Warrior", oldB545Searblades, 2, 4)
		line := "Creature — Bear"
		if elemental {
			line = "Creature — Elemental Warrior"
		}
		brawler := effectstest.B12Push(g, me.ID, "Flamekin Brawler", line, oldB545Brawler, 0, 2)
		effectstest.B06AddMana(me, "R")
		oldB545Activate(t, g, me, brawler, game.ActivateAbilityParams{})
		want := 2
		if elemental {
			want = 3
		}
		if p, _ := effectstest.PtOf(t, g, s); p != want {
			t.Errorf("elemental=%v: Searblades has power %d, want %d", elemental, p, want)
		}
	}
}

// --- Lowland Oaf -----------------------------------------------------------------

func TestOldB545OafGivesAGoblinFlyingThenSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oaf := effectstest.B12Push(g, me.ID, "Lowland Oaf", "Creature — Giant Warrior", oldB545Oaf, 3, 3)
	gob := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	giant := effectstest.B12Creature(g, me.ID, "Giant", "Creature — Giant", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, oaf, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: giant}}}); err == nil {
		t.Fatal("targeted a non-Goblin")
	}
	effectstest.ApaActivate(t, g, me, oaf, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: gob}}})
	oldB545Stats(t, g, gob, 2, 1)
	if !effectstest.HasKeywordOf(g, gob, "flying") {
		t.Error("the Goblin did not gain flying")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(gob) {
		t.Error("the Goblin was not sacrificed at the beginning of the end step")
	}
	if !g.Battlefield.Contains(oaf) {
		t.Error("the Oaf left the battlefield")
	}
}
