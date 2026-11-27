package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b440_cards_test.go — Dark Ascension Auras, batch 440 (#5571, S82
// OLD). Every test plays the real catalog entry through the engine.

const (
	oldB440Favor      = "285ba9af-bdb8-4b31-b0c5-59aa40d4e3e9"
	oldB440Bloodlet   = "3383a72f-fb83-4b9a-aca6-b736ed929c2e"
	oldB440Exhaustion = "6659fae5-09d9-4095-8de3-643afd6ad3be"
	oldB440Seizer     = "98950aaf-d0b8-420b-9e6a-c7d07d0f8588"
	oldB440Misfortune = "b03c7c8d-8206-45cd-8ead-8f7b35c03a7a"
	oldB440Shadows    = "bc1c2203-4cc6-4131-9d3c-6f3da88b97ce"
	oldB440Kindred    = "c2fb8d4c-b973-42a0-9e9c-7abd5cc2c22e"
	oldB440Echoes     = "c99bac4b-7eec-4fdf-b344-0cf9639ff073"
	oldB440Talons     = "e79b1224-1baf-460f-bf12-92b51cd8071f"
	oldB440Thirst     = "ffe3ad06-50d8-4ca3-a1f7-aa31dc2c486e"
	oldB440Aura       = "Enchantment — Aura"
	oldB440Curse      = "Enchantment — Aura Curse"
	oldB440Bear       = "Creature — Bear"
)

func TestOldB440CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB440Favor: "Favor of the Woods", oldB440Bloodlet: "Curse of Bloodletting", oldB440Exhaustion: "Curse of Exhaustion",
		oldB440Seizer: "Soul Seizer", oldB440Seizer + "#1": "Ghastly Haunting", oldB440Misfortune: "Curse of Misfortunes",
		oldB440Shadows: "Spiteful Shadows", oldB440Kindred: "Call to the Kindred", oldB440Echoes: "Curse of Echoes",
		oldB440Talons: "Talons of Falkenrath", oldB440Thirst: "Curse of Thirst",
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

// oldB440Curse casts the Curse on `player` and resolves it.
func oldB440CastCurse(t *testing.T, g *game.Game, name, oracle string, player uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, name, oldB440Curse, oracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: player}})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, id); c == nil || c.AttachedTo.ID != player {
		t.Fatalf("%s did not attach to the player", name)
	}
	return id
}

func oldB440CastAura(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, name, oldB440Aura, oracle, effectstest.CardTargets(host))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardByID(g, id); c == nil || !c.IsAttachedTo(host) {
		t.Fatalf("%s did not attach to its host", name)
	}
	return id
}

func TestOldB440FavorOfTheWoodsGainsLifeWhenEnchantedCreatureBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB440Bear, 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", oldB440Bear, 2, 2)
	attacker := effectstest.BrCreature(g, me.ID, "Attacker", oldB440Bear, 1, 1)
	oldB440CastAura(t, g, "Favor of the Woods", oldB440Favor, blocker)
	life := me.Life
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: blocker, Attacker: attacker}}); err != nil {
		t.Fatalf("DeclareBlockers: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// The Aura was cast by the active player, so it is theirs: they gain.
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3 for the Aura's controller", life, me.Life)
	}
	_ = other
}

func TestOldB440SpitefulShadowsReflectsDamageToTheCreaturesController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Host", oldB440Bear, 2, 5)
	src := effectstest.B12Creature(g, me.ID, "Source", oldB440Bear, 3, 3)
	oldB440CastAura(t, g, "Spiteful Shadows", oldB440Shadows, host)
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(src, host, 3); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the controller's life %d → %d, want 3 damage", life, opp.Life)
	}
}

func TestOldB440CurseOfExhaustionLimitsOnlyTheEnchantedPlayerToOneSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB440CastCurse(t, g, "Curse of Exhaustion", oldB440Exhaustion, opp.ID)
	// Not the enchanted player: unlimited.
	g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{me.ID: {Total: 3}}
	spec, _ := effects.Lookup(oldB440Exhaustion)
	curse := g.Battlefield.Cards[len(g.Battlefield.Cards)-1]
	forbids := func(who uuid.UUID, cast int) bool {
		g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{who: {Total: cast}}
		return spec.CastRestrictions[0].Forbids(game.CastQuery{Game: g, Controller: who, Source: curse})
	}
	if forbids(me.ID, 3) {
		t.Error("the Curse limited a player it is not attached to")
	}
	if forbids(opp.ID, 0) {
		t.Error("the enchanted player's first spell was refused")
	}
	if !forbids(opp.ID, 1) {
		t.Error("the enchanted player's second spell was allowed")
	}
}

func TestOldB440CurseOfThirstDealsDamageForEachCurseAtTheirUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	oldB440CastCurse(t, g, "Curse of Thirst", oldB440Thirst, opp.ID)
	oldB440CastCurse(t, g, "Curse of Thirst", oldB440Thirst, opp.ID)
	life := opp.Life
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, g.Seats[0].ID)
	effectstest.PassPriorityAroundTable(t, g)
	// Each Curse deals damage equal to the two Curses attached.
	if opp.Life != life-4 {
		t.Errorf("life %d → %d, want 4 (two Curses, 2 damage each)", life, opp.Life)
	}
}

func TestOldB440CurseOfBloodlettingDoublesDamageToTheEnchantedPlayerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.B12Creature(g, me.ID, "Source", oldB440Bear, 3, 3)
	oldB440CastCurse(t, g, "Curse of Bloodletting", oldB440Bloodlet, opp.ID)
	mine, theirs := me.Life, opp.Life
	effectstest.Pr7Hit(t, g, src, opp.ID, 3)
	effectstest.Pr7Hit(t, g, src, me.ID, 3)
	if opp.Life != theirs-6 {
		t.Errorf("enchanted player's life %d → %d, want 6", theirs, opp.Life)
	}
	if me.Life != mine-3 {
		t.Errorf("another player's life %d → %d, want 3", mine, me.Life)
	}
}

func TestOldB440TalonsOfFalkenrathGrantsThePumpToTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB440Bear, 2, 2)
	oldB440CastAura(t, g, "Talons of Falkenrath", oldB440Talons, bear)
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "R", "C")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activating the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 4 {
		t.Errorf("the creature has %d power, want 4", p)
	}
}

func TestOldB440CurseOfMisfortunesFetchesADifferentlyNamedCurse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB440CastCurse(t, g, "Curse of Misfortunes", oldB440Misfortune, opp.ID)
	same := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Curse of Misfortunes", TypeLine: oldB440Curse})
	other := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Curse of Thirst", OracleID: oldB440Thirst, TypeLine: oldB440Curse})
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no library search prompt")
	}
	for _, id := range c.SearchCards {
		if id == same {
			t.Error("a Curse with the same name as an attached Curse was offered")
		}
	}
	effectstest.AnswerSearchByID(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	fetched := effectstest.FindBattlefieldCardByID(g, other)
	if fetched == nil || fetched.AttachedTo.Kind != game.TargetPlayer || fetched.AttachedTo.ID != opp.ID {
		t.Fatal("the fetched Curse is not attached to the enchanted player")
	}
}

func TestOldB440CurseOfEchoesOffersACopyToEachOtherPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	oldB440CastCurse(t, g, "Curse of Echoes", oldB440Echoes, opp.ID)
	// The enchanted player casts an instant: the others (not them) are asked.
	g.WithWriteLock(func() {
		spell := uuid.New()
		g.Stack.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
		g.EmitEvent(game.Event{Kind: game.EventCast, Actor: opp.ID, CardID: spell, Source: spell})
	})
	effectstest.PassPriorityAroundTable(t, g)
	asked := map[uuid.UUID]bool{}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm {
			asked[c.Chooser] = true
		}
	}
	if !asked[me.ID] || !asked[third.ID] {
		t.Errorf("the other players were not each asked: %v", asked)
	}
	if asked[opp.ID] {
		t.Error("the enchanted player was offered a copy")
	}
}

func TestOldB440CallToTheKindredPutsASharedTypeCreatureOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Elf Warrior", 2, 2)
	oldB440CastAura(t, g, "Call to the Kindred", oldB440Kindred, host)
	top := effectstest.SetLibraryTopFirst(me,
		game.Card{Name: "Elf", TypeLine: "Creature — Elf Druid", Power: 1, Toughness: 1},
		game.Card{Name: "Bear", TypeLine: oldB440Bear, Power: 2, Toughness: 2})
	elf, bear := top[0], top[1]
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no choose-cards prompt for the creature to put onto the battlefield")
	}
	for _, id := range c.ChooseCards {
		if id == bear {
			t.Error("a creature sharing no type with the enchanted creature was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(elf) {
		t.Error("the Elf did not enter the battlefield")
	}
}

func TestOldB440SoulSeizerTransformsAndAttachesToACreatureOfThePlayerItHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seizer := oldB440PutSeizer(t, g, me)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB440Bear, 2, 2)
	effectstest.DealCombatDamageToPlayer(g, seizer, opp.ID, 1)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)
	c := effectstest.FindBattlefieldCardByID(g, seizer)
	if c == nil || !c.IsAttachedTo(victim) {
		t.Fatal("the transformed Seizer is not attached to the creature")
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if v := effectstest.FindBattlefieldCardByID(g, victim); v == nil || v.Controller != me.ID {
		t.Error("you do not control the enchanted creature")
	}
}

func TestOldB440SoulSeizerDeclinedStaysACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seizer := oldB440PutSeizer(t, g, me)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB440Bear, 2, 2)
	effectstest.DealCombatDamageToPlayer(g, seizer, opp.ID, 1)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, false)
	if c := effectstest.FindBattlefieldCardByID(g, seizer); c == nil || c.IsAttachedTo(victim) {
		t.Error("the Seizer was attached though the player declined")
	}
}

// oldB440PutSeizer puts the two-faced card on the battlefield face up.
func oldB440PutSeizer(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	row := cards.Card{
		ID: uuid.New(), Name: "Soul Seizer // Ghastly Haunting", Layout: "transform",
		TypeLine: "Creature — Spirit // Enchantment — Aura",
		OracleID: uuid.MustParse(oldB440Seizer), Keywords: []string{"Transform"},
		CardFaces: []cards.CardFace{
			{Name: "Soul Seizer", TypeLine: "Creature — Spirit", Power: "1", Toughness: "3", ManaCost: "{3}{U}{U}"},
			{Name: "Ghastly Haunting", TypeLine: "Enchantment — Aura"},
		},
	}
	id := effectstest.ImportToHand(row, p)
	c, ok := effectstest.CardInZone(p.Hand, id)
	if !ok {
		t.Fatal("card did not reach the hand")
	}
	p.Hand.Remove(id)
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	return id
}
