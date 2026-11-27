package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB453Spurnmage = "0477b923-9f32-4e7f-ac08-c715cca819a4"
	oldB453Trap      = "1c6f95fd-f8d0-4d22-aa21-29e5c70b2f85"
	oldB453Cacodemon = "1d34e78e-4fe9-44e1-88d3-a98fd64555a9"
	oldB453Fault     = "4aa34391-d532-48e5-a125-2619d04f27ff"
	oldB453Reiver    = "7dcb95a9-c114-4333-8b45-4380f92d9695"
	oldB453Prophet   = "87c44c23-f209-494a-ab66-c9f5cf7bb4ea"
	oldB453Monk      = "8df64d89-64ee-4388-826d-dd091db4510a"
)

func TestOldB453CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB453Spurnmage: "Spurnmage Advocate", oldB453Trap: "Nemesis Trap", oldB453Cacodemon: "Dread Cacodemon",
		oldB453Fault: "Faultgrinder", oldB453Reiver: "Reiver Demon", oldB453Prophet: "False Prophet",
		oldB453Monk: "Monk Realist",
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

func oldB453Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

func oldB453Zone(g *game.Game, id uuid.UUID) game.ZoneKind { return effectstest.ZoneOfCard(g, id) }

func TestOldB453MonkRealistDestroysATargetEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ench := effectstest.B12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	effectstest.CastCatalogSpell(t, g, "Monk Realist", "Creature — Human Monk Cleric", oldB453Monk, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, ench)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB453Zone(g, ench) != game.ZoneGraveyard {
		t.Errorf("the enchantment is in zone %v, want the graveyard", oldB453Zone(g, ench))
	}
}

func TestOldB453FaultgrinderDestroysALandAndEvokeSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	spare := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	card := effectstest.CastCatalogSpell(t, g, "Faultgrinder", "Creature — Elemental", oldB453Fault, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB453Zone(g, land) != game.ZoneGraveyard || oldB453Zone(g, spare) != game.ZoneBattlefield {
		t.Errorf("land in %v, spare in %v", oldB453Zone(g, land), oldB453Zone(g, spare))
	}
	if oldB453Zone(g, card) != game.ZoneBattlefield {
		t.Errorf("a hard-cast Faultgrinder is in zone %v", oldB453Zone(g, card))
	}
	if !effectstest.HasEffectiveKeyword(t, g, card, "trample") {
		t.Error("no trample")
	}

	// Evoke {4}{R}: it is sacrificed when it enters, and its trigger still destroys.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	land2 := effectstest.B12Permanent(g2, opp2.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g2)
	id := effectstest.B13HandCard(me2, "Faultgrinder", "Creature — Elemental", oldB453Fault)
	effectstest.FloatForTest(g2, me2, "CCCCR")
	if err := g2.CastSpell(me2.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "evoke"}); err != nil {
		t.Fatalf("evoke cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerTriggerTargets(t, g2, me2.ID, land2)
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g2)
	effectstest.PassPriorityAroundTable(t, g2)
	if oldB453Zone(g2, land2) != game.ZoneGraveyard {
		t.Errorf("evoked: the land is in zone %v, want the graveyard", oldB453Zone(g2, land2))
	}
	if oldB453Zone(g2, id) != game.ZoneGraveyard {
		t.Errorf("evoked: Faultgrinder is in zone %v, want the graveyard", oldB453Zone(g2, id))
	}
}

func TestOldB453CacodemonWipesOpponentsAndTapsYoursOnlyWhenCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	theirs2 := effectstest.B12Creature(g, g.Seats[2].ID, "Theirs too", "Creature — Bear", 2, 2)
	demon := effectstest.CastCatalogSpell(t, g, "Dread Cacodemon", "Creature — Demon", oldB453Cacodemon, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB453Zone(g, theirs) != game.ZoneGraveyard || oldB453Zone(g, theirs2) != game.ZoneGraveyard {
		t.Error("an opponent's creature survived")
	}
	if oldB453Zone(g, mine) != game.ZoneBattlefield || oldB453Zone(g, demon) != game.ZoneBattlefield {
		t.Fatal("one of your creatures died")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, mine); !tapped {
		t.Error("your other creature was not tapped")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, demon); tapped {
		t.Error("the Cacodemon tapped itself")
	}

	// Reanimated, it enters quietly.
	bear := effectstest.B12Creature(g, opp.ID, "Another", "Creature — Bear", 2, 2)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dread Cacodemon", TypeLine: "Creature — Demon",
		OracleID: oldB453Cacodemon, Owner: me.ID, Controller: me.ID})
	effectstest.B30Reanimate(t, g, dead, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB453Zone(g, bear) != game.ZoneBattlefield {
		t.Error("a reanimated Cacodemon destroyed creatures")
	}
}

func TestOldB453ReiverDemonSparesBlackAndArtifactCreaturesAndOnlyWhenCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	green := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Green", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	black := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Black", TypeLine: "Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}})
	golem := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 3, Toughness: 3})
	white := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "White", TypeLine: "Creature — Cleric", Power: 1, Toughness: 1, Colors: []string{"W"}})
	effectstest.AdvanceToMain(t, g)
	demon := effectstest.HandCardFull(me, "Reiver Demon", "Creature — Demon", "", oldB453Reiver, []string{"B"})
	if err := g.CastSpell(me.ID, demon, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"green": green, "white": white} {
		if oldB453Zone(g, id) != game.ZoneGraveyard {
			t.Errorf("the %s creature survived", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"black": black, "artifact": golem, "the Demon": demon} {
		if oldB453Zone(g, id) != game.ZoneBattlefield {
			t.Errorf("the %s creature died", name)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, demon, "flying") {
		t.Error("no flying")
	}
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Another", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Reiver Demon", TypeLine: "Creature — Demon",
		OracleID: oldB453Reiver, Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
	effectstest.B30Reanimate(t, g, dead, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB453Zone(g, bear) != game.ZoneBattlefield {
		t.Error("a reanimated Reiver Demon destroyed creatures")
	}
}

func TestOldB453FalseProphetExilesEveryCreatureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prophet := effectstest.B12Push(g, me.ID, "False Prophet", "Creature — Human Cleric", oldB453Prophet, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.KillCreature(t, g, me.ID, prophet)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"yours": mine, "theirs": theirs} {
		if oldB453Zone(g, id) != game.ZoneExile {
			t.Errorf("%s: zone %v, want exile", name, oldB453Zone(g, id))
		}
	}
	if oldB453Zone(g, land) != game.ZoneBattlefield {
		t.Error("a land was exiled")
	}
	if oldB453Zone(g, prophet) != game.ZoneGraveyard {
		t.Errorf("the Prophet is in zone %v, want the graveyard", oldB453Zone(g, prophet))
	}
}

func oldB453Attack(t *testing.T, g *game.Game, me, opp *game.Player, colors ...string) uuid.UUID {
	t.Helper()
	atk := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Attacker", TypeLine: "Creature — Knight", Power: 2, Toughness: 2, Colors: colors})
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	return atk
}

func TestOldB453NemesisTrapExilesTheAttackerAndMakesAShortLivedCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := oldB453Attack(t, g, me, opp, "W")
	trap := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: trap, Name: "Nemesis Trap", TypeLine: "Instant — Trap", ManaCost: "{4}{B}{B}",
		OracleID: oldB453Trap, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "BB")
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{Strict: true, AlternativeCost: "trap", Targets: effectstest.CardTarget(atk)}); err != nil {
		t.Fatalf("trap cast with a white attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB453Zone(g, atk) != game.ZoneExile {
		t.Fatalf("the attacker is in zone %v, want exile", oldB453Zone(g, atk))
	}
	copyID := effectstest.FindBattlefieldByName(g, "Attacker")
	if copyID == uuid.Nil {
		t.Fatal("no token copy was made")
	}
	if c, _ := effectstest.BattlefieldCard(g, copyID); !c.IsToken() || c.Controller != me.ID {
		t.Errorf("the copy is token=%v controlled by %v", c.IsToken(), c.Controller)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Attacker") != uuid.Nil {
		t.Error("the copy was not exiled at the beginning of the next end step")
	}
}

func TestOldB453NemesisTrapNeedsAWhiteAttackerForTheCheapCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := oldB453Attack(t, g, me, opp, "R")
	trap := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: trap, Name: "Nemesis Trap", TypeLine: "Instant — Trap", ManaCost: "{4}{B}{B}",
		OracleID: oldB453Trap, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "BB")
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{Strict: true, AlternativeCost: "trap", Targets: effectstest.CardTarget(atk)}); err == nil {
		t.Fatal("the trap cost was accepted with no white creature attacking")
	}
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(uuid.New())}); err == nil {
		t.Fatal("cast at a target that is not an attacking creature")
	}
}

func TestOldB453SpurnmageReturnsTwoCardsAndDestroysAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adv := effectstest.B12Push(g, me.ID, "Spurnmage Advocate", "Creature — Human Nomad", oldB453Spurnmage, 1, 1)
	oldB453Untap(g, adv)
	a := effectstest.PushGraveyardPermanent(opp, "A", "Creature — Bear", "{1}")
	b := effectstest.PushGraveyardPermanent(opp, "B", "Creature — Bear", "{1}")
	mine := effectstest.PushGraveyardPermanent(me, "Mine", "Creature — Bear", "{1}")
	atk := oldB453Attack(t, g, me, opp)
	two := []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 0},
		{Kind: game.TargetCard, ID: atk, Slot: 1},
	}
	bad := []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: atk, Slot: 1},
	}
	if err := g.ActivateCatalogAbility(me.ID, adv, 0, game.ActivateAbilityParams{Strict: true, Targets: bad}); err == nil {
		t.Fatal("targeted cards from two different graveyards")
	}
	if err := g.ActivateCatalogAbility(me.ID, adv, 0, game.ActivateAbilityParams{Strict: true, Targets: two}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHand(opp, a) || !effectstest.InHand(opp, b) {
		t.Error("the two cards did not return to their owner's hand")
	}
	if oldB453Zone(g, atk) != game.ZoneGraveyard {
		t.Errorf("the attacker is in zone %v, want the graveyard", oldB453Zone(g, atk))
	}
}
