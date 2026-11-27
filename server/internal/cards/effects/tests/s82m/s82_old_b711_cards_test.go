package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b711_cards_test.go — OLD batch 711 (Portal Three Kingdoms, #5622):
// instants and sorceries. Each test plays the real catalog entry through the
// engine.

const (
	oldB711Counterintel = "0c4e64d1-cdcf-4e96-8bf4-ed1414f6ad62"
	oldB711Rally        = "555927b3-bbc8-48b7-9a04-e8ebaabccf3e"
	oldB711BrokenDam    = "75fffc1b-db65-4c6b-9503-db50fccb837c"
	oldB711ImperialEd   = "ac2b3adb-45bb-4f56-9d52-e08771164b35"
	oldB711Deception    = "ae1b44df-1d8c-453a-9efb-6d1c93e52418"
	oldB711ForcedRetr   = "ccf1336c-2ea5-4582-84a5-0b29b3cca6c2"
	oldB711Eunuchs      = "e34f8aae-bf55-4bba-9524-c841591bb14b"
)

func oldB711Player(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func TestOldB711EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB711Counterintel: "Counterintelligence", oldB711Rally: "Rally the Troops", oldB711BrokenDam: "Broken Dam",
		oldB711ImperialEd: "Imperial Edict", oldB711Deception: "Deception", oldB711ForcedRetr: "Forced Retreat",
		oldB711Eunuchs: "Eunuchs' Intrigues",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB711CounterintelligenceReturnsOneOrTwoCreatures(t *testing.T) {
	for _, n := range []int{1, 2} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
		b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
		targets := []uuid.UUID{a, b}[:n]
		effectstest.CastCatalogSpell(t, g, "Counterintelligence", "Sorcery", oldB711Counterintel, effectstest.Pr7bTargets(targets...))
		effectstest.PassPriorityAroundTable(t, g)
		for i, id := range []uuid.UUID{a, b} {
			wantBack := i < n
			if wantBack && (g.Battlefield.Contains(id) || !opp.Hand.Contains(id)) {
				t.Errorf("%d targets: creature %d was not returned to its owner's hand", n, i)
			}
			if !wantBack && !g.Battlefield.Contains(id) {
				t.Errorf("%d targets: the untargeted creature %d was returned", n, i)
			}
		}
	}
}

func TestOldB711CounterintelligenceRefusesNoTargetsAndThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ids := []uuid.UUID{
		effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2),
		effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2),
		effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2),
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Counterintelligence", "Sorcery", oldB711Counterintel, effectstest.Pr7bTargets(ids...)); err == nil {
		t.Error("three targets were accepted")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Counterintelligence", "Sorcery", oldB711Counterintel, nil); err == nil {
		t.Error("no target was accepted")
	}
}

func TestOldB711RallyTheTroopsUntapsOnlyWhenYouWereAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if err := effectstest.CastCatalogSpellErr(t, g, "Rally the Troops", "Instant", oldB711Rally, nil); err == nil {
		t.Fatal("Rally the Troops was cast outside the declare attackers step")
	}
	att := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, opp.ID, "Tapped One", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Tapped Two", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, me.ID, "Tapped Foe", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{mine, other, theirs} {
		effectstest.TapCard(g, id)
	}
	effectstest.SfAttackHold(t, g, opp.ID, att)
	effectstest.SfCast(t, g, opp, game.Card{Name: "Rally the Troops", TypeLine: "Instant", OracleID: oldB711Rally}, game.CastSpellParams{})
	if effectstest.HmlTapped(t, g, mine) || effectstest.HmlTapped(t, g, other) {
		t.Error("a creature the caster controls stayed tapped")
	}
	if !effectstest.HmlTapped(t, g, theirs) {
		t.Error("a creature the caster does not control was untapped")
	}
}

func TestOldB711RallyTheTroopsRefusesAPlayerNobodyAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	att := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	effectstest.SfAttackHold(t, g, third.ID, att)
	card := game.Card{InstanceID: uuid.New(), Name: "Rally the Troops", TypeLine: "Instant", OracleID: oldB711Rally, Owner: opp.ID, Controller: opp.ID}
	opp.Hand.PushTop(card)
	if err := g.CastSpell(opp.ID, card.InstanceID, game.CastSpellParams{}); err == nil {
		t.Fatal("Rally the Troops was cast by a player nobody attacked")
	}
}

func TestOldB711BrokenDamTapsCreaturesWithoutHorsemanship(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	rider := effectstest.Pr7bCreature(g, opp.ID, "Rider", 2, "{2}", nil, "horsemanship")
	if err := effectstest.CastCatalogSpellErr(t, g, "Broken Dam", "Sorcery", oldB711BrokenDam, effectstest.Pr7bTargets(rider)); err == nil {
		t.Fatal("a creature with horsemanship was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Broken Dam", "Sorcery", oldB711BrokenDam, effectstest.Pr7bTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, a) || !effectstest.HmlTapped(t, g, b) {
		t.Error("a targeted creature stayed untapped")
	}
	if effectstest.HmlTapped(t, g, rider) {
		t.Error("the creature with horsemanship was tapped")
	}
}

func TestOldB711ImperialEdictDestroysTheCreatureTheOpponentChooses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	weak := effectstest.B12Creature(g, opp.ID, "Weak", "Creature — Bear", 1, 1)
	strong := effectstest.B12Creature(g, opp.ID, "Strong", "Creature — Giant", 5, 5)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Imperial Edict", "Sorcery", oldB711ImperialEd, oldB711Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, weak)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(weak) {
		t.Error("the chosen creature survived")
	}
	if !g.Battlefield.Contains(strong) || !g.Battlefield.Contains(mine) {
		t.Error("a creature that was not chosen was destroyed")
	}
}

func TestOldB711ImperialEdictWithNoCreatureDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Imperial Edict", "Sorcery", oldB711ImperialEd, oldB711Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's own creature was destroyed")
	}
}

func TestOldB711ImperialEdictRefusesYourselfAsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if err := effectstest.CastCatalogSpellErr(t, g, "Imperial Edict", "Sorcery", oldB711ImperialEd, oldB711Player(me.ID)); err == nil {
		t.Error("the caster was a legal target")
	}
}

func TestOldB711DeceptionMakesAnOpponentDiscardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	keep := effectstest.HandCardFull(opp, "Keep", "Instant", "{U}", "", nil)
	d1 := effectstest.HandCardFull(opp, "Pitch One", "Instant", "{U}", "", nil)
	d2 := effectstest.HandCardFull(opp, "Pitch Two", "Instant", "{U}", "", nil)
	hand := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Deception", "Sorcery", oldB711Deception, oldB711Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, opp.ID); got != 2 {
		t.Fatalf("the opponent owes %d discards, want 2", got)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, d1, d2)
	if got := opp.Hand.Size(); got != hand-2 {
		t.Errorf("hand %d, want %d", got, hand-2)
	}
	if !opp.Hand.Contains(keep) || !opp.Graveyard.Contains(d1) || !opp.Graveyard.Contains(d2) {
		t.Error("the opponent's own choice of cards was not honoured")
	}
}

func TestOldB711DeceptionRefusesYourselfAsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if err := effectstest.CastCatalogSpellErr(t, g, "Deception", "Sorcery", oldB711Deception, oldB711Player(me.ID)); err == nil {
		t.Error("the caster was a legal target")
	}
}

func TestOldB711ForcedRetreatPutsTheCreatureOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	size := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Forced Retreat", "Sorcery", oldB711ForcedRetr, effectstest.Pr7bTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Library.Contains(bear) {
		t.Fatal("the creature is not in its owner's library")
	}
	if opp.Library.Size() != size+1 {
		t.Errorf("library size %d, want %d", opp.Library.Size(), size+1)
	}
	if top := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID; top != bear {
		t.Error("the creature is not on top of the library")
	}
}

func TestOldB711EunuchsIntriguesLetsOnlyTheChosenCreatureBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chosen := effectstest.B12Creature(g, opp.ID, "Chosen", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	att := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Eunuchs' Intrigues", "Sorcery", oldB711Eunuchs, oldB711Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.B12Creature(g, opp.ID, "Late", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, att)
	if err := g.DeclareBlocker(other, att); err == nil {
		t.Error("another creature of the opponent blocked")
	}
	if err := g.DeclareBlocker(late, att); err == nil {
		t.Error("a creature that arrived later blocked")
	}
	if err := g.DeclareBlocker(chosen, att); err != nil {
		t.Errorf("the chosen creature could not block: %v", err)
	}
}

func TestOldB711EunuchsIntriguesDoesNotBindOtherPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	chosen := effectstest.B12Creature(g, opp.ID, "Chosen", "Creature — Bear", 2, 2)
	thirdBear := effectstest.B12Creature(g, third.ID, "Third", "Creature — Bear", 2, 2)
	att := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Eunuchs' Intrigues", "Sorcery", oldB711Eunuchs, oldB711Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, third.ID, att)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(thirdBear, att); err != nil {
		t.Errorf("a third player's creature was barred from blocking: %v", err)
	}
}
