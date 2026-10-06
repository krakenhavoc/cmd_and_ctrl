package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cost_commander_cards_test.go — #1397's proof cards: five catalog
// cards whose COST can move a commander, one per cost component the
// fix touched. Each one used to settle without asking (the discard,
// return and exile payers) or pause half way through the payment (the
// sacrifice). The engine half — every component, both answers, undo —
// is game/cost_commander_choice_test.go.
//
// Since ADR 0115 only a cost that puts a card into a HAND or a LIBRARY
// (CR 903.9b) parks the announcement on a question to the commander's
// owner and pays nothing until it is answered: ninjutsu below. A commander
// discarded, exiled or sacrificed to pay a cost is paid like any other
// card, and the CR 903.9a state-based action asks its owner afterwards;
// the other four cards here pin that order.

// costCommander builds a commander card owned by `owner`.
func costCommander(owner uuid.UUID, name string) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Human Wizard",
		Power: 3, Toughness: 3, Owner: owner, Controller: owner, IsCommander: true,
		KnownBy: map[uuid.UUID]bool{owner: true},
	}
}

// answerCostCommander answers the one open CR 903.9b prompt as `owner`,
// after checking it is addressed to them and nothing has been paid.
func answerCostCommander(t *testing.T, g *game.Game, owner uuid.UUID, apply bool) {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the one CR 903.9 prompt", len(g.PendingChoices))
	}
	c := g.PendingChoices[0]
	if c.Kind != game.PendingChoiceOptionalReplacement || c.Chooser != owner {
		t.Fatalf("prompt %q owed by %s, want an optional_replacement owed by the commander's owner %s", c.Kind, c.Chooser, owner)
	}
	if err := g.ResolveOptionalReplacement(c.ID, owner, apply); err != nil {
		t.Fatalf("answer: %v", err)
	}
}

// Ninja of the Deep Hours — ninjutsu's "Return an unblocked attacker
// you control to hand" (CR 702.49a) spent on YOUR attacking commander.
// Answered "command zone", the commander goes there, and the Ninja still
// enters attacking the player the commander was attacking: the re-made
// activation reads the attack before the commander leaves, exactly as
// the first one would have.
func TestNinjutsuOnAnAttackingCommanderAsksItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmd := pushBattlefieldCardWithTimestamp(g, costCommander(me.ID, "My Commander"))
	ninja := pushNinjaToHand(me, "Ninja of the Deep Hours", "Creature — Human Ninja", ninjaOfTheDeepHoursOracle, 2, 2)
	declareAttack(t, g, opp.ID, cmd)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatal(err)
	}

	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{cmd}}); err != nil {
		t.Fatalf("ninjutsu: %v", err)
	}
	if !g.Battlefield.Contains(cmd) || len(g.StackMeta) != 0 {
		t.Fatal("ninjutsu was paid before the commander's owner answered")
	}
	answerCostCommander(t, g, me.ID, true)
	if !me.Command.Contains(cmd) || me.Hand.Contains(cmd) {
		t.Fatal("the returned commander did not take the command zone")
	}
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(ninja)
	if !ok || !g.Battlefield.Contains(ninja) {
		t.Fatal("the ninja did not enter")
	}
	if c.AttackingTarget != opp.ID {
		t.Errorf("the ninja is attacking %s, want %s — the player the commander was attacking", c.AttackingTarget, opp.ID)
	}
}

// Thrill of Possibility — "As an additional cost to cast this spell,
// discard a card." A commander discarded to it goes to the graveyard as
// the cost is paid, the spell is cast, and its owner is asked about the
// command zone afterwards (CR 903.9a); declined, it stays there.
func TestThrillOfPossibilityDiscardingACommanderPaysThenAsksItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	cmd := costCommander(me.ID, "My Commander")
	me.Hand.PushTop(cmd)
	spell := pushCatalogHandCard(me, "Thrill of Possibility", "Instant", thrillOfPossibilityOracle)
	before := me.Library.Size()

	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{DiscardIDs: []uuid.UUID{cmd.InstanceID}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !me.Graveyard.Contains(cmd.InstanceID) || !g.Stack.Contains(spell) {
		t.Fatal("the discard was not paid and the spell cast without asking first")
	}
	answerCommanderReturn(t, g, me.ID, false)
	if !me.Graveyard.Contains(cmd.InstanceID) {
		t.Fatal("the declined commander is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

// Cadaverous Bloom — "Exile a card from your hand: Add {B}{B} or
// {G}{G}." A mana ability (CR 605.3a): the commander is exiled as the
// cost is paid and the ability asks its colour; its owner is then asked
// about the command zone (CR 903.9a) and takes it.
func TestCadaverousBloomExilingACommanderPaysThenAsksItsOwner(t *testing.T) {
	g, me, bloom, _ := seatBloomWithHand(t, 0)
	cmd := costCommander(me.ID, "My Commander")
	me.Hand.PushTop(cmd)

	if err := g.ActivateManaAbility(me.ID, bloom, 0, game.ManaAbilityParams{ExileIDs: []uuid.UUID{cmd.InstanceID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(cmd.InstanceID) {
		t.Fatal("the commander was not exiled to pay for the Bloom")
	}
	if latestChoiceOfKind(g, game.PendingChoiceMana) == nil {
		t.Fatal("want the Bloom's {B}{B} / {G}{G} pick")
	}
	answerCommanderReturn(t, g, me.ID, true)
	if !me.Command.Contains(cmd.InstanceID) || g.Exile.Contains(cmd.InstanceID) {
		t.Fatal("the commander did not take the command zone")
	}
}

// Grim Lavamancer — "{R}, {T}, Exile two cards from your graveyard:
// This creature deals 2 damage to any target." One of the two is a
// commander its owner let die. Both are exiled as the cost is paid,
// the ability goes on the stack, and the owner is asked again about
// the command zone (CR 903.9a: it was put into exile) and takes it.
func TestGrimLavamancerExilingACommanderFromTheGraveyardPaysThenAsksItsOwner(t *testing.T) {
	g, me, opp := exileCostTable(t)
	lava := pushCatalogPermanent(g, me.ID, "Grim Lavamancer", "Creature — Human Wizard", grimLavamancerOracle, false)
	cmd := costCommander(me.ID, "My Commander")
	me.Graveyard.PushTop(cmd)
	other := pushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	floatMana(t, g, me, "{R}")
	params := game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{cmd.InstanceID, other},
		Targets:  []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}

	if err := g.ActivateCatalogAbility(me.ID, lava, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(cmd.InstanceID) || !g.Exile.Contains(other) {
		t.Fatal("the two named cards were not exiled to pay for the Lavamancer")
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("StackMeta = %d, want the Lavamancer's ability", len(g.StackMeta))
	}
	answerCommanderReturn(t, g, me.ID, true)
	if !me.Command.Contains(cmd.InstanceID) {
		t.Error("the commander did not take the command zone")
	}
	life := opp.Life
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
}

// Viscera Seer — "Sacrifice a creature: Scry 1." The creature is an
// OPPONENT's commander I have stolen. It is sacrificed as the cost is
// paid (so it cannot be spent twice), and the question afterwards goes
// to its OWNER, not to me (CR 903.9a, "its owner").
func TestVisceraSeerSacrificingAStolenCommanderPaysThenAsksItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seer := pushCatalogPermanent(g, me.ID, "Viscera Seer", "Creature — Vampire Wizard", visceraSeerOracle, false)
	stolen := costCommander(opp.ID, "Their Commander")
	stolen.Controller = me.ID
	cmd := pushBattlefieldCardWithTimestamp(g, stolen)

	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{cmd}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(cmd) || !opp.Graveyard.Contains(cmd) {
		t.Fatal("the stolen commander was not sacrificed into its owner's graveyard")
	}
	if len(g.StackMeta) != 1 {
		t.Errorf("StackMeta = %d, want the Seer's scry", len(g.StackMeta))
	}
	if commanderReturnPromptFor(g, me.ID) != nil {
		t.Error("the controller was asked; the question is the owner's")
	}
	answerCommanderReturn(t, g, opp.ID, true)
	if !opp.Command.Contains(cmd) {
		t.Fatal("the stolen commander did not reach its OWNER's command zone")
	}
}
