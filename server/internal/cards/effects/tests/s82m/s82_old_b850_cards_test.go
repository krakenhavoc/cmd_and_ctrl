package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b850_cards_test.go — Ice Age lands, counterspell creatures and
// Hallowed Ground, S82 OLD batch 850 (#5636), played through the real catalog
// entries.

const (
	o850Veldt     = "10b9bfb3-c478-45c6-b227-9c66b63bc79b"
	o850LavaTubes = "2ffb2647-60bb-4916-b919-cbcf18e5e424"
	o850Delta     = "4c6c064c-da61-4c7a-9607-8fb4490ff9a6"
	o850Ridge     = "a597d2b7-484b-4cfd-89a6-d166cb1a3420"
	o850LandCap   = "bfec4d0a-3792-4bc3-bae1-e639da5bb9a6"
	o850Torch     = "313d7b2f-05fb-49f3-8d8a-361be4412d3c"
	o850Shaman    = "3e3e2f71-2159-4f83-a9c1-a67ecac8a711"
	o850Cabal     = "c8312964-5cf0-435a-8757-ea1bb67258a4"
	o850Mistfolk  = "d57409dc-22fa-40c1-a043-cab94d5f3b5f"
	o850Hallowed  = "ee8fd844-4551-4e44-9740-00d2e76eb337"
)

func TestS82OldB850CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o850Veldt: "Veldt", o850LavaTubes: "Lava Tubes", o850Delta: "River Delta", o850Ridge: "Timberline Ridge",
		o850LandCap: "Land Cap", o850Torch: "Order of the Sacred Torch", o850Shaman: "Brine Shaman",
		o850Cabal: "Stromgald Cabal", o850Mistfolk: "Mistfolk", o850Hallowed: "Hallowed Ground",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82OldB850DepletionLandsTapForEitherColourAndSitOutAnUntapStep(t *testing.T) {
	for _, c := range []struct {
		name, oracle, a, b string
	}{
		{"Veldt", o850Veldt, "G", "W"}, {"Lava Tubes", o850LavaTubes, "B", "R"}, {"River Delta", o850Delta, "U", "B"},
		{"Timberline Ridge", o850Ridge, "R", "G"}, {"Land Cap", o850LandCap, "W", "U"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			land := effectstest.PushCatalogPermanent(g, me.ID, c.name, "Land", c.oracle, false)
			for i, col := range []string{c.a, c.b} {
				g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
				if i == 1 {
					g.WithWriteLock(func() { effectstest.BattlefieldCardFor(g, land).Tapped = false })
					g.WithWriteLock(func() { effectstest.BattlefieldCardFor(g, land).Counters = nil })
				}
				if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{col}}); err != nil {
					t.Fatalf("tap for %s: %v", col, err)
				}
				if cols := effectstest.Batch01PoolColors(me); len(cols) != 1 || cols[0] != col {
					t.Errorf("pool %v, want {%s}", cols, col)
				}
				if n := effectstest.B12Counter(t, g, land, "depletion"); n != 1 {
					t.Errorf("%d depletion counters after tapping, want 1", n)
				}
			}
			g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
			if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{"X"}}); err == nil {
				t.Error("a tapped land made mana, or a colour it does not print was accepted")
			}
			// An off-colour pick is refused while untapped.
			g.WithWriteLock(func() { effectstest.BattlefieldCardFor(g, land).Tapped = false })
			if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{"C"}}); err == nil {
				t.Error("an unprinted colour was accepted")
			}
		})
	}
}

func TestS82OldB850VeldtStaysTappedThroughTheNextUntapStepAndLosesItsCounterInUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Veldt", "Land", o850Veldt, false)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap: %v", err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.BattlefieldCardFor(g, land).Tapped {
		t.Error("the Veldt untapped with a depletion counter on it")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, land, "depletion"); n != 0 {
		t.Errorf("%d depletion counters after upkeep, want 0", n)
	}
	if !effectstest.BattlefieldCardFor(g, land).Tapped {
		t.Error("the Veldt untapped mid-turn")
	}
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 0)
	if effectstest.BattlefieldCardFor(g, land).Tapped {
		t.Error("the Veldt, with no counter, stayed tapped through the following untap step")
	}
}

func o850Mana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("mana: %v", err)
		}
	})
}

func o850Countered(g *game.Game, spell uuid.UUID) bool {
	return effectstest.CountEventsOf(g, game.EventCounterSpell, spell) > 0
}

// o850Cast has opp cast a spell of the given type and colour from hand at the
// precombat main step, returning its id.
func o850Cast(t *testing.T, g *game.Game, opp *game.Player, name, typeLine, cost string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func o850Counter(t *testing.T, g *game.Game, who *game.Player, src, spell uuid.UUID, idx int, strict bool) error {
	t.Helper()
	return g.ActivateCatalogAbility(who.ID, src, idx, game.ActivateAbilityParams{Strict: strict,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}})
}

func TestS82OldB850OrderOfTheSacredTorchCountersABlackSpellForALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	torch := effectstest.B12Push(g, me.ID, "Order of the Sacred Torch", "Creature — Human Knight", o850Torch, 2, 2)
	white := o850Cast(t, g, opp, "White Thing", "Instant", "{W}", nil)
	if err := o850Counter(t, g, me, torch, white, 0, true); err == nil {
		t.Fatal("countered a white spell")
	}
	black := o850Cast(t, g, opp, "Black Thing", "Instant", "{B}", nil)
	life := me.Life
	if err := o850Counter(t, g, me, torch, black, 0, true); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d (1 paid)", me.Life, life-1)
	}
	if !effectstest.BattlefieldCardFor(g, torch).Tapped {
		t.Error("the Order did not tap")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !o850Countered(g, black) {
		t.Error("the black spell was not countered")
	}
	if o850Countered(g, white) {
		t.Error("the white spell was countered")
	}
}

func TestS82OldB850StromgaldCabalCountersAWhiteSpellForALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cabal := effectstest.B12Push(g, me.ID, "Stromgald Cabal", "Creature — Human Knight", o850Cabal, 2, 2)
	black := o850Cast(t, g, opp, "Black Thing", "Instant", "{B}", nil)
	if err := o850Counter(t, g, me, cabal, black, 0, true); err == nil {
		t.Fatal("countered a black spell")
	}
	white := o850Cast(t, g, opp, "White Thing", "Instant", "{W}", nil)
	life := me.Life
	if err := o850Counter(t, g, me, cabal, white, 0, true); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 || !o850Countered(g, white) || o850Countered(g, black) {
		t.Errorf("life %d (want %d), white countered %v, black countered %v", me.Life, life-1, o850Countered(g, white), o850Countered(g, black))
	}
}

func TestS82OldB850StromgaldCabalWillNotActivateWhileSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cabal := effectstest.PushCatalogPermanent(g, me.ID, "Stromgald Cabal", "Creature — Human Knight", o850Cabal, true)
	white := o850Cast(t, g, opp, "White Thing", "Instant", "{W}", nil)
	if err := o850Counter(t, g, me, cabal, white, 0, true); err == nil {
		t.Error("a summoning-sick Cabal paid {T}")
	}
}

func TestS82OldB850BrineShamanPumpsAndCountersCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0] // the active seat casts at sorcery speed
	shaman := effectstest.B12Push(g, me.ID, "Brine Shaman", "Creature — Human Cleric Shaman", o850Shaman, 1, 1)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	buddy := effectstest.B12Creature(g, me.ID, "Buddy", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardRefs(buddy), SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the sacrificed creature is not in the graveyard")
	}
	if p, tough := effectstest.PtOf(t, g, buddy); p != 4 || tough != 4 {
		t.Errorf("target is %d/%d, want 4/4", p, tough)
	}
	// The counter: a creature spell is a legal target, a noncreature spell is not.
	inst := o850Cast(t, g, opp, "Bolt-ish", "Instant", "{R}", nil)
	o850Mana(t, g, me, "{C}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, shaman, 1, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardRefs(inst), SacrificeIDs: []uuid.UUID{buddy}}); err == nil {
		t.Fatal("countered a noncreature spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	o850Mana(t, g, me, "{C}{U}{U}")
	bear := o850Cast(t, g, opp, "Cast Bear", "Creature — Bear", "{G}", nil)
	if err := g.ActivateCatalogAbility(me.ID, shaman, 1, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardRefs(bear), SacrificeIDs: []uuid.UUID{buddy}}); err != nil {
		t.Fatalf("counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !o850Countered(g, bear) {
		t.Error("the creature spell was not countered")
	}
	if !me.Graveyard.Contains(buddy) {
		t.Error("the second sacrifice did not happen")
	}
}

func TestS82OldB850MistfolkCountersOnlyASpellThatTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mist := effectstest.B12Push(g, me.ID, "Mistfolk", "Creature — Illusion", o850Mistfolk, 1, 2)
	other := effectstest.B12Push(g, me.ID, "Mistfolk", "Creature — Illusion", o850Mistfolk, 1, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elsewhere := o850Cast(t, g, opp, "Aimed At Bear", "Instant", "{R}", effectstest.CardRefs(bear))
	o850Mana(t, g, me, "{U}{U}")
	if err := o850Counter(t, g, me, mist, elsewhere, 0, true); err == nil {
		t.Fatal("countered a spell that does not target this creature")
	}
	aimedAtOther := o850Cast(t, g, opp, "Aimed At Other", "Instant", "{R}", effectstest.CardRefs(other))
	if err := o850Counter(t, g, me, mist, aimedAtOther, 0, true); err == nil {
		t.Fatal("countered a spell that targets a different Mistfolk")
	}
	aimed := o850Cast(t, g, opp, "Aimed At Mist", "Instant", "{R}", effectstest.CardRefs(mist))
	if err := o850Counter(t, g, me, mist, aimed, 0, true); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !o850Countered(g, aimed) {
		t.Error("the spell that targeted the Mistfolk was not countered")
	}
	if o850Countered(g, elsewhere) || o850Countered(g, aimedAtOther) {
		t.Error("a spell that did not target this Mistfolk was countered")
	}
}

func TestS82OldB850HallowedGroundBouncesANonsnowLandYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hg := effectstest.B12Push(g, me.ID, "Hallowed Ground", "Enchantment", o850Hallowed, 0, 0)
	forest := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	snow := effectstest.PushCatalogPermanent(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	o850Mana(t, g, me, "{W}{W}{W}{W}{W}{W}")
	for _, bad := range []uuid.UUID{snow, theirs} {
		if err := g.ActivateCatalogAbility(me.ID, hg, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bad)}); err == nil {
			t.Error("bounced a snow land or a land the activator does not control")
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, hg, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(forest)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) || !me.Hand.Contains(forest) {
		t.Error("the Forest did not return to its owner's hand")
	}
}
