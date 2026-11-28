package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b827_cards_test.go — Homelands: instants and sorceries, life gain and
// loss, removal, graveyard recursion, artifacts, enters and dies triggers,
// batch 827 (S82 older sets, #5635). Every test plays the real catalog entry
// through the engine.

const (
	oldB827Ambush     = "05828c1e-9185-4b70-94fe-753eacdf45a1"
	oldB827ChainStas  = "1ec365f8-bfb8-416f-b4db-868a8f5d042a"
	oldB827AnHavvaInn = "3cd72848-7b99-4d46-96f5-53f297c9d2a0"
	oldB827Albatross  = "8b59cacb-46ff-4400-a7f6-b258c81a9b14"
	oldB827Visage     = "8e1d750e-44a7-47b1-8e0a-5f8b15d48b41"
	oldB827Chime      = "995e2412-3f69-4aaf-a5d6-ae0ecbb95061"
	oldB827Headstone  = "cbf767d4-9be4-417a-bbf7-0c72ede727b4"
	oldB827Drudge     = "d63f4eeb-3fa5-4e65-8414-e3fd3c03c128"
	oldB827Ban        = "445dd350-5165-4516-8669-9233f5558605"
	oldB827Ruins      = "f805b100-49b8-4d23-a768-e8de8bb0daa3"
)

func TestOldB827EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB827Ambush: "Ambush", oldB827ChainStas: "Chain Stasis", oldB827AnHavvaInn: "An-Havva Inn",
		oldB827Albatross: "Giant Albatross", oldB827Visage: "Broken Visage", oldB827Chime: "Apocalypse Chime",
		oldB827Headstone: "Headstone", oldB827Drudge: "Drudge Spell", oldB827Ban: "Feroz's Ban",
		oldB827Ruins: "An-Zerrin Ruins",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB827Cast puts a card in the active seat's hand with a pool and casts it
// strictly where the game stands.
func oldB827Cast(g *game.Game, c game.Card, pool string, p game.CastSpellParams) (uuid.UUID, error) {
	me := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	me.Hand.PushTop(c)
	effectstest.FloatForTest(g, me, pool)
	p.Strict = true
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, p)
}

func oldB827Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB827DrawsAtTheNextUpkeep(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	before := p.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p.Hand.Size() != before {
		t.Fatalf("drew before the upkeep trigger resolved: hand %d -> %d", before, p.Hand.Size())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d at the next upkeep, want +1", before, p.Hand.Size())
	}
}

// --- Ambush ------------------------------------------------------------------

func oldB827AmbushCard() game.Card {
	return game.Card{Name: "Ambush", TypeLine: "Instant", OracleID: oldB827Ambush, ManaCost: "{3}{R}", Colors: []string{"R"}}
}

func TestOldB827AmbushGivesBlockingCreaturesFirstStrikeAndNobodyElse(t *testing.T) {
	g, me, opp := oldB827Main(t)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: attacker})
	if effectstest.P1g2HasKeyword(g, blocker, "first strike") {
		t.Fatal("the blocker had first strike before Ambush")
	}
	if _, err := oldB827Cast(g, oldB827AmbushCard(), "CCCR", game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, blocker, "first strike") {
		t.Error("the blocking creature did not gain first strike")
	}
	if effectstest.P1g2HasKeyword(g, attacker, "first strike") {
		t.Error("the attacking creature gained first strike")
	}
	if effectstest.P1g2HasKeyword(g, bystander, "first strike") {
		t.Error("a creature that did not block gained first strike")
	}
}

func TestOldB827AmbushWithNoBlockerDoesNothingAndLaterBlockersMissOut(t *testing.T) {
	g, me, opp := oldB827Main(t)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	late := effectstest.B12Creature(g, opp.ID, "Late Blocker", "Creature — Bear", 2, 2)
	if _, err := oldB827Cast(g, oldB827AmbushCard(), "CCCR", game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{late: attacker})
	if effectstest.P1g2HasKeyword(g, late, "first strike") {
		t.Error("a creature that blocked after Ambush resolved gained first strike")
	}
}

func TestOldB827AmbushIsRefusedWithoutItsMana(t *testing.T) {
	g, _, _ := oldB827Main(t)
	if _, err := oldB827Cast(g, oldB827AmbushCard(), "CCR", game.CastSpellParams{}); err == nil {
		t.Error("Ambush was cast with three mana")
	}
}

// --- An-Havva Inn ------------------------------------------------------------

func oldB827InnCard() game.Card {
	return game.Card{Name: "An-Havva Inn", TypeLine: "Sorcery", OracleID: oldB827AnHavvaInn, ManaCost: "{1}{G}{G}", Colors: []string{"G"}}
}

func TestOldB827AnHavvaInnGainsOnePlusEveryGreenCreatureOnTheBattlefield(t *testing.T) {
	g, me, opp := oldB827Main(t)
	effectstest.SeedColoredCreature(g, me.ID, "My Elf", []string{"G"}, 0)
	effectstest.SeedColoredCreature(g, opp.ID, "Their Elf", []string{"G"}, 0)
	effectstest.SeedColoredCreature(g, opp.ID, "Their Goblin", []string{"R"}, 0)
	effectstest.B12Permanent(g, opp.ID, "Green Rock", "Artifact")
	life := me.Life
	if _, err := oldB827Cast(g, oldB827InnCard(), "CGG", game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3 (two green creatures plus one)", life, me.Life)
	}
}

func TestOldB827AnHavvaInnGainsOneWithNoGreenCreature(t *testing.T) {
	g, me, _ := oldB827Main(t)
	life := me.Life
	if _, err := oldB827Cast(g, oldB827InnCard(), "CGG", game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
}

func TestOldB827AnHavvaInnIsRefusedWithoutGreenMana(t *testing.T) {
	g, _, _ := oldB827Main(t)
	if _, err := oldB827Cast(g, oldB827InnCard(), "CCC", game.CastSpellParams{}); err == nil {
		t.Error("An-Havva Inn was cast without green mana")
	}
}

// --- Feroz's Ban -------------------------------------------------------------

func oldB827Price(t *testing.T, g *game.Game, seat *game.Player, typeLine, cost string) int {
	t.Helper()
	base, err := game.ParseCost(cost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", cost, err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(), Name: "Spell", TypeLine: typeLine, ManaCost: cost,
			Owner: seat.ID, Controller: seat.ID,
		},
		Controller: seat.ID,
		FromZone:   game.ZoneHand,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers: %v", err)
	}
	return out.ManaValue()
}

func TestOldB827FerozsBanTaxesEveryPlayersCreatureSpellsByTwo(t *testing.T) {
	g, me, opp := oldB827Main(t)
	if got := oldB827Price(t, g, me, "Creature — Bear", "{1}{G}"); got != 2 {
		t.Fatalf("a creature spell costs %d before the Ban, want 2", got)
	}
	effectstest.B12Push(g, me.ID, "Feroz's Ban", "Artifact", oldB827Ban, 0, 0)
	for _, tc := range []struct {
		name     string
		who      *game.Player
		typeLine string
		cost     string
		want     int
	}{
		{"own creature", me, "Creature — Bear", "{1}{G}", 4},
		{"opponent's creature", opp, "Creature — Bear", "{1}{G}", 4},
		{"artifact creature", opp, "Artifact Creature — Golem", "{3}", 5},
		{"sorcery", me, "Sorcery", "{1}{G}", 2},
		{"artifact", me, "Artifact", "{2}", 2},
	} {
		if got := oldB827Price(t, g, tc.who, tc.typeLine, tc.cost); got != tc.want {
			t.Errorf("%s: costs %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestOldB827FerozsBanRefusesACreatureSpellWithoutTheExtraTwo(t *testing.T) {
	g, me, _ := oldB827Main(t)
	effectstest.B12Push(g, me.ID, "Feroz's Ban", "Artifact", oldB827Ban, 0, 0)
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2}
	if _, err := oldB827Cast(g, bear, "CG", game.CastSpellParams{}); err == nil {
		t.Error("a creature spell was cast for its printed cost under Feroz's Ban")
	}
	if _, err := oldB827Cast(g, bear, "CCCG", game.CastSpellParams{}); err != nil {
		t.Errorf("a creature spell was refused with the extra {2} paid: %v", err)
	}
}

// --- An-Zerrin Ruins ---------------------------------------------------------

func TestOldB827AnZerrinRuinsKeepsTheChosenTypeTappedForEveryController(t *testing.T) {
	g, me, opp := oldB827Main(t)
	if _, err := oldB827Cast(g, game.Card{Name: "An-Zerrin Ruins", TypeLine: "Enchantment", OracleID: oldB827Ruins, ManaCost: "{2}{R}{R}", Colors: []string{"R"}}, "CCRR", game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Bear")
	effectstest.PassPriorityAroundTable(t, g)
	mineBear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	mineElf := effectstest.B12Creature(g, me.ID, "My Elf", "Creature — Elf", 2, 2)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirElf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{mineBear, mineElf, theirBear, theirElf} {
			effectstest.FindBattlefieldCardForTest(g, id).Tapped = true
		}
	})
	effectstest.B12ToMyNextUpkeep(t, g)
	if !effectstest.B16Tapped(t, g, mineBear) {
		t.Error("my Bear untapped in my untap step")
	}
	if effectstest.B16Tapped(t, g, mineElf) {
		t.Error("my Elf stayed tapped though it is not the chosen type")
	}
	// The opponent's untap step came and went while the turns went round.
	if !effectstest.B16Tapped(t, g, theirBear) {
		t.Error("their Bear untapped in their untap step")
	}
	if effectstest.B16Tapped(t, g, theirElf) {
		t.Error("their Elf stayed tapped though it is not the chosen type")
	}
}

func TestOldB827AnZerrinRuinsIsRefusedWithoutItsMana(t *testing.T) {
	g, _, _ := oldB827Main(t)
	if _, err := oldB827Cast(g, game.Card{Name: "An-Zerrin Ruins", TypeLine: "Enchantment", OracleID: oldB827Ruins, ManaCost: "{2}{R}{R}", Colors: []string{"R"}}, "CCCR", game.CastSpellParams{}); err == nil {
		t.Error("An-Zerrin Ruins was cast with one red mana")
	}
}

// --- Headstone ---------------------------------------------------------------

func oldB827HeadstoneCard() game.Card {
	return game.Card{Name: "Headstone", TypeLine: "Instant", OracleID: oldB827Headstone, ManaCost: "{1}{B}", Colors: []string{"B"}}
}

func TestOldB827HeadstoneExilesAnyGraveyardCardAndDrawsAtTheNextUpkeep(t *testing.T) {
	for _, theirs := range []bool{true, false} {
		g, me, opp := oldB827Main(t)
		owner := me
		if theirs {
			owner = opp
		}
		victim := effectstest.PushGraveyardCard(g, owner, "Doomed Card")
		if _, err := oldB827Cast(g, oldB827HeadstoneCard(), "CB", game.CastSpellParams{Targets: effectstest.CardRefs(victim)}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if z := effectstest.ZoneOf(g, victim); z != game.ZoneExile {
			t.Errorf("target of %s graveyard is in %q, want exile", map[bool]string{true: "their", false: "my"}[theirs], z)
		}
		oldB827DrawsAtTheNextUpkeep(t, g, me)
	}
}

func TestOldB827HeadstoneRefusesAPermanentAndNeedsATarget(t *testing.T) {
	g, me, _ := oldB827Main(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if _, err := oldB827Cast(g, oldB827HeadstoneCard(), "CB", game.CastSpellParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Error("Headstone targeted a creature on the battlefield")
	}
	if _, err := oldB827Cast(g, oldB827HeadstoneCard(), "CB", game.CastSpellParams{}); err == nil {
		t.Error("Headstone was cast without a target")
	}
}

func TestOldB827HeadstoneDrawsNothingWhenItsTargetLeftTheGraveyard(t *testing.T) {
	g, me, _ := oldB827Main(t)
	victim := effectstest.PushGraveyardCard(g, me, "Fleeting Card")
	if _, err := oldB827Cast(g, oldB827HeadstoneCard(), "CB", game.CastSpellParams{Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { _, _ = me.Graveyard.Remove(victim) })
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: a spell with no legal target does nothing, the draw included", hand, me.Hand.Size())
	}
}

// --- Apocalypse Chime --------------------------------------------------------

func TestOldB827ApocalypseChimeDestroysNontokenHomelandsPermanentsAndIgnoresRegeneration(t *testing.T) {
	g, me, opp := oldB827Main(t)
	chime := effectstest.PushCatalogPermanent(g, me.ID, "Apocalypse Chime", "Artifact", oldB827Chime, false)
	matron := effectstest.B12Creature(g, me.ID, "Abbey Matron", "Creature — Human Cleric", 1, 1)
	baron := effectstest.B12Creature(g, opp.ID, "Baron Sengir", "Legendary Creature — Vampire", 5, 5)
	shielded := effectstest.B12Creature(g, opp.ID, "Sengir Bats", "Creature — Bat", 2, 2)
	effectstest.P1Regenerate(t, g, shielded)
	highway := effectstest.B12Permanent(g, opp.ID, "Aysen Highway", "Land")
	apostrophe := effectstest.B12Creature(g, me.ID, "Ihsan’s Shade", "Creature — Shade", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	token := effectstest.B12Creature(g, opp.ID, "Joven", "Token Creature — Human Rogue", 1, 1)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, chime, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(chime) {
		t.Error("the Chime was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{
		"Abbey Matron": matron, "Baron Sengir": baron, "Sengir Bats (regeneration shield)": shielded,
		"Aysen Highway": highway, "Ihsan’s Shade": apostrophe,
	} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived the Chime", name)
		}
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("Grizzly Bears (not a Homelands name) was destroyed")
	}
	if !g.Battlefield.Contains(token) {
		t.Error("a token named like a Homelands card was destroyed")
	}
}

func TestOldB827ApocalypseChimeNeedsItsManaAndATappedChimeCannotActivate(t *testing.T) {
	g, me, _ := oldB827Main(t)
	chime := effectstest.PushCatalogPermanent(g, me.ID, "Apocalypse Chime", "Artifact", oldB827Chime, false)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, chime, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the Chime was activated with one mana")
	}
	effectstest.FloatForTest(g, me, "C")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, chime).Tapped = true })
	if err := g.ActivateCatalogAbility(me.ID, chime, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a tapped Chime was activated")
	}
}

// --- Drudge Spell ------------------------------------------------------------

func oldB827Skeletons(g *game.Game) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Skeleton" && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}

func TestOldB827DrudgeSpellExilesTwoCreatureCardsForARegeneratingSkeleton(t *testing.T) {
	g, me, _ := oldB827Main(t)
	drudge := effectstest.PushCatalogPermanent(g, me.ID, "Drudge Spell", "Enchantment", oldB827Drudge, false)
	a := effectstest.PushGraveyardCard(g, me, "Dead Bear")
	b := effectstest.PushGraveyardCard(g, me, "Dead Elf")
	for _, id := range []uuid.UUID{a, b} {
		g.WithWriteLock(func() {
			c, _ := me.Graveyard.Remove(id)
			c.TypeLine = "Creature — Test"
			me.Graveyard.PushTop(c)
		})
	}
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, drudge, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if z := effectstest.ZoneOf(g, id); z != game.ZoneExile {
			t.Errorf("a creature card paid for the ability is in %q, want exile", z)
		}
	}
	sk := oldB827Skeletons(g)
	if len(sk) != 1 {
		t.Fatalf("%d Skeleton tokens, want 1", len(sk))
	}
	if sk[0].Power != 1 || sk[0].Toughness != 1 || sk[0].Controller != me.ID {
		t.Errorf("Skeleton is %d/%d under %s, want a 1/1 under me", sk[0].Power, sk[0].Toughness, sk[0].Controller)
	}
	if row := effectstest.ActivatedRowOf(t, g, me.ID, sk[0].InstanceID, 0); row.Label == "" {
		t.Error("the Skeleton has no regenerate ability")
	}
}

func TestOldB827DrudgeSpellIsRefusedWithoutTwoCreatureCardsOrBlackMana(t *testing.T) {
	g, me, _ := oldB827Main(t)
	drudge := effectstest.PushCatalogPermanent(g, me.ID, "Drudge Spell", "Enchantment", oldB827Drudge, false)
	creature := effectstest.PushGraveyardCard(g, me, "Dead Bear")
	land := effectstest.PushGraveyardCard(g, me, "Dead Land")
	g.WithWriteLock(func() {
		c, _ := me.Graveyard.Remove(creature)
		c.TypeLine = "Creature — Bear"
		me.Graveyard.PushTop(c)
		c, _ = me.Graveyard.Remove(land)
		c.TypeLine = "Land"
		me.Graveyard.PushTop(c)
	})
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, drudge, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{creature, land}}); err == nil {
		t.Error("a land card was exiled to pay for a creature-card cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, drudge, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{creature}}); err == nil {
		t.Error("one creature card paid for a two-card cost")
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	second := effectstest.PushGraveyardCard(g, me, "Dead Elf")
	g.WithWriteLock(func() {
		c, _ := me.Graveyard.Remove(second)
		c.TypeLine = "Creature — Elf"
		me.Graveyard.PushTop(c)
	})
	if err := g.ActivateCatalogAbility(me.ID, drudge, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{creature, second}}); err == nil {
		t.Error("the ability was activated with no {B}")
	}
}

func TestOldB827DrudgeSpellLeavingDestroysEverySkeletonTokenEvenARegeneratingOne(t *testing.T) {
	g, me, opp := oldB827Main(t)
	drudge := effectstest.PushCatalogPermanent(g, me.ID, "Drudge Spell", "Enchantment", oldB827Drudge, false)
	mine := effectstest.B12Creature(g, me.ID, "Skeleton", "Token Creature — Skeleton", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Skeleton", "Token Creature — Skeleton", 2, 1)
	card := effectstest.B12Creature(g, opp.ID, "Skeleton Warrior", "Creature — Skeleton Warrior", 2, 2)
	zombie := effectstest.B12Creature(g, me.ID, "Zombie", "Token Creature — Zombie", 2, 2)
	effectstest.P1Regenerate(t, g, mine)
	effectstest.P1Destroy(t, g, drudge)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a Skeleton token survived Drudge Spell leaving the battlefield")
	}
	if !g.Battlefield.Contains(card) {
		t.Error("a Skeleton card (not a token) was destroyed")
	}
	if !g.Battlefield.Contains(zombie) {
		t.Error("a Zombie token was destroyed")
	}
}

// --- Broken Visage -----------------------------------------------------------

func oldB827VisageCard() game.Card {
	return game.Card{Name: "Broken Visage", TypeLine: "Instant", OracleID: oldB827Visage, ManaCost: "{4}{B}", Colors: []string{"B"}}
}

func oldB827Spirits(g *game.Game) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}

func TestOldB827BrokenVisageDestroysTheAttackerAndMakesASpiritOfItsSizeSacrificedAtTheEndStep(t *testing.T) {
	g, me, opp := oldB827Main(t)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 4)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	if _, err := oldB827Cast(g, oldB827VisageCard(), "CCCCB", game.CastSpellParams{Targets: effectstest.CardRefs(attacker)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking creature was not destroyed")
	}
	sp := oldB827Spirits(g)
	if len(sp) != 1 {
		t.Fatalf("%d Spirit tokens, want 1", len(sp))
	}
	if sp[0].Power != 3 || sp[0].Toughness != 4 || sp[0].Controller != me.ID {
		t.Errorf("Spirit is %d/%d under %s, want a 3/4 under the caster", sp[0].Power, sp[0].Toughness, sp[0].Controller)
	}
	if len(sp[0].Colors) != 1 || sp[0].Colors[0] != "B" || !sp[0].HasSubtype("Spirit") {
		t.Errorf("Spirit is %v %q, want a black Spirit", sp[0].Colors, sp[0].TypeLine)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(oldB827Spirits(g)); got != 0 {
		t.Errorf("%d Spirit tokens left after the end step began, want 0", got)
	}
}

func TestOldB827BrokenVisageCannotBeRegeneratedAroundAndStillMakesTheSpiritFromAnIndestructibleTarget(t *testing.T) {
	g, me, opp := oldB827Main(t)
	shielded := effectstest.B12Creature(g, me.ID, "Shielded", "Creature — Bear", 2, 2)
	effectstest.P1Regenerate(t, g, shielded)
	hard := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hard", TypeLine: "Creature — Golem", Power: 5, Toughness: 6,
		Keywords: []string{"indestructible"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.Pr7bAttack(t, g, opp.ID, shielded, hard)
	if _, err := oldB827Cast(g, oldB827VisageCard(), "CCCCB", game.CastSpellParams{Targets: effectstest.CardRefs(shielded)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shielded) {
		t.Error("a regeneration shield saved the creature")
	}
	if _, err := oldB827Cast(g, oldB827VisageCard(), "CCCCB", game.CastSpellParams{Targets: effectstest.CardRefs(hard)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hard) {
		t.Error("an indestructible creature was destroyed")
	}
	var sizes []int
	for _, s := range oldB827Spirits(g) {
		sizes = append(sizes, s.Power*10+s.Toughness)
	}
	if len(sizes) != 2 {
		t.Fatalf("Spirit sizes %v, want two (2/2 and 5/6)", sizes)
	}
}

func TestOldB827BrokenVisageRefusesAnArtifactCreatureAndANonattacker(t *testing.T) {
	g, me, opp := oldB827Main(t)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, golem, attacker)
	if _, err := oldB827Cast(g, oldB827VisageCard(), "CCCCB", game.CastSpellParams{Targets: effectstest.CardRefs(golem)}); err == nil {
		t.Error("Broken Visage targeted an attacking artifact creature")
	}
	if _, err := oldB827Cast(g, oldB827VisageCard(), "CCCCB", game.CastSpellParams{Targets: effectstest.CardRefs(idle)}); err == nil {
		t.Error("Broken Visage targeted a creature that is not attacking")
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	if _, err := oldB827Cast(g, oldB827VisageCard(), "CCCB", game.CastSpellParams{Targets: effectstest.CardRefs(attacker)}); err == nil {
		t.Error("Broken Visage was cast with four mana")
	}
}

// --- Chain Stasis ------------------------------------------------------------

func oldB827StasisCard() game.Card {
	return game.Card{Name: "Chain Stasis", TypeLine: "Instant", OracleID: oldB827ChainStas, ManaCost: "{U}", Colors: []string{"U"}}
}

// oldB827StasisCast casts Chain Stasis at `target`, resolves it and picks the
// tap-or-untap option (0 tap, 1 untap, 2 neither).
func oldB827StasisCast(t *testing.T, g *game.Game, target uuid.UUID, option int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	if _, err := oldB827Cast(g, oldB827StasisCard(), "U", game.CastSpellParams{Targets: effectstest.CardRefs(target)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, option)
}

func TestOldB827ChainStasisTapsAndATargetsControllerWhoPaysMayCopyIt(t *testing.T) {
	g, me, opp := oldB827Main(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, opp, "CCU")
	oldB827StasisCast(t, g, theirs, 0)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("the creature was not tapped")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, true)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PickCard(t, g, opp.ID, mine)
	copies := effectstest.CopiesOnStack(g)
	if len(copies) != 1 || copies[0].Controller != opp.ID {
		t.Fatalf("want one copy controlled by the creature's controller, got %+v", copies)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	if !effectstest.B16Tapped(t, g, mine) {
		t.Error("the copy did not tap its new target")
	}
}

func TestOldB827ChainStasisUntapsAndAnUnpaidOfferMakesNoCopy(t *testing.T) {
	g, _, opp := oldB827Main(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, theirs).Tapped = true })
	effectstest.FloatForTest(g, opp, "CCU")
	oldB827StasisCast(t, g, theirs, 1)
	if effectstest.B16Tapped(t, g, theirs) {
		t.Error("the creature was not untapped")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	if len(effectstest.CopiesOnStack(g)) != 0 || effectstest.LatestConfirmFor(g, opp.ID) != nil {
		t.Error("a copy was offered although the controller did not pay")
	}
}

func TestOldB827ChainStasisLeavesItAloneAndAControllerWhoCannotPayMakesNoCopy(t *testing.T) {
	g, _, opp := oldB827Main(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oldB827StasisCast(t, g, theirs, 2)
	if effectstest.B16Tapped(t, g, theirs) {
		t.Error("the creature was tapped though the caster left it alone")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, true)
	if len(effectstest.CopiesOnStack(g)) != 0 || effectstest.LatestConfirmFor(g, opp.ID) != nil {
		t.Error("a copy was offered although the controller had nothing to pay with")
	}
}

func TestOldB827ChainStasisRefusesAPlayerAndNeedsItsMana(t *testing.T) {
	g, _, opp := oldB827Main(t)
	if _, err := oldB827Cast(g, oldB827StasisCard(), "U", game.CastSpellParams{Targets: effectstest.PlayerRefs(opp.ID)}); err == nil {
		t.Error("Chain Stasis targeted a player")
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.Seats[g.Turn.ActiveSeat].ManaPool = game.ManaPool{} })
	if _, err := oldB827Cast(g, oldB827StasisCard(), "C", game.CastSpellParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Error("Chain Stasis was cast without blue mana")
	}
}

// --- Giant Albatross ---------------------------------------------------------

// oldB827AlbatrossDies puts the Albatross on the battlefield for me, has each
// of `dealers` damage it (the last one for lethal) and returns with the
// Albatross dead and its trigger resolved to the {1}{U} question.
func oldB827AlbatrossDies(t *testing.T, g *game.Game, me *game.Player, dealers ...uuid.UUID) uuid.UUID {
	t.Helper()
	bird := effectstest.B12Push(g, me.ID, "Giant Albatross", "Creature — Bird", oldB827Albatross, 1, len(dealers))
	for _, d := range dealers {
		effectstest.Pr6Damage(t, g, d, bird, 1)
	}
	if g.Battlefield.Contains(bird) {
		t.Fatal("the Albatross survived")
	}
	effectstest.PassPriorityAroundTable(t, g)
	return bird
}

func TestOldB827GiantAlbatrossDestroysEachDamagerUnlessItsControllerPaysTwoLife(t *testing.T) {
	g, me, opp := oldB827Main(t)
	payer := effectstest.B12Creature(g, opp.ID, "Payer", "Creature — Bear", 2, 4)
	decliner := effectstest.B12Creature(g, opp.ID, "Decliner", "Creature — Bear", 2, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Biter", "Creature — Bear", 2, 4)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 4)
	effectstest.FloatForTest(g, me, "CU")
	oldB827AlbatrossDies(t, g, me, decliner, payer, mine)
	oppLife, myLife := opp.Life, me.Life
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	// Each damaging creature's controller is asked once, for that creature.
	for i := 0; i < 3; i++ {
		var who *game.Player
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePayUnless {
				if c.Chooser == opp.ID {
					who = opp
				} else {
					who = me
				}
				break
			}
		}
		if who == nil {
			t.Fatalf("only %d of 3 life prompts were opened", i)
		}
		// Opponent pays for the first creature they are asked about; declines the second.
		pay := who == me || i == 0
		if who == opp && i == 1 {
			pay = false
		}
		effectstest.AnswerPayUnless(t, g, who.ID, pay)
	}
	effectstest.PassPriorityAroundTable(t, g)
	destroyed := 0
	for _, id := range []uuid.UUID{payer, decliner, mine} {
		if !g.Battlefield.Contains(id) {
			destroyed++
		}
	}
	if destroyed != 1 {
		t.Errorf("%d damaging creatures were destroyed, want exactly the one whose controller declined", destroyed)
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a creature that dealt no damage to the Albatross was destroyed")
	}
	if myLife-me.Life != 2 {
		t.Errorf("I paid %d life for my creature, want 2", myLife-me.Life)
	}
	if oppLife-opp.Life != 2 {
		t.Errorf("the opponent paid %d life in all, want 2 (one of two creatures)", oppLife-opp.Life)
	}
}

func TestOldB827GiantAlbatrossCannotBeRegeneratedAround(t *testing.T) {
	g, me, opp := oldB827Main(t)
	biter := effectstest.B12Creature(g, opp.ID, "Biter", "Creature — Bear", 2, 4)
	effectstest.P1Regenerate(t, g, biter)
	effectstest.FloatForTest(g, me, "CU")
	oldB827AlbatrossDies(t, g, me, biter)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(biter) {
		t.Error("a regeneration shield saved the creature")
	}
}

func TestOldB827GiantAlbatrossDoesNothingIfYouDoNotPayAndIgnoresDamagersThatAreGone(t *testing.T) {
	g, me, opp := oldB827Main(t)
	biter := effectstest.B12Creature(g, opp.ID, "Biter", "Creature — Bear", 2, 4)
	oldB827AlbatrossDies(t, g, me, biter)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(biter) {
		t.Error("the damager was destroyed although the {1}{U} was not paid")
	}
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("the opponent was asked to pay life although nothing was paid")
	}

	g2, me2, opp2 := oldB827Main(t)
	gone := effectstest.B12Creature(g2, opp2.ID, "Gone", "Creature — Bear", 2, 4)
	effectstest.FloatForTest(g2, me2, "CU")
	oldB827AlbatrossDies(t, g2, me2, gone)
	g2.WithWriteLock(func() { g2.Battlefield.Remove(gone) })
	effectstest.AnswerPayUnless(t, g2, me2.ID, true)
	if effectstest.HasPayUnlessFor(g2, opp2.ID) {
		t.Error("the opponent was asked to pay life for a creature that is no longer on the battlefield")
	}
}

func TestOldB827GiantAlbatrossFliesAndAnUndamagedDeathAsksNobody(t *testing.T) {
	g, me, opp := oldB827Main(t)
	bird := effectstest.B12Push(g, me.ID, "Giant Albatross", "Creature — Bird", oldB827Albatross, 1, 1)
	if !effectstest.P1g2HasKeyword(g, bird, "flying") {
		t.Error("Giant Albatross does not fly")
	}
	effectstest.FloatForTest(g, me, "CU")
	effectstest.P1Destroy(t, g, bird)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("someone was asked to pay life though nothing damaged the Albatross")
	}
}
