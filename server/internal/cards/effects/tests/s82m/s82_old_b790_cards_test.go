package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b790_cards_test.go — Visions, S82 OLD batch 790 (#5631). Every test
// plays the real catalog entry through the engine.

const (
	oldB790Hearth   = "4ce95005-82b3-4921-b1c4-60d775a012a2"
	oldB790Legacy   = "9888d67d-fcad-427f-8ba7-b6d421b3d88e"
	oldB790Hope     = "a1b5c83e-cf8b-4644-8315-35387a4f491f"
	oldB790Ants     = "c39112ab-ee1f-4f00-b18b-692b5fe32b80"
	oldB790Knight   = "e9893534-3d05-42c2-93ff-f28ffae89aa3"
	oldB790Watch    = "6f5b6cba-3b57-4e6a-afa7-0cb62e5cc7d4"
	oldB790Mob      = "90e8fb2c-6473-42c3-92f4-f46a744920fa"
	oldB790Betrayal = "9c3d2e1f-a5da-47f4-aea8-5861ef7e1b90"
	oldB790Vanish   = "cd2791e8-0b61-49b8-8b72-ff2737df1049"
	oldB790Wound    = "f7b70077-5518-45f9-9833-01a4f8afac71"
)

func TestOldB790EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB790Hearth: "Hearth Charm", oldB790Legacy: "Suleiman's Legacy", oldB790Hope: "Hope Charm",
		oldB790Ants: "Army Ants", oldB790Knight: "Knight of the Mists", oldB790Watch: "Death Watch",
		oldB790Mob: "Mob Mentality", oldB790Betrayal: "Betrayal", oldB790Vanish: "Vanishing",
		oldB790Wound: "Mortal Wound",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB790TryModal casts a modal spell and returns the announcement's refusal,
// leaving a legal cast on the stack.
func oldB790TryModal(g *game.Game, name, typeLine, oracle string, modes []int, ref ...game.TargetRef) error {
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: active.ID, Controller: active.ID})
	return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: ref, Strict: true})
}

func oldB790CastMode(t *testing.T, g *game.Game, name, oracle string, mode int, ref ...game.TargetRef) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.P1g3CastModal(t, g, name, "Instant", oracle, []int{mode}, ref...)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB790HearthCharmEachMode(t *testing.T) {
	t.Run("destroys an artifact creature only", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		robot := effectstest.B12Creature(g, g.Seats[1].ID, "Robot", "Artifact Creature — Construct", 2, 2)
		bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		if err := oldB790TryModal(g, "Hearth Charm", "Instant", oldB790Hearth, []int{0}, effectstest.P1g3ModeCard(bear, 0)); err == nil {
			t.Error("Hearth Charm targeted a nonartifact creature")
		}
		oldB790CastMode(t, g, "Hearth Charm", oldB790Hearth, 0, effectstest.P1g3ModeCard(robot, 0))
		if g.Battlefield.Contains(robot) {
			t.Error("the artifact creature survived")
		}
	})
	t.Run("pumps attackers only", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		att := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
		idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
		effectstest.BrAttack(t, g, att)
		if err := oldB790TryModal(g, "Hearth Charm", "Instant", oldB790Hearth, []int{1}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p, _ := effectstest.EffectivePTOf(t, g, att); p != 3 {
			t.Errorf("attacker power %d, want 3", p)
		}
		if p, _ := effectstest.EffectivePTOf(t, g, idle); p != 2 {
			t.Errorf("non-attacker power %d, want 2", p)
		}
	})
	t.Run("can't be blocked", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		small := effectstest.B12Creature(g, g.Seats[0].ID, "Small", "Creature — Bear", 2, 2)
		big := effectstest.B12Creature(g, g.Seats[0].ID, "Big", "Creature — Giant", 3, 3)
		effectstest.AdvanceToMain(t, g)
		if err := oldB790TryModal(g, "Hearth Charm", "Instant", oldB790Hearth, []int{2}, effectstest.P1g3ModeCard(big, 0)); err == nil {
			t.Error("Hearth Charm targeted a creature with power 3")
		}
		oldB790CastMode(t, g, "Hearth Charm", oldB790Hearth, 2, effectstest.P1g3ModeCard(small, 0))
		if r := effectstest.AuraRestrictions(t, g, small); r&game.CantBeBlocked == 0 {
			t.Error("the small creature can still be blocked")
		}
	})
}

func TestOldB790HopeCharmEachMode(t *testing.T) {
	t.Run("first strike", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		bear := effectstest.B12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
		oldB790CastMode(t, g, "Hope Charm", oldB790Hope, 0, effectstest.P1g3ModeCard(bear, 0))
		if !effectstest.P1g2HasKeyword(g, bear, "first strike") {
			t.Error("no first strike")
		}
	})
	t.Run("two life", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		life := effectstest.LifeOf(g, opp.ID)
		oldB790CastMode(t, g, "Hope Charm", oldB790Hope, 1, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0})
		if got := effectstest.LifeOf(g, opp.ID); got != life+2 {
			t.Errorf("life %d → %d, want +2", life, got)
		}
	})
	t.Run("destroys an Aura only", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
		aura := effectstest.B12Permanent(g, g.Seats[1].ID, "Some Aura", "Enchantment — Aura")
		oldB790CastMode(t, g, "Hope Charm", oldB790Hope, 2, effectstest.P1g3ModeCard(aura, 0))
		if g.Battlefield.Contains(aura) {
			t.Error("the Aura survived")
		}
		if !g.Battlefield.Contains(bear) {
			t.Error("the creature went with it")
		}
	})
}

func TestOldB790SuleimansLegacyDestroysDjinnsAndEfreetsNowAndLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	djinn := effectstest.B12Creature(g, opp.ID, "Djinn", "Creature — Djinn", 3, 3)
	efreet := effectstest.B12Creature(g, me.ID, "Efreet", "Creature — Efreet", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// Entering through a cast: the enters trigger of the Legacy itself.
	effectstest.B12PlayFromHand(t, g, "Suleiman's Legacy", "Enchantment", oldB790Legacy, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{djinn, efreet} {
		if g.Battlefield.Contains(id) {
			t.Error("a Djinn or Efreet survived the Legacy entering")
		}
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear was destroyed")
	}
	late := effectstest.B12PlayFromHand(t, g, "Late Djinn", "Creature — Djinn", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(late) {
		t.Error("a Djinn entering later survived")
	}
}

func TestOldB790ArmyAntsSacrificesALandToDestroyALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ants := effectstest.PushCatalogPermanent(g, me.ID, "Army Ants", "Creature — Insect", oldB790Ants, false)
	mine := effectstest.B12Permanent(g, me.ID, "My Island", "Basic Land — Island")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	effectstest.B16Activate(t, g, me.ID, ants, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(theirs), SacrificeIDs: []uuid.UUID{mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the targeted land survived")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the land was not sacrificed")
	}
}

func TestOldB790KnightOfTheMistsPaysOrDestroysAKnight(t *testing.T) {
	enter := func(t *testing.T) (*game.Game, uuid.UUID, uuid.UUID, *game.Player) {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		other := effectstest.B12Creature(g, opp.ID, "Other Knight", "Creature — Human Knight", 2, 2)
		knight := effectstest.B12PlayFromHand(t, g, "Knight of the Mists", "Creature — Human Knight", oldB790Knight, game.CastSpellParams{})
		if pick := effectstest.PassUntilPickTarget(t, g, me.ID); pick != nil {
			if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: other}); err != nil {
				t.Fatalf("naming the target Knight: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		return g, knight, other, me
	}
	t.Run("declined", func(t *testing.T) {
		g, knight, other, me := enter(t)
		prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
		if err := g.ResolvePayUnless(prompt.ID, me.ID, false); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(other) {
			t.Error("the target Knight survived the declined payment")
		}
		if !g.Battlefield.Contains(knight) {
			t.Error("the Knight of the Mists itself was destroyed")
		}
	})
	t.Run("paid", func(t *testing.T) {
		g, _, other, me := enter(t)
		prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
		effectstest.B06AddMana(me, "U")
		if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(other) {
			t.Error("the target Knight was destroyed although {U} was paid")
		}
	})
}

func TestOldB790DeathWatchDrainsByPowerAndGainsByToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := effectstest.B12Creature(g, opp.ID, "Brute", "Creature — Ogre", 3, 5)
	effectstest.AuraCast(t, g, "Death Watch", oldB790Watch, big)
	myLife, theirLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(big); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != theirLife-3 {
		t.Errorf("the creature's controller: %d → %d, want -3 (its power)", theirLife, got)
	}
	if got := effectstest.LifeOf(g, me.ID); got != myLife+5 {
		t.Errorf("the Aura's controller: %d → %d, want +5 (its toughness)", myLife, got)
	}
}

func TestOldB790MobMentalityGrantsTrampleAndPumpsWhenEveryNonWallAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AuraCast(t, g, "Mob Mentality", oldB790Mob, host)
	if !effectstest.P1g2HasKeyword(g, host, "trample") {
		t.Error("the enchanted creature has no trample")
	}
	effectstest.BrAttack(t, g, host, other)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, host); p != 4 {
		t.Errorf("host power %d, want 4 (+2 for two attackers; the Wall stays home)", p)
	}
	_ = wall
}

func TestOldB790MobMentalityDoesNothingWhenACreatureStaysHome(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Stays Home", "Creature — Bear", 2, 2)
	effectstest.AuraCast(t, g, "Mob Mentality", oldB790Mob, host)
	effectstest.BrAttack(t, g, host)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, host); p != 2 {
		t.Errorf("host power %d, want 2", p)
	}
}

func TestOldB790BetrayalDrawsWhenTheEnchantedCreatureTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AuraSeedLibrary(me, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Betrayal", effectstest.AuraTypeLine, oldB790Betrayal, effectstest.CardRefs(mine)); err == nil {
		t.Error("Betrayal enchanted a creature its controller controls")
	}
	effectstest.AuraCast(t, g, "Betrayal", oldB790Betrayal, theirs)
	hand := me.Hand.Size()
	if err := g.TapCard(theirs, true); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, got)
	}
}

func TestOldB790VanishingPhasesTheEnchantedCreatureOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.AuraCast(t, g, "Vanishing", oldB790Vanish, bear)
	effectstest.B06AddMana(me, "U", "U")
	effectstest.B16Activate(t, g, me.ID, aura, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PhasedOutInCatalogGame(g, bear) {
		t.Error("the enchanted creature did not phase out")
	}
}

func TestOldB790MortalWoundDestroysTheCreatureThatIsDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 4, 4)
	src := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.AuraCast(t, g, "Mortal Wound", oldB790Wound, bear)
	effectstest.Damage(t, g, src, bear, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the damaged creature survived Mortal Wound")
	}
}
