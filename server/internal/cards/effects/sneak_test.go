package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// sneak_test.go — #4586: sneak (CR 702.190) and its first card, Elektra,
// Daughter of the Hand. The engine half is game/sneak.go; the cost is
// ninjutsu's unblocked attacker (ninjutsu_test.go) paid as an
// alternative cost (Daze's ReturnToHand).

const (
	elektraDaughterOracle = "b46647a1-0c91-451b-8c0b-be8fe02faee2"
	elektraDaughterName   = "Elektra, Daughter of the Hand"
	elektraDaughterType   = "Legendary Creature — Human Ninja Villain"
	// A test-only instant with sneak: no printed instant with the keyword
	// is in the catalog yet (TMNT, #4457, has them).
	sneakInstantOracle = "test-sneak-instant-4586"
)

// pushElektraToHand puts Elektra in p's hand with her printed cost.
func pushElektraToHand(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: elektraDaughterName, TypeLine: elektraDaughterType, ManaCost: "{2}{B}{B}",
			OracleID: elektraDaughterOracle, Power: 3, Toughness: 3,
			Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
		})
	})
	return id
}

// sneakBoard is ninjutsuBoard's sneak twin: my 1/1 swings at `defender`
// (seat 1, or a planeswalker of seat 1's), every defender finishes
// declaring with no blocks, Elektra is in my hand and {B}{B}{B} floats.
// The cursor is in MY declare blockers step with priority back to me.
func sneakBoard(t *testing.T, g *game.Game, defender uuid.UUID) (elektra, attacker uuid.UUID, me, opp *game.Player) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	attacker = b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	if defender == uuid.Nil {
		defender = opp.ID
	}
	elektra = pushElektraToHand(g, me)
	declareAttack(t, g, defender, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	finishBlockDeclarations(t, g)
	floatMana(t, g, me, "{B}{B}{B}")
	return elektra, attacker, me, opp
}

func sneakCast(g *game.Game, me *game.Player, card, returned uuid.UUID) error {
	return g.CastSpell(me.ID, card, game.CastSpellParams{AlternativeCost: "sneak", AltCostIDs: []uuid.UUID{returned}})
}

func TestElektraDaughterOfTheHandIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(elektraDaughterOracle)
	if !ok {
		t.Fatal("Elektra, Daughter of the Hand is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	alt := game.AlternativeCostByKey(elektraDaughterOracle, "sneak")
	if alt == nil || !alt.Sneak || alt.ManaCost != "{1}{B}{B}" || alt.ReturnToHand == nil || alt.Label != "Sneak {1}{B}{B}" {
		t.Fatalf("sneak offer %+v, want Sneak {1}{B}{B} returning a creature", alt)
	}
}

// The headline: the unblocked attacker goes to its owner's hand as the
// cost, and Elektra enters tapped and attacking the same player
// (CR 702.190b); her enter trigger still destroys a small creature.
func TestElektraSneaksInTappedAndAttackingTheSamePlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	elektra, attacker, _, _ := sneakBoard(t, g, uuid.Nil)

	if err := sneakCast(g, me, elektra, attacker); err != nil {
		t.Fatalf("sneak cast: %v", err)
	}
	if !me.Hand.Contains(attacker) {
		t.Fatal("the unblocked attacker was not returned to its owner's hand as the cost")
	}
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)

	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("the test left the declare blockers step (at %s)", g.Turn.Step)
	}
	c, ok := g.LookupCardForEffect(elektra)
	if !ok || !g.Battlefield.Contains(elektra) {
		t.Fatal("Elektra is not on the battlefield")
	}
	if !c.Tapped {
		t.Error("Elektra entered untapped")
	}
	if c.AttackingTarget != opp.ID {
		t.Errorf("Elektra attacks %s, want %s, the player the returned creature attacked", c.AttackingTarget, opp.ID)
	}
	if n := attackEventsFor(g, elektra); n != 0 {
		t.Errorf("%d EventAttack for Elektra, want 0: she was never declared (CR 506.3c)", n)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("her enter trigger did not destroy the 2/2")
	}
}

// The same defender when it is a planeswalker.
func TestElektraSneaksInAttackingTheSamePlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	walker := pushWalkerForTest(g, g.Seats[1].ID, "Test Walker", "", 5)
	elektra, attacker, me, _ := sneakBoard(t, g, walker)

	if err := sneakCast(g, me, elektra, attacker); err != nil {
		t.Fatalf("sneak cast: %v", err)
	}
	assistCardSettle(t, g, me.ID, nil, true)
	c, ok := g.LookupCardForEffect(elektra)
	if !ok || !c.Tapped || c.AttackingTarget != walker {
		t.Fatalf("Elektra tapped=%v attacking %s, want tapped and attacking the walker %s", c.Tapped, c.AttackingTarget, walker)
	}
}

// Sneaked in after blocks, she deals her combat damage unblocked.
func TestElektraSneakedInDealsCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	elektra, attacker, me, opp := sneakBoard(t, g, uuid.Nil)
	life := opp.Life
	if err := sneakCast(g, me, elektra, attacker); err != nil {
		t.Fatalf("sneak cast: %v", err)
	}
	assistCardSettle(t, g, me.ID, nil, true)
	advanceTo(t, g, game.StepCombatDamage)
	if got := opp.Life; got != life-3 {
		t.Errorf("defender life %d, want %d after Elektra's 3", got, life-3)
	}
}

// A blocked attacker is not an unblocked creature (CR 509.1h): refused,
// and the enumerator never names it.
func TestSneakRefusesABlockedAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blocked := b12Creature(g, me.ID, "Blocked Rat", "Creature — Rat", 1, 1)
	free := b12Creature(g, me.ID, "Free Rat", "Creature — Rat", 1, 1)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	elektra := pushElektraToHand(g, me)
	declareAttack(t, g, opp.ID, blocked, free)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, blocked); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	finishBlockDeclarations(t, g)
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("left the declare blockers step (at %s)", g.Turn.Step)
	}
	floatMana(t, g, me, "{B}{B}{B}")

	offered := sneakMoves(t, g, me.ID, elektra)
	if len(offered) != 1 || offered[0] != free {
		t.Errorf("enumerator offers sneak paying %v, want only the unblocked %s", offered, free)
	}
	if err := sneakCast(g, me, elektra, blocked); err == nil {
		t.Fatal("sneak paid with a blocked attacker")
	}
	if !g.Battlefield.Contains(blocked) || !me.Hand.Contains(elektra) {
		t.Fatal("the refused cast moved something")
	}
	if err := sneakCast(g, me, elektra, free); err != nil {
		t.Fatalf("sneak with the unblocked attacker: %v", err)
	}
}

// sneakMoves lists the creature each enumerated sneak cast of `card`
// returns.
func sneakMoves(t *testing.T, g *game.Game, seat, card uuid.UUID) []uuid.UUID {
	t.Helper()
	var out []uuid.UUID
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != legal.TypeCastSpell || m.Source != card {
			continue
		}
		var p struct {
			AlternativeCost string   `json:"alternative_cost"`
			AltCostIDs      []string `json:"alt_cost_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		if p.AlternativeCost != "sneak" {
			t.Errorf("enumerator offers %q for Elektra outside a main phase, want only sneak", p.AlternativeCost)
			continue
		}
		for _, s := range p.AltCostIDs {
			out = append(out, uuid.MustParse(s))
		}
	}
	return out
}

// The bot's move list offers the sneak cast in the window, and the move
// as enumerated is one the engine accepts.
func TestSneakIsOfferedToTheBotLegally(t *testing.T) {
	g := newCatalogGame(t)
	elektra, attacker, me, opp := sneakBoard(t, g, uuid.Nil)
	var move *legal.Move
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == legal.TypeCastSpell && m.Source == elektra {
			m := m
			move = &m
		}
	}
	if move == nil {
		t.Fatal("the enumerator offers no sneak cast of Elektra in the declare blockers step")
	}
	var p struct {
		AlternativeCost string   `json:"alternative_cost"`
		AltCostIDs      []string `json:"alt_cost_ids"`
	}
	if err := json.Unmarshal(move.Params, &p); err != nil {
		t.Fatalf("move params: %v", err)
	}
	if p.AlternativeCost != "sneak" || len(p.AltCostIDs) != 1 || p.AltCostIDs[0] != attacker.String() {
		t.Fatalf("move %+v, want sneak returning %s", p, attacker)
	}
	if err := g.CastSpell(me.ID, elektra, game.CastSpellParams{
		AlternativeCost: p.AlternativeCost, AltCostIDs: []uuid.UUID{attacker}, AutoTap: true, Strict: true,
	}); err != nil {
		t.Fatalf("the enumerated move is refused: %v", err)
	}
	assistCardSettle(t, g, me.ID, nil, true)
	if c, _ := g.LookupCardForEffect(elektra); c.AttackingTarget != opp.ID {
		t.Errorf("Elektra attacks %s, want %s", c.AttackingTarget, opp.ID)
	}
}

// The window is the caster's own declare blockers step (CR 702.190a):
// not the main phase, not the declare attackers step, not the combat
// damage step (where ninjutsu would still be legal), and not an
// opponent's declare blockers step.
func TestSneakIsCastableOnlyInYourDeclareBlockersStep(t *testing.T) {
	alt := game.AlternativeCostByKey(elektraDaughterOracle, "sneak")
	open := func(g *game.Game, p *game.Player, card uuid.UUID) bool {
		var ok bool
		g.ReadSnapshot(func() {
			c, _ := g.LookupCardForEffect(card)
			ok = g.CastTimingForOfferOpenLocked(p.ID, c, game.ZoneHand, nil, alt)
		})
		return ok
	}

	t.Run("main phase", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		rat := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
		elektra := pushElektraToHand(g, me)
		floatMana(t, g, me, "{B}{B}{B}")
		if open(g, me, elektra) {
			t.Error("the sneak window is open in the main phase")
		}
		if err := sneakCast(g, me, elektra, rat); err == nil {
			t.Error("sneak cast in the main phase was accepted")
		}
	})

	t.Run("declare attackers step", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rat := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
		elektra := pushElektraToHand(g, me)
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(rat, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		floatMana(t, g, me, "{B}{B}{B}")
		if open(g, me, elektra) {
			t.Error("the sneak window is open before blockers")
		}
		if err := sneakCast(g, me, elektra, rat); err == nil {
			t.Error("sneak cast before blockers was accepted")
		}
	})

	t.Run("while the defender is still declaring", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rat := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
		b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
		elektra := pushElektraToHand(g, me)
		declareAttack(t, g, opp.ID, rat)
		advanceTo(t, g, game.StepDeclareBlockers)
		floatMana(t, g, me, "{B}{B}{B}")
		if err := sneakCast(g, me, elektra, rat); err == nil {
			t.Error("sneak cast before the defender finished blocking was accepted")
		}
		if !g.Battlefield.Contains(rat) {
			t.Error("the refused cast returned the attacker")
		}
	})

	t.Run("combat damage step", func(t *testing.T) {
		g := newCatalogGame(t)
		elektra, attacker, me, _ := sneakBoard(t, g, uuid.Nil)
		advanceTo(t, g, game.StepCombatDamage)
		if open(g, me, elektra) {
			t.Error("the sneak window is open in the combat damage step")
		}
		if err := sneakCast(g, me, elektra, attacker); !errors.Is(err, game.ErrSorcerySpeedRequired) {
			t.Errorf("sneak cast in the damage step: %v, want ErrSorcerySpeedRequired", err)
		}
	})

	t.Run("an opponent's declare blockers step", func(t *testing.T) {
		g := newCatalogGame(t)
		_, _, me, opp := sneakBoard(t, g, uuid.Nil)
		theirs := pushElektraToHand(g, opp)
		if open(g, opp, theirs) {
			t.Error("the sneak window is open for the defending player")
		}
		if mine := pushElektraToHand(g, me); !open(g, me, mine) {
			t.Error("the sneak window is shut for the active player in their declare blockers step")
		}
	})

	t.Run("the printed cost stays sorcery-speed", func(t *testing.T) {
		g := newCatalogGame(t)
		elektra, _, me, _ := sneakBoard(t, g, uuid.Nil)
		floatMana(t, g, me, "{B}")
		if err := g.CastSpell(me.ID, elektra, game.CastSpellParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
			t.Errorf("printed-cost cast in combat: %v, want ErrSorcerySpeedRequired", err)
		}
	})
}

// A returned creature goes to its OWNER's hand, not the caster's: a
// creature I control but an opponent owns.
func TestSneakReturnsTheCreatureToItsOwnersHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Borrowed Rat", TypeLine: "Creature — Rat",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: me.ID,
	})
	elektra := pushElektraToHand(g, me)
	declareAttack(t, g, opp.ID, stolen)
	advanceTo(t, g, game.StepDeclareBlockers)
	finishBlockDeclarations(t, g)
	floatMana(t, g, me, "{B}{B}{B}")
	if err := sneakCast(g, me, elektra, stolen); err != nil {
		t.Fatalf("sneak cast: %v", err)
	}
	if !opp.Hand.Contains(stolen) || me.Hand.Contains(stolen) {
		t.Error("the returned creature did not go to its owner's hand")
	}
}

// The defender rides the spell through an undo clone and a snapshot
// round trip taken while Elektra is on the stack.
func TestSneakDefenderSurvivesUndoAndRestore(t *testing.T) {
	for _, tc := range []struct {
		name    string
		reshape func(t *testing.T, g *game.Game) *game.Game
	}{
		{"clone", func(_ *testing.T, g *game.Game) *game.Game { return g.Clone() }},
		{"snapshot", func(t *testing.T, g *game.Game) *game.Game { return restoreRoundTrip(t, g, true) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			elektra, attacker, me, opp := sneakBoard(t, g, uuid.Nil)
			if err := sneakCast(g, me, elektra, attacker); err != nil {
				t.Fatalf("sneak cast: %v", err)
			}
			restored := tc.reshape(t, g)
			assistCardSettle(t, restored, me.ID, nil, true)
			c, ok := restored.LookupCardForEffect(elektra)
			if !ok || !c.Tapped || c.AttackingTarget != opp.ID {
				t.Fatalf("restored: tapped=%v attacking %s, want tapped and attacking %s", c.Tapped, c.AttackingTarget, opp.ID)
			}
		})
	}
}

// An instant with sneak takes the window and the cost and resolves as an
// ordinary instant: nothing enters, nothing attacks.
func TestAnInstantWithSneakResolvesNormally(t *testing.T) {
	registerForTest(t, Spec{
		OracleID: sneakInstantOracle, Name: "Sneak Strike", Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 3)
		},
	})
	g := newCatalogGame(t)
	_, attacker, me, _ := sneakBoard(t, g, uuid.Nil)
	floatMana(t, g, me, "{R}")
	strike := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: strike, Name: "Sneak Strike", TypeLine: "Instant", ManaCost: "{3}{R}",
			OracleID: sneakInstantOracle, Owner: me.ID, Controller: me.ID,
		})
	})
	life := me.Life
	if err := sneakCast(g, me, strike, attacker); err != nil {
		t.Fatalf("sneak cast of an instant: %v", err)
	}
	if !me.Hand.Contains(attacker) {
		t.Error("the attacker was not returned as the cost")
	}
	assistCardSettle(t, g, me.ID, nil, true)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d: the instant did not resolve", me.Life, life+3)
	}
	if !me.Graveyard.Contains(strike) {
		t.Error("the instant did not go to the graveyard")
	}
}

// The client's cost picker: in the window the sneak offer is on the wire,
// open, with the unblocked attacker (and not the blocked one) as its
// payment, while the printed cost is shut for timing.
func TestSneakOfferOnTheWire(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blocked := b12Creature(g, me.ID, "Blocked Rat", "Creature — Rat", 1, 1)
	free := b12Creature(g, me.ID, "Free Rat", "Creature — Rat", 1, 1)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	elektra := pushElektraToHand(g, me)
	declareAttack(t, g, opp.ID, blocked, free)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, blocked); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	finishBlockDeclarations(t, g)
	floatMana(t, g, me, "{B}{B}{B}")

	card, offer := sneakOfferView(g, me.ID, elektra)
	if offer == nil {
		t.Fatal("the sneak offer is not on the wire")
	}
	if offer.TimingClosed || !card.PrintedCostTimingClosed {
		t.Errorf("sneak timing_closed=%v, printed_cost_timing_closed=%v, want the sneak row open and the printed one shut",
			offer.TimingClosed, card.PrintedCostTimingClosed)
	}
	if offer.PayOptions == nil || len(offer.PayOptions.Cards) != 1 || offer.PayOptions.Cards[0] != free.String() {
		t.Errorf("pay_options %+v, want only the unblocked %s", offer.PayOptions, free)
	}
}

// sneakOfferView finds a hand card and its sneak offer in viewer's frame.
func sneakOfferView(g *game.Game, viewer, card uuid.UUID) (*protocol.CardView, *protocol.AlternativeCostView) {
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			c := &s.Hand.Cards[i]
			if c.InstanceID != card.String() {
				continue
			}
			for j := range c.AlternativeCosts {
				if c.AlternativeCosts[j].Key == "sneak" {
					return c, &c.AlternativeCosts[j]
				}
			}
			return c, nil
		}
	}
	return nil, nil
}
