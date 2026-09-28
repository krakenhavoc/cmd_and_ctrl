package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// duration_copy_test.go — #1593's cards end to end: a permanent already
// on the battlefield becomes a copy of something until end of turn (or
// for good), takes on the copied card's catalog behaviour while it is
// the copy, and gives it back when the copy ends. The engine shape is
// pinned in server/internal/game/duration_copy_test.go.

const (
	oracleMirageMirror         = "6a84eda1-7e72-4b5a-849f-b6e177565aeb"
	oracleCytoshape            = "14221c18-7801-49c9-a2c8-53b8c5181d63"
	oracleMirrorweave          = "026b7221-0caf-4c8b-8c1b-e7de836797b7"
	oracleUnstableShapeshifter = "6a71deb3-6659-47ea-8687-136cf75fa381"
	oracleLazavDimirMastermind = "8027a610-613e-4640-9840-c8778694f312"
	oracleZulaportCutthroat    = "76b003e0-15af-4f22-bdf2-1ade5430964a"
)

// dcSeed puts a permanent on the battlefield through the zone-move
// event, so it carries an entry stamp like a real one.
func dcSeed(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracle,
		ManaCost:   "{2}",
		Power:      p,
		Toughness:  tough,
		Owner:      owner,
		Controller: owner,
	})
}

// dcCard reads a battlefield permanent after settling the layer pass.
func dcCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c, true
			}
		}
	})
	if !found {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return out
}

// dcCleanup runs the CR 514.2 sweep the cleanup step runs.
func dcCleanup(g *game.Game) {
	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
}

func dcMana(t *testing.T, g *game.Game, player uuid.UUID, mana string) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddManaForEffect(player, uuid.Nil, mana) })
	if err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
}

// dcSacrificeWithBombardment sacrifices `fodder` to a Goblin
// Bombardment-shaped ability on `source` and reports how many triggered
// abilities `watcher` put on the stack in response.
func dcSacrificeWithBombardment(t *testing.T, g *game.Game, player, source, fodder, victim, watcher uuid.UUID) int {
	t.Helper()
	if err := g.ActivateCatalogAbility(player, source, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: victim}},
	}); err != nil {
		t.Fatalf("activate the sacrifice ability: %v", err)
	}
	// The sacrifice is paid at announce, so its death triggers are
	// already queued.
	n := 0
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == watcher {
			n++
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemTriggered && it.SourceCardID == watcher {
			n++
		}
	}
	for i := 0; i < 64 && !stackFullyEmpty(g); i++ {
		if p := triggerOrderPromptFor(g, player); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, player, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return n
}

// --- Mirage Mirror --------------------------------------------------

// TestMirageMirrorTakesOnTheCopiedTriggerAndLosesItAtCleanup — the
// copy's triggered ability works while the Mirror is the copy and is
// gone once cleanup puts the Mirror back.
func TestMirageMirrorTakesOnTheCopiedTriggerAndLosesItAtCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mirror := dcSeed(g, me.ID, "Mirage Mirror", "Artifact", oracleMirageMirror, 0, 0)
	zulaport := dcSeed(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", oracleZulaportCutthroat, 1, 1)
	bomb := dcSeed(g, me.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, 0, 0)
	fodderA := dcSeed(g, me.ID, "Fodder A", "Creature — Goblin", "", 1, 1)
	fodderB := dcSeed(g, me.ID, "Fodder B", "Creature — Goblin", "", 1, 1)

	dcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, mirror, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: zulaport}},
	}); err != nil {
		t.Fatalf("activate Mirage Mirror: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := dcCard(t, g, mirror)
	if c.Name != "Zulaport Cutthroat" || !c.IsCreature() || c.Effective().Power != 1 {
		t.Fatalf("Mirage Mirror did not become the Cutthroat: %q creature %v", c.Name, c.IsCreature())
	}
	if got := dcSacrificeWithBombardment(t, g, me.ID, bomb, fodderA, opp.ID, mirror); got != 1 {
		t.Errorf("while a copy of Zulaport Cutthroat, the Mirror put %d death triggers on the stack, want 1", got)
	}

	dcCleanup(g)
	c = dcCard(t, g, mirror)
	if c.Name != "Mirage Mirror" || c.IsCreature() || c.IsCopy() {
		t.Fatalf("after cleanup the Mirror is %q (creature %v, copy %v)", c.Name, c.IsCreature(), c.IsCopy())
	}
	if got := dcSacrificeWithBombardment(t, g, me.ID, bomb, fodderB, opp.ID, mirror); got != 0 {
		t.Errorf("after cleanup the Mirror still put %d death triggers on the stack", got)
	}
}

// TestMirageMirrorTakesOnTheCopiedActivatedAbility — the copy's
// activated ability replaces the Mirror's own, and the Mirror's own
// comes back at cleanup.
func TestMirageMirrorTakesOnTheCopiedActivatedAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mirror := dcSeed(g, me.ID, "Mirage Mirror", "Artifact", oracleMirageMirror, 0, 0)
	bomb := dcSeed(g, me.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, 0, 0)
	fodder := dcSeed(g, me.ID, "Fodder", "Creature — Goblin", "", 1, 1)

	dcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, mirror, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bomb}},
	}); err != nil {
		t.Fatalf("activate Mirage Mirror: %v", err)
	}
	passPriorityAroundTable(t, g)
	before := opp.Life
	dcSacrificeWithBombardment(t, g, me.ID, mirror, fodder, opp.ID, uuid.Nil)
	if opp.Life != before-1 {
		t.Fatalf("the Mirror-as-Bombardment dealt %d, want 1", before-opp.Life)
	}

	dcCleanup(g)
	rows := game.ActivatedAbilitiesForCard(dcCard(t, g, mirror))
	if len(rows) != 1 || rows[0].Cost.Mana != "{2}" {
		t.Errorf("after cleanup the Mirror's ability is %+v, want its own {2} copy ability", rows)
	}
}

// TestMirageMirrorRefusesATargetItCannotCopy — "target artifact,
// creature, enchantment, or land": an instant card is none of those,
// and neither is a player.
func TestMirageMirrorRefusesATargetItCannotCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mirror := dcSeed(g, me.ID, "Mirage Mirror", "Artifact", oracleMirageMirror, 0, 0)
	dcMana(t, g, me.ID, "{C}{C}")
	err := g.ActivateCatalogAbility(me.ID, mirror, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if err == nil {
		t.Fatal("Mirage Mirror targeted a player")
	}
	if c := dcCard(t, g, mirror); c.Name != "Mirage Mirror" {
		t.Errorf("a refused activation changed the Mirror into %q", c.Name)
	}
}

// --- Shifting Woodland ----------------------------------------------

func dcGraveyardCard(g *game.Game, owner *game.Player, name, typeLine, oracle string, p, tough int) uuid.UUID {
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner.ID, Controller: owner.ID,
	})
	return id
}

// TestShiftingWoodlandBecomesAGraveyardCardUntilCleanup — the delirium
// copy lands, survives the copied card being exiled, and ends at
// cleanup.
func TestShiftingWoodlandBecomesAGraveyardCardUntilCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	woodland := dcSeed(g, me.ID, "Shifting Woodland", "Land", slice294cShiftingWoodlandOracle, 0, 0)
	hoof := dcGraveyardCard(g, me, "Craterhoof Behemoth", "Creature — Beast", "oracle-hoof", 5, 5)
	dcGraveyardCard(g, me, "Sol Ring", "Artifact", "oracle-sol", 0, 0)
	dcGraveyardCard(g, me, "Rhystic Study", "Enchantment", "oracle-study", 0, 0)
	dcGraveyardCard(g, me, "Brainstorm", "Instant", "oracle-brainstorm", 0, 0)

	dcMana(t, g, me.ID, "{G}{G}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, woodland, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hoof}},
	}); err != nil {
		t.Fatalf("activate Shifting Woodland: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := dcCard(t, g, woodland); c.Name != "Craterhoof Behemoth" || c.Effective().Power != 5 || c.IsLand() {
		t.Fatalf("Shifting Woodland did not become the Behemoth: %q %d land %v", c.Name, c.Effective().Power, c.IsLand())
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, hoof); err != nil {
		t.Fatalf("exile the copied card: %v", err)
	}
	if c := dcCard(t, g, woodland); c.Name != "Craterhoof Behemoth" {
		t.Fatalf("exiling the copied card ended the copy: %q", c.Name)
	}
	dcCleanup(g)
	if c := dcCard(t, g, woodland); c.Name != "Shifting Woodland" || !c.IsLand() || c.IsCreature() {
		t.Fatalf("after cleanup the land is %q (land %v creature %v)", c.Name, c.IsLand(), c.IsCreature())
	}
}

// TestShiftingWoodlandNeedsDelirium — three card types is not four.
func TestShiftingWoodlandNeedsDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	woodland := dcSeed(g, me.ID, "Shifting Woodland", "Land", slice294cShiftingWoodlandOracle, 0, 0)
	hoof := dcGraveyardCard(g, me, "Craterhoof Behemoth", "Creature — Beast", "oracle-hoof", 5, 5)
	dcGraveyardCard(g, me, "Sol Ring", "Artifact", "oracle-sol", 0, 0)
	dcGraveyardCard(g, me, "Rhystic Study", "Enchantment", "oracle-study", 0, 0)

	dcMana(t, g, me.ID, "{G}{G}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, woodland, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hoof}},
	}); err == nil {
		t.Fatal("Shifting Woodland activated with only three card types in the graveyard")
	}
}

// --- Cursed Mirror --------------------------------------------------

// TestCursedMirrorIsAHastyCopyUntilCleanup — the copy lands as the
// artifact enters, has haste, and ends at cleanup, leaving an artifact
// that taps for {R}.
func TestCursedMirrorIsAHastyCopyUntilCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	mirror := castCatalogSpell(t, g, "Cursed Mirror", "Artifact", cursedMirrorOracle, nil)
	resolveWithCopyChoice(t, g, bears)

	c := dcCard(t, g, mirror)
	if c.Name != "Grizzly Bears" || !c.IsCreature() || !game.HasKeyword(&c, "haste") {
		t.Fatalf("Cursed Mirror entered as %q (creature %v, haste %v)", c.Name, c.IsCreature(), game.HasKeyword(&c, "haste"))
	}
	if game.HasSummoningSickness(&c) {
		t.Error("the hasty copy is summoning-sick")
	}

	dcCleanup(g)
	c = dcCard(t, g, mirror)
	if c.Name != "Cursed Mirror" || c.IsCreature() || c.IsCopy() {
		t.Fatalf("after cleanup Cursed Mirror is %q (creature %v, copy %v)", c.Name, c.IsCreature(), c.IsCopy())
	}
	if err := g.ActivateManaAbility(me.ID, mirror, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("after cleanup Cursed Mirror does not tap for mana: %v", err)
	}
}

// TestCursedMirrorDeclinedEntersAsItself — "you may": declining is an
// artifact with no copy and nothing to end.
func TestCursedMirrorDeclinedEntersAsItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	mirror := castCatalogSpell(t, g, "Cursed Mirror", "Artifact", cursedMirrorOracle, nil)
	resolveWithCopyChoice(t, g, uuid.Nil)
	if c := dcCard(t, g, mirror); c.Name != "Cursed Mirror" || c.IsCopy() || c.IsCreature() {
		t.Fatalf("a declined Cursed Mirror is %q (copy %v)", c.Name, c.IsCopy())
	}
	if len(g.ScopedEffects) != 0 {
		t.Errorf("a declined copy left %d records", len(g.ScopedEffects))
	}
}

// --- Cytoshape -------------------------------------------------------

// dcAnswerChooseCards answers the open ChooseCards prompt with `pick`.
func dcAnswerChooseCards(t *testing.T, g *game.Game, pick uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 16; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseCards {
				if err := g.ResolveChooseCards(c.ID, c.Chooser, []uuid.UUID{pick}); err != nil {
					t.Fatalf("ResolveChooseCards: %v", err)
				}
				return c
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("no ChooseCards prompt was queued")
	return nil
}

// TestCytoshapeOnACloneRevertsToTheClonesCopy — issue #1593's named
// case, through the real cards: a Clone of Grizzly Bears, Cytoshaped
// into a Hill Giant, is Grizzly Bears again at cleanup.
func TestCytoshapeOnACloneRevertsToTheClonesCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	giant := seedCopyableCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	clone := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, bears)
	if c := dcCard(t, g, clone); c.Name != "Grizzly Bears" {
		t.Fatalf("setup: the Clone entered as %q", c.Name)
	}

	castCatalogSpell(t, g, "Cytoshape", "Instant", oracleCytoshape,
		[]game.TargetRef{{Kind: game.TargetCard, ID: clone}})
	prompt := dcAnswerChooseCards(t, g, giant)
	passPriorityAroundTable(t, g)
	for _, id := range prompt.ChooseCards {
		if id == clone || id == giant || id == bears {
			continue
		}
		t.Errorf("Cytoshape offered %s, which is not a nonlegendary creature on the battlefield", id)
	}
	if c := dcCard(t, g, clone); c.Name != "Hill Giant" || c.Effective().Power != 3 {
		t.Fatalf("Cytoshaped Clone is %q %d/%d", c.Name, c.Effective().Power, c.Effective().Toughness)
	}
	dcCleanup(g)
	c := dcCard(t, g, clone)
	if c.Name != "Grizzly Bears" || c.Effective().Power != 2 {
		t.Fatalf("after cleanup the Clone is %q, want its entry copy, Grizzly Bears", c.Name)
	}
	if c.PrintedSelf == nil || c.PrintedSelf.Name != "Clone" {
		t.Errorf("the Clone lost its own printed values: %+v", c.PrintedSelf)
	}
}

// TestCytoshapeDoesNotOfferALegendaryCreature — "choose a NONLEGENDARY
// creature": a legendary one is not on the list.
func TestCytoshapeDoesNotOfferALegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	legend := seedCopyableCreature(g, me.ID, "Isamaru", "Legendary Creature — Dog", 2, 2)
	castCatalogSpell(t, g, "Cytoshape", "Instant", oracleCytoshape,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	var prompt *game.PendingChoice
	for i := 0; i < 16 && prompt == nil; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseCards {
				prompt = c
			}
		}
		if prompt == nil {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
	}
	if prompt == nil {
		t.Fatal("Cytoshape never asked")
	}
	for _, id := range prompt.ChooseCards {
		if id == legend {
			t.Fatal("Cytoshape offered a legendary creature")
		}
	}
	if err := g.ResolveChooseCards(prompt.ID, prompt.Chooser, []uuid.UUID{legend}); err == nil {
		t.Error("Cytoshape accepted a legendary creature")
	}
}

// --- Mirrorweave -----------------------------------------------------

// TestMirrorweaveCopiesEveryOtherCreatureUntilCleanup — the affected
// set is every OTHER creature, locked as the spell resolves.
func TestMirrorweaveCopiesEveryOtherCreatureUntilCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	model := seedCopyableCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	mine := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := seedCopyableCreature(g, opp.ID, "Isamaru", "Legendary Creature — Dog", 2, 2)
	land := dcSeed(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)

	castCatalogSpell(t, g, "Mirrorweave", "Instant", oracleMirrorweave,
		[]game.TargetRef{{Kind: game.TargetCard, ID: model}})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if c := dcCard(t, g, id); c.Name != "Hill Giant" || c.Effective().Power != 3 {
			t.Errorf("%s is %q, want a Hill Giant", id, c.Name)
		}
	}
	if c := dcCard(t, g, model); c.IsCopy() {
		t.Error("the target itself became a copy — it is not an OTHER creature")
	}
	if c := dcCard(t, g, land); c.Name != "Forest" {
		t.Errorf("a land became %q", c.Name)
	}
	late := seedCopyableCreature(g, opp.ID, "Latecomer", "Creature — Elf", 1, 1)
	if c := dcCard(t, g, late); c.Name != "Latecomer" {
		t.Errorf("a creature that entered after Mirrorweave resolved became %q (CR 611.2c)", c.Name)
	}
	dcCleanup(g)
	if c := dcCard(t, g, mine); c.Name != "Grizzly Bears" {
		t.Errorf("after cleanup: %q", c.Name)
	}
	if c := dcCard(t, g, theirs); c.Name != "Isamaru" || !c.IsLegendary() {
		t.Errorf("after cleanup: %q (legendary %v)", c.Name, c.IsLegendary())
	}
}

// TestMirrorweaveCannotTargetALegendaryCreature — "target NONLEGENDARY
// creature".
func TestMirrorweaveCannotTargetALegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	legend := seedCopyableCreature(g, me.ID, "Isamaru", "Legendary Creature — Dog", 2, 2)
	seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	err := castCatalogSpellErr(t, g, "Mirrorweave", "Instant", oracleMirrorweave,
		[]game.TargetRef{{Kind: game.TargetCard, ID: legend}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Mirrorweave on a legendary creature: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Unstable Shapeshifter -------------------------------------------

// dcEnter puts a creature onto the battlefield with the zone-move AND
// ETB events a real entry emits.
func dcEnter(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine string, p, tough int) uuid.UUID {
	t.Helper()
	id := seedCopyableCreature(g, owner, name, typeLine, p, tough)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: owner})
	})
	return id
}

// TestUnstableShapeshifterCopiesEachCreatureAndKeepsTheAbility — the
// copy carries the granted trigger, so it goes on copying; counters
// on the Shapeshifter stay (CR 707.2).
func TestUnstableShapeshifterCopiesEachCreatureAndKeepsTheAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shifter := dcSeed(g, me.ID, "Unstable Shapeshifter", "Creature — Shapeshifter", oracleUnstableShapeshifter, 0, 1)
	if err := g.AddCounter(shifter, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}

	dcEnter(t, g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	passPriorityAroundTable(t, g)
	c := dcCard(t, g, shifter)
	if c.Name != "Grizzly Bears" || c.CurrentPower() != 3 {
		t.Fatalf("after one entry the Shapeshifter is %q with power %d, want Grizzly Bears 2+1", c.Name, c.CurrentPower())
	}
	if game.CatalogKey(c) != "oracle-Grizzly Bears|grant:"+unstableShapeshifterGrant {
		t.Errorf("the copy lost the granted ability: key %q", game.CatalogKey(c))
	}

	dcEnter(t, g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	passPriorityAroundTable(t, g)
	if c := dcCard(t, g, shifter); c.Name != "Hill Giant" || c.CurrentPower() != 4 {
		t.Fatalf("after a second entry the Shapeshifter is %q (%d), want Hill Giant — the granted trigger has to keep working", c.Name, c.CurrentPower())
	}
	dcCleanup(g)
	if c := dcCard(t, g, shifter); c.Name != "Hill Giant" {
		t.Errorf("an indefinite copy ended at cleanup: %q", c.Name)
	}
	if n := len(g.ScopedEffects); n != 1 {
		t.Errorf("registry holds %d records, want only the newest copy", n)
	}
}

// TestUnstableShapeshifterIgnoresANoncreature — "another CREATURE".
func TestUnstableShapeshifterIgnoresANoncreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shifter := dcSeed(g, me.ID, "Unstable Shapeshifter", "Creature — Shapeshifter", oracleUnstableShapeshifter, 0, 1)
	rock := dcSeed(g, me.ID, "Mind Stone", "Artifact", "", 0, 0)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: rock, Actor: me.ID}) })
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: shifter, Actor: me.ID}) })
	passPriorityAroundTable(t, g)
	if c := dcCard(t, g, shifter); c.Name != "Unstable Shapeshifter" || c.IsCopy() {
		t.Fatalf("a noncreature or its own entry made the Shapeshifter %q", c.Name)
	}
}

// --- Lazav, Dimir Mastermind -----------------------------------------

func dcPassUntilTriggerPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) bool {
	t.Helper()
	for i := 0; i < 16; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
				return true
			}
		}
		if stackFullyEmpty(g) {
			return false
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return false
}

// TestLazavBecomesTheCreatureCardKeepingItsNameAndAbility — the except
// clause: named Lazav, legendary, hexproof, and still watching.
func TestLazavBecomesTheCreatureCardKeepingItsNameAndAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lazav := dcSeed(g, me.ID, "Lazav, Dimir Mastermind", "Legendary Creature — Shapeshifter", oracleLazavDimirMastermind, 3, 3)
	hoof := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: hoof, Name: "Craterhoof Behemoth", TypeLine: "Creature — Beast",
		OracleID: "oracle-hoof", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})

	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, hoof); err != nil {
		t.Fatalf("discard the creature card: %v", err)
	}
	if !dcPassUntilTriggerPrompt(t, g, me.ID) {
		t.Fatal("Lazav did not trigger on a creature card going to an opponent's graveyard")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	c := dcCard(t, g, lazav)
	if c.Name != "Lazav, Dimir Mastermind" || !c.IsLegendary() || c.Effective().Power != 5 || !c.HasSubtype("Beast") {
		t.Fatalf("Lazav became %q (legendary %v, power %d, Beast %v)", c.Name, c.IsLegendary(), c.Effective().Power, c.HasSubtype("Beast"))
	}
	if !game.HasKeyword(&c, "hexproof") {
		t.Error("the copied Lazav lost hexproof")
	}

	// The granted ability keeps watching.
	giant := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: giant, Name: "Hill Giant", TypeLine: "Creature — Giant",
		OracleID: "oracle-giant", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, giant); err != nil {
		t.Fatalf("discard the second creature card: %v", err)
	}
	if !dcPassUntilTriggerPrompt(t, g, me.ID) {
		t.Fatal("the copied Lazav no longer triggers — the granted ability was lost")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if c := dcCard(t, g, lazav); c.Name != "Lazav, Dimir Mastermind" || c.Effective().Power != 3 || !c.HasSubtype("Giant") {
		t.Fatalf("second copy: %q power %d", c.Name, c.Effective().Power)
	}
}

// TestLazavIgnoresItsOwnGraveyardAndNoncreatures — "a CREATURE card
// … into an OPPONENT's graveyard".
func TestLazavIgnoresItsOwnGraveyardAndNoncreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dcSeed(g, me.ID, "Lazav, Dimir Mastermind", "Legendary Creature — Shapeshifter", oracleLazavDimirMastermind, 3, 3)
	mine := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mine, Name: "Grizzly Bears", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	rock := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: rock, Name: "Sol Ring", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, mine); err != nil {
		t.Fatal(err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, rock); err != nil {
		t.Fatal(err)
	}
	if dcPassUntilTriggerPrompt(t, g, me.ID) {
		t.Fatal("Lazav triggered on its controller's own card or on a noncreature card")
	}
}

// --- Snapshot ---------------------------------------------------------

// TestADurationCopyWithAGrantIsARestorePoint — a copy whose except
// clause granted a catalog bundle survives a restore point written to
// disk, and the granted trigger still keys.
func TestADurationCopyWithAGrantIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shifter := dcSeed(g, me.ID, "Unstable Shapeshifter", "Creature — Shapeshifter", oracleUnstableShapeshifter, 0, 1)
	dcEnter(t, g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	passPriorityAroundTable(t, g)

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded game.GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	c := dcCard(t, restored, shifter)
	if c.Name != "Grizzly Bears" || game.CatalogKey(c) != "oracle-Grizzly Bears|grant:"+unstableShapeshifterGrant {
		t.Fatalf("restored Shapeshifter is %q keyed %q", c.Name, game.CatalogKey(c))
	}
}

// TestATreasureThatBecomesACreatureLosesItsManaAbilityUntilCleanup — a
// token carries its abilities on the card as well as in the catalog. A
// copy has to take them away with the rest of its text and give them
// back when it ends.
func TestATreasureThatBecomesACreatureLosesItsManaAbilityUntilCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	tok := TreasureToken()
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = me.ID, me.ID
	treasure := pushBattlefieldCardWithTimestamp(g, tok)
	if len(game.ManaAbilitiesForCard(dcCard(t, g, treasure))) == 0 {
		t.Fatal("setup: the Treasure has no mana ability")
	}
	g.WithWriteLock(func() {
		v, _ := g.CopiableValuesForEffect(bears)
		g.BecomeCopyForEffect(uuid.Nil, bears, []uuid.UUID{treasure}, v, g.UntilEndOfTurnDuration(), "probe")
	})
	if n := len(game.ManaAbilitiesForCard(dcCard(t, g, treasure))); n != 0 {
		t.Errorf("the Treasure-as-Bears still has %d mana abilities", n)
	}
	dcCleanup(g)
	if n := len(game.ManaAbilitiesForCard(dcCard(t, g, treasure))); n == 0 {
		t.Error("after cleanup the Treasure has no mana ability")
	}
}

// --- Dimir Doppelganger ------------------------------------------------

const oracleDimirDoppelganger = "1916f120-4418-404a-aed5-b95ebf60d3a3"

// TestDimirDoppelgangerExilesAndBecomesTheCardKeepingTheAbility — the
// card is exiled, the Doppelganger becomes it for good, and the granted
// ability can be activated again from the copy.
func TestDimirDoppelgangerExilesAndBecomesTheCardKeepingTheAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dopp := dcSeed(g, me.ID, "Dimir Doppelganger", "Creature — Shapeshifter", oracleDimirDoppelganger, 0, 2)
	giant := dcGraveyardCard(g, opp, "Hill Giant", "Creature — Giant", "oracle-giant", 3, 3)
	hoof := dcGraveyardCard(g, me, "Craterhoof Behemoth", "Creature — Beast", "oracle-hoof", 5, 5)

	dcMana(t, g, me.ID, "{C}{U}{B}")
	if err := g.ActivateCatalogAbility(me.ID, dopp, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: giant}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(giant); z == nil || z.Kind != game.ZoneExile {
		t.Error("the targeted card was not exiled")
	}
	if c := dcCard(t, g, dopp); c.Name != "Hill Giant" || c.Effective().Power != 3 {
		t.Fatalf("the Doppelganger is %q %d/%d", c.Name, c.Effective().Power, c.Effective().Toughness)
	}
	dcCleanup(g)
	if c := dcCard(t, g, dopp); c.Name != "Hill Giant" {
		t.Fatalf("an indefinite copy ended at cleanup: %q", c.Name)
	}

	dcMana(t, g, me.ID, "{C}{U}{B}")
	if err := g.ActivateCatalogAbility(me.ID, dopp, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hoof}},
	}); err != nil {
		t.Fatalf("the granted ability cannot be activated from the copy: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := dcCard(t, g, dopp); c.Name != "Craterhoof Behemoth" {
		t.Fatalf("second activation: %q", c.Name)
	}
}

// TestDimirDoppelgangerRefusesANoncreatureCard — "target CREATURE card".
func TestDimirDoppelgangerRefusesANoncreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dopp := dcSeed(g, me.ID, "Dimir Doppelganger", "Creature — Shapeshifter", oracleDimirDoppelganger, 0, 2)
	ring := dcGraveyardCard(g, me, "Sol Ring", "Artifact", "oracle-sol", 0, 0)
	dcMana(t, g, me.ID, "{C}{U}{B}")
	if err := g.ActivateCatalogAbility(me.ID, dopp, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ring}},
	}); err == nil {
		t.Fatal("Dimir Doppelganger targeted an artifact card")
	}
}
