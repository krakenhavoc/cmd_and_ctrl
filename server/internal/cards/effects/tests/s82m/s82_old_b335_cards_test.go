package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o335Aspirant  = "334178f6-cf4b-4ea6-a24f-c7375d1b38de"
	o335TailSlash = "69c51904-f758-4853-a86c-317c04462ea2"
	o335Smasher   = "3425598f-d51c-4f92-8fda-42e96e4c537a"
	o335Rig       = "51c7beca-ca19-4a4d-9512-0255146a9cd0"
	o335Behemoth  = "636e9863-29ac-476a-86db-d8f73b4adbf4"
	o335Rager     = "a7036d08-6ee4-49b1-bbbf-c96a937afa15"
	o335Scorn     = "1fb3b5a3-ca4b-4cbf-ab58-71c960f3efd0"
	o335Dunecast  = "5ac44e0a-782f-49ee-b263-2fb65359faaa"
	o335Keeper    = "84b28100-ad70-4a78-b557-99f5ef6126a3"
)

func TestO335CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		o335Aspirant: "Kolaghan Aspirant", o335TailSlash: "Tail Slash", o335Smasher: "Gate Smasher",
		o335Rig: "Stormrider Rig", o335Behemoth: "Qal Sisma Behemoth", o335Rager: "Blood-Chin Rager",
		o335Scorn: "Silumgar's Scorn", o335Dunecast: "Dromoka Dunecaster", o335Keeper: "Keeper of the Lens",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func o335Card(id uuid.UUID) []game.TargetRef { return effectstest.CardRefs(id) }

func TestO335AspirantPingsEachCreatureThatBlocksIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	asp := effectstest.B12Push(g, me.ID, "Kolaghan Aspirant", "Creature — Human Warrior", o335Aspirant, 2, 1)
	chump := effectstest.B12Creature(g, opp.ID, "Chump", "Creature — Bear", 1, 1)
	sturdy := effectstest.B12Creature(g, opp.ID, "Sturdy", "Creature — Bear", 3, 3)
	effectstest.BrAttack(t, g, asp)
	for _, b := range []uuid.UUID{chump, sturdy} {
		if err := g.DeclareBlocker(b, asp); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(chump) {
		t.Error("the 1/1 blocker survived the Aspirant's trigger")
	}
	if !g.Battlefield.Contains(sturdy) {
		t.Error("the 3/3 blocker died to a 1-damage trigger")
	}
}

func TestO335TailSlashIsOneSided(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 3, 3)
	mine2 := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 3)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Tail Slash", "Instant", o335TailSlash, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: mine2, Slot: 1},
	}); err == nil {
		t.Fatal("Tail Slash took its caster's own creature as the second target")
	}
	effectstest.CastCatalogSpell(t, g, "Tail Slash", "Instant", o335TailSlash, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 3-toughness creature survived 3 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the biting creature took damage back")
	}
}

func TestO335GateSmasherEquipsOnlyToughnessFourOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	smasher := effectstest.PushCatalogPermanent(g, me.ID, "Gate Smasher", "Artifact — Equipment", o335Smasher, false)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 3)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 2, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, smasher, 0, game.ActivateAbilityParams{Strict: true, Targets: o335Card(small)}); err == nil {
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.AttachedTo(t, g, smasher, small) {
			t.Fatal("Gate Smasher attached to a creature with toughness 3")
		}
	}
	effectstest.B06AddMana(me, "R", "R", "R")
	effectstest.B16Activate(t, g, me.ID, smasher, 0, game.ActivateAbilityParams{Strict: true, Targets: o335Card(big)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, smasher, big) {
		t.Fatal("Gate Smasher did not attach to a creature with toughness 4")
	}
	if p := effectstest.EffectivePower(t, g, big); p != 5 {
		t.Errorf("the equipped creature has power %d, want 5", p)
	}
	effectstest.AssertKeywords(t, g, big, "trample")
}

func TestO335StormriderRigMayAttachToACreatureThatEnters(t *testing.T) {
	for _, attach := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		rig := effectstest.PushCatalogPermanent(g, me.ID, "Stormrider Rig", "Artifact — Equipment", o335Rig, false)
		asp := effectstest.CastCatalogSpell(t, g, "Kolaghan Aspirant", "Creature — Human Warrior", o335Aspirant, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, attach)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.AttachedTo(t, g, rig, asp); got != attach {
			t.Errorf("attach=%v: attached %v", attach, got)
		}
	}
}

func TestO335BehemothNeedsTwoToAttackAndToBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := effectstest.B12Push(g, me.ID, "Qal Sisma Behemoth", "Creature — Ogre Warrior", o335Behemoth, 5, 5)
	blocker := effectstest.B12Push(g, opp.ID, "Qal Sisma Behemoth", "Creature — Ogre Warrior", o335Behemoth, 5, 5)
	attacker := effectstest.B12Creature(g, me.ID, "Plain Attacker", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(beast, opp.ID); err == nil {
		t.Fatal("the Behemoth attacked without paying {2}")
	}
	effectstest.B06AddMana(me, "R", "R")
	if err := g.DeclareAttacker(beast, opp.ID); err != nil {
		t.Fatalf("the Behemoth could not attack with {2} available: %v", err)
	}
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("an untaxed creature could not attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("staging the block: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); !errors.Is(err, game.ErrBlockTaxUnpaid) {
		t.Fatalf("finishing the block with no mana: %v, want ErrBlockTaxUnpaid", err)
	}
	effectstest.B06AddMana(opp, "R", "R")
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Errorf("the Behemoth could not block with {2} available: %v", err)
	}
}

func TestO335RagerGivesEveryWarriorYouControlMenaceWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rager := effectstest.B12Push(g, me.ID, "Blood-Chin Rager", "Creature — Human Warrior", o335Rager, 2, 2)
	warrior := effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Warrior", "Creature — Human Warrior", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, rager)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, rager, "menace")
	effectstest.AssertKeywords(t, g, warrior, "menace")
	for _, id := range []uuid.UUID{bear, theirs} {
		if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), "menace") {
			t.Error("menace reached a non-Warrior or an opponent's creature")
		}
	}
}

func TestO335ScornCountersOnlyUnlessPaidWithoutADragon(t *testing.T) {
	for _, tc := range []string{"declined", "paid", "controlled", "revealed"} {
		t.Run(tc, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.ToMainForCost(t, g)
			victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
			params := game.CastSpellParams{Targets: o335Card(victim)}
			switch tc {
			case "controlled":
				effectstest.B12Creature(g, opp.ID, "Dragon", "Creature — Dragon", 4, 4)
			case "revealed":
				dragon := effectstest.PutInHand(opp, game.Card{Name: "Hand Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4})
				params.OptionalCosts, params.RevealIDs = []int{0}, []uuid.UUID{dragon}
			}
			scorn := effectstest.PutInHand(opp, game.Card{Name: "Silumgar's Scorn", TypeLine: "Instant", OracleID: o335Scorn})
			if err := g.CastSpell(opp.ID, scorn, params); err != nil {
				t.Fatalf("cast: %v", err)
			}
			if tc == "declined" || tc == "paid" {
				effectstest.B06AddMana(me, "U")
				effectstest.PassUntilTaxed(t, g, me.ID)
				effectstest.AnswerPayUnless(t, g, me.ID, tc == "paid")
			} else {
				effectstest.PassPriorityAroundTable(t, g)
				if effectstest.HasPayUnlessFor(g, me.ID) {
					t.Fatal("a payment was offered although a Dragon was revealed or controlled")
				}
			}
			effectstest.PassPriorityAroundTable(t, g)
			countered := effectstest.CountEvents(g, game.EventCounterSpell) > 0
			if want := tc != "paid"; countered != want {
				t.Errorf("%s: countered=%v, want %v", tc, countered, want)
			}
		})
	}
}

func TestO335DunecasterTapsAGroundCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dune := effectstest.PushCatalogPermanent(g, me.ID, "Dromoka Dunecaster", "Creature — Human Wizard", o335Dunecast, false)
	ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 2)
	flier := effectstest.BrCreature(g, opp.ID, "Flier", "Creature — Bird", 1, 1, "flying")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, dune, 0, game.ActivateAbilityParams{Strict: true, Targets: o335Card(flier)}); err == nil {
		t.Fatal("the Dunecaster targeted a creature with flying")
	}
	effectstest.B16Activate(t, g, me.ID, dune, 0, game.ActivateAbilityParams{Strict: true, Targets: o335Card(ground)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, ground) {
		t.Error("the ground creature stayed untapped")
	}
}

func TestO335KeeperLooksAtOpposingFaceDownCreaturesWhileItStays(t *testing.T) {
	spec, ok := effects.Lookup(o335Keeper)
	if !ok || !spec.LooksAtFaceDownCreatures {
		t.Fatal("Keeper of the Lens does not declare the standing look at face-down creatures")
	}
}
