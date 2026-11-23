package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch02_second_pass_test.go — the eleven cards of roadmap batch 02
// (#295, `edhrec_rank` 238–360) that were NOT covered by the first
// pass in batch02_test.go / #351.
//
// Two sessions worked #295 in parallel without knowing it. #351
// landed the 25-card "no new machinery" group; this file covers the
// eleven that group did not reach — nine of them because the batch
// triage had filed them under a blocker that had already moved, and
// two (Entomb, Buried Alive) because they needed a one-case engine
// fix to `searchDestZoneLocked`, which ships alongside.
//
// Naming is `b02b`-prefixed to stay clear of batch02_test.go's `b02`
// helpers, which this file deliberately does not touch.

const (
	B02bKetriaTriomeOracle      = "6bae00e8-06cf-4ac4-a1cc-757e454109fe"
	B02bDarksteelCitadelOracle  = "8dc067bf-f78f-4ac4-b6e7-b305c42cf0bc"
	B02bEntombOracle            = "299fc083-0834-4064-8344-f895aff68867"
	B02bBuriedAliveOracle       = "8203c621-a1a0-4865-8c9a-0d4064c86107"
	B02bSeethingSongOracle      = "64bf8929-f5f2-4d50-8667-13b1d007bcfc"
	B02bManaGeyserOracle        = "a8dba58b-2956-492e-ae30-49db2ae68e53"
	B02bHarrowOracle            = "705509e9-a034-4a5a-9c65-66f58748b8a2"
	B02bDiabolicIntentOracle    = "038519b9-bca8-4b27-b5ac-2409595469d0"
	B02bGrayMerchantOracle      = "38f3b157-0df4-409b-89cc-086e1531cd5b"
	B02bCraterhoofOracle        = "8c52bd39-0586-48ca-b263-17210cf9feb6"
	B02bDecanterOfEndlessWaterO = "8ae98ef8-8f52-4877-a08c-1fae5514184e"
)

// --- local helpers -------------------------------------------------

func B02bGraveyardHasNamed(p *game.Player, name string) bool {
	for _, c := range p.Graveyard.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func B02bHandHasNamed(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

func B02bSetTapped(g *game.Game, id uuid.UUID, tapped bool) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = tapped
				return
			}
		}
	})
}
