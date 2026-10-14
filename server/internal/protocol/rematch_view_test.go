package protocol

import "testing"

func TestFilterRematchFor(t *testing.T) {
	full := &RematchView{
		GameID: "new", Host: "h", HostName: "Hana", Open: true,
		Accepted: []string{"a"}, Left: []string{"l"},
	}
	cases := []struct {
		viewer string
		want   string
	}{
		{"h", RematchHost},
		{"a", RematchAccepted},
		{"l", RematchDeclined},
		{"u", RematchUndecided},
	}
	for _, c := range cases {
		got := filterRematchFor(full, c.viewer)
		if got == nil || got.You != c.want {
			t.Errorf("viewer %q: %+v, want you=%q", c.viewer, got, c.want)
			continue
		}
		if got.Accepted != nil || got.Left != nil {
			t.Errorf("viewer %q sees other seats' answers: %+v", c.viewer, got)
		}
		if got.GameID != "new" || got.HostName != "Hana" || !got.Open {
			t.Errorf("viewer %q lost the shared fields: %+v", c.viewer, got)
		}
	}
	if got := filterRematchFor(full, SpectatorViewerID); got != nil {
		t.Errorf("a spectator sees %+v", got)
	}
	if got := filterRematchFor(nil, "h"); got != nil {
		t.Errorf("no offer, but %+v", got)
	}
	admin := filterRematchFor(full, "")
	if admin == nil || len(admin.Accepted) != 1 || len(admin.Left) != 1 || admin.You != "" {
		t.Errorf("the admin's view = %+v, want the whole offer", admin)
	}
	// The filter copies: editing a viewer's view leaves the stamp alone.
	admin.Accepted[0] = "x"
	if full.Accepted[0] != "a" {
		t.Error("the filtered view aliases the stamp")
	}
}
