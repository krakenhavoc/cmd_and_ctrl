package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestAbilityCopyCardsAreWired — the declarations, before any board.
// Each card's clause is what its oracle text says and nothing wider.
func TestAbilityCopyCardsAreWired(t *testing.T) {
	res := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.StrionicResonatorOracle})
	if len(res) != 1 || res[0].Targets == nil || !res[0].Targets.Abilities {
		t.Fatalf("Strionic Resonator: one ability targeting an ability item, got %+v", res)
	}
	if !res[0].Cost.Tap || res[0].Cost.Mana != "{2}" {
		t.Errorf("Strionic Resonator's cost is {2}, {T}: %+v", res[0].Cost)
	}

	eng := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.LithoformEngineOracle})
	if len(eng) != 3 {
		t.Fatalf("Lithoform Engine prints three abilities, got %d", len(eng))
	}
	if eng[0].Targets == nil || !eng[0].Targets.Abilities {
		t.Error("Lithoform Engine's first ability targets an ability item")
	}
	for i, want := range []string{"{2}", "{3}", "{4}"} {
		if eng[i].Cost.Mana != want || !eng[i].Cost.Tap {
			t.Errorf("Lithoform Engine ability %d costs %q, {T}: %+v", i, want, eng[i].Cost)
		}
	}

	rings := game.TriggersForCard(game.Card{OracleID: effectstest.RingsOfBrighthearthOracle})
	if len(rings) != 1 {
		t.Fatalf("Rings of Brighthearth declares one trigger, got %d", len(rings))
	}
	// "If it isn't a mana ability" is the event KIND, not a
	// predicate: a mana ability announces EventManaAbilityActivated.
	if len(rings[0].Watches) != 1 || rings[0].Watches[0] != game.EventActivateAbility {
		t.Errorf("the Rings watch only the non-mana activation: %v", rings[0].Watches)
	}
	if len(game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.RingsOfBrighthearthOracle})) != 0 {
		t.Error("the Rings have no activated ability of their own")
	}

	// Scrap Trawler's clause is a fact about the event — a target,
	// built from it rather than from a static spec.
	trawler := game.TriggersForCard(game.Card{OracleID: effectstest.ScrapTrawlerOracle})
	if len(trawler) != 1 {
		t.Fatalf("Scrap Trawler: one trigger, got %d", len(trawler))
	}
	if trawler[0].TargetsFrom == nil || trawler[0].Targets != nil {
		t.Error("Scrap Trawler: the clause must come from the event, not from a static spec")
	}

	// Cloudstone Curio's bounce is not a target at all (#1337): no
	// "target" in the printed text, so the pick is a resolution-time
	// ChoosePermanents call inside the Effect, and the trigger itself
	// declares neither Targets nor TargetsFrom.
	curio := game.TriggersForCard(game.Card{OracleID: effectstest.CloudstoneCurioOracle})
	if len(curio) != 1 {
		t.Fatalf("Cloudstone Curio: one trigger, got %d", len(curio))
	}
	if curio[0].TargetsFrom != nil || curio[0].Targets != nil {
		t.Error("Cloudstone Curio: the bounce is a resolution-time choice, not a declared target clause")
	}
}

// TestStrionicResonatorCopiesATriggeredAbilityOnly — the clause over
// a real stack. A triggered ability of yours is offered; an activated
// one and an opponent's trigger are not, which is the whole
// difference between this card and Lithoform Engine.
func TestStrionicResonatorCopiesATriggeredAbilityOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	res := effectstest.PushCatalogPermanent(g, me.ID, "Strionic Resonator", "Artifact", effectstest.StrionicResonatorOracle, false)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Theirs", "Artifact", "", false)

	if err := g.AnnounceTrigger(me.ID, other, game.AbilityParams{Label: "a trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	if err := g.ActivateAbility(me.ID, other, game.AbilityParams{Label: "an activation"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	if err := g.AnnounceTrigger(opp.ID, theirs, game.AbilityParams{Label: "their trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	mine := effectstest.AcItemLabelled(g, "a trigger")
	activation := effectstest.AcItemLabelled(g, "an activation")
	notMine := effectstest.AcItemLabelled(g, "their trigger")

	offered := effectstest.AcLegalAbilityTargets(g, me.ID, effectstest.StrionicResonatorOracle, 0)
	if !offered[mine] {
		t.Error("a triggered ability you control must be offered")
	}
	if offered[activation] {
		t.Error("an activated ability is not a legal target for Strionic Resonator")
	}
	if offered[notMine] {
		t.Error("an opponent's trigger is not 'you control'")
	}

	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, res, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	// Pass priority only until the copy exists. Passing all the way
	// would resolve it too, and an item that has resolved is off the
	// stack — the assertion has to catch the copy while it is there.
	effectstest.AcPassUntilCopy(t, g)
	if n := effectstest.AcCopiesOnStack(g); n != 1 {
		t.Fatalf("the Resonator made %d copies, want 1", n)
	}
	cp := effectstest.AcCopyOnStack(g)
	if cp == nil || cp.Kind != game.StackItemTriggered {
		t.Fatalf("the copy of a triggered ability is a triggered ability: %+v", cp)
	}
	if cp.Controller != me.ID {
		t.Error("the copy is controlled by the player who created it (CR 707.10b)")
	}
}

// TestLithoformEngineCopiesBothKinds — the first ability's clause is
// the wider one: an activated ability of yours is a legal target too.
func TestLithoformEngineCopiesBothKinds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Lithoform Engine", "Legendary Artifact", effectstest.LithoformEngineOracle, false)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)

	if err := g.AnnounceTrigger(me.ID, other, game.AbilityParams{Label: "a trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	if err := g.ActivateAbility(me.ID, other, game.AbilityParams{Label: "an activation"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	offered := effectstest.AcLegalAbilityTargets(g, me.ID, effectstest.LithoformEngineOracle, 0)
	trig, act := effectstest.AcItemLabelled(g, "a trigger"), effectstest.AcItemLabelled(g, "an activation")
	if !offered[trig] || !offered[act] {
		t.Errorf("'activated or triggered' must offer both: trig=%v act=%v", offered[trig], offered[act])
	}
}

// TestRingsOfBrighthearthCopiesTheAbilityYouJustActivated — the
// untargeted member of the family, and the proof that the trigger
// data reaches the resolution: the Rings name the ability to copy
// with ctx.Trigger().Event.StackItemID and have no other handle on it.
func TestRingsOfBrighthearthCopiesTheAbilityYouJustActivated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Rings of Brighthearth", "Artifact", effectstest.RingsOfBrighthearthOracle, false)
	monolith := effectstest.PushCatalogPermanent(g, me.ID, "Basalt Monolith", "Artifact", effectstest.AcBasaltMonolithOracle, false)

	effectstest.FloatForTest(g, me, "CCCCC")
	if err := g.ActivateCatalogAbility(me.ID, monolith, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	// The activation's own runStateChecks drains the queue, so by the
	// time this reads it the Rings trigger is already on the stack.
	trigger := effectstest.AcItemOnStack(g, "Rings of Brighthearth")
	if trigger == nil {
		t.Fatal("activating an ability did not trigger the Rings")
	}
	if trigger.Trigger == nil || trigger.Trigger.Event.StackItemID == uuid.Nil {
		t.Fatalf("the Rings' trigger carries no activation to copy: %+v", trigger.Trigger)
	}

	effectstest.PassPriorityAroundTable(t, g)
	pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("the Rings did not offer the {2}")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if n := effectstest.AcCopiesOnStack(g); n != 1 {
		t.Fatalf("paying the {2} made %d copies, want 1", n)
	}
}

// TestRingsOfBrighthearthDoesNotTriggerOffItsOwnCopy — CR 707.10a. A
// copy is created, not activated, so the Rings see nothing and the
// loop terminates. This is the "an ability-activation watch must not
// fire for a copy" rule, tested through the one card in the catalog
// that watches activations.
func TestRingsOfBrighthearthDoesNotTriggerOffItsOwnCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Rings of Brighthearth", "Artifact", effectstest.RingsOfBrighthearthOracle, false)
	monolith := effectstest.PushCatalogPermanent(g, me.ID, "Basalt Monolith", "Artifact", effectstest.AcBasaltMonolithOracle, false)

	effectstest.FloatForTest(g, me, "CCCCC")
	if err := g.ActivateCatalogAbility(me.ID, monolith, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("the Rings did not offer the {2}")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	// One copy, and the table saw exactly ONE activation — the real
	// one. A second would mean the copy announced itself, and the
	// Rings would be copying their own copy forever.
	if n := effectstest.AcCopiesOnStack(g); n != 1 {
		t.Fatalf("%d copies on the stack, want 1", n)
	}
	if n := effectstest.AcCountEvents(g, game.EventActivateAbility); n != 1 {
		t.Fatalf("%d activations announced, want 1 — a copy is created, not activated (CR 707.10a)", n)
	}
	// One Rings announcement, too. EventTrigger is also the CR 732
	// loop breaker's breadcrumb for the activation itself, so count
	// only the ones carrying the Rings' own label.
	if n := effectstest.AcCountTriggerAnnouncements(g, "Rings of Brighthearth"); n != 1 {
		t.Fatalf("the Rings announced %d triggers for one activation, want 1", n)
	}
}

// TestRingsOfBrighthearthIgnoresAManaAbility — "if it isn't a mana
// ability", enforced by CR 605.3b rather than by a predicate: a mana
// ability announces a different event and never reaches the stack.
func TestRingsOfBrighthearthIgnoresAManaAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Rings of Brighthearth", "Artifact", effectstest.RingsOfBrighthearthOracle, false)
	monolith := effectstest.PushCatalogPermanent(g, me.ID, "Basalt Monolith", "Artifact", effectstest.AcBasaltMonolithOracle, false)

	if err := g.ActivateManaAbility(me.ID, monolith, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if n := len(g.PendingTriggers); n != 0 {
		t.Errorf("a mana ability triggered the Rings %d times (CR 605.3b)", n)
	}
}

// TestScrapTrawlerReadsTheDeadArtifactsManaValue — the trigger-data
// row's own card. The clause is cut to the artifact that died: a
// {3} leaves the {1} reachable and the {5} out.
func TestScrapTrawlerReadsTheDeadArtifactsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Scrap Trawler", "Artifact Creature — Construct", effectstest.ScrapTrawlerOracle, false)
	cheap := acGraveyardArtifact(g, me, "Cheap", "{1}")
	pricey := acGraveyardArtifact(g, me, "Pricey", "{5}")
	dying := acBattlefieldArtifact(g, me, "Dying", "{3}")

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(dying); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})

	choice := effectstest.LatestChoiceOfKind(g, game.PendingChoicePickTarget)
	if choice == nil {
		t.Fatal("Scrap Trawler queued no target pick")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range choice.PickTargetCards {
		offered[id] = true
	}
	if !offered[cheap] {
		t.Error("a cheaper artifact in your graveyard must be offered")
	}
	if offered[pricey] {
		t.Error("'lesser mana value' excluded nothing — the {5} was offered against a {3}")
	}
	if offered[dying] {
		t.Error("the artifact that died is not of LESSER mana value than itself")
	}
}

// TestCloudstoneCurioReadsTheEnteringPermanentsTypes — the other
// trigger-data card: "shares a permanent type with it", where "it" is
// the permanent that just entered. #1337: the bounce is a
// resolution-time choice (ChoosePermanents), not a target picked when
// the trigger goes on the stack, so the prompt only appears after the
// trigger resolves.
func TestCloudstoneCurioReadsTheEnteringPermanentsTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Cloudstone Curio", "Artifact", effectstest.CloudstoneCurioOracle, false)
	creature := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)

	entering := effectstest.PushCatalogPermanent(g, me.ID, "Elf", "Creature — Elf", "", false)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: entering, Actor: me.ID})
	})
	if effectstest.PendingOfKind(g, game.PendingChoicePickTarget) != nil {
		t.Fatal("the bounce asked for a target; the printed text names none")
	}
	effectstest.PassPriorityAroundTable(t, g)

	choice := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if choice == nil {
		t.Fatal("Cloudstone Curio queued no pick")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range choice.ChooseCards {
		offered[id] = true
	}
	if !offered[creature] {
		t.Error("another creature you control shares a permanent type with the entering creature")
	}
	if offered[land] {
		t.Error("a land shares no permanent type with a creature")
	}
	if offered[entering] {
		t.Error("'another' must exclude the permanent that entered")
	}
	if choice.ChooseMin != 0 {
		t.Error("the 'you may' is an up-to-one, so declining must be an answer")
	}

	handBefore := len(me.Hand.Cards)
	if err := g.ResolveOwnPermanents(choice.ID, me.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if _, stillOut := effectstest.AangCardOnBF(g, creature); stillOut {
		t.Error("the chosen creature is still on the battlefield")
	}
	if len(me.Hand.Cards) != handBefore+1 {
		t.Errorf("hand %d, want %d", len(me.Hand.Cards), handBefore+1)
	}
}

// TestCloudstoneCurioDeclineReturnsNothing — "you may" declined
// leaves the board untouched.
func TestCloudstoneCurioDeclineReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Cloudstone Curio", "Artifact", effectstest.CloudstoneCurioOracle, false)
	creature := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	entering := effectstest.PushCatalogPermanent(g, me.ID, "Elf", "Creature — Elf", "", false)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: entering, Actor: me.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)

	choice := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if choice == nil {
		t.Fatal("Cloudstone Curio queued no pick")
	}
	if err := g.ResolveOwnPermanents(choice.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveOwnPermanents(decline): %v", err)
	}
	if _, ok := effectstest.AangCardOnBF(g, creature); !ok {
		t.Error("declining the 'you may' still bounced a permanent")
	}
}

// acBattlefieldArtifact seats an artifact with a printed mana cost,
// so its mana value is a real number rather than zero.
func acBattlefieldArtifact(g *game.Game, p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Artifact", ManaCost: cost,
			Owner: p.ID, Controller: p.ID,
		})
	})
	return id
}

// acGraveyardArtifact puts an artifact CARD in a seat's graveyard.
func acGraveyardArtifact(g *game.Game, p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Artifact", ManaCost: cost,
			Owner: p.ID, Controller: p.ID,
		})
	})
	return id
}
