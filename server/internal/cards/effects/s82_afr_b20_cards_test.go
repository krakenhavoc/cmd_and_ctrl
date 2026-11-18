package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b20_cards_test.go — Adventures in the Forgotten Realms, batch 20:
// complex legends (#5499).

const (
	afrB20OrcusOracle      = "9d02b5e1-bd29-45ad-b11d-f2cd21c48979"
	afrB20FroghemothOracle = "66e66543-1f56-4e62-9183-3804843db101"
	afrB20CubeOracle       = "3c2e2e69-cd44-458f-a5b0-8591baa24e65"
	afrB20KarazikarOracle  = "47fde936-be71-4550-826b-b98d139cc57a"
	afrB20LorcanOracle     = "6c87c261-00a8-46f2-92c8-42009a0a2faf"
	afrB20NihiloorOracle   = "148271de-3351-41a1-bcf5-3ecaadc51540"
	afrB20KrydleOracle     = "77212081-0b49-4d92-9c5a-1f21e6907fce"
	afrB20WulfgarOracle    = "b19d1459-eb08-4e75-95f4-402d86371a78"
)

func TestAfrB20EveryBuiltCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		afrB20FroghemothOracle: "Froghemoth",
		afrB20CubeOracle:       "Gelatinous Cube",
		afrB20KarazikarOracle:  "Karazikar, the Eye Tyrant",
		afrB20LorcanOracle:     "Lorcan, Warlock Collector",
		afrB20NihiloorOracle:   "Nihiloor",
		afrB20KrydleOracle:     "Krydle of Baldur's Gate",
		afrB20WulfgarOracle:    "Wulfgar of Icewind Dale",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Gelatinous Cube -----------------------------------------------------

func TestAfrB20CubeExilesNonOozeUntilItLeavesAndDissolvesForX(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ooze := b12Creature(g, opp.ID, "Their Ooze", "Creature — Ooze", 1, 1)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	cube := castHoldCreature(t, g, "Gelatinous Cube", "Creature — Ooze", afrB20CubeOracle, 4, 3)
	if opts := pickTargetPrompt(t, g).PickTargetCards; len(opts) != 1 || opts[0] != bear {
		t.Fatalf("Cube's target options = %v, want only the non-Ooze Bear", opts)
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Cube did not exile the Bear")
	}
	if !g.Battlefield.Contains(ooze) {
		t.Error("the Ooze was exiled")
	}
	mkmAC2AddMana(t, g, me, "{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, cube, 0, game.ActivateAbilityParams{XValue: 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("a mana value 2 card was a legal target for X = 1")
	}
	p7Activate(t, g, me, cube, 0, game.ActivateAbilityParams{XValue: 2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if !opp.Graveyard.Contains(bear) {
		t.Error("the Bear is not in its owner's graveyard")
	}
	cmdr2S32Sacrifice(t, g, cube)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear returned from the graveyard")
	}
}

func TestAfrB20CubeReturnsWhatItHoldsWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	cube := castHoldCreature(t, g, "Gelatinous Cube", "Creature — Ooze", afrB20CubeOracle, 4, 3)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	cmdr2S32Sacrifice(t, g, cube)
	if got := tmtB1Count(g, opp.ID, "Their Bear"); got != 1 {
		t.Errorf("%d Bears on the battlefield after the Cube left, want 1", got)
	}
}

// --- Wulfgar of Icewind Dale ---------------------------------------------

func TestAfrB20WulfgarDoublesYourCreaturesAttackTriggersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Wulfgar of Icewind Dale", "Legendary Creature — Human Barbarian", afrB20WulfgarOracle, false)
	pushCatalogPermanent(g, me.ID, "Adeline, Resplendent Cathar", "Legendary Creature — Human Knight", "38515f89-348b-4cf3-b7bd-1f6fe4ce2fba", false)
	attacker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Attacker", TypeLine: "Creature — Soldier", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: attacker, Actor: me.ID, Target: opp.ID})
	})
	passPriorityAroundTable(t, g)
	if got := countTokensControlled(g, me.ID, "Human"); got != 6 {
		t.Errorf("your Adeline tokens with Wulfgar = %d, want 6", got)
	}

	// An opponent's attacking creature does not qualify, even for their own watcher.
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Wulfgar of Icewind Dale", "Legendary Creature — Human Barbarian", afrB20WulfgarOracle, false)
	pushCatalogPermanent(g, me.ID, "Adeline, Resplendent Cathar", "Legendary Creature — Human Knight", "38515f89-348b-4cf3-b7bd-1f6fe4ce2fba", false)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Attacker", TypeLine: "Creature — Soldier", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: theirs, Actor: opp.ID, Target: me.ID})
	})
	passPriorityAroundTable(t, g)
	if got := countTokensControlled(g, me.ID, "Human"); got != 0 {
		t.Errorf("Adeline tokens for an opponent's attack = %d, want 0", got)
	}
}

func TestAfrB20WulfgarMeleeTriggersTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wulfgar := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Wulfgar of Icewind Dale",
		TypeLine: "Legendary Creature — Human Barbarian", OracleID: afrB20WulfgarOracle, Power: 4, Toughness: 4,
		Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wulfgar, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, wulfgar); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("Wulfgar is %d/%d, want 6/6 (melee triggers twice for one opponent attacked)", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Krydle of Baldur's Gate ---------------------------------------------

func TestAfrB20KrydleDrainsMillsGainsAndScrys(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	krydle := pushCatalogPermanent(g, me.ID, "Krydle of Baldur's Gate", "Legendary Creature — Human Elf Rogue", afrB20KrydleOracle, false)
	seedLibrary(me, "Mine One", "Mine Two")
	seedLibrary(opp, "Theirs One", "Theirs Two")
	myLife, theirLife, theirLibrary := me.Life, opp.Life, opp.Library.Size()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: krydle, Target: opp.ID, Amount: 1, Combat: true})
	})
	passPriorityAroundTable(t, g)
	if got := theirLife - opp.Life; got != 1 {
		t.Errorf("the damaged player lost %d life, want 1", got)
	}
	if got := theirLibrary - opp.Library.Size(); got != 1 {
		t.Errorf("the damaged player milled %d cards, want 1", got)
	}
	if got := me.Life - myLife; got != 1 {
		t.Errorf("I gained %d life, want 1", got)
	}
	if c := scryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 1 {
		t.Errorf("Krydle's controller did not scry 1: %+v", c)
	}
	if c := scryChoiceFor(g, opp.ID); c != nil {
		t.Error("the damaged player scried")
	}
}

func TestAfrB20KrydleDoesNotTriggerForNoncombatOrOthersDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Krydle of Baldur's Gate", "Legendary Creature — Human Elf Rogue", afrB20KrydleOracle, false)
	other := pushVanillaCreature(g, me.ID, "Other Bear", 2, 2)
	theirLife := opp.Life
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: other, Target: opp.ID, Amount: 2, Combat: true})
	})
	passPriorityAroundTable(t, g)
	if opp.Life != theirLife {
		t.Errorf("another creature's damage drained: %d -> %d", theirLife, opp.Life)
	}
}

func afrB20KrydleAttack(t *testing.T, pay bool) (g *game.Game, attacker uuid.UUID, me *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	toMain(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Krydle of Baldur's Gate", "Legendary Creature — Human Elf Rogue", afrB20KrydleOracle, false)
	attacker = pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	other := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(other, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	// one trigger for the whole declaration, however many attackers
	answerPickTarget(t, g, attacker)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a second Krydle trigger asked for a target")
	}
	mkmAC2AddMana(t, g, me, "{C}{C}")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, pay)
	return g, attacker, me
}

func TestAfrB20KrydlePayingTwoMakesTheTargetUnblockable(t *testing.T) {
	g, attacker, me := afrB20KrydleAttack(t, true)
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, attacker, game.CantBeBlocked)
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left in the pool, want the {2} spent", me.ManaPool.Total())
	}
}

func TestAfrB20KrydleDecliningToPayLeavesTheTargetBlockable(t *testing.T) {
	g, attacker, _ := afrB20KrydleAttack(t, false)
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, attacker, 0)
}

// --- Karazikar, the Eye Tyrant -------------------------------------------

func TestAfrB20KarazikarTapsAndGoadsACreatureOfTheAttackedPlayer(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp1 := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp2 := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Karazikar, the Eye Tyrant", "Legendary Creature — Beholder", afrB20KarazikarOracle, false)
	attacker := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	theirs := pushVanillaCreature(g, opp1.ID, "Their Bear", 2, 2)
	bystander := pushVanillaCreature(g, opp2.ID, "Bystander", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp1.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if opts := pickTargetPrompt(t, g).PickTargetCards; len(opts) != 1 || opts[0] != theirs {
		t.Fatalf("Karazikar's target options = %v, want only the attacked player's creature", opts)
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, theirs); !c.Tapped {
		t.Error("the targeted creature was not tapped")
	}
	if got := b33Goaded(t, g, theirs); got != me.ID {
		t.Errorf("the targeted creature is goaded by %v, want me", got)
	}
	if got := b33Goaded(t, g, bystander); got != uuid.Nil {
		t.Errorf("an uninvolved creature is goaded by %v", got)
	}
}

func afrB20OpponentAttacks(t *testing.T, defenderSeat int) (g *game.Game, me, attackerPlayer, defender *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	me, attackerPlayer = g.Seats[0], g.Seats[1]
	defender = g.Seats[defenderSeat]
	pushCatalogPermanent(g, me.ID, "Karazikar, the Eye Tyrant", "Legendary Creature — Beholder", afrB20KarazikarOracle, false)
	bear := pushVanillaCreature(g, attackerPlayer.ID, "Their Attacker", 2, 2)
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(bear, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	return g, me, attackerPlayer, defender
}

func TestAfrB20KarazikarDrawsAndDrainsWhenAnOpponentAttacksAnotherOpponent(t *testing.T) {
	g, me, attacker, defender := afrB20OpponentAttacks(t, 2)
	myLife, theirLife, defLife := me.Life, attacker.Life, defender.Life
	myHand, theirHand, defHand := me.Hand.Size(), attacker.Hand.Size(), defender.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := myLife - me.Life; got != 1 {
		t.Errorf("I lost %d life, want 1", got)
	}
	if got := theirLife - attacker.Life; got != 1 {
		t.Errorf("the attacking player lost %d life, want 1", got)
	}
	if got := me.Hand.Size() - myHand; got != 1 {
		t.Errorf("I drew %d cards, want 1", got)
	}
	if got := attacker.Hand.Size() - theirHand; got != 1 {
		t.Errorf("the attacking player drew %d cards, want 1", got)
	}
	if defender.Life != defLife || defender.Hand.Size() != defHand {
		t.Error("the attacked opponent was affected")
	}
}

func TestAfrB20KarazikarIgnoresAnOpponentAttackingMe(t *testing.T) {
	g, me, attacker, _ := afrB20OpponentAttacks(t, 0)
	myLife, theirLife := me.Life, attacker.Life
	passPriorityAroundTable(t, g)
	if me.Life != myLife || attacker.Life != theirLife {
		t.Errorf("an attack on Karazikar's controller drained: %d -> %d, %d -> %d", myLife, me.Life, theirLife, attacker.Life)
	}
}

// --- Froghemoth ----------------------------------------------------------

func TestAfrB20FroghemothExilesUpToThatManyAndCountsByType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frog := pushCatalogPermanent(g, me.ID, "Froghemoth", "Creature — Frog Horror", afrB20FroghemothOracle, false)
	c1 := pushToGraveyard(g, opp, game.Card{Name: "Dead Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	c2 := pushToGraveyard(g, opp, game.Card{Name: "Dead Wolf", TypeLine: "Creature — Wolf", ManaCost: "{2}{G}", Power: 3, Toughness: 2})
	s1 := pushToGraveyard(g, opp, game.Card{Name: "Old Spell", TypeLine: "Sorcery", ManaCost: "{1}{R}"})
	left := pushToGraveyard(g, opp, game.Card{Name: "Left Behind", TypeLine: "Creature — Rat", ManaCost: "{B}", Power: 1, Toughness: 1})
	mine := pushToGraveyard(g, me, game.Card{Name: "My Corpse", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 1, Toughness: 1})
	life := me.Life
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: frog, Target: opp.ID, Amount: 3, Combat: true})
	})
	p := pickTargetPrompt(t, g)
	if p.PickTargetMax != 3 {
		t.Errorf("the picker allows up to %d targets, want 3 (the damage dealt)", p.PickTargetMax)
	}
	if hasID(p.PickTargetCards, mine) {
		t.Error("the picker offered a card from my own graveyard")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: c1}, {Kind: game.TargetCard, ID: c2}, {Kind: game.TargetCard, ID: s1},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{c1, c2, s1} {
		if graveyardHas(opp, id) {
			t.Errorf("%s is still in their graveyard", id)
		}
	}
	if !graveyardHas(opp, left) {
		t.Error("an untargeted card left their graveyard")
	}
	if got := counterCountOn(g, frog, game.CounterPlusOne); got != 2 {
		t.Errorf("Froghemoth has %d +1/+1 counters, want 2 (two creature cards)", got)
	}
	if got := me.Life - life; got != 1 {
		t.Errorf("I gained %d life, want 1 (one noncreature card)", got)
	}
}

func TestAfrB20FroghemothTargetCountIsCappedByTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frog := pushCatalogPermanent(g, me.ID, "Froghemoth", "Creature — Frog Horror", afrB20FroghemothOracle, false)
	a := pushToGraveyard(g, opp, game.Card{Name: "A", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	b := pushToGraveyard(g, opp, game.Card{Name: "B", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: frog, Target: opp.ID, Amount: 1, Combat: true})
	})
	p := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	}); err == nil {
		t.Fatal("two targets were accepted for one damage")
	}
}

// --- Lorcan, Warlock Collector -------------------------------------------

func afrB20LorcanTheirBearDies(t *testing.T, myLife int) (g *game.Game, me, opp *game.Player, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Lorcan, Warlock Collector", "Legendary Creature — Devil", afrB20LorcanOracle, false)
	bear = pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	me.Life = myLife
	cmdr2S32Sacrifice(t, g, bear)
	return g, me, opp, bear
}

func TestAfrB20LorcanPaysLifeToTakeTheCreatureAsAWarlock(t *testing.T) {
	g, me, _, bear := afrB20LorcanTheirBearDies(t, 20)
	answerMayChoice(t, g, me.ID, true)
	if me.Life != 18 {
		t.Errorf("life is %d, want 18 (paid its mana value, 2)", me.Life)
	}
	c, ok := battlefieldCard(g, bear)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the Bear is not on the battlefield under my control: %+v %v", c, ok)
	}
	if !c.HasSubtype("Warlock") || !c.HasSubtype("Bear") {
		t.Errorf("the Bear's subtypes are %v, want Bear and Warlock", c.Effective().Subtypes)
	}
}

func TestAfrB20LorcanDecliningLeavesTheCardInTheGraveyard(t *testing.T) {
	g, me, opp, bear := afrB20LorcanTheirBearDies(t, 20)
	answerMayChoice(t, g, me.ID, false)
	if me.Life != 20 {
		t.Errorf("life is %d, want 20", me.Life)
	}
	if !graveyardHas(opp, bear) {
		t.Error("the Bear left their graveyard")
	}
}

func TestAfrB20LorcanCannotPayMoreLifeThanItHas(t *testing.T) {
	g, me, opp, bear := afrB20LorcanTheirBearDies(t, 1)
	if c := latestConfirmFor(g, me.ID); c != nil {
		answerMayChoice(t, g, me.ID, true)
	}
	if me.Life != 1 {
		t.Errorf("life is %d, want 1", me.Life)
	}
	if !graveyardHas(opp, bear) {
		t.Error("the Bear came back though its mana value exceeded my life")
	}
}

func TestAfrB20LorcanIgnoresYourOwnCreaturesDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Lorcan, Warlock Collector", "Legendary Creature — Devil", afrB20LorcanOracle, false)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	cmdr2S32Sacrifice(t, g, mine)
	if c := latestConfirmFor(g, me.ID); c != nil {
		t.Error("Lorcan asked about my own creature")
	}
}

func TestAfrB20LorcanExilesYourDyingWarlocks(t *testing.T) {
	g, me, _, bear := afrB20LorcanTheirBearDies(t, 20)
	answerMayChoice(t, g, me.ID, true)
	cmdr2S32Sacrifice(t, g, bear)
	if graveyardHas(g.Seats[1], bear) {
		t.Error("the Warlock went to the graveyard")
	}
	if z := g.FindCardZoneForEffect(bear); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the Warlock is in %v, want exile", z)
	}
	// a creature that is not a Warlock dies as usual.
	plain := pushVanillaCreature(g, me.ID, "Plain Bear", 2, 2)
	cmdr2S32Sacrifice(t, g, plain)
	if z := g.FindCardZoneForEffect(plain); z == nil || z.Kind != game.ZoneGraveyard {
		t.Errorf("a non-Warlock is in %v, want its graveyard", z)
	}
}

// --- Nihiloor ------------------------------------------------------------

func TestAfrB20NihiloorTapsForEachOpponentAndStealsAtMostThatPower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp1 := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Tapper", "Creature — Bear", 3, 3)
	big := b12Creature(g, opp1.ID, "Their Big", "Creature — Giant", 5, 5)
	small := b12Creature(g, opp1.ID, "Their Small", "Creature — Bear", 2, 2)
	nihiloor := castHoldCreature(t, g, "Nihiloor", "Legendary Creature — Horror", afrB20NihiloorOracle, 3, 5)
	for i := 0; i < 3; i++ {
		p := latestChooseCardsFor(g, me.ID)
		if p == nil {
			break
		}
		var picks []uuid.UUID
		if i == 0 {
			picks = []uuid.UUID{mine}
		}
		if err := g.ResolveChooseCards(p.ID, me.ID, picks); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	if c := tmtB1Card(t, g, mine); !c.Tapped {
		t.Fatal("the chosen creature was not tapped")
	}
	if opts := pickTargetPrompt(t, g).PickTargetCards; len(opts) != 1 || opts[0] != small {
		t.Fatalf("steal options = %v, want only the Small (power at most 3)", opts)
	}
	answerPickTarget(t, g, small)
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, small); c.Controller != me.ID {
		t.Fatal("I did not gain control of the Small")
	}
	if c := tmtB1Card(t, g, big); c.Controller != opp1.ID {
		t.Error("the Big changed hands")
	}

	// Attacking with a creature an opponent owns drains its owner.
	myLife, theirLife := me.Life, opp1.Life
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: small, Actor: me.ID, Target: g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)].ID})
	})
	passPriorityAroundTable(t, g)
	if got := me.Life - myLife; got != 2 {
		t.Errorf("I gained %d life, want 2", got)
	}
	if got := theirLife - opp1.Life; got != 2 {
		t.Errorf("the owner lost %d life, want 2", got)
	}
	// A creature I own does not.
	myLife = me.Life
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: mine, Actor: me.ID, Target: opp1.ID})
	})
	passPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Error("attacking with my own creature drained")
	}

	// Control lasts only while I control Nihiloor.
	cmdr2S32Sacrifice(t, g, nihiloor)
	if c := tmtB1Card(t, g, small); c.Controller != opp1.ID {
		t.Error("the Small did not return when Nihiloor left")
	}
}

func TestAfrB20NihiloorTappingNothingStealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp1 := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Creature(g, me.ID, "Tapper", "Creature — Bear", 3, 3)
	small := b12Creature(g, opp1.ID, "Their Small", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Nihiloor", "Legendary Creature — Horror", afrB20NihiloorOracle, 3, 5)
	for i := 0; i < 3; i++ {
		p := latestChooseCardsFor(g, me.ID)
		if p == nil {
			break
		}
		if err := g.ResolveChooseCards(p.ID, me.ID, nil); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, small); c.Controller != opp1.ID {
		t.Error("a creature was stolen though nothing was tapped")
	}
}
