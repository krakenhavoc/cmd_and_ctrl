package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b326_test.go — Dragons of Tarkir, batch 326 (#5551): the life gain
// and life loss cards. Each test plays the real catalog entry through the
// engine and asserts what its text says.

const (
	oldB326Student = "0adcf8d2-6eb9-4679-8ad5-f28e2f4aba39"
	oldB326Wander  = "06d63395-6d05-44d8-bae6-0155f8026e57"
	oldB326Glee    = "bd637245-52bc-4cc9-810d-253e69dc3394"
	oldB326Shriek  = "a4dc1fea-427c-424c-99c8-af2e6d3ad2a9"
	oldB326Invoke  = "fbf27d1b-2b2a-49f3-abcd-20a6c79c44ee"
	oldB326Wound   = "28956db0-985c-402d-abae-1075c68b97e7"
	oldB326Fanatic = "b704f231-f64c-49fc-84f6-0fc53081bcd0"
)

func TestOldB326CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB326Student: "Student of Ojutai", oldB326Wander: "Deadly Wanderings", oldB326Glee: "Butcher's Glee",
		oldB326Shriek: "Foul-Tongue Shriek", oldB326Invoke: "Foul-Tongue Invocation",
		oldB326Wound: "Self-Inflicted Wound", oldB326Fanatic: "Blood-Chin Fanatic",
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

func TestOldB326StudentOfOjutaiGainsTwoPerNoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Student of Ojutai", "Creature — Human Monk", oldB326Student, 2, 4)
	start := me.Life
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.SettleProwess(t, g)
	if me.Life != start {
		t.Errorf("a creature spell gained %d life", me.Life-start)
	}
	effectstest.CastNoncreature(t, g)
	effectstest.SettleProwess(t, g)
	if me.Life != start+2 {
		t.Errorf("life %d after a noncreature spell, want %d", me.Life, start+2)
	}
}

func TestOldB326DeadlyWanderingsNeedsExactlyOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Deadly Wanderings", "Enchantment", oldB326Wander, 0, 0)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if p, tou := effectstest.PtOf(t, g, theirs); p != 2 || tou != 2 {
		t.Errorf("an opponent's lone creature is %d/%d, want 2/2", p, tou)
	}
	only := effectstest.B12Creature(g, me.ID, "Only", "Creature — Bear", 2, 2)
	if p, tou := effectstest.PtOf(t, g, only); p != 4 || tou != 2 {
		t.Errorf("the lone creature is %d/%d, want 4/2", p, tou)
	}
	for _, kw := range []string{"deathtouch", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, only, kw) {
			t.Errorf("the lone creature has no %s", kw)
		}
	}
	second := effectstest.B12Creature(g, me.ID, "Second", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{only, second} {
		if p, _ := effectstest.PtOf(t, g, id); p != 2 {
			t.Errorf("with two creatures a Bear has %d power, want 2", p)
		}
		if effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") || effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
			t.Error("with two creatures a Bear still has deathtouch or lifelink")
		}
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(second) })
	if p, _ := effectstest.PtOf(t, g, only); p != 4 {
		t.Errorf("back to one creature it has %d power, want 4", p)
	}
}

func TestOldB326ButchersGleePumpsGrantsLifelinkAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Butcher's Glee", "Instant", oldB326Glee, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, bear); p != 5 || tou != 2 {
		t.Errorf("the Bear is %d/%d, want 5/2", p, tou)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the Bear has no lifelink")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, bear, 5) })
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(bear) {
		t.Error("the regeneration shield did not save the Bear")
	}
}

func TestOldB326GleeNeedsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Butcher's Glee", "Instant", oldB326Glee, effectstest.CardTarget(rock)); err == nil {
		t.Error("Butcher's Glee accepted a noncreature target")
	}
}

func TestOldB326FoulTongueShriekDrainsOnePerAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Home", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.ArnSetCombat(g, a, opp.ID, uuid.Nil)
	effectstest.ArnSetCombat(g, b, opp.ID, uuid.Nil)
	effectstest.ArnSetCombat(g, theirs, me.ID, uuid.Nil)
	mine, theirLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Foul-Tongue Shriek", "Instant", oldB326Shriek, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 || me.Life != mine+2 {
		t.Errorf("life: me %d (want %d), opponent %d (want %d)", me.Life, mine+2, opp.Life, theirLife-2)
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Foul-Tongue Shriek", "Instant", oldB326Shriek, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Error("Foul-Tongue Shriek targeted its own caster")
	}
}

func TestOldB326FoulTongueShriekWithNoAttackersDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine, theirLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Foul-Tongue Shriek", "Instant", oldB326Shriek, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife || me.Life != mine {
		t.Errorf("life changed with no attackers: me %d, opponent %d", me.Life, opp.Life)
	}
}

func oldB326Invocation(t *testing.T, g *game.Game, me *game.Player, target, victim uuid.UUID, params game.CastSpellParams) {
	t.Helper()
	card := game.Card{InstanceID: uuid.New(), Name: "Foul-Tongue Invocation", TypeLine: "Instant", OracleID: oldB326Invoke, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	params.Targets = []game.TargetRef{{Kind: game.TargetPlayer, ID: target}}
	if err := g.CastSpell(me.ID, card.InstanceID, params); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, target, victim)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB326FoulTongueInvocationGainsFourOnlyWithADragon(t *testing.T) {
	for _, tc := range []string{"nothing", "controlled", "revealed"} {
		t.Run(tc, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 1, 1)
			params := game.CastSpellParams{}
			switch tc {
			case "controlled":
				effectstest.B12Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4)
			case "revealed":
				dragon := effectstest.PutInHand(me, game.Card{Name: "Hand Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4})
				params = game.CastSpellParams{OptionalCosts: []int{0}, RevealIDs: []uuid.UUID{dragon}}
			}
			life := me.Life
			oldB326Invocation(t, g, me, opp.ID, victim, params)
			if g.Battlefield.Contains(victim) {
				t.Error("the opponent's only creature was not sacrificed")
			}
			want := life
			if tc != "nothing" {
				want = life + 4
			}
			if me.Life != want {
				t.Errorf("%s: life %d, want %d", tc, me.Life, want)
			}
		})
	}
}

// oldB326Colored is a 1/1 creature of one colour.
func oldB326Colored(g *game.Game, owner uuid.UUID, name, typeLine, color string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: []string{color},
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

func TestOldB326SelfInflictedWoundCostsAGreenOrWhiteCreatureAndTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	green := oldB326Colored(g, opp.ID, "Elf", "Creature — Elf", "G")
	red := oldB326Colored(g, opp.ID, "Imp", "Creature — Imp", "R")
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Self-Inflicted Wound", "Sorcery", oldB326Wound, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID) != nil {
		effectstest.AnswerOwnPermanents(t, g, opp.ID, green)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(green) {
		t.Error("the green creature was not sacrificed")
	}
	if !g.Battlefield.Contains(red) {
		t.Error("the red creature was sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
}

func TestOldB326SelfInflictedWoundCostsNothingWithoutAGreenOrWhiteCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	red := oldB326Colored(g, opp.ID, "Imp", "Creature — Imp", "R")
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Self-Inflicted Wound", "Sorcery", oldB326Wound, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(red) || opp.Life != life {
		t.Errorf("nothing should happen: red creature present %v, life %d (was %d)", g.Battlefield.Contains(red), opp.Life, life)
	}
}

func TestOldB326BloodChinFanaticDrainsForASacrificedWarriorsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fan := effectstest.B12Push(g, me.ID, "Blood-Chin Fanatic", "Creature — Orc Warrior", oldB326Fanatic, 3, 3)
	warrior := effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 4, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	tgt := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	effectstest.B06AddMana(me, "B", "C")
	if err := g.ActivateCatalogAbility(me.ID, fan, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}, Targets: tgt}); err == nil {
		t.Fatal("sacrificed a non-Warrior")
	}
	effectstest.B06AddMana(me, "B", "C")
	if err := g.ActivateCatalogAbility(me.ID, fan, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fan}, Targets: tgt}); err == nil {
		t.Fatal("sacrificed itself, which is not another Warrior")
	}
	mine, theirs := me.Life, opp.Life
	effectstest.B06AddMana(me, "B", "C")
	effectstest.B16Activate(t, g, me.ID, fan, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{warrior}, Targets: tgt})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-4 || me.Life != mine+4 {
		t.Errorf("life: me %d (want %d), opponent %d (want %d)", me.Life, mine+4, opp.Life, theirs-4)
	}
	if g.Battlefield.Contains(warrior) {
		t.Error("the Warrior was not sacrificed")
	}
}
