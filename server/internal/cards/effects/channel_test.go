package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// channel_test.go — #4885: Channel and Yavimaya Bloomsage // Channel,
// through the catalog and the real cast, activation and turn paths. The
// seam is a mana ability the PLAYER has until end of turn
// (game/player_mana_ability.go).

const channelOracle = "d1b815d1-2848-40d4-a555-66822d1becbc"

// castChannel casts the Alpha Channel for the active seat and resolves
// it, returning the spell's id (the grant's source).
func castChannel(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Channel", "Sorcery", channelOracle, nil)
	passPriorityAroundTable(t, g)
	return id
}

func playerManaOffers(g *game.Game, player uuid.UUID) []game.PlayerManaAbilityOffer {
	var out []game.PlayerManaAbilityOffer
	g.ReadSnapshot(func() { out = g.PlayerManaAbilitiesLocked(player) })
	return out
}

func colorlessFloating(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.ManaPool.CountColor("C") })
	return n
}

func TestChannelGivesPayOneLifeAddColorlessRepeatedly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := castChannel(t, g)

	offers := playerManaOffers(g, me.ID)
	if len(offers) != 1 || offers[0].Source != src || offers[0].LifeCost != 1 ||
		offers[0].Produced != "{C}" || !offers[0].Payable || offers[0].Label != "Channel" {
		t.Fatalf("offers = %+v, want one payable Channel: pay 1 life, add {C}", offers)
	}
	// Only the caster has it.
	if o := playerManaOffers(g, g.Seats[1].ID); len(o) != 0 {
		t.Errorf("an opponent has %+v", o)
	}

	life := me.Life
	for i := 1; i <= 3; i++ {
		if err := g.ActivatePlayerManaAbility(me.ID, src); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		if me.Life != life-i || colorlessFloating(g, me) != i {
			t.Fatalf("after %d: life %d (want %d), {C} floating %d (want %d)",
				i, me.Life, life-i, colorlessFloating(g, me), i)
		}
	}
	// An opponent cannot use the caster's grant.
	if err := g.ActivatePlayerManaAbility(g.Seats[1].ID, src); err == nil {
		t.Error("an opponent activated the caster's Channel")
	}
}

// The mana floats into the pool and pays a cost from there (CR 601.2g):
// a strict cast short of mana is refused, the auto-tapper never spends
// the life for the player, and the cast goes through once the player
// has made the mana themselves.
func TestChannelManaPaysASpellsCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := castChannel(t, g)

	rock := game.Card{InstanceID: uuid.New(), Name: "Test Rock", TypeLine: "Artifact", ManaCost: "{2}",
		Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(rock)
	life := me.Life
	if err := g.CastSpell(me.ID, rock.InstanceID, game.CastSpellParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("a strict auto-tapped cast went through with no mana but Channel's")
	}
	if me.Life != life {
		t.Fatalf("the auto-tapper spent %d life", life-me.Life)
	}
	for range 2 {
		if err := g.ActivatePlayerManaAbility(me.ID, src); err != nil {
			t.Fatalf("activate: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, rock.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the {C}{C} Channel made did not pay {2}: %v", err)
	}
	if colorlessFloating(g, me) != 0 || me.Life != life-2 {
		t.Errorf("pool {C} %d, life %d; want 0 and %d", colorlessFloating(g, me), me.Life, life-2)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock.InstanceID) {
		t.Error("the rock did not resolve")
	}
}

// CR 119.4: you can't pay more life than you have. At 0 life the
// activation is refused with nothing paid and nothing made, the view
// greys it and the enumerator stops offering it. Paying down to exactly
// 0 is legal.
func TestChannelIsRefusedWithoutTheLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := castChannel(t, g)

	g.WithWriteLock(func() { me.Life = 1 })
	if !hasPlayerManaMove(g, me.ID, src) {
		t.Error("not offered at 1 life")
	}
	if err := g.ActivatePlayerManaAbility(me.ID, src); err != nil {
		t.Fatalf("at 1 life: %v", err)
	}
	// Paying down to exactly 0 is legal; the state-based actions then
	// end the player (CR 704.5a), which is what the payment risked.
	if me.Life != 0 || !me.Eliminated {
		t.Fatalf("at 1 life: life %d, eliminated %v; want 0 and out", me.Life, me.Eliminated)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	src2 := castChannel(t, g2)
	g2.WithWriteLock(func() { me2.Life = 0 })
	err := g2.ActivatePlayerManaAbility(me2.ID, src2)
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("at 0 life: err = %v, want ErrInvalidParam", err)
	}
	if me2.Life != 0 || colorlessFloating(g2, me2) != 0 {
		t.Errorf("at 0 life something was paid or made: life %d, {C} %d", me2.Life, colorlessFloating(g2, me2))
	}
	if o := playerManaOffers(g2, me2.ID); len(o) != 1 || o[0].Payable {
		t.Errorf("at 0 life the offer is %+v, want one unpayable", o)
	}
	if hasPlayerManaMove(g2, me2.ID, src2) {
		t.Error("the enumerator offered it at 0 life")
	}
}

// "Until end of turn": the cleanup step ends it.
func TestChannelEndsAtEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := castChannel(t, g)
	for i := 0; g.Turn.ActiveSeat == 0; i++ {
		if i > 20 {
			t.Fatal("the turn never ended")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if o := playerManaOffers(g, me.ID); len(o) != 0 {
		t.Fatalf("still offered next turn: %+v", o)
	}
	life := me.Life
	if err := g.ActivatePlayerManaAbility(me.ID, src); err == nil {
		t.Error("activated after end of turn")
	}
	if me.Life != life {
		t.Error("an expired activation paid life")
	}
}

// The grant is plain data on the player: it survives a strict restore
// and still works, and still ends at cleanup.
func TestChannelSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	src := castChannel(t, g)
	r := restoreRoundTrip(t, g, true)
	me := r.Seats[0]
	life := me.Life
	if o := playerManaOffers(r, me.ID); len(o) != 1 || o[0].Source != src {
		t.Fatalf("after restore: %+v", o)
	}
	if err := r.ActivatePlayerManaAbility(me.ID, src); err != nil {
		t.Fatalf("after restore: %v", err)
	}
	if me.Life != life-1 || colorlessFloating(r, me) != 1 {
		t.Errorf("after restore: life %d, {C} %d", me.Life, colorlessFloating(r, me))
	}
}

func hasPlayerManaMove(g *game.Game, player, src uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, player) {
		if m.Type == legal.TypeActivateManaAbility && m.Source == uuid.Nil &&
			string(m.Params) == `{"player_ability":"`+src.String()+`"}` {
			if m.Cost == nil || m.Cost.Life != 1 {
				return false
			}
			return true
		}
	}
	return false
}

// --- Yavimaya Bloomsage // Channel ---------------------------------

func yavimayaBloomsage(owner uuid.UUID) game.Card {
	return preparationCard(owner, yavimayaBloomsageOracleID, nil,
		game.Face{Name: "Yavimaya Bloomsage", TypeLine: "Creature — Dryad Druid", ManaCost: "{2}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2},
		game.Face{Name: "Channel", TypeLine: "Sorcery", ManaCost: "{G}{G}", Colors: []string{"G"}})
}

// At the beginning of your end step the target gets a +1/+1 counter; the
// Bloomsage becomes prepared only when that leaves the creature at power
// 7 or greater.
func TestYavimayaBloomsagePreparesAtPowerSeven(t *testing.T) {
	for _, tc := range []struct {
		power int
		want  bool
	}{{5, false}, {6, true}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceTo(t, g, game.StepPrecombatMain)
		sage := yavimayaBloomsage(me.ID)
		castFromHandAndResolve(t, g, me, sage)
		if g.IsPreparedForEffect(sage.InstanceID) {
			t.Fatal("Yavimaya Bloomsage entered prepared")
		}
		big := pushVanillaCreature(g, me.ID, "Wurm", tc.power, tc.power)
		advanceToEndStep(t, g)
		pickCard(t, g, me.ID, big)
		passPriorityAroundTable(t, g)
		if n := fraHasCounters(t, g, big); n != 1 {
			t.Errorf("power %d: %d counters on the target, want 1", tc.power, n)
		}
		if got := g.IsPreparedForEffect(sage.InstanceID); got != tc.want {
			t.Errorf("power %d+1: prepared = %v, want %v", tc.power, got, tc.want)
		}
	}
}

// The prepare spell is Channel: casting the copy unprepares the
// Bloomsage and gives its caster pay 1 life: add {C}.
func TestYavimayaBloomsageCastsTheChannelCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	sage := yavimayaBloomsage(me.ID)
	castFromHandAndResolve(t, g, me, sage)
	if _, err := g.BecomePreparedForEffect(sage.InstanceID); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	cp, ok := prepareCopyOf(g, "Channel")
	if !ok {
		t.Fatal("no copy of Channel in exile")
	}
	if err := fraCastCopy(t, g, me, cp); err != nil {
		t.Fatalf("cast the Channel copy: %v", err)
	}
	if g.IsPreparedForEffect(sage.InstanceID) {
		t.Error("casting the copy did not unprepare the Bloomsage")
	}
	passPriorityAroundTable(t, g)
	offers := playerManaOffers(g, me.ID)
	if len(offers) != 1 || offers[0].LifeCost != 1 || offers[0].Produced != "{C}" {
		t.Fatalf("offers = %+v, want Channel's", offers)
	}
	life := me.Life
	if err := g.ActivatePlayerManaAbility(me.ID, offers[0].Source); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != life-1 || colorlessFloating(g, me) != 1 {
		t.Errorf("life %d, {C} %d; want %d and 1", me.Life, colorlessFloating(g, me), life-1)
	}
}
