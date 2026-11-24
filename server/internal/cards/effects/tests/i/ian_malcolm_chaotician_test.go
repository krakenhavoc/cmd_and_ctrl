package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ianMalcolmOracle = "57ff1665-0a7b-4d98-bd8f-cc07cf04f701"

func TestIanMalcolmChaoticianIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(ianMalcolmOracle)
	if !ok || spec.Name != "Ian Malcolm, Chaotician" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// ianDrawTo makes `p` draw until their turn tally reaches `n` and, when
// that crosses the second draw, resolves Ian's trigger.
func ianDrawTo(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	var drawn int
	g.WithWriteLock(func() { drawn = g.TurnTallyFor(p.ID).CardsDrawn })
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(p.ID, n-drawn); err != nil {
			t.Fatalf("draw: %v", err)
		}
	})
	if drawn < 2 && n >= 2 { // the second draw triggers Ian
		effectstest.SettleTriggers(t, g)
	}
}

// ianRef is Ian Malcolm's object reference as it is on the battlefield.
func ianRef(g *game.Game, ian uuid.UUID) game.ObjectRef {
	c, _ := effectstest.CardOnBattlefield(g, ian)
	return game.ObjectRef{ID: ian, Epoch: c.ObjectEpoch}
}

func ianTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	for _, p := range g.Seats {
		for i := 0; i < 6; i++ {
			p.Library.PushBottom(game.NewCard("Filler", p.ID))
		}
	}
	ian := effectstest.B12Push(g, me.ID, "Ian Malcolm, Chaotician", "Legendary Creature — Human Scientist", ianMalcolmOracle, 2, 2)
	return g, me, opp, ian
}

// A player's second draw of the turn exiles the top card of their library,
// linked to Ian; their first and third draws do not.
func TestIanMalcolmSecondDrawExilesThatPlayersTopCard(t *testing.T) {
	g, _, opp, ian := ianTable(t)
	ianDrawTo(t, g, opp, 1)
	if n := len(g.ExiledWithForEffect(game.PermissionCardRef(ianRef(g, ian)))); n != 0 {
		t.Fatalf("%d cards exiled after one draw", n)
	}
	top := effectstest.TopOfLibrary(opp, "Their Next Top", "Sorcery", "{U}")
	_ = effectstest.TopOfLibrary(opp, "Their Drawn Card", "Sorcery", "{U}")
	ianDrawTo(t, g, opp, 2)
	if got := effectstest.CardInExile(t, g, top).ExiledWith.ID; got != ian {
		t.Fatalf("their top card is linked to %s, want Ian Malcolm", got)
	}
	ianDrawTo(t, g, opp, 3)
	var linked int
	g.WithWriteLock(func() {
		for _, c := range g.Exile.Cards {
			if c.ExiledWith.ID == ian {
				linked++
			}
		}
	})
	if linked != 1 {
		t.Errorf("%d cards linked after a third draw, want 1", linked)
	}
}

// On their turn, a player casts one spell they don't own from among the
// cards exiled with Ian, paying with mana of any type; a card they own is
// refused, and so is a second spell that turn. Another player cannot cast
// during this player's turn.
func TestIanMalcolmEachPlayerCastsOneSpellTheyDontOwnOnTheirTurn(t *testing.T) {
	g, me, opp, ian := ianTable(t)
	ianDrawTo(t, g, opp, 1)
	theirs := effectstest.TopOfLibrary(opp, "Their Sorcery", "Sorcery", "{U}")
	_ = effectstest.TopOfLibrary(opp, "Their Draw", "Sorcery", "{U}")
	ianDrawTo(t, g, opp, 2)
	theirs2 := effectstest.TopOfLibrary(opp, "Their Second Sorcery", "Sorcery", "{U}")
	g.WithWriteLock(func() {
		src := ianRef(g, ian)
		if err := g.ExileLinkedForEffect(src, theirs2); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	mine := effectstest.TopOfLibrary(me, "My Sorcery", "Sorcery", "{U}")
	g.WithWriteLock(func() {
		src := ianRef(g, ian)
		if err := g.ExileLinkedForEffect(src, mine); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("cast a card its caster owns")
	}
	effectstest.FloatMana(t, g, opp, "{U}")
	if err := g.CastSpell(opp.ID, mine, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("a player cast from among the cards off their own turn")
	}
	if !effectstest.EnumeratedFor(g, me.ID, theirs) {
		t.Fatal("the bot is not offered the opponent's exiled sorcery")
	}
	if err := g.CastSpell(me.ID, theirs, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the opponent's card with colorless mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, theirs2, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("a second spell from among the cards this turn")
	}
}
