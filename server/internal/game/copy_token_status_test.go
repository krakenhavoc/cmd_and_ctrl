package game

import "testing"

// #4296: token status is the copier's own (CR 111.1, CR 111.6, CR 707.2).
// The end-to-end cases (a token copy of Clone, a Metamorph copying a
// Treasure, a duration copy) are in cards/effects/token_status_copy_test.go;
// this pins the value helper they share.

func TestWithTokenStatusAddsAndStripsTheSupertype(t *testing.T) {
	for _, tc := range []struct {
		name, in string
		token    bool
		want     string
	}{
		{"card to token", "Creature — Bear", true, "Token Creature — Bear"},
		{"token to card", "Token Creature — Goblin", false, "Creature — Goblin"},
		{"token stays one token", "Token Creature — Goblin", true, "Token Creature — Goblin"},
		{"card stays a card", "Creature — Bear", false, "Creature — Bear"},
		{"other supertypes survive", "Legendary Creature — Human", true, "Token Legendary Creature — Human"},
		{"other supertypes survive stripping", "Token Legendary Artifact Creature — Golem", false, "Legendary Artifact Creature — Golem"},
		{"no subtypes", "Token Artifact", false, "Artifact"},
	} {
		got := PrintedValues{TypeLine: tc.in}.withTokenStatus(tc.token).TypeLine
		if got != tc.want {
			t.Errorf("%s: withTokenStatus(%q, %v) = %q, want %q", tc.name, tc.in, tc.token, got, tc.want)
		}
	}
}

// Every face of a double-faced value carries the status, or a transform
// would turn a token into a card (or the reverse).
func TestWithTokenStatusCoversEveryFaceAndNeverMutatesTheInput(t *testing.T) {
	in := PrintedValues{
		TypeLine: "Creature — Human Werewolf",
		Faces: []Face{
			{TypeLine: "Creature — Human Werewolf"},
			{TypeLine: "Creature — Werewolf"},
		},
	}
	got := in.withTokenStatus(true)
	for i, f := range got.Faces {
		if !typeLineHas(f.TypeLine, "token") {
			t.Errorf("face %d = %q, want a token face", i, f.TypeLine)
		}
	}
	if !typeLineHas(got.TypeLine, "token") {
		t.Errorf("type line = %q, want a token", got.TypeLine)
	}
	if typeLineHas(in.TypeLine, "token") || typeLineHas(in.Faces[1].TypeLine, "token") {
		t.Errorf("the input was mutated: %+v", in)
	}
	back := got.withTokenStatus(false)
	for i, f := range back.Faces {
		if typeLineHas(f.TypeLine, "token") {
			t.Errorf("face %d = %q, want a card face", i, f.TypeLine)
		}
	}
}

// applyCopy is the choke point of every entry copy: the copier's token
// status survives it both ways, and PrintedSelf (what the permanent
// reverts to) keeps the copier's own line.
func TestApplyCopyKeepsTheCopiersTokenStatus(t *testing.T) {
	bear := Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}
	gob := Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1}

	tok := Card{Name: "Clone", TypeLine: "Token Creature — Shapeshifter"}
	tok.applyCopy(CopiableValuesOf(bear), bear)
	if !tok.IsToken() || tok.Name != "Grizzly Bears" || tok.TypeLine != "Token Creature — Bear" {
		t.Errorf("token copying a card: %q %q, want a Grizzly Bears token", tok.Name, tok.TypeLine)
	}
	tok.restorePrintedSelf()
	if !tok.IsToken() || tok.Name != "Clone" {
		t.Errorf("after the copy ends: %q %q, want the Clone token", tok.Name, tok.TypeLine)
	}

	card := Card{Name: "Clone", TypeLine: "Creature — Shapeshifter"}
	card.applyCopy(CopiableValuesOf(gob), gob)
	if card.IsToken() || card.Name != "Goblin" || card.TypeLine != "Creature — Goblin" {
		t.Errorf("card copying a token: %q %q, want a Goblin card", card.Name, card.TypeLine)
	}
	card.restorePrintedSelf()
	if card.IsToken() || card.Name != "Clone" {
		t.Errorf("after the copy ends: %q %q, want the Clone card", card.Name, card.TypeLine)
	}
}
