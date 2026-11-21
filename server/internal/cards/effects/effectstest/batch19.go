package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch19_test.go — card-level coverage for the card-coverage
// roadmap's batch 19 (#312, `edhrec_rank` 2032–2132): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play, attack or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b19-prefixed.

const (
	B19TrinketMageOracle            = "ffc95093-24a2-4616-a44a-24788e8df9c8"
	B19BirthingPodOracle            = "f8b9dd54-0837-47f4-ad14-7a0322d46d5f"
	B19CancelOracle                 = "7d00fb28-ea6c-49a9-b4af-ffb38860a9a7"
	B19SelesnyaGuildgateOracle      = "75b235d3-595a-4859-be45-9559d8445db5"
	B19ElvishReclaimerOracle        = "94f01534-7bb3-4b1e-8f77-e84156408686"
	B19DeathreapRitualOracle        = "b26e3596-5b28-4eb6-b3e2-03f63d8c6d49"
	B19GrimBackwoodsOracle          = "5effaa94-7f87-4485-8959-473d584c5034"
	B19WindingConstrictorOracle     = "c9404d7d-a026-4082-9fcb-1ab571a136b5"
	B19GenerousPlundererOracle      = "91c835d1-22ca-4c90-9ba6-c8e01bbc0347"
	B19AzoriusGuildgateOracle       = "ad1712d8-809f-410c-8b91-ffe6fb8a69a1"
	B19WeaverOfHarmonyOracle        = "494b31b2-27ef-4ca1-ac72-c2fcfc8a23a1"
	B19RuthlessTechnomancerOracle   = "4e58ad76-37c7-4531-b207-6890b39a2679"
	B19DrownInDreamsOracle          = "d3cec4b5-bc93-44a2-a29d-3478f0a5dac6"
	B19IllustriousWanderglyphOracle = "72adf091-4491-4afe-b893-b8d20ad38a86"
	B19LiesaOracle                  = "efcaadbe-24e3-4dfc-b08c-a910f003d427"
	B19ErodeOracle                  = "2e467fab-e808-44d3-99bf-e3621baeb7cb"
	B19PerplexingTestOracle         = "8682266c-2b0f-496e-b1c6-1fb338feac24"
	B19ZendikarsRoilOracle          = "a842cc2b-52eb-4dc5-86c6-6575c2ed913d"
	B19CracklingDoomOracle          = "ee81f37b-2a81-46d7-8d2f-9091123846c4"
	B19RecklessImpulseOracle        = "584cf0fd-112a-4ca6-9c0e-1f3228a7d325"
	B19DiregrafColossusOracle       = "fd62ad01-601f-4250-bc2a-8ef3982e45c4"
	B19WrennsResolveOracle          = "a7d8c99c-cd44-4bad-82e6-7c7cff9bd89f"
)

// b19CastXModal casts an X spell with a mode choice from the active
// seat's hand — castXSpell and castModal in one, for Drown in Dreams.
func B19CastXModal(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, x int, modes []int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets, XValue: x}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// b19HasTriggerPromptFor reports whether an optional-trigger prompt
// is waiting on `chooser`.
func B19HasTriggerPromptFor(g *game.Game, chooser uuid.UUID) bool {
	return OwedMays(g, chooser) > 0
}

// b19TriggerQueuedFrom reports whether a triggered ability from
// `source` is waiting — harvested into PendingTriggers, or already on
// the stack.
func B19TriggerQueuedFrom(g *game.Game, source uuid.UUID) bool {
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source {
			return true
		}
	}
	return TriggerOnStack(g, source) != nil
}
