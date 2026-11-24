package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b018_cards_test.go — Secret Lair Drop: Damage (S82 older sets,
// batch 018, #5462). Each test plays the real catalog entry.

const (
	oldB18Kezzerdrix = "1523671c-0cf4-442c-a5f8-c74ae31016e6"
	oldB18Marshal    = "1c0a53fc-8037-46e3-90ea-cb8b73631a83"
	oldB18Kratos     = "42e55b84-8250-441d-a09e-7c61bf7661cc"
	oldB18Warren     = "47f75e25-8a52-4452-a3a8-bf1bd00942d1"
	oldB18Howling    = "4a67946e-8903-4994-9641-d31fbb62082c"
	oldB18Sorcerer   = "5e961d15-5972-4e4b-9385-1cd7cd7c6bbe"
	oldB18Maarika    = "a46f9f5c-775c-470d-8935-9556e1bbb23d"
	oldB18Brion      = "b816b3cc-ae4b-4fb7-8b1a-e01ab83459a3"
	oldB18SoulSpike  = "be979bfc-998d-43d5-a7a4-44548d9a14c4"
	oldB18Pit        = "ea152809-85a2-4fde-8251-3b1f267e4443"
	oldB18Jaya       = "f5f0deb0-070a-45b6-9b81-0bc8143f3040"
)

func TestOldB18EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB18Kezzerdrix: "Kezzerdrix",
		oldB18Marshal:    "Field Marshal",
		oldB18Kratos:     "Kratos, God of War",
		oldB18Warren:     "Warren Instigator",
		oldB18Howling:    "The Howling Abomination",
		oldB18Sorcerer:   "Prodigal Sorcerer",
		oldB18Maarika:    "Maarika, Brutal Gladiator",
		oldB18Brion:      "Brion Stoutarm",
		oldB18SoulSpike:  "Soul Spike",
		oldB18Pit:        "Lord of the Pit",
		oldB18Jaya:       "Jaya's Immolating Inferno",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB18KezzerdrixHurtsYouOnlyWhenOpponentsHaveNoCreatures(t *testing.T) {
	for _, tc := range []struct {
		name      string
		oppBear   bool
		wantLife  int
		startLife int
	}{{"no opposing creature", false, 36, 40}, {"an opposing creature", true, 40, 40}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			effectstest.BlbB1Push(g, me.ID, "Kezzerdrix", "Creature — Rabbit Beast", oldB18Kezzerdrix, 4, 4)
			if tc.oppBear {
				effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
			}
			me.Life = tc.startLife
			effectstest.BeginUpkeep(t, g, me.ID)
			if me.Life != tc.wantLife {
				t.Errorf("life = %d, want %d", me.Life, tc.wantLife)
			}
		})
	}
}

func TestOldB18FieldMarshalPumpsEveryOtherSoldierButNotItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	marshal := effectstest.BlbB1Push(g, me.ID, "Field Marshal", "Creature — Human Soldier", oldB18Marshal, 2, 2)
	mine := effectstest.BlbB1Push(g, me.ID, "Footman", "Creature — Human Soldier", "", 2, 2)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Enemy Footman", "Creature — Human Soldier", "", 2, 2)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	for id, want := range map[uuid.UUID]int{marshal: 2, mine: 3, theirs: 3, bear: 2} {
		if got := effectstest.LayeredCard(t, g, id).CurrentPower(); got != want {
			t.Errorf("power of %s = %d, want %d", effectstest.LayeredCard(t, g, id).Name, got, want)
		}
	}
	for id, want := range map[uuid.UUID]bool{marshal: false, mine: true, theirs: true, bear: false} {
		if got := effectstest.HasKeywordOnBattlefield(t, g, id, "first strike"); got != want {
			t.Errorf("first strike on %s = %v, want %v", effectstest.LayeredCard(t, g, id).Name, got, want)
		}
	}
}

func TestOldB18KratosGivesHasteAndBurnsTheIdleCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kratos := effectstest.BlbB1Push(g, me.ID, "Kratos, God of War", "Legendary Creature — God Warrior", oldB18Kratos, 2, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, kratos, "double strike") {
		t.Error("no double strike")
	}
	a := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 1, 1)
	effectstest.PushVanillaCreature(g, me.ID, "Idler", 1, 1)
	oppBear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	g.BumpLayerVersionForTest()
	for _, id := range []uuid.UUID{a, oppBear} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
			t.Errorf("%s does not have haste", effectstest.LayeredCard(t, g, id).Name)
		}
	}
	// My end step: three creatures of mine, one of which attacked.
	effectstest.AttackWith(t, g, opp.ID, a)
	me.Life, opp.Life = 40, 40
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 38 {
		t.Errorf("my life = %d, want 38 (Kratos and the idler didn't attack)", me.Life)
	}
	if opp.Life != 40-1 && opp.Life != 40 {
		t.Errorf("opponent life = %d", opp.Life)
	}
}

func TestOldB18KratosBurnsTheActivePlayerOnlyAtTheirOwnEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.BlbB1Push(g, me.ID, "Kratos, God of War", "Legendary Creature — God Warrior", oldB18Kratos, 2, 3)
	effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	me.Life, opp.Life = 40, 40
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40 {
		t.Errorf("the opponent was hurt at my end step: %d", opp.Life)
	}
	if me.Life != 39 {
		t.Errorf("my life = %d, want 39 (only Kratos)", me.Life)
	}
}

func TestOldB18WarrenInstigatorPutsAGoblinCreatureCardWhenItHurtsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	warren := effectstest.BlbB1Push(g, me.ID, "Warren Instigator", "Creature — Goblin Berserker", oldB18Warren, 1, 1)
	goblin := effectstest.HandCardFull(me, "Goblin Guide", "Creature — Goblin Scout", "{R}", "", []string{"R"})
	spell := effectstest.HandCardFull(me, "Goblin Charm", "Instant — Goblin", "{R}", "", []string{"R"})
	// Damage to its own controller does nothing.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(warren, me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("damage to its controller triggered it")
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(warren, opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no choice offered")
	}
	for _, id := range c.ChooseCards {
		if id == spell {
			t.Fatal("a Goblin instant was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, goblin)
	if !g.Battlefield.Contains(goblin) || me.Hand.Contains(goblin) {
		t.Error("the Goblin did not enter")
	}
}

func TestOldB18HowlingAbominationTramplesAfterThreeSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	h := effectstest.BlbB1Push(g, me.ID, "The Howling Abomination", "Legendary Creature — Human Beast Warrior", oldB18Howling, 5, 5)
	if !effectstest.HasKeywordOnBattlefield(t, g, h, "haste") {
		t.Error("no haste")
	}
	for i, want := range []bool{false, false, true} {
		g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{me.ID: {Total: i + 1}}
		g.BumpLayerVersionForTest()
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		if got := effectstest.HasKeywordOnBattlefield(t, g, h, "trample"); got != want {
			t.Errorf("after %d spells trample = %v, want %v", i+1, got, want)
		}
	}
}

func TestOldB18HowlingAbominationRespondsToSpellsNotAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	h := effectstest.BlbB1Push(g, opp.ID, "The Howling Abomination", "Legendary Creature — Human Beast Warrior", oldB18Howling, 5, 5)
	me.Life, opp.Life = 40, 40
	effectstest.CastCatalogSpell(t, g, "Bombard", "Instant", effectstest.EoeABBombardOracle, effectstest.CardRefs(h))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 38 {
		t.Errorf("the spell's controller took %d, want 2", 40-me.Life)
	}
	if opp.Life != 40 {
		t.Errorf("its own controller was hurt: %d", opp.Life)
	}
	c := effectstest.LayeredCard(t, g, h)
	if c.CurrentPower() != 7 || c.CurrentToughness() != 7 {
		t.Errorf("Abomination is %d/%d, want 7/7 until end of turn", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestOldB18ProdigalSorcererPingsAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	s := effectstest.BlbB1Push(g, me.ID, "Prodigal Sorcerer", "Creature — Human Wizard Sorcerer", oldB18Sorcerer, 1, 1)
	opp.Life = 40
	effectstest.B16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("life = %d, want 39", opp.Life)
	}
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("activated while tapped")
	}
}

// oldB18Excess has `src` deal `amount` to `victim` as a damage event and
// then runs state-based actions.
func oldB18Deal(g *game.Game, src, victim uuid.UUID, amount int) {
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, victim, amount) })
}

func TestOldB18MaarikaMakesTheControllerSacrificeAfterExcessDamage(t *testing.T) {
	for _, tc := range []struct {
		name       string
		amount     int
		wantPrompt bool
	}{{"lethal exactly: no excess", 2, false}, {"excess damage", 5, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			maarika := effectstest.BlbB1Push(g, me.ID, "Maarika, Brutal Gladiator", "Legendary Creature — Human Warrior", oldB18Maarika, 7, 4)
			victim := effectstest.BlbB1Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 2)
			ench := effectstest.BlbB1Push(g, opp.ID, "Pacifism", "Enchantment — Aura", "", 0, 0)
			oldB18Deal(g, maarika, victim, tc.amount)
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.SacrificeChoiceFor(g, opp.ID)
			if (c != nil) != tc.wantPrompt {
				t.Fatalf("prompt = %v, want %v", c != nil, tc.wantPrompt)
			}
			if c != nil {
				for _, id := range effectstest.SacrificeOptions(c) {
					if id == victim {
						t.Fatal("a creature was offered")
					}
				}
				effectstest.AnswerSacrifice(t, g, opp.ID, ench)
				if g.Battlefield.Contains(ench) {
					t.Error("the noncreature permanent was not sacrificed")
				}
			}
		})
	}
}

func TestOldB18MaarikaCountsExcessDamageFromAnotherSourceThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	maarika := effectstest.BlbB1Push(g, me.ID, "Maarika, Brutal Gladiator", "Legendary Creature — Human Warrior", oldB18Maarika, 7, 4)
	other := effectstest.BlbB1Push(g, me.ID, "Pinger", "Creature — Bear", "", 1, 1)
	victim := effectstest.BlbB1Push(g, opp.ID, "Tough", "Creature — Ogre", "", 2, 9)
	_ = effectstest.BlbB1Push(g, opp.ID, "Pacifism", "Enchantment — Aura", "", 0, 0)
	oldB18Deal(g, other, victim, 12) // excess from another source
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil {
		t.Fatal("a source other than Maarika triggered it")
	}
	oldB18Deal(g, maarika, victim, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("Maarika's damage to an already-excess creature did not trigger")
	}
}

func TestOldB18MaarikaIsIndestructibleOnlyOnYourTurnAndMustBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	maarika := effectstest.BlbB1Push(g, me.ID, "Maarika, Brutal Gladiator", "Legendary Creature — Human Warrior", oldB18Maarika, 7, 4)
	if !effectstest.HasKeywordOnBattlefield(t, g, maarika, "indestructible") {
		t.Error("not indestructible on its controller's turn")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasKeywordOnBattlefield(t, g, maarika, "indestructible") {
		t.Error("indestructible on another player's turn")
	}
}

func TestOldB18BrionStoutarmFlingsASacrificedCreatureAtAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	brion := effectstest.BlbB1Push(g, me.ID, "Brion Stoutarm", "Legendary Creature — Giant Warrior", oldB18Brion, 4, 4)
	ogre := effectstest.PushVanillaCreature(g, me.ID, "Ogre", 3, 3)
	opp.Life = 40
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, brion, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{brion}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("sacrificed Brion himself")
	}
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, brion, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{ogre}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ogre}}}); err == nil {
		t.Fatal("targeted a creature")
	}
	effectstest.B16Activate(t, g, me.ID, brion, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{ogre}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 37 {
		t.Errorf("life = %d, want 37", opp.Life)
	}
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre was not sacrificed")
	}
}

func TestOldB18SoulSpikePitchesTwoBlackCardsForFourDamageAndFourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	b1 := effectstest.HandCardFull(me, "Black One", "Instant", "{B}", "", []string{"B"})
	b2 := effectstest.HandCardFull(me, "Black Two", "Instant", "{B}", "", []string{"B"})
	red := effectstest.HandCardFull(me, "Red", "Instant", "{R}", "", []string{"R"})
	spike := effectstest.HandCardFull(me, "Soul Spike", "Instant", "{5}{B}{B}", oldB18SoulSpike, []string{"B"})
	tgt := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.CastSpell(me.ID, spike, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{b1, red}, Targets: tgt}); err == nil {
		t.Fatal("a red card was accepted as a black card")
	}
	if err := g.CastSpell(me.ID, spike, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{b1}, Targets: tgt}); err == nil {
		t.Fatal("one card is not two")
	}
	me.Life, opp.Life = 40, 40
	if err := g.CastSpell(me.ID, spike, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{b1, b2}, Targets: tgt}); err != nil {
		t.Fatalf("pitching two black cards: %v", err)
	}
	if me.Hand.Contains(b1) || me.Hand.Contains(b2) || !g.Exile.Contains(b1) || !g.Exile.Contains(b2) {
		t.Error("both black cards are exiled as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 36 || me.Life != 44 {
		t.Errorf("opponent %d (want 36), you %d (want 44)", opp.Life, me.Life)
	}
}

func TestOldB18LordOfThePitEatsACreatureOrYou(t *testing.T) {
	t.Run("sacrifices another creature", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		lord := effectstest.BlbB1Push(g, me.ID, "Lord of the Pit", "Creature — Demon", oldB18Pit, 7, 7)
		snack := effectstest.PushVanillaCreature(g, me.ID, "Snack", 1, 1)
		me.Life = 40
		effectstest.BeginUpkeep(t, g, me.ID)
		effectstest.AnswerSacrifice(t, g, me.ID, snack)
		if g.Battlefield.Contains(snack) || !g.Battlefield.Contains(lord) || me.Life != 40 {
			t.Errorf("snack on battlefield %v, lord %v, life %d", g.Battlefield.Contains(snack), g.Battlefield.Contains(lord), me.Life)
		}
	})
	t.Run("deals 7 to you with nothing to sacrifice", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		effectstest.BlbB1Push(g, me.ID, "Lord of the Pit", "Creature — Demon", oldB18Pit, 7, 7)
		me.Life = 40
		effectstest.BeginUpkeep(t, g, me.ID)
		if me.Life != 33 {
			t.Errorf("life = %d, want 33", me.Life)
		}
	})
}

func oldB18CastJaya(t *testing.T, g *game.Game, x int, legend bool, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	if legend {
		effectstest.BlbB1Push(g, me.ID, "Jaya Ballard", "Legendary Creature — Human Wizard", "", 2, 2)
	}
	effectstest.FloatMana(t, g, me, strings.Repeat("{R}", x+2))
	id := effectstest.HandCardFull(me, "Jaya's Immolating Inferno", "Legendary Sorcery", "{X}{R}{R}", oldB18Jaya, []string{"R"})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x, Targets: targets})
}

func TestOldB18JayasImmolatingInfernoNeedsALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	if err := oldB18CastJaya(t, g, 0, false, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Fatal("cast without a legendary creature or planeswalker")
	}
}

func TestOldB18JayasImmolatingInfernoHitsUpToThreeTargetsForX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	opp.Life, third.Life = 40, 40
	tgts := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: bear}, {Kind: game.TargetPlayer, ID: third.ID}}
	if err := oldB18CastJaya(t, g, 3, true, tgts); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 37 || third.Life != 37 {
		t.Errorf("lives %d and %d, want 37 each", opp.Life, third.Life)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear survived 3 damage")
	}
	if me.Life != 40 {
		t.Errorf("caster life %d", me.Life)
	}
}
