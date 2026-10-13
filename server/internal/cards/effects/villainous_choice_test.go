package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// villainous_choice_test.go — #3140, CR 701.55. The keyword action as a
// shape, tested through the cards that use it: the question is the
// facing player's and not the controller's, both options are always on
// offer (701.55b), "an additional time" is counted when the choice is
// faced and asked one at a time (701.55c), several players go in APNAP
// order (701.55d), and a table waiting on any question is a restore
// point.

const (
	villainDamocles     = "544c9ad2-86f9-483e-a53c-f64c8602c393"
	villainDalekEmperor = "07799012-7051-400e-adde-b54b41b80c70"
	villainValeyard     = "ed202f1d-3511-45f4-bfbc-e6b76d4311d1"
)

// villainFace runs FaceVillainousChoice for `players` as `controller`'s
// effect, sourced from `source`, the way a resolving trigger would.
func villainFace(t *testing.T, g *game.Game, controller, source uuid.UUID, c *VillainousChoice, players ...uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{
			Kind: game.StackItemSpell, Controller: controller, Owner: controller, SourceCardID: source,
		})
		err = FaceVillainousChoice{Players: players, Choice: c}.Apply(ctx)
	})
	if err != nil {
		t.Fatalf("FaceVillainousChoice: %v", err)
	}
}

// villainPrompt is the one open option pick, which must be `chooser`'s.
func villainPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	var open []*game.PendingChoice
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionPick {
				open = append(open, c)
			}
		}
	})
	if len(open) != 1 {
		t.Fatalf("%d option picks open, want exactly the one for %s: %+v", len(open), chooser, open)
	}
	if open[0].Chooser != chooser {
		t.Fatalf("the question is %s's, want %s's", open[0].Chooser, chooser)
	}
	return open[0]
}

func villainAnswer(t *testing.T, g *game.Game, c *game.PendingChoice, index int) {
	t.Helper()
	if err := g.ResolveOptionPick(c.ID, c.Chooser, index); err != nil {
		t.Fatalf("ResolveOptionPick(%d): %v", index, err)
	}
}

func villainNoPrompt(t *testing.T, g *game.Game) {
	t.Helper()
	if c := pendingOfKind(g, game.PendingChoiceOptionPick); c != nil {
		t.Fatalf("an option pick is still open for %s: %+v", c.Chooser, c.PickOptions)
	}
}

func villainHandSize(p *game.Player) int { return p.Hand.Size() }

func villainFillLibrary(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		pushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Basic Land — Forest"})
	}
}

// ---- who is asked, and what -------------------------------------------

func TestAVillainousChoiceIsAskedOfTheFacingPlayerNotTheController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, me.ID, "Damocles Base, Sword of Kang", "Legendary Artifact — Vehicle", villainDamocles, false)
	pushVanillaCreature(g, opp.ID, "Bear", 2, 2)

	villainFace(t, g, me.ID, src, damoclesBaseChoice, opp.ID)

	c := villainPrompt(t, g, opp.ID)
	if len(c.PickOptions) != 2 {
		t.Fatalf("a villainous choice has two options, got %+v", c.PickOptions)
	}
	if c.PickOptions[0].Label != "Sacrifice a nontoken creature of your choice" ||
		c.PickOptions[1].LifeCost != 2 {
		t.Errorf("options as printed, with the second's life cost for a bot: %+v", c.PickOptions)
	}
	if c.PickOptions[0].Harm == 0 {
		t.Errorf("a facing player who has a creature to lose is told it costs something: %+v", c.PickOptions[0])
	}
	if c.PickThen == "" {
		t.Error("the question must be a keyed continuation, or the table stops being a restore point")
	}
}

func TestChoosingTheSecondOptionRunsItsBody(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, me.ID, "Damocles Base, Sword of Kang", "Legendary Artifact — Vehicle", villainDamocles, false)
	villainFillLibrary(g, me, 5)
	lifeBefore, handBefore := opp.Life, villainHandSize(me)

	villainFace(t, g, me.ID, src, damoclesBaseChoice, opp.ID)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)

	if opp.Life != lifeBefore-2 {
		t.Errorf("the facing player lost %d life, want 2", lifeBefore-opp.Life)
	}
	if got := villainHandSize(me) - handBefore; got != 2 {
		t.Errorf("the controller drew %d cards, want 2", got)
	}
	villainNoPrompt(t, g)
}

func TestChoosingTheFirstOptionRunsItsBodyAndOffersOnlyWhatQualifies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, me.ID, "Damocles Base, Sword of Kang", "Legendary Artifact — Vehicle", villainDamocles, false)
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	token := pushDiesCreatureForTest(g, opp.ID, "Soldier", "", "Token Creature — Soldier", 1, 1)

	villainFace(t, g, me.ID, src, damoclesBaseChoice, opp.ID)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 0)

	c := pendingOfKind(g, game.PendingChoiceSacrifice)
	if c == nil {
		t.Fatalf("no sacrifice prompt: %+v", g.PendingChoices)
	}
	if c.Chooser != opp.ID {
		t.Errorf("the sacrifice is %s's, want the facing player's", c.Chooser)
	}
	for _, id := range c.SacrificeOptions {
		if id == token {
			t.Error("a token was offered for 'a nontoken creature'")
		}
	}
	if len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != bear {
		t.Errorf("eligible = %v, want just the nontoken Bear", c.SacrificeOptions)
	}
}

// CR 701.55b: a player may choose an option that is impossible and
// performs as much as it can. No creature to sacrifice is no cost.
func TestAnImpossibleOptionIsStillOnOfferAndCostsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, me.ID, "Damocles Base, Sword of Kang", "Legendary Artifact — Vehicle", villainDamocles, false)
	lifeBefore := opp.Life

	villainFace(t, g, me.ID, src, damoclesBaseChoice, opp.ID)
	c := villainPrompt(t, g, opp.ID)
	if len(c.PickOptions) != 2 {
		t.Fatalf("both options are offered even when one cannot be carried out: %+v", c.PickOptions)
	}
	if c.PickOptions[0].Harm != 0 || c.PickOptions[0].LifeCost != 0 {
		t.Errorf("an option that would do nothing is priced as free for a bot: %+v", c.PickOptions[0])
	}
	if c.PickOptions[1].LifeCost != 2 {
		t.Errorf("the possible option keeps its price: %+v", c.PickOptions[1])
	}
	villainAnswer(t, g, c, 0)
	if pendingOfKind(g, game.PendingChoiceSacrifice) != nil {
		t.Error("a sacrifice was asked of a player with nothing to sacrifice")
	}
	if opp.Life != lifeBefore {
		t.Errorf("choosing the impossible option cost %d life", lifeBefore-opp.Life)
	}
	villainNoPrompt(t, g)
}

// ---- CR 701.55c: an additional time -----------------------------------

func TestTheValeyardMakesAnOpponentFaceTheChoiceAnAdditionalTime(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, me.ID, "Damocles Base, Sword of Kang", "Legendary Artifact — Vehicle", villainDamocles, false)
	b43Catalog(g, me.ID, "The Valeyard", "Legendary Creature — Time Lord Noble", villainValeyard, 4, 5)
	villainFillLibrary(g, me, 8)
	lifeBefore := opp.Life

	villainFace(t, g, me.ID, src, damoclesBaseChoice, opp.ID)

	first := villainPrompt(t, g, opp.ID)
	villainAnswer(t, g, first, 1)
	// One at a time: the second question is asked only now.
	second := villainPrompt(t, g, opp.ID)
	if second.ID == first.ID {
		t.Fatal("the second question is the first again")
	}
	villainAnswer(t, g, second, 1)

	if opp.Life != lifeBefore-4 {
		t.Errorf("two answers of 'lose 2 life' cost %d, want 4", lifeBefore-opp.Life)
	}
	villainNoPrompt(t, g)
}

// "They can make the same or different choices."
func TestTheAdditionalTimeMayBeADifferentChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, me.ID, "Damocles Base, Sword of Kang", "Legendary Artifact — Vehicle", villainDamocles, false)
	b43Catalog(g, me.ID, "The Valeyard", "Legendary Creature — Time Lord Noble", villainValeyard, 4, 5)
	villainFillLibrary(g, me, 8)
	pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	lifeBefore := opp.Life

	villainFace(t, g, me.ID, src, damoclesBaseChoice, opp.ID)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 0)

	if opp.Life != lifeBefore-2 {
		t.Errorf("one 'lose 2 life' answer cost %d, want 2", lifeBefore-opp.Life)
	}
	if pendingOfKind(g, game.PendingChoiceSacrifice) == nil {
		t.Error("the second answer, sacrifice, was not carried out")
	}
}

func TestTheAdditionalTimeIsCountedWhenTheChoiceIsFaced(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	times := func(p uuid.UUID) (n int) {
		g.WithWriteLock(func() { n = g.VillainousChoiceTimesForEffect(p) })
		return n
	}
	if got := times(opp.ID); got != 1 {
		t.Fatalf("with no Valeyard a choice is faced %d times, want 1", got)
	}
	first := b43Catalog(g, me.ID, "The Valeyard", "Legendary Creature — Time Lord Noble", villainValeyard, 4, 5)
	if got := times(opp.ID); got != 2 {
		t.Errorf("one Valeyard: faced %d times, want 2", got)
	}
	if got := times(me.ID); got != 1 {
		t.Errorf("its controller is not an opponent of itself: faced %d times, want 1", got)
	}
	b43Catalog(g, me.ID, "The Valeyard", "Legendary Creature — Time Lord Noble", villainValeyard, 4, 5)
	if got := times(opp.ID); got != 3 {
		t.Errorf("two Valeyards: faced %d times, want 3 (the second applies to the modified instruction)", got)
	}
	// A third player's own Valeyard does not double their own choices,
	// but it does double mine.
	b43Catalog(g, g.Seats[2].ID, "The Valeyard", "Legendary Creature — Time Lord Noble", villainValeyard, 4, 5)
	if got := times(g.Seats[2].ID); got != 3 {
		t.Errorf("seat 2 controls one and faces the other two: %d, want 3", got)
	}
	if got := times(me.ID); got != 2 {
		t.Errorf("I control two and face seat 2's one: %d, want 2", got)
	}
	// A permanent the catalog no longer reads adds nothing (the static is
	// judged through CatalogAbilityKey, as every other one is).
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == first {
				g.Battlefield.Cards[i].OracleID = ""
			}
		}
	})
	if got := times(opp.ID); got != 3 {
		t.Errorf("one Valeyard no longer read: %d, want 3 (my other one and seat 2's, plus one)", got)
	}
}

// ---- CR 701.55d: several players, one at a time, in APNAP order ------

func TestSeveralPlayersFaceItOneAtATimeInTurnOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	order := []*game.Player{g.Seats[1], g.Seats[2], g.Seats[3]}
	src := pushCatalogPermanent(g, me.ID, "The Dalek Emperor", "Legendary Artifact Creature — Dalek", villainDalekEmperor, false)

	// Listed out of order on purpose: the engine asks them in APNAP order.
	villainFace(t, g, me.ID, src, daleksEmperorChoice, order[2].ID, order[0].ID, order[1].ID)

	for i, p := range order {
		c := villainPrompt(t, g, p.ID)
		villainAnswer(t, g, c, 1) // the Dalek token for the controller
		if i < len(order)-1 {
			// And only one is asked at a time.
			villainPrompt(t, g, order[i+1].ID)
		}
	}
	villainNoPrompt(t, g)
	if n := battlefieldCountNamed(g, me.ID, "Dalek"); n != 3 {
		t.Errorf("the controller has %d Dalek tokens, want one per opponent", n)
	}
}

func TestEachPlayersAdditionalTimesAreAskedBeforeTheNextPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := pushCatalogPermanent(g, me.ID, "The Dalek Emperor", "Legendary Artifact Creature — Dalek", villainDalekEmperor, false)
	b43Catalog(g, me.ID, "The Valeyard", "Legendary Creature — Time Lord Noble", villainValeyard, 4, 5)

	villainFace(t, g, me.ID, src, daleksEmperorChoice, g.Seats[1].ID, g.Seats[2].ID)

	for _, want := range []uuid.UUID{g.Seats[1].ID, g.Seats[1].ID, g.Seats[2].ID, g.Seats[2].ID} {
		villainAnswer(t, g, villainPrompt(t, g, want), 1)
	}
	villainNoPrompt(t, g)
}

func TestAPlayerWhoLeavesBeforeTheirTurnToChooseIsSkipped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := pushCatalogPermanent(g, me.ID, "The Dalek Emperor", "Legendary Artifact Creature — Dalek", villainDalekEmperor, false)

	villainFace(t, g, me.ID, src, daleksEmperorChoice, g.Seats[1].ID, g.Seats[2].ID)
	g.Seats[2].Eliminated = true
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[1].ID), 1)

	villainNoPrompt(t, g)
}

// ---- restore points ---------------------------------------------------

func TestAVillainousChoiceSurvivesCloneAndARestoreAndCarriesOn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, next := g.Seats[0], g.Seats[1], g.Seats[2]
	src := pushCatalogPermanent(g, me.ID, "The Dalek Emperor", "Legendary Artifact Creature — Dalek", villainDalekEmperor, false)
	b43Catalog(g, me.ID, "The Valeyard", "Legendary Creature — Time Lord Noble", villainValeyard, 4, 5)
	villainFace(t, g, me.ID, src, daleksEmperorChoice, opp.ID, next.ID)
	first := villainPrompt(t, g, opp.ID)

	// Clone (undo): the copy answers independently of the original.
	clone := g.Clone()
	villainAnswer(t, clone, villainPrompt(t, clone, opp.ID), 1)
	villainPrompt(t, clone, opp.ID) // its additional time
	villainPrompt(t, g, opp.ID)     // the original is untouched
	if c := villainPrompt(t, g, opp.ID); c.ID != first.ID {
		t.Fatal("answering the clone moved the original")
	}

	// Snapshot: an open villainous choice is a restore point.
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("an open villainous choice blocks the restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded game.GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	got := villainPrompt(t, restored, opp.ID)
	if got.ID != first.ID || len(got.PickOptions) != 2 {
		t.Fatalf("the restored question is not the same one: %+v", got)
	}
	// The whole chain carries on after the restore: the additional time,
	// then the next player.
	villainAnswer(t, restored, got, 1)
	villainAnswer(t, restored, villainPrompt(t, restored, opp.ID), 1)
	villainAnswer(t, restored, villainPrompt(t, restored, next.ID), 1)
	villainAnswer(t, restored, villainPrompt(t, restored, next.ID), 1)
	villainNoPrompt(t, restored)
	if n := battlefieldCountNamed(restored, me.ID, "Dalek"); n != 4 {
		t.Errorf("the restored chain made %d Dalek tokens, want 4 (two players, twice each)", n)
	}
}

// ---- the registered body is the card's -------------------------------

func TestTheEmperorsCreatesADalekTokenForTheController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, me.ID, "The Dalek Emperor", "Legendary Artifact Creature — Dalek", villainDalekEmperor, false)
	villainFace(t, g, me.ID, src, daleksEmperorChoice, opp.ID)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)

	var found *game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Dalek" {
			c := c
			found = &c
		}
	}
	if found == nil {
		t.Fatal("no Dalek token")
	}
	if found.Controller != me.ID {
		t.Errorf("the Dalek is controlled by %s, want the Emperor's controller", found.Controller)
	}
	if found.Power != 3 || found.Toughness != 3 || !found.HasSubtype("Dalek") {
		t.Errorf("token = %s %d/%d, want a 3/3 Dalek", found.TypeLine, found.Power, found.Toughness)
	}
}

// battlefieldCountNamed counts the battlefield permanents a player
// controls with this name.
func battlefieldCountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}
