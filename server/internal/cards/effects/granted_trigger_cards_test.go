package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_trigger_cards_test.go — ADR 0093 PR 3: granted TRIGGERED
// abilities on real cards, and the two behaviour changes the ADR says
// its PR 3 must pin — after a control change the trigger belongs to the
// recipient's controller, and a recipient whose own abilities are
// removed later loses the grant.

const thornbiteStaffOracle = "dae4815e-9025-4993-ab46-52a3f1a7219e"

// A commander an opponent has stolen still carries the Background's
// grant (you still own it), and the trigger is the THIEF's: the
// thief's creature dying drains the thief's opponents, you among
// them; your creature dying drains nobody.
func TestAgentOfTheIronThroneStolenCommanderDrainsForTheThief(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Agent of the Iron Throne", "Legendary Enchantment — Background", b21AgentOfTheIronThroneOracle, 0, 0, "B")
	commander := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human Knight", 3, 3)
	pushAuraOnLand(t, g, opp.ID, "Mind Control", mindControlOracle, commander)
	if got := controllerOf(t, g, commander); got != opp.ID {
		t.Fatalf("setup: the commander should be the opponent's now, controller %s", got)
	}
	if grantedTriggerCount(g, commander) != 1 {
		t.Fatal("a stolen commander you own still has the Background's ability")
	}
	before := b17Life(g)
	theirBear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	b18Kill(t, g, theirBear)
	if me.Life != before[0]-1 {
		t.Errorf("the thief's creature died: you are the thief's opponent, life %d want %d", me.Life, before[0]-1)
	}
	if opp.Life != before[1] {
		t.Errorf("the thief is not their own opponent: life %d want %d", opp.Life, before[1])
	}
	myBear := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	b18Kill(t, g, myBear)
	if opp.Life != before[1] {
		t.Error("your creature dying is not an artifact or creature the ability's controller controls")
	}
}

// Dionus's grant is each Elf's own ability: an Elf that LATER loses
// all its abilities loses it (CR 613.6), and the other Elves keep it.
func TestDionusGrantIsLostToALaterAbilityRemoval(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Dionus, Elvish Archdruid", "Legendary Creature — Elf Druid", b23DionusOracle, 3, 3, "G")
	elf := b21Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, 1, 1, "G")
	other := b21Push(g, me.ID, "Fyndhorn Elves", "Creature — Elf Druid", llanowarElvesOracle, 1, 1, "G")
	if grantedTriggerCount(g, elf) != 1 {
		t.Fatal("setup: an Elf you control has Dionus's ability")
	}
	pushAuraOnLand(t, g, me.ID, "Darksteel Mutation", darksteelMutationOracle, elf)
	if grantedTriggerCount(g, elf) != 0 {
		t.Error("a later ability removal takes the granted trigger")
	}
	if grantedTriggerCount(g, other) != 1 {
		t.Error("the other Elf keeps it")
	}
	advanceToMain(t, g)
	b23TapMana(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, other) || counterCount(g, other, "+1/+1") != 1 {
		t.Error("the Elf that kept the ability untaps and grows")
	}
}

// Dionus's trigger belongs to the Elf's controller: an Elf stolen by
// an opponent is no longer an Elf YOU control, so it has no grant from
// your Dionus — and a Dionus the opponent steals grants to THEIR Elves.
func TestDionusGrantFollowsControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dionus := b21Push(g, me.ID, "Dionus, Elvish Archdruid", "Legendary Creature — Elf Druid", b23DionusOracle, 3, 3, "G")
	mine := b21Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, 1, 1, "G")
	theirs := b21Push(g, opp.ID, "Their Elves", "Creature — Elf Druid", llanowarElvesOracle, 1, 1, "G")
	if grantedTriggerCount(g, mine) != 1 || grantedTriggerCount(g, theirs) != 0 {
		t.Fatal("setup: your Dionus grants to your Elves only")
	}
	pushAuraOnLand(t, g, opp.ID, "Mind Control", mindControlOracle, dionus)
	if grantedTriggerCount(g, mine) != 0 || grantedTriggerCount(g, theirs) != 1 {
		t.Error("a Dionus the opponent controls grants to the opponent's Elves")
	}
}

// Thornbite Staff: the equipped creature pings, taps and is untapped
// by a death; the ping is the host's damage.
func TestThornbiteStaffGrantsThePingAndTheUntap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := b16Creature(g, me.ID, "Shaman", "Creature — Goblin Shaman", 1, 1, "R")
	staff := b21Push(g, me.ID, "Thornbite Staff", "Kindred Artifact — Shaman Equipment", thornbiteStaffOracle, 0, 0)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}{C}")
	equipTo(t, g, me.ID, staff, shaman)
	idx, ref := grantedActivatedIndex(t, g, shaman)
	if idx < 0 {
		t.Fatal("the equipped creature has the ping")
	}
	if grantedTriggerCount(g, shaman) != 1 {
		t.Fatal("the equipped creature has the untap trigger")
	}
	before := opp.Life
	floatMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, shaman, idx, game.ActivateAbilityParams{
		Ref:     ref,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if opp.Life != before-1 {
		t.Errorf("the ping: life %d want %d", opp.Life, before-1)
	}
	if !b16Tapped(t, g, shaman) {
		t.Fatal("the ping taps the equipped creature")
	}
	bear := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	b18Kill(t, g, bear)
	if b16Tapped(t, g, shaman) {
		t.Error("a creature died: the equipped creature untaps")
	}
}

// "Whenever a Shaman creature enters, you may attach this Equipment to
// it" — a real "you may".
func TestThornbiteStaffMayAttachToAnEnteringShaman(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	staff := b21Push(g, me.ID, "Thornbite Staff", "Kindred Artifact — Shaman Equipment", thornbiteStaffOracle, 0, 0)
	// A real entry, so EventETB fires.
	card := game.NewCard("Shaman", me.ID)
	card.TypeLine, card.Power, card.Toughness = "Creature — Elf Shaman", 1, 1
	shaman := card.InstanceID
	me.Hand.PushTop(card)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, shaman); err != nil {
		t.Fatalf("put the Shaman onto the battlefield: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, staff); host.ID != shaman {
		t.Errorf("the Staff is attached to %v, want the Shaman", host)
	}
	if grantedTriggerCount(g, shaman) != 1 {
		t.Error("attached, the Shaman has the Staff's granted trigger")
	}
}

// grantedTriggerCount counts the GRANTED abilities a permanent lists
// that no activated or mana row accounts for — its granted triggers.
func grantedTriggerCount(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return
		}
		for _, info := range game.GrantedAbilitiesOf(c) {
			if len(game.CatalogTriggers(info.Key)) > 0 {
				n++
			}
		}
	})
	return n
}
