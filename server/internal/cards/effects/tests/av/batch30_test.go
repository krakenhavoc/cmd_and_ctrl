package testsav

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b30HandCard seeds a card into a player's hand.
func b30HandCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. River's Rebuke
// was already on main from the S23 boardwipe work and is pinned here
// so the table matches the issue's 32 minus the four declared skips.
func TestBatch30CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B30StonespeakerCrystalOracle: "Stonespeaker Crystal",
		effectstest.B30LunarConvocationOracle:    "Lunar Convocation",
		effectstest.B30TolarianWindsOracle:       "Tolarian Winds",
		effectstest.B30WoodlandChasmOracle:       "Woodland Chasm",
		effectstest.B30RiversRebukeAlreadyOracle: "River's Rebuke",
		effectstest.B30DamnablePactOracle:        "Damnable Pact",
		effectstest.B30HonoredDreyleaderOracle:   "Honored Dreyleader",
		effectstest.B30GoblinTrashmasterOracle:   "Goblin Trashmaster",
		effectstest.B30KhalniGardenOracle:        "Khalni Garden",
		effectstest.B30WindgracesJudgmentOracle:  "Windgrace's Judgment",
		effectstest.B30ThrabenCharmOracle:        "Thraben Charm",
		effectstest.B30SurlyBadgersaurOracle:     "Surly Badgersaur",
		effectstest.B30PromisingVeinOracle:       "Promising Vein",
		effectstest.B30OppressionOracle:          "Oppression",
		effectstest.B30PickYourPoisonOracle:      "Pick Your Poison",
		effectstest.B30KeeperOfFablesOracle:      "Keeper of Fables",
		effectstest.B30IronSpiderOracle:          "Iron Spider, Stark Upgrade",
		effectstest.B30FavorableWindsOracle:      "Favorable Winds",
		effectstest.B30CanopyTacticianOracle:     "Canopy Tactician",
		effectstest.B30VorelOracle:               "Vorel of the Hull Clade",
		effectstest.B30HourOfPromiseOracle:       "Hour of Promise",
		effectstest.B30SatoruOracle:              "Satoru, the Infiltrator",
		effectstest.B30BetorOracle:               "Betor, Kin to All",
		effectstest.B30WakeningSunsAvatarOracle:  "Wakening Sun's Avatar",
		effectstest.B30SavvyHunterOracle:         "Savvy Hunter",
		effectstest.B30ExtractFromDarknessOracle: "Extract from Darkness",
		effectstest.B30DocksideChefOracle:        "Dockside Chef",
		effectstest.B30GrimGuardianOracle:        "Grim Guardian",
		// #920 shipped the resolving-item slot, which is the one
		// thing "a spell copying itself at resolution" was waiting on.
		effectstest.B30ChainOfSmogSkipOracle: "Chain of Smog",
	}
	if len(want) != 29 {
		t.Fatalf("the batch registers 28 cards plus one already on main, the table lists %d", len(want))
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
	// Gemhide Sliver came off the declared-skip list with ADR 0093 and
	// ships in gemhide_sliver.go; Ultima, Origin of Oblivion came off it
	// with ADR 0109 PR 3 and ships in ultima_origin_of_oblivion.go;
	// Zimone, Paradox Sculptor came off it once its beginning-of-combat
	// trigger had an event (zimone_paradox_sculptor.go).
}

// --- lands ---------------------------------------------------------

func TestB30WoodlandChasmEntersTappedAndTapsForEitherColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12PlayFromHand(t, g, "Woodland Chasm", "Snow Land — Swamp Forest", effectstest.B30WoodlandChasmOracle, game.CastSpellParams{})
	if !effectstest.B12Card(t, g, land).Tapped {
		t.Fatal("the snow dual enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	effectstest.B28TapForMana(t, g, me.ID, land, "B")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("tapped for {B}: pool %v", got)
	}
}

func TestB30KhalniGardenEntersTappedWithAPlant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12PlayFromHand(t, g, "Khalni Garden", "Land", effectstest.B30KhalniGardenOracle, game.CastSpellParams{})
	if !effectstest.B12Card(t, g, land).Tapped {
		t.Fatal("the Garden enters tapped")
	}
	if len(g.PendingTriggers) == 0 && effectstest.TriggerOnStack(g, land) == nil {
		t.Fatal("the Plant is an enters trigger, not a silent side effect")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B30TokensNamed(g, me.ID, "Plant"); got != 1 {
		t.Fatalf("one Plant: %d", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Plant" && effects.IsToken(c) {
			if p, tough := effectstest.EffectivePower(t, g, c.InstanceID), effectstest.EffectiveToughness(t, g, c.InstanceID); p != 0 || tough != 1 {
				t.Errorf("a 0/1 Plant: %d/%d", p, tough)
			}
		}
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	effectstest.B28TapForMana(t, g, me.ID, land, "")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("tapped for {G}: pool %v", got)
	}
}

func TestB30PromisingVeinTapsForColorlessAndFetchesABasicTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vein := effectstest.SeedLand(g, me.ID, "Promising Vein", "Land — Cave", effectstest.B30PromisingVeinOracle)
	effectstest.B28TapForMana(t, g, me.ID, vein, "")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("{T}: Add {C}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vein) })
	lib := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Snow-Covered Forest", TypeLine: "Basic Snow Land — Forest"},
		game.Card{Name: "Bayou", TypeLine: "Land — Swamp Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, vein, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(vein) {
		t.Fatal("the Vein is sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("one basic in the library is no choice — it is taken outright")
	}
	if !g.Battlefield.Contains(lib[0]) {
		t.Fatal("a snow-covered basic is a basic land card and is fetched")
	}
	if !effectstest.B12Card(t, g, lib[0]).Tapped {
		t.Error("the fetched basic enters tapped")
	}
	if g.Battlefield.Contains(lib[1]) {
		t.Error("a nonbasic Forest is not a basic land card")
	}
}

// --- spells --------------------------------------------------------

func TestB30TolarianWindsDiscardsTheHandAndDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	lib := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Tolarian Winds", "Instant", effectstest.B30TolarianWindsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	life, hand := opp.Life, opp.Hand.Size()
	effectstest.CastXSpell(t, g, "Damnable Pact", "Sorcery", effectstest.B30DamnablePactOracle, "{X}{B}{B}", 3, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+3 {
		t.Errorf("target player draws X: %d → %d", hand, opp.Hand.Size())
	}
	if opp.Life != life-3 {
		t.Errorf("and loses X: %d → %d", life, opp.Life)
	}
}

func TestB30WindgracesJudgmentDestroysOnePermanentPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	aBear := effectstest.B12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	aRock := effectstest.B12Permanent(g, a.ID, "A's Rock", "Artifact")
	bAura := effectstest.B12Permanent(g, b.ID, "B's Enchantment", "Enchantment")
	bLand := effectstest.SeedLand(g, b.ID, "Forest", "Basic Land — Forest", "")
	legal := effectstest.LegalCards(g, me.ID, effectstest.B30WindgracesJudgmentOracle)
	if legal[mine] || legal[bLand] || !legal[aBear] {
		t.Fatal("only nonland permanents opponents control are legal")
	}
	// One clause per opponent: two of one opponent's permanents are
	// refused at announce (CR 601.2c).
	err := effectstest.CastCatalogSpellErr(t, g, "Windgrace's Judgment", "Instant", effectstest.B30WindgracesJudgmentOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: aBear}, {Kind: game.TargetCard, ID: aRock}, {Kind: game.TargetCard, ID: bAura},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two of one opponent's permanents are refused, got %v", err)
	}
	effectstest.CastCatalogSpell(t, g, "Windgrace's Judgment", "Instant", effectstest.B30WindgracesJudgmentOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: aBear}, {Kind: game.TargetCard, ID: bAura},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aBear) || g.Battlefield.Contains(bAura) {
		t.Error("one permanent of each opponent's is destroyed")
	}
	if !g.Battlefield.Contains(aRock) {
		t.Error("the permanent nobody named stays")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own permanents are untouched")
	}
	if spec, _ := effects.Lookup(effectstest.B30WindgracesJudgmentOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("each pick is bound to its opponent now")
	}
}

func TestB30ThrabenCharmThreeModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushVanillaCreature(g, me.ID, "Soldier A", 1, 1)
	effectstest.PushVanillaCreature(g, me.ID, "Soldier B", 1, 1)
	effectstest.PushVanillaCreature(g, me.ID, "Soldier C", 1, 1)
	golem := effectstest.B12Creature(g, opp.ID, "Their Golem", "Creature — Golem", 6, 6)
	effectstest.CastModal(t, g, "Thraben Charm", "Instant", effectstest.B30ThrabenCharmOracle, []int{0}, effectstest.B16TargetCard(golem))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Error("three creatures is 6 damage — the 6/6 dies")
	}
	aura := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	effectstest.CastModal(t, g, "Thraben Charm", "Instant", effectstest.B30ThrabenCharmOracle, []int{1}, effectstest.B16TargetCard(aura))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("mode two destroys the enchantment")
	}
	dead := effectstest.B17GraveyardCard(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	spare := effectstest.B17GraveyardCard(g.Seats[2], "Spare Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastModal(t, g, "Thraben Charm", "Instant", effectstest.B30ThrabenCharmOracle, []int{2}, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) || !g.Exile.Contains(dead) {
		t.Error("the targeted player's graveyard is exiled")
	}
	if !g.Seats[2].Graveyard.Contains(spare) {
		t.Error("a player not named keeps their graveyard")
	}
}

func TestB30PickYourPoisonEdictsEachOpponentByMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	myFlyer := effectstest.B30Flyer(g, me.ID, "My Bird", "Creature — Bird", 1, 1)
	aFlyer := effectstest.B30Flyer(g, a.ID, "A's Bird", "Creature — Bird", 1, 1)
	aBear := effectstest.B12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	bRock := effectstest.B12Permanent(g, b.ID, "B's Rock", "Artifact")
	effectstest.CastModal(t, g, "Pick Your Poison", "Sorcery", effectstest.B30PickYourPoisonOracle, []int{2}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// A's only flyer is a forced pick, collected unasked (ADR 0175,
	// #3664): only creatures with flying are candidates.
	if effectstest.SacrificeChoiceFor(g, a.ID) != nil || effectstest.SacrificeChoiceFor(g, b.ID) != nil || effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("nobody has a real choice, so nobody is asked")
	}
	if g.Battlefield.Contains(aFlyer) || !g.Battlefield.Contains(aBear) || !g.Battlefield.Contains(myFlyer) {
		t.Error("their flyer is sacrificed; their Bear and your flyer stay")
	}
	effectstest.CastModal(t, g, "Pick Your Poison", "Sorcery", effectstest.B30PickYourPoisonOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bRock) {
		t.Error("mode one sacrifices the artifact")
	}
}

func TestB30HourOfPromiseFetchesTwoLandsTappedAndMakesZombiesWithThreeDeserts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLand(g, me.ID, "Desert of the Glorified", "Land — Desert", "")
	lib := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Desert A", TypeLine: "Land — Desert"},
		game.Card{Name: "Desert B", TypeLine: "Land — Desert"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	effectstest.CastCatalogSpell(t, g, "Hour of Promise", "Sorcery", effectstest.B30HourOfPromiseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !g.Battlefield.Contains(lib[0]) || !g.Battlefield.Contains(lib[1]) {
		t.Fatal("both land cards are fetched")
	}
	if !effectstest.B12Card(t, g, lib[0]).Tapped || !effectstest.B12Card(t, g, lib[1]).Tapped {
		t.Error("they enter tapped")
	}
	if got := effectstest.B30TokensNamed(g, me.ID, "Zombie"); got != 2 {
		t.Errorf("three Deserts after the search is two Zombies: %d", got)
	}
	// Without three Deserts, no Zombies.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.SeedSearchLibrary(me2,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Desert", TypeLine: "Land — Desert"},
	)
	effectstest.CastCatalogSpell(t, g2, "Hour of Promise", "Sorcery", effectstest.B30HourOfPromiseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.B30TokensNamed(g2, me2.ID, "Zombie"); got != 0 {
		t.Errorf("one Desert is no Zombies: %d", got)
	}
}

func TestB30ExtractFromDarknessMillsEveryoneThenReanimatesTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := effectstest.B17GraveyardCard(opp, "Their Dead Golem", "Creature — Golem", "{4}")
	bolt := effectstest.B17GraveyardCard(opp, "Spent Bolt", "Instant", "{R}")
	libs := []int{me.Library.Size(), opp.Library.Size(), g.Seats[2].Library.Size(), g.Seats[3].Library.Size()}
	effectstest.CastCatalogSpell(t, g, "Extract from Darkness", "Sorcery", effectstest.B30ExtractFromDarknessOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, dead) || effectstest.HasID(pick.ChooseCards, bolt) {
		t.Fatal("a creature card in any graveyard is offered after the mill; an instant is not")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, dead)
	for i, p := range g.Seats {
		if p.Library.Size() != libs[i]-2 {
			t.Errorf("seat %d mills two: %d → %d", i, libs[i], p.Library.Size())
		}
	}
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the Golem is on the battlefield")
	}
	if c := effectstest.B12Card(t, g, dead); c.Controller != me.ID {
		t.Error("under the caster's control")
	}
	if spec, _ := effects.Lookup(effectstest.B30ExtractFromDarknessOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("the choice is made after the mill, as printed: nothing is left to declare")
	}
}

// The choice follows the mill, so a creature the spell itself milled
// can be the one put onto the battlefield, and the spell can be cast
// when no graveyard holds a creature yet.
func TestB30ExtractFromDarknessCanTakeACreatureItJustMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	milled := effectstest.B10LibraryTop(opp, "Milled Golem", "Creature — Golem", "{4}", 3, 3)
	effectstest.B10LibraryTop(opp, "Milled Filler", "Instant", "{R}", 0, 0)
	for _, p := range g.Seats {
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() {
				t.Fatal("setup: no graveyard holds a creature before the spell")
			}
		}
	}
	effectstest.CastCatalogSpell(t, g, "Extract from Darkness", "Sorcery", effectstest.B30ExtractFromDarknessOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, milled) {
		t.Fatal("the creature the spell milled is a legal pick")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, milled)
	if c := effectstest.B12Card(t, g, milled); c.Controller != me.ID || !g.Battlefield.Contains(milled) {
		t.Error("the milled creature enters under the caster's control")
	}
}

// --- artifacts and enchantments -------------------------------------

func TestB30StonespeakerCrystalTapsForTwoAndEatsGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	crystal := effectstest.B12Push(g, me.ID, "Stonespeaker Crystal", "Artifact", effectstest.B30StonespeakerCrystalOracle, 0, 0)
	effectstest.B28TapForMana(t, g, me.ID, crystal, "")
	if got := effectstest.PoolColors(me); len(got) != 2 || got[0] != "C" || got[1] != "C" {
		t.Errorf("{T}: Add {C}{C}: pool %v", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(crystal) })
	aDead := effectstest.B17GraveyardCard(a, "A's Dead Bear", "Creature — Bear", "{1}{G}")
	bDead := effectstest.B17GraveyardCard(b, "B's Dead Bear", "Creature — Bear", "{1}{G}")
	mine := effectstest.B17GraveyardCard(me, "My Dead Bear", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, crystal, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: a.ID}, {Kind: game.TargetPlayer, ID: b.ID},
	}})
	if g.Battlefield.Contains(crystal) {
		t.Fatal("sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	crystal2 := effectstest.B12Push(g, me.ID, "Stonespeaker Crystal", "Artifact", effectstest.B30StonespeakerCrystalOracle, 0, 0)
	hand = me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, crystal2, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || !me.Graveyard.Contains(mine) {
		t.Error("with no targets it exiles nothing and still draws")
	}
}

func TestB30LunarConvocationDrainsOnGainAndMakesABatOnGainAndLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	conv := effectstest.B12Push(g, me.ID, "Lunar Convocation", "Enchantment", effectstest.B30LunarConvocationOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B14Gain(t, g, me.ID, 3)
	before := effectstest.B30Lives(g)
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	effectstest.B30SettleOrder(t, g)
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != before[i]-1 {
			t.Errorf("gained life this turn: opponent %d loses 1: %d → %d", i, before[i], g.Seats[i].Life)
		}
	}
	if got := effectstest.B30TokensNamed(g, me.ID, "Bat"); got != 0 {
		t.Errorf("gained but not lost: no Bat: %d", got)
	}
	// Next turn: pay 2 life for a card (a loss) and gain — both fire.
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.AdvanceToMain(t, g)
	hand, life := me.Hand.Size(), me.Life
	effectstest.B16Activate(t, g, me.ID, conv, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-2 {
		t.Fatalf("{1}{B}, Pay 2 life: draw a card: hand %d → %d, life %d → %d", hand, me.Hand.Size(), life, me.Life)
	}
	effectstest.B14Gain(t, g, me.ID, 1)
	before = effectstest.B30Lives(g)
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	// Both triggers fire. Their "if"s read only this turn's life tallies,
	// which nothing can lower, and a drain and a Bat commute, so the
	// batch goes on with no order prompt (#3033).
	if ch := effectstest.TriggerOrderPrompt(g); ch != nil {
		t.Fatalf("a drain and a Bat behind life-tally conditions asked for an order: %+v", ch)
	}
	effectstest.B30SettleOrder(t, g)
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != before[i]-1 {
			t.Errorf("opponent %d loses 1 again: %d → %d", i, before[i], g.Seats[i].Life)
		}
	}
	if got := effectstest.B30TokensNamed(g, me.ID, "Bat"); got != 1 {
		t.Errorf("gained and lost: one Bat: %d", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bat" && effects.IsToken(c) {
			effectstest.AssertKeywords(t, g, c.InstanceID, "flying")
		}
	}
	// A turn with neither: nothing.
	effectstest.B12ToMyNextUpkeep(t, g)
	before = effectstest.B30Lives(g)
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	effectstest.B30SettleOrder(t, g)
	if g.Seats[1].Life != before[1] || effectstest.B30TokensNamed(g, me.ID, "Bat") != 1 {
		t.Error("no life gained this turn: no drain, no Bat")
	}
}

func TestB30OppressionMakesEachCasterDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Oppression", "Enchantment", effectstest.B30OppressionOracle, 0, 0)
	effectstest.CastCatalogSpell(t, g, "My Bear", "Creature — Bear", "", nil)
	if len(g.PendingTriggers) == 0 && effectstest.TriggerOnStack(g, effectstest.FindBattlefieldByName(g, "Oppression")) == nil {
		t.Fatal("casting a spell triggers Oppression")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, me.ID); got != 1 {
		t.Errorf("the caster owes a discard, got %d", got)
	}
	if effectstest.DiscardOwed(g, opp.ID) != 0 {
		t.Error("nobody else does")
	}
	card, _ := me.Hand.Top()
	effectstest.AnswerDiscard(t, g, me.ID, card.InstanceID)
	if !me.Graveyard.Contains(card.InstanceID) {
		t.Error("the chosen card is discarded")
	}
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, opp.ID); got != 1 {
		t.Errorf("an opponent's spell costs the opponent a card, got %d", got)
	}
}

func TestB30FavorableWindsPumpsYourFlyersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Favorable Winds", "Enchantment", effectstest.B30FavorableWindsOracle, 0, 0)
	bird := effectstest.B30Flyer(g, me.ID, "My Bird", "Creature — Bird", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B30Flyer(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1)
	if p, tough := effectstest.EffectivePower(t, g, bird), effectstest.EffectiveToughness(t, g, bird); p != 2 || tough != 2 {
		t.Errorf("a creature you control with flying is +1/+1: %d/%d", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("one without flying is unchanged: %d", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's flyer is unchanged: %d", p)
	}
}

// --- creatures -----------------------------------------------------

func TestB30HonoredDreyleaderCountsSquirrelsAndFoodAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "My Squirrel", "Creature — Squirrel", 1, 1)
	effectstest.B12Permanent(g, me.ID, "Food", "Token Artifact — Food")
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Squirrel", "Creature — Squirrel", 1, 1)
	leader := effectstest.CastCatalogSpell(t, g, "Honored Dreyleader", "Creature — Squirrel Warrior", effectstest.B30HonoredDreyleaderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, leader, "+1/+1"); got != 2 {
		t.Fatalf("one Squirrel and one Food you control: %d counters", got)
	}
	effectstest.AssertKeywords(t, g, leader, "trample")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.FoodToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, leader, "+1/+1"); got != 3 {
		t.Errorf("a Food entering grows it: %d", got)
	}
	effectstest.CastCatalogSpell(t, g, "Another Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, leader, "+1/+1"); got != 3 {
		t.Errorf("a Bear does not: %d", got)
	}
	effectstest.B13PlayAs(t, g, 1, "Their Other Squirrel", "Creature — Squirrel", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, leader, "+1/+1"); got != 3 {
		t.Errorf("an opponent's Squirrel does not: %d", got)
	}
}

func TestB30GoblinTrashmasterLordsGoblinsAndEatsOneToKillAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := effectstest.B12Push(g, me.ID, "Goblin Trashmaster", "Creature — Goblin Warrior", effectstest.B30GoblinTrashmasterOracle, 3, 3)
	goblin := effectstest.B12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if p := effectstest.EffectivePower(t, g, goblin); p != 2 {
		t.Errorf("another Goblin you control is +1/+1: %d", p)
	}
	if p := effectstest.EffectivePower(t, g, master); p != 3 {
		t.Errorf("the Trashmaster is 'other': %d", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Goblin is unchanged: %d", p)
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear}, Targets: effectstest.B16TargetCard(rock),
	}); err == nil {
		t.Fatal("a Bear is not a Goblin")
	}
	effectstest.B16Activate(t, g, me.ID, master, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{goblin}, Targets: effectstest.B16TargetCard(rock)})
	if g.Battlefield.Contains(goblin) {
		t.Fatal("the Goblin is sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
}

func TestB30SurlyBadgersaurThreeDiscardPayoffs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	badger := effectstest.B12Push(g, me.ID, "Surly Badgersaur", "Creature — Badger Dinosaur", effectstest.B30SurlyBadgersaurOracle, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.EmptyHandToLibrary(g, me)
	effectstest.AdvanceToMain(t, g)
	b30HandCard(me, "Some Bear", "Creature — Bear")
	effectstest.B30DiscardOne(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, badger, "+1/+1"); got != 1 {
		t.Errorf("discarding a creature card grows it: %d", got)
	}
	b30HandCard(me, "Forest", "Basic Land — Forest")
	effectstest.B30DiscardOne(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B30TokensNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("discarding a land card makes a Treasure: %d", got)
	}
	b30HandCard(me, "Some Bolt", "Instant")
	effectstest.B30DiscardOne(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("discarding a noncreature, nonland card asks which creature to fight")
	}
	if effectstest.HasID(p.PickTargetCards, badger) || !effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("only creatures you don't control are offered")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 4/4 Badgersaur kills the 2/2")
	}
	if c := effectstest.B12Card(t, g, badger); c.DamageMarked != 2 {
		t.Errorf("and takes 2 back: %d marked", c.DamageMarked)
	}
	if got := effectstest.CounterCount(g, badger, "+1/+1"); got != 1 || effectstest.B30TokensNamed(g, me.ID, "Treasure") != 1 {
		t.Error("the other two triggers do not fire on an instant")
	}
	// Nothing to fight: the trigger is removed without a prompt.
	b30HandCard(me, "Another Bolt", "Instant")
	effectstest.B30DiscardOne(t, g, me.ID)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("with no creature you don't control there is nothing to ask")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// An opponent's discard is not yours.
	b30HandCard(opp, "Their Land", "Basic Land — Swamp")
	effectstest.B30DiscardOne(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B30TokensNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("an opponent's discard makes nothing: %d", got)
	}
}

func TestB30KeeperOfFablesDrawsOncePerNonHumanCombatDamageBatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Keeper of Fables", "Creature — Cat", effectstest.B30KeeperOfFablesOracle, 4, 5)
	cat := effectstest.B12Creature(g, me.ID, "My Cat", "Creature — Cat", 2, 2)
	dog := effectstest.B12Creature(g, me.ID, "My Dog", "Creature — Dog", 2, 2)
	human := effectstest.B12Creature(g, me.ID, "My Human", "Creature — Human Soldier", 2, 2)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, cat, dog, human)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("two non-Humans connecting is one draw: %d → %d", hand, me.Hand.Size())
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.AdvanceToMain(t, g)
	hand = me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, human)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a Human alone draws nothing: %d → %d", hand, me.Hand.Size())
	}
}

func TestB30IronSpiderPumpsArtifactCreaturesAndVehicles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := effectstest.B12Push(g, me.ID, "Iron Spider, Stark Upgrade", "Legendary Artifact Creature — Spider Hero", effectstest.B30IronSpiderOracle, 2, 3)
	construct := effectstest.B12Creature(g, me.ID, "My Construct", "Artifact Creature — Construct", 1, 1)
	vehicle := effectstest.B12Permanent(g, me.ID, "My Vehicle", "Artifact — Vehicle")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Construct", "Artifact Creature — Construct", 1, 1)
	effectstest.AssertKeywords(t, g, spider, "vigilance")
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, spider, 0, game.ActivateAbilityParams{})
	if !effectstest.B12Card(t, g, spider).Tapped {
		t.Error("tapped to activate")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, want := range []struct {
		id   uuid.UUID
		n    int
		what string
	}{
		{spider, 1, "the Spider itself"}, {construct, 1, "an artifact creature"}, {vehicle, 1, "a Vehicle"},
		{bear, 0, "a non-artifact creature"}, {rock, 0, "a non-creature, non-Vehicle artifact"}, {theirs, 0, "an opponent's artifact creature"},
	} {
		if got := effectstest.CounterCount(g, want.id, "+1/+1"); got != want.n {
			t.Errorf("%s: %d counters, want %d", want.what, got, want.n)
		}
	}
	// #789 made the draw ability real: "from among artifacts you
	// control" is a counter removal split across permanents, and the
	// card no longer ships with a gap.
	if spec, _ := effects.Lookup(effectstest.B30IronSpiderOracle); len(spec.Activated) != 2 || spec.Completeness != effects.CompletenessFull {
		t.Error("the counter-removal draw should be live and the card complete (#789)")
	}
}

func TestB30CanopyTacticianLordsElvesAndTapsForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tactician := effectstest.B12Push(g, me.ID, "Canopy Tactician", "Creature — Elf Warrior", effectstest.B30CanopyTacticianOracle, 3, 3)
	elf := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf Druid", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	if p := effectstest.EffectivePower(t, g, elf); p != 2 {
		t.Errorf("another Elf you control is +1/+1: %d", p)
	}
	if p := effectstest.EffectivePower(t, g, tactician); p != 3 {
		t.Errorf("the Tactician is 'other': %d", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 || effectstest.EffectivePower(t, g, theirs) != 1 {
		t.Error("a Bear and an opponent's Elf are unchanged")
	}
	effectstest.B28TapForMana(t, g, me.ID, tactician, "")
	if got := effectstest.PoolColors(me); len(got) != 3 || got[0] != "G" || got[1] != "G" || got[2] != "G" {
		t.Errorf("{T}: Add {G}{G}{G}: pool %v", got)
	}
}

func TestB30VorelDoublesEveryKindOfCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vorel := effectstest.B12Push(g, me.ID, "Vorel of the Hull Clade", "Legendary Creature — Human Merfolk", effectstest.B30VorelOracle, 1, 4)
	hydra := effectstest.B12Creature(g, opp.ID, "Their Hydra", "Creature — Hydra", 0, 0)
	effectstest.B28AddCounter(t, g, hydra, 3)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hydra, "charge", 2) })
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, vorel, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(hydra)})
	if !effectstest.B12Card(t, g, vorel).Tapped {
		t.Error("tapped to activate")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, hydra, "+1/+1"); got != 6 {
		t.Errorf("three +1/+1 counters become six: %d", got)
	}
	if got := effectstest.CounterCount(g, hydra, "charge"); got != 4 {
		t.Errorf("two charge counters become four: %d", got)
	}
	// No counters: nothing to double, and the ability still resolves.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vorel) })
	effectstest.B16Activate(t, g, me.ID, vorel, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 0 {
		t.Errorf("zero doubled is zero: %d", got)
	}
}

func TestB30SatoruDrawsWhenCreaturesArriveUncast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	satoru := effectstest.CastFromHandForTest(t, g, me, "Satoru, the Infiltrator", "Legendary Creature — Human Ninja Rogue", "{U}{B}", effectstest.B30SatoruOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("a cast Satoru draws nothing: %d → %d", hand, me.Hand.Size())
	}
	effectstest.AssertKeywords(t, g, satoru, "menace")
	// ADR 0109 §11: Satoru reads the whole event batch (CR 603.2c), and a
	// direct engine call outside a resolution shares the batch of the
	// last one — so each arrival below happens in a step of its own, as
	// it would in play, rather than alongside Satoru's own cast.
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepBeginCombat)
	dead := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	effectstest.B30Reanimate(t, g, dead, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a reanimated creature was not cast: draw: %d → %d", hand, me.Hand.Size())
	}
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEndCombat)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("tokens draw nothing: %d", me.Hand.Size())
	}
	theirs := effectstest.B17GraveyardCard(opp, "Their Dead Bear", "Creature — Bear", "{1}{G}")
	effectstest.B30Reanimate(t, g, theirs, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("an opponent's reanimation draws nothing: %d", me.Hand.Size())
	}
	// #3855: a creature cast with an insufficient pool under Satoru asks
	// what was paid outside the pool. Declared paid, it draws nothing.
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepPostcombatMain)
	asking := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: asking, Name: "Asking Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})
	})
	var ask *game.PaperPaymentRequiredError
	if err := g.CastSpell(me.ID, asking, game.CastSpellParams{}); !errors.As(err, &ask) {
		t.Fatalf("an unfunded cast under Satoru = %v, want a paper payment prompt", err)
	}
	g.WithWriteLock(func() { me.Hand.Remove(asking) })
	effectstest.CastFromHandForTest(t, g, me, "Cast Bear", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{
		PaperPayment: &game.PaperPayment{Mana: []game.PaperMana{{Color: "G"}, {Color: "G"}}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a creature declared paid draws nothing: %d", me.Hand.Size())
	}
	// Declared cast without paying its mana cost, it draws.
	effectstest.CastFromHandForTest(t, g, me, "Free Bear", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{
		PaperPayment: &game.PaperPayment{Free: true},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("a creature declared free draws: %d → %d", hand+1, me.Hand.Size())
	}
	// Satoru itself arriving uncast draws.
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(satoru) })
	me.Hand.Remove(satoru)
	me.Graveyard.PushTop(game.Card{InstanceID: satoru, Name: "Satoru, the Infiltrator", TypeLine: "Legendary Creature — Human Ninja Rogue",
		OracleID: effectstest.B30SatoruOracle, Owner: me.ID, Controller: me.ID})
	hand = me.Hand.Size()
	effectstest.B30Reanimate(t, g, satoru, me.ID)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a reanimated Satoru draws for itself: %d → %d", hand, me.Hand.Size())
	}
	if spec, _ := effects.Lookup(effectstest.B30SatoruOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("Satoru is complete since a waived payment asks (#3855)")
	}
}

func TestB30BetorThreeToughnessThresholds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Betor, Kin to All", "Legendary Creature — Spirit Dragon", effectstest.B30BetorOracle, 5, 7)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// 7 + 2 = 9: nothing.
	hand := me.Hand.Size()
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	if len(g.PendingTriggers) != 0 || effectstest.TriggerOrderPrompt(g) != nil {
		t.Fatal("total toughness 9 does not trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("no draw at 9: %d → %d", hand, me.Hand.Size())
	}
	// 7 + 2 + 1 = 10: draw, but no untap.
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "My Rat", "Creature — Rat", 1, 1)
	effectstest.B16Tap(g, bear)
	hand = me.Hand.Size()
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("total toughness 10 draws: %d → %d", hand, me.Hand.Size())
	}
	if !effectstest.B12Card(t, g, bear).Tapped {
		t.Error("below 20, no untap")
	}
	// 10 + 10 = 20: draw and untap, no life loss.
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 10)
	effectstest.B16Tap(g, bear)
	hand = me.Hand.Size()
	before := effectstest.B30Lives(g)
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || effectstest.B12Card(t, g, bear).Tapped {
		t.Error("at 20: draw and untap each creature you control")
	}
	if g.Seats[1].Life != before[1] {
		t.Error("below 40, no life loss")
	}
	// 20 + 20 = 40: everything.
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "My Wall 2", "Creature — Wall", 0, 10)
	effectstest.B12Creature(g, me.ID, "My Wall 3", "Creature — Wall", 0, 10)
	g.Seats[1].Life = 21
	before = effectstest.B30Lives(g)
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dino := effectstest.B12Creature(g, me.ID, "My Dino", "Creature — Dinosaur", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirDino := effectstest.B12Creature(g, opp.ID, "Their Dino", "Creature — Dinosaur", 3, 3)
	avatar := effectstest.CastCatalogSpell(t, g, "Wakening Sun's Avatar", "Creature — Dinosaur Avatar", effectstest.B30WakeningSunsAvatarOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(theirs) {
		t.Error("cast from hand: every non-Dinosaur creature is destroyed")
	}
	if !g.Battlefield.Contains(dino) || !g.Battlefield.Contains(theirDino) || !g.Battlefield.Contains(avatar) {
		t.Error("Dinosaurs survive, the Avatar included")
	}
	// Reanimated: no wipe.
	bear2 := effectstest.B12Creature(g, me.ID, "My Bear 2", "Creature — Bear", 2, 2)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Wakening Sun's Avatar", TypeLine: "Creature — Dinosaur Avatar",
		OracleID: effectstest.B30WakeningSunsAvatarOracle, ManaCost: "{5}{W}{W}{W}", Owner: me.ID, Controller: me.ID})
	effectstest.B30Reanimate(t, g, dead, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear2) {
		t.Error("a reanimated Avatar enters quietly")
	}
	// Cast from the command zone: not from hand.
	bear3 := effectstest.B12Creature(g, me.ID, "My Bear 3", "Creature — Bear", 2, 2)
	active := g.Seats[g.Turn.ActiveSeat]
	cmd := uuid.New()
	active.Command.PushTop(game.Card{InstanceID: cmd, Name: "Wakening Sun's Avatar", TypeLine: "Creature — Dinosaur Avatar",
		OracleID: effectstest.B30WakeningSunsAvatarOracle, Owner: active.ID, Controller: active.ID, IsCommander: true})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, cmd, game.CastSpellParams{FromZone: "command"}); err != nil {
		t.Fatalf("commander cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear3) {
		t.Error("cast from the command zone is not cast from hand")
	}
}

func TestB30SavvyHunterMakesFoodOnAttackAndOnBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunter := effectstest.B12Push(g, me.ID, "Savvy Hunter", "Creature — Human Warrior", effectstest.B30SavvyHunterOracle, 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, hunter)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B30TokensNamed(g, me.ID, "Food"); got != 1 {
		t.Fatalf("attacking makes a Food: %d", got)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.AangAdvanceToMain(t, g, 1)
	raider := effectstest.PushVanillaCreature(g, opp.ID, "Raider", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// The Hunter is still tapped from its own attack, and a tapped
	// creature can't block (CR 509.1a). #1501: with no legal block the
	// defender's declaration would complete as the step began and the
	// block below be refused as late, so it untaps first.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(hunter) })
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(hunter, raider); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	// #830: the block declaration announces at its lock-in.
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B30TokensNamed(g, me.ID, "Food"); got != 2 {
		t.Errorf("blocking makes another: %d", got)
	}
	// #747: the two-Food draw ships at its printed count; its engine
	// test is in sacrifice_n_cards_test.go.
	if spec, _ := effects.Lookup(effectstest.B30SavvyHunterOracle); len(spec.Activated) != 1 || spec.Completeness != effects.CompletenessFull {
		t.Error("the two-Food draw ships whole")
	}
}

func TestB30DocksideChefEatsAnArtifactOrCreatureToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chef := effectstest.B12Push(g, me.ID, "Dockside Chef", "Enchantment Creature — Human Citizen", effectstest.B30DocksideChefOracle, 1, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	land := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, chef, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land is neither an artifact nor a creature")
	}
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, chef, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Fatal("the artifact is sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("draw a card: %d → %d", hand, me.Hand.Size())
	}
	// The Chef can eat itself.
	effectstest.B16Activate(t, g, me.ID, chef, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{chef}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(chef) || me.Hand.Size() != hand+2 {
		t.Error("sacrificing the Chef to its own ability still draws")
	}
}

func TestB30GrimGuardianDrainsOnItselfAndOtherEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := effectstest.B30Lives(g)
	effectstest.CastCatalogSpell(t, g, "Grim Guardian", "Enchantment Creature — Zombie", effectstest.B30GrimGuardianOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != before[i]-1 {
			t.Errorf("its own entry counts: opponent %d %d → %d", i, before[i], g.Seats[i].Life)
		}
	}
	effectstest.CastCatalogSpell(t, g, "Some Aura", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[1].Life != before[1]-2 {
		t.Errorf("another enchantment you control entering: %d", g.Seats[1].Life)
	}
	effectstest.CastCatalogSpell(t, g, "Some Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[1].Life != before[1]-2 {
		t.Errorf("a creature is not an enchantment: %d", g.Seats[1].Life)
	}
	effectstest.B13PlayAs(t, g, 1, "Their Enchantment", "Enchantment", "")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[1].Life != before[1]-2 {
		t.Errorf("an opponent's enchantment is not yours: %d", g.Seats[1].Life)
	}
	if me.Life != before[0] {
		t.Error("the controller loses nothing")
	}
}
