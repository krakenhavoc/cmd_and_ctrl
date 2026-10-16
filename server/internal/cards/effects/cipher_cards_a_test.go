package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cipher_cards_a_test.go — the cipher cards of #3867's batch A (CR
// 702.99): each card's own text, and the whole cipher loop (encode,
// combat damage to a player, the free copy with new targets) for Call
// of the Nightwing and Hands of Binding.

const (
	ciaArcaneHeistOracle       = "839748e7-ccb4-421e-9626-b6d8be9390ab"
	ciaCallOfTheNightwingOracl = "919d21ad-3d1e-4f80-8005-64e3f8bf64d1"
	ciaHandsOfBindingOracle    = "9bc82078-2f3e-44ee-906d-707f3430d414"
	ciaHiddenStringsOracle     = "a50737d7-dd2d-40da-817c-f913f76b5011"
	ciaMentalVaporsOracle      = "e0151f37-790b-47f4-9928-cd8f32f68e4c"
	ciaMidnightRecoveryOracle  = "7904025c-7120-4362-abe9-a9f5fb64447f"
	ciaParanoidDelusionsOracle = "129cec2c-fc1b-4ce6-98f2-2123e4dbd1ae"
)

// ciaSeats is the active player and the next one, at the active
// player's main phase.
func ciaSeats(t *testing.T, g *game.Game) (*game.Player, *game.Player) {
	t.Helper()
	advanceToMain(t, g)
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
}

// ciaCipherPick is `chooser`'s open cipher question, or nil.
func ciaCipherPick(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	if c := latestOptionPickFor(g, chooser); c != nil && strings.Contains(c.Reason, "cipher") {
		return c
	}
	return nil
}

// ciaAnswerCipher answers the open cipher question: encode on
// `creature`, or decline when it is uuid.Nil.
func ciaAnswerCipher(t *testing.T, g *game.Game, chooser, creature uuid.UUID) {
	t.Helper()
	pick := ciaCipherPick(g, chooser)
	if pick == nil {
		t.Fatal("no cipher question is open")
	}
	idx := len(pick.PickOptions) - 1
	if creature != uuid.Nil {
		idx = -1
		for i, o := range pick.PickOptions {
			if len(o.Cards) == 1 && o.Cards[0] == creature {
				idx = i
			}
		}
		if idx < 0 {
			t.Fatalf("the creature is not offered: %+v", pick.PickOptions)
		}
	}
	answerOptionPick(t, g, chooser, idx)
	g.SettleResolution()
}

// ciaPermanentOf is the battlefield permanent named `name` controlled
// by `controller`, other than `not`.
func ciaPermanentOf(g *game.Game, controller uuid.UUID, name string, not ...uuid.UUID) uuid.UUID {
	var out uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller != controller || c.Name != name {
				continue
			}
			skip := false
			for _, n := range not {
				skip = skip || c.InstanceID == n
			}
			if !skip && out == uuid.Nil {
				out = c.InstanceID
			}
		}
	})
	return out
}

func ciaTapped(g *game.Game, id uuid.UUID) bool {
	tapped := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				tapped = c.Tapped
			}
		}
	})
	return tapped
}

func ciaLibrary(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
}

// Call of the Nightwing makes its Horror before cipher is asked, so the
// Horror can carry the card; the Horror's combat damage offers a copy,
// and the copy makes a second Horror.
func TestCiaCallOfTheNightwingEncodesOnItsOwnHorror(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ciaSeats(t, g)
	id := castCatalogSpell(t, g, "Call of the Nightwing", "Sorcery", ciaCallOfTheNightwingOracl, nil)
	if cipherResolveToPrompt(t, g, me.ID) == nil {
		t.Fatal("the Horror is a creature you control, so cipher asks")
	}
	horror := ciaPermanentOf(g, me.ID, "Horror")
	if horror == uuid.Nil {
		t.Fatal("a Horror token is created before the cipher question")
	}
	var c game.Card
	g.ReadSnapshot(func() { c, _ = g.LookupCardForEffect(horror) })
	if c.Power != 1 || c.Toughness != 1 || !strings.Contains(strings.Join(c.Colors, ""), "U") || !strings.Contains(strings.Join(c.Colors, ""), "B") {
		t.Errorf("Horror %+v: a 1/1 blue and black token", c)
	}
	ciaAnswerCipher(t, g, me.ID, horror)
	if on, ok := encodedOn(g, id); !ok || on != horror {
		t.Fatal("Call of the Nightwing is encoded on the Horror")
	}

	dealCombatDamageToPlayer(g, horror, opp.ID, 1)
	passPriorityAroundTable(t, g)
	cp := copyOf(t, g, "Call of the Nightwing")
	if err := castCopy(g, me.ID, cp.InstanceID, true); err != nil {
		t.Fatalf("the free copy: %v", err)
	}
	passPriorityAroundTable(t, g)
	if ciaCipherPick(g, me.ID) != nil {
		t.Error("a copy is not a card: cipher asks nothing")
	}
	if second := ciaPermanentOf(g, me.ID, "Horror", horror); second == uuid.Nil {
		t.Error("the copy makes a second Horror")
	}
}

// Hands of Binding taps a creature an opponent controls; through the
// encoded creature, the copy taps a different one.
func TestCiaHandsOfBindingTapsAndTheCopyTapsAnother(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ciaSeats(t, g)
	mine := seedBear(g, me.ID)
	first := seedBear(g, opp.ID)
	second := seedBear(g, opp.ID)
	castCatalogSpell(t, g, "Hands of Binding", "Sorcery", ciaHandsOfBindingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: first}})
	if cipherResolveToPrompt(t, g, me.ID) == nil {
		t.Fatal("cipher asks")
	}
	if !ciaTapped(g, first) {
		t.Error("the targeted creature is tapped before cipher is asked")
	}
	ciaAnswerCipher(t, g, me.ID, mine)

	dealCombatDamageToPlayer(g, mine, opp.ID, 2)
	passPriorityAroundTable(t, g)
	cp := copyOf(t, g, "Hands of Binding")
	if err := castCopy(g, me.ID, cp.InstanceID, true, game.TargetRef{Kind: game.TargetCard, ID: second}); err != nil {
		t.Fatalf("the free copy with a new target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !ciaTapped(g, second) {
		t.Error("the copy taps the creature it targeted")
	}
	if ciaTapped(g, mine) {
		t.Error("your own creature is not a legal target and stays untapped")
	}
}

// Hands of Binding cannot target a creature you control.
func TestCiaHandsOfBindingTargetsOnlyAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ciaSeats(t, g)
	mine := seedBear(g, me.ID)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Hands of Binding", TypeLine: "Sorcery",
		OracleID: ciaHandsOfBindingOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}}); err == nil {
		t.Error("a creature you control is not \"a creature an opponent controls\"")
	}
}

// Mental Vapors: the target player discards a card they choose, and the
// cipher question waits for that answer.
func TestCiaMentalVaporsDiscardsThenAsksCipher(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ciaSeats(t, g)
	seedBear(g, me.ID)
	opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	hand := opp.Hand.Size()
	id := castCatalogSpell(t, g, "Mental Vapors", "Sorcery", ciaMentalVaporsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	for i := 0; i < 8 && discardChoiceFor(g, opp.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if ciaCipherPick(g, me.ID) != nil {
		t.Fatal("cipher is asked only after the discard")
	}
	discardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != hand-1 {
		t.Errorf("hand %d → %d: the target discards one", hand, opp.Hand.Size())
	}
	ciaAnswerCipher(t, g, me.ID, uuid.Nil)
	if !me.Graveyard.Contains(id) {
		t.Error("declined, Mental Vapors goes to the graveyard")
	}
}

// Midnight Recovery returns a creature card from your graveyard, and
// never a noncreature card.
func TestCiaMidnightRecoveryReturnsACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ciaSeats(t, g)
	dead := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	spell := b17GraveyardCard(me, "Old Spell", "Sorcery", "{1}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Midnight Recovery", TypeLine: "Sorcery",
		OracleID: ciaMidnightRecoveryOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}}); err == nil {
		t.Error("a sorcery card is not a creature card")
	}
	castCatalogSpell(t, g, "Midnight Recovery", "Sorcery", ciaMidnightRecoveryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card returns to your hand")
	}
}

// Paranoid Delusions mills three.
func TestCiaParanoidDelusionsMillsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ciaSeats(t, g)
	ciaLibrary(opp, 5)
	lib, yard := opp.Library.Size(), opp.Graveyard.Size()
	id := castCatalogSpell(t, g, "Paranoid Delusions", "Sorcery", ciaParanoidDelusionsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-3 || opp.Graveyard.Size() != yard+3 {
		t.Errorf("library %d → %d, graveyard %d → %d: mills three", lib, opp.Library.Size(), yard, opp.Graveyard.Size())
	}
	if !me.Graveyard.Contains(id) {
		t.Error("with no creature to encode on, the card goes to the graveyard")
	}
}

// Hidden Strings: two different targets, a tap-or-untap-or-neither
// question for each in turn, the second after the first.
func TestCiaHiddenStringsTapsOrUntapsEachTargetInTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ciaSeats(t, g)
	theirs := seedBear(g, opp.ID)
	mine := seedBear(g, me.ID)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Hidden Strings", TypeLine: "Sorcery",
		OracleID: ciaHiddenStringsOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 0},
	}}); err == nil {
		t.Error("\"another target permanent\": the same permanent twice is refused")
		passPriorityAroundTable(t, g)
	}
	castCatalogSpell(t, g, "Hidden Strings", "Sorcery", ciaHiddenStringsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: mine}})
	first := cipherResolveToPrompt(t, g, me.ID)
	if first == nil || len(first.PickOptions) != 3 || first.PickOptions[0].Cards[0] != theirs {
		t.Fatalf("the first target is asked about first, with three answers; got %+v", first)
	}
	answerOptionPick(t, g, me.ID, 0) // tap it
	if !ciaTapped(g, theirs) {
		t.Error("the first target is tapped")
	}
	second := latestOptionPickFor(g, me.ID)
	if second == nil || second.PickOptions[0].Cards[0] != mine {
		t.Fatalf("the second target is asked about once the first is answered; got %+v", second)
	}
	answerOptionPick(t, g, me.ID, 1) // untap it
	if ciaTapped(g, mine) {
		t.Error("the second target is untapped")
	}
	// The untapped bear is a creature you control: cipher is asked last.
	ciaAnswerCipher(t, g, me.ID, uuid.Nil)
}

// Hidden Strings' "you may": leaving both as they are changes nothing.
func TestCiaHiddenStringsMayLeaveBoth(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ciaSeats(t, g)
	a := seedBear(g, opp.ID)
	b := seedBear(g, opp.ID)
	castCatalogSpell(t, g, "Hidden Strings", "Sorcery", ciaHiddenStringsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	if cipherResolveToPrompt(t, g, me.ID) == nil {
		t.Fatal("the first question")
	}
	answerOptionPick(t, g, me.ID, 2)
	answerOptionPick(t, g, me.ID, 2)
	passPriorityAroundTable(t, g)
	if ciaTapped(g, a) || ciaTapped(g, b) {
		t.Error("left as they were, neither is tapped")
	}
}

// Arcane Heist: cast an instant from an opponent's graveyard for free;
// it is exiled instead of going back to their graveyard. A card in your
// own graveyard is not a legal target.
func TestCiaArcaneHeistCastsFromAnOpponentsGraveyardThenExiles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ciaSeats(t, g)
	third := g.Seats[(g.Turn.ActiveSeat+2)%4]
	bolt := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	own := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Arcane Heist", TypeLine: "Sorcery",
		OracleID: ciaArcaneHeistOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: own}}}); err == nil {
		t.Error("a card in your own graveyard is not in an opponent's")
		passPriorityAroundTable(t, g)
	}
	_, _ = me.Hand.Remove(id)

	castCatalogSpell(t, g, "Arcane Heist", "Sorcery", ciaArcaneHeistOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)
	life := third.Life
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}},
	}); err != nil {
		t.Fatalf("casting the opponent's instant for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if third.Life != life-3 {
		t.Errorf("life %d → %d: the stolen Bolt deals 3", life, third.Life)
	}
	if opp.Graveyard.Contains(bolt) || !inExile(g, bolt) {
		t.Error("the spell is exiled instead of going back to its owner's graveyard")
	}
}
