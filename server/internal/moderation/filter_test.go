package moderation

import (
	"strings"
	"testing"
)

// The test words are real entries of denylist.txt: a test that checked a
// word the list does not hold would pass for the wrong reason. "retarded"
// and "nigger" stand for the substring entries, "kys" and "slut" for the
// whole-word ones, "kill yourself" for a phrase.

func TestNameRefusesListedWordsAndTheirDisguises(t *testing.T) {
	refused := []string{
		"retarded",
		"RETARDED",
		"xxretardedxx",    // a substring entry matches inside a word
		"r3t4rd3d",        // leetspeak
		"rétårdéd",        // diacritics
		"reeetaaarded",    // stretched letters
		"r e t a r d e d", // spaced out
		"r.e.t.a.r.d.e.d",
		"re\u200btarded", // zero-width space
		"ｒｅｔａｒｄｅｄ",       // fullwidth
		"rеtarded",       // Cyrillic е
		"slut",
		"Sl_u_t",
		"SLUUUT",
		"kys",
		"k y s",
		"Kill Yourself",
		"kill   yourself!!",
		"N1gg3r",
	}
	for _, in := range refused {
		if err := Name(in); err == nil {
			t.Errorf("Name(%q) allowed, want refused", in)
		}
	}
}

func TestNameAllowsOrdinaryNamesAndWordsThatContainListedOnes(t *testing.T) {
	allowed := []string{
		"Alice",
		"Bob the Builder",
		"Sol Ring",
		"Wrath of God",
		"Cultivate",
		"classic",      // contains "ass"-like letters; "ass" is not listed
		"Scunthorpe",   // the classic false positive
		"raccoon",      // contains coon: a whole-word entry
		"coonskin cap", // ditto
		"Pakistan",     // contains paki
		"slutsky",      // a whole-word entry only matches the word
		"kyoto",        // contains ky
		"ni gh t",      // spaced pieces that are not a listed word
		"bone rope",    // two real words must not run together
		"wow!",         // a trailing ! is punctuation
		"C++ fan",      // a trailing + is punctuation
		"Player 2",
		"Dr4g0n",
		"",
		"   ",
		"日本語の名前",
	}
	for _, in := range allowed {
		if err := Name(in); err != nil {
			t.Errorf("Name(%q) refused, want allowed", in)
		}
	}
}

func TestChatMasksTheWordAndKeepsTheRest(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"you are retarded lol", "you are ******** lol"},
		{"RETARDED", "********"},
		{"r.e.t.a.r.d.e.d now", "*************** now"},
		{"nothing to see here", "nothing to see here"},
		{"go kys and a slut", "go *** and a ****"},
		{"cast Wrath of God", "cast Wrath of God"},
		{"rétardéd", "********"},
		{"kill yourself", "**** ********"},
	} {
		got, changed := Chat(tc.in)
		if got != tc.want {
			t.Errorf("Chat(%q) = %q, want %q", tc.in, got, tc.want)
		}
		if changed != (tc.in != tc.want) {
			t.Errorf("Chat(%q) changed = %v", tc.in, changed)
		}
	}
}

func TestChatMasksEveryOccurrence(t *testing.T) {
	got, _ := Chat("slut slut and slut")
	if strings.Contains(got, "slut") || strings.Count(got, "****") != 3 {
		t.Errorf("got %q", got)
	}
}

func TestKeyFoldsSpellingsOfOneName(t *testing.T) {
	want := Key("Bob")
	for _, in := range []string{"bob", "BOB", "b0b", "B O B", " bob ", "Bööb", "Bobbb", "b.o.b"} {
		if got := Key(in); got != want {
			t.Errorf("Key(%q) = %q, want %q", in, got, want)
		}
	}
	if Key("Bob") == Key("Rob") {
		t.Error("different names share a key")
	}
	if got := Key("  日本  "); got != "日本" {
		t.Errorf("Key of a name with no latin letters = %q", got)
	}
}

func TestDenylistLinesAreWellFormed(t *testing.T) {
	if len(entries) < 20 {
		t.Fatalf("denylist has %d entries; did the embed break?", len(entries))
	}
	for _, e := range entries {
		if len(e.flat) == 0 {
			t.Error("an entry has no letters")
		}
		if !e.whole && len(e.flat) > 0 {
			n := 0
			for _, r := range e.flat {
				n += r.n
			}
			if n < 5 {
				t.Errorf("substring entry of %d letters: a short word must be a whole-word entry (=word)", n)
			}
		}
	}
}
