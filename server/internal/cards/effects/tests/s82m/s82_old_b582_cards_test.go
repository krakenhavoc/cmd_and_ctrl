package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b582_cards_test.go — Ravnica: City of Guilds, batch 582 (S82 older
// sets, #5600). Every test plays the real catalog entry through the engine.

const (
	oldB582DryadsCaress   = "7df07d16-e5cf-4c58-980a-632909aa1357"
	oldB582PsychicDrain   = "ac5427d7-9af9-4f94-a8c3-da019ec0b5c5"
	oldB582DarkHeart      = "c44f40da-867e-4237-b4b1-ed6feb1f37b7"
	oldB582BloodletterQ   = "79cf9c52-56e0-4237-a595-ec07824e5eed"
	oldB582TrophyHunter   = "d2c68c15-761e-4aea-89d8-f7fe8f44905c"
	oldB582PeregrineMask  = "a742308c-4047-487f-b95b-0fa5bec2f09a"
	oldB582GriftersBlade  = "fe8b7835-a9b2-4a8d-964a-471430f1d143"
	oldB582Spelunkers     = "0534a514-2847-4dbc-8178-6a2886c75bf6"
	oldB582Gargoyle       = "7b175a33-a05c-49fa-802a-36f7013c892a"
	oldB582InduceParanoia = "ec9475f8-235a-4c85-9b96-9c2085a9b732"
	oldB582SeedSpark      = "fc2e441c-bc0b-439f-af71-daba3b033381"
)

func TestOldB582EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB582DryadsCaress: "Dryad's Caress", oldB582PsychicDrain: "Psychic Drain",
		oldB582DarkHeart: "Dark Heart of the Wood", oldB582BloodletterQ: "Bloodletter Quill",
		oldB582TrophyHunter: "Trophy Hunter", oldB582PeregrineMask: "Peregrine Mask",
		oldB582GriftersBlade: "Grifter's Blade", oldB582Spelunkers: "Goblin Spelunkers",
		oldB582Gargoyle: "Nullstone Gargoyle", oldB582InduceParanoia: "Induce Paranoia",
		oldB582SeedSpark: "Seed Spark",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB582Hand puts a card in the player's hand.
func oldB582Hand(p *game.Player, name, typeLine, oracle, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

// oldB582Cast casts a hand card paying exactly `pool` and lets the stack settle.
func oldB582Cast(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, pool string, params game.CastSpellParams) {
	t.Helper()
	effectstest.AddPool(t, g, p.ID, pool)
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Dryad's Caress ----------------------------------------------------------

func TestOldB582DryadsCaressGainsPerCreatureAndUntapsOnlyWithWhite(t *testing.T) {
	for _, white := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
		effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		effectstest.TapForTest(t, g, mine, other)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Tapped", "Creature — Bear", 2, 2)
		effectstest.TapForTest(t, g, theirs)
		pool := "{G}{G}{C}{C}{C}{C}"
		if white {
			pool = "{G}{G}{W}{C}{C}{C}"
		}
		life := me.Life
		id := oldB582Hand(me, "Dryad's Caress", "Instant", oldB582DryadsCaress, "{4}{G}{G}")
		oldB582Cast(t, g, me, id, pool, game.CastSpellParams{})
		if me.Life != life+4 {
			t.Errorf("white=%v: life %d, want %d (one per creature on the battlefield)", white, me.Life, life+4)
		}
		if tapped := effectstest.IsTapped(g, mine); tapped == white {
			t.Errorf("white=%v: my creature tapped=%v", white, tapped)
		}
		if tapped := effectstest.IsTapped(g, other); tapped == white {
			t.Errorf("white=%v: my other creature tapped=%v", white, tapped)
		}
		if !effectstest.IsTapped(g, theirs) {
			t.Errorf("white=%v: an opponent's creature was untapped", white)
		}
	}
}

// --- Seed Spark --------------------------------------------------------------

func TestOldB582SeedSparkDestroysAndMakesSaprolingsOnlyWithGreen(t *testing.T) {
	for _, green := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
		pool := "{W}{C}{C}{C}"
		if green {
			pool = "{W}{G}{C}{C}"
		}
		id := oldB582Hand(me, "Seed Spark", "Instant", oldB582SeedSpark, "{3}{W}")
		oldB582Cast(t, g, me, id, pool, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(rock)})
		if g.Battlefield.Contains(rock) {
			t.Errorf("green=%v: the artifact survived", green)
		}
		want := 0
		if green {
			want = 2
		}
		if n, _ := effectstest.CountTokensNamed(g, me.ID, "Saproling"); n != want {
			t.Errorf("green=%v: %d Saprolings, want %d", green, n, want)
		}
	}
}

func TestOldB582SeedSparkRefusesACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	id := oldB582Hand(me, "Seed Spark", "Instant", oldB582SeedSpark, "{3}{W}")
	effectstest.AddPool(t, g, me.ID, "{W}{G}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("Seed Spark accepted a nonartifact, nonenchantment target")
	}
}

// --- Psychic Drain -----------------------------------------------------------

func TestOldB582PsychicDrainMillsTheTargetAndGainsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lib, mine, life := opp.Library.Size(), me.Library.Size(), me.Life
	id := oldB582Hand(me, "Psychic Drain", "Sorcery", oldB582PsychicDrain, "{X}{U}{B}")
	oldB582Cast(t, g, me, id, "{U}{B}{C}{C}{C}", game.CastSpellParams{XValue: 3, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if got := lib - opp.Library.Size(); got != 3 {
		t.Errorf("the target milled %d, want 3", got)
	}
	if me.Library.Size() != mine {
		t.Error("the caster's library changed")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

// --- Dark Heart of the Wood --------------------------------------------------

func TestOldB582DarkHeartSacrificesAForestForThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	heart := effectstest.SeedPermanentWithOracle(g, me.ID, "Dark Heart of the Wood", "Enchantment", oldB582DarkHeart)
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	swamp := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	if err := g.ActivateCatalogAbility(me.ID, heart, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{swamp}}); err == nil {
		t.Fatal("a Swamp paid for a Forest sacrifice")
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, heart, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{forest}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) {
		t.Error("the Forest was not sacrificed")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

// --- Bloodletter Quill -------------------------------------------------------

func TestOldB582BloodletterQuillDrawsThenLosesLifePerCounterAndUBRemoves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	quill := effectstest.SeedPermanentWithOracle(g, me.ID, "Bloodletter Quill", "Artifact", oldB582BloodletterQ)
	effectstest.UntapForTest(g, quill)
	hand, life := me.Hand.Size(), me.Life
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, quill, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, quill, "blood"); got != 1 {
		t.Fatalf("blood counters %d, want 1", got)
	}
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d life %d, want %d and %d", me.Hand.Size(), me.Life, hand+1, life-1)
	}
	effectstest.UntapForTest(g, quill)
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, quill, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("second activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life after the second activation %d, want %d (1 then 2)", me.Life, life-3)
	}
	effectstest.AddPool(t, g, me.ID, "{U}{B}")
	if err := g.ActivateCatalogAbility(me.ID, quill, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("remove: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, quill, "blood"); got != 1 {
		t.Errorf("blood counters after removal %d, want 1", got)
	}
	// With none left to remove the ability does nothing, and a pool that
	// can't pay is refused.
	if err := g.ActivateCatalogAbility(me.ID, quill, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the removal was paid with no mana")
	}
}

// --- Trophy Hunter -----------------------------------------------------------

func TestOldB582TrophyHunterPingsFlyersAndGrowsWhenOneDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hunter := effectstest.B12Push(g, me.ID, "Trophy Hunter", "Creature — Human Archer", oldB582TrophyHunter, 2, 3)
	bird := effectstest.B12Push(g, opp.ID, "Little Bird", "Creature — Bird", "", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bird {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ping := func(target uuid.UUID, strict bool) error {
		effectstest.AddPool(t, g, me.ID, "{C}{G}")
		return g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{
			Strict: strict, Targets: effectstest.CardTarget(target),
		})
	}
	if err := ping(bear, true); err == nil {
		t.Fatal("the ping accepted a creature without flying")
	}
	if err := ping(bird, true); err != nil {
		t.Fatalf("ping the flyer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("the flyer survived 1 damage")
	}
	if got := effectstest.CounterCount(g, hunter, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	// A creature the Hunter did not damage dying does not count.
	other := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 1, 1)
	effectstest.KillCreature(t, g, opp.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, hunter, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters after a bystander died %d, want 1", got)
	}
}

// --- Peregrine Mask ----------------------------------------------------------

func TestOldB582PeregrineMaskGrantsDefenderFlyingAndFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mask := effectstest.SeedPermanentWithOracle(g, me.ID, "Peregrine Mask", "Artifact — Equipment", oldB582PeregrineMask)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, kw := range []string{"defender", "flying", "first strike"} {
		if effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the bear has %s before it is equipped", kw)
		}
	}
	if err := effectstest.StrictEquip(t, g, me.ID, mask, bear, "{C}"); err == nil {
		t.Fatal("equip {2} was paid with one mana")
	}
	if err := effectstest.StrictEquip(t, g, me.ID, mask, bear, "{C}"); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"defender", "flying", "first strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the equipped bear lacks %s", kw)
		}
	}
}

// --- Goblin Spelunkers -------------------------------------------------------

func TestOldB582GoblinSpelunkersCantBeBlockedOnlyWhileTheDefenderHasAMountain(t *testing.T) {
	for _, hasMountain := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, them := effectstest.SeatsForCombat(g)
		walker := effectstest.B12Push(g, me.ID, "Goblin Spelunkers", "Creature — Goblin Warrior", oldB582Spelunkers, 2, 2)
		bear := effectstest.B12Creature(g, them.ID, "Bear", "Creature — Bear", 2, 2)
		land := "Plains"
		if hasMountain {
			land = "Mountain"
		}
		effectstest.SeedLandOnBattlefield(g, them.ID, land, "Basic Land — "+land)
		effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
		effectstest.AttackIntoBlocks(t, g, them.ID, walker)
		r := effectstest.BlockRefusal(t, g, bear, walker)
		if hasMountain && r != game.BlockReasonLandwalk {
			t.Errorf("blocked while the defender has a Mountain: reason %q, want landwalk", r)
		}
		if !hasMountain && r != "" {
			t.Errorf("block refused without a Mountain: %q", r)
		}
	}
}

// --- Nullstone Gargoyle ------------------------------------------------------

func TestOldB582NullstoneGargoyleCountersOnlyTheFirstNoncreatureSpellOfTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, opp.ID, "Nullstone Gargoyle", "Artifact Creature — Gargoyle", oldB582Gargoyle, 4, 5)

	// A creature spell is not the first noncreature spell, and does not use it up.
	bear := oldB582Hand(me, "Bear", "Creature — Bear", "", "")
	oldB582Cast(t, g, me, bear, "", game.CastSpellParams{})
	if !g.Battlefield.Contains(bear) {
		t.Fatal("a creature spell was countered")
	}
	first := oldB582Hand(me, "First Rock", "Artifact", "", "")
	oldB582Cast(t, g, me, first, "", game.CastSpellParams{})
	if g.Battlefield.Contains(first) || !me.Graveyard.Contains(first) {
		t.Error("the first noncreature spell of the turn was not countered")
	}
	second := oldB582Hand(me, "Second Rock", "Artifact", "", "")
	oldB582Cast(t, g, me, second, "", game.CastSpellParams{})
	if !g.Battlefield.Contains(second) {
		t.Error("the second noncreature spell of the turn was countered")
	}
}

func TestOldB582NullstoneGargoyleCountsSpellsCastBeforeItArrived(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	first := oldB582Hand(me, "First Rock", "Artifact", "", "")
	oldB582Cast(t, g, me, first, "", game.CastSpellParams{})
	effectstest.B12Push(g, me.ID, "Nullstone Gargoyle", "Artifact Creature — Gargoyle", oldB582Gargoyle, 4, 5)
	second := oldB582Hand(me, "Second Rock", "Artifact", "", "")
	oldB582Cast(t, g, me, second, "", game.CastSpellParams{})
	if !g.Battlefield.Contains(second) {
		t.Error("a spell that was not the turn's first noncreature spell was countered")
	}
}

// --- Induce Paranoia ---------------------------------------------------------

func TestOldB582InduceParanoiaCountersAndMillsOnlyWithBlack(t *testing.T) {
	for _, black := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		spell := oldB582Hand(me, "Big Sorcery", "Sorcery", "", "{3}")
		effectstest.AddPool(t, g, me.ID, "{C}{C}{C}")
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast the sorcery: %v", err)
		}
		pool := "{U}{U}{C}{C}"
		if black {
			pool = "{U}{U}{B}{C}"
		}
		lib, theirs := me.Library.Size(), opp.Library.Size()
		paranoia := oldB582Hand(opp, "Induce Paranoia", "Instant", oldB582InduceParanoia, "{2}{U}{U}")
		oldB582Cast(t, g, opp, paranoia, pool, game.CastSpellParams{Targets: effectstest.CardTarget(spell)})
		if !me.Graveyard.Contains(spell) {
			t.Errorf("black=%v: the spell was not countered", black)
		}
		want := 0
		if black {
			want = 3
		}
		// The countered spell itself lands in the graveyard; the library loses the mill.
		if got := lib - me.Library.Size(); got != want {
			t.Errorf("black=%v: the spell's controller milled %d, want %d", black, got, want)
		}
		if opp.Library.Size() != theirs {
			t.Errorf("black=%v: the caster's library changed", black)
		}
	}
}

// --- Grifter's Blade ---------------------------------------------------------

func TestOldB582GriftersBladeEntersAttachedToTheChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	id := oldB582Hand(me, "Grifter's Blade", "Artifact — Equipment", oldB582GriftersBlade, "{3}")
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID)
	if c == nil {
		t.Fatal("the Blade did not ask for a creature")
	}
	if _, ok := effectstest.BattlefieldCard(g, id); ok {
		t.Fatal("the Blade entered before the choice (CR 614.12)")
	}
	offered := map[uuid.UUID]bool{}
	for _, cand := range c.ChooseCards {
		offered[cand] = true
	}
	if !offered[a] || !offered[b] || offered[theirs] || len(offered) != 2 {
		t.Errorf("candidates %v: want exactly my two creatures", c.ChooseCards)
	}
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, id, b) {
		t.Error("the Blade is not attached to the chosen creature")
	}
	if effectstest.AttachedTo(t, g, id, a) {
		t.Error("the Blade is attached to the other creature")
	}
	if p := effectstest.EffectivePower(t, g, b); p != 3 {
		t.Errorf("the equipped creature's power %d, want 3", p)
	}
}

func TestOldB582GriftersBladeWithNoCreatureEntersUnattachedAndCanBeEquipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := oldB582Hand(me, "Grifter's Blade", "Artifact — Equipment", oldB582GriftersBlade, "{3}")
	oldB582Cast(t, g, me, id, "{C}{C}{C}", game.CastSpellParams{Strict: true})
	if effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID) != nil {
		t.Fatal("asked a question with no creature to name")
	}
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("the Blade is not on the battlefield")
	}
	if c.IsAttached() {
		t.Error("the Blade entered attached with no creature to attach to")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.StrictEquip(t, g, me.ID, id, bear, "{C}"); err != nil {
		t.Fatalf("equip {1}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, id, bear) {
		t.Error("equip did not attach the Blade")
	}
}

func TestOldB582GriftersBladeDoesNotOfferACreatureWithProtectionFromArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	guarded := effectstest.B12Creature(g, me.ID, "Guarded", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == guarded {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, game.ProtectionFromCardType("artifact"))
			}
		}
	})
	id := oldB582Hand(me, "Grifter's Blade", "Artifact — Equipment", oldB582GriftersBlade, "{3}")
	oldB582Cast(t, g, me, id, "{C}{C}{C}", game.CastSpellParams{Strict: true})
	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID)
	if c == nil {
		// A single legal creature is still a choice; none is not.
		t.Fatal("the Blade did not ask for a creature")
	}
	for _, cand := range c.ChooseCards {
		if cand == guarded {
			t.Error("a creature with protection from artifacts was offered")
		}
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != plain {
		t.Errorf("candidates %v, want only the plain creature", c.ChooseCards)
	}
}
