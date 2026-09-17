package model

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestPermanentLabelExplainsMissedUntap(t *testing.T) {
	for _, tc := range []struct {
		name    string
		tapped  bool
		noUntap *protocol.NoUntapView
		want    bool
	}{
		{"static", true, &protocol.NoUntapView{Static: true}, true},
		{"controller marker", true, &protocol.NoUntapView{Next: []string{"me"}}, true},
		{"other player marker", true, &protocol.NoUntapView{Next: []string{"other"}}, false},
		{"untapped", false, &protocol.NoUntapView{Static: true}, false},
		{"ordinary", true, nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := protocol.CardView{Name: "Mana source", KnownByYou: true, Controller: "me", Tapped: tc.tapped, NoUntap: tc.noUntap}
			if label := permanentLabel(&c); strings.Contains(label, "won't untap") != tc.want {
				t.Fatalf("label = %q, want won't-untap flag %v", label, tc.want)
			}
		})
	}
}
