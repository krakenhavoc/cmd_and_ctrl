package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0107_pr6_damage_cards_test.go — ADR 0107 PR 6 (#1853): the cards
// "damage can't be prevented" unblocks, each against a prevention
// effect it must get through.

const (
	pr6SkullcrackOracle     = "a6fc0b0c-9ba2-47f2-ab92-26fd20d0f86d"
	pr6CallInOracle         = "ca54e14c-67ee-4de2-bc0f-19a5f9461d37"
	pr6CombustOracle        = "f1ba56e3-3d09-4610-88e7-8681b5736105"
	pr6PinpointOracle       = "d619cce4-036a-4da2-ae64-f43fc095ebe9"
	pr6ArrowStormOracle     = "a0ce2f5a-9c0b-4801-9105-72f84d8a4a1f"
	pr6UrzasRageOracle      = "363f8c66-fe0c-44b9-987d-1d160e3f9c54"
	pr6LavaBurstOracle      = "f7be3da5-55b2-46f2-a5aa-277dee242b94"
	pr6FearFireFoesOracle   = "35a0e085-df04-454d-b4d3-42db387bbad0"
	pr6FlamesOracle         = "4e9df979-c1c2-4de1-944e-c5e2d782e66e"
	pr6FlaringPainOracle    = "cad19ad2-3ea9-43f5-bb5a-b0793402fab0"
	pr6ImpracticalOracle    = "c8da73ed-2834-4094-8723-a4ad75660290"
	pr6WildSlashOracle      = "f76a0e5b-74c1-4ee8-b502-e8b988086de4"
	pr6UnstableOracle       = "fb901095-2ea4-4b9e-ad5a-2892894dc4c3"
	pr6ExcruciatorOracle    = "9911c2f8-fafb-4979-898e-e86b726e3538"
	pr6MalignusOracle       = "f90174b3-effc-4c81-8cb0-856a0646516d"
	pr6LeylineOracle        = "2608df54-dfbe-417d-aef5-49afbdfb03da"
	pr6SunspineOracle       = "479adf64-f42c-4cb6-bf25-648c044bca8b"
	pr6BalothOracle         = "6a4ef075-9254-4d36-9572-622833fae54e"
	pr6GearhulkOracle       = "c4c6083a-b0b5-4c6b-9f11-9f4fe38417c2"
	pr6QuestingBeastOracle  = "b685757b-521e-4353-a233-97052359723d"
	pr6LightningSurgeOracle = "dc55e69e-e1b8-4129-902c-c71bcb952418"
)

// pr6Shielded puts a creature with a charged prevention shield on the
// battlefield for `owner` — a target any spell may choose, and a
// prevention effect that "can't be prevented" must get through.
func pr6Shielded(g *game.Game, owner uuid.UUID, toughness int) uuid.UUID {
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shielded Knight", TypeLine: "Creature — Knight", Colors: []string{"W"},
		Power: 2, Toughness: toughness, Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() { g.PreventNextDamageThisTurnForEffect(uuid.Nil, id, 20, false, "shield") })
	return id
}

// pr6ProRed puts a creature with protection from red on the
// battlefield for `owner`. Used only as an untargeted damage victim: a
// red spell can't target it (CR 702.16b).
func pr6ProRed(g *game.Game, owner uuid.UUID, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pro-Red Knight", TypeLine: "Creature — Knight",
		Power: 2, Toughness: toughness, Colors: []string{"W"}, Keywords: []string{"protection from red"},
		Owner: owner, Controller: owner,
	})
}

func pr6Fog(g *game.Game) {
	g.WithWriteLock(func() { g.PreventCombatDamageThisTurnForEffect(uuid.Nil, uuid.Nil, "Fog") })
}

func pr6ShieldPlayer(g *game.Game, player uuid.UUID, n int) {
	g.WithWriteLock(func() { g.PreventNextDamageThisTurnForEffect(uuid.Nil, player, n, false, "shield") })
}

func pr6Player(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func pr6Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func pr6CantGainLife(g *game.Game, p *game.Player) bool {
	var out bool
	g.ReadSnapshot(func() { out = g.PlayerCantGainLifeLocked(p) })
	return out
}

// Skullcrack: through a shield, and nobody gains life for the rest of
// the turn — the caster included.
func TestPR6SkullcrackDealsThroughAShieldAndStopsLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pr6ShieldPlayer(g, opp.ID, 10)
	before := opp.Life
	castCatalogSpell(t, g, "Skullcrack", "Instant", pr6SkullcrackOracle, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Fatalf("opponent's life %d → %d, want 3 through the shield", before, opp.Life)
	}
	if !pr6CantGainLife(g, me) || !pr6CantGainLife(g, opp) {
		t.Error("players can't gain life this turn, the caster included")
	}
	mine := me.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 5) })
	if me.Life != mine {
		t.Errorf("the caster gained life under Skullcrack: %d → %d", mine, me.Life)
	}
}

func TestPR6CallInAProfessionalHitsACreatureThroughProtection(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	knight := pr6Shielded(g, opp.ID, 5)
	castCatalogSpell(t, g, "Call In a Professional", "Instant", pr6CallInOracle, pr6Card(knight))
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, knight); got != 3 {
		t.Errorf("the protected creature has %d damage, want 3", got)
	}
	if !pr6CantGainLife(g, opp) {
		t.Error("players can't gain life this turn")
	}
}

// Combust's damage gets through protection; the stack shows it.
func TestPR6CombustIsMarkedOnTheStackAndGetsThrough(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	knight := pr6Shielded(g, opp.ID, 6)
	id := castCatalogSpell(t, g, "Combust", "Instant", pr6CombustOracle, pr6Card(knight))
	var chip, uncounterable bool
	g.ReadSnapshot(func() {
		chip = g.SpellDamageCantBePreventedForEffect(id)
		uncounterable = g.SpellCantBeCounteredForEffect(id)
	})
	if !chip || !uncounterable {
		t.Errorf("Combust on the stack: damage chip %v, can't be countered %v; want both", chip, uncounterable)
	}
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, knight); got != 5 {
		t.Errorf("knight has %d damage, want 5", got)
	}
}

func TestPR6PinpointAvalancheGetsThroughProtection(t *testing.T) {
	g := newCatalogGame(t)
	knight := pr6Shielded(g, g.Seats[1].ID, 6)
	castCatalogSpell(t, g, "Pinpoint Avalanche", "Instant", pr6PinpointOracle, pr6Card(knight))
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, knight); got != 4 {
		t.Errorf("knight has %d damage, want 4", got)
	}
}

// Banefire's riders are both X of 5 or more.
func TestPR6BanefireRidersFollowX(t *testing.T) {
	for _, tc := range []struct {
		x       int
		through bool
	}{{4, false}, {5, true}} {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		pr6ShieldPlayer(g, opp.ID, 20)
		before := opp.Life
		id := b19CastXModal(t, g, "Banefire", "Sorcery", b39BanefireOracle, "{X}{R}", tc.x, nil, pr6Player(opp.ID))
		var uncounterable bool
		g.ReadSnapshot(func() { uncounterable = g.SpellCantBeCounteredForEffect(id) })
		if uncounterable != tc.through {
			t.Errorf("X=%d: can't be countered = %v, want %v", tc.x, uncounterable, tc.through)
		}
		passPriorityAroundTable(t, g)
		got := before - opp.Life
		want := 0
		if tc.through {
			want = tc.x
		}
		if got != want {
			t.Errorf("X=%d: opponent lost %d through a 20-point shield, want %d", tc.x, got, want)
		}
	}
}

// Stomp: the turn grant covers its own 2 damage and a later Fog'd
// combat damage alike.
func TestPR6StompGrantsForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pr6ShieldPlayer(g, opp.ID, 10)
	before := opp.Life
	castCatalogSpell(t, g, "Stomp", "Instant", bonecrusherGiantOracleID+"#1", pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent's life %d → %d, want 2 through the shield", before, opp.Life)
	}
	var lines []string
	g.ReadSnapshot(func() { lines = g.DamageCantBePreventedThisTurnLabels() })
	if len(lines) != 1 {
		t.Errorf("turn grants %v, want one", lines)
	}
}

// Arrow Storm: raid decides both the amount and the rider.
func TestPR6ArrowStormWithoutRaidIsPreventable(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pr6ShieldPlayer(g, opp.ID, 2)
	before := opp.Life
	castCatalogSpell(t, g, "Arrow Storm", "Sorcery", pr6ArrowStormOracle, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("no raid: 4 damage less a 2-point shield, opponent lost %d", before-opp.Life)
	}
}

// Urza's Rage unkicked is 3 preventable damage.
func TestPR6UrzasRageUnkickedIsPreventable(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pr6ShieldPlayer(g, opp.ID, 10)
	before := opp.Life
	id := castCatalogSpell(t, g, "Urza's Rage", "Instant", pr6UrzasRageOracle, pr6Player(opp.ID))
	var chip bool
	g.ReadSnapshot(func() { chip = g.SpellDamageCantBePreventedForEffect(id) })
	if chip {
		t.Error("an unkicked Urza's Rage is not marked")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("unkicked Rage: the shield takes all 3, opponent lost %d", before-opp.Life)
	}
}

// Lava Burst: unpreventable only at a creature.
func TestPR6LavaBurstIsUnpreventableAtACreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	knight := pr6Shielded(g, opp.ID, 9)
	b19CastXModal(t, g, "Lava Burst", "Sorcery", pr6LavaBurstOracle, "{X}{R}", 3, nil, pr6Card(knight))
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, knight); got != 3 {
		t.Errorf("knight has %d damage, want 3", got)
	}
	pr6ShieldPlayer(g, opp.ID, 10)
	before := opp.Life
	b19CastXModal(t, g, "Lava Burst", "Sorcery", pr6LavaBurstOracle, "{X}{R}", 3, nil, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("at a player the damage is preventable: opponent lost %d", before-opp.Life)
	}
}

// Fear, Fire, Foes!: X to the target, 1 to each other creature its
// controller controls, none to anyone else's — all through protection.
func TestPR6FearFireFoesSplashesTheSameController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := pr6Shielded(g, opp.ID, 9)
	friend := pr6Shielded(g, opp.ID, 9)
	mine := pr6Shielded(g, me.ID, 9)
	b19CastXModal(t, g, "Fear, Fire, Foes!", "Sorcery", pr6FearFireFoesOracle, "{X}{R}", 4, nil, pr6Card(target))
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, target); got != 4 {
		t.Errorf("target has %d, want 4", got)
	}
	if got := damageMarkedOn(g, friend); got != 1 {
		t.Errorf("the target's controller's other creature has %d, want 1", got)
	}
	if got := damageMarkedOn(g, mine); got != 0 {
		t.Errorf("another player's creature has %d, want 0", got)
	}
}

// Flames of the Blood Hand: 4 through a shield, and the player gains no
// life this turn.
func TestPR6FlamesOfTheBloodHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pr6ShieldPlayer(g, opp.ID, 10)
	before := opp.Life
	castCatalogSpell(t, g, "Flames of the Blood Hand", "Instant", pr6FlamesOracle, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Fatalf("opponent lost %d, want 4", before-opp.Life)
	}
	theirs, mine := opp.Life, me.Life
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3)
	})
	if opp.Life != theirs {
		t.Errorf("the damaged player gained life: %d → %d", theirs, opp.Life)
	}
	if me.Life != mine+3 {
		t.Errorf("the caster's gain was stopped: %d → %d", mine, me.Life)
	}
}

func TestPR6FlaringPainAndImpracticalJokeGrantTheTurn(t *testing.T) {
	for _, tc := range []struct{ name, oracle, typeLine string }{
		{"Flaring Pain", pr6FlaringPainOracle, "Instant"},
		{"Impractical Joke", pr6ImpracticalOracle, "Sorcery"},
	} {
		g := newCatalogGame(t)
		castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
		passPriorityAroundTable(t, g)
		var lines []string
		g.ReadSnapshot(func() { lines = g.DamageCantBePreventedThisTurnLabels() })
		if len(lines) != 1 || lines[0] != tc.name {
			t.Errorf("%s: turn grants %v, want [%s]", tc.name, lines, tc.name)
		}
	}
}

// Wild Slash's grant needs ferocious; the 2 damage does not.
func TestPR6WildSlashGrantsOnlyWithFerocious(t *testing.T) {
	for _, ferocious := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if ferocious {
			pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Beast",
				Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
		}
		before := opp.Life
		castCatalogSpell(t, g, "Wild Slash", "Instant", pr6WildSlashOracle, pr6Player(opp.ID))
		passPriorityAroundTable(t, g)
		var lines []string
		g.ReadSnapshot(func() { lines = g.DamageCantBePreventedThisTurnLabels() })
		if (len(lines) == 1) != ferocious {
			t.Errorf("ferocious %v: grants %v", ferocious, lines)
		}
		if opp.Life != before-2 {
			t.Errorf("ferocious %v: opponent lost %d, want 2", ferocious, before-opp.Life)
		}
	}
}

// Unstable Footing unkicked targets nothing and only grants.
func TestPR6UnstableFootingUnkickedOnlyGrants(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	castCatalogSpell(t, g, "Unstable Footing", "Instant", pr6UnstableOracle, nil)
	passPriorityAroundTable(t, g)
	var lines []string
	g.ReadSnapshot(func() { lines = g.DamageCantBePreventedThisTurnLabels() })
	if len(lines) != 1 {
		t.Errorf("grants %v, want one", lines)
	}
	if opp.Life != before {
		t.Errorf("unkicked Footing dealt damage: %d", before-opp.Life)
	}
}

// Excruciator's damage is unpreventable, its opponent's is not.
func TestPR6ExcruciatorsDamageGetsThrough(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ex := pushCatalogPermanent(g, me.ID, "Excruciator", "Creature — Avatar", pr6ExcruciatorOracle, false)
	other := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 20, Owner: opp.ID, Controller: opp.ID})
	pr6Fog(g)
	if err := g.MarkCombatDamage(ex, victim, 7); err != nil {
		t.Fatal(err)
	}
	if err := g.MarkCombatDamage(other, victim, 2); err != nil {
		t.Fatal(err)
	}
	if got := damageMarkedOn(g, victim); got != 7 {
		t.Errorf("victim has %d, want 7 — Excruciator's damage through the Fog, the Bear's prevented", got)
	}
}

// Malignus is half the highest opponent life total, rounded up.
func TestPR6MalignusSizesOffTheHighestOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		p.Life = 20
	}
	opp.Life = 25
	me.Life = 40
	id := pushCatalogPermanent(g, me.ID, "Malignus", "Creature — Elemental Spirit", pr6MalignusOracle, false)
	var p, tough int
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.LookupCardForEffect(id)
		p, tough = c.CurrentPower(), c.CurrentToughness()
	})
	if p != 13 || tough != 13 {
		t.Errorf("Malignus is %d/%d with an opponent at 25, want 13/13", p, tough)
	}
}

// Leyline of Punishment: nobody gains life, and protection stops nothing.
func TestPR6LeylineOfPunishment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Leyline of Punishment", "Enchantment", pr6LeylineOracle, false)
	knight := pr6ProRed(g, opp.ID, 9)
	bolt := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Pinger", TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	mine := me.Life
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(bolt, knight, 2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 4)
	})
	if got := damageMarkedOn(g, knight); got != 2 {
		t.Errorf("knight has %d, want 2 through protection", got)
	}
	if me.Life != mine {
		t.Errorf("the Leyline's controller gained life: %d → %d", mine, me.Life)
	}
}

// Sunspine Lynx counts each player's nonbasic lands.
func TestPR6SunspineLynxPunishesNonbasics(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Temple", TypeLine: "Land",
			Owner: opp.ID, Controller: opp.ID})
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain",
		Owner: me.ID, Controller: me.ID})
	mine, theirs := me.Life, opp.Life
	castCatalogSpell(t, g, "Sunspine Lynx", "Creature — Elemental Cat", pr6SunspineOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != theirs-2 {
		t.Errorf("opponent with two nonbasics lost %d, want 2", theirs-opp.Life)
	}
	if me.Life != mine {
		t.Errorf("controller with only a basic lost %d, want 0", mine-me.Life)
	}
}

// Frenzied Baloth: combat damage through a Fog, a spell's damage not.
func TestPR6FrenziedBalothBreaksFogForCombatOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Frenzied Baloth", "Creature — Beast", pr6BalothOracle, false)
	attacker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Opp Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	blocker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Wall", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 9, Owner: me.ID, Controller: me.ID})
	pr6Fog(g)
	if err := g.MarkCombatDamage(attacker, blocker, 2); err != nil {
		t.Fatal(err)
	}
	if got := damageMarkedOn(g, blocker); got != 2 {
		t.Errorf("combat damage under Fog + Baloth: %d, want 2", got)
	}
	knight := pr6ProRed(g, opp.ID, 9)
	red := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red", TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(red, knight, 2) })
	if got := damageMarkedOn(g, knight); got != 0 {
		t.Errorf("noncombat damage got through protection under the Baloth: %d", got)
	}
}

// Pyrewood Gearhulk: the others get +2/+2, it does not, and the turn
// grant is live.
func TestPR6PyrewoodGearhulkPumpsTheOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Pyrewood Gearhulk", "Artifact Creature — Construct", pr6GearhulkOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	var p int
	var lines []string
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.LookupCardForEffect(bear)
		p = c.CurrentPower()
		lines = g.DamageCantBePreventedThisTurnLabels()
	})
	if p != 4 {
		t.Errorf("the Bear's power is %d, want 4", p)
	}
	if len(lines) != 1 {
		t.Errorf("turn grants %v, want one", lines)
	}
}

// Questing Beast's trigger targets only a planeswalker the damaged
// opponent controls.
func TestPR6QuestingBeastTargetsTheDamagedOpponentsPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := pushCatalogPermanent(g, me.ID, "Questing Beast", "Legendary Creature — Beast", pr6QuestingBeastOracle, false)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 9}, Owner: opp.ID, Controller: opp.ID})
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 9}, Owner: me.ID, Controller: me.ID})
	spec, _ := Lookup(pr6QuestingBeastOracle)
	clause := spec.Triggered[0].TargetsFrom(game.TriggerContext{Event: game.Event{
		Kind: game.EventDealDamage, Source: beast, Target: opp.ID, Amount: 4, Combat: true,
	}}, nil, g)
	if clause == nil {
		t.Fatal("no clause for a damaged opponent")
	}
	var their, own game.Card
	g.ReadSnapshot(func() {
		their, _ = g.LookupCardForEffect(theirs)
		own, _ = g.LookupCardForEffect(mine)
	})
	if !clause.CardOK(g, me.ID, their, game.ZoneBattlefield) {
		t.Error("the damaged opponent's planeswalker must be a legal target")
	}
	if clause.CardOK(g, me.ID, own, game.ZoneBattlefield) {
		t.Error("another player's planeswalker must not be")
	}
}

// Lightning Surge without threshold is 4 preventable damage.
func TestPR6LightningSurgeWithoutThreshold(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pr6ShieldPlayer(g, opp.ID, 1)
	before := opp.Life
	castCatalogSpell(t, g, "Lightning Surge", "Sorcery", pr6LightningSurgeOracle, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("no threshold: 4 less a 1-point shield, opponent lost %d", before-opp.Life)
	}
}
