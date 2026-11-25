package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b255_cards_test.go — Aether Revolt, batch 255 (S82 older sets,
// #5536). Every test plays the real catalog entry through the engine.

const (
	oldB255ArchangelOracle    = "472efb69-47c8-4c93-b60a-f6dba4364b8c"
	oldB255SimulacrumOracle   = "dd886509-7455-4c8f-976f-0f5fafcb97be"
	oldB255TechnicianOracle   = "47a66451-8a4e-4fc6-8685-96e1feb8f997"
	oldB255AlleyEvasionOracle = "b59c7c62-ce2f-4062-9790-566191aded69"
	oldB255CrackdownOracle    = "128f7e5a-17e0-4069-a185-c9f874c7d234"
	oldB255ProwlerOracle      = "da18d97c-85f3-4395-9af5-bca72e33ebfc"
)

func TestOldB255EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB255ArchangelOracle, oldB255SimulacrumOracle, oldB255TechnicianOracle,
		oldB255AlleyEvasionOracle, oldB255CrackdownOracle, oldB255ProwlerOracle,
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

// oldB255Settle lets an enters trigger go on the stack, picks its target
// and answers its "may", then resolves everything.
func oldB255Settle(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID, yes bool) {
	t.Helper()
	for i := 0; i < 12; i++ {
		effectstest.PassPriorityUntilChoice(t, g)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			effectstest.AnswerPickTarget(t, g, target)
			continue
		}
		if p := effectstest.LatestTriggerPromptFor(g, me.ID); p != nil {
			if err := g.ResolveTriggerPrompt(p.ID, me.ID, yes); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			continue
		}
		break
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB255ConsulateCrackdownExilesOpposingArtifactsUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mine := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "My Rock", TypeLine: "Artifact"})
	theirs := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Consulate Crackdown", "Enchantment", oldB255CrackdownOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, theirs) != game.ZoneExile {
		t.Errorf("their artifact is in %v, want exile", effectstest.ZoneOf(g, theirs))
	}
	if effectstest.ZoneOf(g, mine) != game.ZoneBattlefield || effectstest.ZoneOf(g, bear) != game.ZoneBattlefield {
		t.Error("my artifact or their non-artifact creature was exiled")
	}
	var crackdown uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Consulate Crackdown" {
			crackdown = c.InstanceID
		}
	}
	if crackdown == uuid.Nil {
		t.Fatal("Consulate Crackdown is not on the battlefield")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(crackdown); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(oldB255OnBattlefield(g, opp.ID, "Their Rock")); got != 1 {
		t.Errorf("their artifact came back %d times, want once, under its owner", got)
	}
}

func oldB255OnBattlefield(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func TestOldB255AlleyEvasionGivesPlusOnePlusTwoOrReturnsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB255AlleyEvasionOracle, []int{0}, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 4 {
		t.Errorf("the Bear is %d/%d, want 3/4", p, tt)
	}
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB255AlleyEvasionOracle, []int{1}, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("the Bear is still on the battlefield")
	}
	if got := len(me.Hand.Cards); got == 0 {
		t.Error("the Bear did not return to hand")
	}
}

func TestOldB255ExquisiteArchangelReplacesTheLossWithStartingLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	angel := effectstest.B12Push(g, me.ID, "Exquisite Archangel", "Creature — Angel", oldB255ArchangelOracle, 5, 5)
	g.WithWriteLock(func() { me.Life = -3 })
	g.RunStateChecksForTest()
	if me.Eliminated {
		t.Fatal("the player lost the game")
	}
	if me.Life != g.Settings.StartingLife {
		t.Errorf("life %d, want the starting life total %d", me.Life, g.Settings.StartingLife)
	}
	if _, ok := effectstest.BattlefieldCard(g, angel); ok {
		t.Error("the Archangel is still on the battlefield")
	}
	g.WithWriteLock(func() { me.Life = 0 })
	g.RunStateChecksForTest()
	if !me.Eliminated {
		t.Error("the player survived a second loss with no Archangel")
	}
}

func TestOldB255DispersalTechnicianMayReturnAnArtifact(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		rock := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
		effectstest.CastCatalogSpell(t, g, "Dispersal Technician", "Creature — Vedalken Artificer", oldB255TechnicianOracle, nil)
		oldB255Settle(t, g, me, rock, yes)
		_, onBattlefield := effectstest.BattlefieldCard(g, rock)
		if yes == onBattlefield || (yes && !opp.Hand.Contains(rock)) {
			t.Errorf("yes=%v: on battlefield %v, in their hand %v", yes, onBattlefield, opp.Hand.Contains(rock))
		}
	}
}

func TestOldB255FourthBridgeProwlerMayShrinkACreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
		effectstest.CastCatalogSpell(t, g, "Fourth Bridge Prowler", "Creature — Human Rogue", oldB255ProwlerOracle, nil)
		oldB255Settle(t, g, me, bear, yes)
		want := 3
		if yes {
			want = 2
		}
		if p, tt := effectstest.PtOf(t, g, bear); p != want || tt != want {
			t.Errorf("yes=%v: the Bear is %d/%d, want %d/%d", yes, p, tt, want, want)
		}
	}
}

func TestOldB255TezzeretsSimulacrumDrainsOneOrThreeWithATezzeret(t *testing.T) {
	for _, tezzeret := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		effectstest.AdvanceToMain(t, g)
		sim := effectstest.B12Push(g, me.ID, "Tezzeret's Simulacrum", "Artifact Creature — Golem", oldB255SimulacrumOracle, 2, 3)
		if tezzeret {
			effectstest.PushWalkerForTest(g, me.ID, "Tezzeret Test", "", 4)
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].Name == "Tezzeret Test" {
						g.Battlefield.Cards[i].TypeLine = "Legendary Planeswalker — Tezzeret"
					}
				}
			})
		}
		life := opp.Life
		if err := g.ActivateCatalogAbility(me.ID, sim, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := 1
		if tezzeret {
			want = 3
		}
		if opp.Life != life-want {
			t.Errorf("tezzeret=%v: life %d -> %d, want %d lost", tezzeret, life, opp.Life, want)
		}
	}
}
