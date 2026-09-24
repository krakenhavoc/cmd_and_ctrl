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
// sacrifice). Each now parks the announcement on a CR 903.9 question
// to the commander's owner, pays nothing until it is answered, and
// finishes the card once it is. The engine half — every component,
// both answers, undo — is game/cost_commander_choice_test.go.

// costCommander builds a commander card owned by `owner`.
func costCommander(owner uuid.UUID, name string) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Human Wizard",
		Power: 3, Toughness: 3, Owner: owner, Controller: owner, IsCommander: true,
		KnownBy: map[uuid.UUID]bool{owner: true},
	}
}

// answerCostCommander answers the one open CR 903.9 prompt as `owner`,
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
// discard a card." A commander pitched to it and declined goes to the
// graveyard, and the spell is cast and draws its two.
func TestThrillOfPossibilityDiscardingACommanderAsksItsOwner(t *testing.T) {
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
	if !me.Hand.Contains(cmd.InstanceID) || g.Stack.Contains(spell) {
		t.Fatal("the spell was paid for before the commander's owner answered")
	}
	answerCostCommander(t, g, me.ID, false)
	if !me.Graveyard.Contains(cmd.InstanceID) {
		t.Fatal("the declined commander is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

// Cadaverous Bloom — "Exile a card from your hand: Add {B}{B} or
// {G}{G}." A mana ability (CR 605.3a) that cannot pause once begun, so
// the question comes before it begins; answered "command zone", the
// commander goes there and the ability then asks its colour.
func TestCadaverousBloomExilingACommanderAsksItsOwner(t *testing.T) {
	g, me, bloom, _ := seatBloomWithHand(t, 0)
	cmd := costCommander(me.ID, "My Commander")
	me.Hand.PushTop(cmd)

	if err := g.ActivateManaAbility(me.ID, bloom, 0, game.ManaAbilityParams{ExileIDs: []uuid.UUID{cmd.InstanceID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Hand.Contains(cmd.InstanceID) {
		t.Fatal("the Bloom was paid before the commander's owner answered")
	}
	answerCostCommander(t, g, me.ID, true)
	if !me.Command.Contains(cmd.InstanceID) || g.Exile.Contains(cmd.InstanceID) {
		t.Fatal("the commander did not take the command zone")
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceMana {
		t.Fatalf("want the Bloom's {B}{B} / {G}{G} pick after the answer, got %d prompt(s)", len(g.PendingChoices))
	}
}

// Grim Lavamancer — "{R}, {T}, Exile two cards from your graveyard:
// This creature deals 2 damage to any target." One of the two is a
// commander its owner let die; asked again as the cost is paid, the
// owner takes the command zone, and the other card is exiled as named.
func TestGrimLavamancerExilingACommanderFromTheGraveyardAsksItsOwner(t *testing.T) {
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
	if !me.Graveyard.Contains(cmd.InstanceID) || !me.Graveyard.Contains(other) || len(g.StackMeta) != 0 {
		t.Fatal("the Lavamancer was paid before the commander's owner answered")
	}
	answerCostCommander(t, g, me.ID, true)
	if !me.Command.Contains(cmd.InstanceID) {
		t.Error("the commander did not take the command zone")
	}
	if !g.Exile.Contains(other) {
		t.Error("the other named card was not exiled")
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("StackMeta = %d, want the Lavamancer's ability", len(g.StackMeta))
	}
	life := opp.Life
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
}

// Viscera Seer — "Sacrifice a creature: Scry 1." The creature is an
// OPPONENT's commander I have stolen, so the question goes to them, not
// to me; and nothing is sacrificed while they decide, which is what
// kept the same commander from being sacrificed twice.
func TestVisceraSeerSacrificingAStolenCommanderAsksItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seer := pushCatalogPermanent(g, me.ID, "Viscera Seer", "Creature — Vampire Wizard", visceraSeerOracle, false)
	stolen := costCommander(opp.ID, "Their Commander")
	stolen.Controller = me.ID
	cmd := pushBattlefieldCardWithTimestamp(g, stolen)

	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{cmd}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Battlefield.Contains(cmd) || len(g.StackMeta) != 0 {
		t.Fatal("the Seer was paid before the commander's owner answered")
	}
	answerCostCommander(t, g, opp.ID, true)
	if !opp.Command.Contains(cmd) {
		t.Fatal("the stolen commander did not reach its OWNER's command zone")
	}
	if len(g.StackMeta) != 1 {
		t.Errorf("StackMeta = %d, want the Seer's scry", len(g.StackMeta))
	}
}
