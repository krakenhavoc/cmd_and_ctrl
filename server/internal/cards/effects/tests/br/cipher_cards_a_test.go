package testsbr

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ciaCipherPick is `chooser`'s open cipher question, or nil.
func ciaCipherPick(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	if c := effectstest.LatestOptionPickFor(g, chooser); c != nil && strings.Contains(c.Reason, "cipher") {
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
	effectstest.AnswerOptionPick(t, g, chooser, idx)
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

func ciaLibrary(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
}

// Call of the Nightwing makes its Horror before cipher is asked, so the
// Horror can carry the card; the Horror's combat damage offers a copy,
// and the copy makes a second Horror.
func TestCiaCallOfTheNightwingEncodesOnItsOwnHorror(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CiaSeats(t, g)
	id := effectstest.CastCatalogSpell(t, g, "Call of the Nightwing", "Sorcery", effectstest.CiaCallOfTheNightwingOracl, nil)
	if effectstest.CipherResolveToPrompt(t, g, me.ID) == nil {
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
	if on, ok := effectstest.EncodedOn(g, id); !ok || on != horror {
		t.Fatal("Call of the Nightwing is encoded on the Horror")
	}

	effectstest.DealCombatDamageToPlayer(g, horror, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	cp := effectstest.CopyOf(t, g, "Call of the Nightwing")
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true); err != nil {
		t.Fatalf("the free copy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CiaSeats(t, g)
	mine := effectstest.SeedBear(g, me.ID)
	first := effectstest.SeedBear(g, opp.ID)
	second := effectstest.SeedBear(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Hands of Binding", "Sorcery", effectstest.CiaHandsOfBindingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: first}})
	if effectstest.CipherResolveToPrompt(t, g, me.ID) == nil {
		t.Fatal("cipher asks")
	}
	if !effectstest.CiaTapped(g, first) {
		t.Error("the targeted creature is tapped before cipher is asked")
	}
	ciaAnswerCipher(t, g, me.ID, mine)

	effectstest.DealCombatDamageToPlayer(g, mine, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	cp := effectstest.CopyOf(t, g, "Hands of Binding")
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, game.TargetRef{Kind: game.TargetCard, ID: second}); err != nil {
		t.Fatalf("the free copy with a new target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.CiaTapped(g, second) {
		t.Error("the copy taps the creature it targeted")
	}
	if effectstest.CiaTapped(g, mine) {
		t.Error("your own creature is not a legal target and stays untapped")
	}
}

// Hands of Binding cannot target a creature you control.
func TestCiaHandsOfBindingTargetsOnlyAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CiaSeats(t, g)
	mine := effectstest.SeedBear(g, me.ID)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Hands of Binding", TypeLine: "Sorcery",
		OracleID: effectstest.CiaHandsOfBindingOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}}); err == nil {
		t.Error("a creature you control is not \"a creature an opponent controls\"")
	}
}

// Mental Vapors: the target player discards a card they choose, and the
// cipher question waits for that answer.
func TestCiaMentalVaporsDiscardsThenAsksCipher(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CiaSeats(t, g)
	effectstest.SeedBear(g, me.ID)
	opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	hand := opp.Hand.Size()
	id := effectstest.CastCatalogSpell(t, g, "Mental Vapors", "Sorcery", effectstest.CiaMentalVaporsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	for i := 0; i < 8 && effectstest.DiscardChoiceFor(g, opp.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if ciaCipherPick(g, me.ID) != nil {
		t.Fatal("cipher is asked only after the discard")
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
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
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CiaSeats(t, g)
	dead := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	spell := effectstest.B17GraveyardCard(me, "Old Spell", "Sorcery", "{1}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Midnight Recovery", TypeLine: "Sorcery",
		OracleID: effectstest.CiaMidnightRecoveryOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}}); err == nil {
		t.Error("a sorcery card is not a creature card")
	}
	effectstest.CastCatalogSpell(t, g, "Midnight Recovery", "Sorcery", effectstest.CiaMidnightRecoveryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card returns to your hand")
	}
}

// Paranoid Delusions mills three.
func TestCiaParanoidDelusionsMillsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CiaSeats(t, g)
	ciaLibrary(opp, 5)
	lib, yard := opp.Library.Size(), opp.Graveyard.Size()
	id := effectstest.CastCatalogSpell(t, g, "Paranoid Delusions", "Sorcery", effectstest.CiaParanoidDelusionsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CiaSeats(t, g)
	theirs := effectstest.SeedBear(g, opp.ID)
	mine := effectstest.SeedBear(g, me.ID)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Hidden Strings", TypeLine: "Sorcery",
		OracleID: effectstest.CiaHiddenStringsOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 0},
	}}); err == nil {
		t.Error("\"another target permanent\": the same permanent twice is refused")
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.CastCatalogSpell(t, g, "Hidden Strings", "Sorcery", effectstest.CiaHiddenStringsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: mine}})
	first := effectstest.CipherResolveToPrompt(t, g, me.ID)
	if first == nil || len(first.PickOptions) != 3 || first.PickOptions[0].Cards[0] != theirs {
		t.Fatalf("the first target is asked about first, with three answers; got %+v", first)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // tap it
	if !effectstest.CiaTapped(g, theirs) {
		t.Error("the first target is tapped")
	}
	second := effectstest.LatestOptionPickFor(g, me.ID)
	if second == nil || second.PickOptions[0].Cards[0] != mine {
		t.Fatalf("the second target is asked about once the first is answered; got %+v", second)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // untap it
	if effectstest.CiaTapped(g, mine) {
		t.Error("the second target is untapped")
	}
	// The untapped bear is a creature you control: cipher is asked last.
	ciaAnswerCipher(t, g, me.ID, uuid.Nil)
}

// Hidden Strings' "you may": leaving both as they are changes nothing.
func TestCiaHiddenStringsMayLeaveBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CiaSeats(t, g)
	a := effectstest.SeedBear(g, opp.ID)
	b := effectstest.SeedBear(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Hidden Strings", "Sorcery", effectstest.CiaHiddenStringsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	if effectstest.CipherResolveToPrompt(t, g, me.ID) == nil {
		t.Fatal("the first question")
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 2)
	effectstest.AnswerOptionPick(t, g, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CiaTapped(g, a) || effectstest.CiaTapped(g, b) {
		t.Error("left as they were, neither is tapped")
	}
}

// Arcane Heist: cast an instant from an opponent's graveyard for free;
// it is exiled instead of going back to their graveyard. A card in your
// own graveyard is not a legal target.
func TestCiaArcaneHeistCastsFromAnOpponentsGraveyardThenExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CiaSeats(t, g)
	third := g.Seats[(g.Turn.ActiveSeat+2)%4]
	bolt := effectstest.GyCard(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	own := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Arcane Heist", TypeLine: "Sorcery",
		OracleID: effectstest.CiaArcaneHeistOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: own}}}); err == nil {
		t.Error("a card in your own graveyard is not in an opponent's")
		effectstest.PassPriorityAroundTable(t, g)
	}
	_, _ = me.Hand.Remove(id)

	effectstest.CastCatalogSpell(t, g, "Arcane Heist", "Sorcery", effectstest.CiaArcaneHeistOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	life := third.Life
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}},
	}); err != nil {
		t.Fatalf("casting the opponent's instant for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if third.Life != life-3 {
		t.Errorf("life %d → %d: the stolen Bolt deals 3", life, third.Life)
	}
	if opp.Graveyard.Contains(bolt) || !effectstest.InExile(g, bolt) {
		t.Error("the spell is exiled instead of going back to its owner's graveyard")
	}
}
