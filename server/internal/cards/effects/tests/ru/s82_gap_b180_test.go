package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b180_test.go — GAP batch 180 (#5948): Commander Starter Decks
// creatures. Every test plays the real catalog entry through the engine.

const (
	gapB180VerdantForce = "7a21ea22-3cd7-4c11-8895-5943c0d93a0d"
	gapB180Disciple     = "781cee5f-f9cb-42ed-b408-543d297bf96d"
	gapB180Emmara       = "c65ba242-3369-48a9-864f-1b1f85238f67"
	gapB180Akoum        = "05b6f9e3-9acd-43a6-acff-caa811aaf0a1"
	gapB180Magmatic     = "ea73153b-1ecb-4593-a9ec-ce4f54e69824"
	gapB180Mordant      = "c02ed1b7-1f8d-4196-a9b6-65e1b5a763d7"
	gapB180Devils       = "b446b485-c10b-4175-bb49-16ac2f9e7718"
	gapB180Drudge       = "1c6a0d19-c7dd-4353-95ad-dfdef2d5d2d9"
	gapB180Isperia      = "c46718dc-24dc-4b77-b455-aa4c89570b8d"
	gapB180Phoenix      = "326d9b5c-ac04-45b5-9a8c-a9be7d8cca1b" // Firewing Phoenix, a graveyard activation
	gapB180BearType     = "Creature — Bear"
)

func TestGapB180EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		gapB180VerdantForce, gapB180Disciple, gapB180Emmara, gapB180Akoum, gapB180Magmatic,
		gapB180Mordant, gapB180Devils, gapB180Drudge, gapB180Isperia,
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

// --- Verdant Force / Magmatic Force ----------------------------------------

func TestGapB180VerdantForceMakesASaprolingEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Verdant Force", "Creature — Elemental", gapB180VerdantForce, 7, 7)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Saproling"); got != 1 {
		t.Fatalf("an opponent's upkeep: %d Saprolings, want 1", got)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Saproling"); got != 4 {
		t.Errorf("a full round of upkeeps: %d Saprolings, want 4", got)
	}
}

func TestGapB180MagmaticForceDealsThreeToAnyTargetEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Magmatic Force", "Creature — Elemental", gapB180Magmatic, 7, 7)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, wall); got != 3 {
		t.Errorf("the target has %d damage, want 3 on an opponent's upkeep", got)
	}
	life := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the player lost %d, want 3", life-opp.Life)
	}
}

// --- Emmara ---------------------------------------------------------------

func TestGapB180EmmaraMakesALifelinkSoldierWhenSheBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	emmara := effectstest.B12Push(g, me.ID, "Emmara, Soul of the Accord", "Legendary Creature — Elf Cleric", gapB180Emmara, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, emmara)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Soldier"); got != 1 {
		t.Fatalf("%d Soldier tokens after attacking, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" && (!effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "lifelink") || c.Power != 1 || c.Toughness != 1) {
			t.Errorf("the Soldier is %d/%d with lifelink=%v, want a 1/1 with lifelink", c.Power, c.Toughness, effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "lifelink"))
		}
	}
}

func TestGapB180EmmaraWithVigilanceDoesNotBecomeTappedByAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	emmara := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Emmara, Soul of the Accord", TypeLine: "Legendary Creature — Elf Cleric",
		OracleID: gapB180Emmara, Power: 2, Toughness: 2, Keywords: []string{"vigilance"}})
	effectstest.DeclareAttack(t, g, opp.ID, emmara)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Soldier"); got != 0 {
		t.Errorf("%d Soldier tokens, want none: an attack with vigilance doesn't tap her", got)
	}
}

// --- Draconic Disciple ----------------------------------------------------

func TestGapB180DraconicDiscipleTapsForAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	disciple := effectstest.B12Push(g, me.ID, "Draconic Disciple", "Creature — Human Shaman", gapB180Disciple, 2, 2)
	for _, colour := range []string{"W", "U", "B", "R", "G"} {
		me.ManaPool.EmptyPool()
		effectstest.FindBattlefieldCardByID(g, disciple).Tapped = false
		if err := g.ActivateManaAbility(me.ID, disciple, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
			t.Fatalf("tap for %s: %v", colour, err)
		}
		if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != colour {
			t.Errorf("pool = %v, want one %s", me.ManaPool.Units(), colour)
		}
	}
}

func TestGapB180DraconicDiscipleSacrificesForAFiveFiveDragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	disciple := effectstest.B12Push(g, me.ID, "Draconic Disciple", "Creature — Human Shaman", gapB180Disciple, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, disciple, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {7}")
	}
	effectstest.ApaMana(me, "C", "C", "C", "C", "C", "C", "C")
	effectstest.ApaActivate(t, g, me, disciple, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(disciple) {
		t.Error("the Disciple was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Dragon"); got != 1 {
		t.Fatalf("%d Dragon tokens, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dragon" && (c.Power != 5 || c.Toughness != 5 || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying")) {
			t.Errorf("the Dragon is %d/%d, want a 5/5 with flying", c.Power, c.Toughness)
		}
	}
}

// --- Isperia, Supreme Judge -----------------------------------------------

func TestGapB180IsperiaDrawsWhenACreatureAttacksYouButNotAnotherPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attacker, you, other := g.Seats[0], g.Seats[1], g.Seats[2]
	isperia := effectstest.B12Push(g, you.ID, "Isperia, Supreme Judge", "Legendary Creature — Sphinx", gapB180Isperia, 6, 4)
	bear := effectstest.B12Creature(g, attacker.ID, "Bear", gapB180BearType, 2, 2)
	effectstest.DeclareAttack(t, g, other.ID, bear)
	if n := effectstest.TriggersOnStackFrom(g, isperia); n != 0 {
		t.Fatalf("%d triggers when another player was attacked, want 0", n)
	}
	g2 := effectstest.NewCatalogGame(t)
	attacker2, you2 := g2.Seats[0], g2.Seats[1]
	isperia2 := effectstest.B12Push(g2, you2.ID, "Isperia, Supreme Judge", "Legendary Creature — Sphinx", gapB180Isperia, 6, 4)
	bear2 := effectstest.B12Creature(g2, attacker2.ID, "Bear", gapB180BearType, 2, 2)
	hand := you2.Hand.Size()
	effectstest.DeclareAttack(t, g2, you2.ID, bear2)
	if n := effectstest.TriggersOnStackFrom(g2, isperia2); n != 1 {
		t.Fatalf("%d triggers when I was attacked, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerLatestTriggerPrompt(t, g2, you2.ID, true)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := you2.Hand.Size(); got != hand+1 {
		t.Errorf("hand grew by %d, want 1", got-hand)
	}
}

// --- Akoum Hellkite -------------------------------------------------------

func gapB180PlayLand(t *testing.T, g *game.Game, name, typeLine string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}

func TestGapB180AkoumHellkiteLandfallDealsOneOrTwoForAMountain(t *testing.T) {
	for _, tc := range []struct {
		land, typeLine string
		want           int
	}{
		{"Forest", "Basic Land — Forest", 1},
		{"Mountain", "Basic Land — Mountain", 2},
	} {
		t.Run(tc.land, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.B12Push(g, me.ID, "Akoum Hellkite", "Creature — Dragon", gapB180Akoum, 4, 4)
			life := opp.Life
			gapB180PlayLand(t, g, tc.land, tc.typeLine)
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickPlayer(t, g, me.ID, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if got := life - opp.Life; got != tc.want {
				t.Errorf("%s: dealt %d, want %d", tc.land, got, tc.want)
			}
		})
	}
}

func TestGapB180AkoumHellkiteIgnoresAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hellkite := effectstest.B12Push(g, me.ID, "Akoum Hellkite", "Creature — Dragon", gapB180Akoum, 4, 4)
	effectstest.B12Permanent(g, opp.ID, "Their Mountain", "Basic Land — Mountain")
	if n := effectstest.TriggersOnStackFrom(g, hellkite); n != 0 {
		t.Errorf("%d triggers for a land I don't control", n)
	}
}

// --- Laboratory Drudge ----------------------------------------------------

func TestGapB180LaboratoryDrudgeDrawsOnlyAfterAGraveyardCastOrActivation(t *testing.T) {
	// Nothing from a graveyard this turn: no draw at the end step.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Laboratory Drudge", "Creature — Zombie Horror", gapB180Drudge, 3, 4)
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand grew by %d with nothing played from a graveyard, want 0", got-hand)
	}
}

func TestGapB180LaboratoryDrudgeDrawsAfterAGraveyardActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.B12Push(g, me.ID, "Laboratory Drudge", "Creature — Zombie Horror", gapB180Drudge, 3, 4)
	effectstest.SeedLibrary(me, "A", "B", "C")
	phoenix := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: phoenix, Name: "Firewing Phoenix", TypeLine: "Creature — Phoenix",
		OracleID: gapB180Phoenix, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "R", "R", "R")
	effectstest.B16Activate(t, g, me.ID, phoenix, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand grew by %d after activating a graveyard card's ability, want 1", got-hand)
	}
}

func TestGapB180LaboratoryDrudgeDrawsAfterAFlashbackCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Laboratory Drudge", "Creature — Zombie Horror", gapB180Drudge, 3, 4)
	effectstest.SeedLibrary(me, "A", "B", "C")
	foe := effectstest.B12Creature(g, opp.ID, "Foe", gapB180BearType, 2, 2)
	effectstest.MidB03Flash(t, g, "Abandon the Post", "Sorcery", "5404c6a5-399d-4237-b03c-41b6a1b79050", effectstest.CardTarget(foe))
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand grew by %d after casting a spell from a graveyard, want 1", got-hand)
	}
}

// --- Mordant Dragon -------------------------------------------------------

func TestGapB180MordantDragonFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	dragon := effectstest.B12Push(g, me.ID, "Mordant Dragon", "Creature — Dragon", gapB180Mordant, 5, 5)
	effectstest.ApaMana(me, "C", "R")
	effectstest.ApaActivate(t, g, me, dragon, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, dragon); p != 6 {
		t.Errorf("power %d, want 6", p)
	}
}

func TestGapB180MordantDragonMayHitACreatureOfTheDamagedPlayerForThatMuch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	dragon := effectstest.B12Push(g, me.ID, "Mordant Dragon", "Creature — Dragon", gapB180Mordant, 5, 5)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	elsewhere := effectstest.B12Creature(g, other.ID, "Elsewhere", "Creature — Wall", 0, 9)
	effectstest.BrAttack(t, g, dragon)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(p.PickTargetCards, elsewhere) || !effectstest.HasID(p.PickTargetCards, wall) {
		t.Errorf("offered %v: only the damaged player's creatures may be picked", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, wall)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, wall); got != 5 {
		t.Errorf("the creature took %d, want 5 (the combat damage)", got)
	}
}

func TestGapB180MordantDragonMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Mordant Dragon", "Creature — Dragon", gapB180Mordant, 5, 5)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.BrAttack(t, g, dragon)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, wall)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, wall); got != 0 {
		t.Errorf("the creature took %d after the caster declined", got)
	}
}

// --- Wildfire Devils ------------------------------------------------------

func gapB180Bolts(g *game.Game, perSeat int) map[uuid.UUID][]uuid.UUID {
	out := map[uuid.UUID][]uuid.UUID{}
	for _, p := range g.Seats {
		for i := 0; i < perSeat; i++ {
			out[p.ID] = append(out[p.ID], effectstest.GyCard(p, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle))
		}
	}
	return out
}

func gapB180ExiledCount(g *game.Game, bolts map[uuid.UUID][]uuid.UUID) int {
	n := 0
	for _, ids := range bolts {
		for _, id := range ids {
			if effectstest.InExile(g, id) {
				n++
			}
		}
	}
	return n
}

func TestGapB180WildfireDevilsExilesOneCardOfARandomPlayerAndCastsTheCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolts := gapB180Bolts(g, 1)
	effectstest.CastAndResolveCreature(t, g, "Wildfire Devils", "Creature — Devil", gapB180Devils)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB180ExiledCount(g, bolts); got != 1 {
		t.Fatalf("%d graveyard cards exiled, want exactly one", got)
	}
	cast := effectstest.ResolutionCastFor(g, me.ID)
	if cast == nil {
		t.Fatal("no offer to cast the copy")
	}
	cp := effectstest.CopyOf(t, g, "Lightning Bolt")
	life := opp.Life
	if err := g.CastForResolutionChoice(me.ID, cast.ID, cp.InstanceID, game.CastSpellParams{
		FromZone: "exile", Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)},
	}); err != nil {
		t.Fatalf("casting the copy for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d → %d, want -3 from the free Bolt copy", life, opp.Life)
	}
}

func TestGapB180WildfireDevilsLetsTheChosenPlayerPickWhichCardIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bolts := gapB180Bolts(g, 2)
	effectstest.CastAndResolveCreature(t, g, "Wildfire Devils", "Creature — Devil", gapB180Devils)
	effectstest.PassPriorityAroundTable(t, g)
	var chooser *game.Player
	for _, p := range g.Seats {
		if effectstest.ChooseCardsChoiceFor(g, p.ID) != nil {
			chooser = p
		}
	}
	if chooser == nil {
		t.Fatal("no player was asked which card to exile")
	}
	effectstest.AnswerChooseCards(t, g, chooser.ID, bolts[chooser.ID][1])
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, bolts[chooser.ID][1]) || effectstest.InExile(g, bolts[chooser.ID][0]) {
		t.Error("the card the chosen player named is not the one exiled")
	}
	if got := gapB180ExiledCount(g, bolts); got != 1 {
		t.Errorf("%d cards exiled, want exactly one", got)
	}
}

func TestGapB180WildfireDevilsDoesNothingWithoutAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		effectstest.GyCard(p, "Bear", gapB180BearType, "{1}{G}", "")
	}
	effectstest.CastAndResolveCreature(t, g, "Wildfire Devils", "Creature — Devil", gapB180Devils)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastFor(g, me.ID) != nil {
		t.Error("a cast was offered with no instant or sorcery card to copy")
	}
	if g.Exile.Size() != 0 {
		t.Error("something was exiled")
	}
}

func TestGapB180WildfireDevilsTriggersOnYourUpkeepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Wildfire Devils", "Creature — Devil", gapB180Devils, 4, 2)
	bolts := gapB180Bolts(g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB180ExiledCount(g, bolts); got != 0 {
		t.Fatalf("%d exiled on an opponent's upkeep, want 0", got)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB180ExiledCount(g, bolts); got != 1 {
		t.Errorf("%d exiled on my upkeep, want 1", got)
	}
}
