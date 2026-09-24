package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// paused_exit_cast_test.go — #1474, the two uses of a paused card that
// neither #1451's `moving` list nor #1478's `tapping` list could see.
//
// A commander an effect is exiling, destroying or discarding sits
// where it was while its owner answers CR 903.9, but as far as the
// rules are concerned it has already left. #1451 refused every cost
// that would MOVE it and #1478 every cost that would TAP it. What was
// left:
//
//   - CASTING the paused card itself (or playing it, if it is a land).
//     CR 601.2a moves the spell before any cost is paid, so it was in
//     no cost list at all, and CastSpell put it on the stack with the
//     exile's prompt still open.
//   - Activating an ability OF the paused permanent through a cost
//     component that neither moves nor taps it — a counter it removes
//     or adds, a loyalty cost, or no cost at all — and naming it as
//     another ability's counter-removal source.
//
// All of them are refused with ErrChoicePending through the one gate,
// refusePausedCostCardsLocked, before anything moves or is paid.

// pausedCastTable is a commander paused on its way out of `zone`, with
// {C}{C} floating and an ordinary card of the same shape in the same
// zone to show the refusal is about the card, not the table.
type pausedCastTable struct {
	g      *Game
	me     *Player
	cmdr   uuid.UUID
	spare  uuid.UUID
	prompt *PendingChoice
}

func newPausedCastTable(t *testing.T, zone func(*Player) *Zone, typeLine string) *pausedCastTable {
	t.Helper()
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	me.Hand.Cards = nil
	z := zone(me)
	cmdr := seatCommander(t, z, me)
	editZoneCard(z, cmdr, func(c *Card) {
		c.TypeLine = typeLine
		c.OracleID = "test-paused-cast"
		if typeLine != "Legendary Land" {
			c.ManaCost = "{1}"
		}
	})
	spare := NewCard("Spare", me.ID)
	spare.TypeLine = typeLine
	spare.OracleID = "test-paused-cast"
	if typeLine != "Legendary Land" {
		spare.ManaCost = "{1}"
	}
	z.PushTop(spare)
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	return &pausedCastTable{g: g, me: me, cmdr: cmdr, spare: spare.InstanceID,
		prompt: exileCommanderPaused(t, g, me, z, cmdr)}
}

// assertNothingSpent checks a refused announcement left the table as
// it found it: the card where it was, no spell or ability on the stack,
// the pool untouched, no land drop spent, and the exile's prompt the
// only one open.
func (tb *pausedCastTable) assertNothingSpent(t *testing.T, z *Zone, pool int) {
	t.Helper()
	g, me := tb.g, tb.me
	if !z.Contains(tb.cmdr) {
		t.Fatalf("the paused commander left the %s", z.Kind)
	}
	if len(g.Stack.Cards) != 0 || len(g.StackMeta) != 0 {
		t.Errorf("stack %d / meta %d — the refused cast reached the stack", len(g.Stack.Cards), len(g.StackMeta))
	}
	if g.Battlefield.Contains(tb.cmdr) || g.LandsPlayedThisTurn[me.ID] != 0 {
		t.Errorf("the refused land play reached the battlefield or spent the land drop")
	}
	if len(me.ManaPool) != pool {
		t.Errorf("pool %v, want %d mana — the refused cast paid", me.ManaPool, pool)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].ID != tb.prompt.ID {
		t.Fatalf("the refusal disturbed the exile's prompt (%d pending)", len(g.PendingChoices))
	}
}

// A commander an effect is exiling out of its owner's hand or
// graveyard cannot be cast, or played if it is a land, while its owner
// decides. Nothing moves and nothing is paid. The same cast of an
// ordinary card beside it goes through with the prompt open, and once
// the owner sends the commander to the command zone it casts from
// there as usual.
func TestAPausedCommanderCannotBeCastWhileItsOwnerIsAsked(t *testing.T) {
	hand := func(p *Player) *Zone { return p.Hand }
	yard := func(p *Player) *Zone { return p.Graveyard }
	for _, tc := range []struct {
		name     string
		zone     func(*Player) *Zone
		from     string
		typeLine string
	}{
		{"a creature spell from hand", hand, "", "Legendary Creature — Angel"},
		{"an instant from hand", hand, "", "Legendary Instant"},
		{"a creature spell from the graveyard", yard, "graveyard", "Legendary Creature — Zombie"},
		{"a land from hand", hand, "", "Legendary Land"},
		{"a land from the graveyard", yard, "graveyard", "Legendary Land"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Gravecrawler's "you may cast this card from your
			// graveyard", and Crucible of Worlds' land half, as the
			// card's own declaration.
			withCatalogCastableZones(t, castableZonesFor("test-paused-cast", ZoneGraveyard))
			tb := newPausedCastTable(t, tc.zone, tc.typeLine)
			g, me := tb.g, tb.me
			z := tc.zone(me)

			err := g.CastSpell(me.ID, tb.cmdr, CastSpellParams{FromZone: tc.from, Strict: true})
			if !errors.Is(err, ErrChoicePending) {
				t.Errorf("casting the paused commander: err = %v, want ErrChoicePending", err)
			}
			tb.assertNothingSpent(t, z, 2)

			// The ordinary card in the same zone is unaffected.
			if err := g.CastSpell(me.ID, tb.spare, CastSpellParams{FromZone: tc.from, Strict: true}); err != nil {
				t.Fatalf("casting an ordinary card with the prompt open: %v", err)
			}
			if z.Contains(tb.spare) {
				t.Fatal("the ordinary card did not leave its zone")
			}

			if err := g.ResolveOptionalReplacement(tb.prompt.ID, me.ID, true); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
			assertOnlyIn(t, tb.cmdr, me.Command, z, g.Exile, g.Stack)
			if tc.typeLine == "Legendary Land" {
				return // no rule plays a land out of the command zone
			}
			// The answer is in: the commander casts again, from where
			// the answer put it.
			for len(g.StackMeta) > 0 {
				resolveTop(t, g)
			}
			me.ManaPool.AddMana(ManaToken{Color: "C"})
			if err := g.CastSpell(me.ID, tb.cmdr, CastSpellParams{FromZone: "command", Strict: true}); err != nil {
				t.Fatalf("casting the commander from the command zone after the answer: %v", err)
			}
			if !g.Stack.Contains(tb.cmdr) {
				t.Error("the commander is not on the stack after the answered cast")
			}
		})
	}
}

// --- ability costs that neither move nor tap -------------------------

// counterMark is an ability with `cost` that marks its source on
// resolution.
func counterMark(label string, cost AbilityCost) ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label: label,
		Cost:  cost,
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}
}

// A destroyed commander, its owner still deciding, cannot be the
// source of an activation — through a counter it removes or adds, a
// loyalty cost, a counter-cost mana ability, or no cost at all — and
// cannot be named as another ability's counter-removal source. Each
// refusal changes no counter, puts nothing on the stack or in the pool,
// and leaves the destroy's prompt alone. An ordinary permanent in the
// same slot still pays, prompt open, and after the answer the
// commander is where the destroy sent it.
func TestADestroyedCommanderCannotPayOrActivateWithCounters(t *testing.T) {
	loyalty := 1
	for _, tc := range []struct {
		name string
		try  func(g *Game, me *Player, id, remover uuid.UUID) error
	}{
		{"its own remove-a-counter ability", func(g *Game, me *Player, id, _ uuid.UUID) error {
			return g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{})
		}},
		{"its own put-a-counter ability", func(g *Game, me *Player, id, _ uuid.UUID) error {
			return g.ActivateCatalogAbility(me.ID, id, 1, ActivateAbilityParams{})
		}},
		{"its own loyalty ability", func(g *Game, me *Player, id, _ uuid.UUID) error {
			return g.ActivateCatalogAbility(me.ID, id, 2, ActivateAbilityParams{})
		}},
		{"its own free ability", func(g *Game, me *Player, id, _ uuid.UUID) error {
			return g.ActivateCatalogAbility(me.ID, id, 3, ActivateAbilityParams{})
		}},
		{"its own remove-a-counter mana ability", func(g *Game, me *Player, id, _ uuid.UUID) error {
			return g.ActivateManaAbility(me.ID, id, 0, ManaAbilityParams{})
		}},
		{"its own put-a-counter mana ability", func(g *Game, me *Player, id, _ uuid.UUID) error {
			return g.ActivateManaAbility(me.ID, id, 1, ManaAbilityParams{})
		}},
		{"the sandbox loyalty verb", func(g *Game, me *Player, id, _ uuid.UUID) error {
			return g.ActivateLoyalty(me.ID, id, "+1", 1)
		}},
		{"another ability's counter-removal source", func(g *Game, me *Player, id, remover uuid.UUID) error {
			return g.ActivateCatalogAbility(me.ID, remover, 0, ActivateAbilityParams{CounterSourceIDs: []uuid.UUID{id}})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			remover := pushCounterCostSource(g, me, AbilityCost{RemoveCounters: &CounterRemovalCost{
				Counter: "+1/+1", N: 1, From: creatureCostSpec(),
			}}, nil)
			withCounterAbilities := func(c *Card) {
				c.TypeLine = "Legendary Planeswalker Creature — Angel"
				c.Counters = map[string]int{"+1/+1": 2, CounterLoyalty: 3}
				c.ActivatedAbilities = []ActivatedAbilityShape{
					counterMark("Remove a +1/+1 counter: mark", AbilityCost{RemoveCounters: &CounterRemovalCost{Counter: "+1/+1", N: 1}}),
					counterMark("Put a -1/-1 counter: mark", AbilityCost{AddCounter: &CounterAddCost{Counter: "-1/-1", N: 1}}),
					counterMark("+1: mark", AbilityCost{Loyalty: &loyalty}),
					counterMark("{0}: mark", AbilityCost{}),
				}
				c.ManaAbilities = []ManaAbilityShape{
					{RemoveCounters: &CounterRemovalCost{Counter: "+1/+1", N: 1}, Produced: "{G}", Label: "Remove a +1/+1 counter: Add {G}"},
					{AddCounter: &CounterAddCost{Counter: "-1/-1", N: 1}, Produced: "{G}", Label: "Put a -1/-1 counter: Add {G}"},
				}
			}
			cmdr := seatCommander(t, g.Battlefield, me)
			editBattlefieldCard(g, cmdr, withCounterAbilities)
			prompt := destroyCommanderPaused(t, g, me, cmdr)

			if err := tc.try(g, me, cmdr, remover); !errors.Is(err, ErrChoicePending) {
				t.Errorf("naming the destroyed commander: err = %v, want ErrChoicePending", err)
			}
			c := findBattlefieldCard(g, cmdr)
			if c == nil {
				t.Fatal("the destroyed commander left the battlefield before its owner answered")
			}
			if c.Counters["+1/+1"] != 2 || c.Counters["-1/-1"] != 0 || c.Counters[CounterLoyalty] != 3 || g.LoyaltyActivatedThisTurn[cmdr] {
				t.Errorf("the refused payment changed the commander's counters: %v", c.Counters)
			}
			if len(g.StackMeta) != 0 || len(me.ManaPool) != 0 {
				t.Errorf("stack %d, pool %v — the refused payment paid something", len(g.StackMeta), me.ManaPool)
			}
			if len(g.PendingChoices) != 1 || g.PendingChoices[0].ID != prompt.ID {
				t.Fatalf("the refusal disturbed the destroy's prompt (%d pending)", len(g.PendingChoices))
			}

			// The same payment with an ordinary permanent, prompt open.
			bear := pushTapOthersPermanent(g, me, "Bear", "Creature — Bear", false)
			editBattlefieldCard(g, bear, withCounterAbilities)
			if err := tc.try(g, me, bear, remover); err != nil {
				t.Errorf("naming an ordinary permanent while the prompt is open: %v", err)
			}

			if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, false); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
			assertOnlyIn(t, cmdr, me.Graveyard, g.Battlefield, me.Command)
		})
	}
}
