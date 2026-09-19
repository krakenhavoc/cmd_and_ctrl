package game

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// cast_price_test.go — #696. ONE pricer, and the table that holds it
// to that.
//
// The bug this file exists to prevent is not a wrong number, it is
// two numbers: the auto-tap preview endpoint parsed `card.ManaCost`
// and re-applied the commander tax and the cost modifiers itself, so
// it knew nothing about the alternative cost claimed at announce, a
// granted permission's flat override, the "spend mana as though any
// colour" fold or the mana half of an optional additional cost. The
// preview then disabled "Auto-tap & cast" on a cast the engine would
// have allowed (an airbent {5}{R}{R} priced at {2}), or planned taps
// for one it would refuse (a Think Twice whose flashback costs more
// than its printed cost).
//
// So every case below is priced THREE ways and all three must agree:
//
//	preview     g.PriceCast — what the lobby endpoint asks
//	enumerator  g.PriceCastForEffect's Base, then the shared cost
//	            modifier pass, which is exactly what legal/cast.go does
//	payment     a pool funded to the previewed price, cast in STRICT
//	            mode: it must succeed and leave the pool empty, and
//	            one mana short it must be refused
//
// The payment leg is the one that cannot be faked. A preview that
// agreed with itself and disagreed with CastSpell would pass the
// first two and fail the third.

// costKey renders a ParsedCost as a comparable string: the generic
// demand, the {X} slots and the colour requirements in sorted order.
// Sorted because the order of the coloured slots is an artefact of
// the cost string and never part of what is owed.
func costKey(c ParsedCost) string {
	opts := make([]string, 0, len(c.Required))
	for _, req := range c.Required {
		o := append([]string(nil), req.Options...)
		sort.Strings(o)
		opts = append(opts, strings.Join(o, "/"))
	}
	sort.Strings(opts)
	return fmt.Sprintf("generic=%d x=%d colors=[%s]", c.Generic, c.XSlots, strings.Join(opts, " "))
}

// fundExactly puts exactly `cost` into the player's pool: one token
// per coloured requirement (its first option, which is what a
// monocoloured symbol has anyway) and a colourless token per generic.
func fundExactly(p *Player, cost ParsedCost, xValue int) {
	for _, req := range cost.Required {
		colour := "C"
		if len(req.Options) > 0 {
			colour = req.Options[0]
		}
		p.ManaPool.AddMana(ManaToken{Color: colour})
	}
	for i := 0; i < cost.Generic+cost.XSlots*xValue; i++ {
		p.ManaPool.AddMana(ManaToken{Color: "C"})
	}
}

// enumeratorPrice is the legal/cast.go reading of the same price: the
// pre-modifier total out of the shared pricer, then the shared
// modifier pass. Kept here rather than imported because the legal
// package imports game and not the other way round; if the two ever
// diverge, this is the line that has to change with it.
func enumeratorPrice(t *testing.T, g *Game, playerID uuid.UUID, card Card, params CastSpellParams) ParsedCost {
	t.Helper()
	price, err := g.PriceCast(playerID, card, params)
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	zone, ok := castZoneFromWire(params.FromZone)
	if !ok {
		t.Fatalf("bad from_zone %q", params.FromZone)
	}
	cost, err := g.ApplyCostModifiers(price.Base, CostQuery{
		Card:       price.Card,
		Controller: playerID,
		FromZone:   zone,
		XValue:     params.XValue,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers: %v", err)
	}
	return cost
}

// castPriceCase is one announced cast, its board, and the price it
// owes.
type castPriceCase struct {
	name string
	// setup builds a fresh game, seeds the card and returns the
	// caster, the card's instance ID and the announcement.
	setup func(t *testing.T) (*Game, *Player, uuid.UUID, CastSpellParams)
	// want is the price, as costKey renders it.
	want string
}

func castPriceCases() []castPriceCase {
	const (
		looting = "test-price-flashback"
		kicker  = "test-price-kicker"
		sphere  = "test-price-sphere"
	)
	return []castPriceCase{
		{
			name: "printed cost from hand",
			setup: func(t *testing.T) (*Game, *Player, uuid.UUID, CastSpellParams) {
				g := newActiveGame(t)
				me := g.Seats[0]
				id := spellInHand(t, g, me, "Test Bear", "Creature — Bear", "{1}{G}")
				return g, me, id, CastSpellParams{}
			},
			want: "generic=1 x=0 colors=[G]",
		},
		{
			// CR 118.9. The preview used to price the {R} in the
			// corner and call a flashback cast affordable that was
			// not — Think Twice's {2}{U} against its printed {1}{U}.
			name: "alternative cost claimed from the graveyard",
			setup: func(t *testing.T) (*Game, *Player, uuid.UUID, CastSpellParams) {
				g := newActiveGame(t)
				me := g.Seats[0]
				withCatalogCastableZones(t, castableZonesFor(looting, ZoneGraveyard))
				withCatalogAlternativeCosts(t, altCostFor(looting, flashbackCost("{2}{R}")))
				id := looterInGraveyard(t, g, me, looting)
				return g, me, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}
			},
			want: "generic=2 x=0 colors=[R]",
		},
		{
			// ADR 0066. The grant's flat price belongs to the
			// INSTANCE, so no amount of reading the card can find it.
			name: "granted exile cast with a cost override",
			setup: func(t *testing.T) (*Game, *Player, uuid.UUID, CastSpellParams) {
				g := newActiveGame(t)
				me := g.Seats[0]
				toMainPhase(t, g)
				id := permanentFor(g, me, "Overpriced Bear", "Creature — Bear", "{5}{G}{G}")
				g.WithWriteLock(func() {
					if err := g.ExileCardWithPermissionForEffect(id, airbendGrant()); err != nil {
						t.Fatalf("airbend: %v", err)
					}
				})
				return g, me, id, CastSpellParams{FromZone: "exile"}
			},
			want: "generic=2 x=0 colors=[]",
		},
		{
			// S21 sub-PR 6: the colour demand folds into generic, so
			// a preview that kept the pips would report {G}{G}
			// missing on a board with two Mountains.
			name: "granted exile cast that spends mana as any colour",
			setup: func(t *testing.T) (*Game, *Player, uuid.UUID, CastSpellParams) {
				g := newActiveGame(t)
				me := g.Seats[0]
				toMainPhase(t, g)
				id := permanentFor(g, me, "Pirate Bear", "Creature — Bear", "{1}{G}{G}")
				g.WithWriteLock(func() {
					if err := g.ExileCardWithPermissionForEffect(id, CastPermission{
						CastOnly: true, Duration: WhileInZoneDuration(), AnyColor: true,
					}); err != nil {
						t.Fatalf("grant: %v", err)
					}
				})
				return g, me, id, CastSpellParams{FromZone: "exile"}
			},
			want: "generic=3 x=0 colors=[]",
		},
		{
			// ADR 0073 §3, CR 601.2f.
			name: "optional additional cost paid twice",
			setup: func(t *testing.T) (*Game, *Player, uuid.UUID, CastSpellParams) {
				g := newActiveGame(t)
				me := g.Seats[0]
				stubOptionalCosts(t, kicker, []AdditionalCost{{
					Optional: true, Key: "multikicker", Label: "Multikicker {1}{R}",
					ManaCost: "{1}{R}", Repeat: 3,
				}})
				id := spellInHand(t, g, me, "Test Wolfbriar", "Creature — Elemental", "{1}{G}")
				g.WithWriteLock(func() {
					for i := range me.Hand.Cards {
						if me.Hand.Cards[i].InstanceID == id {
							me.Hand.Cards[i].OracleID = kicker
						}
					}
				})
				return g, me, id, CastSpellParams{OptionalCosts: []int{0, 0}}
			},
			want: "generic=3 x=0 colors=[G R R]",
		},
		{
			// CR 903.8, layered on the cost being PAID.
			name: "commander tax on a third cast",
			setup: func(t *testing.T) (*Game, *Player, uuid.UUID, CastSpellParams) {
				g := newActiveGame(t)
				me := g.Seats[0]
				toMainPhase(t, g)
				c := NewCard("Test Commander", me.ID)
				c.TypeLine = "Legendary Creature — Bear"
				c.ManaCost = "{1}{G}"
				me.Command.PushTop(c)
				me.CommanderCasts[c.InstanceID] = 2
				return g, me, c.InstanceID, CastSpellParams{FromZone: "command"}
			},
			want: "generic=5 x=0 colors=[G]",
		},
		{
			// S28 / ADR 0048. The preview DID know about these, and
			// the point of the case is that routing through the one
			// pricer did not lose them.
			name: "board cost modifier on top of an alternative cost",
			setup: func(t *testing.T) (*Game, *Player, uuid.UUID, CastSpellParams) {
				g := newActiveGame(t)
				me := g.Seats[0]
				withCatalogCastableZones(t, castableZonesFor(looting, ZoneGraveyard))
				withCatalogAlternativeCosts(t, altCostFor(looting, flashbackCost("{2}{R}")))
				withCatalogCostModifiers(t, modifiersFor(sphere, CostModifier{
					Kind: CostIncrease, Label: "Spells cost {1} more", Amount: fixed(1),
				}))
				modifierSource(t, g, me, "Test Sphere", sphere)
				id := looterInGraveyard(t, g, me, looting)
				return g, me, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}
			},
			want: "generic=3 x=0 colors=[R]",
		},
	}
}

// The preview, the enumerator and the payment agree, case by case.
func TestOneCastPricerAgreesAcrossItsConsumers(t *testing.T) {
	for _, tc := range castPriceCases() {
		t.Run(tc.name, func(t *testing.T) {
			g, me, id, params := tc.setup(t)
			card, ok := g.LookupCardForEffect(id)
			if !ok {
				t.Fatalf("card %s not found", id)
			}

			price, err := g.PriceCast(me.ID, card, params)
			if err != nil {
				t.Fatalf("PriceCast: %v", err)
			}
			if got := costKey(price.Total); got != tc.want {
				t.Errorf("preview price = %s, want %s", got, tc.want)
			}
			if got := costKey(enumeratorPrice(t, g, me.ID, card, params)); got != tc.want {
				t.Errorf("enumerator price = %s, want %s", got, tc.want)
			}

			// The payment leg. Fund the pool to exactly the previewed
			// price and cast in strict mode: an engine charging
			// anything else either refuses or leaves change.
			strict := params
			strict.Strict = true
			fundExactly(me, price.Total, params.XValue)
			if err := g.CastSpell(me.ID, id, strict); err != nil {
				t.Fatalf("cast funded to exactly the previewed price: %v", err)
			}
			if len(me.ManaPool) != 0 {
				t.Errorf("pool after the cast = %+v, want empty — the payment charged less than the preview", me.ManaPool)
			}
		})
	}
}

// …and one mana short of the previewed price, the cast is refused.
// The other half of the same claim: a preview that under-priced would
// pass the test above by leaving the pool empty and still hand the
// player a cast the engine rejects.
func TestACastOneManaShortOfThePreviewIsRefused(t *testing.T) {
	for _, tc := range castPriceCases() {
		t.Run(tc.name, func(t *testing.T) {
			g, me, id, params := tc.setup(t)
			card, ok := g.LookupCardForEffect(id)
			if !ok {
				t.Fatalf("card %s not found", id)
			}
			price, err := g.PriceCast(me.ID, card, params)
			if err != nil {
				t.Fatalf("PriceCast: %v", err)
			}
			fundExactly(me, price.Total, params.XValue)
			if len(me.ManaPool) == 0 {
				t.Skip("a free cast has nothing to be short of")
			}
			me.ManaPool = me.ManaPool[:len(me.ManaPool)-1]

			strict := params
			strict.Strict = true
			err = g.CastSpell(me.ID, id, strict)
			var short *InsufficientManaError
			if !errorsAs(err, &short) {
				t.Fatalf("one mana short of the previewed price: got %v, want *InsufficientManaError", err)
			}
		})
	}
}

// The preview prices the FACE being announced (ADR 0034). A modal DFC
// whose back half costs something else was previewed at the front
// half's price, because the endpoint never saw `face`.
func TestPriceCastReadsTheAnnouncedFace(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Restoration", me.ID)
	c.Layout = LayoutModalDFC
	c.Faces = []Face{
		{Name: "Front Half", TypeLine: "Sorcery", ManaCost: "{4}{U}"},
		{Name: "Back Half", TypeLine: "Land", ManaCost: ""},
	}
	c.SetFace(0)
	me.Hand.PushTop(c)

	card, _ := g.LookupCardForEffect(c.InstanceID)
	front, err := g.PriceCast(me.ID, card, CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast front: %v", err)
	}
	if got := costKey(front.Total); got != "generic=4 x=0 colors=[U]" {
		t.Errorf("front face price = %s, want {4}{U}", got)
	}
	if front.Card.Name != "Front Half" {
		t.Errorf("priced card name = %q, want the front face", front.Card.Name)
	}
	back, err := g.PriceCast(me.ID, card, CastSpellParams{Face: 1})
	if err != nil {
		t.Fatalf("PriceCast back: %v", err)
	}
	if got := costKey(back.Total); got != "generic=0 x=0 colors=[]" {
		t.Errorf("back face price = %s, want the back face's own (free) cost", got)
	}
	if back.Card.Name != "Back Half" {
		t.Errorf("priced card name = %q, want the back face — the caller reads the spend context off it", back.Card.Name)
	}
}

// Base is the pre-tapping total and Total is what the payment
// charges. The split is load-bearing for the bot enumerator, which
// searches an X against Base: convoke settles the {X} slot into
// generic as part of PAYING, and an enumerator reading Total would
// see XSlots=0 and never offer Chord of Calling at X above nothing.
func TestCastPriceKeepsTheXSlotBeforeTheConvokeFold(t *testing.T) {
	const chord = "test-price-convoke"
	g := newActiveGame(t)
	me := g.Seats[0]
	prev := CatalogTapPermanentsCost
	CatalogTapPermanentsCost = func(id string) *TapPermanentsCost {
		if id == chord {
			return convokeForTest()
		}
		return nil
	}
	t.Cleanup(func() { CatalogTapPermanentsCost = prev })

	id := spellInHand(t, g, me, "Test Chord", "Instant", "{X}{G}{G}{G}")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].OracleID = chord
			}
		}
	})
	card, _ := g.LookupCardForEffect(id)
	price, err := g.PriceCast(me.ID, card, CastSpellParams{XValue: 2})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if price.Base.XSlots != 1 {
		t.Errorf("Base.XSlots = %d, want the printed {X} still open for the enumerator's search", price.Base.XSlots)
	}
	if price.Total.XSlots != 0 || price.Total.Generic != 2 {
		t.Errorf("Total = %s, want X settled into generic by the convoke fold", costKey(price.Total))
	}
}
