package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB022Thran     = "36282c29-a44b-4156-9441-03aa4bf386a8"
	oldB022TimeSieve = "3da5977a-36d4-4f32-ab9b-8b93809d818d"
	oldB022Lotus     = "41dd29b9-f08d-4ccc-8dc0-da11d2d456e9"
	oldB022Voldaren  = "4399c284-e3fe-46b1-aa1f-15a2846858f5"
	oldB022Sigil     = "98388eba-31ab-4302-be9e-ea67ea2b8c1a"
	oldB022Student   = "afcc68e2-3634-43ce-ae99-dd2ec735d146"
	oldB022Samut     = "b1eee275-5106-4c8d-a497-d7945ad81d74"
	oldB022Charix    = "c38cc5fe-936a-4d61-a816-9be773ae1d02"
)

func TestOldB022CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB022Thran:           "Thran Temporal Gateway",
		oldB022TimeSieve:       "Time Sieve",
		oldB022Lotus:           "Blacker Lotus",
		oldB022Voldaren:        "Voldaren Pariah",
		oldB022Voldaren + "#1": "Abolisher of Bloodlines",
		oldB022Sigil:           "Leshrac's Sigil",
		oldB022Student:         "Sakashima's Student",
		oldB022Samut:           "Samut, Voice of Dissent",
		oldB022Charix:          "Charix, the Raging Isle",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB022ThranTemporalGatewayPutsOnlyAHistoricPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gate := effectstest.B12Push(g, me.ID, "Thran Temporal Gateway", "Legendary Artifact", oldB022Thran, 0, 0)
	add := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		})
		return id
	}
	rock := add("Test Rock", "Artifact")
	king := add("Test King", "Legendary Creature — Human")
	saga := add("Test Saga", "Enchantment — Saga")
	bear := add("Test Bear", "Creature — Bear")
	spell := add("Test Legend Spell", "Legendary Sorcery")
	effectstest.ApaMana(me, "G", "G", "G", "G")
	effectstest.ApaActivate(t, g, me, gate, 0, game.ActivateAbilityParams{})
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no card was offered")
	}
	for _, id := range []uuid.UUID{rock, king, saga} {
		if !slices.Contains(pick.ChooseCards, id) {
			t.Errorf("the historic card %s is not offered: %v", id, pick.ChooseCards)
		}
	}
	for _, id := range []uuid.UUID{bear, spell} {
		if slices.Contains(pick.ChooseCards, id) {
			t.Errorf("a card that is not a historic permanent card was offered: %s", id)
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, king)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(king) || !me.Hand.Contains(rock) {
		t.Error("only the chosen card should have entered the battlefield")
	}
	if !effectstest.IsTapped(g, gate) {
		t.Error("the Gateway did not tap")
	}
}

func TestOldB022TimeSieveSacrificesFiveArtifactsForAnExtraTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	sieve := effectstest.B12Push(g, me.ID, "Time Sieve", "Artifact", oldB022TimeSieve, 0, 0)
	var rocks []uuid.UUID
	for range 3 {
		rocks = append(rocks, effectstest.B12Permanent(g, me.ID, "Rock", "Artifact"))
	}
	// Four artifacts in all: five cannot be paid.
	if err := g.ActivateCatalogAbility(me.ID, sieve, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: append([]uuid.UUID{sieve}, rocks...)}); err == nil {
		t.Fatal("four artifacts paid for five")
	}
	rocks = append(rocks, effectstest.B12Permanent(g, me.ID, "Rock", "Artifact"))
	creature := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, sieve, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: append([]uuid.UUID{creature}, rocks...)}); err == nil {
		t.Fatal("a creature was accepted as one of the five artifacts")
	}
	effectstest.ApaActivate(t, g, me, sieve, 0, game.ActivateAbilityParams{SacrificeIDs: append([]uuid.UUID{sieve}, rocks...)})
	for _, id := range append([]uuid.UUID{sieve}, rocks...) {
		if g.Battlefield.Contains(id) {
			t.Fatalf("artifact %s was not sacrificed", id)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, seat, true, "Time Sieve's extra turn")
}

func TestOldB022BlackerLotusAddsFourManaOfOneColorAndIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	lotus := effectstest.B12Push(g, me.ID, "Blacker Lotus", "Artifact", oldB022Lotus, 0, 0)
	if err := g.ActivateManaAbility(me.ID, lotus, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	color := effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	if got := me.ManaPool.CountColor(color); got != 4 || me.ManaPool.Total() != 4 {
		t.Fatalf("pool holds %d %s of %d mana, want four of one colour", got, color, me.ManaPool.Total())
	}
	if g.Battlefield.Contains(lotus) {
		t.Fatal("the Lotus stayed on the battlefield")
	}
	if !g.Exile.Contains(lotus) || me.Graveyard.Contains(lotus) {
		t.Fatal("the Lotus was not removed from the game (exiled, not sacrificed)")
	}
}

func TestOldB022SamutGivesOtherCreaturesHasteAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	samut := effectstest.B12Push(g, me.ID, "Samut, Voice of Dissent", "Legendary Creature — Human Warrior", oldB022Samut, 3, 4)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("another creature of mine has no haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, theirs).Tapped = true })
	effectstest.ApaMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, samut, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: samut}}}); err == nil {
		t.Fatal("Samut was accepted as the target of her own untap")
	}
	effectstest.ApaActivate(t, g, me, samut, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}})
	if effectstest.IsTapped(g, theirs) {
		t.Error("the target stayed tapped")
	}
	if !effectstest.IsTapped(g, samut) {
		t.Error("Samut did not tap for her own ability")
	}
}

func TestOldB022CharixTaxesSpellsThatTargetItAndGrowsForIslands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	charix := effectstest.B12Push(g, me.ID, "Charix, the Raging Isle", "Legendary Creature — Leviathan Crab", oldB022Charix, 0, 17)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	at := func(id uuid.UUID) []game.TargetRef { return []game.TargetRef{{Kind: game.TargetCard, ID: id}} }
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, at(charix)); got != 3 {
		t.Errorf("an opponent's spell at Charix costs %d, want 3", got)
	}
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, at(bear)); got != 1 {
		t.Errorf("an opponent's spell at another creature costs %d, want 1", got)
	}
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, me, at(charix)); got != 1 {
		t.Errorf("my own spell at Charix costs %d, want 1", got)
	}
	for range 2 {
		effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	}
	effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.ApaMana(me, "U", "U", "U")
	effectstest.ApaActivate(t, g, me, charix, 0, game.ActivateAbilityParams{})
	c := effectstest.ApaLive(g, charix)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 15 {
		t.Errorf("Charix is %d/%d after the pump with two Islands, want 2/15", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestOldB022SakashimasStudentEntersAsACopyThatIsAlsoANinja(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Their Bear", "Creature — Bear", 2, 3)
	student := effectstest.CastGraveyardCopier(t, g, "Sakashima's Student", "Creature — Human Ninja", oldB022Student, 0, 0)
	effectstest.ResolveWithCopyChoice(t, g, bear)
	got := effectstest.CopyBattlefieldCard(t, g, student)
	if got.Name != "Their Bear" || got.CurrentPower() != 2 || got.CurrentToughness() != 3 {
		t.Fatalf("entered as %q %d/%d, want a 2/3 Their Bear", got.Name, got.CurrentPower(), got.CurrentToughness())
	}
	if !got.HasSubtype("Bear") || !got.HasSubtype("Ninja") {
		t.Errorf("the copy is %q, want a Bear and a Ninja", got.TypeLine)
	}
	if got.Controller != me.ID {
		t.Error("the copy is not under my control")
	}
	abs := game.ActivatedAbilitiesForCard(game.Card{OracleID: oldB022Student})
	if len(abs) != 1 || abs[0].Cost.Mana != "{1}{U}" || abs[0].Cost.ReturnToHand.Empty() {
		t.Errorf("want one ninjutsu ability costing {1}{U} with a return cost, got %d abilities", len(abs))
	}
}

func TestOldB022LeshracsSigilLetsYouTakeACardFromAGreenCastersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AangAdvanceToMain(t, g, 1)
	effectstest.B12Push(g, me.ID, "Leshrac's Sigil", "Enchantment", oldB022Sigil, 0, 0)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	keep := uuid.New()
	take := uuid.New()
	spell := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: keep, Name: "Keep", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		opp.Hand.PushTop(game.Card{InstanceID: take, Name: "Take", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Green Spell", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}, Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.ApaMana(opp, "G")
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.ApaMana(me, "B", "B")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	pick := effectstest.OpenVariant(t, g)
	if pick.Chooser != me.ID || !slices.Contains(pick.DiscardOptions, take) || !slices.Contains(pick.DiscardOptions, keep) {
		t.Fatalf("pick: chooser %v options %v; want the Sigil's controller choosing from the caster's hand", pick.Chooser, pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{take}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(take) || !opp.Hand.Contains(keep) {
		t.Error("the chosen card should have been discarded and the other kept")
	}
	if me.ManaPool.CountColor("B") != 0 {
		t.Error("{B}{B} was not spent")
	}
}

func TestOldB022LeshracsSigilDeclinedOrNonGreenDoesNothingAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AangAdvanceToMain(t, g, 1)
	sigil := effectstest.B12Push(g, me.ID, "Leshrac's Sigil", "Enchantment", oldB022Sigil, 0, 0)
	red := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: red, Name: "Red Spell", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.ApaMana(opp, "R")
	if err := g.CastSpell(opp.ID, red, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a red spell triggered the Sigil")
	}
	effectstest.ApaMana(me, "B", "B")
	effectstest.ApaActivate(t, g, me, sigil, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(sigil) || !me.Hand.Contains(sigil) {
		t.Error("the Sigil was not returned to its owner's hand")
	}
}

func TestOldB022VoldarenPariahTransformsForThreeOthersAndForcesAnEdict(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	pariah := effectstest.ImportToHand(oldB022VoldarenRow(), me)
	card, ok := effectstest.CardInZone(me.Hand, pariah)
	if !ok {
		t.Fatal("the Pariah did not reach the hand")
	}
	me.Hand.Remove(pariah)
	effectstest.PushBattlefieldCardWithTimestamp(g, card)
	var fodder []uuid.UUID
	for range 2 {
		fodder = append(fodder, effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1))
	}
	var theirs []uuid.UUID
	for range 4 {
		theirs = append(theirs, effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	}
	if err := g.ActivateCatalogAbility(me.ID, pariah, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: fodder}); err == nil {
		t.Fatal("two creatures paid for three")
	}
	if err := g.ActivateCatalogAbility(me.ID, pariah, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: append(slices.Clone(fodder), pariah)}); err == nil {
		t.Fatal("the Pariah was accepted as one of the three OTHER creatures")
	}
	fodder = append(fodder, effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1))
	effectstest.ApaActivate(t, g, me, pariah, 0, game.ActivateAbilityParams{SacrificeIDs: fodder})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ApaLive(g, pariah)
	if c.Name != "Abolisher of Bloodlines" || c.CurrentPower() != 6 || c.CurrentToughness() != 5 {
		t.Fatalf("transformed into %q %d/%d, want a 6/5 Abolisher of Bloodlines", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, theirs[0], theirs[1], theirs[2])
	effectstest.PassPriorityAroundTable(t, g)
	left := 0
	for _, id := range theirs {
		if g.Battlefield.Contains(id) {
			left++
		}
	}
	if left != 1 {
		t.Errorf("%d of the opponent's four creatures remain, want 1 after sacrificing three", left)
	}
	if !g.Battlefield.Contains(pariah) {
		t.Error("the transformed creature is gone")
	}
}

func oldB022VoldarenRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Voldaren Pariah // Abolisher of Bloodlines",
		Layout:   "transform",
		TypeLine: "Creature — Vampire Horror // Creature — Eldrazi Vampire",
		OracleID: uuid.MustParse(oldB022Voldaren),
		Keywords: []string{"Flying", "Madness", "Transform"},
		CardFaces: []cards.CardFace{
			{Name: "Voldaren Pariah", TypeLine: "Creature — Vampire Horror", ManaCost: "{3}{B}{B}", Colors: []string{"B"}, Power: "3", Toughness: "3",
				OracleText: "Flying\nSacrifice three other creatures: Transform this creature.\nMadness {B}{B}{B}"},
			{Name: "Abolisher of Bloodlines", TypeLine: "Creature — Eldrazi Vampire", Colors: []string{"B"}, Power: "6", Toughness: "5",
				OracleText: "Flying\nWhen this creature transforms into Abolisher of Bloodlines, target opponent sacrifices three creatures of their choice."},
		},
	}
}
