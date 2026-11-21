package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsco "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/co"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprB20AjaniOracle      = "8a70ddc5-bb51-45eb-96f2-8ecaecac0cc2"
	rprB20ArchonOracle     = "61f55e39-4028-4c91-910e-16b88c73beb5"
	rprB20HuntbonderOracle = "4957c166-9734-4ba0-a4b9-2a248c0ff38f"
	rprB20BantOracle       = "0a1d817d-dce8-4e83-a380-909f7c9eee46"
	rprB20ChampionOracle   = "cad09970-14c8-4d80-82fe-6c855efb0191"
	rprB20CamaraderieOrcl  = "9aab386c-d48c-4611-b757-aa69b26cc1b1"
	rprB20GiantOracle      = "49a46cfa-f2e9-4718-a6cb-47dbb0792b56"
	rprB20DaxosOracle      = "4cf10cb6-b082-4c0d-89e9-43cd39716201"
	rprB20RegentOracle     = "71b5fdf2-4475-4102-85eb-9be164a9ae29"
	rprB20DeclarationOrcl  = "49554441-bb48-4829-b8d6-55e687a6c2e2"
	rprB20DoubleVisionOrcl = "6629a48b-be9a-4636-8add-69d096a1ca7e"
)

func TestRprB20EveryCardIsRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB20AjaniOracle:      "Ajani Unyielding",
		rprB20ArchonOracle:     "Archon of Coronation",
		rprB20HuntbonderOracle: "Avenging Huntbonder",
		rprB20BantOracle:       "Bant Panorama",
		rprB20ChampionOracle:   "Bloodsoaked Champion",
		rprB20CamaraderieOrcl:  "Camaraderie",
		rprB20GiantOracle:      "Crystalline Giant",
		rprB20DaxosOracle:      "Daxos of Meletis",
		rprB20RegentOracle:     "Deathbringer Regent",
		rprB20DeclarationOrcl:  "Declaration in Stone",
		rprB20DoubleVisionOrcl: "Double Vision",
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

// rprB20Creature seeds a creature (catalog row when oracle is set).
func rprB20Creature(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	return effectstest.B12Push(g, owner, name, "Creature — Test", oracle, power, toughness)
}

func rprB20Counters(g *game.Game, id uuid.UUID) map[string]int {
	c, _ := g.LookupCardForEffect(id)
	return c.Counters
}

// --- Bant Panorama ---------------------------------------------------------------

func TestRprB20BantPanoramaFetchesForestPlainsOrIslandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	panorama := effectstest.B43Catalog(g, me.ID, "Bant Panorama", "Land", rprB20BantOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, panorama, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the fetch opens a search")
	}
	if effectstest.SearchOptionNamed(g, c, "Mountain") != uuid.Nil {
		t.Error("a Mountain is none of Forest, Plains or Island")
	}
	island := effectstest.SearchOptionNamed(g, c, "Island")
	if island == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Forest") == uuid.Nil {
		t.Fatal("the Island and the Forest are offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, island)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(panorama) {
		t.Error("the Panorama sacrificed itself to pay")
	}
	if !g.Battlefield.Contains(island) || !effectstest.B16Tapped(t, g, island) {
		t.Error("the basic arrives on the battlefield tapped")
	}
}

// --- Camaraderie -----------------------------------------------------------------

func TestRprB20CamaraderieGainsAndDrawsPerCreatureAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 12)
	a := rprB20Creature(g, me.ID, "Bear A", "", 2, 2)
	b := rprB20Creature(g, me.ID, "Bear B", "", 2, 2)
	other := rprB20Creature(g, g.Seats[1].ID, "Their Bear", "", 2, 2)
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Camaraderie", "Sorcery", rprB20CamaraderieOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d: two creatures", me.Life, life+2)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		c, _ := g.LookupCardForEffect(id)
		if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
			t.Errorf("%s is %d/%d, want 3/3", c.Name, c.CurrentPower(), c.CurrentToughness())
		}
	}
	if c, _ := g.LookupCardForEffect(other); c.CurrentPower() != 2 {
		t.Error("an opponent's creature was pumped")
	}
}

// --- Double Vision ---------------------------------------------------------------

func TestRprB20DoubleVisionCopiesOnlyTheFirstInstantOrSorceryEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 20)
	effectstest.PushPermanentForTest(g, me.ID, "Double Vision", rprB20DoubleVisionOrcl, "Enchantment")
	hand := me.Hand.Size()
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Fatalf("drew %d, want 4: the Divination and its copy", got)
	}
	hand = me.Hand.Size()
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d on the second spell, want 2: no copy", got)
	}
}

func TestRprB20DoubleVisionIgnoresACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 20)
	effectstest.PushPermanentForTest(g, me.ID, "Double Vision", rprB20DoubleVisionOrcl, "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d, want 4: the creature spell was not the first instant or sorcery", got)
	}
}

// --- Declaration in Stone --------------------------------------------------------

func TestRprB20DeclarationExilesEverySameNamedCreatureAndCluesForNontoken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := rprB20Creature(g, opp.ID, "Goblin Raider", "", 2, 2)
	second := rprB20Creature(g, opp.ID, "Goblin Raider", "", 2, 2)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Raider", TypeLine: "Token Creature — Goblin",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	different := rprB20Creature(g, opp.ID, "Goblin Chief", "", 2, 2)
	mine := rprB20Creature(g, me.ID, "Goblin Raider", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Declaration in Stone", "Sorcery", rprB20DeclarationOrcl, []game.TargetRef{{Kind: game.TargetCard, ID: first}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{first, second, token} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s still on the battlefield", id)
		}
	}
	if !effectstest.InExile(g, first) || !effectstest.InExile(g, second) {
		t.Error("the nontoken copies are in exile")
	}
	if !g.Battlefield.Contains(different) || !g.Battlefield.Contains(mine) {
		t.Error("a differently named creature, or another player's same-named one, was exiled")
	}
	clues := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == opp.ID && c.Name == "Clue" {
			clues++
		}
	}
	if clues != 2 {
		t.Errorf("the controller made %d Clues, want 2 (two nontoken creatures)", clues)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Clue" {
			t.Error("the caster got a Clue")
		}
	}
}

// --- Archon of Coronation --------------------------------------------------------

func TestRprB20ArchonBecomesMonarchOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Archon of Coronation", TypeLine: "Creature — Archon",
		OracleID: rprB20ArchonOracle, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	effectstest.ToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Errorf("monarch = %v, want the Archon's controller", g.Monarch)
	}
}

func rprB20ArchonBoard(t *testing.T, monarch bool) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rprB20Creature(g, opp.ID, "Archon of Coronation", rprB20ArchonOracle, 5, 5)
	if monarch {
		g.Monarch = opp.ID
	}
	return g, opp
}

func TestRprB20ArchonStopsDamageLifeLossWhileMonarchOnly(t *testing.T) {
	for _, monarch := range []bool{true, false} {
		g, opp := rprB20ArchonBoard(t, monarch)
		before := opp.Life
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
		effectstest.PassPriorityAroundTable(t, g)
		want := before - 3
		if monarch {
			want = before
		}
		if opp.Life != want {
			t.Errorf("monarch=%v: life %d, want %d", monarch, opp.Life, want)
		}
	}
}

func TestRprB20ArchonDoesNotStopLifeLossThatIsNotDamage(t *testing.T) {
	g, opp := rprB20ArchonBoard(t, true)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -4) })
	if opp.Life != before-4 {
		t.Errorf("life %d, want %d: a drain is not damage", opp.Life, before-4)
	}
}

func TestRprB20ArchonCombatDamageStillMakesTheAttackerTheMonarch(t *testing.T) {
	g, opp := rprB20ArchonBoard(t, true)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("life %d, want %d: combat damage costs no life", opp.Life, before)
	}
	if g.Monarch != me.ID {
		t.Errorf("monarch = %v, want the attacker's controller", g.Monarch)
	}
}

// --- Avenging Huntbonder ---------------------------------------------------------

func TestRprB20HuntbonderPutsADoubleStrikeCounterOnAnotherAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunt := rprB20Creature(g, me.ID, "Avenging Huntbonder", rprB20HuntbonderOracle, 3, 3)
	other := rprB20Creature(g, me.ID, "Other Attacker", "", 2, 2)
	idle := rprB20Creature(g, me.ID, "Stay Home", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, hunt, other)
	pick := effectstest.PickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{hunt, idle} {
		for _, offered := range pick.PickTargetCards {
			if offered == bad {
				t.Errorf("%s was offered as the target", bad)
			}
		}
	}
	effectstest.AnswerPickTarget(t, g, other)
	effectstest.PassPriorityAroundTable(t, g)
	if rprB20Counters(g, other)[game.CounterDoubleStrike] != 1 {
		t.Errorf("counters on the other attacker: %v", rprB20Counters(g, other))
	}
	if rprB20Counters(g, hunt)[game.CounterDoubleStrike] != 0 {
		t.Error("the Huntbonder put a counter on itself")
	}
	if !effectstest.P1g2HasKeyword(g, other, "double strike") {
		t.Error("the counter grants double strike")
	}
}

// --- Deathbringer Regent ---------------------------------------------------------

func rprB20CastRegent(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Deathbringer Regent", TypeLine: "Creature — Dragon",
		OracleID: rprB20RegentOracle, Power: 5, Toughness: 6, Owner: me.ID, Controller: me.ID})
	effectstest.ToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestRprB20RegentDestroysAllOtherCreaturesWithFiveOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var others []uuid.UUID
	for i := 0; i < 3; i++ {
		others = append(others, rprB20Creature(g, me.ID, "Mine", "", 2, 2))
	}
	for i := 0; i < 2; i++ {
		others = append(others, rprB20Creature(g, g.Seats[1].ID, "Theirs", "", 2, 2))
	}
	regent := rprB20CastRegent(t, g, me)
	if !g.Battlefield.Contains(regent) {
		t.Fatal("the Regent was destroyed with the rest")
	}
	for _, id := range others {
		if g.Battlefield.Contains(id) {
			t.Error("another creature survived")
		}
	}
}

func TestRprB20RegentDoesNothingWithOnlyFourOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var others []uuid.UUID
	for i := 0; i < 4; i++ {
		others = append(others, rprB20Creature(g, g.Seats[1+i%3].ID, "Other", "", 2, 2))
	}
	rprB20CastRegent(t, g, me)
	for _, id := range others {
		if !g.Battlefield.Contains(id) {
			t.Error("a creature was destroyed with fewer than five others")
		}
	}
}

func TestRprB20RegentDoesNothingWhenNotCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var others []uuid.UUID
	for i := 0; i < 5; i++ {
		others = append(others, rprB20Creature(g, g.Seats[1].ID, "Other", "", 2, 2))
	}
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Deathbringer Regent", TypeLine: "Creature — Dragon",
		OracleID: rprB20RegentOracle, Power: 5, Toughness: 6, Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("move: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, o := range others {
		if !g.Battlefield.Contains(o) {
			t.Error("a creature was destroyed though the Regent was not cast")
		}
	}
}

// --- Bloodsoaked Champion --------------------------------------------------------

func TestRprB20ChampionCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := rprB20Creature(g, opp.ID, "Bloodsoaked Champion", rprB20ChampionOracle, 2, 1)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(champ, bear); err == nil {
		t.Error("the Champion blocked")
	}
}

func TestRprB20ChampionReturnsFromGraveyardOnlyAfterYouAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogGraveyardCard(me, "Bloodsoaked Champion", "Creature — Human Warrior", rprB20ChampionOracle, 2, 1)
	effectstest.ToMain(t, g)
	effectstest.B06AddMana(me, "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("raid: activated without having attacked")
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate after attacking: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("the Champion did not return to the battlefield")
	}
}

// --- Crystalline Giant -----------------------------------------------------------

func TestRprB20GiantGainsTenDifferentCountersThenStops(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	giant := rprB20Creature(g, g.Seats[0].ID, "Crystalline Giant", rprB20GiantOracle, 3, 3)
	for turn := 1; turn <= 11; turn++ {
		rprB20NextOwnCombat(t, g, turn == 1)
		total := 0
		for kind, n := range rprB20Counters(g, giant) {
			if n > 1 {
				t.Fatalf("turn %d: %d %s counters", turn, n, kind)
			}
			total += n
		}
		want := min(turn, len(cardsco.CrystallineGiantKinds))
		if total != want {
			t.Fatalf("turn %d: %d counters, want %d (%v)", turn, total, want, rprB20Counters(g, giant))
		}
	}
}

// rprB20NextOwnCombat walks to seat 0's next beginning of combat (the
// current turn's when first is set) and lets the trigger resolve.
func rprB20NextOwnCombat(t *testing.T, g *game.Game, first bool) {
	t.Helper()
	seq := g.Turn.Seq
	for i := 0; i < 400; i++ {
		if g.Turn.Step == game.StepBeginCombat && g.Turn.ActiveSeat == 0 && (first || g.Turn.Seq != seq) {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Step != game.StepBeginCombat || g.Turn.ActiveSeat != 0 {
		t.Fatalf("never reached seat 0's beginning of combat (at %s, seat %d)", g.Turn.Step, g.Turn.ActiveSeat)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Daxos of Meletis ------------------------------------------------------------

func TestRprB20DaxosCantBeBlockedByPowerThreeOrGreater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	daxos := rprB20Creature(g, me.ID, "Daxos of Meletis", rprB20DaxosOracle, 2, 2)
	big := rprB20Creature(g, opp.ID, "Big", "", 3, 3)
	small := rprB20Creature(g, opp.ID, "Small", "", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(daxos, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(big, daxos); err == nil {
		t.Error("a power-3 creature blocked Daxos")
	}
	if err := g.DeclareBlocker(small, daxos); err != nil {
		t.Errorf("a power-2 creature could not block: %v", err)
	}
}

func TestRprB20DaxosExilesTopCardGainsLifeAndMayCastIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	daxos := rprB20Creature(g, me.ID, "Daxos of Meletis", rprB20DaxosOracle, 2, 2)
	top := effectstest.SeedLibraryTop(opp, "Their Spell", "Sorcery", "{3}{R}")
	life := me.Life
	effectstest.AttackWith(t, g, opp.ID, daxos)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, top) {
		t.Fatal("the top card of the damaged player's library was not exiled")
	}
	if me.Life != life+4 {
		t.Errorf("life = %d, want %d: the card's mana value", me.Life, life+4)
	}
	c, _ := g.LookupCardForEffect(top)
	if c.Owner != opp.ID {
		t.Error("the exiled card changed owner")
	}
}

// --- Ajani Unyielding ------------------------------------------------------------

func rprB20Ajani(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.PushCatalogWalkerWithLine(g, me.ID, "Ajani Unyielding", "Legendary Planeswalker — Ajani", rprB20AjaniOracle, 4)
}

func TestRprB20AjaniPlusTwoTakesNonlandPermanentsAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ajani := rprB20Ajani(g, me)
	effectstest.Msc18Library(me, 6)
	land := effectstest.SeedLibraryTop(me, "A Land", "Basic Land — Forest", "")
	spell := effectstest.SeedLibraryTop(me, "A Spell", "Sorcery", "{1}")
	bear := effectstest.SeedLibraryTop(me, "A Bear", "Creature — Bear", "{1}{G}")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, ajani, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate +2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyCount(g, ajani) != 6 {
		t.Errorf("loyalty %d, want 6", effectstest.LoyaltyCount(g, ajani))
	}
	if !me.Hand.Contains(bear) {
		t.Error("the creature card did not go to hand")
	}
	if me.Hand.Contains(land) || me.Hand.Contains(spell) {
		t.Error("a land or an instant or sorcery card went to hand")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1", got)
	}
}

func TestRprB20AjaniMinusTwoExilesAndGivesItsControllerLifeEqualToPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	ajani := rprB20Ajani(g, me)
	victim := rprB20Creature(g, opp.ID, "Big Bear", "", 5, 5)
	life := opp.Life
	effectstest.ActivateBolas(t, g, me, ajani, 1, game.TargetRef{Kind: game.TargetCard, ID: victim})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, victim) {
		t.Error("the creature was not exiled")
	}
	if opp.Life != life+5 {
		t.Errorf("its controller's life = %d, want %d", opp.Life, life+5)
	}
	if effectstest.LoyaltyCount(g, ajani) != 2 {
		t.Errorf("loyalty %d, want 2", effectstest.LoyaltyCount(g, ajani))
	}
}

func TestRprB20AjaniMinusNineCountersEveryCreatureAndOtherPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	ajani := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Ajani Unyielding", "Legendary Planeswalker — Ajani", rprB20AjaniOracle, 9)
	a := rprB20Creature(g, me.ID, "Bear A", "", 2, 2)
	theirs := rprB20Creature(g, opp.ID, "Their Bear", "", 2, 2)
	otherWalker := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Other Walker", "Legendary Planeswalker — Test", "", 3)
	theirWalker := effectstest.PushCatalogWalkerWithLine(g, opp.ID, "Their Walker", "Legendary Planeswalker — Test", "", 3)
	effectstest.ActivateBolas(t, g, me, ajani, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if rprB20Counters(g, a)[game.CounterPlusOne] != 5 {
		t.Errorf("my creature has %v", rprB20Counters(g, a))
	}
	if rprB20Counters(g, theirs)[game.CounterPlusOne] != 0 {
		t.Error("an opponent's creature got counters")
	}
	if effectstest.LoyaltyCount(g, otherWalker) != 8 {
		t.Errorf("my other planeswalker has %d loyalty, want 8", effectstest.LoyaltyCount(g, otherWalker))
	}
	if effectstest.LoyaltyCount(g, theirWalker) != 3 {
		t.Error("an opponent's planeswalker got loyalty")
	}
}
