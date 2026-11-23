package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b218_test.go — Commander Legends: Battle for Baldur's Gate legends and
// red creatures, GAP batch 218 (#5948). Each test plays the real catalog entry.

const (
	gapB218FaldornOracle       = "d8e6294d-51cc-44ff-8bc6-b0a168e4f423"
	gapB218AmethystOracle      = "0961b68d-bb0c-4bfb-ad78-01f90b3b97c8"
	gapB218GorionOracle        = "195bd8fe-581c-44ed-b7f8-e800df3502ff"
	gapB218KorlessaOracle      = "2ceac1be-32ab-4a04-ade3-852926a5ca45"
	gapB218BhaalsInvokerOracle = "92512acd-bf1e-4133-93cd-6c6af5affb95"
	gapB218AuroraOracle        = "a98cc4d0-a322-4142-aee0-83f39b137616"
	gapB218AmberOracle         = "d1435f7b-a694-4c7f-a0f2-34c4c646e04b"
	gapB218MazzyOracle         = "9fa75b3c-a512-4318-9c03-e2d958acb6ef"
	gapB218RaggadraggaOracle   = "a7641ee0-9c0d-4201-9827-a775ed03797c"
	gapB218BalorOracle         = "6f931c1e-f1d2-4dc7-a653-c29adfc8fb2c"
	gapB218AlloyMyrOracle      = "efb0394c-2a45-4dd8-bca3-08704056fa31"
)

func TestGapB218EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB218FaldornOracle:         "Faldorn, Dread Wolf Herald",
		gapB218AmethystOracle:        "Amethyst Dragon",
		gapB218AmethystOracle + "#1": "Explosive Crystal",
		gapB218GorionOracle:          "Gorion, Wise Mentor",
		gapB218KorlessaOracle:        "Korlessa, Scale Singer",
		gapB218BhaalsInvokerOracle:   "Bhaal's Invoker",
		gapB218AuroraOracle:          "Aurora Phoenix",
		gapB218AmberOracle:           "Amber Gristle O'Maul",
		gapB218MazzyOracle:           "Mazzy, Truesword Paladin",
		gapB218RaggadraggaOracle:     "Raggadragga, Goreguts Boss",
		gapB218BalorOracle:           "Balor",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB218Wolves(g *game.Game, owner uuid.UUID) int {
	return effectstest.TmtB1Count(g, owner, "Wolf")
}

func gapB218Read(g *game.Game, id uuid.UUID) (c game.Card) {
	g.ReadSnapshot(func() { c, _ = g.LookupCardForEffect(id) })
	return c
}

// --- Faldorn ---

func TestGapB218FaldornMakesAWolfWhenYouCastASpellFromExileOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, first := effectstest.ImpulseExileCost(t, g, "Exiled Spell", "{R}")
	effectstest.AdvanceToMainForTest(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Faldorn, Dread Wolf Herald", "Legendary Creature — Human Druid", gapB218FaldornOracle, false)
	effectstest.B06AddMana(me, "R", "R")
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB218Wolves(g, me.ID); got != 1 {
		t.Fatalf("%d Wolf tokens after casting from exile, want 1", got)
	}
	hand := effectstest.HandSorcery(me, "Hand Spell", "{R}")
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB218Wolves(g, me.ID); got != 1 {
		t.Errorf("%d Wolf tokens after a cast from hand, want still 1", got)
	}
}

func TestGapB218FaldornMakesAWolfWhenALandYouControlEntersFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Faldorn, Dread Wolf Herald", "Legendary Creature — Human Druid", gapB218FaldornOracle, false)
	move := func(owner *game.Player, typeLine string) {
		id := uuid.New()
		g.WithWriteLock(func() {
			g.Exile.PushTop(game.Card{InstanceID: id, Name: "Exiled", TypeLine: typeLine, Owner: owner.ID, Controller: owner.ID})
		})
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
			t.Fatalf("move: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	move(opp, "Basic Land — Forest")
	move(me, "Artifact")
	if got := gapB218Wolves(g, me.ID); got != 0 {
		t.Fatalf("%d Wolf tokens for an opponent's land and a nonland from exile, want 0", got)
	}
	move(me, "Basic Land — Forest")
	if got := gapB218Wolves(g, me.ID); got != 1 {
		t.Errorf("%d Wolf tokens after my land entered from exile, want 1", got)
	}
}

func TestGapB218FaldornDiscardsToExileTheTopCardAndMayPlayIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	faldorn := effectstest.PushCatalogPermanent(g, me.ID, "Faldorn, Dread Wolf Herald", "Legendary Creature — Human Druid", gapB218FaldornOracle, false)
	fodder := effectstest.HandSorcery(me, "Fodder", "{R}")
	top := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Card", TypeLine: "Sorcery", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	})
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, faldorn, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(fodder) || !g.Exile.Contains(top) {
		t.Fatal("the card was not discarded, or the top card was not exiled")
	}
	if effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneExile) == nil {
		t.Error("no permission to play the exiled card")
	}
}

// --- Amethyst Dragon // Explosive Crystal ---

func TestGapB218ExplosiveCrystalDealsFourDamageDivided(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R", "R", "R", "R")
	effectstest.CastFromHandForTest(t, g, me, "Explosive Crystal", "Sorcery — Adventure", "{4}{R}", gapB218AmethystOracle+"#1", game.CastSpellParams{
		Strict:       true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetPlayer, ID: opp.ID}},
		Distribution: map[uuid.UUID]int{bear: 2, opp.ID: 2},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature took 2 damage and survived")
	}
	if opp.Life != life-2 {
		t.Errorf("the player's life is %d, want %d", opp.Life, life-2)
	}
	if spec, _ := effects.Lookup(gapB218AmethystOracle); len(spec.PrintedKeywords) != 2 {
		t.Errorf("Amethyst Dragon keywords %v, want flying and haste", spec.PrintedKeywords)
	}
}

// --- Gorion ---

func TestGapB218GorionCopiesAnAdventureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Gorion, Wise Mentor", "Legendary Creature — Human Wizard", gapB218GorionOracle, false)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	adventure := game.Card{InstanceID: uuid.New(), Name: "Explosive Crystal", TypeLine: "Sorcery — Adventure", Layout: game.LayoutAdventure,
		OracleID: gapB218AmethystOracle + "#1", ManaCost: "{4}{R}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(adventure)
	effectstest.B06AddMana(me, "R", "R", "R", "R", "R")
	if err := g.CastSpell(me.ID, adventure.InstanceID, game.CastSpellParams{Strict: true,
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Distribution: map[uuid.UUID]int{opp.ID: 4}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	// "You may choose new targets for the copy": keep the same player.
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 8 {
		t.Errorf("the opponent lost %d life, want 8 (the spell and its copy)", got)
	}
}

func TestGapB218GorionDoesNotCopyACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Gorion, Wise Mentor", "Legendary Creature — Human Wizard", gapB218GorionOracle, false)
	effectstest.AdvanceToMain(t, g)
	bear := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(bear)
	effectstest.B06AddMana(me, "G")
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectstest.OwedMays(g, me.ID) != 0 {
		t.Error("Gorion offered to copy a creature spell")
	}
}

// --- Korlessa ---

func TestGapB218KorlessaLooksAtTheTopAndCastsDragonsFromIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Korlessa, Scale Singer", "Legendary Creature — Dragon Bard", gapB218KorlessaOracle, false)
	effectstest.AdvanceToMainForTest(t, g)
	spec, _ := effects.Lookup(gapB218KorlessaOracle)
	if spec.LibraryTopVisible != game.LibraryTopOwner {
		t.Errorf("top visibility %v, want only the owner", spec.LibraryTopVisible)
	}
	dragon := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Dragon", TypeLine: "Creature — Dragon", ManaCost: "{R}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "R")
	if err := g.CastSpell(me.ID, dragon, game.CastSpellParams{Strict: true, FromZone: "library"}); err != nil {
		t.Fatalf("cast a Dragon from the top: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) {
		t.Error("the Dragon did not resolve onto the battlefield")
	}
	bear := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{R}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "R")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Error("a non-Dragon spell was cast from the top of the library")
	}
}

// --- Bhaal's Invoker ---

func TestGapB218BhaalsInvokerDealsFourToEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	invoker := effectstest.PushCatalogPermanent(g, me.ID, "Bhaal's Invoker", "Creature — Dragon Shaman", gapB218BhaalsInvokerOracle, false)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, invoker, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 4
		if s.ID == me.ID {
			want = lives[s.ID]
		}
		if s.Life != want {
			t.Errorf("%s life %d, want %d", s.Name, s.Life, want)
		}
	}
}

// --- Aurora Phoenix ---

func TestGapB218AuroraPhoenixReturnsFromTheGraveyardWhenYouCastACascadeSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	phoenix := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: phoenix, Name: "Aurora Phoenix", TypeLine: "Creature — Phoenix", OracleID: gapB218AuroraOracle,
			ManaCost: "{4}{R}{R}", Power: 5, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	plain := effectstest.HandSorcery(me, "Plain Spell", "{R}")
	effectstest.B06AddMana(me, "R", "R")
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(phoenix) {
		t.Fatal("the Phoenix returned for a spell without cascade")
	}
	// A cascade spell: Annoyed Altisaur's cascade trigger is the keyword's.
	altisaur := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: altisaur, Name: "Annoyed Altisaur", TypeLine: "Creature — Dinosaur", OracleID: "a8135179-51ef-454e-98ff-69137440339f",
		ManaCost: "{4}{G}{G}{G}", Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "G", "G", "G", "G", "G", "G", "G")
	if err := g.CastSpell(me.ID, altisaur, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the cascade spell: %v", err)
	}
	for i := 0; i < 4 && !me.Hand.Contains(phoenix); i++ {
		effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerAllMayCast(t, g, me.ID, false)
	}
	if !me.Hand.Contains(phoenix) {
		t.Error("the Phoenix did not return to its owner's hand when a cascade spell was cast")
	}
}

// --- Amber Gristle O'Maul ---

func TestGapB218AmberGristleDiscardsHandAndDrawsPerPlayerBeingAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	amber := effectstest.B12Push(g, me.ID, "Amber Gristle O'Maul", "Legendary Creature — Dwarf Cleric", gapB218AmberOracle, 3, 3)
	effectstest.HandSorcery(me, "Card A", "{R}")
	effectstest.HandSorcery(me, "Card B", "{R}")
	effectstest.DeclareAttack(t, g, opp.ID, amber)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 1 {
		t.Errorf("hand has %d cards after discarding it all and attacking one player, want 1", got)
	}
}

func TestGapB218AmberGristleDecliningKeepsTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	amber := effectstest.B12Push(g, me.ID, "Amber Gristle O'Maul", "Legendary Creature — Dwarf Cleric", gapB218AmberOracle, 3, 3)
	effectstest.HandSorcery(me, "Card A", "{R}")
	before := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, amber)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand has %d cards after declining, want %d", got, before)
	}
}

// --- Mazzy ---

func TestGapB218MazzyPumpsAnEnchantedAttackerAndRecastsADyingAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mazzy, Truesword Paladin", "Legendary Creature — Halfling Knight", gapB218MazzyOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Enchanted Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Test Aura", "Enchantment — Aura", "", false)
	effectstest.BrAttach(t, g, aura, bear)
	effectstest.DeclareAttack(t, g, opp.ID, bear, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 4 {
		t.Errorf("the enchanted attacker has power %d, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the enchanted attacker has no trample")
	}
	if p, _ := effectstest.PtOf(t, g, plain); p != 2 || effectstest.HasEffectiveKeyword(t, g, plain, "trample") {
		t.Error("the unenchanted attacker was pumped")
	}
	// The Aura dies: it is exiled and may be cast until the end of my next turn.
	if err := g.DestroyPermanentForEffect(aura); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(aura) || me.Graveyard.Contains(aura) {
		t.Fatal("the Aura was not exiled from the graveyard")
	}
	if effectstest.GrantedPermissionOn(g, me.ID, aura, game.ZoneExile) == nil {
		t.Error("no permission to cast the exiled Aura")
	}
}

func TestGapB218MazzyIgnoresAnAuraAnOpponentOwns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Mazzy, Truesword Paladin", "Legendary Creature — Halfling Knight", gapB218MazzyOracle, false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Aura", "Enchantment — Aura", "", false)
	if err := g.DestroyPermanentForEffect(theirs); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(theirs) {
		t.Error("an opponent's Aura was exiled by Mazzy")
	}
}

// --- Raggadragga ---

func TestGapB218RaggadraggaGivesManaCreaturesPlusTwoAndUntapsThemWhenTheyAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Raggadragga, Goreguts Boss", "Legendary Creature — Human Boar", gapB218RaggadraggaOracle, false)
	myr := effectstest.B12Push(g, me.ID, "Alloy Myr", "Artifact Creature — Myr", gapB218AlloyMyrOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Alloy Myr", "Artifact Creature — Myr", gapB218AlloyMyrOracle, 2, 2)
	if p, tough := effectstest.PtOf(t, g, myr); p != 4 || tough != 4 {
		t.Errorf("my mana creature is %d/%d, want 4/4", p, tough)
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("a creature without a mana ability has power %d, want 2", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's mana creature has power %d, want 2", p)
	}
	effectstest.DeclareAttack(t, g, opp.ID, myr)
	effectstest.PassPriorityAroundTable(t, g)
	if c := gapB218Read(g, myr); c.Tapped {
		t.Error("the attacking mana creature was not untapped")
	}
}

func TestGapB218RaggadraggaPumpsOnASevenManaSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Raggadragga, Goreguts Boss", "Legendary Creature — Human Boar", gapB218RaggadraggaOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Big Sorcery", "Sorcery", "{7}", "", game.CastSpellParams{Strict: true})
	gapB218PickTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 9 || tough != 9 {
		t.Errorf("the target is %d/%d, want 9/9", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the target has no trample")
	}
}

func TestGapB218RaggadraggaIgnoresASixManaSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Raggadragga, Goreguts Boss", "Legendary Creature — Human Boar", gapB218RaggadraggaOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Mid Sorcery", "Sorcery", "{6}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("a six-mana spell pumped the creature to %d", p)
	}
}

func gapB218PickTarget(t *testing.T, g *game.Game, chooser uuid.UUID, id uuid.UUID) {
	t.Helper()
	if pick := effectstest.LatestPickTarget(g, chooser); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
}

// --- Balor ---

func TestGapB218BalorDiesAndDifferentOpponentsAreHitByDifferentModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := len(g.Seats)
	opp, other := g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n]
	effectstest.AdvanceToMain(t, g)
	balor := effectstest.B12Push(g, me.ID, "Balor", "Creature — Demon", gapB218BalorOracle, 5, 5)
	for i := 0; i < 4; i++ {
		effectstest.HandSorcery(opp, "Opp Card", "{R}")
	}
	handBefore := opp.Hand.Size()
	otherLife := other.Life
	effectstest.B18Kill(t, g, balor)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the trigger did not ask for modes")
	}
	// First (draw three, discard three at random) and third (damage equal to hand size).
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if err := effectstest.TmcB7TryPickPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("first bullet: %v", err)
	}
	if err := effectstest.TmcB7TryPickPlayer(t, g, me.ID, opp.ID); err == nil {
		t.Fatal("the damage bullet accepted the opponent the first bullet targets")
	}
	if err := effectstest.TmcB7TryPickPlayer(t, g, me.ID, other.ID); err != nil {
		t.Fatalf("third bullet: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != handBefore {
		t.Errorf("the first opponent's hand is %d, want %d (drew three, discarded three)", got, handBefore)
	}
	if want := otherLife - other.Hand.Size(); other.Life != want {
		t.Errorf("the second opponent's life is %d, want %d (damage equal to their hand)", other.Life, want)
	}
	if other.Life == otherLife && other.Hand.Size() > 0 {
		t.Error("no damage was dealt by the dead Balor")
	}
}

func TestGapB218BalorEdictTakesANontokenArtifactOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	balor := effectstest.B12Push(g, me.ID, "Balor", "Creature — Demon", gapB218BalorOracle, 5, 5)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	token := effectstest.PushCatalogPermanent(g, opp.ID, "Their Treasure", "Token Artifact — Treasure", "", false)
	effectstest.DeclareAttack(t, g, opp.ID, balor)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the attack trigger did not ask for modes")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if err := effectstest.TmcB7TryPickPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the nontoken artifact was not sacrificed")
	}
	if !g.Battlefield.Contains(token) {
		t.Error("the token artifact was sacrificed")
	}
}
