package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entwine_cards_a_test.go — the first batch of entwine cards (CR 702.42,
// #3866): each cast entwined resolves every bullet in the order written
// (CR 702.42b), and unentwined one bullet alone does what it prints.

const (
	ewaAbunasChantOracle      = "2d4721f1-1a49-45aa-99d5-ab8c6f3f3ab6"
	ewaBarbedLightningOracle  = "f6e1eb52-9f13-425c-8309-1b08ed6557db"
	ewaBetrayalOfFleshOracle  = "43cee9b6-9853-4337-869a-45540c9db10f"
	ewaDreamsGripOracle       = "8b239262-c392-4ec6-9ae1-8b4a076e4bf0"
	ewaFlourishingStrikeOrcl  = "98c44298-7b6a-4f98-bcb7-1a63a4c29f87"
	ewaGoblinWarPartyOracle   = "fa4417e5-229a-4465-9bc5-0b3d6209b4a4"
	ewaGrabTheReinsOracle     = "3ceda405-8175-416a-beda-9a95f5acc60c"
	ewaGravedigOracle         = "c77c3618-bc12-46a8-a4ef-073e359ff3f9"
	ewaInciteWarOracle        = "8a640369-59ab-4506-b473-1f804721414c"
	ewaBlackManaForBetrayal   = "{C}{C}{C}{C}{C}{B}"
	ewaRedPoolBarbedEntwined  = "{C}{C}{C}{C}{R}"
	ewaRedPoolBarbedUntwined  = "{C}{C}{R}"
	ewaGreenPoolStrikeEntwine = "{C}{C}{C}{G}{G}"
)

func ewaCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	return gaLayered(t, g, id)
}

func ewaInGraveyard(p *game.Player, id uuid.UUID) bool { return p.Graveyard.Contains(id) }

func ewaTarget(id uuid.UUID, occ int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id, Mode: occ}
}

func ewaPlayer(id uuid.UUID, occ int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id, Mode: occ}
}

// --- Abuna's Chant ---------------------------------------------------

func TestEwaAbunasChantEntwinedGainsLifeAndShields(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	bear := pushBear(g, me.ID, "Shielded Bear", 10)
	life := me.Life
	if _, err := castEntwineSpell(g, me, "Abuna's Chant", "Instant", "{3}{W}", ewaAbunasChantOracle, []string{"W"},
		"{C}{C}{C}{C}{C}{W}", []int{0, 1}, true, []game.TargetRef{ewaTarget(bear, 1)}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d, want %d", me.Life, life+5)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 6) })
	if got := damageMarkedOn(g, bear); got != 1 {
		t.Errorf("6 into a 5-point shield marked %d, want 1", got)
	}
}

func TestEwaAbunasChantUnentwinedShieldOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	bear := pushBear(g, me.ID, "Shielded Bear", 10)
	life := me.Life
	if _, err := castEntwineSpell(g, me, "Abuna's Chant", "Instant", "{3}{W}", ewaAbunasChantOracle, []string{"W"},
		"{C}{C}{C}{C}{C}{W}", []int{0, 1}, false, []game.TargetRef{ewaTarget(bear, 1)}); err == nil {
		t.Fatal("unentwined: both bullets must be refused")
	}
	if _, err := castEntwineSpell(g, me, "Abuna's Chant", "Instant", "{3}{W}", ewaAbunasChantOracle, []string{"W"},
		"{C}{C}{C}{W}", []int{1}, false, []game.TargetRef{ewaTarget(bear, 0)}); err != nil {
		t.Fatalf("unentwined shield: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d — the life bullet was not chosen", me.Life, life)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 5) })
	if got := damageMarkedOn(g, bear); got != 0 {
		t.Errorf("5 into a 5-point shield marked %d, want 0", got)
	}
}

// --- Barbed Lightning ------------------------------------------------

func TestEwaBarbedLightningEntwinedHitsCreatureAndPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 3, 3)
	life := opp.Life
	if _, err := castEntwineSpell(g, me, "Barbed Lightning", "Instant", "{2}{R}", ewaBarbedLightningOracle, []string{"R"},
		ewaRedPoolBarbedEntwined, []int{0, 1}, true, []game.TargetRef{ewaTarget(bear, 0), ewaPlayer(opp.ID, 1)}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !ewaInGraveyard(opp, bear) {
		t.Error("the 3/3 took 3 and should have died")
	}
	if opp.Life != life-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-3)
	}
}

func TestEwaBarbedLightningUnentwinedPlayerOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 3, 3)
	life := opp.Life
	if _, err := castEntwineSpell(g, me, "Barbed Lightning", "Instant", "{2}{R}", ewaBarbedLightningOracle, []string{"R"},
		ewaRedPoolBarbedUntwined, []int{1}, false, []game.TargetRef{ewaPlayer(opp.ID, 0)}); err != nil {
		t.Fatalf("unentwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 || ewaInGraveyard(opp, bear) {
		t.Errorf("player bullet only: life %d (want %d), bear dead = %v", opp.Life, life-3, ewaInGraveyard(opp, bear))
	}
}

// --- Betrayal of Flesh -----------------------------------------------

type ewaBetrayalBoard struct {
	me, opp *game.Player
	lands   []uuid.UUID
	theirs  uuid.UUID
	dead    uuid.UUID
}

func ewaNewBetrayalBoard(t *testing.T, lands int) (*game.Game, ewaBetrayalBoard) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	b := ewaBetrayalBoard{me: me, opp: opp}
	for range lands {
		b.lands = append(b.lands, pushLand(g, me.ID, "Island"))
	}
	b.theirs = pushVanillaCreature(g, opp.ID, "Their Ogre", 3, 3)
	b.dead = b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	return g, b
}

func (b ewaBetrayalBoard) cast(g *game.Game, modes []int, entwined bool, sacrifice []uuid.UUID) error {
	id := handCardFull(b.me, "Betrayal of Flesh", "Instant", "{5}{B}", ewaBetrayalOfFleshOracle, []string{"B"})
	b.me.ManaPool.EmptyPool()
	if err := g.AddManaForEffect(b.me.ID, uuid.Nil, ewaBlackManaForBetrayal); err != nil {
		return err
	}
	var optional []int
	if entwined {
		optional = []int{0}
	}
	var targets []game.TargetRef
	for occ, m := range modes {
		if m == 0 {
			targets = append(targets, ewaTarget(b.theirs, occ))
		} else {
			targets = append(targets, ewaTarget(b.dead, occ))
		}
	}
	err := g.CastSpell(b.me.ID, id, game.CastSpellParams{
		Strict: true, Modes: modes, OptionalCosts: optional, Targets: targets, SacrificeIDs: sacrifice,
	})
	if err != nil {
		_, _ = b.me.Hand.Remove(id)
		b.me.ManaPool.EmptyPool()
	}
	return err
}

func TestEwaBetrayalOfFleshEntwinedSacrificesThreeLands(t *testing.T) {
	g, b := ewaNewBetrayalBoard(t, 3)
	if err := b.cast(g, []int{0, 1}, true, b.lands[:2]); err == nil {
		t.Fatal("entwined with two lands named: the entwine cost is three")
	}
	if err := b.cast(g, []int{0, 1}, false, nil); err == nil {
		t.Fatal("both bullets without the entwine cost must be refused")
	}
	if err := b.cast(g, []int{0, 1}, true, b.lands); err != nil {
		t.Fatalf("entwined with three lands: %v", err)
	}
	for _, l := range b.lands {
		if !b.me.Graveyard.Contains(l) {
			t.Error("a land named for the entwine cost was not sacrificed")
		}
	}
	passPriorityAroundTable(t, g)
	if !ewaInGraveyard(b.opp, b.theirs) {
		t.Error("the destroy bullet left the Ogre alive")
	}
	if !g.Battlefield.Contains(b.dead) {
		t.Error("the return bullet left the bear in the graveyard")
	}
}

func TestEwaBetrayalOfFleshRefusesEntwineWithTooFewLands(t *testing.T) {
	g, b := ewaNewBetrayalBoard(t, 2)
	if err := b.cast(g, []int{0, 1}, true, b.lands); err == nil {
		t.Fatal("two lands cannot pay \"Sacrifice three lands\"")
	}
	if err := b.cast(g, []int{1}, false, nil); err != nil {
		t.Fatalf("unentwined return: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(b.dead) || ewaInGraveyard(b.opp, b.theirs) {
		t.Error("unentwined, only the return bullet resolves")
	}
	for _, l := range b.lands {
		if !g.Battlefield.Contains(l) {
			t.Error("an unentwined cast sacrificed a land")
		}
	}
}

// --- Dream's Grip ----------------------------------------------------

func TestEwaDreamsGripEntwinedTapsAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := pushLand(g, me.ID, "Island")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	if _, err := castEntwineSpell(g, me, "Dream's Grip", "Instant", "{U}", ewaDreamsGripOracle, []string{"U"},
		"{C}{U}", []int{0, 1}, true, []game.TargetRef{ewaTarget(theirs, 0), ewaTarget(mine, 1)}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !ewaCard(t, g, theirs).Tapped {
		t.Error("the tap bullet left the bear untapped")
	}
	if ewaCard(t, g, mine).Tapped {
		t.Error("the untap bullet left the land tapped")
	}
}

func TestEwaDreamsGripUnentwinedUntapsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(theirs) })
	if _, err := castEntwineSpell(g, me, "Dream's Grip", "Instant", "{U}", ewaDreamsGripOracle, []string{"U"},
		"{U}", []int{1}, false, []game.TargetRef{ewaTarget(theirs, 0)}); err != nil {
		t.Fatalf("unentwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if ewaCard(t, g, theirs).Tapped {
		t.Error("the untap bullet left the bear tapped")
	}
}

// --- Flourishing Strike ----------------------------------------------

func TestEwaFlourishingStrikeEntwinedKillsFlierAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Drake", TypeLine: "Creature — Drake",
		Power: 2, Toughness: 5, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	if _, err := castEntwineSpell(g, me, "Flourishing Strike", "Instant", "{1}{G}", ewaFlourishingStrikeOrcl, []string{"G"},
		ewaGreenPoolStrikeEntwine, []int{0, 1}, true, []game.TargetRef{ewaTarget(flier, 0), ewaTarget(mine, 1)}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !ewaInGraveyard(opp, flier) {
		t.Error("the 5-toughness flier took 5 and should have died")
	}
	if p, tg := effectivePTOf(t, g, mine); p != 5 || tg != 5 {
		t.Errorf("pumped bear is %d/%d, want 5/5", p, tg)
	}
}

func TestEwaFlourishingStrikeDamageNeedsAFlier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	ground := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if _, err := castEntwineSpell(g, me, "Flourishing Strike", "Instant", "{1}{G}", ewaFlourishingStrikeOrcl, []string{"G"},
		"{C}{G}", []int{0}, false, []game.TargetRef{ewaTarget(ground, 0)}); err == nil {
		t.Fatal("\"target creature with flying\" accepted a creature without flying")
	}
	if _, err := castEntwineSpell(g, me, "Flourishing Strike", "Instant", "{1}{G}", ewaFlourishingStrikeOrcl, []string{"G"},
		"{C}{G}", []int{1}, false, []game.TargetRef{ewaTarget(ground, 0)}); err != nil {
		t.Fatalf("the pump bullet on any creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, ground); p != 5 {
		t.Errorf("pumped bear has power %d, want 5", p)
	}
}

// --- Goblin War Party ------------------------------------------------

func ewaGoblinsOf(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == RedGoblinToken().Name {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

func TestEwaGoblinWarPartyEntwinedPumpsItsOwnGoblins(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	if _, err := castEntwineSpell(g, me, "Goblin War Party", "Sorcery", "{3}{R}", ewaGoblinWarPartyOracle, []string{"R"},
		"{C}{C}{C}{C}{C}{R}{R}", []int{1, 0}, true, nil); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	gobs := ewaGoblinsOf(g, me.ID)
	if len(gobs) != 3 {
		t.Fatalf("%d Goblins, want 3", len(gobs))
	}
	// CR 702.42b: the tokens come first, so the pump reaches them.
	for _, id := range append(gobs, bear) {
		c := ewaCard(t, g, id)
		if !game.HasKeyword(&c, "haste") {
			t.Errorf("%s has no haste", c.Name)
		}
	}
	if p, tg := effectivePTOf(t, g, gobs[0]); p != 2 || tg != 2 {
		t.Errorf("a Goblin is %d/%d, want 2/2", p, tg)
	}
	if p, _ := effectivePTOf(t, g, bear); p != 3 {
		t.Errorf("the bear has power %d, want 3", p)
	}
}

func TestEwaGoblinWarPartyUnentwinedTokensOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	if _, err := castEntwineSpell(g, me, "Goblin War Party", "Sorcery", "{3}{R}", ewaGoblinWarPartyOracle, []string{"R"},
		"{C}{C}{C}{R}", []int{0}, false, nil); err != nil {
		t.Fatalf("unentwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	gobs := ewaGoblinsOf(g, me.ID)
	if len(gobs) != 3 {
		t.Fatalf("%d Goblins, want 3", len(gobs))
	}
	if c := ewaCard(t, g, gobs[0]); c.Power != 1 || game.HasKeyword(&c, "haste") {
		t.Errorf("an unpumped Goblin is %d power, haste %v", c.Power, game.HasKeyword(&c, "haste"))
	}
}

// --- Grab the Reins --------------------------------------------------

func TestEwaGrabTheReinsEntwinedStealsAndFlingsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	ogre := pushVanillaCreature(g, opp.ID, "Their Ogre", 4, 4)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	life := opp.Life
	if _, err := castEntwineSpell(g, me, "Grab the Reins", "Instant", "{3}{R}", ewaGrabTheReinsOracle, []string{"R"},
		"{C}{C}{C}{C}{C}{R}{R}", []int{0, 1}, true, []game.TargetRef{ewaTarget(ogre, 0), ewaPlayer(opp.ID, 1)}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := ewaCard(t, g, ogre); c.Controller != me.ID {
		t.Fatalf("the steal bullet ran first (CR 702.42b): the Ogre is controlled by %s", c.Controller)
	}
	// The prompt is asked (two creatures to choose from), and the
	// stolen Ogre is among them.
	if sacrificeChoiceFor(g, me.ID) == nil {
		t.Fatal("the sacrifice bullet did not ask which creature")
	}
	answerSacrifice(t, g, me.ID, ogre)
	g.SettleResolution()
	if !g.Battlefield.Contains(mine) {
		t.Error("the bear that was not chosen was sacrificed")
	}
	if !ewaInGraveyard(opp, ogre) {
		t.Error("the stolen Ogre was not sacrificed")
	}
	if opp.Life != life-4 {
		t.Errorf("opponent life %d, want %d — the Ogre's power", opp.Life, life-4)
	}
}

func TestEwaGrabTheReinsUnentwinedStealsWithHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	ogre := pushVanillaCreature(g, opp.ID, "Their Ogre", 4, 4)
	if _, err := castEntwineSpell(g, me, "Grab the Reins", "Instant", "{3}{R}", ewaGrabTheReinsOracle, []string{"R"},
		"{C}{C}{C}{R}", []int{0}, false, []game.TargetRef{ewaTarget(ogre, 0)}); err != nil {
		t.Fatalf("unentwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := ewaCard(t, g, ogre)
	if c.Controller != me.ID || !game.HasKeyword(&c, "haste") {
		t.Errorf("the Ogre: controller %s (want me), haste %v", c.Controller, game.HasKeyword(&c, "haste"))
	}
	if choiceFor(g, me.ID) {
		t.Error("the unchosen sacrifice bullet asked for a sacrifice")
	}
}

// --- Gravedig --------------------------------------------------------

func TestEwaGravedigEntwinedTokenForTargetPlayerAndReturn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	dead := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	if _, err := castEntwineSpell(g, me, "Gravedig", "Sorcery", "{1}{B}", ewaGravedigOracle, []string{"B"},
		"{C}{C}{C}{B}", []int{0, 1}, true, []game.TargetRef{ewaPlayer(opp.ID, 0), ewaTarget(dead, 1)}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	zombies := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == BlackZombieToken().Name && c.Controller == opp.ID {
				zombies++
			}
		}
	})
	if zombies != 1 {
		t.Errorf("the target player has %d Zombies, want 1", zombies)
	}
	if !me.Hand.Contains(dead) {
		t.Error("the return bullet left the bear in the graveyard")
	}
}

func TestEwaGravedigUnentwinedReturnsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	dead := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	if _, err := castEntwineSpell(g, me, "Gravedig", "Sorcery", "{1}{B}", ewaGravedigOracle, []string{"B"},
		"{C}{B}", []int{1}, false, []game.TargetRef{ewaTarget(dead, 0)}); err != nil {
		t.Fatalf("unentwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the return bullet left the bear in the graveyard")
	}
}

// --- Incite War ------------------------------------------------------

// Cast in the opponent's upkeep: their creatures must attack this turn,
// and my own creature has first strike.
func TestEwaInciteWarEntwinedForcesTheirAttackAndGivesFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % len(g.Seats)
	opp := g.Seats[oppSeat]
	third := g.Seats[(seat+2)%len(g.Seats)]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	advanceToUpkeepOf(t, g, oppSeat)
	if _, err := castEntwineSpell(g, me, "Incite War", "Instant", "{2}{R}", ewaInciteWarOracle, []string{"R"},
		"{C}{C}{C}{C}{R}", []int{0, 1}, true, []game.TargetRef{ewaPlayer(opp.ID, 0)}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := ewaCard(t, g, mine); !game.HasKeyword(&c, "first strike") {
		t.Error("my bear has no first strike")
	}
	advanceToDeclareAttackersOf(t, g, oppSeat)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("the opponent's pass with their bear home = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(theirs, third.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the bear attacking: %v", err)
	}
}

func TestEwaInciteWarUnentwinedFirstStrikeOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	if _, err := castEntwineSpell(g, me, "Incite War", "Instant", "{2}{R}", ewaInciteWarOracle, []string{"R"},
		"{C}{C}{R}", []int{1}, false, nil); err != nil {
		t.Fatalf("unentwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := ewaCard(t, g, mine); !game.HasKeyword(&c, "first strike") {
		t.Error("my bear has no first strike")
	}
	if hasRequirementRecord(g) {
		t.Error("the unchosen attack bullet registered a requirement")
	}
}
