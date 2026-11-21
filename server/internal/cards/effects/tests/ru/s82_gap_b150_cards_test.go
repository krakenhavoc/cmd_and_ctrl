package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b150_cards_test.go — GAP batch 150, Jumpstart 2022 creatures
// (#5948). Every card is played through the engine.

const (
	gap150ShadowbornDemonOracle     = "7dacc5aa-ea4d-414e-8324-6a7d24bb7671"
	gap150SinuousVerminOracle       = "658f5cfb-326d-4da9-a39b-2ecd393a1f54"
	gap150SkullslitherWormOracle    = "a36dda2e-4cb8-4821-9c93-bd637ae7b333"
	gap150SoulcageFiendOracle       = "ecbb660a-4013-410a-b683-82c3ef1690cd"
	gap150TakenumaBleederOracle     = "ee397a87-6933-4d2d-aac6-8dc4e96d524b"
	gap150TerminationFacilitatorOra = "56edd306-8676-49c4-b3c3-5ff840536c3a"
	gap150TivashOracle              = "b5641fcf-135e-43e7-a547-344fac61e279"
)

func TestGap150EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		gap150ShadowbornDemonOracle, gap150SinuousVerminOracle, gap150SkullslitherWormOracle,
		gap150SoulcageFiendOracle, gap150TakenumaBleederOracle, gap150TerminationFacilitatorOra,
		gap150TivashOracle,
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

func TestGap150ShadowbornDemonDestroysANonDemonCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := gap150Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	demon := effectstest.B12Creature(g, opp.ID, "Other Demon", "Creature — Demon", 2, 2)
	effectstest.CastHoldCreature(t, g, "Shadowborn Demon", "Creature — Demon", gap150ShadowbornDemonOracle, 5, 6)
	p := effectstest.LatestPickTarget(g, g.Seats[g.Turn.ActiveSeat].ID)
	if p != nil {
		for _, id := range p.PickTargetCards {
			if id == demon {
				t.Fatal("a Demon was offered as a target")
			}
		}
		effectstest.PickCard(t, g, p.Chooser, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the non-Demon creature survived")
	}
	if !g.Battlefield.Contains(demon) {
		t.Error("the Demon was destroyed")
	}
}

func TestGap150ShadowbornDemonUpkeepSacrificeNeedsFewerThanSix(t *testing.T) {
	for _, tc := range []struct {
		inGraveyard int
		sacrifices  bool
	}{{5, true}, {6, false}} {
		g := effectstest.NewCatalogGame(t)
		me, _ := gap150Seats(g)
		demon := effectstest.B12Push(g, me.ID, "Shadowborn Demon", "Creature — Demon", gap150ShadowbornDemonOracle, 5, 6)
		for i := 0; i < tc.inGraveyard; i++ {
			effectstest.PushGraveyardCardTyped(me, "Dead", "Creature — Bear")
		}
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID, Step: game.StepUpkeep}) })
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSacrifice(t, g, me.ID, demon)
		effectstest.PassPriorityAroundTable(t, g)
		if gone := !g.Battlefield.Contains(demon); gone != tc.sacrifices {
			t.Errorf("%d creature cards: Demon sacrificed = %v, want %v", tc.inGraveyard, gone, tc.sacrifices)
		}
	}
}

func TestGap150SinuousVerminMonstrosityGivesMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := gap150Seats(g)
	vermin := effectstest.B12Push(g, me.ID, "Sinuous Vermin", "Creature — Rat Horror", gap150SinuousVerminOracle, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, vermin, "menace") {
		t.Fatal("menace before it is monstrous")
	}
	effectstest.ApaMana(me, "B", "B", "C", "C", "C")
	effectstest.ApaActivate(t, g, me, vermin, 0, game.ActivateAbilityParams{})
	if got := effectstest.CountersOn(g, vermin, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters, want 3", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, vermin, "menace") {
		t.Error("no menace once monstrous")
	}
}

func TestGap150SkullslitherWormDiscardsOrGrows(t *testing.T) {
	// Every opponent has a card: each discards, the Worm stays 3/3.
	g := effectstest.NewCatalogGame(t)
	me, _ := gap150Seats(g)
	var opps []*game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID {
			opps = append(opps, s)
			effectstest.D2dSetHand(s, 2)
		}
	}
	worm := effectstest.CastHoldCreature(t, g, "Skullslither Worm", "Creature — Worm", gap150SkullslitherWormOracle, 3, 3)
	for _, o := range opps {
		for i := 0; i < 4 && effectstest.DiscardChoiceFor(g, o.ID) != nil; i++ {
			effectstest.AnswerDiscard(t, g, o.ID, o.Hand.Cards[0].InstanceID)
		}
		if o.Hand.Size() != 1 {
			t.Errorf("%s holds %d cards, want 1", o.Name, o.Hand.Size())
		}
	}
	if got := effectstest.CountersOn(g, worm, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters though every opponent discarded", got)
	}

	// Every opponent has an empty hand: two counters for each.
	g = effectstest.NewCatalogGame(t)
	me, _ = gap150Seats(g)
	n := 0
	for _, s := range g.Seats {
		if s.ID != me.ID {
			effectstest.D2dSetHand(s, 0)
			n++
		}
	}
	worm = effectstest.CastHoldCreature(t, g, "Skullslither Worm", "Creature — Worm", gap150SkullslitherWormOracle, 3, 3)
	if got := effectstest.CountersOn(g, worm, game.CounterPlusOne); got != 2*n {
		t.Errorf("%d counters with %d empty-handed opponents, want %d", got, n, 2*n)
	}
}

func TestGap150SoulcageFiendDrainsEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap150Seats(g)
	fiend := effectstest.B12Push(g, me.ID, "Soulcage Fiend", "Creature — Demon", gap150SoulcageFiendOracle, 3, 2)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.B27Kill(g, fiend)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if s.Life != lives[s.ID]-3 {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, lives[s.ID]-3)
		}
	}
}

func TestGap150TakenumaBleederCostsLifeUnlessYouControlADemon(t *testing.T) {
	for _, demon := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bleeder := effectstest.B12Push(g, me.ID, "Takenuma Bleeder", "Creature — Ogre Shaman", gap150TakenumaBleederOracle, 3, 3)
		if demon {
			effectstest.B12Creature(g, me.ID, "A Demon", "Creature — Demon", 1, 1)
		}
		life := me.Life
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(bleeder, opp.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		want := life - 1
		if demon {
			want = life
		}
		if me.Life != want {
			t.Errorf("demon=%v: life %d, want %d", demon, me.Life, want)
		}
	}
}

func TestGap150TakenumaBleederCostsLifeWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bleeder := effectstest.B12Push(g, opp.ID, "Takenuma Bleeder", "Creature — Ogre Shaman", gap150TakenumaBleederOracle, 3, 3)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(bleeder, attacker); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("blocker's life %d, want %d", opp.Life, life-1)
	}
}

func TestGap150TerminationFacilitatorBountyThenDamageDestroys(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := gap150Seats(g)
	fac := effectstest.B12Push(g, me.ID, "Termination Facilitator", "Creature — Human Assassin", gap150TerminationFacilitatorOra, 1, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 4, 4)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 4, 4)
	// Sorcery speed only.
	effectstest.ApaActivate(t, g, me, fac, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}})
	if got := effectstest.CountersOn(g, theirs, "bounty"); got != 1 {
		t.Fatalf("%d bounty counters, want 1", got)
	}
	// An opponent's creature with no counter is unaffected by damage.
	effectstest.B27Damage(g, mine, other, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(other) {
		t.Error("a creature without a bounty counter was destroyed")
	}
	effectstest.B27Damage(g, mine, theirs, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the creature with a bounty counter survived being dealt damage")
	}
}

func TestGap150TerminationFacilitatorOnlyAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	me, opp := gap150Seats(g)
	fac := effectstest.B12Push(g, me.ID, "Termination Facilitator", "Creature — Human Assassin", gap150TerminationFacilitatorOra, 1, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 4, 4)
	if err := g.ActivateCatalogAbility(me.ID, fac, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Error("activated outside a main phase")
	}
}

func TestGap150TivashPaysLifeForADemonOfThatSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap150Seats(g)
	tivash := effectstest.B12Push(g, me.ID, "Tivash, Gloom Summoner", "Legendary Creature — Human Warlock", gap150TivashOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(tivash, me.ID, 3) })
	life := me.Life
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginEndStep, Actor: me.ID, Step: game.StepEnd}) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Demon" && c.IsToken() && c.Controller == me.ID {
			found = true
			if c.Power != 3 || c.Toughness != 3 {
				t.Errorf("the Demon is %d/%d, want 3/3", c.Power, c.Toughness)
			}
			if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Error("the Demon has no flying")
			}
		}
	}
	if !found {
		t.Error("no Demon token")
	}
}

func TestGap150TivashDoesNothingWithoutLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap150Seats(g)
	effectstest.B12Push(g, me.ID, "Tivash, Gloom Summoner", "Legendary Creature — Human Warlock", gap150TivashOracle, 4, 4)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginEndStep, Actor: me.ID, Step: game.StepEnd}) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("asked about a Demon with no life gained")
	}
}

// gap150Seats is the active seat and the next one.
func gap150Seats(g *game.Game) (me, opp *game.Player) {
	seat := g.Turn.ActiveSeat
	return g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
}
