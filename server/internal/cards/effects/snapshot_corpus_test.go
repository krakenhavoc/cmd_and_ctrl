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
		// v7, added by tier 3b-2 (#1497) as new files: block-rule
		// effects a spell or ability creates, which only became restore
		// points with it — the cantBeBlockedExceptBy and
		// limitBlockersPerDefender mods, and Text, on disk.
		{"gingerbrute", corpusGingerbrute},
		{"mirri_limit", corpusMirriLimit},
		// v7, added by #1650 as a new file: a restriction whose
		// affected set is a live rule — addRestrictions under the
		// creaturesWithoutFlying scope on disk.
		{"falter", corpusFalter},
		// v7, added by #1651: a hexproof waiver under the live
		// opponentsAndTheirCreatures scope, and a cantHaveKeywords record
		// pinned to an opponent's creature. Written by #1707's v7
		// regeneration (ADR 0044's 2026-09-28 #1698 amendment).
		{"detection_tower", corpusDetectionTower},
		{"arcane_lighthouse", corpusArcaneLighthouse},
		// v7, added by tier 4's second slice (#1497, ADR 0041 P9) as new
		// files: declared triggered abilities waiting to resolve, named
		// by their catalog row — a card's own row, a granted bundle's,
		// a token's, an emblem's, a modal one, one whose clause is
		// built from the trigger context, and a storm trigger over a
		// countered spell's last-known information.
		{"etb_trigger_on_stack", corpusETBTriggerOnStack},
		{"granted_dies_trigger", corpusGrantedDiesTrigger},
		{"token_trigger", corpusTokenTrigger},
		{"emblem_trigger", corpusEmblemTrigger},
		{"modal_trigger", corpusModalTrigger},
		{"targets_from_trigger", corpusTargetsFromTrigger},
		{"storm_after_counter", corpusStormAfterCounter},
		// Tier 4-0's prowess/pump body, which merged while 4-2 was open:
		// an engine trigger with no catalog row, keyed by its body.
		{"prowess_on_stack", corpusProwessOnStack},
		// v7, added by #1805 (ADR 0106 §3) as a new file: an evolve
		// trigger — the second engine keyword trigger, keyed by its
		// evolve/grow body — waiting on the stack, carrying the entered
		// creature on its trigger context.
		{"evolve_on_stack", corpusEvolveOnStack},
		// v7, added by #1858 (ADR 0107 §1) as a new file: a CR 603.8
		// state trigger — a catalog row with a State condition and no
		// event — waiting on the stack. Its latch is derived from this
		// item, so the restored table must not trigger it again.
		{"state_trigger_on_stack", corpusStateTriggerOnStack},
		// v7, added by #1858 as a new file: a CR 603.8 state trigger
		// about one of several objects (Bomb Squad) waiting on the
		// stack. The creature it is about rides its trigger context,
		// which is also what the per-object latch reads.
		{"state_each_trigger_on_stack", corpusStateEachTriggerOnStack},
		// v7, added by #1593: duration copy effects — the becomeCopy mod
		// carrying its copied values, and the carried durationCopyBase
		// under a Cytoshaped Clone. Written by #1712, alongside the
		// regeneration of gingerbrute.json and whip_redirect.json that
		// unblocked it (ADR 0044's 2026-09-28 #1712 amendment).
		{"duration_copy", corpusDurationCopy},
		// v7, added by ADR 0101 (#1753) as a new file: keyword counters
		// and their CR 613.7c timestamps (counterStampedAt), one older and
		// one newer than a "loses all abilities" record on the same
		// creature, so a restore keeps the order.
		{"keyword_counters", corpusKeywordCounters},
		// v7, added by ADR 0100 sub-PR 1 as a new file: a delved spell
		// on the stack — PaidCost.Delved, the objects in exile, on disk.
		{"delved_spell_on_stack", corpusDelvedSpellOnStack},
		// v7, added by ADR 0100 sub-PR 3 as a new file: an either/or
		// spell on the stack — PaidCost.CostBranch and
		// PaidCost.Discarded on disk.
		{"either_or_spell_on_stack", corpusEitherOrSpellOnStack},
		// v7, added by ADR 0103 as a new file: a Room on the battlefield
		// with one door unlocked (Card.Unlocked on disk) and a Room spell
		// on the stack cast as its RIGHT half (ActiveFace 1 on a split
		// card whose catalog key stays bare).
		{"room_doors", corpusRoomDoors},
		// v7, added by ADR 0104 (#1745) as a new file: control of a
		// spell — a setController record pinned to a spell on the stack
		// (affected onStack/epoch, duration PinnedOnStack/PinnedEpoch,
		// StackItem.baseController), and a stolen permanent spell that
		// has resolved, its record re-pinned to the permanent by epoch
		// and its baseController the caster.
		{"stolen_spell", corpusStolenSpell},
		// v7, added by ADR 0100 sub-PR 2 as a new file: the CR 607.2q
		// delve link on a permanent (CastProvenance.Delved) and on a
		// departed one's last-known information (PermanentInfo.Delved).
		{"delve_linked_permanents", corpusDelveLinkedPermanents},
		// v7, added by ADR 0106 PR 3 (#1806) as a new file: "can't be
		// countered" as data — a turn grant and an unspent one-use
		// promise on a seat (PlayerStatic.CantBeCountered), and a spell
		// on the stack with two marks (StackItem.CantBeCountered), one
		// from a spent promise and one from Vexing Shusher.
		{"counter_shields", corpusCounterShields},
		// v7, added by ADR 0107 PR 3 (#1854) as new files: rebound as
		// data — a rebound card in exile with its upkeep delayed
		// trigger queued (body rebound/cast, the exiled object in
		// Params.Object), the same trigger fired and waiting on the
		// stack, and the free cast granted after "yes" (a CastPermission
		// whose LapseOnPass is "exile").
		{"rebound_waiting", corpusReboundWaiting},
		{"rebound_on_stack", corpusReboundOnStack},
		{"rebound_free_cast_grant", corpusReboundFreeCastGrant},
		// v7, added by ADR 0107 PR 6 (#1853, #1880) as a new file: the
		// rules gates as data — a resolved Skullcrack's
		// damageCantBePrevented and cantGainLife turn grants, Flames of
		// the Blood Hand's gainNoLife replacement, a pinned
		// damageCantBePrevented + damageCantBeRedirected pair (Whippoorwill's
		// shape) and a rest-of-the-game cantGainLife on one player.
		{"rules_gates", corpusRulesGates},
		// v7, added by ADR 0107 PR 4 (#1854) as a new file: a spell
		// GIVEN rebound — a ScopedEffect pinned to the spell on the
		// stack with an addKeywords mod, the record Taigam's trigger
		// writes, with a Taigam on the battlefield as its source.
		{"granted_rebound_on_stack", corpusGrantedReboundOnStack},
		// v7, added by ADR 0107 PR 7 (#1860) as a new file: the next-damage
		// shield as data — a chosen source with a colour recheck and a
		// follow-up body protecting a player and their creatures, already
		// spent in this batch (SpentBatch); a no-choice "creature of the
		// chosen type" shield; and a shield pinned to one permanent.
		{"next_damage_shields", corpusNextDamageShields},
		// v7, added by ADR 0108 PR 1 (#1886, #1887) as a new file: the
		// turn-scoped death and regeneration marks as data — a pinned
		// exileIfWouldDie (Lava Coil), the same kind over the live
		// creatures and opponentsCreatures scopes (Flaying Tendrils,
		// Malicious Eclipse), a pinned cantBeRegenerated (Incinerate),
		// and Whippoorwill's "when the creature dies this turn" delayed
		// trigger waiting on its object.
		{"exile_if_dies", corpusExileIfDies},
		// v7, added by ADR 0108 PR 3 (#1823) as a new file: a resolved
		// Yawgmoth's Will — the exileInsteadOfYourGraveyard record (a
		// game-wide replacement naming one player) and the stored
		// ScopeStanding graveyard cast permission written beside it, with
		// the Will itself already exiled by its own replacement.
		{"yawgmoths_will", corpusYawgmothsWill},
		// v7, added by ADR 0108 PR 2 (#1890) as a new file: the
		// multiplyDamage kind in each of its shapes — Insult's "your
		// sources" (a resolved Insult, beside its can't-be-prevented
		// grant), Isengard's triple to opponents and their permanents,
		// Lightning's "that player and their permanents" until your next
		// turn, Blind Fury's combat-only creature-to-creature, and a
		// pinned "next time" multiplier already spent by its instance.
		{"multiply_damage", corpusMultiplyDamage},
		// v7, added by ADR 0108 PR 2 (#1890) as a new file: Impulsive
		// Maneuvers' losing flip — a preventNextCombatFromSource shield on
		// an attacking creature, its own kind so that a binary from before
		// it refuses the file rather than reading an ordinary next-damage
		// shield.
		{"next_combat_damage_shield", corpusNextCombatDamageShield},
		// v7, added by ADR 0109 PR 5 (#1895) as a new file: a resolved
		// Turf Wound — the cantPlayLands record (a game-scope rule kind
		// naming the one banned player) beside a standing Territorial
		// Dispute, whose land-play restriction is catalog data and so adds
		// nothing to the file but the permanent.
		{"cant_play_lands", corpusCantPlayLands},
		// v7, added by ADR 0109 PR 1 (#1881) as a new file: CR 305.7
		// from a resolved effect as data — setBasicLandTypes records
		// until end of turn (Tidal Warrior), until the land's controller's
		// next turn (Orcish Farmer), for as long as the source remains
		// (Gaea's Liege) and indefinitely (Thelonite Monk), and an
		// addSubtypes Forest in addition to the land's own types
		// (Navigator's Compass).
		{"land_types", corpusLandTypes},
		// v7, added by ADR 0109 PR 2 (#1894, #1604) as a new file: a
		// duration with two conditions (Seasinger's "for as long as you
		// control this creature and this creature remains tapped",
		// Duration.Also) and a counter-held one (Minas Morgul's "for as
		// long as that creature has a shadow counter on it",
		// WhilePinnedHasCounter with its CounterKind).
		{"durations", corpusDurations},
		// v7, added by ADR 0109 PR 3 (#1604) as a new file: the
		// loseLandTypes kind, in Ultima, Origin of Oblivion's record
		// (loses all land types and abilities and has "{T}: Add {C}"
		// for as long as the land has a blight counter on it).
		{"lose_land_types", corpusLoseLandTypes},
	}
}

// corpusLoseLandTypes is ADR 0109 §2's loseLandTypes kind, made by the
// card that writes it: Ultima attacks and blights a land.
func corpusLoseLandTypes(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	llAttackWithUltima(t, g, me.ID, opp.ID, land)
	advanceTo(t, g, game.StepPostcombatMain)
	if len(g.ScopedEffects) != 1 {
		t.Fatalf("setup: %d scoped records, want 1", len(g.ScopedEffects))
	}
	return g
}

// corpusDurations is ADR 0109's new duration fields, made by the cards
// that write them.
func corpusDurations(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushLandFor(g, me.ID, "Island", "Basic Land — Island") // or Seasinger's state trigger sacrifices it
	pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	seasinger := pushCatalogPermanent(g, me.ID, "Seasinger", "Creature — Merfolk", rtSeasingerOracle, false)
	victim := ctrlPushCreature(g, opp.ID, "Bear")
	if err := g.ActivateCatalogAbility(me.ID, seasinger, 0, game.ActivateAbilityParams{Targets: ltCardTarget(victim)}); err != nil {
		t.Fatalf("setup: Seasinger: %v", err)
	}
	passPriorityAroundTable(t, g)
	morgul := pushCatalogPermanent(g, me.ID, "Minas Morgul, Dark Fortress", "Legendary Land", rtMinasMorgulOracle, false)
	mine := ctrlPushCreature(g, me.ID, "Wolf")
	b06AddMana(me, "B", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, morgul, 0, game.ActivateAbilityParams{Targets: ltCardTarget(mine)}); err != nil {
		t.Fatalf("setup: Minas Morgul: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.ScopedEffects) != 2 {
		t.Fatalf("setup: %d scoped records, want 2", len(g.ScopedEffects))
	}
	return g
}

// corpusNextCombatDamageShield is Impulsive Maneuvers on the
// battlefield and the combat-only next-damage shield its losing flip
// makes on an attacker.
func corpusNextCombatDamageShield(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	maneuvers := b12Push(g, me.ID, "Impulsive Maneuvers", "Enchantment", "39a9323d-dddc-42ac-929d-3f4fa7c87567", 0, 0)
	attacker := pushBattlefieldCardWithTimestamp(g, corpusCreature(me.ID, "Raider", 3, 3))
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		ref, zone, ok := g.DamageSourceRefLocked(attacker)
		if !ok {
			t.Fatal("setup: the attacker is in no zone")
		}
		g.PreventNextDamageFromSourceForEffect(game.NextDamageShield{
			EffectSource: maneuvers, Controller: me.ID, Source: ref, SourceZone: zone, CombatOnly: true,
			Label: "Impulsive Maneuvers — prevent the next damage",
		})
	})
	if n := len(g.ScopedEffects); n != 1 || g.ScopedEffects[0].Mods[0].Kind != game.ModPreventNextCombatFromSource {
		t.Fatalf("setup: %d scoped records, want the one combat-only shield", n)
	}
	return g
}

// corpusMultiplyDamage is ADR 0108 §3's multiplier in each of its shapes.
func corpusMultiplyDamage(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	castCatalogSpell(t, g, "Insult", "Sorcery", "47543892-4d60-4c6b-a6a4-69b9172af01e", nil)
	passPriorityAroundTable(t, g)
	gambler := pushBattlefieldCardWithTimestamp(g, corpusCreature(me.ID, "Gambler", 2, 2))
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		g.MultiplyDamageForEffect(game.DamageMultiplier{Controller: me.ID, Factor: 3, Sources: game.DamageSourcesYours,
			Recipients: game.DamageRecipientsOpponentsAndTheirPermanents, Label: "Isengard Unleashed"})
		g.MultiplyDamageForEffect(game.DamageMultiplier{Controller: me.ID, Factor: 2,
			Recipients: game.DamageRecipientsPlayerAndTheirPermanents, Player: opp.ID, UntilNextTurnOf: me.ID,
			Label: "Lightning, Army of One — Stagger"})
		g.MultiplyDamageForEffect(game.DamageMultiplier{Controller: me.ID, Factor: 2, Sources: game.DamageSourcesCreatures,
			Recipients: game.DamageRecipientsCreatures, CombatOnly: true, Label: "Blind Fury"})
		ref, zone, ok := g.DamageSourceRefLocked(gambler)
		if !ok {
			t.Fatal("setup: the gambler is in no zone")
		}
		g.MultiplyDamageForEffect(game.DamageMultiplier{Controller: me.ID, Factor: 2, Source: ref, SourceZone: zone,
			Next: true, Label: "Desperate Gambit — double the next damage"})
		// The next-time multiplier doubles the gambler's damage and is
		// spent for the rest of this batch.
		if err := g.DealDamageToPlayerForEffect(gambler, opp.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	n := 0
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModMultiplyDamage {
				n++
			}
		}
	}
	if n != 5 {
		t.Fatalf("setup: %d multipliers, want 5", n)
	}
	return g
}

// corpusLandTypes is ADR 0109 §1's setBasicLandTypes kind in each
// duration the catalog writes it with, made by the cards that write it.
func corpusLandTypes(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lands := make([]uuid.UUID, 4)
	for i := range lands {
		lands[i] = pushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	}
	mine := pushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	// Gaea's Liege is a */* counting your Forests: one keeps it alive.
	pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	elf := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Llanowar Elves",
		TypeLine: "Creature — Elf Druid", Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	activate := func(name, typeLine, oracle string, params game.ActivateAbilityParams) {
		t.Helper()
		src := pushCatalogPermanent(g, me.ID, name, typeLine, oracle, false)
		if err := g.ActivateCatalogAbility(me.ID, src, 0, params); err != nil {
			t.Fatalf("setup: %s: %v", name, err)
		}
		passPriorityAroundTable(t, g)
	}
	activate("Tidal Warrior", "Creature — Merfolk Warrior", ltTidalWarriorOracle, game.ActivateAbilityParams{Targets: ltCardTarget(lands[0])})
	activate("Orcish Farmer", "Creature — Orc", ltOrcishFarmerOracle, game.ActivateAbilityParams{Targets: ltCardTarget(lands[1])})
	activate("Gaea's Liege", "Creature — Avatar", ltGaeasLiegeOracle, game.ActivateAbilityParams{Targets: ltCardTarget(lands[2])})
	activate("Thelonite Monk", "Creature — Insect Monk Cleric", ltTheloniteMonkOracle,
		game.ActivateAbilityParams{Targets: ltCardTarget(lands[3]), SacrificeIDs: []uuid.UUID{elf}})
	activate("Navigator's Compass", "Artifact", ltNavigatorsCompassOracle, game.ActivateAbilityParams{Targets: ltCardTarget(mine)})
	answerOptionPick(t, g, me.ID, 4) // Forest
	if len(g.ScopedEffects) != 5 {
		t.Fatalf("setup: %d scoped records, want 5", len(g.ScopedEffects))
	}
	return g
}

// corpusExileIfDies is ADR 0108 §1 and §2's two kinds in each of their
// shapes, made by the cards that make them.
func corpusExileIfDies(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := pushBattlefieldCardWithTimestamp(g, corpusCreature(opp.ID, "Hill Giant", 3, 10))
	troll := pushBattlefieldCardWithTimestamp(g, corpusCreature(opp.ID, "Troll Ascetic", 3, 10))
	bear := pushBattlefieldCardWithTimestamp(g, corpusCreature(opp.ID, "Grizzly Bears", 2, 8))
	castCatalogSpell(t, g, "Lava Coil", "Sorcery", p1LavaCoilOracle, pr6Card(ogre))
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Incinerate", "Instant", p1IncinerateOracle, pr6Card(troll))
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Flaying Tendrils", "Sorcery", p1FlayingOracle, nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Malicious Eclipse", "Sorcery", p1EclipseOracle, nil)
	passPriorityAroundTable(t, g)
	bird := pushCatalogPermanent(g, me.ID, "Whippoorwill", "Creature — Bird", p1WhippoorwillOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{Targets: pr6Card(bear)}); err != nil {
		t.Fatalf("setup: Whippoorwill: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("setup: %d delayed triggers, want Whippoorwill's", len(g.DelayedTriggers))
	}
	return g
}

// corpusYawgmothsWill is a real Yawgmoth's Will after it resolved.
func corpusYawgmothsWill(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest")
	will := castCatalogSpell(t, g, "Yawgmoth's Will", "Sorcery", yawgmothsWillOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(will) || len(g.ScopedEffects) != 1 || len(me.CastPermissions) != 1 {
		t.Fatalf("setup: Will exiled=%v, %d scoped records, %d permissions; want exiled, 1 and 1",
			g.Exile.Contains(will), len(g.ScopedEffects), len(me.CastPermissions))
	}
	return g
}

// corpusGrantedReboundOnStack is Lightning Bolt cast from hand and
// given rebound on the stack by Taigam, Ojutai Master (#1854).
func corpusGrantedReboundOnStack(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	foe := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	taigam := pushCatalogPermanent(g, me.ID, "Taigam, Ojutai Master", "Legendary Creature — Human Monk", taigamOracle, false)
	id := castCatalogSpell(t, g, "Lightning Bolt", "Instant", boltOracleCombat,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: foe.ID}})
	var ok bool
	g.WithWriteLock(func() {
		ok = g.GrantKeywordsToSpellForEffect(taigam, id, []string{game.KeywordRebound},
			"Taigam, Ojutai Master — that spell gains rebound")
	})
	if !ok {
		t.Fatal("setup: Lightning Bolt was not given rebound")
	}
	return g
}

// corpusNextDamageShields is ADR 0107 §6's ModPreventNextFromSource in
// each of its shapes.
func corpusNextDamageShields(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dragon := pushBattlefieldCardWithTimestamp(g, corpusCreature(opp.ID, "Shivan Dragon", 5, 5))
	knight := pushBattlefieldCardWithTimestamp(g, corpusCreature(me.ID, "Knight", 2, 2))
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		ref, zone, ok := g.DamageSourceRefLocked(dragon)
		if !ok {
			t.Fatal("setup: the dragon is in no zone")
		}
		g.PreventNextDamageFromSourceForEffect(game.NextDamageShield{
			Controller: me.ID, Source: ref, SourceZone: zone, Queries: []game.PermanentQuery{QueryColors("R")},
			ProtectPlayer: me.ID, ProtectTypes: []string{"creature"}, Then: preventedGainLifeBody,
			Label: "Shadowbane — prevent the next damage from a source",
		})
		g.PreventNextDamageFromSourceForEffect(game.NextDamageShield{
			Controller: me.ID, ProtectPlayer: me.ID,
			Queries: []game.PermanentQuery{{Types: []string{"creature"}, Subtypes: []string{"Dragon"}}},
			Label:   "Circle of Solace — prevent the next damage from a source",
		})
		g.PreventNextDamageFromSourceForEffect(game.NextDamageShield{
			Controller: me.ID, Source: ref, SourceZone: zone, ProtectPermanent: knight,
			Label: "Charm Peddler — prevent the next damage from a source",
		})
		// The first shield takes the dragon's damage to the knight and is
		// spent for the rest of this batch.
		if err := g.DealDamageToCreatureForEffect(dragon, knight, 1); err != nil {
			t.Fatal(err)
		}
	})
	if n := len(g.ScopedEffects); n != 3 {
		t.Fatalf("setup: %d scoped records, want 3", n)
	}
	return g
}

// corpusReboundWaiting is Staggershock cast from hand, resolved and
// exiled by rebound, its delayed trigger queued for its controller's
// next upkeep (#1854).
func corpusReboundWaiting(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	foe := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := castCatalogSpell(t, g, "Staggershock", "Instant", "056c3b7d-b603-40b8-8404-18c2eb7e7129",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: foe.ID}})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) || len(g.DelayedTriggers) != 1 {
		t.Fatalf("setup: exiled %v, delayed triggers %d", g.Exile.Contains(id), len(g.DelayedTriggers))
	}
	return g
}

// corpusReboundOnStack is that trigger fired at the controller's next
// upkeep and waiting on the stack, a keyed item (#1854).
func corpusReboundOnStack(t *testing.T) *game.Game {
	g := corpusReboundWaiting(t)
	seat := g.Turn.ActiveSeat
	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	advanceToUpkeepOf(t, g, seat)
	found := false
	for _, it := range g.PendingTriggers {
		if it != nil && it.Body == "rebound/cast" {
			found = true
		}
	}
	for _, it := range g.StackMeta {
		if it != nil && it.Body == "rebound/cast" {
			found = true
		}
	}
	if !found {
		t.Fatal("setup: the rebound trigger is not waiting")
	}
	return g
}

// corpusReboundFreeCastGrant is the offer accepted: the free cast is a
// per-object CastPermission that closes on its holder's pass and leaves
// the card in exile (#1854).
func corpusReboundFreeCastGrant(t *testing.T) *game.Game {
	g := corpusReboundOnStack(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 8 && latestChoiceOfKind(g, game.PendingChoiceMayCast) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	offer := latestChoiceOfKind(g, game.PendingChoiceMayCast)
	if offer == nil {
		t.Fatal("setup: no rebound offer")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	granted := false
	for _, perm := range me.CastPermissions {
		if perm.LapseOnPass == game.LapseStaysInExile {
			granted = true
		}
	}
	if !granted {
		t.Fatal("setup: no rebound grant")
	}
	return g
}

// corpusRulesGates is ADR 0107 §5's four new mod kinds on one board.
func corpusRulesGates(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	castCatalogSpell(t, g, "Skullcrack", "Instant", pr6SkullcrackOracle, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Flames of the Blood Hand", "Instant", pr6FlamesOracle, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	bear := pushBattlefieldCardWithTimestamp(g, corpusCreature(opp.ID, "Grizzly Bears", 2, 2))
	g.WithWriteLock(func() {
		g.DamageToCantBePreventedThisTurnForEffect(uuid.Nil, bear, true, "Whippoorwill")
		g.PlayerCantGainLifeForEffect(uuid.Nil, me.ID, game.IndefiniteDuration(), "Screaming Nemesis")
	})
	if n := len(g.ScopedEffects); n != 5 {
		t.Fatalf("setup: %d scoped records, want 5 (two grants, a replacement, a pinned pair, a rest-of-game)", n)
	}
	return g
}

// corpusCantPlayLands is ADR 0109 §4's one stored shape: Turf Wound
// resolved on the opponent, so a cantPlayLands record names them.
func corpusCantPlayLands(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Territorial Dispute", "Enchantment", lpTerritorialDisputeOracle, false)
	castCatalogSpell(t, g, "Turf Wound", "Instant", lpTurfWoundOracle, pr6Player(opp.ID))
	passPriorityAroundTable(t, g)
	if n := len(g.ScopedEffects); n != 1 || g.ScopedEffects[0].Mods[0].Kind != game.ModCantPlayLands {
		t.Fatalf("setup: scoped records = %+v, want one cantPlayLands", g.ScopedEffects)
	}
	return g
}

// corpusCounterShields is ADR 0106 §4's three stored shapes at once.
// Insist resolved and its promise was spent by the Grizzly Bears on the
// stack, which Vexing Shusher then marked again; Veil of Summer's turn
// grant is on the caster, and so is Mistrise Village's promise, which
// nothing has spent yet.
func corpusCounterShields(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Insist", "Sorcery", cgInsistOracle, nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Veil of Summer", "Instant", cgVeilOfSummerOracle, nil)
	passPriorityAroundTable(t, g)
	village := pushCatalogPermanent(g, me.ID, "Mistrise Village", "Land", cgMistriseVillageOracle, false)
	shusher := pushCatalogPermanent(g, me.ID, "Vexing Shusher", "Creature — Goblin Shaman", cgVexingShusherOracle, false)
	bear := castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, village, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("setup: Mistrise Village: %v", err)
	}
	cgResolveTop(t, g)
	if err := g.ActivateCatalogAbility(me.ID, shusher, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("setup: Vexing Shusher: %v", err)
	}
	cgResolveTop(t, g)
	if it := g.StackMeta[bear]; it == nil || len(it.CantBeCountered) != 2 || len(me.Statics) != 2 {
		t.Fatalf("setup: want the Bears with two marks and two statics on the caster, got %+v / %+v", it, me.Statics)
	}
	return g
}

// corpusDelveLinkedPermanents is a Murktide Regent on the battlefield
// that delved two instants and entered with their counters, and an
// Ethereal Forager that delved one and has since left: the permanent
// carries its link to the cards in exile, and the Forager's last-known
// information carries its own, which is how its attack trigger finds
// them after it has gone (ADR 0100 sub-PR 2).
func corpusDelveLinkedPermanents(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	murktide, err := castWithTapParams(t, g, "Murktide Regent", "Creature — Dragon", "{5}{U}{U}",
		murktideRegentOracle, game.CastSpellParams{DelveIDs: delveFuel(me, 2)})
	if err != nil {
		t.Fatalf("setup: cast Murktide Regent: %v", err)
	}
	passPriorityAroundTable(t, g)
	forager, err := castWithTapParams(t, g, "Ethereal Forager", "Creature — Elemental Whale", "{4}{U}{U}",
		etherealForagerOracle, game.CastSpellParams{DelveIDs: delveFuel(me, 1)})
	if err != nil {
		t.Fatalf("setup: cast Ethereal Forager: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(forager); err != nil {
			t.Fatalf("setup: bounce the Forager: %v", err)
		}
	})
	var linked int
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == murktide {
			linked = len(c.Delved())
		}
	}
	if linked != 2 {
		t.Fatalf("setup: want Murktide Regent on the battlefield linked to two cards, got %d", linked)
	}
	var lki game.PermanentInfo
	var ok bool
	g.WithWriteLock(func() { lki, ok = g.LastKnownPermanentForEffect(forager) })
	if !ok || len(lki.Delved) != 1 {
		t.Fatal("setup: want the departed Forager's last-known information to carry its delve link")
	}
	return g
}

// corpusStolenSpell is ADR 0104 on disk: a creature spell stolen on the
// stack and resolved under the thief, then an instant stolen and still
// waiting on the stack.
func corpusStolenSpell(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	advanceToMain(t, g)
	caster := g.Seats[g.Turn.ActiveSeat]
	thief := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cast := func(name, typeLine string) uuid.UUID {
		c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: "{0}",
			Owner: caster.ID, Controller: caster.ID}
		if c.IsCreature() {
			c.Power, c.Toughness = 2, 2
		}
		g.WithWriteLock(func() { caster.Hand.PushTop(c) })
		if err := g.CastSpell(caster.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
			t.Fatalf("setup: cast %s: %v", name, err)
		}
		g.WithWriteLock(func() {
			if !g.GainControlOfSpellForEffect(uuid.Nil, c.InstanceID, thief.ID, "corpus — Aethersnatch") {
				t.Fatalf("setup: the steal of %s registered nothing", name)
			}
		})
		return c.InstanceID
	}
	cast("Grizzly Bears", "Creature — Bear")
	passPriorityAroundTable(t, g)
	cast("Divination", "Sorcery")
	return g
}

// corpusRoomDoors is a Room that entered with its right door unlocked
// (CR 709.5d) and a second Room's right half waiting on the stack.
func corpusRoomDoors(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	// An oracle ID no test registers: the board is the engine's Room
	// lifecycle, not a card file, and a test that registered this
	// Room's ID would leave its definition behind for the corpus run.
	room := func() game.Card {
		c := testRoomCard(me.ID)
		c.OracleID = "corpus-room-doors-oracle"
		return c
	}
	first, second := room(), room()
	me.Hand.PushTop(first)
	me.Hand.PushTop(second)
	if err := g.CastSpell(me.ID, first.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("setup: cast the first Room's right half: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(first.InstanceID) {
		t.Fatal("setup: the first Room did not resolve")
	}
	if err := g.CastSpell(me.ID, second.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("setup: cast the second Room's right half: %v", err)
	}
	return g
}

// corpusEitherOrSpellOnStack is a Demand Answers on the stack that paid
// its discard branch (ADR 0100 §2): the record names the branch and the
// discarded card, which a restore has to carry for "the discarded card"
// (Grab the Prize) and "if the modified creature was sacrificed"
// (Lethal Throwdown).
func corpusEitherOrSpellOnStack(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pitch := handCard(me, "Pitch", "Instant")
	id, err := castWithTapParams(t, g, "Demand Answers", "Instant", "{1}{R}", demandAnswersOracle,
		game.CastSpellParams{CostBranch: branch(1), DiscardIDs: []uuid.UUID{pitch}})
	if err != nil {
		t.Fatalf("setup: CastSpell: %v", err)
	}
	if it := g.StackMeta[id]; it == nil || it.Paid.CostBranch != 2 || len(it.Paid.Discarded) != 1 {
		t.Fatal("setup: want Demand Answers on the stack with its discard branch recorded")
	}
	return g
}

// corpusDelvedSpellOnStack is a Treasure Cruise on the stack that
// delved three cards (CR 702.66a, ADR 0100): its payment record names
// the three objects that landed in exile, which is what a restore has
// to carry for sub-PR 2's "exiled with it" readers (CR 607.2q).
func corpusDelvedSpellOnStack(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fuel := delveFuel(me, 3)
	id, err := castCruise(t, g, game.CastSpellParams{DelveIDs: fuel})
	if err != nil {
		t.Fatalf("setup: CastSpell: %v", err)
	}
	if it := g.StackMeta[id]; it == nil || len(it.Paid.Delved) != 3 {
		t.Fatal("setup: want Treasure Cruise on the stack with three delved cards")
	}
	return g
}

// corpusKeywordCounters is two Bears, each with a flying counter and a
// "loses all abilities" scoped effect. On the first the counter came
// first, so the removal takes flying; on the second the removal came
// first, so the counter's flying applies after it (CR 613.3, 613.7c).
func corpusKeywordCounters(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[0].ID
	grounded := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Grizzly Bears", 2, 2))
	flier := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Runeclaw Bear", 2, 2))
	lose := func(id uuid.UUID) {
		g.WithWriteLock(func() {
			if !g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(id),
				[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "corpus — loses all abilities") {
				t.Fatal("setup: the removal registered nothing")
			}
		})
	}
	counter := func(id uuid.UUID) {
		g.WithWriteLock(func() {
			if err := g.AddCounterForEffect(id, game.CounterFlying, 1); err != nil {
				t.Fatalf("setup: flying counter: %v", err)
			}
		})
	}
	counter(grounded)
	lose(grounded)
	lose(flier)
	counter(flier)
	return g
}

// corpusDurationCopy is a Clone of Grizzly Bears that Cytoshape turned
// into a Hill Giant until end of turn (a becomeCopy record, and the
// Clone's entry copy as its durationCopyBase), beside an Unstable
// Shapeshifter that has become a copy of a Runeclaw Bear for good (an
// indefinite record whose copied values carry a granted bundle).
func corpusDurationCopy(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	bears := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Grizzly Bears", 2, 2))
	giant := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Hill Giant", 3, 3))
	clone := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, bears)
	castCatalogSpell(t, g, "Cytoshape", "Instant", oracleCytoshape,
		[]game.TargetRef{{Kind: game.TargetCard, ID: clone}})
	dcAnswerChooseCards(t, g, giant)
	passPriorityAroundTable(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Unstable Shapeshifter", TypeLine: "Creature — Shapeshifter",
		OracleID: oracleUnstableShapeshifter, Power: 0, Toughness: 1, Owner: me, Controller: me,
	})
	dcEnter(t, g, me, "Runeclaw Bear", "Creature — Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if len(g.ScopedEffects) != 2 {
		t.Fatalf("setup: want the Cytoshape and the Shapeshifter records, have %d", len(g.ScopedEffects))
	}
	return g
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

// ---------------------------------------------------------------
// v7 boards added by ADR 0041 phase 3's tier 3b-2 (#1497)
// ---------------------------------------------------------------
//
// NEW files in v7/: a block rule a spell or ability created held the
// restore point back until cleanup before tier 3b-2, so neither of
// these could be a fixture.

// corpusGingerbrute is a real Gingerbrute activation: the
// cantBeBlockedExceptBy record pinned to it, Keywords ["haste"] and
// Text on disk.
func corpusGingerbrute(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	brute := pushCatalogPermanent(g, me, "Gingerbrute", "Artifact Creature — Food Golem", gingerbruteOracle, false)
	gingerbruteActivate(t, g, me, brute)
	if n := scopedBlockRuleCount(g); n != 1 {
		t.Fatalf("setup: Gingerbrute registered %d scoped block rules, want 1", n)
	}
	return g
}

// corpusMirriLimit is a real Mirri, Weatherlight Duelist attack: the
// limitBlockersPerDefender record (ScopeOpponentsCreatures) her attack
// trigger registered.
func corpusMirriLimit(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	mirri := b12Push(g, me.ID, "Mirri, Weatherlight Duelist", "Legendary Creature — Cat Warrior", mirriWeatherlightDuelistOracle, 3, 2)
	declareAttack(t, g, opp.ID, mirri)
	passPriorityAroundTable(t, g)
	if n := scopedBlockRuleCount(g); n != 1 {
		t.Fatalf("setup: Mirri's trigger registered %d scoped block rules, want 1", n)
	}
	return g
}

// corpusFalter is a real Falter: one addRestrictions record over the
// live creaturesWithoutFlying scope (#1650), with a creature on the
// board it reaches.
func corpusFalter(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	seat := g.Turn.ActiveSeat
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pushSizedCreature(g, opp.ID, "Grizzly Bears", 2, 2)
	castCatalogSpell(t, g, "Falter", "Instant", falterOracle, nil)
	passPriorityAroundTable(t, g)
	if len(g.ScopedEffects) != 1 || g.ScopedEffects[0].Scope != game.ScopeCreaturesWithoutFlying {
		t.Fatalf("setup: Falter registered %+v, want one creaturesWithoutFlying record", g.ScopedEffects)
	}
	return g
}

// corpusDetectionTower is a real Detection Tower activation: one
// waiveHexproof record over the live opponentsAndTheirCreatures scope
// (#1651), with a hexproof creature on the board it reaches.
func corpusDetectionTower(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushSizedCreature(g, opp.ID, "Slippery Bogle", 1, 1, "hexproof")
	activateLand(t, g, "Detection Tower", detectionTowerOracle)
	if len(g.ScopedEffects) != 1 || g.ScopedEffects[0].Scope != game.ScopeOpponentsAndTheirCreatures {
		t.Fatalf("setup: Detection Tower registered %+v, want one opponentsAndTheirCreatures record", g.ScopedEffects)
	}
	return g
}

// corpusArcaneLighthouse is a real Arcane Lighthouse activation: one
// cantHaveKeywords record pinned to the opponent's hexproof creature
// (#1651), so CantHave is on disk in the characteristic's shape too.
func corpusArcaneLighthouse(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushSizedCreature(g, opp.ID, "Slippery Bogle", 1, 1, "hexproof")
	activateLand(t, g, "Arcane Lighthouse", arcaneLighthouseOracle)
	if len(g.ScopedEffects) != 1 || g.ScopedEffects[0].Mods[0].Kind != game.ModCantHaveKeywords {
		t.Fatalf("setup: Arcane Lighthouse registered %+v, want one cantHaveKeywords record", g.ScopedEffects)
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
// v7 boards added by ADR 0041 phase 3's tier 4, second slice (#1497)
// ---------------------------------------------------------------
//
// NEW files in v7/: a triggered ability waiting on the stack held the
// restore point back until it resolved before this slice. Each board
// asserts the stamp it is there to freeze, so a card that slips back to
// a hand-written Build fails here rather than writing a fixture that
// proves nothing.

const (
	corpusMulldrifterOracle = "24d0f5e7-0d9e-4b76-900e-a7274e80312d"
	corpusTeferiHeroOracle  = "f2f165b6-ef0a-42ad-9352-ba68be8248b0"
)

// corpusSettleTrigger passes priority until a triggered item from
// `source` is on the stack, and returns it.
func corpusSettleTrigger(t *testing.T, g *game.Game, source uuid.UUID) *game.StackItem {
	t.Helper()
	for i := 0; i < 8; i++ {
		if it := triggerOnStack(g, source); it != nil {
			return it
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("no trigger from %s reached the stack", source)
	return nil
}

// corpusRequireTriggeredStamp fails unless the item names a triggered
// catalog row whose ref starts with `refPrefix`.
func corpusRequireTriggeredStamp(t *testing.T, it *game.StackItem, refPrefix string) {
	t.Helper()
	if it == nil || it.Body != game.CatalogTriggeredBodyKey || it.Params.Ability == nil ||
		it.Params.Ability.Slot != game.AbilitySlotTriggered || !strings.HasPrefix(it.Params.Ability.Ref, refPrefix) {
		t.Fatalf("setup: the trigger is not stamped with a %s… triggered ref: %+v", refPrefix, it)
	}
}

// corpusETBTriggerOnStack is a real Mulldrifter's "when this enters,
// draw two cards" waiting on the stack: an own:0 triggered ref.
func corpusETBTriggerOnStack(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	drifter := castCatalogSpell(t, g, "Mulldrifter", "Creature — Elemental", corpusMulldrifterOracle, nil)
	corpusRequireTriggeredStamp(t, corpusSettleTrigger(t, g, drifter), "own:")
	return g
}

// corpusGrantedDiesTrigger is Feign Death's GRANTED "when this creature
// dies, return it" waiting on the stack: the key comes off the dead
// creature's last-known information, and the ref names the bundle.
func corpusGrantedDiesTrigger(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	bear := pushBattlefieldCardWithTimestamp(g, corpusCreature(me, "Grizzly Bears", 2, 2))
	castCatalogSpell(t, g, "Feign Death", "Instant", feignDeathOracle, dgCardRef(bear))
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	corpusRequireTriggeredStamp(t, corpusSettleTrigger(t, g, bear), "grant:")
	return g
}

// corpusTokenTrigger is a real Pest token's "when this token dies, you
// gain 1 life" waiting on the stack: a token template's row, whose
// source has ceased to exist.
func corpusTokenTrigger(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	pest := PestToken()
	pest.InstanceID, pest.Owner, pest.Controller = uuid.New(), me, me
	id := pushBattlefieldCardWithTimestamp(g, pest)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	it := corpusSettleTrigger(t, g, id)
	corpusRequireTriggeredStamp(t, it, "own:")
	if it.Params.Ability.Key != game.TokenKey("pest") {
		t.Fatalf("setup: the Pest's trigger names %q, want the token key", it.Params.Ability.Key)
	}
	return g
}

// corpusEmblemTrigger is Teferi, Hero of Dominaria's emblem —
// "whenever you draw a card, exile target permanent an opponent
// controls" — triggered by a draw, its target chosen, waiting on the
// stack: an emblem's row, with a target clause.
func corpusEmblemTrigger(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	victim := pushBattlefieldCardWithTimestamp(g, corpusCreature(opp, "Grizzly Bears", 2, 2))
	teferi := uuid.New()
	g.Seats[g.Turn.ActiveSeat].Graveyard.PushTop(game.Card{
		InstanceID: teferi, Name: "Teferi, Hero of Dominaria",
		TypeLine: "Legendary Planeswalker — Teferi", OracleID: corpusTeferiHeroOracle,
		Owner: me, Controller: me,
	})
	var err error
	g.WithWriteLock(func() {
		if err = g.CreateEmblemForEffect(me, teferi); err == nil {
			err = g.DrawNForEffect(me, 1)
		}
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	pickTriggerTarget(t, g, me, victim)
	var it *game.StackItem
	for _, item := range g.StackMeta {
		if item.Kind == game.StackItemTriggered {
			it = item
		}
	}
	corpusRequireTriggeredStamp(t, it, "own:")
	if it.Params.Ability.Key != game.EmblemKey(corpusTeferiHeroOracle) {
		t.Fatalf("setup: the emblem's trigger names %q, want the emblem key", it.Params.Ability.Key)
	}
	return g
}

// corpusModalTrigger is a real Gala Greeters' alliance trigger with its
// mode chosen (CR 603.3c), waiting on the stack: the row's mode clause
// comes back from the catalog on restore.
func corpusModalTrigger(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	greeters := b12Push(g, me, "Gala Greeters", "Creature — Elf Druid", b15GalaGreetersOracle, 1, 1)
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me)
	if c == nil {
		t.Fatal("setup: the alliance trigger asked for no mode")
	}
	if err := g.ResolveModePick(c.ID, me, []int{1}); err != nil {
		t.Fatalf("setup: ResolveModePick: %v", err)
	}
	corpusRequireTriggeredStamp(t, triggerOnStack(g, greeters), "own:")
	return g
}

// corpusTargetsFromTrigger is a real Gixian Puppeteer's dies trigger —
// "return ANOTHER target creature card …", whose clause TargetsFrom
// builds from the trigger's own source — with its target chosen,
// waiting on the stack. Restore calls TargetsFrom again.
func corpusTargetsFromTrigger(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	me := seat.ID
	puppeteer := b12Push(g, me, "Gixian Puppeteer", "Creature — Phyrexian Warlock", b40GixianPuppeteerOracle, 2, 3)
	elf := uuid.New()
	seat.Graveyard.PushTop(game.Card{
		InstanceID: elf, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		ManaCost: "{G}", Power: 1, Toughness: 1, Owner: me, Controller: me,
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(puppeteer) })
	pickTriggerTarget(t, g, me, elf)
	corpusRequireTriggeredStamp(t, triggerOnStack(g, puppeteer), "own:")
	return g
}

// corpusProwessOnStack is a prowess trigger — an engine trigger with no
// catalog row, keyed by tier 4-0's prowess/pump body — waiting on the
// stack above the Lightning Bolt that triggered it.
func corpusProwessOnStack(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	monk := pushProwessCreature(g, me, "Monastery Swiftspear", 1, 2, game.KeywordProwess, "haste")
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	if it := corpusSettleTrigger(t, g, monk); it.Body != "prowess/pump" {
		t.Fatalf("setup: the prowess trigger names body %q, want prowess/pump", it.Body)
	}
	return g
}

// corpusEvolveOnStack is an evolve trigger — an engine trigger with no
// catalog row, keyed by its evolve/grow body (#1805) — waiting on the
// stack after a bigger creature entered under its controller's control.
func corpusEvolveOnStack(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	raptor := pushEvolveCreature(g, me, "Cloudfin Raptor", 0, 1, "flying", game.KeywordEvolve)
	enterCreature(t, g, me, "Grizzly Bears", 2, 2)
	if it := corpusSettleTrigger(t, g, raptor); it.Body != "evolve/grow" {
		t.Fatalf("setup: the evolve trigger names body %q, want evolve/grow", it.Body)
	}
	return g
}

// corpusStateTriggerOnStack is a real Emperor Crocodile alone on its
// controller's side, its "When you control no other creatures, sacrifice
// this creature" waiting on the stack: an own:0 triggered ref with no
// trigger context.
func corpusStateTriggerOnStack(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	croc := corpusCreature(me, "Emperor Crocodile", 5, 5)
	croc.TypeLine, croc.OracleID = "Creature — Crocodile", stEmperorCrocodile
	id := pushBattlefieldCardWithTimestamp(g, croc)
	it := corpusSettleTrigger(t, g, id)
	corpusRequireTriggeredStamp(t, it, "own:")
	if it.Trigger != nil {
		t.Fatalf("setup: a state trigger carries a trigger context %+v; it fired off no event", it.Trigger)
	}
	return g
}

// corpusStateEachTriggerOnStack is a real Bomb Squad beside a bear with
// four fuse counters, its "Whenever a creature has four or more fuse
// counters on it" waiting on the stack: an own: triggered ref whose
// trigger context names the bear and no event.
func corpusStateEachTriggerOnStack(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	squad := corpusCreature(me, "Bomb Squad", 1, 1)
	squad.TypeLine, squad.OracleID = "Creature — Dwarf", bsBombSquad
	id := pushBattlefieldCardWithTimestamp(g, squad)
	bear := corpusCreature(me, "Grizzly Bears", 2, 2)
	bear.Counters = map[string]int{"fuse": 4}
	bearID := pushBattlefieldCardWithTimestamp(g, bear)
	it := corpusSettleTrigger(t, g, id)
	corpusRequireTriggeredStamp(t, it, "own:")
	if it.Trigger == nil || it.Trigger.Object == nil || it.Trigger.Object.ID != bearID || it.Trigger.Fired() {
		t.Fatalf("setup: the trigger does not name the bear with no event: %+v", it.Trigger)
	}
	return g
}

// corpusStormAfterCounter is a real Grapeshot's storm trigger waiting
// on the stack after Counterspell countered the Grapeshot: the trigger
// is data, and the spell it copies is in lastKnownStack.
func corpusStormAfterCounter(t *testing.T) *game.Game {
	g := newCorpusGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	passPriorityAroundTable(t, g)
	shot := castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	corpusRequireTriggeredStamp(t, corpusSettleTrigger(t, g, shot), "own:")
	counter := castCatalogSpell(t, g, "Counterspell", "Instant", corpusCounterspellOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: shot}})
	for i := 0; i < 8 && g.Stack.Contains(counter); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	snap := g.CaptureSnapshot()
	if g.Stack.Contains(shot) || triggerOnStack(g, shot) == nil || len(snap.LastKnownStack) != 1 {
		t.Fatalf("setup: grapeshot on the stack %v, storm trigger %v, %d last-known spells — want the storm trigger over a countered Grapeshot",
			g.Stack.Contains(shot), triggerOnStack(g, shot) != nil, len(snap.LastKnownStack))
	}
	return g
}

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

// renderCorpusBoards renders every scripted board, twice, and fails if a
// board comes out different the second time: a board that does not pin
// all of its nondeterminism cannot be a fixture.
func renderCorpusBoards(t *testing.T) map[string][]byte {
	t.Helper()
	rendered := map[string][]byte{}
	for _, b := range corpusBoards() {
		first, second := renderBoard(t, b), renderBoard(t, b)
		if !bytes.Equal(first, second) {
			t.Fatalf("board %q renders differently twice in a row, so it cannot be a fixture — something in it is not pinned:\n%s",
				b.name, firstDifferingLine(first, second))
		}
		rendered[b.name] = first
	}
	return rendered
}

// corpusFixtureDrift compares every rendered board with the fixture
// already written for it in dir. It returns the boards that have no
// file yet, and one entry per existing fixture that is NOT a subset of
// its board's fresh render (#1801).
//
// Subset, not byte identity, because the rule within a schema version is
// additive only (SnapshotSchemaVersion's comment): a field added later in
// the version renders in the fresh capture and is absent from an older
// fixture, and that is exactly what the version allows. What it does not
// allow — a key in the fixture that the build no longer writes, or writes
// with another type or value — is what this reports. The comparison is
// the restore guard's own (corpusSubset, with the same ignore lists), so
// "may this build still write into v<N>?" and "does this build still
// read v<N>?" are one rule, not two.
func corpusFixtureDrift(t *testing.T, dir string, rendered map[string][]byte) (fresh map[string][]byte, drift []string) {
	t.Helper()
	fresh = map[string][]byte{}
	for name, raw := range rendered {
		existing, err := os.ReadFile(filepath.Join(dir, name+".json"))
		if err != nil {
			fresh[name] = raw
			continue
		}
		if diffs := corpusFixtureNotInRender(t, existing, raw); len(diffs) > 0 {
			drift = append(drift, name+".json:\n    "+strings.Join(diffs, "\n    "))
		}
	}
	sort.Strings(drift)
	return fresh, drift
}

// corpusFixtureNotInRender lists every key or value in a fixture's
// snapshot that a fresh render of the same board does not carry.
func corpusFixtureNotInRender(t *testing.T, fixtureRaw, renderRaw []byte) []string {
	t.Helper()
	fixture, err := decodeGeneric(fixtureRaw)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	render, err := decodeGeneric(renderRaw)
	if err != nil {
		t.Fatalf("decode render: %v", err)
	}
	want, _ := fixture.(map[string]any)["snapshot"]
	got, _ := render.(map[string]any)["snapshot"]
	if want == nil {
		return []string{`no "snapshot" in the fixture`}
	}
	var diffs []string
	corpusSubset(want, got, "", game.SnapshotSchemaVersion, &diffs)
	return diffs
}

// corpusDriftAdvice is what to do when a fixture is not a subset of its
// board's render. Shared by the writer and the always-on test.
func corpusDriftAdvice() string {
	return fmt.Sprintf(`Each line is a key or value in a fixture that a fresh render of its
board no longer carries. New keys in the render are fine (additive, the
v%[1]d rule); these are not.

If the snapshot's shape changed — a key renamed, removed or retyped —
that is a new schema version: bump SnapshotSchemaVersion (see its
comment for when that is required) and run the writer again to write
v%[2]d beside the old set. If a scripted board now builds a different
game on purpose, keep the old board building what its fixture holds and
add the new game as a new board. Never edit the fixture.`,
		game.SnapshotSchemaVersion, game.SnapshotSchemaVersion+1)
}

// TestWriteSnapshotCorpus writes the generated fixture set for the
// current schema version. It runs only when asked:
//
//	go test ./internal/cards/effects -run TestWriteSnapshotCorpus -args -write-corpus
//
// The writer NEVER TOUCHES AN EXISTING FILE (ADR 0041's 2026-09-24
// phase 3 amendment, owner decision 6 — narrowed from "never touches an
// existing directory"). For a version whose directory already exists it
// re-renders every board and fails unless every file already there is a
// SUBSET of its board's fresh render (corpusFixtureDrift, #1801): a key
// added later in the version is fine, a key lost or changed is a new
// version, in a new directory. A board with no file yet — one a later
// change under the same version added — is written beside the others,
// which is how a shape that becomes a restore point after its version
// was introduced gets frozen without a bump.
//
// TestSnapshotCorpusBoardsStillMatchTheirFixtures runs the same check on
// every CI run, so this refusal is never the first place drift shows up.
func TestWriteSnapshotCorpus(t *testing.T) {
	if !*writeCorpus {
		t.Skip("writes fixtures; run with -args -write-corpus")
	}
	dir := filepath.Join(corpusRoot, fmt.Sprintf("v%d", game.SnapshotSchemaVersion))
	fresh, drift := corpusFixtureDrift(t, dir, renderCorpusBoards(t))
	if len(drift) > 0 {
		t.Fatalf("%s holds fixtures written by an earlier build of schema v%d, and fixtures are never rewritten.\n\n  %s\n\n%s\n\nNothing was written.",
			dir, game.SnapshotSchemaVersion, strings.Join(drift, "\n  "), corpusDriftAdvice())
	}
	if len(fresh) == 0 {
		t.Logf("%s already holds a fixture for every board; nothing to write", dir)
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(fresh))
	for name := range fresh {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(dir, name+".json")
		// O_EXCL: the one guarantee this function makes is that it
		// never overwrites a fixture, so it asks the filesystem to
		// refuse rather than trusting the read above.
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(fresh[name]); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("wrote %d new fixtures to %s: %s", len(fresh), dir, strings.Join(names, ", "))
}

// TestSnapshotCorpusBoardsStillMatchTheirFixtures is the writer's
// precondition, on every CI run (#1801): each fixture in the current
// version's set is a subset of a fresh render of its board. Without it,
// a change that stops rendering a key (or renders it differently) passes
// CI and is found only by the next person who adds a board — which is
// how #1801 surfaced, two PRs after the change that caused it.
//
// A registered board with no fixture fails too: the writer only adds
// files, so it runs in the change that registers the board. The
// restore guard (TestSnapshotCorpusRestores) is the other half:
// it checks the fixture against a RESTORED game, this against a BUILT
// one.
func TestSnapshotCorpusBoardsStillMatchTheirFixtures(t *testing.T) {
	dir := filepath.Join(corpusRoot, fmt.Sprintf("v%d", game.SnapshotSchemaVersion))
	fresh, drift := corpusFixtureDrift(t, dir, renderCorpusBoards(t))
	if len(drift) > 0 {
		t.Errorf("%s:\n\n  %s\n\n%s", dir, strings.Join(drift, "\n  "), corpusDriftAdvice())
	}
	if len(fresh) > 0 {
		names := make([]string, 0, len(fresh))
		for name := range fresh {
			names = append(names, name)
		}
		sort.Strings(names)
		t.Errorf(`these boards are registered in corpusBoards but %s has no fixture for them: %s

Write them in the same change that registers them (it only adds files):

  go test ./internal/cards/effects -run TestWriteSnapshotCorpus -args -write-corpus`,
			dir, strings.Join(names, ", "))
	}
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
