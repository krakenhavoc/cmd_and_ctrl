package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// damping_engine_test.go — #1986, CR 116.2d: Damping Engine's restriction
// (a cast gate and a land-play gate over one rule) and its escape (a
// special action that sacrifices a permanent of the player's choice and
// writes a per-player, per-source record both gates read).

const dampingEngineOracle = "3584a055-5220-4721-89f8-de33f779122f"

// deTable is a four-seat game with a Damping Engine under seat 0, the
// player whose turn it is. Seat 0 controls the Engine and nothing else,
// the others nothing at all, so seat 0 controls more permanents than each
// other player and is the restricted one.
type deTable struct {
	g      *game.Game
	me     *game.Player
	opp    *game.Player
	engine uuid.UUID
}

func newDampingEngineTable(t *testing.T) *deTable {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	engine := pushCatalogPermanent(g, me.ID, "Damping Engine", "Artifact", dampingEngineOracle, false)
	return &deTable{g: g, me: me, opp: opp, engine: engine}
}

// castGate asks the cast gate about a spell of the given type line.
func (tb *deTable) castGate(p *game.Player, typeLine string) error {
	var err error
	tb.g.WithWriteLock(func() {
		err = tb.g.CastGateLocked(p.ID, game.Card{Name: "X", TypeLine: typeLine}, game.ZoneHand, game.CastSpellParams{})
	})
	return err
}

// landGate asks the land-play gate.
func (tb *deTable) landGate(p *game.Player) error {
	return lpGate(tb.g, p, "Forest", game.ZoneHand)
}

// bear puts a creature under p's control and returns its id.
func (tb *deTable) bear(p *game.Player) uuid.UUID {
	return seedCreature(tb.g, "Bear", p.ID)
}

// ignore takes the CR 116.2d special action for `p` against `source`,
// paying with `sacrifice`.
func (tb *deTable) ignore(p *game.Player, source, sacrifice uuid.UUID) error {
	return tb.g.PerformSpecialAction(p.ID, source, game.SpecialActionIgnoreEffect,
		game.SpecialActionParams{SacrificeID: sacrifice})
}

// offers is the player's CR 116.2d offers right now.
func (tb *deTable) offers(p *game.Player) []game.IgnoreEffectOffer {
	var out []game.IgnoreEffectOffer
	tb.g.WithWriteLock(func() { out = tb.g.IgnoreEffectOffersLocked(p.ID) })
	return out
}

func (tb *deTable) inGraveyard(p *game.Player, id uuid.UUID) bool {
	return p.Graveyard != nil && p.Graveyard.Contains(id)
}

func assertDECantCast(t *testing.T, err error, what string) {
	t.Helper()
	var cant *game.CantCastError
	if !errors.As(err, &cant) || !strings.Contains(cant.Reason, "controls more permanents than each other player") {
		t.Errorf("%s: err = %v, want Damping Engine's clause", what, err)
	}
}

// The restriction: a player with strictly more permanents than every other
// player can't cast artifact, creature or enchantment spells or play
// lands. Other spells are free, a tie restricts nobody, and the Engine's
// own controller is bound like anyone.
func TestDampingEngineRestrictsThePlayerWithTheMostPermanents(t *testing.T) {
	tb := newDampingEngineTable(t)
	for _, typeLine := range []string{"Artifact", "Creature — Bear", "Enchantment", "Artifact Creature — Golem", "Enchantment — Aura"} {
		assertDECantCast(t, tb.castGate(tb.me, typeLine), typeLine)
	}
	for _, typeLine := range []string{"Instant", "Sorcery", "Planeswalker — Jace", "Battle — Siege"} {
		if err := tb.castGate(tb.me, typeLine); err != nil {
			t.Errorf("a %s spell is refused: %v", typeLine, err)
		}
	}
	err := tb.landGate(tb.me)
	if err == nil || !strings.Contains(err.Error(), "controls more permanents than each other player can't play lands") ||
		!strings.Contains(err.Error(), "Damping Engine") {
		t.Errorf("land play: err = %v, want Damping Engine's clause naming the card", err)
	}
	// An opponent with nothing is not the player with the most.
	if tb.castGate(tb.opp, "Creature — Bear") != nil || tb.landGate(tb.opp) != nil {
		t.Error("a player with fewer permanents is restricted")
	}

	// A tie restricts nobody, even the Engine's controller.
	tb.bear(tb.opp)
	if tb.castGate(tb.me, "Creature — Bear") != nil || tb.landGate(tb.me) != nil {
		t.Error("a tie for the most permanents restricts")
	}
	// Strictly more than EACH other player: a third player ahead of the
	// second keeps the second unrestricted, and the one in front is bound.
	tb.bear(tb.opp)
	assertDECantCast(t, tb.castGate(tb.opp, "Creature — Bear"), "the opponent now has the most")
	if tb.landGate(tb.opp) == nil {
		t.Error("the opponent with the most permanents may play a land")
	}
	if tb.castGate(tb.me, "Creature — Bear") != nil {
		t.Error("the Engine's controller is restricted while an opponent controls more")
	}
}

// Tokens are permanents; a player who has left the game is not a player.
func TestDampingEngineCountsTokensAndSkipsEliminatedPlayers(t *testing.T) {
	tb := newDampingEngineTable(t)
	// Every opponent has one permanent, a token, and so has seat 0 (the
	// Engine): a tie.
	for _, p := range tb.g.Seats {
		if p.ID != tb.me.ID {
			seedLandOnBattlefield(tb.g, p.ID, "Spirit", "Token Creature — Spirit")
		}
	}
	if tb.castGate(tb.me, "Creature — Bear") != nil {
		t.Fatal("setup: a tie restricts")
	}
	// An opponent who left the game no longer counts: the other two
	// still tie with seat 0.
	tb.g.Seats[(tb.g.Turn.ActiveSeat+1)%len(tb.g.Seats)].Eliminated = true
	if tb.castGate(tb.me, "Creature — Bear") != nil {
		t.Error("an eliminated player's permanents were counted")
	}
	tb.bear(tb.me)
	assertDECantCast(t, tb.castGate(tb.me, "Creature — Bear"), "one more than the two who remain")
}

// The escape: the restricted player takes the special action, sacrificing
// a permanent of THEIR choice, and both gates open for them until the
// turn ends. Nobody else's restriction changes.
func TestIgnoringDampingEngineOpensBothGatesUntilEndOfTurn(t *testing.T) {
	tb := newDampingEngineTable(t)
	g := tb.g
	cheap, dear := tb.bear(tb.me), tb.bear(tb.me)
	offers := tb.offers(tb.me)
	if len(offers) != 1 || offers[0].Source != tb.engine || offers[0].SourceName != "Damping Engine" {
		t.Fatalf("offers = %+v, want one from the Engine", offers)
	}
	if got := len(offers[0].Candidates); got != 3 {
		t.Errorf("%d candidates, want the Engine and both bears", got)
	}
	if len(tb.offers(tb.opp)) != 0 {
		t.Error("a player the effect does not bear on is offered an ignore")
	}

	// The player picks which permanent: not the first of the payment
	// order, the one they name.
	if err := tb.ignore(tb.me, tb.engine, dear); err != nil {
		t.Fatalf("ignore: %v", err)
	}
	if !tb.inGraveyard(tb.me, dear) || g.Battlefield.Contains(dear) {
		t.Error("the chosen permanent was not sacrificed")
	}
	if !g.Battlefield.Contains(cheap) || !g.Battlefield.Contains(tb.engine) {
		t.Error("something other than the chosen permanent was sacrificed")
	}
	if tb.castGate(tb.me, "Creature — Bear") != nil || tb.castGate(tb.me, "Artifact") != nil ||
		tb.castGate(tb.me, "Enchantment") != nil || tb.landGate(tb.me) != nil {
		t.Error("the restricted player is still refused after ignoring the Engine")
	}
	if len(tb.offers(tb.me)) != 0 {
		t.Error("an ignore already taken is offered again")
	}
	// The effect is per player: the opponent, once they have the most,
	// is restricted exactly as before.
	tb.bear(tb.opp)
	tb.bear(tb.opp)
	tb.bear(tb.opp)
	assertDECantCast(t, tb.castGate(tb.opp, "Creature — Bear"), "the opponent")

	// The cast and the land play themselves go through, not just the gates.
	advanceToMain(t, g)
	land := lpHandLand(tb.me, "Forest")
	if err := g.CastSpell(tb.me.ID, land, game.CastSpellParams{}); err != nil {
		t.Errorf("the land play after ignoring: %v", err)
	}
	creature := uuid.New()
	tb.me.Hand.PushTop(game.Card{InstanceID: creature, Name: "Free Bear", TypeLine: "Creature — Bear",
		ManaCost: "{0}", Power: 1, Toughness: 1, Owner: tb.me.ID, Controller: tb.me.ID})
	if err := g.CastSpell(tb.me.ID, creature, game.CastSpellParams{}); err != nil {
		t.Errorf("the creature cast after ignoring: %v", err)
	}
}

// "Until end of turn": the record is swept at cleanup, and the player who
// still has the most permanents is restricted again.
func TestIgnoringDampingEngineEndsWithTheTurn(t *testing.T) {
	tb := newDampingEngineTable(t)
	bear := tb.bear(tb.me)
	if err := tb.ignore(tb.me, tb.engine, bear); err != nil {
		t.Fatalf("ignore: %v", err)
	}
	if tb.landGate(tb.me) != nil {
		t.Fatal("setup: the ignore did not open the gate")
	}
	records := len(tb.g.ScopedEffects)
	if records != 1 || tb.g.ScopedEffects[0].Mods[0].Kind != game.ModIgnoreEffect {
		t.Fatalf("scoped records = %+v, want one ignoreEffect", tb.g.ScopedEffects)
	}
	endTurn(t, tb.g)
	if n := len(tb.g.ScopedEffects); n != 0 {
		t.Errorf("%d records survived the end of the turn", n)
	}
	assertDECantCast(t, tb.castGate(tb.me, "Creature — Bear"), "after the turn ended")
	if tb.landGate(tb.me) == nil {
		t.Error("the land gate stayed open after the turn ended")
	}
	if len(tb.offers(tb.me)) != 1 {
		t.Error("the ignore is not on offer again the next turn")
	}
}

// Another Engine is another effect: each is ignored on its own, for its own
// sacrifice, and the other keeps refusing in the meantime.
func TestEachDampingEngineIsIgnoredOnItsOwn(t *testing.T) {
	tb := newDampingEngineTable(t)
	second := pushCatalogPermanent(tb.g, tb.me.ID, "Damping Engine", "Artifact", dampingEngineOracle, false)
	a, b := tb.bear(tb.me), tb.bear(tb.me)
	if got := len(tb.offers(tb.me)); got != 2 {
		t.Fatalf("%d offers, want one per Engine", got)
	}
	if err := tb.ignore(tb.me, tb.engine, a); err != nil {
		t.Fatalf("ignore the first: %v", err)
	}
	assertDECantCast(t, tb.castGate(tb.me, "Creature — Bear"), "the second Engine still binds")
	if tb.landGate(tb.me) == nil {
		t.Error("the second Engine stopped binding the land play")
	}
	offers := tb.offers(tb.me)
	if len(offers) != 1 || offers[0].Source != second {
		t.Fatalf("offers = %+v, want only the second Engine's", offers)
	}
	if err := tb.ignore(tb.me, second, b); err != nil {
		t.Fatalf("ignore the second: %v", err)
	}
	if tb.castGate(tb.me, "Creature — Bear") != nil || tb.landGate(tb.me) != nil {
		t.Error("both Engines are ignored and the player is still refused")
	}
}

// An Engine that left and came back is a new object (CR 400.7): the record
// names the old one, so the player is bound again.
func TestIgnoringDampingEngineDoesNotFollowItThroughAZoneChange(t *testing.T) {
	// The exile return re-mints the instance ID; the graveyard return
	// keeps it and relies on the object epoch, which only the record's
	// epoch comparison can tell apart.
	for name, away := range map[string]func(tb *deTable) error{
		"through exile": func(tb *deTable) error { return tb.g.ExileCardForEffect(tb.engine) },
		"through the graveyard": func(tb *deTable) error {
			return tb.g.SacrificePermanentForEffect(tb.engine)
		},
	} {
		tb := newDampingEngineTable(t)
		if err := tb.ignore(tb.me, tb.engine, tb.bear(tb.me)); err != nil {
			t.Fatalf("%s: ignore: %v", name, err)
		}
		if tb.landGate(tb.me) != nil {
			t.Fatalf("%s: setup: not ignoring", name)
		}
		tb.g.WithWriteLock(func() {
			if err := away(tb); err != nil {
				t.Fatalf("%s: leave: %v", name, err)
			}
			if id, err := tb.g.ReturnToBattlefieldForEffect(tb.engine, tb.me.ID, false); err != nil || id == uuid.Nil {
				t.Fatalf("%s: return: %v (%v)", name, err, id)
			}
		})
		if tb.landGate(tb.me) == nil {
			t.Errorf("%s: the returned Engine is a new object and must not be ignored", name)
		}
		if len(tb.offers(tb.me)) != 1 {
			t.Errorf("%s: the returned Engine does not offer its ignore", name)
		}
	}
}

// What the action refuses, and that a refusal leaves the board as it was.
func TestIgnoringDampingEngineRefusals(t *testing.T) {
	tb := newDampingEngineTable(t)
	mine, theirs := tb.bear(tb.me), tb.bear(tb.opp)

	snapshotOK := func(what string) {
		t.Helper()
		if !tb.g.Battlefield.Contains(mine) || !tb.g.Battlefield.Contains(theirs) || !tb.g.Battlefield.Contains(tb.engine) {
			t.Errorf("%s: a refused action sacrificed something", what)
		}
		if len(tb.g.ScopedEffects) != 0 {
			t.Errorf("%s: a refused action wrote a record", what)
		}
	}

	// A permanent the player does not control can't be the cost.
	if err := tb.ignore(tb.me, tb.engine, theirs); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("sacrificing an opponent's permanent: %v, want ErrInvalidParam", err)
	}
	// The cost is named: the engine does not choose for the player.
	if err := tb.ignore(tb.me, tb.engine, uuid.Nil); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("no sacrifice named: %v, want ErrInvalidParam", err)
	}
	if err := tb.ignore(tb.me, tb.engine, uuid.New()); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("a sacrifice that is not on the board: %v, want ErrInvalidParam", err)
	}
	// Nothing to ignore: the opponent is not bound by the Engine.
	if err := tb.ignore(tb.opp, tb.engine, theirs); !errors.Is(err, game.ErrSpecialActionNotOffered) {
		t.Errorf("an unaffected player: %v, want ErrSpecialActionNotOffered", err)
	}
	// A card with no ignorable effect.
	if err := tb.ignore(tb.me, mine, mine); !errors.Is(err, game.ErrSpecialActionNotOffered) {
		t.Errorf("a permanent with nothing to ignore: %v, want ErrSpecialActionNotOffered", err)
	}
	// And a card that is not on the battlefield.
	if err := tb.ignore(tb.me, uuid.New(), mine); !errors.Is(err, game.ErrCardNotFound) {
		t.Errorf("a card that does not exist: %v, want ErrCardNotFound", err)
	}
	// A sacrifice named on a different special action is refused, and so
	// is another kind's picker on this one.
	if err := tb.g.PerformSpecialAction(tb.me.ID, tb.engine, game.SpecialActionForetell,
		game.SpecialActionParams{SacrificeID: mine}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("sacrifice_id on foretell: %v, want ErrInvalidParam", err)
	}
	if err := tb.g.PerformSpecialAction(tb.me.ID, tb.engine, game.SpecialActionIgnoreEffect,
		game.SpecialActionParams{SacrificeID: mine, Cost: "{2}"}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("a cost on ignore_effect: %v, want ErrInvalidParam", err)
	}
	snapshotOK("refusals")
}

// A permanent that can't be sacrificed can't pay, and a player with
// nothing they can sacrifice is offered nothing.
func TestAPermanentThatCantBeSacrificedCantPayForTheEngineIgnore(t *testing.T) {
	tb := newDampingEngineTable(t)
	locked := tb.bear(tb.me)
	free := tb.bear(tb.me)
	tb.g.WithWriteLock(func() { tb.g.CantBeSacrificedThisTurnForEffect(uuid.Nil, locked, "Test") })
	for _, c := range tb.offers(tb.me)[0].Candidates {
		if c == locked {
			t.Error("a permanent that can't be sacrificed is a candidate")
		}
	}
	if err := tb.ignore(tb.me, tb.engine, locked); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("paying with it: %v, want ErrInvalidParam", err)
	}
	tb.g.WithWriteLock(func() { tb.g.CantBeSacrificedThisTurnForEffect(uuid.Nil, free, "Test") })
	tb.g.WithWriteLock(func() { tb.g.CantBeSacrificedThisTurnForEffect(uuid.Nil, tb.engine, "Test") })
	if len(tb.offers(tb.me)) != 0 {
		t.Error("an ignore the player cannot pay for is offered")
	}
}

// The Engine may be the permanent given up: it ends the effect for
// everybody, and no record of an ignore is left naming a dead object.
func TestSacrificingTheEngineItselfEndsItWithoutARecord(t *testing.T) {
	tb := newDampingEngineTable(t)
	if err := tb.ignore(tb.me, tb.engine, tb.engine); err != nil {
		t.Fatalf("ignore: %v", err)
	}
	if tb.g.Battlefield.Contains(tb.engine) || !tb.inGraveyard(tb.me, tb.engine) {
		t.Error("the Engine was not sacrificed")
	}
	if len(tb.g.ScopedEffects) != 0 {
		t.Errorf("a record names the sacrificed Engine: %+v", tb.g.ScopedEffects)
	}
	if tb.castGate(tb.me, "Creature — Bear") != nil || tb.landGate(tb.me) != nil {
		t.Error("the Engine is gone and still restricts")
	}
}

// The sacrifice is a real sacrifice: the event is emitted and a
// "whenever you sacrifice" watcher sees it; the special action itself is
// logged with the card's row.
func TestTheDampingEngineIgnoreCostIsASacrificeAndTheActionIsLogged(t *testing.T) {
	tb := newDampingEngineTable(t)
	bear := tb.bear(tb.me)
	before := len(tb.g.EventsThisTurn())
	if err := tb.ignore(tb.me, tb.engine, bear); err != nil {
		t.Fatalf("ignore: %v", err)
	}
	var sacrificed, acted bool
	for _, ev := range tb.g.EventsThisTurn()[before:] {
		switch ev.Kind {
		case game.EventSacrifice:
			sacrificed = sacrificed || (ev.CardID == bear && ev.Actor == tb.me.ID)
		case game.EventSpecialAction:
			acted = acted || (ev.CardID == tb.engine && ev.Actor == tb.me.ID && strings.Contains(ev.Label, "Damping Engine"))
		}
	}
	if !sacrificed {
		t.Error("no sacrifice event for the cost")
	}
	if !acted {
		t.Error("no special-action event naming the Engine")
	}
}

// A restore point taken while the ignore is live restores it: the record
// names the Engine as an object and the player, and both gates stay open
// in the restored game, which still lifts at the end of the turn.
func TestADampingEngineIgnoreSurvivesARestore(t *testing.T) {
	tb := newDampingEngineTable(t)
	if err := tb.ignore(tb.me, tb.engine, tb.bear(tb.me)); err != nil {
		t.Fatalf("ignore: %v", err)
	}
	g := restoreRoundTrip(t, tb.g, true)
	rt := &deTable{g: g, me: g.Seats[0], opp: g.Seats[1], engine: tb.engine}
	if rt.me.ID != tb.me.ID {
		t.Fatal("setup: the seats moved")
	}
	if rt.castGate(rt.me, "Creature — Bear") != nil || rt.landGate(rt.me) != nil {
		t.Error("after a restore the ignore is gone")
	}
	if len(rt.offers(rt.me)) != 0 {
		t.Error("after a restore the ignore is offered again")
	}
	endTurn(t, g)
	assertDECantCast(t, rt.castGate(rt.me, "Creature — Bear"), "restored game, next turn")
}

// Losing the Engine's abilities stops it restricting, and stops the offer.
func TestAnEngineThatLostItsAbilitiesRestrictsAndOffersNothing(t *testing.T) {
	tb := newDampingEngineTable(t)
	if len(tb.offers(tb.me)) != 1 {
		t.Fatal("setup: no offer")
	}
	tb.g.WithWriteLock(func() {
		tb.g.RegisterScopedEffectForEffect(tb.engine, tb.g.PinnedObjectsLocked(tb.engine),
			[]game.Mod{game.LoseAllAbilitiesMod()}, tb.g.UntilEndOfTurnDuration(), "Humility")
		tb.g.RecomputeLayersIfStaleLocked()
	})
	if len(tb.offers(tb.me)) != 0 {
		t.Error("a permanent that lost its abilities still offers an ignore")
	}
	if tb.landGate(tb.me) != nil || tb.castGate(tb.me, "Creature — Bear") != nil {
		t.Error("and still restricts")
	}
}

// Register refuses an ignore that switches nothing off, and a marked
// restriction with no ignore.
func TestRegisterRefusesAnEngineIgnoreWithoutItsRestrictionAndTheOtherWayRound(t *testing.T) {
	ig := &game.IgnorableEffect{
		Label:   "Sacrifice a permanent: ignore X",
		Applies: func(*game.Game, *game.Card, uuid.UUID) bool { return true },
	}
	forbid := func(game.LandPlayQuery) bool { return true }
	expectRegisterPanic(t, "no restriction is Ignorable", func() {
		checkIgnorableEffect(Spec{Name: "Fixture", IgnorableEffect: ig})
	})
	expectRegisterPanic(t, "declares no IgnorableEffect", func() {
		checkIgnorableEffect(Spec{Name: "Fixture", LandPlayRestrictions: []game.LandPlayRestriction{
			{Label: "x", Forbids: forbid, Ignorable: true}}})
	})
	expectRegisterPanic(t, "without its Label or Applies", func() {
		checkIgnorableEffect(Spec{Name: "Fixture", IgnorableEffect: &game.IgnorableEffect{Label: "x"},
			LandPlayRestrictions: []game.LandPlayRestriction{{Label: "x", Forbids: forbid, Ignorable: true}}})
	})
	// The right shape passes, with either gate's restriction marked.
	checkIgnorableEffect(Spec{Name: "Fixture", IgnorableEffect: ig,
		CastRestrictions: []game.CastRestriction{{Label: "x", Forbids: func(game.CastQuery) bool { return true }, Ignorable: true}}})
	checkIgnorableEffect(Spec{Name: "Fixture"})
}
