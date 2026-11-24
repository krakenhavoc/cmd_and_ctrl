package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const tashaOracle = "863a148c-596a-492b-b68e-3e1cbfcb140f"

func TestTashaTheWitchQueenIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(tashaOracle)
	if !ok || spec.Name != "Tasha, the Witch Queen" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || spec.StartingLoyalty != 4 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// tashaTable puts Tasha (loyalty 4) on the active player's battlefield in
// their main phase.
func tashaTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ToMain(t, g)
	tasha := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Tasha, the Witch Queen", "Legendary Planeswalker — Tasha", tashaOracle, 4)
	return g, me, opp, tasha
}

// tashaExiled puts a sorcery card `owner` owns into exile, with a page
// counter when `page` is set, and returns it.
func tashaExiled(g *game.Game, owner *game.Player, name string, page bool) uuid.UUID {
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.ManaCost = "Sorcery", "{5}{U}"
	if page {
		c.Counters = map[string]int{"page": 1}
	}
	g.WithWriteLock(func() { g.Exile.PushTop(c) })
	return c.InstanceID
}

// Casting a spell you don't own makes a 3/3 black Demon; casting your own
// does not.
func TestTashaMakesADemonWhenYouCastASpellYouDontOwn(t *testing.T) {
	g, me, opp, _ := tashaTable(t)
	theirs := tashaExiled(g, opp, "Their Sorcery", false)
	mine := tashaExiled(g, me, "My Sorcery", false)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{theirs, mine} {
			c, _ := g.LookupCardForEffect(id)
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Cost: "{0}", CastOnly: true}, []game.Card{c})
		}
	})
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast my own card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Demon"); n != 0 {
		t.Fatalf("%d Demons after casting my own card, want 0", n)
	}
	if err := g.CastSpell(me.ID, theirs, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast their card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g) // the trigger
	effectstest.PassPriorityAroundTable(t, g) // the spell
	if n := effectstest.OnBattlefieldNamed(g, "Demon"); n != 1 {
		t.Fatalf("%d Demons after casting their card, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Demon" && (c.Power != 3 || c.Toughness != 3 || len(c.Colors) != 1 || c.Colors[0] != "B") {
			t.Errorf("Demon token %d/%d %v, want a 3/3 black", c.Power, c.Toughness, c.Colors)
		}
	}
}

// +1 draws a card and exiles the targeted instant from an opponent's
// graveyard with a page counter; a creature card is not a legal target.
func TestTashaPlusOneDrawsAndExilesWithAPageCounter(t *testing.T) {
	g, me, opp, tasha := tashaTable(t)
	instant := effectstest.GraveyardCardOf(g, opp, "Their Instant", "Instant", "{U}", "")
	creature := effectstest.GraveyardCardOf(g, opp, "Their Bear", "Creature — Bear", "{1}{G}", "")
	mine := effectstest.GraveyardCardOf(g, me, "My Instant", "Instant", "{U}", "")
	for _, bad := range []uuid.UUID{creature, mine} {
		if err := g.ActivateCatalogAbility(me.ID, tasha, 0, game.ActivateAbilityParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad}}}); err == nil {
			t.Fatalf("targeted %v", bad)
		}
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, tasha, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: instant}}}); err != nil {
		t.Fatalf("+1 on the opponent's instant: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if got := effectstest.CounterCountInExile(g, instant, "page"); got != 1 {
		t.Errorf("the exiled instant has %d page counters, want 1", got)
	}
}

// −3 casts one spell from among the cards in exile with page counters,
// free; a card without one is not offered, and a second is not cast.
func TestTashaMinusThreeCastsOnePageCardFree(t *testing.T) {
	g, me, opp, tasha := tashaTable(t)
	first := tashaExiled(g, opp, "Paged One", true)
	second := tashaExiled(g, opp, "Paged Two", true)
	plain := tashaExiled(g, opp, "Unpaged", false)
	if err := g.ActivateCatalogAbility(me.ID, tasha, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	for i := 0; i < 8 && effectstest.ResolutionCastOfferFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) == nil {
		t.Fatal("no free cast was offered")
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, plain, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast a card with no page counter")
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, first, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast a page card free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.Cmdr2S41InExile(g, first) != nil {
		t.Error("the cast card is still in exile")
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("a second page card was cast")
	}
}
