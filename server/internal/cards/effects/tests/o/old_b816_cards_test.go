package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// old_b816_cards_test.go — S82 OLD batch 816, Alliances Auras (#5634).

const (
	b816Presence = "d6efff61-2087-4f48-a8a8-6c465190dad9"
	b816Pride    = "515fb257-ec07-4978-af16-47a10a22ba3c"
	b816Bones    = "5a747256-4215-4334-98ab-0c2e4ed92e47"
	b816Fury     = "dfc08963-132f-4ebf-9384-15d17d1098f7"
	b816Woods    = "491c6f69-9b75-4d5d-b41f-a14b2fbfdd07"
	b816Voice    = "68d25ba4-d710-4090-bceb-5cfbd6bb63a1"
	b816Chosen   = "7f425406-2746-45a1-86bf-f9645015ffc2"
	b816Plague   = "609c3b1c-d9ad-4616-b685-52b85528fde8"
)

func TestB816CardsAreRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		b816Presence: "Awesome Presence", b816Pride: "Kjeldoran Pride", b816Bones: "Casting of Bones",
		b816Fury: "Bestial Fury", b816Woods: "Gift of the Woods", b816Voice: "Veteran's Voice",
		b816Chosen: "Nature's Chosen", b816Plague: "Krovikan Plague",
	} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// b816Aura puts an Aura onto the battlefield attached to host (no cast, so no
// enters trigger).
func b816Aura(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := effectstest.ApaPush(g, owner, owner, game.Card{Name: name, OracleID: oracle, TypeLine: "Enchantment — Aura"})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatal(err)
		}
	})
	return aura
}

// b816CastAura casts an Aura from the hand onto host and resolves it (so the
// enters trigger runs too).
func b816CastAura(t *testing.T, g *game.Game, me *game.Player, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, OracleID: oracle, TypeLine: "Enchantment — Aura", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}}}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("%s did not enter", name)
	}
	return id
}

func b816ToNextUpkeep(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep {
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestB816AwesomePresenceRefusesABlockThatIsNotPaidFor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b816Aura(t, g, me.ID, "Awesome Presence", b816Presence, atk)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, atk)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, atk); err != nil {
		return // refused up front
	}
	if err := g.FinishBlocks(opp.ID); err == nil {
		t.Fatal("the block stood with no mana to pay {3}")
	}
}

func TestB816AwesomePresenceLetsAPaidBlockStand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b816Aura(t, g, me.ID, "Awesome Presence", b816Presence, atk)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, atk)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.ApaMana(opp, "G", "G", "G", "G")
	if err := g.DeclareBlocker(blocker, atk); err != nil {
		t.Fatalf("a block with {3} to pay: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	if got := effectstest.ApaLive(g, blocker).BlockedAttackers(); len(got) != 1 {
		t.Errorf("the paid block is not standing: %v", got)
	}
	if got := effectstest.PoolSize(g, opp.ID); got != 1 {
		t.Errorf("pool %d, want 1 left (the block cost {3} of four)", got)
	}
}

func TestB816KjeldoranPridePumpsAndMovesToAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Bear", 2, 2)
	two := effectstest.B12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	aura := b816Aura(t, g, me.ID, "Kjeldoran Pride", b816Pride, one)
	if p, tt := effectstest.EffectivePTOf(t, g, one); p != 3 || tt != 4 {
		t.Fatalf("the enchanted creature is %d/%d, want 3/4", p, tt)
	}
	effectstest.ApaMana(me, "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: one}}}); err == nil {
		t.Fatal("moved the Aura onto the creature it already enchants")
	}
	effectstest.P7Activate(t, g, me, aura, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: two}}})
	if p, tt := effectstest.EffectivePTOf(t, g, two); p != 3 || tt != 4 {
		t.Errorf("the new host is %d/%d, want 3/4", p, tt)
	}
	if p, tt := effectstest.EffectivePTOf(t, g, one); p != 2 || tt != 2 {
		t.Errorf("the old host is %d/%d, want 2/2", p, tt)
	}
}

func TestB816CastingOfBonesDrawsThreeThenDiscardsOneOfThemWhenTheHostDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(me)
	host := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b816Aura(t, g, me.ID, "Casting of Bones", b816Bones, host)
	before := me.Hand.Size()
	effectstest.B14Kill(g, host)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	pk := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pk == nil || len(pk.ChooseCards) != 3 {
		t.Fatalf("pick = %+v, want a choice among the three drawn cards", pk)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, pk.ChooseCards[0])
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d, want %d (three drawn, one discarded)", got, before+2)
	}
}

func TestB816BestialFuryDrawsAtTheNextUpkeepAndPumpsWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(me)
	atk := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	b816CastAura(t, g, me, "Bestial Fury", b816Fury, atk)
	effectstest.Pr7bAttack(t, g, opp.ID, atk)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: atk})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, atk); p != 6 {
		t.Errorf("the blocked creature has power %d, want 6", p)
	}
	if c := effectstest.LayeredCard(t, g, atk); !game.HasKeyword(&c, "trample") {
		t.Error("the blocked creature has no trample")
	}
	before := me.Hand.Size()
	b816ToNextUpkeep(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d from the delayed draw", got, before+1)
	}
}

func TestB816GiftOfTheWoodsTriggersWhenBlockedOrBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b816Aura(t, g, me.ID, "Gift of the Woods", b816Woods, atk)
	life := me.Life
	effectstest.Pr7bAttack(t, g, opp.ID, atk)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: atk})
	effectstest.PassPriorityAroundTable(t, g)
	if _, tt := effectstest.EffectivePTOf(t, g, atk); tt != 5 {
		t.Errorf("the blocked creature has toughness %d, want 5", tt)
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func TestB816GiftOfTheWoodsTriggersWhenTheEnchantedCreatureBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b816Aura(t, g, me.ID, "Gift of the Woods", b816Woods, blocker)
	life := me.Life
	effectstest.EndTheTurn(t, g)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(atk, me.ID); err != nil {
		t.Fatalf("declare the attacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, atk); err != nil {
		t.Fatalf("declare the block: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if _, tt := effectstest.EffectivePTOf(t, g, blocker); tt != 5 {
		t.Errorf("the blocker has toughness %d, want 5", tt)
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d (the Aura's controller gains it)", me.Life, life+1)
	}
}

func TestB816VeteransVoiceTapsTheHostToPumpAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	aura := b816Aura(t, g, me.ID, "Veteran's Voice", b816Voice, host)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{host}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}}}); err == nil {
		t.Fatal("pumped the creature tapped for the cost")
	}
	effectstest.P7Activate(t, g, me, aura, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{host}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}})
	if !b816Tapped(g, host) {
		t.Error("the enchanted creature was not tapped")
	}
	if p, tt := effectstest.EffectivePTOf(t, g, other); p != 4 || tt != 3 {
		t.Errorf("the target is %d/%d, want 4/3", p, tt)
	}
	// Tapped now: the cost can't be paid again.
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{host}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}}); err == nil {
		t.Error("activated with the enchanted creature already tapped")
	}
}

func TestB816NaturesChosenUntapsOncePerTurnOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	aura := b816Aura(t, g, me.ID, "Nature's Chosen", b816Chosen, host)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, host).Tapped = true })
	effectstest.P7Activate(t, g, me, aura, 0, game.ActivateAbilityParams{})
	if b816Tapped(g, host) {
		t.Fatal("the enchanted creature was not untapped")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, host).Tapped = true })
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("untapped a second time in one turn")
	}
}

func TestB816NaturesChosenSecondAbilityNeedsAWhiteHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	green := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2})
	white := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "White Knight", TypeLine: "Creature — Knight", Colors: []string{"W"}, Power: 2, Toughness: 2})
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, land).Tapped = true })
	greenAura := b816Aura(t, g, me.ID, "Nature's Chosen", b816Chosen, green)
	whiteAura := b816Aura(t, g, me.ID, "Nature's Chosen", b816Chosen, white)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: land}}
	if err := g.ActivateCatalogAbility(me.ID, greenAura, 1, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{green}, Targets: target}); err == nil {
		t.Fatal("a non-white host paid for the second ability")
	}
	effectstest.P7Activate(t, g, me, whiteAura, 1, game.ActivateAbilityParams{TapIDs: []uuid.UUID{white}, Targets: target})
	if b816Tapped(g, land) {
		t.Error("the land was not untapped")
	}
	if !b816Tapped(g, white) {
		t.Error("the white host was not tapped for the cost")
	}
}

func TestB816KrovikanPlaguePingsTapsAndShrinksTheHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(me)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 3)
	aura := b816CastAura(t, g, me, "Krovikan Plague", b816Plague, host)
	life := opp.Life
	effectstest.P7Activate(t, g, me, aura, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{host}, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Errorf("their life %d, want %d", opp.Life, life-1)
	}
	if !b816Tapped(g, host) {
		t.Error("the host did not tap")
	}
	if c := effectstest.ApaLive(g, host); c == nil || c.CurrentPower() != 2 || c.CurrentToughness() != 2 || c.Counters["-0/-1"] != 1 {
		t.Errorf("the host is %v, want a 2/2 carrying one -0/-1 counter", c)
	}
	before := me.Hand.Size()
	b816ToNextUpkeep(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d from the delayed draw", got, before+1)
	}
}

func TestB816KrovikanPlagueCannotEnchantAWallOrACreatureYouDoNotControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	for _, host := range []uuid.UUID{wall, theirs} {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Krovikan Plague", OracleID: b816Plague, TypeLine: "Enchantment — Aura", Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}}}); err == nil {
			t.Error("cast onto a Wall or a creature the caster does not control")
		}
	}
}

func b816Tapped(g *game.Game, id uuid.UUID) bool {
	c := effectstest.FindBattlefieldCardForTest(g, id)
	return c != nil && c.Tapped
}
