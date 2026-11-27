package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b778_test.go — Weatherlight, batch 778 (#5629): enchantments, pump and
// combat tricks. Each test plays the real catalog entry.

const (
	oldB778Ground   = "3b92c4dc-2605-42f4-a374-c3be4de3279e"
	oldB778Mists    = "b18b8415-a0b1-459d-8c6b-896381184935"
	oldB778Veil     = "c6db2ea4-954c-443b-9603-c41596555c31"
	oldB778Renewal  = "f86bcdde-8b1f-4efa-a01d-39e09dc826d0"
	oldB778Foliage  = "fb1369aa-25a2-4e2a-9cbe-6cfe7e238b1b"
	oldB778Banner   = "79560d80-1745-4773-a5c4-4c8e98475135"
	oldB778Avizoa   = "91520db8-2d50-4c17-9f0e-2f92f63285f7"
	oldB778Necratog = "9f820f1e-3cd1-4a3e-8d00-194c8a80b296"
	oldB778Forge    = "cb4414a6-27d0-4922-938b-391283535f24"
	oldB778Dwarf    = "cc43507a-ca61-45ae-ac0a-4ddb851c27ef"

	oldB778Bear = "Creature — Bear"
)

func TestOldB778EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB778Ground: "Familiar Ground", oldB778Mists: "Pendrell Mists", oldB778Veil: "Teferi's Veil",
		oldB778Renewal: "Angelic Renewal", oldB778Foliage: "Dense Foliage", oldB778Banner: "Jabari's Banner",
		oldB778Avizoa: "Avizoa", oldB778Necratog: "Necratog", oldB778Forge: "Thran Forge",
		oldB778Dwarf: "Dwarven Berserker",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v; want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB778HasType(t *testing.T, g *game.Game, id uuid.UUID, want string) bool {
	t.Helper()
	for _, ty := range effectstest.EffectiveTypes(t, g, id) {
		if strings.EqualFold(ty, want) {
			return true
		}
	}
	return false
}

func oldB778InZone(z *game.Zone, id uuid.UUID) bool {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestOldB778FamiliarGroundLimitsEachOfYourCreaturesToOneBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Familiar Ground", "Enchantment", oldB778Ground, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "Attacker A", oldB778Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Attacker B", oldB778Bear, 2, 2)
	x := effectstest.B12Creature(g, opp.ID, "Blocker X", oldB778Bear, 2, 2)
	y := effectstest.B12Creature(g, opp.ID, "Blocker Y", oldB778Bear, 2, 2)
	effectstest.BrAttack(t, g, a, b)
	effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: x, Attacker: a}, {Blocker: y, Attacker: a},
	}), game.BlockReasonTooManyBlockers)
	effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: x, Attacker: b}, {Blocker: y, Attacker: b},
	}), game.BlockReasonTooManyBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: x, Attacker: a}, {Blocker: y, Attacker: b}}); err != nil {
		t.Errorf("one blocker each is legal: %v", err)
	}
}

func TestOldB778WithoutFamiliarGroundTwoCreaturesMayBlockOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Attacker A", oldB778Bear, 2, 2)
	x := effectstest.B12Creature(g, opp.ID, "Blocker X", oldB778Bear, 2, 2)
	y := effectstest.B12Creature(g, opp.ID, "Blocker Y", oldB778Bear, 2, 2)
	effectstest.BrAttack(t, g, a)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: x, Attacker: a}, {Blocker: y, Attacker: a}}); err != nil {
		t.Errorf("a double block is legal with no Familiar Ground: %v", err)
	}
}

func TestOldB778DenseFoliageBarsSpellsButNotAbilitiesFromTargetingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB778Bear, 2, 9)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB778Bear, 2, 9)
	banner := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Jabari's Banner", "Artifact", oldB778Banner)
	bolt := func() error {
		return effectstest.CastCatalogSpellErrStrict(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	}
	if err := bolt(); err != nil {
		t.Fatalf("a spell could target a creature before Dense Foliage: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B12Push(g, me.ID, "Dense Foliage", "Enchantment", oldB778Foliage, 0, 0)
	if err := bolt(); err == nil {
		t.Error("a spell targeted an opponent's creature under Dense Foliage")
	}
	for _, id := range []uuid.UUID{mine, theirs} {
		if err := effectstest.CastCatalogSpellErrStrict(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: id}}); err == nil {
			t.Errorf("a spell targeted creature %s under Dense Foliage", id)
		}
	}
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, banner, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}}); err != nil {
		t.Errorf("an ability could not target a creature under Dense Foliage: %v", err)
	}
}

func TestOldB778PendrellMistsMakesEachCreatureCostItsControllerOneAtUpkeep(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Pendrell Mists", "Enchantment", oldB778Mists, 0, 0)
		mine := effectstest.B12Creature(g, me.ID, "Mine", oldB778Bear, 2, 2)
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.FloatMana(t, g, me, "{G}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(mine); got != pay {
			t.Errorf("paid %v: my creature on the battlefield = %v", pay, got)
		}
	}
}

func TestOldB778TeferisVeilPhasesAnAttackerOutAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.B12Push(g, me.ID, "Teferi's Veil", "Enchantment", oldB778Veil, 0, 0)
	atk := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	idle := effectstest.PushVanillaCreature(g, me.ID, "Idle", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PhasedOutInCatalogGame(g, atk) {
		t.Fatal("the attacker phased out before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PhasedOutInCatalogGame(g, atk) {
		t.Error("the attacker did not phase out at end of combat")
	}
	if effectstest.PhasedOutInCatalogGame(g, idle) {
		t.Error("a creature that did not attack phased out")
	}
}

func TestOldB778AngelicRenewalSacrificesItselfToReturnACreatureThatDied(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		renewal := effectstest.B12Push(g, me.ID, "Angelic Renewal", "Enchantment", oldB778Renewal, 0, 0)
		bear := effectstest.B12Creature(g, me.ID, "Bear", oldB778Bear, 2, 2)
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bear); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(bear); got != yes {
			t.Errorf("answered %v: the creature is on the battlefield = %v", yes, got)
		}
		if got := g.Battlefield.Contains(renewal); got == yes {
			t.Errorf("answered %v: Angelic Renewal is on the battlefield = %v", yes, got)
		}
	}
}

func TestOldB778AngelicRenewalIgnoresAnOpponentsCreatureAndAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Angelic Renewal", "Enchantment", oldB778Renewal, 0, 0)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB778Bear, 2, 2)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, theirs); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("Angelic Renewal asked about a creature put into an opponent's graveyard")
	}
}

func TestOldB778JabarisBannerGrantsFlankingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	banner := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Jabari's Banner", "Artifact", oldB778Banner)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB778Bear, 2, 2)
	effectstest.FillPool(me, 1)
	effectstest.Pr7Activate(t, g, me.ID, banner, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flanking") {
		t.Error("the creature did not gain flanking")
	}
	if !effectstest.PermanentTapped(g, banner) {
		t.Error("the Banner was not tapped to pay")
	}
}

func TestOldB778AvizoaPumpsOncePerTurnAndSkipsYourNextUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jelly := effectstest.B12Push(g, me.ID, "Avizoa", "Creature — Jellyfish", oldB778Avizoa, 2, 2)
	effectstest.AssertKeywords(t, g, jelly, "flying")
	land := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	effectstest.Pr7Activate(t, g, me.ID, jelly, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, jelly); p != 4 || tough != 4 {
		t.Errorf("Avizoa is %d/%d, want 4/4", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, jelly, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("Avizoa activated twice in one turn")
	}
	if err := g.TapTargetForEffect(land); err != nil {
		t.Fatalf("tap the land: %v", err)
	}
	effectstest.NextTurnOfMine(t, g)
	if !effectstest.PermanentTapped(g, land) {
		t.Error("the land untapped although the untap step should have been skipped")
	}
	effectstest.NextTurnOfMine(t, g)
	if effectstest.PermanentTapped(g, land) {
		t.Error("the land stayed tapped through a second untap step")
	}
}

func TestOldB778NecratogExilesTheTopCreatureCardOfYourGraveyardToGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tog := effectstest.B12Push(g, me.ID, "Necratog", "Creature — Atog", oldB778Necratog, 1, 2)
	older := effectstest.PushGraveyardCardTyped(me, "Older Bear", oldB778Bear)
	top := effectstest.PushGraveyardCardTyped(me, "Top Bear", oldB778Bear)
	spell := effectstest.PushGraveyardCardTyped(me, "Spell on top", "Sorcery")
	effectstest.Pr7Activate(t, g, me.ID, tog, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{top}})
	if p, tough := effectstest.PtOf(t, g, tog); p != 3 || tough != 4 {
		t.Errorf("Necratog is %d/%d, want 3/4", p, tough)
	}
	if !effectstest.InExile(g, top) {
		t.Error("the top creature card was not exiled")
	}
	if effectstest.InExile(g, older) || effectstest.InExile(g, spell) {
		t.Error("a card other than the top creature card was exiled")
	}
	// Once more takes the next creature card down; a third has nothing to pay.
	effectstest.Pr7Activate(t, g, me.ID, tog, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{older}})
	if !effectstest.InExile(g, older) {
		t.Error("the second activation did not exile the next creature card")
	}
	if err := g.ActivateCatalogAbility(me.ID, tog, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Necratog activated with no creature card in the graveyard")
	}
}

func TestOldB778ThranForgePumpsAndMakesANonartifactCreatureAnArtifactUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forge := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Thran Forge", "Artifact", oldB778Forge)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB778Bear, 2, 2)
	golem := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Golem", "Artifact Creature — Golem", "")
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, forge, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: golem}}}); err == nil {
		t.Error("the Forge targeted an artifact creature")
	}
	effectstest.Pr7Activate(t, g, me.ID, forge, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, tough)
	}
	if !oldB778HasType(t, g, bear, "Artifact") || !oldB778HasType(t, g, bear, "Creature") {
		t.Error("the Bear is not an artifact creature")
	}
	effectstest.NextTurnOfMine(t, g)
	if oldB778HasType(t, g, bear, "Artifact") {
		t.Error("the Bear is still an artifact after the turn ended")
	}
}

func TestOldB778DwarvenBerserkerGetsPlusThreeAndTrampleWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwarf := effectstest.B12Push(g, me.ID, "Dwarven Berserker", "Creature — Dwarf Berserker", oldB778Dwarf, 1, 1)
	wall := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4})
	effectstest.DeclareAttack(t, g, opp.ID, dwarf)
	if p, _ := effectstest.PtOf(t, g, dwarf); p != 1 {
		t.Fatalf("the Berserker is %d power before being blocked", p)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, dwarf); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, dwarf); p != 4 || tough != 1 {
		t.Errorf("the Berserker is %d/%d once blocked, want 4/1", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dwarf, "trample") {
		t.Error("the Berserker did not gain trample")
	}
}

func TestOldB778DwarvenBerserkerUnblockedStaysSmall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwarf := effectstest.B12Push(g, me.ID, "Dwarven Berserker", "Creature — Dwarf Berserker", oldB778Dwarf, 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, dwarf)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, dwarf); p != 1 {
		t.Errorf("the unblocked Berserker has %d power, want 1", p)
	}
}
