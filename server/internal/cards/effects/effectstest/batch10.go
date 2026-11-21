package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch10_test.go — card-level coverage for the card-coverage
// roadmap's batch 10 (#303, `edhrec_rank` 1111–1213): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b10-prefixed.

const (
	B10MaraudingBlightPriestOracle = "814b87fe-2a75-4ff2-8637-7e69e3fb285b"
	B10DrakusethOracle             = "060deaff-44d6-4f03-9568-bcb7add80255"
	B10KalonianHydraOracle         = "7bd36106-04fe-481f-b16e-e076dcbb183b"
	B10MarbleDiamondOracle         = "910488bf-66ab-415e-973b-1262b2ab7454"
	B10FirebrandArcherOracle       = "2e9289d6-dbc6-456d-88cf-d1f534e731d6"
	B10LumraOracle                 = "97a84e9d-bfc4-4ca2-b1e8-908dba56ccdb"
	B10SecretRendezvousOracle      = "34b90d3e-f48c-41ff-b3e4-5cab7a9cd597"
	B10WaterloggedGroveOracle      = "70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa"
	B10RadiantSummitOracle         = "5dd0cc44-4647-4857-ad3b-22494099d08a"
	B10MagusOfTheWheelOracle       = "df13dafb-f82e-42ac-b6dc-efe82af5db58"
	B10RecruiterOfTheGuardOracle   = "d521a329-a53a-4962-810a-2abed80df260"
	B10SheoldredsEdictOracle       = "217062f5-96f1-454c-9507-17f34ef37070"
	B10TitaniaOracle               = "d0ade00d-a496-441d-9b7e-7dc033d3292c"
	B10SoulShatterOracle           = "615927c2-3fb0-4e64-a1a7-55fb56de1423"
	B10VindicateOracle             = "63c1ac21-e3d8-40c2-8c09-3f31c52992ef"
	B10KedissOracle                = "d9c87cc2-943e-49b6-becc-748857549617"
	B10TheShireOracle              = "9abf9a0e-8e7d-406b-a01d-d4870b30134e"
	B10MikokoroOracle              = "a4580a1d-141e-449b-9018-e0258130634b"
	B10DamningVerdictOracle        = "355ed7ef-bfa6-4538-99fb-a2d203eb7005"
	B10CityOfTraitorsOracle        = "f161111d-9747-47b3-bb10-3c8bded32e21"
	B10EssenceWardenOracle         = "6ca2a89e-7032-4864-b4e9-66f3178f90ab"
	B10SuturePriestOracle          = "c4d36522-3ace-4bfb-bf2d-1a366f458698"
	B10GoblinEngineerOracle        = "c1d6cce8-085f-42cb-8b0c-b6fbbf88b16a"
	B10SirenStormtamerOracle       = "59c7496b-19a2-478b-b28f-6f153c2458ae"
	B10CrackleWithPowerOracle      = "273f5483-b67e-4dd6-bba8-c0a047fa34d7"
	B10NotionThiefOracle           = "f8dab16e-1d50-443e-9431-8b6f1cf61c9c"
	B10EzurisPredationOracle       = "d0dd425b-fdba-41b4-b9e6-f5161610bd7e"
	B10DragonsHoardOracle          = "8cf77dc4-763b-41b7-a5da-0ef2734f08e6"
	B10ArchonOfSunsGraceOracle     = "48a99dce-0aa9-4aac-81df-cec5f94c639d"
	B10PatriarsSealOracle          = "ad95eac2-5b48-4102-9988-a6d7b1ec30e0"
	B10MarwynOracle                = "ee35de1c-aef1-4bd4-85fd-fe77bc927790"
	B10MountDoomOracle             = "995c8dac-fd27-468a-abd4-02372cf0c850"
	B10WastelandOracle             = "09a70ae8-3859-4a09-901d-dce063fa3b5f"
	B10FireLitThicketOracle        = "d99a1d9a-7721-4331-bf22-1c6ee0bd825a"
	B10PrimalVigorOracle           = "c665544f-557b-4631-a1dc-39571470ca2e"

	B10RampantGrowthOracle = "8539f295-5d58-4436-a73a-b9277c4c7795"
)

// b10Awake clears summoning sickness on a creature that was cast
// this turn, so a test can attack or tap with it in the same turn.
func B10Awake(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

// b10Creature pushes a live creature with the given type line, P/T
// and (optionally) mana cost, with a real entry timestamp and no
// summoning sickness.
func B10Creature(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b10LibraryTop pushes a card onto the TOP of a player's library — the
// next card drawn or milled.
func B10LibraryTop(p *game.Player, name, typeLine, manaCost string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// b10AddMana floats the named mana in a player's pool.
func B10AddMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

// b10ResolveAllManaPicks answers every open colour pick for a player
// with `color`, returning how many were answered.
func B10ResolveAllManaPicks(t *testing.T, g *game.Game, chooser uuid.UUID, color string) int {
	t.Helper()
	// #3939: a filter land's picks are one choice asked once per pick,
	// so drain until none is open rather than walking a snapshot.
	n := 0
	for c := RiderLatestManaPick(g, chooser); c != nil && n < 8; c = RiderLatestManaPick(g, chooser) {
		if err := g.ResolveManaChoice(c.ID, chooser, color); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
		n++
	}
	return n
}

// b10OpponentCastsBolt puts a Lightning Bolt in an opponent's hand and
// casts it at the given target, returning the spell's ID.
func B10OpponentCastsBolt(t *testing.T, g *game.Game, opp *game.Player, target game.TargetRef) uuid.UUID {
	t.Helper()
	return Batch01OpponentCasts(t, g, opp, "Lightning Bolt", LightningBoltOracle, "{R}", []game.TargetRef{target})
}
