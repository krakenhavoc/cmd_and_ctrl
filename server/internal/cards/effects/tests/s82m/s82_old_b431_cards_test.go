package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 431 (Planechase 2012): Noggle Ransacker, Auratouched Mage, Last
// Stand, See Beyond, Whirlpool Warrior, Brutalizer Exarch.

const (
	oldB431Noggle   = "162aed84-f87f-4bd4-9cac-5008766bdfb7"
	oldB431Mage     = "3c7ffc6c-88be-4127-a673-fba8c1aa5fa1"
	oldB431Stand    = "4d2a465e-9ebd-4002-b6cd-e0eab08bad54"
	oldB431Beyond   = "4f9942ee-d3fd-4c01-93eb-870d642ebd83"
	oldB431Warrior  = "b7a1c54b-bb2f-4423-865a-ad114cd18c76"
	oldB431Brutaliz = "bdec14d1-1b38-4223-869e-a0ffc76b2b0c"
)

func TestOldB431EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB431Noggle: "Noggle Ransacker", oldB431Mage: "Auratouched Mage", oldB431Stand: "Last Stand",
		oldB431Beyond: "See Beyond", oldB431Warrior: "Whirlpool Warrior", oldB431Brutaliz: "Brutalizer Exarch",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

func oldB431Fill(g *game.Game, n int) {
	for _, p := range g.Seats {
		effectstest.SeedLibrary(p, "L1", "L2", "L3", "L4", "L5", "L6", "L7", "L8", "L9", "L10")
		_ = n
	}
}

// --- Noggle Ransacker -------------------------------------------------------

func TestOldB431NoggleEveryPlayerDrawsTwoAndDiscardsOneAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	oldB431Fill(g, 10)
	hands := make([]int, len(g.Seats))
	graves := make([]int, len(g.Seats))
	effectstest.B12PlayFromHand(t, g, "Noggle Ransacker", "Creature — Noggle Rogue", oldB431Noggle, game.CastSpellParams{})
	for i, p := range g.Seats {
		hands[i], graves[i] = len(p.Hand.Cards), len(p.Graveyard.Cards)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if got := len(p.Hand.Cards) - hands[i]; got != 1 {
			t.Errorf("%s: hand changed by %d, want +1 (two drawn, one discarded)", p.Name, got)
		}
		if got := len(p.Graveyard.Cards) - graves[i]; got != 1 {
			t.Errorf("%s: graveyard changed by %d, want +1", p.Name, got)
		}
	}
}

// --- See Beyond -------------------------------------------------------------

func TestOldB431SeeBeyondDrawsTwoThenShufflesAHandCardIntoTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	effectstest.B12PlayFromHand(t, g, "See Beyond", "Sorcery", oldB431Beyond, game.CastSpellParams{})
	handBefore, libBefore := len(me.Hand.Cards), len(me.Library.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no pick for the card to shuffle away: %+v", g.PendingChoices)
	}
	if len(me.Hand.Cards) != handBefore+2 {
		t.Fatalf("hand %d after the draws, want %d", len(me.Hand.Cards), handBefore+2)
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("pick bounds %d..%d, want exactly one", c.ChooseMin, c.ChooseMax)
	}
	pick := me.Hand.Cards[0].InstanceID
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pick}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if !me.Library.Contains(pick) {
		t.Error("the picked card is not in the library")
	}
	if len(me.Hand.Cards) != handBefore+1 || len(me.Library.Cards) != libBefore-2+1 {
		t.Errorf("hand %d library %d, want %d and %d", len(me.Hand.Cards), len(me.Library.Cards), handBefore+1, libBefore-1)
	}
}

// --- Whirlpool Warrior ------------------------------------------------------

func TestOldB431WarriorEntersAndRedrawsOnlyYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	oldB431Fill(g, 10)
	me, opp := g.Seats[0], g.Seats[1]
	oppHand := append([]game.Card(nil), opp.Hand.Cards...)
	effectstest.B12PlayFromHand(t, g, "Whirlpool Warrior", "Creature — Merfolk Warrior", oldB431Warrior, game.CastSpellParams{})
	handBefore, libBefore := len(me.Hand.Cards), len(me.Library.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != handBefore || len(me.Library.Cards) != libBefore {
		t.Errorf("hand %d library %d, want %d and %d unchanged in size", len(me.Hand.Cards), len(me.Library.Cards), handBefore, libBefore)
	}
	if len(opp.Hand.Cards) != len(oppHand) {
		t.Errorf("an opponent's hand changed size: %d, want %d", len(opp.Hand.Cards), len(oppHand))
	}
	for i, c := range opp.Hand.Cards {
		if c.InstanceID != oppHand[i].InstanceID {
			t.Error("an opponent's hand was shuffled by the enters trigger")
			break
		}
	}
}

func TestOldB431WarriorSacrificeRedrawsEveryHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB431Fill(g, 10)
	w := effectstest.PushCatalogPermanent(g, me.ID, "Whirlpool Warrior", "Creature — Merfolk Warrior", oldB431Warrior, false)
	effectstest.FillPoolColored(me, "R", 1)
	sizes := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		sizes[i] = len(p.Hand.Cards)
	}
	before := make([][]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		for _, c := range p.Hand.Cards {
			before[i] = append(before[i], c.InstanceID)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(w) {
		t.Error("the Warrior was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if len(p.Hand.Cards) != sizes[i] {
			t.Errorf("%s: hand %d, want %d", p.Name, len(p.Hand.Cards), sizes[i])
		}
	}
	for i, p := range g.Seats {
		for _, id := range before[i] {
			if p.Hand.Contains(id) {
				continue
			}
			if !p.Library.Contains(id) {
				t.Errorf("%s: an old hand card is neither in hand nor library", p.Name)
			}
		}
	}
}

// --- Last Stand -------------------------------------------------------------

func oldB431Lands(g *game.Game, p *game.Player, subtype string, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushCatalogPermanent(g, p.ID, subtype, "Basic Land — "+subtype, "", false)
	}
}

func TestOldB431LastStandScalesEachSentenceWithItsBasicLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	oldB431Lands(g, me, "Swamp", 2)
	oldB431Lands(g, me, "Mountain", 3)
	oldB431Lands(g, me, "Forest", 2)
	oldB431Lands(g, me, "Plains", 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
	survivor := effectstest.B12Creature(g, opp.ID, "Survivor", "Creature — Ogre", 4, 4)
	lifeMe, lifeOpp := me.Life, opp.Life
	effectstest.B12PlayFromHand(t, g, "Last Stand", "Sorcery", oldB431Stand, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: victim}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != lifeOpp-4 {
		t.Errorf("opponent life %d, want %d (2 per Swamp)", opp.Life, lifeOpp-4)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the 3/3 survived damage equal to three Mountains")
	}
	if !g.Battlefield.Contains(survivor) {
		t.Error("an untargeted creature died")
	}
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 2 {
		t.Errorf("%d Saprolings, want 2 (one per Forest)", n)
	}
	if me.Life != lifeMe+2 {
		t.Errorf("my life %d, want %d (2 per Plains)", me.Life, lifeMe+2)
	}
}

func TestOldB431LastStandDrawsAndDiscardsOnePerIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	oldB431Lands(g, me, "Island", 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 1, 1)
	effectstest.B12PlayFromHand(t, g, "Last Stand", "Sorcery", oldB431Stand, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: victim}}})
	handBefore := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, me.ID); got != 2 {
		t.Fatalf("owed %d discards, want 2", got)
	}
	if len(me.Hand.Cards) != handBefore+2 {
		t.Errorf("hand %d after drawing two, want %d", len(me.Hand.Cards), handBefore+2)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if len(me.Hand.Cards) != handBefore {
		t.Errorf("hand %d after discarding two, want %d", len(me.Hand.Cards), handBefore)
	}
}

func TestOldB431LastStandNeedsAnOpponentAndACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 1, 1)
	id := effectstest.PushHandCardWithManaCost(me, "Last Stand", "Sorcery", "")
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == id {
			me.Hand.Cards[i].OracleID = oldB431Stand
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID}, {Kind: game.TargetCard, ID: victim}}}); err == nil {
		t.Error("Last Stand targeted its own controller as the opponent")
	}
}

// --- Auratouched Mage -------------------------------------------------------

func oldB431AuraLibrary(p *game.Player) (aura, other uuid.UUID) {
	other = effectstest.SeedLibraryTop(p, "Plain Card", "Instant", "{U}")
	aura = effectstest.SeedLibraryTop(p, "Spectral Gift", "Enchantment — Aura", "{1}{W}")
	return aura, other
}

func TestOldB431MageFetchesAnAuraOntoItselfWhileItIsStillThere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	aura, other := oldB431AuraLibrary(me)
	mage := effectstest.B12PlayFromHand(t, g, "Auratouched Mage", "Creature — Human Wizard", oldB431Mage, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if effectstest.SearchOptionNamed(g, c, "Plain Card") != uuid.Nil {
		t.Error("a non-Aura card was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Spectral Gift")
	card, ok := effectstest.BattlefieldCard(g, aura)
	if !ok {
		t.Fatalf("the Aura is in %q, want the battlefield", effectstest.B12ZoneOf(g, aura))
	}
	if card.AttachedTo.ID != mage {
		t.Errorf("the Aura is attached to %v, want the Mage", card.AttachedTo.ID)
	}
	if effectstest.B12ZoneOf(g, other) != game.ZoneLibrary {
		t.Error("the other card left the library")
	}
}

func TestOldB431MageGoneBeforeTheSearchPutsTheAuraInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	aura, _ := oldB431AuraLibrary(me)
	// The enters trigger resolves with the Mage no longer on the battlefield.
	mage := effectstest.PushGraveyardCardTyped(me, "Auratouched Mage", "Creature — Human Wizard")
	item := &game.StackItem{Controller: me.ID, SourceCardID: mage}
	g.WithWriteLock(func() {
		if err := effects.S82OldB431AuratouchedMageSearch(g, item); err != nil {
			t.Fatalf("resolve: %v", err)
		}
	})
	effectstest.AnswerSearchNamed(t, g, me.ID, "Spectral Gift")
	if got := effectstest.B12ZoneOf(g, aura); got != game.ZoneHand {
		t.Errorf("the Aura is in %q, want hand", got)
	}
}

// --- Brutalizer Exarch ------------------------------------------------------

func TestOldB431ExarchSearchesForACreatureAndPutsItOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	effectstest.SeedLibraryTop(me, "Filler", "Instant", "{U}")
	bear := effectstest.SeedLibraryTop(me, "Library Bear", "Creature — Bear", "{1}{G}")
	effectstest.SeedLibraryTop(me, "Filler Two", "Sorcery", "{R}")
	effectstest.CastAndResolveCreature(t, g, "Brutalizer Exarch", "Creature — Phyrexian Cleric", oldB431Brutaliz)
	effectstest.Cmdr2S43ChooseMode(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if effectstest.SearchOptionNamed(g, c, "Filler") != uuid.Nil {
		t.Error("a noncreature card was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Library Bear")
	_ = bear
	if got := effectstest.LibraryTopNames(me, 1); len(got) != 1 || got[0] != "Library Bear" {
		t.Errorf("library top is %v, want the creature found", got)
	}
}

func TestOldB431ExarchTucksANoncreaturePermanentUnderItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	art := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Brutalizer Exarch", "Creature — Phyrexian Cleric", oldB431Brutaliz)
	effectstest.Cmdr2S43ChooseMode(t, g, me.ID, 1)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, bear) {
			t.Error("a creature was offered as the noncreature permanent")
		}
		effectstest.PickCard(t, g, me.ID, art)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(art) {
		t.Fatal("the artifact is still on the battlefield")
	}
	if n := len(opp.Library.Cards); n == 0 || opp.Library.Cards[0].Name != "Their Relic" {
		t.Errorf("the artifact is not at the bottom of its owner's library (%d cards)", n)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature left the battlefield")
	}
}
