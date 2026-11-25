package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b104_cards_test.go — S82 OLD batch 104 (#5510, Jumpstart): creatures
// with keywords, instants and sorceries, counterspells and bounce, removal.
// Each test plays the real catalog entry.

const (
	oldB104Airship   = "3c4661da-db2e-4834-9bba-d9a5592465d2"
	oldB104Shrieker  = "a5196310-f449-4d5e-8ea8-901a83bad6ff"
	oldB104Stinger   = "c57edca9-b71e-4ec1-84d5-4b178969ccb7"
	oldB104Flurry    = "0c8ab7bf-7b6a-4a34-b988-9f3f4d84a9ea"
	oldB104Uprising  = "d9d3f6fc-5087-451a-94b1-ffe1a5d30f5e"
	oldB104Companion = "668f646b-94f0-4f1a-badd-2d69b559d8a2"
	oldB104Ruffian   = "c74d72ee-4795-41ca-b676-4d5a05bd12b7"
	oldB104Horror    = "b1eb7a92-3055-426e-9716-cc6d2a69f0af"
	oldB104Sweep     = "c6901a89-ac7d-4797-a027-97bb1cf42e60"
	oldB104Strike    = "1ad6339d-6206-49bf-bf97-05e6d3bcaab7"
	oldB104Sethron   = "561c1ce9-41e7-474a-9476-bb9a7d24f0d5"
	oldB104Chariot   = "90f4ec87-d33d-46c1-8cc7-3e7ad3f8ad70"
)

func TestOldB104EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB104Airship:   "Rishadan Airship",
		oldB104Shrieker:  "Lightning Shrieker",
		oldB104Stinger:   "Nightshade Stinger",
		oldB104Flurry:    "Flurry of Horns",
		oldB104Uprising:  "Elemental Uprising",
		oldB104Companion: "Cathar's Companion",
		oldB104Ruffian:   "Slate Street Ruffian",
		oldB104Horror:    "Dinrova Horror",
		oldB104Sweep:     "Sweep Away",
		oldB104Strike:    "Assassin's Strike",
		oldB104Sethron:   "Sethron, Hurloon General",
		oldB104Chariot:   "Warmonger's Chariot",
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

func TestOldB104RishadanAirshipBlocksOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	ship := effectstest.B12Push(g, them.ID, "Rishadan Airship", "Creature — Human Pirate", oldB104Airship, 3, 1)
	ground := effectstest.B12Push(g, me.ID, "Ground attacker", "Creature — Bear", "", 2, 2)
	flyer := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Flier", TypeLine: "Creature — Bird", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Keywords: []string{"flying"}})
	effectstest.AttackIntoBlocks(t, g, them.ID, ground, flyer)
	if effectstest.BlockRefusal(t, g, ship, ground) == "" {
		t.Error("the Airship blocked a creature without flying")
	}
	if got := effectstest.BlockRefusal(t, g, ship, flyer); got != "" {
		t.Errorf("the Airship was refused a flyer: %q", got)
	}
}

func TestOldB104NightshadeStingerCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Nightshade Stinger", "Creature — Faerie Rogue", oldB104Stinger, 1, 1)
	effectstest.AssertRestrictions(t, g, id, game.CantBlock)
}

func TestOldB104CatharsCompanionGainsIndestructibleOnlyFromNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dog := effectstest.B12Push(g, me.ID, "Cathar's Companion", "Creature — Dog", oldB104Companion, 3, 1)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, dog, "indestructible") {
		t.Fatal("a creature spell gave indestructible")
	}
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", effectstest.AcPlayerRef(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, dog, "indestructible") {
		t.Error("a noncreature spell did not give indestructible")
	}
}

func TestOldB104SlateStreetRuffianMakesTheDefenderDiscardWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ruffian := effectstest.B12Push(g, me.ID, "Slate Street Ruffian", "Creature — Human Warrior", oldB104Ruffian, 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	held := effectstest.PushHandCardOfType(opp, "Held", "Sorcery")
	effectstest.UnblockedAttackIntoBlockers(t, g, ruffian)
	if err := g.DeclareBlocker(wall, ruffian); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
		t.Fatalf("the defender was not asked to discard: %+v", g.PendingChoices)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, held)
	if !opp.Graveyard.Contains(held) {
		t.Error("the chosen card was not discarded")
	}
}

func TestOldB104SlateStreetRuffianDoesNothingWhenUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ruffian := effectstest.B12Push(g, me.ID, "Slate Street Ruffian", "Creature — Human Warrior", oldB104Ruffian, 2, 2)
	effectstest.PushHandCardOfType(opp, "Held", "Sorcery")
	effectstest.UnblockedAttackIntoBlockers(t, g, ruffian)
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		t.Error("an unblocked Ruffian made the defender discard")
	}
}

func TestOldB104LightningShriekerShufflesItselfAwayAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Lightning Shrieker", "Creature — Dragon", oldB104Shrieker, 5, 5)
	lib := me.Library.Size()
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Fatal("the Shrieker stayed on the battlefield")
	}
	if me.Library.Size() != lib+1 {
		t.Errorf("library %d, want %d", me.Library.Size(), lib+1)
	}
	for _, c := range me.Library.Cards {
		if c.InstanceID == id {
			return
		}
	}
	t.Error("the Shrieker is not in its owner's library")
}

func TestOldB104FlurryOfHornsMakesTwoHastyMinotaurs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Flurry of Horns", "Sorcery", oldB104Flurry, nil)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Minotaur" || c.Controller != me.ID {
			continue
		}
		n++
		if p, tough := effectstest.EffectivePTOf(t, g, c.InstanceID); p != 2 || tough != 3 || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "haste") {
			t.Errorf("token is %d/%d haste=%v, want 2/3 haste", p, tough, effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "haste"))
		}
	}
	if n != 2 {
		t.Errorf("%d Minotaur tokens, want 2", n)
	}
}

func TestOldB104ElementalUprisingAnimatesALandUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.CastCatalogSpellErr(t, g, "Elemental Uprising", "Instant", oldB104Uprising, effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("an opponent's land was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Elemental Uprising", "Instant", oldB104Uprising, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	types := effectstest.EffectiveTypes(t, g, land)
	has := func(s string) bool {
		for _, x := range types {
			if x == s {
				return true
			}
		}
		return false
	}
	if !has("land") && !has("Land") || !has("creature") && !has("Creature") {
		t.Errorf("types %v: want it a land and a creature", types)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, land); p != 4 || tough != 4 {
		t.Errorf("land is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, land, "haste") {
		t.Error("the land has no haste")
	}
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepPrecombatMain)
	if effectstest.EffectiveOf(t, g, land).Power == 4 {
		t.Error("the land stayed a 4/4 after the turn")
	}
}

func TestOldB104DinrovaHorrorBouncesAPermanentAndItsOwnerDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Push(g, opp.ID, "Foe", "Creature — Bear", "", 3, 3)
	held := effectstest.PushHandCardOfType(opp, "Held", "Sorcery")
	effectstest.DftABCast(t, g, me, game.Card{Name: "Dinrova Horror", TypeLine: "Creature — Horror", OracleID: oldB104Horror, Power: 4, Toughness: 4})
	effectstest.PickCard(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) || !effectstest.HandHasID(opp, foe) {
		t.Fatal("the permanent was not returned to its owner's hand")
	}
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
		t.Fatal("the permanent's owner was not asked to discard")
	}
	effectstest.AnswerDiscard(t, g, opp.ID, held)
	if !opp.Graveyard.Contains(held) {
		t.Error("the chosen card was not discarded")
	}
}

func TestOldB104SweepAwayReturnsAnIdleCreatureToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Push(g, opp.ID, "Foe", "Creature — Bear", "", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Sweep Away", "Instant", oldB104Sweep, effectstest.CardRefs(foe))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("a creature that is not attacking was offered the library")
	}
	if !effectstest.HandHasID(opp, foe) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestOldB104SweepAwayMayTuckAnAttackerOnTopOfItsLibrary(t *testing.T) {
	for _, tuck := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		atk := effectstest.B12Push(g, opp.ID, "Raider", "Creature — Bear", "", 3, 3)
		effectstest.AdvanceToMainOf(t, g, 1)
		effectstest.DeclareAttack(t, g, me.ID, atk)
		id := uuid.New()
		opp2 := g.Seats[0]
		opp2.Hand.PushTop(game.Card{InstanceID: id, Name: "Sweep Away", TypeLine: "Instant", OracleID: oldB104Sweep, Owner: opp2.ID, Controller: opp2.ID})
		if err := g.CastSpell(opp2.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(atk)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestConfirmFor(g, opp2.ID) == nil {
			t.Fatalf("tuck=%v: the attacking creature's bounce asked nothing", tuck)
		}
		effectstest.AnswerMayChoice(t, g, opp2.ID, tuck)
		effectstest.PassPriorityAroundTable(t, g)
		topped := len(opp.Library.Cards) > 0 && opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID == atk
		if tuck && (!topped || effectstest.HandHasID(opp, atk)) {
			t.Errorf("the attacker is not on top of its owner's library (topped=%v)", topped)
		}
		if !tuck && !effectstest.HandHasID(opp, atk) {
			t.Error("declining did not return the attacker to hand")
		}
		_ = me
	}
}

func TestOldB104AssassinsStrikeDestroysAndItsControllerDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Push(g, opp.ID, "Foe", "Creature — Bear", "", 3, 3)
	held := effectstest.PushHandCardOfType(opp, "Held", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Assassin's Strike", "Sorcery", oldB104Strike, effectstest.CardRefs(foe))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) || !opp.Graveyard.Contains(foe) {
		t.Fatal("the creature was not destroyed")
	}
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
		t.Fatal("the creature's controller was not asked to discard")
	}
	effectstest.AnswerDiscard(t, g, opp.ID, held)
	if !opp.Graveyard.Contains(held) {
		t.Error("the chosen card was not discarded")
	}
}

func oldB104Minotaurs(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == "Minotaur" {
			n++
		}
	}
	return n
}

func TestOldB104SethronMakesAMinotaurForItselfAndEachOtherNontokenMinotaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.DftABCast(t, g, me, game.Card{Name: "Sethron, Hurloon General", TypeLine: "Legendary Creature — Minotaur Warrior", OracleID: oldB104Sethron, Power: 4, Toughness: 4})
	if got := oldB104Minotaurs(g, me.ID); got != 1 {
		t.Fatalf("%d Minotaur tokens after Sethron entered, want 1", got)
	}
	effectstest.DftABCast(t, g, me, game.Card{Name: "Other Minotaur", TypeLine: "Creature — Minotaur", Power: 2, Toughness: 2})
	if got := oldB104Minotaurs(g, me.ID); got != 2 {
		t.Errorf("%d tokens after another nontoken Minotaur entered, want 2", got)
	}
	effectstest.DftABCast(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if got := oldB104Minotaurs(g, me.ID); got != 2 {
		t.Errorf("a non-Minotaur made a token: %d", got)
	}
}

func TestOldB104SethronPumpsOnlyYourMinotaurs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seth := effectstest.B12Push(g, me.ID, "Sethron, Hurloon General", "Legendary Creature — Minotaur Warrior", oldB104Sethron, 4, 4)
	mine := effectstest.B12Push(g, me.ID, "Minotaur", "Creature — Minotaur", "", 2, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Minotaur", "Creature — Minotaur", "", 2, 2)
	effectstest.FillPool(me, 3)
	effectstest.Pr7Activate(t, g, me.ID, seth, 0, game.ActivateAbilityParams{})
	for _, id := range []uuid.UUID{seth, mine} {
		if p, _ := effectstest.EffectivePTOf(t, g, id); !effectstest.HasEffectiveKeyword(t, g, id, "menace") || !effectstest.HasEffectiveKeyword(t, g, id, "haste") || p < 3 {
			t.Errorf("a Minotaur you control: power %d menace=%v haste=%v", p, effectstest.HasEffectiveKeyword(t, g, id, "menace"), effectstest.HasEffectiveKeyword(t, g, id, "haste"))
		}
	}
	for _, id := range []uuid.UUID{bear, theirs} {
		if p, _ := effectstest.EffectivePTOf(t, g, id); p != 2 || effectstest.HasEffectiveKeyword(t, g, id, "menace") {
			t.Errorf("a creature that is not yours or not a Minotaur was affected: power %d", p)
		}
	}
}

func TestOldB104WarmongersChariotPumpsAndLetsADefenderAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	chariot := effectstest.PushCatalogPermanent(g, me.ID, "Warmonger's Chariot", "Artifact — Equipment", oldB104Chariot, false)
	wall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 4, Keywords: []string{"defender"}, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.DeclareAttacker(wall, opp.ID); err == nil {
		t.Fatal("an unequipped defender attacked")
	}
	effectstest.AdvanceToStepOf(t, g, g.Turn.ActiveSeat, game.StepPrecombatMain)
	effectstest.FillPool(me, 3)
	effectstest.EquipTo(t, g, me.ID, chariot, wall)
	if p, tough := effectstest.EffectivePTOf(t, g, wall); p != 2 || tough != 6 {
		t.Errorf("equipped Wall is %d/%d, want 2/6", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, wall, "defender") {
		t.Error("the keyword should stay")
	}
	effectstest.DeclareAttack(t, g, opp.ID, wall)
	if c, ok := effectstest.BattlefieldCard(g, wall); !ok || c.AttackingTarget != opp.ID {
		t.Error("the equipped defender could not attack")
	}
}
