package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b280_cards_test.go — Eternal Masters, batch 280 (#5542). Every test
// plays the real catalog entry through the engine.

const (
	b280RootsOracle      = "0dcee3a3-6cb1-45ce-958c-1f57ced80be4"
	b280GaseousOracle    = "896af9a9-12a2-4dd6-957c-e52150b5f3a2"
	b280ElephantOracle   = "c138bd7f-8751-4e96-b54a-d5082e1a491f"
	b280AnnihilateOracle = "28f55673-7aa0-497e-8434-aa70df1b5e04"
	b280TomeOracle       = "9b7634bf-067b-4b5c-be72-ff1b80247969"
	b280DiminishOracle   = "f9b02a7c-c122-4282-90e2-81fcfa663193"
	b280OriflammeOracle  = "0b16a650-68b0-44dc-a9e1-15b7966e0b18"
	b280NauseaOracle     = "2b8abae2-dd15-41ee-81ae-6a463723b43e"
	b280WelkinOracle     = "0ce31ff1-6cfe-4b34-bd02-7d727934e754"
	b280HookmasterOracle = "e6941577-8262-42ef-bf32-f6505d90b73a"
	b280GlasskiteOracle  = "65b61373-835a-4b22-90d1-d019446e98f7"
	b280GhostOracle      = "cd97f8b4-e751-4b52-987b-a442aef57809"
)

func TestB280EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		b280RootsOracle: "Roots", b280GaseousOracle: "Gaseous Form", b280ElephantOracle: "Elephant Guide",
		b280AnnihilateOracle: "Annihilate", b280TomeOracle: "Emmessi Tome", b280DiminishOracle: "Diminishing Returns",
		b280OriflammeOracle: "Orcish Oriflamme", b280NauseaOracle: "Nausea", b280WelkinOracle: "Welkin Guide",
		b280HookmasterOracle: "Kor Hookmaster", b280GlasskiteOracle: "Jetting Glasskite", b280GhostOracle: "Skulking Ghost",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func b280Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// b280Cast casts a card of the active player's with its real cost from the pool.
func b280Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle, pool string, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, p, pool)
	id := effectstest.HandCardOf(g, p, name, typeLine, cost, oracle)
	return id, g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

func b280MustCast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle, pool string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id, err := b280Cast(t, g, p, name, typeLine, cost, oracle, pool, targets)
	if err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// --- Roots ---

func TestB280RootsTapsTheCreatureAndKeepsItTappedThroughItsControllersUntap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	flier := effectstest.B12Creature(g, opp.ID, "Bird", "Creature — Bird", 1, 1)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, flier).Keywords = append(effectstest.FindBattlefieldCardForTest(g, flier).Keywords, "flying")
		g.BumpLayerVersionForTest()
	})
	if _, err := b280Cast(t, g, me, "Roots", "Enchantment — Aura", "{3}{G}", b280RootsOracle, "CCCG", b280Target(flier)); err == nil {
		t.Fatal("Roots enchanted a creature with flying")
	}
	b280MustCast(t, g, me, "Roots", "Enchantment — Aura", "{3}{G}", b280RootsOracle, "CCCG", b280Target(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LayeredCard(t, g, bear).Tapped {
		t.Fatal("the enchanted creature was not tapped as Roots entered")
	}
	for i := 0; i < 60 && !(g.Turn.Step == game.StepPrecombatMain && g.Turn.ActiveSeat == 1); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != 1 {
		t.Fatal("never reached the opponent's turn")
	}
	if !effectstest.LayeredCard(t, g, bear).Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

// --- Gaseous Form ---

func b280Combat(t *testing.T, g *game.Game, attacker, blocker uuid.UUID) {
	t.Helper()
	opp := g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
}

func b280Gaseous(t *testing.T, attackerP, attackerT, blockerP, blockerT int, enchantAttacker bool) (g *game.Game, attacker, blocker uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker = effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", attackerP, attackerT)
	blocker = effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", blockerP, blockerT)
	host := blocker
	if enchantAttacker {
		host = attacker
	}
	b280MustCast(t, g, me, "Gaseous Form", "Enchantment — Aura", "{2}{U}", b280GaseousOracle, "CCU", b280Target(host))
	effectstest.PassPriorityAroundTable(t, g)
	b280Combat(t, g, attacker, blocker)
	return
}

func TestB280GaseousFormPreventsCombatDamageDealtByTheEnchantedCreature(t *testing.T) {
	g, attacker, blocker := b280Gaseous(t, 5, 5, 2, 2, true)
	if !g.Battlefield.Contains(blocker) {
		t.Error("the blocker died to an enchanted attacker's combat damage")
	}
	if !g.Battlefield.Contains(attacker) {
		t.Error("the attacker died to a 2/2")
	}
}

func TestB280GaseousFormPreventsCombatDamageDealtToTheEnchantedCreature(t *testing.T) {
	g, attacker, blocker := b280Gaseous(t, 1, 1, 5, 5, true)
	if !g.Battlefield.Contains(attacker) {
		t.Error("the enchanted attacker died to a 5/5's combat damage")
	}
	if !g.Battlefield.Contains(blocker) {
		t.Error("the blocker died though the enchanted creature deals no damage")
	}
}

func TestB280GaseousFormOnABlockerWorksTheSameWay(t *testing.T) {
	g, attacker, blocker := b280Gaseous(t, 3, 3, 3, 3, false)
	if !g.Battlefield.Contains(attacker) || !g.Battlefield.Contains(blocker) {
		t.Error("an enchanted blocker and its attacker should both survive")
	}
}

func TestB280GaseousFormDoesNotPreventNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b280MustCast(t, g, me, "Gaseous Form", "Enchantment — Aura", "{2}{U}", b280GaseousOracle, "CCU", b280Target(bear))
	effectstest.PassPriorityAroundTable(t, g)
	b280MustCast(t, g, me, "Slaying Fire", "Instant", "{2}{R}", effectstest.SlayingFireOracle, "CCR", b280Target(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("noncombat damage was prevented")
	}
}

// --- Elephant Guide ---

func TestB280ElephantGuidePumpsAndLeavesAnElephantWhenTheCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b280MustCast(t, g, me, "Elephant Guide", "Enchantment — Aura", "{2}{G}", b280ElephantOracle, "CCG", b280Target(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Fatalf("enchanted creature is %d/%d, want 5/5", p, tough)
	}
	effectstest.KillCreature(t, g, opp.ID, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Elephant"); got != 1 {
		t.Errorf("the Aura's controller has %d Elephants, want 1", got)
	}
	if got := effectstest.CountBattlefieldNamed(g, opp.ID, "Elephant"); got != 0 {
		t.Errorf("the creature's controller got %d Elephants, want 0", got)
	}
	toks := effectstest.TokensNamed(g, "Elephant")
	if len(toks) != 1 {
		t.Fatalf("%d Elephant tokens", len(toks))
	}
	if p, tough := effectstest.PtOf(t, g, toks[0].InstanceID); p != 3 || tough != 3 {
		t.Errorf("the Elephant is %d/%d, want 3/3", p, tough)
	}
}

// --- Annihilate ---

func TestB280AnnihilateDestroysANonblackCreatureAndDrawsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Shade", TypeLine: "Creature — Shade", Power: 1, Toughness: 1,
		Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Colors: []string{"G"}, Owner: opp.ID, Controller: opp.ID})
	if _, err := b280Cast(t, g, me, "Annihilate", "Instant", "{3}{B}{B}", b280AnnihilateOracle, "CCCBB", b280Target(black)); err == nil {
		t.Fatal("Annihilate targeted a black creature")
	}
	hand := len(me.Hand.Cards)
	b280MustCast(t, g, me, "Annihilate", "Instant", "{3}{B}{B}", b280AnnihilateOracle, "CCCBB", b280Target(green))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the nonblack creature survived")
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d -> %d, want one drawn card", hand, len(me.Hand.Cards))
	}
}

// --- Emmessi Tome ---

func TestB280EmmessiTomeDrawsTwoThenDiscardsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tome := effectstest.PushCatalogPermanent(g, me.ID, "Emmessi Tome", "Artifact — Book", b280TomeOracle, false)
	effectstest.FloatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, tome, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("four mana paid for {5}")
	}
	effectstest.FloatForTest(g, me, "C")
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, tome, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrDiscard(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d -> %d, want +1 (two drawn, one discarded)", hand, len(me.Hand.Cards))
	}
	if !effectstest.LayeredCard(t, g, tome).Tapped {
		t.Error("the Tome is not tapped")
	}
}

// --- Diminishing Returns ---

func TestB280DiminishingReturnsShufflesExilesTenAndLetsEachPlayerDrawUpToSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mk := func(p *game.Player, n int) {
		g.WithWriteLock(func() {
			p.Library.Cards = nil
			for i := 0; i < n; i++ {
				p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lib", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
			}
		})
	}
	mk(me, 20)
	mk(opp, 20)
	for _, p := range g.Seats {
		if p != me && p != opp {
			mk(p, 20)
		}
	}
	oppGrave := effectstest.PushCatalogGraveyardCard(opp, "Old", "Sorcery", "", 0, 0)
	oppHand := effectstest.HandCard(opp, "Held", "Sorcery")
	meGrave := effectstest.PushCatalogGraveyardCard(me, "Gone", "Sorcery", "", 0, 0)
	owned := len(me.Hand.Cards) + len(me.Graveyard.Cards)
	b280MustCast(t, g, me, "Diminishing Returns", "Sorcery", "{2}{U}{U}", b280DiminishOracle, "CCUU", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Exile.Cards) != 10 {
		t.Fatalf("%d cards exiled, want the top ten of the caster's library", len(g.Exile.Cards))
	}
	for _, c := range g.Exile.Cards {
		if c.Owner != me.ID {
			t.Errorf("an exiled card belongs to someone else")
		}
	}
	for _, id := range []uuid.UUID{oppGrave, meGrave} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind == game.ZoneGraveyard {
			t.Errorf("a graveyard card was not shuffled into its library")
		}
	}
	if opp.Hand.Contains(oppHand) {
		t.Error("the opponent's hand card was not shuffled away")
	}
	counts := map[uuid.UUID]int{me.ID: 7, opp.ID: 3}
	for _, p := range g.Seats {
		n, ok := counts[p.ID]
		if !ok {
			n = 0
		}
		pr := effectstest.PayAmountPrompt(g, p.ID)
		if pr == nil {
			t.Fatalf("no number prompt for %s", p.ID)
		}
		effectstest.AnswerNumber(t, g, p.ID, n)
	}
	if len(me.Hand.Cards) != 7 {
		t.Errorf("caster hand %d, want 7", len(me.Hand.Cards))
	}
	if len(opp.Hand.Cards) != 3 {
		t.Errorf("opponent hand %d, want 3", len(opp.Hand.Cards))
	}
	if want := 20 + owned - 10 - 7; len(me.Library.Cards) != want {
		t.Errorf("caster library %d, want %d", len(me.Library.Cards), want)
	}
}

// --- Orcish Oriflamme ---

func TestB280OrcishOriflammePumpsOnlyYourAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Orcish Oriflamme", "Enchantment", b280OriflammeOracle, false)
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if p, _ := effectstest.PtOf(t, g, atk); p != 2 {
		t.Fatalf("a creature that is not attacking has power %d, want 2", p)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if p, _ := effectstest.PtOf(t, g, atk); p != 3 {
		t.Errorf("the attacker has power %d, want 3", p)
	}
	if p, _ := effectstest.PtOf(t, g, idle); p != 2 {
		t.Errorf("a non-attacker has power %d, want 2", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("an opposing creature has power %d, want 2", p)
	}
}

// --- Nausea ---

func TestB280NauseaGivesEveryCreatureMinusOneMinusOneUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	b280MustCast(t, g, me, "Nausea", "Sorcery", "{1}{B}", b280NauseaOracle, "CB", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("my creature is %d/%d, want 2/2", p, tough)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the opposing 1/1 survived")
	}
	late := effectstest.B12Creature(g, me.ID, "Late", "Creature — Bear", 3, 3)
	if p, _ := effectstest.PtOf(t, g, late); p != 3 {
		t.Errorf("a creature that arrived later has power %d, want 3 (the effect fixes its set)", p)
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 {
		t.Errorf("power %d after end of turn, want 3", p)
	}
}

// --- Welkin Guide ---

func TestB280WelkinGuideGivesATargetPlusTwoPlusTwoAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	guide := b280MustCast(t, g, me, "Welkin Guide", "Creature — Bird Cleric", "{4}{W}", b280WelkinOracle, "CCCCW", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(guide) {
		t.Fatal("the Guide did not enter")
	}
	effectstest.AssertKeywords(t, g, guide, "flying")
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	effectstest.AssertKeywords(t, g, bear, "flying")
}

// --- Kor Hookmaster ---

func TestB280KorHookmasterTapsAnOpposingCreatureAndFreezesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	b280MustCast(t, g, me, "Kor Hookmaster", "Creature — Kor Soldier", "{2}{W}", b280HookmasterOracle, "CCW", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LayeredCard(t, g, theirs).Tapped {
		t.Fatal("the opposing creature was not tapped")
	}
	if effectstest.LayeredCard(t, g, mine).Tapped {
		t.Error("my own creature was tapped")
	}
	for i := 0; i < 60 && !(g.Turn.Step == game.StepPrecombatMain && g.Turn.ActiveSeat == 1); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if !effectstest.LayeredCard(t, g, theirs).Tapped {
		t.Error("the creature untapped in its controller's next untap step")
	}
}

// --- Jetting Glasskite ---

func TestB280GlasskiteCountersOnlyTheFirstTargetingEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kite := effectstest.B12Push(g, opp.ID, "Jetting Glasskite", "Creature — Spirit", b280GlasskiteOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)
	first := b280MustCast(t, g, me, "Slaying Fire", "Instant", "{2}{R}", effectstest.SlayingFireOracle, "CCR", b280Target(kite))
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(first); z == nil || z.Kind != game.ZoneGraveyard {
		t.Errorf("the first spell is in %v, want the graveyard (countered)", z)
	}
	if got := effectstest.LayeredCard(t, g, kite).DamageMarked; got != 0 {
		t.Errorf("the Glasskite has %d damage after a countered spell", got)
	}
	b280MustCast(t, g, me, "Slaying Fire", "Instant", "{2}{R}", effectstest.SlayingFireOracle, "CCR", b280Target(kite))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, kite).DamageMarked; got != 3 {
		t.Errorf("the Glasskite has %d damage, want 3 from the second spell this turn", got)
	}
}

// --- Skulking Ghost ---

func TestB280SkulkingGhostIsSacrificedWhenAnyoneTargetsIt(t *testing.T) {
	for _, owner := range []int{0, 1} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		ghost := effectstest.B12Push(g, g.Seats[owner].ID, "Skulking Ghost", "Creature — Spirit", b280GhostOracle, 2, 1)
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, ghost).Toughness = 9 })
		b280MustCast(t, g, me, "Slaying Fire", "Instant", "{2}{R}", effectstest.SlayingFireOracle, "CCR", b280Target(ghost))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(ghost) {
			t.Errorf("owner seat %d: the Ghost survived being targeted", owner)
		}
	}
}
