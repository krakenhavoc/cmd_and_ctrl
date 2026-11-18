package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// twinning_staff_test.go — #5918: a copy instruction is a replaceable
// event (RepEventCopySpell), and Twinning Staff adds one copy to each
// instruction its controller carries out, whatever made it: storm,
// casualty, conspire, a copy for each Golem (CR 707.10e), the Staff's own
// ability, or a copy with no new targets of its own.

const twinningStaffOracle = "3ca8113b-f5b8-41a7-aae6-ebb32554dfe3"

// twinningStaffOn puts a Twinning Staff onto `owner`'s battlefield.
func twinningStaffOn(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Twinning Staff", "Artifact", twinningStaffOracle, false)
}

// twinningStaffCopies is how many spell copies were put on the stack.
func twinningStaffCopies(g *game.Game) int {
	return acCountEvents(g, game.EventSpellCopied)
}

// Storm count two is ONE instruction of two copies, so the Staff makes it
// three, not four: four resolutions in all.
func TestTwinningStaffAddsOneCopyToAStormInstruction(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	twinningStaffOn(g, me)
	before := lifeOf(g, opp)

	stormFiller(t, g)
	stormFiller(t, g)
	castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	answerCopyPrompts(t, g, me, opp, 3)

	if got := twinningStaffCopies(g); got != 3 {
		t.Errorf("%d copies, want 3 (storm count 2, plus one)", got)
	}
	if got := lifeOf(g, opp); got != before-4 {
		t.Errorf("life %d → %d, want -4 (the spell and three copies)", before, got)
	}
}

// Two Staffs each apply once to the instruction (CR 614.5).
func TestTwoTwinningStaffsAddTwoCopies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	twinningStaffOn(g, me)
	twinningStaffOn(g, me)
	before := lifeOf(g, opp)

	stormFiller(t, g)
	castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	answerCopyPrompts(t, g, me, opp, 3)

	if got := twinningStaffCopies(g); got != 3 {
		t.Errorf("%d copies, want 3 (storm count 1, plus one per Staff)", got)
	}
	if got := lifeOf(g, opp); got != before-4 {
		t.Errorf("life %d → %d, want -4", before, got)
	}
}

// "One or more times": a storm count of zero copies nothing, so there is
// nothing for the Staff to add to.
func TestTwinningStaffDoesNotCopyWhenNothingIsCopied(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	twinningStaffOn(g, me)
	before := lifeOf(g, opp)

	castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	answerCopyPrompts(t, g, me, opp, 0)

	if got := twinningStaffCopies(g); got != 0 {
		t.Errorf("%d copies, want 0", got)
	}
	if got := lifeOf(g, opp); got != before-1 {
		t.Errorf("life %d → %d, want -1", before, got)
	}
}

// "If YOU would copy": an opponent's Staff does nothing to my storm.
func TestAnOpponentsTwinningStaffDoesNotAddToYourCopies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	twinningStaffOn(g, opp)
	before := lifeOf(g, opp)

	stormFiller(t, g)
	castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	answerCopyPrompts(t, g, me, opp, 1)

	if got := twinningStaffCopies(g); got != 1 {
		t.Errorf("%d copies, want 1", got)
	}
	if got := lifeOf(g, opp); got != before-2 {
		t.Errorf("life %d → %d, want -2", before, got)
	}
}

// The Staff's own ability is a copy instruction, so with the Staff out it
// makes two copies of a Lightning Bolt you control, each offered new
// targets. It can't target an opponent's spell, and it costs {7}.
func TestTwinningStaffsAbilityCopiesTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	staff := twinningStaffOn(g, me.ID)
	theirs := castFromSeat(t, g, opp, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	bolt := castFromSeat(t, g, me, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})

	g.WithWriteLock(func() { fillPool(me, 6) })
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{Targets: cardRefs(bolt), Strict: true}); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("activating with six mana: %v, want ErrInsufficientMana", err)
	}
	g.WithWriteLock(func() { fillPool(me, 1) })
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs), Strict: true}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting an opponent's spell: %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{Targets: cardRefs(bolt), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	oppBefore, thirdBefore := lifeOf(g, opp.ID), lifeOf(g, third.ID)
	// The ability resolves first: two re-target prompts, one sent at the
	// third seat and one left on the opponent.
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	for _, to := range []uuid.UUID{third.ID, opp.ID} {
		p := latestPickTarget(g, me.ID)
		if p == nil {
			t.Fatal("a copy was not offered new targets")
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: to}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("a third copy was offered new targets")
	}
	passPriorityAroundTable(t, g)

	if got := twinningStaffCopies(g); got != 2 {
		t.Errorf("%d copies, want 2 (one, plus the Staff's)", got)
	}
	if got := oppBefore - lifeOf(g, opp.ID); got != 6 {
		t.Errorf("the opponent lost %d, want 6 (the Bolt and one copy)", got)
	}
	if got := thirdBefore - lifeOf(g, third.ID); got != 3 {
		t.Errorf("the third seat lost %d, want 3 (the retargeted copy)", got)
	}
}

// The additional copy may get new targets even when the instruction's own
// copy may not.
func TestTwinningStaffsAdditionalCopyMayChooseNewTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	twinningStaffOn(g, me)
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	oppBefore, thirdBefore := lifeOf(g, opp), lifeOf(g, third)

	var err error
	g.WithWriteLock(func() { err = g.CopySpellForEffect(bolt, me, false, nil) })
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if got := twinningStaffCopies(g); got != 1 {
		t.Fatalf("%d copies before the prompt, want 1 (the instruction's own)", got)
	}
	p := latestPickTarget(g, me)
	if p == nil {
		t.Fatal("the additional copy was not offered new targets")
	}
	if err := g.ResolvePickTarget(p.ID, me, game.TargetRef{Kind: game.TargetPlayer, ID: third}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	if latestPickTarget(g, me) != nil {
		t.Fatal("the instruction's own copy was offered new targets")
	}
	passPriorityAroundTable(t, g)

	if got := oppBefore - lifeOf(g, opp); got != 6 {
		t.Errorf("the opponent lost %d, want 6 (the Bolt and the kept copy)", got)
	}
	if got := thirdBefore - lifeOf(g, third); got != 3 {
		t.Errorf("the third seat lost %d, want 3 (the additional copy)", got)
	}
}

// Casualty's "copy it" is one copy instruction: two copies with the Staff.
func TestTwinningStaffAddsACopyToCasualty(t *testing.T) {
	g := newCatalogGame(t)
	me, id := snc5Hand(t, g, "Light 'Em Up", "Sorcery", "{1}{R}", snc5LightEmUp)
	twinningStaffOn(g, me.ID)
	victim := pushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 6)
	fodder := snc5Sac(g, me.ID, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	snc5Drive(t, g, game.TargetRef{Kind: game.TargetCard, ID: victim}, false, false)

	if got := twinningStaffCopies(g); got != 2 {
		t.Errorf("%d copies, want 2 (casualty's one, plus one)", got)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("2 damage from the spell and 2 from each of its two copies should kill a 0/6")
	}
}

// Conspire's "copy it" is one copy instruction too.
func TestTwinningStaffAddsACopyToConspire(t *testing.T) {
	g, me, opp := rsTapTable(t)
	twinningStaffOn(g, me.ID)
	burn := handCardOf(g, me, "Burn at the Stake", "Sorcery", "{2}{R}{R}{R}", burnAtTheStakeOracle)
	team := taoCreatures(g, me.ID, 3)
	g.WithWriteLock(func() {
		fillPoolColored(me, "R", 3)
		fillPool(me, 2)
	})
	life := opp.Life
	if err := g.CastSpell(me.ID, burn, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		OptionalCosts: []int{0},
		TeamworkIDs:   team[:1],
		ConspireIDs:   team[1:],
		Strict:        true,
	}); err != nil {
		t.Fatalf("Burn at the Stake with conspire: %v", err)
	}
	if prompts := sqResolveAll(t, g, me); prompts != 2 {
		t.Errorf("%d re-target prompts, want 2 (one per copy)", prompts)
	}
	if got := twinningStaffCopies(g); got != 2 {
		t.Errorf("%d copies, want 2 (conspire's one, plus one)", got)
	}
	if lost := life - opp.Life; lost != 9 {
		t.Errorf("the opponent lost %d, want 9: three from the spell and three from each copy", lost)
	}
}

// "Copy that spell for each other Golem" is ONE instruction (CR 707.10e):
// two Golem copies, and one more that starts from the original's target
// and may be sent elsewhere.
func TestTwinningStaffAddsOneCopyToACopyForEachGolem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	twinningStaffOn(g, me.ID)
	pre := b12Push(g, opp.ID, "Precursor Golem", "Artifact Creature — Golem", rprB13PrecursorOracle, 3, 3)
	a := b12Creature(g, me.ID, "Golem A", "Artifact Creature — Golem", 3, 3)
	b := b12Creature(g, opp.ID, "Golem B", "Artifact Creature — Golem", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Giant Growth", "Instant", rprB13GiantGrowthOracle, cardTarget(a))
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no order prompt for the caster")
	}
	answerChooseCards(t, g, me.ID, b, pre)
	if got := twinningStaffCopies(g); got != 2 {
		t.Fatalf("%d copies before the additional one is placed, want 2", got)
	}
	pickCard(t, g, me.ID, bear)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a Golem's copy was offered new targets")
	}
	for i := 0; i < 4 && len(g.Stack.Cards) > 0; i++ {
		passPriorityAroundTable(t, g)
	}
	if got := twinningStaffCopies(g); got != 3 {
		t.Errorf("%d copies, want 3 (one per other Golem, plus one)", got)
	}
	for name, id := range map[string]uuid.UUID{"A": a, "B": b, "Precursor": pre} {
		if p := effectivePower(t, g, id); p != 6 {
			t.Errorf("Golem %s is %d power, want 6", name, p)
		}
	}
	if p := effectivePower(t, g, bear); p != 5 {
		t.Errorf("the Bear is %d power, want 5 (the additional copy)", p)
	}
}
