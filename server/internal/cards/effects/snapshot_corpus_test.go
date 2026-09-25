package effects

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// snapshot_corpus_test.go is #522's compatibility guard: the test that
// decodes a snapshot this binary did not write.
//
// The round-trip tests in internal/game have the same binary on both
// sides, so a renamed mirror field, a changed unit or a repurposed key
// passes them all while yesterday's restore point decodes into today's
// binary as a zero value. This file restores a COMMITTED corpus of
// restore points on every CI run:
//
//	internal/game/testdata/snapshots/v<N>/*.json   generated, one frozen set per schema version
//	internal/game/testdata/snapshots/real/*.json   scrubbed real files from cmd-dev (cmd/snapshotscrub)
//
// It lives in the effects package, not in internal/game, because a
// fixture holds tokens, a Clone, an Equipment and a planeswalker, and
// restoring those means rebuilding their abilities from the real
// catalog — which internal/game cannot import.
//
// For every fixture it checks, in order:
//
//  1. no card comes back with fewer catalog abilities than the file
//     recorded (the #522 parity check, GameSnapshot.AbilityShortfalls);
//  2. RestoreStrict accepts it;
//  3. every key and value in the fixture is still in a fresh capture of
//     the restored game (corpusSubset). A key this binary no longer
//     reads — renamed, removed, retyped — is gone from the recapture,
//     and that is the failure. New keys in the recapture are fine;
//  4. the restored game round-trips exactly through this binary;
//  5. it is a working game: layers recompute and it accepts an action.
//
// Fixtures are APPEND-ONLY. Each generated file is written once, by
// TestWriteSnapshotCorpus, and never regenerated (a later board under
// the same version may ADD a file, never change one): rewriting a
// fixture to make this test pass converts the guard back into the
// comment it replaced (ADR 0044 decision 7). A schema bump gets a NEW
// directory; the old ones stay and keep being restored.

var writeCorpus = flag.Bool("write-corpus", false,
	"write the generated snapshot fixtures for the current SnapshotSchemaVersion (TestWriteSnapshotCorpus)")

// corpusRoot is the fixture corpus, in the game package's testdata
// because the files are game.GameSnapshot's on-disk format.
var corpusRoot = filepath.Join("..", "..", "game", "testdata", "snapshots")

// corpusFile is the restore-point envelope ws writes to
// <dataDir>/restore/<id>.json, so a real file drops straight in.
type corpusFile struct {
	Seq      uint64             `json:"seq"`
	Snapshot *game.GameSnapshot `json:"snapshot"`
}

// corpusEpoch is every wall-clock time in a generated fixture.
var corpusEpoch = time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC)

// corpusMigrations lists, per schema version K, the JSON paths a
// fixture written BELOW K may legitimately disagree on after a restore
// — the paths K's migration rewrites. A bump that migrates a field adds
// its paths here, with the reason, in the same change; nothing else
// may. Paths are normalised: indices become [] and UUID map keys {}.
var corpusMigrations = map[int]map[string]string{}

// corpusIgnored are paths no fixture is held to, and why.
var corpusIgnored = map[string]string{
	".takenAt":      "capture metadata, not game state",
	".schema":       "the one field a fixture is expected to disagree on once the version moves",
	".layerVersion": "restore advances it by one so the first read recomputes the layers",
}

// corpusIgnoredLeaves are keys ignored wherever they appear. All three
// measure the running binary's CATALOG, not the game: a build that adds
// an ability to a card changes them legitimately. The parity check is
// what holds them to anything, and it only objects to FEWER.
var corpusIgnoredLeaves = map[string]string{
	"catalogAbilities":      "a measurement of the catalog; checked by AbilityShortfalls",
	"manaAbilityCount":      "a measurement of the catalog; checked by AbilityShortfalls",
	"activatedAbilityCount": "a measurement of the catalog; checked by AbilityShortfalls",
}

// ---------------------------------------------------------------
// The scripted boards
// ---------------------------------------------------------------

type corpusBoard struct {
	name  string
	build func(t *testing.T) *game.Game
}

const (
	corpusElspethOracle = "05e6b243-48a6-4a42-bc5f-413441de9c33"
	corpusCollarOracle  = "f5f4dd28-f4ae-4d39-b9b8-6ebfd63c93fe"
)

// corpusBoards is the generated half of the corpus. A board added here
// is written as a NEW file in the current version's set, or in the next
// version's; a file already written is never touched.
func corpusBoards() []corpusBoard {
	return []corpusBoard{
		{"fresh", func(t *testing.T) *game.Game { return newCorpusGame(t) }},
		{"tokens", corpusTokens},
		{"counters_planeswalker", corpusCountersAndPlaneswalker},
		{"commanders", corpusCommanders},
		{"face_down", corpusFaceDown},
		{"attached", corpusAttached},
		{"combat", corpusCombat},
		{"clone", corpusClone},
		// v7 (#1497): ADR 0041 phase 3's data-backed scoped effects.
		{"control", corpusControl},
		{"amass", corpusAmass},
		{"scoped_effect_kinds", corpusScopedEffectKinds},
		{"mass_diminish", corpusMassDiminish},
		// v7, added by tier 2 (#1497) as new files: delayed triggers as data.
		{"earthbend", corpusEarthbend},
		{"suspend_haste", corpusSuspendHaste},
		// v7, added by the #1568 review as new files: params, condParams
		// and a fired keyed item on disk.
		{"queued_mana_drain", corpusQueuedManaDrain},
		{"queued_doublecast", corpusQueuedDoublecast},
		{"fired_arcane_denial", corpusFiredArcaneDenial},
		// v7, added by tier 3a (#1497) as new files: until-end-of-turn
		// effects, which only became restore points with it.
		{"until_eot_pump", corpusUntilEOTPump},
		{"crewed_vehicle", corpusCrewedVehicle},
		// v7, added by ADR 0093 PR 4 (#1584) as a new file: duration
		// grants — the grantAbilities mod on disk.
		{"duration_grants", corpusDurationGrants},
		// v7, added by tier 4-0 (#1497) as a new file: a CR 603.12
		// reflexive trigger on the stack, its target already chosen —
		// a keyed stack item whose clause is re-derived from its Body
		// rather than carried as a captured closure.
		{"reflexive_trigger", corpusReflexiveTrigger},
		// v7, added by tier 4's first slice (#1497, ADR 0041 P9) as new
		// files: a stamped activated ability waiting on the stack (an
		// own row and a granted row), and a countered spell's last-known
		// information, which the snapshot carries from this slice on.
		{"activated_on_stack", corpusActivatedOnStack},
		{"granted_activated_on_stack", corpusGrantedActivatedOnStack},
		{"countered_spell_lki", corpusCounteredSpellLKI},
		// v7, added by tier 3b-1 (#1497) as new files: replacement
		// effects a spell creates, which only became restore points with
		// it — the replacement mods, ScopeGame, ScopeYourPermanents,
		// seq, amount and then on disk.
		{"fog", corpusFog},
		{"mending_hands_partial", corpusMendingHandsPartial},
		{"whip_redirect", corpusWhipRedirect},
		{"cosmic_intervention", corpusCosmicIntervention},
	}
}

// ---------------------------------------------------------------
// v7 boards added by ADR 0041 phase 3's tier 3b-1 (#1497)
// ---------------------------------------------------------------
//
// NEW files in v7/: a replacement effect a spell created held the
// restore point back until cleanup (and the Whip's redirect for as long
// as it lasted) before tier 3b, so none of these could be a fixture.

// corpusFog is a real Fog: one preventCombatDamage record, ScopeGame.
func corpusFog(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	castCatalogSpell(t, g, "Fog", "Instant", fogOracle, nil)
	passPriorityAroundTable(t, g)
	if n := scopedReplacementCount(g); n != 1 {
		t.Fatalf("setup: Fog registered %d scoped replacements, want 1", n)
	}
	return g
}

// corpusMendingHandsPartial is a real Mending Hands on a creature that
// has since been dealt 3: a preventDamage shield with 1 charge left.
func corpusMendingHandsPartial(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	bear := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Grizzly Bears", 2, 6))
	castCatalogSpell(t, g, "Mending Hands", "Instant", mendingHandsOracl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(bear, bear, 3) })
	if len(g.ScopedEffects) != 1 || g.ScopedEffects[0].Mods[0].Amount != 1 {
		t.Fatalf("setup: want one shield with 1 charge left, have %+v", g.ScopedEffects)
	}
	return g
}

// corpusWhipRedirect is a real Whip of Erebos activation: the returned
// creature, its end-step exile queued, and the exileInsteadOfLeaving
// record pinned to it with an indefinite duration.
func corpusWhipRedirect(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	advanceToMainOf(t, g, seat)
	whip := pushCatalogPermanent(g, me.ID, "Whip of Erebos", "Legendary Enchantment Artifact", b06WhipOfErebosOracle, false)
	dead := seedGraveyardCreature(me, "Giant", "{4}{B}")
	b06AddMana(me, "B", "B", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) || scopedReplacementCount(g) != 1 {
		t.Fatalf("setup: the creature is back %v, scoped replacements %d", g.Battlefield.Contains(dead), scopedReplacementCount(g))
	}
	return g
}

// corpusCosmicIntervention is a real Cosmic Intervention after it
// saved a creature: the exileInsteadOfGraveyard record
// (ScopeYourPermanents, with its then body) and the delayed return it
// scheduled.
func corpusCosmicIntervention(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	bear := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Grizzly Bears", 2, 2))
	castCatalogSpell(t, g, "Cosmic Intervention", "Instant", cosmicInterventionOracle, nil)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !g.Exile.Contains(bear) || len(g.DelayedTriggers) != 1 {
		t.Fatalf("setup: exiled %v, delayed triggers %d", g.Exile.Contains(bear), len(g.DelayedTriggers))
	}
	return g
}

// newCorpusGame is a started, mulligans-closed two-seat game: a
// commander and eleven Forests each, seeded so the deal is fixed.
func newCorpusGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i, name := range []string{"P1", "P2"} {
		deck := []game.Card{{
			InstanceID:  uuid.New(),
			Name:        fmt.Sprintf("Test Commander %d", i+1),
			TypeLine:    "Legendary Creature — Test",
			Power:       3,
			Toughness:   3,
			IsCommander: true,
		}}
		for range 11 {
			deck = append(deck, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest"})
		}
		if _, err := g.AddPlayer(name, deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithSource(rand.NewPCG(522, 6)); err != nil {
		t.Fatalf("StartWithSource: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

func corpusCreature(owner uuid.UUID, name string, p, tough int) game.Card {
	return game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Creature — Bear",
		Power:      p,
		Toughness:  tough,
		Owner:      owner,
		Controller: owner,
	}
}

func corpusTokens(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[0].ID
	for _, tok := range []game.Card{TreasureToken(), FoodToken(), ClueToken(), BloodToken()} {
		pushToken(g, me, tok)
	}
	return g
}

func corpusCountersAndPlaneswalker(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[0].ID
	bear := corpusCreature(me, "Grizzly Bears", 2, 2)
	bear.Counters = map[string]int{"+1/+1": 2}
	pushBattlefieldCardWithTimestamp(g, bear)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID:      uuid.New(),
		Name:            "Elspeth, Sun's Champion",
		TypeLine:        "Legendary Planeswalker — Elspeth",
		OracleID:        corpusElspethOracle,
		StartingLoyalty: 4,
		Counters:        map[string]int{"loyalty": 5},
		Owner:           me,
		Controller:      me,
	})
	return g
}

func corpusCommanders(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	// Seat 0's commander stays in the command zone; seat 1's has been
	// exiled.
	p := g.Seats[1]
	if p.Command.Size() == 0 {
		t.Fatal("setup: seat 1 has no commander in the command zone")
	}
	id := p.Command.Cards[0].InstanceID
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(p.Command, g.Exile, id); err != nil {
			t.Fatalf("exile the commander: %v", err)
		}
	})
	return g
}

func corpusFaceDown(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[0].ID
	manifested := corpusCreature(me, "Grizzly Bears", 2, 2)
	manifested.SetFaceDown(game.FaceDownManifested)
	manifested.KnownBy = map[uuid.UUID]bool{me: true}
	pushBattlefieldCardWithTimestamp(g, manifested)

	foretold := game.Card{
		InstanceID: uuid.New(),
		Name:       "Behold the Multitude",
		TypeLine:   "Sorcery",
		Owner:      me,
		Controller: me,
		KnownBy:    map[uuid.UUID]bool{me: true},
	}
	foretold.SetFaceDown(game.FaceDownForetold)
	g.Exile.PushTop(foretold)
	return g
}

func corpusAttached(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[0].ID
	host := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Grizzly Bears", 2, 2))
	collar := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Basilisk Collar",
		TypeLine:   "Artifact — Equipment",
		OracleID:   corpusCollarOracle,
		Owner:      me,
		Controller: me,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(collar, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	return g
}

func corpusCombat(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	attacker := g.Seats[g.Turn.ActiveSeat]
	defender := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := pushBattlefieldCardWithTimestamp(g, corpusCreature(attacker.ID, "Grizzly Bears", 2, 2))
	b := pushBattlefieldCardWithTimestamp(g, corpusCreature(attacker.ID, "Runeclaw Bear", 2, 2))
	pushBattlefieldCardWithTimestamp(g, corpusCreature(defender.ID, "Balduvian Bears", 2, 2))
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to declare attackers: %v", err)
		}
	}
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, defender.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	return g
}

func corpusClone(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bears := pushBattlefieldCardWithTimestamp(g, corpusCreature(active.ID, "Grizzly Bears", 2, 2))
	castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, bears)
	return g
}

// ---------------------------------------------------------------
// v7 boards: ADR 0041 phase 3's data-backed scoped effects (#1497)
// ---------------------------------------------------------------

// corpusControl is the two control shapes: a Sower-of-Temptation-style
// theft that lasts while its source remains, and a Switcheroo exchange
// (two records, one timestamp, indefinite and pinned).
func corpusControl(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	sower := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Sower of Temptation", 2, 2))
	stolen := pushBattlefieldCardWithTimestamp(g, corpusCreature(opp, "Grizzly Bears", 2, 2))
	mine := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Runeclaw Bear", 2, 2))
	theirs := pushBattlefieldCardWithTimestamp(g, corpusCreature(opp, "Balduvian Bears", 2, 2))
	g.WithWriteLock(func() {
		d, ok := g.ForAsLongAsOnBattlefieldDuration(sower)
		if !ok {
			t.Fatal("setup: the Sower is not on the battlefield")
		}
		if !g.GainControlForEffect(sower, stolen, me, d, "Sower of Temptation — gain control") {
			t.Fatal("GainControlForEffect registered nothing")
		}
		if !g.ExchangeControlForEffect(uuid.Nil, mine, theirs, "Switcheroo — exchange control") {
			t.Fatal("ExchangeControlForEffect registered nothing")
		}
	})
	return g
}

// corpusAmass is amass's "it's also an Orc" on a Zombie Army: the
// indefinite, pinned addSubtypes record.
func corpusAmass(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[0].ID
	g.WithWriteLock(func() {
		for _, subtype := range []string{"Zombie", "Orc"} {
			if err := g.AmassForEffect(me, uuid.Nil, ArmyToken("Zombie"), subtype, 1, nil); err != nil {
				t.Fatalf("amass %s: %v", subtype, err)
			}
		}
	})
	if len(g.ScopedEffects) != 1 {
		t.Fatalf("setup: the Orc amass registered %d scoped effects, want 1", len(g.ScopedEffects))
	}
	return g
}

// corpusScopedEffectKinds freezes every mod kind and every duration
// kind the v7 vocabulary has, registered through the one write path so
// the file is exactly what a live game would write. One record per
// kind, all on one creature; the suspend-shaped member (EnteredAt 0,
// a controller) is on a second.
func corpusScopedEffectKinds(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	target := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Grizzly Bears", 2, 2))
	hasty := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Rift Bolt Bear", 3, 3))
	g.WithWriteLock(func() {
		pinned := g.PinnedObjectsLocked(target)
		type entry struct {
			mods []game.Mod
			d    game.Duration
		}
		whileOnBattlefield, ok := g.ForAsLongAsOnBattlefieldDuration(target)
		if !ok {
			t.Fatal("setup: the target is not on the battlefield")
		}
		entries := []entry{
			{[]game.Mod{game.SetControllerMod(opp)}, g.UntilEndOfTurnDuration()},
			{[]game.Mod{game.AddTypesMod("Artifact")}, g.UntilYourNextTurnDuration(me)},
			{[]game.Mod{game.RemoveTypesMod("Creature")}, whileOnBattlefield},
			{[]game.Mod{game.AddSubtypesMod("Island")}, g.PinnedTo(game.IndefiniteDuration(), target)},
			{[]game.Mod{game.AllCreatureTypesMod()}, g.UntilEndOfYourNextTurnDuration(me)},
			{[]game.Mod{game.SetColorsMod("U")}, g.UntilEndOfTurnDuration()},
			{[]game.Mod{game.AddKeywordsMod("flying", "toxic 1")}, g.UntilEndOfTurnDuration()},
			{[]game.Mod{game.RemoveKeywordsMod("hexproof")}, g.UntilEndOfTurnDuration()},
			{[]game.Mod{game.LoseAllAbilitiesMod("defender")}, g.UntilEndOfTurnDuration()},
			{[]game.Mod{game.AddRestrictionsMod(game.CantAttackOrBlock)}, g.UntilEndOfTurnDuration()},
			{game.SetBasePTMods(0, 2), g.UntilEndOfTurnDuration()},
			{[]game.Mod{game.ModifyPTMod(3, -1)}, g.UntilEndOfTurnDuration()},
		}
		for i, e := range entries {
			if !g.RegisterScopedEffectForEffect(uuid.Nil, pinned, e.mods, e.d,
				fmt.Sprintf("corpus scoped effect %d", i)) {
				t.Fatalf("entry %d registered nothing", i)
			}
		}
		// Suspend's shape: any entry of the object, while its caster
		// controls it.
		if !g.RegisterScopedEffectForEffect(hasty,
			[]game.AffectedObject{{ID: hasty, Controller: me}},
			[]game.Mod{game.AddKeywordsMod("haste")},
			game.UntilYouLoseControlOfDuration(hasty, me), "Suspend — haste (CR 702.62a)") {
			t.Fatal("the suspend-shaped entry registered nothing")
		}
	})
	return g
}

// corpusMassDiminish casts the real card, so the file holds what the
// catalog writes rather than what a test registered by hand.
func corpusMassDiminish(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushBattlefieldCardWithTimestamp(g, corpusCreature(other.ID, "Craw Wurm", 6, 4))
	pushBattlefieldCardWithTimestamp(g, corpusCreature(other.ID, "Grizzly Bears", 2, 2))
	castCatalogSpell(t, g, "Mass Diminish", "Sorcery", massDiminishOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: other.ID}})
	passPriorityAroundTable(t, g)
	if len(g.ScopedEffects) != 1 {
		t.Fatalf("setup: Mass Diminish registered %d scoped effects, want 1", len(g.ScopedEffects))
	}
	return g
}

// ---------------------------------------------------------------
// v7 boards added by ADR 0041 phase 3's tier 2 (#1497)
// ---------------------------------------------------------------
//
// NEW files in v7/, under the narrowed rule: the writer never touches
// an existing file. Both shapes only became restore points with tier 2.

// corpusEarthbend is a real earthbend: the animation record AND the
// "return it tapped when it dies or is exiled" delayed trigger, whose
// body and event condition are registered keys.
func corpusEarthbend(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[0].ID
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me, Controller: me,
	})
	g.WithWriteLock(func() {
		if err := g.EarthbendForEffect(me, uuid.Nil, land, 2); err != nil {
			t.Fatalf("EarthbendForEffect: %v", err)
		}
	})
	if len(g.DelayedTriggers) != 1 || len(g.ScopedEffects) != 1 {
		t.Fatalf("setup: earthbend left %d delayed triggers and %d scoped effects, want 1 and 1",
			len(g.DelayedTriggers), len(g.ScopedEffects))
	}
	return g
}

// corpusSuspendHaste is a real suspend cast (#1558 item 6): Judoon
// Enforcers suspended, its countdown run to the last counter, the free
// cast taken, and the creature on the battlefield with CR 702.62a's
// haste — the record whose member follows the card from the stack
// (EnteredAt 0) while its caster controls it.
func corpusSuspendHaste(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Judoon Enforcers", TypeLine: "Creature — Rhino Soldier",
		OracleID: corpusJudoonOracle, ManaCost: "{3}{R}{W}", Power: 5, Toughness: 5,
		Owner: active.ID, Controller: active.ID,
	})
	for _, c := range []string{"C", "R", "W"} {
		active.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	if err := g.PerformSpecialAction(active.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend Judoon Enforcers: %v", err)
	}
	// Six counters is six turns of draws from an eleven-card library;
	// take all but the last off by effect, so the next upkeep's tick is
	// the one that offers the cast.
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterTime, -5); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	offer := latestChoiceOfKind(g, game.PendingChoiceMayCast)
	if offer == nil {
		t.Fatal("setup: the last time counter offered no cast")
	}
	if err := g.ResolveMayCast(offer.ID, active.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.ScopedEffects) != 1 {
		t.Fatalf("setup: the suspended creature carries %d scoped effects, want its haste", len(g.ScopedEffects))
	}
	return g
}

// corpusQueuedManaDrain is a queued Mana Drain refund: a delayed trigger
// whose body reads PARAMS (the mana value), waiting for its controller's
// next main phase (#1568 review).
func corpusQueuedManaDrain(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller: me, Label: "Mana Drain — add {C} × 3",
			At: game.StepPrecombatMain, ControllerTurnOnly: true,
			Body: manaDrainRefundBody, Params: game.EffectParams{Amount: 3},
		})
	})
	return g
}

// corpusQueuedDoublecast is a queued Doublecast: an event-conditioned
// delayed trigger whose condition reads CONDPARAMS (the spell filter).
func corpusQueuedDoublecast(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller: me, Label: "Doublecast — copy that spell",
			On:         []game.EventKind{game.EventCast},
			Condition:  youNextCastCondition,
			CondParams: game.EffectParams{Filter: game.CastFilter{Types: []string{"Instant", "Sorcery"}}},
			Body:       copyTheSpellBody,
		})
	})
	return g
}

// corpusFiredArcaneDenial is a FIRED keyed trigger waiting on the stack:
// Arcane Denial's upkeep draws, with its victim as params, put on the
// stack by the upkeep that fires it — so the file holds a stack item
// with `body` and `params`.
func corpusFiredArcaneDenial(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller: me, Label: "Arcane Denial — its controller draws two cards, you draw a card",
			At: game.StepEnd, Body: arcaneDenialDrawsBody, Params: game.EffectParams{Player: victim},
		})
	})
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if len(g.StackMeta)+len(g.PendingTriggers) == 0 {
		t.Fatal("setup: the delayed trigger did not fire at the end step")
	}
	return g
}

// corpusDurationGrants is two real duration grants (ADR 0093 PR 4,
// #1584): Feign Death's lone grantAbilities mod, and Fake Your Own
// Death's "+2/+0 and gains …" — a modifyPT and a grantAbilities mod in
// one record. The file holds `grants` keys a restore must resolve.
func corpusDurationGrants(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	bear := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Grizzly Bears", 2, 2))
	wurm := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Craw Wurm", 6, 4))
	castCatalogSpell(t, g, "Feign Death", "Instant", feignDeathOracle, dgCardRef(bear))
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Fake Your Own Death", "Instant", fakeYourOwnDeathOracle, dgCardRef(wurm))
	passPriorityAroundTable(t, g)
	if len(g.ScopedEffects) != 2 {
		t.Fatalf("setup: the two spells registered %d scoped effects, want 2", len(g.ScopedEffects))
	}
	return g
}

// corpusReflexiveTrigger is a real CR 603.12 reflexive trigger sitting
// on the stack with its target already chosen (ADR 0041 P9, #1497,
// tier 4): Undead Butler dies, its controller exiles it, and "when you
// do" — the reflexive half — has picked the creature card to return
// and is waiting to resolve. The file holds a keyed stack item whose
// clause is re-derived from its Body ("undead-butler/return-to-hand")
// rather than carried as a captured *TargetSpec.
func corpusReflexiveTrigger(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushGraveyardPermanent(me, "Dead Fatty", "Creature — Bear", "{5}{B}")
	butler := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Undead Butler", TypeLine: "Creature — Zombie",
		OracleID: b41UndeadButlerOracle, Power: 1, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(butler) })
	passPriorityAroundTable(t, g)
	// "You may exile it" is the parent's optional prompt; the
	// reflexive "when you do" that follows is mandatory.
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, target)
	if len(g.StackMeta) == 0 {
		t.Fatal("setup: the reflexive trigger is not on the stack")
	}
	return g
}

const corpusJudoonOracle = "ca04089c-24b6-465e-9303-ea28c0d6f3c7"

// ---------------------------------------------------------------
// v7 boards added by ADR 0041 phase 3's tier 3a (#1497)
// ---------------------------------------------------------------
//
// NEW files in v7/: an until-end-of-turn effect held the restore point
// back until cleanup before tier 3a, so neither shape could be a
// fixture until now.

// corpusUntilEOTPump is a real Giant Growth cast on a prowess creature:
// the spell's +3/+3 and the prowess trigger's +1/+1, two modifyPT
// records ending at this turn's cleanup — what the catalog and the
// engine write, not what a test registered by hand.
func corpusUntilEOTPump(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	monk := pushProwessCreature(g, me, "Monastery Swiftspear", 1, 2)
	castCatalogSpell(t, g, "Giant Growth", "Instant", giantGrowthOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: monk}})
	passPriorityAroundTable(t, g)
	if len(g.ScopedEffects) != 2 {
		t.Fatalf("setup: Giant Growth on a prowess creature registered %d scoped effects, want 2", len(g.ScopedEffects))
	}
	return g
}

// corpusCrewedVehicle is a real crew activation: Smuggler's Copter
// crewed, one record adding Artifact Creature until end of turn.
func corpusCrewedVehicle(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	copter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Smuggler's Copter", OracleID: smugglersCopterOracle,
		TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 3, Owner: me, Controller: me,
	})
	crewer := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Grizzly Bears", 2, 2))
	if err := g.ActivateCatalogAbility(me, copter, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.ScopedEffects) != 1 {
		t.Fatalf("setup: crew registered %d scoped effects, want 1", len(g.ScopedEffects))
	}
	return g
}

// ---------------------------------------------------------------
// v7 boards added by ADR 0041 phase 3's tier 4, first slice (#1497)
// ---------------------------------------------------------------
//
// NEW files in v7/: an ability waiting on the stack held the restore
// point back until it resolved before this slice, so none of these
// could be a fixture until now.

// corpusActivatedOnStack is a real Goblin Bombardment activation waiting
// on the stack: the sacrifice paid, the ping at the opponent unresolved.
// The file holds a stack item with body catalog/activated and an own:0
// ability ref.
func corpusActivatedOnStack(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	bomb := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Bombardment", OracleID: goblinBombardmentOracle,
		TypeLine: "Enchantment", Owner: me, Controller: me,
	})
	fodder := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Grizzly Bears", 2, 2))
	if err := g.ActivateCatalogAbility(me, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp}},
	}); err != nil {
		t.Fatalf("activate Goblin Bombardment: %v", err)
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("setup: %d stack items, want the Bombardment's ping", len(g.StackMeta))
	}
	return g
}

// corpusGrantedActivatedOnStack is a real Squirrel Nest's granted
// ability, activated on the enchanted land and waiting on the stack: a
// grant:<bundle>:<i>:<n> ability ref on disk.
func corpusGrantedActivatedOnStack(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me, Controller: me,
	})
	pushAuraOnLand(t, g, me, "Squirrel Nest", gaSquirrelNestOracle, land)
	idx, ref := grantedActivatedIndex(t, g, land)
	if idx < 0 {
		t.Fatal("setup: the enchanted land has no granted ability")
	}
	if err := g.ActivateCatalogAbility(me, land, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("setup: %d stack items, want the Squirrel", len(g.StackMeta))
	}
	return g
}

// corpusCounteredSpellLKI is a real Lightning Bolt countered by a real
// Counterspell: the stack is empty, and the Bolt as it last stood on it
// is in the game's last-known information — `lastKnownStack` on disk.
func corpusCounteredSpellLKI(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	castCatalogSpell(t, g, "Counterspell", "Instant", corpusCounterspellOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)
	snap := g.CaptureSnapshot()
	if len(snap.StackMeta) != 0 || len(snap.LastKnownStack) != 1 {
		t.Fatalf("setup: %d stack items and %d last-known spells, want 0 and the Bolt",
			len(snap.StackMeta), len(snap.LastKnownStack))
	}
	return g
}

const corpusCounterspellOracle = "cc187110-1148-4090-bbb8-e205694a39f5"

// ---------------------------------------------------------------
// Rendering a board deterministically
// ---------------------------------------------------------------

// detReader is a counter-mode SHA-256 stream: the same seed gives the
// same bytes, so uuid.New gives the same IDs.
type detReader struct {
	seed []byte
	ctr  uint64
	buf  []byte
}

func (r *detReader) Read(p []byte) (int, error) {
	for i := range p {
		if len(r.buf) == 0 {
			var c [8]byte
			binary.BigEndian.PutUint64(c[:], r.ctr)
			r.ctr++
			sum := sha256.Sum256(append(append([]byte(nil), r.seed...), c[:]...))
			r.buf = sum[:]
		}
		p[i] = r.buf[0]
		r.buf = r.buf[1:]
	}
	return len(p), nil
}

// renderBoard builds one board with every source of nondeterminism
// pinned — IDs, the engine clock, the RNG, the wall-clock fields — and
// returns the fixture bytes.
func renderBoard(t *testing.T, b corpusBoard) []byte {
	t.Helper()
	uuid.SetRand(&detReader{seed: []byte("cmdctrl-corpus/" + b.name)})
	defer uuid.SetRand(nil)
	tick := corpusEpoch.UnixNano()
	restoreClock := game.SetClockForTest(func() int64 { tick += 1000; return tick })
	defer restoreClock()

	g := b.build(t)
	g.CreatedAt = corpusEpoch
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("board %q is not a restore point, so it cannot be a fixture: %+v", b.name, snap.Continuations)
	}
	snap.TakenAt = corpusEpoch
	for i := range snap.Seats {
		for j := range snap.Seats[i].LifeHistory {
			snap.Seats[i].LifeHistory[j].At = corpusEpoch
		}
	}
	raw, err := json.MarshalIndent(corpusFile{Seq: 1, Snapshot: snap}, "", "  ")
	if err != nil {
		t.Fatalf("marshal %q: %v", b.name, err)
	}
	return append(raw, '\n')
}

func firstDifferingLine(a, b []byte) string {
	al, bl := strings.Split(string(a), "\n"), strings.Split(string(b), "\n")
	for i := 0; i < len(al) && i < len(bl); i++ {
		if al[i] != bl[i] {
			return fmt.Sprintf("line %d:\n  -%s\n  +%s", i+1, al[i], bl[i])
		}
	}
	return fmt.Sprintf("lengths differ: %d vs %d lines", len(al), len(bl))
}

// TestWriteSnapshotCorpus writes the generated fixture set for the
// current schema version. It runs only when asked:
//
//	go test ./internal/cards/effects -run TestWriteSnapshotCorpus -args -write-corpus
//
// The writer NEVER TOUCHES AN EXISTING FILE (ADR 0041's 2026-09-24
// phase 3 amendment, owner decision 6 — narrowed from "never touches an
// existing directory"). For a version whose directory already exists it
// re-renders every board and fails if any file already there would come
// out different, because a fixture is never rewritten: a changed shape
// is a new version, in a new directory. A board with no file yet — one a
// later change under the same version added — is written beside the
// others, which is how a shape that becomes a restore point after its
// version was introduced gets frozen without a bump.
func TestWriteSnapshotCorpus(t *testing.T) {
	if !*writeCorpus {
		t.Skip("writes fixtures; run with -args -write-corpus")
	}
	dir := filepath.Join(corpusRoot, fmt.Sprintf("v%d", game.SnapshotSchemaVersion))
	rendered := map[string][]byte{}
	for _, b := range corpusBoards() {
		first, second := renderBoard(t, b), renderBoard(t, b)
		if !bytes.Equal(first, second) {
			t.Fatalf("board %q renders differently twice in a row, so it cannot be a fixture — something in it is not pinned:\n%s",
				b.name, firstDifferingLine(first, second))
		}
		rendered[b.name] = first
	}

	var changed []string
	fresh := map[string][]byte{}
	for name, raw := range rendered {
		existing, err := os.ReadFile(filepath.Join(dir, name+".json"))
		if err != nil {
			fresh[name] = raw
			continue
		}
		if !bytes.Equal(existing, raw) {
			changed = append(changed, name+".json: "+firstDifferingLine(existing, raw))
		}
	}
	if len(changed) > 0 {
		sort.Strings(changed)
		t.Fatalf(`%s holds fixtures written by an earlier build of schema v%d, and
fixtures are never rewritten. The writer would now produce:

  %s

If the snapshot's shape changed, that is a new schema version: bump
SnapshotSchemaVersion (see its comment for when that is required) and
run this again to write v%d beside the old set. If only the scripted
boards changed, nothing needs doing — the file on disk is still the
fixture. Nothing was written.`,
			dir, game.SnapshotSchemaVersion, strings.Join(changed, "\n  "), game.SnapshotSchemaVersion+1)
	}
	if len(fresh) == 0 {
		t.Logf("%s already holds this build's output; nothing to write", dir)
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, raw := range fresh {
		path := filepath.Join(dir, name+".json")
		// O_EXCL: the one guarantee this function makes is that it
		// never overwrites a fixture, so it asks the filesystem to
		// refuse rather than trusting the read above.
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(raw); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("wrote %d new fixtures to %s", len(fresh), dir)
}

// TestSnapshotCorpusBoardsStillBuild keeps the writer honest between
// bumps: every scripted board still builds and is still a restore
// point, so the next version's set can be written when it is needed
// rather than debugged then.
func TestSnapshotCorpusBoardsStillBuild(t *testing.T) {
	for _, b := range corpusBoards() {
		t.Run(b.name, func(t *testing.T) { renderBoard(t, b) })
	}
}

// ---------------------------------------------------------------
// Restoring the corpus
// ---------------------------------------------------------------

var versionDirRe = regexp.MustCompile(`^v([0-9]+)$`)

// TestSnapshotCorpusRestores is the guard. See the file comment.
func TestSnapshotCorpusRestores(t *testing.T) {
	entries, err := os.ReadDir(corpusRoot)
	if err != nil {
		t.Fatalf("read %s: %v", corpusRoot, err)
	}
	haveCurrent := false
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		version := 0
		if m := versionDirRe.FindStringSubmatch(e.Name()); m != nil {
			version, _ = strconv.Atoi(m[1])
			if version == game.SnapshotSchemaVersion {
				haveCurrent = true
			}
			if version > game.SnapshotSchemaVersion {
				t.Errorf("%s holds fixtures for schema v%d, newer than this binary's v%d", e.Name(), version, game.SnapshotSchemaVersion)
				continue
			}
		} else if e.Name() != "real" {
			t.Errorf("%s: unexpected directory in the corpus (want v<N> or real)", e.Name())
			continue
		}
		dir := filepath.Join(corpusRoot, e.Name())
		files, err := filepath.Glob(filepath.Join(dir, "*.json"))
		if err != nil {
			t.Fatal(err)
		}
		if version > 0 && len(files) == 0 {
			t.Errorf("%s holds no fixtures", dir)
		}
		for _, f := range files {
			t.Run(e.Name()+"/"+filepath.Base(f), func(t *testing.T) { checkCorpusFixture(t, f, version) })
		}
	}
	if !haveCurrent {
		t.Errorf(`SnapshotSchemaVersion is %d and %s/v%d does not exist.

Every schema version gets one frozen fixture set, written when the
version is introduced. Write it:

  go test ./internal/cards/effects -run TestWriteSnapshotCorpus -args -write-corpus`,
			game.SnapshotSchemaVersion, corpusRoot, game.SnapshotSchemaVersion)
	}
}

func checkCorpusFixture(t *testing.T, path string, version int) {
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := decodeGeneric(raw)
	if err != nil {
		t.Fatalf("decode as JSON: %v", err)
	}
	var file corpusFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("decode as a restore point: %v", err)
	}
	if file.Snapshot == nil {
		t.Fatal(`no "snapshot" in the file`)
	}
	schema := file.Snapshot.Schema
	if version > 0 && schema != version {
		t.Fatalf("fixture in v%d/ says schema %d", version, schema)
	}

	// 1. The parity check.
	for _, sf := range file.Snapshot.AbilityShortfalls() {
		t.Errorf("%s (%s%s) in %s comes back with fewer catalog abilities than the fixture recorded: captured %+v, restored %+v, entry missing %v — a deploy would restore it flagged manual",
			sf.Name, sf.OracleID, sf.TokenKey, sf.Zone, sf.Captured, sf.Restored, sf.EntryMissing)
	}

	// 2. RestoreStrict.
	g, err := file.Snapshot.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}

	// 3. Nothing in the fixture was lost on the way in.
	recap := g.CaptureSnapshot()
	recapAny, err := decodeGeneric(mustMarshal(t, recap))
	if err != nil {
		t.Fatal(err)
	}
	want, _ := fixture.(map[string]any)["snapshot"]
	var diffs []string
	corpusSubset(want, recapAny, "", schema, &diffs)
	if len(diffs) > 0 {
		t.Errorf(`this binary does not read %s the way the binary that wrote it did.
Each line is a key or value in the fixture that is not in a fresh capture
of the restored game — a mirror field renamed, removed or retyped, or a
change of meaning, under schema v%d:

  %s

Do not edit the fixture. If the change is deliberate it is a schema bump
with a migration: raise SnapshotSchemaVersion, migrate the old shape in
restore, and list the migrated paths in corpusMigrations.`,
			path, schema, strings.Join(diffs, "\n  "))
	}

	// 4. The restored game round-trips exactly through this binary.
	again, err := decodeRestorePoint(t, mustMarshal(t, recap)).RestoreStrict()
	if err != nil {
		t.Fatalf("second RestoreStrict: %v", err)
	}
	recap2 := again.CaptureSnapshot()
	if recap2.LayerVersion != recap.LayerVersion+1 {
		t.Errorf("second restore layerVersion %d, want %d", recap2.LayerVersion, recap.LayerVersion+1)
	}
	recap2.LayerVersion, recap2.TakenAt = recap.LayerVersion, recap.TakenAt
	if a, b := mustMarshal(t, recap), mustMarshal(t, recap2); !bytes.Equal(a, b) {
		t.Errorf("the restored game does not round-trip through this binary:\n%s", firstDifferingLine(a, b))
	}

	// 5. A usable game: the layers recompute against the restored
	// board (every catalog static on it runs), and it takes an action.
	g.ReadSnapshot(func() {})
	if len(g.Seats) == 0 {
		t.Fatal("restored game has no seats")
	}
	seat := g.Seats[0]
	if seat.Library.Size() > 0 {
		if err := g.DrawCard(seat.ID); err != nil {
			t.Errorf("the restored game refused an action: %v", err)
		}
	}
}

func decodeRestorePoint(t *testing.T, raw []byte) *game.GameSnapshot {
	t.Helper()
	var s game.GameSnapshot
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	return &s
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func decodeGeneric(raw []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	err := dec.Decode(&v)
	return v, err
}

var (
	indexRe   = regexp.MustCompile(`\[[0-9]+\]`)
	uuidKeyRe = regexp.MustCompile(`\.[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
)

// normalisePath turns ".seats[1].commanderDamage.<uuid>" into
// ".seats[].commanderDamage.{}", the form the ignore lists use.
func normalisePath(p string) string {
	return uuidKeyRe.ReplaceAllString(indexRe.ReplaceAllString(p, "[]"), ".{}")
}

func corpusPathIgnored(path string, schema int) bool {
	norm := normalisePath(path)
	if _, ok := corpusIgnored[norm]; ok {
		return true
	}
	leaf := norm[strings.LastIndex(norm, ".")+1:]
	if _, ok := corpusIgnoredLeaves[leaf]; ok {
		return true
	}
	for k, paths := range corpusMigrations {
		if schema < k {
			if _, ok := paths[norm]; ok {
				return true
			}
		}
	}
	return false
}

const corpusDiffCap = 40

// corpusSubset appends a line for every key or value in want that got
// does not carry. Keys only in got are additions and are fine.
func corpusSubset(want, got any, path string, schema int, out *[]string) {
	if len(*out) >= corpusDiffCap || corpusPathIgnored(path, schema) {
		return
	}
	switch w := want.(type) {
	case map[string]any:
		g, ok := got.(map[string]any)
		if !ok {
			*out = append(*out, fmt.Sprintf("%s: an object in the fixture, %s in the recapture", path, describe(got)))
			return
		}
		keys := make([]string, 0, len(w))
		for k := range w {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child := path + "." + k
			gv, ok := g[k]
			if !ok {
				if !corpusPathIgnored(child, schema) {
					*out = append(*out, fmt.Sprintf("%s: in the fixture (%s), absent from the recapture", child, describe(w[k])))
				}
				continue
			}
			corpusSubset(w[k], gv, child, schema, out)
		}
	case []any:
		g, ok := got.([]any)
		if !ok {
			*out = append(*out, fmt.Sprintf("%s: a list in the fixture, %s in the recapture", path, describe(got)))
			return
		}
		if len(w) != len(g) {
			*out = append(*out, fmt.Sprintf("%s: %d entries in the fixture, %d in the recapture", path, len(w), len(g)))
			return
		}
		for i := range w {
			corpusSubset(w[i], g[i], fmt.Sprintf("%s[%d]", path, i), schema, out)
		}
	default:
		if !reflect.DeepEqual(want, got) {
			*out = append(*out, fmt.Sprintf("%s: %s in the fixture, %s in the recapture", path, describe(want), describe(got)))
		}
	}
}

func describe(v any) string {
	switch x := v.(type) {
	case map[string]any:
		return fmt.Sprintf("an object of %d keys", len(x))
	case []any:
		return fmt.Sprintf("a list of %d", len(x))
	case nil:
		return "null"
	default:
		s := fmt.Sprintf("%v", x)
		if len(s) > 60 {
			s = s[:60] + "…"
		}
		return fmt.Sprintf("%q", s)
	}
}

// TestCorpusSubsetSeesARename proves the comparison is not vacuous: a
// key the binary no longer reads shows up as absent.
func TestCorpusSubsetSeesARename(t *testing.T) {
	want, _ := decodeGeneric([]byte(`{"cards":[{"counters":{"+1/+1":2},"name":"Bear"}],"takenAt":"x"}`))
	got, _ := decodeGeneric([]byte(`{"cards":[{"counterMap":{"+1/+1":2},"name":"Bear","extra":1}],"takenAt":"y"}`))
	var diffs []string
	corpusSubset(want, got, "", game.SnapshotSchemaVersion, &diffs)
	if len(diffs) != 1 || !strings.Contains(diffs[0], ".cards[0].counters") {
		t.Fatalf("diffs = %v, want exactly the renamed key", diffs)
	}
}
