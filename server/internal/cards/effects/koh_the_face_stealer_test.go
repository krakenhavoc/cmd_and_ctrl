package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// koh_the_face_stealer_test.go — Koh, the Face Stealer and ADR 0155's
// 2026-11-05 amendment (borrowed triggered abilities). Each test plays the
// card through the engine: Koh's own exiles, the "Pay 1 life" choice, and
// what Koh then has — a lent trigger that fires, stops and survives a
// restore, and the lent activated abilities beside it.

const (
	kohOracle            = "2bbb33e3-7d03-457c-92bb-b0c943615fa1"
	kohDoomedOracle      = "a30907c0-fbde-4fd3-a8c7-f304305fcea7"
	kohCryptGhastOracle  = "a3c8d817-7949-4dae-b9f5-f9d952479270"
	kohCrimsonOpOracle   = "d8863742-6c1c-415a-ad27-7def6d119c6a"
	kohSoulWardenType    = "Creature — Human Cleric"
	kohLegendaryTypeLine = "Legendary Creature — Shapeshifter Spirit"
)

// kohTable is a game with Koh on the battlefield under the active seat.
func kohTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	koh := gaPut(g, me.ID, "Koh, the Face Stealer", kohLegendaryTypeLine, kohOracle)
	return g, me, opp, koh
}

// kohTakesTheDead puts a catalog creature on the battlefield under
// `owner`, destroys it, and answers Koh's "you may exile it" with yes.
// It returns the card's ID, which is the exiled card's.
func kohTakesTheDead(t *testing.T, g *game.Game, me, owner *game.Player, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := gaPut(g, owner.ID, name, typeLine, oracle)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy %s: %v", name, err)
		}
	})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if c := elExiledCard(t, g, id); c.ExiledWith.ID == uuid.Nil {
		t.Fatalf("%s was exiled with no link to Koh", name)
	}
	return id
}

// kohChoose activates "Pay 1 life: Choose a creature card exiled with
// Koh" and resolves it, answering the pick with `pick` when one is asked.
func kohChoose(t *testing.T, g *game.Game, me *game.Player, koh, pick uuid.UUID) {
	t.Helper()
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, koh, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Koh: %v", err)
	}
	if me.Life != life-1 {
		t.Fatalf("life %d → %d, want 1 paid", life, me.Life)
	}
	passPriorityAroundTable(t, g)
	if ask := latestChooseCardsFor(g, me.ID); ask != nil {
		if err := g.ResolveChooseCards(ask.ID, me.ID, []uuid.UUID{pick}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	if got := gaLayered(t, g, koh).ChosenCard.ID; got != pick {
		t.Fatalf("Koh's last chosen card is %s, want %s", got, pick)
	}
}

// kohCreatureEnters has a vanilla creature enter under `p` through the
// real token path, so it carries an EventETB, and settles the stack.
func kohCreatureEnters(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(p.ID, game.Card{
			Name: "Koh Test Bear", TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2,
		}, 1)
	})
	passPriorityAroundTable(t, g)
}

// kohQueuedTrigger is a triggered item from `source` waiting to be put on
// the stack (CR 603.3) or already on it.
func kohQueuedTrigger(g *game.Game, source uuid.UUID) *game.StackItem {
	for _, it := range g.PendingTriggers {
		if it != nil && it.SourceCardID == source {
			return it
		}
	}
	return triggerOnStack(g, source)
}

func TestKohEntersAndExilesUpToOneOtherTargetCreatureLinkedToIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	warden := gaPut(g, opp.ID, "Soul Warden", kohSoulWardenType, b04SoulWardenOracle)
	koh := rfdCast(t, g, "Koh, the Face Stealer", kohLegendaryTypeLine, kohOracle, 6, 6)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("Koh's enters trigger asked for no target")
	}
	if hasID(pick.PickTargetCards, koh) {
		t.Error("Koh is a legal target of its own \"other target creature\"")
	}
	pickCard(t, g, me.ID, warden)
	passPriorityAroundTable(t, g)
	if got := elExiledCard(t, g, warden).ExiledWith; got != elRef(t, g, koh) {
		t.Fatalf("the Warden is exiled linked to %+v, want this Koh object", got)
	}
	// The exile is one of Koh's two: the choice offers it.
	kohChoose(t, g, me, koh, warden)
}

func TestKohBorrowsTheChosenCardsTriggerUntilAnotherCardIsChosen(t *testing.T) {
	g, me, opp, koh := kohTable(t)
	warden := kohTakesTheDead(t, g, me, opp, "Soul Warden", kohSoulWardenType, b04SoulWardenOracle)
	elves := kohTakesTheDead(t, g, me, opp, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)

	// Nothing chosen yet: Koh has nothing of either card.
	life := me.Life
	kohCreatureEnters(t, g, me)
	if me.Life != life {
		t.Fatalf("life %d → %d before any choice, want no Soul Warden trigger on Koh", life, me.Life)
	}

	kohChoose(t, g, me, koh, warden)
	life = me.Life
	kohCreatureEnters(t, g, opp)
	if me.Life != life+1 {
		t.Fatalf("life %d → %d, want the lent \"whenever another creature enters, you gain 1 life\" to fire", life, me.Life)
	}
	if idx, _ := blBorrowedMana(t, g, koh); len(idx) != 0 {
		t.Errorf("Koh has %d lent mana rows with the Warden chosen, want none", len(idx))
	}

	// Choosing the Elves replaces the Warden: the trigger stops, and the
	// Elves' activated mana ability is lent.
	kohChoose(t, g, me, koh, elves)
	life = me.Life
	kohCreatureEnters(t, g, opp)
	if me.Life != life {
		t.Errorf("life %d → %d after the Elves were chosen, want the Warden's trigger gone", life, me.Life)
	}
	if idx, _ := blBorrowedMana(t, g, koh); len(idx) != 1 {
		t.Fatalf("Koh has %d lent mana rows, want the Elves' {T}: Add {G}", len(idx))
	}
	blTapLentMana(t, g, me, koh)
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool = %v, want [G] from the lent mana ability", got)
	}
}

func TestKohLosesTheTriggerWhenTheChosenCardLeavesExile(t *testing.T) {
	g, me, opp, koh := kohTable(t)
	warden := kohTakesTheDead(t, g, me, opp, "Soul Warden", kohSoulWardenType, b04SoulWardenOracle)
	kohChoose(t, g, me, koh, warden)
	elToHand(t, g, opp.ID, warden)
	life := me.Life
	kohCreatureEnters(t, g, opp)
	if me.Life != life {
		t.Errorf("life %d → %d, want nothing: the chosen card left exile (CR 400.7)", life, me.Life)
	}
	if n := len(game.TriggersForCard(gaLayered(t, g, koh))); n != 2 {
		t.Errorf("Koh has %d triggered abilities, want only its own 2", n)
	}
}

func TestKohsLentTriggerSurvivesASnapshotRoundTrip(t *testing.T) {
	g, me, opp, koh := kohTable(t)
	warden := kohTakesTheDead(t, g, me, opp, "Soul Warden", kohSoulWardenType, b04SoulWardenOracle)
	kohChoose(t, g, me, koh, warden)

	// A lent trigger waiting on the stack names the LENDER's catalog row,
	// with Koh as its source.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{
			Name: "Koh Test Bear", TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2,
		}, 1)
	})
	item := kohQueuedTrigger(g, koh)
	if item == nil {
		t.Fatal("no lent trigger from Koh waiting")
	}
	if a := item.Params.Ability; a == nil || a.Key != b04SoulWardenOracle || a.Slot != game.AbilitySlotTriggered {
		t.Fatalf("the lent trigger's ref = %+v, want the Soul Warden's triggered row", item.Params.Ability)
	}

	restored := restoreThroughJSON(t, g)
	if kohQueuedTrigger(restored, koh) == nil {
		t.Fatal("the restored game lost the lent trigger")
	}
	rMe := restored.Seats[g.Turn.ActiveSeat]
	life := rMe.Life
	passPriorityAroundTable(t, restored)
	if rMe.Life != life+1 {
		t.Errorf("restored: life %d → %d, want the lent trigger to resolve", life, rMe.Life)
	}
	// The choice itself is carried: Koh still has the trigger.
	life = rMe.Life
	kohCreatureEnters(t, restored, restored.Seats[(g.Turn.ActiveSeat+1)%len(restored.Seats)])
	if rMe.Life != life+1 {
		t.Errorf("restored: life %d → %d, want Koh to keep the chosen card's trigger", life, rMe.Life)
	}
}

func TestKohDyingHasTheChosenCardsDiesTrigger(t *testing.T) {
	g, me, opp, koh := kohTable(t)
	traveler := kohTakesTheDead(t, g, me, opp, "Doomed Traveler", "Creature — Human Soldier", kohDoomedOracle)
	kohChoose(t, g, me, koh, traveler)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(koh); err != nil {
			t.Fatalf("destroy Koh: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	spirits := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Spirit" {
			spirits++
		}
	}
	if spirits != 1 {
		t.Errorf("%d Spirits, want 1: Koh died with the Traveler's \"when this creature dies, create a 1/1 Spirit\" (CR 603.10a)", spirits)
	}
}

func TestKohBorrowsATriggeredKeywordAndATriggeredManaAbility(t *testing.T) {
	g, me, opp, koh := kohTable(t)
	op := kohTakesTheDead(t, g, me, opp, "Crimson Operative", "Artifact Creature — Human Villain", kohCrimsonOpOracle)
	kohChoose(t, g, me, koh, op)
	prowess := 0
	for _, tr := range game.TriggersForCard(gaLayered(t, g, koh)) {
		if tr.Keyword == game.KeywordProwess {
			prowess++
		}
	}
	if prowess != 1 {
		t.Errorf("Koh has %d prowess triggers with Crimson Operative chosen, want 1", prowess)
	}

	ghast := kohTakesTheDead(t, g, me, opp, "Crypt Ghast", "Creature — Spirit", kohCryptGhastOracle)
	kohChoose(t, g, me, koh, ghast)
	swamp := b31Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", "", 0, 0)
	tapForMana(t, g, me.ID, swamp)
	if got := sortedPool(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("Swamp pool = %v, want {B}{B}: Koh has Crypt Ghast's triggered mana ability", got)
	}
}

func TestKohChoosesOnlyCreatureCardsItsOwnAbilitiesExiled(t *testing.T) {
	g, me, opp, koh := kohTable(t)
	// Linked to the Koh object but by no ability of Koh's: not "exiled
	// with Koh" for the choice (CR 607.2a).
	stray := elPushExiled(g, opp.ID, elRef(t, g, koh), "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	warden := kohTakesTheDead(t, g, me, opp, "Soul Warden", kohSoulWardenType, b04SoulWardenOracle)
	// One candidate is chosen without a prompt.
	kohChoose(t, g, me, koh, warden)
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("a pick was asked with one creature card to choose")
	}
	if gaLayered(t, g, koh).ChosenCard.ID == stray {
		t.Error("Koh chose a card its own abilities did not exile")
	}
	// A token that dies is not exiled: "nontoken".
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{
			Name: "Koh Test Token", TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2,
		}, 1)
	})
	passPriorityAroundTable(t, g)
	tok := battlefieldIDNamed(g, opp.ID, "Koh Test Token")
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(tok); err != nil {
			t.Fatalf("destroy token: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("Koh offered to exile a token that died")
	}
}

func TestKohDeclinedOrGoneCardIsNotExiled(t *testing.T) {
	g, me, opp, koh := kohTable(t)
	_ = koh
	// Declined: the card stays in the graveyard.
	warden := gaPut(g, opp.ID, "Soul Warden", kohSoulWardenType, b04SoulWardenOracle)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(warden); err != nil {
			t.Fatal(err)
		}
	})
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(warden) {
		t.Error("a declined \"you may exile it\" exiled the card")
	}
	// Left the graveyard in response: not exiled (the ruling).
	elves := gaPut(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", blLlanowarOracle)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(elves); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, elves); err != nil {
		t.Fatalf("move the Elves to hand: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(elves) {
		t.Error("Koh exiled a card that had left the graveyard")
	}
}
