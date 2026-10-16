package game

import (
	"testing"

	"github.com/google/uuid"
)

// exiled_with_test.go — #2530: Card.ExiledWith and the filter that reads it.

func TestPermissionFilterExiledWithSourceNamesTheObject(t *testing.T) {
	src := PermissionCardRef{ID: uuid.New(), Epoch: 2}
	f := PermissionFilter{ExiledWithSource: true, ExiledWith: src}

	if !f.Matches(Card{ExiledWith: src}) {
		t.Error("a card linked to the source object must match")
	}
	if f.Matches(Card{}) {
		t.Error("a card linked to nothing must not match")
	}
	if f.Matches(Card{ExiledWith: PermissionCardRef{ID: src.ID, Epoch: src.Epoch + 1}}) {
		t.Error("a card linked to an earlier incarnation of the source must not match (CR 400.7)")
	}
	if f.Matches(Card{ExiledWith: PermissionCardRef{ID: uuid.New(), Epoch: src.Epoch}}) {
		t.Error("a card linked to another permanent must not match")
	}
	// A flag with no derived object grants nothing, and in particular
	// not to a card that is linked to nothing either.
	bare := PermissionFilter{ExiledWithSource: true}
	if bare.Matches(Card{}) {
		t.Error("ExiledWithSource with no derived source matched an unlinked card")
	}
	// The flag off leaves the filter exactly as it was.
	if !(PermissionFilter{}).Matches(Card{ExiledWith: src}) {
		t.Error("a filter without the flag must ignore the link")
	}
}

func TestStampExiledWithOnlyLandsInExile(t *testing.T) {
	ref := PermissionCardRef{ID: uuid.New(), Epoch: 1}
	id := uuid.New()

	exile := &Zone{Kind: ZoneExile, Cards: []Card{{InstanceID: id}}}
	stampExiledWithLocked(exile, id, ref, AbilityRef{})
	if exile.Cards[0].ExiledWith != ref {
		t.Errorf("exile card link = %+v, want %+v", exile.Cards[0].ExiledWith, ref)
	}

	grave := &Zone{Kind: ZoneGraveyard, Cards: []Card{{InstanceID: id}}}
	stampExiledWithLocked(grave, id, ref, AbilityRef{})
	if grave.Cards[0].ExiledWith.ID != uuid.Nil {
		t.Error("a card that did not land in exile was linked")
	}

	none := &Zone{Kind: ZoneExile, Cards: []Card{{InstanceID: id}}}
	stampExiledWithLocked(none, id, PermissionCardRef{}, AbilityRef{})
	if none.Cards[0].ExiledWith.ID != uuid.Nil {
		t.Error("the zero ref stamped a link")
	}
	stampExiledWithLocked(nil, id, ref, AbilityRef{}) // must not panic
}

func TestMoveCardClearsTheExiledWithLink(t *testing.T) {
	id := uuid.New()
	exile := &Zone{Kind: ZoneExile, Cards: []Card{{InstanceID: id, ExiledWith: PermissionCardRef{ID: uuid.New(), Epoch: 1}}}}
	hand := &Zone{Kind: ZoneHand}
	if _, err := MoveCard(exile, hand, id); err != nil {
		t.Fatal(err)
	}
	if got := hand.Cards[0].ExiledWith; got.ID != uuid.Nil {
		t.Errorf("a card that left exile still carries %+v", got)
	}
}

// #2651: the ability half of the link (CR 607.2a).

func TestStampExiledWithCarriesTheAbility(t *testing.T) {
	ref := PermissionCardRef{ID: uuid.New(), Epoch: 1}
	ab := AbilityRef{Key: "k", Slot: AbilitySlotTriggered, Ref: "own:0", Name: "exile it"}
	id := uuid.New()
	exile := &Zone{Kind: ZoneExile, Cards: []Card{{InstanceID: id}}}
	stampExiledWithLocked(exile, id, ref, ab)
	if got := exile.Cards[0].ExiledWithAbility; got != ab {
		t.Errorf("ExiledWithAbility = %+v, want %+v", got, ab)
	}
	if !exile.Cards[0].ExiledWithAbility.IsRow(AbilitySlotTriggered, "exile it") {
		t.Error("the stamped ref does not name its own row")
	}
	if exile.Cards[0].ExiledWithAbility.IsRow(AbilitySlotActivated, "exile it") {
		t.Error("a triggered row answered for the activated list")
	}
	if (AbilityRef{}).IsRow("", "") {
		t.Error("the zero ref names a row")
	}

	hand := &Zone{Kind: ZoneHand}
	if _, err := MoveCard(exile, hand, id); err != nil {
		t.Fatal(err)
	}
	if got := hand.Cards[0].ExiledWithAbility; got != (AbilityRef{}) {
		t.Errorf("a card that left exile still names the ability %+v", got)
	}
}

// The route's ability rides only its own link: a replacement that
// rewrote the event's link to its own source (Valgavoth) is not that
// ability's exile.
func TestReplacementEventKeepsTheAbilityOnlyForTheRoutesLink(t *testing.T) {
	src := PermissionCardRef{ID: uuid.New(), Epoch: 1}
	ab := AbilityRef{Key: "k", Slot: AbilitySlotActivated, Ref: "own:0", Name: "exile"}
	r := &zoneRoute{ExiledWith: src, ExiledWithAbility: ab}
	ev := &ReplacementEvent{ExiledWith: src, zoneRoute: r}
	if got := ev.exiledWithAbility(); got != ab {
		t.Errorf("ability = %+v, want the route's %+v", got, ab)
	}
	ev.ExiledWith = PermissionCardRef{ID: uuid.New(), Epoch: 1}
	if got := ev.exiledWithAbility(); got != (AbilityRef{}) {
		t.Errorf("a rewritten link kept the route's ability %+v", got)
	}
	if got := (&ReplacementEvent{ExiledWith: src}).exiledWithAbility(); got != (AbilityRef{}) {
		t.Errorf("an event with no route named an ability %+v", got)
	}
}

func TestPlayedCardExiledWithRow(t *testing.T) {
	src := PermissionCardRef{ID: uuid.New(), Epoch: 3}
	ab := AbilityRef{Key: "k", Slot: AbilitySlotTriggered, Ref: "own:0", Name: "exile"}
	in := Card{ExiledWith: src, ExiledWithAbility: ab}

	cast := Event{Kind: EventCast, OldZone: ZoneExile, NewZone: ZoneStack}
	stampPlayedExiledWith(&cast, &in)
	if !PlayedCardExiledWithRow(cast, src, AbilitySlotTriggered, "exile") {
		t.Error("a cast from exile of a linked card did not match")
	}
	if PlayedCardExiledWithRow(cast, src, AbilitySlotTriggered, "another row") {
		t.Error("another row of the same object matched")
	}
	if PlayedCardExiledWithRow(cast, PermissionCardRef{ID: src.ID, Epoch: 4}, AbilitySlotTriggered, "exile") {
		t.Error("another object of the same card matched (CR 400.7)")
	}

	land := Event{Kind: EventZoneMove, OldZone: ZoneExile, NewZone: ZoneBattlefield, Played: true}
	stampPlayedExiledWith(&land, &in)
	if !PlayedCardExiledWithRow(land, src, AbilitySlotTriggered, "exile") {
		t.Error("a land played from exile did not match")
	}
	put := Event{Kind: EventZoneMove, OldZone: ZoneExile, NewZone: ZoneBattlefield}
	stampPlayedExiledWith(&put, &in)
	if PlayedCardExiledWithRow(put, src, AbilitySlotTriggered, "exile") {
		t.Error("a card PUT onto the battlefield from exile is not played")
	}

	bare := Event{Kind: EventCast, OldZone: ZoneExile, NewZone: ZoneStack}
	stampPlayedExiledWith(&bare, &Card{ExiledWith: src})
	if bare.ExiledWithAbility != nil || PlayedCardExiledWithRow(bare, src, AbilitySlotTriggered, "exile") {
		t.Error("a link with no ability matched a row")
	}
	none := Event{Kind: EventCast, OldZone: ZoneExile, NewZone: ZoneStack}
	stampPlayedExiledWith(&none, &Card{})
	if none.ExiledWith != nil {
		t.Error("an unlinked card stamped a link on its play")
	}
}

func TestExiledWithAbilitySurvivesASnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := PermissionCardRef{ID: uuid.New(), Epoch: 2}
	ab := AbilityRef{Key: "k", Slot: AbilitySlotTriggered, Ref: "own:1", Name: "exile"}
	linked, plain := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(Card{InstanceID: linked, Name: "Linked", Owner: me.ID, Controller: me.ID, ExiledWith: src, ExiledWithAbility: ab})
		g.Exile.PushTop(Card{InstanceID: plain, Name: "Plain", Owner: me.ID, Controller: me.ID, ExiledWith: src})
	})
	_, restored := roundTrip(t, g)
	restored.WithWriteLock(func() {
		got := restored.ExiledWithRowForEffect(src, AbilitySlotTriggered, "exile")
		if len(got) != 1 || got[0].InstanceID != linked {
			t.Fatalf("restored row read = %d cards, want the linked one", len(got))
		}
		if n := len(restored.ExiledWithForEffect(src)); n != 2 {
			t.Errorf("restored object read = %d cards, want 2", n)
		}
	})
}
