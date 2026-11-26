package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b295_cards_test.go — Oath of the Gatewatch devoid cards, S82 OLD
// batch 295, played through the real catalog entries.

const (
	oldB295VoidGrafter  = "01d202b9-1ac6-484c-a208-a8bc69bccdb2"
	oldB295Sinkhole     = "03fde598-5b6b-4c7f-ba13-c06ed1fa918a"
	oldB295Harvester    = "13f9a845-d0e9-4790-99e0-b857204ad1da"
	oldB295Negator      = "18dcf3db-d6c6-4780-a931-546e71a3ae3f"
	oldB295BirthingHulk = "33899675-0de2-4adb-bf3f-55e18b493015"
	oldB295Witness      = "340d2261-9454-4ab0-827c-49c7e76c7781"
	oldB295Immobilizer  = "416f9656-9247-4ef2-971b-1515fc8e031b"
	oldB295Endurance    = "63701aa6-d3b8-4af2-96a4-2dec27d6a661"
	oldB295Visions      = "66553485-4e9d-4049-9e80-3ac903f8127a"
	oldB295Translator   = "6766c5e4-baa7-4db4-a900-c08fbcd98f5d"
	oldB295ReaverDrone  = "6dd887be-484b-4ecc-8c02-8cce7e43ec89"
)

func TestOldB295CardsAreFullWithNoCaveats(t *testing.T) {
	for id, name := range map[string]string{
		oldB295VoidGrafter: "Void Grafter", oldB295Sinkhole: "Consuming Sinkhole", oldB295Harvester: "Thought Harvester",
		oldB295Negator: "Gravity Negator", oldB295BirthingHulk: "Birthing Hulk", oldB295Witness: "Witness the End",
		oldB295Immobilizer: "Immobilizer Eldrazi", oldB295Endurance: "Unnatural Endurance", oldB295Visions: "Visions of Brutality",
		oldB295Translator: "Kozilek's Translator", oldB295ReaverDrone: "Reaver Drone",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
		if !effectstest.HasKeyword(spec.PrintedKeywords, game.KeywordDevoid) {
			t.Errorf("%s does not declare devoid", name)
		}
	}
}

func oldB295PickTarget(t *testing.T, g *game.Game, me uuid.UUID, ref game.TargetRef) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestPickTarget(g, me); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me, ref); err != nil {
			t.Fatalf("pick target: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
}

func oldB295HasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	return effectstest.HasEffectiveKeyword(t, g, id, kw)
}

func TestOldB295VoidGrafterGivesAnotherCreatureHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	grafter := effectstest.CastCatalogSpell(t, g, "Void Grafter", "Creature — Eldrazi Drone", oldB295VoidGrafter, nil)
	oldB295PickTarget(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if !oldB295HasKeyword(t, g, bear, "hexproof") {
		t.Error("the other creature did not gain hexproof")
	}
	if oldB295HasKeyword(t, g, grafter, "hexproof") {
		t.Error("the Grafter gained hexproof itself")
	}
}

func TestOldB295ConsumingSinkholeModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dryad := effectstest.PushCatalogPermanent(g, opp.ID, "Dryad Arbor", "Land Creature — Forest Dryad", "", false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Consuming Sinkhole", "Instant", oldB295Sinkhole, effectstest.CardRefs(bear)); err == nil {
		t.Error("the first mode targeted a creature that is not a land")
	}
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB295Sinkhole, []int{0}, effectstest.CardRefs(dryad))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dryad) {
		t.Error("the land creature was not exiled")
	}
	before := opp.Life
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB295Sinkhole, []int{1}, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("life %d, want %d", opp.Life, before-4)
	}
}

func TestOldB295ThoughtHarvesterExilesOnlyForColorlessSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Thought Harvester", "Creature — Eldrazi Drone", oldB295Harvester, 2, 4)
	effectstest.AdvanceToMain(t, g)
	before := opp.Library.Size()
	effectstest.HandCardFull(me, "Red Spell", "Sorcery", "", "", []string{"R"})
	effectstest.CastCatalogSpell(t, g, "Colorless Rock", "Artifact", "", nil)
	oldB295PickTarget(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if got := opp.Library.Size(); got != before-1 {
		t.Errorf("library %d, want %d after a colorless spell", got, before-1)
	}
}

func TestOldB295GravityNegatorGivesFlyingForColorlessMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	negator := effectstest.B12Push(g, me.ID, "Gravity Negator", "Creature — Eldrazi Drone", oldB295Negator, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, negator)
	effectstest.FloatMana(t, g, me, "{C}") // after the step change: the pool empties between steps
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatalf("pick target: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB295HasKeyword(t, g, bear, "flying") {
		t.Error("the other creature did not gain flying after {C} was paid")
	}
}

func TestOldB295BirthingHulkMakesTwoScionsAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Birthing Hulk", "Creature — Eldrazi Drone", oldB295BirthingHulk, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Eldrazi Scion"); n != 2 {
		t.Errorf("%d Scions, want 2", n)
	}
	spec, _ := effects.Lookup(oldB295BirthingHulk)
	if len(spec.Activated) != 1 {
		t.Fatalf("%d activated abilities, want the regenerate row", len(spec.Activated))
	}
}

func TestOldB295WitnessTheEndExilesTwoFromHandAndDrainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	extra := []uuid.UUID{effectstest.HandCard(opp, "Spare A", "Sorcery"), effectstest.HandCard(opp, "Spare B", "Sorcery")}
	before, life := opp.Hand.Size(), opp.Life
	effectstest.CastCatalogSpell(t, g, "Witness the End", "Sorcery", oldB295Witness, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, extra...)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != before-2 {
		t.Errorf("hand %d, want %d", got, before-2)
	}
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
}

func TestOldB295ImmobilizerEldraziStopsTopHeavyBlockersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imm := effectstest.B12Push(g, me.ID, "Immobilizer Eldrazi", "Creature — Eldrazi Drone", oldB295Immobilizer, 2, 1)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.FloatMana(t, g, me, "{C}")
	}
	effectstest.S82Activate(t, g, me.ID, imm, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BrAttack(t, g, bear)
	effectstest.BrRefusal(t, g.Clone().DeclareBlocker(wall, bear), game.BlockReasonCantBlock)
	if err := g.DeclareBlocker(ogre, bear); err != nil {
		t.Errorf("a creature with power over toughness could not block: %v", err)
	}
}

func TestOldB295UnnaturalEnduranceGivesPowerAndAShield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Unnatural Endurance", "Instant", oldB295Endurance, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.PowerOf(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	effectstest.DestroyNow(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regeneration shield did not save the creature")
	}
}

func TestOldB295VisionsOfBrutalityBlocksAndDrainsTheHostsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Visions of Brutality", TypeLine: "Enchantment — Aura", OracleID: oldB295Visions,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, attacker)
	effectstest.BrRefusal(t, g.Clone().DeclareBlocker(host, attacker), game.BlockReasonCantBlock)
	before := opp.Life
	effectstest.DealNoncombatDamage(g, opp.ID, host, me.ID)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("the host's controller is at %d, want %d", opp.Life, before-2)
	}
}

func TestOldB295KozileksTranslatorMakesOneColorlessPerTurnForALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tr := effectstest.SeedPermanentWithOracle(g, me.ID, "Kozilek's Translator", "Creature — Eldrazi Drone", oldB295Translator)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, tr, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "C" || me.Life != life-1 {
		t.Errorf("pool %v, life %d: want one {C} for 1 life", me.ManaPool, me.Life)
	}
	if err := g.ActivateManaAbility(me.ID, tr, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the Translator's ability was used twice in one turn")
	}
}

func TestOldB295ReaverDroneCostsALifeUnlessAnotherColorlessCreature(t *testing.T) {
	for _, tc := range []struct {
		name  string
		extra func(g *game.Game, me uuid.UUID)
		want  int
	}{
		{"alone", func(*game.Game, uuid.UUID) {}, -1},
		{"beside a coloured creature", func(g *game.Game, me uuid.UUID) {
			effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear",
				Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me, Controller: me})
		}, -1},
		{"beside a colorless creature", func(g *game.Game, me uuid.UUID) {
			effectstest.B12Creature(g, me, "Golem", "Artifact Creature — Golem", 2, 2)
		}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.B12Push(g, me.ID, "Reaver Drone", "Creature — Eldrazi Drone", oldB295ReaverDrone, 2, 1)
			tc.extra(g, me.ID)
			before := me.Life
			// walk to this player's next own upkeep
			left := false
			for i := 0; i < 200; i++ {
				if _, err := g.AdvanceStep(); err != nil {
					t.Fatalf("AdvanceStep: %v", err)
				}
				if g.Turn.ActiveSeat != 0 {
					left = true
				}
				if left && g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 0 {
					break
				}
			}
			effectstest.MonSettle(t, g)
			if me.Life-before != tc.want {
				t.Errorf("life change %d, want %d", me.Life-before, tc.want)
			}
		})
	}
}
