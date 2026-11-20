package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsv "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/v"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vengeful_strangler_test.go — Vengeful Strangler // Strangling Grasp
// (#5677), and the engine door "return it to the battlefield transformed
// attached to …" (Game.ReturnFromGraveyardTransformedAttachedForEffect).

func vengefulStranglerRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Vengeful Strangler // Strangling Grasp",
		Layout:   "transform",
		TypeLine: "Creature — Human Rogue // Enchantment — Aura",
		OracleID: uuid.MustParse(cardsv.VengefulStranglerOracleID),
		CardFaces: []cards.CardFace{
			{Name: "Vengeful Strangler", TypeLine: "Creature — Human Rogue", ManaCost: "{1}{B}", Colors: []string{"B"}, Power: "2", Toughness: "1"},
			{Name: "Strangling Grasp", TypeLine: "Enchantment — Aura", Colors: []string{"B"}},
		},
	}
}

// --- Vengeful Strangler -------------------------------------------

func TestVengefulStranglerCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	strangler := effectstest.ImportToBattlefield(t, g, vengefulStranglerRow(), me)
	c := effectstest.AdCard(t, g, strangler)
	if c.Effective().Restrictions&game.CantBlock == 0 {
		t.Error("Vengeful Strangler can block")
	}
	if c.Effective().Power != 2 || c.Effective().Toughness != 1 {
		t.Errorf("it is %d/%d, want 2/1", c.Effective().Power, c.Effective().Toughness)
	}
}

// It dies, targets a creature or planeswalker an opponent controls
// (not one of its controller's), and comes back as Strangling Grasp
// attached to it under its controller's control.
func TestVengefulStranglerReturnsAsStranglingGraspOnTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.RtaBear(g, me.ID, "My Bear")
	theirs := effectstest.RtaBear(g, opp.ID, "Their Bear")
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 4}, Owner: opp.ID, Controller: opp.ID,
	})
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	strangler := effectstest.ImportToBattlefield(t, g, vengefulStranglerRow(), me)

	effectstest.RtaDestroy(t, g, strangler)
	prompt := effectstest.RtaWaitForTarget(t, g, me.ID)
	if !effectstest.HasID(prompt.PickTargetCards, theirs) || !effectstest.HasID(prompt.PickTargetCards, walker) {
		t.Errorf("an opponent's creature and planeswalker are both targets: %v", prompt.PickTargetCards)
	}
	if effectstest.HasID(prompt.PickTargetCards, mine) || effectstest.HasID(prompt.PickTargetCards, rock) {
		t.Errorf("its controller's creature and an opponent's artifact are not targets: %v", prompt.PickTargetCards)
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, effectstest.CardRef(mine)); err == nil {
		t.Fatal("targeting your own creature was accepted")
	}
	effectstest.PickCard(t, g, me.ID, walker)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.AuraHostChoiceFor(g, me.ID) != nil {
		t.Fatal("the effect names the host: nothing is chosen as it enters")
	}
	c := effectstest.AdCard(t, g, strangler)
	if effectstest.AdZone(g, strangler) != game.ZoneBattlefield || c.ActiveFace != 1 || c.Name != "Strangling Grasp" || !c.IsAura() || c.IsCreature() {
		t.Fatalf("returned as %q (%s) face %d in %v, want the Aura Strangling Grasp", c.Name, c.TypeLine, c.ActiveFace, effectstest.AdZone(g, strangler))
	}
	if !c.IsAttachedTo(walker) || c.Controller != me.ID {
		t.Errorf("attached to %+v under %v, want the planeswalker under its controller", c.AttachedTo, c.Controller)
	}
}

// With nothing to target the trigger is removed and the card stays in
// the graveyard, on its front face.
func TestVengefulStranglerWithNoTargetStaysInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.RtaBear(g, me.ID, "My Bear")
	strangler := effectstest.ImportToBattlefield(t, g, vengefulStranglerRow(), me)
	effectstest.RtaDestroy(t, g, strangler)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.AdCard(t, g, strangler)
	if effectstest.AdZone(g, strangler) != game.ZoneGraveyard || c.ActiveFace != 0 || c.Name != "Vengeful Strangler" {
		t.Fatalf("%q face %d in %v, want Vengeful Strangler in the graveyard", c.Name, c.ActiveFace, effectstest.AdZone(g, strangler))
	}
}

// rtaGraspOn kills a Vengeful Strangler of p's and returns it as
// Strangling Grasp on host.
func rtaGraspOn(t *testing.T, g *game.Game, p *game.Player, host uuid.UUID) uuid.UUID {
	t.Helper()
	strangler := effectstest.ImportToBattlefield(t, g, vengefulStranglerRow(), p)
	effectstest.RtaDestroy(t, g, strangler)
	effectstest.RtaWaitForTarget(t, g, p.ID)
	effectstest.PickCard(t, g, p.ID, host)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.AdCard(t, g, strangler); !c.IsAttachedTo(host) {
		t.Fatalf("setup: Strangling Grasp is attached to %+v", c.AttachedTo)
	}
	return strangler
}

// At its controller's upkeep the enchanted permanent's controller
// chooses a nonland permanent of theirs to sacrifice, then loses 1
// life. Lands are not offered, and nobody else's upkeep triggers it.
func TestStranglingGraspUpkeepSacrificeThenLoseLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.RtaBear(g, opp.ID, "Their Bear")
	spare := effectstest.RtaBear(g, opp.ID, "Their Spare Bear")
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Land", TypeLine: "Basic Land — Swamp", Owner: opp.ID, Controller: opp.ID,
	})
	myBear := effectstest.RtaBear(g, me.ID, "My Bear")
	grasp := rtaGraspOn(t, g, me, theirs)

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TriggerOnStack(g, grasp) != nil {
		t.Fatal("it triggered at another player's upkeep")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if effectstest.TriggerOnStack(g, grasp) == nil {
		t.Fatal("it did not trigger at its controller's upkeep")
	}
	mineBefore, theirsBefore := me.Life, opp.Life
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the enchanted permanent's controller is not asked: %+v", g.PendingChoices)
	}
	if opp.Life != theirsBefore {
		t.Error("the life was lost before the sacrifice")
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{land}); err == nil {
		t.Fatal("a land was accepted as the nonland permanent")
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{myBear}); err == nil {
		t.Fatal("another player's permanent was accepted")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, spare)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(spare) {
		t.Error("the chosen permanent was not sacrificed")
	}
	if !g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(land) || !g.Battlefield.Contains(myBear) {
		t.Error("something else left the battlefield")
	}
	if opp.Life != theirsBefore-1 || me.Life != mineBefore {
		t.Errorf("life: enchanted player %d → %d (want -1), controller %d → %d", theirsBefore, opp.Life, mineBefore, me.Life)
	}
	if effectstest.AdZone(g, grasp) != game.ZoneBattlefield {
		t.Error("Strangling Grasp left the battlefield")
	}
}

// Sacrificing the enchanted permanent itself is allowed; the Aura then
// goes to its owner's graveyard as Vengeful Strangler (CR 712.8a), and
// the life is still lost.
func TestStranglingGraspEnchantedPermanentMayBeTheSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.RtaBear(g, opp.ID, "Their Bear")
	grasp := rtaGraspOn(t, g, me, theirs)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	before := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) {
		t.Fatal("the enchanted creature was not sacrificed")
	}
	if opp.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, opp.Life)
	}
	c := effectstest.AdCard(t, g, grasp)
	if effectstest.AdZone(g, grasp) != game.ZoneGraveyard || !me.Graveyard.Contains(grasp) || c.ActiveFace != 0 || c.Name != "Vengeful Strangler" {
		t.Errorf("the Aura: %q face %d in %v, want Vengeful Strangler in its owner's graveyard", c.Name, c.ActiveFace, effectstest.AdZone(g, grasp))
	}
}

// "An opponent controls" is part of the enchant ability: if the Aura's
// controller gains control of the enchanted permanent, the Aura is put
// into the graveyard (CR 704.5m).
func TestStranglingGraspFallsOffAPermanentItsControllerControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.RtaBear(g, opp.ID, "Their Bear")
	grasp := rtaGraspOn(t, g, me, theirs)
	g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, theirs, me.ID, game.IndefiniteDuration(), "theft") })
	g.RunStateChecksForTest()
	if effectstest.AdZone(g, grasp) != game.ZoneGraveyard {
		t.Errorf("Strangling Grasp is in %v, want the graveyard", effectstest.AdZone(g, grasp))
	}
}

// --- the engine door ----------------------------------------------

// CR 303.4i: a host the back face can't legally enchant leaves the card
// in the graveyard on its front face, with no entry and no prompt.
func TestReturnTransformedAttachedToAnIllegalHostStaysInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.RtaBear(g, me.ID, "My Bear")
	effectstest.RtaBear(g, opp.ID, "Their Bear")
	card := effectstest.ImportToHand(vengefulStranglerRow(), me)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, card); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}

	for name, host := range map[string]game.TargetRef{
		"a creature its controller controls": effectstest.CardRef(mine),
		"a player":                           {Kind: game.TargetPlayer, ID: opp.ID},
		"an object that is gone":             effectstest.CardRef(uuid.New()),
	} {
		var got uuid.UUID
		var err error
		g.WithWriteLock(func() {
			got, err = g.ReturnFromGraveyardTransformedAttachedForEffect(card, me.ID, host)
		})
		if err != nil || got != uuid.Nil {
			t.Fatalf("%s: got (%v, %v), want (Nil, nil)", name, got, err)
		}
		c := effectstest.AdCard(t, g, card)
		if effectstest.AdZone(g, card) != game.ZoneGraveyard || c.ActiveFace != 0 || c.Name != "Vengeful Strangler" {
			t.Fatalf("%s: %q face %d in %v, want the front face in the graveyard", name, c.Name, c.ActiveFace, effectstest.AdZone(g, card))
		}
		if effectstest.AuraHostChoiceFor(g, me.ID) != nil {
			t.Fatalf("%s: a host was asked for", name)
		}
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventETB && ev.CardID == card {
			t.Fatal("it entered")
		}
	}
}

// A legal host: back face up, attached, under the named controller,
// though the card's owner is somebody else.
func TestReturnTransformedAttachedUnderAnotherPlayersControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.RtaBear(g, me.ID, "My Bear")
	card := effectstest.ImportToHand(vengefulStranglerRow(), other)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: other.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: other.ID}, card); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	var got uuid.UUID
	var err error
	g.WithWriteLock(func() {
		got, err = g.ReturnFromGraveyardTransformedAttachedForEffect(card, opp.ID, effectstest.CardRef(mine))
	})
	if err != nil || got != card {
		t.Fatalf("got (%v, %v), want the card", got, err)
	}
	c := effectstest.AdCard(t, g, card)
	if c.Controller != opp.ID || c.Owner != other.ID || !c.IsAttachedTo(mine) || c.ActiveFace != 1 {
		t.Errorf("controller %v owner %v host %+v face %d", c.Controller, c.Owner, c.AttachedTo, c.ActiveFace)
	}
}
