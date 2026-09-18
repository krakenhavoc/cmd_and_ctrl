package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch30_test.go — card-level coverage for the card-coverage
// roadmap's batch 30 (#393, `edhrec_rank` 3145–3244): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, block, land play or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b30-prefixed.

const (
	b30StonespeakerCrystalOracle    = "a494fcee-6885-434c-aad5-6f83640c4472"
	b30LunarConvocationOracle       = "cb68d1e4-36aa-4a00-b671-8959e7d526d4"
	b30TolarianWindsOracle          = "686cce6e-18ec-45b0-8d2e-74fa353a905e"
	b30WoodlandChasmOracle          = "bf5482b6-dd3e-4fb7-bc62-29e23b417a5f"
	b30RiversRebukeAlreadyOracle    = "c52cfb41-18f3-4e73-b5e7-d75baf74e578"
	b30DamnablePactOracle           = "28dd0ae4-fc56-4abf-af24-6c6c8d0a10cd"
	b30HonoredDreyleaderOracle      = "44bffa3e-4d31-4f9b-829c-e2066ab2f641"
	b30GoblinTrashmasterOracle      = "0283bf5e-ddf2-4a4a-a7cf-d3e27eed7e7d"
	b30KhalniGardenOracle           = "b2d5ba45-8674-4428-89db-c2bbbf0bf5c5"
	b30WindgracesJudgmentOracle     = "de1ca6ed-b275-4f62-ba05-f31b3659b352"
	b30ThrabenCharmOracle           = "7eccd10d-f224-4467-af25-f2b768993f9f"
	b30SurlyBadgersaurOracle        = "0209dc74-ac49-4deb-907a-e9fa49d27a0f"
	b30PromisingVeinOracle          = "861eb7d7-7616-4620-a4fd-4b8c3bf00dd1"
	b30OppressionOracle             = "f488f679-f5d5-4e61-b0c8-0b0eb622df0d"
	b30PickYourPoisonOracle         = "9af4a832-d634-47aa-91ed-79d44fe08864"
	b30KeeperOfFablesOracle         = "c8ca3116-e0f0-4e27-aa0f-99ed85927040"
	b30IronSpiderOracle             = "e123fd7d-ace9-48a4-9510-eedcc837d8e8"
	b30FavorableWindsOracle         = "2361ca87-6352-4ba3-8d91-b3d71242914d"
	b30CanopyTacticianOracle        = "8b20d6f4-6322-4435-92d5-acaae74774f4"
	b30VorelOracle                  = "3001f971-c4db-4176-89b5-e7f4a8890c0f"
	b30HourOfPromiseOracle          = "51ff3e53-90bb-41bf-80e4-bb3fd51d574a"
	b30SatoruOracle                 = "7555c429-5f2d-4171-b6b0-8e3c8da7f314"
	b30BetorOracle                  = "826a0db6-195a-4697-9bd6-f544b279e3cf"
	b30WakeningSunsAvatarOracle     = "3c2aec69-ffd9-4a34-888c-58adbbb99bb5"
	b30SavvyHunterOracle            = "602132c2-8ee8-41f8-bfac-cb17d32203f5"
	b30ExtractFromDarknessOracle    = "e597d8a1-3bbc-4001-b642-f4421447970f"
	b30DocksideChefOracle           = "fed12a16-8920-403c-be63-0601a9d864b0"
	b30GrimGuardianOracle           = "c1f1babf-13d0-4fc4-b192-127d2d5db7f1"
	b30UltimaSkipOracle             = "baa337ce-edc6-4ee5-a898-68e9dbb4ab93"
	b30GemhideSliverSkipOracle      = "2c09ca09-8e62-4fe3-9b3d-61573dd2ffbc"
	b30ChainOfSmogSkipOracle        = "ea14c26b-bf2f-48b4-b879-6e63069ded1f"
	b30ZimoneParadoxSculptorSkipOID = "9dd674a7-becf-4106-b53f-bca88426d92d"
)

// b30Lives snapshots every seat's life total.
func b30Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// b30TokensNamed counts the battlefield tokens `controller` controls
// with the given name.
func b30TokensNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && IsToken(c) {
			n++
		}
	}
	return n
}

// b30Flyer seeds a non-catalog creature with printed flying.
func b30Flyer(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Keywords: []string{"flying"},
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b30HandCard seeds a card into a player's hand.
func b30HandCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// b30DiscardOne discards one card from a player's hand through the
// effect path, so discard triggers see it. With one card in hand the
// random pick is that card.
func b30DiscardOne(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(player, 1); err != nil {
			t.Fatalf("DiscardRandomForEffect: %v", err)
		}
	})
}

// b30Reanimate puts a graveyard card onto the battlefield under
// `controller` through the effect path.
func b30Reanimate(t *testing.T, g *game.Game, card, controller uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardUnderControlForEffect(card, game.ZoneBattlefield, controller); err != nil {
			t.Fatalf("ReturnFromGraveyardUnderControlForEffect: %v", err)
		}
	})
}

// b30SettleOrder answers a trigger-order prompt if one is open, then
// passes priority until the stack is empty.
func b30SettleOrder(t *testing.T, g *game.Game) {
	t.Helper()
	if triggerOrderPrompt(g) != nil {
		answerTriggerOrderInOfferedOrder(t, g)
	}
	passPriorityAroundTable(t, g)
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. River's Rebuke
// was already on main from the S23 boardwipe work and is pinned here
// so the table matches the issue's 32 minus the four declared skips.
func TestBatch30CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b30StonespeakerCrystalOracle: "Stonespeaker Crystal",
		b30LunarConvocationOracle:    "Lunar Convocation",
		b30TolarianWindsOracle:       "Tolarian Winds",
		b30WoodlandChasmOracle:       "Woodland Chasm",
		b30RiversRebukeAlreadyOracle: "River's Rebuke",
		b30DamnablePactOracle:        "Damnable Pact",
		b30HonoredDreyleaderOracle:   "Honored Dreyleader",
		b30GoblinTrashmasterOracle:   "Goblin Trashmaster",
		b30KhalniGardenOracle:        "Khalni Garden",
		b30WindgracesJudgmentOracle:  "Windgrace's Judgment",
		b30ThrabenCharmOracle:        "Thraben Charm",
		b30SurlyBadgersaurOracle:     "Surly Badgersaur",
		b30PromisingVeinOracle:       "Promising Vein",
		b30OppressionOracle:          "Oppression",
		b30PickYourPoisonOracle:      "Pick Your Poison",
		b30KeeperOfFablesOracle:      "Keeper of Fables",
		b30IronSpiderOracle:          "Iron Spider, Stark Upgrade",
		b30FavorableWindsOracle:      "Favorable Winds",
		b30CanopyTacticianOracle:     "Canopy Tactician",
		b30VorelOracle:               "Vorel of the Hull Clade",
		b30HourOfPromiseOracle:       "Hour of Promise",
		b30SatoruOracle:              "Satoru, the Infiltrator",
		b30BetorOracle:               "Betor, Kin to All",
		b30WakeningSunsAvatarOracle:  "Wakening Sun's Avatar",
		b30SavvyHunterOracle:         "Savvy Hunter",
		b30ExtractFromDarknessOracle: "Extract from Darkness",
		b30DocksideChefOracle:        "Dockside Chef",
		b30GrimGuardianOracle:        "Grim Guardian",
	}
	if len(want) != 28 {
		t.Fatalf("the batch registers 27 cards plus one already on main, the table lists %d", len(want))
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
	// The four declared skips must NOT be registered — each needs a
	// seam the engine does not have, and a spec would ship the card
	// stronger than printed or as something other than itself.
	for _, skipped := range []string{
		b30UltimaSkipOracle,             // a land losing all types and abilities and gaining a mana ability; a tap-for-{C} rider
		b30GemhideSliverSkipOracle,      // a mana ability granted to other permanents by a static
		b30ChainOfSmogSkipOracle,        // a spell copying itself at resolution on the target player's say-so
		b30ZimoneParadoxSculptorSkipOID, // a beginning-of-combat trigger event
	} {
		if _, ok := Lookup(skipped); ok {
			t.Errorf("%s is a declared skip and must not be registered", skipped)
		}
	}
}

// --- lands ---------------------------------------------------------

func TestB30WoodlandChasmEntersTappedAndTapsForEitherColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12PlayFromHand(t, g, "Woodland Chasm", "Snow Land — Swamp Forest", b30WoodlandChasmOracle, game.CastSpellParams{})
	if !b12Card(t, g, land).Tapped {
		t.Fatal("the snow dual enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b28TapForMana(t, g, me.ID, land, "B")
	if got := poolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("tapped for {B}: pool %v", got)
	}
}

func TestB30KhalniGardenEntersTappedWithAPlant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12PlayFromHand(t, g, "Khalni Garden", "Land", b30KhalniGardenOracle, game.CastSpellParams{})
	if !b12Card(t, g, land).Tapped {
		t.Fatal("the Garden enters tapped")
	}
	if len(g.PendingTriggers) == 0 && triggerOnStack(g, land) == nil {
		t.Fatal("the Plant is an enters trigger, not a silent side effect")
	}
	passPriorityAroundTable(t, g)
	if got := b30TokensNamed(g, me.ID, "Plant"); got != 1 {
		t.Fatalf("one Plant: %d", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Plant" && IsToken(c) {
			if p, tough := effectivePower(t, g, c.InstanceID), effectiveToughness(t, g, c.InstanceID); p != 0 || tough != 1 {
				t.Errorf("a 0/1 Plant: %d/%d", p, tough)
			}
		}
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b28TapForMana(t, g, me.ID, land, "")
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("tapped for {G}: pool %v", got)
	}
}

func TestB30PromisingVeinTapsForColorlessAndFetchesABasicTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vein := seedLand(g, me.ID, "Promising Vein", "Land — Cave", b30PromisingVeinOracle)
	b28TapForMana(t, g, me.ID, vein, "")
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("{T}: Add {C}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vein) })
	lib := seedSearchLibrary(me,
		game.Card{Name: "Snow-Covered Forest", TypeLine: "Basic Snow Land — Forest"},
		game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, vein, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(vein) {
		t.Fatal("the Vein is sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("one basic in the library is no choice — it is taken outright")
	}
	if !g.Battlefield.Contains(lib[0]) {
		t.Fatal("a snow-covered basic is a basic land card and is fetched")
	}
	if !b12Card(t, g, lib[0]).Tapped {
		t.Error("the fetched basic enters tapped")
	}
	if g.Battlefield.Contains(lib[1]) {
		t.Error("a nonbasic Forest is not a basic land card")
	}
}

// --- spells --------------------------------------------------------

func TestB30TolarianWindsDiscardsTheHandAndDrawsThatMany(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	lib := me.Library.Size()
	castCatalogSpell(t, g, "Tolarian Winds", "Instant", b30TolarianWindsOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Errorf("discard the hand, draw that many: %d → %d", before, me.Hand.Size())
	}
	if me.Library.Size() != lib-before {
		t.Errorf("%d cards drawn: library %d → %d", before, lib, me.Library.Size())
	}
	if me.Graveyard.Size() != before+1 {
		t.Errorf("the old hand and the Winds are in the graveyard: %d", me.Graveyard.Size())
	}
}

func TestB30DamnablePactDrawsXAndLosesX(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	life, hand := opp.Life, opp.Hand.Size()
	castXSpell(t, g, "Damnable Pact", "Sorcery", b30DamnablePactOracle, "{X}{B}{B}", 3, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+3 {
		t.Errorf("target player draws X: %d → %d", hand, opp.Hand.Size())
	}
	if opp.Life != life-3 {
		t.Errorf("and loses X: %d → %d", life, opp.Life)
	}
}

func TestB30WindgracesJudgmentDestroysOnePermanentPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	aBear := b12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	aRock := b12Permanent(g, a.ID, "A's Rock", "Artifact")
	bAura := b12Permanent(g, b.ID, "B's Enchantment", "Enchantment")
	bLand := seedLand(g, b.ID, "Forest", "Basic Land — Forest", "")
	legal := legalCards(g, me.ID, b30WindgracesJudgmentOracle)
	if legal[mine] || legal[bLand] || !legal[aBear] || !legal[bAura] {
		t.Fatal("only nonland permanents opponents control are legal")
	}
	castCatalogSpell(t, g, "Windgrace's Judgment", "Instant", b30WindgracesJudgmentOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: aBear}, {Kind: game.TargetCard, ID: aRock}, {Kind: game.TargetCard, ID: bAura},
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aBear) || g.Battlefield.Contains(bAura) {
		t.Error("the first permanent named for each opponent is destroyed")
	}
	if !g.Battlefield.Contains(aRock) {
		t.Error("a second permanent of the same opponent's is skipped — one per opponent")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own permanents are untouched")
	}
	if spec, _ := Lookup(b30WindgracesJudgmentOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the one-per-opponent enforcement at resolution is a declared gap")
	}
}

func TestB30ThrabenCharmThreeModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushVanillaCreature(g, me.ID, "Soldier A", 1, 1)
	pushVanillaCreature(g, me.ID, "Soldier B", 1, 1)
	pushVanillaCreature(g, me.ID, "Soldier C", 1, 1)
	golem := b12Creature(g, opp.ID, "Their Golem", "Creature — Golem", 6, 6)
	castModal(t, g, "Thraben Charm", "Instant", b30ThrabenCharmOracle, []int{0}, b16TargetCard(golem))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Error("three creatures is 6 damage — the 6/6 dies")
	}
	aura := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	castModal(t, g, "Thraben Charm", "Instant", b30ThrabenCharmOracle, []int{1}, b16TargetCard(aura))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("mode two destroys the enchantment")
	}
	dead := b17GraveyardCard(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	spare := b17GraveyardCard(g.Seats[2], "Spare Bear", "Creature — Bear", "{1}{G}")
	castModal(t, g, "Thraben Charm", "Instant", b30ThrabenCharmOracle, []int{2}, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) || !g.Exile.Contains(dead) {
		t.Error("the targeted player's graveyard is exiled")
	}
	if !g.Seats[2].Graveyard.Contains(spare) {
		t.Error("a player not named keeps their graveyard")
	}
}

func TestB30PickYourPoisonEdictsEachOpponentByMode(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	myFlyer := b30Flyer(g, me.ID, "My Bird", "Creature — Bird", 1, 1)
	aFlyer := b30Flyer(g, a.ID, "A's Bird", "Creature — Bird", 1, 1)
	aBear := b12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	bRock := b12Permanent(g, b.ID, "B's Rock", "Artifact")
	castModal(t, g, "Pick Your Poison", "Sorcery", b30PickYourPoisonOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)
	c := sacrificeChoiceFor(g, a.ID)
	if c == nil {
		t.Fatal("an opponent with a flyer is asked to sacrifice one")
	}
	if hasID(c.SacrificeOptions, aBear) || !hasID(c.SacrificeOptions, aFlyer) {
		t.Error("only creatures with flying are offered")
	}
	if sacrificeChoiceFor(g, b.ID) != nil || sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("an opponent with no flyer, and the caster, are not asked")
	}
	answerSacrifice(t, g, a.ID, aFlyer)
	if g.Battlefield.Contains(aFlyer) || !g.Battlefield.Contains(myFlyer) {
		t.Error("their flyer is sacrificed; yours stays")
	}
	castModal(t, g, "Pick Your Poison", "Sorcery", b30PickYourPoisonOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if c := sacrificeChoiceFor(g, b.ID); c == nil || !hasID(c.SacrificeOptions, bRock) {
		t.Fatal("mode one asks for an artifact")
	}
	answerSacrifice(t, g, b.ID, bRock)
	if g.Battlefield.Contains(bRock) {
		t.Error("the artifact is sacrificed")
	}
}

func TestB30HourOfPromiseFetchesTwoLandsTappedAndMakesZombiesWithThreeDeserts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLand(g, me.ID, "Desert of the Glorified", "Land — Desert", "")
	lib := seedSearchLibrary(me,
		game.Card{Name: "Desert A", TypeLine: "Land — Desert"},
		game.Card{Name: "Desert B", TypeLine: "Land — Desert"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	castCatalogSpell(t, g, "Hour of Promise", "Sorcery", b30HourOfPromiseOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(lib[0]) || !g.Battlefield.Contains(lib[1]) {
		t.Fatal("both land cards are fetched")
	}
	if !b12Card(t, g, lib[0]).Tapped || !b12Card(t, g, lib[1]).Tapped {
		t.Error("they enter tapped")
	}
	if got := b30TokensNamed(g, me.ID, "Zombie"); got != 2 {
		t.Errorf("three Deserts after the search is two Zombies: %d", got)
	}
	// Without three Deserts, no Zombies.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	seedSearchLibrary(me2,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Desert", TypeLine: "Land — Desert"},
	)
	castCatalogSpell(t, g2, "Hour of Promise", "Sorcery", b30HourOfPromiseOracle, nil)
	passPriorityAroundTable(t, g2)
	if got := b30TokensNamed(g2, me2.ID, "Zombie"); got != 0 {
		t.Errorf("one Desert is no Zombies: %d", got)
	}
}

func TestB30ExtractFromDarknessMillsEveryoneThenReanimatesTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := b17GraveyardCard(opp, "Their Dead Golem", "Creature — Golem", "{4}")
	bolt := b17GraveyardCard(opp, "Spent Bolt", "Instant", "{R}")
	if legal := legalCards(g, me.ID, b30ExtractFromDarknessOracle); !legal[dead] || legal[bolt] {
		t.Fatal("a creature card in any graveyard is legal; an instant is not")
	}
	libs := []int{me.Library.Size(), opp.Library.Size(), g.Seats[2].Library.Size(), g.Seats[3].Library.Size()}
	castCatalogSpell(t, g, "Extract from Darkness", "Sorcery", b30ExtractFromDarknessOracle, b16TargetCard(dead))
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Library.Size() != libs[i]-2 {
			t.Errorf("seat %d mills two: %d → %d", i, libs[i], p.Library.Size())
		}
	}
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the Golem is on the battlefield")
	}
	if c := b12Card(t, g, dead); c.Controller != me.ID {
		t.Error("under the caster's control")
	}
	if spec, _ := Lookup(b30ExtractFromDarknessOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the announce-time choice is a declared gap")
	}
}

// --- artifacts and enchantments -------------------------------------

func TestB30StonespeakerCrystalTapsForTwoAndEatsGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	crystal := b12Push(g, me.ID, "Stonespeaker Crystal", "Artifact", b30StonespeakerCrystalOracle, 0, 0)
	b28TapForMana(t, g, me.ID, crystal, "")
	if got := poolColors(me); len(got) != 2 || got[0] != "C" || got[1] != "C" {
		t.Errorf("{T}: Add {C}{C}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(crystal) })
	aDead := b17GraveyardCard(a, "A's Dead Bear", "Creature — Bear", "{1}{G}")
	bDead := b17GraveyardCard(b, "B's Dead Bear", "Creature — Bear", "{1}{G}")
	mine := b17GraveyardCard(me, "My Dead Bear", "Creature — Bear", "{1}{G}")
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, crystal, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetPlayer, ID: b.ID},
	}})
	if g.Battlefield.Contains(crystal) {
		t.Fatal("sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(aDead) || !g.Exile.Contains(bDead) {
		t.Error("both named players' graveyards are exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("a player not named keeps their graveyard")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("then draw a card: %d → %d", hand, me.Hand.Size())
	}
	// "Any number" includes none: just the draw.
	crystal2 := b12Push(g, me.ID, "Stonespeaker Crystal", "Artifact", b30StonespeakerCrystalOracle, 0, 0)
	hand = me.Hand.Size()
	b16Activate(t, g, me.ID, crystal2, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || !me.Graveyard.Contains(mine) {
		t.Error("with no targets it exiles nothing and still draws")
	}
}

func TestB30LunarConvocationDrainsOnGainAndMakesABatOnGainAndLoss(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	conv := b12Push(g, me.ID, "Lunar Convocation", "Enchantment", b30LunarConvocationOracle, 0, 0)
	advanceToMain(t, g)
	b14Gain(t, g, me.ID, 3)
	before := b30Lives(g)
	batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	b30SettleOrder(t, g)
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != before[i]-1 {
			t.Errorf("gained life this turn: opponent %d loses 1: %d → %d", i, before[i], g.Seats[i].Life)
		}
	}
	if got := b30TokensNamed(g, me.ID, "Bat"); got != 0 {
		t.Errorf("gained but not lost: no Bat: %d", got)
	}
	// Next turn: pay 2 life for a card (a loss) and gain — both fire.
	b12ToMyNextUpkeep(t, g)
	advanceToMain(t, g)
	hand, life := me.Hand.Size(), me.Life
	b16Activate(t, g, me.ID, conv, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-2 {
		t.Fatalf("{1}{B}, Pay 2 life: draw a card: hand %d → %d, life %d → %d", hand, me.Hand.Size(), life, me.Life)
	}
	b14Gain(t, g, me.ID, 1)
	before = b30Lives(g)
	batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	if triggerOrderPrompt(g) == nil {
		t.Fatal("both triggers fire and the controller orders them")
	}
	b30SettleOrder(t, g)
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != before[i]-1 {
			t.Errorf("opponent %d loses 1 again: %d → %d", i, before[i], g.Seats[i].Life)
		}
	}
	if got := b30TokensNamed(g, me.ID, "Bat"); got != 1 {
		t.Errorf("gained and lost: one Bat: %d", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bat" && IsToken(c) {
			assertKeywords(t, g, c.InstanceID, "flying")
		}
	}
	// A turn with neither: nothing.
	b12ToMyNextUpkeep(t, g)
	before = b30Lives(g)
	batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	b30SettleOrder(t, g)
	if g.Seats[1].Life != before[1] || b30TokensNamed(g, me.ID, "Bat") != 1 {
		t.Error("no life gained this turn: no drain, no Bat")
	}
}

func TestB30OppressionMakesEachCasterDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Oppression", "Enchantment", b30OppressionOracle, 0, 0)
	castCatalogSpell(t, g, "My Bear", "Creature — Bear", "", nil)
	if len(g.PendingTriggers) == 0 && triggerOnStack(g, findBattlefieldByName(g, "Oppression")) == nil {
		t.Fatal("casting a spell triggers Oppression")
	}
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, me.ID); got != 1 {
		t.Errorf("the caster owes a discard, got %d", got)
	}
	if discardOwed(g, opp.ID) != 0 {
		t.Error("nobody else does")
	}
	card, _ := me.Hand.Top()
	answerDiscard(t, g, me.ID, card.InstanceID)
	if !me.Graveyard.Contains(card.InstanceID) {
		t.Error("the chosen card is discarded")
	}
	b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, opp.ID); got != 1 {
		t.Errorf("an opponent's spell costs the opponent a card, got %d", got)
	}
}

func TestB30FavorableWindsPumpsYourFlyersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Favorable Winds", "Enchantment", b30FavorableWindsOracle, 0, 0)
	bird := b30Flyer(g, me.ID, "My Bird", "Creature — Bird", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b30Flyer(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1)
	if p, tough := effectivePower(t, g, bird), effectiveToughness(t, g, bird); p != 2 || tough != 2 {
		t.Errorf("a creature you control with flying is +1/+1: %d/%d", p, tough)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("one without flying is unchanged: %d", p)
	}
	if p := effectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's flyer is unchanged: %d", p)
	}
}

// --- creatures -----------------------------------------------------

func TestB30HonoredDreyleaderCountsSquirrelsAndFoodAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "My Squirrel", "Creature — Squirrel", 1, 1)
	b12Permanent(g, me.ID, "Food", "Token Artifact — Food")
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Squirrel", "Creature — Squirrel", 1, 1)
	leader := castCatalogSpell(t, g, "Honored Dreyleader", "Creature — Squirrel Warrior", b30HonoredDreyleaderOracle, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, leader, "+1/+1"); got != 2 {
		t.Fatalf("one Squirrel and one Food you control: %d counters", got)
	}
	assertKeywords(t, g, leader, "trample")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, FoodToken(), 1) })
	passPriorityAroundTable(t, g)
	if got := counterCount(g, leader, "+1/+1"); got != 3 {
		t.Errorf("a Food entering grows it: %d", got)
	}
	castCatalogSpell(t, g, "Another Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, leader, "+1/+1"); got != 3 {
		t.Errorf("a Bear does not: %d", got)
	}
	b13PlayAs(t, g, 1, "Their Other Squirrel", "Creature — Squirrel", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, leader, "+1/+1"); got != 3 {
		t.Errorf("an opponent's Squirrel does not: %d", got)
	}
}

func TestB30GoblinTrashmasterLordsGoblinsAndEatsOneToKillAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := b12Push(g, me.ID, "Goblin Trashmaster", "Creature — Goblin Warrior", b30GoblinTrashmasterOracle, 3, 3)
	goblin := b12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if p := effectivePower(t, g, goblin); p != 2 {
		t.Errorf("another Goblin you control is +1/+1: %d", p)
	}
	if p := effectivePower(t, g, master); p != 3 {
		t.Errorf("the Trashmaster is 'other': %d", p)
	}
	if p := effectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Goblin is unchanged: %d", p)
	}
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear}, Targets: b16TargetCard(rock),
	}); err == nil {
		t.Fatal("a Bear is not a Goblin")
	}
	b16Activate(t, g, me.ID, master, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{goblin}, Targets: b16TargetCard(rock)})
	if g.Battlefield.Contains(goblin) {
		t.Fatal("the Goblin is sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
}

func TestB30SurlyBadgersaurThreeDiscardPayoffs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	badger := b12Push(g, me.ID, "Surly Badgersaur", "Creature — Badger Dinosaur", b30SurlyBadgersaurOracle, 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	emptyHandToLibrary(g, me)
	advanceToMain(t, g)
	b30HandCard(me, "Some Bear", "Creature — Bear")
	b30DiscardOne(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, badger, "+1/+1"); got != 1 {
		t.Errorf("discarding a creature card grows it: %d", got)
	}
	b30HandCard(me, "Forest", "Basic Land — Forest")
	b30DiscardOne(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := b30TokensNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("discarding a land card makes a Treasure: %d", got)
	}
	b30HandCard(me, "Some Bolt", "Instant")
	b30DiscardOne(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("discarding a noncreature, nonland card asks which creature to fight")
	}
	if hasID(p.PickTargetCards, badger) || !hasID(p.PickTargetCards, theirs) {
		t.Error("only creatures you don't control are offered")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 4/4 Badgersaur kills the 2/2")
	}
	if c := b12Card(t, g, badger); c.DamageMarked != 2 {
		t.Errorf("and takes 2 back: %d marked", c.DamageMarked)
	}
	if got := counterCount(g, badger, "+1/+1"); got != 1 || b30TokensNamed(g, me.ID, "Treasure") != 1 {
		t.Error("the other two triggers do not fire on an instant")
	}
	// Nothing to fight: the trigger is removed without a prompt.
	b30HandCard(me, "Another Bolt", "Instant")
	b30DiscardOne(t, g, me.ID)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("with no creature you don't control there is nothing to ask")
	}
	passPriorityAroundTable(t, g)
	// An opponent's discard is not yours.
	b30HandCard(opp, "Their Land", "Basic Land — Swamp")
	b30DiscardOne(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := b30TokensNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("an opponent's discard makes nothing: %d", got)
	}
}

func TestB30KeeperOfFablesDrawsOncePerNonHumanCombatDamageBatch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Keeper of Fables", "Creature — Cat", b30KeeperOfFablesOracle, 4, 5)
	cat := b12Creature(g, me.ID, "My Cat", "Creature — Cat", 2, 2)
	dog := b12Creature(g, me.ID, "My Dog", "Creature — Dog", 2, 2)
	human := b12Creature(g, me.ID, "My Human", "Creature — Human Soldier", 2, 2)
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, cat, dog, human)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("two non-Humans connecting is one draw: %d → %d", hand, me.Hand.Size())
	}
	b12ToMyNextUpkeep(t, g)
	advanceToMain(t, g)
	hand = me.Hand.Size()
	attackWith(t, g, opp.ID, human)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a Human alone draws nothing: %d → %d", hand, me.Hand.Size())
	}
}

func TestB30IronSpiderPumpsArtifactCreaturesAndVehicles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := b12Push(g, me.ID, "Iron Spider, Stark Upgrade", "Legendary Artifact Creature — Spider Hero", b30IronSpiderOracle, 2, 3)
	construct := b12Creature(g, me.ID, "My Construct", "Artifact Creature — Construct", 1, 1)
	vehicle := b12Permanent(g, me.ID, "My Vehicle", "Artifact — Vehicle")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := b12Creature(g, opp.ID, "Their Construct", "Artifact Creature — Construct", 1, 1)
	assertKeywords(t, g, spider, "vigilance")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, spider, 0, game.ActivateAbilityParams{})
	if !b12Card(t, g, spider).Tapped {
		t.Error("tapped to activate")
	}
	passPriorityAroundTable(t, g)
	for _, want := range []struct {
		id   uuid.UUID
		n    int
		what string
	}{
		{spider, 1, "the Spider itself"}, {construct, 1, "an artifact creature"}, {vehicle, 1, "a Vehicle"},
		{bear, 0, "a non-artifact creature"}, {rock, 0, "a non-creature, non-Vehicle artifact"}, {theirs, 0, "an opponent's artifact creature"},
	} {
		if got := counterCount(g, want.id, "+1/+1"); got != want.n {
			t.Errorf("%s: %d counters, want %d", want.what, got, want.n)
		}
	}
	// #789 made the draw ability real: "from among artifacts you
	// control" is a counter removal split across permanents, and the
	// card no longer ships with a gap.
	if spec, _ := Lookup(b30IronSpiderOracle); len(spec.Activated) != 2 || spec.Completeness != CompletenessFull {
		t.Error("the counter-removal draw should be live and the card complete (#789)")
	}
}

func TestB30CanopyTacticianLordsElvesAndTapsForThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tactician := b12Push(g, me.ID, "Canopy Tactician", "Creature — Elf Warrior", b30CanopyTacticianOracle, 3, 3)
	elf := b12Creature(g, me.ID, "My Elf", "Creature — Elf Druid", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	if p := effectivePower(t, g, elf); p != 2 {
		t.Errorf("another Elf you control is +1/+1: %d", p)
	}
	if p := effectivePower(t, g, tactician); p != 3 {
		t.Errorf("the Tactician is 'other': %d", p)
	}
	if p := effectivePower(t, g, bear); p != 2 || effectivePower(t, g, theirs) != 1 {
		t.Error("a Bear and an opponent's Elf are unchanged")
	}
	b28TapForMana(t, g, me.ID, tactician, "")
	if got := poolColors(me); len(got) != 3 || got[0] != "G" || got[1] != "G" || got[2] != "G" {
		t.Errorf("{T}: Add {G}{G}{G}: pool %v", got)
	}
}

func TestB30VorelDoublesEveryKindOfCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vorel := b12Push(g, me.ID, "Vorel of the Hull Clade", "Legendary Creature — Human Merfolk", b30VorelOracle, 1, 4)
	hydra := b12Creature(g, opp.ID, "Their Hydra", "Creature — Hydra", 0, 0)
	b28AddCounter(t, g, hydra, 3)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hydra, "charge", 2) })
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, vorel, 0, game.ActivateAbilityParams{Targets: b16TargetCard(hydra)})
	if !b12Card(t, g, vorel).Tapped {
		t.Error("tapped to activate")
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, hydra, "+1/+1"); got != 6 {
		t.Errorf("three +1/+1 counters become six: %d", got)
	}
	if got := counterCount(g, hydra, "charge"); got != 4 {
		t.Errorf("two charge counters become four: %d", got)
	}
	// No counters: nothing to double, and the ability still resolves.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vorel) })
	b16Activate(t, g, me.ID, vorel, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, "+1/+1"); got != 0 {
		t.Errorf("zero doubled is zero: %d", got)
	}
}

func TestB30SatoruDrawsWhenCreaturesArriveUncast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	satoru := castCatalogSpell(t, g, "Satoru, the Infiltrator", "Legendary Creature — Human Ninja Rogue", b30SatoruOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("a cast Satoru draws nothing: %d → %d", hand, me.Hand.Size())
	}
	assertKeywords(t, g, satoru, "menace")
	dead := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	b30Reanimate(t, g, dead, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a reanimated creature was not cast: draw: %d → %d", hand, me.Hand.Size())
	}
	castCatalogSpell(t, g, "Cast Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a cast creature draws nothing: %d", me.Hand.Size())
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 2) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("tokens draw nothing: %d", me.Hand.Size())
	}
	theirs := b17GraveyardCard(opp, "Their Dead Bear", "Creature — Bear", "{1}{G}")
	b30Reanimate(t, g, theirs, opp.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("an opponent's reanimation draws nothing: %d", me.Hand.Size())
	}
	// Satoru itself arriving uncast draws.
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(satoru) })
	me.Hand.Remove(satoru)
	me.Graveyard.PushTop(game.Card{InstanceID: satoru, Name: "Satoru, the Infiltrator", TypeLine: "Legendary Creature — Human Ninja Rogue",
		OracleID: b30SatoruOracle, Owner: me.ID, Controller: me.ID})
	hand = me.Hand.Size()
	b30Reanimate(t, g, satoru, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a reanimated Satoru draws for itself: %d → %d", hand, me.Hand.Size())
	}
	if spec, _ := Lookup(b30SatoruOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the no-mana-spent half is a declared gap")
	}
}

func TestB30BetorThreeToughnessThresholds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Betor, Kin to All", "Legendary Creature — Spirit Dragon", b30BetorOracle, 5, 7)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// 7 + 2 = 9: nothing.
	hand := me.Hand.Size()
	batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	if len(g.PendingTriggers) != 0 || triggerOrderPrompt(g) != nil {
		t.Fatal("total toughness 9 does not trigger")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("no draw at 9: %d → %d", hand, me.Hand.Size())
	}
	// 7 + 2 + 1 = 10: draw, but no untap.
	b12ToMyNextUpkeep(t, g)
	advanceToMain(t, g)
	b12Creature(g, me.ID, "My Rat", "Creature — Rat", 1, 1)
	b16Tap(g, bear)
	hand = me.Hand.Size()
	batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("total toughness 10 draws: %d → %d", hand, me.Hand.Size())
	}
	if !b12Card(t, g, bear).Tapped {
		t.Error("below 20, no untap")
	}
	// 10 + 10 = 20: draw and untap, no life loss.
	b12ToMyNextUpkeep(t, g)
	advanceToMain(t, g)
	b12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 10)
	b16Tap(g, bear)
	hand = me.Hand.Size()
	before := b30Lives(g)
	batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || b12Card(t, g, bear).Tapped {
		t.Error("at 20: draw and untap each creature you control")
	}
	if g.Seats[1].Life != before[1] {
		t.Error("below 40, no life loss")
	}
	// 20 + 20 = 40: everything.
	b12ToMyNextUpkeep(t, g)
	advanceToMain(t, g)
	b12Creature(g, me.ID, "My Wall 2", "Creature — Wall", 0, 10)
	b12Creature(g, me.ID, "My Wall 3", "Creature — Wall", 0, 10)
	g.Seats[1].Life = 21
	before = b30Lives(g)
	batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Seats[1].Life != 10 {
		t.Errorf("at 40 each opponent loses half their life rounded up: 21 → %d", g.Seats[1].Life)
	}
	if g.Seats[2].Life != before[2]-(before[2]+1)/2 {
		t.Errorf("seat 2: %d → %d", before[2], g.Seats[2].Life)
	}
	if me.Life != before[0] {
		t.Error("the controller loses nothing")
	}
}

func TestB30WakeningSunsAvatarWipesNonDinosaursOnlyWhenCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dino := b12Creature(g, me.ID, "My Dino", "Creature — Dinosaur", 3, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirDino := b12Creature(g, opp.ID, "Their Dino", "Creature — Dinosaur", 3, 3)
	avatar := castCatalogSpell(t, g, "Wakening Sun's Avatar", "Creature — Dinosaur Avatar", b30WakeningSunsAvatarOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(theirs) {
		t.Error("cast from hand: every non-Dinosaur creature is destroyed")
	}
	if !g.Battlefield.Contains(dino) || !g.Battlefield.Contains(theirDino) || !g.Battlefield.Contains(avatar) {
		t.Error("Dinosaurs survive, the Avatar included")
	}
	// Reanimated: no wipe.
	bear2 := b12Creature(g, me.ID, "My Bear 2", "Creature — Bear", 2, 2)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Wakening Sun's Avatar", TypeLine: "Creature — Dinosaur Avatar",
		OracleID: b30WakeningSunsAvatarOracle, ManaCost: "{5}{W}{W}{W}", Owner: me.ID, Controller: me.ID})
	b30Reanimate(t, g, dead, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear2) {
		t.Error("a reanimated Avatar enters quietly")
	}
	// Cast from the command zone: not from hand.
	bear3 := b12Creature(g, me.ID, "My Bear 3", "Creature — Bear", 2, 2)
	active := g.Seats[g.Turn.ActiveSeat]
	cmd := uuid.New()
	active.Command.PushTop(game.Card{InstanceID: cmd, Name: "Wakening Sun's Avatar", TypeLine: "Creature — Dinosaur Avatar",
		OracleID: b30WakeningSunsAvatarOracle, Owner: active.ID, Controller: active.ID, IsCommander: true})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, cmd, game.CastSpellParams{FromZone: "command"}); err != nil {
		t.Fatalf("commander cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear3) {
		t.Error("cast from the command zone is not cast from hand")
	}
}

func TestB30SavvyHunterMakesFoodOnAttackAndOnBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunter := b12Push(g, me.ID, "Savvy Hunter", "Creature — Human Warrior", b30SavvyHunterOracle, 3, 3)
	declareAttack(t, g, opp.ID, hunter)
	passPriorityAroundTable(t, g)
	if got := b30TokensNamed(g, me.ID, "Food"); got != 1 {
		t.Fatalf("attacking makes a Food: %d", got)
	}
	advanceTo(t, g, game.StepCombatDamage)
	aangAdvanceToMain(t, g, 1)
	raider := pushVanillaCreature(g, opp.ID, "Raider", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(hunter, raider); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	// #830: the block declaration announces at its lock-in.
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if got := b30TokensNamed(g, me.ID, "Food"); got != 2 {
		t.Errorf("blocking makes another: %d", got)
	}
	// #747: the two-Food draw ships at its printed count; its engine
	// test is in sacrifice_n_cards_test.go.
	if spec, _ := Lookup(b30SavvyHunterOracle); len(spec.Activated) != 1 || spec.Completeness != CompletenessFull {
		t.Error("the two-Food draw ships whole")
	}
}

func TestB30DocksideChefEatsAnArtifactOrCreatureToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chef := b12Push(g, me.ID, "Dockside Chef", "Enchantment Creature — Human Citizen", b30DocksideChefOracle, 1, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	land := seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, chef, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land is neither an artifact nor a creature")
	}
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, chef, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Fatal("the artifact is sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("draw a card: %d → %d", hand, me.Hand.Size())
	}
	// The Chef can eat itself.
	b16Activate(t, g, me.ID, chef, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{chef}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(chef) || me.Hand.Size() != hand+2 {
		t.Error("sacrificing the Chef to its own ability still draws")
	}
}

func TestB30GrimGuardianDrainsOnItselfAndOtherEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := b30Lives(g)
	castCatalogSpell(t, g, "Grim Guardian", "Enchantment Creature — Zombie", b30GrimGuardianOracle, nil)
	passPriorityAroundTable(t, g)
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != before[i]-1 {
			t.Errorf("its own entry counts: opponent %d %d → %d", i, before[i], g.Seats[i].Life)
		}
	}
	castCatalogSpell(t, g, "Some Aura", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	if g.Seats[1].Life != before[1]-2 {
		t.Errorf("another enchantment you control entering: %d", g.Seats[1].Life)
	}
	castCatalogSpell(t, g, "Some Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if g.Seats[1].Life != before[1]-2 {
		t.Errorf("a creature is not an enchantment: %d", g.Seats[1].Life)
	}
	b13PlayAs(t, g, 1, "Their Enchantment", "Enchantment", "")
	passPriorityAroundTable(t, g)
	if g.Seats[1].Life != before[1]-2 {
		t.Errorf("an opponent's enchantment is not yours: %d", g.Seats[1].Life)
	}
	if me.Life != before[0] {
		t.Error("the controller loses nothing")
	}
}
