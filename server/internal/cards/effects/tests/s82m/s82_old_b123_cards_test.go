package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander 2020, OLD batch 123: card draw, cycling and library.

const (
	oldB123Nimble     = "750d846e-2774-4296-8d9f-5207477d842d"
	oldB123Drift      = "7aa5db50-5057-45fd-bce9-fb4a5e874263"
	oldB123Decoy      = "94d9878e-9786-438c-8b33-c1dff730b748"
	oldB123Mystic     = "9652f79c-3060-4e9c-9623-0ebe4ab4094c"
	oldB123Fireflux   = "d53fecd7-aa12-482e-8f45-5efe6348070d"
	oldB123Perspect   = "e1b8a89a-b308-4fc5-ada7-61de6e21b7f9"
	oldB123Gladiator  = "4da2425d-2118-4116-855d-2ef80693a7b7" // creature with cycling {1}{B}
	oldB123GladTL     = "Creature — Zombie Gladiator"
	oldB123BearLine   = "Creature — Bear"
	oldB123SorceryTL  = "Sorcery"
	oldB123InstantTL  = "Instant"
	oldB123AnyTargets = 0
)

func TestOldB123CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB123Nimble: "Nimble Obstructionist", oldB123Drift: "Astral Drift", oldB123Decoy: "Decoy Gambit",
		oldB123Mystic: "Lunar Mystic", oldB123Fireflux: "Fireflux Squad", oldB123Perspect: "New Perspectives",
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

// oldB123Drive answers whatever prompts are open, in a loop, until none is:
// a target prompt with `target`, an optional-trigger or confirm prompt with
// `yes`, and lets the stack settle between answers.
func oldB123Drive(t *testing.T, g *game.Game, chooser uuid.UUID, yes bool, targets ...uuid.UUID) {
	t.Helper()
	for i := 0; i < 20; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
			}
		}
		if open == nil {
			return
		}
		var err error
		switch open.Kind {
		case game.PendingChoicePickTarget:
			refs := make([]game.TargetRef, 0, len(targets))
			for _, id := range targets {
				refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
			}
			err = g.ResolvePickTargets(open.ID, open.Chooser, refs)
		case game.PendingChoiceTriggerPrompt:
			err = g.ResolveTriggerPrompt(open.ID, open.Chooser, yes)
		case game.PendingChoiceConfirm:
			err = g.ResolveConfirm(open.ID, open.Chooser, yes)
		default:
			t.Fatalf("an unexpected prompt is open: %s", open.Kind)
		}
		if err != nil {
			t.Fatalf("answering %s: %v", open.Kind, err)
		}
	}
	t.Fatal("prompts kept coming")
}

func TestOldB123LunarMysticPaysOneToDrawOffAnInstantOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Lunar Mystic", "Creature — Human Wizard", oldB123Mystic, 2, 2)
	before := me.Hand.Size()
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Quick Thing", TypeLine: oldB123InstantTL}, game.CastSpellParams{})
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("casting an instant did not offer the payment")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand grew by %d, want 1", me.Hand.Size()-before)
	}
	// A sorcery does not trigger it.
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Slow Thing", TypeLine: oldB123SorceryTL}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a sorcery offered the payment")
	}
}

func TestOldB123NimbleObstructionistCountersAnOpponentsAbilityWhenCycled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	boiler := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.StCard("Plague Boiler", "fef502af-6e79-4c55-a86a-b45adb3fc64a", "Artifact", 0, 0))
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: opp.ID}) })
	g.RunStateChecksForTest()
	var item uuid.UUID
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemTriggered && it.SourceCardID == boiler {
			item = it.ID
		}
	}
	if item == uuid.Nil {
		t.Fatal("the opponent's trigger is not on the stack")
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Nimble Obstructionist", TypeLine: "Creature — Bird Wizard", OracleID: oldB123Nimble, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	oldB123Drive(t, g, me.ID, true, item)
	if n := effectstest.CounterCount(g, boiler, "plague"); n != 0 {
		t.Errorf("the countered trigger still put a plague counter on the Boiler (%d)", n)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled card is not in the graveyard")
	}
}

func TestOldB123AstralDriftBlinksOnAnyCycle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Astral Drift", oldB123Drift, "Enchantment", 0, 0))
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB123BearLine, 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Undead Gladiator", TypeLine: oldB123GladTL, OracleID: oldB123Gladiator, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	oldB123Drive(t, g, me.ID, true, bear)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	oldB123Drive(t, g, me.ID, true)
	back := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" && c.Owner == opp.ID {
			back = true
		}
	}
	if !back {
		t.Error("the creature did not return at the next end step")
	}
}

func TestOldB123AstralDriftTriggersOffItselfBeingCycled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB123BearLine, 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Astral Drift", TypeLine: "Enchantment", OracleID: oldB123Drift, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	oldB123Drive(t, g, me.ID, true, bear)
	if g.Battlefield.Contains(bear) {
		t.Error("cycling Astral Drift itself did not exile the creature")
	}
}

func TestOldB123DecoyGambitBouncesUnlessTheirControllerLetsYouDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	a := effectstest.B12Creature(g, o1.ID, "Bear A", oldB123BearLine, 2, 2)
	b := effectstest.B12Creature(g, o2.ID, "Bear B", oldB123BearLine, 2, 2)
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Decoy Gambit", oldB123InstantTL, oldB123Decoy, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	answers := map[uuid.UUID]bool{o1.ID: true, o2.ID: false}
	for i := 0; i < 4; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceConfirm {
				open = c
			}
		}
		if open == nil {
			break
		}
		if err := g.ResolveConfirm(open.ID, open.Chooser, answers[open.Chooser]); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
	}
	if !g.Battlefield.Contains(a) {
		t.Error("the creature whose controller let me draw was returned")
	}
	if g.Battlefield.Contains(b) {
		t.Error("the creature whose controller declined stayed")
	}
	if me.Hand.Size() != before+1 {
		t.Errorf("hand changed by %d, want +1 for the one draw", me.Hand.Size()-before)
	}
}

func TestOldB123FirefluxSquadSwapsAnAttackerForARevealedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squad := effectstest.B12Push(g, me.ID, "Fireflux Squad", "Creature — Human Soldier", oldB123Fireflux, 4, 3)
	other := effectstest.B12Creature(g, me.ID, "Other Attacker", oldB123BearLine, 2, 2)
	me.Library.Cards = nil
	filler := uuid.New()
	hit := uuid.New()
	skipA := uuid.New()
	// Bottom first: the library's top is the last element.
	me.Library.PushTop(game.Card{InstanceID: filler, Name: "Bottom Land", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: hit, Name: "Hit Elk", TypeLine: "Creature — Elk", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: skipA, Name: "Top Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.DeclareAttack(t, g, opp.ID, squad, other)
	oldB123Drive(t, g, me.ID, true, other)
	if g.Battlefield.Contains(other) {
		t.Fatal("the exiled attacker is still on the battlefield")
	}
	var elk *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == hit {
			elk = &g.Battlefield.Cards[i]
		}
	}
	if elk == nil {
		t.Fatal("the revealed creature did not enter")
	}
	if !elk.Tapped || elk.AttackingTarget != opp.ID {
		t.Errorf("the creature entered tapped=%v attacking=%v, want tapped and attacking the opponent", elk.Tapped, elk.AttackingTarget)
	}
	if len(me.Library.Cards) != 2 {
		t.Errorf("library holds %d cards, want the 2 revealed non-hits on the bottom", len(me.Library.Cards))
	}
}

func TestOldB123NewPerspectivesDrawsThreeAndMakesCyclingFreeAtSevenCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	me.Hand.Cards = nil
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "New Perspectives", TypeLine: "Enchantment", OracleID: oldB123Perspect}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := me.Hand.Size(); n != 3 {
		t.Fatalf("hand holds %d cards after the draw, want 3", n)
	}
	// Seven cards in hand: cycling for {0}.
	for me.Hand.Size() < 6 {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Undead Gladiator", TypeLine: oldB123GladTL, OracleID: oldB123Gladiator, Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the printed cycling cost was payable with an empty pool")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, game.AltEquipIndexBase+1, game.ActivateAbilityParams{Strict: true, Ref: game.AltEquipRefPrefix + "own:1"}); err != nil {
		t.Fatalf("the {0} cycling twin was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled card was not discarded")
	}
	// Six cards left: the alternative is closed.
	id2 := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id2, Name: "Undead Gladiator", TypeLine: oldB123GladTL, OracleID: oldB123Gladiator, Owner: me.ID, Controller: me.ID})
	me.Hand.Cards = me.Hand.Cards[len(me.Hand.Cards)-3:]
	if err := g.ActivateCatalogAbility(me.ID, id2, game.AltEquipIndexBase+1, game.ActivateAbilityParams{Strict: true, Ref: game.AltEquipRefPrefix + "own:1"}); err == nil {
		t.Error("the {0} twin worked with fewer than seven cards in hand")
	}
}
