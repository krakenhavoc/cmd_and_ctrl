package testsbr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Rhystic Study --------------------------------------------

func TestRhysticStudyOpponentCastDeclineDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	studyID := effectstest.PushPermanentForTest(g, owner.ID, "Rhystic Study", effectstest.RhysticStudyOracle, "Enchantment")
	handBefore := owner.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})

	// The Study's trigger sits above the Bolt (cast in response to
	// the cast event, higher Seq).
	item := effectstest.TriggerOnStack(g, studyID)
	if item == nil {
		t.Fatalf("Rhystic Study trigger not on the stack after an opponent cast")
	}
	if item.Controller != owner.ID {
		t.Errorf("trigger controller = %s, want the Study's controller %s", item.Controller, owner.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Resolution asks the CASTER to pay, not the Study's owner.
	if !effectstest.HasPayUnlessFor(g, caster.ID) {
		t.Fatalf("no pay_unless prompt for the caster after the trigger resolved")
	}
	if owner.Hand.Size() != handBefore {
		t.Fatalf("Study drew before the payer answered")
	}
	prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{1}" {
		t.Errorf("PayCost = %q, want {1}", prompt.PayCost)
	}
	// #1565: the decline no longer draws automatically — it opens a
	// second, independent "you may draw a card?" for the Study's
	// controller (MayChoice, #796).
	if owner.Hand.Size() != handBefore {
		t.Fatalf("Study drew before its controller answered 'draw a card?'")
	}
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != owner.ID {
		t.Fatalf("no 'draw a card?' prompt addressed to the Study's controller")
	}
	if err := g.ResolveConfirm(ask.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveConfirm(draw): %v", err)
	}
	if got := owner.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Rhystic Study on decline, then 'yes': owner hand delta %d, want 1", got)
	}
}

// TestRhysticStudyControllerMayDeclineTheDraw closes the caveat
// #1565 audited: "you may draw a card" really is optional now, not a
// forced draw — the controller can say "no" instead of being handed
// a card (or, on an empty library, the loss CR 704.5b would give
// them for drawing from nothing).
func TestRhysticStudyControllerMayDeclineTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Rhystic Study", effectstest.RhysticStudyOracle, "Enchantment")
	handBefore := owner.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, caster.ID, false)

	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != owner.ID {
		t.Fatalf("no 'draw a card?' prompt addressed to the Study's controller")
	}
	if err := g.ResolveConfirm(ask.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveConfirm(decline): %v", err)
	}
	if owner.Hand.Size() != handBefore {
		t.Errorf("owner hand changed by %d after declining the draw, want 0", owner.Hand.Size()-handBefore)
	}
}

func TestRhysticStudyOpponentPaysNoDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Rhystic Study", effectstest.RhysticStudyOracle, "Enchantment")
	handBefore := owner.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	// Float the {1} now — castCatalogSpell walks steps to reach the
	// main phase, and pools empty on every step change.
	caster.ManaPool.AddMana(game.ManaToken{Color: "C"})
	effectstest.AnswerPayUnless(t, g, caster.ID, true)

	if owner.Hand.Size() != handBefore {
		t.Errorf("Study drew even though the caster paid: delta %d", owner.Hand.Size()-handBefore)
	}
	if len(caster.ManaPool) != 0 {
		t.Errorf("payment did not leave the caster's pool: %d tokens left", len(caster.ManaPool))
	}
}

func TestRhysticStudyIgnoresControllersOwnCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	studyID := effectstest.PushPermanentForTest(g, caster.ID, "Rhystic Study", effectstest.RhysticStudyOracle, "Enchantment")

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	if effectstest.TriggerOnStack(g, studyID) != nil {
		t.Errorf("Rhystic Study triggered on its own controller's cast")
	}
}

// --- Smothering Tithe -----------------------------------------

func TestSmotheringTitheOpponentDrawDeclineMakesTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	drawer := g.Seats[0]
	owner := g.Seats[1]
	titheID := effectstest.PushPermanentForTest(g, owner.ID, "Smothering Tithe", effectstest.SmotheringTitheOracle, "Enchantment")

	// The manual draw has to happen off the active seat's own draw
	// step: Game.DrawCard is a deliberate no-op there, because the
	// turn-based draw has already fired. newCatalogGame parks the
	// cursor on that step (#692), so walk to the main phase first.
	effectstest.AdvanceToMain(t, g)

	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	// The draw is a sandbox action outside any resolution — the
	// trigger has to land on the stack at that boundary.
	if effectstest.TriggerOnStack(g, titheID) == nil {
		t.Fatalf("Tithe trigger not on the stack after an opponent drew")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, drawer.ID) {
		t.Fatalf("no pay_unless prompt for the drawer")
	}
	prompt := effectstest.AnswerPayUnless(t, g, drawer.ID, false)
	if prompt.PayCost != "{2}" {
		t.Errorf("PayCost = %q, want {2}", prompt.PayCost)
	}
	if n := effectstest.CountOnBattlefield(g, "Treasure", owner.ID); n != 1 {
		t.Errorf("Treasures under the Tithe's controller: %d, want 1", n)
	}
}

func TestSmotheringTitheOpponentPaysNoTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	drawer := g.Seats[0]
	owner := g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Smothering Tithe", effectstest.SmotheringTitheOracle, "Enchantment")
	// The manual draw has to happen off the active seat's own draw
	// step: Game.DrawCard is a deliberate no-op there, because the
	// turn-based draw has already fired. newCatalogGame parks the
	// cursor on that step (#692), so walk to the main phase first.
	// The mana goes in AFTER the walk — the pool empties at the end
	// of every step (CR 500.5).
	effectstest.AdvanceToMain(t, g)
	drawer.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, drawer.ID, true)
	if n := effectstest.CountOnBattlefield(g, "Treasure", owner.ID); n != 0 {
		t.Errorf("Treasure created despite payment: %d", n)
	}
}

func TestSmotheringTitheIgnoresOwnDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0]
	titheID := effectstest.PushPermanentForTest(g, owner.ID, "Smothering Tithe", effectstest.SmotheringTitheOracle, "Enchantment")
	if err := g.DrawCard(owner.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if effectstest.TriggerOnStack(g, titheID) != nil || len(g.PendingTriggers) != 0 {
		t.Errorf("Tithe triggered on its own controller's draw")
	}
}

// --- Esper Sentinel -------------------------------------------

func TestEsperSentinelFirstNoncreatureSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	sentinelID := effectstest.PushDiesCreatureForTest(g, owner.ID, "Esper Sentinel", effectstest.EsperSentinelOracle,
		"Artifact Creature — Human Soldier", 1, 1)
	handBefore := owner.Hand.Size()

	// A creature spell first: not a noncreature spell, no trigger.
	effectstest.CastCatalogSpell(t, g, "Colossal Dreadmaw", "Creature — Dinosaur", effectstest.ColossalDreadmawOracle, nil)
	if effectstest.TriggerOnStack(g, sentinelID) != nil {
		t.Fatalf("Sentinel triggered on a creature spell")
	}
	effectstest.PassPriorityAroundTable(t, g)

	// First noncreature spell this turn: trigger.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	if effectstest.TriggerOnStack(g, sentinelID) == nil {
		t.Fatalf("Sentinel did not trigger on the first noncreature spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{1}" {
		t.Errorf("PayCost = %q, want {1} (Sentinel's power)", prompt.PayCost)
	}
	if got := owner.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Sentinel decline: owner hand delta %d, want 1", got)
	}

	// Second noncreature spell the same turn: no trigger.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	if effectstest.TriggerOnStack(g, sentinelID) != nil {
		t.Errorf("Sentinel triggered on the second noncreature spell of the turn")
	}
}

func TestEsperSentinelXReadsPumpedPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	owner := g.Seats[1]
	sentinelID := effectstest.PushDiesCreatureForTest(g, owner.ID, "Esper Sentinel", effectstest.EsperSentinelOracle,
		"Artifact Creature — Human Soldier", 1, 1)
	// Two +1/+1 counters → power 3 → the ask is {3}.
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(sentinelID, "+1/+1", 2); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{3}" {
		t.Errorf("PayCost = %q, want {3}", prompt.PayCost)
	}
}

// --- Beast Whisperer ------------------------------------------

func TestBeastWhispererDrawsOnOwnCreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	whispererID := effectstest.PushDiesCreatureForTest(g, caster.ID, "Beast Whisperer", effectstest.BeastWhispererOracle,
		"Creature — Elf Druid", 2, 3)

	// Noncreature: nothing.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	if effectstest.TriggerOnStack(g, whispererID) != nil {
		t.Fatalf("Beast Whisperer triggered on an instant")
	}
	effectstest.PassPriorityAroundTable(t, g)
	handBefore := caster.Hand.Size()

	dreadmawID := effectstest.CastCatalogSpell(t, g, "Colossal Dreadmaw", "Creature — Dinosaur", effectstest.ColossalDreadmawOracle, nil)
	trig := effectstest.TriggerOnStack(g, whispererID)
	if trig == nil {
		t.Fatalf("Beast Whisperer did not trigger on a creature spell")
	}
	spell := g.StackMeta[dreadmawID]
	if spell == nil || !(trig.Seq > spell.Seq) {
		t.Errorf("trigger must sit above the creature spell (resolves first)")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := caster.Hand.Size() - handBefore; got != 1 {
		t.Errorf("Beast Whisperer: hand delta %d, want 1", got)
	}
	if !g.Battlefield.Contains(dreadmawID) {
		t.Errorf("the creature spell itself did not resolve")
	}
}

// --- Consecrated Sphinx ---------------------------------------

func TestConsecratedSphinxOpponentDrawYesDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	drawer := g.Seats[0]
	owner := g.Seats[1]
	sphinxID := effectstest.PushDiesCreatureForTest(g, owner.ID, "Consecrated Sphinx", effectstest.ConsecratedSphinxOracle,
		"Creature — Sphinx", 4, 6)
	// The manual draw has to happen off the active seat's own draw
	// step: Game.DrawCard is a deliberate no-op there, because the
	// turn-based draw has already fired. newCatalogGame parks the
	// cursor on that step (#692), so walk to the main phase first.
	effectstest.AdvanceToMain(t, g)
	handBefore := owner.Hand.Size()

	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	g.RunStateChecksForTest()
	if effectstest.TriggerOnStack(g, sphinxID) == nil {
		t.Fatalf("Sphinx trigger not on the stack before it resolved")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, owner.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := owner.Hand.Size() - handBefore; got != 2 {
		t.Errorf("Sphinx: owner hand delta %d, want 2", got)
	}
	// The Sphinx's own draws must not re-trigger it.
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Errorf("Sphinx re-triggered on its controller's own draws")
		}
	}
}

func TestConsecratedSphinxDeclineDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	drawer := g.Seats[0]
	owner := g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, owner.ID, "Consecrated Sphinx", effectstest.ConsecratedSphinxOracle,
		"Creature — Sphinx", 4, 6)
	// The manual draw has to happen off the active seat's own draw
	// step: Game.DrawCard is a deliberate no-op there, because the
	// turn-based draw has already fired. newCatalogGame parks the
	// cursor on that step (#692), so walk to the main phase first.
	effectstest.AdvanceToMain(t, g)
	handBefore := owner.Hand.Size()

	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, owner.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if owner.Hand.Size() != handBefore {
		t.Errorf("Sphinx drew after a decline")
	}
}
