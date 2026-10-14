package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// closes_on_pass_view_test.go — #3490. A pass-closed free cast (rebound,
// suspend, discover, cascade) ends when its holder passes priority, so
// the client has to know which exile grants those are: auto-pass would
// otherwise pass for the holder 80 ms after "Cast it free", before the
// target picker is answered, and the card would stay in exile.

func TestExileGrantShipsClosesOnPass(t *testing.T) {
	for _, tc := range []struct {
		name  string
		grant game.CastPermission
		want  bool
	}{
		{"rebound's free cast", game.CastPermission{
			Cost: "{0}", Timing: game.TimingFlash, CastOnly: true, LapseOnPass: game.LapseStaysInExile,
		}, true},
		{"an impulse grant that outlasts a pass", game.CastPermission{CastOnly: true}, false},
		{"a free cast that does not close on a pass", game.CastPermission{Cost: "{0}", CastOnly: true}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := buildActiveGame(t)
			me := g.Seats[0]
			c := game.NewCard("Test Misalignment", me.ID)
			c.TypeLine = "Sorcery"
			c.ManaCost = "{4}{U}"
			c.KnownBy = map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}
			grant := tc.grant
			grant.Player = me.ID
			g.Exile.PushTop(c)
			g.GrantCastPermissionOverCardForEffect(c.InstanceID, grant)

			v := ViewOfGameFor(g, me.ID.String())
			var card *CardView
			for i := range v.Exile.Cards {
				if v.Exile.Cards[i].InstanceID == c.InstanceID.String() {
					card = &v.Exile.Cards[i]
				}
			}
			if card == nil || card.ExilePlay == nil {
				t.Fatalf("exile view missing the grant")
			}
			if card.ExilePlay.ClosesOnPass != tc.want {
				t.Errorf("closes_on_pass = %v, want %v", card.ExilePlay.ClosesOnPass, tc.want)
			}
		})
	}
}
