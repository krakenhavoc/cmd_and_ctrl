package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// any_color_spend_view_test.go — #1600: the client greys a hand card
// off the view's legal_actions digest, so under Chromatic Orrery the
// view must call a {U}{U} spell castable off two Mountains exactly when
// the engine's payment accepts it. Both answers are asked of one board,
// with and without the Orrery.
func TestAnyColorSpendViewAndPaymentAgree(t *testing.T) {
	const oracleChromaticOrrery = "95c3976c-33f3-490b-bfd3-7f1af2fe0416"
	for _, withOrrery := range []bool{true, false} {
		g := buildMainPhaseGame(t)
		seat := g.Seats[g.Turn.ActiveSeat]
		seat.Hand.Cards = nil
		push := func(c game.Card) uuid.UUID {
			c.InstanceID, c.Owner, c.Controller = uuid.New(), seat.ID, seat.ID
			g.Battlefield.PushTop(c)
			return c.InstanceID
		}
		push(game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
		push(game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
		if withOrrery {
			push(game.Card{Name: "Chromatic Orrery", TypeLine: "Legendary Artifact", OracleID: oracleChromaticOrrery})
		}
		spell := game.NewCard("Blue Spell", seat.ID)
		spell.TypeLine, spell.ManaCost = "Sorcery", "{U}{U}"
		spell.KnownBy = map[uuid.UUID]bool{seat.ID: true}
		seat.Hand.PushTop(spell)

		v := ViewOfGameFor(g, seat.ID.String())
		castable := false
		if v.LegalActions != nil {
			if src := v.LegalActions.Sources[spell.InstanceID.String()]; src != nil {
				castable = slices.Contains(src.Kinds, legal.KindCast)
			}
		}
		err := g.CastSpell(seat.ID, spell.InstanceID, game.CastSpellParams{Strict: true, AutoTap: true})
		if castable != withOrrery || (err == nil) != withOrrery {
			t.Errorf("Orrery %v: the view says castable %v, the payment says %v", withOrrery, castable, err)
		}
	}
}
