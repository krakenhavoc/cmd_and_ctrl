package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// old_b815_cards_test.go — S82 OLD batch 815, Alliances activated abilities
// (#5634).

const (
	b815Dam       = "a3947473-0f57-4937-bb1c-f20969144415"
	b815Explorers = "bc0518fc-d34b-474a-92c5-2cb013828396"
	b815Engine    = "bf6ea878-044b-40a8-a21b-0191a948a69a"
	b815Defender  = "02197d14-8d9f-4b1f-9ee0-ce8ef690843c"
	b815Browse    = "49491601-e272-4603-9348-185a2953bd4b"
	b815Digger    = "b526962e-be8a-4803-b038-57869bb96751"
	b815Breaker   = "84fecd54-c08a-4f5a-b2fa-b254069ccadb"
	b815Scepter   = "120e9013-b07b-49d9-8b86-f6c26315581d"
	b815Adnate    = "e3174070-542f-429e-9149-b572f6614537"
	b815Crusader  = "45351d9d-453f-4245-84c0-053eb8aca345"
	b815Cylix     = "fdf4e4c3-fe17-4098-9f76-e98f3f00bd3e"
)

func TestB815CardsAreRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		b815Dam: "Floodwater Dam", b815Explorers: "Benthic Explorers", b815Engine: "Urza's Engine",
		b815Defender: "Sworn Defender", b815Browse: "Browse", b815Digger: "Soldevi Digger",
		b815Breaker: "Mishra's Groundbreaker", b815Scepter: "Gustha's Scepter", b815Adnate: "Soldevi Adnate",
		b815Crusader: "Varchild's Crusader", b815Cylix: "Ashnod's Cylix",
	} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func b815Pool(p *game.Player, color string) int {
	n := 0
	for _, m := range p.ManaPool {
		if m.Color == color {
			n += max(m.Count, 1)
		}
	}
	return n
}

func b815Tapped(g *game.Game, id uuid.UUID) bool {
	c := effectstest.FindBattlefieldCardForTest(g, id)
	return c != nil && c.Tapped
}

func TestB815FloodwaterDamTapsXTargetLandsForTwoXPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dam := effectstest.PushCatalogPermanent(g, me.ID, "Floodwater Dam", "Artifact", b815Dam, false)
	l1 := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	l2 := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	l3 := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.ApaMana(me, "G", "G", "G", "G", "G")
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: l1}, {Kind: game.TargetCard, ID: l2}}
	effectstest.P7Activate(t, g, me, dam, 0, game.ActivateAbilityParams{XValue: 2, Targets: targets})
	if !b815Tapped(g, l1) || !b815Tapped(g, l2) || b815Tapped(g, l3) {
		t.Errorf("tapped: %v %v (untouched %v)", b815Tapped(g, l1), b815Tapped(g, l2), b815Tapped(g, l3))
	}
	if !b815Tapped(g, dam) {
		t.Error("the Dam did not tap for its own cost")
	}
	if got := effectstest.PoolSize(g, me.ID); got != 0 {
		t.Errorf("%d mana left, want 0 (X=2 costs five)", got)
	}
}

func TestB815FloodwaterDamRefusesTooFewTargetsForX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dam := effectstest.PushCatalogPermanent(g, me.ID, "Floodwater Dam", "Artifact", b815Dam, false)
	l1 := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.ApaMana(me, "G", "G", "G", "G", "G")
	err := g.ActivateCatalogAbility(me.ID, dam, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: l1}}})
	if err == nil {
		t.Error("X = 2 with one target was accepted")
	}
}

func TestB815BenthicExplorersUntapsAnOpponentsTappedLandForItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	explorers := effectstest.B12Push(g, me.ID, "Benthic Explorers", "Creature — Merfolk Scout", b815Explorers, 2, 4)
	island := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	untapped := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	mine := effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, island).Tapped = true
		effectstest.FindBattlefieldCardForTest(g, mine).Tapped = true
	})
	for name, id := range map[string]uuid.UUID{"their untapped land": untapped, "my own tapped land": mine} {
		if err := g.ActivateManaAbility(me.ID, explorers, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{id}}); err == nil {
			t.Errorf("%s paid the cost", name)
		}
	}
	if err := g.ActivateManaAbility(me.ID, explorers, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{island}}); err != nil {
		t.Fatalf("untap their tapped Island: %v", err)
	}
	if b815Tapped(g, island) {
		t.Error("their Island was not untapped")
	}
	if b815Pool(me, "U") != 1 {
		t.Errorf("pool %v, want one {U}", me.ManaPool)
	}
	if !b815Tapped(g, explorers) {
		t.Error("the Explorers did not tap")
	}
}

func TestB815UrzasEngineGainsBandingAndGivesItsBandTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engine := effectstest.B12Push(g, me.ID, "Urza's Engine", "Artifact Creature — Juggernaut", b815Engine, 1, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "G", "G", "G", "G", "G", "G")
	effectstest.P7Activate(t, g, me, engine, 0, game.ActivateAbilityParams{})
	if c := effectstest.LayeredCard(t, g, engine); !game.HasKeyword(&c, game.KeywordBanding) {
		t.Fatal("the Engine did not gain banding")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	decls := []game.AttackDeclaration{
		{Attacker: engine, Target: opp.ID, Band: 1},
		{Attacker: bear, Target: opp.ID, Band: 1},
	}
	if _, err := g.DeclareAttackersWith(decls, game.DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare the band: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P7Activate(t, g, me, engine, 1, game.ActivateAbilityParams{})
	if c := effectstest.LayeredCard(t, g, bear); !game.HasKeyword(&c, "trample") {
		t.Error("the banded Bear did not gain trample")
	}
	if c := effectstest.LayeredCard(t, g, engine); !game.HasKeyword(&c, "trample") {
		t.Error("the Engine lost its printed trample")
	}
}

func TestB815SwornDefenderTradesPowerAndToughnessWithItsOpponentInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	defender := effectstest.B12Push(g, me.ID, "Sworn Defender", "Creature — Human Knight", b815Defender, 1, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 2, 5)
	effectstest.Pr7bAttack(t, g, opp.ID, defender)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: defender})
	effectstest.ApaMana(me, "W")
	effectstest.P7Activate(t, g, me, defender, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wall}}})
	if p, tt := effectstest.EffectivePTOf(t, g, defender); p != 4 || tt != 3 {
		t.Errorf("Sworn Defender is %d/%d, want 4/3 (toughness 5 minus 1, power 2 plus 1)", p, tt)
	}
}

func TestB815SwornDefenderCannotTargetACreatureNotInCombatWithIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	defender := effectstest.B12Push(g, me.ID, "Sworn Defender", "Creature — Human Knight", b815Defender, 1, 3)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "W")
	err := g.ActivateCatalogAbility(me.ID, defender, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bystander}}})
	if err == nil {
		t.Error("targeted a creature that is not blocking or blocked by the Defender")
	}
}

func TestB815BrowseTakesOneOfTheTopFiveAndExilesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AwbPadLibrary(me)
	var top []uuid.UUID
	for _, n := range []string{"A", "B", "C", "D", "E"} {
		top = append(top, effectstest.B02TopOfLibrary(me, "Look "+n, "Sorcery"))
	}
	browse := effectstest.PushCatalogPermanent(g, me.ID, "Browse", "Enchantment", b815Browse, false)
	effectstest.ApaMana(me, "U", "U", "G", "G")
	before := me.Hand.Size()
	effectstest.P7Activate(t, g, me, browse, 0, game.ActivateAbilityParams{})
	effectstest.AnswerChooseCards(t, g, me.ID, top[2])
	if got := me.Hand.Size(); got != before+1 || !me.Hand.Contains(top[2]) {
		t.Errorf("hand %d, want %d with the chosen card", got, before+1)
	}
	for i, id := range top {
		if i == 2 {
			continue
		}
		if !g.Exile.Contains(id) {
			t.Errorf("look card %d was not exiled", i)
		}
	}
}

func TestB815SoldeviDiggerPutsTheTopGraveyardCardOnTheBottomOfTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	digger := effectstest.PushCatalogPermanent(g, me.ID, "Soldevi Digger", "Artifact", b815Digger, false)
	older, newer := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: older, Name: "Older", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: newer, Name: "Newer", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	effectstest.ApaMana(me, "G", "G")
	effectstest.P7Activate(t, g, me, digger, 0, game.ActivateAbilityParams{})
	if b, _ := me.Library.Bottom(); b.InstanceID != newer {
		t.Errorf("the bottom of the library is %q, want the top graveyard card", b.Name)
	}
	if !me.Graveyard.Contains(older) || me.Graveyard.Contains(newer) {
		t.Error("the wrong card left the graveyard")
	}
}

func TestB815MishrasGroundbreakerMakesALandAThreeThreeArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	breaker := effectstest.PushCatalogPermanent(g, me.ID, "Mishra's Groundbreaker", "Artifact", b815Breaker, false)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.P7Activate(t, g, me, breaker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}})
	if g.Battlefield.Contains(breaker) {
		t.Error("the Groundbreaker was not sacrificed")
	}
	c := effectstest.LayeredCard(t, g, land)
	if !c.IsCreature() || !c.IsArtifact() || !c.IsLand() {
		t.Errorf("the land is %q, want an artifact creature that is still a land", c.TypeLine)
	}
	if p, tt := effectstest.EffectivePTOf(t, g, land); p != 3 || tt != 3 {
		t.Errorf("the land is %d/%d, want 3/3", p, tt)
	}
	// It lasts: through a turn's end.
	effectstest.EndTheTurn(t, g)
	if c := effectstest.LayeredCard(t, g, land); !c.IsCreature() {
		t.Error("the animation ended with the turn")
	}
}

func TestB815VarchildsCrusaderIsBlockableOnlyByWallsThenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := effectstest.B12Push(g, me.ID, "Varchild's Crusader", "Creature — Human Knight", b815Crusader, 3, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.P7Activate(t, g, me, crusader, 0, game.ActivateAbilityParams{})
	effectstest.Pr7bAttack(t, g, opp.ID, crusader)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, crusader); err == nil {
		t.Error("a non-Wall blocked the Crusader")
	}
	if err := g.DeclareBlocker(wall, crusader); err != nil {
		t.Errorf("a Wall could not block it: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(crusader) {
		t.Error("the Crusader was not sacrificed at the end step")
	}
}

func TestB815SoldeviAdnateAddsBlackEqualToTheSacrificedCreaturesManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	adnate := effectstest.B12Push(g, me.ID, "Soldevi Adnate", "Creature — Human Cleric", b815Adnate, 1, 2)
	black := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Black Four", TypeLine: "Creature — Zombie", ManaCost: "{3}{B}", Colors: []string{"B"}, Power: 2, Toughness: 2})
	green := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2})
	if err := g.ActivateManaAbility(me.ID, adnate, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{green}}); err == nil {
		t.Fatal("sacrificed a creature that is neither black nor an artifact")
	}
	if err := g.ActivateManaAbility(me.ID, adnate, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{black}}); err != nil {
		t.Fatalf("sacrifice the black creature: %v", err)
	}
	if b815Pool(me, "B") != 4 {
		t.Errorf("pool %v, want four {B}", me.ManaPool)
	}
	if g.Battlefield.Contains(black) {
		t.Error("the creature was not sacrificed")
	}
}

func TestB815SoldeviAdnateAcceptsAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	adnate := effectstest.B12Push(g, me.ID, "Soldevi Adnate", "Creature — Human Cleric", b815Adnate, 1, 2)
	golem := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{5}", Power: 3, Toughness: 3})
	if err := g.ActivateManaAbility(me.ID, adnate, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{golem}}); err != nil {
		t.Fatalf("sacrifice the artifact creature: %v", err)
	}
	if b815Pool(me, "B") != 5 {
		t.Errorf("pool %v, want five {B}", me.ManaPool)
	}
}

func TestB815AshnodsCylixKeepsOneOfTheTopThreeAndExilesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AwbPadLibrary(opp)
	cylix := effectstest.PushCatalogPermanent(g, me.ID, "Ashnod's Cylix", "Artifact", b815Cylix, false)
	var top []uuid.UUID
	for _, n := range []string{"A", "B", "C"} {
		top = append(top, effectstest.B02TopOfLibrary(opp, "Top "+n, "Sorcery"))
	}
	effectstest.ApaMana(me, "G", "G", "G")
	effectstest.P7Activate(t, g, me, cylix, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.AnswerChooseCards(t, g, opp.ID, top[0])
	if got, _ := opp.Library.Top(); got.InstanceID != top[0] {
		t.Errorf("the top of their library is %q, want the card they kept", got.Name)
	}
	if !g.Exile.Contains(top[1]) || !g.Exile.Contains(top[2]) || g.Exile.Contains(top[0]) {
		t.Error("the other two were not exiled")
	}
}

func TestB815GusthasScepterExilesReturnsAndLosesItsPile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scepter := effectstest.PushCatalogPermanent(g, me.ID, "Gustha's Scepter", "Artifact", b815Scepter, false)
	hidden := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: hidden, Name: "Hidden Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.P7Activate(t, g, me, scepter, 0, game.ActivateAbilityParams{})
	effectstest.AnswerChooseCards(t, g, me.ID, hidden)
	if me.Hand.Contains(hidden) || !g.Exile.Contains(hidden) {
		t.Fatal("the card was not exiled from the hand")
	}
	// Return it.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, scepter).Tapped = false })
	effectstest.P7Activate(t, g, me, scepter, 1, game.ActivateAbilityParams{})
	effectstest.AnswerChooseCards(t, g, me.ID, hidden)
	if !me.Hand.Contains(hidden) {
		t.Fatal("the card did not come back to the hand")
	}
	// Exile it again, then lose the Scepter: the pile goes to the graveyard.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, scepter).Tapped = false })
	effectstest.P7Activate(t, g, me, scepter, 0, game.ActivateAbilityParams{})
	effectstest.AnswerChooseCards(t, g, me.ID, hidden)
	effectstest.B14Kill(g, scepter)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(hidden) || g.Exile.Contains(hidden) {
		t.Error("the exiled card did not go to its owner's graveyard when the Scepter left")
	}
}
