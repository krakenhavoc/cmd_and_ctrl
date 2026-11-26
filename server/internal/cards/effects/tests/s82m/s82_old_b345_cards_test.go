package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b345_cards_test.go — Khans of Tarkir, batch 345 (S82 older sets,
// #5554): twelve morph cards. Each is cast face down for {3} and turned face
// up through the real special action, then its trigger resolves.

const (
	oldB345Jeering    = "170adfe9-ffbe-40b1-b665-e756e7eb0302"
	oldB345Abomin     = "3d98af5f-7a0b-4a5a-b3e4-f3c9d150c993"
	oldB345Pine       = "4c92b862-34c7-4195-bee3-bcc0cafad140"
	oldB345Rattleclaw = "84c9d8b8-8dd9-41fc-a2b6-af1ffb9090f9"
	oldB345Efreet     = "882a9dc9-b1bb-4be8-850f-e28bad8f5840"
	oldB345Pony       = "8cfb5181-8ab7-484b-8821-a62c561f762f"
	oldB345Watcher    = "92cf18ab-dd7b-4bc8-8b85-a81b8cb5192c"
	oldB345Mistfire   = "9af9c353-6b71-46e7-8815-a920e012dbb8"
	oldB345Charger    = "a0665c10-e652-4608-920b-66206c0e5905"
	oldB345Savants    = "ca588f4e-4d2f-4aae-84de-be9862e13f1c"
	oldB345Ripper     = "cfbcd70b-d3a9-47e9-a8b3-b8134d00dff8"
	oldB345Ambusher   = "fb7f45eb-c134-4493-af05-e3b14d80e48c"
)

func TestOldB345EveryCardIsRegisteredFullWithAMorph(t *testing.T) {
	for name, oracle := range map[string]string{
		"Jeering Instigator": oldB345Jeering, "Abomination of Gudul": oldB345Abomin, "Pine Walker": oldB345Pine,
		"Rattleclaw Mystic": oldB345Rattleclaw, "Efreet Weaponmaster": oldB345Efreet, "Ponyback Brigade": oldB345Pony,
		"Watcher of the Roost": oldB345Watcher, "Mistfire Weaver": oldB345Mistfire, "Temur Charger": oldB345Charger,
		"Dragon's Eye Savants": oldB345Savants, "Ruthless Ripper": oldB345Ripper, "Horde Ambusher": oldB345Ambusher,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not in the catalog", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %d, want Full with none", name, spec.Name, spec.Completeness, len(spec.Caveats))
		}
		if ac := game.AlternativeCostByKey(oracle, "morph"); ac == nil || ac.FaceDown == nil || ac.ManaCost != "{3}" {
			t.Errorf("%s declares no morph cast for {3}", name)
		}
	}
}

// oldB345FaceDown casts the card from hand face down for its morph and
// resolves it; it returns the permanent.
func oldB345FaceDown(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "morph"}); err != nil {
		t.Fatalf("cast %s face down: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatalf("%s did not enter face down", name)
	}
	return id
}

// oldB345TurnUp pays the mana morph cost from conjured mana (or reveals
// `reveal`) and turns the permanent face up.
func oldB345TurnUp(t *testing.T, g *game.Game, me *game.Player, id, reveal uuid.UUID, mana ...string) {
	t.Helper()
	g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens(mana...)...) })
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true, RevealID: reveal}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	if effectstest.FindCardAnywhere(t, g, id).FaceDown {
		t.Fatal("still face down after the special action")
	}
}

func oldB345Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB345Colored(me *game.Player, color string) uuid.UUID {
	return effectstest.HandCardFull(me, "Revealed "+color, "Creature — Bear", "", "", []string{color})
}

func TestOldB345JeeringInstigatorStealsAnotherCreatureOnYourTurn(t *testing.T) {
	g, me, opp := oldB345Table(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.TapCard(g, theirs)
	id := oldB345FaceDown(t, g, "Jeering Instigator", "Creature — Goblin Rogue", oldB345Jeering, 2, 1)
	oldB345TurnUp(t, g, me, id, uuid.Nil, "R", "C", "C")
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target was asked for")
	}
	if effectstest.HasID(p.PickTargetCards, id) {
		t.Error("\"another\": the Instigator was offered as its own target")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardForTest(g, theirs)
	if c.Controller != me.ID {
		t.Fatal("control did not change")
	}
	if effectstest.TappedNow(g, theirs) {
		t.Error("the stolen creature was not untapped")
	}
	effectstest.AssertKeywords(t, g, theirs, "haste")
}

func TestOldB345JeeringInstigatorDoesNothingOnAnotherPlayersTurn(t *testing.T) {
	g, me, opp := oldB345Table(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := oldB345FaceDown(t, g, "Jeering Instigator", "Creature — Goblin Rogue", oldB345Jeering, 2, 1)
	g.WithWriteLock(func() { g.Turn.ActiveSeat = (g.Turn.ActiveSeat + 1) % len(g.Seats) })
	oldB345TurnUp(t, g, me, id, uuid.Nil, "R", "C", "C")
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("the trigger asked for a target although it is not the controller's turn (CR 603.4)")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, theirs).Controller != opp.ID {
		t.Error("control changed on an opponent's turn")
	}
}

func TestOldB345AbominationOfGudulLootsOnCombatDamage(t *testing.T) {
	g, me, opp := oldB345Table(t)
	id := effectstest.B12Push(g, me.ID, "Abomination of Gudul", "Creature — Horror", oldB345Abomin, 3, 4)
	for i := 0; i < 3; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMain(t, g)
	before := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, id)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand = %d after drawing, want %d before the discard choice", got, before+1)
	}
}

func TestOldB345PineWalkerUntapsAnyCreatureYouTurnFaceUp(t *testing.T) {
	g, me, _ := oldB345Table(t)
	effectstest.B12Push(g, me.ID, "Pine Walker", "Creature — Elemental", oldB345Pine, 5, 5)
	id := oldB345FaceDown(t, g, "Rattleclaw Mystic", "Creature — Human Shaman", oldB345Rattleclaw, 2, 1)
	effectstest.TapCard(g, id)
	oldB345TurnUp(t, g, me, id, uuid.Nil, "C", "C")
	effectstest.B30SettleOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedNow(g, id) {
		t.Error("the creature turned face up was not untapped")
	}
}

func TestOldB345RattleclawMysticAddsThreeColorsWhenTurnedFaceUp(t *testing.T) {
	g, me, _ := oldB345Table(t)
	id := oldB345FaceDown(t, g, "Rattleclaw Mystic", "Creature — Human Shaman", oldB345Rattleclaw, 2, 1)
	oldB345TurnUp(t, g, me, id, uuid.Nil, "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	got := map[string]int{}
	for _, m := range me.ManaPool {
		got[m.Color]++
	}
	if got["G"] != 1 || got["U"] != 1 || got["R"] != 1 || len(me.ManaPool) != 3 {
		t.Errorf("pool = %v, want one each of G, U and R", got)
	}
}

func TestOldB345EfreetWeaponmasterPumpsAnotherCreatureOnEnteringAndTurningUp(t *testing.T) {
	g, me, _ := oldB345Table(t)
	ally := effectstest.B12Creature(g, me.ID, "Ally Bear", "Creature — Bear", 2, 2)
	id := oldB345FaceDown(t, g, "Efreet Weaponmaster", "Creature — Efreet Monk", oldB345Efreet, 4, 3)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("the face-down Efreet has no enters trigger (CR 708.2a)")
	}
	oldB345TurnUp(t, g, me, id, uuid.Nil, "U", "R", "W", "C", "C")
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target was asked for")
	}
	if effectstest.HasID(p.PickTargetCards, id) {
		t.Error("\"another\": the Efreet was offered as its own target")
	}
	effectstest.PickCard(t, g, me.ID, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if pw, _ := effectstest.MkmB1PT(t, g, ally); pw != 5 {
		t.Errorf("power = %d, want 5", pw)
	}
}

func TestOldB345PonybackBrigadeMakesThreeGoblinsWhenTurnedFaceUp(t *testing.T) {
	g, me, _ := oldB345Table(t)
	id := oldB345FaceDown(t, g, "Ponyback Brigade", "Creature — Goblin Warrior", oldB345Pony, 2, 2)
	n := func() int {
		c := 0
		for _, x := range g.Battlefield.Cards {
			if x.Name == "Goblin" && x.IsToken() && x.Controller == me.ID {
				c++
			}
		}
		return c
	}
	if n() != 0 {
		t.Fatal("a face-down Brigade made tokens")
	}
	oldB345TurnUp(t, g, me, id, uuid.Nil, "R", "W", "B", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if n() != 3 {
		t.Errorf("goblin tokens = %d, want 3", n())
	}
}

func TestOldB345WatcherOfTheRoostRevealsAWhiteCardAndGainsTwo(t *testing.T) {
	g, me, _ := oldB345Table(t)
	id := oldB345FaceDown(t, g, "Watcher of the Roost", "Creature — Bird Soldier", oldB345Watcher, 2, 1)
	green := oldB345Colored(me, "G")
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true, RevealID: green}); err == nil {
		t.Fatal("turned face up by revealing a green card")
	}
	white := oldB345Colored(me, "W")
	life := me.Life
	oldB345TurnUp(t, g, me, id, white)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	if !me.Hand.Contains(white) {
		t.Error("the revealed card left the hand")
	}
}

func TestOldB345MistfireWeaverGivesHexproofToACreatureYouControl(t *testing.T) {
	g, me, opp := oldB345Table(t)
	ally := effectstest.B12Creature(g, me.ID, "Ally Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := oldB345FaceDown(t, g, "Mistfire Weaver", "Creature — Djinn Wizard", oldB345Mistfire, 3, 1)
	oldB345TurnUp(t, g, me, id, uuid.Nil, "U", "C", "C")
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target was asked for")
	}
	if effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("an opponent's creature was offered for \"creature you control\"")
	}
	effectstest.PickCard(t, g, me.ID, ally)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, ally, "hexproof")
}

func TestOldB345TemurChargerGivesTrampleToAnyCreature(t *testing.T) {
	g, me, opp := oldB345Table(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := oldB345FaceDown(t, g, "Temur Charger", "Creature — Horse", oldB345Charger, 3, 1)
	oldB345TurnUp(t, g, me, id, oldB345Colored(me, "G"))
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, theirs, "trample")
}

func TestOldB345DragonsEyeSavantsLooksAtAnOpponentsHand(t *testing.T) {
	g, me, opp := oldB345Table(t)
	theirCard := effectstest.HandCardFull(opp, "Secret", "Creature — Bear", "", "", nil)
	id := oldB345FaceDown(t, g, "Dragon's Eye Savants", "Creature — Human Wizard", oldB345Savants, 0, 6)
	oldB345TurnUp(t, g, me, id, oldB345Colored(me, "U"))
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	c, _, _ := func() (game.Card, int, bool) {
		x := effectstest.FindCardAnywhere(t, g, theirCard)
		return x, 0, true
	}()
	if !c.IsKnownTo(me.ID) {
		t.Error("the controller was not shown the opponent's hand")
	}
	if !opp.Hand.Contains(theirCard) {
		t.Error("the card left the hand")
	}
}

func TestOldB345RuthlessRipperMakesATargetPlayerLoseTwoWithDeathtouch(t *testing.T) {
	g, me, opp := oldB345Table(t)
	id := oldB345FaceDown(t, g, "Ruthless Ripper", "Creature — Human Assassin", oldB345Ripper, 1, 1)
	if effectstest.HasKeywordOnBattlefield(t, g, id, "deathtouch") {
		t.Error("the face-down Ripper has deathtouch (CR 708.2a)")
	}
	life := opp.Life
	oldB345TurnUp(t, g, me, id, oldB345Colored(me, "B"))
	effectstest.AssertKeywords(t, g, id, "deathtouch")
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want -2", life, opp.Life)
	}
}

func TestOldB345HordeAmbusherStopsATargetBlockingWhenTurnedFaceUp(t *testing.T) {
	g, me, opp := oldB345Table(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.B12Push(g, me.ID, "Horde Ambusher", "Creature — Human Berserker", oldB345Ambusher, 2, 2)
	fd := oldB345FaceDown(t, g, "Horde Ambusher", "Creature — Human Berserker", oldB345Ambusher, 2, 2)
	oldB345TurnUp(t, g, me, fd, oldB345Colored(me, "R"))
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.DeclareAttack(t, g, opp.ID, id)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(theirs, id); err == nil {
		t.Error("the creature that can't block blocked")
	}
}

func TestOldB345HordeAmbusherHurtsItsControllerWhenItBlocks(t *testing.T) {
	g, me, opp := oldB345Table(t)
	attacker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	amb := effectstest.B12Push(g, me.ID, "Horde Ambusher", "Creature — Human Berserker", oldB345Ambusher, 2, 2)
	other := effectstest.B12Push(g, me.ID, "Bystander", "Creature — Bear", "", 2, 2)
	life := me.Life
	effectstest.WoeS2Block(g, me.ID, other, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("another creature's block hurt the Ambusher's controller (%d → %d)", life, me.Life)
	}
	effectstest.WoeS2Block(g, me.ID, amb, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
}
