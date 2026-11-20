package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// c19_5517_b1_cards_test.go — Commander 2019, batch 1 (#5517). Every
// test plays the real catalog entry through the engine.

const (
	c19MireInMiseryOracle        = "da1d1f64-e239-4b92-9eb9-fe9ea06aa18a"
	c19VoiceOfManyOracle         = "9b40faad-1b96-4461-89d8-9a207b58e832"
	c19ThoughtSpongeOracle       = "256e094a-a9e8-42d9-90e2-7a24ed85e426"
	c19SkyfirePhoenixOracle      = "8fdb3350-a6ff-4d02-a649-10663436c40f"
	c19TectonicHellionOracle     = "e67b8eb2-7b39-4fdb-9ee2-686d1d781cfa"
	c19GrismoldOracle            = "d0d28e2c-416e-432c-b4bf-07f6302c52ea"
	c19DoomedArtisanOracle       = "70c92f25-32d6-4cda-b961-cf2680bd59ab"
	c19CliffsideRescuerOracle    = "c53f2d1c-5a53-4020-aace-a1d9f75e47c7"
	c19LeadershipVacuumOracle    = "5f9dd005-1648-451e-b888-69ef81331c95"
	c19MarisiOracle              = "17475cf0-661d-4361-827b-1e4d638089f1"
	c19BackdraftHellkiteOracle   = "fcf9057a-4975-4a7d-85f5-3fdc80942fcd"
	c19GerrardOracle             = "8a1e8435-16f2-475b-a8a9-7b9f0cf4594a"
	c19ThaliasGeistcallerOracle  = "ee8cf4b4-2e23-4f01-a9fc-7b47ddda31b7"
	c19GrevenOracle              = "b22f4797-c74f-4d95-9528-c16c2bc4c05b"
	c19ElshaOfTheInfiniteOracle  = "da339200-ed26-42e4-b24f-b96bc4daef7c"
	c19BearType                  = "Creature — Bear"
	c19ForestType                = "Basic Land — Forest"
	c19SculptureCreatureTypeLine = "Artifact Creature — Sculpture"
)

// c19GraveyardCard puts a card into p's graveyard directly (no event, so
// it was not put there from the battlefield this turn).
func c19GraveyardCard(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id
}

// c19ExileHolds reports whether a card named name is in exile.
func c19ExileHolds(g *game.Game, name string) bool {
	for _, c := range g.Exile.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func TestC195517EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		c19MireInMiseryOracle, c19VoiceOfManyOracle, c19ThoughtSpongeOracle, c19SkyfirePhoenixOracle,
		c19TectonicHellionOracle, c19GrismoldOracle, c19DoomedArtisanOracle, c19CliffsideRescuerOracle,
		c19LeadershipVacuumOracle, c19MarisiOracle, c19BackdraftHellkiteOracle, c19GerrardOracle,
		c19ThaliasGeistcallerOracle, c19GrevenOracle, c19ElshaOfTheInfiniteOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestC195517PrintedKeywords(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Thought Sponge", "Creature — Sponge", c19ThoughtSpongeOracle, []string{"flash"}},
		{"Skyfire Phoenix", "Creature — Phoenix", c19SkyfirePhoenixOracle, []string{"flying", "haste"}},
		{"Tectonic Hellion", "Creature — Hellion", c19TectonicHellionOracle, []string{"haste"}},
		{"Grismold, the Dreadsower", "Legendary Creature — Troll Shaman", c19GrismoldOracle, []string{"trample"}},
		{"Cliffside Rescuer", "Creature — Kor Soldier", c19CliffsideRescuerOracle, []string{"vigilance"}},
		{"Backdraft Hellkite", "Creature — Dragon", c19BackdraftHellkiteOracle, []string{"flying"}},
		{"Gerrard, Weatherlight Hero", "Legendary Creature — Human Soldier", c19GerrardOracle, []string{"first strike"}},
		{"Thalia's Geistcaller", "Creature — Human Cleric", c19ThaliasGeistcallerOracle, []string{"lifelink"}},
		{"Greven, Predator Captain", "Legendary Creature — Phyrexian Human Warrior", c19GrevenOracle, []string{"menace"}},
		{"Elsha of the Infinite", "Legendary Creature — Djinn Monk", c19ElshaOfTheInfiniteOracle, []string{"prowess"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Each opponent has one creature or enchantment at most, so nobody is
// asked; the caster's own creature and an opponent's land stay.
func TestC195517MireInMiseryEachOpponentSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, a, b := blbB1Seats(g)
	c := g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Mine", c19BearType, 2, 2)
	bear := b12Creature(g, a.ID, "Their Bear", c19BearType, 2, 2)
	aura := b12Permanent(g, b.ID, "Their Enchantment", "Enchantment")
	land := b12Permanent(g, c.ID, "Their Forest", c19ForestType)
	castCatalogSpell(t, g, "Mire in Misery", "Sorcery", c19MireInMiseryOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("an opponent kept their only creature")
	}
	if g.Battlefield.Contains(aura) {
		t.Error("an opponent kept their only enchantment")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was sacrificed")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster sacrificed a creature")
	}
}

// Voice of Many: with the Voice in play you control one creature; two
// opponents control none and one controls two, so you draw two.
func TestC195517VoiceOfManyDrawsPerOpponentBehind(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	b12Creature(g, opp.ID, "One", c19BearType, 2, 2)
	b12Creature(g, opp.ID, "Two", c19BearType, 2, 2)
	hand := handSize(me)
	id := castHoldCreature(t, g, "Voice of Many", "Creature — Elf Druid", c19VoiceOfManyOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Voice of Many did not resolve")
	}
	if got := handSize(me) - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

// Thought Sponge enters with counters for the opponent who drew the most
// this turn (its controller's own draws don't count), and draws its power
// when it dies.
func TestC195517ThoughtSpongeCountsTheBiggestOpponentDraw(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, other := blbB1Seats(g)
	if g.DrawnThisTurn == nil {
		g.DrawnThisTurn = map[uuid.UUID][]uuid.UUID{}
	}
	g.DrawnThisTurn[opp.ID] = append(g.DrawnThisTurn[opp.ID], uuid.New(), uuid.New())
	g.DrawnThisTurn[other.ID] = append(g.DrawnThisTurn[other.ID], uuid.New())
	g.DrawnThisTurn[me.ID] = append(g.DrawnThisTurn[me.ID], uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	sponge := castHoldCreature(t, g, "Thought Sponge", "Creature — Sponge", c19ThoughtSpongeOracle, 1, 1)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, sponge); got != 2 {
		t.Fatalf("Thought Sponge has %d +1/+1 counters, want 2", got)
	}
	if p, tough := ptOf(t, g, sponge); p != 3 || tough != 3 {
		t.Fatalf("Thought Sponge is %d/%d, want 3/3", p, tough)
	}
	hand := handSize(me)
	b27Kill(g, sponge)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := handSize(me) - hand; got != 3 {
		t.Errorf("drew %d when it died, want 3", got)
	}
}

func c19CastFromHand(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	advanceToMain(t, g)
	if err := g.CastSpell(p.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", c.Name, err)
	}
	return c.InstanceID
}

func TestC195517SkyfirePhoenixReturnsWhenYouCastYourCommander(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	phoenix := c19GraveyardCard(me, "Skyfire Phoenix", "Creature — Phoenix", c19SkyfirePhoenixOracle)
	c19CastFromHand(t, g, me, game.Card{Name: "Plain Bear", TypeLine: c19BearType, Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(phoenix) {
		t.Fatal("a spell that is not a commander returned the Phoenix")
	}
	c19CastFromHand(t, g, me, game.Card{Name: "My Commander", TypeLine: "Legendary Creature — Elf", Power: 2, Toughness: 2, IsCommander: true})
	passPriorityAroundTable(t, g)
	if _, ok := visBattlefieldNamed(g, me.ID, "Skyfire Phoenix"); !ok {
		t.Error("casting your commander did not return the Phoenix")
	}
	if me.Graveyard.Contains(phoenix) {
		t.Error("the Phoenix is still in the graveyard")
	}
}

// Two players tie for the most lands with two each, so each sacrifices
// both without being asked; the player with fewer keeps theirs.
func TestC195517TectonicHellionHitsTheMostLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	hellion := b12Push(g, me.ID, "Tectonic Hellion", "Creature — Hellion", c19TectonicHellionOracle, 8, 5)
	var mine, theirs []uuid.UUID
	for i := 0; i < 2; i++ {
		mine = append(mine, b12Permanent(g, me.ID, "My Forest", c19ForestType))
		theirs = append(theirs, b12Permanent(g, opp.ID, "Their Forest", c19ForestType))
	}
	lone := b12Permanent(g, other.ID, "Lone Forest", c19ForestType)
	declareAttack(t, g, opp.ID, hellion)
	passPriorityAroundTable(t, g)
	for _, id := range append(mine, theirs...) {
		if g.Battlefield.Contains(id) {
			t.Error("a land of a player with the most lands survived")
		}
	}
	if !g.Battlefield.Contains(lone) {
		t.Error("the player with fewer lands sacrificed one")
	}
}

func TestC195517GrismoldPlantsAndGrowsOnTokenDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grismold := b12Push(g, me.ID, "Grismold, the Dreadsower", "Legendary Creature — Troll Shaman", c19GrismoldOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", c19BearType, 2, 2)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	plants := battlefieldIDsNamed(g, "Plant")
	if len(plants) != len(g.Seats) {
		t.Fatalf("%d Plants, want one per player (%d)", len(plants), len(g.Seats))
	}
	owners := map[uuid.UUID]bool{}
	for _, id := range plants {
		c, _ := battlefieldCard(g, id)
		owners[c.Controller] = true
		if c.Power != 1 || c.Toughness != 1 {
			t.Errorf("a Plant is %d/%d, want 1/1", c.Power, c.Toughness)
		}
	}
	if len(owners) != len(g.Seats) {
		t.Errorf("Plants went to %d players, want %d", len(owners), len(g.Seats))
	}
	b27Kill(g, bear)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, grismold); got != 0 {
		t.Fatalf("a nontoken creature dying gave Grismold %d counters", got)
	}
	for _, id := range plants[:2] {
		b27Kill(g, id)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, grismold); got != 2 {
		t.Errorf("Grismold has %d +1/+1 counters after two tokens died, want 2", got)
	}
}

func TestC195517DoomedArtisanSculptures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Doomed Artisan", "Creature — Human Artificer", c19DoomedArtisanOracle, 1, 1)
	mine := b12Creature(g, me.ID, "My Sculpture", c19SculptureCreatureTypeLine, 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Sculpture", c19SculptureCreatureTypeLine, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", c19BearType, 2, 2)
	if r := effectiveOf(t, g, mine).Restrictions; r&game.CantAttack == 0 || r&game.CantBlock == 0 {
		t.Error("your Sculpture can attack or block")
	}
	if r := effectiveOf(t, g, theirs).Restrictions; r&(game.CantAttack|game.CantBlock) != 0 {
		t.Error("an opponent's Sculpture is restricted")
	}
	if r := effectiveOf(t, g, bear).Restrictions; r&(game.CantAttack|game.CantBlock) != 0 {
		t.Error("a non-Sculpture is restricted")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, id := range battlefieldIDsNamed(g, "Sculpture") {
		if c, _ := battlefieldCard(g, id); c.Controller == me.ID && IsToken(c) {
			token = id
		}
	}
	if token == uuid.Nil {
		t.Fatal("no Sculpture token")
	}
	// Two Sculptures you control: the seeded one and the token.
	if p, tough := ptOf(t, g, token); p != 2 || tough != 2 {
		t.Errorf("the Sculpture token is %d/%d, want 2/2", p, tough)
	}
	if r := effectiveOf(t, g, token).Restrictions; r&game.CantAttack == 0 {
		t.Error("the Sculpture token can attack")
	}
}

func TestC195517CliffsideRescuerGrantsProtection(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	rescuer := b12Push(g, me.ID, "Cliffside Rescuer", "Creature — Kor Soldier", c19CliffsideRescuerOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", c19BearType, 2, 2)
	apaActivate(t, g, me, rescuer, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if g.Battlefield.Contains(rescuer) {
		t.Error("the Rescuer was not sacrificed")
	}
	if !hasEffectiveKeyword(t, g, bear, game.ProtectionFromEachOfYourOpponents) {
		t.Error("the target has no protection from each of your opponents")
	}
}

func TestC195517LeadershipVacuumSendsCommandersHome(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Commander", TypeLine: "Legendary Creature — Elf",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, IsCommander: true,
	})
	other := b12Creature(g, opp.ID, "Their Bear", c19BearType, 2, 2)
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Elf",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	hand := handSize(me)
	castCatalogSpell(t, g, "Leadership Vacuum", "Instant", c19LeadershipVacuumOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Command.Contains(theirs) {
		t.Error("the target player's commander is not in the command zone")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a creature that is not a commander left")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's commander left")
	}
	if got := handSize(me) - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestC195517MarisiOpponentsCantCastDuringCombat(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	b12Push(g, me.ID, "Marisi, Breaker of the Coil", "Legendary Creature — Cat Warrior", c19MarisiOracle, 5, 4)
	if err := e7CastInstant(g, opp); err != nil {
		t.Fatalf("an opponent was refused in the main phase: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	assertCantCast(t, e7CastInstant(g, opp), "during combat")
	if err := e7CastInstant(g, me); err != nil {
		t.Errorf("Marisi's controller was refused in combat: %v", err)
	}
}

func TestC195517MarisiGoadsTheDamagedPlayersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, other := blbB1Seats(g)
	b12Push(g, me.ID, "Marisi, Breaker of the Coil", "Legendary Creature — Cat Warrior", c19MarisiOracle, 5, 4)
	bear := b12Creature(g, me.ID, "Bear", c19BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", c19BearType, 2, 2)
	bystander := b12Creature(g, other.ID, "Bystander", c19BearType, 2, 2)
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		if c, _ := battlefieldCard(g, theirs); !c.IsGoaded() {
			t.Error("the damaged player's creature is not goaded")
		}
		if b, _ := battlefieldCard(g, bystander); b.IsGoaded() {
			t.Error("another player's creature was goaded")
		}
	})
}

func TestC195517BackdraftHellkiteGrantsFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	opp := top100Opponent(g)
	instant := seedGraveyardCard(t, g, "Old Instant", "Instant", "test-c19-instant")
	land := c19GraveyardCard(me, "Old Forest", c19ForestType, "")
	hellkite := b12Push(g, me.ID, "Backdraft Hellkite", "Creature — Dragon", c19BackdraftHellkiteOracle, 4, 4)
	if grantedPermissionOn(g, me.ID, instant, game.ZoneGraveyard).Granted() {
		t.Fatal("the instant had flashback before the attack")
	}
	declareAttack(t, g, opp.ID, hellkite)
	passPriorityAroundTable(t, g)
	perm := grantedPermissionOn(g, me.ID, instant, game.ZoneGraveyard)
	if !perm.Granted() {
		t.Fatal("the instant card gained no flashback")
	}
	if perm.AltCostKey != "flashback" || !perm.ExileOnResolution {
		t.Errorf("granted %q (exile %v), want flashback that exiles", perm.AltCostKey, perm.ExileOnResolution)
	}
	if grantedPermissionOn(g, me.ID, land, game.ZoneGraveyard).Granted() {
		t.Error("a land card gained flashback")
	}
}

func TestC195517GerrardReturnsWhatDiedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	gerrard := b12Push(g, me.ID, "Gerrard, Weatherlight Hero", "Legendary Creature — Human Soldier", c19GerrardOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Fallen Bear", c19BearType, 2, 2)
	rock := b12Permanent(g, me.ID, "Fallen Rock", "Artifact")
	grove := b12Permanent(g, me.ID, "Fallen Grove", "Enchantment")
	old := c19GraveyardCard(me, "Old Bear", c19BearType, "")
	for _, id := range []uuid.UUID{bear, rock, grove} {
		b27Kill(g, id)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	b27Kill(g, gerrard)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	for _, name := range []string{"Fallen Bear", "Fallen Rock"} {
		if _, ok := visBattlefieldNamed(g, me.ID, name); !ok {
			t.Errorf("%s did not return", name)
		}
	}
	if _, ok := visBattlefieldNamed(g, me.ID, "Fallen Grove"); ok {
		t.Error("an enchantment card returned")
	}
	if !me.Graveyard.Contains(old) {
		t.Error("a card that was not put there from the battlefield this turn left the graveyard")
	}
	if !c19ExileHolds(g, "Gerrard, Weatherlight Hero") {
		t.Error("Gerrard was not exiled")
	}
	if _, ok := visBattlefieldNamed(g, me.ID, "Gerrard, Weatherlight Hero"); ok {
		t.Error("Gerrard returned himself")
	}
}

func TestC195517ThaliasGeistcallerMakesSpiritsFromGraveyardCasts(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	spell := seedGraveyardCard(t, g, "Old Instant", "Instant", "test-c19-geist-instant")
	geist := b12Push(g, me.ID, "Thalia's Geistcaller", "Creature — Human Cleric", c19ThaliasGeistcallerOracle, 3, 1)
	c19CastFromHand(t, g, me, game.Card{Name: "Hand Instant", TypeLine: "Instant"})
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Spirit")); n != 0 {
		t.Fatalf("a spell cast from hand made %d Spirits", n)
	}
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(spell, game.CastPermission{
			Player: me.ID, Zone: game.ZoneGraveyard, AltCostKey: "flashback",
			ExileOnResolution: true, Label: "test flashback",
		})
	})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	spirits := battlefieldIDsNamed(g, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirits after a graveyard cast, want 1", len(spirits))
	}
	if !hasEffectiveKeyword(t, g, spirits[0], "flying") {
		t.Error("the Spirit has no flying")
	}
	apaActivate(t, g, me, geist, 0, game.ActivateAbilityParams{SacrificeIDs: spirits})
	if g.Battlefield.Contains(spirits[0]) {
		t.Error("the Spirit was not sacrificed")
	}
	if !hasEffectiveKeyword(t, g, geist, "indestructible") {
		t.Error("the Geistcaller did not gain indestructible")
	}
}

func TestC195517GrevenGrowsWithLifeLostAndSacrificesToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	greven := b12Push(g, me.ID, "Greven, Predator Captain", "Legendary Creature — Phyrexian Human Warrior", c19GrevenOracle, 5, 5)
	fodder := b12Creature(g, me.ID, "Fodder", c19BearType, 3, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3) })
	if p, tough := ptOf(t, g, greven); p != 8 || tough != 5 {
		t.Fatalf("Greven is %d/%d after losing 3 life, want 8/5", p, tough)
	}
	hand, life := handSize(me), me.Life
	declareAttack(t, g, opp.ID, greven)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{greven}); err == nil {
		t.Fatal("Greven sacrificed himself")
	}
	answerOwnPermanents(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed")
	}
	if got := handSize(me) - hand; got != 3 {
		t.Errorf("drew %d, want 3 (the creature's power)", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d (the creature's toughness lost)", me.Life, life-2)
	}
	if p, _ := ptOf(t, g, greven); p != 10 {
		t.Errorf("Greven's power is %d after losing 5 life this turn, want 10", p)
	}
}

func TestC195517ElshaCastsNoncreatureSpellsFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	b12Push(g, me.ID, "Elsha of the Infinite", "Legendary Creature — Djinn Monk", c19ElshaOfTheInfiniteOracle, 3, 3)
	bear := seedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: c19BearType, Power: 2, Toughness: 2})
	if grantedPermissionOn(g, me.ID, bear, game.ZoneLibrary).Granted() {
		t.Fatal("a creature card on top may be cast")
	}
	sorcery := seedTopOfLibrary(me, game.Card{Name: "Top Sorcery", TypeLine: "Sorcery"})
	if !grantedPermissionOn(g, me.ID, sorcery, game.ZoneLibrary).Granted() {
		t.Fatal("a noncreature card on top may not be cast")
	}
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: string(game.ZoneLibrary)}); err != nil {
		t.Fatalf("casting a sorcery from the top at the beginning of combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Library.Contains(sorcery) {
		t.Error("the sorcery is still on top of the library")
	}
}
