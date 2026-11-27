package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 583 (Saviors of Kamigawa): Burning-Eye Zubera, Kiyomaro, First
// to Stand, Matsu-Tribe Birdstalker, Descendant of Kiyomaro, Godo's
// Irregulars, Kiku's Shadow, Kashi-Tribe Elite, Ebony Owl Netsuke.

const (
	oldB583Zubera     = "24bf8773-c8bb-494c-a0af-aef80f5b3d47"
	oldB583Kiyomaro   = "340ec369-1761-4134-a96d-5537b4e08703"
	oldB583Birdstalk  = "4c90ca5c-551c-4da5-9482-62521b1ae17c"
	oldB583Descendant = "bbf07bec-9c02-455b-885d-7cb5fc77b5de"
	oldB583Godo       = "bcf0ca90-134f-43af-890a-558c5e82974e"
	oldB583Shadow     = "ce1678a1-441a-4303-a83a-d7d138b366fd"
	oldB583Elite      = "dac9bbfc-62e9-42a9-8ad6-2530ba47b990"
	oldB583Netsuke    = "dcacb717-4ca4-47c0-9b23-26e4eb4ffdc8"
)

func TestOldB583EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB583Zubera: "Burning-Eye Zubera", oldB583Kiyomaro: "Kiyomaro, First to Stand",
		oldB583Birdstalk: "Matsu-Tribe Birdstalker", oldB583Descendant: "Descendant of Kiyomaro",
		oldB583Godo: "Godo's Irregulars", oldB583Shadow: "Kiku's Shadow",
		oldB583Elite: "Kashi-Tribe Elite", oldB583Netsuke: "Ebony Owl Netsuke",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

// --- Burning-Eye Zubera -------------------------------------------------------

func oldB583ZuberaDies(t *testing.T, hits ...int) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	z := effectstest.InvB1Push(g, me.ID, "Burning-Eye Zubera", "Creature — Zubera Spirit", oldB583Zubera, 3, 3)
	src := effectstest.InvB1Push(g, opp.ID, "Pinger", "Creature — Goblin", "", 1, 9)
	for _, n := range hits {
		effectstest.Pr6Damage(t, g, src, z, n)
	}
	return g, z, opp.ID
}

func TestOldB583ZuberaDealsThreeWhenFourDamageWasDealtToIt(t *testing.T) {
	// 2 then 2: it survives the first hit and dies to the second with 4 marked.
	g, z, oppID := oldB583ZuberaDies(t, 2, 2)
	me, opp := g.Seats[0], g.Seats[1]
	if g.Battlefield.Contains(z) {
		t.Fatal("the Zubera survived 4 damage")
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickPlayer(t, g, me.ID, oppID)
	}
	before := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, before-3)
	}
}

func TestOldB583ZuberaDoesNothingWhenOnlyThreeDamageWasDealt(t *testing.T) {
	g, z, _ := oldB583ZuberaDies(t, 3)
	me, opp := g.Seats[0], g.Seats[1]
	if g.Battlefield.Contains(z) {
		t.Fatal("the Zubera survived 3 damage")
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("the Zubera asked for a target after 3 damage: %+v", p)
	}
	before := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("the opponent lost %d life", before-opp.Life)
	}
}

func TestOldB583ZuberaSacrificedWithoutDamageDoesNothing(t *testing.T) {
	g, z, _ := oldB583ZuberaDies(t)
	me := g.Seats[0]
	effectstest.KillCreature(t, g, me.ID, z)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("the Zubera asked for a target with no damage: %+v", p)
	}
}

// --- Kiyomaro, First to Stand -------------------------------------------------

func oldB583Kiyomaro1(t *testing.T, hand int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	k := effectstest.InvB1Push(g, me.ID, "Kiyomaro, First to Stand", "Legendary Creature — Spirit", oldB583Kiyomaro, 0, 0)
	effectstest.D2dSetHand(me, hand)
	return g, k
}

func TestOldB583KiyomaroIsAsBigAsYourHandAndHasVigilanceAtFour(t *testing.T) {
	g, k := oldB583Kiyomaro1(t, 3)
	me := g.Seats[0]
	if p, tough := effectstest.PtOf(t, g, k); p != 3 || tough != 3 {
		t.Errorf("with 3 cards Kiyomaro is %d/%d, want 3/3", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, k, "vigilance") {
		t.Error("vigilance with 3 cards in hand")
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, k); p != 4 || tough != 4 {
		t.Errorf("with 4 cards Kiyomaro is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, k, "vigilance") {
		t.Error("no vigilance with 4 cards in hand")
	}
	if err := g.DrawCard(g.Seats[1].ID); err != nil {
		t.Fatalf("DrawCard (opponent): %v", err)
	}
	if p, _ := effectstest.PtOf(t, g, k); p != 4 {
		t.Errorf("an opponent's hand changed Kiyomaro's power to %d", p)
	}
}

func TestOldB583KiyomaroGainsSevenWhenItDealsDamageWithSevenCards(t *testing.T) {
	g, k := oldB583Kiyomaro1(t, 7)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Life
	effectstest.EmitNoncombatDamage(g, k, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+7 {
		t.Errorf("life %d, want %d", me.Life, before+7)
	}
}

func TestOldB583KiyomaroGainsNothingWithSixCards(t *testing.T) {
	g, k := oldB583Kiyomaro1(t, 6)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Life
	effectstest.EmitNoncombatDamage(g, k, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life %d, want %d", me.Life, before)
	}
}

// --- Descendant of Kiyomaro ---------------------------------------------------

func oldB583Descendant1(t *testing.T, mine int, theirs ...int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	d := effectstest.InvB1Push(g, me.ID, "Descendant of Kiyomaro", "Creature — Human Soldier", oldB583Descendant, 2, 3)
	effectstest.D2dSetHand(me, mine)
	for i, n := range theirs {
		effectstest.D2dSetHand(g.Seats[i+1], n)
	}
	return g, d
}

func TestOldB583DescendantGrowsAndGainsLifeWhileYouHoldTheMostCards(t *testing.T) {
	g, d := oldB583Descendant1(t, 5, 4, 3, 2)
	me, opp := g.Seats[0], g.Seats[1]
	if p, tough := effectstest.PtOf(t, g, d); p != 3 || tough != 5 {
		t.Fatalf("Descendant is %d/%d with the most cards, want 3/5", p, tough)
	}
	before := me.Life
	effectstest.DealCombatDamageToPlayer(g, d, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("life %d, want %d", me.Life, before+3)
	}
}

func TestOldB583DescendantNeedsMoreThanEachOpponent(t *testing.T) {
	g, d := oldB583Descendant1(t, 4, 4, 1, 0)
	me, opp := g.Seats[0], g.Seats[1]
	if p, tough := effectstest.PtOf(t, g, d); p != 2 || tough != 3 {
		t.Fatalf("Descendant is %d/%d on a tie, want 2/3", p, tough)
	}
	before := me.Life
	effectstest.DealCombatDamageToPlayer(g, d, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life %d on a tie, want %d", me.Life, before)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, d); p != 3 || tough != 5 {
		t.Errorf("Descendant is %d/%d after drawing ahead, want 3/5", p, tough)
	}
}

// --- Kiku's Shadow ------------------------------------------------------------

func TestOldB583KikusShadowMakesACreatureDamageItselfByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.InvB1Push(g, opp.ID, "Big", "Creature — Ogre", "", 4, 4)
	small := effectstest.InvB1Push(g, opp.ID, "Small", "Creature — Ogre", "", 2, 3)
	effectstest.CastCatalogSpell(t, g, "Kiku's Shadow", "Sorcery", oldB583Shadow, effectstest.CardTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 4/4 survived dealing 4 damage to itself")
	}
	effectstest.CastCatalogSpell(t, g, "Kiku's Shadow", "Sorcery", oldB583Shadow, effectstest.CardTarget(small))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Fatal("the 2/3 died to 2 damage")
	}
	if got := effectstest.Pr6Marked(g, small); got != 2 {
		t.Errorf("damage marked on the 2/3 is %d, want 2", got)
	}
}

// --- Ebony Owl Netsuke --------------------------------------------------------

func oldB583NetsukeUpkeep(t *testing.T, upkeepSeat, hand int) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Ebony Owl Netsuke", oldB583Netsuke, "Artifact")
	victim := g.Seats[upkeepSeat]
	effectstest.D2dSetHand(victim, hand)
	effectstest.AdvanceToUpkeepOf(t, g, upkeepSeat)
	return g, victim
}

func TestOldB583NetsukeDealsFourToAnOpponentWithSevenCards(t *testing.T) {
	g, opp := oldB583NetsukeUpkeep(t, 1, 7)
	start := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != start-4 {
		t.Errorf("life %d -> %d, want -4", start, opp.Life)
	}
}

func TestOldB583NetsukeIgnoresSixCardsAndItsOwnersUpkeep(t *testing.T) {
	g, opp := oldB583NetsukeUpkeep(t, 1, 6)
	start := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != start {
		t.Errorf("six cards: lost %d", start-opp.Life)
	}
	g, me := oldB583NetsukeUpkeep(t, 0, 9)
	start = me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start {
		t.Errorf("own upkeep: lost %d", start-me.Life)
	}
}

// --- Godo's Irregulars --------------------------------------------------------

func TestOldB583GodoPingsABlockerForR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	godo := effectstest.InvB1Push(g, me.ID, "Godo's Irregulars", "Creature — Human Warrior", oldB583Godo, 0, 1) // power 0: the blocker's damage is the ability's alone
	wall := effectstest.InvB1Push(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
	other := effectstest.InvB1Push(g, opp.ID, "Bystander", "Creature — Wall", "", 0, 4)
	effectstest.Mrd3BlockedBy(t, g, opp.ID, godo, wall)
	if err := effectstest.InvB1Activate(g, me.ID, godo, 0, effectstest.CardTarget(other)...); err == nil {
		t.Fatal("targeted a creature that is not blocking it")
	}
	effectstest.FillPoolColored(me, "R", 1)
	if err := effectstest.InvB1Activate(g, me.ID, godo, 0, effectstest.CardTarget(wall)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, wall); got != 1 {
		t.Errorf("the blocker has %d damage, want 1", got)
	}
}

func TestOldB583GodoCannotBeActivatedWithNothingBlockingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	godo := effectstest.InvB1Push(g, me.ID, "Godo's Irregulars", "Creature — Human Warrior", oldB583Godo, 1, 1)
	wall := effectstest.InvB1Push(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
	effectstest.FillPoolColored(me, "R", 1)
	if err := effectstest.InvB1Activate(g, me.ID, godo, 0, effectstest.CardTarget(wall)...); err == nil {
		t.Error("activated with no blocker")
	}
}

// --- Matsu-Tribe Birdstalker and Kashi-Tribe Elite ----------------------------

func oldB583FrozenBy(t *testing.T, name, typeLine, oracle string) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hitter := effectstest.InvB1Push(g, me.ID, name, typeLine, oracle, 2, 2)
	wall := effectstest.InvB1Push(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
	effectstest.Mrd3BlockedBy(t, g, opp.ID, hitter, wall)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	return g, hitter, wall
}

func oldB583Holds(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("battlefield card %s is missing", id)
	}
	return len(c.NextUntapSkips)
}

func TestOldB583BirdstalkerTapsAndFreezesTheCreatureItDamaged(t *testing.T) {
	g, _, wall := oldB583FrozenBy(t, "Matsu-Tribe Birdstalker", "Creature — Snake Warrior Archer", oldB583Birdstalk)
	if !effectstest.IsTapped(g, wall) {
		t.Error("the damaged creature is not tapped")
	}
	if got := oldB583Holds(t, g, wall); got != 1 {
		t.Errorf("untap holds on the damaged creature: %d, want 1", got)
	}
}

func TestOldB583EliteTapsAndFreezesTheCreatureItDamaged(t *testing.T) {
	g, _, wall := oldB583FrozenBy(t, "Kashi-Tribe Elite", "Creature — Snake Warrior", oldB583Elite)
	if !effectstest.IsTapped(g, wall) {
		t.Error("the damaged creature is not tapped")
	}
	if got := oldB583Holds(t, g, wall); got != 1 {
		t.Errorf("untap holds on the damaged creature: %d, want 1", got)
	}
}

func TestOldB583BirdstalkerIgnoresNoncombatDamageAndPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := effectstest.InvB1Push(g, me.ID, "Matsu-Tribe Birdstalker", "Creature — Snake Warrior Archer", oldB583Birdstalk, 2, 2)
	wall := effectstest.InvB1Push(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
	effectstest.EmitNoncombatDamage(g, bird, wall, 1)
	effectstest.DealCombatDamageToPlayer(g, bird, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, wall) || oldB583Holds(t, g, wall) != 0 {
		t.Error("the creature was frozen without combat damage dealt to it")
	}
}

func TestOldB583BirdstalkerGainsReachForG(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bird := effectstest.InvB1Push(g, me.ID, "Matsu-Tribe Birdstalker", "Creature — Snake Warrior Archer", oldB583Birdstalk, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bird, "reach") {
		t.Fatal("reach before the ability")
	}
	if err := effectstest.InvB1Activate(g, me.ID, bird, 0); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.FillPoolColored(me, "G", 1)
	if err := effectstest.InvB1Activate(g, me.ID, bird, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bird, "reach") {
		t.Error("no reach after {G}")
	}
}

func TestOldB583EliteGivesShroudToYourLegendarySnakesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elite := effectstest.InvB1Push(g, me.ID, "Kashi-Tribe Elite", "Creature — Snake Warrior", oldB583Elite, 2, 3)
	legend := effectstest.InvB1Push(g, me.ID, "Legendary Snake", "Legendary Creature — Snake", "", 2, 2)
	plain := effectstest.InvB1Push(g, me.ID, "Plain Snake", "Creature — Snake", "", 2, 2)
	legendOther := effectstest.InvB1Push(g, me.ID, "Legendary Ogre", "Legendary Creature — Ogre", "", 2, 2)
	theirs := effectstest.InvB1Push(g, opp.ID, "Their Legendary Snake", "Legendary Creature — Snake", "", 2, 2)
	for id, want := range map[uuid.UUID]bool{elite: false, legend: true, plain: false, legendOther: false, theirs: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "shroud"); got != want {
			t.Errorf("shroud on %s: %v, want %v", id, got, want)
		}
	}
}
