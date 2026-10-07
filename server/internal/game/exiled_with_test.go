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
	stampExiledWithLocked(exile, id, ref)
	if exile.Cards[0].ExiledWith != ref {
		t.Errorf("exile card link = %+v, want %+v", exile.Cards[0].ExiledWith, ref)
	}

	grave := &Zone{Kind: ZoneGraveyard, Cards: []Card{{InstanceID: id}}}
	stampExiledWithLocked(grave, id, ref)
	if grave.Cards[0].ExiledWith.ID != uuid.Nil {
		t.Error("a card that did not land in exile was linked")
	}

	none := &Zone{Kind: ZoneExile, Cards: []Card{{InstanceID: id}}}
	stampExiledWithLocked(none, id, PermissionCardRef{})
	if none.Cards[0].ExiledWith.ID != uuid.Nil {
		t.Error("the zero ref stamped a link")
	}
	stampExiledWithLocked(nil, id, ref) // must not panic
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
