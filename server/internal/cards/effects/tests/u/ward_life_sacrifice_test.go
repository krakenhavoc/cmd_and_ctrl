package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- ward—pay N life -------------------------------------------------

// Declining counters the spell and costs nothing. The prompt goes to
// the caster and declares the life it would charge, so the bot's move
// list can price the accept branch.
func TestLifeWardDeclinedCountersTheSpell(t *testing.T) {
	g, me, opp, elemental := effectstest.WardTable(t, "Refraction Elemental", "Creature — Elemental", effectstest.RefractionElementalOracle)
	blade := effectstest.CastAtWardedCreature(t, g, opp, elemental)
	life := opp.Life

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if effectstest.ConfirmChoiceFor(g, me.ID) != nil {
		t.Error("the ward permanent's controller must not be asked to pay")
	}
	if prompt.LifeCost != 2 {
		t.Errorf("prompt LifeCost = %d, want 2", prompt.LifeCost)
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(elemental) {
		t.Error("a declined ward must counter the spell — the Elemental should live")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell should be in its owner's graveyard")
	}
	if opp.Life != life {
		t.Errorf("declining cost %d life, want 0", life-opp.Life)
	}
}

// Paying takes the life and the spell resolves.
func TestLifeWardPaidTakesTheLifeAndTheSpellResolves(t *testing.T) {
	g, _, opp, elemental := effectstest.WardTable(t, "Refraction Elemental", "Creature — Elemental", effectstest.RefractionElementalOracle)
	effectstest.CastAtWardedCreature(t, g, opp, elemental)
	life := opp.Life

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != life-2 {
		t.Errorf("paying the ward: life %d -> %d, want -2", life, opp.Life)
	}
	if g.Battlefield.Contains(elemental) {
		t.Error("the ward was paid — the removal spell should have resolved")
	}
}

// CR 119.4: a caster below the payment cannot pay, so the spell is
// countered and nobody is asked.
func TestLifeWardCountersWithoutAPromptWhenTheCasterCannotPay(t *testing.T) {
	g, _, opp, elemental := effectstest.WardTable(t, "Refraction Elemental", "Creature — Elemental", effectstest.RefractionElementalOracle)
	opp.Life = 1
	blade := effectstest.CastAtWardedCreature(t, g, opp, elemental)

	if prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID); prompt != nil {
		t.Fatalf("a caster at 1 life was offered a 2-life payment: %q", prompt.Reason)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(elemental) || !opp.Graveyard.Contains(blade) {
		t.Error("a ward the caster cannot pay must counter the spell")
	}
	if opp.Life != 1 {
		t.Errorf("life %d, want 1 — nothing was paid", opp.Life)
	}
}

// Sedgemoor Witch carries the same cost at 3.
func TestSedgemoorWitchWardAsksForThreeLife(t *testing.T) {
	g, _, opp, witch := effectstest.WardTable(t, "Sedgemoor Witch", "Creature — Human Warlock", effectstest.SedgemoorWitchOracle)
	effectstest.CastAtWardedCreature(t, g, opp, witch)

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if prompt.LifeCost != 3 {
		t.Errorf("prompt LifeCost = %d, want 3", prompt.LifeCost)
	}
}

// Magecraft: an instant its controller casts makes a Pest; an
// opponent's does not.
func TestSedgemoorWitchMakesAPestWhenYouCastAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Sedgemoor Witch", "Creature — Human Warlock", effectstest.SedgemoorWitchOracle, false)

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Pest"); n != 1 {
		t.Errorf("%d Pests after casting an instant, want 1", n)
	}
}

// --- ward—sacrifice a creature ------------------------------------

// Paying is a two-link chain: the confirm, then a pick over the
// CASTER's creatures only — never the Ripper's controller's. The pick
// is sacrificed, and the removal resolves. Both deaths feed the
// Ripper's drain, which is how the test sees them.
func TestSacrificeWardPaidSacrificesTheCastersCreature(t *testing.T) {
	g, me, opp, ripper := effectstest.WardTable(t, "Vein Ripper", "Creature — Vampire Assassin", effectstest.VeinRipperOracle)
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastAtWardedCreature(t, g, opp, ripper)
	myLife, oppLife := me.Life, opp.Life

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if prompt.LifeCost != 0 {
		t.Errorf("a sacrifice ward declared LifeCost %d", prompt.LifeCost)
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("accepting the ward queued no sacrifice pick")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != bear {
		t.Fatalf("candidates = %v, want only the caster's creature %s", pick.ChooseCards, bear)
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("pick bounds %d..%d, want exactly one", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Fatal("the chosen creature was not sacrificed")
	}
	effectstest.SettleDrainingAt(t, g, me.ID, opp.ID)

	if !opp.Graveyard.Contains(bear) {
		t.Error("the sacrificed creature should be in its owner's graveyard")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the Ripper's controller's own creature must be untouched")
	}
	if g.Battlefield.Contains(ripper) {
		t.Error("the ward was paid — the removal spell should have resolved")
	}
	// Two creatures died (the sacrifice, then the Ripper): two drains.
	if opp.Life != oppLife-4 || me.Life != myLife+4 {
		t.Errorf("two drains of 2: me %+d opp %+d, want +4 / -4", me.Life-myLife, opp.Life-oppLife)
	}
}

// Declining counters the spell and nothing is sacrificed.
func TestSacrificeWardDeclinedCountersTheSpell(t *testing.T) {
	g, _, opp, ripper := effectstest.WardTable(t, "Vein Ripper", "Creature — Vampire Assassin", effectstest.VeinRipperOracle)
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	blade := effectstest.CastAtWardedCreature(t, g, opp, ripper)

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) != nil {
		t.Error("declining must not ask for a sacrifice")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(ripper) || !opp.Graveyard.Contains(blade) {
		t.Error("a declined ward must counter the spell")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("declining sacrificed a creature")
	}
}

// TestSacrificeWardCountersWhenItsPickIsWithdrawn — #1225, and the
// third way a sacrifice ward can go unpaid. The caster accepted, and
// then their last creature left the battlefield while the "which
// creature?" prompt was open (removal in response, a state-based
// death, an admin move). No answer the resolver would take is left, so
// the prompt is withdrawn (#1045) — and the drop runs the pick's frame
// with nothing picked, which is the SAME branch the guard one line
// above takes for a caster who never had a creature: the cost is not
// paid, so the spell is countered.
//
// Before #1225 the drop ran nothing at all and the spell sat on the
// stack neither paid for nor countered — the ward silently did not
// happen.
func TestSacrificeWardCountersWhenItsPickIsWithdrawn(t *testing.T) {
	g, _, opp, ripper := effectstest.WardTable(t, "Vein Ripper", "Creature — Vampire Assassin", effectstest.VeinRipperOracle)
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	blade := effectstest.CastAtWardedCreature(t, g, opp, ripper)

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) == nil {
		t.Fatal("accepting the ward queued no sacrifice pick")
	}

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bear); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})

	if c := effectstest.ChooseCardsChoiceFor(g, opp.ID); c != nil {
		t.Fatalf("the pick survives a board with none of its candidates on it: %+v", c)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(ripper) || !opp.Graveyard.Contains(blade) {
		t.Error("a ward whose payment could not be completed must counter the spell")
	}
}

// A caster with no creature cannot pay: countered, no prompt. The
// Ripper's controller having creatures does not count.
func TestSacrificeWardCountersWithoutAPromptWhenTheCasterHasNoCreature(t *testing.T) {
	g, me, opp, ripper := effectstest.WardTable(t, "Vein Ripper", "Creature — Vampire Assassin", effectstest.VeinRipperOracle)
	effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	blade := effectstest.CastAtWardedCreature(t, g, opp, ripper)

	if prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID); prompt != nil {
		t.Fatalf("a caster with no creature was offered the sacrifice: %q", prompt.Reason)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(ripper) || !opp.Graveyard.Contains(blade) {
		t.Error("a ward the caster cannot pay must counter the spell")
	}
}

// --- construction ----------------------------------------------------

// A ward cost is one component or mana plus fixed life. Other mixtures fail at catalog
// init rather than charging whichever component the resolver checks
// first.
func TestWardCostRejectsUnsupportedMixtures(t *testing.T) {
	for name, cost := range map[string]effects.WardCost{
		"empty":              {},
		"negative life":      {Mana: "{2}", Life: -1},
		"mana and sacrifice": {Mana: "{1}", Sacrifice: &effects.WardSacrificeCost{Label: "a creature", OK: effects.Creature()}},
		"life and sac":       {Life: 2, Sacrifice: &effects.WardSacrificeCost{Label: "a creature", OK: effects.Creature()}},
		"nil predicate":      {Sacrifice: &effects.WardSacrificeCost{Label: "a creature"}},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: Ward accepted %+v", name, cost)
				}
			}()
			effects.Ward(cost, "")
		}()
	}
	for _, ok := range []effects.WardCost{effects.WardMana("{2}"), effects.WardManaAndLife("{2}", 2), effects.WardLife(3), effects.WardSacrifice("a creature", effects.Creature())} {
		effects.Ward(ok, "")
	}
}
