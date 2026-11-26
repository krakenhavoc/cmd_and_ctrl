package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o333Display    = "4a2233de-50bd-49be-ac49-4a3c0f02dd52"
	o333Surge      = "5c9a2789-4b8c-4706-a95a-99561cd6d09c"
	o333Vandalize  = "8c71cbb3-e900-4ea3-aced-8439b5803654"
	o333Silkwrap   = "909becbc-1d84-4e94-b589-e58438569403"
	o333Purge      = "b61718ad-7102-4ed2-b884-733bc46df715"
	o333Renewal    = "6beaf310-c623-4ab6-a0d3-f4c8221f255b"
	o333Executor   = "8587ec98-f35d-476c-ba67-0ef8bcd1f253"
	o333Corpseweft = "9ffb5615-124e-4848-9e0d-120851bd50e5"
	o333Command    = "c1fafe93-8fd2-44c1-a4a0-5c9ecee550cc"
	o333Gravepurge = "cc94462f-a207-4fdb-8704-aa979c244cab"
)

func TestO333CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		o333Display: "Display of Dominance", o333Surge: "Surge of Righteousness", o333Vandalize: "Vandalize",
		o333Silkwrap: "Silkwrap", o333Purge: "Radiant Purge", o333Renewal: "Foul Renewal",
		o333Executor: "Risen Executioner", o333Corpseweft: "Corpseweft", o333Command: "Ojutai's Command",
		o333Gravepurge: "Gravepurge",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func o333Colored(g *game.Game, owner uuid.UUID, name, typeLine string, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

func o333CastInHand(g *game.Game, p *game.Player, name, typeLine, oracle string, targets []game.TargetRef) error {
	id := effectstest.PutInHand(p, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Targets: targets, Strict: true})
}

func o333ModalErr(t *testing.T, g *game.Game, name, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	_, err := effectstest.CastModalErr(t, g, name, "Instant", oracle, modes, targets)
	return err
}

func o333Card(id uuid.UUID) []game.TargetRef { return effectstest.CardRefs(id) }

// --- Display of Dominance --------------------------------------------------

func TestO333DisplayDestroysABlueOrBlackNoncreaturePermanentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	blueArt := o333Colored(g, opp.ID, "Blue Relic", "Artifact", "U")
	blueCreature := o333Colored(g, opp.ID, "Blue Bear", "Creature — Bear", "U")
	redArt := o333Colored(g, opp.ID, "Red Relic", "Artifact", "R")
	for _, bad := range []uuid.UUID{blueCreature, redArt} {
		if err := o333ModalErr(t, g, "Display of Dominance", o333Display, []int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: bad, Mode: 0}}); err == nil {
			t.Fatal("Display of Dominance took a creature or a red permanent")
		}
	}
	effectstest.CastModal(t, g, "Display of Dominance", "Instant", o333Display, []int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: blueArt, Mode: 0}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blueArt) {
		t.Error("the blue artifact survived")
	}
	if !g.Battlefield.Contains(blueCreature) || !g.Battlefield.Contains(redArt) {
		t.Error("an illegal permanent was destroyed")
	}
}

func TestO333DisplayShieldsPermanentsFromOpposingBlueAndBlackSpellsThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastModal(t, g, "Display of Dominance", "Instant", o333Display, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModCantBeTargetedByOpponentSpellsOfColors && e.Scope == game.ScopeYourPermanents && e.Controller == me.ID &&
				len(m.Colors) == 2 && m.Colors[0] == "U" && m.Colors[1] == "B" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("no opponents-only blue/black targeting bar over my permanents was recorded")
	}
}

// --- Surge of Righteousness ------------------------------------------------

func TestO333SurgeDestroysAnAttackingBlackOrRedCreatureAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	blackAttacker := o333Colored(g, me.ID, "Black Attacker", "Creature — Zombie", "B")
	greenAttacker := o333Colored(g, me.ID, "Green Attacker", "Creature — Bear", "G")
	idle := o333Colored(g, me.ID, "Idle Red", "Creature — Goblin", "R")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{blackAttacker, greenAttacker} {
		if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	for _, bad := range []uuid.UUID{greenAttacker, idle} {
		if err := o333CastInHand(g, me, "Surge of Righteousness", "Instant", o333Surge, o333Card(bad)); err == nil {
			t.Fatal("Surge of Righteousness took a green attacker or a non-attacking creature")
		}
	}
	life := me.Life
	if err := o333CastInHand(g, me, "Surge of Righteousness", "Instant", o333Surge, o333Card(blackAttacker)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blackAttacker) {
		t.Error("the black attacker survived")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

// --- Vandalize ---------------------------------------------------------------

func TestO333VandalizeDestroysEachChosenTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	art := o333Colored(g, opp.ID, "Relic", "Artifact")
	land := o333Colored(g, opp.ID, "Plains", "Basic Land — Plains")
	effectstest.CastModal(t, g, "Vandalize", "Sorcery", o333Vandalize, []int{0, 1}, []game.TargetRef{
		{Kind: game.TargetCard, ID: art, Mode: 0}, {Kind: game.TargetCard, ID: land, Mode: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(art) || g.Battlefield.Contains(land) {
		t.Error("both the artifact and the land should be destroyed when both are chosen")
	}
}

func TestO333VandalizeOneModeLeavesTheOtherTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	art := o333Colored(g, opp.ID, "Relic", "Artifact")
	land := o333Colored(g, opp.ID, "Plains", "Basic Land — Plains")
	effectstest.CastModal(t, g, "Vandalize", "Sorcery", o333Vandalize, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: land, Mode: 0}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !g.Battlefield.Contains(art) {
		t.Error("only the land should have been destroyed")
	}
}

// --- Silkwrap ----------------------------------------------------------------

func TestO333SilkwrapExilesAnOpposingSmallCreatureUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Small", TypeLine: "Creature — Bear",
		ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant",
		ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mine", TypeLine: "Creature — Bear",
		ManaCost: "{1}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	wrap := effectstest.CastCatalogSpell(t, g, "Silkwrap", "Enchantment", o333Silkwrap, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	for _, bad := range []uuid.UUID{big, mine} {
		if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(bad)); err == nil {
			t.Fatal("Silkwrap targeted a creature with mana value above 3 or its controller's own creature")
		}
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(small) {
		t.Fatal("the small creature was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wrap) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Small", opp.ID); n != 1 {
		t.Errorf("after Silkwrap left, %d of the exiled creature under its owner's control, want 1", n)
	}
}

// --- Radiant Purge -----------------------------------------------------------

func TestO333RadiantPurgeExilesOnlyMulticoloredCreaturesAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	gold := o333Colored(g, opp.ID, "Gold Bear", "Creature — Bear", "W", "U")
	mono := o333Colored(g, opp.ID, "Mono Bear", "Creature — Bear", "W")
	goldArt := o333Colored(g, opp.ID, "Gold Relic", "Artifact", "W", "U")
	goldEnch := o333Colored(g, opp.ID, "Gold Ward", "Enchantment", "G", "B")
	for _, bad := range []uuid.UUID{mono, goldArt} {
		if err := effectstest.CastCatalogSpellErrStrict(t, g, "Radiant Purge", "Instant", o333Purge, o333Card(bad)); err == nil {
			t.Fatal("Radiant Purge took a monocolored creature or a multicolored artifact")
		}
	}
	for _, good := range []uuid.UUID{gold, goldEnch} {
		effectstest.CastCatalogSpell(t, g, "Radiant Purge", "Instant", o333Purge, o333Card(good))
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Exile.Contains(good) {
			t.Errorf("%v was not exiled", good)
		}
	}
}

// --- Foul Renewal ------------------------------------------------------------

func TestO333FoulRenewalReturnsTheCardAndShrinksByItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tough := effectstest.GraveyardCardOf(g, me, "Tough Dead", "Creature — Wall", "{2}", "")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == tough {
				me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = 0, 3
			}
		}
	})
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Giant", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Foul Renewal", "Instant", o333Renewal, []game.TargetRef{
		{Kind: game.TargetCard, ID: tough, Slot: 0}, {Kind: game.TargetCard, ID: victim, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, tough) {
		t.Error("the creature card was not returned to hand")
	}
	if p, th := effectstest.EffectivePower(t, g, victim), effectstest.EffectiveToughness(t, g, victim); p != 2 || th != 2 {
		t.Errorf("the second target is %d/%d, want 2/2 (-3/-3)", p, th)
	}
}

// --- Risen Executioner -------------------------------------------------------

func TestO333RisenExecutionerCantBlockAndPumpsOtherZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ex := effectstest.PushCatalogPermanent(g, me.ID, "Risen Executioner", "Creature — Zombie Warrior", o333Executor, false)
	zombie := effectstest.B12Creature(g, me.ID, "Zed", "Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, zombie); p != 3 {
		t.Errorf("a Zombie is %d power, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Zombie is %d power, want 2", p)
	}
	if got := effectstest.EffectivePower(t, g, ex); got != 1 {
		t.Errorf("the Executioner gave itself the bonus: power %d, want its printed base 1", got)
	}
	if !effectstest.SbCantBlock(t, g, ex) {
		t.Error("the Executioner can block")
	}
}

func TestO333RisenExecutionerGraveyardCastCostsOneMorePerOtherCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ex := effectstest.GraveyardCardOf(g, me, "Risen Executioner", "Creature — Zombie Warrior", "{2}{B}{B}", o333Executor)
	effectstest.GraveyardCardOf(g, me, "Dead A", "Creature — Bear", "{1}", "")
	effectstest.GraveyardCardOf(g, me, "Dead B", "Creature — Bear", "{1}", "")
	effectstest.GraveyardCardOf(g, me, "Not A Creature", "Sorcery", "{1}", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "B", "B", "B")
	if err := g.CastSpell(me.ID, ex, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("cast with five mana; the two other creature cards add {2} for six")
	}
	effectstest.B06AddMana(me, "B")
	if err := g.CastSpell(me.ID, ex, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast from the graveyard with six mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ex) {
		t.Error("the Executioner did not enter")
	}
}

// --- Corpseweft --------------------------------------------------------------

func TestO333CorpseweftMakesATappedZombieHorrorTwiceTheCardsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cw := effectstest.PushCatalogPermanent(g, me.ID, "Corpseweft", "Enchantment", o333Corpseweft, false)
	a := effectstest.GraveyardCardOf(g, me, "Dead A", "Creature — Bear", "{1}", "")
	b := effectstest.GraveyardCardOf(g, me, "Dead B", "Creature — Bear", "{1}", "")
	land := effectstest.GraveyardCardOf(g, me, "Dead Land", "Land", "", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, cw, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated exiling no cards")
	}
	if err := g.ActivateCatalogAbility(me.ID, cw, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("activated exiling a noncreature card")
	}
	effectstest.B16Activate(t, g, me.ID, cw, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{a, b}})
	effectstest.PassPriorityAroundTable(t, g)
	toks := effectstest.BattlefieldIDsNamed(g, "Zombie Horror")
	if len(toks) != 1 {
		t.Fatalf("%d Zombie Horror tokens, want 1", len(toks))
	}
	if p, th := effectstest.EffectivePower(t, g, toks[0]), effectstest.EffectiveToughness(t, g, toks[0]); p != 4 || th != 4 {
		t.Errorf("the token is %d/%d, want 4/4", p, th)
	}
	if !effectstest.IsTapped(g, toks[0]) {
		t.Error("the token entered untapped")
	}
	if me.Graveyard.Contains(a) || me.Graveyard.Contains(b) || !me.Graveyard.Contains(land) {
		t.Error("the cost exiled the wrong cards")
	}
}

// --- Ojutai's Command --------------------------------------------------------

func TestO333CommandReturnsASmallCreatureAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	small := effectstest.GraveyardCardOf(g, me, "Small Dead", "Creature — Bear", "{1}{G}", "")
	big := effectstest.GraveyardCardOf(g, me, "Big Dead", "Creature — Giant", "{3}{G}", "")
	if err := o333ModalErr(t, g, "Ojutai's Command", o333Command, []int{0, 3}, []game.TargetRef{{Kind: game.TargetCard, ID: big, Mode: 0}}); err == nil {
		t.Fatal("returned a creature card with mana value 4")
	}
	hand := me.Hand.Size()
	effectstest.CastModal(t, g, "Ojutai's Command", "Instant", o333Command, []int{0, 3}, []game.TargetRef{{Kind: game.TargetCard, ID: small, Mode: 0}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the small creature did not return to the battlefield")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after drawing a card", me.Hand.Size(), hand+1)
	}
}

func TestO333CommandGainsLifeAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.CastModal(t, g, "Ojutai's Command", "Instant", o333Command, []int{1, 3}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

// --- Gravepurge --------------------------------------------------------------

func TestO333GravepurgePutsTheChosenCreatureCardsOnTopThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.GraveyardCardOf(g, me, "Dead A", "Creature — Bear", "{1}", "")
	b := effectstest.GraveyardCardOf(g, me, "Dead B", "Creature — Bear", "{1}", "")
	sorc := effectstest.GraveyardCardOf(g, me, "Old Spell", "Sorcery", "{1}", "")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Gravepurge", "Instant", o333Gravepurge, o333Card(sorc)); err == nil {
		t.Fatal("Gravepurge took a noncreature card")
	}
	hand, lib := me.Hand.Size(), me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Gravepurge", "Instant", o333Gravepurge, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(a) || me.Graveyard.Contains(b) {
		t.Error("a chosen card stayed in the graveyard")
	}
	if me.Hand.Size() != hand+1 || me.Library.Size() != lib+1 {
		t.Errorf("hand %d (want %d) library %d (want %d): two went on top and one was drawn", me.Hand.Size(), hand+1, me.Library.Size(), lib+1)
	}
}
