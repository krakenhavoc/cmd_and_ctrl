package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b780_test.go — Weatherlight, batch 780 (#5629): counters, creatures
// with keywords, lands and token makers. Each test plays the real catalog entry.

const (
	oldB780Bomb      = "061e2ef6-ce8d-4319-b525-96b9c2824001"
	oldB780Fungus    = "3d3c5dd5-ad4e-48ca-92b1-91e1f4ad5c3b"
	oldB780Roc       = "6764d4d9-1fe5-4b42-830c-a3b4c3b37f1f"
	oldB780Infantry  = "543af1f4-f93f-4c7f-a067-ad2db599a4ae"
	oldB780Vultures  = "7fd6fe21-8fb8-499b-9dd4-806d0d660265"
	oldB780Fog       = "b304fcc0-899e-433f-b6df-5cea504dc4ba"
	oldB780Canyons   = "622e2561-48b1-4aca-9abb-9a3c284dcceb"
	oldB780Liege     = "bef7441e-5a44-46bc-8141-8e02ecd06fce"
	oldB780BearType  = "Creature — Bear"
	oldB780FuseKind  = "fuse"
	oldB780ShellKind = "shell"
)

func TestOldB780EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB780Bomb: "Goblin Bomb", oldB780Fungus: "Fungus Elemental", oldB780Roc: "Roc Hatchling",
		oldB780Infantry: "Benalish Infantry", oldB780Vultures: "Circling Vultures", oldB780Fog: "Fog Elemental",
		oldB780Canyons: "Winding Canyons", oldB780Liege: "Liege of the Hollows",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v; want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB780GoblinBombFlipsAtUpkeepAndGrowsOrShrinks(t *testing.T) {
	won, lost, declined := 0, 0, 0
	for table := 0; table < 60 && (won == 0 || lost == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bomb := effectstest.B12Push(g, me.ID, "Goblin Bomb", "Enchantment", oldB780Bomb, 0, 0)
		effectstest.SetCountersOn(g, bomb, map[string]int{oldB780FuseKind: 2})
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if table == 0 {
			// Declining the "may" flips nothing.
			effectstest.AnswerMayChoice(t, g, me.ID, false)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, bomb, oldB780FuseKind); got != 2 || effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall) != nil {
				t.Fatalf("declining still flipped: %d fuse counters", got)
			}
			declined++
			continue
		}
		effectstest.AnswerMayChoice(t, g, me.ID, true)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		flips := len(effectstest.RandomEvents(g, game.EventFlipCoin))
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) != flips+1 {
			t.Fatalf("%d coins flipped, want 1", len(evs)-flips)
		}
		got := effectstest.CountersOn(g, bomb, oldB780FuseKind)
		if evs[len(evs)-1].Won {
			won++
			if got != 3 {
				t.Errorf("a won flip left %d fuse counters, want 3", got)
			}
		} else {
			lost++
			if got != 1 {
				t.Errorf("a lost flip left %d fuse counters, want 1", got)
			}
		}
	}
	if declined == 0 || won == 0 || lost == 0 {
		t.Fatalf("saw %d declined, %d won and %d lost; want all three", declined, won, lost)
	}
}

func TestOldB780GoblinBombTwentyDamageNeedsFiveFuseCountersAndSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bomb := effectstest.B12Push(g, me.ID, "Goblin Bomb", "Enchantment", oldB780Bomb, 0, 0)
	effectstest.SetCountersOn(g, bomb, map[string]int{oldB780FuseKind: 4})
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("the Bomb went off with four fuse counters")
	}
	effectstest.SetCountersOn(g, bomb, map[string]int{oldB780FuseKind: 5})
	before := opp.Life
	effectstest.Pr7Activate(t, g, me.ID, bomb, 0, game.ActivateAbilityParams{Targets: target})
	if g.Battlefield.Contains(bomb) {
		t.Error("the Bomb was not sacrificed")
	}
	if opp.Life != before-20 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, before-20)
	}
}

func TestOldB780FungusElementalGrowsOnlyOnTheTurnItEnteredAndCostsAForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fungus := effectstest.CastHoldCreature(t, g, "Fungus Elemental", "Creature — Fungus Elemental", oldB780Fungus, 3, 3)
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	plains := effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.FloatMana(t, g, me, "{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, fungus, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{plains}}); err == nil {
		t.Error("a Plains paid for the Forest sacrifice")
	}
	effectstest.Pr7Activate(t, g, me.ID, fungus, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{forest}})
	if g.Battlefield.Contains(forest) {
		t.Error("the Forest was not sacrificed")
	}
	if p, tough := effectstest.PtOf(t, g, fungus); p != 5 || tough != 5 {
		t.Errorf("Fungus Elemental is %d/%d, want 5/5", p, tough)
	}
	effectstest.NextTurnOfMine(t, g)
	forest2 := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, fungus, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{forest2}}); err == nil {
		t.Error("Fungus Elemental grew on a later turn")
	}
}

func TestOldB780RocHatchlingEntersShelledAndHatchesAfterFourUpkeeps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	roc := effectstest.CastHoldCreature(t, g, "Roc Hatchling", "Creature — Bird", oldB780Roc, 0, 1)
	if got := effectstest.CountersOn(g, roc, oldB780ShellKind); got != 4 {
		t.Fatalf("it entered with %d shell counters, want 4", got)
	}
	if p, tough := effectstest.PtOf(t, g, roc); p != 0 || tough != 1 {
		t.Errorf("the shelled Hatchling is %d/%d, want 0/1", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, roc, "flying") {
		t.Error("the shelled Hatchling flies")
	}
	for want := 3; want >= 0; want-- {
		effectstest.B12ToMyNextUpkeep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CountersOn(g, roc, oldB780ShellKind); got != want {
			t.Fatalf("after an upkeep it has %d shell counters, want %d", got, want)
		}
		if want > 0 {
			if p, tough := effectstest.PtOf(t, g, roc); p != 0 || tough != 1 {
				t.Errorf("with %d shell counters it is %d/%d, want 0/1", want, p, tough)
			}
		}
	}
	if p, tough := effectstest.PtOf(t, g, roc); p != 3 || tough != 3 {
		t.Errorf("the hatched Roc is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, roc, "flying") {
		t.Error("the hatched Roc does not fly")
	}
}

func TestOldB780BenalishInfantryHasBanding(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	inf := effectstest.B12Push(g, me.ID, "Benalish Infantry", "Creature — Human Soldier", oldB780Infantry, 1, 3)
	effectstest.AssertKeywords(t, g, inf, "banding")
}

func TestOldB780CirclingVulturesFliesDiscardsAtInstantSpeedAndTaxesItsUpkeep(t *testing.T) {
	t.Run("flies, and pays with the top creature card", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		v := effectstest.B12Push(g, me.ID, "Circling Vultures", "Creature — Bird", oldB780Vultures, 3, 2)
		effectstest.AssertKeywords(t, g, v, "flying")
		older := effectstest.PushGraveyardCardTyped(me, "Older Bear", oldB780BearType)
		top := effectstest.PushGraveyardCardTyped(me, "Top Bear", oldB780BearType)
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.UpkeepPromptFor(t, g, me.ID)
		if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{top}); err != nil {
			t.Fatalf("ResolvePayUnlessWithCards: %v", err)
		}
		if !g.Battlefield.Contains(v) || !effectstest.InExile(g, top) || effectstest.InExile(g, older) {
			t.Error("the Vultures should stay and only the top creature card be exiled")
		}
	})
	t.Run("declined", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		v := effectstest.B12Push(g, me.ID, "Circling Vultures", "Creature — Bird", oldB780Vultures, 3, 2)
		effectstest.PushGraveyardCardTyped(me, "Top Bear", oldB780BearType)
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, false)
		if g.Battlefield.Contains(v) {
			t.Error("the Vultures stayed although nothing was exiled")
		}
	})
	t.Run("discard from hand at instant speed", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		card := effectstest.HandCardFull(me, "Circling Vultures", "Creature — Bird", "{B}", oldB780Vultures, []string{"B"})
		if err := g.PerformSpecialAction(me.ID, card, game.SpecialActionDiscard, game.SpecialActionParams{Strict: true}); err != nil {
			t.Fatalf("the discard special action: %v", err)
		}
		if me.Hand.Contains(card) || !me.Graveyard.Contains(card) {
			t.Error("the card was not discarded")
		}
	})
}

func TestOldB780FogElementalIsSacrificedAtEndOfCombatWhenItAttacksOrBlocks(t *testing.T) {
	t.Run("attacks", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		fog := effectstest.B12Push(g, me.ID, "Fog Elemental", "Creature — Elemental", oldB780Fog, 4, 4)
		effectstest.AssertKeywords(t, g, fog, "flying")
		effectstest.DeclareAttack(t, g, opp.ID, fog)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(fog) {
			t.Fatal("the Elemental was sacrificed before end of combat")
		}
		effectstest.AdvanceTo(t, g, game.StepEndCombat)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(fog) {
			t.Error("the Elemental survived end of combat after attacking")
		}
	})
	t.Run("blocks", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		fog := effectstest.B12Push(g, opp.ID, "Fog Elemental", "Creature — Elemental", oldB780Fog, 4, 4)
		atk := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 1, 1)
		effectstest.DeclareAttack(t, g, opp.ID, atk)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(fog, atk); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
		effectstest.LockInBlocks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AdvanceTo(t, g, game.StepEndCombat)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(fog) {
			t.Error("the Elemental survived end of combat after blocking")
		}
	})
	t.Run("idle", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		fog := effectstest.B12Push(g, me.ID, "Fog Elemental", "Creature — Elemental", oldB780Fog, 4, 4)
		effectstest.AdvanceTo(t, g, game.StepEndCombat)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(fog) {
			t.Error("an idle Elemental was sacrificed")
		}
	})
}

func TestOldB780WindingCanyonsLetsCreatureSpellsBeCastAsThoughTheyHadFlashThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	canyons := effectstest.PushLandFor(g, me.ID, "Winding Canyons", "Land")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == canyons {
				g.Battlefield.Cards[i].OracleID = oldB780Canyons
			}
		}
	})
	creature := effectstest.HandCardFull(me, "Test Bear", oldB780BearType, "", "", nil)
	sorcery := effectstest.HandCardFull(me, "Test Sorcery", "Sorcery", "", "", nil)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a creature spell was cast at instant speed with no flash")
	}
	effectstest.FillPool(me, 2)
	effectstest.Pr7Activate(t, g, me.ID, canyons, 0, game.ActivateAbilityParams{})
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a sorcery was cast at instant speed")
	}
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a creature spell could not be cast with flash: %v", err)
	}
}

func TestOldB780LiegeOfTheHollowsLetsEachPlayerPayForTheirOwnSquirrels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liege := effectstest.B12Push(g, me.ID, "Liege of the Hollows", "Creature — Spirit", oldB780Liege, 3, 4)
	effectstest.FillPool(me, 3)
	effectstest.FillPool(opp, 2)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, liege); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, seat := range []struct {
		p    *game.Player
		pays int
	}{{me, 2}, {opp, 1}} {
		c := effectstest.PayAmountPrompt(g, seat.p.ID)
		if c == nil {
			t.Fatalf("seat %v was not asked how much to pay", seat.p.ID)
		}
		if err := g.ResolvePayAmount(c.ID, seat.p.ID, seat.pays); err != nil {
			t.Fatalf("ResolvePayAmount(%d): %v", seat.pays, err)
		}
		effectstest.AnswerPayUnless(t, g, seat.p.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.ManaPool.Total() != 1 || opp.ManaPool.Total() != 1 {
		t.Errorf("pools left %d and %d, want 1 each (3-2 and 2-1)", me.ManaPool.Total(), opp.ManaPool.Total())
	}
	// A player with no mana to pay has nothing to choose, so is not asked.
	for _, p := range g.Seats[2:] {
		if effectstest.PayAmountPrompt(g, p.ID) != nil {
			t.Errorf("seat %v with no mana was asked how much to pay", p.ID)
		}
	}
	if n := effectstest.CountOnBattlefield(g, "Squirrel", me.ID); n != 2 {
		t.Errorf("the controller has %d Squirrels, want 2", n)
	}
	if n := effectstest.CountOnBattlefield(g, "Squirrel", opp.ID); n != 1 {
		t.Errorf("the opponent has %d Squirrels, want 1", n)
	}
	if n := effectstest.CountOnBattlefield(g, "Squirrel", g.Seats[2].ID) + effectstest.CountOnBattlefield(g, "Squirrel", g.Seats[3].ID); n != 0 {
		t.Errorf("players who paid nothing got %d Squirrels", n)
	}
}
