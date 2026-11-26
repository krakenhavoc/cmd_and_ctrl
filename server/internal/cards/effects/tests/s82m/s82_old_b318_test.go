package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b318_test.go — Magic Origins batch 318 (S82, #5549): the spell
// mastery cards. Every test plays the real catalog entry through the engine.

const (
	oldB318PsychicRebuttalOracle    = "08dd393f-12d8-4801-882c-6d76009b4f12"
	oldB318NecromanticSummonsOracle = "2f921740-de98-4276-8196-47a38a1e4591"
	oldB318DarkDabblingOracle       = "342daae1-f7f0-4bce-94bc-e9bdce53e48e"
	oldB318SwiftReckoningOracle     = "82b0c006-bf58-4f05-8801-e5d2857a6fd6"
	oldB318KytheonsTacticsOracle    = "8ecbbf14-68c3-4378-9fce-383f9d901bfa"
	oldB318GatherThePackOracle      = "8f31d870-872d-4671-80d1-8d2ecd2c38d2"
	oldB318SendToSleepOracle        = "9e3fd1e9-7db6-40de-b1de-cd8cc9f60590"
	oldB318TalentOracle             = "a3a26df4-a9d5-48f5-942f-76862bdcd146"
	oldB318CalculatedDismissal      = "d8e63065-9ebc-419d-b7af-5867fd81edff"
	oldB318ExquisiteFirecraft       = "e83c617f-1e07-40db-9b12-ae01122a3e24"
	oldB318GideonsPhalanxOracle     = "f9489831-a9a2-47ab-abae-2f389c9fae2c"
	oldB318LightningStrikeOracle    = "f34b9bc4-7bfe-47fd-ba23-4eeeb46026eb"
)

func TestOldB318EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB318PsychicRebuttalOracle:    "Psychic Rebuttal",
		oldB318NecromanticSummonsOracle: "Necromantic Summons",
		oldB318DarkDabblingOracle:       "Dark Dabbling",
		oldB318SwiftReckoningOracle:     "Swift Reckoning",
		oldB318KytheonsTacticsOracle:    "Kytheon's Tactics",
		oldB318GatherThePackOracle:      "Gather the Pack",
		oldB318SendToSleepOracle:        "Send to Sleep",
		oldB318TalentOracle:             "Talent of the Telepath",
		oldB318CalculatedDismissal:      "Calculated Dismissal",
		oldB318ExquisiteFirecraft:       "Exquisite Firecraft",
		oldB318GideonsPhalanxOracle:     "Gideon's Phalanx",
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

// oldB318Mastery puts n instant cards in p's graveyard; a mastery needs two.
func oldB318Mastery(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.WthGraveyardCard(p, "Spent Spell", "Instant")
	}
}

func oldB318Creature(g *game.Game, owner uuid.UUID, name string, p, tough int) uuid.UUID {
	return effectstest.B12Creature(g, owner, name, "Creature — Bear", p, tough)
}

func TestOldB318CalculatedDismissalCountersUnlessThreeAndScriesWithMastery(t *testing.T) {
	for _, tc := range []struct {
		name    string
		spells  int
		pay     bool
		scrying bool
	}{
		{"no mastery, unpaid", 1, false, false},
		{"mastery, unpaid", 2, false, true},
		{"mastery, paid", 2, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			oldB318Mastery(opp, tc.spells)
			effectstest.ToMainForCost(t, g)
			victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
			effectstest.CastFromSeat(t, g, opp, "Calculated Dismissal", "Instant", oldB318CalculatedDismissal, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
			if tc.pay {
				effectstest.B06AddMana(me, "C", "C", "C")
			}
			effectstest.PassUntilTaxed(t, g, me.ID)
			effectstest.AnswerPayUnless(t, g, me.ID, tc.pay)
			if tc.scrying {
				// The scry is asked as the spell resolves, before anything else.
				c := effectstest.ScryChoiceFor(g, opp.ID)
				if c == nil {
					t.Fatal("no scry prompt")
				}
				if len(c.ScryCards) != 2 {
					t.Errorf("scrying %d cards, want 2", len(c.ScryCards))
				}
				if err := g.ResolveScry(c.ID, opp.ID, nil, c.ScryCards); err != nil {
					t.Fatalf("ResolveScry: %v", err)
				}
			} else if effectstest.ScryChoiceFor(g, opp.ID) != nil {
				t.Fatal("a scry prompt without spell mastery")
			}
			effectstest.PassPriorityAroundTable(t, g)
			_, resolved := effectstest.BattlefieldCard(g, victim)
			if tc.pay != resolved {
				t.Errorf("paid %v but the spell resolved %v", tc.pay, resolved)
			}
		})
	}
}

func TestOldB318DarkDabblingRegeneratesTheTargetAndWithMasteryTheOthers(t *testing.T) {
	for _, tc := range []struct {
		name   string
		spells int
		others int
	}{{"no mastery", 1, 0}, {"mastery", 2, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := g.Seats[0], g.Seats[1]
			oldB318Mastery(me, tc.spells)
			target := oldB318Creature(g, me.ID, "Target", 2, 2)
			other := oldB318Creature(g, me.ID, "Other", 2, 2)
			theirs := oldB318Creature(g, opp.ID, "Theirs", 2, 2)
			hand := me.Hand.Size()
			effectstest.CastCatalogSpell(t, g, "Dark Dabbling", "Instant", oldB318DarkDabblingOracle, effectstest.CardRefs(target))
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.LayeredCard(t, g, target).RegenerationShields; got != 1 {
				t.Errorf("the target has %d regeneration shields, want 1", got)
			}
			if got := effectstest.LayeredCard(t, g, other).RegenerationShields; got != tc.others {
				t.Errorf("the other creature has %d shields, want %d", got, tc.others)
			}
			if got := effectstest.LayeredCard(t, g, theirs).RegenerationShields; got != 0 {
				t.Errorf("an opponent's creature has %d shields, want 0", got)
			}
			if me.Hand.Size() != hand+1 {
				t.Errorf("hand is %d, want %d", me.Hand.Size(), hand+1)
			}
		})
	}
}

func TestOldB318SwiftReckoningDestroysATappedCreatureAndFlashesWithMastery(t *testing.T) {
	t.Run("destroys only a tapped creature", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		opp := g.Seats[1]
		tapped := oldB318Creature(g, opp.ID, "Tapped", 2, 2)
		untapped := oldB318Creature(g, opp.ID, "Untapped", 2, 2)
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
		if err := effectstest.CastCatalogSpellErrStrict(t, g, "Swift Reckoning", "Sorcery", oldB318SwiftReckoningOracle, effectstest.CardRefs(untapped)); err == nil {
			t.Fatal("Swift Reckoning targeted an untapped creature")
		}
		effectstest.CastCatalogSpell(t, g, "Swift Reckoning", "Sorcery", oldB318SwiftReckoningOracle, effectstest.CardRefs(tapped))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(tapped) || !g.Battlefield.Contains(untapped) {
			t.Error("the tapped creature should be destroyed and the untapped one left")
		}
	})
	for _, tc := range []struct {
		name   string
		spells int
		flash  bool
	}{{"no mastery has no flash", 1, false}, {"mastery has flash", 2, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			oldB318Mastery(opp, tc.spells)
			victim := oldB318Creature(g, me.ID, "Victim", 2, 2)
			g.WithWriteLock(func() { _ = g.TapTargetForEffect(victim) })
			effectstest.ToMainForCost(t, g)
			effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
			id := uuid.New()
			opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Swift Reckoning", TypeLine: "Sorcery", OracleID: oldB318SwiftReckoningOracle, Owner: opp.ID, Controller: opp.ID})
			err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: effectstest.CardRefs(victim), Strict: true})
			if tc.flash && err != nil {
				t.Fatalf("with spell mastery it may be cast at instant speed: %v", err)
			}
			if !tc.flash && err == nil {
				t.Fatal("without spell mastery it was cast at instant speed")
			}
		})
	}
}

func TestOldB318KytheonsTacticsPumpsAndWithMasteryGivesVigilance(t *testing.T) {
	for _, tc := range []struct {
		name      string
		spells    int
		vigilance bool
	}{{"no mastery", 1, false}, {"mastery", 2, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := g.Seats[0], g.Seats[1]
			oldB318Mastery(me, tc.spells)
			mine := oldB318Creature(g, me.ID, "Mine", 2, 2)
			theirs := oldB318Creature(g, opp.ID, "Theirs", 2, 2)
			effectstest.CastCatalogSpell(t, g, "Kytheon's Tactics", "Sorcery", oldB318KytheonsTacticsOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.HasKeywordOnBattlefield(t, g, mine, "vigilance"); got != tc.vigilance {
				t.Errorf("vigilance = %v, want %v", got, tc.vigilance)
			}
			if p, tough := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 4 || tough != 3 {
				t.Errorf("my creature is %d/%d, want 4/3", p, tough)
			}
			if effectstest.EffectivePower(t, g, theirs) != 2 || effectstest.EffectiveToughness(t, g, theirs) != 2 || effectstest.HasKeywordOnBattlefield(t, g, theirs, "vigilance") {
				t.Errorf("an opponent's creature changed")
			}
		})
	}
}

func TestOldB318SendToSleepTapsAndWithMasteryKeepsThemTapped(t *testing.T) {
	for _, tc := range []struct {
		name   string
		spells int
		frozen int
	}{{"no mastery", 1, 0}, {"mastery", 2, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := g.Seats[0], g.Seats[1]
			oldB318Mastery(me, tc.spells)
			a := oldB318Creature(g, me.ID, "A", 2, 2)
			b := oldB318Creature(g, opp.ID, "B", 2, 2)
			c := oldB318Creature(g, opp.ID, "C", 2, 2)
			effectstest.CastCatalogSpell(t, g, "Send to Sleep", "Instant", oldB318SendToSleepOracle, effectstest.CardRefs(a, b))
			effectstest.PassPriorityAroundTable(t, g)
			for _, id := range []uuid.UUID{a, b} {
				card, _ := effectstest.BattlefieldCard(g, id)
				if !card.Tapped || len(card.NextUntapSkips) != tc.frozen {
					t.Errorf("target: tapped %v, %d untap markers, want tapped and %d", card.Tapped, len(card.NextUntapSkips), tc.frozen)
				}
			}
			if card, _ := effectstest.BattlefieldCard(g, c); card.Tapped {
				t.Error("an untargeted creature was tapped")
			}
		})
	}
}

func TestOldB318ExquisiteFirecraftDealsFourAndIsUncounterableWithMastery(t *testing.T) {
	for _, tc := range []struct {
		name      string
		spells    int
		countered bool
	}{{"no mastery is countered", 1, true}, {"mastery can't be countered", 2, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			oldB318Mastery(me, tc.spells)
			effectstest.ToMainForCost(t, g)
			life := opp.Life
			fire := effectstest.CastCatalogSpell(t, g, "Exquisite Firecraft", "Sorcery", oldB318ExquisiteFirecraft, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
			effectstest.CastFromSeat(t, g, opp, "Mana Leak", "Instant", "c61fe162-2202-4e56-9ba0-393547f9875f", []game.TargetRef{{Kind: game.TargetCard, ID: fire}})
			effectstest.PassUntilTaxed(t, g, me.ID)
			effectstest.AnswerPayUnless(t, g, me.ID, false)
			effectstest.PassPriorityAroundTable(t, g)
			wantLife := life - 4
			if tc.countered {
				wantLife = life
			}
			if opp.Life != wantLife {
				t.Errorf("opponent is at %d, want %d", opp.Life, wantLife)
			}
		})
	}
}

func TestOldB318GideonsPhalanxMakesFourKnightsAndWithMasteryIndestructible(t *testing.T) {
	for _, tc := range []struct {
		name         string
		spells       int
		indestructed bool
	}{{"no mastery", 1, false}, {"mastery", 2, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := g.Seats[0], g.Seats[1]
			oldB318Mastery(me, tc.spells)
			mine := oldB318Creature(g, me.ID, "Mine", 2, 2)
			theirs := oldB318Creature(g, opp.ID, "Theirs", 2, 2)
			effectstest.CastCatalogSpell(t, g, "Gideon's Phalanx", "Instant", oldB318GideonsPhalanxOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			var knightIDs []uuid.UUID
			g.ReadSnapshot(func() {
				for _, c := range g.Battlefield.Cards {
					if c.Name != "Knight" || c.Controller != me.ID {
						continue
					}
					knightIDs = append(knightIDs, c.InstanceID)
					if !c.IsToken() || c.Power != 2 || c.Toughness != 2 {
						t.Errorf("Knight is %d/%d token=%v", c.Power, c.Toughness, c.IsToken())
					}
				}
			})
			if len(knightIDs) != 4 {
				t.Fatalf("%d Knights, want 4", len(knightIDs))
			}
			if got := effectstest.HasKeywordOnBattlefield(t, g, mine, "indestructible"); got != tc.indestructed {
				t.Errorf("my creature indestructible = %v, want %v", got, tc.indestructed)
			}
			if effectstest.HasKeywordOnBattlefield(t, g, theirs, "indestructible") {
				t.Error("an opponent's creature became indestructible")
			}
			for _, id := range knightIDs {
				if !effectstest.HasKeywordOnBattlefield(t, g, id, "vigilance") || effectstest.HasKeywordOnBattlefield(t, g, id, "indestructible") != tc.indestructed {
					t.Error("a Knight's vigilance or indestructible is wrong")
				}
			}
		})
	}
}

func TestOldB318GatherThePackTakesOneCreatureOrTwoWithMasteryAndMillsTheRest(t *testing.T) {
	for _, tc := range []struct {
		name   string
		spells int
		take   int
	}{{"no mastery", 1, 1}, {"mastery", 2, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[0]
			oldB318Mastery(me, tc.spells)
			deep := effectstest.PutOnTopOfLibrary(g, me, "Deep Card", "Creature — Bear", "")
			var top []uuid.UUID
			var creatures []uuid.UUID
			for i, tl := range []string{"Land", "Creature — Bear", "Instant", "Creature — Elf", "Creature — Wolf"} {
				id := effectstest.PutOnTopOfLibrary(g, me, "Revealed", tl, "")
				top = append(top, id)
				if i == 1 || i == 3 || i == 4 {
					creatures = append(creatures, id)
				}
			}
			effectstest.CastCatalogSpell(t, g, "Gather the Pack", "Sorcery", oldB318GatherThePackOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.ChooseCardsChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no choose-cards prompt")
			}
			if len(c.ChooseCards) != 3 {
				t.Errorf("%d cards offered, want the 3 creature cards", len(c.ChooseCards))
			}
			if c.ChooseMax != tc.take {
				t.Errorf("may take up to %d, want %d", c.ChooseMax, tc.take)
			}
			picks := creatures[:tc.take]
			effectstest.AnswerChooseCards(t, g, me.ID, picks...)
			for _, id := range picks {
				if !effectstest.HandHasID(me, id) {
					t.Errorf("creature %s is not in hand", id)
				}
			}
			for _, id := range top {
				inHand := effectstest.HandHasID(me, id)
				if !inHand && !me.Graveyard.Contains(id) {
					t.Errorf("revealed card %s is neither in hand nor in the graveyard", id)
				}
			}
			if !me.Library.Contains(deep) {
				t.Error("the sixth card left the library")
			}
		})
	}
}

func TestOldB318NecromanticSummonsReturnsFromAnyGraveyardWithMasteryCounters(t *testing.T) {
	for _, tc := range []struct {
		name     string
		spells   int
		counters int
	}{{"no mastery", 1, 0}, {"mastery", 2, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := g.Seats[0], g.Seats[1]
			oldB318Mastery(me, tc.spells)
			dead := effectstest.WthGraveyardCard(opp, "Dead Bear", "Creature — Bear")
			if err := effectstest.CastCatalogSpellErrStrict(t, g, "Necromantic Summons", "Sorcery", oldB318NecromanticSummonsOracle, effectstest.CardRefs(effectstest.WthGraveyardCard(opp, "Spell", "Sorcery"))); err == nil {
				t.Fatal("it targeted a non-creature card")
			}
			effectstest.CastCatalogSpell(t, g, "Necromantic Summons", "Sorcery", oldB318NecromanticSummonsOracle, effectstest.CardRefs(dead))
			effectstest.PassPriorityAroundTable(t, g)
			c, ok := effectstest.BattlefieldCard(g, dead)
			if !ok {
				t.Fatal("the creature did not enter")
			}
			if c.Controller != me.ID {
				t.Error("it did not enter under the caster's control")
			}
			if got := c.Counters[game.CounterPlusOne]; got != tc.counters {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.counters)
			}
		})
	}
}

func TestOldB318PsychicRebuttalCountersASpellAimedAtYouAndMayCopyIt(t *testing.T) {
	setup := func(t *testing.T, spells int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		oldB318Mastery(me, spells)
		effectstest.ToMain(t, g)
		strike := effectstest.G2CastFrom(t, g, opp, "Lightning Strike", "Instant", oldB318LightningStrikeOracle,
			game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}})
		return g, me, opp, strike
	}
	t.Run("refuses a spell that does not target you", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.ToMain(t, g)
		strike := effectstest.G2CastFrom(t, g, opp, "Lightning Strike", "Instant", oldB318LightningStrikeOracle,
			game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
		if err := effectstest.CastCatalogSpellErrStrict(t, g, "Psychic Rebuttal", "Instant", oldB318PsychicRebuttalOracle, effectstest.CardRefs(strike)); err == nil {
			t.Fatal("Psychic Rebuttal targeted a spell that does not target its caster")
		}
		_ = me
	})
	t.Run("no mastery counters and does not copy", func(t *testing.T) {
		g, me, opp, strike := setup(t, 1)
		life, mine := opp.Life, me.Life
		effectstest.CastCatalogSpell(t, g, "Psychic Rebuttal", "Instant", oldB318PsychicRebuttalOracle, effectstest.CardRefs(strike))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Stack.Contains(strike) || me.Life != mine || opp.Life != life {
			t.Errorf("the spell was not countered cleanly (life %d/%d)", me.Life, opp.Life)
		}
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) != nil {
			t.Error("a copy was offered without spell mastery")
		}
	})
	t.Run("mastery copies it with new targets", func(t *testing.T) {
		g, me, opp, strike := setup(t, 2)
		life, mine := opp.Life, me.Life
		effectstest.CastCatalogSpell(t, g, "Psychic Rebuttal", "Instant", oldB318PsychicRebuttalOracle, effectstest.CardRefs(strike))
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerConfirmFor(t, g, me, true)
		pick := effectstest.LatestPickTarget(g, me.ID)
		if pick == nil {
			t.Fatal("no new-target choice for the copy")
		}
		if err := g.ResolvePickTarget(pick.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != mine {
			t.Errorf("the caster took damage: %d", me.Life)
		}
		if opp.Life != life-3 {
			t.Errorf("the copy did not hit the opponent: %d, want %d", opp.Life, life-3)
		}
	})
	t.Run("mastery may decline the copy", func(t *testing.T) {
		g, me, opp, strike := setup(t, 2)
		life := opp.Life
		effectstest.CastCatalogSpell(t, g, "Psychic Rebuttal", "Instant", oldB318PsychicRebuttalOracle, effectstest.CardRefs(strike))
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerConfirmFor(t, g, me, false)
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != life {
			t.Errorf("a copy resolved after the caster declined: %d", opp.Life)
		}
	})
}

func TestOldB318TalentOfTheTelepath(t *testing.T) {
	setup := func(t *testing.T, spells int) (g *game.Game, me, opp *game.Player, strikeA, strikeB, creature, tail uuid.UUID, revealed []uuid.UUID) {
		g = effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp = g.Seats[0], g.Seats[1]
		oldB318Mastery(me, spells)
		tail = effectstest.PutOnTopOfLibrary(g, opp, "Unrevealed", "Creature — Bear", "")
		for i := 0; i < 4; i++ {
			revealed = append(revealed, effectstest.PutOnTopOfLibrary(g, opp, "Filler", "Land", ""))
		}
		creature = effectstest.PutOnTopOfLibrary(g, opp, "Their Creature", "Creature — Bear", "")
		strikeB = effectstest.PutOnTopOfLibrary(g, opp, "Lightning Strike", "Instant", oldB318LightningStrikeOracle)
		strikeA = effectstest.PutOnTopOfLibrary(g, opp, "Lightning Strike", "Instant", oldB318LightningStrikeOracle)
		revealed = append(revealed, creature, strikeB, strikeA)
		effectstest.CastCatalogSpell(t, g, "Talent of the Telepath", "Sorcery", oldB318TalentOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
		effectstest.PassPriorityAroundTable(t, g)
		return
	}
	castAt := func(t *testing.T, g *game.Game, me *game.Player, card uuid.UUID, target uuid.UUID) {
		t.Helper()
		if err := effectstest.CastThroughResolutionCast(t, g, me.ID, card, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(target)}}); err != nil {
			t.Fatalf("cast through the offer: %v", err)
		}
	}
	t.Run("one spell, then the rest go to the graveyard", func(t *testing.T) {
		g, me, opp, a, b, creature, tail, revealed := setup(t, 1)
		effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, a, b)
		if offer := effectstest.ResolutionCastOfferFor(g, me.ID); slices.Contains(offer.DiscardOptions, creature) {
			t.Error("a creature card was offered")
		}
		life := opp.Life
		castAt(t, g, me, a, opp.ID)
		if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
			t.Fatal("a second cast was offered without spell mastery")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != life-3 {
			t.Errorf("the cast Strike left the opponent at %d, want %d", opp.Life, life-3)
		}
		for _, id := range revealed {
			if id == a {
				continue
			}
			if !opp.Graveyard.Contains(id) {
				t.Errorf("revealed card %s did not go to the opponent's graveyard", id)
			}
		}
		if !opp.Library.Contains(tail) {
			t.Error("the eighth card left the library")
		}
	})
	t.Run("mastery casts a second spell", func(t *testing.T) {
		g, me, opp, a, b, _, _, revealed := setup(t, 2)
		effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, a, b)
		castAt(t, g, me, a, opp.ID)
		effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, b)
		castAt(t, g, me, b, opp.ID)
		if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
			t.Fatal("a third cast was offered")
		}
		life := opp.Life
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != life-6 {
			t.Errorf("opponent at %d, want %d after two Strikes", opp.Life, life-6)
		}
		for _, id := range revealed {
			if id != a && id != b && !opp.Graveyard.Contains(id) {
				t.Errorf("revealed card %s did not go to the graveyard", id)
			}
		}
	})
	t.Run("declining casts nothing and mills all seven", func(t *testing.T) {
		g, me, opp, a, b, _, _, revealed := setup(t, 2)
		effectstest.DeclineResolutionCast(t, g, me.ID)
		if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
			t.Fatal("asked again after declining")
		}
		effectstest.PassPriorityAroundTable(t, g)
		for _, id := range revealed {
			if !opp.Graveyard.Contains(id) {
				t.Errorf("revealed card %s is not in the graveyard", id)
			}
		}
		_, _ = a, b
	})
}
