package game

import "testing"

// TestWarpInTextFindsTheKeywordLine pins the reader against Oracle text
// as Scryfall ships it: warp is on a card when a line of its own starts
// "Warp [cost]" (CR 702.185a), and a card that grants warp or asks about
// it does not have it.
func TestWarpInTextFindsTheKeywordLine(t *testing.T) {
	cases := []struct {
		name string
		text string
		want bool
	}{
		{"Anticausal Vestige", "When this creature leaves the battlefield, draw a card, then you may put a permanent card with mana value less than or equal to the number of lands you control from your hand onto the battlefield tapped.\nWarp {4} (You may cast this card from your hand for its warp cost. Exile this creature at the beginning of the next end step, then you may cast it from exile on a later turn.)", true},
		{"warp first, no reminder text", "Warp {1}{R}\nHaste", true},
		{"a non-mana warp cost", "Warp—Pay 3 life.", true},
		{"Tannuk, Steadfast Second (grants warp)", "Other creatures you control have haste.\nArtifact cards and red creature cards in your hand have warp {2}{R}. (You may cast a card from your hand for its warp cost. Exile that creature at the beginning of the next end step, then you may cast it from exile on a later turn.)", false},
		{"Blade of the Swarm (asks about warp)", "When this creature enters, choose one —\n• Put two +1/+1 counters on this creature.\n• Put target exiled card with warp on the bottom of its owner's library.", false},
		{"a word that only starts with warp", "Warped creatures you control get +1/+1.", false},
		{"a vanilla creature", "", false},
	}
	for _, tc := range cases {
		if got := WarpInText(tc.text); got != tc.want {
			t.Errorf("%s: WarpInText = %v, want %v", tc.name, got, tc.want)
		}
	}
	if !WarpInText("", "Flying", "Warp {2}{U}") {
		t.Error("warp on any one of the texts counts")
	}
}

// TestHasWarpReadsTheStampAndIsSilentFaceDown: a card the catalog does
// not run has warp when the importer's stamp says so (#5688), a card the
// importer never read is not known to, and a face-down card has no
// characteristics to ask (CR 406.3a, CR 708.2a).
func TestHasWarpReadsTheStampAndIsSilentFaceDown(t *testing.T) {
	stamped := Card{Name: "Uncatalogued Warper", TypeLine: "Creature — Drix", WarpText: PrintedYes}
	if !HasWarp(stamped) {
		t.Error("a card whose printed text has warp has warp")
	}
	if HasWarp(Card{Name: "Read, No Warp", TypeLine: "Creature — Bear", WarpText: PrintedNo}) {
		t.Error("a card whose printed text has no warp does not have it")
	}
	if HasWarp(Card{Name: "Never Read", TypeLine: "Creature — Bear"}) {
		t.Error("a card nobody read and the catalog does not know is not known to have warp")
	}
	hidden := stamped
	hidden.SetFaceDown(FaceDownExiled)
	if HasWarp(hidden) {
		t.Error("a card exiled face down has no characteristics (CR 406.3a)")
	}
}

// TestSetFaceRederivesWarpText: the face that is up is the face whose
// text is read, on a card the importer read; a card it never read keeps
// PrintedUnknown.
func TestSetFaceRederivesWarpText(t *testing.T) {
	c := Card{
		Name: "Front // Back", Layout: LayoutTransform, WarpText: PrintedYes,
		Faces: []Face{
			{Name: "Front", TypeLine: "Creature — Drix", OracleText: "Warp {1}{U}"},
			{Name: "Back", TypeLine: "Creature — Drix", OracleText: "Flying"},
		},
	}
	c.SetFace(1)
	if c.WarpText != PrintedNo {
		t.Errorf("back face up: WarpText = %v, want PrintedNo", c.WarpText)
	}
	c.SetFace(0)
	if c.WarpText != PrintedYes {
		t.Errorf("front face up: WarpText = %v, want PrintedYes", c.WarpText)
	}
	unread := Card{Name: "Front // Back", Layout: LayoutTransform, Faces: c.Faces}
	unread.SetFace(0)
	if unread.WarpText != PrintedUnknown {
		t.Errorf("a card the importer never read: WarpText = %v, want PrintedUnknown", unread.WarpText)
	}
}
