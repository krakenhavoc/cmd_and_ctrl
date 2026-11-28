package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b808_test.go — Mirage cards with activated abilities, S82 OLD batch
// 808, played through the real catalog entries.

const (
	oldB808Raging   = "137e0ad2-e10a-4327-86a3-5ac7bc7fc469"
	oldB808Mist     = "448f5c55-7334-4fa4-9d90-0e57bae19444"
	oldB808Locust   = "9963b6a9-d715-430e-9b5e-576fb688d944"
	oldB808Canopy   = "a9e59ee4-02a6-498c-82f0-f4f2ff334d65"
	oldB808Elder    = "adea6cb4-1e74-4137-9c80-1dde004a6b26"
	oldB808Chariot  = "b0af249d-9255-42f5-bafc-ac61ee281930"
	oldB808Efreet   = "c21d3360-8be0-477e-a592-7e64a76570d3"
	oldB808Wave     = "d8eed5f7-8724-45c7-b6fc-dcfcc400f09a"
	oldB808Nomad    = "db130ff4-7525-421f-b7e7-200921ecbdec"
	oldB808Wildcats = "dd25afe0-a69a-4e0b-b364-b88d3a522ed3"
	oldB808Healer   = "e2117373-6f0a-41c8-ab6a-9014b9628e64"
	oldB808Gnomes   = "ea016b84-7f77-42e7-bd14-2e62a1e35f3b"
	oldB808Walker   = "f1768dd5-c5c8-4059-9695-55b8a11eeffd"
	oldB808Totem    = "fa798a3d-7797-4795-b251-a45f0429d169"
)

func TestOldB808CardsAreFullWithNoCaveats(t *testing.T) {
	for id, name := range map[string]string{
		oldB808Raging: "Raging Spirit", oldB808Mist: "Mist Dragon", oldB808Locust: "Locust Swarm",
		oldB808Canopy: "Canopy Dragon", oldB808Elder: "Village Elder", oldB808Chariot: "Chariot of the Sun",
		oldB808Efreet: "Frenetic Efreet", oldB808Wave: "Wave Elemental", oldB808Nomad: "Dwarven Nomad",
		oldB808Wildcats: "Uktabi Wildcats", oldB808Healer: "Femeref Healer", oldB808Gnomes: "Ersatz Gnomes",
		oldB808Walker: "Unseen Walker", oldB808Totem: "Grinning Totem",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB808Perm seeds a ready-to-use catalog permanent for `owner`.
func oldB808Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	id := effectstest.PushCatalogPermanent(g, owner, name, typeLine, oracle, false)
	g.WithWriteLock(func() {
		if c := effectstest.FindBattlefieldCardForTest(g, id); c != nil {
			c.Power, c.Toughness = p, tough
		}
	})
	return id
}

func oldB808Activate(g *game.Game, me *game.Player, id uuid.UUID, idx int, params game.ActivateAbilityParams) error {
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, id, idx, params)
}

func oldB808Colored(g *game.Game, owner uuid.UUID, name string, p, tough int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Colors: colors,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func TestOldB808RagingSpiritBecomesColorlessUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	spirit := oldB808Perm(g, me.ID, "Raging Spirit", "Creature — Spirit", oldB808Raging, 3, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, spirit).Colors = []string{"R"} })
	if got := effectstest.EffectiveColors(t, g, spirit); len(got) != 1 {
		t.Fatalf("colours before = %v, want red", got)
	}
	if err := oldB808Activate(g, me, spirit, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without paying {2}")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := oldB808Activate(g, me, spirit, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveColors(t, g, spirit); len(got) != 0 {
		t.Errorf("colours = %v, want colorless", got)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if got := effectstest.EffectiveColors(t, g, spirit); len(got) != 1 {
		t.Errorf("colours next turn = %v, want red again", got)
	}
}

func TestOldB808MistDragonGainsAndLosesFlyingIndefinitelyAndPhasesOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dragon := oldB808Perm(g, me.ID, "Mist Dragon", "Creature — Dragon", oldB808Mist, 4, 4)
	if effectstest.HasEffectiveKeyword(t, g, dragon, "flying") {
		t.Fatal("flies before it gains flying")
	}
	for i, step := range []struct {
		idx  int
		want bool
	}{{0, true}, {1, false}, {0, true}} {
		if err := oldB808Activate(g, me, dragon, step.idx, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HasEffectiveKeyword(t, g, dragon, "flying"); got != step.want {
			t.Fatalf("step %d: flying = %v, want %v", i, got, step.want)
		}
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("the gained flying did not last past the turn")
	}
	effectstest.FloatMana(t, g, me, "{U}{U}{C}{C}{C}")
	if err := oldB808Activate(g, me, dragon, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("phase out: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PhasedOutInCatalogGame(g, dragon) {
		t.Error("the Dragon did not phase out")
	}
}

func TestOldB808LocustSwarmRegeneratesAndUntapsOnlyOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	swarm := oldB808Perm(g, me.ID, "Locust Swarm", "Creature — Insect", oldB808Locust, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, swarm, "flying") {
		t.Error("no flying")
	}
	effectstest.FloatMana(t, g, me, "{G}")
	if err := oldB808Activate(g, me, swarm, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B27Kill(g, swarm)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(swarm) {
		t.Fatal("the regeneration shield did not save the Swarm")
	}
	effectstest.TapCard(g, swarm)
	effectstest.FloatMana(t, g, me, "{G}{G}")
	if err := oldB808Activate(g, me, swarm, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("untap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedNow(g, swarm) {
		t.Error("the Swarm is still tapped")
	}
	effectstest.TapCard(g, swarm)
	if err := oldB808Activate(g, me, swarm, 1, game.ActivateAbilityParams{}); err == nil {
		t.Error("untapped a second time in one turn")
	}
}

func TestOldB808CanopyDragonSwapsTrampleForFlyingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dragon := oldB808Perm(g, me.ID, "Canopy Dragon", "Creature — Dragon", oldB808Canopy, 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "trample") || effectstest.HasEffectiveKeyword(t, g, dragon, "flying") {
		t.Fatal("wrong starting keywords")
	}
	effectstest.FloatMana(t, g, me, "{C}{G}")
	if err := oldB808Activate(g, me, dragon, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "flying") || effectstest.HasEffectiveKeyword(t, g, dragon, "trample") {
		t.Error("did not gain flying and lose trample")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, dragon, "flying") || !effectstest.HasEffectiveKeyword(t, g, dragon, "trample") {
		t.Error("the swap lasted past the turn")
	}
}

func TestOldB808VillageElderSacrificesAForestToRegenerateATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	elder := oldB808Perm(g, me.ID, "Village Elder", "Creature — Human Druid", oldB808Elder, 1, 1)
	bear := effectstest.SeedBear(g, me.ID)
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	swamp := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.FloatMana(t, g, me, "{G}")
	if err := oldB808Activate(g, me, elder, 0, game.ActivateAbilityParams{
		Targets: effectstest.RtCardTarget(bear), SacrificeIDs: []uuid.UUID{swamp},
	}); err == nil {
		t.Fatal("a Swamp paid for the Forest sacrifice")
	}
	if err := oldB808Activate(g, me, elder, 0, game.ActivateAbilityParams{
		Targets: effectstest.RtCardTarget(bear), SacrificeIDs: []uuid.UUID{forest},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) {
		t.Error("the Forest was not sacrificed")
	}
	effectstest.B27Kill(g, bear)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(bear) {
		t.Error("the targeted creature was not regenerated")
	}
}

func TestOldB808ChariotOfTheSunGivesFlyingAndBaseToughnessOneToYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	chariot := oldB808Perm(g, me.ID, "Chariot of the Sun", "Artifact", oldB808Chariot, 0, 0)
	mine := oldB808Colored(g, me.ID, "Big Bear", 4, 6, "G")
	theirs := oldB808Colored(g, opp.ID, "Their Bear", 2, 6, "G")
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := oldB808Activate(g, me, chariot, 0, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(theirs)}); err == nil {
		t.Fatal("targeted a creature an opponent controls")
	}
	if err := oldB808Activate(g, me, chariot, 0, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(mine)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "flying") {
		t.Error("no flying")
	}
	if got := effectstest.EffectiveToughness(t, g, mine); got != 1 {
		t.Errorf("toughness = %d, want 1", got)
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 4 {
		t.Errorf("power = %d, want its own 4", got)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, mine, "flying") || effectstest.EffectiveToughness(t, g, mine) != 6 {
		t.Error("the effect lasted past the turn")
	}
}

func TestOldB808FreneticEfreetPhasesOutOnAWinAndIsSacrificedOnALoss(t *testing.T) {
	sawWin, sawLoss := false, false
	for table := 0; table < 80 && !(sawWin && sawLoss); table++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		efreet := oldB808Perm(g, me.ID, "Frenetic Efreet", "Creature — Efreet", oldB808Efreet, 2, 1)
		if !effectstest.HasEffectiveKeyword(t, g, efreet, "flying") {
			t.Fatal("no flying")
		}
		if err := oldB808Activate(g, me, efreet, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		flips := len(effectstest.RandomEvents(g, game.EventFlipCoin))
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) != flips+1 {
			t.Fatalf("%d coins flipped, want 1", len(evs)-flips)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if evs[len(evs)-1].Won {
			sawWin = true
			if !effectstest.PhasedOutInCatalogGame(g, efreet) {
				t.Error("a won flip did not phase the Efreet out")
			}
			if effectstest.GraveyardHas(me, efreet) {
				t.Error("a won flip sacrificed the Efreet")
			}
		} else {
			sawLoss = true
			if g.Battlefield.Contains(efreet) || !effectstest.GraveyardHas(me, efreet) {
				t.Error("a lost flip did not sacrifice the Efreet")
			}
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("never saw both outcomes (win %v, loss %v)", sawWin, sawLoss)
	}
}

func TestOldB808WaveElementalSacrificesToTapUpToThreeCreaturesWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wave := oldB808Perm(g, me.ID, "Wave Elemental", "Creature — Elemental", oldB808Wave, 2, 3)
	a := effectstest.SeedBear(g, opp.ID)
	b := effectstest.SeedBear(g, opp.ID)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.FloatMana(t, g, me, "{U}")
	if err := oldB808Activate(g, me, wave, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: flier}},
	}); err == nil {
		t.Fatal("targeted a creature with flying")
	}
	effectstest.FloatMana(t, g, me, "{U}")
	if err := oldB808Activate(g, me, wave, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wave) {
		t.Error("the Elemental was not sacrificed")
	}
	if !effectstest.TappedNow(g, a) || !effectstest.TappedNow(g, b) {
		t.Error("the targets were not tapped")
	}
	if effectstest.TappedNow(g, flier) {
		t.Error("the flier was tapped")
	}
}

func TestOldB808DwarvenNomadMakesAPowerTwoOrLessCreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	nomad := oldB808Perm(g, me.ID, "Dwarven Nomad", "Creature — Dwarf Nomad", oldB808Nomad, 1, 1)
	big := oldB808Colored(g, me.ID, "Big Bear", 3, 3)
	small := oldB808Colored(g, me.ID, "Small Bear", 2, 2)
	blocker := oldB808Colored(g, opp.ID, "Blocker", 2, 2)
	if err := oldB808Activate(g, me, nomad, 0, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(big)}); err == nil {
		t.Fatal("targeted a creature with power 3")
	}
	if err := oldB808Activate(g, me, nomad, 0, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(small)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(small, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, small); err == nil {
		t.Error("an unblockable creature was blocked")
	}
}

func TestOldB808UktabiWildcatsAreAsBigAsYourForestsAndRegenerateBySacrificingOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cats := oldB808Perm(g, me.ID, "Uktabi Wildcats", "Creature — Cat", oldB808Wildcats, 0, 0)
	f1 := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad")
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if p, tough := effectstest.EffectivePower(t, g, cats), effectstest.EffectiveToughness(t, g, cats); p != 3 || tough != 3 {
		t.Fatalf("size = %d/%d, want 3/3 (my three Forests)", p, tough)
	}
	effectstest.FloatMana(t, g, me, "{G}")
	if err := oldB808Activate(g, me, cats, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{f1}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, cats); p != 2 {
		t.Errorf("power after the sacrifice = %d, want 2", p)
	}
	effectstest.B27Kill(g, cats)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(cats) {
		t.Error("the Wildcats were not regenerated")
	}
}

func TestOldB808FemerefHealerPreventsTheNextOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	healer := oldB808Perm(g, me.ID, "Femeref Healer", "Creature — Human Cleric", oldB808Healer, 1, 1)
	bear := oldB808Colored(g, me.ID, "Bear", 2, 9)
	src := oldB808Colored(g, opp.ID, "Pinger", 1, 1)
	if err := oldB808Activate(g, me, healer, 0, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B27Damage(g, src, bear, 3)
	if got := effectstest.ApaLive(g, bear).DamageMarked; got != 2 {
		t.Errorf("damage marked = %d, want 2 (one prevented)", got)
	}
	effectstest.B27Damage(g, src, bear, 3)
	if got := effectstest.ApaLive(g, bear).DamageMarked; got != 5 {
		t.Errorf("damage marked = %d after a second hit, want 5 (the shield is spent)", got)
	}
}

func TestOldB808ErsatzGnomesMakeAPermanentColorlessUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gnomes := oldB808Perm(g, me.ID, "Ersatz Gnomes", "Artifact Creature — Gnome", oldB808Gnomes, 1, 1)
	bear := oldB808Colored(g, opp.ID, "Bear", 2, 2, "G")
	if err := oldB808Activate(g, me, gnomes, 1, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveColors(t, g, bear); len(got) != 0 {
		t.Errorf("colours = %v, want none", got)
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if got := effectstest.EffectiveColors(t, g, bear); len(got) != 1 {
		t.Errorf("colours next turn = %v, want green again", got)
	}
}

func TestOldB808ErsatzGnomesMakeASpellColorlessAndTheCreatureItBecomesStaysSo(t *testing.T) {
	for _, withGnomes := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		me := g.Seats[0]
		gnomes := oldB808Perm(g, me.ID, "Ersatz Gnomes", "Artifact Creature — Gnome", oldB808Gnomes, 1, 1)
		spell := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"},
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		if withGnomes {
			if err := oldB808Activate(g, me, gnomes, 0, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(spell)}); err != nil {
				t.Fatalf("activate on the spell: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(spell) {
			t.Fatal("the creature spell did not resolve")
		}
		got := effectstest.EffectiveColors(t, g, spell)
		if withGnomes && len(got) != 0 {
			t.Errorf("the creature it became is %v, want colorless (CR 400.7a)", got)
		}
		if !withGnomes && len(got) != 1 {
			t.Errorf("without the Gnomes the creature is %v, want green", got)
		}
		if withGnomes {
			effectstest.AdvanceToMainOf(t, g, 1)
			if got := effectstest.EffectiveColors(t, g, spell); len(got) != 0 {
				t.Errorf("next turn the creature is %v, want still colorless (no duration)", got)
			}
		}
	}
}

func TestOldB808UnseenWalkerGivesAnyCreatureForestwalkUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	walker := oldB808Perm(g, me.ID, "Unseen Walker", "Creature — Dryad", oldB808Walker, 1, 1)
	other := oldB808Colored(g, me.ID, "Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, walker, "forestwalk") || effectstest.HasEffectiveKeyword(t, g, other, "forestwalk") {
		t.Fatal("wrong starting keywords")
	}
	effectstest.FloatMana(t, g, me, "{C}{G}{G}")
	if err := oldB808Activate(g, me, walker, 0, game.ActivateAbilityParams{Targets: effectstest.RtCardTarget(other)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, other, "forestwalk") {
		t.Error("the target did not gain forestwalk")
	}
	effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, other, "forestwalk") {
		t.Error("forestwalk lasted past the turn")
	}
}

func TestOldB808GrinningTotemExilesACardForYouToPlayUntilYourNextUpkeepThenBinsIt(t *testing.T) {
	for _, play := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.SeedLibrary(opp, "Their One", "Their Two", "Their Three")
		effectstest.AdvanceToMain(t, g)
		totem := oldB808Perm(g, me.ID, "Grinning Totem", "Artifact", oldB808Totem, 0, 0)
		libBefore := opp.Library.Size()
		effectstest.FloatMana(t, g, me, "{C}{C}")
		if err := oldB808Activate(g, me, totem, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSearchNamed(t, g, me.ID, "Their Two")
		if g.Battlefield.Contains(totem) {
			t.Error("the Totem was not sacrificed")
		}
		if opp.Library.Size() != libBefore-1 {
			t.Fatalf("library %d, want %d", opp.Library.Size(), libBefore-1)
		}
		var exiled uuid.UUID
		for _, c := range g.Exile.Cards {
			if c.Owner == opp.ID {
				exiled = c.InstanceID
			}
		}
		if exiled == uuid.Nil {
			t.Fatal("nothing was exiled from their library")
		}
		if play {
			if err := effectstest.CastFromExile(t, g, me.ID, exiled); err != nil {
				t.Fatalf("could not play the exiled card: %v", err)
			}
		}
		if !play {
			effectstest.AdvanceToUpkeepOf(t, g, 0)
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.GraveyardHas(opp, exiled) {
				t.Error("the unplayed card was not put into its owner's graveyard")
			}
		}
	}
}
