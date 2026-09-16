package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s23_boardwipes_test.go — S23: the mass-effect primitives and the
// cards built on them.
//
// The tests worth reading first are the SIMULTANEITY ones. Every
// other assertion here would have passed against the pre-S23
// hand-rolled loops; the Zulaport pair is the pair that would not,
// and it is the reason the sprint touched the engine at all.

const (
	wrathOfGodOracle       = "34515b16-c9a4-4f98-8c77-416a7a523407"
	supremeVerdictOracle   = "0230de18-8d15-4cfa-9d42-7ccddd9f9570"
	fumigateOracle         = "b17ea905-0696-4e58-b564-557e87236e27"
	deadlyTempestOracle    = "b5516bc9-ec8d-4323-8748-96c49d7d0622"
	inGarruksWakeOracle    = "a6899b94-427d-4851-a474-4087e0a0918a"
	ritualOfSootOracle     = "29e9cf1c-a6bd-4bee-9000-ac1b4e19d6b0"
	planarCleansingOracle  = "a98c2d81-4add-4292-bbdd-e1b69ff936d4"
	shatterstormOracle     = "96ce2403-4607-440a-92ae-80aceb458c5d"
	cleansingNovaOracle    = "aff34f28-f707-4458-8af3-1bd5b13a6b10"
	cruxOfFateOracle       = "52a0dae4-2a95-487e-acd4-eabdb2d031e2"
	depopulateOracle       = "4a83c2fa-f2d2-4b86-8407-4269f127936d"
	evacuationOracle       = "fdd94383-b573-439a-8e1c-925af887c5a6"
	riversRebukeOracle     = "c52cfb41-18f3-4e73-b5e7-d75baf74e578"
	whelmingWaveOracle     = "e510eaaf-6497-480f-baa8-f4796b5f1086"
	toxicDelugeOracle      = "afaef788-34d1-460b-b884-9d7ae6ddeb18"
	languishOracle         = "ef1a83f2-6707-41a2-b5ed-861c8e45ae07"
	massacreWurmOracle     = "93cf50cf-0ecc-4d3e-abea-778c1ebacec4"
	chandrasIgnitionOracle = "f61680da-606e-4d16-b0a0-361aa5210901"
	baneOfProgressOracle   = "51f9a6cc-8eb2-44ed-a2d9-913ac514ad67"
	nevinyrralsDiskOracle  = "96230edf-568a-47dd-b877-9d92aa58fac8"
	zulaportOracle         = "76b003e0-15af-4f22-bdf2-1ade5430964a"
)

// pushWipeCreature seeds a plain vanilla creature onto the
// battlefield with an explicit size, since most of these tests care
// about toughness thresholds.
func pushWipeCreature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Power:      power,
		Toughness:  toughness,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// castWipe seeds a sorcery into the active seat's hand and casts it
// with explicit params, for the cards castCatalogSpell's
// targets-only signature cannot express (an X value, a mode).
func castWipe(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracle,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// --- simultaneity: the reason S23 touched the engine ---------------

// A wrath kills every creature AT THE SAME TIME (CR 700.4), so an
// aristocrats payoff that dies in the wipe still sees every other
// creature die — including itself.
//
// Zulaport Cutthroat rather than Blood Artist because Zulaport is
// untargeted, so the assertion is about the batching and not about
// answering four target prompts.
//
// Before S23 this was wrong in a way that depended on battlefield
// ORDER: each DestroyTarget emitted its own LTB and the harvester
// only walked the live battlefield, so a Zulaport processed first
// saw nothing and one processed last saw everything. The number
// below is the rules-correct one and is order-independent.
func TestWrathDeathsAreSimultaneousForAristocratsPayoffs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	// Zulaport is pushed FIRST, so it is early in the battlefield
	// order — the position that used to lose the most triggers.
	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue", zulaportOracle, false)
	pushWipeCreature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	pushWipeCreature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	pushWipeCreature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)

	meBefore, oppBefore := me.Life, opp.Life
	castCatalogSpell(t, g, "Wrath of God", "Sorcery", wrathOfGodOracle, nil)
	passPriorityAroundTable(t, g)

	// Four creatures I controlled died together; Zulaport triggers
	// on each, itself included.
	if want := oppBefore - 4; opp.Life != want {
		t.Errorf("opponent life %d -> %d, want %d (four simultaneous deaths)", oppBefore, opp.Life, want)
	}
	if want := meBefore + 4; me.Life != want {
		t.Errorf("caster life %d -> %d, want %d", meBefore, me.Life, want)
	}
}

// The same guarantee through the OTHER death path: Languish does not
// destroy anything, it shrinks everything and lets the zero-toughness
// state-based action do the killing. That SBA sweep is batched too,
// so the payoff counts the same four deaths.
func TestShrinkWipeDeathsAreSimultaneousToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue", zulaportOracle, false)
	pushWipeCreature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	pushWipeCreature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	pushWipeCreature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)

	meBefore, oppBefore := me.Life, opp.Life
	castCatalogSpell(t, g, "Languish", "Sorcery", languishOracle, nil)
	passPriorityAroundTable(t, g)

	if want := oppBefore - 4; opp.Life != want {
		t.Errorf("opponent life %d -> %d, want %d (four simultaneous SBA deaths)", oppBefore, opp.Life, want)
	}
	if want := meBefore + 4; me.Life != want {
		t.Errorf("caster life %d -> %d, want %d", meBefore, me.Life, want)
	}
}

// --- destroy sweeps -----------------------------------------------

func TestWrathOfGodDestroysEveryCreatureAndNothingElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushWipeCreature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := pushWipeCreature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rock := pushWipeCreature(g, me.ID, "Sol Ring", "Artifact", 0, 0)

	castCatalogSpell(t, g, "Wrath of God", "Sorcery", wrathOfGodOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs} {
		if g.Battlefield.Contains(id) {
			t.Errorf("creature %s survived a wrath", id)
		}
	}
	if !g.Battlefield.Contains(rock) {
		t.Errorf("Wrath of God destroyed a non-creature")
	}
}

// Fumigate's "you gain 1 life for each creature destroyed this way"
// is the primitive's count, not a re-scan of the battlefield.
func TestFumigateGainsOneLifePerCreatureDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushWipeCreature(g, me.ID, "Mine A", "Creature — Bear", 2, 2)
	pushWipeCreature(g, me.ID, "Mine B", "Creature — Bear", 2, 2)
	pushWipeCreature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	pushWipeCreature(g, me.ID, "Rock", "Artifact", 0, 0)

	before := me.Life
	castCatalogSpell(t, g, "Fumigate", "Sorcery", fumigateOracle, nil)
	passPriorityAroundTable(t, g)

	if want := before + 3; me.Life != want {
		t.Errorf("life %d -> %d, want %d (three creatures, the artifact does not count)", before, me.Life, want)
	}
}

// Deadly Tempest charges each player for THEIR OWN board, which
// means the tally has to be taken from the pre-move copies.
func TestDeadlyTempestChargesEachPlayerForTheirOwnBoard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	pushWipeCreature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	pushWipeCreature(g, opp.ID, "Theirs A", "Creature — Bear", 2, 2)
	pushWipeCreature(g, opp.ID, "Theirs B", "Creature — Bear", 2, 2)
	pushWipeCreature(g, opp.ID, "Theirs C", "Creature — Bear", 2, 2)

	meBefore, oppBefore, thirdBefore := me.Life, opp.Life, third.Life
	castCatalogSpell(t, g, "Deadly Tempest", "Sorcery", deadlyTempestOracle, nil)
	passPriorityAroundTable(t, g)

	if want := meBefore - 1; me.Life != want {
		t.Errorf("caster life %d -> %d, want %d", meBefore, me.Life, want)
	}
	if want := oppBefore - 3; opp.Life != want {
		t.Errorf("opponent life %d -> %d, want %d", oppBefore, opp.Life, want)
	}
	if third.Life != thirdBefore {
		t.Errorf("a player with no creatures lost life: %d -> %d", thirdBefore, third.Life)
	}
}

// "You don't control" is a resolution-time read of control, and it
// is what makes a nine-mana sorcery worth casting.
func TestInGarruksWakeSparesYourOwnBoard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushWipeCreature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := pushWipeCreature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	theirWalker := pushWipeCreature(g, opp.ID, "Their Walker", "Planeswalker — Test", 0, 0)
	theirRock := pushWipeCreature(g, opp.ID, "Their Rock", "Artifact", 0, 0)

	castCatalogSpell(t, g, "In Garruk's Wake", "Sorcery", inGarruksWakeOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(mine) {
		t.Errorf("In Garruk's Wake destroyed a creature you control")
	}
	if g.Battlefield.Contains(theirs) {
		t.Errorf("opponent's creature survived")
	}
	if g.Battlefield.Contains(theirWalker) {
		t.Errorf("opponent's planeswalker survived")
	}
	if !g.Battlefield.Contains(theirRock) {
		t.Errorf("In Garruk's Wake hit an artifact, which it does not mention")
	}
}

func TestRitualOfSootSparesTheExpensiveHalfOfTheBoard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cheap := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: cheap, Name: "Cheap", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	pricey := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: pricey, Name: "Pricey", TypeLine: "Creature — Wurm",
		ManaCost: "{4}{G}{G}", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Ritual of Soot", "Sorcery", ritualOfSootOracle, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(cheap) {
		t.Errorf("a mana-value-2 creature survived Ritual of Soot")
	}
	if !g.Battlefield.Contains(pricey) {
		t.Errorf("a mana-value-6 creature was destroyed by Ritual of Soot")
	}
}

func TestPlanarCleansingLeavesOnlyLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	creature := pushWipeCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := pushWipeCreature(g, me.ID, "Rock", "Artifact", 0, 0)
	ench := pushWipeCreature(g, me.ID, "Aura", "Enchantment", 0, 0)
	land := pushWipeCreature(g, me.ID, "Island", "Basic Land — Island", 0, 0)

	castCatalogSpell(t, g, "Planar Cleansing", "Sorcery", planarCleansingOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{creature, rock, ench} {
		if g.Battlefield.Contains(id) {
			t.Errorf("nonland permanent %s survived Planar Cleansing", id)
		}
	}
	if !g.Battlefield.Contains(land) {
		t.Errorf("Planar Cleansing destroyed a land")
	}
}

func TestShatterstormDestroysOnlyArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := pushWipeCreature(g, me.ID, "Rock", "Artifact", 0, 0)
	bear := pushWipeCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Shatterstorm", "Sorcery", shatterstormOracle, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Errorf("artifact survived Shatterstorm")
	}
	if !g.Battlefield.Contains(bear) {
		t.Errorf("Shatterstorm destroyed a creature")
	}
}

// --- exclusions ----------------------------------------------------

// "Except for Krakens, Leviathans, Octopuses, and Serpents" — the
// Except / AnySubtype pair, read against effective subtypes.
func TestWhelmingWaveSparesTheSeaMonsters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushWipeCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	kraken := pushWipeCreature(g, me.ID, "Kraken", "Creature — Kraken", 9, 9)
	serpent := pushWipeCreature(g, me.ID, "Serpent", "Creature — Serpent", 5, 5)

	castCatalogSpell(t, g, "Whelming Wave", "Sorcery", whelmingWaveOracle, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Errorf("a Bear survived Whelming Wave")
	}
	if !me.Hand.Contains(bear) {
		t.Errorf("the bounced Bear did not reach its owner's hand")
	}
	for _, id := range []uuid.UUID{kraken, serpent} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("an excluded sea monster %s was bounced anyway", id)
		}
	}
}

// The two Crux of Fate modes partition the creatures exactly, which
// is what building both from one Subtype("Dragon") predicate buys.
func TestCruxOfFateModesArePerfectlyComplementary(t *testing.T) {
	for _, tc := range []struct {
		name        string
		mode        int
		dragonLives bool
		bearLives   bool
	}{
		{"destroy all Dragons", 0, false, true},
		{"destroy all non-Dragons", 1, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			dragon := pushWipeCreature(g, me.ID, "Dragon", "Creature — Dragon", 5, 5)
			bear := pushWipeCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

			castWipe(t, g, "Crux of Fate", "Sorcery", cruxOfFateOracle,
				game.CastSpellParams{Modes: []int{tc.mode}})
			passPriorityAroundTable(t, g)

			if got := g.Battlefield.Contains(dragon); got != tc.dragonLives {
				t.Errorf("dragon alive = %v, want %v", got, tc.dragonLives)
			}
			if got := g.Battlefield.Contains(bear); got != tc.bearLives {
				t.Errorf("bear alive = %v, want %v", got, tc.bearLives)
			}
		})
	}
}

func TestCleansingNovaSecondModeTakesArtifactsAndEnchantmentsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushWipeCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := pushWipeCreature(g, me.ID, "Rock", "Artifact", 0, 0)
	ench := pushWipeCreature(g, me.ID, "Aura", "Enchantment", 0, 0)

	castWipe(t, g, "Cleansing Nova", "Sorcery", cleansingNovaOracle,
		game.CastSpellParams{Modes: []int{1}})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Errorf("mode 2 destroyed a creature")
	}
	for _, id := range []uuid.UUID{rock, ench} {
		if g.Battlefield.Contains(id) {
			t.Errorf("permanent %s survived mode 2", id)
		}
	}
}

// --- bounce --------------------------------------------------------

func TestEvacuationReturnsEveryCreatureToItsOwnersHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushWipeCreature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := pushWipeCreature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rock := pushWipeCreature(g, me.ID, "Rock", "Artifact", 0, 0)

	castCatalogSpell(t, g, "Evacuation", "Instant", evacuationOracle, nil)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(mine) {
		t.Errorf("caster's creature did not return to their hand")
	}
	if !opp.Hand.Contains(theirs) {
		t.Errorf("opponent's creature went to the wrong hand")
	}
	if !g.Battlefield.Contains(rock) {
		t.Errorf("Evacuation bounced a non-creature")
	}
}

// River's Rebuke sweeps one seat, chosen by the target, and the
// predicate is built from the resolved target rather than from the
// caster.
func TestRiversRebukeOnlySweepsTheTargetedPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := pushWipeCreature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	theirRock := pushWipeCreature(g, opp.ID, "Their Rock", "Artifact", 0, 0)
	theirLand := pushWipeCreature(g, opp.ID, "Their Island", "Basic Land — Island", 0, 0)
	mine := pushWipeCreature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	thirds := pushWipeCreature(g, third.ID, "Third's", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "River's Rebuke", "Sorcery", riversRebukeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{theirs, theirRock} {
		if g.Battlefield.Contains(id) {
			t.Errorf("targeted player kept nonland permanent %s", id)
		}
		// River's Rebuke RETURNS TO HAND. Absence from the
		// battlefield is equally satisfied by a destroy or an exile.
		if !opp.Hand.Contains(id) {
			t.Errorf("nonland permanent %s did not reach its owner's hand", id)
		}
	}
	if !g.Battlefield.Contains(theirLand) {
		t.Errorf("River's Rebuke bounced a land")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(thirds) {
		t.Errorf("River's Rebuke swept a player it did not target")
	}
}

// --- Toxic Deluge: the additional cost and the layer effect ---------

func TestToxicDelugePaysXLifeAndShrinksEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := pushWipeCreature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := pushWipeCreature(g, opp.ID, "Big", "Creature — Wurm", 6, 6)

	before := me.Life
	castWipe(t, g, "Toxic Deluge", "Sorcery", toxicDelugeOracle,
		game.CastSpellParams{XValue: 3})

	// CR 601.2f: the additional cost is paid at CAST, with the spell
	// still on the stack — so the life is already gone before the
	// creatures shrink.
	if want := before - 3; me.Life != want {
		t.Fatalf("life at cast %d -> %d, want %d (X life is a cast cost)", before, me.Life, want)
	}

	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(small) {
		t.Errorf("a 2/2 survived -3/-3")
	}
	if !g.Battlefield.Contains(big) {
		t.Errorf("a 6/6 died to -3/-3")
	}
	// Survivors stay shrunk for the turn — the half that separates a
	// -X/-X from a destroy.
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == big && c.CurrentToughness() != 3 {
			t.Errorf("surviving 6/6 toughness = %d, want 3", c.CurrentToughness())
		}
	}
}

// CR 119.4 — you may pay N life only with a life total of at least N,
// so an unpayable X is a rejected cast rather than a player at -3.
func TestToxicDelugeRejectsMoreLifeThanYouHave(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Toxic Deluge", TypeLine: "Sorcery",
		OracleID: toxicDelugeOracle, Owner: active.ID, Controller: active.ID,
	})
	err := g.CastSpell(active.ID, id, game.CastSpellParams{XValue: active.Life + 1})
	if err == nil {
		t.Fatalf("casting Toxic Deluge for more life than you have was allowed")
	}
	if !active.Hand.Contains(id) {
		t.Errorf("a rejected cast still moved the card off the hand")
	}
}

// --- Supreme Verdict -----------------------------------------------

// The counterspell RESOLVES — it is a legal target — and then does
// nothing (CR 701.6a). The distinction from "illegal target" is
// observable: an illegal target would fizzle the counterspell
// instead.
func TestSupremeVerdictCannotBeCountered(t *testing.T) {
	g := newCatalogGame(t)
	caster, opponent := g.Seats[0], g.Seats[1]
	bear := pushWipeCreature(g, caster.ID, "Bear", "Creature — Bear", 2, 2)

	verdictID := castCatalogSpell(t, g, "Supreme Verdict", "Sorcery", supremeVerdictOracle, nil)

	counterID := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: counterID,
		Name:       "Counterspell",
		TypeLine:   "Instant",
		OracleID:   "cc187110-1148-4090-bbb8-e205694a39f5",
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, counterID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: verdictID}},
	}); err != nil {
		t.Fatalf("Counterspell on Supreme Verdict was refused at announce: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !opponent.Graveyard.Contains(counterID) {
		t.Errorf("Counterspell did not resolve — it should have, and done nothing")
	}
	if g.Battlefield.Contains(bear) {
		t.Errorf("Supreme Verdict was countered")
	}
	if !caster.Graveyard.Contains(verdictID) {
		t.Errorf("Supreme Verdict did not resolve to its owner's graveyard")
	}
}

// --- counted sweeps on permanents ----------------------------------

func TestBaneOfProgressGrowsByWhatItDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushWipeCreature(g, me.ID, "My Rock", "Artifact", 0, 0)
	pushWipeCreature(g, opp.ID, "Their Rock", "Artifact", 0, 0)
	pushWipeCreature(g, opp.ID, "Their Aura", "Enchantment", 0, 0)
	bear := pushWipeCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	// Cast it rather than seeding it: the ETB trigger only fires on a
	// real battlefield entry.
	bane := castCatalogSpell(t, g, "Bane of Progress", "Creature — Elemental", baneOfProgressOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bane) {
		t.Fatalf("Bane of Progress destroyed itself")
	}
	if !g.Battlefield.Contains(bear) {
		t.Errorf("Bane of Progress destroyed a creature")
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != bane {
			continue
		}
		if got := c.Counters[game.CounterPlusOne]; got != 3 {
			t.Errorf("+1/+1 counters = %d, want 3 (two artifacts + one enchantment)", got)
		}
	}
}

// Massacre Wurm is a one-sided Languish plus a per-death tax, and
// both halves read "opponent" relative to the Wurm's controller.
func TestMassacreWurmShrinksOnlyOpponentsAndTaxesEachDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushWipeCreature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirsA := pushWipeCreature(g, opp.ID, "Theirs A", "Creature — Bear", 2, 2)
	theirsB := pushWipeCreature(g, opp.ID, "Theirs B", "Creature — Bear", 2, 2)
	theirsBig := pushWipeCreature(g, opp.ID, "Theirs Big", "Creature — Wurm", 5, 5)

	oppBefore := opp.Life
	castCatalogSpell(t, g, "Massacre Wurm", "Creature — Phyrexian Wurm", massacreWurmOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(mine) {
		t.Errorf("Massacre Wurm shrank a creature its controller owns")
	}
	for _, id := range []uuid.UUID{theirsA, theirsB} {
		if g.Battlefield.Contains(id) {
			t.Errorf("opponent's 2/2 %s survived -2/-2", id)
		}
	}
	if !g.Battlefield.Contains(theirsBig) {
		t.Errorf("opponent's 5/5 died to -2/-2")
	}
	// Two deaths, two life each.
	if want := oppBefore - 4; opp.Life != want {
		t.Errorf("opponent life %d -> %d, want %d", oppBefore, opp.Life, want)
	}
}

// Chandra's Ignition's damage is dealt BY the chosen creature, and
// it spares that creature and nobody else.
//
// The card itself landed on main in the roadmap batch while this
// sprint was in flight (S23's own copy was dropped at rebase in favour
// of it). The test stays because it is the only coverage of the
// clause, and because "each other creature" plus the SBA deaths that
// follow is a mass effect whether or not it uses the primitives.
func TestChandrasIgnitionBurnsEveryOtherCreatureAndEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := pushWipeCreature(g, me.ID, "Big", "Creature — Wurm", 5, 5)
	myOther := pushWipeCreature(g, me.ID, "My Other", "Creature — Bear", 2, 2)
	theirs := pushWipeCreature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	oppBefore := opp.Life
	castCatalogSpell(t, g, "Chandra's Ignition", "Sorcery", chandrasIgnitionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(big) {
		t.Errorf("the source creature damaged itself")
	}
	for _, id := range []uuid.UUID{myOther, theirs} {
		if g.Battlefield.Contains(id) {
			t.Errorf("creature %s survived 5 damage", id)
		}
	}
	if want := oppBefore - 5; opp.Life != want {
		t.Errorf("opponent life %d -> %d, want %d", oppBefore, opp.Life, want)
	}
}

// The Disk's fairness is entirely in "enters tapped" — it is a CR 614
// self-replacement, so it is never untapped on the battlefield and
// cannot be cracked the turn it lands.
func TestNevinyrralsDiskEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	disk := castCatalogSpell(t, g, "Nevinyrral's Disk", "Artifact", nevinyrralsDiskOracle, nil)
	passPriorityAroundTable(t, g)

	c, ok := g.LookupCardForEffect(disk)
	if !ok || !g.Battlefield.Contains(disk) {
		t.Fatalf("Nevinyrral's Disk is not on the battlefield")
	}
	if !c.Tapped {
		t.Errorf("Nevinyrral's Disk entered untapped")
	}
}

// --- registration canary -------------------------------------------

// Every card this sprint claims is in the catalog. A card that fails
// to Register (a typo'd oracle ID, an init() that never ran) is
// silently a non-catalog card, which is the failure mode this sprint
// is least able to notice from behaviour alone.
func TestS23BoardwipesAreRegistered(t *testing.T) {
	for name, oracle := range map[string]string{
		"Wrath of God":     wrathOfGodOracle,
		"Supreme Verdict":  supremeVerdictOracle,
		"Fumigate":         fumigateOracle,
		"Deadly Tempest":   deadlyTempestOracle,
		"In Garruk's Wake": inGarruksWakeOracle,
		"Ritual of Soot":   ritualOfSootOracle,
		"Planar Cleansing": planarCleansingOracle,
		"Shatterstorm":     shatterstormOracle,
		"Cleansing Nova":   cleansingNovaOracle,
		"Crux of Fate":     cruxOfFateOracle,
		"Depopulate":       depopulateOracle,
		"Evacuation":       evacuationOracle,
		"River's Rebuke":   riversRebukeOracle,
		"Whelming Wave":    whelmingWaveOracle,
		"Toxic Deluge":     toxicDelugeOracle,
		"Languish":         languishOracle,
		"Massacre Wurm":    massacreWurmOracle,
		// Not an S23 registration — see the note on its test above.
		"Chandra's Ignition": chandrasIgnitionOracle,
		"Bane of Progress":   baneOfProgressOracle,
		"Nevinyrral's Disk":  nevinyrralsDiskOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not in the catalog", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered under the name %q", oracle, spec.Name)
		}
	}
}
