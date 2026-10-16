package game

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// autotap_restricted_test.go — #3874, ADR 0040 amendment 2026-10-16:
// auto-pay plans a mana ability whose output is restricted when the
// spell being cast is one the restriction admits, and for no other
// payment. The engine half: the candidate rule, its ranking and the
// payability bound. The printed cards (Cavern of Souls, Delighted
// Halfling) are pinned in cards/effects/autotap_restricted_test.go and
// the enumerator's half in legal/restricted_autopay_test.go.

// cantBeCountered is Cavern of Souls' and Delighted Halfling's rider.
func uncounterableRider() ManaSpendRider {
	return ManaSpendRider{Kind: ManaRiderCantBeCountered, When: []string{ManaRestrictCast}}
}

// halflingShape is Delighted Halfling's two rows: a plain {C}, and any
// colour for a legendary spell, which can't be countered.
func halflingShape() []ManaAbilityShape {
	return []ManaAbilityShape{
		{TapCost: true, Produced: "{C}", Label: "Add {C}"},
		{TapCost: true, Produced: "{W|U|B|R|G}", Label: "Add one mana of any color (legendary spells only)",
			Restrictions: []string{ManaRestrictCast, ManaRestrictSupertype("Legendary")},
			SpendRiders:  []ManaSpendRider{uncounterableRider()}},
	}
}

// cavernShape is Cavern of Souls named for `tribe` (empty: nothing named
// yet), its restriction computed at activation as the card's is.
func cavernShape(tribe string) []ManaAbilityShape {
	return []ManaAbilityShape{
		{TapCost: true, Produced: "{C}", Label: "Add {C}"},
		{TapCost: true, Produced: "{W|U|B|R|G}", Label: "Add one mana of any color (chosen-type creature spells only)",
			RestrictionsFunc: func(*Game, uuid.UUID, uuid.UUID) []string {
				return []string{ManaRestrictCast, ManaRestrictType("Creature"), ManaRestrictSubtype(tribe)}
			},
			SpendRiders: []ManaSpendRider{uncounterableRider()}},
	}
}

// zigguratShape is Ancient Ziggurat: any colour, creature spells only,
// no rider.
func zigguratShape() []ManaAbilityShape {
	return []ManaAbilityShape{{TapCost: true, Produced: "{W|U|B|R|G}",
		Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")}}}
}

// jeganthaShape is Jegantha's symbol-level clause: never plannable.
func jeganthaShape() []ManaAbilityShape {
	return []ManaAbilityShape{{TapCost: true, Produced: "{W}{U}{B}{R}{G}",
		Restrictions: []string{ManaRestrictNoGeneric}}}
}

// activateOnlyShape is mana for activations only (a Signet-fed Eldrazi
// Temple half in the abstract): an activation admits it, and the planner
// still never takes it, because it plans restricted mana for casts only.
func activateOnlyShape() []ManaAbilityShape {
	return []ManaAbilityShape{{TapCost: true, Produced: "{G}", Restrictions: []string{ManaRestrictActivate}}}
}

func restrictedBoard(t *testing.T) (*Game, *Player) {
	t.Helper()
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	return g, g.Seats[0]
}

// tryAutoCast is autoCast returning the error.
func tryAutoCast(g *Game, p *Player, name, typeLine, cost string) (uuid.UUID, error) {
	c := NewCard(name, p.ID)
	c.TypeLine, c.ManaCost, c.Controller = typeLine, cost, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID, g.CastSpell(p.ID, c.InstanceID, CastSpellParams{Strict: true, AutoTap: true})
}

// The headline: the restricted row is a source for the spell it admits,
// and for no other. A Halfling is the only source, so whether the cast
// goes through IS whether the planner took its legendary-only row.
func TestAutoPayPlansRestrictedManaOnlyForASpellItAdmits(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		want           bool
	}{
		{"a legendary creature spell", "Legendary Creature — Elf Druid", true},
		{"a legendary sorcery", "Legendary Sorcery", true},
		{"a nonlegendary creature spell", "Creature — Elf Druid", false},
		{"an instant", "Instant", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me := restrictedBoard(t)
			halfling := pushIntrinsicPermanent(g, me, "Halfling", "Creature — Halfling Citizen", halflingShape(), nil)
			id, err := tryAutoCast(g, me, "Spell", tc.typeLine, "{G}")
			if !tc.want {
				var short *InsufficientManaError
				if !errors.As(err, &short) {
					t.Fatalf("cast = %v, want insufficient mana: the restriction refuses this spell", err)
				}
				if tappedOnBattlefield(t, g, halfling) {
					t.Error("a refused cast tapped the Halfling")
				}
				return
			}
			if err != nil {
				t.Fatalf("cast: %v", err)
			}
			if !tappedOnBattlefield(t, g, halfling) {
				t.Error("the Halfling was not tapped")
			}
			item := g.StackMeta[id]
			if item == nil || !item.SpellCantBeCounteredByMana() {
				t.Errorf("the legendary spell does not record the rider: %+v", item)
			}
			if me.ManaPool.Total() != 0 {
				t.Errorf("mana left floating: %+v", me.ManaPool)
			}
		})
	}
}

// Cavern of Souls: the chosen type, read at planning time exactly as the
// mint reads it. A Cavern with nothing named admits nothing.
func TestAutoPayPlansCavernForACreatureOfTheChosenType(t *testing.T) {
	for _, tc := range []struct {
		name, tribe, typeLine string
		want                  bool
	}{
		{"an Elf, Elf named", "Elf", "Creature — Elf Warrior", true},
		{"a Goblin, Elf named", "Elf", "Creature — Goblin", false},
		{"an Elf sorcery is not a creature spell", "Elf", "Kindred Sorcery — Elf", false},
		{"an Elf, nothing named", "", "Creature — Elf Warrior", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me := restrictedBoard(t)
			cavern := pushIntrinsicPermanent(g, me, "Cavern", "Land", cavernShape(tc.tribe), nil)
			id, err := tryAutoCast(g, me, "Spell", tc.typeLine, "{G}")
			if (err == nil) != tc.want {
				t.Fatalf("cast err = %v, want success %v", err, tc.want)
			}
			if !tc.want {
				if tappedOnBattlefield(t, g, cavern) {
					t.Error("a refused cast tapped the Cavern")
				}
				return
			}
			if !g.StackMeta[id].SpellCantBeCounteredByMana() {
				t.Error("the Elf cast with Cavern mana can be countered")
			}
		})
	}
}

// The rider that fires is preferred over a basic that pays the same pip,
// whichever arrived first; for a spell the restriction refuses, the basic
// pays and the Halfling stays untapped.
func TestAutoPayPrefersTheRestrictedRowWhoseRiderFires(t *testing.T) {
	for _, halflingFirst := range []bool{true, false} {
		for _, tc := range []struct {
			name, typeLine string
			wantHalfling   bool
		}{
			{"legendary", "Legendary Creature — Elf", true},
			{"nonlegendary", "Creature — Elf", false},
		} {
			t.Run(fmt.Sprintf("%s/halfling first %v", tc.name, halflingFirst), func(t *testing.T) {
				g, me := restrictedBoard(t)
				var halfling, forest uuid.UUID
				if halflingFirst {
					halfling = pushIntrinsicPermanent(g, me, "Halfling", "Creature — Halfling", halflingShape(), nil)
					forest = pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
				} else {
					forest = pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
					halfling = pushIntrinsicPermanent(g, me, "Halfling", "Creature — Halfling", halflingShape(), nil)
				}
				item := autoCast(t, g, me, "Spell", tc.typeLine, "{G}", "")
				if got := tappedOnBattlefield(t, g, halfling); got != tc.wantHalfling {
					t.Errorf("Halfling tapped = %v, want %v", got, tc.wantHalfling)
				}
				if got := tappedOnBattlefield(t, g, forest); got == tc.wantHalfling {
					t.Errorf("Forest tapped = %v, want %v", got, !tc.wantHalfling)
				}
				if item.SpellCantBeCounteredByMana() != tc.wantHalfling {
					t.Errorf("uncounterable = %v, want %v", item.SpellCantBeCounteredByMana(), tc.wantHalfling)
				}
			})
		}
	}
}

// Restricted mana with no rider is spent on a spell it admits before
// mana any spell could use (the pool's restricted-first rule, ADR 0040
// §4), in the coloured pass and the generic recruit alike.
func TestAutoPaySpendsAdmittedRestrictedManaBeforeOrdinaryMana(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost string
		wantZiggurat         bool
	}{
		{"a creature's coloured pip", "Creature — Bear", "{G}", true},
		{"a creature's generic pip", "Creature — Bear", "{1}", true},
		{"an instant's pip", "Instant", "{G}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me := restrictedBoard(t)
			forest := pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
			zig := pushIntrinsicPermanent(g, me, "Ziggurat", "Land", zigguratShape(), nil)
			autoCast(t, g, me, "Spell", tc.typeLine, tc.cost, "")
			if got := tappedOnBattlefield(t, g, zig); got != tc.wantZiggurat {
				t.Errorf("Ziggurat tapped = %v, want %v", got, tc.wantZiggurat)
			}
			if got := tappedOnBattlefield(t, g, forest); got == tc.wantZiggurat {
				t.Errorf("Forest tapped = %v, want %v", got, !tc.wantZiggurat)
			}
		})
	}
}

// No cast, no restricted plan: a payment with no stated object, an
// activation (even one the restriction admits), and the planner's own
// context-free entry points all leave restricted output to the click.
// So does a symbol-level clause, for a cast too.
func TestRestrictedManaIsPlannedForACastAndNothingElse(t *testing.T) {
	g, me := restrictedBoard(t)
	pushIntrinsicPermanent(g, me, "Halfling", "Creature — Halfling", halflingShape()[1:], nil)
	pushIntrinsicPermanent(g, me, "Activator", "Land", activateOnlyShape(), nil)
	legendary := NewCard("Legend", me.ID)
	legendary.TypeLine, legendary.ManaCost = "Legendary Creature — Elf", "{G}"
	source := NewCard("Relic", me.ID)
	source.TypeLine = "Artifact"
	green := costFor(t, "{G}")
	g.WithWriteLock(func() {
		if _, _, ok := g.autoTapTopUpLocked(me.ID, green, 0, ManaSpendForCast(legendary), nil, 0); !ok {
			t.Error("a legendary cast found no plan from the Halfling")
		}
		if _, _, ok := g.autoTapTopUpLocked(me.ID, green, 0, ManaSpendContext{}, nil, 0); ok {
			t.Error("a payment with no object planned restricted mana")
		}
		if _, _, ok := g.autoTapTopUpLocked(me.ID, green, 0, ManaSpendForAbility(source), nil, 0); ok {
			t.Error("an activation planned restricted mana (the activate-only land admits it, the planner must not take it)")
		}
	})
	if _, ok := g.AutoTapForCost(me.ID, green, 0); ok {
		t.Error("AutoTapForCost planned restricted mana")
	}

	g2, me2 := restrictedBoard(t)
	pushIntrinsicPermanent(g2, me2, "Jegantha", "Creature — Elemental Elk", jeganthaShape(), nil)
	if _, err := tryAutoCast(g2, me2, "Bear", "Creature — Bear", "{G}"); err == nil {
		t.Error("a can't-pay-generic source was planned: the planner cannot tell which symbol it pays")
	}
}

// The costed planner (#2455) never books restricted mana, not even for a
// cast it admits: it may fund a Signet's {1} with it, which the
// restriction forbids. The ordinary plan still takes the Halfling.
func TestTheCostedPlannerLeavesRestrictedManaOut(t *testing.T) {
	g, me := restrictedBoard(t)
	pushIntrinsicPermanent(g, me, "Halfling", "Creature — Halfling", halflingShape()[1:], nil)
	pushIntrinsicPermanent(g, me, "Signet", "Artifact",
		[]ManaAbilityShape{{TapCost: true, ManaCost: "{1}", Produced: "{G}{U}"}}, nil)
	legendary := NewCard("Legend", me.ID)
	legendary.TypeLine = "Legendary Creature — Elf"
	g.WithWriteLock(func() {
		wish := g.autoTapWishFor(me.ID, ManaSpendForCast(legendary), 0)
		// {G}{U} needs the Signet, and the Signet's {1} only the
		// Halfling could fund.
		if _, ok := g.autoTapCostedLocked(me, costFor(t, "{G}{U}"), 0, nil, wish); ok {
			t.Error("the costed planner funded a Signet with legendary-only mana")
		}
	})
}

// The payability bound (#3874). The candidate set now depends on the
// spell, so the answers that must agree are spelled out and checked over
// every board, cost and spell:
//
//  1. the enumerator's probe (the planner under the cast's context) and
//     the cast gate (CastSpell with auto-pay, on a clone) give the same
//     answer — a cast is offered exactly when it goes through;
//  2. adding candidates never takes a payment away: payable with no
//     context implies payable with the cast's;
//  3. for a spell no restricted row on the board admits, the answer is
//     exactly the context-free one: the restriction refuses it, so the
//     row is not a candidate.
func TestRestrictedCandidatesKeepThePayabilityBound(t *testing.T) {
	type row struct {
		name   string
		shape  func() []ManaAbilityShape
		admits func(c Card) bool
	}
	legendaryOnly := row{"Halfling", halflingShape, func(c Card) bool { return c.IsLegendary() }}
	elfCavern := row{"Cavern", func() []ManaAbilityShape { return cavernShape("Elf") }, func(c Card) bool {
		return c.IsCreature() && c.HasSubtype("Elf")
	}}
	ziggurat := row{"Ziggurat", zigguratShape, func(c Card) bool { return c.IsCreature() }}
	jegantha := row{"Jegantha", jeganthaShape, func(Card) bool { return false }}
	activateOnly := row{"Activator", activateOnlyShape, func(Card) bool { return false }}

	boards := map[string]struct {
		restricted []row
		basics     []string
	}{
		"halfling and a swamp":         {[]row{legendaryOnly}, []string{"Swamp"}},
		"cavern and a mountain":        {[]row{elfCavern}, []string{"Mountain"}},
		"ziggurat, cavern and forest":  {[]row{ziggurat, elfCavern}, []string{"Forest"}},
		"jegantha and a forest":        {[]row{jegantha}, []string{"Forest"}},
		"activate-only and two swamps": {[]row{activateOnly}, []string{"Swamp", "Swamp"}},
		"everything":                   {[]row{legendaryOnly, elfCavern, ziggurat, jegantha, activateOnly}, []string{"Island"}},
	}
	basicType := map[string]string{
		"Swamp": "Basic Land — Swamp", "Mountain": "Basic Land — Mountain",
		"Forest": "Basic Land — Forest", "Island": "Basic Land — Island",
	}
	costs := []string{"{G}", "{1}", "{2}", "{G}{G}", "{1}{G}", "{B}", "{R}{R}", "{2}{G}", "{U}{G}", "{4}", "{W}{U}{B}"}
	spells := []struct{ name, typeLine string }{
		{"Legendary Elf", "Legendary Creature — Elf Druid"},
		{"Elf", "Creature — Elf Warrior"},
		{"Goblin", "Creature — Goblin"},
		{"Legendary Sorcery", "Legendary Sorcery"},
		{"Instant", "Instant"},
	}
	for bname, board := range boards {
		for _, cost := range costs {
			for _, sp := range spells {
				name := fmt.Sprintf("%s: %s %s", bname, sp.name, cost)
				g, me := restrictedBoard(t)
				for _, r := range board.restricted {
					pushIntrinsicPermanent(g, me, r.name, "Land", r.shape(), nil)
				}
				for _, b := range board.basics {
					pushBattlefieldForTest(g, me.ID, b, basicType[b], "")
				}
				card := NewCard(sp.name, me.ID)
				card.TypeLine, card.ManaCost, card.Controller = sp.typeLine, cost, me.ID
				me.Hand.PushTop(card)
				parsed := costFor(t, cost)

				var probe, bare bool
				g.WithWriteLock(func() {
					probe = g.AutoTapTopUpForEffectExcluding(me.ID, parsed, 0, ManaSpendForCastFrom(card, ZoneHand), nil)
					_, _, bare = g.autoTapTopUpLocked(me.ID, parsed, 0, ManaSpendContext{}, nil, 0)
				})
				clone := g.Clone()
				gate := clone.CastSpell(me.ID, card.InstanceID, CastSpellParams{Strict: true, AutoTap: true, FromZone: "hand"})
				if probe != (gate == nil) {
					t.Errorf("%s: the enumerator's probe says %v, the cast gate says %v", name, probe, gate)
				}
				if bare && !probe {
					t.Errorf("%s: payable with no context and not for the cast", name)
				}
				admitted := false
				for _, r := range board.restricted {
					admitted = admitted || r.admits(card)
				}
				if !admitted && probe != bare {
					t.Errorf("%s: no restricted row admits this spell, yet payable %v (context-free %v)", name, probe, bare)
				}
			}
		}
	}
}
