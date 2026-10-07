package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// spend_grant_convoke_test.go — #1928 on a real catalog card. The
// Wandering Rescuer ({3}{W}{W}, convoke) exiled under an any-colour
// cast permission: the grant widens MANA (CR 609.4b), and a tapped
// creature is not mana (CR 702.51a), so convoke still follows the
// colour rule. The view, the bot enumerator and the payment must say
// the same thing about it.

func stolenRescuer(t *testing.T, perm game.CastPermission) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	me.Hand.Cards = nil
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "The Wandering Rescuer",
		TypeLine: "Legendary Creature — Human Samurai Noble", ManaCost: wanderingRescuerCost,
		OracleID: theWanderingRescuerOracle, Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	perm.Player = me.ID
	perm.Duration = game.WhileInZoneDuration()
	g.WithWriteLock(func() {
		if err := g.ExileCardWithPermissionForEffect(id, perm); err != nil {
			t.Fatal(err)
		}
	})
	return g, me, id
}

// viewCastCost is the first cast_prices entry the viewer sees on an
// exiled card.
func viewCastCost(t *testing.T, g *game.Game, viewer, card uuid.UUID) string {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, c := range v.Exile.Cards {
		if c.InstanceID == card.String() {
			if len(c.CastPrices) == 0 {
				t.Fatal("the exiled card carries no cast price")
			}
			return c.CastPrices[0].Cost
		}
	}
	t.Fatalf("card %s is not in the exile view", card)
	return ""
}

func fundOffColour(me *game.Player, n int) {
	for i := 0; i < n; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	}
}

func TestStolenConvokeSpellViewBotAndPaymentAgree(t *testing.T) {
	for _, tc := range []struct {
		name string
		perm game.CastPermission
	}{
		{"any color", game.CastPermission{AnyColor: true}},
		{"any type", game.CastPermission{AnyType: true}},
	} {
		t.Run(tc.name+"/five off-colour mana", func(t *testing.T) {
			g, me, id := stolenRescuer(t, tc.perm)
			fundOffColour(me, 5)
			if !enumeratedFor(g, me.ID, id) {
				t.Error("the bot enumerator does not offer a cast the pool can pay")
			}
			if !viewCastable(t, g, me.ID, id) {
				t.Error("the view does not mark castable a cast the pool can pay")
			}
			if got := viewCastCost(t, g, me.ID, id); got != wanderingRescuerCost {
				t.Errorf("the view's cast price = %s, want the printed %s (CR 609.4b)", got, wanderingRescuerCost)
			}
			if err := castStash(g, me, id); err != nil {
				t.Fatalf("payment: %v", err)
			}
		})
		t.Run(tc.name+"/four off-colour mana", func(t *testing.T) {
			g, me, id := stolenRescuer(t, tc.perm)
			fundOffColour(me, 4)
			if enumeratedFor(g, me.ID, id) {
				t.Error("the bot enumerator offers a cast four mana cannot pay")
			}
			if err := castStash(g, me, id); err == nil {
				t.Fatal("payment accepted four mana for a five-mana cost")
			}
		})
		t.Run(tc.name+"/off-colour creatures", func(t *testing.T) {
			g, me, id := stolenRescuer(t, tc.perm)
			var taps []uuid.UUID
			for i := 0; i < 5; i++ {
				taps = append(taps, pushTapCostPermanent(g, me.ID, "Bear", "Creature — Bear", "G"))
			}
			err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true, TapIDs: taps})
			if err == nil {
				t.Fatal("five green creatures convoked a {3}{W}{W} under an any-colour grant")
			}
		})
		t.Run(tc.name+"/on-colour creatures and mana", func(t *testing.T) {
			g, me, id := stolenRescuer(t, tc.perm)
			taps := []uuid.UUID{
				pushTapCostPermanent(g, me.ID, "Soldier", "Creature — Soldier", "W"),
				pushTapCostPermanent(g, me.ID, "Soldier", "Creature — Soldier", "W"),
				pushTapCostPermanent(g, me.ID, "Bear", "Creature — Bear", "G"),
			}
			fundOffColour(me, 2)
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true, TapIDs: taps}); err != nil {
				t.Fatalf("two white creatures, one green and two mana: %v", err)
			}
		})
	}
}
