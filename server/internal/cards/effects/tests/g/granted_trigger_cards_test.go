package testsg

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Agent of the Iron Throne", "Legendary Enchantment — Background", effectstest.B21AgentOfTheIronThroneOracle, 0, 0, "B")
	commander := effectstest.B21Commander(g, me.ID, "My Commander", "Legendary Creature — Human Knight", 3, 3)
	effectstest.PushAuraOnLand(t, g, opp.ID, "Mind Control", effectstest.MindControlOracle, commander)
	if got := effectstest.ControllerOf(t, g, commander); got != opp.ID {
		t.Fatalf("setup: the commander should be the opponent's now, controller %s", got)
	}
	if effectstest.GrantedTriggerCount(g, commander) != 1 {
		t.Fatal("a stolen commander you own still has the Background's ability")
	}
	before := effectstest.B17Life(g)
	theirBear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B18Kill(t, g, theirBear)
	if me.Life != before[0]-1 {
		t.Errorf("the thief's creature died: you are the thief's opponent, life %d want %d", me.Life, before[0]-1)
	}
	if opp.Life != before[1] {
		t.Errorf("the thief is not their own opponent: life %d want %d", opp.Life, before[1])
	}
	myBear := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B18Kill(t, g, myBear)
	if opp.Life != before[1] {
		t.Error("your creature dying is not an artifact or creature the ability's controller controls")
	}
}

// Dionus's grant is each Elf's own ability: an Elf that LATER loses
// all its abilities loses it (CR 613.6), and the other Elves keep it.
func TestDionusGrantIsLostToALaterAbilityRemoval(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Dionus, Elvish Archdruid", "Legendary Creature — Elf Druid", effectstest.B23DionusOracle, 3, 3, "G")
	elf := effectstest.B21Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, 1, 1, "G")
	other := effectstest.B21Push(g, me.ID, "Fyndhorn Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, 1, 1, "G")
	if effectstest.GrantedTriggerCount(g, elf) != 1 {
		t.Fatal("setup: an Elf you control has Dionus's ability")
	}
	effectstest.PushAuraOnLand(t, g, me.ID, "Darksteel Mutation", effectstest.DarksteelMutationOracle, elf)
	if effectstest.GrantedTriggerCount(g, elf) != 0 {
		t.Error("a later ability removal takes the granted trigger")
	}
	if effectstest.GrantedTriggerCount(g, other) != 1 {
		t.Error("the other Elf keeps it")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B23TapMana(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, other) || effectstest.CounterCount(g, other, "+1/+1") != 1 {
		t.Error("the Elf that kept the ability untaps and grows")
	}
}

// Dionus's trigger belongs to the Elf's controller: an Elf stolen by
// an opponent is no longer an Elf YOU control, so it has no grant from
// your Dionus — and a Dionus the opponent steals grants to THEIR Elves.
func TestDionusGrantFollowsControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dionus := effectstest.B21Push(g, me.ID, "Dionus, Elvish Archdruid", "Legendary Creature — Elf Druid", effectstest.B23DionusOracle, 3, 3, "G")
	mine := effectstest.B21Push(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, 1, 1, "G")
	theirs := effectstest.B21Push(g, opp.ID, "Their Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, 1, 1, "G")
	if effectstest.GrantedTriggerCount(g, mine) != 1 || effectstest.GrantedTriggerCount(g, theirs) != 0 {
		t.Fatal("setup: your Dionus grants to your Elves only")
	}
	effectstest.PushAuraOnLand(t, g, opp.ID, "Mind Control", effectstest.MindControlOracle, dionus)
	if effectstest.GrantedTriggerCount(g, mine) != 0 || effectstest.GrantedTriggerCount(g, theirs) != 1 {
		t.Error("a Dionus the opponent controls grants to the opponent's Elves")
	}
}

// Thornbite Staff: the equipped creature pings, taps and is untapped
// by a death; the ping is the host's damage.
func TestThornbiteStaffGrantsThePingAndTheUntap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shaman := effectstest.B16Creature(g, me.ID, "Shaman", "Creature — Goblin Shaman", 1, 1, "R")
	staff := effectstest.B21Push(g, me.ID, "Thornbite Staff", "Kindred Artifact — Shaman Equipment", thornbiteStaffOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	effectstest.EquipTo(t, g, me.ID, staff, shaman)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, shaman)
	if idx < 0 {
		t.Fatal("the equipped creature has the ping")
	}
	if effectstest.GrantedTriggerCount(g, shaman) != 1 {
		t.Fatal("the equipped creature has the untap trigger")
	}
	before := opp.Life
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, shaman, idx, game.ActivateAbilityParams{
		Ref:     ref,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if opp.Life != before-1 {
		t.Errorf("the ping: life %d want %d", opp.Life, before-1)
	}
	if !effectstest.B16Tapped(t, g, shaman) {
		t.Fatal("the ping taps the equipped creature")
	}
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B18Kill(t, g, bear)
	if effectstest.B16Tapped(t, g, shaman) {
		t.Error("a creature died: the equipped creature untaps")
	}
}

// "Whenever a Shaman creature enters, you may attach this Equipment to
// it" — a real "you may".
func TestThornbiteStaffMayAttachToAnEnteringShaman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	staff := effectstest.B21Push(g, me.ID, "Thornbite Staff", "Kindred Artifact — Shaman Equipment", thornbiteStaffOracle, 0, 0)
	// A real entry, so EventETB fires.
	card := game.NewCard("Shaman", me.ID)
	card.TypeLine, card.Power, card.Toughness = "Creature — Elf Shaman", 1, 1
	shaman := card.InstanceID
	me.Hand.PushTop(card)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, shaman); err != nil {
		t.Fatalf("put the Shaman onto the battlefield: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, staff); host.ID != shaman {
		t.Errorf("the Staff is attached to %v, want the Shaman", host)
	}
	if effectstest.GrantedTriggerCount(g, shaman) != 1 {
		t.Error("attached, the Shaman has the Staff's granted trigger")
	}
}
