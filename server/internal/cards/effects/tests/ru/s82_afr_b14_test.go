package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 — Adventures in the Forgotten Realms, batch 14: Classes II and cards
// played from exile.

const (
	afrB14RogueClass    = "76f0f39a-ce1b-4935-a6b8-5c4e1a27ab24"
	afrB14BardClass     = "5bbc3ad0-4865-43f0-8baf-e7f4af5db656"
	afrB14MonkClass     = "65b28ec7-1099-4f9d-8d57-2ffd8a5b0e91"
	afrB14SorcererClass = "19ba0f2d-4de3-49a0-be7d-da0aa220cd8a"
	afrB14Flameskull    = "f3d935cd-d582-464a-a13b-bb84a53da4a1"
	afrB14Hurl          = "4254c048-7207-4ca0-bd20-ebb418d9ed1a"
	afrB14Asmodeus      = "1e4032f6-9741-4845-9325-781c9e489172"
	afrB14ShareSpoils   = "d39af66c-4997-48b8-a704-f21d93247d7d"
)

func TestAfrB14EveryBuiltCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		afrB14Flameskull: "Flameskull",
		afrB14Hurl:       "Hurl Through Hell",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Hurl Through Hell ---------------------------------------------------

func TestAfrB14HurlExilesTheCreatureAndLetsYouCastItAsAnyColourUntilYourNextTurnEnds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Hurl Through Hell", "Instant", afrB14Hurl, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, bear) {
		t.Fatal("the target was not exiled")
	}
	perm := effectstest.ExiledPermission(g, bear)
	if perm.Player != me.ID || !perm.CastOnly || !perm.AnyColor {
		t.Errorf("permission %+v: the caster may CAST it with mana of any colour", perm)
	}
	if !effectstest.PermissionLive(g, perm, me.ID) {
		t.Error("the caster cannot cast it this turn")
	}
	if effectstest.PermissionLive(g, perm, opp.ID) {
		t.Error("its owner may cast it too")
	}
	effectstest.NextTurnOfMine(t, g)
	if !effectstest.B12ExiledPlayableBy(t, g, bear, me.ID) {
		t.Error("the window closed before the end of the caster's next turn")
	}
	effectstest.AdvanceToEndStep(t, g)
	for i := 0; i < 4 && effectstest.B12ExiledPlayableBy(t, g, bear, me.ID); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if effectstest.B12ExiledPlayableBy(t, g, bear, me.ID) {
		t.Error("the permission outlived the end of the caster's next turn")
	}
}

func TestAfrB14HurlRefusesANonCreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Rock", "Artifact", "", false)
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Hurl Through Hell", "Instant", afrB14Hurl, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	if err == nil {
		t.Error("a noncreature permanent was accepted")
	}
}

// --- Flameskull ----------------------------------------------------------

func afrB14FlameskullDies(t *testing.T, g *game.Game, me *game.Player, topName, topType string) (skull, top uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	skull = effectstest.PushCatalogPermanent(g, me.ID, "Flameskull", "Creature — Skeleton", afrB14Flameskull, false)
	top = effectstest.TopLibraryCard(me, topName, topType)
	effectstest.KillCreature(t, g, me.ID, skull)
	effectstest.PassPriorityAroundTable(t, g)
	return skull, top
}

func TestAfrB14FlameskullExilesItselfAndTheTopCardAndYouPlayOnlyOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skull, top := afrB14FlameskullDies(t, g, me, "Top Land", "Basic Land — Forest")
	if !effectstest.InExile(g, skull) || !effectstest.InExile(g, top) {
		t.Fatalf("Flameskull exiled: %v, top card exiled: %v", effectstest.InExile(g, skull), effectstest.InExile(g, top))
	}
	if !effectstest.B12ExiledPlayableBy(t, g, skull, me.ID) || !effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Fatal("both exiled cards should be playable at first")
	}
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing the exiled land: %v", err)
	}
	if effectstest.B12ExiledPlayableBy(t, g, skull, me.ID) {
		t.Error("Flameskull is still playable after the other card was played")
	}
}

func TestAfrB14FlameskullPlayedFirstLocksTheOtherCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skull, top := afrB14FlameskullDies(t, g, me, "Top Land", "Basic Land — Forest")
	if err := g.CastSpell(me.ID, skull, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("casting Flameskull from exile: %v", err)
	}
	if effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the other card is still playable after Flameskull was cast")
	}
}

func TestAfrB14FlameskullCannotBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skull := effectstest.PushCatalogPermanent(g, me.ID, "Flameskull", "Creature — Skeleton", afrB14Flameskull, false)
	if effectstest.AuraRestrictions(t, g, skull)&game.CantBlock == 0 {
		t.Error("Flameskull can block")
	}
}

func TestAfrB14FlameskullExiledFromTheGraveyardByAnotherEffectDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	skull := effectstest.PushCatalogPermanent(g, me.ID, "Flameskull", "Creature — Skeleton", afrB14Flameskull, false)
	top := effectstest.TopLibraryCard(me, "Top Card", "Sorcery")
	effectstest.ExileCreature(t, g, skull)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.InExile(g, top) {
		t.Error("an exiled Flameskull did not die, yet the top card was exiled")
	}
}

// afrB14Class seeds a Class at `level` under p (CR 716.2).
func afrB14Class(g *game.Game, p *game.Player, name, oracle string, level int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Enchantment — Class", OracleID: oracle,
		Owner: p.ID, Controller: p.ID, ClassLevel: level,
	})
}

// afrB14Perm asks the engine's one permission query: may `player` cast or play
// the exiled card now?
func afrB14Perm(g *game.Game, player uuid.UUID, id uuid.UUID) *game.CastPermission {
	var perm *game.CastPermission
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.InstanceID == id {
				perm = g.CastPermissionForLocked(player, c, game.ZoneExile)
			}
		}
	})
	return perm
}

// --- Rogue Class ---------------------------------------------------------

func TestAfrB14RogueClassExilesTheTopCardFaceDownAndYouMayLookButNotPlayIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	class := afrB14Class(g, me, "Rogue Class", afrB14RogueClass, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	top := effectstest.TopLibraryCard(opp, "Their Top Card", "Sorcery")
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, top) {
		t.Fatal("the top card of the damaged player's library was not exiled")
	}
	c := effectstest.CardInExile(t, g, top)
	if !c.FaceDown {
		t.Error("the card is not face down")
	}
	if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
		t.Errorf("only the Class's controller may look at it: known to me %v, to them %v", c.IsKnownTo(me.ID), c.IsKnownTo(opp.ID))
	}
	linked := false
	g.ReadSnapshot(func() {
		for _, e := range g.ExiledWithForEffect(game.PermissionCardRef{ID: class, Epoch: effectstest.FindBattlefieldCardByID(g, class).ObjectEpoch}) {
			linked = linked || e.InstanceID == top
		}
	})
	if !linked {
		t.Error("the card is not exiled with this Class")
	}
	if afrB14Perm(g, me.ID, top) != nil {
		t.Error("at level 1 the exiled card may be played")
	}
}

func TestAfrB14RogueClassTriggersForEachCreatureThatConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	afrB14Class(g, me, "Rogue Class", afrB14RogueClass, 1)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	first := effectstest.TopLibraryCard(opp, "First", "Sorcery")
	second := effectstest.TopLibraryCard(opp, "Second", "Sorcery")
	effectstest.DealCombatDamageToPlayer(g, a, opp.ID, 1)
	effectstest.DealCombatDamageToPlayer(g, b, opp.ID, 1)
	effectstest.DealCombatDamageToPlayer(g, theirs, me.ID, 1) // an opponent's creature: no trigger
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, first) || !effectstest.InExile(g, second) {
		t.Errorf("two creatures dealt damage; exiled: first %v, second %v", effectstest.InExile(g, first), effectstest.InExile(g, second))
	}
}

func TestAfrB14RogueClassLevelsGrantMenaceThenPlayingTheExiledCards(t *testing.T) {
	for _, level := range []int{1, 2, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		afrB14Class(g, me, "Rogue Class", afrB14RogueClass, level)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		top := effectstest.TopLibraryCard(opp, "Their Top Card", "Sorcery")
		effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HasKeywordOnBattlefield(t, g, bear, "menace"); got != (level >= 2) {
			t.Errorf("level %d: menace %v", level, got)
		}
		perm := afrB14Perm(g, me.ID, top)
		if (perm != nil) != (level >= 3) {
			t.Errorf("level %d: permission %+v", level, perm)
			continue
		}
		if perm != nil && (!perm.AnyColor || perm.CastOnly) {
			t.Errorf("level 3: permission %+v must be play (not cast only) with any colour", perm)
		}
		if afrB14Perm(g, opp.ID, top) != nil {
			t.Errorf("level %d: the card's owner may play it", level)
		}
	}
}

// --- Bard Class ----------------------------------------------------------

func afrB14Price(t *testing.T, g *game.Game, p *game.Player, spell game.Card) game.ParsedCost {
	t.Helper()
	base, err := game.ParseCost(spell.ManaCost)
	if err != nil {
		t.Fatal(err)
	}
	priced, err := g.ApplyCostModifiersForEffect(base, game.CostQuery{Game: g, Card: spell, Controller: p.ID})
	if err != nil {
		t.Fatal(err)
	}
	return priced
}

func TestAfrB14BardClassLegendaryCreaturesEnterWithAnExtraCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	afrB14Class(g, me, "Bard Class", afrB14BardClass, 1)
	effectstest.CastCatalogSpell(t, g, "Legend", "Legendary Creature — Elf", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Commoner", "Creature — Elf", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	var legend, commoner *game.Card
	for i := range g.Battlefield.Cards {
		switch g.Battlefield.Cards[i].Name {
		case "Legend":
			legend = &g.Battlefield.Cards[i]
		case "Commoner":
			commoner = &g.Battlefield.Cards[i]
		}
	}
	if legend == nil || commoner == nil {
		t.Fatal("the creatures did not enter")
	}
	if got := legend.Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("the legendary creature has %d +1/+1 counters, want 1", got)
	}
	if got := commoner.Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("a nonlegendary creature has %d counters", got)
	}
}

func TestAfrB14BardClassLevelTwoReducesOnlyTheColouredManaOfLegendarySpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	legend := game.Card{InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Elf", ManaCost: "{2}{R}{G}", Owner: me.ID, Controller: me.ID}
	mono := game.Card{InstanceID: uuid.New(), Name: "Mono Legend", TypeLine: "Legendary Creature — Elf", ManaCost: "{3}{G}", Owner: me.ID, Controller: me.ID}
	plain := game.Card{InstanceID: uuid.New(), Name: "Plain", TypeLine: "Creature — Elf", ManaCost: "{2}{R}{G}", Owner: me.ID, Controller: me.ID}
	class := afrB14Class(g, me, "Bard Class", afrB14BardClass, 1)
	if got := afrB14Price(t, g, me, legend); got.Generic != 2 || len(got.Required) != 2 {
		t.Errorf("level 1 prices {2}{R}{G} as %+v", got)
	}
	g.WithWriteLock(func() { _ = g.SetClassLevelForEffect(class, 2) })
	if got := afrB14Price(t, g, me, legend); got.Generic != 2 || len(got.Required) != 0 {
		t.Errorf("level 2 prices a legendary {2}{R}{G} as %+v, want {2}", got)
	}
	if got := afrB14Price(t, g, me, mono); got.Generic != 3 || len(got.Required) != 0 {
		t.Errorf("level 2 prices a legendary {3}{G} as %+v, want {3}: the green is taken off and nothing else", got)
	}
	if got := afrB14Price(t, g, me, plain); got.Generic != 2 || len(got.Required) != 2 {
		t.Errorf("a nonlegendary spell is reduced: %+v", got)
	}
}

func TestAfrB14BardClassLevelThreeExilesTwoWhenYouCastALegendarySpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	afrB14Class(g, me, "Bard Class", afrB14BardClass, 3)
	one := effectstest.TopLibraryCard(me, "One", "Sorcery")
	two := effectstest.TopLibraryCard(me, "Two", "Land")
	effectstest.CastCatalogSpell(t, g, "Legend", "Legendary Creature — Elf", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, one) || !effectstest.InExile(g, two) {
		t.Fatal("the top two cards were not exiled")
	}
	for _, id := range []uuid.UUID{one, two} {
		if !effectstest.B12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("%s is not playable this turn", id)
		}
	}
	// A nonlegendary spell does nothing.
	three := effectstest.TopLibraryCard(me, "Three", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Commoner", "Creature — Elf", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.InExile(g, three) {
		t.Error("a nonlegendary spell triggered the Class")
	}
}

func TestAfrB14BardClassLevelThreeDoesNothingBelowLevelThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	afrB14Class(g, me, "Bard Class", afrB14BardClass, 2)
	top := effectstest.TopLibraryCard(me, "Top", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Legend", "Legendary Creature — Elf", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.InExile(g, top) {
		t.Error("the level-3 trigger fired at level 2")
	}
}

// --- Monk Class ----------------------------------------------------------

func TestAfrB14MonkClassSecondSpellCostsOneLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	afrB14Class(g, me, "Monk Class", afrB14MonkClass, 1)
	spell := game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", ManaCost: "{2}{W}", Owner: me.ID, Controller: me.ID}
	if got := afrB14Price(t, g, me, spell); got.Generic != 2 {
		t.Errorf("the first spell costs %+v, want full price", got)
	}
	effectstest.CastCatalogSpell(t, g, "First", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := afrB14Price(t, g, me, spell); got.Generic != 1 {
		t.Errorf("the second spell costs %+v, want {1} less", got)
	}
	effectstest.CastCatalogSpell(t, g, "Second", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := afrB14Price(t, g, me, spell); got.Generic != 2 {
		t.Errorf("the third spell costs %+v, want full price", got)
	}
}

func TestAfrB14MonkClassLevelTwoReturnsUpToOneNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	class := afrB14Class(g, me, "Monk Class", afrB14MonkClass, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{W}{U}") })
	if err := g.ActivateCatalogAbility(me.ID, class, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LevelOf(g, class); got != 2 {
		t.Fatalf("level %d, want 2", got)
	}
	if effectstest.OnBattlefieldByID(g, theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the target was not returned to its owner's hand")
	}
}

func TestAfrB14MonkClassLevelThreeExilesWithAConditionalCastPermission(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	afrB14Class(g, me, "Monk Class", afrB14MonkClass, 3)
	top := effectstest.TopLibraryCard(me, "Top Spell", "Sorcery")
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, top) {
		t.Fatal("the top card was not exiled at the beginning of the upkeep")
	}
	perm := effectstest.ExiledPermission(g, top)
	if perm.Player != me.ID || !perm.CastOnly || perm.Condition == nil || perm.Condition.Key != game.CondCastASpellThisTurn {
		t.Fatalf("permission %+v: cast only, while exiled, if another spell was cast this turn", perm)
	}
	if effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("playable before another spell was cast this turn")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Other Spell", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("not castable after another spell was cast this turn")
	}
}

func TestAfrB14MonkClassLevelThreeCardStaysCastableOnLaterTurnsOnlyAfterASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	afrB14Class(g, me, "Monk Class", afrB14MonkClass, 3)
	top := effectstest.TopLibraryCard(me, "Top Spell", "Sorcery")
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NextTurnOfMine(t, g)
	if effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("castable on a later turn before any spell was cast")
	}
	effectstest.CastCatalogSpell(t, g, "Other Spell", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the grant did not last for as long as the card remained exiled")
	}
}

// --- Sorcerer Class ------------------------------------------------------

func TestAfrB14SorcererClassEntersAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	before := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Sorcerer Class", "Enchantment — Class", afrB14SorcererClass, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - before; got != 0 && len(g.PendingChoices) == 0 {
		t.Errorf("hand changed by %d with no discard prompt", got)
	}
	if len(g.PendingChoices) == 0 {
		t.Fatal("no discard prompt after drawing two")
	}
}

func TestAfrB14SorcererClassLevelTwoGivesCreaturesRestrictedMana(t *testing.T) {
	for _, level := range []int{1, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		afrB14Class(g, me, "Sorcerer Class", afrB14SorcererClass, level)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Elf", 1, 1)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Elf", 1, 1)
		hasMana := func(id uuid.UUID) bool {
			g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
			c := effectstest.FindBattlefieldCardByID(g, id)
			return c != nil && len(game.ManaAbilitiesOf(c)) > 0
		}
		if got := hasMana(mine); got != (level >= 2) {
			t.Errorf("level %d: my creature has a mana ability: %v", level, got)
		}
		if hasMana(theirs) {
			t.Errorf("level %d: an opponent's creature has a mana ability", level)
		}
	}
}

func TestAfrB14SorcererClassManaPaysForInstantsSorceriesAndClassLevelsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	afrB14Class(g, me, "Sorcerer Class", afrB14SorcererClass, 2)
	other := afrB14Class(g, me, "Wizard Class", effectstest.WizardClassOracle, 1)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	tapForRestrictedMana := func(color string) {
		t.Helper()
		if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{Colors: []string{color}}); err != nil {
			t.Fatalf("tapping the elf: %v", err)
		}
	}
	// {U} from the elf (plus {1}{U}... of ordinary mana) pays Wizard Class's level 2.
	tapForRestrictedMana("U")
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{1}{U}") })
	if err := g.ActivateCatalogAbility(me.ID, other, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("level-up paid with the restricted mana: %v", err)
	}
}

func TestAfrB14SorcererClassLevelThreeDealsDamageEqualToTheSpellsCastThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	afrB14Class(g, me, "Sorcerer Class", afrB14SorcererClass, 3)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Bolt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 1 {
		t.Errorf("the first instant dealt %d, want 1", got)
	}
	effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 3 {
		t.Errorf("after the second spell the opponent has lost %d in all, want 1+2", got)
	}
	effectstest.CastCatalogSpell(t, g, "Creature", "Creature — Elf", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 3 {
		t.Errorf("a creature spell triggered the Class: lost %d", got)
	}
	if got := g.Seats[2].Life; got != 40 && got != g.Seats[3].Life {
		t.Errorf("opponents lost different amounts")
	}
}

// --- Asmodeus the Archfiend ----------------------------------------------

func TestAfrB14AsmodeusExilesInsteadOfDrawing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	asm := effectstest.PushCatalogPermanent(g, me.ID, "Asmodeus the Archfiend", "Legendary Creature — Devil God", afrB14Asmodeus, false)
	top := effectstest.TopLibraryCard(me, "Top", "Sorcery")
	hand := len(me.Hand.Cards)
	_ = g.DrawCard(me.ID)
	if len(me.Hand.Cards) != hand {
		t.Error("a card was drawn")
	}
	c := effectstest.CardInExile(t, g, top)
	if !c.FaceDown || c.IsKnownTo(me.ID) {
		t.Errorf("the card must be exiled face down with nobody looking: face down %v, known to me %v", c.FaceDown, c.IsKnownTo(me.ID))
	}
	n := 0
	g.ReadSnapshot(func() {
		n = len(g.ExiledWithForEffect(game.PermissionCardRef{ID: asm, Epoch: effectstest.FindBattlefieldCardByID(g, asm).ObjectEpoch}))
	})
	if n != 1 {
		t.Errorf("%d cards exiled with Asmodeus, want 1", n)
	}
}

func TestAfrB14AsmodeusDrawSevenExilesSevenAndReturnLosesThatMuchLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	asm := effectstest.PushCatalogPermanent(g, me.ID, "Asmodeus the Archfiend", "Legendary Creature — Devil God", afrB14Asmodeus, false)
	hand := len(me.Hand.Cards)
	life := me.Life
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}") })
	if err := g.ActivateCatalogAbility(me.ID, asm, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("draw seven: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand changed by %d, want 0", len(me.Hand.Cards)-hand)
	}
	if got := len(g.Exile.Cards); got != 7 {
		t.Fatalf("%d cards exiled, want 7", got)
	}
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{B}") })
	if err := g.ActivateCatalogAbility(me.ID, asm, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("return: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - hand; got != 7 {
		t.Errorf("%d cards returned to hand, want 7", got)
	}
	if got := life - me.Life; got != 7 {
		t.Errorf("lost %d life, want 7", got)
	}
	if len(g.Exile.Cards) != 0 {
		t.Errorf("%d cards still exiled", len(g.Exile.Cards))
	}
}

func TestAfrB14AsmodeusReturnsEachCardToItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	asm := effectstest.PushCatalogPermanent(g, me.ID, "Asmodeus the Archfiend", "Legendary Creature — Devil God", afrB14Asmodeus, false)
	mine := effectstest.TopLibraryCard(me, "Mine", "Sorcery")
	_ = g.DrawCard(me.ID)
	// A card another player owns, exiled with Asmodeus (a stolen draw).
	theirs := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: theirs, Name: "Theirs", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() { _, _ = g.ExileFaceDownLinkedForEffect(asm, me.ID, theirs) })
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{B}") })
	if err := g.ActivateCatalogAbility(me.ID, asm, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("return: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Errorf("cards did not reach their owners' hands: mine %v, theirs %v", me.Hand.Contains(mine), opp.Hand.Contains(theirs))
	}
}

func TestAfrB14AsmodeusReplacesOnlyItsControllersDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Asmodeus the Archfiend", "Legendary Creature — Devil God", afrB14Asmodeus, false)
	hand := len(opp.Hand.Cards)
	_ = g.DrawCard(opp.ID)
	if len(opp.Hand.Cards) != hand+1 {
		t.Error("an opponent's draw was replaced")
	}
}
