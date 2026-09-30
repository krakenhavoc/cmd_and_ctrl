package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tokens_297a_test.go — slice 297-a, the cards that make tokens (or
// copies of tokens) as they enter or as things happen around them.

const (
	hornOfGondorOracle         = "88777fc0-286d-47fc-8e94-1e289b4964c4"
	agateInstigatorOracle      = "e1a2bd31-b800-47cd-8866-1b849a601b09"
	lathlissOracle             = "15fef45d-f1a9-49b2-abaa-fe77bb9d1afd"
	miirymOracle               = "880fcf86-e6a9-482c-b91d-750f293127b2"
	marionetteApprenticeOracle = "726d9d2c-736a-4852-9938-a0f50d8fd89f"
	marionetteMasterOracle     = "dabbf796-3b88-499e-8839-06fa36fe01ac"
	tataruTaruOracle           = "70dd5013-f18f-4501-882f-70590c424e20"
	renewedSolidarityOracle    = "bea3ff6e-7649-4e51-b2ad-763f9ac2d4b8"
	weddingRingOracle          = "0c34e962-99d9-4163-b852-4f61886546aa"
)

// castFromHand casts `card` from the active seat's hand with the given
// announcement, after advancing to a main phase.
func castFromHand(t *testing.T, g *game.Game, card game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = active.ID, active.ID
	active.Hand.PushTop(card)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, card.InstanceID, params); err != nil {
		t.Fatalf("CastSpell %s: %v", card.Name, err)
	}
	return card.InstanceID
}

// castFlashCreatureFrom casts a creature with flash out of `seat`'s
// hand whoever's turn it is, so a test can have an opponent's creature
// enter during the active player's turn.
func castFlashCreatureFrom(t *testing.T, g *game.Game, seat *game.Player, card game.Card) uuid.UUID {
	t.Helper()
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = seat.ID, seat.ID
	card.Keywords = append(card.Keywords, "flash")
	seat.Hand.PushTop(card)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(seat.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s from seat %s: %v", card.Name, seat.ID, err)
	}
	return card.InstanceID
}

// cardsNamed counts battlefield cards with a name under a controller.
func cardsNamed(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

// --- Horn of Gondor -----------------------------------------------

func TestHornOfGondorMakesASoldierThenOneTokenPerHuman(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	horn := castFromHand(t, g, game.Card{Name: "Horn of Gondor", TypeLine: "Legendary Artifact", OracleID: hornOfGondorOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Human Soldier", me.ID); n != 1 {
		t.Fatalf("%d Human Soldier tokens after the Horn entered, want 1", n)
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Town Guard", TypeLine: "Creature — Human Cleric",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Not A Human", TypeLine: "Creature — Elf",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
		t.Fatalf("fund: %v", err)
	}

	if err := g.ActivateCatalogAbility(me.ID, horn, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Two Humans (the Soldier and the Cleric): two new tokens.
	if n := cardsNamed(g, "Human Soldier", me.ID); n != 3 {
		t.Errorf("%d Human Soldier tokens, want 3", n)
	}
}

// --- Agate Instigator ---------------------------------------------

func TestAgateInstigatorPingsForEachCreatureAndOffspringMakesAPinger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castFromHand(t, g, game.Card{
		Name: "Agate Instigator", TypeLine: "Creature — Lizard Rogue", OracleID: agateInstigatorOracle,
		ManaCost: "{1}{R}", Power: 2, Toughness: 2,
	}, game.CastSpellParams{OptionalCosts: []int{0}})
	passPriorityAroundTable(t, g)

	if n := cardsNamed(g, "Agate Instigator", me.ID); n != 2 {
		t.Fatalf("%d Agate Instigators, want the card and its offspring token", n)
	}
	// The token entering is another creature you control: the original
	// pinged once, and the token's own trigger was not yet on the board.
	if opp.Life != 39 {
		t.Errorf("an opponent is at %d, want 39 (one ping for the offspring token)", opp.Life)
	}

	castFromHand(t, g, game.Card{Name: "Goblin Scout", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != 37 {
		t.Errorf("an opponent is at %d after another creature entered, want 37 (both Instigators ping)", opp.Life)
	}
}

func TestAgateInstigatorUnpaidMakesNoToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castFromHand(t, g, game.Card{
		Name: "Agate Instigator", TypeLine: "Creature — Lizard Rogue", OracleID: agateInstigatorOracle,
		ManaCost: "{1}{R}", Power: 2, Toughness: 2,
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Agate Instigator", me.ID); n != 1 {
		t.Errorf("%d Agate Instigators without paying offspring, want 1", n)
	}
}

// --- Lathliss, Dragon Queen ---------------------------------------

func TestLathlissMakesADragonForEachOtherNontokenDragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lathliss, Dragon Queen", TypeLine: "Legendary Creature — Dragon",
		OracleID: lathlissOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})

	castFromHand(t, g, game.Card{Name: "Test Dragon", TypeLine: "Creature — Dragon", Power: 3, Toughness: 3}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Dragon", me.ID); n != 1 {
		t.Fatalf("%d Dragon tokens after a Dragon entered, want 1", n)
	}
	// The token is a Dragon too, but a token: no further trigger.
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Dragon", me.ID); n != 1 {
		t.Errorf("%d Dragon tokens, the token triggered Lathliss", n)
	}
	// An opponent's Dragon, and a creature of mine that is not a Dragon,
	// make nothing.
	castFlashCreatureFrom(t, g, opp, game.Card{Name: "Their Dragon", TypeLine: "Creature — Dragon", Power: 3, Toughness: 3})
	passPriorityAroundTable(t, g)
	castFromHand(t, g, game.Card{Name: "Plain Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Dragon", me.ID); n != 1 {
		t.Errorf("%d Dragon tokens of mine, want 1", n)
	}
	if n := cardsNamed(g, "Dragon", opp.ID); n != 0 {
		t.Errorf("%d Dragon tokens for the opponent, want 0", n)
	}
}

func TestLathlissPumpsDragonsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lathliss := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lathliss, Dragon Queen", TypeLine: "Legendary Creature — Dragon",
		OracleID: lathlissOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Dragon", TypeLine: "Creature — Dragon", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Dragon", TypeLine: "Creature — Dragon", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	bear := auraBear(g, me.ID)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}"); err != nil {
		t.Fatalf("fund: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, lathliss, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	for name, want := range map[string]struct {
		id    uuid.UUID
		power int
	}{"Lathliss": {lathliss, 7}, "my Dragon": {mine, 3}, "their Dragon": {theirs, 2}, "a Bear": {bear, 2}} {
		if got := effectivePower(t, g, want.id); got != want.power {
			t.Errorf("%s has power %d, want %d", name, got, want.power)
		}
	}
}

// --- Miirym, Sentinel Wyrm ----------------------------------------

func TestMiirymCopiesAnEnteringNontokenDragonAsANonlegendaryToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Miirym, Sentinel Wyrm", TypeLine: "Legendary Creature — Dragon Spirit",
		OracleID: miirymOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	castFromHand(t, g, game.Card{
		Name: "Big Wyrm", TypeLine: "Legendary Creature — Dragon", Power: 4, Toughness: 4,
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)

	var copies []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Big Wyrm" {
			copies = append(copies, c)
		}
	}
	if len(copies) != 2 {
		t.Fatalf("%d Big Wyrms on the battlefield, want the card and one token copy", len(copies))
	}
	for _, c := range copies {
		if !c.IsToken() {
			if !c.IsLegendary() {
				t.Error("the Dragon that was cast lost its legendary supertype")
			}
			continue
		}
		if c.IsLegendary() {
			t.Errorf("the token copy is legendary: %q", c.TypeLine)
		}
		if c.Power != 4 || c.Toughness != 4 {
			t.Errorf("the token copy is %d/%d, want 4/4", c.Power, c.Toughness)
		}
	}
	// Tokens do not trigger Miirym, so the count is stable.
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Big Wyrm", me.ID); n != 2 {
		t.Errorf("%d Big Wyrms after settling, want 2", n)
	}
}

func TestMiirymIgnoresAnOpponentsDragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Miirym, Sentinel Wyrm", TypeLine: "Legendary Creature — Dragon Spirit",
		OracleID: miirymOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	castFlashCreatureFrom(t, g, opp, game.Card{Name: "Their Dragon", TypeLine: "Creature — Dragon", Power: 3, Toughness: 3})
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Their Dragon", me.ID) + cardsNamed(g, "Their Dragon", opp.ID); n != 1 {
		t.Errorf("%d copies of the opponent's Dragon, want only the card", n)
	}
}

// --- Marionette Apprentice ----------------------------------------

func castMarionetteApprentice(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castFromHand(t, g, game.Card{
		Name: "Marionette Apprentice", TypeLine: "Creature — Human Artificer", OracleID: marionetteApprenticeOracle,
		ManaCost: "{1}{B}", Power: 1, Toughness: 2,
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	return id
}

func TestMarionetteApprenticeFabricatesACounterOrAServo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	apprentice := castMarionetteApprentice(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, apprentice, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters after choosing the counter, want 1", n)
	}
	if n := cardsNamed(g, "Servo", me.ID); n != 0 {
		t.Errorf("%d Servos after choosing the counter", n)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	apprentice2 := castMarionetteApprentice(t, g2)
	answerMayChoice(t, g2, me2.ID, false)
	passPriorityAroundTable(t, g2)
	if n := countersOn(g2, apprentice2, game.CounterPlusOne); n != 0 {
		t.Errorf("%d +1/+1 counters when a Servo was made instead, want 0", n)
	}
	if n := cardsNamed(g2, "Servo", me2.ID); n != 1 {
		t.Errorf("%d Servos after declining the counter, want 1", n)
	}
}

func TestMarionetteApprenticeDrainsWhenYourCreaturesAndArtifactsDie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	apprentice := castMarionetteApprentice(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	servos := battlefieldIDsNamed(g, "Servo")
	if len(servos) != 1 {
		t.Fatalf("%d Servos, want 1", len(servos))
	}
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	before := opp.Life

	killCreature(t, g, me.ID, servos[0])
	passPriorityAroundTable(t, g)
	killCreature(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("an opponent is at %d, want %d (a Servo and an artifact died)", opp.Life, before-2)
	}

	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("an opponent's creature dying drained: %d", opp.Life)
	}
	killCreature(t, g, me.ID, apprentice)
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("the Apprentice itself dying drained: %d (it says \"another\")", opp.Life)
	}
}

// --- Marionette Master --------------------------------------------

func castMarionetteMaster(t *testing.T, g *game.Game, counters bool) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := castFromHand(t, g, game.Card{
		Name: "Marionette Master", TypeLine: "Creature — Human Artificer", OracleID: marionetteMasterOracle,
		ManaCost: "{4}{B}{B}", Power: 1, Toughness: 1,
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, counters)
	passPriorityAroundTable(t, g)
	return id
}

func pickPlayerTarget(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: target}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

func TestMarionetteMasterDrainsForItsPowerWhenAnArtifactDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castMarionetteMaster(t, g, false)
	servos := battlefieldIDsNamed(g, "Servo")
	if len(servos) != 3 {
		t.Fatalf("%d Servos, want 3", len(servos))
	}
	before := opp.Life

	killCreature(t, g, me.ID, servos[0])
	passPriorityAroundTable(t, g)
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	// The Master is a 1/1 (the Servos were chosen over the counters).
	if opp.Life != before-1 {
		t.Errorf("the target opponent is at %d, want %d", opp.Life, before-1)
	}
}

// CR 608.2h: the power is last-known information. The Master dies with
// the trigger still on the stack and the drain is for the 4/4 it was.
func TestMarionetteMasterDrainsForItsLastKnownPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := castMarionetteMaster(t, g, true) // three +1/+1 counters: a 4/4
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	before := opp.Life

	killCreature(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	pickPlayerTarget(t, g, me.ID, opp.ID)
	if triggerOnStack(g, master) == nil {
		t.Fatal("the drain trigger is not on the stack")
	}
	killCreature(t, g, me.ID, master)
	passPriorityAroundTable(t, g)

	if opp.Life != before-4 {
		t.Errorf("the target opponent is at %d, want %d (the Master's last-known power)", opp.Life, before-4)
	}
}

// --- Tataru Taru --------------------------------------------------

func castTataru(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castFromHand(t, g, game.Card{
		Name: "Tataru Taru", TypeLine: "Legendary Creature — Dwarf Advisor", OracleID: tataruTaruOracle,
		ManaCost: "{1}{W}", Power: 1, Toughness: 2,
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	return id
}

func TestTataruTaruDrawsAndLetsAnOpponentDrawForATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	auraSeedLibrary(me, 5)
	auraSeedLibrary(opp, 5)
	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()

	castTataru(t, g)
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, opp.ID, true)
	passPriorityAroundTable(t, g)

	// The cast moved Tataru out of my hand: net effect is one drawn card.
	if got := me.Hand.Size() - myHand; got != 1 {
		t.Errorf("I drew %d cards, want 1", got)
	}
	if got := opp.Hand.Size() - theirHand; got != 1 {
		t.Errorf("the opponent drew %d cards, want 1", got)
	}
	// The opponent drew on MY turn: Scions' Secretary makes a tapped Treasure.
	treasures := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == me.ID {
			treasures++
			if !c.Tapped {
				t.Error("the Treasure entered untapped")
			}
		}
	}
	if treasures != 1 {
		t.Errorf("%d Treasures, want 1", treasures)
	}

	// Only once each turn.
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasures after a second off-turn draw, want still 1", n)
	}
}

func TestTataruTaruOpponentMayDeclineAndOwnTurnDrawsPayNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	auraSeedLibrary(me, 5)
	auraSeedLibrary(opp, 8)
	theirHand := opp.Hand.Size()

	castTataru(t, g)
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - theirHand; got != 0 {
		t.Errorf("the opponent drew %d after declining", got)
	}

	// Their own draw step is their turn: no Treasure.
	advanceToDrawStepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Treasure", me.ID); n != 0 {
		t.Errorf("%d Treasures from a draw in the drawer's own turn, want 0", n)
	}
}

func TestTataruTaruAnyOpponentsOffTurnDrawMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, third := g.Seats[0], g.Seats[2]
	auraSeedLibrary(me, 5)
	auraSeedLibrary(third, 5)
	castTataru(t, g)
	// Nobody is targeted or asked here: decline the enter trigger's draw.
	pickPlayerTarget(t, g, me.ID, g.Seats[1].ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, g.Seats[1].ID, false)
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.DrawNForEffect(third.ID, 1) })
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasures after another opponent drew off-turn, want 1", n)
	}
}

// --- Renewed Solidarity -------------------------------------------

func TestRenewedSolidarityPumpsAndCopiesNewTokensOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushNamedTribePermanent(t, g, me.ID, "Renewed Solidarity", "Enchantment", renewedSolidarityOracle, "Goblin")
	goblin := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Card", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := auraBear(g, me.ID)
	if p := effectivePower(t, g, goblin); p != 2 {
		t.Errorf("a Goblin has power %d, want 2", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a Bear has power %d, want 2", p)
	}

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 red Goblin"), 2)
		_ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Human Soldier"), 1)
	})
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	if n := cardsNamed(g, "Goblin", me.ID); n != 4 {
		t.Errorf("%d Goblin tokens, want the two new ones doubled to 4", n)
	}
	if n := cardsNamed(g, "Human Soldier", me.ID); n != 1 {
		t.Errorf("%d Human Soldier tokens, want 1 (not the chosen type)", n)
	}
	if n := cardsNamed(g, "Goblin Card", me.ID); n != 1 {
		t.Errorf("a nontoken Goblin was copied (%d)", n)
	}
}

// --- Wedding Ring -------------------------------------------------

func castWeddingRing(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	id := castFromHand(t, g, game.Card{
		Name: "Wedding Ring", TypeLine: "Artifact", OracleID: weddingRingOracle, ManaCost: "{2}{W}{W}",
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	return id
}

func TestWeddingRingGivesTheChosenOpponentATokenCopyAndOnlyOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castWeddingRing(t, g)

	if n := cardsNamed(g, "Wedding Ring", me.ID); n != 1 {
		t.Errorf("%d Rings under my control, want 1", n)
	}
	var tokens int
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Wedding Ring" && c.Controller == opp.ID {
			tokens++
			if !c.IsToken() {
				t.Error("the opponent's Ring is not a token")
			}
		}
	}
	if tokens != 1 {
		t.Errorf("the opponent has %d Rings, want one token copy (the copy was not cast)", tokens)
	}
}

func TestWeddingRingDrawsAndGainsLifeWhenTheHolderDoesOnTheirTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	auraSeedLibrary(me, 5)
	auraSeedLibrary(opp, 5)
	castWeddingRing(t, g)

	// On my turn, my own draws and gains are my own business, and the
	// opponent's ring-holder drawing now is not on their turn.
	myHand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != myHand {
		t.Errorf("I drew because an opponent drew on MY turn (hand %d -> %d)", myHand, me.Hand.Size())
	}

	// Their draw step is their turn, and they hold a Ring: I draw.
	advanceToDrawStepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - myHand; got != 1 {
		t.Errorf("I drew %d cards off the opponent's draw-step draw, want 1", got)
	}

	life := me.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3) })
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("my life is %d, want %d (I gain as much as they did on their turn)", me.Life, life+3)
	}
}

func TestRemoveLegendaryFromTypeLineHandlesATokenTemplate(t *testing.T) {
	for in, want := range map[string]string{
		"Token Legendary Creature — Dragon":         "Token Creature — Dragon",
		"Legendary Creature — Dragon":               "Creature — Dragon",
		"Token Creature — Dragon":                   "Token Creature — Dragon",
		"Token Legendary Artifact Creature — Golem": "Token Artifact Creature — Golem",
	} {
		if got := removeLegendaryFromTypeLine(in); got != want {
			t.Errorf("removeLegendaryFromTypeLine(%q) = %q, want %q", in, got, want)
		}
	}
}
