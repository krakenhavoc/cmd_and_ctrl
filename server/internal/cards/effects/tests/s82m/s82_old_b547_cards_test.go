package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b547_cards_test.go — S82 OLD batch 547 (#5593, Lorwyn: activated
// abilities): Black Poplar Shaman, Elvish Branchbender, Sygg, River Guide,
// Twinning Glass, Drowner of Secrets, Silvergill Douser, Deeptread Merrow,
// Boggart Forager, Kithkin Healer and Gilt-Leaf Seer. Each test plays the real
// catalog entry through the engine.

const (
	oldB547Poplar       = "39d60cd8-bebc-407d-b361-8e49e70e339a"
	oldB547Branchbender = "6e9fe674-44c8-4ab4-a0ef-ae71d65bbcc0"
	oldB547Sygg         = "73f6da1e-b8ce-4b5a-9da0-84ca67adaf1c"
	oldB547Glass        = "907baed5-499b-4f20-bbb9-8f5bead37c30"
	oldB547Drowner      = "a5ce3fb6-9c70-4433-a079-467c911bfd4d"
	oldB547Douser       = "b24f2917-b96c-4348-b033-4926df3bd4b0"
	oldB547Merrow       = "b301b371-e99e-43cc-af7f-dbd7b2abe5f8"
	oldB547Forager      = "cef9e0d3-5e28-4c5f-88fc-0d78af5a4db1"
	oldB547Healer       = "e3bd74f3-1f38-4162-a0be-c76ad63e2d9b"
	oldB547Seer         = "eab27405-f590-40b7-ad7c-0e6c52410c0a"
)

func TestOldB547EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB547Poplar: "Black Poplar Shaman", oldB547Branchbender: "Elvish Branchbender", oldB547Sygg: "Sygg, River Guide", oldB547Glass: "Twinning Glass", oldB547Drowner: "Drowner of Secrets", oldB547Douser: "Silvergill Douser", oldB547Merrow: "Deeptread Merrow", oldB547Forager: "Boggart Forager", oldB547Healer: "Kithkin Healer", oldB547Seer: "Gilt-Leaf Seer",
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

func oldB547Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB547Player(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func oldB547Changeling(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Changeling", TypeLine: "Creature — Shapeshifter", Power: 1, Toughness: 1,
		Keywords: []string{game.KeywordChangeling}, Owner: owner, Controller: owner,
	})
}

// --- Black Poplar Shaman -----------------------------------------------------

func TestOldB547PoplarRegeneratesOnlyATreefolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shaman := effectstest.B12Push(g, me.ID, "Black Poplar Shaman", "Creature — Treefolk Shaman", oldB547Poplar, 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	changeling := oldB547Changeling(g, me.ID)
	effectstest.B06AddMana(me, "B", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB547Card(bear)}); err == nil {
		t.Fatal("regenerated a Bear")
	}
	effectstest.ApaActivate(t, g, me, shaman, 0, game.ActivateAbilityParams{Targets: oldB547Card(changeling)})
	if n := effectstest.RegenShieldsOn(t, g, changeling); n != 1 {
		t.Fatalf("the changeling has %d regeneration shields, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(changeling) })
	if !g.Battlefield.Contains(changeling) {
		t.Error("the regenerated changeling was destroyed")
	}
}

// --- Elvish Branchbender -----------------------------------------------------

func TestOldB547BranchbenderAnimatesAForestAsAnXByXTreefolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bender := effectstest.B12Push(g, me.ID, "Elvish Branchbender", "Creature — Elf Druid", oldB547Branchbender, 2, 2)
	effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 1, 1)
	oldB547Changeling(g, me.ID)
	effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf Warrior", 1, 1)
	forest := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	island := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	if err := g.ActivateCatalogAbility(me.ID, bender, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB547Card(island)}); err == nil {
		t.Fatal("animated an Island")
	}
	effectstest.ApaActivate(t, g, me, bender, 0, game.ActivateAbilityParams{Targets: oldB547Card(forest)})
	c, ok := effectstest.BattlefieldCard(g, forest)
	if !ok || !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Treefolk") || !c.HasSubtype("Forest") {
		t.Fatalf("the Forest is %+v, want a Forest land Treefolk creature", c.TypeLine)
	}
	if p, tough := effectstest.PtOf(t, g, forest); p != 3 || tough != 3 {
		t.Errorf("the Forest is %d/%d, want 3/3: the Branchbender, an Elf and a changeling are three Elves you control", p, tough)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if c, _ := effectstest.BattlefieldCard(g, forest); c.IsCreature() {
		t.Error("the Forest is still a creature after the turn ended")
	}
}

// --- Sygg, River Guide ---------------------------------------------------------

func TestOldB547SyggGivesAMerfolkProtectionFromAChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sygg := effectstest.B12Push(g, me.ID, "Sygg, River Guide", "Legendary Creature — Merfolk Wizard", oldB547Sygg, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasKeywordOf(g, sygg, "islandwalk") {
		t.Error("Sygg lacks islandwalk")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, sygg, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB547Card(bear)}); err == nil {
		t.Fatal("targeted a Bear")
	}
	if err := g.ActivateCatalogAbility(me.ID, sygg, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB547Card(sygg)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < 8 && effectstest.PendingOfKind(g, game.PendingChoiceColor) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerColor(t, g, me.ID, "R")
	c, _ := effectstest.BattlefieldCard(g, sygg)
	if !game.HasKeyword(&c, game.ProtectionFromColor("R")) {
		t.Error("Sygg did not gain protection from red")
	}
	if game.HasKeyword(&c, game.ProtectionFromColor("U")) {
		t.Error("Sygg gained protection from a color that was not chosen")
	}
}

// --- Drowner of Secrets --------------------------------------------------------

func TestOldB547DrownerTapsAMerfolkToMillAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drowner := effectstest.B12Push(g, me.ID, "Drowner of Secrets", "Creature — Merfolk Wizard", oldB547Drowner, 1, 3)
	library, graveyard := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.ApaActivate(t, g, me, drowner, 0, game.ActivateAbilityParams{Targets: oldB547Player(opp.ID), TapIDs: []uuid.UUID{drowner}})
	if opp.Library.Size() != library-1 || opp.Graveyard.Size() != graveyard+1 {
		t.Errorf("library %d graveyard %d, want %d and %d", opp.Library.Size(), opp.Graveyard.Size(), library-1, graveyard+1)
	}
	if !effectstest.HmlTapped(t, g, drowner) {
		t.Error("the Drowner (the only Merfolk) was not tapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, drowner, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB547Player(opp.ID), TapIDs: []uuid.UUID{drowner}}); err == nil {
		t.Fatal("activated with no untapped Merfolk")
	}
}

func TestOldB547DrownerMaySpendASummoningSickChangeling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drowner := effectstest.B12Push(g, me.ID, "Drowner of Secrets", "Creature — Merfolk Wizard", oldB547Drowner, 1, 3)
	effectstest.TapCard(g, drowner)
	sick := oldB547Changeling(g, me.ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == sick {
				g.Battlefield.Cards[i].SummonedThisTurn = true
			}
		}
	})
	library := opp.Library.Size()
	effectstest.ApaActivate(t, g, me, drowner, 0, game.ActivateAbilityParams{Targets: oldB547Player(opp.ID), TapIDs: []uuid.UUID{sick}})
	if opp.Library.Size() != library-1 {
		t.Error("a summoning-sick Merfolk could not pay the cost")
	}
}

// --- Silvergill Douser ---------------------------------------------------------

func TestOldB547DouserShrinksByTheMerfolkAndFaeriesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	douser := effectstest.B12Push(g, me.ID, "Silvergill Douser", "Creature — Merfolk Wizard", oldB547Douser, 1, 1)
	effectstest.B12Creature(g, me.ID, "Faerie", "Creature — Faerie Rogue", 1, 1)
	effectstest.B12Creature(g, me.ID, "Faerie Merfolk", "Creature — Faerie Merfolk", 1, 1)
	oldB547Changeling(g, me.ID)
	effectstest.B12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 1, 1)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	effectstest.ApaActivate(t, g, me, douser, 0, game.ActivateAbilityParams{Targets: oldB547Card(ogre)})
	if p, tough := effectstest.PtOf(t, g, ogre); p != 1 || tough != 5 {
		t.Errorf("the Ogre is %d/%d, want 1/5: the Douser, a Faerie, a Faerie Merfolk and a changeling are four", p, tough)
	}
}

// --- Deeptread Merrow ----------------------------------------------------------

func TestOldB547MerrowGainsIslandwalkUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	merrow := effectstest.B12Push(g, me.ID, "Deeptread Merrow", "Creature — Merfolk Rogue", oldB547Merrow, 2, 1)
	if effectstest.HasKeywordOf(g, merrow, "islandwalk") {
		t.Fatal("the Merrow has islandwalk before the ability")
	}
	effectstest.B06AddMana(me, "U")
	effectstest.ApaActivate(t, g, me, merrow, 0, game.ActivateAbilityParams{})
	if !effectstest.HasKeywordOf(g, merrow, "islandwalk") {
		t.Error("the Merrow did not gain islandwalk")
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if effectstest.HasKeywordOf(g, merrow, "islandwalk") {
		t.Error("islandwalk outlasted the turn")
	}
}

// --- Boggart Forager -----------------------------------------------------------

func TestOldB547ForagerSacrificesToShuffleAPlayersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forager := effectstest.B12Push(g, me.ID, "Boggart Forager", "Creature — Goblin Rogue", oldB547Forager, 1, 1)
	for i := 0; i < 20; i++ {
		effectstest.TopOfLibrary(opp, "Card", "Sorcery", "{1}")
	}
	order := func() []uuid.UUID {
		var out []uuid.UUID
		for _, c := range opp.Library.Cards {
			out = append(out, c.InstanceID)
		}
		return out
	}
	before := order()
	effectstest.B06AddMana(me, "R")
	effectstest.ApaActivate(t, g, me, forager, 0, game.ActivateAbilityParams{Targets: oldB547Player(opp.ID)})
	if g.Battlefield.Contains(forager) {
		t.Error("the Forager was not sacrificed")
	}
	after := order()
	if len(after) != len(before) || slices.Equal(before, after) {
		t.Error("the target library was not shuffled")
	}
}

// --- Kithkin Healer ------------------------------------------------------------

func TestOldB547HealerPreventsTheNextDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	healer := effectstest.B12Push(g, me.ID, "Kithkin Healer", "Creature — Kithkin Cleric", oldB547Healer, 2, 2)
	effectstest.ApaActivate(t, g, me, healer, 0, game.ActivateAbilityParams{Targets: oldB547Player(opp.ID)})
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.BoltOracleCombat, oldB547Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d: one of Bolt's three damage is prevented", opp.Life, life-2)
	}
	if !effectstest.HmlTapped(t, g, healer) {
		t.Error("the Healer did not tap")
	}
}

// --- Gilt-Leaf Seer ------------------------------------------------------------

func TestOldB547SeerLooksAtTheTopTwoAndReordersThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	seer := effectstest.B12Push(g, me.ID, "Gilt-Leaf Seer", "Creature — Elf Shaman", oldB547Seer, 2, 2)
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LookAtTopChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("the look-at prompt = %+v, want two cards", c)
	}
	top := effectstest.LibraryTopNames(me, 2)
	if err := g.ResolveLookAtTop(c.ID, me.ID, []uuid.UUID{c.ScryCards[1], c.ScryCards[0]}); err != nil {
		t.Fatalf("ResolveLookAtTop: %v", err)
	}
	if got := effectstest.LibraryTopNames(me, 2); got[0] != top[1] || got[1] != top[0] {
		t.Errorf("library top is %v, want %v reversed", got, top)
	}
}

// --- Twinning Glass ------------------------------------------------------------

func TestOldB547TwinningGlassCastsAHandCardWithTheNameOfASpellCastThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glass := effectstest.PushCatalogPermanent(g, me.ID, "Twinning Glass", "Artifact", oldB547Glass, false)
	boltInHand := effectstest.AvB1Hand(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.BoltOracleCombat, ManaCost: "{R}"})
	other := effectstest.AvB1Hand(me, game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"})
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.BoltOracleCombat, oldB547Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.B06AddMana(me, "C")
	effectstest.ApaActivate(t, g, me, glass, 0, game.ActivateAbilityParams{})
	offer := effectstest.KeywordCastPromptFor(g, me.ID)
	if offer == nil {
		t.Fatal("Twinning Glass offered no free cast")
	}
	if slices.Contains(offer.DiscardOptions, other) || !slices.Contains(offer.DiscardOptions, boltInHand) {
		t.Fatalf("the offer %v must hold only the Lightning Bolt (%s), not Shock (%s)", offer.DiscardOptions, boltInHand, other)
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, boltInHand, game.CastSpellParams{FromZone: "hand", Targets: oldB547Player(opp.ID)}); err != nil {
		t.Fatalf("cast the free Bolt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d: the second Bolt was free and resolved", opp.Life, life-3)
	}
}

func TestOldB547TwinningGlassOffersNothingWithoutASameNamedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	glass := effectstest.PushCatalogPermanent(g, me.ID, "Twinning Glass", "Artifact", oldB547Glass, false)
	effectstest.AvB1Hand(me, game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"})
	effectstest.B06AddMana(me, "C")
	effectstest.ApaActivate(t, g, me, glass, 0, game.ActivateAbilityParams{})
	if effectstest.KeywordCastPromptFor(g, me.ID) != nil || effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("Twinning Glass offered a cast although no spell was cast this turn")
	}
}
