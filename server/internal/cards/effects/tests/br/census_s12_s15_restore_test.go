package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// census_s12_s15_restore_test.go — #4361 S12 and S15, with real cards.
// Each test takes a STRICT restore while the prompt is open, answers it on
// the restored table and checks the card still finishes off the disk.

// passUntilCopyOnStack passes priority until a spell copy is on the stack
// and returns its stack ID.
func passUntilCopyOnStack(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	for i := 0; i < 16; i++ {
		for id, it := range g.StackMeta {
			if it != nil && it.IsCopy && it.Kind == game.StackItemSpell {
				return id
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("no spell copy reached the stack")
	return uuid.Nil
}

// S12 (owner decision D6, ADR 0104): Double Major copies a Grizzly Bears
// spell, and Aethersnatch steals the copy on the stack. The copy becomes a
// token as it resolves (CR 608.3f), and two entry replacements ask the
// thief, its controller, for the CR 616 order of its entry. The copy's
// control records ride the entry's tail as data, so the order prompt is a
// restore point, and the restored answer puts one token onto the
// battlefield that is the thief's with the copy's controller as its
// default (CR 110.2b). #4994: the token is not created (CR 111.13), so the
// thief's Doubling Season and Parallel Lives neither double it nor give
// the prompt; they stay on the board to show it.
func TestCensusS12AStolenPermanentSpellCopysTokenIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, thief := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.TokenOrderBoard(g, thief.ID)
	pa, pb := effectstest.SeedTokenEntryOrderProbes(g, thief.ID)
	spell := effectstest.CastCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	cp := passUntilCopyOnStack(t, g)
	effectstest.RespondWith(t, g, thief, game.Card{Name: "Aethersnatch", TypeLine: "Instant", OracleID: effectstest.AethersnatchOracle},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: cp}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	if g.PendingChoices[0].Chooser != thief.ID {
		t.Fatalf("the order prompt went to %s, want the thief %s", g.PendingChoices[0].Chooser, thief.ID)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	if tokens, _ := effectstest.CopiesOnBattlefield(g, "Grizzly Bears"); len(tokens) != 0 {
		t.Fatalf("the restored table made %d tokens before the order", len(tokens))
	}
	effectstest.AnswerOrderBySource(t, g, pa, pb)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	tokens, _ := effectstest.CopiesOnBattlefield(g, "Grizzly Bears")
	if len(tokens) != 1 {
		t.Fatalf("%d token copies, want one: the copy's token is not created, so nothing doubles it (CR 111.13)", len(tokens))
	}
	for _, tok := range tokens {
		if tok.Controller != thief.ID {
			t.Errorf("a token is controlled by %s, want the thief %s", tok.Controller, thief.ID)
		}
		if tok.BaseController != caster.ID {
			t.Errorf("a token's default controller is %s, want the copy's controller %s: the restored tail did not re-pin the copy's control records",
				tok.BaseController, caster.ID)
		}
	}
}

// reverberateUntilCopyPrompt casts Reverberate on spell and passes
// priority until the copy's new-targets prompt is open.
func reverberateUntilCopyPrompt(t *testing.T, g *game.Game, me uuid.UUID, spell uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Reverberate", "Instant", effectstest.ReverberateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	for i := 0; i < 8 && effectstest.CopyPromptFor(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.CopyPromptFor(g, me) == nil {
		t.Fatal("Reverberate's copy opened no new-targets prompt")
	}
}

// S15: Reverberate on a Lightning Bolt aimed at one opponent. The copy's
// prompt is a restore point, and the restored answer points the copy at
// another opponent: each takes 3.
func TestCensusS15AReverberateCopysNewTargetIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim, other := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	reverberateUntilCopyPrompt(t, g, me, bolt)
	before := effectstest.LifeOf(g, other)

	g = effectstest.StrictRestoreWithCopyWalk(t, g)
	p := effectstest.CopyPromptFor(g, me)
	if p == nil {
		t.Fatal("the restored table has no copy prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me, game.TargetRef{Kind: game.TargetPlayer, ID: other}); err != nil {
		t.Fatalf("ResolvePickTarget on the restored prompt: %v", err)
	}
	if effectstest.StackCopies(g) != 1 {
		t.Fatalf("the restored answer made %d copies, want 1", effectstest.StackCopies(g))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, other); got != before-3 {
		t.Errorf("the copy's new target has %d life, want %d: the restored walk did not build the copy", got, before-3)
	}
	if got := effectstest.LifeOf(g, victim); got != before-3 {
		t.Errorf("the Bolt's target has %d life, want %d", got, before-3)
	}
}

// S15: Reverberate on Bite Down, a copy walk of two steps (#2622). The
// walk is a restore point at its first step and, after that answer, at
// its second, and the copy is built from both restored answers.
func TestCensusS15AReverberateCopyOfBiteDownIsARestorePointAtEachStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.Pr7Creature(g, me.ID, "Mine", 3)
	mine2 := effectstest.Pr7Creature(g, me.ID, "Mine 2", 2)
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2)
	theirs2 := effectstest.Pr7Creature(g, opp.ID, "Theirs 2", 2)
	bite := effectstest.CastCopyTarget(t, g, me, "Bite Down", "Instant", effectstest.B26BiteDownOracle,
		game.CastSpellParams{Targets: effectstest.CardTargets(mine, theirs)})
	reverberateUntilCopyPrompt(t, g, me.ID, bite)

	g = effectstest.StrictRestoreWithCopyWalk(t, g)
	p := effectstest.CopyPromptFor(g, me.ID)
	if p == nil {
		t.Fatal("the restored table has no copy prompt at the first step")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine2}); err != nil {
		t.Fatalf("first step on the restored prompt: %v", err)
	}
	if effectstest.CopyPromptFor(g, me.ID) == nil {
		t.Fatal("the restored walk did not ask its second step")
	}

	g = effectstest.StrictRestoreWithCopyWalk(t, g)
	p = effectstest.CopyPromptFor(g, me.ID)
	if p == nil {
		t.Fatal("the restored table has no copy prompt at the second step")
	}
	for _, id := range p.PickTargetCards {
		if id == mine || id == mine2 {
			t.Errorf("the second step offers %s, a creature you control: the restored walk asked the wrong clause", id)
		}
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs2}); err != nil {
		t.Fatalf("second step on the restored prompt: %v", err)
	}
	var cp *game.StackItem
	for _, it := range g.StackMeta {
		if it != nil && it.IsCopy {
			cp = it
		}
	}
	if cp == nil {
		t.Fatal("the restored walk made no copy")
	}
	if len(cp.Targets) != 2 || cp.Targets[0].ID != mine2 || cp.Targets[1].ID != theirs2 {
		t.Fatalf("the copy targets %+v, want Mine 2 then Theirs 2", cp.Targets)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, theirs2); ok && c.DamageMarked != 2 {
		t.Errorf("Theirs 2 has %d damage, want Mine 2's 2 from the copy", c.DamageMarked)
	}
}

// redcapTriggerTargeting casts a live black-red Murderous Redcap and
// aims its enter trigger at target.
func redcapTriggerTargeting(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) uuid.UUID {
	t.Helper()
	redcap := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: redcap, Name: "Murderous Redcap", TypeLine: "Creature — Goblin Assassin",
		OracleID: effectstest.B40MurderousRedcapOracle, Power: 2, Toughness: 2, Colors: []string{"B", "R"},
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, redcap, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Redcap: %v", err)
	}
	effectstest.PassUntilOnBattlefield(t, g, redcap)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID,
		game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	return redcap
}

// S15: Strionic Resonator copies a live black-red Murderous Redcap's enter
// trigger (an ability copy, its item stamped with its catalog row). The
// copy's prompt is a restore point: the restored walk still judges the
// new target against the Redcap (protection from red refuses it, from
// blue does not), and the copy it builds deals the Redcap's 2.
func TestCensusS15AStrionicResonatorCopysNewTargetIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	plain, proRed, proBlue := effectstest.ProtectedCreatures(g, opp)
	redcap := redcapTriggerTargeting(t, g, me, plain)
	effectstest.ResonatorCopiesTheRedcapTrigger(t, g, me, redcap)

	g = effectstest.StrictRestoreWithCopyWalk(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the restored table has no copy prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: proRed}}); err == nil {
		t.Fatal("the restored walk accepted the pro-red creature for a black-red Redcap's copy")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: proBlue}}); err != nil {
		t.Fatalf("the restored walk refused the pro-blue creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, proBlue); c.DamageMarked != 2 {
		t.Errorf("the pro-blue creature has %d damage, want 2 from the restored copy", c.DamageMarked)
	}
	if c, _ := effectstest.BattlefieldCard(g, plain); c.DamageMarked != 2 {
		t.Errorf("the plain creature has %d damage, want 2 from the original trigger", c.DamageMarked)
	}
}

// S15: the departed-source line (#1449) across a restore. The Redcap was
// turned blue and died with its trigger waiting, and the Resonator's copy
// is judged against the Redcap as it last existed. The restored walk
// reads the same last-known record off the written item.
func TestCensusS15AResonatorCopyOfADepartedSourceIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	plain, proRed, proBlue := effectstest.ProtectedCreatures(g, opp)
	redcap := effectstest.RedcapTurnedBlueAndKilled(t, g, game.TargetRef{Kind: game.TargetCard, ID: plain})
	effectstest.ResonatorCopiesTheRedcapTrigger(t, g, me, redcap)

	g = effectstest.StrictRestoreWithCopyWalk(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the restored table has no copy prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: proBlue}}); err == nil {
		t.Fatal("the restored walk accepted the pro-blue creature: it judged the black-red graveyard card (#1449)")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: proRed}}); err != nil {
		t.Fatalf("the restored walk refused the pro-red creature: %v (#1449)", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, proRed); c.DamageMarked != 2 {
		t.Errorf("the pro-red creature has %d damage, want 2 from the restored copy", c.DamageMarked)
	}
}
