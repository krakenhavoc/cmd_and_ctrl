package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

const liesaOracle = "2270daaf-252d-40c4-bed5-bd46ae15f041"

// liesaInCommand puts Liesa in the active seat's command zone, cast
// `casts` times before, at a main phase.
func liesaInCommand(t *testing.T, casts int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	knownBy := map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		knownBy[s.ID] = true
	}
	me.Command.PushTop(game.Card{
		InstanceID: id, Name: "Liesa, Shroud of Dusk", TypeLine: "Legendary Creature — Angel",
		ManaCost: "{2}{W}{W}{B}", OracleID: liesaOracle, Power: 5, Toughness: 5,
		Owner: me.ID, Controller: me.ID, IsCommander: true, KnownBy: knownBy,
	})
	me.CommanderCasts[id] = casts
	effectstest.AdvanceToMain(t, g)
	return g, me, id
}

// CR 903.8, replaced: Liesa's third cast from the command zone costs
// {2}{W}{W}{B} and 4 life, not {6}{W}{W}{B}.
func TestLiesaPaysTheCommanderTaxInLife(t *testing.T) {
	g, me, id := liesaInCommand(t, 2)
	price, err := g.PriceCast(me.ID, game.Card{InstanceID: id, Name: "Liesa, Shroud of Dusk", ManaCost: "{2}{W}{W}{B}", OracleID: liesaOracle, Owner: me.ID, Controller: me.ID, IsCommander: true},
		game.CastSpellParams{FromZone: "command"})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if got := price.Total.String(); got != "{2}{W}{W}{B}" {
		t.Errorf("price %s, want {2}{W}{W}{B} (no mana tax)", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{W}{W}{B}")
	life := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "command", Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if got := life - me.Life; got != 4 {
		t.Errorf("paid %d life, want 4 (2 for each of two previous casts)", got)
	}
	if me.CommanderCasts[id] != 3 {
		t.Errorf("CommanderCasts = %d, want 3", me.CommanderCasts[id])
	}
}

// CR 119.4: with 3 life the 4-life tax cannot be paid, so the cast is
// refused and nothing moves.
func TestLiesaCannotBeCastWithoutTheLifeForTheTax(t *testing.T) {
	g, me, id := liesaInCommand(t, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}{W}{W}{B}")
	g.WithWriteLock(func() { me.Life = 3 })
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "command", Strict: true}); err == nil {
		t.Fatal("the cast was accepted with 3 life for a 4-life tax")
	}
	if !me.Command.Contains(id) || me.Life != 3 {
		t.Errorf("a refused cast moved the card or paid life (life %d)", me.Life)
	}
}

// The first cast pays no tax at all, and the view's command-zone price
// carries the life once there is a tax.
func TestLiesaFirstCastIsUntaxedAndTheViewShowsTheLife(t *testing.T) {
	g, me, id := liesaInCommand(t, 0)
	effectstest.FloatMana(t, g, me, "{C}{C}{W}{W}{B}")
	life := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "command", Strict: true}); err != nil {
		t.Fatalf("first cast: %v", err)
	}
	if me.Life != life {
		t.Errorf("first cast paid %d life", life-me.Life)
	}

	g2, me2, id2 := liesaInCommand(t, 1)
	v := protocol.ViewOfGameFor(g2, me2.ID.String())
	for _, s := range v.Seats {
		for _, c := range s.Command.Cards {
			if c.InstanceID != id2.String() {
				continue
			}
			if len(c.CastPrices) == 0 {
				t.Fatal("no cast price on the command-zone card")
			}
			p := c.CastPrices[0]
			if p.Cost != "{2}{W}{W}{B}" || p.Life != 2 || p.CommanderTaxLife != 2 {
				t.Errorf("price %+v, want {2}{W}{W}{B} and 2 life of commander tax", p)
			}
			return
		}
	}
	t.Fatal("Liesa not in the viewer's command zone")
}

// "Whenever a player casts a spell, they lose 2 life" — any player.
func TestLiesaMakesEachCasterLoseTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Liesa, Shroud of Dusk", "Legendary Creature — Angel", liesaOracle, false)
	spell := effectstest.HandSpell(me, "Opt Shape", "Instant", "{0}")
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - me.Life; got != 2 {
		t.Errorf("caster lost %d life, want 2", got)
	}
}
