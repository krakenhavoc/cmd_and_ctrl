package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b161_cards_test.go — Jumpstart 2022 creatures, GAP batch 161 (#5948).
// Every test plays the real catalog entry through the engine.

const (
	gapPranksters = "4b02df52-d6c0-4735-b2b8-d637cb41f0ba"
	gapKinseekers = "67ff2880-1998-48e9-88c7-2aa012697401"
	gapSentinel   = "ede92fb6-dfd6-40b4-b58f-af1f7a155efc"
	gapMirrorImg  = "7275fb8c-571a-4c4c-889a-829419332e51"
	gapApprentice = "e68d6459-2f32-4320-a062-6d33f85348f9"
	gapHawk       = "8dbe7a06-3935-44a3-ac47-4ef00ed990f1"
	gapPiratedCpy = "09ea5508-78fd-4bbc-8395-38ec13202d9a"
	gapRiverSneak = "678095cc-ec8f-471a-b675-e59905d65c33"
)

func TestGapB161EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{gapPranksters, gapKinseekers, gapSentinel, gapMirrorImg, gapApprentice, gapHawk, gapPiratedCpy, gapRiverSneak} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func gapEmitCast(g *game.Game, caster uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventCast, Actor: caster, CardID: uuid.New()})
	})
}

func TestGapB161PrankstersBouncesOnlyOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Glen Elendra Pranksters", "Creature — Faerie Wizard", gapPranksters, 1, 3)
	bear := effectstest.SeedBear(g, me.ID)

	// Your own turn: casting a spell does nothing.
	if g.Seats[g.Turn.ActiveSeat].ID != me.ID {
		t.Fatalf("active seat is not seat 0")
	}
	gapEmitCast(g, me.ID)
	effectstest.AlaB1Settle(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("the Pranksters triggered on its controller's own turn")
	}
	if effectstest.FindBattlefieldCardByID(g, bear) == nil {
		t.Fatal("the Bear left the battlefield on the controller's own turn")
	}

	// An opponent's turn: the trigger targets, and the may is answered yes.
	effectstest.AdvanceToMainOf(t, g, 1)
	gapEmitCast(g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AlaB1Settle(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AlaB1Settle(t, g)
	if effectstest.FindBattlefieldCardByID(g, bear) != nil {
		t.Error("the Bear is still on the battlefield")
	}
	if !me.Hand.Contains(bear) {
		t.Error("the Bear did not return to its owner's hand")
	}
}

func TestGapB161PrankstersMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.B12Push(g, me.ID, "Glen Elendra Pranksters", "Creature — Faerie Wizard", gapPranksters, 1, 3)
	bear := effectstest.SeedBear(g, me.ID)
	gapEmitCast(g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AlaB1Settle(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.AlaB1Settle(t, g)
	if effectstest.FindBattlefieldCardByID(g, bear) == nil {
		t.Error("the Bear was returned although the may was declined")
	}
}

func gapKinseekersEnter(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return effectstest.CastHoldCreature(t, g, "Littjara Kinseekers", "Creature — Shapeshifter", gapKinseekers, 2, 4)
}

func TestGapB161KinseekersGrowsWithThreeSharingAType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedCopyableCreature(g, me.ID, "Elf A", "Creature — Elf", 1, 1)
	effectstest.SeedCopyableCreature(g, me.ID, "Elf B", "Creature — Elf", 1, 1)
	id := gapKinseekersEnter(t, g)
	// The Kinseekers is a changeling, so it is an Elf too: three Elves.
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, id); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry) == nil {
		t.Error("no scry was asked")
	}
}

func TestGapB161KinseekersDoesNothingWithoutThreeSharing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedCopyableCreature(g, me.ID, "Elf A", "Creature — Elf", 1, 1)
	effectstest.SeedCopyableCreature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	id := gapKinseekersEnter(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// Only two creatures share Elf (or Bear) with the changeling: Elf A and the
	// Kinseekers; the Kinseekers + Bear share Bear. No three share a type.
	if got := effectstest.PlusOneCounters(g, id); got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0", got)
	}
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry) != nil {
		t.Error("a scry was asked although the condition was false")
	}
}

func TestGapB161SentinelMayTapWhenAnArtifactEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Lumengrid Sentinel", "Creature — Human Wizard", gapSentinel, 1, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// A nonartifact entering does nothing.
	effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	effectstest.AlaB1Settle(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the Sentinel triggered for a nonartifact permanent")
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.AlaB1Settle(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AlaB1Settle(t, g)
	if !effectstest.IsTapped(g, victim) {
		t.Error("the target permanent was not tapped")
	}
}

func TestGapB161MirrorImageCopiesOnlyYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := effectstest.SeedCopyableCreature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)

	effectstest.CastCatalogSpell(t, g, "Mirror Image", "Creature — Shapeshifter", gapMirrorImg, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.CopyPrompt(g)
	if prompt == nil {
		t.Fatal("no copy prompt")
	}
	if effectstest.HasID(prompt.CopyOptions, theirs) {
		t.Error("an opponent's creature was offered")
	}
	if !effectstest.HasID(prompt.CopyOptions, mine) {
		t.Error("my own creature was not offered")
	}
	effectstest.ResolveWithCopyChoice(t, g, mine)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Grizzly Bears" {
			copies++
		}
	}
	if copies != 2 {
		t.Errorf("%d Grizzly Bears under my control, want 2", copies)
	}
}

func TestGapB161ApprenticeMillsOpponentsThenScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		effectstest.AwbPadLibrary(p)
	}
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Library.Size()
	}
	effectstest.CastHoldCreature(t, g, "Overwhelmed Apprentice", "Creature — Human Wizard", gapApprentice, 1, 2)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 8 && effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry) == nil; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, p := range g.Seats {
		want := before[p.ID]
		if p.ID != me.ID {
			want -= 2
		}
		if got := p.Library.Size(); got != want {
			t.Errorf("%s's library is %d, want %d", p.Name, got, want)
		}
	}
	scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil || len(scry.ScryCards) != 2 || scry.Chooser != me.ID {
		t.Fatalf("scry prompt = %+v, want my scry 2", scry)
	}
}

func TestGapB161HawkLootsForSnowMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.AwbPadLibrary(me)
	hawk := effectstest.BlbB1Push(g, me.ID, "Pilfering Hawk", "Snow Creature — Bird", gapHawk, 1, 2)
	effectstest.BlbB1Push(g, me.ID, "Plains", "Basic Land — Plains", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, hawk, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("the snow cost was paid with a nonsnow land")
	}
	effectstest.BlbB1Push(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island", "", 0, 0)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, hawk, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("activating with a snow land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d after the draw, want %d (the discard is still owed)", got, hand+1)
	}
	if !effectstest.IsTapped(g, hawk) {
		t.Error("the Hawk is untapped after paying {T}")
	}
}

func TestGapB161PiratedCopyDrawsOnSameNameCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(me)
	bears := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id := effectstest.CastCatalogSpell(t, g, "Pirated Copy", "Creature — Shapeshifter Pirate", gapPiratedCpy, nil)
	effectstest.ResolveWithCopyChoice(t, g, bears)
	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Grizzly Bears" {
		t.Fatalf("copy is %q, want Grizzly Bears", got.Name)
	}
	if !hasSubtype(got.TypeLine, "Pirate") || !hasSubtype(got.TypeLine, "Bear") {
		t.Errorf("type line = %q, want Bear and Pirate", got.TypeLine)
	}
	// The original Bears deals combat damage to a player: same name, so draw.
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: bears, Target: opp.ID, Amount: 2, Combat: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d after a same-name creature hit a player", got, hand+1)
	}
	// A differently named creature does not draw.
	other := effectstest.SeedCopyableCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	hand = me.Hand.Size()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: other, Target: opp.ID, Amount: 3, Combat: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand is %d, want %d: a differently named creature drew", got, hand)
	}
}

func hasSubtype(typeLine, sub string) bool {
	for i := 0; i+len(sub) <= len(typeLine); i++ {
		if typeLine[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestGapB161RiverSneakIsUnblockableAndGrowsForMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sneak := effectstest.B12Push(g, me.ID, "River Sneak", "Creature — Merfolk Warrior", gapRiverSneak, 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AlaB1Settle(t, g)
	if got := effectstest.EffectivePower(t, g, sneak); got != 1 {
		t.Fatalf("power = %d after a non-Merfolk entered, want 1", got)
	}
	effectstest.CastHoldCreature(t, g, "Merfolk Pal", "Creature — Merfolk", "", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, sneak); got != 2 {
		t.Errorf("power = %d after another Merfolk entered, want 2", got)
	}
	effectstest.AssertRestrictions(t, g, sneak, game.CantBeBlocked)
}
