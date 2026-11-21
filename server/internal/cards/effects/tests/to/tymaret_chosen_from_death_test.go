package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const tymaretOracle = "64b4e91f-83a4-4b3f-87d4-8c48907e628d"

func tymaretOnBattlefield(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tymaret, Chosen from Death", TypeLine: "Legendary Enchantment Creature — Demigod",
		OracleID: tymaretOracle, ManaCost: "{B}{B}", Colors: []string{"B"}, Power: 2,
		Owner: owner, Controller: owner,
	})
}

func TestTymaretToughnessIsYourDevotionToBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ty := tymaretOnBattlefield(g, me.ID)
	if got := effectstest.EffectiveToughness(t, g, ty); got != 2 {
		t.Errorf("alone Tymaret has toughness %d, want 2 (her own {B}{B})", got)
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Black Thing", TypeLine: "Creature — Zombie",
		ManaCost: "{2}{B}{B}", Colors: []string{"B"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Black Thing", TypeLine: "Creature — Zombie",
		ManaCost: "{B}{B}{B}", Colors: []string{"B"}, Power: 1, Toughness: 1, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	if got := effectstest.EffectiveToughness(t, g, ty); got != 4 {
		t.Errorf("toughness %d, want 4 (an opponent's pips do not count)", got)
	}
}

func tymaretActivate(t *testing.T, g *game.Game, me *game.Player, ty uuid.UUID, targets ...uuid.UUID) error {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	return g.ActivateCatalogAbility(me.ID, ty, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(targets...), Strict: true})
}

func TestTymaretExilesTwoCardsAndGainsLifePerCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ty := tymaretOnBattlefield(g, me.ID)
	bear := effectstest.PushGraveyardPermanent(opp, "Bear", "Creature — Bear", "{1}")
	shock := effectstest.PushGraveyardPermanent(me, "Shock", "Instant", "{R}")
	left := effectstest.PushGraveyardPermanent(opp, "Left Behind", "Creature — Cat", "{1}")
	if err := tymaretActivate(t, g, me, ty, bear, shock); err == nil {
		t.Fatal("activated with an empty mana pool")
	}
	for _, c := range []string{"B", "B"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	life := me.Life
	if err := tymaretActivate(t, g, me, ty, bear, shock); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) || !g.Exile.Contains(shock) {
		t.Fatal("the two targets were not exiled")
	}
	if !opp.Graveyard.Contains(left) {
		t.Error("an untargeted card left its graveyard")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d (one creature card exiled)", me.Life, life+1)
	}
}

func TestTymaretWithNoTargetGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ty := tymaretOnBattlefield(g, me.ID)
	for _, c := range []string{"B", "B"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	life := me.Life
	if err := tymaretActivate(t, g, me, ty); err != nil {
		t.Fatalf("activate with no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}
