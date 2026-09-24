package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_gate_view_test.go — the VIEW half of #664 and #760 (ADR 0073
// §7 and §9). The client greys a card and names the clause from
// server data, so the stamp has to be the same gate's answer and the
// offers have to be the same catalog's.
//
// Both tests stub the catalog hooks rather than registering real
// cards: this package must not import the effects catalog, and the
// question being asked is about the projection, not about any card.

const (
	viewKickerOracle      = "test-view-kicker"
	viewRestrictionOracle = "test-view-restriction"
)

func stubViewOptionalCosts(t *testing.T, oracleID string, costs []game.AdditionalCost) {
	t.Helper()
	prev := game.CatalogOptionalCosts
	game.CatalogOptionalCosts = func(id string) []game.AdditionalCost {
		if id == oracleID {
			return costs
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogOptionalCosts = prev })
}

func stubViewCastRestrictions(t *testing.T, oracleID string, rules []game.CastRestriction) {
	t.Helper()
	prev := game.CatalogCastRestrictions
	game.CatalogCastRestrictions = func(id string) []game.CastRestriction {
		if id == oracleID {
			return rules
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogCastRestrictions = prev })
}

func handCardIn(t *testing.T, v GameView, seat, instanceID string) CardView {
	t.Helper()
	for _, s := range v.Seats {
		if s.ID != seat {
			continue
		}
		for _, c := range s.Hand.Cards {
			if c.InstanceID == instanceID {
				return c
			}
		}
	}
	t.Fatalf("card %s not found in %s's hand view", instanceID, seat)
	return CardView{}
}

// TestOptionalCostsAreOfferedOnTheCard is #664's view half: the
// client cannot render a kicker toggle it was never told about.
func TestOptionalCostsAreOfferedOnTheCard(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	stubViewOptionalCosts(t, viewKickerOracle, []game.AdditionalCost{
		{Optional: true, Key: game.KickerKey, ManaCost: "{4}", Label: "Kicker {4}"},
		{Optional: true, Key: game.MultikickerKey, ManaCost: "{G}", Repeat: 7, Label: "Multikicker {G}"},
	})

	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id,
			Name:       "Kickable",
			TypeLine:   "Instant",
			OracleID:   viewKickerOracle,
			ManaCost:   "{R}",
			Owner:      me.ID,
			Controller: me.ID,
			// The owner knows their own hand card. Without it the
			// per-viewer filter redacts the whole cost surface, which
			// is correct behaviour and would make this test vacuous.
			KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if len(c.OptionalCosts) != 2 {
		t.Fatalf("optional_costs = %+v, want two offers", c.OptionalCosts)
	}
	if c.OptionalCosts[0].Index != 0 || c.OptionalCosts[0].Key != game.KickerKey {
		t.Errorf("first offer = %+v, want index 0 keyed kicker", c.OptionalCosts[0])
	}
	if c.OptionalCosts[0].MaxTimes != 1 {
		t.Errorf("kicker max_times = %d, want 1", c.OptionalCosts[0].MaxTimes)
	}
	// The multikicker's cap is what turns the client's checkbox into
	// a stepper, so it is the one number this projection must not
	// drop.
	if c.OptionalCosts[1].MaxTimes != 7 {
		t.Errorf("multikicker max_times = %d, want 7", c.OptionalCosts[1].MaxTimes)
	}
	if c.CantCast != "" {
		t.Errorf("nothing restricts casting, but cant_cast = %q", c.CantCast)
	}
}

// TestCantCastIsStampedFromTheSameGate is #760's view half: the
// refusal the engine would give, on the card, before the click.
func TestCantCastIsStampedFromTheSameGate(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const clause = "Test Warden — each player can't cast more than one spell each turn."
	stubViewCastRestrictions(t, viewRestrictionOracle, []game.CastRestriction{{
		Label: clause,
		Forbids: func(q game.CastQuery) bool {
			return q.Game.CastTallyFor(q.Controller).Total >= 1
		},
	}})

	id := uuid.New()
	warden := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id,
			Name:       "Second Spell",
			TypeLine:   "Instant",
			ManaCost:   "{R}",
			Owner:      me.ID,
			Controller: me.ID,
			KnownBy:    map[uuid.UUID]bool{me.ID: true},
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: warden,
			Name:       "Test Warden",
			TypeLine:   "Enchantment",
			OracleID:   viewRestrictionOracle,
			Owner:      me.ID,
			Controller: me.ID,
		})
	})

	// Nothing cast yet: the card is clean.
	if got := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String()).CantCast; got != "" {
		t.Errorf("before any cast, cant_cast = %q, want empty", got)
	}

	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = make(map[uuid.UUID]game.CastTally)
		}
		g.SpellsCastThisTurn[me.ID] = game.CastTally{Total: 1}
	})

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if c.CantCast == "" {
		t.Fatalf("a restricted card carries no cant_cast")
	}
	if !strings.Contains(c.CantCast, "more than one spell") {
		t.Errorf("cant_cast = %q, want the printed clause", c.CantCast)
	}
	if c.CastableHere {
		t.Errorf("a card the gate refuses is still marked castable_here")
	}
	// And the engine agrees, which is the whole point of the stamp
	// being the same function's answer.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
		t.Errorf("the engine allowed a cast the view greyed out")
	}
}

// TestCantCastIsStampedFromAGrantedBan is #1316's view half: the same
// `cant_cast` stamp, this time from a PLAYER-scoped grant rather than
// a permanent's static — Avatar's Wrath and Mandate of Peace's shape,
// with no card on the battlefield at all to explain the refusal.
func TestCantCastIsStampedFromAGrantedBan(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const clause = "Test Mandate — can't cast spells this turn"

	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id,
			Name:       "Any Instant",
			TypeLine:   "Instant",
			ManaCost:   "{R}",
			Owner:      me.ID,
			Controller: me.ID,
			KnownBy:    map[uuid.UUID]bool{me.ID: true},
		})
	})

	if got := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String()).CantCast; got != "" {
		t.Errorf("before any grant, cant_cast = %q, want empty", got)
	}

	g.WithWriteLock(func() {
		g.GrantCastBanForEffect(me.ID, game.CastBanRule{Kind: game.CastBanOutright}, clause, uuid.Nil, game.Duration{})
	})

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if c.CantCast != clause {
		t.Errorf("cant_cast = %q, want %q", c.CantCast, clause)
	}
	if c.CastableHere {
		t.Error("a card a granted ban refuses is still marked castable_here")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
		t.Error("the engine allowed a cast the view greyed out")
	}
}

// TestCantCastIsNeverStampedOnALand is #1439's view half. Playing a
// land is a special action (CR 305.1, CR 116.2a), never a cast, so
// CastGateLocked must never even be asked about one. Before the fix
// castStampsFor asked it unconditionally, so a restriction shaped
// like Rule of Law stamped `cant_cast` onto a land in hand — which
// every reader of that field (`castIsForbidden` and friends) would
// then read as "this land can't be played", even though it always
// could.
func TestCantCastIsNeverStampedOnALand(t *testing.T) {
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	advanceTo(t, g, game.StepPrecombatMain)
	active := g.Seats[g.Turn.ActiveSeat]

	const clause = "Test Warden — each player can't cast more than one spell each turn."
	stubViewCastRestrictions(t, viewRestrictionOracle, []game.CastRestriction{{
		Label: clause,
		Forbids: func(q game.CastQuery) bool {
			return q.Game.CastTallyFor(q.Controller).Total >= 1
		},
	}})

	land := uuid.New()
	g.WithWriteLock(func() {
		active.Hand.PushTop(game.Card{
			InstanceID: land,
			Name:       "Forest",
			TypeLine:   "Basic Land — Forest",
			Owner:      active.ID,
			Controller: active.ID,
			// The owner knows their own hand card — otherwise the
			// redaction pass strips the whole card and the test
			// would pass for the wrong reason.
			KnownBy: map[uuid.UUID]bool{active.ID: true},
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Test Warden", TypeLine: "Enchantment",
			OracleID: viewRestrictionOracle, Owner: active.ID, Controller: active.ID,
		})
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = make(map[uuid.UUID]game.CastTally)
		}
		g.SpellsCastThisTurn[active.ID] = game.CastTally{Total: 1}
	})

	c := handCardIn(t, ViewOfGameFor(g, active.ID.String()), active.ID.String(), land.String())
	if c.CantCast != "" {
		t.Errorf("a land carries cant_cast = %q, want empty — playing a land is not a cast (CR 116.2a)", c.CantCast)
	}

	// And the engine agrees: CastSpell actually accepts the play the
	// view left un-greyed.
	if err := g.CastSpell(active.ID, land, game.CastSpellParams{}); err != nil {
		t.Errorf("the engine refused the land play the view left un-greyed: %v", err)
	}
}
