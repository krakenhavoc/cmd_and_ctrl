package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b705_cards_test.go — S82 OLD batch 705 (#5621, Urza's Destiny):
// card draw, evasion while enchanted, the Scents. Each test plays the real
// catalog entry.

const (
	oldB705Fatigue    = "0a88dcb6-a391-408f-8bfc-7b5b2cc34267"
	oldB705Rayne      = "6933a7bf-10ed-46c8-a49c-ce200bce3eea"
	oldB705Glass      = "ca16b32c-3f1e-4f8d-821a-064f922c45d8"
	oldB705Rector     = "e3c85068-b4b6-40b9-a16c-5c3b2d059ec4"
	oldB705Horror     = "4344abeb-47ab-4638-b50f-dd502f14c1af"
	oldB705Beebles    = "75903c3b-2dbe-4662-b6a8-39599e49adde"
	oldB705Elite      = "c8a1b3dd-bc36-4d97-8f93-1151446f04e2"
	oldB705Osprey     = "d5482ccf-7c3f-45b8-aaa8-8f6686fe584d"
	oldB705Donate     = "049d1fb2-cce7-40e7-a633-e84cf9782d89"
	oldB705Nightshade = "2e88f39d-6177-42d9-b756-144219dc0961"
	oldB705Ivy        = "b6db81e5-b7ca-45dc-a72d-1ff12935a3e4"
)

func TestOldB705EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB705Fatigue: "Fatigue", oldB705Rayne: "Rayne, Academy Chancellor", oldB705Glass: "Scrying Glass",
		oldB705Rector: "Academy Rector", oldB705Horror: "Skittering Horror", oldB705Beebles: "Bubbling Beebles",
		oldB705Elite: "Metathran Elite", oldB705Osprey: "Fledgling Osprey", oldB705Donate: "Donate",
		oldB705Nightshade: "Scent of Nightshade", oldB705Ivy: "Scent of Ivy",
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

func oldB705Library(p *game.Player, n int) {
	p.Library.Cards = nil
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: p.ID, Controller: p.ID})
	}
}

func oldB705Hand(g *game.Game, p *game.Player, cards ...game.Card) []uuid.UUID {
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		p.Hand.Cards = nil
		for _, c := range cards {
			c.InstanceID, c.Owner, c.Controller = uuid.New(), p.ID, p.ID
			p.Hand.PushTop(c)
			ids = append(ids, c.InstanceID)
		}
	})
	return ids
}

// --- Fatigue -----------------------------------------------------------------

func TestOldB705FatigueSkipsTheTargetsNextDrawStepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oppSeat := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	oldB705Library(opp, 12)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UUU")
	effectstest.CastFromHandForTest(t, g, me, "Fatigue", "Sorcery", "{1}{U}", oldB705Fatigue,
		game.CastSpellParams{Strict: true, Targets: effectstest.P8Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, oppSeat)
	before := opp.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if opp.Hand.Size() != before {
		t.Errorf("the target drew on their next draw step: hand %d, was %d", opp.Hand.Size(), before)
	}
	// The skip is spent: their following draw step draws.
	effectstest.AdvanceToUpkeepOfSeat(t, g, oppSeat)
	before = opp.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if opp.Hand.Size() != before+1 {
		t.Errorf("the second draw step after Fatigue drew nothing: hand %d, want %d", opp.Hand.Size(), before+1)
	}
}

// --- Skittering Horror -------------------------------------------------------

func TestOldB705SkitteringHorrorIsSacrificedWhenYouCastACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	horror := effectstest.PushCatalogPermanent(g, me.ID, "Skittering Horror", "Creature — Phyrexian Horror", oldB705Horror, false)
	effectstest.CastWithCost(t, g, "Plain Sorcery", "Sorcery", "{2}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(horror) {
		t.Fatal("the Horror was sacrificed for a noncreature spell")
	}
	effectstest.CastWithCost(t, g, "Plain Bear", "Creature — Bear", "{2}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(horror) {
		t.Error("the Horror survived its controller casting a creature spell")
	}
}

func TestOldB705SkitteringHorrorIgnoresAnOpponentsCreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	horror := effectstest.PushCatalogPermanent(g, me.ID, "Skittering Horror", "Creature — Phyrexian Horror", oldB705Horror, false)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventCast, Actor: opp.ID, CardID: uuid.New()})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(horror) {
		t.Error("the Horror was sacrificed for an opponent's spell")
	}
}

// --- Bubbling Beebles --------------------------------------------------------

func oldB705Attack(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
}

func TestOldB705BubblingBeeblesIsUnblockableWhileTheDefenderControlsAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	beebles := effectstest.PushCatalogPermanent(g, me.ID, "Bubbling Beebles", "Creature — Beeble", oldB705Beebles, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PushCatalogPermanent(g, opp.ID, "Pacifism-ish", "Enchantment", "", false)
	oldB705Attack(t, g, beebles, opp.ID)
	effectstest.BrRefusal(t, g.DeclareBlocker(bear, beebles), game.BlockReasonCantBeBlocked)
}

func TestOldB705BubblingBeeblesCanBeBlockedWithoutAnEnchantmentOnTheDefendersSide(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	beebles := effectstest.PushCatalogPermanent(g, me.ID, "Bubbling Beebles", "Creature — Beeble", oldB705Beebles, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	// The attacker's own enchantment is not the defender's.
	effectstest.PushCatalogPermanent(g, me.ID, "My Enchantment", "Enchantment", "", false)
	oldB705Attack(t, g, beebles, opp.ID)
	if err := g.DeclareBlocker(bear, beebles); err != nil {
		t.Errorf("the Beebles could not be blocked: %v", err)
	}
}

// --- Metathran Elite and Fledgling Osprey ------------------------------------

func TestOldB705MetathranEliteIsUnblockableOnlyWhileEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elite := effectstest.PushCatalogPermanent(g, me.ID, "Metathran Elite", "Creature — Metathran Soldier", oldB705Elite, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "Aura", "", elite)
	oldB705Attack(t, g, elite, opp.ID)
	effectstest.BrRefusal(t, g.DeclareBlocker(bear, elite), game.BlockReasonCantBeBlocked)
}

func TestOldB705MetathranEliteCanBeBlockedWhileUnenchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elite := effectstest.PushCatalogPermanent(g, me.ID, "Metathran Elite", "Creature — Metathran Soldier", oldB705Elite, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	oldB705Attack(t, g, elite, opp.ID)
	if err := g.DeclareBlocker(bear, elite); err != nil {
		t.Errorf("an unenchanted Elite could not be blocked: %v", err)
	}
}

func TestOldB705FledglingOspreyHasFlyingOnlyWhileEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	osprey := effectstest.PushCatalogPermanent(g, me.ID, "Fledgling Osprey", "Creature — Bird", oldB705Osprey, false)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Other Bird", "Creature — Bird", "", false)
	if effectstest.EffectiveAbilitiesContain(t, g, osprey, "flying") {
		t.Fatal("an unenchanted Osprey has flying")
	}
	aura := effectstest.Enchant(t, g, "Aura", "", osprey)
	if !effectstest.EffectiveAbilitiesContain(t, g, osprey, "flying") {
		t.Error("an enchanted Osprey lacks flying")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, other, "flying") {
		t.Error("the Osprey's flying leaked to another creature")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	if effectstest.EffectiveAbilitiesContain(t, g, osprey, "flying") {
		t.Error("the Osprey kept flying after its Aura left")
	}
}

// --- Donate ------------------------------------------------------------------

func TestOldB705DonateGivesAPermanentYouControlAway(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UUUUUU")
	if _, err := effectstest.CastFromHandErr(g, me, "Donate", "Sorcery", "{2}{U}", oldB705Donate,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Error("Donate targeted a permanent its caster does not control")
	}
	effectstest.CastFromHandForTest(t, g, me, "Donate", "Sorcery", "{2}{U}", oldB705Donate,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardByID(g, bear)
	if c == nil || c.Controller != opp.ID {
		t.Fatalf("the Bear's controller is %v, want the target player", c)
	}
	if c.Owner != me.ID {
		t.Error("ownership changed: Donate gives control only")
	}
	for i := 0; i < 12; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if c := effectstest.FindBattlefieldCardByID(g, bear); c == nil || c.Controller != opp.ID {
		t.Error("the Bear came back: Donate has no duration")
	}
}

// --- Academy Rector ----------------------------------------------------------

func TestOldB705AcademyRectorExilesItselfToFetchAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rector := effectstest.PushCatalogPermanent(g, me.ID, "Academy Rector", "Creature — Human Cleric", oldB705Rector, false)
	oldB705Library(me, 4)
	wanted := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: wanted, Name: "Some Enchantment", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	})
	effectstest.P1Destroy(t, g, rector)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, wanted)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wanted) {
		t.Error("the fetched enchantment did not enter")
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == rector {
			t.Error("the Rector is still in the graveyard: it was not exiled")
		}
	}
}

func TestOldB705AcademyRectorDeclinedDoesNotSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rector := effectstest.PushCatalogPermanent(g, me.ID, "Academy Rector", "Creature — Human Cleric", oldB705Rector, false)
	oldB705Library(me, 3)
	effectstest.P1Destroy(t, g, rector)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range me.Graveyard.Cards {
		found = found || c.InstanceID == rector
	}
	if !found {
		t.Error("declining still exiled the Rector")
	}
}

// --- Scent of Nightshade / Scent of Ivy --------------------------------------

func TestOldB705ScentOfIvyPumpsByTheGreenCardsRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ids := oldB705Hand(g, me,
		game.Card{Name: "Green A", TypeLine: "Creature — Elf", Colors: []string{"G"}},
		game.Card{Name: "Green B", TypeLine: "Sorcery", Colors: []string{"G"}},
		game.Card{Name: "Blue C", TypeLine: "Sorcery", Colors: []string{"U"}})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "G")
	effectstest.CastFromHandForTest(t, g, me, "Scent of Ivy", "Instant", "{G}", oldB705Ivy,
		game.CastSpellParams{Strict: true, Targets: effectstest.Pr7bTargets(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsPrompt(t, g)
	if len(prompt.ChooseCards) != 2 {
		t.Fatalf("offered %d cards to reveal, want the 2 green ones", len(prompt.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("Bear is %d/%d, want 4/4 after revealing two green cards", p, tough)
	}
	revealed := false
	for _, ev := range g.Events {
		revealed = revealed || ev.Kind == game.EventRevealCards
	}
	if !revealed {
		t.Error("the cards were not revealed")
	}
}

func TestOldB705ScentOfIvyMayRevealNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB705Hand(g, me, game.Card{Name: "Green A", TypeLine: "Creature — Elf", Colors: []string{"G"}})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "G")
	effectstest.CastFromHandForTest(t, g, me, "Scent of Ivy", "Instant", "{G}", oldB705Ivy,
		game.CastSpellParams{Strict: true, Targets: effectstest.Pr7bTargets(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("Bear is %d/%d, want 2/2 after revealing nothing", p, tough)
	}
}

func TestOldB705ScentOfNightshadeShrinksByTheBlackCardsRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	ids := oldB705Hand(g, me,
		game.Card{Name: "Black A", TypeLine: "Creature — Zombie", Colors: []string{"B"}},
		game.Card{Name: "Black B", TypeLine: "Sorcery", Colors: []string{"B"}},
		game.Card{Name: "Red C", TypeLine: "Sorcery", Colors: []string{"R"}})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "BB")
	effectstest.CastFromHandForTest(t, g, me, "Scent of Nightshade", "Instant", "{1}{B}", oldB705Nightshade,
		game.CastSpellParams{Strict: true, Targets: effectstest.Pr7bTargets(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("Victim is %d/%d, want 1/1 after revealing two black cards", p, tough)
	}
}

// --- Rayne, Academy Chancellor -----------------------------------------------

func oldB705Targeted(g *game.Game, actor, target, stackItem uuid.UUID, card bool) {
	ev := game.Event{Kind: game.EventBecomesTarget, Actor: actor, Target: target, StackItemID: stackItem, Source: uuid.New()}
	if card {
		ev.CardID = target
	}
	g.WithWriteLock(func() { g.EmitEvent(ev) })
}

func TestOldB705RayneDrawsWhenAnOpponentTargetsYouOrYourPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Rayne, Academy Chancellor", "Legendary Creature — Human Wizard", oldB705Rayne, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oldB705Library(me, 8)

	// A permanent they control: no trigger.
	hand := me.Hand.Size()
	oldB705Targeted(g, opp.ID, theirs, uuid.New(), true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) != nil || me.Hand.Size() != hand {
		t.Fatal("Rayne triggered for an opponent's own permanent")
	}
	// My own spell targeting my permanent: no trigger.
	oldB705Targeted(g, me.ID, bear, uuid.New(), true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) != nil || me.Hand.Size() != hand {
		t.Fatal("Rayne triggered for my own spell")
	}
	// An opponent targets my creature: may draw a card.
	oldB705Targeted(g, opp.ID, bear, uuid.New(), true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand = %d, want %d after an opponent targeted my creature", me.Hand.Size(), hand+1)
	}
	// An opponent targets me: may draw, and declining draws nothing.
	oldB705Targeted(g, opp.ID, me.ID, uuid.New(), false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("declining still drew: hand = %d", me.Hand.Size())
	}
}

func TestOldB705RayneTriggersOncePerObjectPerSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Rayne, Academy Chancellor", "Legendary Creature — Human Wizard", oldB705Rayne, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bear2 := effectstest.B12Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	oldB705Library(me, 8)
	spell := uuid.New()
	oldB705Targeted(g, opp.ID, bear, spell, true)
	oldB705Targeted(g, opp.ID, bear, spell, true)
	oldB705Targeted(g, opp.ID, bear2, spell, true)
	effectstest.PassPriorityAroundTable(t, g)
	asked := 0
	for i := 0; i < 4 && effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) != nil; i++ {
		effectstest.AnswerMayChoice(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
		asked++
	}
	if asked != 2 {
		t.Errorf("Rayne asked %d times, want 2 (one per object targeted)", asked)
	}
}

func TestOldB705RayneOffersAnAdditionalDrawWhileEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rayne := effectstest.PushCatalogPermanent(g, me.ID, "Rayne, Academy Chancellor", "Legendary Creature — Human Wizard", oldB705Rayne, false)
	effectstest.Enchant(t, g, "Aura", "", rayne)
	oldB705Library(me, 8)
	hand := me.Hand.Size()
	oldB705Targeted(g, opp.ID, rayne, uuid.New(), true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand = %d, want %d: one draw and the additional one", me.Hand.Size(), hand+2)
	}
}

// --- Scrying Glass -----------------------------------------------------------

func oldB705GlassTable(t *testing.T, theirHand ...game.Card) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB705Library(me, 6)
	glass := effectstest.PushCatalogPermanent(g, me.ID, "Scrying Glass", "Artifact", oldB705Glass, false)
	oldB705Hand(g, opp, theirHand...)
	effectstest.FloatForTest(g, me, "CCC")
	return g, me, opp, glass
}

func oldB705Glassed(t *testing.T, g *game.Game, me, opp *game.Player, glass uuid.UUID, number int, color string) {
	t.Helper()
	effectstest.Pr7Activate(t, g, me.ID, glass, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.P8Player(opp.ID)})
	effectstest.AnswerNumber(t, g, me.ID, number)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, color)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB705ScryingGlassDrawsOnAnExactMatch(t *testing.T) {
	g, me, opp, glass := oldB705GlassTable(t,
		game.Card{Name: "Red A", TypeLine: "Sorcery", Colors: []string{"R"}},
		game.Card{Name: "Red B", TypeLine: "Sorcery", Colors: []string{"R"}},
		game.Card{Name: "Blue C", TypeLine: "Sorcery", Colors: []string{"U"}})
	hand := me.Hand.Size()
	oldB705Glassed(t, g, me, opp, glass, 2, "R")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d after naming exactly two red cards", me.Hand.Size(), hand+1)
	}
	if c, _ := effectstest.BattlefieldCard(g, glass); !c.Tapped {
		t.Error("Scrying Glass was not tapped for the cost")
	}
}

func TestOldB705ScryingGlassDrawsNothingOffByOne(t *testing.T) {
	g, me, opp, glass := oldB705GlassTable(t,
		game.Card{Name: "Red A", TypeLine: "Sorcery", Colors: []string{"R"}},
		game.Card{Name: "Red B", TypeLine: "Sorcery", Colors: []string{"R"}})
	hand := me.Hand.Size()
	oldB705Glassed(t, g, me, opp, glass, 1, "R")
	if me.Hand.Size() != hand {
		t.Errorf("hand = %d, want %d: two red cards is not exactly one", me.Hand.Size(), hand)
	}
}

func TestOldB705ScryingGlassCountsAMulticolouredCardForEachColour(t *testing.T) {
	g, me, opp, glass := oldB705GlassTable(t,
		game.Card{Name: "Gold", TypeLine: "Creature — Elf", Colors: []string{"R", "G"}},
		game.Card{Name: "Colorless", TypeLine: "Artifact"})
	hand := me.Hand.Size()
	oldB705Glassed(t, g, me, opp, glass, 1, "G")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d: a red-green card is one green card", me.Hand.Size(), hand+1)
	}
}

func TestOldB705ScryingGlassRefusesYourselfAndNeedsAnOpponent(t *testing.T) {
	g, me, _, glass := oldB705GlassTable(t)
	if err := g.ActivateCatalogAbility(me.ID, glass, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.P8Player(me.ID)}); err == nil {
		t.Error("Scrying Glass targeted its own controller")
	}
}
