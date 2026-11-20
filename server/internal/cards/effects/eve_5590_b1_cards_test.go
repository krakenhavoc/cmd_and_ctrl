package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eve_5590_b1_cards_test.go — Eventide cards, batch 1 (#5590). Every test
// plays the real catalog entry through the engine.

const (
	eveDuskdaleWurmOracle        = "aba4e290-b597-4c93-b3dd-745ca45ad72d"
	eveHarvestGwyllionOracle     = "bcdd64d7-8414-488b-9b13-d5678d06625f"
	eveHearthfireHobgoblinOracle = "bc940eda-efa8-4227-9b56-0c6fbf1398d2"
	eveHobgoblinDragoonOracle    = "3bbfceba-9f3a-4988-a009-13c9472138b1"
	eveNipGwyllionOracle         = "cd53f7cf-e4da-4e93-8550-60556300b56c"
	eveSmolderingButcherOracle   = "b0f56eef-1abc-437e-8e2e-c079731fe783"
	eveLingeringTormentorOracle  = "276b1bd8-8918-4226-bab3-100cf5aa3c24"
	eveRendclawTrowOracle        = "86177c7e-a6eb-448b-baf8-aedd06f0a2f3"
	eveTrapjawKelpieOracle       = "0d6fccd5-ec14-4da6-972e-b2f5502818a4"
	eveStalkerHagOracle          = "c6aa6091-5503-4c22-91af-f999112b0081"
	eveBloodiedGhostOracle       = "1195de01-f0ee-42c1-8e9e-a4bb90dd4932"
	eveSpiritOfTheHearthOracle   = "30cd69a8-7893-4075-94ca-04450ff821d3"
	eveTalonrendOracle           = "85c6290e-2bf2-4666-a606-326a8209ace6"
	eveTwinbladeSlasherOracle    = "2375a9c6-28ef-4c8f-909c-0aa7e5fc085c"
	eveDuergarCaveGuardOracle    = "79d7950b-92c1-4431-a69e-26544eb0c7f0"
	eveStreamHopperOracle        = "153cd93c-7bba-4482-893c-55a602c4439b"
	eveStillmoonCavalierOracle   = "9c430b5a-83b6-4f4f-b1d3-5737e1f61947"
	eveOdiousTrowOracle          = "0e9864e2-1bb4-4527-b2af-62433a90b669"
	eveSutureSpiritOracle        = "4b2eba61-3937-4dc2-ad0a-e51fff8ecb80"
	eveHoofSkulkinOracle         = "7f60549f-7a0a-4009-a51a-d3f51214be11"
	eveJawboneSkulkinOracle      = "1cbcc073-1645-4d7f-98f4-7309d7b35692"
	eveShellSkulkinOracle        = "82df38bc-f29b-49f8-b734-e4df6d0f645a"
	eveFangSkulkinOracle         = "54bed864-fe2e-4359-b0ac-c27a2c225487"
	eveWildernessHypnotistOracle = "4c6870f3-4b70-4050-bf22-b1f348f89360"
	eveSpitemareOracle           = "cf4d8b19-fd11-4200-9871-896772019f4c"
	eveOverbeingOfMythOracle     = "5391f601-b778-413f-a31f-9f8b23b40f5c"
	eveDeityOfScarsOracle        = "bf7c1fa8-0460-4be5-b00d-72c0960c4149"
	eveNoggleBanditOracle        = "3e7707df-54e3-4307-9695-2de4e3edcf89"
	eveNobilisOfWarOracle        = "4789ac35-ca91-4046-8262-24c34e751cea"
	eveNettleSentinelOracle      = "5c367b7d-3cb4-45db-b836-dad953efce15"
	eveCinderPyromancerOracle    = "47bfee3e-34b4-4b89-aaf8-f74e0c3d3c62"
	eveBallynockTrapperOracle    = "96232342-3ec7-4740-a114-969537b8f2fc"
	eveMerrowLevitatorOracle     = "a4154dac-3f6d-4f8f-a769-9f5b8acecda7"
	eveBelligerentHatchlingOracl = "7bd4e504-541e-4b2c-9f1d-3c67957aef2e"
	eveNoxiousHatchlingOracle    = "36072943-7823-434d-b898-f0a2c4e00af2"
	eveVoraciousHatchlingOracle  = "eab5eb29-bed6-456b-8929-5b9b820aa5a3"
	eveSturdyHatchlingOracle     = "08e95b6c-9f4c-42d9-ad64-2d905d781d49"
	eveShrewdHatchlingOracle     = "011371f0-a7f1-47a5-9883-22a14cfe4de9"
	eveLoyalGyrfalconOracle      = "f693cf76-0919-4b89-92de-e619be006abb"
	eveDoubleCleaveOracle        = "d86f3d5e-838c-4787-9417-85a6997c3bde"
	eveUnmakeOracle              = "c551ba63-2a74-439d-8996-d566e2efbd5f"
	eveHallowedBurialOracle      = "dcaaddee-00b1-495c-9161-40c91c900aac"
	eveInundateOracle            = "0390a3dc-386c-402f-a89c-b94867bfb0a2"
	eveBanishingKnackOracle      = "6f23621e-eceb-402e-a8d3-8e2b734762b3"
	eveWistfulSelkieOracle       = "4503d0fb-842d-49f1-9769-8509ff0f2ce1"
	eveFlickerwispOracle         = "b23a3d30-6b8e-4aad-890f-db0c3af43ace"
	eveNoggleBridgebreakerOracle = "d7e127e4-56fd-4116-9c1a-421beab58e5c"
	eveCacheRaidersOracle        = "57947037-91b2-4058-be22-158d3257317e"

	eveBearType      = "Creature — Bear"
	eveElementalType = "Creature — Elemental"
)

// eveCastColoured has `p` cast a vanilla instant whose mana cost is `cost`
// (its colours are read off the cost), floating exactly that mana first.
func eveCastColoured(t *testing.T, g *game.Game, p *game.Player, cost string) uuid.UUID {
	t.Helper()
	spec := ""
	for _, r := range cost {
		if r != '{' && r != '}' {
			spec += string(r)
		}
	}
	floatForTest(g, p, spec)
	return castFromHandForTest(t, g, p, "Spell", "Instant", cost, "", strict)
}

// --- every card is in the catalog ----------------------------------------

func TestEve5590B1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eveDuskdaleWurmOracle: "Duskdale Wurm", eveHarvestGwyllionOracle: "Harvest Gwyllion",
		eveHearthfireHobgoblinOracle: "Hearthfire Hobgoblin", eveHobgoblinDragoonOracle: "Hobgoblin Dragoon",
		eveNipGwyllionOracle: "Nip Gwyllion", eveSmolderingButcherOracle: "Smoldering Butcher",
		eveLingeringTormentorOracle: "Lingering Tormentor", eveRendclawTrowOracle: "Rendclaw Trow",
		eveTrapjawKelpieOracle: "Trapjaw Kelpie", eveStalkerHagOracle: "Stalker Hag",
		eveBloodiedGhostOracle: "Bloodied Ghost", eveSpiritOfTheHearthOracle: "Spirit of the Hearth",
		eveTalonrendOracle: "Talonrend", eveTwinbladeSlasherOracle: "Twinblade Slasher",
		eveDuergarCaveGuardOracle: "Duergar Cave-Guard", eveStreamHopperOracle: "Stream Hopper",
		eveStillmoonCavalierOracle: "Stillmoon Cavalier", eveOdiousTrowOracle: "Odious Trow",
		eveSutureSpiritOracle: "Suture Spirit", eveHoofSkulkinOracle: "Hoof Skulkin",
		eveJawboneSkulkinOracle: "Jawbone Skulkin", eveShellSkulkinOracle: "Shell Skulkin",
		eveFangSkulkinOracle: "Fang Skulkin", eveWildernessHypnotistOracle: "Wilderness Hypnotist",
		eveSpitemareOracle: "Spitemare", eveOverbeingOfMythOracle: "Overbeing of Myth",
		eveDeityOfScarsOracle: "Deity of Scars", eveNoggleBanditOracle: "Noggle Bandit",
		eveNobilisOfWarOracle: "Nobilis of War", eveNettleSentinelOracle: "Nettle Sentinel",
		eveCinderPyromancerOracle: "Cinder Pyromancer", eveBallynockTrapperOracle: "Ballynock Trapper",
		eveMerrowLevitatorOracle: "Merrow Levitator", eveBelligerentHatchlingOracl: "Belligerent Hatchling",
		eveNoxiousHatchlingOracle: "Noxious Hatchling", eveVoraciousHatchlingOracle: "Voracious Hatchling",
		eveSturdyHatchlingOracle: "Sturdy Hatchling", eveShrewdHatchlingOracle: "Shrewd Hatchling",
		eveLoyalGyrfalconOracle: "Loyal Gyrfalcon", eveDoubleCleaveOracle: "Double Cleave",
		eveUnmakeOracle: "Unmake", eveHallowedBurialOracle: "Hallowed Burial",
		eveInundateOracle: "Inundate", eveBanishingKnackOracle: "Banishing Knack",
		eveWistfulSelkieOracle: "Wistful Selkie", eveFlickerwispOracle: "Flickerwisp",
		eveNoggleBridgebreakerOracle: "Noggle Bridgebreaker", eveCacheRaidersOracle: "Cache Raiders",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- printed keywords -------------------------------------------------------

func TestEve5590B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Duskdale Wurm", "Creature — Wurm", eveDuskdaleWurmOracle, []string{"trample"}},
		{"Harvest Gwyllion", "Creature — Hag", eveHarvestGwyllionOracle, []string{game.KeywordWither}},
		{"Hearthfire Hobgoblin", "Creature — Goblin Soldier", eveHearthfireHobgoblinOracle, []string{"double strike"}},
		{"Hobgoblin Dragoon", "Creature — Goblin Knight", eveHobgoblinDragoonOracle, []string{"flying", "first strike"}},
		{"Nip Gwyllion", "Creature — Hag", eveNipGwyllionOracle, []string{"lifelink"}},
		{"Smoldering Butcher", "Creature — Elemental Warrior", eveSmolderingButcherOracle, []string{game.KeywordWither}},
		{"Lingering Tormentor", "Creature — Spirit", eveLingeringTormentorOracle, []string{"fear", game.KeywordPersist}},
		{"Rendclaw Trow", "Creature — Troll", eveRendclawTrowOracle, []string{game.KeywordWither, game.KeywordPersist}},
		{"Trapjaw Kelpie", "Creature — Beast", eveTrapjawKelpieOracle, []string{"flash", game.KeywordPersist}},
		{"Stalker Hag", "Creature — Hag", eveStalkerHagOracle, []string{"swampwalk", "forestwalk"}},
		{"Bloodied Ghost", "Creature — Spirit", eveBloodiedGhostOracle, []string{"flying"}},
		{"Spirit of the Hearth", "Creature — Cat Spirit", eveSpiritOfTheHearthOracle, []string{"flying"}},
		{"Talonrend", eveElementalType, eveTalonrendOracle, []string{"flying"}},
		{"Twinblade Slasher", "Creature — Elf Warrior", eveTwinbladeSlasherOracle, []string{game.KeywordWither}},
		{"Duergar Cave-Guard", "Creature — Dwarf Warrior", eveDuergarCaveGuardOracle, []string{game.KeywordWither}},
		{"Stillmoon Cavalier", "Creature — Zombie Knight", eveStillmoonCavalierOracle, []string{"protection from white", "protection from black"}},
		{"Suture Spirit", "Creature — Spirit", eveSutureSpiritOracle, []string{"flying"}},
		{"Deity of Scars", "Creature — Spirit Avatar", eveDeityOfScarsOracle, []string{"trample"}},
		{"Nobilis of War", "Creature — Spirit Avatar", eveNobilisOfWarOracle, []string{"flying"}},
		{"Belligerent Hatchling", eveElementalType, eveBelligerentHatchlingOracl, []string{"first strike"}},
		{"Noxious Hatchling", eveElementalType, eveNoxiousHatchlingOracle, []string{game.KeywordWither}},
		{"Voracious Hatchling", eveElementalType, eveVoraciousHatchlingOracle, []string{"lifelink"}},
		{"Loyal Gyrfalcon", "Creature — Bird", eveLoyalGyrfalconOracle, []string{"defender", "flying"}},
		{"Flickerwisp", eveElementalType, eveFlickerwispOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s (has %v)", tc.name, kw, abilities)
				}
			}
		})
	}
}

// Fear and forestwalk are block rules: a green creature can't block the
// Tormentor, and nothing can block the Hag while its defender has a Forest.
func TestEve5590B1EvasionKeywordsRefuseBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tormentor := b12Push(g, me.ID, "Lingering Tormentor", "Creature — Spirit", eveLingeringTormentorOracle, 2, 2)
	hag := b12Push(g, me.ID, "Stalker Hag", "Creature — Hag", eveStalkerHagOracle, 3, 2)
	b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := b16Creature(g, opp.ID, "Bear", eveBearType, 2, 2, "G")
	other := b16Creature(g, opp.ID, "Other Bear", eveBearType, 2, 2, "G")
	brAttack(t, g, tormentor, hag)
	if err := g.DeclareBlocker(bear, tormentor); err == nil {
		t.Error("a green creature blocked a creature with fear")
	}
	if err := g.DeclareBlocker(other, hag); err == nil {
		t.Error("a creature blocked a forestwalker while its controller has a Forest")
	}
}

// --- entry counters ---------------------------------------------------------

func TestEve5590B1EntersWithMinusCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness, want int
	}{
		{"Bloodied Ghost", "Creature — Spirit", eveBloodiedGhostOracle, 3, 3, 1},
		{"Deity of Scars", "Creature — Spirit Avatar", eveDeityOfScarsOracle, 7, 7, 2},
		{"Belligerent Hatchling", eveElementalType, eveBelligerentHatchlingOracl, 6, 6, 4},
		{"Noxious Hatchling", eveElementalType, eveNoxiousHatchlingOracle, 6, 6, 4},
		{"Voracious Hatchling", eveElementalType, eveVoraciousHatchlingOracle, 6, 6, 4},
		{"Sturdy Hatchling", eveElementalType, eveSturdyHatchlingOracle, 6, 6, 4},
		{"Shrewd Hatchling", eveElementalType, eveShrewdHatchlingOracle, 6, 6, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s did not resolve", tc.name)
			}
			if got := minusCounters(g, id); got != tc.want {
				t.Errorf("%d -1/-1 counters, want %d", got, tc.want)
			}
		})
	}
}

// Each colour is its own trigger: a spell of one colour removes one
// counter, a spell of both removes two.
func TestEve5590B1HatchlingsLoseACounterPerSpellOfTheirColours(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, first, second string
	}{
		{"Belligerent Hatchling", eveBelligerentHatchlingOracl, "R", "W"},
		{"Noxious Hatchling", eveNoxiousHatchlingOracle, "B", "G"},
		{"Voracious Hatchling", eveVoraciousHatchlingOracle, "W", "B"},
		{"Sturdy Hatchling", eveSturdyHatchlingOracle, "G", "U"},
		{"Shrewd Hatchling", eveShrewdHatchlingOracle, "U", "R"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := castHoldCreature(t, g, tc.name, eveElementalType, tc.oracle, 6, 6)
			eveCastColoured(t, g, me, "{"+tc.first+"}")
			passPriorityAroundTable(t, g)
			if got := minusCounters(g, id); got != 3 {
				t.Fatalf("after a one-colour spell: %d counters, want 3", got)
			}
			eveCastColoured(t, g, me, "{"+tc.first+"}{"+tc.second+"}")
			passPriorityAroundTable(t, g)
			// The two triggers go on the stack together, and their
			// controller is asked to order them (CR 603.3b).
			if triggerOrderPrompt(g) == nil {
				t.Fatalf("a two-colour spell put no trigger-order prompt up: one trigger, not two")
			}
			answerTriggerOrderInOfferedOrder(t, g)
			passPriorityAroundTable(t, g)
			if got := minusCounters(g, id); got != 1 {
				t.Errorf("after a two-colour spell: %d counters, want 1", got)
			}
		})
	}
}

// --- cast triggers ----------------------------------------------------------

// A spell of another colour asks nothing; one of the printed colour asks,
// and a yes untaps the creature.
func TestEve5590B1UntapWhenYouCastASpellOfItsColour(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, colour, other string
	}{
		{"Nettle Sentinel", "Creature — Elf Warrior", eveNettleSentinelOracle, "G", "U"},
		{"Cinder Pyromancer", "Creature — Elemental Shaman", eveCinderPyromancerOracle, "R", "G"},
		{"Ballynock Trapper", "Creature — Kithkin Soldier", eveBallynockTrapperOracle, "W", "B"},
		{"Merrow Levitator", "Creature — Merfolk Wizard", eveMerrowLevitatorOracle, "U", "R"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			arnTap(t, g, id)
			eveCastColoured(t, g, me, "{"+tc.other+"}")
			passPriorityAroundTable(t, g)
			if n := b18TriggerPromptCount(g, me.ID); n != 0 {
				t.Fatalf("a spell of another colour asked %d questions", n)
			}
			if !arnTapped(t, g, id) {
				t.Fatal("it untapped on a spell of another colour")
			}
			eveCastColoured(t, g, me, "{"+tc.colour+"}")
			answerLatestTriggerPrompt(t, g, me.ID, true)
			passPriorityAroundTable(t, g)
			if arnTapped(t, g, id) {
				t.Error("it did not untap")
			}
		})
	}
}

func TestEve5590B1NettleSentinelDoesntUntapInYourUntapStep(t *testing.T) {
	spec, ok := Lookup(eveNettleSentinelOracle)
	if !ok || len(spec.UntapStepRestrictions) != 1 {
		t.Fatalf("Nettle Sentinel declares %d untap-step restrictions, want 1", len(spec.UntapStepRestrictions))
	}
	g := newCatalogGame(t)
	owner := g.Seats[1]
	sentinel := b12Push(g, owner.ID, "Nettle Sentinel", "Creature — Elf Warrior", eveNettleSentinelOracle, 2, 2)
	arnTap(t, g, sentinel)
	advanceToUpkeepOf(t, g, 1)
	if !arnTapped(t, g, sentinel) {
		t.Error("Nettle Sentinel untapped in its controller's untap step")
	}
}

func TestEve5590B1LoyalGyrfalconLosesDefenderOnAWhiteSpell(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bird := b12Push(g, me.ID, "Loyal Gyrfalcon", "Creature — Bird", eveLoyalGyrfalconOracle, 3, 3)
	if !hasEffectiveKeyword(t, g, bird, "defender") {
		t.Fatal("no defender to begin with")
	}
	eveCastColoured(t, g, me, "{W}")
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, bird, "defender") {
		t.Error("it kept defender after a white spell")
	}
	if !hasEffectiveKeyword(t, g, bird, "flying") {
		t.Error("it lost flying")
	}
}

// --- tap abilities ----------------------------------------------------------

func TestEve5590B1CinderPyromancerPingsAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	pyro := b12Push(g, me.ID, "Cinder Pyromancer", "Creature — Elemental Shaman", eveCinderPyromancerOracle, 0, 1)
	life := opp.Life
	apaActivate(t, g, me, pyro, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
	if !hmlTapped(t, g, pyro) {
		t.Error("the Pyromancer did not tap")
	}
}

func TestEve5590B1BallynockTrapperTapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	trapper := b12Push(g, me.ID, "Ballynock Trapper", "Creature — Kithkin Soldier", eveBallynockTrapperOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", eveBearType, 2, 2)
	apaActivate(t, g, me, trapper, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !hmlTapped(t, g, bear) {
		t.Error("the target is untapped")
	}
}

func TestEve5590B1MerrowLevitatorGrantsFlying(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	levitator := b12Push(g, me.ID, "Merrow Levitator", "Creature — Merfolk Wizard", eveMerrowLevitatorOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", eveBearType, 2, 2)
	apaActivate(t, g, me, levitator, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target has no flying")
	}
}

func TestEve5590B1WildernessHypnotistShrinksOnlyRedOrGreen(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	hypnotist := b12Push(g, me.ID, "Wilderness Hypnotist", "Creature — Merfolk Wizard", eveWildernessHypnotistOracle, 1, 3)
	blue := b16Creature(g, opp.ID, "Blue", "Creature — Merfolk", 2, 2, "U")
	red := b16Creature(g, opp.ID, "Red", "Creature — Goblin", 2, 2, "R")
	if err := g.ActivateCatalogAbility(me.ID, hypnotist, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(blue)}); err == nil {
		t.Fatal("the Hypnotist targeted a blue creature")
	}
	apaActivate(t, g, me, hypnotist, 0, game.ActivateAbilityParams{Targets: cardTarget(red)})
	if p, tough := ptOf(t, g, red); p != 0 || tough != 2 {
		t.Errorf("the red creature is %d/%d, want 0/2", p, tough)
	}
}

// --- mana activations on the creature itself ----------------------------------

func TestEve5590B1SelfActivations(t *testing.T) {
	for _, tc := range []struct {
		label, name, typeLine, oracle string
		power, toughness, row         int
		mana                          []string
		keyword                       string
		wantP, wantT                  int
	}{
		{"Talonrend +1/-1", "Talonrend", eveElementalType, eveTalonrendOracle, 0, 5, 0, []string{"R"}, "", 1, 4},
		{"Duergar Cave-Guard +1/+0", "Duergar Cave-Guard", "Creature — Dwarf Warrior", eveDuergarCaveGuardOracle, 1, 3, 0, []string{"W"}, "", 2, 3},
		{"Stream Hopper flying", "Stream Hopper", "Creature — Goblin", eveStreamHopperOracle, 1, 1, 0, []string{"U"}, "flying", 1, 1},
		{"Stillmoon Cavalier flying", "Stillmoon Cavalier", "Creature — Zombie Knight", eveStillmoonCavalierOracle, 2, 1, 0, []string{"W"}, "flying", 2, 1},
		{"Stillmoon Cavalier first strike", "Stillmoon Cavalier", "Creature — Zombie Knight", eveStillmoonCavalierOracle, 2, 1, 1, []string{"B"}, "first strike", 2, 1},
		{"Stillmoon Cavalier +1/+0", "Stillmoon Cavalier", "Creature — Zombie Knight", eveStillmoonCavalierOracle, 2, 1, 2, []string{"W", "B"}, "", 3, 1},
		{"Sturdy Hatchling shroud", "Sturdy Hatchling", eveElementalType, eveSturdyHatchlingOracle, 6, 6, 0, []string{"U"}, "shroud", 6, 6},
	} {
		t.Run(tc.label, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, id, tc.row, game.ActivateAbilityParams{})
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("no %s", tc.keyword)
			}
			if p, tough := ptOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestEve5590B1TwinbladeSlasherPumpsOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	slasher := b12Push(g, me.ID, "Twinblade Slasher", "Creature — Elf Warrior", eveTwinbladeSlasherOracle, 1, 1)
	apaMana(me, "C", "G")
	apaActivate(t, g, me, slasher, 0, game.ActivateAbilityParams{})
	if p, tough := ptOf(t, g, slasher); p != 3 || tough != 3 {
		t.Fatalf("%d/%d after the pump, want 3/3", p, tough)
	}
	apaMana(me, "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, slasher, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the pump was activated twice in one turn")
	}
}

func TestEve5590B1ShrewdHatchlingStopsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hatchling := b12Push(g, me.ID, "Shrewd Hatchling", eveElementalType, eveShrewdHatchlingOracle, 6, 6)
	bear := b12Creature(g, opp.ID, "Bear", eveBearType, 2, 2)
	other := b12Creature(g, opp.ID, "Other Bear", eveBearType, 2, 2)
	apaMana(me, "R")
	apaActivate(t, g, me, hatchling, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	brAttack(t, g, hatchling)
	if err := g.DeclareBlocker(bear, hatchling); err == nil {
		t.Error("the target blocked the Hatchling")
	}
	if err := g.DeclareBlocker(other, hatchling); err != nil {
		t.Errorf("another creature can't block it: %v", err)
	}
}

// --- regeneration -----------------------------------------------------------

func TestEve5590B1OdiousTrowRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	trow := b12Push(g, me.ID, "Odious Trow", "Creature — Troll", eveOdiousTrowOracle, 1, 1)
	apaMana(me, "C", "G")
	apaActivate(t, g, me, trow, 0, game.ActivateAbilityParams{})
	hobDestroy(t, g, trow)
	if !g.Battlefield.Contains(trow) {
		t.Error("the shielded Trow was destroyed")
	}
}

func TestEve5590B1DeityOfScarsRemovesACounterToRegenerate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	deity := castHoldCreature(t, g, "Deity of Scars", "Creature — Spirit Avatar", eveDeityOfScarsOracle, 7, 7)
	apaMana(me, "B")
	apaActivate(t, g, me, deity, 0, game.ActivateAbilityParams{CounterKind: game.CounterMinusOne})
	if got := minusCounters(g, deity); got != 1 {
		t.Fatalf("%d -1/-1 counters after paying, want 1", got)
	}
	hobDestroy(t, g, deity)
	if !g.Battlefield.Contains(deity) {
		t.Error("the shielded Deity was destroyed")
	}
}

func TestEve5590B1SutureSpiritRegeneratesTarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	spirit := b12Push(g, me.ID, "Suture Spirit", "Creature — Spirit", eveSutureSpiritOracle, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", eveBearType, 2, 2)
	apaMana(me, "W", "B", "W")
	apaActivate(t, g, me, spirit, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	hobDestroy(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the shielded Bear was destroyed")
	}
}

// --- the Skulkins -----------------------------------------------------------

func TestEve5590B1SkulkinsAffectOnlyTheirColour(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		mana         []string
		colour, bad  string
		keyword      string
		wantP, wantT int
	}{
		{"Hoof Skulkin", eveHoofSkulkinOracle, []string{"C", "C", "C"}, "G", "R", "", 3, 3},
		{"Jawbone Skulkin", eveJawboneSkulkinOracle, []string{"C", "C"}, "R", "G", "haste", 2, 2},
		{"Shell Skulkin", eveShellSkulkinOracle, []string{"C", "C", "C"}, "U", "B", "shroud", 2, 2},
		{"Fang Skulkin", eveFangSkulkinOracle, []string{"C", "C"}, "B", "U", game.KeywordWither, 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			skulkin := b12Push(g, me.ID, tc.name, "Artifact Creature — Scarecrow", tc.oracle, 2, 2)
			good := b16Creature(g, me.ID, "Good", eveBearType, 2, 2, tc.colour)
			bad := b16Creature(g, me.ID, "Bad", eveBearType, 2, 2, tc.bad)
			apaMana(me, tc.mana...)
			if err := g.ActivateCatalogAbility(me.ID, skulkin, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bad)}); err == nil {
				t.Fatal("it targeted a creature of the wrong colour")
			}
			apaActivate(t, g, me, skulkin, 0, game.ActivateAbilityParams{Targets: cardTarget(good)})
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, good, tc.keyword) {
				t.Errorf("the target has no %s", tc.keyword)
			}
			if p, tough := ptOf(t, g, good); p != tc.wantP || tough != tc.wantT {
				t.Errorf("the target is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

// --- statics and other triggers --------------------------------------------------

func TestEve5590B1SpitemareReflectsTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mare := b12Push(g, me.ID, "Spitemare", eveElementalType, eveSpitemareOracle, 3, 3)
	src := b16Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5, "R")
	before := opp.Life
	cmdr2B08Damage(g, src, mare, 2)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("the opponent lost %d, want 2", got)
	}
}

func TestEve5590B1OverbeingOfMythCountsYourHandAndDrawsExtra(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	d2dSetHand(me, 3)
	seedLibrary(me, "A", "B", "C", "D", "E")
	overbeing := b12Push(g, me.ID, "Overbeing of Myth", "Creature — Spirit Avatar", eveOverbeingOfMythOracle, 0, 0)
	if p, tough := ptOf(t, g, overbeing); p != 3 || tough != 3 {
		t.Fatalf("%d/%d with three cards in hand, want 3/3", p, tough)
	}
	advancePastCleanupForTest(t, g)
	hand := me.Hand.Size()
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
		if g.Turn.Step == game.StepDraw && g.Turn.ActiveSeat == seat {
			break
		}
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand = %d, want %d (the draw step's card and one more)", got, hand+2)
	}
}

func TestEve5590B1NoggleBanditIsBlockedOnlyByDefenders(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bandit := b12Push(g, me.ID, "Noggle Bandit", "Creature — Noggle Rogue", eveNoggleBanditOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", eveBearType, 2, 2)
	wall := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}})
	brAttack(t, g, bandit)
	if err := g.DeclareBlocker(bear, bandit); err == nil {
		t.Error("a creature without defender blocked the Bandit")
	}
	if err := g.DeclareBlocker(wall, bandit); err != nil {
		t.Errorf("a creature with defender can't block it: %v", err)
	}
}

func TestEve5590B1NobilisOfWarPumpsOnlyAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Nobilis of War", "Creature — Spirit Avatar", eveNobilisOfWarOracle, 3, 4)
	attacker := b12Creature(g, me.ID, "Attacker", eveBearType, 2, 2)
	home := b12Creature(g, me.ID, "Home", eveBearType, 2, 2)
	brAttack(t, g, attacker)
	if got := effectivePower(t, g, attacker); got != 4 {
		t.Errorf("the attacker has power %d, want 4", got)
	}
	if got := effectivePower(t, g, home); got != 2 {
		t.Errorf("a creature at home has power %d, want 2", got)
	}
}

func TestEve5590B1SpiritOfTheHearthGivesYouHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spirit := b12Push(g, me.ID, "Spirit of the Hearth", "Creature — Cat Spirit", eveSpiritOfTheHearthOracle, 4, 5)
	if !hasPlayerAbility(playerAbilities(g, me), KeywordHexproof) {
		t.Fatal("no hexproof with the Spirit out")
	}
	hobDestroy(t, g, spirit)
	if hasPlayerAbility(playerAbilities(g, me), KeywordHexproof) {
		t.Error("the hexproof outlived the Spirit")
	}
}

// --- the spells -------------------------------------------------------------

func TestEve5590B1DoubleCleaveGrantsDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := b12Creature(g, me.ID, "Bear", eveBearType, 2, 2)
	castCatalogSpell(t, g, "Double Cleave", "Instant", eveDoubleCleaveOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("the target has no double strike")
	}
}

func TestEve5590B1UnmakeExilesACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	bear := b12Creature(g, opp.ID, "Bear", eveBearType, 2, 2)
	castCatalogSpell(t, g, "Unmake", "Instant", eveUnmakeOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || opp.Graveyard.Contains(bear) {
		t.Error("the creature was not exiled")
	}
}

func TestEve5590B1HallowedBurialTucksEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := b12Creature(g, me.ID, "Mine", eveBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", eveBearType, 2, 2)
	myLib, theirLib := me.Library.Size(), opp.Library.Size()
	castCatalogSpell(t, g, "Hallowed Burial", "Sorcery", eveHallowedBurialOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("a creature stayed on the battlefield")
	}
	if me.Library.Size() != myLib+1 || opp.Library.Size() != theirLib+1 {
		t.Errorf("libraries %d and %d, want %d and %d", me.Library.Size(), opp.Library.Size(), myLib+1, theirLib+1)
	}
}

func TestEve5590B1InundateReturnsOnlyNonblueCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	blue := b16Creature(g, opp.ID, "Blue", "Creature — Merfolk", 2, 2, "U")
	green := b16Creature(g, opp.ID, "Green", eveBearType, 2, 2, "G")
	hand := opp.Hand.Size()
	castCatalogSpell(t, g, "Inundate", "Sorcery", eveInundateOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(blue) {
		t.Error("the blue creature was returned")
	}
	if g.Battlefield.Contains(green) {
		t.Error("the green creature stayed")
	}
	if got := opp.Hand.Size(); got != hand+1 {
		t.Errorf("the owner's hand is %d, want %d", got, hand+1)
	}
}

func TestEve5590B1BanishingKnackGrantsABounce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b16Creature(g, me.ID, "Bear", eveBearType, 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", eveBearType, 2, 2, "G")
	dgCastOn(t, g, "Banishing Knack", eveBanishingKnackOracle, mine)
	idx, ref := grantedActivatedIndex(t, g, mine)
	if idx < 0 {
		t.Fatal("the target has no granted bounce")
	}
	b16Activate(t, g, me.ID, mine, idx, game.ActivateAbilityParams{Ref: ref, Targets: dgCardRef(theirs)})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, theirs) != nil {
		t.Error("the target nonland permanent was not bounced")
	}
	if !b16Tapped(t, g, mine) {
		t.Error("the bounce did not tap the creature that has it")
	}
	advancePastCleanupForTest(t, g)
	if idx, _ := grantedActivatedIndex(t, g, mine); idx >= 0 {
		t.Error("the granted bounce outlived the turn")
	}
}

// --- enters and upkeep triggers --------------------------------------------------

func TestEve5590B1WistfulSelkieDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Wistful Selkie", "Creature — Merfolk Wizard", eveWistfulSelkieOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestEve5590B1FlickerwispReturnsTheCardAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	bear := b12Creature(g, opp.ID, "Bear", eveBearType, 2, 2)
	castHoldCreature(t, g, "Flickerwisp", eveElementalType, eveFlickerwispOracle, 3, 1)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if _, ok := visBattlefieldNamed(g, opp.ID, "Bear"); ok {
		t.Fatal("the target was not exiled")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if _, ok := visBattlefieldNamed(g, opp.ID, "Bear"); !ok {
		t.Error("the exiled card did not return at the end step")
	}
}

func TestEve5590B1NoggleBridgebreakerReturnsALand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	land := b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	castHoldCreature(t, g, "Noggle Bridgebreaker", "Creature — Noggle Rogue", eveNoggleBridgebreakerOracle, 4, 3)
	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("no land return was asked")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if !me.Hand.Contains(land) {
		t.Error("the land is not in its owner's hand")
	}
}

func TestEve5590B1CacheRaidersReturnsAPermanentEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Cache Raiders", "Creature — Merfolk Rogue", eveCacheRaidersOracle, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", eveBearType, 2, 2)
	walkToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("no return was asked")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if !me.Hand.Contains(bear) {
		t.Error("the Bear is not in its owner's hand")
	}
}
