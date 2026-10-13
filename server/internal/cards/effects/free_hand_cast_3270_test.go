package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3270: "you may cast a spell from your hand without paying its mana cost"
// is cast WHILE the ability resolves (CR 608.2g), ignoring timing, and the
// chance ends when the holder passes. These tests drive Buster Sword through
// a real combat to the combat damage step, where a sorcery is not castable
// on its own, and then play the client's half (the cast_spell the free-cast
// chain sends) against the engine.

const mindRotOracle = "ad44cf74-b717-48fb-9fa2-77512024d76a"

// busterTable is a combat with Buster Sword's carrier connecting for 5: the
// game sits in the combat damage step with the draw already resolved and the
// free-cast pick pending. `black` makes the carrier black, so Doom Blade has
// no legal target on the table.
type busterTable struct {
	g       *game.Game
	me, opp *game.Player
	bolt    uuid.UUID
	mindRot uuid.UUID
	doom    uuid.UUID
	fodder  uuid.UUID
}

func newBusterTable(t *testing.T, black bool) *busterTable {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	colors := []string(nil)
	if black {
		colors = []string{"B"}
	}
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Colors: colors, Owner: me.ID, Controller: me.ID,
	})
	sword := seedEquipment(g, me.ID, "Buster Sword", busterSwordOracle)
	equipTo(t, g, me.ID, sword, bear)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	tb := &busterTable{g: g, me: me, opp: opp}
	tb.bolt = handCardFull(me, "Lightning Bolt", "Instant", "{R}", boltOracleCombat, []string{"R"})
	tb.mindRot = handCardFull(me, "Mind Rot", "Sorcery", "{2}{B}", mindRotOracle, []string{"B"})
	tb.doom = handCardFull(me, "Doom Blade", "Instant", "{1}{B}", doomBladeOracle, []string{"B"})
	tb.fodder = handCardFull(me, "Fodder", "Sorcery", "{6}", "", nil)

	attackWith(t, g, opp.ID, bear)
	if opp.Life != 35 {
		t.Fatalf("opponent life %d, want 35 (Buster Sword's 5 connected)", opp.Life)
	}
	passPriorityAroundTable(t, g)
	return tb
}

// pick answers the free-cast prompt with `id` (uuid.Nil for none).
func (tb *busterTable) pick(t *testing.T, id uuid.UUID) {
	t.Helper()
	p := latestChoiceOfKindFor(tb.g, game.PendingChoiceChooseCards, tb.me.ID)
	if p == nil {
		t.Fatal("no free-cast pick is waiting")
	}
	if p.ChooseDestination != game.ChooseCastFree {
		t.Fatalf("the pick declares destination %q, want %q (the client opens the cast on it)",
			p.ChooseDestination, game.ChooseCastFree)
	}
	if id == uuid.Nil {
		if err := tb.g.ResolveChooseCards(p.ID, tb.me.ID, nil); err != nil {
			t.Fatalf("choosing nothing: %v", err)
		}
		return
	}
	if !hasID(p.ChooseCards, id) {
		t.Fatalf("card not offered: %v", p.ChooseCards)
	}
	if err := tb.g.ResolveChooseCards(p.ID, tb.me.ID, []uuid.UUID{id}); err != nil {
		t.Fatalf("choosing the free cast: %v", err)
	}
}

// castFree is the client's cast_spell once the pick opens the chain: the
// granted offer claimed, an empty pool, strict mana.
func (tb *busterTable) castFree(id uuid.UUID, targets ...game.TargetRef) error {
	return tb.g.CastSpell(tb.me.ID, id, game.CastSpellParams{
		Strict:          true,
		Targets:         targets,
		AlternativeCost: game.AltCostKeyGrantedCost,
	})
}

func (tb *busterTable) permissionOn(id uuid.UUID) *game.CastPermission {
	var p *game.CastPermission
	tb.g.ReadSnapshot(func() { p = tb.g.CastPermissionOnCardByIDForEffect(id) })
	return p
}

func TestBusterSwordPickedInstantIsCastWithItsTargetAndResolves(t *testing.T) {
	tb := newBusterTable(t, false)
	if tb.g.Turn.Step != game.StepCombatDamage {
		t.Fatalf("step %s, want the combat damage step", tb.g.Turn.Step)
	}
	// Mana value 1 is within the 5 damage dealt, and the hand also holds
	// Fodder ({6}), which must not be offered.
	p := latestChoiceOfKindFor(tb.g, game.PendingChoiceChooseCards, tb.me.ID)
	if p == nil || hasID(p.ChooseCards, tb.fodder) {
		t.Fatalf("the pick offers a card above the damage: %v", p)
	}
	tb.pick(t, tb.bolt)

	if err := tb.castFree(tb.bolt, game.TargetRef{Kind: game.TargetPlayer, ID: tb.opp.ID}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	if tb.me.Hand.Contains(tb.bolt) {
		t.Fatal("the spell did not leave the hand")
	}
	passPriorityAroundTable(t, tb.g)
	if tb.opp.Life != 32 {
		t.Errorf("opponent life %d, want 32 (5 combat + 3 from the free Bolt)", tb.opp.Life)
	}
	if tb.permissionOn(tb.bolt) != nil {
		t.Error("the grant outlived the cast")
	}
}

func TestBusterSwordPickedSorceryIsCastInTheCombatDamageStep(t *testing.T) {
	tb := newBusterTable(t, false)
	// Control: a sorcery nobody picked is refused here on its own timing.
	err := tb.g.CastSpell(tb.me.ID, tb.mindRot, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: tb.opp.ID}},
	})
	if err == nil {
		t.Fatal("an unpicked sorcery was castable in the combat damage step")
	}
	tb.pick(t, tb.mindRot)
	if err := tb.castFree(tb.mindRot, game.TargetRef{Kind: game.TargetPlayer, ID: tb.opp.ID}); err != nil {
		t.Fatalf("a sorcery picked in the combat damage step must be castable: %v", err)
	}
	passPriorityAroundTable(t, tb.g)
	// Mind Rot resolved: the target now owes its two discards.
	owed := false
	for _, c := range tb.g.PendingChoices {
		if c != nil && c.Chooser == tb.opp.ID && c.Source == tb.mindRot {
			owed = true
		}
	}
	if !owed {
		t.Error("Mind Rot did not resolve: its target owes no discard")
	}
}

func TestBusterSwordPickingNoneOpensNothing(t *testing.T) {
	tb := newBusterTable(t, false)
	tb.pick(t, uuid.Nil)
	for _, id := range []uuid.UUID{tb.bolt, tb.mindRot} {
		if tb.permissionOn(id) != nil {
			t.Errorf("card %s holds a cast permission after picking nothing", id)
		}
	}
	err := tb.castFree(tb.bolt, game.TargetRef{Kind: game.TargetPlayer, ID: tb.opp.ID})
	if err == nil {
		t.Fatal("the free cast opened without a pick")
	}
}

func TestBusterSwordUncastablePickStaysInHand(t *testing.T) {
	// Doom Blade needs a nonblack creature; the only one on the table is
	// black, so the pick can not be cast.
	tb := newBusterTable(t, true)
	tb.pick(t, tb.doom)
	err := tb.castFree(tb.doom)
	if err == nil {
		t.Fatal("Doom Blade was cast with no legal target")
	}
	if !tb.me.Hand.Contains(tb.doom) {
		t.Fatal("the uncastable pick left the hand")
	}
	// And it can not be held: passing ends the chance.
	if err := tb.g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	if tb.permissionOn(tb.doom) != nil {
		t.Error("the grant survived the holder's pass")
	}
	if !tb.me.Hand.Contains(tb.doom) {
		t.Error("the lapse moved the card out of the hand")
	}
}

func TestBusterSwordPassingEndsTheChance(t *testing.T) {
	tb := newBusterTable(t, false)
	tb.pick(t, tb.bolt)
	if tb.permissionOn(tb.bolt) == nil {
		t.Fatal("the pick opened no cast")
	}
	if err := tb.g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	if tb.permissionOn(tb.bolt) != nil {
		t.Fatal("the free cast outlived the holder's pass")
	}
	if !tb.me.Hand.Contains(tb.bolt) {
		t.Fatal("passing moved the card")
	}
	err := tb.castFree(tb.bolt, game.TargetRef{Kind: game.TargetPlayer, ID: tb.opp.ID})
	if err == nil || strings.Contains(err.Error(), "illegal target") {
		t.Fatalf("a free cast after the pass must be refused for its price, got %v", err)
	}
}

// The other four offers share the helper, so one table proves each one wires
// it: the pick declares the cast, a sorcery picked in the combat damage step
// is castable free, and passing instead ends the chance with the card still
// in hand.
func TestHandFreeCastOffersOpenTheCastAtAnyTiming(t *testing.T) {
	offers := map[string]func(t *testing.T, g *game.Game, me, opp *game.Player){
		"Kari Zev's Expertise": func(t *testing.T, g *game.Game, me, _ *game.Player) {
			offerFrom(t, g, me, func(ctx *Context) error { return kariZevsExpertiseOfferFreeCast(ctx, me.ID) })
		},
		"Rishkar's Expertise": func(t *testing.T, g *game.Game, me, _ *game.Player) {
			offerFrom(t, g, me, func(ctx *Context) error { return rishkarsExpertiseOfferFreeCast(ctx, me.ID) })
		},
		"Surtland Elementalist": func(t *testing.T, g *game.Game, me, _ *game.Player) {
			offerFrom(t, g, me, func(ctx *Context) error { return surtlandElementalistOfferFreeCast(ctx, me.ID) })
		},
		"Great Intelligence's Plan": func(t *testing.T, g *game.Game, me, opp *game.Player) {
			villainFillLibrary(g, me, 6)
			villainCastPlan(t, g, opp)
			villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)
		},
	}
	for name, offer := range offers {
		for _, cast := range []bool{true, false} {
			label := name + "/pass"
			if cast {
				label = name + "/cast"
			}
			t.Run(label, func(t *testing.T) {
				g := newCatalogGame(t)
				me, opp := g.Seats[0], g.Seats[1]
				g.WithWriteLock(func() { me.Hand.Cards = nil })
				rot := handCardFull(me, "Mind Rot", "Sorcery", "{B}", mindRotOracle, []string{"B"})
				// Great Intelligence's Plan is a sorcery, so its own
				// resolution is in a main phase; the others are offered
				// in the combat damage step, where a sorcery is the
				// point.
				if name != "Great Intelligence's Plan" {
					bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
					attackWith(t, g, opp.ID, bear)
				}
				offer(t, g, me, opp)

				pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
				if pick == nil {
					t.Fatal("no free-cast pick")
				}
				if pick.ChooseDestination != game.ChooseCastFree {
					t.Fatalf("destination %q, want %q", pick.ChooseDestination, game.ChooseCastFree)
				}
				if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{rot}); err != nil {
					t.Fatalf("pick: %v", err)
				}
				params := game.CastSpellParams{
					Strict:          true,
					Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
					AlternativeCost: game.AltCostKeyGrantedCost,
				}
				if !cast {
					if err := g.PassPriority(); err != nil {
						t.Fatalf("pass: %v", err)
					}
					if !me.Hand.Contains(rot) {
						t.Fatal("passing moved the card out of the hand")
					}
					if err := g.CastSpell(me.ID, rot, params); err == nil {
						t.Fatal("the free cast outlived the pass")
					}
					return
				}
				if err := g.CastSpell(me.ID, rot, params); err != nil {
					t.Fatalf("a sorcery picked in the combat damage step must be castable: %v", err)
				}
				if me.Hand.Contains(rot) {
					t.Fatal("the spell did not leave the hand")
				}
			})
		}
	}
}
