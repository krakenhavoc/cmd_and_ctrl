package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b05_cards_test.go — S82 ONE batch 5 (#5479): toxic matters and
// toxic grants. Every test plays the real catalog card through the engine.

const (
	oneB05Raptor      = "5d55c397-3828-4e4b-8a61-8c7272ba4800"
	oneB05Singer      = "1c832491-6b78-49c6-baf2-63f17847f5f5"
	oneB05Zealot      = "4c11bda3-903f-4b7c-9e5e-f6c5c91977cf"
	oneB05Devotion    = "c2a6cfb7-e5bd-49c8-aa51-fd5203eca8b8"
	oneB05Slash       = "573a3afb-3a2a-41b7-ac1a-9685b90bffa3"
	oneB05Mantle      = "b4e85b7c-3047-46d1-a6e6-2ec5eac0268a"
	oneB05Ascent      = "ed9c488a-9d44-4d2d-bcf8-81352193b8c2"
	oneB05Injector    = "8026533d-c0e7-48b7-9e1a-9a370899aa57"
	oneB05Communion   = "ff58459a-0015-4012-86c0-580da4e02676"
	oneB05Nurse       = "af7abd22-3134-4872-bbe2-0ad84742d9e5"
	oneB05Rotpriest   = "7d9a424e-cab1-410f-9c04-ee0025d04169"
	oneB05Ribskiff    = "8f761657-7127-4552-a200-532e063e2419"
	oneB05SingerType  = "Creature — Phyrexian Cleric"
	oneB05AuraType    = "Enchantment — Aura"
	oneB05InstantType = "Instant"
)

func TestOneB05EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oneB05Raptor: "Flensing Raptor", oneB05Singer: "Slaughter Singer", oneB05Zealot: "Porcelain Zealot",
		oneB05Devotion: "Compleat Devotion", oneB05Slash: "Hexgold Slash", oneB05Mantle: "Maze's Mantle",
		oneB05Ascent: "Aspirant's Ascent", oneB05Injector: "Prosthetic Injector",
		oneB05Communion: "Necrogen Communion", oneB05Nurse: "Plague Nurse",
		oneB05Rotpriest: "Necrogen Rotpriest", oneB05Ribskiff: "Ribskiff",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oneB05Toxic(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := -1
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				n = game.ToxicTotal(&g.Battlefield.Cards[i])
			}
		}
	})
	return n
}

// A creature with toxic, as the catalog gives it (Slaughter Singer, toxic 2).
func oneB05ToxicCreature(g *game.Game, owner uuid.UUID, power, toughness int) uuid.UUID {
	return b12Push(g, owner, "Slaughter Singer", oneB05SingerType, oneB05Singer, power, toughness)
}

func TestOneB05HexgoldSlashDealsTwoOrFourIfToxic(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	plain := b12Creature(g, opp.ID, "Plain", "Creature — Bear", 2, 3)
	toxic := oneB05ToxicCreature(g, opp.ID, 2, 4)
	castCatalogSpell(t, g, "Hexgold Slash", oneB05InstantType, oneB05Slash, b16TargetCard(plain))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(plain) {
		t.Error("2 damage killed a 3-toughness creature without toxic")
	}
	castCatalogSpell(t, g, "Hexgold Slash", oneB05InstantType, oneB05Slash, b16TargetCard(toxic))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(toxic) {
		t.Error("4 damage should have killed the 4-toughness creature with toxic")
	}
}

func TestOneB05CompleatDevotionPumpsAndDrawsOnlyWithToxic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B")
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	toxic := oneB05ToxicCreature(g, me.ID, 2, 2)
	hand := handSize(me)
	castCatalogSpell(t, g, "Compleat Devotion", oneB05InstantType, oneB05Devotion, b16TargetCard(plain))
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, plain), effectiveToughness(t, g, plain); p != 4 || tg != 4 {
		t.Errorf("plain creature is %d/%d, want 4/4", p, tg)
	}
	if handSize(me) != hand {
		t.Errorf("drew %d cards for a creature without toxic", handSize(me)-hand)
	}
	castCatalogSpell(t, g, "Compleat Devotion", oneB05InstantType, oneB05Devotion, b16TargetCard(toxic))
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, toxic); p != 4 {
		t.Errorf("toxic creature power %d, want 4", p)
	}
	if handSize(me)-hand != 1 {
		t.Errorf("drew %d cards, want 1", handSize(me)-hand)
	}
	// Refused on an opponent's creature.
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Compleat Devotion", oneB05InstantType, oneB05Devotion, b16TargetCard(theirs)); err == nil {
		t.Error("targeted a creature the caster does not control")
	}
}

func castCatalogSpellErrStrict(t *testing.T, g *game.Game, name, typeLine, oracleID string, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: active.ID, Controller: active.ID})
	return g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets, Strict: true})
}

func TestOneB05AspirantsAscentGrantsFlyingAndToxicThatPoisons(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	toxic := oneB05ToxicCreature(g, me.ID, 2, 2)
	castCatalogSpell(t, g, "Aspirant's Ascent", oneB05InstantType, oneB05Ascent, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Aspirant's Ascent", oneB05InstantType, oneB05Ascent, b16TargetCard(toxic))
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 5 {
		t.Errorf("Bear is %d/%d, want 3/5", p, tg)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") || oneB05Toxic(t, g, bear) != 1 {
		t.Errorf("Bear flying=%v toxic=%d, want flying and toxic 1", hasEffectiveKeyword(t, g, bear, "flying"), oneB05Toxic(t, g, bear))
	}
	if got := oneB05Toxic(t, g, toxic); got != 3 {
		t.Errorf("toxic 2 creature has toxic %d under Ascent, want 3 (cumulative)", got)
	}
	attackWith(t, g, opp.ID, bear)
	if poisonOn(opp) != 1 {
		t.Errorf("opponent has %d poison, want 1", poisonOn(opp))
	}
}

func TestOneB05AspirantsAscentEndsAtEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Aspirant's Ascent", oneB05InstantType, oneB05Ascent, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, 1)
	if oneB05Toxic(t, g, bear) != 0 || hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the grant outlasted the turn")
	}
}

func TestOneB05PorcelainZealotPumpsOneOrTwoAtBeginningOfCombat(t *testing.T) {
	for _, tc := range []struct {
		name  string
		toxic bool
		want  int
	}{{"plain", false, 3}, {"toxic", true, 4}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			b12Push(g, me.ID, "Porcelain Zealot", "Creature — Phyrexian Soldier", oneB05Zealot, 2, 3)
			var c uuid.UUID
			if tc.toxic {
				c = oneB05ToxicCreature(g, me.ID, 2, 2)
			} else {
				c = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			}
			theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
			advanceTo(t, g, game.StepBeginCombat)
			b04WaitForPick(t, g, me.ID)
			p := latestPickTarget(g, me.ID)
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
				t.Fatal("targeted a creature an opponent controls")
			}
			pickCard(t, g, me.ID, c)
			passPriorityAroundTable(t, g)
			if got := effectivePower(t, g, c); got != tc.want {
				t.Errorf("power %d, want %d", got, tc.want)
			}
		})
	}
}

func TestOneB05FlensingRaptorTargetsAnotherCreatureWithToxic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	toxic := oneB05ToxicCreature(g, me.ID, 2, 2)
	raptor := castHoldCreature(t, g, "Flensing Raptor", "Creature — Phyrexian Bird", oneB05Raptor, 2, 2)
	answerTriggerTargets(t, g, me.ID, toxic)
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, toxic), effectiveToughness(t, g, toxic); p != 3 || tg != 3 {
		t.Errorf("target is %d/%d, want 3/3", p, tg)
	}
	if !hasEffectiveKeyword(t, g, toxic, "flying") {
		t.Error("the target did not gain flying")
	}
	if p := effectivePower(t, g, raptor); p != 2 {
		t.Errorf("the Raptor itself was pumped: power %d", p)
	}
	if oneB05Toxic(t, g, raptor) != 1 {
		t.Errorf("the Raptor's toxic is %d, want 1", oneB05Toxic(t, g, raptor))
	}
}

func TestOneB05FlensingRaptorHasNoTargetWithoutAnotherToxicCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Flensing Raptor", "Creature — Phyrexian Bird", oneB05Raptor, 2, 2)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, plain); p != 2 || hasEffectiveKeyword(t, g, plain, "flying") {
		t.Errorf("a creature without toxic was affected: power %d", p)
	}
}

func TestOneB05SlaughterSingerPumpsAnotherAttackingToxicCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	singer := b12Push(g, me.ID, "Slaughter Singer", oneB05SingerType, oneB05Singer, 2, 2)
	other := b12Push(g, me.ID, "Flensing Raptor", "Creature — Phyrexian Bird", oneB05Raptor, 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, singer, other, plain)
	settleOrderingTriggers(t, g)
	if p := effectivePower(t, g, other); p != 3 {
		t.Errorf("the other toxic attacker has power %d, want 3", p)
	}
	if p := effectivePower(t, g, singer); p != 2 {
		t.Errorf("Slaughter Singer pumped itself: power %d", p)
	}
	if p := effectivePower(t, g, plain); p != 2 {
		t.Errorf("an attacker without toxic was pumped: power %d", p)
	}
}

func TestOneB05ProstheticInjectorEquipsToughnessAndToxic(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	inj := b12Push(g, me.ID, "Prosthetic Injector", "Artifact — Equipment", oneB05Injector, 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, inj, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("equipped with no mana")
	}
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, inj, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear), Strict: true})
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2 || tg != 4 {
		t.Errorf("equipped Bear is %d/%d, want 2/4", p, tg)
	}
	if oneB05Toxic(t, g, bear) != 1 {
		t.Errorf("equipped Bear has toxic %d, want 1", oneB05Toxic(t, g, bear))
	}
	attackWith(t, g, opp.ID, bear)
	if poisonOn(opp) != 1 {
		t.Errorf("opponent has %d poison, want 1", poisonOn(opp))
	}
}

func TestOneB05RibskiffDrawsOnEnteringAndCarriesToxic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B")
	hand := handSize(me)
	id := castCatalogSpell(t, g, "Ribskiff", "Artifact — Vehicle", oneB05Ribskiff, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Ribskiff did not enter")
	}
	// The cast card leaves the hand and the trigger draws one.
	if handSize(me) != hand+1 {
		t.Errorf("hand %d -> %d, want the draw", hand, handSize(me))
	}
	if oneB05Toxic(t, g, id) != 2 {
		t.Errorf("Ribskiff toxic %d, want 2", oneB05Toxic(t, g, id))
	}
}

func TestOneB05RibskiffCrew3NeedsPowerThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ship := b12Push(g, me.ID, "Ribskiff", "Artifact — Vehicle", oneB05Ribskiff, 4, 4)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{small}, Strict: true}); err == nil {
		t.Fatal("crewed with total power 2")
	}
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{big}, Strict: true}); err != nil {
		t.Fatalf("crew with power 3: %v", err)
	}
	passPriorityAroundTable(t, g)
	attackWith(t, g, opp.ID, ship)
	if poisonOn(opp) != 2 || opp.Life != game.StartingLife-4 {
		t.Errorf("opponent: %d poison, %d life; want 2 and %d", poisonOn(opp), opp.Life, game.StartingLife-4)
	}
}

func TestOneB05NecrogenRotpriestAddsAPoisonCounterPerToxicCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	priest := b12Push(g, me.ID, "Necrogen Rotpriest", "Creature — Phyrexian Zombie Cleric", oneB05Rotpriest, 1, 5)
	other := oneB05ToxicCreature(g, me.ID, 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, priest, other, plain)
	passPriorityAroundTable(t, g)
	// Rotpriest: 2 toxic + 1 extra; Singer: 2 toxic + 1 extra; plain: none.
	if poisonOn(opp) != 6 {
		t.Errorf("opponent has %d poison, want 6", poisonOn(opp))
	}
}

func TestOneB05NecrogenRotpriestGrantsDeathtouchToToxicCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	priest := b12Push(g, me.ID, "Necrogen Rotpriest", "Creature — Phyrexian Zombie Cleric", oneB05Rotpriest, 1, 5)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	b06AddMana(me, "C", "B", "G")
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{
		Strict: true, Targets: b16TargetCard(plain)}); err == nil {
		t.Fatal("targeted a creature without toxic")
	}
	b16Activate(t, g, me.ID, priest, 0, game.ActivateAbilityParams{Targets: b16TargetCard(priest), Strict: true})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, priest, "deathtouch") {
		t.Error("Rotpriest did not gain deathtouch")
	}
}

func TestOneB05PlagueNurseGivesOtherToxicCreaturesToxicOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	nurse := b12Push(g, me.ID, "Plague Nurse", "Creature — Phyrexian Cleric", oneB05Nurse, 3, 4)
	singer := oneB05ToxicCreature(g, me.ID, 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	theirs := b12Push(g, opp.ID, "Slaughter Singer", oneB05SingerType, oneB05Singer, 2, 2)
	b06AddMana(me, "G", "C", "C", "G", "C", "C")
	b16Activate(t, g, me.ID, nurse, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if got := oneB05Toxic(t, g, singer); got != 3 {
		t.Errorf("the other toxic creature has toxic %d, want 3", got)
	}
	if got := oneB05Toxic(t, g, nurse); got != 2 {
		t.Errorf("Plague Nurse has toxic %d, want 2 (it says each OTHER creature)", got)
	}
	if got := oneB05Toxic(t, g, plain); got != 0 {
		t.Errorf("a creature without toxic got toxic %d", got)
	}
	if got := oneB05Toxic(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature has toxic %d, want 2", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, nurse, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a second time in one turn")
	}
	advanceToUpkeepOf(t, g, 1)
	if got := oneB05Toxic(t, g, singer); got != 2 {
		t.Errorf("toxic %d after the turn, want 2", got)
	}
}

func TestOneB05NecrogenCommunionGivesToxicAndReturnsTheCreatureUnderYourControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	// A creature I control but a friend owns would be a Threaten; keep it simple:
	// my own creature, then a kill.
	bear := b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "14c8eb6a-6a47-4a2b-b0a1-6b5a8b0f4d11", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Necrogen Communion", oneB05AuraType, oneB05Communion, b16TargetCard(theirs)); err == nil {
		t.Fatal("enchanted a creature the caster does not control")
	}
	castCatalogSpell(t, g, "Necrogen Communion", oneB05AuraType, oneB05Communion, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if oneB05Toxic(t, g, bear) != 2 {
		t.Fatalf("enchanted creature has toxic %d, want 2", oneB05Toxic(t, g, bear))
	}
	castCatalogSpell(t, g, "Hexgold Slash", oneB05InstantType, oneB05Slash, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Grizzly Bears" && c.Controller == me.ID {
				n++
			}
		}
	})
	if n != 1 {
		t.Errorf("the dead creature came back %d times under my control, want 1", n)
	}
}

func TestOneB05MazesMantlePumpsAndGivesHexproofOnlyWithToxic(t *testing.T) {
	for _, tc := range []struct {
		name     string
		toxic    bool
		hexproof bool
	}{{"toxic", true, true}, {"plain", false, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			var c uuid.UUID
			if tc.toxic {
				c = oneB05ToxicCreature(g, me.ID, 2, 2)
			} else {
				c = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			}
			castCatalogSpell(t, g, "Maze's Mantle", oneB05AuraType, oneB05Mantle, b16TargetCard(c))
			passPriorityAroundTable(t, g)
			if p, tg := effectivePower(t, g, c), effectiveToughness(t, g, c); p != 4 || tg != 4 {
				t.Errorf("enchanted creature is %d/%d, want 4/4", p, tg)
			}
			if got := hasEffectiveKeyword(t, g, c, "hexproof"); got != tc.hexproof {
				t.Errorf("hexproof = %v, want %v", got, tc.hexproof)
			}
			if tc.hexproof {
				advanceToUpkeepOf(t, g, 1)
				if hasEffectiveKeyword(t, g, c, "hexproof") {
					t.Error("hexproof outlasted the turn")
				}
				if p := effectivePower(t, g, c); p != 4 {
					t.Errorf("the Aura's +2/+2 ended with the turn: power %d", p)
				}
			}
		})
	}
}
