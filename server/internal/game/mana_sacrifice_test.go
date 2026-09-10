package game

import (
	"testing"

	"github.com/google/uuid"
)

// mana_sacrifice_test.go — the S21 mana-cost pass: sacrifice-ANOTHER
// as a mana ability's cost (Ashnod's Altar), and the CR 302.1
// summoning-sickness check that tap-for-mana abilities were missing.

// altarAbility is Ashnod's Altar's shape: no tap, sacrifice a creature
// you control, add {C}{C}.
func altarAbility() []ManaAbilityShape {
	return []ManaAbilityShape{{
		SacrificeOther: &TargetSpec{
			Mode:  "permanent",
			Label: "a creature",
			Zones: []ZoneKind{ZoneBattlefield},
			CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
				return c.IsCreature()
			},
			Min: 1,
			Max: 1,
		},
		Produced: "{C}{C}",
		Label:    "Sacrifice a creature: Add {C}{C}",
	}}
}

func TestManaAbilitySacrificesAnotherCreatureForMana(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	altar := pushIntrinsicPermanent(g, me, "Ashnod's Altar", "Artifact", altarAbility(), nil)
	food := pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)

	if err := g.ActivateManaAbility(me.ID, altar, 0, ManaAbilityParams{
		SacrificeIDs: []uuid.UUID{food},
	}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if g.Battlefield.Contains(food) {
		t.Error("the creature was not sacrificed")
	}
	if !me.Graveyard.Contains(food) {
		t.Error("the sacrificed creature should be in its owner's graveyard")
	}
	// The Altar itself survives — it eats another creature, not itself.
	if !g.Battlefield.Contains(altar) {
		t.Error("the Altar sacrificed itself")
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("mana pool = %d, want 2", len(me.ManaPool))
	}
	if !hasEvent(g, EventSacrifice, food) {
		t.Error("no EventSacrifice for the eaten creature")
	}
}

// CR 605.3a: a mana ability resolves immediately, so the mana must be
// in the pool by the time any dies-trigger the sacrifice queued gets
// to resolve. That ordering is what makes an Altar plus a payoff an
// engine rather than a coin flip.
func TestManaAbilitySacrificeDrainsTriggersAfterManaLands(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	altar := pushIntrinsicPermanent(g, me, "Ashnod's Altar", "Artifact", altarAbility(), nil)
	food := pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)

	if err := g.ActivateManaAbility(me.ID, altar, 0, ManaAbilityParams{
		SacrificeIDs: []uuid.UUID{food},
	}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Fatalf("mana pool = %d, want 2 — mana must land even with triggers queued", len(me.ManaPool))
	}
}

// Every cost is validated before any is paid: a clause the chosen
// permanent doesn't satisfy must not eat it.
func TestManaAbilitySacrificeRejectsIllegalChoices(t *testing.T) {
	for _, tc := range []struct {
		name string
		// setup returns the altar and the id to name as the sacrifice.
		setup func(t *testing.T, g *Game) (uuid.UUID, uuid.UUID)
	}{
		{"not a creature", func(t *testing.T, g *Game) (uuid.UUID, uuid.UUID) {
			me := g.Seats[0]
			altar := pushIntrinsicPermanent(g, me, "Ashnod's Altar", "Artifact", altarAbility(), nil)
			rock := pushIntrinsicPermanent(g, me, "Sol Ring", "Artifact", nil, nil)
			return altar, rock
		}},
		{"opponent's creature", func(t *testing.T, g *Game) (uuid.UUID, uuid.UUID) {
			me, opp := g.Seats[0], g.Seats[1]
			altar := pushIntrinsicPermanent(g, me, "Ashnod's Altar", "Artifact", altarAbility(), nil)
			theirs := pushIntrinsicPermanent(g, opp, "Bear", "Creature — Bear", nil, nil)
			return altar, theirs
		}},
		{"nonexistent card", func(t *testing.T, g *Game) (uuid.UUID, uuid.UUID) {
			me := g.Seats[0]
			altar := pushIntrinsicPermanent(g, me, "Ashnod's Altar", "Artifact", altarAbility(), nil)
			return altar, uuid.New()
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			altar, chosen := tc.setup(t, g)

			err := g.ActivateManaAbility(me.ID, altar, 0, ManaAbilityParams{
				SacrificeIDs: []uuid.UUID{chosen},
			})
			if err == nil {
				t.Fatal("illegal sacrifice choice was accepted")
			}
			if len(me.ManaPool) != 0 {
				t.Errorf("mana was produced for an illegal cost: pool %d", len(me.ManaPool))
			}
			if chosen != uuid.Nil && g.Battlefield.Contains(chosen) == false {
				// Only meaningful for the two cases where the card exists.
				if tc.name != "nonexistent card" {
					t.Error("the named permanent was sacrificed despite an illegal cost")
				}
			}
		})
	}
}

// A sacrifice-another cost needs exactly one choice — none, or two,
// is not a legal payment.
func TestManaAbilitySacrificeNeedsExactlyOneChoice(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	altar := pushIntrinsicPermanent(g, me, "Ashnod's Altar", "Artifact", altarAbility(), nil)
	a := pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)
	b := pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)

	for _, tc := range []struct {
		name string
		ids  []uuid.UUID
	}{
		{"none", nil},
		{"two", []uuid.UUID{a, b}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := g.ActivateManaAbility(me.ID, altar, 0, ManaAbilityParams{
				SacrificeIDs: tc.ids,
			}); err == nil {
				t.Error("expected an error")
			}
		})
	}
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("a rejected payment ate a creature anyway")
	}
}

// An ability with no sacrifice component must reject a stray choice
// rather than silently eating the named permanent.
func TestManaAbilityRejectsUnwantedSacrificeIDs(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	rock := pushIntrinsicPermanent(g, me, "Sol Ring", "Artifact",
		[]ManaAbilityShape{{TapCost: true, Produced: "{C}{C}", Label: "Add {C}{C}"}}, nil)
	bear := pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)

	if err := g.ActivateManaAbility(me.ID, rock, 0, ManaAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
	}); err == nil {
		t.Error("a stray sacrifice choice should be rejected")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the stray choice was sacrificed")
	}
}

// --- CR 302.1 on tap-for-mana abilities -------------------------

// Before this pass, Birds of Paradise / Llanowar Merfolk / Palladium
// Myr all tapped for mana the turn they landed. ActivateCatalogAbility
// had the check; ActivateManaAbility didn't.
func TestManaAbilityRespectsSummoningSickness(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushIntrinsicPermanent(g, me, "Birds of Paradise", "Creature — Bird",
		[]ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}}, nil)
	setSummonedThisTurn(g, id, true)

	if err := g.ActivateManaAbility(me.ID, id, 0, ManaAbilityParams{}); err != ErrSummoningSick {
		t.Errorf("err = %v, want ErrSummoningSick", err)
	}
	if len(me.ManaPool) != 0 {
		t.Error("a summoning-sick creature produced mana")
	}
	if cardTapped(g, id) {
		t.Error("a rejected activation tapped the creature anyway")
	}
}

// Haste exempts it, as everywhere else.
func TestManaAbilityHasteBeatsSummoningSickness(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushIntrinsicPermanent(g, me, "Hasty Bird", "Creature — Bird",
		[]ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}},
		[]string{"haste"})
	setSummonedThisTurn(g, id, true)

	if err := g.ActivateManaAbility(me.ID, id, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("haste should allow it: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("mana pool = %d, want 1", len(me.ManaPool))
	}
}

// A NONCREATURE source is never summoning-sick, however new it is —
// Sol Ring taps the turn it lands.
func TestManaAbilityNoncreatureIgnoresSummoningSickness(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushIntrinsicPermanent(g, me, "Sol Ring", "Artifact",
		[]ManaAbilityShape{{TapCost: true, Produced: "{C}{C}", Label: "Add {C}{C}"}}, nil)
	setSummonedThisTurn(g, id, true)

	if err := g.ActivateManaAbility(me.ID, id, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("an artifact is never summoning-sick: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("mana pool = %d, want 2", len(me.ManaPool))
	}
}

func setSummonedThisTurn(g *Game, id uuid.UUID, v bool) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].SummonedThisTurn = v
			return
		}
	}
}

func cardTapped(g *Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	return false
}
