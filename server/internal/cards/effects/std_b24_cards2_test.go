package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b24_cards2_test.go — Standard batch 24, the second half: Norin, Trial
// of Agony, Waltz of Rage, Altanak, Anthropede, Balustrade Wurm and
// Cathartic Parting.

// b24Attack declares `attacker` at `defender` and walks to the declare-blockers
// step, ready for blocks.
func b24Attack(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
}

// --- Norin, Swift Survivalist ----------------------------------------------------

func TestStdB24NorinCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 3, 3)
	norin := b12Push(g, them.ID, "Norin, Swift Survivalist", b24NorinType, b24NorinOracle, 2, 1)
	b24Attack(t, g, attacker, them.ID)
	if err := g.DeclareBlocker(norin, attacker); err == nil {
		t.Error("Norin blocked")
	}
	if _, cantBlock := technodromeRestricted(t, g, norin); !cantBlock {
		t.Error("Norin does not carry the can't-block restriction")
	}
}

func TestStdB24NorinExilesTheBlockedCreatureAndYouMayPlayItThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Norin, Swift Survivalist", b24NorinType, b24NorinOracle, 2, 1)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	wall := pushVanillaCreature(g, them.ID, "Wall", 0, 9)
	b24Attack(t, g, bear, them.ID)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Fatal("the blocked creature was not exiled")
	}
	perm := exiledPermission(g, bear)
	if perm.Player != me.ID || !permissionLive(g, perm, me.ID) {
		t.Errorf("%+v: you may play the exiled card this turn", perm)
	}
	initSettle(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, wall); got != 0 {
		t.Errorf("the blocker took %d: the exiled creature still dealt combat damage", got)
	}
}

func TestStdB24NorinGrantsTheOwnersCardToItsControllerNotItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, them, third := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Norin, Swift Survivalist", b24NorinType, b24NorinOracle, 2, 1)
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: third.ID, Controller: me.ID,
	})
	wall := pushVanillaCreature(g, them.ID, "Wall", 4, 9)
	b24Attack(t, g, stolen, them.ID)
	if err := g.DeclareBlocker(wall, stolen); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(stolen) {
		t.Fatal("the creature was not exiled")
	}
	perm := exiledPermission(g, stolen)
	if perm.Player != me.ID {
		t.Errorf("the permission names %s, want you (Norin's controller), not the owner", perm.Player)
	}
}

func TestStdB24NorinDecliningLeavesTheCreatureInCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Norin, Swift Survivalist", b24NorinType, b24NorinOracle, 2, 1)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	wall := pushVanillaCreature(g, them.ID, "Wall", 0, 9)
	b24Attack(t, g, bear, them.ID)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	initSettle(t, g)
	if g.Exile.Contains(bear) || !g.Battlefield.Contains(bear) {
		t.Error("declining exiled the creature")
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, wall); got != 2 {
		t.Errorf("the wall took %d, want 2: combat damage was dealt", got)
	}
}

func TestStdB24NorinTriggersForAnyCreatureYouControlIncludingNorin(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	norin := b12Push(g, me.ID, "Norin, Swift Survivalist", b24NorinType, b24NorinOracle, 2, 1)
	wall := pushVanillaCreature(g, them.ID, "Wall", 0, 9)
	// An opponent's blocked creature is not yours: only creatures you control.
	b24Attack(t, g, norin, them.ID)
	if err := g.DeclareBlocker(wall, norin); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(norin) {
		t.Error("Norin, blocked, was not exiled: it is a creature you control")
	}
}

// --- Trial of Agony --------------------------------------------------------------

func b24TrialTargets(a, b uuid.UUID) []game.TargetRef {
	return []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0},
		{Kind: game.TargetCard, ID: b, Slot: 1},
	}
}

func TestStdB24TrialOfAgonyThatPlayerChoosesWhichTakesFiveAndTheOtherCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 9)
	b := b12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 9)
	castCatalogSpell(t, g, "Trial of Agony", "Sorcery", b24TrialOracle, b24TrialTargets(a, b))
	passPriorityAroundTable(t, g)
	if socB20ChooseCardsPrompt(g, me.ID) != nil {
		t.Fatal("the caster was asked to choose")
	}
	prompt := socB20ChooseCardsPrompt(g, opp.ID)
	if prompt == nil {
		t.Fatal("the opponent was not asked to choose one of the creatures")
	}
	if err := g.ResolveChooseCards(prompt.ID, opp.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b24Damage(t, g, b); got != 5 {
		t.Errorf("the chosen creature took %d, want 5", got)
	}
	if got := b24Damage(t, g, a); got != 0 {
		t.Errorf("the other creature took %d", got)
	}
	if _, cantBlock := technodromeRestricted(t, g, a); !cantBlock {
		t.Error("the other creature can still block")
	}
	if _, cantBlock := technodromeRestricted(t, g, b); cantBlock {
		t.Error("the chosen creature can't block")
	}
}

func TestStdB24TrialOfAgonyAnIllegalTargetLeavesTheOtherAsTheChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 9)
	b := b12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 9)
	castCatalogSpell(t, g, "Trial of Agony", "Sorcery", b24TrialOracle, b24TrialTargets(a, b))
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(a) })
	passPriorityAroundTable(t, g)
	if socB20ChooseCardsPrompt(g, opp.ID) != nil || socB20ChooseCardsPrompt(g, me.ID) != nil {
		t.Fatal("a choice was offered with one creature left")
	}
	if got := b24Damage(t, g, b); got != 5 {
		t.Errorf("the remaining creature took %d, want 5", got)
	}
}

func TestStdB24TrialOfAgonyNeedsTwoCreaturesOfTheSameOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	a := b12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 9)
	b := b12Creature(g, other.ID, "Wall B", "Creature — Wall", 0, 9)
	mine := b12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 9)
	for name, targets := range map[string][]game.TargetRef{
		"two different opponents":  b24TrialTargets(a, b),
		"your own creature":        b24TrialTargets(mine, a),
		"the same creature twice":  b24TrialTargets(a, a),
		"only one target":          b24TrialTargets(a, a)[:1],
		"a creature of each owner": b24TrialTargets(b, mine),
	} {
		id := putInHand(me, game.Card{Name: "Trial of Agony", TypeLine: "Sorcery", OracleID: b24TrialOracle})
		toMain(t, g)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
	}
}

// --- Waltz of Rage ---------------------------------------------------------------

func TestStdB24WaltzOfRageTargetDealsItsPowerToEachOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, me.ID, "Big", "Creature — Giant", 3, 9)
	mine := b12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 9)
	theirs := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	castCatalogSpell(t, g, "Waltz of Rage", "Sorcery", b24WaltzOracle, dskDGTarget(big))
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"your other creature": mine, "their creature": theirs} {
		if got := b24Damage(t, g, id); got != 3 {
			t.Errorf("%s took %d, want 3", name, got)
		}
	}
	if got := b24Damage(t, g, big); got != 0 {
		t.Errorf("the target took %d: it is not an \"other\" creature", got)
	}
}

func TestStdB24WaltzOfRageExilesTheTopCardForEachCreatureOfYoursThatDiesThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, me.ID, "Big", "Creature — Giant", 3, 9)
	small1 := b12Creature(g, me.ID, "Small One", "Creature — Elf", 1, 1)
	small2 := b12Creature(g, me.ID, "Small Two", "Creature — Elf", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	top := samiLibraryTopFirst(me,
		game.Card{Name: "Top One", TypeLine: "Sorcery"},
		game.Card{Name: "Top Two", TypeLine: "Sorcery"},
		game.Card{Name: "Third", TypeLine: "Sorcery"},
	)
	castCatalogSpell(t, g, "Waltz of Rage", "Sorcery", b24WaltzOracle, dskDGTarget(big))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{small1, small2, theirs} {
		if g.Battlefield.Contains(id) {
			t.Fatal("a 1/1 survived 3 damage")
		}
	}
	// Two of your creatures died: two cards exiled. The opponent's death is
	// not "a creature you control".
	if !g.Exile.Contains(top[0]) || !g.Exile.Contains(top[1]) {
		t.Fatal("the top two cards were not exiled for the two creatures that died")
	}
	if g.Exile.Contains(top[2]) {
		t.Error("a third card was exiled: the opponent's creature dying doesn't count")
	}
	for _, id := range top[:2] {
		perm := exiledPermission(g, id)
		if perm.Player != me.ID || !permissionLive(g, perm, me.ID) {
			t.Errorf("%+v: you may play the exiled card", perm)
		}
	}
	// Still this turn: another creature of yours dying exiles another card.
	late := b12Creature(g, me.ID, "Late", "Creature — Elf", 1, 1)
	rooncDestroy(t, g, late)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top[2]) {
		t.Error("a creature dying later this turn did not trigger the delayed ability")
	}
}

func TestStdB24WaltzOfRageEndsWithTheTurnAndFizzlesWithoutItsTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	big := b12Creature(g, me.ID, "Big", "Creature — Giant", 3, 9)
	other := b12Creature(g, me.ID, "Other", "Creature — Elf", 1, 1)
	top := samiLibraryTopFirst(me, game.Card{Name: "Top One", TypeLine: "Sorcery"}, game.Card{Name: "Top Two", TypeLine: "Sorcery"})
	castCatalogSpell(t, g, "Waltz of Rage", "Sorcery", b24WaltzOracle, dskDGTarget(big))
	rooncDestroy(t, g, big) // the target is gone: the spell does nothing
	// Not even a delayed ability: another creature dying exiles nothing.
	passPriorityAroundTable(t, g)
	if got := b24Damage(t, g, other); got != 0 {
		t.Errorf("a fizzled Waltz dealt %d damage", got)
	}
	rooncDestroy(t, g, other)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(top[0]) {
		t.Error("a fizzled Waltz still set up its delayed ability")
	}
}

func TestStdB24WaltzOfRageDelayedAbilityLastsOnlyUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	big := b12Creature(g, me.ID, "Big", "Creature — Giant", 3, 9)
	top := samiLibraryTopFirst(me, game.Card{Name: "Top One", TypeLine: "Sorcery"}, game.Card{Name: "Top Two", TypeLine: "Sorcery"})
	castCatalogSpell(t, g, "Waltz of Rage", "Sorcery", b24WaltzOracle, dskDGTarget(big))
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	for i := 0; i < 6 && g.Turn.Step != game.StepUpkeep; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	late := b12Creature(g, me.ID, "Late", "Creature — Elf", 1, 1)
	rooncDestroy(t, g, late)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(top[0]) {
		t.Error("the delayed ability outlived the turn")
	}
}

// --- Altanak, the Thrice-Called --------------------------------------------------

func TestStdB24AltanakDrawsWhenAnOpponentTargetsItBeforeTheSpellResolves(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	alt := b12Push(g, opp.ID, "Altanak, the Thrice-Called", b24AltanakType, b24AltanakOracle, 9, 9)
	if !hasKeywordOf(g, alt, "trample") {
		t.Error("Altanak does not have trample")
	}
	before := opp.Hand.Size()
	spell := castCatalogSpell(t, g, "Giant Growth", "Instant", b24GiantGrowthOracle, cardRefs(alt))
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != before+1 {
		t.Errorf("Altanak's controller has %d cards, want %d", got, before+1)
	}
	// The draw resolves before the spell that caused it (the ruling).
	draw, resolve := -1, -1
	for i, ev := range g.Events {
		if ev.Kind == game.EventDrawCard && ev.Actor == opp.ID && draw < 0 {
			draw = i
		}
		if ev.Kind == game.EventResolve && ev.CardID == spell {
			resolve = i
		}
	}
	if draw < 0 || resolve < 0 || draw > resolve {
		t.Errorf("draw event at %d, spell resolution at %d: the draw must come first", draw, resolve)
	}
}

func TestStdB24AltanakIgnoresItsOwnControllersSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	alt := b12Push(g, me.ID, "Altanak, the Thrice-Called", b24AltanakType, b24AltanakOracle, 9, 9)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Giant Growth", "Instant", b24GiantGrowthOracle, cardRefs(alt))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("targeting your own Altanak drew %d cards", got-before)
	}
}

func TestStdB24AltanakTriggersOnceForASpellThatTargetsItTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	alt := b12Push(g, me.ID, "Altanak, the Thrice-Called", b24AltanakType, b24AltanakOracle, 9, 9)
	item := uuid.New()
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, Actor: opp.ID, Source: uuid.New(),
				StackItemID: item, Target: alt, CardID: alt})
		}
	})
	if got := len(g.PendingTriggers); got != 1 {
		t.Errorf("%d triggers for one spell that targets Altanak twice, want 1", got)
	}
	// A different spell targeting it is a new trigger.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, Actor: opp.ID, Source: uuid.New(),
			StackItemID: uuid.New(), Target: alt, CardID: alt})
	})
	if got := len(g.PendingTriggers); got != 2 {
		t.Errorf("%d triggers after a second spell, want 2", got)
	}
}

func TestStdB24AltanakDiscardReturnsALandFromYourGraveyardTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	land := pushCatalogGraveyardCard(me, "Old Forest", "Basic Land — Forest", "", 0, 0)
	nonland := pushCatalogGraveyardCard(me, "Old Bear", "Creature — Bear", "", 2, 2)
	alt := pushCatalogHandCard(me, "Altanak, the Thrice-Called", b24AltanakType, b24AltanakOracle)
	if err := dskRT2Activate(g, me.ID, alt, 0, "{G}{C}", game.ActivateAbilityParams{Targets: cardRefs(nonland)}); err == nil {
		t.Fatal("a nonland card was accepted as the target")
	}
	me.ManaPool = nil
	if err := dskRT2Activate(g, me.ID, alt, 0, "{G}{C}", game.ActivateAbilityParams{Targets: cardRefs(land)}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if me.Hand.Contains(alt) || !me.Graveyard.Contains(alt) {
		t.Fatal("Altanak was not discarded as the cost")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land did not return")
	}
	if c, _ := battlefieldCard(g, land); !c.Tapped {
		t.Error("the land returned untapped")
	}
}

// --- Anthropede ------------------------------------------------------------------

func b24OpenPayPrompt(g *game.Game, p *game.Player) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == p.ID {
			return c
		}
	}
	return nil
}

func b24EnterAnthropede(t *testing.T, g *game.Game) {
	t.Helper()
	castAndResolveCreature(t, g, "Anthropede", b24AnthropedeType, b24AnthropedeOrcl)
	passPriorityAroundTable(t, g)
}

func TestStdB24AnthropedeDiscardACardToDestroyARoom(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	room := b12Permanent(g, opp.ID, "Test Room", "Enchantment — Room")
	other := b12Permanent(g, opp.ID, "Test Aura", "Enchantment")
	hand := b24HandCards(me, "Pitch Me")
	b24EnterAnthropede(t, g)
	prompt := b24OpenPayPrompt(g, me)
	if prompt == nil {
		t.Fatal("no discard-or-pay prompt")
	}
	option := 0
	if err := g.ResolvePayUnlessWithOption(prompt.ID, me.ID, true, &option, []uuid.UUID{hand[0]}); err != nil {
		t.Fatalf("pay by discarding: %v", err)
	}
	if me.Hand.Contains(hand[0]) {
		t.Error("the discarded card is still in hand")
	}
	answerPickTargetIfOpen(t, g, room)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(room) {
		t.Error("the Room was not destroyed")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a non-Room enchantment was destroyed")
	}
}

func TestStdB24AnthropedePayTwoToDestroyARoom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	room := b12Permanent(g, me.ID, "My Room", "Enchantment — Room")
	b24HandCards(me, "Keep Me")
	toMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	b24EnterAnthropede(t, g)
	prompt := b24OpenPayPrompt(g, me)
	if prompt == nil {
		t.Fatal("no discard-or-pay prompt")
	}
	option := 1
	if err := g.ResolvePayUnlessWithOption(prompt.ID, me.ID, true, &option, nil); err != nil {
		t.Fatalf("pay {2}: %v", err)
	}
	answerPickTargetIfOpen(t, g, room)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(room) {
		t.Error("the Room (your own is a legal target) was not destroyed")
	}
	if got := poolSize(g, me.ID); got != 0 {
		t.Errorf("{2} was not spent: %d mana left in the pool", got)
	}
}

func TestStdB24AnthropedeDecliningDestroysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	room := b12Permanent(g, opp.ID, "Test Room", "Enchantment — Room")
	hand := b24HandCards(me, "Keep Me")
	b24EnterAnthropede(t, g)
	prompt := b24OpenPayPrompt(g, me)
	if prompt == nil {
		t.Fatal("no discard-or-pay prompt")
	}
	if err := g.ResolvePayUnless(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(room) || !me.Hand.Contains(hand[0]) {
		t.Error("declining destroyed the Room or discarded a card")
	}
}

func TestStdB24AnthropedeAsksNobodyWhoCantPayEither(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Permanent(g, opp.ID, "Test Room", "Enchantment — Room")
	me.Hand.Cards = nil
	b24EnterAnthropede(t, g)
	if b24OpenPayPrompt(g, me) != nil {
		t.Error("asked to pay with an empty hand and no mana")
	}
}

// --- Balustrade Wurm -------------------------------------------------------------

func TestStdB24BalustradeWurmCantBeCounteredAndHasTrampleAndHaste(t *testing.T) {
	spec, _ := Lookup(b24WurmOracle)
	if !spec.CantBeCountered {
		t.Error("Balustrade Wurm can be countered")
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	wurm := b12Push(g, me.ID, "Balustrade Wurm", b24WurmType, b24WurmOracle, 5, 5)
	assertKeywords(t, g, wurm, "trample", "haste")
}

func TestStdB24BalustradeWurmReturnsOnlyWithDeliriumAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	wurm := pushCatalogGraveyardCard(me, "Balustrade Wurm", b24WurmType, b24WurmOracle, 5, 5)
	// The Wurm is the creature card; two more types is not enough.
	gyCard(me, "Old Bolt", "Instant", "{R}", "")
	gyCard(me, "Old Ritual", "Sorcery", "{B}", "")
	if err := dskRT2Activate(g, me.ID, wurm, 0, "{G}{G}{C}{C}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with three card types in the graveyard")
	}
	me.ManaPool = nil
	gyCard(me, "Old Forest", "Basic Land — Forest", "", "")
	if err := dskRT2Activate(g, me.ID, wurm, 0, "{G}{G}{C}{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with delirium: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Fatal("the Wurm did not return")
	}
	if got := countersOn(g, wurm, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
	rooncDestroy(t, g, wurm)
	if me.Graveyard.Contains(wurm) || !g.Exile.Contains(wurm) {
		t.Error("the Wurm with a finality counter went to the graveyard instead of exile")
	}
}

func TestStdB24BalustradeWurmCantReturnOutsideYourMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	wurm := pushCatalogGraveyardCard(me, "Balustrade Wurm", b24WurmType, b24WurmOracle, 5, 5)
	b24SeedTypes(me, 4)
	advanceTo(t, g, game.StepBeginCombat)
	if err := dskRT2Activate(g, me.ID, wurm, 0, "{G}{G}{C}{C}", game.ActivateAbilityParams{}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Cathartic Parting -----------------------------------------------------------

func b24PartingTargets(permanent uuid.UUID, graveyard ...uuid.UUID) []game.TargetRef {
	out := []game.TargetRef{{Kind: game.TargetCard, ID: permanent, Slot: 0}}
	for _, id := range graveyard {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id, Slot: 1})
	}
	return out
}

func b24LibraryContains(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Library.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestStdB24CatharticPartingShufflesTheirPermanentAndYourGraveyardCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	gy1 := gyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	gy2 := gyCard(me, "Old Bolt", "Instant", "{R}", "")
	castCatalogSpell(t, g, "Cathartic Parting", "Sorcery", b24ParTingOracle, b24PartingTargets(relic, gy1, gy2))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || !b24LibraryContains(opp, relic) {
		t.Fatal("the artifact was not shuffled into its owner's library")
	}
	answerConfirmFor(t, g, me, true)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(gy1) || me.Graveyard.Contains(gy2) {
		t.Error("a targeted graveyard card stayed in the graveyard")
	}
	if !b24LibraryContains(me, gy1) || !b24LibraryContains(me, gy2) {
		t.Error("the targeted graveyard cards are not in your library")
	}
}

func TestStdB24CatharticPartingTheOwnerShufflesAStolenPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp, third := g.Seats[1], g.Seats[2]
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Relic", TypeLine: "Artifact",
		Owner: third.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Cathartic Parting", "Sorcery", b24ParTingOracle, b24PartingTargets(stolen))
	passPriorityAroundTable(t, g)
	if !b24LibraryContains(third, stolen) || b24LibraryContains(opp, stolen) {
		t.Error("the permanent did not go to its OWNER's library")
	}
}

func TestStdB24CatharticPartingTheGraveyardHalfIsOptional(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := b12Permanent(g, opp.ID, "Relic", "Enchantment")
	gy := gyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	castCatalogSpell(t, g, "Cathartic Parting", "Sorcery", b24ParTingOracle, b24PartingTargets(relic, gy))
	passPriorityAroundTable(t, g)
	answerConfirmFor(t, g, me, false)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(gy) || b24LibraryContains(me, gy) {
		t.Error("declining moved the graveyard card")
	}
	if b24LibraryContains(opp, relic) == false {
		t.Error("the first half is not optional")
	}
}

func TestStdB24CatharticPartingMayShuffleEvenWithNoGraveyardTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	castCatalogSpell(t, g, "Cathartic Parting", "Sorcery", b24ParTingOracle, b24PartingTargets(relic))
	passPriorityAroundTable(t, g)
	answerConfirmFor(t, g, me, true)
	passPriorityAroundTable(t, g)
	if g.Stack.Size() != 0 {
		t.Error("the spell did not finish")
	}
}

func TestStdB24CatharticPartingTargetLimits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	mine := b12Permanent(g, me.ID, "My Relic", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	var gy []uuid.UUID
	for i := 0; i < 5; i++ {
		gy = append(gy, gyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", ""))
	}
	for name, targets := range map[string][]game.TargetRef{
		"five graveyard cards":      b24PartingTargets(relic, gy...),
		"your own permanent":        b24PartingTargets(mine),
		"an opponent's nonartifact": b24PartingTargets(land),
		"no permanent at all":       b24PartingTargets(relic)[:0],
		"an opponent's graveyard":   {{Kind: game.TargetCard, ID: relic, Slot: 0}, {Kind: game.TargetCard, ID: gyCard(opp, "Their Bear", "Creature — Bear", "", ""), Slot: 1}},
	} {
		id := putInHand(me, game.Card{Name: "Cathartic Parting", TypeLine: "Sorcery", OracleID: b24ParTingOracle})
		toMain(t, g)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
	}
	// Four graveyard cards is the most.
	id := putInHand(me, game.Card{Name: "Cathartic Parting", TypeLine: "Sorcery", OracleID: b24ParTingOracle})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: b24PartingTargets(relic, gy[:4]...)}); err != nil {
		t.Errorf("four graveyard cards: %v", err)
	}
}
