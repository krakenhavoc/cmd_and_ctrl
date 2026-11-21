package testscm

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s30_cards_test.go — batch 30 of the wave-2 Commander-deck cards
// (#5368): reanimation and graveyard creatures.

const (
	s30StormOfSoulsOracle = "7022ae9a-2589-493b-a050-a230fd02f80b"
)

func TestCmdr2S30StormOfSoulsReturnsEveryCreatureAsAOneOneSpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	golem := effectstest.PushGraveyardCardTyped(me, "Dead Golem", "Artifact Creature — Golem")
	stone := effectstest.PushGraveyardCardTyped(me, "Dead Stone", "Artifact")
	theirs := effectstest.PushGraveyardCardTyped(opp, "Their Elf", "Creature — Elf")
	storm := effectstest.CastCatalogSpell(t, g, "Storm of Souls", "Sorcery", s30StormOfSoulsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, golem} {
		if !g.Battlefield.Contains(id) {
			t.Fatalf("creature card %s did not return", id)
		}
		c := effectstest.LayeredCard(t, g, id)
		eff := c.Effective()
		if !c.HasSubtype("Spirit") || !slices.Contains(eff.Abilities, "flying") || eff.Power != 1 || eff.Toughness != 1 {
			t.Errorf("%s: subtypes %v abilities %v %d/%d, want a 1/1 Spirit with flying", c.Name, eff.Subtypes, eff.Abilities, eff.Power, eff.Toughness)
		}
	}
	if !effectstest.LayeredCard(t, g, bear).HasSubtype("Bear") {
		t.Error("it keeps its other types")
	}
	if g.Battlefield.Contains(stone) || g.Battlefield.Contains(theirs) {
		t.Error("a noncreature card and an opponent's creature card stay put")
	}
	if !effectstest.InExile(g, storm) {
		t.Error("Storm of Souls is exiled")
	}
}

const (
	s30MarshalsAnthemOracle = "41fb2b0e-d0a0-4e47-a6f3-fc9860ade65a"
	s30RotHulkOracle        = "36117e99-a2bc-46f9-8e17-bd2e2a9c31f4"
)

// s30PickTargets answers the latest pick_target prompt with every card given.
func s30PickTargets(t *testing.T, g *game.Game, chooser uuid.UUID, cards ...uuid.UUID) {
	t.Helper()
	pick := effectstest.LatestPickTarget(g, chooser)
	if pick == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	var refs []game.TargetRef
	for _, c := range cards {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: c})
	}
	if err := g.ResolvePickTargets(pick.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

func TestCmdr2S30RotHulkReturnsUpToOneZombiePerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opponents := len(g.Seats) - 1
	var zombies []uuid.UUID
	for i := 0; i < opponents+1; i++ {
		zombies = append(zombies, effectstest.PushGraveyardCardTyped(me, "Zombie", "Creature — Zombie"))
	}
	elf := effectstest.PushGraveyardCardTyped(me, "Dead Elf", "Creature — Elf")
	effectstest.CastCatalogSpell(t, g, "Rot Hulk", "Creature — Zombie", s30RotHulkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// One more pick than there are opponents is refused.
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	var tooMany []game.TargetRef
	for _, z := range zombies {
		tooMany = append(tooMany, game.TargetRef{Kind: game.TargetCard, ID: z})
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, tooMany); err == nil {
		t.Error("more targets than opponents was accepted")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: elf}}); err == nil {
		t.Error("a non-Zombie was accepted")
	}
	s30PickTargets(t, g, me.ID, zombies[:opponents]...)
	effectstest.PassPriorityAroundTable(t, g)
	for _, z := range zombies[:opponents] {
		if !g.Battlefield.Contains(z) {
			t.Errorf("Zombie %s did not return", z)
		}
	}
	if g.Battlefield.Contains(zombies[opponents]) || g.Battlefield.Contains(elf) {
		t.Error("only the chosen Zombies return")
	}
}

func TestCmdr2S30MarshalsAnthemPumpsAndReturnsOnePerKick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushGraveyardCardTyped(me, "Dead A", "Creature — Bear")
	b := effectstest.PushGraveyardCardTyped(me, "Dead B", "Creature — Bear")
	c := effectstest.PushGraveyardCardTyped(me, "Dead C", "Creature — Bear")
	anthem, err := effectstest.CastWithOptionalCosts(t, g, "Marshal's Anthem", "Enchantment", s30MarshalsAnthemOracle, nil, []int{0, 0}, nil)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Kicked twice: three targets are too many.
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}); err == nil {
		t.Error("three targets accepted for a twice-kicked Anthem")
	}
	s30PickTargets(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(anthem) || !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) || g.Battlefield.Contains(c) {
		t.Fatal("the Anthem and exactly the two chosen creatures are on the battlefield")
	}
	if e := effectstest.LayeredCard(t, g, a).Effective(); e.Power != 3 || e.Toughness != 3 {
		t.Errorf("returned creature is %d/%d, want 3/3 with the Anthem", e.Power, e.Toughness)
	}
}

func TestCmdr2S30MarshalsAnthemUnkickedReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushGraveyardCardTyped(me, "Dead A", "Creature — Bear")
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Marshal's Anthem", "Enchantment", s30MarshalsAnthemOracle, nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || g.Battlefield.Contains(a) {
		t.Error("an unkicked Anthem returns nothing")
	}
}

const s30MoldgrafMonstrosityOracle = "3c4d5392-27cf-492e-9ae1-dee330b82223"

func TestCmdr2S30MoldgrafMonstrosityExilesItselfAndReturnsTwoAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hulk := effectstest.PushCatalogPermanent(g, me.ID, "Moldgraf Monstrosity", "Creature — Insect", s30MoldgrafMonstrosityOracle, false)
	var dead []uuid.UUID
	for i := 0; i < 3; i++ {
		dead = append(dead, effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear"))
	}
	land := effectstest.PushGraveyardCardTyped(me, "Dead Land", "Land")
	effectstest.KillCreature(t, g, me.ID, hulk)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, hulk) {
		t.Error("the Monstrosity is exiled")
	}
	back := 0
	for _, id := range dead {
		if g.Battlefield.Contains(id) {
			back++
		}
	}
	if back != 2 {
		t.Errorf("%d creature cards returned, want exactly 2", back)
	}
	if g.Battlefield.Contains(land) || g.Battlefield.Contains(hulk) {
		t.Error("a land card and the Monstrosity itself stay out")
	}
}

func TestCmdr2S30MoldgrafMonstrosityWithOneCreatureCardReturnsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hulk := effectstest.PushCatalogPermanent(g, me.ID, "Moldgraf Monstrosity", "Creature — Insect", s30MoldgrafMonstrosityOracle, false)
	only := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.KillCreature(t, g, me.ID, hulk)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(only) || !effectstest.InExile(g, hulk) {
		t.Error("the single creature card returns and the Monstrosity is exiled")
	}
}

const (
	s30CarrionGrubOracle       = "c54b375d-a456-4886-96f2-8dc7a0024a5b"
	s30MoldgrafMillipedeOracle = "48edac92-0dd8-4274-bf1b-122243770540"
)

func TestCmdr2S30CarrionGrubReadsTheGreatestPowerInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grub := effectstest.PushCatalogPermanent(g, me.ID, "Carrion Grub", "Creature — Insect", s30CarrionGrubOracle, false)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == grub {
			g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 0, 5
		}
	}
	if e := effectstest.LayeredCard(t, g, grub).Effective(); e.Power != 0 {
		t.Fatalf("empty graveyard: power %d, want 0", e.Power)
	}
	toGraveyard := func(p *game.Player, name, typeLine string, power int) {
		id := uuid.New()
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: power, Toughness: 1, Owner: p.ID, Controller: p.ID})
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: p.ID}, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}
	toGraveyard(me, "Small", "Creature — Bear", 2)
	toGraveyard(me, "Big", "Creature — Bear", 6)
	toGraveyard(me, "Huge Artifact", "Artifact", 9)
	toGraveyard(opp, "Their Giant", "Creature — Giant", 9)
	if e := effectstest.LayeredCard(t, g, grub).Effective(); e.Power != 6 || e.Toughness != 5 {
		t.Errorf("got %d/%d, want 6/5: only creature cards in my graveyard count", e.Power, e.Toughness)
	}
}

func TestCmdr2S30CarrionGrubMillsFourOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 6; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery"})
	}
	before, gy := me.Library.Size(), me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Carrion Grub", "Creature — Insect", s30CarrionGrubOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != before-4 || me.Graveyard.Size() != gy+4 {
		t.Errorf("library %d->%d graveyard %d->%d, want four milled", before, me.Library.Size(), gy, me.Graveyard.Size())
	}
}

// s30OnTop puts a card on top of a library (pushLibraryCardForTest goes to the bottom).
func s30OnTop(p *game.Player, name, typeLine string) {
	p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
}

func TestCmdr2S30MoldgrafMillipedeCountsTheCardsItJustMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	s30OnTop(me, "Deep", "Creature — Bear")
	for i := 0; i < 2; i++ {
		s30OnTop(me, "Milled Bear", "Creature — Bear")
	}
	s30OnTop(me, "Spell", "Sorcery")
	id := effectstest.CastCatalogSpell(t, g, "Moldgraf Millipede", "Creature — Insect Horror", s30MoldgrafMillipedeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The top three (two creatures and a sorcery) are milled; with the
	// creature already there that is three creature cards.
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
}

const (
	s30SkaabRuinatorOracle   = "83734824-1cad-4649-87ae-a5c338771c4c"
	s30ThrillingEncoreOracle = "2a114ff8-fa7d-41b8-b412-6232fbb68efc"
	s30BygoneMarvelsOracle   = "494dc919-e429-491b-b19b-5037499e2dd6"
)

func TestCmdr2S30SkaabRuinatorIsCastFromTheGraveyardExilingThreeCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	var fuel []uuid.UUID
	for i := 0; i < 3; i++ {
		fuel = append(fuel, effectstest.PushGraveyardCardTyped(me, "Fuel", "Creature — Zombie"))
	}
	art := effectstest.PushGraveyardCardTyped(me, "Trinket", "Artifact")
	ruin := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: ruin, Name: "Skaab Ruinator", TypeLine: "Creature — Zombie Horror", OracleID: s30SkaabRuinatorOracle,
		ManaCost: "{1}{U}{U}", Power: 5, Toughness: 6, Owner: me.ID, Controller: me.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	// An artifact cannot pay for a creature-card cost.
	if err := g.CastSpell(me.ID, ruin, game.CastSpellParams{Strict: true, FromZone: "graveyard", ExileIDs: []uuid.UUID{fuel[0], fuel[1], art}}); err == nil {
		t.Fatal("cast with a noncreature card exiled was accepted")
	}
	// Two cards cannot pay for three.
	if err := g.CastSpell(me.ID, ruin, game.CastSpellParams{Strict: true, FromZone: "graveyard", ExileIDs: fuel[:2]}); err == nil {
		t.Fatal("cast with two cards exiled was accepted")
	}
	if err := g.CastSpell(me.ID, ruin, game.CastSpellParams{Strict: true, FromZone: "graveyard", ExileIDs: fuel}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	for _, id := range fuel {
		if !g.Exile.Contains(id) {
			t.Errorf("card %s was not exiled", id)
		}
	}
	if !me.Graveyard.Contains(art) {
		t.Error("the artifact card stays")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ruin) {
		t.Fatal("Skaab Ruinator did not resolve")
	}
	if !slices.Contains(effectstest.LayeredCard(t, g, ruin).Effective().Abilities, "flying") {
		t.Error("Skaab Ruinator has flying")
	}
}

func TestCmdr2S30SkaabRuinatorIsNotCastableFromHandWithoutTheCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushGraveyardCardTyped(me, "Fuel", "Creature — Zombie")
	_, err := effectstest.CastWithTapParams(t, g, "Skaab Ruinator", "Creature — Zombie Horror", "{1}{U}{U}", s30SkaabRuinatorOracle,
		game.CastSpellParams{Strict: true})
	if err == nil {
		t.Error("cast without exiling three creature cards was accepted")
	}
}

func TestCmdr2S30ThrillingEncoreTakesWhatDiedFromTheBattlefieldThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 3, 3)
	old := effectstest.PushGraveyardCardTyped(opp, "Old Bear", "Creature — Bear")
	discarded := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: discarded, Name: "Discarded", TypeLine: "Creature — Elf", Owner: opp.ID, Controller: opp.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, discarded); err != nil {
		t.Fatalf("discard: %v", err)
	}
	artifact := effectstest.PushCatalogPermanent(g, opp.ID, "Dead Trinket", "Artifact", "", false)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.KillCreature(t, g, opp.ID, artifact)
	effectstest.CastCatalogSpell(t, g, "Thrilling Encore", "Instant", s30ThrillingEncoreOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("creature %s that died this turn did not return", id)
		}
		if c := effectstest.LayeredCard(t, g, id); c.Controller != me.ID {
			t.Errorf("%s returned under %s's control, want the caster's", c.Name, c.Controller)
		}
	}
	if g.Battlefield.Contains(old) || g.Battlefield.Contains(discarded) || g.Battlefield.Contains(artifact) {
		t.Error("a card from an earlier turn, a discarded creature and a dead artifact stay put")
	}
}

func s30CopiesOnTheStack(g *game.Game) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, m := range g.StackMeta {
			if m != nil && m.IsCopy {
				n++
			}
		}
	})
	return n
}

func TestCmdr2S30BygoneMarvelsCopiesTwiceWithEightPermanentCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var targets []uuid.UUID
	for i := 0; i < 8; i++ {
		targets = append(targets, effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear"))
	}
	marvels := effectstest.CastCatalogSpell(t, g, "Bygone Marvels", "Sorcery", s30BygoneMarvelsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: targets[0]}})
	// The cast trigger resolves first and makes two copies, each asking for
	// its own new target (CR 707.10c).
	prompts := 0
	for i := 0; i < 80 && (g.Stack.Size() > 0 || len(g.PendingChoices) > 0); i++ {
		if ch := effectstest.LatestPickTarget(g, me.ID); ch != nil {
			prompts++
			if prompts > 2 {
				t.Fatal("more than two new-target prompts")
			}
			if err := g.ResolvePickTarget(ch.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: targets[prompts]}); err != nil {
				t.Fatalf("new target for copy %d: %v", prompts, err)
			}
			continue
		}
		if ch := effectstest.LatestConfirmFor(g, me.ID); ch != nil {
			if err := g.ResolveConfirm(ch.ID, me.ID, true); err != nil {
				t.Fatalf("confirm: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if prompts != 2 {
		t.Fatalf("%d new-target prompts, want one for each of the two copies", prompts)
	}
	for _, id := range targets[:3] {
		if !me.Hand.Contains(id) {
			t.Errorf("target %s was not returned to hand by the spell or a copy", id)
		}
	}
	if me.Graveyard.Contains(marvels) || !effectstest.InExile(g, marvels) {
		t.Error("the spell is exiled, and the copies leave nothing in the graveyard")
	}
	if me.Graveyard.Size() != 5 {
		t.Errorf("graveyard has %d cards, want the 5 untouched ones", me.Graveyard.Size())
	}
}

func TestCmdr2S30BygoneMarvelsSevenPermanentCardsMakesNoCopyAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var targets []uuid.UUID
	for i := 0; i < 7; i++ {
		targets = append(targets, effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear"))
	}
	hand := me.Hand.Size()
	marvels := effectstest.CastCatalogSpell(t, g, "Bygone Marvels", "Sorcery", s30BygoneMarvelsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: targets[0]}})
	if s30CopiesOnTheStack(g) != 0 {
		t.Fatal("seven permanent cards made a copy")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(targets[0]) || me.Hand.Size() != hand+1 {
		t.Errorf("the target is not back in hand (hand %d -> %d)", hand, me.Hand.Size())
	}
	if !effectstest.InExile(g, marvels) {
		t.Error("Bygone Marvels exiles itself")
	}
}

const (
	s30OldStickfingersOracle = "1d85eec2-80db-466c-bd48-d2a12866e8bd"
	s30GateOracle            = "5a682ec5-78e6-4238-b5c2-e94b7a8f8066"
	s30SpiritSistersOracle   = "83914a83-4de2-4d64-9770-f68a12014167"
)

func TestCmdr2S30OldStickfingersRevealsUntilXCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	// Top to bottom: land, creature, sorcery, creature, creature, land.
	var order []uuid.UUID
	for _, c := range []struct{ name, tl string }{
		{"Bottom Land", "Land"}, {"Deep Bear", "Creature — Bear"}, {"Third Bear", "Creature — Bear"},
		{"Spell", "Sorcery"}, {"Second Bear", "Creature — Bear"}, {"Top Land", "Land"},
	} {
		id := uuid.New()
		order = append(order, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: c.name, TypeLine: c.tl, Owner: me.ID, Controller: me.ID})
	}
	// order is bottom-first: the top two creatures are order[4] and order[2].
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Old Stickfingers", TypeLine: "Legendary Creature — Horror", OracleID: s30OldStickfingersOracle,
		ManaCost: "{X}{B}{G}", Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{G}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Old Stickfingers did not resolve")
	}
	if !me.Graveyard.Contains(order[4]) || !me.Graveyard.Contains(order[2]) {
		t.Error("the first two creature cards revealed go to the graveyard")
	}
	if me.Graveyard.Contains(order[3]) || me.Graveyard.Contains(order[5]) {
		t.Error("the revealed land and sorcery do not")
	}
	// The revealed noncreature cards are on the bottom of the library; the
	// unrevealed ones sit above them in order.
	lib := effectstest.LibraryOrder(me)
	if len(lib) != 4 {
		t.Fatalf("library has %d cards, want 4", len(lib))
	}
	bottomTwo := lib[:2]
	if !slices.Contains(bottomTwo, order[3]) || !slices.Contains(bottomTwo, order[5]) {
		t.Errorf("the revealed noncreature cards are not on the bottom: %v", bottomTwo)
	}
	// 2/2: two creature cards in the graveyard.
	if e := effectstest.LayeredCard(t, g, id).Effective(); e.Power != 2 || e.Toughness != 2 {
		t.Errorf("Old Stickfingers is %d/%d, want 2/2", e.Power, e.Toughness)
	}
}

func TestCmdr2S30OldStickfingersWithAShortLibraryRevealsItAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	bear := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: bear, Name: "Lone Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Old Stickfingers", TypeLine: "Legendary Creature — Horror", OracleID: s30OldStickfingersOracle,
		ManaCost: "{X}{B}{G}", Owner: me.ID, Controller: me.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{G}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bear) || me.Library.Size() != 1 || !me.Library.Contains(land) {
		t.Errorf("graveyard has bear %v, library %d cards (land back on the bottom %v)", me.Graveyard.Contains(bear), me.Library.Size(), me.Library.Contains(land))
	}
}

func TestCmdr2S30GateToTheAfterlifeGainsLifeForANontokenCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Gate to the Afterlife", "Artifact", s30GateOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("life %d -> %d, want +1", life, me.Life)
	}
	// The loot is asked and may be declined: nothing is drawn.
	hand := me.Hand.Size()
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d after declining the loot", hand, me.Hand.Size())
	}
	// A token does not trigger it.
	life = me.Life
	token := effectstest.PushVanillaCreature(g, me.ID, "Token Bear", 1, 1)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == token {
			g.Battlefield.Cards[i].TypeLine = "Token Creature — Bear"
		}
	}
	effectstest.KillCreature(t, g, me.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a token's death changed life %d -> %d", life, me.Life)
	}
}

func TestCmdr2S30GateToTheAfterlifeNeedsSixCreatureCardsAndFindsTheGift(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gate := effectstest.PushCatalogPermanent(g, me.ID, "Gate to the Afterlife", "Artifact", s30GateOracle, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	gift := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: gift, Name: "God-Pharaoh's Gift", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	for i := 0; i < 5; i++ {
		effectstest.PushGraveyardCardTyped(me, "Fuel", "Creature — Zombie")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, gate, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only five creature cards in the graveyard")
	}
	effectstest.PushGraveyardCardTyped(me, "Fuel", "Creature — Zombie")
	if err := g.ActivateCatalogAbility(me.ID, gate, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with six creature cards: %v", err)
	}
	if g.Battlefield.Contains(gate) {
		t.Error("the Gate is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary)
	effectstest.AnswerSearchByID(t, g, me.ID, gift)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(gift) {
		t.Error("God-Pharaoh's Gift did not enter the battlefield")
	}
}

func TestCmdr2S30SpiritSistersCallReturnsTheChosenCardForASacrificeAndExilesItAfter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	call := effectstest.PushCatalogPermanent(g, me.ID, "Spirit-Sister's Call", "Enchantment", s30SpiritSistersOracle, false)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PickTriggerTarget(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	// Only a permanent sharing a card type with the chosen creature card may be
	// sacrificed: the enchantment is no candidate.
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{call}); err == nil {
		t.Error("the Call itself (an enchantment) was accepted as a sacrifice for a creature card")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(dead) {
		t.Fatal("the creature was sacrificed and the chosen card returned")
	}
	effectstest.KillCreature(t, g, me.ID, dead)
	if !effectstest.InExile(g, dead) {
		t.Error("the returned card is exiled instead of going to the graveyard")
	}
}

func TestCmdr2S30SpiritSistersCallDecliningTheSacrificeReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Spirit-Sister's Call", "Enchantment", s30SpiritSistersOracle, false)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PickTriggerTarget(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) || g.Battlefield.Contains(dead) {
		t.Error("with no sacrifice nothing returns")
	}
}

func TestCmdr2S30SpiritSistersCallAnEnchantmentCardCanBePaidForByTheCallItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	call := effectstest.PushCatalogPermanent(g, me.ID, "Spirit-Sister's Call", "Enchantment", s30SpiritSistersOracle, false)
	effectstest.PushVanillaCreature(g, me.ID, "Only Creature", 1, 1)
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Aura", "Enchantment")
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PickTriggerTarget(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, call)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(call) || !g.Battlefield.Contains(dead) {
		t.Error("the Call may sacrifice itself to return an enchantment card")
	}
}
