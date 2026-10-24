package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3270: "you may cast a spell from your hand without paying its mana cost"
// is cast WHILE the ability resolves (CR 608.2g), ignoring timing. These
// tests drive Buster Sword through a real combat to the combat damage step,
// where a sorcery is not castable on its own.
//
// #4250: the cast is the resolution_cast prompt's answer, made inside the
// trigger's resolution: one permission over every card the offer names,
// spent by the one cast, and a decline that leaves the cards in hand.

const mindRotOracle = "ad44cf74-b717-48fb-9fa2-77512024d76a"

// busterTable is a combat with Buster Sword's carrier connecting for 5: the
// game sits in the combat damage step with the draw already resolved and the
// free-cast offer pending. `black` makes the carrier black, so Doom Blade
// has no legal target on the table.
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

// offer is the free-cast offer waiting inside the trigger's resolution.
func (tb *busterTable) offer(t *testing.T) *game.PendingChoice {
	t.Helper()
	return mustResolutionCastOffer(t, tb.g, tb.me.ID, game.MayCastKeywordHandCast)
}

// castFree is the cast_spell the sheet sends: strict mana, an empty pool.
// It names no claim: the prompt holds a hand card's cast to the granted
// offer (#4250).
func (tb *busterTable) castFree(id uuid.UUID, targets ...game.TargetRef) error {
	return tb.g.CastSpell(tb.me.ID, id, game.CastSpellParams{
		Strict:  true,
		Targets: targets,
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
	p := tb.offer(t)
	if hasID(p.DiscardOptions, tb.fodder) || !hasID(p.DiscardOptions, tb.bolt) {
		t.Fatalf("the offer is %v, want every spell within the damage and not Fodder", p.DiscardOptions)
	}
	if err := tb.castFree(tb.bolt, game.TargetRef{Kind: game.TargetPlayer, ID: tb.opp.ID}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	if tb.me.Hand.Contains(tb.bolt) || !tb.g.Stack.Contains(tb.bolt) {
		t.Fatal("the spell did not go from the hand to the stack")
	}
	// One spell: the cast spent the offer, and the resolution let go.
	if resolutionCastOfferFor(tb.g, tb.me.ID) != nil || tb.g.ResolutionPaused() {
		t.Fatal("the offer outlived its one cast")
	}
	if tb.permissionOn(tb.mindRot) != nil {
		t.Error("the other candidates kept the free cast")
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
	// Control: a card the offer does not name is refused.
	if err := tb.castFree(tb.fodder); err == nil {
		t.Fatal("a card above the damage was cast during the offer")
	}
	tb.offer(t)
	if err := tb.castFree(tb.mindRot, game.TargetRef{Kind: game.TargetPlayer, ID: tb.opp.ID}); err != nil {
		t.Fatalf("a sorcery offered in the combat damage step must be castable: %v", err)
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
	tb.offer(t)
	declineResolutionCast(t, tb.g, tb.me.ID)
	for _, id := range []uuid.UUID{tb.bolt, tb.mindRot} {
		if tb.permissionOn(id) != nil {
			t.Errorf("card %s holds a cast permission after the decline", id)
		}
		if !tb.me.Hand.Contains(id) {
			t.Errorf("card %s left the hand on the decline", id)
		}
	}
	if tb.g.ResolutionPaused() {
		t.Error("the decline did not let the resolution finish")
	}
	err := tb.g.CastSpell(tb.me.ID, tb.bolt, game.CastSpellParams{
		Strict:          true,
		Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: tb.opp.ID}},
		AlternativeCost: game.AltCostKeyGrantedCost,
	})
	if err == nil {
		t.Fatal("the free cast opened after the decline")
	}
}

func TestBusterSwordUncastablePickStaysInHand(t *testing.T) {
	// Doom Blade needs a nonblack creature; the only one on the table is
	// black, so it can not be cast.
	tb := newBusterTable(t, true)
	tb.offer(t)
	if err := tb.castFree(tb.doom); err == nil {
		t.Fatal("Doom Blade was cast with no legal target")
	}
	if !tb.me.Hand.Contains(tb.doom) {
		t.Fatal("the uncastable card left the hand")
	}
	// The refused cast leaves the offer open for another card or a decline.
	tb.offer(t)
	declineResolutionCast(t, tb.g, tb.me.ID)
	if tb.permissionOn(tb.doom) != nil {
		t.Error("the grant survived the decline")
	}
	if !tb.me.Hand.Contains(tb.doom) {
		t.Error("the decline moved the card out of the hand")
	}
}

func TestBusterSwordPassingEndsTheChance(t *testing.T) {
	tb := newBusterTable(t, false)
	tb.offer(t)
	if tb.permissionOn(tb.bolt) == nil {
		t.Fatal("the offer opened no cast")
	}
	// Nobody receives priority while the offer is open (CR 608.2g).
	if err := tb.g.PassPriority(); err == nil {
		t.Fatal("priority was passed while the free cast was being offered")
	}
	declineResolutionCast(t, tb.g, tb.me.ID)
	if tb.permissionOn(tb.bolt) != nil {
		t.Fatal("the free cast outlived the decline")
	}
	if !tb.me.Hand.Contains(tb.bolt) {
		t.Fatal("declining moved the card")
	}
	err := tb.g.CastSpell(tb.me.ID, tb.bolt, game.CastSpellParams{
		Strict:          true,
		Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: tb.opp.ID}},
		AlternativeCost: game.AltCostKeyGrantedCost,
	})
	if err == nil {
		t.Fatal("a free cast after the decline must be refused")
	}
}

// The other four offers share the helper, so one table proves each one wires
// it: the offer is the resolution_cast, a sorcery offered in the combat
// damage step is castable free, and declining instead ends the chance with
// the card still in hand. Kari Zev's, Rishkar's and Surtland's run here
// outside a resolution, through the keyed pick (#4246) that is still the
// offer there; Great Intelligence's Plan resolves for real.
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
			label := name + "/decline"
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
				params := game.CastSpellParams{
					Strict:          true,
					Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
					AlternativeCost: game.AltCostKeyGrantedCost,
				}
				if resolutionCastOfferFor(g, me.ID) != nil {
					mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, rot)
					if !cast {
						declineResolutionCast(t, g, me.ID)
						if !me.Hand.Contains(rot) {
							t.Fatal("declining moved the card out of the hand")
						}
						if err := g.CastSpell(me.ID, rot, params); err == nil {
							t.Fatal("the free cast outlived the decline")
						}
						return
					}
				} else {
					pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
					if pick == nil {
						t.Fatal("no free-cast offer")
					}
					if pick.ChooseDestination != game.ChooseCastFree {
						t.Fatalf("destination %q, want %q", pick.ChooseDestination, game.ChooseCastFree)
					}
					if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{rot}); err != nil {
						t.Fatalf("pick: %v", err)
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
				}
				if err := g.CastSpell(me.ID, rot, params); err != nil {
					t.Fatalf("a sorcery offered in the combat damage step must be castable: %v", err)
				}
				if me.Hand.Contains(rot) {
					t.Fatal("the spell did not leave the hand")
				}
			})
		}
	}
}
