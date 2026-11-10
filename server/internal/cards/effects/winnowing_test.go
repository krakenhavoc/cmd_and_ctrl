package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const winnowingOracle = "8b834f83-951b-4b5c-b9f1-3d860baa3c0e"

func TestWinnowingIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(winnowingOracle)
	if !ok || spec.Name != "Winnowing" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
	if spec.Purpose.Sweep.How != game.SweepSacrifice || spec.Purpose.Sweep.Matches != game.SweepCreatures {
		t.Errorf("purpose %+v, want a creature sacrifice sweep", spec.Purpose)
	}
}

// winnowingCreature parks a creature with a type line (and keywords) for
// the shared-type read.
func winnowingCreature(g *game.Game, owner uuid.UUID, name, typeLine string, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Keywords: keywords,
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

// Winnowing is cast with convoke; the caster picks one creature for each
// player, and each player loses the creatures that share no creature
// type with their chosen one. A changeling shares with anything typed; a
// creature with no creature type shares nothing; lands stay.
func TestWinnowingKeepsWhatSharesATypeWithTheChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	soldiers := pushTapCostSoldiers(g, me.ID, 4)
	myElf := winnowingCreature(g, me.ID, "My Elf Warrior", "Creature — Elf Warrior")
	myDruid := winnowingCreature(g, me.ID, "My Elf Druid", "Creature — Elf Druid")
	myChangeling := winnowingCreature(g, me.ID, "My Shapeshifter", "Creature — Shapeshifter", "changeling")
	myGoblin := winnowingCreature(g, me.ID, "My Goblin", "Creature — Goblin")
	theirGoblin := winnowingCreature(g, opp.ID, "Their Goblin", "Creature — Goblin")
	theirShaman := winnowingCreature(g, opp.ID, "Their Goblin Shaman", "Creature — Goblin Shaman")
	theirElf := winnowingCreature(g, opp.ID, "Their Elf Warrior", "Creature — Elf Warrior")
	theirGolem := winnowingCreature(g, opp.ID, "Their Typeless", "Artifact Creature")
	var land uuid.UUID
	g.WithWriteLock(func() { land = pushNamedPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest") })

	fillPoolColored(me, "W", 2)
	castFromHandForTest(t, g, me, "Winnowing", "Sorcery", "{4}{W}{W}", winnowingOracle,
		game.CastSpellParams{Strict: true, TapIDs: soldiers})
	passPriorityAroundTable(t, g)

	mine := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	theirs := latestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID)
	if mine == nil || theirs == nil {
		t.Fatalf("want one leg for each player with creatures, all answered by the caster: %+v", g.PendingChoices)
	}
	if mine.ChooseMin != 1 || mine.ChooseMax != 1 || theirs.ChooseMin != 1 || theirs.ChooseMax != 1 {
		t.Errorf("each leg chooses exactly one creature: %d..%d and %d..%d", mine.ChooseMin, mine.ChooseMax, theirs.ChooseMin, theirs.ChooseMax)
	}
	if err := g.ResolveTheirPermanents(theirs.ID, me.ID, []uuid.UUID{land}); err == nil {
		t.Error("a land was accepted as the chosen creature")
	}
	for _, c := range append([]*game.PendingChoice(nil), g.PendingChoices...) {
		if c == nil {
			continue
		}
		switch c.Kind {
		case game.PendingChoiceOwnPermanents:
			if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{myElf}); err != nil {
				t.Fatalf("ResolveOwnPermanents: %v", err)
			}
		case game.PendingChoiceTheirPermanents:
			if c.FromPlayer != opp.ID {
				t.Fatalf("an unexpected leg over %s", c.FromPlayer)
			}
			if err := g.ResolveTheirPermanents(c.ID, me.ID, []uuid.UUID{theirGoblin}); err != nil {
				t.Fatalf("ResolveTheirPermanents: %v", err)
			}
		}
	}

	for id, want := range map[uuid.UUID]bool{
		myElf: true, myDruid: true, myChangeling: true, myGoblin: false,
		theirGoblin: true, theirShaman: true, theirElf: false, theirGolem: false, land: true,
		soldiers[0]: false, soldiers[3]: false,
	} {
		if got := g.Battlefield.Contains(id); got != want {
			t.Errorf("%s on the battlefield = %v, want %v", cardNameOf(g, id), got, want)
		}
	}
}

// Without the creatures to convoke, two white mana can't cast it.
func TestWinnowingNeedsItsWholeCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	fillPoolColored(me, "W", 2)
	id := rfdHandCard(me, "Winnowing", "Sorcery", "{4}{W}{W}", winnowingOracle, 0, 0)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("Winnowing was cast for {W}{W} with nothing to convoke")
	}
}
