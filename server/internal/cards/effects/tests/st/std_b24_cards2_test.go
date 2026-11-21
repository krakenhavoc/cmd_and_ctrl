package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b24_cards2_test.go — Standard batch 24, the second half: Norin, Trial
// of Agony, Waltz of Rage, Altanak, Anthropede, Balustrade Wurm and
// Cathartic Parting.

// b24Attack declares `attacker` at `defender` and walks to the declare-blockers
// step, ready for blocks.
func b24Attack(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
}

// --- Norin, Swift Survivalist ----------------------------------------------------

func TestStdB24NorinCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 3, 3)
	norin := effectstest.B12Push(g, them.ID, "Norin, Swift Survivalist", b24NorinType, b24NorinOracle, 2, 1)
	b24Attack(t, g, attacker, them.ID)
	if err := g.DeclareBlocker(norin, attacker); err == nil {
		t.Error("Norin blocked")
	}
	if _, cantBlock := effectstest.TechnodromeRestricted(t, g, norin); !cantBlock {
		t.Error("Norin does not carry the can't-block restriction")
	}
}

func TestStdB24NorinExilesTheBlockedCreatureAndYouMayPlayItThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Norin, Swift Survivalist", b24NorinType, b24NorinOracle, 2, 1)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	wall := effectstest.PushVanillaCreature(g, them.ID, "Wall", 0, 9)
	b24Attack(t, g, bear, them.ID)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Fatal("the blocked creature was not exiled")
	}
	perm := effectstest.ExiledPermission(g, bear)
	if perm.Player != me.ID || !effectstest.PermissionLive(g, perm, me.ID) {
		t.Errorf("%+v: you may play the exiled card this turn", perm)
	}
	effectstest.InitSettle(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, wall); got != 0 {
		t.Errorf("the blocker took %d: the exiled creature still dealt combat damage", got)
	}
}

func TestStdB24NorinGrantsTheOwnersCardToItsControllerNotItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Norin, Swift Survivalist", b24NorinType, b24NorinOracle, 2, 1)
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: third.ID, Controller: me.ID,
	})
	wall := effectstest.PushVanillaCreature(g, them.ID, "Wall", 4, 9)
	b24Attack(t, g, stolen, them.ID)
	if err := g.DeclareBlocker(wall, stolen); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(stolen) {
		t.Fatal("the creature was not exiled")
	}
	perm := effectstest.ExiledPermission(g, stolen)
	if perm.Player != me.ID {
		t.Errorf("the permission names %s, want you (Norin's controller), not the owner", perm.Player)
	}
}

func TestStdB24NorinDecliningLeavesTheCreatureInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Norin, Swift Survivalist", b24NorinType, b24NorinOracle, 2, 1)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	wall := effectstest.PushVanillaCreature(g, them.ID, "Wall", 0, 9)
	b24Attack(t, g, bear, them.ID)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.InitSettle(t, g)
	if g.Exile.Contains(bear) || !g.Battlefield.Contains(bear) {
		t.Error("declining exiled the creature")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, wall); got != 2 {
		t.Errorf("the wall took %d, want 2: combat damage was dealt", got)
	}
}

func TestStdB24NorinTriggersForAnyCreatureYouControlIncludingNorin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	norin := effectstest.B12Push(g, me.ID, "Norin, Swift Survivalist", b24NorinType, b24NorinOracle, 2, 1)
	wall := effectstest.PushVanillaCreature(g, them.ID, "Wall", 0, 9)
	// An opponent's blocked creature is not yours: only creatures you control.
	b24Attack(t, g, norin, them.ID)
	if err := g.DeclareBlocker(wall, norin); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 9)
	b := effectstest.B12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 9)
	effectstest.CastCatalogSpell(t, g, "Trial of Agony", "Sorcery", b24TrialOracle, b24TrialTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SocB20ChooseCardsPrompt(g, me.ID) != nil {
		t.Fatal("the caster was asked to choose")
	}
	prompt := effectstest.SocB20ChooseCardsPrompt(g, opp.ID)
	if prompt == nil {
		t.Fatal("the opponent was not asked to choose one of the creatures")
	}
	if err := g.ResolveChooseCards(prompt.ID, opp.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B24Damage(t, g, b); got != 5 {
		t.Errorf("the chosen creature took %d, want 5", got)
	}
	if got := effectstest.B24Damage(t, g, a); got != 0 {
		t.Errorf("the other creature took %d", got)
	}
	if _, cantBlock := effectstest.TechnodromeRestricted(t, g, a); !cantBlock {
		t.Error("the other creature can still block")
	}
	if _, cantBlock := effectstest.TechnodromeRestricted(t, g, b); cantBlock {
		t.Error("the chosen creature can't block")
	}
}

func TestStdB24TrialOfAgonyAnIllegalTargetLeavesTheOtherAsTheChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 9)
	b := effectstest.B12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 9)
	effectstest.CastCatalogSpell(t, g, "Trial of Agony", "Sorcery", b24TrialOracle, b24TrialTargets(a, b))
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(a) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SocB20ChooseCardsPrompt(g, opp.ID) != nil || effectstest.SocB20ChooseCardsPrompt(g, me.ID) != nil {
		t.Fatal("a choice was offered with one creature left")
	}
	if got := effectstest.B24Damage(t, g, b); got != 5 {
		t.Errorf("the remaining creature took %d, want 5", got)
	}
}

func TestStdB24TrialOfAgonyNeedsTwoCreaturesOfTheSameOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	a := effectstest.B12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 9)
	b := effectstest.B12Creature(g, other.ID, "Wall B", "Creature — Wall", 0, 9)
	mine := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 9)
	for name, targets := range map[string][]game.TargetRef{
		"two different opponents":  b24TrialTargets(a, b),
		"your own creature":        b24TrialTargets(mine, a),
		"the same creature twice":  b24TrialTargets(a, a),
		"only one target":          b24TrialTargets(a, a)[:1],
		"a creature of each owner": b24TrialTargets(b, mine),
	} {
		id := effectstest.PutInHand(me, game.Card{Name: "Trial of Agony", TypeLine: "Sorcery", OracleID: b24TrialOracle})
		effectstest.ToMain(t, g)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
	}
}

// --- Waltz of Rage ---------------------------------------------------------------

func TestStdB24WaltzOfRageTargetDealsItsPowerToEachOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 3, 9)
	mine := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 9)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	effectstest.CastCatalogSpell(t, g, "Waltz of Rage", "Sorcery", b24WaltzOracle, effectstest.DskDGTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"your other creature": mine, "their creature": theirs} {
		if got := effectstest.B24Damage(t, g, id); got != 3 {
			t.Errorf("%s took %d, want 3", name, got)
		}
	}
	if got := effectstest.B24Damage(t, g, big); got != 0 {
		t.Errorf("the target took %d: it is not an \"other\" creature", got)
	}
}

func TestStdB24WaltzOfRageExilesTheTopCardForEachCreatureOfYoursThatDiesThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 3, 9)
	small1 := effectstest.B12Creature(g, me.ID, "Small One", "Creature — Elf", 1, 1)
	small2 := effectstest.B12Creature(g, me.ID, "Small Two", "Creature — Elf", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	top := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "Top One", TypeLine: "Sorcery"},
		game.Card{Name: "Top Two", TypeLine: "Sorcery"},
		game.Card{Name: "Third", TypeLine: "Sorcery"},
	)
	effectstest.CastCatalogSpell(t, g, "Waltz of Rage", "Sorcery", b24WaltzOracle, effectstest.DskDGTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
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
		perm := effectstest.ExiledPermission(g, id)
		if perm.Player != me.ID || !effectstest.PermissionLive(g, perm, me.ID) {
			t.Errorf("%+v: you may play the exiled card", perm)
		}
	}
	// Still this turn: another creature of yours dying exiles another card.
	late := effectstest.B12Creature(g, me.ID, "Late", "Creature — Elf", 1, 1)
	effectstest.RooncDestroy(t, g, late)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top[2]) {
		t.Error("a creature dying later this turn did not trigger the delayed ability")
	}
}

func TestStdB24WaltzOfRageEndsWithTheTurnAndFizzlesWithoutItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 3, 9)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Elf", 1, 1)
	top := effectstest.SamiLibraryTopFirst(me, game.Card{Name: "Top One", TypeLine: "Sorcery"}, game.Card{Name: "Top Two", TypeLine: "Sorcery"})
	effectstest.CastCatalogSpell(t, g, "Waltz of Rage", "Sorcery", b24WaltzOracle, effectstest.DskDGTarget(big))
	effectstest.RooncDestroy(t, g, big) // the target is gone: the spell does nothing
	// Not even a delayed ability: another creature dying exiles nothing.
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B24Damage(t, g, other); got != 0 {
		t.Errorf("a fizzled Waltz dealt %d damage", got)
	}
	effectstest.RooncDestroy(t, g, other)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(top[0]) {
		t.Error("a fizzled Waltz still set up its delayed ability")
	}
}

func TestStdB24WaltzOfRageDelayedAbilityLastsOnlyUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 3, 9)
	top := effectstest.SamiLibraryTopFirst(me, game.Card{Name: "Top One", TypeLine: "Sorcery"}, game.Card{Name: "Top Two", TypeLine: "Sorcery"})
	effectstest.CastCatalogSpell(t, g, "Waltz of Rage", "Sorcery", b24WaltzOracle, effectstest.DskDGTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	for i := 0; i < 6 && g.Turn.Step != game.StepUpkeep; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	late := effectstest.B12Creature(g, me.ID, "Late", "Creature — Elf", 1, 1)
	effectstest.RooncDestroy(t, g, late)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(top[0]) {
		t.Error("the delayed ability outlived the turn")
	}
}

// --- Altanak, the Thrice-Called --------------------------------------------------

func TestStdB24AltanakDrawsWhenAnOpponentTargetsItBeforeTheSpellResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	alt := effectstest.B12Push(g, opp.ID, "Altanak, the Thrice-Called", b24AltanakType, b24AltanakOracle, 9, 9)
	if !effectstest.HasKeywordOf(g, alt, "trample") {
		t.Error("Altanak does not have trample")
	}
	before := opp.Hand.Size()
	spell := effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", b24GiantGrowthOracle, effectstest.CardRefs(alt))
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	alt := effectstest.B12Push(g, me.ID, "Altanak, the Thrice-Called", b24AltanakType, b24AltanakOracle, 9, 9)
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", b24GiantGrowthOracle, effectstest.CardRefs(alt))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("targeting your own Altanak drew %d cards", got-before)
	}
}

func TestStdB24AltanakTriggersOnceForASpellThatTargetsItTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	alt := effectstest.B12Push(g, me.ID, "Altanak, the Thrice-Called", b24AltanakType, b24AltanakOracle, 9, 9)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogGraveyardCard(me, "Old Forest", "Basic Land — Forest", "", 0, 0)
	nonland := effectstest.PushCatalogGraveyardCard(me, "Old Bear", "Creature — Bear", "", 2, 2)
	alt := effectstest.PushCatalogHandCard(me, "Altanak, the Thrice-Called", b24AltanakType, b24AltanakOracle)
	if err := effectstest.DskRT2Activate(g, me.ID, alt, 0, "{G}{C}", game.ActivateAbilityParams{Targets: effectstest.CardRefs(nonland)}); err == nil {
		t.Fatal("a nonland card was accepted as the target")
	}
	me.ManaPool = nil
	if err := effectstest.DskRT2Activate(g, me.ID, alt, 0, "{G}{C}", game.ActivateAbilityParams{Targets: effectstest.CardRefs(land)}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if me.Hand.Contains(alt) || !me.Graveyard.Contains(alt) {
		t.Fatal("Altanak was not discarded as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land did not return")
	}
	if c, _ := effectstest.BattlefieldCard(g, land); !c.Tapped {
		t.Error("the land returned untapped")
	}
}

func b24EnterAnthropede(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastAndResolveCreature(t, g, "Anthropede", b24AnthropedeType, b24AnthropedeOrcl)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestStdB24AnthropedeDiscardACardToDestroyARoom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	room := effectstest.B12Permanent(g, opp.ID, "Test Room", "Enchantment — Room")
	other := effectstest.B12Permanent(g, opp.ID, "Test Aura", "Enchantment")
	hand := b24HandCards(me, "Pitch Me")
	b24EnterAnthropede(t, g)
	prompt := effectstest.B24OpenPayPrompt(g, me)
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
	effectstest.AnswerPickTargetIfOpen(t, g, room)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(room) {
		t.Error("the Room was not destroyed")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a non-Room enchantment was destroyed")
	}
}

func TestStdB24AnthropedePayTwoToDestroyARoom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	room := effectstest.B12Permanent(g, me.ID, "My Room", "Enchantment — Room")
	b24HandCards(me, "Keep Me")
	effectstest.ToMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	b24EnterAnthropede(t, g)
	prompt := effectstest.B24OpenPayPrompt(g, me)
	if prompt == nil {
		t.Fatal("no discard-or-pay prompt")
	}
	option := 1
	if err := g.ResolvePayUnlessWithOption(prompt.ID, me.ID, true, &option, nil); err != nil {
		t.Fatalf("pay {2}: %v", err)
	}
	effectstest.AnswerPickTargetIfOpen(t, g, room)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(room) {
		t.Error("the Room (your own is a legal target) was not destroyed")
	}
	if got := effectstest.PoolSize(g, me.ID); got != 0 {
		t.Errorf("{2} was not spent: %d mana left in the pool", got)
	}
}

func TestStdB24AnthropedeDecliningDestroysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	room := effectstest.B12Permanent(g, opp.ID, "Test Room", "Enchantment — Room")
	hand := b24HandCards(me, "Keep Me")
	b24EnterAnthropede(t, g)
	prompt := effectstest.B24OpenPayPrompt(g, me)
	if prompt == nil {
		t.Fatal("no discard-or-pay prompt")
	}
	if err := g.ResolvePayUnless(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(room) || !me.Hand.Contains(hand[0]) {
		t.Error("declining destroyed the Room or discarded a card")
	}
}

func TestStdB24AnthropedeAsksNobodyWhoCantPayEither(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Permanent(g, opp.ID, "Test Room", "Enchantment — Room")
	me.Hand.Cards = nil
	b24EnterAnthropede(t, g)
	if effectstest.B24OpenPayPrompt(g, me) != nil {
		t.Error("asked to pay with an empty hand and no mana")
	}
}

// --- Balustrade Wurm -------------------------------------------------------------

func TestStdB24BalustradeWurmCantBeCounteredAndHasTrampleAndHaste(t *testing.T) {
	spec, _ := effects.Lookup(b24WurmOracle)
	if !spec.CantBeCountered {
		t.Error("Balustrade Wurm can be countered")
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wurm := effectstest.B12Push(g, me.ID, "Balustrade Wurm", b24WurmType, b24WurmOracle, 5, 5)
	effectstest.AssertKeywords(t, g, wurm, "trample", "haste")
}

func TestStdB24BalustradeWurmReturnsOnlyWithDeliriumAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	wurm := effectstest.PushCatalogGraveyardCard(me, "Balustrade Wurm", b24WurmType, b24WurmOracle, 5, 5)
	// The Wurm is the creature card; two more types is not enough.
	effectstest.GyCard(me, "Old Bolt", "Instant", "{R}", "")
	effectstest.GyCard(me, "Old Ritual", "Sorcery", "{B}", "")
	if err := effectstest.DskRT2Activate(g, me.ID, wurm, 0, "{G}{G}{C}{C}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with three card types in the graveyard")
	}
	me.ManaPool = nil
	effectstest.GyCard(me, "Old Forest", "Basic Land — Forest", "", "")
	if err := effectstest.DskRT2Activate(g, me.ID, wurm, 0, "{G}{G}{C}{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with delirium: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Fatal("the Wurm did not return")
	}
	if got := effectstest.CountersOn(g, wurm, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
	effectstest.RooncDestroy(t, g, wurm)
	if me.Graveyard.Contains(wurm) || !g.Exile.Contains(wurm) {
		t.Error("the Wurm with a finality counter went to the graveyard instead of exile")
	}
}

func TestStdB24BalustradeWurmCantReturnOutsideYourMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	wurm := effectstest.PushCatalogGraveyardCard(me, "Balustrade Wurm", b24WurmType, b24WurmOracle, 5, 5)
	effectstest.B24SeedTypes(me, 4)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := effectstest.DskRT2Activate(g, me.ID, wurm, 0, "{G}{G}{C}{C}", game.ActivateAbilityParams{}); err == nil {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	gy1 := effectstest.GyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	gy2 := effectstest.GyCard(me, "Old Bolt", "Instant", "{R}", "")
	effectstest.CastCatalogSpell(t, g, "Cathartic Parting", "Sorcery", b24ParTingOracle, b24PartingTargets(relic, gy1, gy2))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || !b24LibraryContains(opp, relic) {
		t.Fatal("the artifact was not shuffled into its owner's library")
	}
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(gy1) || me.Graveyard.Contains(gy2) {
		t.Error("a targeted graveyard card stayed in the graveyard")
	}
	if !b24LibraryContains(me, gy1) || !b24LibraryContains(me, gy2) {
		t.Error("the targeted graveyard cards are not in your library")
	}
}

func TestStdB24CatharticPartingTheOwnerShufflesAStolenPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, third := g.Seats[1], g.Seats[2]
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Relic", TypeLine: "Artifact",
		Owner: third.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Cathartic Parting", "Sorcery", b24ParTingOracle, b24PartingTargets(stolen))
	effectstest.PassPriorityAroundTable(t, g)
	if !b24LibraryContains(third, stolen) || b24LibraryContains(opp, stolen) {
		t.Error("the permanent did not go to its OWNER's library")
	}
}

func TestStdB24CatharticPartingTheGraveyardHalfIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Enchantment")
	gy := effectstest.GyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.CastCatalogSpell(t, g, "Cathartic Parting", "Sorcery", b24ParTingOracle, b24PartingTargets(relic, gy))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(gy) || b24LibraryContains(me, gy) {
		t.Error("declining moved the graveyard card")
	}
	if b24LibraryContains(opp, relic) == false {
		t.Error("the first half is not optional")
	}
}

func TestStdB24CatharticPartingMayShuffleEvenWithNoGraveyardTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Cathartic Parting", "Sorcery", b24ParTingOracle, b24PartingTargets(relic))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Size() != 0 {
		t.Error("the spell did not finish")
	}
}

func TestStdB24CatharticPartingTargetLimits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	mine := effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	var gy []uuid.UUID
	for i := 0; i < 5; i++ {
		gy = append(gy, effectstest.GyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", ""))
	}
	for name, targets := range map[string][]game.TargetRef{
		"five graveyard cards":      b24PartingTargets(relic, gy...),
		"your own permanent":        b24PartingTargets(mine),
		"an opponent's nonartifact": b24PartingTargets(land),
		"no permanent at all":       b24PartingTargets(relic)[:0],
		"an opponent's graveyard":   {{Kind: game.TargetCard, ID: relic, Slot: 0}, {Kind: game.TargetCard, ID: effectstest.GyCard(opp, "Their Bear", "Creature — Bear", "", ""), Slot: 1}},
	} {
		id := effectstest.PutInHand(me, game.Card{Name: "Cathartic Parting", TypeLine: "Sorcery", OracleID: b24ParTingOracle})
		effectstest.ToMain(t, g)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err == nil {
			t.Errorf("%s: the cast was accepted", name)
		}
	}
	// Four graveyard cards is the most.
	id := effectstest.PutInHand(me, game.Card{Name: "Cathartic Parting", TypeLine: "Sorcery", OracleID: b24ParTingOracle})
	effectstest.ToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: b24PartingTargets(relic, gy[:4]...)}); err != nil {
		t.Errorf("four graveyard cards: %v", err)
	}
}
