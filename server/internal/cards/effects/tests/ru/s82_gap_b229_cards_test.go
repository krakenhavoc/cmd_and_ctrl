package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander Legends: Battle for Baldur's Gate, GAP batch 229 (S82, #5948):
// the instants.

const (
	gapDelayedBlastFireballOracle = "99915b5a-5092-4900-8824-6de2aab1b2c1"
	gapEarthTremorOracle          = "1dd7d11b-e3f3-4886-b412-0c085187869c"
	gapPackAttackOracle           = "84b385cd-9dfd-4513-af3c-20fc1cc8fea0"
	gapStormKingsThunderOracle    = "f5c5f462-10fa-422a-90da-07f46257f22f"
	gapTwoHandedAxeOracle         = "b012cec9-2aff-4dfb-8db1-61e830c33926"
	gapWildMagicSurgeOracle       = "41d16975-d6c0-4343-bea2-4f360ead9a1c"
	gapBanesContingencyOracle     = "8089c3eb-99a8-40ca-882a-c1d4e30075ee"
	gapDomineeringWillOracle      = "2c438ecb-b127-4450-9395-2d1eea294f30"
	gapElminstersSimulacrumOracle = "3bb32408-81d3-4d0c-8674-1a1c5e3feb3b"
	gapGalesRedirectionOracle     = "8676f007-bf5f-4312-968c-59f4549f7d51"
	gapDemonBoltOracle            = "6e786f0b-e2cb-40ee-96bc-81945a0b35a3"
)

func TestS82GapB229CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapDelayedBlastFireballOracle: "Delayed Blast Fireball",
		gapEarthTremorOracle:          "Earth Tremor",
		gapPackAttackOracle:           "Pack Attack",
		gapStormKingsThunderOracle:    "Storm King's Thunder",
		gapTwoHandedAxeOracle:         "Two-Handed Axe",
		gapTwoHandedAxeOracle + "#1":  "Sweeping Cleave",
		gapWildMagicSurgeOracle:       "Wild Magic Surge",
		gapBanesContingencyOracle:     "Bane's Contingency",
		gapDomineeringWillOracle:      "Domineering Will",
		gapElminstersSimulacrumOracle: "Elminster's Simulacrum",
		gapGalesRedirectionOracle:     "Gale's Redirection",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// gapCast puts the card in the active seat's hand under id and casts it for
// free, strictly.
func gapCast(t *testing.T, g *game.Game, id uuid.UUID, name, typeLine, oracle string, params game.CastSpellParams) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.D5681Spell(g, me, id, name, typeLine, "", oracle)
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}

func TestGapDelayedBlastFireballDealsTwoFromHandAndFiveFromExile(t *testing.T) {
	for _, fromExile := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp, other := effectstest.BlbB1Seats(g)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 9)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 9)
		third := effectstest.B12Creature(g, other.ID, "Third", "Creature — Bear", 1, 9)
		want := 2
		id := uuid.New()
		if fromExile {
			want = 5
			c := game.Card{InstanceID: id, Name: "Delayed Blast Fireball", TypeLine: "Instant", OracleID: gapDelayedBlastFireballOracle, Owner: me.ID, Controller: me.ID}
			g.WithWriteLock(func() {
				g.Exile.PushTop(c)
				g.GrantCastPermissionToCardsForEffect(game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Cost: "{0}", Duration: g.UntilEndOfTurnDuration()}, []game.Card{c})
			})
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
				t.Fatalf("cast from exile: %v", err)
			}
		} else if err := gapCast(t, g, id, "Delayed Blast Fireball", "Instant", gapDelayedBlastFireballOracle, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		before := [2]int{opp.Life, other.Life}
		myLife := me.Life
		effectstest.PassPriorityAroundTable(t, g)
		if got := before[0] - opp.Life; got != want {
			t.Errorf("fromExile=%v: first opponent lost %d, want %d", fromExile, got, want)
		}
		if got := before[1] - other.Life; got != want {
			t.Errorf("fromExile=%v: second opponent lost %d, want %d", fromExile, got, want)
		}
		if me.Life != myLife {
			t.Errorf("fromExile=%v: the caster lost life", fromExile)
		}
		if got := effectstest.FindCardAnywhere(t, g, theirs).DamageMarked; got != want {
			t.Errorf("fromExile=%v: an opponent's creature has %d damage, want %d", fromExile, got, want)
		}
		if got := effectstest.FindCardAnywhere(t, g, third).DamageMarked; got != want {
			t.Errorf("fromExile=%v: another opponent's creature has %d damage, want %d", fromExile, got, want)
		}
		if got := effectstest.FindCardAnywhere(t, g, mine).DamageMarked; got != 0 {
			t.Errorf("fromExile=%v: the caster's creature has %d damage", fromExile, got)
		}
	}
}

func TestGapDelayedBlastFireballForetellsFor4RR(t *testing.T) {
	sa := effectstest.ForetellOf(gapDelayedBlastFireballOracle)
	if sa == nil || sa.CastCost != "{4}{R}{R}" || sa.Cost != game.ForetellExileCost {
		t.Fatalf("foretell = %+v", sa)
	}
}

func TestGapEarthTremorDealsDamageEqualToYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := gapCast(t, g, uuid.New(), "Earth Tremor", "Instant", gapEarthTremorOracle, game.CastSpellParams{Targets: effectstest.CardTarget(rock)}); err == nil {
		t.Fatal("Earth Tremor targeted an artifact")
	}
	if err := gapCast(t, g, uuid.New(), "Earth Tremor", "Instant", gapEarthTremorOracle, game.CastSpellParams{Targets: effectstest.CardTarget(wall)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, wall).DamageMarked; got != 3 {
		t.Errorf("damage = %d, want 3 (the caster's lands only)", got)
	}
}

func TestGapPackAttackPumpsByThePlayersBeingAttackedAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, tc := range []struct{ id, def uuid.UUID }{{a, opp.ID}, {b, opp.ID}, {c, other.ID}} {
		if err := g.DeclareAttacker(tc.id, tc.def); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	hand := me.Hand.Size()
	if err := gapCast(t, g, uuid.New(), "Pack Attack", "Instant", gapPackAttackOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if got := effectstest.EffectivePower(t, g, id); got != 4 {
			t.Errorf("an attacker has power %d, want 2+2 (two players attacked)", got)
		}
	}
	if got := effectstest.EffectivePower(t, g, idle); got != 2 {
		t.Errorf("a creature that is not attacking has power %d, want 2", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d (a card drawn)", got, hand+1)
	}
}

func TestGapPackAttackWithNoAttackersOnlyDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	if err := gapCast(t, g, uuid.New(), "Pack Attack", "Instant", gapPackAttackOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("power = %d, want 2", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
}

func TestGapTwoHandedAxeDoublesTheEquippedAttackersPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	axe := effectstest.SeedEquipment(g, me.ID, "Two-Handed Axe", gapTwoHandedAxeOracle)
	effectstest.EquipTo(t, g, me.ID, axe, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, axe, bear) {
		t.Fatal("the Axe is not attached")
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 6 {
		t.Errorf("power = %d, want 6", got)
	}
}

func TestGapTwoHandedAxeEquipCostsOneRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	axe := effectstest.SeedEquipment(g, me.ID, "Two-Handed Axe", gapTwoHandedAxeOracle)
	if err := effectstest.StrictEquip(t, g, me.ID, axe, bear, "{G}{G}"); err == nil {
		t.Fatal("equipped without red mana")
	}
}

func TestGapSweepingCleaveGivesDoubleStrikeToYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := gapCast(t, g, uuid.New(), "Sweeping Cleave", "Instant — Adventure", gapTwoHandedAxeOracle+"#1", game.CastSpellParams{Targets: effectstest.CardTarget(theirs)}); err == nil {
		t.Fatal("Sweeping Cleave targeted an opponent's creature")
	}
	if err := gapCast(t, g, uuid.New(), "Sweeping Cleave", "Instant — Adventure", gapTwoHandedAxeOracle+"#1", game.CastSpellParams{Targets: effectstest.CardTarget(mine)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "double strike") {
		t.Error("no double strike")
	}
}

func TestGapWildMagicSurgeDestroysAndPutsASharedTypeCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Artifact Creature — Golem", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := gapCast(t, g, uuid.New(), "Wild Magic Surge", "Instant", gapWildMagicSurgeOracle, game.CastSpellParams{Targets: effectstest.CardTarget(mine)}); err == nil {
		t.Fatal("Wild Magic Surge targeted the caster's own permanent")
	}
	landID, spellID, bearID := uuid.New(), uuid.New(), uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: bearID, Name: "Library Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	opp.Library.PushTop(game.Card{InstanceID: landID, Name: "Library Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	opp.Library.PushTop(game.Card{InstanceID: spellID, Name: "Library Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	if err := gapCast(t, g, uuid.New(), "Wild Magic Surge", "Instant", gapWildMagicSurgeOracle, game.CastSpellParams{Targets: effectstest.CardTarget(victim)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(victim) {
		t.Error("the target was not destroyed")
	}
	if c := effectstest.FindBattlefieldCardByID(g, bearID); c == nil || c.Controller != opp.ID {
		t.Fatalf("the creature card did not enter under its owner's control: %+v", c)
	}
	if effectstest.OnBattlefieldByID(g, landID) {
		t.Error("the land (not a shared type) entered")
	}
	for _, id := range []uuid.UUID{landID, spellID} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
			t.Errorf("a revealed card is not on the bottom of the library: %v", z)
		}
	}
}

func TestGapWildMagicSurgeSharesAnyOfTheTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	_, opp, _ := effectstest.BlbB1Seats(g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Artifact Creature — Golem", 2, 2)
	artifactID, creatureID := uuid.New(), uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: creatureID, Name: "Deep Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	opp.Library.PushTop(game.Card{InstanceID: artifactID, Name: "Library Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	if err := gapCast(t, g, uuid.New(), "Wild Magic Surge", "Instant", gapWildMagicSurgeOracle, game.CastSpellParams{Targets: effectstest.CardTarget(victim)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, artifactID) {
		t.Error("an artifact card does not share a type with an artifact creature")
	}
	if effectstest.OnBattlefieldByID(g, creatureID) {
		t.Error("the reveal went on past the first shared-type card")
	}
}

func TestGapBanesContingencyCountersAndOnlyScriesForYourCommander(t *testing.T) {
	for _, commander := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		leader := effectstest.B12Creature(g, me.ID, "Leader", "Legendary Creature — Bear", 2, 9)
		g.WithWriteLock(func() {
			if c := effectstest.FindBattlefieldCardByID(g, leader); c != nil {
				c.IsCommander = commander
			}
		})
		for i := 0; i < 3; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		}
		bolt := uuid.New()
		if err := gapCast(t, g, bolt, "Demon Bolt", "Instant", gapDemonBoltOracle, game.CastSpellParams{Targets: effectstest.CardTarget(leader), HoldPriority: true}); err != nil {
			t.Fatalf("cast bolt: %v", err)
		}
		hand := me.Hand.Size()
		if err := gapCast(t, g, uuid.New(), "Bane's Contingency", "Instant", gapBanesContingencyOracle, game.CastSpellParams{Targets: effectstest.CardTarget(bolt)}); err != nil {
			t.Fatalf("cast Bane's Contingency: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
		if commander {
			if scry == nil || len(scry.ScryCards) != 2 {
				t.Fatalf("commander: scry prompt %+v, want 2 cards", scry)
			}
			effectstest.AnswerScryKeepAll(t, g, me.ID)
			effectstest.PassPriorityAroundTable(t, g)
		} else if scry != nil {
			t.Fatalf("not a commander: scry prompt %+v", scry)
		}
		if got := effectstest.FindCardAnywhere(t, g, leader).DamageMarked; got != 0 {
			t.Errorf("commander=%v: the bolt was not countered (%d damage)", commander, got)
		}
		wantHand := hand
		if commander {
			wantHand++
		}
		if got := me.Hand.Size(); got != wantHand {
			t.Errorf("commander=%v: hand = %d, want %d", commander, got, wantHand)
		}
	}
}

func TestGapDomineeringWillStealsUntapsAndForcesBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, other := effectstest.BlbB1Seats(g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, other.ID, "B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if c := effectstest.FindBattlefieldCardByID(g, a); c != nil {
			c.Tapped = true
		}
	})
	targets := []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
		{Kind: game.TargetCard, ID: a, Slot: 1},
		{Kind: game.TargetCard, ID: b, Slot: 1},
	}
	if err := gapCast(t, g, uuid.New(), "Domineering Will", "Instant", gapDomineeringWillOracle, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c := effectstest.LayeredCard(t, g, id)
		if c.Controller != me.ID {
			t.Errorf("%s is controlled by %v, want the caster", c.Name, c.Controller)
		}
		if c.Tapped {
			t.Errorf("%s is tapped", c.Name)
		}
		if !effectstest.MkmB1MustBlock(t, g, id) {
			t.Errorf("%s has no blocks-if-able requirement", c.Name)
		}
	}
}

func TestGapDomineeringWillCannotTakeAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	targets := []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
		{Kind: game.TargetCard, ID: attacker, Slot: 1},
	}
	if err := gapCast(t, g, uuid.New(), "Domineering Will", "Instant", gapDomineeringWillOracle, game.CastSpellParams{Targets: targets}); err == nil {
		t.Fatal("Domineering Will targeted an attacking creature")
	}
}

func TestGapElminstersSimulacrumCopiesOneCreaturePerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var picks []game.TargetRef
	for i := 1; i <= 3; i++ {
		p := g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)]
		id := effectstest.B12Creature(g, p.ID, "Ogre", "Creature — Ogre", i+1, i+1)
		picks = append(picks, game.TargetRef{Kind: game.TargetCard, ID: id, Slot: i - 1})
	}
	if err := gapCast(t, g, uuid.New(), "Elminster's Simulacrum", "Instant", gapElminstersSimulacrumOracle, game.CastSpellParams{Targets: picks[:2]}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	copies := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Ogre" && c.IsToken() {
			copies++
		}
	}
	if copies != 2 {
		t.Errorf("tokens = %d, want 2 (up to one per opponent)", copies)
	}
}

func TestGapStormKingsThunderCopiesTheNextInstantXTimes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	storm := uuid.New()
	effectstest.D5681Spell(g, me, storm, "Storm King's Thunder", "Instant", "{X}{R}{R}{R}", gapStormKingsThunderOracle)
	if err := g.CastSpell(me.ID, storm, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want 1", len(g.DelayedTriggers))
	}
	before := effectstest.LifeOf(g, opp.ID)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, opp.ID, 2)
	if got := effectstest.LifeOf(g, opp.ID); got != before-9 {
		t.Errorf("opponent life = %d, want %d (the Bolt and its two copies)", got, before-9)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("the trigger is still waiting after the first instant")
	}
}

func TestGapStormKingsThunderIgnoresACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	storm := uuid.New()
	effectstest.D5681Spell(g, me, storm, "Storm King's Thunder", "Instant", "{X}{R}{R}{R}", gapStormKingsThunderOracle)
	if err := g.CastSpell(me.ID, storm, game.CastSpellParams{XValue: 1}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	bear := uuid.New()
	effectstest.D5681Spell(g, me, bear, "Bear", "Creature — Bear", "", "")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast bear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("a creature spell was copied (%d copies)", n)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Errorf("delayed triggers = %d, want 1 (still waiting for an instant or sorcery)", len(g.DelayedTriggers))
	}
}

func TestGapGalesRedirectionExilesTheSpellAndGrantsACastByTheRoll(t *testing.T) {
	for _, tc := range []struct {
		natural int
		free    bool
	}{{1, false}, {5, false}, {12, true}, {20, true}} {
		gale := uuid.New()
		g := effectstest.D5681Table(t, gale, tc.natural)
		me, opp, _ := effectstest.BlbB1Seats(g)
		wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 20)
		for i := 0; i < 4; i++ {
			effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
		}
		pack := uuid.New()
		effectstest.D5681Spell(g, me, pack, "Earth Tremor", "Instant", "{3}{R}", gapEarthTremorOracle)
		if err := g.CastSpell(me.ID, pack, game.CastSpellParams{Strict: false, HoldPriority: true, Targets: effectstest.CardTarget(wall)}); err != nil {
			t.Fatalf("cast Earth Tremor: %v", err)
		}
		if err := gapCastWith(t, g, gale, "Gale's Redirection", "Instant", gapGalesRedirectionOracle, effectstest.CardTarget(pack)); err != nil {
			t.Fatalf("cast Gale's Redirection: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if z := g.FindCardZoneForEffect(pack); z == nil || z.Kind != game.ZoneExile {
			t.Fatalf("natural %d: the spell is in %v, want exile", tc.natural, z)
		}
		if got := effectstest.FindCardAnywhere(t, g, wall).DamageMarked; got != 0 {
			t.Fatalf("natural %d: the exiled spell resolved", tc.natural)
		}
		// The sum is the natural die plus the mana value (4), so 12+4 and 20+4 are 15+.
		err := g.CastSpell(me.ID, pack, game.CastSpellParams{Strict: true, FromZone: "exile", Targets: effectstest.CardTarget(wall)})
		if tc.free {
			if err != nil {
				t.Errorf("natural %d: a {0} cast was refused: %v", tc.natural, err)
			}
			continue
		}
		if err == nil {
			t.Errorf("natural %d: cast for free on a result under 15", tc.natural)
		}
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}{G}{G}"); err != nil {
			t.Fatalf("mana: %v", err)
		}
		if err := g.CastSpell(me.ID, pack, game.CastSpellParams{Strict: true, FromZone: "exile", Targets: effectstest.CardTarget(wall)}); err != nil {
			t.Errorf("natural %d: any-color mana was refused: %v", tc.natural, err)
		}
	}
}

// gapCastWith is gapCast for a spell with targets.
func gapCastWith(t *testing.T, g *game.Game, id uuid.UUID, name, typeLine, oracle string, targets []game.TargetRef) error {
	t.Helper()
	return gapCast(t, g, id, name, typeLine, oracle, game.CastSpellParams{Targets: targets})
}
