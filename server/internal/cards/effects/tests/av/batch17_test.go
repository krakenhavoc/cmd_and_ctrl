package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b17Hands reads every seat's hand size.
func b17Hands(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		out = append(out, p.Hand.Size())
	}
	return out
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Boros
// Guildgate is a row in the Guildgate table, so a transposed row is
// invisible until someone plays that exact card. Korvold and
// Evacuation were already on main and are not in this table.
func TestBatch17CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B17JaradOracle:                "Jarad, Golgari Lich Lord",
		effectstest.B17TormodsCryptOracle:         "Tormod's Crypt",
		effectstest.B17BorosGuildgateOracle:       "Boros Guildgate",
		effectstest.B17BattleOfBywaterOracle:      "The Battle of Bywater",
		effectstest.B17DeepglowSkateOracle:        "Deepglow Skate",
		effectstest.B17ClaimJumperOracle:          "Claim Jumper",
		effectstest.B17UndeadAugurOracle:          "Undead Augur",
		effectstest.B17MultaniOracle:              "Multani, Yavimaya's Avatar",
		effectstest.B17ThreefoldThunderhulkOracle: "Threefold Thunderhulk",
		effectstest.B17FiremaneCommandoOracle:     "Firemane Commando",
		effectstest.B17ContainmentConstructOracle: "Containment Construct",
		effectstest.B17RazakethOracle:             "Razaketh, the Foulblooded",
		effectstest.B17CordialVampireOracle:       "Cordial Vampire",
		effectstest.B17ArwenOracle:                "Arwen, Weaver of Hope",
		effectstest.B17DictateOfKruphixOracle:     "Dictate of Kruphix",
		effectstest.B17KazuulOracle:               "Kazuul, Tyrant of the Cliffs",
		effectstest.B17PhyrexianObliteratorOracle: "Phyrexian Obliterator",
		effectstest.B17BreenaOracle:               "Breena, the Demagogue",
		effectstest.B17NighthawkScavengerOracle:   "Nighthawk Scavenger",
		effectstest.B17HarvesterOfSoulsOracle:     "Harvester of Souls",
		effectstest.B17TempleBellOracle:           "Temple Bell",
		effectstest.B17ColossalGraveReaverOracle:  "Colossal Grave-Reaver",
		effectstest.B17JacesArchivistOracle:       "Jace's Archivist",
		effectstest.B17InsidiousFungusOracle:      "Insidious Fungus",
		effectstest.B17VileEntomberOracle:         "Vile Entomber",
		effectstest.B17RefuteOracle:               "Refute",
		effectstest.B17LeafCrownedVisionaryOracle: "Leaf-Crowned Visionary",
		effectstest.B17TesharOracle:               "Teshar, Ancestor's Apostle",
		effectstest.B17RiseOfTheWitchKingOracle:   "Rise of the Witch-king",
	}
	if len(want) != 29 {
		t.Fatalf("the batch registers 29 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- the land and the artifacts ------------------------------------

func TestB17BorosGuildgateEntersTappedAndTapsForRedOrWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	gate := effectstest.PlayLandFromHand(t, g, "Boros Guildgate", effectstest.B17BorosGuildgateOracle)
	effectstest.Top100AssertEnteredTapped(t, g, gate, "Boros Guildgate")
	if effectstest.TapEventsFor(g, gate) != 0 {
		t.Error("enters-tapped is a replacement, not a tap")
	}
	spec, _ := effects.Lookup(effectstest.B17BorosGuildgateOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{R|W}" {
		t.Errorf("mana abilities %+v, want one producing {R|W}", spec.ManaAbilities)
	}
}

func TestB17TormodsCryptExilesATargetPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	crypt := effectstest.B12Push(g, me.ID, "Tormod's Crypt", "Artifact", effectstest.B17TormodsCryptOracle, 0, 0)
	a := effectstest.PushGraveyardCardForTest(opp, "Dead A")
	b := effectstest.PushGraveyardCardForTest(opp, "Dead B")
	c := effectstest.PushGraveyardCardForTest(other, "Dead C")
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, crypt, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetPlayer(opp.ID)})
	if g.Battlefield.Contains(crypt) {
		t.Error("the Crypt is sacrificed as a cost")
	}
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) || opp.Graveyard.Size() != 0 {
		t.Error("the target player's whole graveyard is exiled")
	}
	if !other.Graveyard.Contains(c) {
		t.Error("another player's graveyard is untouched")
	}
}

func TestB17TempleBellDrawsForEachPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bell := effectstest.B12Push(g, me.ID, "Temple Bell", "Artifact", effectstest.B17TempleBellOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	before := b17Hands(g)
	effectstest.B16Activate(t, g, me.ID, bell, 0, game.ActivateAbilityParams{})
	for i, p := range g.Seats {
		if p.Hand.Size() != before[i]+1 {
			t.Errorf("seat %d drew %d, want 1", i, p.Hand.Size()-before[i])
		}
	}
	if !effectstest.B16Tapped(t, g, bell) {
		t.Error("the Bell taps as its cost")
	}
}

func TestB17JacesArchivistWheelsToTheGreatestDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archivist := effectstest.PushCatalogPermanent(g, me.ID, "Jace's Archivist", "Creature — Vedalken Wizard", effectstest.B17JacesArchivistOracle, false)
	effectstest.AdvanceToMain(t, g)
	// Hands of 8 (me — seat 0 drew for turn 1, CR 103.8c at four
	// seats), 7 (opp) and 7, 7: pad the opponent's to 9 so the
	// greatest discard is theirs.
	for i := 0; i < 2; i++ {
		effectstest.HandCardFull(opp, "Extra", "Instant", "", "", nil)
	}
	effectstest.B06AddMana(me, "U")
	effectstest.B16Activate(t, g, me.ID, archivist, 0, game.ActivateAbilityParams{})
	for i, p := range g.Seats {
		if p.Hand.Size() != 9 {
			t.Errorf("seat %d holds %d after the wheel, want 9 (the greatest discard)", i, p.Hand.Size())
		}
	}
	if opp.Graveyard.Size() != 9 || me.Graveyard.Size() != 8 {
		t.Errorf("discards: me %d opp %d, want 8 and 9", me.Graveyard.Size(), opp.Graveyard.Size())
	}
}

// --- the spells ----------------------------------------------------

func TestB17RefuteCountersASpellAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetPlayer(me.ID))
	effectstest.CastCatalogSpell(t, g, "Refute", "Instant", effectstest.B17RefuteOracle, effectstest.B16TargetCard(bolt))
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != 40 {
		t.Fatal("the Bolt is countered")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d, want 1", me.Hand.Size()-hand)
	}
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", effectstest.DiscardOwed(g, me.ID))
	}
}

func TestB17BattleOfBywaterKillsBigCreaturesAndFeedsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B16Creature(g, me.ID, "Hobbit", "Creature — Halfling", 1, 1, "W")
	big := effectstest.B16Creature(g, me.ID, "Ent", "Creature — Treefolk", 4, 4, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 5, 5, "R")
	theirSmall := effectstest.B16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 2, 2, "R")
	darksteel := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Darksteel Colossus", TypeLine: "Artifact Creature — Golem",
		Power: 11, Toughness: 11, Keywords: []string{"indestructible"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "The Battle of Bywater", "Sorcery", effectstest.B17BattleOfBywaterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || g.Battlefield.Contains(theirs) {
		t.Error("creatures with power 3 or greater are destroyed")
	}
	if !g.Battlefield.Contains(small) || !g.Battlefield.Contains(theirSmall) {
		t.Error("smaller creatures survive")
	}
	if !g.Battlefield.Contains(darksteel) {
		t.Error("an indestructible creature survives (#446 — DestroyAllMatching honours it)")
	}
	foods := effectstest.BattlefieldIDsNamed(g, "Food")
	if len(foods) != 2 {
		t.Fatalf("two creatures you control survived: %d Foods, want 2", len(foods))
	}
	for _, id := range foods {
		if effectstest.ControllerOf(t, g, id) != me.ID {
			t.Error("the Foods are the caster's")
		}
	}
}

// A commander of yours with power 3 or more is destroyed by The Battle
// of Bywater, so it is not a creature you control when the Foods are
// counted. Since ADR 0115 it is in its owner's graveyard when the
// Foods are made, and the CR 903.9a question that follows changes
// nothing about the count, whichever answer it gets.
func TestB17BattleOfBywaterDoesNotFeedADestroyedCommander(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{{"to the command zone", true}, {"left in the graveyard", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			commander := effectstest.B36Commander(g, me.ID, "My Commander") // a 3/3
			hobbit := effectstest.B16Creature(g, me.ID, "Hobbit", "Creature — Halfling", 1, 1, "W")

			effectstest.CastCatalogSpell(t, g, "The Battle of Bywater", "Sorcery", effectstest.B17BattleOfBywaterOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(hobbit) {
				t.Fatal("the Hobbit is too small to be destroyed")
			}
			if !me.Graveyard.Contains(commander) {
				t.Fatal("the destroyed commander is not in its owner's graveyard")
			}
			if n := len(effectstest.BattlefieldIDsNamed(g, "Food")); n != 1 {
				t.Errorf("%d Foods before the CR 903.9a answer, want 1 (only the Hobbit)", n)
			}

			effectstest.AnswerCommanderReturn(t, g, me.ID, tc.commandZone)
			if tc.commandZone && !me.Command.Contains(commander) {
				t.Error("the commander goes to the command zone")
			}
			if !tc.commandZone && !me.Graveyard.Contains(commander) {
				t.Error("the commander stays in the graveyard")
			}
			if n := len(effectstest.BattlefieldIDsNamed(g, "Food")); n != 1 {
				t.Errorf("after the CR 903.9a answer: %d Foods, want 1 (only the Hobbit)", n)
			}
		})
	}
}

// The Battle of Bywater destroys its creatures at the same time
// (CR 700.4), so a Zulaport Cutthroat big enough to be swept triggers
// for every creature its controller lost, itself included. Zulaport is
// pushed first, the battlefield position where the old one-at-a-time
// loop destroyed it before the others and it saw only its own death.
func TestB17BattleOfBywaterDeathsAreSimultaneousForAristocratsPayoffs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zulaport := effectstest.PushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue", effectstest.ZulaportOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(zulaport, "+1/+1", 2) })
	effectstest.PushWipeCreature(g, me.ID, "Ent A", "Creature — Treefolk", 3, 3)
	effectstest.PushWipeCreature(g, me.ID, "Ent B", "Creature — Treefolk", 4, 4)
	effectstest.PushWipeCreature(g, me.ID, "Ent C", "Creature — Treefolk", 5, 5)
	effectstest.PushWipeCreature(g, me.ID, "Hobbit", "Creature — Halfling", 1, 1)

	meBefore, oppBefore := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "The Battle of Bywater", "Sorcery", effectstest.B17BattleOfBywaterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(zulaport) {
		t.Fatal("Zulaport with two +1/+1 counters has power 3 and is destroyed")
	}
	if want := oppBefore - 4; opp.Life != want {
		t.Errorf("opponent life %d -> %d, want %d (four simultaneous deaths; the Hobbit is too small to die)", oppBefore, opp.Life, want)
	}
	if want := meBefore + 4; me.Life != want {
		t.Errorf("caster life %d -> %d, want %d", meBefore, me.Life, want)
	}
	if n := len(effectstest.BattlefieldIDsNamed(g, "Food")); n != 1 {
		t.Errorf("only the Hobbit is left: %d Foods, want 1", n)
	}
}

func TestB17RiseOfTheWitchKingEdictsEveryoneAndReturnsThePick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	spare := effectstest.B16Creature(g, me.ID, "Spare Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B16Creature(g, opp.ID, "Their Spare", "Creature — Bear", 2, 2, "G")
	rock := effectstest.B17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	// returnOffer is the return pick, never the sacrifice pick (both are
	// choose_cards since #3664).
	returnOffer := func() *game.PendingChoice {
		if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c != nil && !effectstest.IsSacrificePrompt(c) {
			return c
		}
		return nil
	}
	effectstest.CastCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery", effectstest.B17RiseOfTheWitchKingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// #3664: one player at a time, in APNAP order.
	if effectstest.SacrificeChoiceFor(g, me.ID) == nil || effectstest.SacrificeChoiceFor(g, opp.ID) != nil {
		t.Fatal("the active player is asked to sacrifice first, and alone")
	}
	// #1019: "if you sacrificed a creature this way" is about the
	// SACRIFICE, so nothing is offered while the prompts are open.
	if returnOffer() != nil {
		t.Error("the return was offered before anybody had chosen a creature")
	}
	effectstest.AnswerSacrifice(t, g, me.ID, mine)
	if returnOffer() != nil {
		t.Error("the run waits for every asked seat, not just the controller")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the controller's pick was sacrificed before the opponent had chosen")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)
	// #2863: the card is chosen now, from the graveyard as it is after
	// the sacrifices. "Another" excludes the Bear you just sacrificed
	// (ruling 2023-06-16), and the opponent's Bear is not yours.
	pick := returnOffer()
	if pick == nil {
		t.Fatal("the controller chooses the permanent card on resolution")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("you MAY return ONE: bounds %d..%d", pick.ChooseMin, pick.ChooseMax)
	}
	if !effectstest.HasID(pick.ChooseCards, rock) || effectstest.HasID(pick.ChooseCards, mine) || effectstest.HasID(pick.ChooseCards, theirs) {
		t.Errorf("offer %v: want your Sol Ring, never the Bear you sacrificed or theirs", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || effectstest.ControllerOf(t, g, rock) != me.ID {
		t.Error("the chosen permanent card returns to the battlefield")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("the sacrifices happen")
	}

	// With no creature of your own there is nothing to sacrifice, and
	// so nothing is offered.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(spare) })
	rock2 := effectstest.B17GraveyardCard(me, "Mind Stone", "Artifact", "{2}")
	effectstest.CastCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery", effectstest.B17RiseOfTheWitchKingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || g.Battlefield.Contains(rock2) {
		t.Error("no creature sacrificed — no return")
	}
}

// --- the creatures: dies triggers ----------------------------------

func TestB17HarvesterOfSoulsMayDrawWhenAnotherNontokenCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harvester := effectstest.B12Push(g, me.ID, "Harvester of Souls", "Creature — Demon", effectstest.B17HarvesterOfSoulsOracle, 5, 5)
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("an opponent's creature dying draws 1, drew %d", me.Hand.Size()-hand)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(goblin) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Fatal("a token dying is not a nontoken creature dying")
		}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(harvester) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Fatal("the Harvester's own death is not ANOTHER creature dying")
		}
	}
}

func TestB17UndeadAugurDrawsAndDrainsOnZombieDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	augur := effectstest.B12Push(g, me.ID, "Undead Augur", "Creature — Zombie Wizard", effectstest.B17UndeadAugurOracle, 2, 2)
	zombie := effectstest.B16Creature(g, me.ID, "Walking Corpse", "Creature — Zombie", 2, 2, "B")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2, "B")
	effectstest.AdvanceToMain(t, g)
	hand, life := me.Hand.Size(), me.Life
	effectstest.B15Destroy(t, g, zombie)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("a Zombie you control dying: draw 1, lose 1; got +%d cards, %d life", me.Hand.Size()-hand, me.Life-life)
	}
	effectstest.B15Destroy(t, g, bear)
	effectstest.B15Destroy(t, g, theirs)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Error("a non-Zombie of yours, or an opponent's Zombie, does nothing")
	}
	effectstest.B15Destroy(t, g, augur)
	if me.Hand.Size() != hand+2 || me.Life != life-2 {
		t.Errorf("the Augur's own death draws and drains too; got +%d cards, %d life", me.Hand.Size()-hand, me.Life-life)
	}
}

func TestB17CordialVampireGrowsYourVampiresOnAnyDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cordial := effectstest.B12Push(g, me.ID, "Cordial Vampire", "Creature — Vampire", effectstest.B17CordialVampireOracle, 1, 1)
	vampire := effectstest.B16Creature(g, me.ID, "Vampire Nighthawk", "Creature — Vampire Shaman", 2, 3, "B")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirVampire := effectstest.B16Creature(g, opp.ID, "Their Vampire", "Creature — Vampire", 2, 2, "B")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.B15Destroy(t, g, theirs)
	if effectstest.CounterCount(g, cordial, "+1/+1") != 1 || effectstest.CounterCount(g, vampire, "+1/+1") != 1 {
		t.Error("another creature dying: a counter on each Vampire you control, the Cordial included")
	}
	if effectstest.CounterCount(g, bear, "+1/+1") != 0 || effectstest.CounterCount(g, theirVampire, "+1/+1") != 0 {
		t.Error("a non-Vampire and an opponent's Vampire get nothing")
	}
	effectstest.B15Destroy(t, g, cordial)
	if effectstest.CounterCount(g, vampire, "+1/+1") != 2 {
		t.Error("the Cordial's own death grows the Vampires it leaves behind")
	}
}

// --- the creatures: ETB and search ---------------------------------

func TestB17VileEntomberSearchesAnyCardIntoTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Griselbrand", TypeLine: "Legendary Creature — Demon"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
	)
	effectstest.CastCatalogSpell(t, g, "Vile Entomber", "Creature — Zombie Warlock", effectstest.B17VileEntomberOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pc := effectstest.SearchChoiceFor(g, me.ID)
	if pc == nil {
		t.Fatal("three candidates for one slot: the searcher chooses")
	}
	// "A card" names no quality (CR 701.23d, #3053): Vile Entomber must find one.
	if pc.SearchMin != 1 {
		t.Errorf("SearchMin = %d, want 1: an unconditional search cannot fail to find", pc.SearchMin)
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Griselbrand")
	if effectstest.FindBattlefieldByName(g, "Griselbrand") != uuid.Nil || !effectstest.B02bGraveyardHasNamed(me, "Griselbrand") {
		t.Error("the chosen card goes to the graveyard")
	}
	if me.Library.Size() != 2 {
		t.Errorf("library %d, want 2", me.Library.Size())
	}
}

func TestB17RazakethPaysLifeAndACreatureToTutor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	razaketh := effectstest.B12Push(g, me.ID, "Razaketh, the Foulblooded", "Legendary Creature — Demon", effectstest.B17RazakethOracle, 8, 8)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	rock := effectstest.B12Permanent(g, me.ID, "Mana Rock", "Artifact")
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Demonic Tutor", TypeLine: "Sorcery"},
	)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, razaketh, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{razaketh}}); err == nil {
		t.Fatal("Razaketh can't sacrifice himself")
	}
	if err := g.ActivateCatalogAbility(me.ID, razaketh, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("an artifact is not a creature")
	}
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, razaketh, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if me.Life != life-2 || g.Battlefield.Contains(bear) {
		t.Error("the cost is 2 life and the creature")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Demonic Tutor")
	if !effectstest.B02bHandHasNamed(me, "Demonic Tutor") {
		t.Error("the chosen card reaches the hand")
	}
}

func TestB17ClaimJumperFetchesPlainsWhileAnOpponentIsAhead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	}
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Tundra", "Land — Plains Island"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
	)
	effectstest.CastCatalogSpell(t, g, "Claim Jumper", "Creature — Rabbit Mercenary", effectstest.B17ClaimJumperOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("behind on lands: the optional search asks")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Tundra") == uuid.Nil {
		t.Error("a Plains card is any land with the Plains subtype, and nothing else")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Tundra")
	tundra := effectstest.FindBattlefieldByName(g, "Tundra")
	if tundra == uuid.Nil || !effectstest.B16Tapped(t, g, tundra) {
		t.Fatal("the Plains card enters tapped")
	}
	// One land against two: still behind, so the process repeats once.
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("still behind after the first Plains: repeat once")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Plains")
	if effectstest.FindBattlefieldByName(g, "Plains") == uuid.Nil {
		t.Error("the second Plains reaches the battlefield")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("only once — no third search")
	}
	jumper := effectstest.FindBattlefieldByName(g, "Claim Jumper")
	if !effectstest.HasEffectiveKeyword(t, g, jumper, "vigilance") {
		t.Error("vigilance")
	}

	// Level on lands: no trigger at all (CR 603.4).
	effectstest.CastCatalogSpell(t, g, "Claim Jumper", "Creature — Rabbit Mercenary", effectstest.B17ClaimJumperOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("no opponent controls more lands — no search")
	}
}

func TestB17ThreefoldThunderhulkEntersWithCountersAndMakesGnomes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, me.ID, "Mana Rock", "Artifact")
	hulk := effectstest.CastCatalogSpell(t, g, "Threefold Thunderhulk", "Artifact Creature — Gnome", effectstest.B17ThreefoldThunderhulkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, hulk, "+1/+1") != 3 {
		t.Fatalf("enters with three +1/+1 counters, has %d", effectstest.CounterCount(g, hulk, "+1/+1"))
	}
	if n := effectstest.B16CountNamed(g, "Gnome"); n != 3 {
		t.Fatalf("the ETB makes Gnomes equal to its power (3), got %d", n)
	}
	gnome := effectstest.FindBattlefieldByName(g, "Gnome")
	if c := effectstest.CardByID(g, gnome); !c.IsArtifact() || !c.IsCreature() || effectstest.EffectivePower(t, g, gnome) != 1 {
		t.Error("a Gnome is a 1/1 artifact creature")
	}

	// {2}, sacrifice another artifact: a fourth counter, so the attack
	// makes four more Gnomes.
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, hulk, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{hulk}}); err == nil {
		t.Fatal("the Thunderhulk can't feed itself")
	}
	effectstest.B16Activate(t, g, me.ID, hulk, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if effectstest.CounterCount(g, hulk, "+1/+1") != 4 {
		t.Errorf("counters %d, want 4", effectstest.CounterCount(g, hulk, "+1/+1"))
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == hulk {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	effectstest.DeclareAttack(t, g, opp.ID, hulk)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Gnome"); n != 7 {
		t.Errorf("the attack makes four more Gnomes: %d total, want 7", n)
	}
}

func TestB17DeepglowSkateDoublesEveryKindOfCounterOnTheTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hydra", TypeLine: "Creature — Hydra", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{"+1/+1": 3},
	})
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jace", TypeLine: "Legendary Planeswalker — Jace", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Coffers", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{"charge": 2, "storage": 1},
	})
	untouched := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bystander", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{"+1/+1": 5},
	})
	effectstest.CastCatalogSpell(t, g, "Deepglow Skate", "Creature — Fish", effectstest.B17DeepglowSkateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("any number of target permanents: the controller picks")
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 0 {
		t.Errorf("min %d max %d, want 0 and unbounded", p.PickTargetMin, p.PickTargetMax)
	}
	effectstest.B17PickCards(t, g, me.ID, hydra, walker, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, hydra, "+1/+1") != 6 {
		t.Errorf("three +1/+1 doubled to %d, want 6", effectstest.CounterCount(g, hydra, "+1/+1"))
	}
	if effectstest.CounterCount(g, walker, game.CounterLoyalty) != 8 {
		t.Errorf("four loyalty doubled to %d, want 8", effectstest.CounterCount(g, walker, game.CounterLoyalty))
	}
	if effectstest.CounterCount(g, theirs, "charge") != 4 || effectstest.CounterCount(g, theirs, "storage") != 2 {
		t.Error("EACH kind doubles, on an opponent's permanent too")
	}
	if effectstest.CounterCount(g, untouched, "+1/+1") != 5 {
		t.Error("a permanent not targeted is untouched")
	}
}

// --- the creatures: statics ----------------------------------------

func TestB17NighthawkScavengerPowerCountsTypesInOpponentsGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	hawk := effectstest.B12Push(g, me.ID, "Nighthawk Scavenger", "Creature — Vampire Rogue", effectstest.B17NighthawkScavengerOracle, 1, 3)
	if effectstest.EffectivePower(t, g, hawk) != 1 {
		t.Fatalf("empty graveyards: power 1, got %d", effectstest.EffectivePower(t, g, hawk))
	}
	effectstest.B17GraveyardCard(opp, "Dead Bear", "Creature — Bear", "")
	effectstest.B17GraveyardCard(opp, "Dead Bolt", "Instant", "")
	effectstest.B17GraveyardCard(other, "Dead Bear Too", "Creature — Bear", "")
	effectstest.B17GraveyardCard(other, "Dead Rock", "Artifact Creature — Golem", "")
	effectstest.B17GraveyardCard(me, "My Dead Land", "Land", "")
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if got := effectstest.EffectivePower(t, g, hawk); got != 4 {
		t.Errorf("creature, instant and artifact across two opponents' graveyards: power 1+3 = 4, got %d", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, hawk, "flying") || !effectstest.HasEffectiveKeyword(t, g, hawk, "deathtouch") || !effectstest.HasEffectiveKeyword(t, g, hawk, "lifelink") {
		t.Error("flying, deathtouch, lifelink")
	}
}

func TestB17MultaniCountsLandsYouControlAndInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	multani := effectstest.B12Push(g, me.ID, "Multani, Yavimaya's Avatar", "Legendary Creature — Elemental Avatar", effectstest.B17MultaniOracle, 0, 0)
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	effectstest.B17GraveyardCard(me, "Dead Forest", "Basic Land — Forest", "")
	effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "")
	effectstest.B17GraveyardCard(opp, "Their Dead Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if got := effectstest.EffectivePower(t, g, multani); got != 3 {
		t.Errorf("two lands you control and one land card in your graveyard: 3/3, got power %d", got)
	}
	if got := effectstest.EffectiveToughness(t, g, multani); got != 3 {
		t.Errorf("toughness %d, want 3", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, multani, "reach") || !effectstest.HasEffectiveKeyword(t, g, multani, "trample") {
		t.Error("reach, trample")
	}
	// #1381: the graveyard ability is registered now
	// (multani_yavimayas_avatar_test.go proves it works) and the card
	// is complete.
	if spec, _ := effects.Lookup(effectstest.B17MultaniOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the card should be complete")
	}
}

func TestB17JaradGrowsWithCreatureCardsAndFlingsTheSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jarad := effectstest.B12Push(g, me.ID, "Jarad, Golgari Lich Lord", "Legendary Creature — Zombie Elf", effectstest.B17JaradOracle, 2, 2)
	effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "")
	effectstest.B17GraveyardCard(me, "Dead Bolt", "Instant", "")
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if got := effectstest.EffectivePower(t, g, jarad); got != 3 {
		t.Errorf("one creature card in the graveyard: 3/3, got power %d", got)
	}

	// Fling a 4/4 with a +1/+1 counter: each opponent loses 5.
	beast := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Beast", TypeLine: "Creature — Beast", Power: 4, Toughness: 4,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(beast, "+1/+1", 1) })
	before := effectstest.B17Life(g)
	effectstest.B06AddMana(me, "C", "B", "G")
	if err := g.ActivateCatalogAbility(me.ID, jarad, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{jarad}}); err == nil {
		t.Fatal("Jarad can't sacrifice himself")
	}
	effectstest.B16Activate(t, g, me.ID, jarad, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{beast}})
	for i, p := range g.Seats {
		want := before[i]
		if i != 0 {
			want -= 5
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d (each opponent loses the sacrificed creature's power)", i, p.Life, want)
		}
	}
	// The Beast is a creature card in the graveyard now: Jarad is 4/4.
	if got := effectstest.EffectivePower(t, g, jarad); got != 4 {
		t.Errorf("after the sacrifice Jarad's power is %d, want 4", got)
	}
}

// The fling reads the sacrificed creature as it last existed, so an
// anthem's bonus counts: a 4/4 under Glorious Anthem drains 5.
func TestB17JaradFlingCountsAnAnthem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jarad := effectstest.B12Push(g, me.ID, "Jarad, Golgari Lich Lord", "Legendary Creature — Zombie Elf", effectstest.B17JaradOracle, 2, 2)
	effectstest.PushGloriousAnthemFor(g, me.ID)
	beast := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Beast", TypeLine: "Creature — Beast", Power: 4, Toughness: 4,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	before := effectstest.B17Life(g)
	effectstest.B06AddMana(me, "C", "B", "G")
	effectstest.B16Activate(t, g, me.ID, jarad, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{beast}})
	for i, p := range g.Seats {
		want := before[i]
		if i != 0 {
			want -= 5
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d (the anthem's +1 is part of the sacrificed power)", i, p.Life, want)
		}
	}
}

func TestB17ArwenGivesOtherCreaturesCountersEqualToHerToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arwen := effectstest.B12Push(g, me.ID, "Arwen, Weaver of Hope", "Legendary Creature — Elf Noble", effectstest.B17ArwenOracle, 2, 1)
	bear := effectstest.Top100CastCreature(t, g, me, "Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, bear, "+1/+1") != 1 {
		t.Fatalf("Arwen's toughness is 1: the Bear enters with one counter, has %d", effectstest.CounterCount(g, bear, "+1/+1"))
	}
	if effectstest.EffectivePower(t, g, bear)+effectstest.CounterCount(g, bear, "+1/+1") != 3 {
		t.Error("the counter is real")
	}
	// A pumped Arwen hands out more.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(arwen, "+1/+1", 2) })
	second := effectstest.Top100CastCreature(t, g, me, "Second Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, second, "+1/+1") != 3 {
		t.Errorf("toughness 3 now: %d counters, want 3", effectstest.CounterCount(g, second, "+1/+1"))
	}
	// An opponent's creature, and a noncreature of yours, get nothing.
	effectstest.AdvanceToMainOf(t, g, 1)
	theirs := effectstest.Top100CastCreature(t, g, opp, "Their Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, theirs, "+1/+1") != 0 {
		t.Error("an opponent's creature is untouched")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	rock := effectstest.HandCardFull(me, "Mana Rock", "Artifact", "", "", nil)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, rock, "+1/+1") != 0 {
		t.Error("a noncreature gets nothing")
	}
	// #762: a created token takes the same entry pipeline, so it gets
	// Arwen's toughness in counters like every other creature.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	if got := effectstest.CounterCount(g, effectstest.FindBattlefieldByName(g, "Goblin"), "+1/+1"); got != 3 {
		t.Errorf("a token: %d +1/+1 counters, want 3 (Arwen's toughness)", got)
	}
	if spec, _ := effects.Lookup(effectstest.B17ArwenOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the token gap is closed — the caveat must be gone")
	}
}

func TestB17LeafCrownedVisionaryBuffsOtherElvesAndPaysToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	visionary := effectstest.B12Push(g, me.ID, "Leaf-Crowned Visionary", "Creature — Elf Druid", effectstest.B17LeafCrownedVisionaryOracle, 1, 1)
	elf := effectstest.B16Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1, "G")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	if effectstest.EffectivePower(t, g, elf) != 2 {
		t.Error("another Elf you control gets +1/+1")
	}
	if effectstest.EffectivePower(t, g, visionary) != 1 || effectstest.EffectivePower(t, g, bear) != 2 || effectstest.EffectivePower(t, g, theirs) != 1 {
		t.Error("not itself, not a non-Elf, not an opponent's Elf")
	}

	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Elvish Mystic", "Creature — Elf Druid", "", nil)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		// The trigger goes on the stack above the spell and asks when
		// it resolves.
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("casting an Elf spell asks to pay {G}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("paying draws 1, drew %d", me.Hand.Size()-hand)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a non-Elf spell asks nothing")
	}
}

// --- the creatures: attack, damage and cast triggers ---------------

func TestB17KazuulMakesAnOgreUnlessTheAttackerPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Kazuul, Tyrant of the Cliffs", "Legendary Creature — Ogre Warrior", effectstest.B17KazuulOracle, 5, 4)
	a := effectstest.B16Creature(g, opp.ID, "Raider A", "Creature — Human", 2, 2, "R")
	b := effectstest.B16Creature(g, opp.ID, "Raider B", "Creature — Human", 2, 2, "R")
	c := effectstest.B16Creature(g, opp.ID, "Raider C", "Creature — Human", 2, 2, "R")
	effectstest.AdvanceToMainOf(t, g, 1)
	// One declaration (CR 508.1, #3621): all three attackers are staged
	// before it locks in.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []struct{ id, at uuid.UUID }{{a, me.ID}, {b, me.ID}, {c, other.ID}} {
		if err := g.DeclareAttacker(d.id, d.at); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// Two creatures attacked Kazuul's controller: two prompts for the
	// attacker. One attacked someone else: no prompt for it.
	paid, declined := 0, 0
	for effectstest.HasPayUnlessFor(g, opp.ID) {
		if paid == 0 {
			effectstest.B06AddMana(opp, "C", "C", "C")
			effectstest.AnswerPayUnless(t, g, opp.ID, true)
			paid++
		} else {
			effectstest.AnswerPayUnless(t, g, opp.ID, false)
			declined++
		}
	}
	if paid != 1 || declined != 1 {
		t.Fatalf("two prompts (one per attacker at you), got %d answered", paid+declined)
	}
	ogres := effectstest.BattlefieldIDsNamed(g, "Ogre")
	if len(ogres) != 1 {
		t.Fatalf("one paid, one declined: %d Ogres, want 1", len(ogres))
	}
	if effectstest.ControllerOf(t, g, ogres[0]) != me.ID || effectstest.EffectivePower(t, g, ogres[0]) != 3 {
		t.Error("the Ogre is Kazuul's controller's 3/3")
	}
}

func TestB17PhyrexianObliteratorMakesTheDamagersControllerSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	obliterator := effectstest.B12Push(g, me.ID, "Phyrexian Obliterator", "Creature — Phyrexian Horror", effectstest.B17PhyrexianObliteratorOracle, 5, 5)
	x := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	y := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	z := effectstest.B12Permanent(g, opp.ID, "Their Land", "Basic Land — Forest")
	w := effectstest.B12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetCard(obliterator))
	effectstest.PassPriorityAroundTable(t, g)
	prompts := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSacrifice && c.Chooser == opp.ID {
			prompts++
		}
	}
	if prompts != 3 {
		t.Fatalf("3 damage: the Bolt's controller owes three sacrifices, %d prompts queued", prompts)
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, x)
	effectstest.AnswerSacrifice(t, g, opp.ID, y)
	effectstest.AnswerSacrifice(t, g, opp.ID, z)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(x) || g.Battlefield.Contains(y) || g.Battlefield.Contains(z) || !g.Battlefield.Contains(w) {
		t.Error("three permanents of their choice are gone, the fourth stays")
	}
	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil {
		t.Error("nothing more is owed")
	}
	if !g.Battlefield.Contains(obliterator) {
		t.Error("3 damage does not kill a 5/5")
	}
	_ = me
}

func TestB17FiremaneCommandoDrawsForWideAttacksAimedElsewhere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Firemane Commando", "Creature — Angel Soldier", effectstest.B17FiremaneCommandoOracle, 4, 3)
	mine := []uuid.UUID{
		effectstest.B16Creature(g, me.ID, "Knight A", "Creature — Knight", 2, 2, "W"),
		effectstest.B16Creature(g, me.ID, "Knight B", "Creature — Knight", 2, 2, "W"),
	}
	theirs := []uuid.UUID{
		effectstest.B16Creature(g, opp.ID, "Raider A", "Creature — Human", 2, 2, "R"),
		effectstest.B16Creature(g, opp.ID, "Raider B", "Creature — Human", 2, 2, "R"),
		effectstest.B16Creature(g, opp.ID, "Raider C", "Creature — Human", 2, 2, "R"),
	}
	// You attack with two: you draw.
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, mine...)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("attacking with two draws you 1, drew %d", me.Hand.Size()-hand)
	}

	// An opponent attacks another player with two: they draw.
	effectstest.AdvanceToMainOf(t, g, 1)
	hand, theirHand := me.Hand.Size(), opp.Hand.Size()
	effectstest.DeclareAttack(t, g, other.ID, theirs[0], theirs[1])
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != theirHand+1 {
		t.Errorf("another player attacking elsewhere with two draws them 1, drew %d", opp.Hand.Size()-theirHand)
	}
	if me.Hand.Size() != hand {
		t.Error("you draw nothing for their attack")
	}

	// Next turn cycle: one of the two comes at you — no card for them.
	effectstest.AdvanceToMainOf(t, g, 1)
	theirHand = opp.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(theirs[0], other.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(theirs[1], me.ID); err != nil {
		t.Fatal(err)
	}
	// One declaration, two defenders — the lock-in harvests off the
	// whole of it (#859), which is what "if none of those creatures
	// attacked you" is asking about.
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != theirHand {
		t.Errorf("one of the attackers came at you: no draw, drew %d", opp.Hand.Size()-theirHand)
	}
}

func TestB17BreenaTriggersOncePerOpponentAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	breena := effectstest.B12Push(g, me.ID, "Breena, the Demagogue", "Legendary Creature — Bird Warlock", effectstest.B17BreenaOracle, 1, 3)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	raiders := []uuid.UUID{
		effectstest.B16Creature(g, opp.ID, "Raider A", "Creature — Human", 2, 2, "R"),
		effectstest.B16Creature(g, opp.ID, "Raider B", "Creature — Human", 2, 2, "R"),
		effectstest.B16Creature(g, opp.ID, "Raider C", "Creature — Human", 2, 2, "R"),
	}
	// `other` is ahead of `third`: attacking `other` pays.
	other.Life, third.Life = 40, 30
	effectstest.AdvanceToMainOf(t, g, 1)
	theirHand := opp.Hand.Size()
	effectstest.DeclareAttack(t, g, other.ID, raiders[0], raiders[1])
	if n := effectstest.B784Triggers(g, breena); n != 1 {
		t.Fatalf("two creatures at one opponent: ONE trigger; %d on the stack", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The creature is chosen as the trigger resolves, not as it is put on
	// the stack: a hexproof creature of yours is as good as any.
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil || pick.ChooseMin != 1 || pick.ChooseMax != 1 || !effectstest.HasID(pick.ChooseCards, bear) || !effectstest.HasID(pick.ChooseCards, breena) {
		t.Fatal("one creature pick at resolution, offering every creature you control")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != theirHand+1 {
		t.Errorf("the attacking player draws 1, drew %d", opp.Hand.Size()-theirHand)
	}
	if effectstest.CounterCount(g, bear, "+1/+1") != 2 {
		t.Errorf("two +1/+1 counters on the picked creature, got %d", effectstest.CounterCount(g, bear, "+1/+1"))
	}
	// (Two creatures at that opponent were ONE trigger above, not one
	// per creature.)

	// Attacking the opponent who is BEHIND does not pay.
	effectstest.AdvanceToMainOf(t, g, 1)
	theirHand = opp.Hand.Size()
	effectstest.DeclareAttack(t, g, third.ID, raiders[0])
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil || opp.Hand.Size() != theirHand {
		t.Error("the attacked opponent has less life than the other: no trigger")
	}

	// Attacking Breena's own controller is not attacking an opponent
	// of hers.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.DeclareAttack(t, g, me.ID, raiders[0])
	if effectstest.B784Triggers(g, breena) != 0 {
		t.Error("you are not one of your own opponents")
	}
}

// #4839, CR 508.3e: attacking an opponent's planeswalker is not attacking
// that opponent. A creature at the planeswalker alone does not trigger
// Breena; one at the planeswalker and one at the player is one trigger,
// so the planeswalker attack did not spend the once-per-opponent slot.
func TestB17BreenaIgnoresAnAttackOnAnOpponentsPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	breena := effectstest.B12Push(g, me.ID, "Breena, the Demagogue", "Legendary Creature — Bird Warlock", effectstest.B17BreenaOracle, 1, 3)
	walker := effectstest.P1g2Planeswalker(g, other.ID, 5)
	a := effectstest.B16Creature(g, opp.ID, "Raider A", "Creature — Human", 2, 2, "R")
	b := effectstest.B16Creature(g, opp.ID, "Raider B", "Creature — Human", 2, 2, "R")
	// `other` is ahead of `third`, so an attack on `other` would pay.
	other.Life, third.Life = 40, 30
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.DeclareAttack(t, g, walker, a)
	if n := effectstest.B784Triggers(g, breena); n != 0 {
		t.Fatalf("an attack on an opponent's planeswalker alone: %d Breena triggers, want 0", n)
	}

	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, at := range []struct{ id, def uuid.UUID }{{a, walker}, {b, other.ID}} {
		if err := g.DeclareAttacker(at.id, at.def); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	if n := effectstest.B784Triggers(g, breena); n != 1 {
		t.Fatalf("one creature at the planeswalker, one at its controller: %d Breena triggers, want 1", n)
	}
}

// An additional combat triggers her again for the same opponent: the
// once-per-combat guard is per combat, and there is no per-turn tally on
// top of it. Aurelia, the Warleader supplies the extra combat.
func TestB17BreenaTriggersAgainInAnAdditionalCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	breena := effectstest.B12Push(g, me.ID, "Breena, the Demagogue", "Legendary Creature — Bird Warlock", effectstest.B17BreenaOracle, 1, 3)
	aurelia := effectstest.PushDiesCreatureForTest(g, me.ID, "Aurelia, the Warleader", effectstest.AureliaWarleaderOracle, "Legendary Creature — Angel", 3, 4)
	opp.Life, third.Life = 40, 30
	handBefore := me.Hand.Size()

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(aurelia, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil {
		t.Fatal("the first combat triggers Breena")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, breena)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, breena, "+1/+1"); got != 2 {
		t.Fatalf("first combat: %d counters on Breena, want 2", got)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %+v, want an additional combat", next)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(aurelia, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker in the second combat: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil {
		t.Fatal("the same opponent attacked in an additional combat triggers Breena again")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, breena)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, breena, "+1/+1"); got != 4 {
		t.Errorf("after the second combat: %d counters, want 4", got)
	}
	if got := me.Hand.Size() - handBefore; got < 2 {
		t.Errorf("the attacker drew %d from Breena over two combats, want at least 2", got)
	}
}

func TestB17TesharReturnsACheapCreatureWhenYouCastAHistoricSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Teshar, Ancestor's Apostle", "Legendary Creature — Bird Cleric", effectstest.B17TesharOracle, 2, 2)
	cheap := effectstest.B17GraveyardCard(me, "Myr", "Artifact Creature — Myr", "{2}")
	big := effectstest.B17GraveyardCard(me, "Wurm", "Creature — Wurm", "{5}{G}")
	rockCard := effectstest.B17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.HandCardFull(me, "Mind Stone", "Artifact", "{2}", "", nil)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("casting an artifact triggers the targeted return")
	}
	if effectstest.HasID(p.PickTargetCards, big) || effectstest.HasID(p.PickTargetCards, rockCard) || !effectstest.HasID(p.PickTargetCards, cheap) {
		t.Error("only creature cards with mana value 3 or less are offered")
	}
	effectstest.PickCard(t, g, me.ID, cheap)
	if effectstest.TriggerOnStack(g, effectstest.FindBattlefieldByName(g, "Teshar, Ancestor's Apostle")) == nil {
		t.Fatal("the trigger sits above the historic spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the creature card returns to the battlefield")
	}
	// A non-historic spell: no trigger.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a plain creature spell is not historic")
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestB17DictateOfKruphixDrawsEachPlayerAnExtraCardAndHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if !effectstest.HasEffectiveKeyword(t, g, effectstest.B12Push(g, me.ID, "Dictate of Kruphix", "Enchantment", effectstest.B17DictateOfKruphixOracle, 0, 0), "flash") {
		t.Error("flash")
	}
	for seat := 1; seat < 4; seat++ {
		p := g.Seats[seat]
		hand := p.Hand.Size()
		effectstest.Batch01AdvanceToStepOf(t, g, seat, game.StepDraw)
		effectstest.PassPriorityAroundTable(t, g)
		if p.Hand.Size() != hand+2 {
			t.Errorf("seat %d drew %d in its draw step, want 2 (the turn's draw plus the Dictate's)", seat, p.Hand.Size()-hand)
		}
	}
	hand := me.Hand.Size()
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepDraw)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("the controller's own draw step too: drew %d, want 2", me.Hand.Size()-hand)
	}
}

func TestB17ContainmentConstructExilesADiscardToPlayThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Containment Construct", "Artifact Creature — Construct", effectstest.B17ContainmentConstructOracle, 2, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.EmptyHandToLibrary(g, me)
	land := effectstest.HandCardFull(me, "Forest", "Basic Land — Forest", "", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(land) {
		t.Fatal("the discarded card is exiled from the graveyard")
	}
	perm := effectstest.ExiledPermission(g, land)
	if perm.Player != me.ID || perm.CastOnly || perm.Duration.Kind != game.UntilEndOfTurn {
		t.Errorf("grant %+v: the controller may PLAY it this turn", perm)
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing the exiled land: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land is played from exile")
	}
	// Declining leaves the card in the graveyard; an opponent's
	// discard is not yours.
	bolt := effectstest.HandCardFull(me, "Bolt", "Instant", "{R}", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bolt) {
		t.Error("declined: the card stays in the graveyard")
	}
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Error("an opponent's discard is not yours")
		}
	}
}

func TestB17ColossalGraveReaverMillsAndReanimatesTheChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// seedSearchLibrary pushes each card under the previous, so the
	// first listed is on top: the top three are what the ETB mills.
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Big Wurm", TypeLine: "Creature — Wurm", ManaCost: "{5}{G}", Power: 6, Toughness: 6},
		game.Card{Name: "Deep Card", TypeLine: "Instant"},
	)
	reaver := effectstest.CastCatalogSpell(t, g, "Colossal Grave-Reaver", "Creature — Dragon", effectstest.B17ColossalGraveReaverOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != 1 {
		t.Fatalf("the ETB mills three: library %d, want 1", me.Library.Size())
	}
	// #2523: the controller chooses — here the SMALLER creature, which
	// the old greatest-mana-value auto-pick would never have returned.
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("two milled creature cards: the controller is asked which one returns")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 2 {
		t.Errorf("prompt is %d..%d over %d cards, want exactly one of the two creature cards",
			pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
	}
	var bearID uuid.UUID
	for _, id := range pick.ChooseCards {
		if c, ok := g.LookupCardForEffect(id); ok && c.Name == "Small Bear" {
			bearID = id
		}
	}
	if bearID == uuid.Nil {
		t.Fatal("the Small Bear is not offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bearID)
	effectstest.PassPriorityAroundTable(t, g)
	wurm, bear := effectstest.FindBattlefieldByName(g, "Big Wurm"), effectstest.FindBattlefieldByName(g, "Small Bear")
	if bear == uuid.Nil || effectstest.ControllerOf(t, g, bear) != me.ID {
		t.Fatal("the chosen creature card comes back")
	}
	if wurm != uuid.Nil {
		t.Error("only ONE of them")
	}
	if !effectstest.B02bGraveyardHasNamed(me, "Big Wurm") || !effectstest.B02bGraveyardHasNamed(me, "Forest") {
		t.Error("the rest stay milled")
	}
	if !effectstest.HasEffectiveKeyword(t, g, reaver, "flying") {
		t.Error("flying")
	}
	// A lone milled creature card returns without a question.
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Lone Elf", TypeLine: "Creature — Elf", ManaCost: "{G}"})
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("one candidate is no choice")
	}
	if effectstest.FindBattlefieldByName(g, "Lone Elf") == uuid.Nil {
		t.Error("the lone milled creature returns")
	}
	// An opponent milling creature cards is not your graveyard.
	effectstest.SeedSearchLibrary(opp, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Their Bear") != uuid.Nil {
		t.Error("an opponent's mill does nothing")
	}
	// A mill with no creature card does nothing either.
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Bolt", TypeLine: "Instant"})
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	if len(g.PendingTriggers)+len(g.StackMeta) != 0 {
		t.Error("no creature card milled — no trigger")
	}
}

func TestB17InsidiousFungusOffersEachModeAsAnAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	shrine := effectstest.B12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	effectstest.AdvanceToMain(t, g)
	spec, _ := effects.Lookup(effectstest.B17InsidiousFungusOracle)
	if len(spec.Activated) != 3 {
		t.Fatalf("three modes, three abilities; got %d", len(spec.Activated))
	}

	fungus := effectstest.B12Push(g, me.ID, "Insidious Fungus", "Creature — Fungus", effectstest.B17InsidiousFungusOracle, 1, 2)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, fungus, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(shrine)}); err == nil {
		t.Fatal("the artifact mode does not take an enchantment")
	}
	effectstest.B16Activate(t, g, me.ID, fungus, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(rock)})
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(fungus) {
		t.Error("the artifact is destroyed and the Fungus is sacrificed")
	}

	fungus = effectstest.B12Push(g, me.ID, "Insidious Fungus", "Creature — Fungus", effectstest.B17InsidiousFungusOracle, 1, 2)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, fungus, 1, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(shrine)})
	if g.Battlefield.Contains(shrine) {
		t.Error("the enchantment is destroyed")
	}

	fungus = effectstest.B12Push(g, me.ID, "Insidious Fungus", "Creature — Fungus", effectstest.B17InsidiousFungusOracle, 1, 2)
	effectstest.B06AddMana(me, "C", "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, fungus, 2, game.ActivateAbilityParams{})
	if me.Hand.Size() != hand+1 {
		t.Errorf("the third mode draws 1, drew %d", me.Hand.Size()-hand)
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Error("the land-from-hand clause landed with #654; nothing is deferred any more")
	}
}
