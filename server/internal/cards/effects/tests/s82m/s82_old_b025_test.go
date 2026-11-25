package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB25KarrthusOID  = "242dfae4-504c-45d4-9a49-f7ab515fd23a"
	oldB25FelidarOID   = "33869ba6-13e5-4e48-8963-92f7965648fb"
	oldB25FrostOID     = "4c5a2443-a038-44d3-83a4-5e9a6db7b1a7"
	oldB25PrelateOID   = "7e0cdba8-45e0-4fe5-b440-2d8fcf4d6b2b"
	oldB25AmarethOID   = "8cccc8d0-bdd8-4853-8330-6f6a57695d11"
	oldB25DoctorOID    = "90aa8344-ffc7-4a1a-a02a-80eceb1babe5"
	oldB25PlagueOID    = "df130616-8efe-42ed-a7f0-b58bef3f11c4"
	oldB25DragonLine   = "Creature — Dragon"
	oldB25BearTypeLine = "Creature — Bear"
)

func TestOldB25CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB25KarrthusOID: "Karrthus, Tyrant of Jund", oldB25FelidarOID: "Felidar Guardian",
		oldB25FrostOID: "Frost Titan", oldB25PrelateOID: "Sanctum Prelate",
		oldB25AmarethOID: "Amareth, the Lustrous", oldB25DoctorOID: "The Fifteenth Doctor",
		oldB25PlagueOID: "Plague Engineer",
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

func oldB25Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tgh int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tgh, Owner: owner, Controller: owner,
	})
}

func oldB25LibraryCard(g *game.Game, p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

// oldB25AnswerConfirm settles the stack and answers the open "you may" prompt.
func oldB25AnswerConfirm(t *testing.T, g *game.Game, me *game.Player, yes bool) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if c == nil {
		t.Fatal("no confirm prompt is open")
	}
	if err := g.ResolveConfirm(c.ID, me.ID, yes); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
}

func oldB25PushTapped(g *game.Game, owner uuid.UUID, name, typeLine string, p, tgh int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: p, Toughness: tgh, Owner: owner, Controller: owner, Tapped: true,
	})
}

// --- Plague Engineer ----------------------------------------------------------

func TestOldB25PlagueEngineerShrinksOpponentsCreaturesOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plague Engineer", TypeLine: "Creature — Phyrexian Carrier", OracleID: oldB25PlagueOID,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, NamedTribe: "Elf",
	})
	mine := oldB25Push(g, me.ID, "My Elf", "Creature — Elf", "", 2, 2)
	theirs := oldB25Push(g, opp.ID, "Their Elf", "Creature — Elf", "", 2, 2)
	other := oldB25Push(g, opp.ID, "Their Bear", oldB25BearTypeLine, "", 2, 2)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if p, tg := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 1 || tg != 1 {
		t.Errorf("an opposing Elf is %d/%d, want 1/1", p, tg)
	}
	if p := effectstest.EffectivePower(t, g, mine); p != 2 {
		t.Errorf("my own Elf has power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, other); p != 2 {
		t.Errorf("an opposing Bear has power %d, want 2", p)
	}
}

func TestOldB25PlagueEngineerAsksForATypeAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := oldB25Push(g, opp.ID, "Their Elf", "Creature — Elf", "", 1, 1)
	engineer := effectstest.CastHoldCreature(t, g, "Plague Engineer", "Creature — Phyrexian Carrier", oldB25PlagueOID, 2, 2)
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !effectstest.HasEffectiveKeyword(t, g, engineer, "deathtouch") {
		t.Error("Plague Engineer has no deathtouch")
	}
	// A 1/1 Elf with -1/-1 dies to state-based actions.
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opposing 1/1 Elf survived the -1/-1")
	}
}

// --- Karrthus, Tyrant of Jund -----------------------------------------------

func TestOldB25KarrthusTakesAndUntapsEveryDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := oldB25PushTapped(g, me.ID, "My Dragon", oldB25DragonLine, 4, 4)
	theirs := oldB25PushTapped(g, opp.ID, "Their Dragon", "Artifact Creature — Dragon", 3, 3)
	bear := oldB25PushTapped(g, opp.ID, "Their Bear", oldB25BearTypeLine, 2, 2)
	karrthus := effectstest.CastHoldCreature(t, g, "Karrthus, Tyrant of Jund", "Legendary Creature — Dragon", oldB25KarrthusOID, 7, 7)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Errorf("the opposing Dragon is controlled by %s, want me", got)
	}
	if got := effectstest.ControllerOf(t, g, bear); got != opp.ID {
		t.Error("a non-Dragon changed hands")
	}
	for name, id := range map[string]uuid.UUID{"my Dragon": mine, "their Dragon": theirs} {
		if effectstest.TappedOnBattlefield(t, g, id) {
			t.Errorf("%s is still tapped", name)
		}
	}
	if !effectstest.TappedOnBattlefield(t, g, bear) {
		t.Error("a non-Dragon was untapped")
	}
	// The theft is for good: it is still mine after the turn rolls over.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Error("the stolen Dragon went back at end of turn")
	}
	_ = karrthus
}

func TestOldB25KarrthusGivesOtherDragonsYouControlHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	karrthus := oldB25Push(g, me.ID, "Karrthus, Tyrant of Jund", "Legendary Creature — Dragon", oldB25KarrthusOID, 7, 7)
	mine := oldB25Push(g, me.ID, "My Dragon", oldB25DragonLine, "", 4, 4)
	bear := oldB25Push(g, me.ID, "My Bear", oldB25BearTypeLine, "", 2, 2)
	theirs := oldB25Push(g, opp.ID, "Their Dragon", oldB25DragonLine, "", 3, 3)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for _, tc := range []struct {
		id   uuid.UUID
		want bool
		note string
	}{{karrthus, true, "Karrthus is printed with haste"}, {mine, true, "another Dragon I control"}, {bear, false, "a non-Dragon"}, {theirs, false, "an opponent's Dragon"}} {
		if got := effectstest.HasEffectiveKeyword(t, g, tc.id, "haste"); got != tc.want {
			t.Errorf("%s: haste = %v, want %v", tc.note, got, tc.want)
		}
	}
}

// --- Felidar Guardian ---------------------------------------------------------

func TestOldB25FelidarGuardianBlinksAnotherPermanentYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := oldB25Push(g, me.ID, "Bear", oldB25BearTypeLine, "", 2, 2)
	theirs := oldB25Push(g, opp.ID, "Their Bear", oldB25BearTypeLine, "", 2, 2)
	guardian := effectstest.CastHoldCreature(t, g, "Felidar Guardian", "Creature — Cat Beast", oldB25FelidarOID, 1, 4)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the Guardian asked for no target")
	}
	for _, bad := range []uuid.UUID{guardian, theirs} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("%s was accepted as a target", bad)
		}
	}
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Bear" {
			n++
			if !c.SummonedThisTurn || c.Controller != me.ID {
				t.Error("the Bear did not re-enter under my control as a new object")
			}
		}
	}
	if n != 1 {
		t.Fatalf("%d Bears on the battlefield, want 1", n)
	}
}

func TestOldB25FelidarGuardianBlinkIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := oldB25Push(g, me.ID, "Bear", oldB25BearTypeLine, "", 2, 2)
	effectstest.CastHoldCreature(t, g, "Felidar Guardian", "Creature — Cat Beast", oldB25FelidarOID, 1, 4)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, false)
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || c.SummonedThisTurn {
		t.Error("declining the blink still exiled the Bear")
	}
}

// --- Frost Titan --------------------------------------------------------------

func TestOldB25FrostTitanTapsAndFreezesATargetWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := oldB25Push(g, opp.ID, "Their Bear", oldB25BearTypeLine, "", 2, 2)
	effectstest.CastHoldCreature(t, g, "Frost Titan", "Creature — Giant", oldB25FrostOID, 6, 6)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if !effectstest.TappedOnBattlefield(t, g, bear) {
		t.Error("the target was not tapped")
	}
	if skips := effectstest.HmlUntapSkips(t, g, bear); len(skips) != 1 {
		t.Errorf("marks = %+v, want one next-untap-step mark", skips)
	}
}

func TestOldB25FrostTitanTapsAndFreezesWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := oldB25Push(g, me.ID, "Frost Titan", "Creature — Giant", oldB25FrostOID, 6, 6)
	bear := oldB25Push(g, opp.ID, "Their Bear", oldB25BearTypeLine, "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, titan)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if !effectstest.TappedOnBattlefield(t, g, bear) {
		t.Error("the attack trigger did not tap the target")
	}
	if skips := effectstest.HmlUntapSkips(t, g, bear); len(skips) != 1 {
		t.Errorf("marks = %+v, want one next-untap-step mark", skips)
	}
}

func TestOldB25FrostTitanCountersAnOpponentsSpellUnlessTwoIsPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := oldB25Push(g, me.ID, "Frost Titan", "Creature — Giant", oldB25FrostOID, 6, 6)
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, titan)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no payment prompt for the spell's controller")
	}
	if c := effectstest.PendingOfKind(g, game.PendingChoicePayUnless); c == nil || c.PayCost != "{2}" {
		t.Errorf("the tax is %+v, want {2}", c)
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(titan) || !opp.Graveyard.Contains(blade) {
		t.Error("declining the tax must counter the spell")
	}
	_ = me
}

// --- Sanctum Prelate ----------------------------------------------------------

func TestOldB25SanctumPrelateBansNoncreatureSpellsWithTheChosenManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sanctum Prelate", TypeLine: "Creature — Human Cleric", OracleID: oldB25PrelateOID,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, ChosenNumber: 3,
	})
	for _, tc := range []struct {
		name, typeLine, cost string
		x                    int
		refused              bool
	}{
		{"Three Sorcery", "Sorcery", "{2}{U}", 0, true},
		{"Three Artifact", "Artifact", "{3}", 0, true},
		{"X Sorcery", "Sorcery", "{X}{U}", 2, true},
		{"Two Sorcery", "Sorcery", "{1}{U}", 0, false},
		{"Four Sorcery", "Sorcery", "{3}{U}", 0, false},
		{"Three Creature", "Creature — Bear", "{2}{G}", 0, false},
	} {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: tc.name, TypeLine: tc.typeLine, ManaCost: tc.cost, Owner: me.ID, Controller: me.ID})
		err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: tc.x})
		if refused := err != nil; refused != tc.refused {
			t.Errorf("%s: cast err = %v, want refused %v", tc.name, err, tc.refused)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
}

func TestOldB25SanctumPrelateAsksForANumberAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastHoldCreature(t, g, "Sanctum Prelate", "Creature — Human Cleric", oldB25PrelateOID, 2, 2)
	c := effectstest.PayAmountPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no number prompt")
	}
	if pa := c.PayAmount; pa.Min != 0 {
		t.Fatalf("prompt %+v, want a floor of 0", *pa)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 0); err != nil {
		t.Fatalf("answer 0: %v", err)
	}
	// Zero is a number: a noncreature spell with mana value 0 is now barred.
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Free Sorcery", TypeLine: "Sorcery", ManaCost: "{0}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
		t.Error("a mana value 0 noncreature spell was cast after choosing 0")
	}
}

// --- Amareth, the Lustrous ----------------------------------------------------

func TestOldB25AmarethOffersATopCardThatSharesATypeWithAnEnteringPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	amareth := oldB25Push(g, me.ID, "Amareth, the Lustrous", "Legendary Creature — Dragon", oldB25AmarethOID, 6, 6)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !effectstest.HasEffectiveKeyword(t, g, amareth, "flying") {
		t.Error("Amareth has no flying")
	}
	top := effectstest.PushLibraryTyped(me, "Top Bear", oldB25BearTypeLine)
	hand := me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Another Bear", oldB25BearTypeLine, "", 2, 2)
	oldB25AnswerConfirm(t, g, me, true)
	if !me.Hand.Contains(top) || me.Hand.Size() != hand+1 {
		t.Errorf("the creature on top was not taken (hand %d, was %d)", me.Hand.Size(), hand)
	}
}

func TestOldB25AmarethLeavesATopCardThatSharesNoType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB25Push(g, me.ID, "Amareth, the Lustrous", "Legendary Creature — Dragon", oldB25AmarethOID, 6, 6)
	top := effectstest.PushLibraryTyped(me, "Top Island", "Basic Land — Island")
	effectstest.CastHoldCreature(t, g, "Another Bear", oldB25BearTypeLine, "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Error("a confirm was asked for a top card that shares no type")
	}
	if me.Hand.Contains(top) || effectstest.LibraryTopIDs(me, 1)[0] != top {
		t.Error("a land on top was taken for a creature entering")
	}
}

func TestOldB25AmarethTakingTheTopCardIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB25Push(g, me.ID, "Amareth, the Lustrous", "Legendary Creature — Dragon", oldB25AmarethOID, 6, 6)
	top := effectstest.PushLibraryTyped(me, "Top Bear", oldB25BearTypeLine)
	effectstest.CastHoldCreature(t, g, "Another Bear", oldB25BearTypeLine, "", 2, 2)
	oldB25AnswerConfirm(t, g, me, false)
	if me.Hand.Contains(top) {
		t.Error("declining still put the card into hand")
	}
}

// --- The Fifteenth Doctor -----------------------------------------------------

func TestOldB25FifteenthDoctorMillsThreeAndTakesAnArtifactWithManaValueTwoOrThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	cheap := oldB25LibraryCard(g, me, "Cheap Rock", "Artifact", "{2}")
	one := oldB25LibraryCard(g, me, "One Rock", "Artifact", "{1}")
	effectstest.PushLibraryTyped(me, "Filler", oldB25BearTypeLine)
	effectstest.CastHoldCreature(t, g, "The Fifteenth Doctor", "Legendary Creature — Time Lord Doctor", oldB25DoctorOID, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no card was offered from the milled cards")
	}
	if len(p.ChooseCards) != 1 || p.ChooseCards[0] != cheap {
		t.Fatalf("offered %v, want only the mana value 2 artifact", p.ChooseCards)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{cheap}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(cheap) || !me.Graveyard.Contains(one) {
		t.Error("the artifact was not put into hand, or the others were not milled")
	}
}

func TestOldB25FifteenthDoctorGivesTheFirstNonartifactSpellImprovise(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB25Push(g, me.ID, "The Fifteenth Doctor", "Legendary Creature — Time Lord Doctor", oldB25DoctorOID, 3, 3)
	rock := oldB25Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	g.WithWriteLock(func() {})
	sorcery := game.Card{InstanceID: uuid.New(), Name: "Sorcery", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID}
	artifactSpell := game.Card{InstanceID: uuid.New(), Name: "Trinket", TypeLine: "Artifact", ManaCost: "{1}", Owner: me.ID, Controller: me.ID}
	var gotSorcery, gotArtifact bool
	g.WithWriteLock(func() {
		if c := g.TapPermanentsCostForCastLocked(me.ID, sorcery, game.ZoneHand); !c.Empty() {
			gotSorcery = c.Key == "improvise"
		}
		if c := g.TapPermanentsCostForCastLocked(me.ID, artifactSpell, game.ZoneHand); !c.Empty() {
			gotArtifact = c.Key == "improvise"
		}
	})
	if !gotSorcery {
		t.Error("the first nonartifact spell has no improvise")
	}
	if gotArtifact {
		t.Error("an artifact spell was given improvise")
	}
	_ = rock
}
