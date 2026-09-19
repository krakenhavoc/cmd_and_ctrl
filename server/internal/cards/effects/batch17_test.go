package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch17_test.go — card-level coverage for the card-coverage
// roadmap's batch 17 (#310, `edhrec_rank` 1830–1929): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play, attack or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b17-prefixed.

const (
	b17JaradOracle                = "87e65e36-9483-49fe-b644-2caca092107f"
	b17TormodsCryptOracle         = "1573f7f9-672c-421a-b1ac-3d0d8aea59ca"
	b17BorosGuildgateOracle       = "73c423b7-cab8-4e69-8070-9edbf96a6c2c"
	b17BattleOfBywaterOracle      = "a94c191d-a938-458e-bc1b-2f44fd8873a3"
	b17DeepglowSkateOracle        = "debce64a-18bd-42f5-9e85-158c2242e9e9"
	b17ClaimJumperOracle          = "18f0cd0b-3e4f-4637-a62e-75dd1b2f3fce"
	b17UndeadAugurOracle          = "c26887e1-27f4-4550-921b-53460e43c079"
	b17MultaniOracle              = "4b8bf64b-4800-45ff-81c6-2857f34999b5"
	b17ThreefoldThunderhulkOracle = "b8020a8b-557b-465d-865d-b59fecd7abc1"
	b17FiremaneCommandoOracle     = "ac675899-25bb-4f06-9c8c-bf188024495f"
	b17ContainmentConstructOracle = "54d48e54-be4a-4a78-a778-b28e74ef7134"
	b17RazakethOracle             = "136c9ecb-59b4-4ef6-bcb2-7b8d3df3ee75"
	b17CordialVampireOracle       = "d61fb9e8-d05a-481a-a90f-5def300c9abb"
	b17ArwenOracle                = "c0891157-f3dd-4866-866f-fd13e9be9633"
	b17DictateOfKruphixOracle     = "be74a4c1-d569-4203-b28a-3e2ac6a82990"
	b17KazuulOracle               = "f8bf3d91-cb50-48ce-88c6-cdcb37b64b57"
	b17PhyrexianObliteratorOracle = "41820f91-27cf-41c0-bb5e-9adf6845a6a4"
	b17BreenaOracle               = "d11e627b-8a48-411d-a261-2c9a02a758ba"
	b17NighthawkScavengerOracle   = "379ed4ec-8f2e-448b-9ed2-a3181e2f877e"
	b17HarvesterOfSoulsOracle     = "5987ce77-10ad-4871-900a-5a005fcf4955"
	b17TempleBellOracle           = "fc8032e9-c83a-4cf9-92e1-b1d2d9642695"
	b17ColossalGraveReaverOracle  = "df8e0d1b-b47c-4807-9c9b-84dcec835254"
	b17JacesArchivistOracle       = "b6c8ac69-daa7-4e2e-a1d9-439731a81870"
	b17InsidiousFungusOracle      = "0a8d0217-ff24-4177-b6be-707eb2b6b9e9"
	b17VileEntomberOracle         = "556dd333-4066-41f7-98f6-41794754de71"
	b17RefuteOracle               = "eb1dfa29-7371-4cb6-bfa2-16f7820b69be"
	b17LeafCrownedVisionaryOracle = "ef693b9f-11e3-49bf-8387-b8f480b9007a"
	b17TesharOracle               = "d4193f50-33b5-4201-ba03-eb135eced7ff"
	b17RiseOfTheWitchKingOracle   = "3c86541c-3601-4a38-8872-39705e41303a"
)

// b17Life reads every seat's life.
func b17Life(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		out = append(out, p.Life)
	}
	return out
}

// b17Hands reads every seat's hand size.
func b17Hands(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		out = append(out, p.Hand.Size())
	}
	return out
}

// b17GraveyardCard seeds a card with a full type line into a
// player's graveyard.
func b17GraveyardCard(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// b17PickCards answers a multi-slot pick_target prompt.
func b17PickCards(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	refs := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.ResolvePickTargets(p.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Boros
// Guildgate is a row in the Guildgate table, so a transposed row is
// invisible until someone plays that exact card. Korvold and
// Evacuation were already on main and are not in this table.
func TestBatch17CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b17JaradOracle:                "Jarad, Golgari Lich Lord",
		b17TormodsCryptOracle:         "Tormod's Crypt",
		b17BorosGuildgateOracle:       "Boros Guildgate",
		b17BattleOfBywaterOracle:      "The Battle of Bywater",
		b17DeepglowSkateOracle:        "Deepglow Skate",
		b17ClaimJumperOracle:          "Claim Jumper",
		b17UndeadAugurOracle:          "Undead Augur",
		b17MultaniOracle:              "Multani, Yavimaya's Avatar",
		b17ThreefoldThunderhulkOracle: "Threefold Thunderhulk",
		b17FiremaneCommandoOracle:     "Firemane Commando",
		b17ContainmentConstructOracle: "Containment Construct",
		b17RazakethOracle:             "Razaketh, the Foulblooded",
		b17CordialVampireOracle:       "Cordial Vampire",
		b17ArwenOracle:                "Arwen, Weaver of Hope",
		b17DictateOfKruphixOracle:     "Dictate of Kruphix",
		b17KazuulOracle:               "Kazuul, Tyrant of the Cliffs",
		b17PhyrexianObliteratorOracle: "Phyrexian Obliterator",
		b17BreenaOracle:               "Breena, the Demagogue",
		b17NighthawkScavengerOracle:   "Nighthawk Scavenger",
		b17HarvesterOfSoulsOracle:     "Harvester of Souls",
		b17TempleBellOracle:           "Temple Bell",
		b17ColossalGraveReaverOracle:  "Colossal Grave-Reaver",
		b17JacesArchivistOracle:       "Jace's Archivist",
		b17InsidiousFungusOracle:      "Insidious Fungus",
		b17VileEntomberOracle:         "Vile Entomber",
		b17RefuteOracle:               "Refute",
		b17LeafCrownedVisionaryOracle: "Leaf-Crowned Visionary",
		b17TesharOracle:               "Teshar, Ancestor's Apostle",
		b17RiseOfTheWitchKingOracle:   "Rise of the Witch-king",
	}
	if len(want) != 29 {
		t.Fatalf("the batch registers 29 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
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
	g := newCatalogGame(t)
	gate := playLandFromHand(t, g, "Boros Guildgate", b17BorosGuildgateOracle)
	top100AssertEnteredTapped(t, g, gate, "Boros Guildgate")
	if tapEventsFor(g, gate) != 0 {
		t.Error("enters-tapped is a replacement, not a tap")
	}
	spec, _ := Lookup(b17BorosGuildgateOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{R|W}" {
		t.Errorf("mana abilities %+v, want one producing {R|W}", spec.ManaAbilities)
	}
}

func TestB17TormodsCryptExilesATargetPlayersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	crypt := b12Push(g, me.ID, "Tormod's Crypt", "Artifact", b17TormodsCryptOracle, 0, 0)
	a := pushGraveyardCardForTest(opp, "Dead A")
	b := pushGraveyardCardForTest(opp, "Dead B")
	c := pushGraveyardCardForTest(other, "Dead C")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, crypt, 0, game.ActivateAbilityParams{Targets: b16TargetPlayer(opp.ID)})
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	bell := b12Push(g, me.ID, "Temple Bell", "Artifact", b17TempleBellOracle, 0, 0)
	advanceToMain(t, g)
	before := b17Hands(g)
	b16Activate(t, g, me.ID, bell, 0, game.ActivateAbilityParams{})
	for i, p := range g.Seats {
		if p.Hand.Size() != before[i]+1 {
			t.Errorf("seat %d drew %d, want 1", i, p.Hand.Size()-before[i])
		}
	}
	if !b16Tapped(t, g, bell) {
		t.Error("the Bell taps as its cost")
	}
}

func TestB17JacesArchivistWheelsToTheGreatestDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archivist := pushCatalogPermanent(g, me.ID, "Jace's Archivist", "Creature — Vedalken Wizard", b17JacesArchivistOracle, false)
	advanceToMain(t, g)
	// Hands of 8 (me — seat 0 drew for turn 1, CR 103.8c at four
	// seats), 7 (opp) and 7, 7: pad the opponent's to 9 so the
	// greatest discard is theirs.
	for i := 0; i < 2; i++ {
		handCardFull(opp, "Extra", "Instant", "", "", nil)
	}
	b06AddMana(me, "U")
	b16Activate(t, g, me.ID, archivist, 0, game.ActivateAbilityParams{})
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetPlayer(me.ID))
	castCatalogSpell(t, g, "Refute", "Instant", b17RefuteOracle, b16TargetCard(bolt))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != 40 {
		t.Fatal("the Bolt is countered")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d, want 1", me.Hand.Size()-hand)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", discardOwed(g, me.ID))
	}
}

func TestB17BattleOfBywaterKillsBigCreaturesAndFeedsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b16Creature(g, me.ID, "Hobbit", "Creature — Halfling", 1, 1, "W")
	big := b16Creature(g, me.ID, "Ent", "Creature — Treefolk", 4, 4, "G")
	theirs := b16Creature(g, opp.ID, "Their Dragon", "Creature — Dragon", 5, 5, "R")
	theirSmall := b16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 2, 2, "R")
	darksteel := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Darksteel Colossus", TypeLine: "Artifact Creature — Golem",
		Power: 11, Toughness: 11, Keywords: []string{"indestructible"}, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "The Battle of Bywater", "Sorcery", b17BattleOfBywaterOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || g.Battlefield.Contains(theirs) {
		t.Error("creatures with power 3 or greater are destroyed")
	}
	if !g.Battlefield.Contains(small) || !g.Battlefield.Contains(theirSmall) {
		t.Error("smaller creatures survive")
	}
	if !g.Battlefield.Contains(darksteel) {
		t.Error("an indestructible creature survives (#446 — DestroyAllMatching honours it)")
	}
	foods := battlefieldIDsNamed(g, "Food")
	if len(foods) != 2 {
		t.Fatalf("two creatures you control survived: %d Foods, want 2", len(foods))
	}
	for _, id := range foods {
		if controllerOf(t, g, id) != me.ID {
			t.Error("the Foods are the caster's")
		}
	}
}

// A commander of yours with power 3 or more is destroyed by The Battle
// of Bywater, so it is not a creature you control when the Foods are
// counted. The engine keeps it on the battlefield until its owner
// answers the CR 903.9 prompt, so a plain board count would pay for it.
//
// #815: the Foods are made from the destruction's continuation, so
// nothing is counted until that answer arrives — and then the
// commander is not among the creatures its controller still has,
// whichever answer it was.
func TestB17BattleOfBywaterDoesNotFeedADestroyedCommander(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
	}{{"to the command zone", true}, {"to the graveyard", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			commander := b36Commander(g, me.ID, "My Commander") // a 3/3
			hobbit := b16Creature(g, me.ID, "Hobbit", "Creature — Halfling", 1, 1, "W")

			castCatalogSpell(t, g, "The Battle of Bywater", "Sorcery", b17BattleOfBywaterOracle, nil)
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(hobbit) {
				t.Fatal("the Hobbit is too small to be destroyed")
			}
			if n := len(battlefieldIDsNamed(g, "Food")); n != 0 {
				t.Errorf("%d Foods while the CR 903.9 prompt is still open, want 0 — "+
					"the sweep is not finished until it is answered", n)
			}

			if tc.commandZone {
				b36AcceptCommandZone(t, g, me.ID)
				if !me.Command.Contains(commander) {
					t.Error("the commander goes to the command zone")
				}
			} else {
				b21DeclineCommandZone(t, g, me.ID)
				if !me.Graveyard.Contains(commander) {
					t.Error("the commander goes to the graveyard")
				}
			}
			if n := len(battlefieldIDsNamed(g, "Food")); n != 1 {
				t.Errorf("after the CR 903.9 answer: %d Foods, want 1 (only the Hobbit)", n)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zulaport := pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue", zulaportOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(zulaport, "+1/+1", 2) })
	pushWipeCreature(g, me.ID, "Ent A", "Creature — Treefolk", 3, 3)
	pushWipeCreature(g, me.ID, "Ent B", "Creature — Treefolk", 4, 4)
	pushWipeCreature(g, me.ID, "Ent C", "Creature — Treefolk", 5, 5)
	pushWipeCreature(g, me.ID, "Hobbit", "Creature — Halfling", 1, 1)

	meBefore, oppBefore := me.Life, opp.Life
	castCatalogSpell(t, g, "The Battle of Bywater", "Sorcery", b17BattleOfBywaterOracle, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(zulaport) {
		t.Fatal("Zulaport with two +1/+1 counters has power 3 and is destroyed")
	}
	if want := oppBefore - 4; opp.Life != want {
		t.Errorf("opponent life %d -> %d, want %d (four simultaneous deaths; the Hobbit is too small to die)", oppBefore, opp.Life, want)
	}
	if want := meBefore + 4; me.Life != want {
		t.Errorf("caster life %d -> %d, want %d", meBefore, me.Life, want)
	}
	if n := len(battlefieldIDsNamed(g, "Food")); n != 1 {
		t.Errorf("only the Hobbit is left: %d Foods, want 1", n)
	}
}

func TestB17RiseOfTheWitchKingEdictsEveryoneAndReturnsThePick(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	rock := b17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	castCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery", b17RiseOfTheWitchKingOracle, b16TargetCard(rock))
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, me.ID) == nil || sacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("each player with a creature is asked to sacrifice one")
	}
	// #1019: "if you sacrificed a creature this way" is about the
	// SACRIFICE, so nothing comes back while the prompts are open.
	if g.Battlefield.Contains(rock) {
		t.Error("the permanent came back before anybody had chosen a creature")
	}
	answerSacrifice(t, g, me.ID, mine)
	if g.Battlefield.Contains(rock) {
		t.Error("the run waits for every asked seat, not just the controller")
	}
	answerSacrifice(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || controllerOf(t, g, rock) != me.ID {
		t.Error("the picked permanent card returns to the battlefield once the sacrifices have landed")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("the sacrifices happen")
	}

	// With no creature of your own there is nothing to sacrifice, and
	// so nothing comes back.
	rock2 := b17GraveyardCard(me, "Mind Stone", "Artifact", "{2}")
	castCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery", b17RiseOfTheWitchKingOracle, b16TargetCard(rock2))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock2) {
		t.Error("no creature sacrificed — no return")
	}
	// And it is castable with no pick at all.
	castCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery", b17RiseOfTheWitchKingOracle, nil)
	passPriorityAroundTable(t, g)
}

// --- the creatures: dies triggers ----------------------------------

func TestB17HarvesterOfSoulsMayDrawWhenAnotherNontokenCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harvester := b12Push(g, me.ID, "Harvester of Souls", "Creature — Demon", b17HarvesterOfSoulsOracle, 5, 5)
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("an opponent's creature dying draws 1, drew %d", me.Hand.Size()-hand)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1) })
	goblin := findBattlefieldByName(g, "Goblin")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(goblin) })
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Fatal("a token dying is not a nontoken creature dying")
		}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(harvester) })
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt {
			t.Fatal("the Harvester's own death is not ANOTHER creature dying")
		}
	}
}

func TestB17UndeadAugurDrawsAndDrainsOnZombieDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	augur := b12Push(g, me.ID, "Undead Augur", "Creature — Zombie Wizard", b17UndeadAugurOracle, 2, 2)
	zombie := b16Creature(g, me.ID, "Walking Corpse", "Creature — Zombie", 2, 2, "B")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2, "B")
	advanceToMain(t, g)
	hand, life := me.Hand.Size(), me.Life
	b15Destroy(t, g, zombie)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("a Zombie you control dying: draw 1, lose 1; got +%d cards, %d life", me.Hand.Size()-hand, me.Life-life)
	}
	b15Destroy(t, g, bear)
	b15Destroy(t, g, theirs)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Error("a non-Zombie of yours, or an opponent's Zombie, does nothing")
	}
	b15Destroy(t, g, augur)
	if me.Hand.Size() != hand+2 || me.Life != life-2 {
		t.Errorf("the Augur's own death draws and drains too; got +%d cards, %d life", me.Hand.Size()-hand, me.Life-life)
	}
}

func TestB17CordialVampireGrowsYourVampiresOnAnyDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cordial := b12Push(g, me.ID, "Cordial Vampire", "Creature — Vampire", b17CordialVampireOracle, 1, 1)
	vampire := b16Creature(g, me.ID, "Vampire Nighthawk", "Creature — Vampire Shaman", 2, 3, "B")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirVampire := b16Creature(g, opp.ID, "Their Vampire", "Creature — Vampire", 2, 2, "B")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	b15Destroy(t, g, theirs)
	if counterCount(g, cordial, "+1/+1") != 1 || counterCount(g, vampire, "+1/+1") != 1 {
		t.Error("another creature dying: a counter on each Vampire you control, the Cordial included")
	}
	if counterCount(g, bear, "+1/+1") != 0 || counterCount(g, theirVampire, "+1/+1") != 0 {
		t.Error("a non-Vampire and an opponent's Vampire get nothing")
	}
	b15Destroy(t, g, cordial)
	if counterCount(g, vampire, "+1/+1") != 2 {
		t.Error("the Cordial's own death grows the Vampires it leaves behind")
	}
}

// --- the creatures: ETB and search ---------------------------------

func TestB17VileEntomberSearchesAnyCardIntoTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Griselbrand", TypeLine: "Legendary Creature — Demon"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
	)
	castCatalogSpell(t, g, "Vile Entomber", "Creature — Zombie Warlock", b17VileEntomberOracle, nil)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("three candidates for one slot: the searcher chooses")
	}
	answerSearchNamed(t, g, me.ID, "Griselbrand")
	if findBattlefieldByName(g, "Griselbrand") != uuid.Nil || !b02bGraveyardHasNamed(me, "Griselbrand") {
		t.Error("the chosen card goes to the graveyard")
	}
	if me.Library.Size() != 2 {
		t.Errorf("library %d, want 2", me.Library.Size())
	}
}

func TestB17RazakethPaysLifeAndACreatureToTutor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	razaketh := b12Push(g, me.ID, "Razaketh, the Foulblooded", "Legendary Creature — Demon", b17RazakethOracle, 8, 8)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	rock := b12Permanent(g, me.ID, "Mana Rock", "Artifact")
	seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Demonic Tutor", TypeLine: "Sorcery"},
	)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, razaketh, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{razaketh}}); err == nil {
		t.Fatal("Razaketh can't sacrifice himself")
	}
	if err := g.ActivateCatalogAbility(me.ID, razaketh, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("an artifact is not a creature")
	}
	life := me.Life
	b16Activate(t, g, me.ID, razaketh, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if me.Life != life-2 || g.Battlefield.Contains(bear) {
		t.Error("the cost is 2 life and the creature")
	}
	answerSearchNamed(t, g, me.ID, "Demonic Tutor")
	if !b02bHandHasNamed(me, "Demonic Tutor") {
		t.Error("the chosen card reaches the hand")
	}
}

func TestB17ClaimJumperFetchesPlainsWhileAnOpponentIsAhead(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		b12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	}
	seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Tundra", "Land — Plains Island"),
		searchTestLand("Forest", "Basic Land — Forest"),
	)
	castCatalogSpell(t, g, "Claim Jumper", "Creature — Rabbit Mercenary", b17ClaimJumperOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("behind on lands: the optional search asks")
	}
	if searchOptionNamed(g, c, "Forest") != uuid.Nil || searchOptionNamed(g, c, "Tundra") == uuid.Nil {
		t.Error("a Plains card is any land with the Plains subtype, and nothing else")
	}
	answerSearchNamed(t, g, me.ID, "Tundra")
	tundra := findBattlefieldByName(g, "Tundra")
	if tundra == uuid.Nil || !b16Tapped(t, g, tundra) {
		t.Fatal("the Plains card enters tapped")
	}
	// One land against two: still behind, so the process repeats once.
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("still behind after the first Plains: repeat once")
	}
	answerSearchNamed(t, g, me.ID, "Plains")
	if findBattlefieldByName(g, "Plains") == uuid.Nil {
		t.Error("the second Plains reaches the battlefield")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("only once — no third search")
	}
	jumper := findBattlefieldByName(g, "Claim Jumper")
	if !hasEffectiveKeyword(t, g, jumper, "vigilance") {
		t.Error("vigilance")
	}

	// Level on lands: no trigger at all (CR 603.4).
	castCatalogSpell(t, g, "Claim Jumper", "Creature — Rabbit Mercenary", b17ClaimJumperOracle, nil)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("no opponent controls more lands — no search")
	}
}

func TestB17ThreefoldThunderhulkEntersWithCountersAndMakesGnomes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, me.ID, "Mana Rock", "Artifact")
	hulk := castCatalogSpell(t, g, "Threefold Thunderhulk", "Artifact Creature — Gnome", b17ThreefoldThunderhulkOracle, nil)
	passPriorityAroundTable(t, g)
	if counterCount(g, hulk, "+1/+1") != 3 {
		t.Fatalf("enters with three +1/+1 counters, has %d", counterCount(g, hulk, "+1/+1"))
	}
	if n := b16CountNamed(g, "Gnome"); n != 3 {
		t.Fatalf("the ETB makes Gnomes equal to its power (3), got %d", n)
	}
	gnome := findBattlefieldByName(g, "Gnome")
	if c := cardByID(g, gnome); !c.IsArtifact() || !c.IsCreature() || effectivePower(t, g, gnome) != 1 {
		t.Error("a Gnome is a 1/1 artifact creature")
	}

	// {2}, sacrifice another artifact: a fourth counter, so the attack
	// makes four more Gnomes.
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, hulk, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{hulk}}); err == nil {
		t.Fatal("the Thunderhulk can't feed itself")
	}
	b16Activate(t, g, me.ID, hulk, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if counterCount(g, hulk, "+1/+1") != 4 {
		t.Errorf("counters %d, want 4", counterCount(g, hulk, "+1/+1"))
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == hulk {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	declareAttack(t, g, opp.ID, hulk)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Gnome"); n != 7 {
		t.Errorf("the attack makes four more Gnomes: %d total, want 7", n)
	}
}

func TestB17DeepglowSkateDoublesEveryKindOfCounterOnTheTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hydra", TypeLine: "Creature — Hydra", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{"+1/+1": 3},
	})
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jace", TypeLine: "Legendary Planeswalker — Jace", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 4},
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Coffers", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{"charge": 2, "storage": 1},
	})
	untouched := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bystander", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{"+1/+1": 5},
	})
	castCatalogSpell(t, g, "Deepglow Skate", "Creature — Fish", b17DeepglowSkateOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("any number of target permanents: the controller picks")
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 0 {
		t.Errorf("min %d max %d, want 0 and unbounded", p.PickTargetMin, p.PickTargetMax)
	}
	b17PickCards(t, g, me.ID, hydra, walker, theirs)
	passPriorityAroundTable(t, g)
	if counterCount(g, hydra, "+1/+1") != 6 {
		t.Errorf("three +1/+1 doubled to %d, want 6", counterCount(g, hydra, "+1/+1"))
	}
	if counterCount(g, walker, game.CounterLoyalty) != 8 {
		t.Errorf("four loyalty doubled to %d, want 8", counterCount(g, walker, game.CounterLoyalty))
	}
	if counterCount(g, theirs, "charge") != 4 || counterCount(g, theirs, "storage") != 2 {
		t.Error("EACH kind doubles, on an opponent's permanent too")
	}
	if counterCount(g, untouched, "+1/+1") != 5 {
		t.Error("a permanent not targeted is untouched")
	}
}

// --- the creatures: statics ----------------------------------------

func TestB17NighthawkScavengerPowerCountsTypesInOpponentsGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	hawk := b12Push(g, me.ID, "Nighthawk Scavenger", "Creature — Vampire Rogue", b17NighthawkScavengerOracle, 1, 3)
	if effectivePower(t, g, hawk) != 1 {
		t.Fatalf("empty graveyards: power 1, got %d", effectivePower(t, g, hawk))
	}
	b17GraveyardCard(opp, "Dead Bear", "Creature — Bear", "")
	b17GraveyardCard(opp, "Dead Bolt", "Instant", "")
	b17GraveyardCard(other, "Dead Bear Too", "Creature — Bear", "")
	b17GraveyardCard(other, "Dead Rock", "Artifact Creature — Golem", "")
	b17GraveyardCard(me, "My Dead Land", "Land", "")
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if got := effectivePower(t, g, hawk); got != 4 {
		t.Errorf("creature, instant and artifact across two opponents' graveyards: power 1+3 = 4, got %d", got)
	}
	if !hasEffectiveKeyword(t, g, hawk, "flying") || !hasEffectiveKeyword(t, g, hawk, "deathtouch") || !hasEffectiveKeyword(t, g, hawk, "lifelink") {
		t.Error("flying, deathtouch, lifelink")
	}
}

func TestB17MultaniCountsLandsYouControlAndInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	multani := b12Push(g, me.ID, "Multani, Yavimaya's Avatar", "Legendary Creature — Elemental Avatar", b17MultaniOracle, 0, 0)
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	b17GraveyardCard(me, "Dead Forest", "Basic Land — Forest", "")
	b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "")
	b17GraveyardCard(opp, "Their Dead Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if got := effectivePower(t, g, multani); got != 3 {
		t.Errorf("two lands you control and one land card in your graveyard: 3/3, got power %d", got)
	}
	if got := effectiveToughness(t, g, multani); got != 3 {
		t.Errorf("toughness %d, want 3", got)
	}
	if !hasEffectiveKeyword(t, g, multani, "reach") || !hasEffectiveKeyword(t, g, multani, "trample") {
		t.Error("reach, trample")
	}
	if spec, _ := Lookup(b17MultaniOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the graveyard ability gap must be declared")
	}
}

func TestB17JaradGrowsWithCreatureCardsAndFlingsTheSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jarad := b12Push(g, me.ID, "Jarad, Golgari Lich Lord", "Legendary Creature — Zombie Elf", b17JaradOracle, 2, 2)
	b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "")
	b17GraveyardCard(me, "Dead Bolt", "Instant", "")
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if got := effectivePower(t, g, jarad); got != 3 {
		t.Errorf("one creature card in the graveyard: 3/3, got power %d", got)
	}

	// Fling a 4/4 with a +1/+1 counter: each opponent loses 5.
	beast := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Beast", TypeLine: "Creature — Beast", Power: 4, Toughness: 4,
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(beast, "+1/+1", 1) })
	before := b17Life(g)
	b06AddMana(me, "C", "B", "G")
	if err := g.ActivateCatalogAbility(me.ID, jarad, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{jarad}}); err == nil {
		t.Fatal("Jarad can't sacrifice himself")
	}
	b16Activate(t, g, me.ID, jarad, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{beast}})
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
	if got := effectivePower(t, g, jarad); got != 4 {
		t.Errorf("after the sacrifice Jarad's power is %d, want 4", got)
	}
}

func TestB17ArwenGivesOtherCreaturesCountersEqualToHerToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arwen := b12Push(g, me.ID, "Arwen, Weaver of Hope", "Legendary Creature — Elf Noble", b17ArwenOracle, 2, 1)
	bear := top100CastCreature(t, g, me, "Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if counterCount(g, bear, "+1/+1") != 1 {
		t.Fatalf("Arwen's toughness is 1: the Bear enters with one counter, has %d", counterCount(g, bear, "+1/+1"))
	}
	if effectivePower(t, g, bear)+counterCount(g, bear, "+1/+1") != 3 {
		t.Error("the counter is real")
	}
	// A pumped Arwen hands out more.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(arwen, "+1/+1", 2) })
	second := top100CastCreature(t, g, me, "Second Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if counterCount(g, second, "+1/+1") != 3 {
		t.Errorf("toughness 3 now: %d counters, want 3", counterCount(g, second, "+1/+1"))
	}
	// An opponent's creature, and a noncreature of yours, get nothing.
	advanceToMainOf(t, g, 1)
	theirs := top100CastCreature(t, g, opp, "Their Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if counterCount(g, theirs, "+1/+1") != 0 {
		t.Error("an opponent's creature is untouched")
	}
	advanceToMainOf(t, g, 0)
	rock := handCardFull(me, "Mana Rock", "Artifact", "", "", nil)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if counterCount(g, rock, "+1/+1") != 0 {
		t.Error("a noncreature gets nothing")
	}
	// #762: a created token takes the same entry pipeline, so it gets
	// Arwen's toughness in counters like every other creature.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	if got := counterCount(g, findBattlefieldByName(g, "Goblin"), "+1/+1"); got != 3 {
		t.Errorf("a token: %d +1/+1 counters, want 3 (Arwen's toughness)", got)
	}
	if spec, _ := Lookup(b17ArwenOracle); spec.Completeness != CompletenessFull {
		t.Error("the token gap is closed — the caveat must be gone")
	}
}

func TestB17LeafCrownedVisionaryBuffsOtherElvesAndPaysToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	visionary := b12Push(g, me.ID, "Leaf-Crowned Visionary", "Creature — Elf Druid", b17LeafCrownedVisionaryOracle, 1, 1)
	elf := b16Creature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1, "G")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	if effectivePower(t, g, elf) != 2 {
		t.Error("another Elf you control gets +1/+1")
	}
	if effectivePower(t, g, visionary) != 1 || effectivePower(t, g, bear) != 2 || effectivePower(t, g, theirs) != 1 {
		t.Error("not itself, not a non-Elf, not an opponent's Elf")
	}

	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Elvish Mystic", "Creature — Elf Druid", "", nil)
	if !hasPayUnlessFor(g, me.ID) {
		// The trigger goes on the stack above the spell and asks when
		// it resolves.
		passPriorityAroundTable(t, g)
	}
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("casting an Elf spell asks to pay {G}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("paying draws 1, drew %d", me.Hand.Size()-hand)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("a non-Elf spell asks nothing")
	}
}

// --- the creatures: attack, damage and cast triggers ---------------

func TestB17KazuulMakesAnOgreUnlessTheAttackerPays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Kazuul, Tyrant of the Cliffs", "Legendary Creature — Ogre Warrior", b17KazuulOracle, 5, 4)
	a := b16Creature(g, opp.ID, "Raider A", "Creature — Human", 2, 2, "R")
	b := b16Creature(g, opp.ID, "Raider B", "Creature — Human", 2, 2, "R")
	c := b16Creature(g, opp.ID, "Raider C", "Creature — Human", 2, 2, "R")
	advanceToMainOf(t, g, 1)
	declareAttack(t, g, me.ID, a, b)
	if err := g.DeclareAttacker(c, other.ID); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	// Two creatures attacked Kazuul's controller: two prompts for the
	// attacker. One attacked someone else: no prompt for it.
	paid, declined := 0, 0
	for hasPayUnlessFor(g, opp.ID) {
		if paid == 0 {
			b06AddMana(opp, "C", "C", "C")
			answerPayUnless(t, g, opp.ID, true)
			paid++
		} else {
			answerPayUnless(t, g, opp.ID, false)
			declined++
		}
	}
	if paid != 1 || declined != 1 {
		t.Fatalf("two prompts (one per attacker at you), got %d answered", paid+declined)
	}
	ogres := battlefieldIDsNamed(g, "Ogre")
	if len(ogres) != 1 {
		t.Fatalf("one paid, one declined: %d Ogres, want 1", len(ogres))
	}
	if controllerOf(t, g, ogres[0]) != me.ID || effectivePower(t, g, ogres[0]) != 3 {
		t.Error("the Ogre is Kazuul's controller's 3/3")
	}
}

func TestB17PhyrexianObliteratorMakesTheDamagersControllerSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	obliterator := b12Push(g, me.ID, "Phyrexian Obliterator", "Creature — Phyrexian Horror", b17PhyrexianObliteratorOracle, 5, 5)
	x := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	y := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	z := b12Permanent(g, opp.ID, "Their Land", "Basic Land — Forest")
	w := b12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetCard(obliterator))
	passPriorityAroundTable(t, g)
	prompts := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSacrifice && c.Chooser == opp.ID {
			prompts++
		}
	}
	if prompts != 3 {
		t.Fatalf("3 damage: the Bolt's controller owes three sacrifices, %d prompts queued", prompts)
	}
	answerSacrifice(t, g, opp.ID, x)
	answerSacrifice(t, g, opp.ID, y)
	answerSacrifice(t, g, opp.ID, z)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(x) || g.Battlefield.Contains(y) || g.Battlefield.Contains(z) || !g.Battlefield.Contains(w) {
		t.Error("three permanents of their choice are gone, the fourth stays")
	}
	if sacrificeChoiceFor(g, opp.ID) != nil {
		t.Error("nothing more is owed")
	}
	if !g.Battlefield.Contains(obliterator) {
		t.Error("3 damage does not kill a 5/5")
	}
	_ = me
}

func TestB17FiremaneCommandoDrawsForWideAttacksAimedElsewhere(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Firemane Commando", "Creature — Angel Soldier", b17FiremaneCommandoOracle, 4, 3)
	mine := []uuid.UUID{
		b16Creature(g, me.ID, "Knight A", "Creature — Knight", 2, 2, "W"),
		b16Creature(g, me.ID, "Knight B", "Creature — Knight", 2, 2, "W"),
	}
	theirs := []uuid.UUID{
		b16Creature(g, opp.ID, "Raider A", "Creature — Human", 2, 2, "R"),
		b16Creature(g, opp.ID, "Raider B", "Creature — Human", 2, 2, "R"),
		b16Creature(g, opp.ID, "Raider C", "Creature — Human", 2, 2, "R"),
	}
	// You attack with two: you draw.
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, mine...)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("attacking with two draws you 1, drew %d", me.Hand.Size()-hand)
	}

	// An opponent attacks another player with two: they draw.
	advanceToMainOf(t, g, 1)
	hand, theirHand := me.Hand.Size(), opp.Hand.Size()
	declareAttack(t, g, other.ID, theirs[0], theirs[1])
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != theirHand+1 {
		t.Errorf("another player attacking elsewhere with two draws them 1, drew %d", opp.Hand.Size()-theirHand)
	}
	if me.Hand.Size() != hand {
		t.Error("you draw nothing for their attack")
	}

	// Next turn cycle: one of the two comes at you — no card for them.
	advanceToMainOf(t, g, 1)
	theirHand = opp.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(theirs[0], other.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(theirs[1], me.ID); err != nil {
		t.Fatal(err)
	}
	// One declaration, two defenders — the lock-in harvests off the
	// whole of it (#859), which is what "if none of those creatures
	// attacked you" is asking about.
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != theirHand {
		t.Errorf("one of the attackers came at you: no draw, drew %d", opp.Hand.Size()-theirHand)
	}
}

func TestB17BreenaTriggersOncePerOpponentAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	breena := b12Push(g, me.ID, "Breena, the Demagogue", "Legendary Creature — Bird Warlock", b17BreenaOracle, 1, 3)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	raiders := []uuid.UUID{
		b16Creature(g, opp.ID, "Raider A", "Creature — Human", 2, 2, "R"),
		b16Creature(g, opp.ID, "Raider B", "Creature — Human", 2, 2, "R"),
		b16Creature(g, opp.ID, "Raider C", "Creature — Human", 2, 2, "R"),
	}
	// `other` is ahead of `third`: attacking `other` pays.
	other.Life, third.Life = 40, 30
	advanceToMainOf(t, g, 1)
	theirHand := opp.Hand.Size()
	declareAttack(t, g, other.ID, raiders[0], raiders[1])
	if n := len(g.PendingChoices); n != 1 || latestPickTarget(g, me.ID) == nil {
		t.Fatalf("two creatures at one opponent: ONE trigger, one creature pick; %d prompts", n)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != theirHand+1 {
		t.Errorf("the attacking player draws 1, drew %d", opp.Hand.Size()-theirHand)
	}
	if counterCount(g, bear, "+1/+1") != 2 {
		t.Errorf("two +1/+1 counters on the picked creature, got %d", counterCount(g, bear, "+1/+1"))
	}
	// A third creature at the same opponent: nothing more this combat.
	if err := g.DeclareAttacker(raiders[2], other.ID); err != nil {
		t.Fatal(err)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the same opponent attacked again this combat does not re-trigger")
	}

	// Attacking the opponent who is BEHIND does not pay.
	advanceToMainOf(t, g, 1)
	theirHand = opp.Hand.Size()
	declareAttack(t, g, third.ID, raiders[0])
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || opp.Hand.Size() != theirHand {
		t.Error("the attacked opponent has less life than the other: no trigger")
	}

	// Attacking Breena's own controller is not attacking an opponent
	// of hers.
	advanceToMainOf(t, g, 1)
	declareAttack(t, g, me.ID, raiders[0])
	if latestPickTarget(g, me.ID) != nil {
		t.Error("you are not one of your own opponents")
	}
	_ = breena
}

func TestB17TesharReturnsACheapCreatureWhenYouCastAHistoricSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Teshar, Ancestor's Apostle", "Legendary Creature — Bird Cleric", b17TesharOracle, 2, 2)
	cheap := b17GraveyardCard(me, "Myr", "Artifact Creature — Myr", "{2}")
	big := b17GraveyardCard(me, "Wurm", "Creature — Wurm", "{5}{G}")
	rockCard := b17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	advanceToMain(t, g)
	rock := handCardFull(me, "Mind Stone", "Artifact", "{2}", "", nil)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("casting an artifact triggers the targeted return")
	}
	if hasID(p.PickTargetCards, big) || hasID(p.PickTargetCards, rockCard) || !hasID(p.PickTargetCards, cheap) {
		t.Error("only creature cards with mana value 3 or less are offered")
	}
	pickCard(t, g, me.ID, cheap)
	if triggerOnStack(g, findBattlefieldByName(g, "Teshar, Ancestor's Apostle")) == nil {
		t.Fatal("the trigger sits above the historic spell")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the creature card returns to the battlefield")
	}
	// A non-historic spell: no trigger.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a plain creature spell is not historic")
	}
	passPriorityAroundTable(t, g)
}

func TestB17DictateOfKruphixDrawsEachPlayerAnExtraCardAndHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if !hasEffectiveKeyword(t, g, b12Push(g, me.ID, "Dictate of Kruphix", "Enchantment", b17DictateOfKruphixOracle, 0, 0), "flash") {
		t.Error("flash")
	}
	for seat := 1; seat < 4; seat++ {
		p := g.Seats[seat]
		hand := p.Hand.Size()
		batch01AdvanceToStepOf(t, g, seat, game.StepDraw)
		passPriorityAroundTable(t, g)
		if p.Hand.Size() != hand+2 {
			t.Errorf("seat %d drew %d in its draw step, want 2 (the turn's draw plus the Dictate's)", seat, p.Hand.Size()-hand)
		}
	}
	hand := me.Hand.Size()
	batch01AdvanceToStepOf(t, g, 0, game.StepDraw)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("the controller's own draw step too: drew %d, want 2", me.Hand.Size()-hand)
	}
}

func TestB17ContainmentConstructExilesADiscardToPlayThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Containment Construct", "Artifact Creature — Construct", b17ContainmentConstructOracle, 2, 1)
	advanceToMain(t, g)
	emptyHandToLibrary(g, me)
	land := handCardFull(me, "Forest", "Basic Land — Forest", "", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(land) {
		t.Fatal("the discarded card is exiled from the graveyard")
	}
	perm := exiledPermission(g, land)
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
	bolt := handCardFull(me, "Bolt", "Instant", "{R}", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
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

func TestB17ColossalGraveReaverMillsAndReanimatesTheBiggestCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// seedSearchLibrary pushes each card under the previous, so the
	// first listed is on top: the top three are what the ETB mills.
	seedSearchLibrary(me,
		game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Big Wurm", TypeLine: "Creature — Wurm", ManaCost: "{5}{G}", Power: 6, Toughness: 6},
		game.Card{Name: "Deep Card", TypeLine: "Instant"},
	)
	reaver := castCatalogSpell(t, g, "Colossal Grave-Reaver", "Creature — Dragon", b17ColossalGraveReaverOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 1 {
		t.Fatalf("the ETB mills three: library %d, want 1", me.Library.Size())
	}
	wurm, bear := findBattlefieldByName(g, "Big Wurm"), findBattlefieldByName(g, "Small Bear")
	if wurm == uuid.Nil || controllerOf(t, g, wurm) != me.ID {
		t.Fatal("one milled creature card comes back — the greatest mana value")
	}
	if bear != uuid.Nil {
		t.Error("only ONE of them")
	}
	if !b02bGraveyardHasNamed(me, "Small Bear") || !b02bGraveyardHasNamed(me, "Forest") {
		t.Error("the rest stay milled")
	}
	if !hasEffectiveKeyword(t, g, reaver, "flying") {
		t.Error("flying")
	}
	// An opponent milling creature cards is not your graveyard.
	seedSearchLibrary(opp, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})
	g.WithWriteLock(func() { _ = g.MillNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Their Bear") != uuid.Nil {
		t.Error("an opponent's mill does nothing")
	}
	// A mill with no creature card does nothing either.
	seedSearchLibrary(me, game.Card{Name: "Bolt", TypeLine: "Instant"})
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	if len(g.PendingTriggers)+len(g.StackMeta) != 0 {
		t.Error("no creature card milled — no trigger")
	}
}

func TestB17InsidiousFungusOffersEachModeAsAnAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	shrine := b12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	advanceToMain(t, g)
	spec, _ := Lookup(b17InsidiousFungusOracle)
	if len(spec.Activated) != 3 {
		t.Fatalf("three modes, three abilities; got %d", len(spec.Activated))
	}

	fungus := b12Push(g, me.ID, "Insidious Fungus", "Creature — Fungus", b17InsidiousFungusOracle, 1, 2)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, fungus, 0, game.ActivateAbilityParams{Targets: b16TargetCard(shrine)}); err == nil {
		t.Fatal("the artifact mode does not take an enchantment")
	}
	b16Activate(t, g, me.ID, fungus, 0, game.ActivateAbilityParams{Targets: b16TargetCard(rock)})
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(fungus) {
		t.Error("the artifact is destroyed and the Fungus is sacrificed")
	}

	fungus = b12Push(g, me.ID, "Insidious Fungus", "Creature — Fungus", b17InsidiousFungusOracle, 1, 2)
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, fungus, 1, game.ActivateAbilityParams{Targets: b16TargetCard(shrine)})
	if g.Battlefield.Contains(shrine) {
		t.Error("the enchantment is destroyed")
	}

	fungus = b12Push(g, me.ID, "Insidious Fungus", "Creature — Fungus", b17InsidiousFungusOracle, 1, 2)
	b06AddMana(me, "C", "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, fungus, 2, game.ActivateAbilityParams{})
	if me.Hand.Size() != hand+1 {
		t.Errorf("the third mode draws 1, drew %d", me.Hand.Size()-hand)
	}
	if spec.Completeness != CompletenessFull {
		t.Error("the land-from-hand clause landed with #654; nothing is deferred any more")
	}
}
