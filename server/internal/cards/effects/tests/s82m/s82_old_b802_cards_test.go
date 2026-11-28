package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b802_cards_test.go — Mirage's removal, played from the real
// catalog entries (S82 OLD batch 802).

const (
	b802MinerOracle     = "0935377b-384f-4e4c-9fbb-8d2a7d5dc280"
	b802TinkererOracle  = "0afc2d2f-4fdb-440f-9c35-45bf103c0437"
	b802CorpsesOracle   = "16a7895a-ca23-4112-a91f-4b1528e75cea"
	b802SandsOracle     = "3bcb7436-f7d9-4bb0-81be-7a3db9d8329f"
	b802ChaosOracle     = "436b5989-9b8c-4e03-a82e-046837c891ca"
	b802FaerieOracle    = "447f1cce-bbc5-473b-a838-34070f8fd2dc"
	b802RashidaOracle   = "4aff0f09-0607-41b7-8f60-4ef35ff6183e"
	b802TerrorOracle    = "67d12a5a-9e0c-4550-86ad-cd1e173fea8f"
	b802CreatureLine    = "Creature — Test"
	b802DragonLine      = "Creature — Dragon"
	b802NonbasicLand    = "Land"
	b802BasicSwamp      = "Basic Land — Swamp"
	b802BasicMountain   = "Basic Land — Mountain"
	b802BasicPlains     = "Basic Land — Plains"
	b802BasicIsland     = "Basic Land — Island"
	b802ColoredBearLine = "Creature — Bear"
)

func TestS82OldB802CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b802MinerOracle:    "Dwarven Miner",
		b802TinkererOracle: "Goblin Tinkerer",
		b802CorpsesOracle:  "Wall of Corpses",
		b802SandsOracle:    "Choking Sands",
		b802ChaosOracle:    "Reign of Chaos",
		b802FaerieOracle:   "Uktabi Faerie",
		b802RashidaOracle:  "Rashida Scalebane",
		b802TerrorOracle:   "Reign of Terror",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// b802Pay adds one mana of each colour word to the pool.
func b802Pay(p *game.Player, colours ...string) {
	for _, c := range colours {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func b802Activate(g *game.Game, p *game.Player, source uuid.UUID, target uuid.UUID) error {
	var targets []game.TargetRef
	if target != uuid.Nil {
		targets = effectstest.CardTarget(target)
	}
	return g.ActivateCatalogAbility(p.ID, source, 0, game.ActivateAbilityParams{Strict: true, Targets: targets})
}

func TestDwarvenMinerDestroysOnlyANonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	miner := effectstest.PushCatalogPermanent(g, me.ID, "Dwarven Miner", "Creature — Dwarf", b802MinerOracle, false)
	basic := effectstest.B12Permanent(g, opp.ID, "Mountain", b802BasicMountain)
	odd := effectstest.B12Permanent(g, opp.ID, "Strange Land", b802NonbasicLand)

	b802Pay(me, "R", "R", "R")
	if err := b802Activate(g, me, miner, basic); err == nil {
		t.Fatal("targeted a basic land: accepted")
	}
	if err := b802Activate(g, me, miner, odd); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(odd) {
		t.Error("the nonbasic land survived")
	}
	if !g.Battlefield.Contains(basic) {
		t.Error("the basic land was destroyed")
	}
	if !effectstest.TappedOnBattlefield(t, g, miner) {
		t.Error("the Miner did not tap for its cost")
	}
}

func TestGoblinTinkererDestroysAnArtifactAndTakesItsManaValueInDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tinkerer := effectstest.B12Push(g, me.ID, "Goblin Tinkerer", "Creature — Goblin Artificer", b802TinkererOracle, 1, 5)
	rock := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}"})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b802CreatureLine, 2, 2)

	b802Pay(me, "R")
	if err := b802Activate(g, me, tinkerer, bear); err == nil {
		t.Fatal("targeted a nonartifact: accepted")
	}
	if err := b802Activate(g, me, tinkerer, rock); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if got := effectstest.DamageOn(t, g, tinkerer); got != 2 {
		t.Errorf("the Tinkerer has %d damage, want its artifact's mana value 2", got)
	}
}

func TestWallOfCorpsesDestroysTheCreatureItIsBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Ogre", b802CreatureLine, 3, 3)
	other := effectstest.B12Creature(g, me.ID, "Other", b802CreatureLine, 3, 3)
	wall := effectstest.PushCatalogPermanent(g, opp.ID, "Wall of Corpses", "Creature — Wall", b802CorpsesOracle, false)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{attacker, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)

	b802Pay(opp, "B")
	if err := b802Activate(g, opp, wall, other); err == nil {
		t.Fatal("targeted a creature it is not blocking: accepted")
	}
	if err := b802Activate(g, opp, wall, attacker); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the blocked creature survived")
	}
	if g.Battlefield.Contains(wall) {
		t.Error("the Wall was not sacrificed")
	}
}

func TestChokingSandsHitsTheControllerOfANonbasicLandOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	swamp := effectstest.B12Permanent(g, opp.ID, "Swamp", b802BasicSwamp)
	basic := effectstest.B12Permanent(g, opp.ID, "Plains", b802BasicPlains)
	odd := effectstest.B12Permanent(g, opp.ID, "Strange Land", b802NonbasicLand)

	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Choking Sands", "Sorcery", b802SandsOracle, effectstest.CardTarget(swamp)); err == nil {
		t.Fatal("targeted a Swamp: accepted")
	}
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Choking Sands", "Sorcery", b802SandsOracle, effectstest.CardTarget(basic))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(basic) || opp.Life != life {
		t.Errorf("basic Plains: on battlefield %v, life %d -> %d, want destroyed and no damage", g.Battlefield.Contains(basic), life, opp.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Choking Sands", "Sorcery", b802SandsOracle, effectstest.CardTarget(odd))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(odd) || opp.Life != life-2 {
		t.Errorf("nonbasic land: on battlefield %v, life %d -> %d, want destroyed and 2 damage", g.Battlefield.Contains(odd), life, opp.Life)
	}
}

func TestReignOfChaosBothModes(t *testing.T) {
	for _, tc := range []struct {
		mode             int
		landLine, colour string
	}{{0, b802BasicPlains, "W"}, {1, b802BasicIsland, "U"}} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		land := effectstest.B12Permanent(g, opp.ID, "Land", tc.landLine)
		mate := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Mate", TypeLine: b802ColoredBearLine, Colors: []string{tc.colour}, Power: 2, Toughness: 2})
		other := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Other", TypeLine: b802ColoredBearLine, Colors: []string{"G"}, Power: 2, Toughness: 2})
		id := uuid.New()
		active := g.Seats[g.Turn.ActiveSeat]
		active.Hand.PushTop(game.Card{InstanceID: id, Name: "Reign of Chaos", TypeLine: "Sorcery", OracleID: b802ChaosOracle, Owner: active.ID, Controller: active.ID})
		effectstest.AdvanceToMain(t, g)
		err := g.CastSpell(active.ID, id, game.CastSpellParams{Strict: true, Modes: []int{tc.mode}, Targets: b802Pair(tc.mode, land, other)})
		if err == nil {
			t.Fatalf("mode %d: a green creature accepted as the coloured target", tc.mode)
		}
		if err := g.CastSpell(active.ID, id, game.CastSpellParams{Strict: true, Modes: []int{tc.mode}, Targets: b802Pair(tc.mode, land, mate)}); err != nil {
			t.Fatalf("mode %d: cast: %v", tc.mode, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(land) || g.Battlefield.Contains(mate) {
			t.Errorf("mode %d: land on battlefield %v, creature %v, want both destroyed", tc.mode, g.Battlefield.Contains(land), g.Battlefield.Contains(mate))
		}
		if !g.Battlefield.Contains(other) {
			t.Errorf("mode %d: the green creature was destroyed", tc.mode)
		}
	}
}

func TestUktabiFaerieSacrificesToDestroyAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	faerie := effectstest.PushCatalogPermanent(g, me.ID, "Uktabi Faerie", "Creature — Faerie", b802FaerieOracle, false)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b802CreatureLine, 2, 2)
	b802Pay(me, "G", "G", "G", "G")
	if err := b802Activate(g, me, faerie, bear); err == nil {
		t.Fatal("targeted a nonartifact: accepted")
	}
	if err := b802Activate(g, me, faerie, rock); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(faerie) {
		t.Errorf("artifact on battlefield %v, faerie %v, want both gone", g.Battlefield.Contains(rock), g.Battlefield.Contains(faerie))
	}
}

func TestRashidaScalebaneDestroysAnAttackingDragonAndGainsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rashida := effectstest.PushCatalogPermanent(g, opp.ID, "Rashida Scalebane", "Legendary Creature — Human Soldier", b802RashidaOracle, false)
	dragon := effectstest.B12Creature(g, me.ID, "Drake", b802DragonLine, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", b802CreatureLine, 2, 2)
	if err := b802Activate(g, opp, rashida, dragon); err == nil {
		t.Fatal("targeted a Dragon that is not attacking: accepted")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{dragon, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	if err := b802Activate(g, opp, rashida, bear); err == nil {
		t.Fatal("targeted an attacking non-Dragon: accepted")
	}
	life := opp.Life
	if err := b802Activate(g, opp, rashida, dragon); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dragon) {
		t.Error("the Dragon survived")
	}
	if opp.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, opp.Life)
	}
}

func TestReignOfTerrorDestroysTheChosenColourAndCostsTwoLifeEach(t *testing.T) {
	for _, tc := range []struct {
		pick   int
		colour string
	}{{0, "G"}, {1, "W"}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		var dead, alive []uuid.UUID
		for i := 0; i < 2; i++ {
			dead = append(dead, effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: b802ColoredBearLine, Colors: []string{tc.colour}, Power: 2, Toughness: 2}))
		}
		dead = append(dead, effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: b802ColoredBearLine, Colors: []string{tc.colour}, Power: 2, Toughness: 2}))
		other := "W"
		if tc.colour == "W" {
			other = "G"
		}
		alive = append(alive, effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bystander", TypeLine: b802ColoredBearLine, Colors: []string{other}, Power: 2, Toughness: 2}))
		life := me.Life
		effectstest.CastCatalogSpell(t, g, "Reign of Terror", "Sorcery", b802TerrorOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick)
		if ask == nil {
			t.Fatal("no colour choice was asked")
		}
		if err := g.ResolveOptionPick(ask.ID, me.ID, tc.pick); err != nil {
			t.Fatalf("ResolveOptionPick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		for _, id := range dead {
			if g.Battlefield.Contains(id) {
				t.Errorf("pick %d: a %s creature survived", tc.pick, tc.colour)
			}
		}
		for _, id := range alive {
			if !g.Battlefield.Contains(id) {
				t.Errorf("pick %d: the other colour's creature was destroyed", tc.pick)
			}
		}
		if me.Life != life-6 {
			t.Errorf("pick %d: life %d -> %d, want -6 for three creatures", tc.pick, life, me.Life)
		}
	}
}

// b802Pair is a modal spell's two targets: slot 0 the land, slot 1 the creature.
func b802Pair(_ int, land, creature uuid.UUID) []game.TargetRef {
	return []game.TargetRef{
		{Kind: game.TargetCard, ID: land, Mode: 0, Slot: 0},
		{Kind: game.TargetCard, ID: creature, Mode: 0, Slot: 1},
	}
}
