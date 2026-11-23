package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_player_batch_b_test.go — the second batch of "Any player may
// activate this ability" cards (ADR 0106 PR 6, #1793). Each test has a
// player who does not control the permanent activate its any-player
// row, pay for it out of their own resources (CR 602.1a), and checks
// that the effect is the printed one. Xantcha's test
// (xantcha_sleeper_agent_test.go) pins the seam itself.

const (
	apbAetherStormOracle       = "ff4297d3-3d96-4bd6-a606-1bdc20a6df2b"
	apbArmageddonClockOracle   = "70d90ef4-0cda-405f-abf1-734fa909efa6"
	apbInfiniteHourglassOracle = "7c59ad65-ad1f-4dc4-bcf2-ea6b2b48fad8"
	apbSamiteSanctuaryOracle   = "9e6e1ea6-4682-4056-b70e-fe45d0c9fdee"
	apbTidalControlOracle      = "855e200e-5375-4aac-b1f7-5113162e7e14"
	apbVolrathsDungeonOracle   = "a8b968ea-76f3-4ee7-9d53-d251d8a9faf6"
	apbWishmongerOracle        = "f68b90af-9502-4d99-aadf-2bfcea8a655a"
)

// apbColorless puts n colourless mana in a player's pool.
func apbColorless(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
}

func TestAetherStormStopsCreatureSpellsAndAnyoneMayPayLifeToEndIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	storm := effectstest.PushCatalogPermanent(g, owner.ID, "Aether Storm", "Enchantment", apbAetherStormOracle, false)
	effectstest.ApbRequireAnyPlayerRows(t, g, storm, 0)

	// "Creature spells can't be cast" binds its own controller …
	if err := effectstest.CastCatalogSpellErr(t, g, "Grizzly Bears", "Creature — Bear", "", nil); err == nil {
		t.Fatal("Aether Storm's controller cast a creature spell")
	}
	// … and everyone else, for creature spells only.
	spec, _ := effects.Lookup(apbAetherStormOracle)
	forbids := spec.CastRestrictions[0].Forbids
	src := *effectstest.FindBattlefieldCardForTest(g, storm)
	g.WithWriteLock(func() {
		if !forbids(game.CastQuery{Game: g, Card: game.Card{Name: "Bear", TypeLine: "Creature — Bear"}, Controller: other.ID, Source: src, FromZone: game.ZoneHand}) {
			t.Error("another player's creature spell is not refused")
		}
		if forbids(game.CastQuery{Game: g, Card: game.Card{Name: "Shock", TypeLine: "Instant"}, Controller: other.ID, Source: src, FromZone: game.ZoneHand}) {
			t.Error("an instant is refused")
		}
	})

	// A regeneration shield does not save it (CR 701.19c).
	g.WithWriteLock(func() {
		if err := (effects.Regenerate{Target: storm}).Apply(effects.NewContext(g, &game.StackItem{Controller: owner.ID, SourceCardID: uuid.New()})); err != nil {
			t.Fatalf("regenerate: %v", err)
		}
	})

	// A player who does not control it pays the 4 life (CR 602.1a).
	ownerLife, otherLife := owner.Life, other.Life
	if err := g.ActivateCatalogAbility(other.ID, storm, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("another player activates Aether Storm: %v", err)
	}
	if other.Life != otherLife-4 || owner.Life != ownerLife {
		t.Errorf("life after paying: activator %d (want %d), controller %d (want %d)", other.Life, otherLife-4, owner.Life, ownerLife)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(storm) {
		t.Fatal("Aether Storm survived its own ability")
	}
	if !owner.Graveyard.Contains(storm) {
		t.Error("Aether Storm is not in its owner's graveyard")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Grizzly Bears", "Creature — Bear", "", nil); err != nil {
		t.Errorf("a creature spell is still refused once Aether Storm is gone: %v", err)
	}
}

func TestArmageddonClockCountsUpAndAnyoneMayWindItBackDuringAnUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, other := g.Seats[1], g.Seats[2]
	clock := effectstest.PushCatalogPermanent(g, owner.ID, "Armageddon Clock", "Artifact", apbArmageddonClockOracle, false)
	effectstest.ApbRequireAnyPlayerRows(t, g, clock, 0)

	// Outside an upkeep step the row is closed to everyone.
	apbColorless(other, 4)
	if err := g.ActivateCatalogAbility(other.ID, clock, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("Armageddon Clock was activated outside an upkeep step")
	}
	other.ManaPool = nil

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TriggerOnStack(g, clock) == nil {
		t.Fatal("no doom-counter trigger at its controller's upkeep")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.ApbCounters(g, clock, "doom"); n != 1 {
		t.Fatalf("doom counters after the upkeep trigger = %d, want 1", n)
	}

	// In the upkeep, a player who does not control it pays {4} out of
	// their own pool and takes the counter off.
	apbColorless(owner, 4)
	apbColorless(other, 4)
	if err := g.ActivateCatalogAbility(other.ID, clock, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("another player activates Armageddon Clock in an upkeep: %v", err)
	}
	if other.ManaPool.Total() != 0 || owner.ManaPool.Total() != 4 {
		t.Errorf("pools after paying {4}: activator %d (want 0), controller %d (want 4)", other.ManaPool.Total(), owner.ManaPool.Total())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.ApbCounters(g, clock, "doom"); n != 0 {
		t.Fatalf("doom counters after removal = %d, want 0", n)
	}
	// With none left it can still be activated, and does nothing.
	apbColorless(other, 4)
	if err := g.ActivateCatalogAbility(other.ID, clock, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with no doom counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.ApbCounters(g, clock, "doom"); n != 0 {
		t.Fatalf("doom counters after removal from none = %d, want 0", n)
	}

	// The draw-step trigger deals the count to every player.
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(clock, "doom", 2); err != nil {
			t.Fatalf("add doom counters: %v", err)
		}
	})
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepDraw)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != before[p.ID]-2 {
			t.Errorf("seat %s life %d, want %d", p.Name, p.Life, before[p.ID]-2)
		}
	}
}

func TestInfiniteHourglassPumpsEveryCreatureAndAnyoneMayRemoveATimeCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, other := g.Seats[1], g.Seats[3]
	bear := effectstest.SeedCreature(g, "Bear", g.Seats[0].ID)
	glass := effectstest.PushCatalogPermanent(g, owner.ID, "Infinite Hourglass", "Artifact", apbInfiniteHourglassOracle, false)
	effectstest.ApbRequireAnyPlayerRows(t, g, glass, 0)

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.ApbCounters(g, glass, game.CounterTime); n != 1 {
		t.Fatalf("time counters after the upkeep trigger = %d, want 1", n)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("another player's creature has power %d with one time counter, want 3", p)
	}

	apbColorless(other, 3)
	if err := g.ActivateCatalogAbility(other.ID, glass, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("another player activates Infinite Hourglass in an upkeep: %v", err)
	}
	if other.ManaPool.Total() != 0 {
		t.Errorf("activator's pool holds %d after paying {3}", other.ManaPool.Total())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.ApbCounters(g, glass, game.CounterTime); n != 0 {
		t.Fatalf("time counters after removal = %d, want 0", n)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("power %d with no time counters, want 2", p)
	}

	effectstest.AdvanceToMainOf(t, g, 1)
	apbColorless(other, 3)
	if err := g.ActivateCatalogAbility(other.ID, glass, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Infinite Hourglass was activated in a main phase")
	}
}

func TestSamiteSanctuaryShieldsAnotherPlayersCreatureForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, ctrl, other := g.Seats[0], g.Seats[1], g.Seats[2]
	sanctuary := effectstest.PushCatalogPermanent(g, owner.ID, "Samite Sanctuary", "Enchantment", apbSamiteSanctuaryOracle, false)
	effectstest.ApbRequireAnyPlayerRows(t, g, sanctuary, 0)
	bear := effectstest.SeedCreature(g, "Bear", ctrl.ID)

	apbColorless(other, 2)
	if err := g.ActivateCatalogAbility(other.ID, sanctuary, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("another player activates Samite Sanctuary: %v", err)
	}
	if other.ManaPool.Total() != 0 {
		t.Errorf("activator's pool holds %d after paying {2}", other.ManaPool.Total())
	}
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		if err := (effects.DealDamage{Source: uuid.New(), Target: bear, Amount: 2}).Apply(effects.NewContext(g, &game.StackItem{Controller: owner.ID})); err != nil {
			t.Fatalf("deal damage: %v", err)
		}
	})
	if c := effectstest.FindBattlefieldCardForTest(g, bear); c == nil || c.DamageMarked != 1 {
		t.Fatalf("damage marked = %+v, want 1 of 2 (the shield holds one)", c)
	}
}

func TestTidalControlEitherCostCountersARedOrGreenSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, caster, payer, owner := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	tidal := effectstest.PushCatalogPermanent(g, owner.ID, "Tidal Control", "Enchantment", apbTidalControlOracle, false)
	effectstest.ApbRequireAnyPlayerRows(t, g, tidal, 0, 1)
	before := me.Life

	// "Pay 2 life": a player who controls neither the spell nor Tidal
	// Control pays the life.
	bolt := effectstest.B10OpponentCastsBolt(t, g, caster, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	payerLife, ownerLife := payer.Life, owner.Life
	if err := g.ActivateCatalogAbility(payer.ID, tidal, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("pay-life row: %v", err)
	}
	if payer.Life != payerLife-2 || owner.Life != ownerLife {
		t.Errorf("life after paying: activator %d (want %d), controller %d (want %d)", payer.Life, payerLife-2, owner.Life, ownerLife)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before || !caster.Graveyard.Contains(bolt) {
		t.Fatalf("the Bolt was not countered: life %d → %d", before, me.Life)
	}

	// "{2}": the Bolt's own target pays the mana.
	bolt2 := effectstest.B10OpponentCastsBolt(t, g, caster, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	apbColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, tidal, 1, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt2}},
	}); err != nil {
		t.Fatalf("mana row: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("activator's pool holds %d after paying {2}", me.ManaPool.Total())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before || !caster.Graveyard.Contains(bolt2) {
		t.Fatalf("the second Bolt was not countered: life %d → %d", before, me.Life)
	}

	// A blue spell is not a legal target.
	blue := effectstest.Batch01OpponentCasts(t, g, caster, "Blue Instant", "", "{U}", nil)
	if err := g.ActivateCatalogAbility(payer.ID, tidal, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: blue}},
	}); err == nil {
		t.Error("Tidal Control targeted a blue spell")
	}
}

func TestVolrathsDungeonAnyoneMayPayLifeOnTheirOwnTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, other := g.Seats[0], g.Seats[1]
	dungeon := effectstest.PushCatalogPermanent(g, owner.ID, "Volrath's Dungeon", "Enchantment", apbVolrathsDungeonOracle, false)
	effectstest.ApbRequireAnyPlayerRows(t, g, dungeon, 0)

	// "But only during their turn": not during the controller's turn …
	if err := g.ActivateCatalogAbility(other.ID, dungeon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("another player destroyed Volrath's Dungeon during its controller's turn")
	}
	// … and the controller's discard row is not theirs at all.
	if len(other.Hand.Cards) == 0 {
		t.Fatal("seat 1 has no hand to discard from")
	}
	if err := g.ActivateCatalogAbility(other.ID, dungeon, 1, game.ActivateAbilityParams{
		Strict:     true,
		DiscardIDs: []uuid.UUID{other.Hand.Cards[0].InstanceID},
		Targets:    []game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}},
	}); err == nil {
		t.Fatal("another player activated the Dungeon's discard row")
	}

	effectstest.AdvanceToMainOf(t, g, 1)
	life := other.Life
	if err := g.ActivateCatalogAbility(other.ID, dungeon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("another player on their own turn: %v", err)
	}
	if other.Life != life-5 {
		t.Errorf("activator life %d, want %d", other.Life, life-5)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dungeon) || !owner.Graveyard.Contains(dungeon) {
		t.Error("Volrath's Dungeon was not destroyed")
	}
}

func TestVolrathsDungeonTargetPlayerPutsACardOfTheirChoiceOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, victim := g.Seats[0], g.Seats[1]
	dungeon := effectstest.PushCatalogPermanent(g, owner.ID, "Volrath's Dungeon", "Enchantment", apbVolrathsDungeonOracle, false)
	effectstest.AdvanceToMainOf(t, g, 0)

	pay := owner.Hand.Cards[0].InstanceID
	if err := g.ActivateCatalogAbility(owner.ID, dungeon, 1, game.ActivateAbilityParams{
		Strict:     true,
		DiscardIDs: []uuid.UUID{pay},
		Targets:    []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
	}); err != nil {
		t.Fatalf("discard row: %v", err)
	}
	if !owner.Graveyard.Contains(pay) {
		t.Error("the discarded card is not in the controller's graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if len(victim.Hand.Cards) < 2 {
		t.Fatal("the victim needs two cards for the choice to be one")
	}
	pick := victim.Hand.Cards[1].InstanceID
	handBefore := victim.Hand.Size()
	effectstest.AnswerChooseCards(t, g, victim.ID, pick)
	effectstest.PassPriorityAroundTable(t, g)
	top, err := victim.Library.Top()
	if err != nil || top.InstanceID != pick {
		t.Fatalf("library top = %v (%v), want the victim's pick %v", top.InstanceID, err, pick)
	}
	if victim.Hand.Size() != handBefore-1 {
		t.Errorf("victim hand %d, want %d", victim.Hand.Size(), handBefore-1)
	}
}

func TestWishmongerColourIsChosenByTheTargetsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, ctrl, other := g.Seats[0], g.Seats[1], g.Seats[2]
	monger := effectstest.PushCatalogPermanent(g, owner.ID, "Wishmonger", "Creature — Unicorn Monger", apbWishmongerOracle, false)
	effectstest.ApbRequireAnyPlayerRows(t, g, monger, 0)
	bear := effectstest.SeedCreature(g, "Bear", ctrl.ID)

	apbColorless(other, 2)
	if err := g.ActivateCatalogAbility(other.ID, monger, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("another player activates Wishmonger: %v", err)
	}
	if other.ManaPool.Total() != 0 {
		t.Errorf("activator's pool holds %d after paying {2}", other.ManaPool.Total())
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.AnswerColor(t, g, ctrl.ID, "R")
	if prompt.Chooser != ctrl.ID {
		t.Errorf("the colour was asked of %v, want the creature's controller %v", prompt.Chooser, ctrl.ID)
	}
	if got := effectstest.ProtectionsOn(t, g, bear); !effectstest.HasString(got, "red") {
		t.Errorf("protections on the bear = %q, want red", got)
	}
}
