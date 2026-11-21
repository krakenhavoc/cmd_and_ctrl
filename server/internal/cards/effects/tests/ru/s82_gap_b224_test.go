package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b224_test.go — Commander Legends: Battle for Baldur's Gate
// creatures, GAP batch 224 (#5948). Every test plays the real catalog entry
// through the engine.

const (
	gap224Invoker    = "daa47749-7a20-4e64-b476-a44eaeb4d844"
	gap224Angels     = "c2bd0c24-3973-48bc-a591-3779e3640c5d"
	gap224Hippogriff = "3372b994-47de-4e89-9bfc-76e1cb3f1cfa"
	gap224Crystal    = "56b0cb64-b886-413e-bb20-392bf12d3926"
	gap224Ellyn      = "ca5ecc26-5ed8-4466-bb8f-df2b39bddd18"
	gap224Officer    = "bf06b2a8-2bb5-4eae-8306-d051c5d9f462"
	gap224Galepowder = "b2457982-6c28-47e9-9b32-9db359eb0ac5"
	gap224Githzerai  = "d4d3cba8-562d-4208-ab0a-c82b4dd8e959"
	gap224Protector  = "6bdc4996-f7d9-4dc4-b70a-aa63f6816b1d"
)

func TestGap224EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gap224Invoker: "Bane's Invoker", gap224Angels: "Battle Angels of Tyr",
		gap224Hippogriff: "Blessed Hippogriff", gap224Hippogriff + "#1": "Tyr's Blessing",
		gap224Crystal: "Crystal Dragon", gap224Crystal + "#1": "Rob the Hoard", gap224Ellyn: "Ellyn Harbreeze, Busybody",
		gap224Officer: "Flaming Fist Officer", gap224Galepowder: "Galepowder Mage", gap224Githzerai: "Githzerai Monk",
		gap224Protector: "Glorious Protector",
	}
	for oracle, name := range want {
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

func TestGap224BanesInvokerPumpsTwoCreaturesForEightMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	inv := effectstest.PushCatalogPermanent(g, me.ID, "Bane's Invoker", "Creature — Human Cleric", gap224Invoker, false)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}}}); err == nil {
		t.Fatal("activated for seven mana")
	}
	effectstest.ApaMana(me, "C")
	err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}})
	if err == nil {
		t.Fatal("three targets were accepted for an \"up to two\" clause")
	}
	effectstest.ApaMana(me, "C", "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, inv, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 4 {
			t.Errorf("%d/%d, want 4/4", p, tough)
		}
		if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), "flying") {
			t.Error("no flying")
		}
	}
	if p, _ := effectstest.PtOf(t, g, c); p != 2 {
		t.Errorf("the untargeted creature has power %d, want 2", p)
	}
}

func gap224Angel(t *testing.T) (*game.Game, uuid.UUID, *game.Player, []*game.Player) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	angel := effectstest.B12Push(g, me.ID, "Battle Angels of Tyr", "Creature — Angel Knight", gap224Angels, 4, 4)
	others := append([]*game.Player(nil), g.Seats[1:]...)
	return g, angel, me, others
}

func gap224Hand(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

func gap224Lands(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushCatalogPermanent(g, p.ID, "Forest", "Basic Land — Forest", "", false)
	}
}

func TestGap224BattleAngelsOfTyrPaysOffAgainstTheLeader(t *testing.T) {
	g, angel, me, others := gap224Angel(t)
	victim := others[0]
	// The victim leads every table statistic; everyone else is behind.
	gap224Hand(victim, 8)
	gap224Lands(g, victim, 4)
	gap224Lands(g, others[1], 1)
	g.WithWriteLock(func() { victim.Life = 60 })
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand, life := me.Hand.Size(), me.Life
	effectstest.DealCombatDamageToPlayer(g, angel, victim.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if n := effectstest.OnBattlefieldNamed(g, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
	if me.Life-life != 3 {
		t.Errorf("gained %d life, want 3", me.Life-life)
	}
}

func TestGap224BattleAngelsOfTyrDoesNothingAgainstAPlayerWhoIsNotAhead(t *testing.T) {
	g, angel, me, others := gap224Angel(t)
	victim := others[0]
	// Ties are not "more": every statistic is level or lower for the victim.
	gap224Hand(victim, 1)
	gap224Hand(me, 1)
	gap224Lands(g, victim, 1)
	gap224Lands(g, me, 1)
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand, life := me.Hand.Size(), me.Life
	effectstest.DealCombatDamageToPlayer(g, angel, victim.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d with a tie, want 0", got)
	}
	if n := effectstest.OnBattlefieldNamed(g, "Treasure"); n != 0 {
		t.Errorf("%d Treasures with a tie, want 0", n)
	}
	if me.Life != life {
		t.Errorf("life changed by %d with a tie, want 0", me.Life-life)
	}
}

func TestGap224BattleAngelsOfTyrHasMyriad(t *testing.T) {
	g, angel, me, _ := gap224Angel(t)
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, angel)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Battle Angels of Tyr" && c.InstanceID != angel && effects.IsToken(c) {
			copies++
		}
	}
	if copies != 2 {
		t.Errorf("%d myriad copies, want 2", copies)
	}
}

func TestGap224BlessedHippogriffGivesAnAttackerFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hippo := effectstest.B12Push(g, me.ID, "Blessed Hippogriff", "Creature — Hippogriff", gap224Hippogriff, 2, 3)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{hippo, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: hippo}); err == nil {
		t.Error("the Hippogriff (which already has flying) was accepted as a target")
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "flying") {
		t.Error("the Bear did not gain flying")
	}
}

func TestGap224TyrsBlessingGivesIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Tyr's Blessing", TypeLine: "Instant — Adventure", OracleID: gap224Hippogriff + "#1",
		ManaCost: "{W}", Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "W")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "indestructible") {
		t.Error("the Bear is not indestructible")
	}
}

func TestGap224RobTheHoardReturnsLegendaryArtifactsAndEnchantmentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	legend := uuid.New()
	rock := uuid.New()
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: legend, Name: "Old Legend", TypeLine: "Legendary Creature — Human", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: rock, Name: "Old Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	cast := func(target uuid.UUID, strict bool) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rob the Hoard", TypeLine: "Sorcery — Adventure", OracleID: gap224Crystal + "#1",
			ManaCost: "{1}{W}", Owner: me.ID, Controller: me.ID})
		effectstest.ApaMana(me, "W", "C")
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: strict, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
	}
	if err := cast(bear, true); err == nil {
		t.Fatal("a plain creature card was accepted")
	}
	for _, id := range []uuid.UUID{legend, rock} {
		if err := cast(id, true); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(id) {
			t.Errorf("%s did not return to the hand", id)
		}
	}
}

func TestGap224EllynLooksAtAsManyCardsAsTokensCreatedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ellyn := effectstest.PushCatalogPermanent(g, me.ID, "Ellyn Harbreeze, Busybody", "Legendary Creature — Human Peasant", gap224Ellyn, false)
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	// No tokens yet: nothing happens, but the cost is paid.
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, ellyn, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand changed with no tokens created")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, ellyn).Tapped = false })
	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 2)
	if err := g.ActivateCatalogAbility(me.ID, ellyn, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the picked card is not in the hand")
	}
	if me.Hand.Contains(ids[0]) {
		t.Error("the other looked-at card went to the hand")
	}
}

func TestGap224FlamingFistOfficerGrowsWhenAnotherCreatureOfYoursLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	officer := effectstest.B12Push(g, me.ID, "Flaming Fist Officer", "Creature — Gnome Soldier", gap224Officer, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	gap224Leave(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, officer); n != 0 {
		t.Fatalf("%d counters after an opponent's creature left, want 0", n)
	}
	gap224Leave(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOnes(g, officer); n != 1 {
		t.Errorf("%d counters after my creature left, want 1", n)
	}
}

func TestGap224GalepowderMageBlinksAnotherCreatureAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Galepowder Mage", "Creature — Kithkin Wizard", gap224Galepowder, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, mage)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mage}); err == nil {
		t.Error("the Mage targeted itself")
	}
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature was not exiled")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Victim" && c.Controller == opp.ID {
			found = true
		}
	}
	if !found {
		t.Error("the creature did not return under its owner's control at the end step")
	}
}

func TestGap224GithzeraiMonkTapsEveryCreatureYouDontControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Githzerai Monk", "Creature — Gith Monk", gap224Githzerai, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(theirs); !c.Tapped {
		t.Error("their creature is untapped")
	}
	if c, _ := g.LookupCardForEffect(mine); c.Tapped {
		t.Error("my creature was tapped")
	}
}

func TestGap224GloriousProtectorExilesNonAngelsUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	angel := effectstest.B12Creature(g, me.ID, "Little Angel", "Creature — Angel", 2, 2)
	keep := effectstest.B12Creature(g, me.ID, "Kept", "Creature — Bear", 2, 2)
	protector := effectstest.CastHoldCreature(t, g, "Glorious Protector", "Creature — Angel Cleric", gap224Protector, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear was not exiled")
	}
	if !g.Battlefield.Contains(angel) || !g.Battlefield.Contains(keep) {
		t.Error("an Angel or an unpicked creature left the battlefield")
	}
	// When the Protector leaves, the Bear returns.
	gap224Leave(t, g, protector)
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" {
			found = true
		}
	}
	if !found {
		t.Error("the Bear did not return when the Protector left")
	}
}

func TestGap224GloriousProtectorHasFlashFlyingAndForetell(t *testing.T) {
	spec, ok := effects.Lookup(gap224Protector)
	if !ok {
		t.Fatal("not registered")
	}
	if len(spec.SpecialActions) != 1 {
		t.Errorf("%d special actions, want the foretell action", len(spec.SpecialActions))
	}
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Glorious Protector", "Creature — Angel Cleric", gap224Protector, 3, 4)
	abilities := effectstest.EffectiveAbilities(t, g, id)
	for _, kw := range []string{"flash", "flying"} {
		if !effectstest.HasAbility(abilities, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

// gap224Leave moves a permanent from the battlefield to its owner's hand, the
// way a bounce does.
func gap224Leave(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var owner uuid.UUID
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			owner = c.Owner
		}
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: owner}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
}
