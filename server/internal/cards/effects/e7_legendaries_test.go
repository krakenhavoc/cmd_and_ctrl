package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// e7_legendaries_test.go — issue #1117, the five hardest cards in the
// Betor life-swap deck. One focused test per card, each asserting the
// clause that could plausibly have been got wrong rather than merely
// that the card registers.

const (
	dragonlordDromokaOracle = "82cdd612-3322-4372-810f-1ff106ea8e6a"
	selvalaHeartOracle      = "1d725121-e50c-42f0-9128-56802f07c89e"
	troubleInPairsOracle    = "f349f58b-8cc8-45e4-9565-2b46fdf976c9"
	gixPraetorOracle        = "928d977e-cff0-4e0e-83bb-16d73a754f35"
	valgavothOracle         = "cae3ec72-436d-4086-9dcb-17b3d92ad5c4"
)

// e7CastInstant announces a plain non-catalog instant from a seat's
// hand and hands back the announce error, so a test can assert either
// outcome. No targets and no catalog entry: the only thing that can
// refuse it is the CR 101.2 cast gate.
func e7CastInstant(g *game.Game, caster *game.Player) error {
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Test Instant", TypeLine: "Instant",
		Owner: caster.ID, Controller: caster.ID,
	})
	return g.CastSpell(caster.ID, id, game.CastSpellParams{})
}

// e7Main walks the cursor to a main phase.
func e7Main(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// --- Dragonlord Dromoka --------------------------------------------

// The lockout is asymmetric in two directions at once: it binds every
// seat but Dromoka's controller, and only while it is that
// controller's turn. Both are asserted, because a predicate that
// dropped either half would pass a one-sided test — forget the seat
// check and Dromoka stops her own controller casting, forget the turn
// check and she is a permanent silence.
func TestDragonlordDromokaLocksOpponentsOutOfHerControllersTurn(t *testing.T) {
	t.Run("an opponent cannot cast during her controller's turn", func(t *testing.T) {
		g := newCatalogGame(t)
		active, opp := g.Seats[0], g.Seats[1]
		seedEnchantment(g, dragonlordDromokaOracle, "Dragonlord Dromoka",
			"Legendary Creature — Elder Dragon", active.ID)
		e7Main(t, g)

		assertCantCast(t, e7CastInstant(g, opp), "during your turn")
	})

	t.Run("her own controller may still cast", func(t *testing.T) {
		g := newCatalogGame(t)
		active := g.Seats[0]
		seedEnchantment(g, dragonlordDromokaOracle, "Dragonlord Dromoka",
			"Legendary Creature — Elder Dragon", active.ID)
		e7Main(t, g)

		if err := e7CastInstant(g, active); err != nil {
			t.Errorf("Dromoka refused her own controller's spell: %v", err)
		}
	})

	t.Run("an opponent may cast on somebody else's turn", func(t *testing.T) {
		g := newCatalogGame(t)
		// Dromoka belongs to a seat that is NOT the active player, so
		// it is not "your turn" and nothing is forbidden.
		owner, caster := g.Seats[1], g.Seats[2]
		seedEnchantment(g, dragonlordDromokaOracle, "Dragonlord Dromoka",
			"Legendary Creature — Elder Dragon", owner.ID)
		e7Main(t, g)

		if err := e7CastInstant(g, caster); err != nil {
			t.Errorf("Dromoka locked out a turn that is not her controller's: %v", err)
		}
	})
}

// The counterspell RESOLVES and does nothing (CR 701.6a); Dromoka
// reaches the battlefield. The distinction from "illegal target" is
// observable — an illegal target would fizzle the counterspell.
func TestDragonlordDromokaCannotBeCountered(t *testing.T) {
	g := newCatalogGame(t)
	caster, opponent := g.Seats[0], g.Seats[1]

	dromokaID := castCatalogSpell(t, g, "Dragonlord Dromoka",
		"Legendary Creature — Elder Dragon", dragonlordDromokaOracle, nil)

	counterID := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: counterID, Name: "Counterspell", TypeLine: "Instant",
		OracleID: "cc187110-1148-4090-bbb8-e205694a39f5",
		Owner:    opponent.ID, Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, counterID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dromokaID}},
	}); err != nil {
		t.Fatalf("Counterspell on Dromoka was refused at announce: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !opponent.Graveyard.Contains(counterID) {
		t.Errorf("Counterspell did not resolve — it should have, and done nothing")
	}
	if !g.Battlefield.Contains(dromokaID) {
		t.Errorf("Dragonlord Dromoka was countered")
	}
	if _, ok := Lookup(dragonlordDromokaOracle); !ok {
		t.Fatal("Dromoka is not registered")
	}
	if spec, _ := Lookup(dragonlordDromokaOracle); len(spec.PrintedKeywords) != 2 {
		t.Errorf("printed keywords = %v, want flying and lifelink", spec.PrintedKeywords)
	}
	_ = caster
}

// --- Selvala, Heart of the Wilds -----------------------------------

// The whole point of the mana ability: X INDEPENDENT any-colour picks,
// not one pick minting X of a colour. Asserted twice — on the produced
// string, which is where the two shapes differ, and on the number of
// prompts a real activation opens, which is what a player experiences.
func TestSelvalaAddsOneAnyColorPickPerPointOfGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	selvala := pushCatalogPermanent(g, me.ID, "Selvala, Heart of the Wilds",
		"Legendary Creature — Elf Scout", selvalaHeartOracle, false)
	b16Creature(g, me.ID, "Big Friend", "Creature — Beast", 5, 5, "G")
	// An opponent's bigger creature must NOT count: "creatures YOU
	// control".
	b16Creature(g, g.Seats[1].ID, "Their Bigger Friend", "Creature — Beast", 9, 9, "G")

	spec, ok := Lookup(selvalaHeartOracle)
	if !ok || len(spec.ManaAbilities) != 1 {
		t.Fatalf("Selvala should register exactly one mana ability, got %+v", spec.ManaAbilities)
	}
	got := spec.ManaAbilities[0].ProducedFunc(g, me.ID, selvala)
	if want := AnyCombinationOfColors(5); got != want {
		t.Fatalf("produced = %q, want %q (five independent any-colour slots)", got, want)
	}
	if got == OneColorOfAmount(5) {
		t.Fatal("Selvala produced 'five mana of any ONE color' — that is Nykthos, not Selvala")
	}

	e7Main(t, g)
	b06AddMana(me, "G")
	if err := g.ActivateManaAbility(me.ID, selvala, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	picks := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			picks++
		}
	}
	if picks != 5 {
		t.Fatalf("%d colour picks, want 5 — one per point of greatest power", picks)
	}
	// Answering them differently is the "any COMBINATION" half.
	for _, color := range []string{"W", "U", "B", "R", "G"} {
		c := riderLatestManaPick(g, me.ID)
		if c == nil {
			t.Fatalf("ran out of picks before %s", color)
		}
		if err := g.ResolveManaChoice(c.ID, me.ID, color); err != nil {
			t.Fatalf("ResolveManaChoice(%s): %v", color, err)
		}
	}
	if got := len(batch01PoolColors(me)); got != 5 {
		t.Fatalf("pool holds %d mana, want 5", got)
	}
	seen := map[string]bool{}
	for _, c := range batch01PoolColors(me) {
		seen[c] = true
	}
	if len(seen) != 5 {
		t.Errorf("pool = %v, want five different colours", batch01PoolColors(me))
	}
}

// The draw belongs to the ENTERING creature's controller, not to
// Selvala's. Seat 1 owns Selvala; seat 0 plays the fatty and is the
// one asked and the one who draws.
func TestSelvalaOffersTheDrawToTheEnteringCreaturesController(t *testing.T) {
	g := newCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	pushDiesCreatureForTest(g, owner.ID, "Selvala, Heart of the Wilds",
		selvalaHeartOracle, "Legendary Creature — Elf Scout", 2, 3)

	casterBefore, ownerBefore := caster.Hand.Size(), owner.Hand.Size()
	e7CastCreature(t, g, caster, "Big Friend", 5, 5)
	passPriorityAroundTable(t, g)

	if confirmForTriggerPrompt(g, owner.ID) != nil {
		t.Error("Selvala's controller was asked — the 'may' belongs to the other seat")
	}
	prompt := confirmForTriggerPrompt(g, caster.ID)
	if prompt == nil {
		t.Fatal("the entering creature's controller was not asked")
	}
	answerLatestTriggerPrompt(t, g, caster.ID, true)
	passPriorityAroundTable(t, g)

	// The creature was seeded into hand and cast straight back out of
	// it, so the only net movement is the draw.
	if got := caster.Hand.Size() - casterBefore; got != 1 {
		t.Errorf("caster hand delta %d, want 1 (the draw)", got)
	}
	if got := owner.Hand.Size() - ownerBefore; got != 0 {
		t.Errorf("Selvala's controller drew %d cards, want 0", got)
	}
}

// A creature that ties the biggest is not "greater than each other
// creature's power" — no trigger at all.
func TestSelvalaDoesNotTriggerOnATie(t *testing.T) {
	g := newCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	pushDiesCreatureForTest(g, owner.ID, "Selvala, Heart of the Wilds",
		selvalaHeartOracle, "Legendary Creature — Elf Scout", 2, 3)
	b16Creature(g, owner.ID, "Incumbent", "Creature — Beast", 5, 5, "G")

	e7CastCreature(t, g, caster, "Equal Friend", 5, 5)
	passPriorityAroundTable(t, g)

	if confirmForTriggerPrompt(g, caster.ID) != nil {
		t.Error("a tie triggered Selvala — the power must be strictly greater")
	}
}

// e7CastCreature casts a plain creature with real P/T from the active
// seat's hand. castCatalogSpell leaves power at zero, which is exactly
// the number Selvala's comparison is about.
func e7CastCreature(t *testing.T, g *game.Game, caster *game.Player, name string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness,
		Owner: caster.ID, Controller: caster.ID,
	})
	e7Main(t, g)
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// confirmForTriggerPrompt returns the newest open "you may" trigger
// prompt owed by a seat, or nil.
func confirmForTriggerPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			out = c
		}
	}
	return out
}

// --- Trouble in Pairs ----------------------------------------------

// "Their SECOND card each turn" is exactly the second: the first draw
// is free and the third is free again.
func TestTroubleInPairsDrawsOnAnOpponentsSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	seedEnchantment(g, troubleInPairsOracle, "Trouble in Pairs", "Enchantment", me.ID)
	e7Main(t, g)

	// Seat 0 already drew for turn, so the tally starts at one and the
	// very next draw is their second.
	if got := g.TurnTallyFor(opp.ID).CardsDrawn; got != 1 {
		t.Fatalf("opponent's draw tally is %d at the main phase, want 1 (the turn-based draw)", got)
	}
	before := me.Hand.Size()
	e7Draw(t, g, opp) // their second
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("the second draw gave me %d cards, want 1", got)
	}

	e7Draw(t, g, opp) // their third
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("the third draw also triggered: total %d, want 1", got)
	}

	// My own draws are not an opponent's.
	e7Draw(t, g, me)
	e7Draw(t, g, me)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("my own draws triggered it: hand delta %d, want 3 (1 trigger + 2 draws)", got)
	}
}

// The same shape one tally over: the second SPELL, and not the first.
func TestTroubleInPairsDrawsOnAnOpponentsSecondSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	seedEnchantment(g, troubleInPairsOracle, "Trouble in Pairs", "Enchantment", me.ID)
	e7Main(t, g)
	before := me.Hand.Size()

	if err := e7CastInstant(g, opp); err != nil {
		t.Fatalf("first spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Fatalf("the FIRST spell drew me %d cards, want 0", got)
	}

	if err := e7CastInstant(g, opp); err != nil {
		t.Fatalf("second spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("the second spell drew me %d cards, want 1", got)
	}

	if err := e7CastInstant(g, opp); err != nil {
		t.Fatalf("third spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("the third spell triggered too: total %d, want 1", got)
	}
}

// "Attacks YOU with two or more creatures" — one attacker is not
// enough, and a batch of three is still one card.
func TestTroubleInPairsDrawsOnceForATwoCreatureAttackAtYou(t *testing.T) {
	t.Run("one attacker draws nothing", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[1], g.Seats[0]
		seedEnchantment(g, troubleInPairsOracle, "Trouble in Pairs", "Enchantment", me.ID)
		lone := pushVanillaCreature(g, opp.ID, "Lone Bear", 2, 2)
		before := me.Hand.Size()

		declareAttack(t, g, me.ID, lone)
		passPriorityAroundTable(t, g)

		if got := me.Hand.Size() - before; got != 0 {
			t.Errorf("a one-creature attack drew %d cards, want 0", got)
		}
	})

	t.Run("three attackers draw exactly one", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[1], g.Seats[0]
		seedEnchantment(g, troubleInPairsOracle, "Trouble in Pairs", "Enchantment", me.ID)
		a := pushVanillaCreature(g, opp.ID, "Bear A", 2, 2)
		b := pushVanillaCreature(g, opp.ID, "Bear B", 2, 2)
		c := pushVanillaCreature(g, opp.ID, "Bear C", 2, 2)
		before := me.Hand.Size()

		declareAttack(t, g, me.ID, a, b, c)
		passPriorityAroundTable(t, g)

		if got := me.Hand.Size() - before; got != 1 {
			t.Errorf("a three-creature attack drew %d cards, want exactly 1", got)
		}
	})

	t.Run("two creatures attacking somebody else draw nothing", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp, other := g.Seats[1], g.Seats[0], g.Seats[2]
		seedEnchantment(g, troubleInPairsOracle, "Trouble in Pairs", "Enchantment", me.ID)
		a := pushVanillaCreature(g, opp.ID, "Bear A", 2, 2)
		b := pushVanillaCreature(g, opp.ID, "Bear B", 2, 2)
		before := me.Hand.Size()

		declareAttack(t, g, other.ID, a, b)
		passPriorityAroundTable(t, g)

		if got := me.Hand.Size() - before; got != 0 {
			t.Errorf("an attack at somebody else drew %d cards, want 0 — it says 'attacks YOU'", got)
		}
	})
}

// e7Draw draws one card for a player outside their own draw step.
func e7Draw(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	if err := g.DrawCard(p.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
}

// --- Gix, Yawgmoth Praetor -----------------------------------------

// The prompt and the payment belong to the DAMAGING creature's
// controller, who here is neither Gix's controller nor the damaged
// player — the political shape the card is played for.
func TestGixAsksTheDamagingCreaturesControllerAndChargesOneLife(t *testing.T) {
	g := newCatalogGame(t)
	attacker, gixOwner, victim := g.Seats[0], g.Seats[1], g.Seats[2]
	pushDiesCreatureForTest(g, gixOwner.ID, "Gix, Yawgmoth Praetor",
		gixPraetorOracle, "Legendary Creature — Phyrexian Praetor", 3, 3)
	bear := pushVanillaCreature(g, attacker.ID, "Bear", 2, 2)

	handBefore, lifeBefore := attacker.Hand.Size(), attacker.Life
	gixHandBefore := gixOwner.Hand.Size()
	attackWith(t, g, victim.ID, bear)
	passPriorityAroundTable(t, g)

	prompt := confirmChoiceFor(g, attacker.ID)
	if prompt == nil {
		t.Fatal("the damaging creature's controller was not offered the payment")
	}
	if confirmChoiceFor(g, gixOwner.ID) != nil {
		t.Error("Gix's controller was asked — the 'may' belongs to the creature's controller")
	}
	if prompt.LifeCost != 1 {
		t.Errorf("prompt LifeCost = %d, want 1 (declared so a bot can price it)", prompt.LifeCost)
	}
	if err := g.ResolveConfirm(prompt.ID, attacker.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := attacker.Hand.Size() - handBefore; got != 1 {
		t.Errorf("the payer drew %d cards, want 1", got)
	}
	if got := lifeBefore - attacker.Life; got != 1 {
		t.Errorf("the payer paid %d life, want 1", got)
	}
	if got := gixOwner.Hand.Size() - gixHandBefore; got != 0 {
		t.Errorf("Gix's controller drew %d cards, want 0", got)
	}
}

// Declining costs nothing and draws nothing.
func TestGixDeclinedTakesNoLifeAndDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	attacker, gixOwner, victim := g.Seats[0], g.Seats[1], g.Seats[2]
	pushDiesCreatureForTest(g, gixOwner.ID, "Gix, Yawgmoth Praetor",
		gixPraetorOracle, "Legendary Creature — Phyrexian Praetor", 3, 3)
	bear := pushVanillaCreature(g, attacker.ID, "Bear", 2, 2)

	handBefore, lifeBefore := attacker.Hand.Size(), attacker.Life
	attackWith(t, g, victim.ID, bear)
	passPriorityAroundTable(t, g)

	prompt := confirmChoiceFor(g, attacker.ID)
	if prompt == nil {
		t.Fatal("no payment prompt")
	}
	if err := g.ResolveConfirm(prompt.ID, attacker.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)

	if attacker.Hand.Size() != handBefore || attacker.Life != lifeBefore {
		t.Errorf("declining cost %d life and drew %d cards, want 0 and 0",
			lifeBefore-attacker.Life, attacker.Hand.Size()-handBefore)
	}
}

// --- Valgavoth, Terror Eater ---------------------------------------

// The replacement asks TWO questions, and the second is the one that
// keeps Valgavoth from being stronger than printed: a creature you
// stole from an opponent dies to its OWNER's graveyard, which is an
// opponent's, but you controlled it and printed Valgavoth does not
// eat it.
func TestValgavothExilesOnlyCardsYouDidNotControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushDiesCreatureForTest(g, me.ID, "Valgavoth, Terror Eater",
		valgavothOracle, "Legendary Creature — Elder Demon", 9, 9)

	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	// Owned by the opponent, controlled by me — a stolen creature.
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID,
	})
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{theirs, stolen, mine} {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Errorf("DestroyPermanentForEffect: %v", err)
			}
		}
	})

	if !g.Exile.Contains(theirs) {
		t.Errorf("an opponent's own dying creature was not exiled")
	}
	if opp.Graveyard.Contains(theirs) {
		t.Errorf("an opponent's dying creature reached their graveyard")
	}
	if !opp.Graveyard.Contains(stolen) {
		t.Errorf("a creature YOU controlled was exiled — Valgavoth says 'a card you didn't control'")
	}
	if !me.Graveyard.Contains(mine) {
		t.Errorf("your own card went somewhere other than your graveyard")
	}
}

// Ward—Sacrifice three nonland permanents: three, not one. The pick
// bounds are the assertion the count generalisation exists for, and
// the sacrificed permanents land in EXILE rather than the graveyard,
// because Valgavoth's own replacement catches them on the way.
func TestValgavothWardDemandsThreeNonlandPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	valgavoth := pushCatalogPermanent(g, me.ID, "Valgavoth, Terror Eater",
		"Legendary Creature — Elder Demon", valgavothOracle, false)
	e7Main(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}

	a := b16Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2, "G")
	b := b16Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2, "G")
	c := b16Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2, "G")
	// A land is not a legal sacrifice for this cost.
	land := seedLand(g, opp.ID, "Forest", "Basic Land — Forest", "")

	castAtWardedCreature(t, g, opp, valgavoth)
	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("accepting the ward queued no sacrifice pick")
	}
	if pick.ChooseMin != 3 || pick.ChooseMax != 3 {
		t.Fatalf("pick bounds %d..%d, want exactly three", pick.ChooseMin, pick.ChooseMax)
	}
	for _, id := range pick.ChooseCards {
		if id == land {
			t.Errorf("a land was offered for 'three nonland permanents'")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{a, b, c}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b, c} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a chosen permanent was not sacrificed")
		}
		if !g.Exile.Contains(id) {
			t.Errorf("a sacrificed permanent did not land in exile — Valgavoth eats it too")
		}
	}
	// The ward was PAID, so the removal spell resolves and does its
	// job — which is the assertion that the payment was accepted.
	if g.Battlefield.Contains(valgavoth) {
		t.Errorf("the ward was paid, so the removal spell should have resolved and killed Valgavoth")
	}
	if !me.Graveyard.Contains(valgavoth) {
		t.Errorf("Valgavoth is your own card, so his own replacement must not exile him")
	}
}

// Fewer than three nonland permanents is not a payment at all: the
// spell is countered without ever offering a prompt (CR 118.4).
func TestValgavothWardCountersWhenTheCasterCannotPayAllThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	valgavoth := pushCatalogPermanent(g, me.ID, "Valgavoth, Terror Eater",
		"Legendary Creature — Elder Demon", valgavothOracle, false)
	e7Main(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	b16Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2, "G")
	b16Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2, "G")

	blade := castAtWardedCreature(t, g, opp, valgavoth)
	if prompt := passUntilConfirmFor(t, g, opp.ID); prompt != nil {
		t.Fatalf("a caster with two permanents was offered a three-permanent payment: %q", prompt.Reason)
	}
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(valgavoth) {
		t.Error("a ward the caster cannot pay must counter the spell")
	}
	if !g.Exile.Contains(blade) && !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell went nowhere")
	}
}

// Every ward that shipped before the count existed still charges
// exactly one — the zero value is one.
func TestWardSacrificeCountDefaultsToOne(t *testing.T) {
	if got := (WardSacrificeCost{}).count(); got != 1 {
		t.Errorf("zero Count reads as %d, want 1", got)
	}
	if got := WardSacrifice("a creature", Creature()).Sacrifice.count(); got != 1 {
		t.Errorf("WardSacrifice charges %d, want 1", got)
	}
	if got := WardSacrificeN(3, "three nonland permanents", Nonland()).Sacrifice.count(); got != 3 {
		t.Errorf("WardSacrificeN(3) charges %d, want 3", got)
	}
}
