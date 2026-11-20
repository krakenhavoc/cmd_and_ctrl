package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// banding_view_test.go — ADR 0224 (#5723): the wire's attack_band. The
// table sees which attackers are banded together, numbered rather than
// keyed, and an attacker in no band carries nothing.
func TestAttackBandNumbersTheBandedAttackers(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	push := func(name string, keywords ...string) uuid.UUID {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Creature — Test",
			Power: 1, Toughness: 1, Owner: me, Controller: me, Keywords: keywords,
		})
		return id
	}
	hero := push("Benalish Hero", game.KeywordBanding)
	bear := push("Grizzly Bears")
	loner := push("Hill Giant")
	for i := 0; i < 20 && g.Turn.Step != game.StepDeclareAttackers; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := g.DeclareAttackersWith([]game.AttackDeclaration{
		{Attacker: hero, Target: opp, Band: 1},
		{Attacker: bear, Target: opp, Band: 1},
		{Attacker: loner, Target: opp},
	}, game.DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare the band: %v", err)
	}
	v := ViewOfGame(g)
	h, b, l := cardInView(t, v, hero), cardInView(t, v, bear), cardInView(t, v, loner)
	if h.AttackBand != 1 || b.AttackBand != 1 {
		t.Errorf("attack_band = %d and %d, want both 1", h.AttackBand, b.AttackBand)
	}
	if l.AttackBand != 0 {
		t.Errorf("an attacker in no band has attack_band %d", l.AttackBand)
	}
}
