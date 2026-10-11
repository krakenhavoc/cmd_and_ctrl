package deckcoverage

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// TestAuditedAltarAndCutthroatLeaveUnreviewed pins #2954: after the audit
// both cards sit in the Automated bucket (declared Full, no caveats), not
// Unreviewed. Registry-backed, so a card that regresses to the zero-value
// declaration, or gains a caveat, fails here by name.
func TestAuditedAltarAndCutthroatLeaveUnreviewed(t *testing.T) {
	idx := cards.NewIndex()
	legal := map[string]string{"commander": "legal"}
	for _, c := range []cards.Card{
		{OracleID: uuid.MustParse("8d02b297-97c4-4379-9862-0a462400f66f"), Name: "Phyrexian Altar",
			TypeLine: "Artifact", OracleText: "Sacrifice a creature: Add one mana of any color."},
		{OracleID: uuid.MustParse("76b003e0-15af-4f22-bdf2-1ade5430964a"), Name: "Zulaport Cutthroat",
			TypeLine: "Creature — Human Rogue Ally", Power: "1", Toughness: "1", ManaCost: "{1}{B}",
			OracleText: "Whenever this creature or another creature you control dies, each opponent loses 1 life and you gain 1 life."},
	} {
		c.ID = uuid.New()
		c.Legalities = legal
		idx.Put(c)
	}
	r, err := Build(idx, Deck{Name: "audit", Entries: []deck.Entry{
		{Name: "Phyrexian Altar", Count: 1},
		{Name: "Zulaport Cutthroat", Count: 1},
	}})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(r.Cards) != 2 {
		t.Fatalf("report has %d cards, want 2", len(r.Cards))
	}
	for _, c := range r.Cards {
		if c.Bucket != Automated {
			t.Errorf("%s is in bucket %q, want %q", c.Name, c.Bucket, Automated)
		}
		if len(c.Caveats) != 0 {
			t.Errorf("%s carries caveats %v", c.Name, c.Caveats)
		}
	}
	if r.Counts[Unreviewed] != 0 || r.Needed() {
		t.Errorf("counts = %v, want nothing unreviewed", r.Counts)
	}
}
