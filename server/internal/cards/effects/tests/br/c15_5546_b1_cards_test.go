package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// c15_5546_b1_cards_test.go — Commander 2015, batch 1 (#5546). Every test
// plays the real catalog entry through the engine.

const (
	c15CallerOfThePackOracle    = "3b6cbd2a-02ba-4afb-a081-0f25d830a8ee"
	c15HeraldOfTheHostOracle    = "4371d5bc-1760-477b-822e-6d4999fb0a6b"
	c15WarchiefGiantOracle      = "589abb40-04b4-4dbe-9c2a-446b6c3075d1"
	c15BroodbirthViperOracle    = "a8455828-9c61-4ef9-a5f1-1f0cc6fbf446"
	c15BansheeOracle            = "282b4dbf-ec00-48ba-8010-73bee489fa7a"
	c15KalemneOracle            = "e100f9fc-b6fb-42ab-b9b3-796bede6804d"
	c15MizzixOracle             = "f787c6cf-a4ff-487b-a7b0-0f516e68510a"
	c15DaxosOracle              = "8a9e8f60-eb13-4ae8-82df-d4c46da74451"
	c15KalemnesCaptainOracle    = "663e6ab0-ed99-4cec-86f7-843207c31e17"
	c15KarlovOracle             = "647d7988-9a5d-4700-9238-938d0d5fae12"
	c15KasetoOracle             = "198b6c6f-a480-4672-9d87-b0cb5d8ed1c0"
	c15GreatOakGuardianOracle   = "5dba7746-c087-4ec1-92ca-ab9882dea24d"
	c15RighteousConfluenceOracl = "ee28937b-7ead-45ad-8448-35c51a09fb32"
	c15VerdantConfluenceOracle  = "0f786c99-d245-4b11-bcce-6e13fb1255cc"
	c15WretchedConfluenceOracle = "fb430640-81b6-4720-84d6-b8ec3af6a2c3"
	c15SandstoneOracleOracle    = "850e2dbd-b904-41b9-8ffe-0aa562837bde"
	c15OreskosExplorerOracle    = "a3055ae8-990d-495a-9cb5-971407990cfa"
	c15MeteorBlastOracle        = "0ba1837b-67a1-40d1-b617-8c94834d02ba"
	c15AwakenTheSkyTyrantOracle = "0607b289-74a6-4f3a-b15c-a639244873b6"
	c15DaxossTormentOracle      = "93ef3098-d1c0-4b05-9d28-1afa133bfc4f"
	c15BloodsporeThrinaxOracle  = "1f4743ed-4e91-4344-9409-c8ede5c4804d"

	c15BearType = "Creature — Bear"
)

// c15B1Cast puts a non-catalog card with a mana cost in p's hand, casts it
// from a main phase and settles the stack (the spell and any trigger it
// causes).
func c15B1Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, manaCost string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// c15B1GiveExperience gives p n experience counters.
func c15B1GiveExperience(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.AddPlayerCounterByForEffect(p.ID, p.ID, game.CounterExperience, n) })
}

// c15B1GraveyardCard puts a card of the given type line into p's graveyard.
func c15B1GraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// c15B1Tapped seeds a tapped vanilla creature under owner.
func c15B1Tapped(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: c15BearType,
		Power: 2, Toughness: 2, Tapped: true, Owner: owner, Controller: owner,
	})
}

func c15B1Power(g *game.Game, id uuid.UUID) int { return effectstest.ApaLive(g, id).CurrentPower() }

// --- every card is in the catalog ----------------------------------------

func TestC15B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		c15CallerOfThePackOracle, c15HeraldOfTheHostOracle, c15WarchiefGiantOracle, c15BroodbirthViperOracle,
		c15BansheeOracle, c15KalemneOracle, c15MizzixOracle, c15DaxosOracle, c15KalemnesCaptainOracle,
		c15KarlovOracle, c15KasetoOracle, c15GreatOakGuardianOracle, c15RighteousConfluenceOracl,
		c15VerdantConfluenceOracle, c15WretchedConfluenceOracle, c15SandstoneOracleOracle,
		c15OreskosExplorerOracle, c15MeteorBlastOracle, c15AwakenTheSkyTyrantOracle,
		c15DaxossTormentOracle, c15BloodsporeThrinaxOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestC15B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Caller of the Pack", "Creature — Beast", c15CallerOfThePackOracle, []string{"trample"}},
		{"Herald of the Host", "Creature — Angel", c15HeraldOfTheHostOracle, []string{"flying", "vigilance"}},
		{"Warchief Giant", "Creature — Giant Warrior", c15WarchiefGiantOracle, []string{"haste"}},
		{"Kalemne, Disciple of Iroas", "Legendary Creature — Giant Soldier", c15KalemneOracle, []string{"double strike", "vigilance"}},
		{"Kalemne's Captain", "Creature — Giant Soldier", c15KalemnesCaptainOracle, []string{"vigilance"}},
		{"Great Oak Guardian", "Creature — Treefolk", c15GreatOakGuardianOracle, []string{"flash", "reach"}},
		{"Sandstone Oracle", "Artifact Creature — Sphinx", c15SandstoneOracleOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 4, 4)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- myriad ---------------------------------------------------------------------

// Attacking one opponent at a four-player table offers a copy attacking
// each of the other two.
func TestC15B1MyriadCopiesAttackTheOtherOpponents(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Caller of the Pack", "Creature — Beast", c15CallerOfThePackOracle},
		{"Herald of the Host", "Creature — Angel", c15HeraldOfTheHostOracle},
		{"Warchief Giant", "Creature — Giant Warrior", c15WarchiefGiantOracle},
		{"Broodbirth Viper", "Creature — Snake", c15BroodbirthViperOracle},
		{"Banshee of the Dread Choir", "Creature — Spirit", c15BansheeOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			attacker := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 4, 4)
			effectstest.DeclareAttack(t, g, opp.ID, attacker)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerMayChoice(t, g, me.ID, true)
			effectstest.AnswerMayChoice(t, g, me.ID, true)
			copies := 0
			for _, c := range g.Battlefield.Cards {
				if c.Name == tc.name && c.InstanceID != attacker && effects.IsToken(c) {
					copies++
				}
			}
			if copies != 2 {
				t.Errorf("%d myriad copies, want 2", copies)
			}
		})
	}
}

func TestC15B1BroodbirthViperMayDrawOnCombatDamage(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		viper := effectstest.B12Push(g, me.ID, "Broodbirth Viper", "Creature — Snake", c15BroodbirthViperOracle, 3, 3)
		hand := me.Hand.Size()
		effectstest.DealCombatDamageToPlayer(g, viper, opp.ID, 3)
		effectstest.AssistCardSettle(t, g, me.ID, nil, yes)
		want := hand
		if yes {
			want++
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("answered %v: hand %d, want %d", yes, got, want)
		}
	}
}

func TestC15B1BansheeMakesTheDamagedPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	banshee := effectstest.B12Push(g, me.ID, "Banshee of the Dread Choir", "Creature — Spirit", c15BansheeOracle, 4, 4)
	effectstest.D2dSetHand(opp, 3)
	effectstest.D2dSetHand(other, 3)
	effectstest.DealCombatDamageToPlayer(g, banshee, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1", n)
	}
	if n := effectstest.DiscardOwed(g, other.ID); n != 0 {
		t.Errorf("a player not dealt damage owes %d discards", n)
	}
}

// --- experience counters -----------------------------------------------------

func TestC15B1KalemneCountsBigCreatureSpellsAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kalemne := effectstest.B12Push(g, me.ID, "Kalemne, Disciple of Iroas", "Legendary Creature — Giant Soldier", c15KalemneOracle, 3, 3)
	c15B1Cast(t, g, me, "Small Ogre", "Creature — Ogre", "{4}", 4, 4)
	if got := effects.ExperienceCounters(g, me.ID); got != 0 {
		t.Fatalf("a mana value 4 creature gave %d experience counters", got)
	}
	c15B1Cast(t, g, me, "Sorcery Five", "Sorcery", "{5}", 0, 0)
	if got := effects.ExperienceCounters(g, me.ID); got != 0 {
		t.Fatalf("a mana value 5 sorcery gave %d experience counters", got)
	}
	c15B1Cast(t, g, me, "Big Ogre", "Creature — Ogre", "{5}", 5, 5)
	if got := effects.ExperienceCounters(g, me.ID); got != 1 {
		t.Fatalf("experience counters %d, want 1", got)
	}
	c15B1GiveExperience(g, me, 1)
	if got := c15B1Power(g, kalemne); got != 5 {
		t.Errorf("Kalemne's power with two experience counters is %d, want 5", got)
	}
}

func TestC15B1MizzixGetsExperienceAndDiscounts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Mizzix of the Izmagnus", "Legendary Creature — Goblin Wizard", c15MizzixOracle, 2, 2)
	c15B1Cast(t, g, me, "Opt", "Instant", "{U}", 0, 0)
	if got := effects.ExperienceCounters(g, me.ID); got != 1 {
		t.Fatalf("a mana value 1 instant with no experience: %d counters, want 1", got)
	}
	c15B1Cast(t, g, me, "Ponder", "Sorcery", "{U}", 0, 0)
	if got := effects.ExperienceCounters(g, me.ID); got != 1 {
		t.Fatalf("a mana value 1 sorcery with one experience: %d counters, want still 1", got)
	}
	c15B1Cast(t, g, me, "Divination", "Sorcery", "{2}{U}", 0, 0)
	if got := effects.ExperienceCounters(g, me.ID); got != 2 {
		t.Fatalf("a mana value 3 sorcery with one experience: %d counters, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Divination", "Sorcery", "{2}{U}"); got != 1 {
		t.Errorf("two experience counters, sorcery: %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Golem", "Artifact Creature — Golem", "{4}"); got != 4 {
		t.Errorf("creature spell: %d, want 4 (undiscounted)", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Divination", "Sorcery", "{2}{U}"); got != 3 {
		t.Errorf("opponent's sorcery: %d, want 3 (undiscounted)", got)
	}
}

func TestC15B1DaxosGetsExperienceAndMakesAGrowingSpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	daxos := effectstest.B12Push(g, me.ID, "Daxos the Returned", "Legendary Creature — Zombie Soldier", c15DaxosOracle, 2, 2)
	c15B1Cast(t, g, me, "Bear", c15BearType, "{1}{G}", 2, 2)
	if got := effects.ExperienceCounters(g, me.ID); got != 0 {
		t.Fatalf("a creature spell gave %d experience counters", got)
	}
	c15B1Cast(t, g, me, "Test Enchantment", "Enchantment", "{1}{W}", 0, 0)
	if got := effects.ExperienceCounters(g, me.ID); got != 1 {
		t.Fatalf("experience counters %d, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, daxos, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	spirits := effectstest.TokensNamed(g, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirit tokens, want 1", len(spirits))
	}
	spirit := spirits[0]
	if !spirit.IsEnchantment() || !spirit.IsCreature() {
		t.Errorf("the Spirit is %q, want an enchantment creature", spirit.TypeLine)
	}
	if got := c15B1Power(g, spirit.InstanceID); got != 1 {
		t.Errorf("the Spirit's power with one experience counter is %d, want 1", got)
	}
	c15B1GiveExperience(g, me, 2)
	if got := c15B1Power(g, spirit.InstanceID); got != 3 {
		t.Errorf("the Spirit's power with three experience counters is %d, want 3", got)
	}
}

// --- monstrosity, counters, activated abilities -------------------------------

func TestC15B1KalemnesCaptainExilesAllArtifactsAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	captain := effectstest.PushMonster(g, me.ID, "Kalemne's Captain", c15KalemnesCaptainOracle, 5, 5)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	aura := effectstest.B12Permanent(g, me.ID, "Anthem", "Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", c15BearType, 2, 2)
	effectstest.ActivateMonstrosity(t, g, me.ID, captain, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.MonstrousNow(g, captain) || effectstest.CounterCount(g, captain, game.CounterPlusOne) != 3 {
		t.Fatalf("not monstrous with three counters")
	}
	if !g.Exile.Contains(rock) || !g.Exile.Contains(aura) {
		t.Error("an artifact or enchantment was not exiled")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was exiled")
	}
}

func TestC15B1KarlovGrowsOnLifeGainAndExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	karlov := effectstest.B12Push(g, me.ID, "Karlov of the Ghost Council", "Legendary Creature — Spirit Advisor", c15KarlovOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", c15BearType, 2, 2)
	effectstest.JmpB3GainLife(g, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, karlov, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d counters after one life gain, want 2", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, karlov, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("activated with only two +1/+1 counters")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(karlov, game.CounterPlusOne, 4) })
	if err := g.ActivateCatalogAbility(me.ID, karlov, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate with six counters: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("the target was not exiled")
	}
	if got := effectstest.CounterCount(g, karlov, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters left, want 0", got)
	}
}

func TestC15B1KasetoMakesUnblockableAndPumpsOnlySnakes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kaseto := effectstest.B12Push(g, me.ID, "Kaseto, Orochi Archmage", "Legendary Creature — Snake Wizard", c15KasetoOracle, 2, 2)
	snake := effectstest.B12Creature(g, me.ID, "Snake", "Creature — Snake", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", c15BearType, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	for _, id := range []uuid.UUID{snake, bear} {
		if err := g.ActivateCatalogAbility(me.ID, kaseto, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(id)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := c15B1Power(g, snake); got != 3 {
		t.Errorf("the Snake's power is %d, want 3", got)
	}
	if got := c15B1Power(g, bear); got != 2 {
		t.Errorf("the Bear's power is %d, want 2 (not a Snake)", got)
	}
	effectstest.BrAttack(t, g, bear)
	if err := g.DeclareBlocker(blocker, bear); err == nil {
		t.Error("the target was blocked")
	}
}

// --- enters triggers -----------------------------------------------------------

func TestC15B1GreatOakGuardianPumpsAndUntapsTargetPlayersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := c15B1Tapped(g, me.ID, "Mine")
	theirs := c15B1Tapped(g, opp.ID, "Theirs")
	effectstest.CastHoldCreature(t, g, "Great Oak Guardian", "Creature — Treefolk", c15GreatOakGuardianOracle, 4, 5)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}, true)
	c := effectstest.ApaLive(g, mine)
	if c.Tapped || c.CurrentPower() != 4 {
		t.Errorf("your creature: tapped %v, power %d, want untapped 4", c.Tapped, c.CurrentPower())
	}
	o := effectstest.ApaLive(g, theirs)
	if !o.Tapped || o.CurrentPower() != 2 {
		t.Errorf("another player's creature: tapped %v, power %d, want tapped 2", o.Tapped, o.CurrentPower())
	}
}

func TestC15B1SandstoneOracleDrawsTheDifference(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.CastHoldCreature(t, g, "Sandstone Oracle", "Artifact Creature — Sphinx", c15SandstoneOracleOracle, 4, 4)
	effectstest.D2dSetHand(me, 2)
	effectstest.D2dSetHand(opp, 5)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 5 {
		t.Errorf("hand %d, want 5 (drew the difference of 3)", got)
	}
}

func TestC15B1OreskosExplorerFindsAPlainsPerPlayerWithMoreLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	for i, s := range g.Seats {
		if s.ID == me.ID {
			continue
		}
		// Two opponents control more lands than you; the third controls none.
		if i%4 != (g.Turn.ActiveSeat+3)%4 {
			effectstest.B12Permanent(g, s.ID, "Forest", "Basic Land — Forest")
		}
	}
	lib := effectstest.B23Library(me,
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	effectstest.CastHoldCreature(t, g, "Oreskos Explorer", "Creature — Cat Scout", c15OreskosExplorerOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	c = effectstest.SearchChoiceFor(g, me.ID)
	if c.SearchMax != 2 {
		t.Fatalf("search max = %d, want 2", c.SearchMax)
	}
	for _, id := range c.SearchCards {
		if id == lib[3] {
			t.Error("an Island was offered")
		}
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, lib[:2]); err != nil {
		t.Fatalf("take two Plains: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range lib[:2] {
		if !me.Hand.Contains(id) {
			t.Error("a found Plains is not in your hand")
		}
	}
}

// --- spells ----------------------------------------------------------------------

func TestC15B1MeteorBlastDealsFourToEachOfXTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	life := opp.Life
	effectstest.CastXSpell(t, g, "Meteor Blast", "Sorcery", c15MeteorBlastOracle, "{X}{R}{R}{R}", 2,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: ogre}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d, want %d", opp.Life, life-4)
	}
	if g.Battlefield.Contains(ogre) {
		t.Error("the 4/4 survived 4 damage")
	}
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Meteor Blast", TypeLine: "Sorcery",
		OracleID: c15MeteorBlastOracle, ManaCost: "{X}{R}{R}{R}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID},
	}}); err == nil {
		t.Error("one target was accepted on X=2")
	}
}

func TestC15B1RighteousConfluenceKnightExileAndLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ench := effectstest.B12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	life := me.Life
	effectstest.CastModal(t, g, "Righteous Confluence", "Sorcery", c15RighteousConfluenceOracl, []int{0, 1, 2}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, ench, 1, 0),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.TokensNamed(g, "Knight")); n != 1 {
		t.Errorf("%d Knights, want 1", n)
	}
	if !g.Exile.Contains(ench) {
		t.Error("the enchantment was not exiled")
	}
	if me.Life != life+5 {
		t.Errorf("life %d, want %d", me.Life, life+5)
	}
}

func TestC15B1RighteousConfluenceRepeatsAMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastModal(t, g, "Righteous Confluence", "Sorcery", c15RighteousConfluenceOracl, []int{0, 0, 2}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.TokensNamed(g, "Knight")); n != 2 {
		t.Errorf("%d Knights, want 2", n)
	}
	if me.Life != life+5 {
		t.Errorf("life %d, want %d", me.Life, life+5)
	}
}

func TestC15B1VerdantConfluenceCountersAndReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "A", c15BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", c15BearType, 2, 2)
	rock := c15B1GraveyardCard(me, "Rock", "Artifact")
	effectstest.CastModal(t, g, "Verdant Confluence", "Sorcery", c15VerdantConfluenceOracle, []int{0, 0, 1}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, a, 0, 0),
		effectstest.ModeRef(game.TargetCard, b, 1, 0),
		effectstest.ModeRef(game.TargetCard, rock, 2, 0),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, a, game.CounterPlusOne) != 2 || effectstest.CounterCount(g, b, game.CounterPlusOne) != 2 {
		t.Errorf("counters %d and %d, want 2 each", effectstest.CounterCount(g, a, game.CounterPlusOne), effectstest.CounterCount(g, b, game.CounterPlusOne))
	}
	if !me.Hand.Contains(rock) {
		t.Error("the permanent card did not return to hand")
	}
}

func TestC15B1VerdantConfluenceSearchesABasicTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lib := effectstest.B23Library(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	bear := effectstest.B12Creature(g, me.ID, "Bear", c15BearType, 2, 2)
	rock := c15B1GraveyardCard(me, "Rock", "Artifact")
	effectstest.CastModal(t, g, "Verdant Confluence", "Sorcery", c15VerdantConfluenceOracle, []int{0, 1, 2}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, bear, 0, 0),
		effectstest.ModeRef(game.TargetCard, rock, 1, 0),
	})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	c = effectstest.SearchChoiceFor(g, me.ID)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, lib[:1]); err != nil {
		t.Fatalf("take the Forest: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	land := effectstest.BattlefieldCardFor(g, lib[0])
	if land == nil || !land.Tapped {
		t.Error("the basic is not on the battlefield tapped")
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
	if !me.Hand.Contains(rock) {
		t.Error("the permanent card did not return to hand")
	}
}

func TestC15B1WretchedConfluenceDrawShrinkAndReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", c15BearType, 2, 2)
	dead := c15B1GraveyardCard(me, "Dead Bear", c15BearType)
	hand, life := opp.Hand.Size(), opp.Life
	effectstest.CastModal(t, g, "Wretched Confluence", "Instant", c15WretchedConfluenceOracle, []int{0, 1, 2}, []game.TargetRef{
		effectstest.ModeRef(game.TargetPlayer, opp.ID, 0, 0),
		effectstest.ModeRef(game.TargetCard, bear, 1, 0),
		effectstest.ModeRef(game.TargetCard, dead, 2, 0),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+1 || opp.Life != life-1 {
		t.Errorf("target player: hand %d → %d, life %d → %d, want +1 card and -1 life", hand, opp.Hand.Size(), life, opp.Life)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived -2/-2")
	}
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
}

// --- enchantments --------------------------------------------------------------

func TestC15B1AwakenTheSkyTyrantBecomesADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tyrant := effectstest.PushCatalogPermanent(g, me.ID, "Awaken the Sky Tyrant", "Enchantment", c15AwakenTheSkyTyrantOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", c15BearType, 2, 2)
	effectstest.DealNoncombatDamage(g, me.ID, mine, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tyrant) {
		t.Fatal("damage from your own source sacrificed it")
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", c15BearType, 2, 2)
	effectstest.DealNoncombatDamage(g, opp.ID, theirs, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tyrant) {
		t.Error("the enchantment was not sacrificed")
	}
	dragons := effectstest.TokensNamed(g, "Dragon")
	if len(dragons) != 1 || dragons[0].Controller != me.ID {
		t.Fatalf("%d Dragons, want 1 under your control", len(dragons))
	}
	if got := c15B1Power(g, dragons[0].InstanceID); got != 5 {
		t.Errorf("the Dragon's power is %d, want 5", got)
	}
}

func TestC15B1DaxossTormentAnimatesOnConstellation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	torment := effectstest.PushCatalogPermanent(g, me.ID, "Daxos's Torment", "Enchantment", c15DaxossTormentOracle, false)
	if effectstest.ApaLive(g, torment).IsCreature() {
		t.Fatal("a creature before any enchantment entered")
	}
	c15B1Cast(t, g, me, "Test Enchantment", "Enchantment", "{1}", 0, 0)
	c := effectstest.ApaLive(g, torment)
	if !c.IsCreature() || !c.HasSubtype("Demon") || c.CurrentPower() != 5 {
		t.Errorf("after an enchantment entered: creature %v, Demon %v, power %d", c.IsCreature(), c.HasSubtype("Demon"), c.CurrentPower())
	}
	if !effectstest.HasKeywordOf(g, torment, "flying") || !effectstest.HasKeywordOf(g, torment, "haste") {
		t.Error("the Demon lacks flying or haste")
	}
}

// --- Bloodspore Thrinax -----------------------------------------------------------

func TestC15B1BloodsporeThrinaxSharesItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bloodspore Thrinax", TypeLine: "Creature — Lizard", OracleID: c15BloodsporeThrinaxOracle,
		Power: 2, Toughness: 2, Counters: map[string]int{game.CounterPlusOne: 2}, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.CastHoldCreature(t, g, "Bear", c15BearType, "", 2, 2)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the entering creature has %d counters, want 2", got)
	}
}
