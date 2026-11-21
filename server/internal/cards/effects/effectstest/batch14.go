package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch14_test.go — card-level coverage for the card-coverage
// roadmap's batch 14 (#307, `edhrec_rank` 1525–1625): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b14-prefixed.

const (
	B14BloomingMarshOracle        = "66fa2326-1b5d-41fb-b919-83bf9f383577"
	B14HangarbackWalkerOracle     = "dde55256-5259-44e7-a267-fca45a7f0d04"
	B14UtvaraHellkiteOracle       = "05a6a571-643e-429e-8e5c-1c3f8b0dc746"
	B14EerieUltimatumOracle       = "5674f6ae-ed5d-441e-a534-b5dd415165fd"
	B14CreepingBloodsuckerOracle  = "0bdddaf9-579a-4588-b5b3-faa189d4bdcc"
	B14HydroidKrasisOracle        = "6bd872b2-5c40-4e11-9a7f-0136a51b0642"
	B14SimicGuildgateOracle       = "e8705df9-6439-4930-91b6-229f818559af"
	B14DragonmasterOutcastOracle  = "b6fb79c3-cd32-4045-8177-e52841eea65b"
	B14TreeOfTalesOracle          = "8b4aa971-b919-4750-8388-33d4f42c9280"
	B14RebuffTheWickedOracle      = "f5822e53-ddec-4c77-bcf4-091e35c1e731"
	B14GanaxOracle                = "a3112971-3af4-46e8-b2d6-a8759b39d0d1"
	B14CorpsejackMenaceOracle     = "ca0cc02b-b106-4eca-9388-d4b48dd3be49"
	B14AdventurersInnOracle       = "232bd88c-ecdb-43dd-b34a-d381cb3bedf2"
	B14KozilekOracle              = "4c1c1537-e519-4e2f-9bc2-d34b289d4487"
	B14ReturnToDustOracle         = "3029df1d-d02a-4fed-8ab4-000a2096f823"
	B14SpirebluffCanalOracle      = "eb0d8093-5f93-4b25-9384-08f9731bfb28"
	B14DesynchronizationOracle    = "dab64d0f-1246-4a4d-9a79-6db2ca0c8882"
	B14ArchfiendOfDepravityOracle = "af247e2f-b271-4f5b-ab98-4579d2c17c21"
	B14RadiantFountainOracle      = "6db442e5-fbcc-4456-a4c5-bea1aee3fc8e"
	B14NullElementalBlastOracle   = "ec7699e3-ce56-46e4-9e4d-5ed2bd5fca83"
	B14KiorasFollowerOracle       = "22c044ad-77d7-4c93-953d-e2daa9686ff7"
	B14SiegeGangCommanderOracle   = "ddc7f59a-bbb1-4ba1-82c8-6813fd191940"
	B14ElectrostaticFieldOracle   = "2fa94a07-c932-4f85-b6e0-a97d2b29eb52"
	B14ScavengingOozeOracle       = "1ff25f67-36a7-4cfa-a2b1-2135b5b6fb67"
	B14NurturingPeatlandOracle    = "8ed932ff-986c-4592-ad70-53b3fac80d69"
	B14ResourcefulDefenseOracle   = "83e78565-e61f-4bbc-b834-f47941f7e3ec"
	B14ReplicatingRingOracle      = "1ff00f5b-4bf9-4724-8bb5-6b9a9eb0ec7f"
	B14DesertedTempleOracle       = "9f12bf9a-6e1a-4377-b4af-e8cabd3ee58a"
	B14ConsumingAberrationOracle  = "9b55fb72-237d-4935-b645-8ebc6eb4140e"
	B14GolgariGuildgateOracle     = "fa2da325-6859-45bb-b185-35526b01bcc1"
	B14UnmarkedGraveOracle        = "134dfb35-0c32-4143-aa63-7701f406b59e"
	B14CloudOracle                = "33d2584b-bf29-4c22-bd45-14ba2fb98c0e"
	B14RhoxFaithmenderOracle      = "2abe9303-d498-4aad-b6b2-8b5064bd2ffd"
	B14PawnOfUlamogAlreadyOID     = "9bcaf141-1f1f-491f-aced-13dc093b9e2c"
	B14CruxOfFateAlreadyOID       = "52a0dae4-2a95-487e-acd4-eabdb2d031e2"
	B14MazirekAlreadyOID          = "e0420f2c-d578-421e-ae75-e7dc5f70661a"
	B14CounterspellOracle         = "cc187110-1148-4090-bbb8-e205694a39f5"
)

// b14CastAt casts a card from `caster`'s hand at the current step,
// without walking the cursor — for an instant cast outside a main
// phase (Return to Dust's clause) or by a non-active player.
func B14CastAt(t *testing.T, g *game.Game, caster *game.Player, name, typeLine, oracle string, colors []string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Colors: colors,
		Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// b14Tokens counts the battlefield tokens with a name under a
// controller and returns the first one.
func B14Tokens(g *game.Game, controller uuid.UUID, name string) (int, game.Card) {
	n := 0
	var first game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller && effects.IsToken(c) {
			if n == 0 {
				first = c
			}
			n++
		}
	}
	return n, first
}

// b14Kill destroys a battlefield permanent through the engine's own
// destroy path, so its LKI and dies-triggers are real.
func B14Kill(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
}

// b14Gain changes a player's life through the public mutator, so a
// replacement watching life sees it.
func B14Gain(t *testing.T, g *game.Game, player uuid.UUID, delta int) {
	t.Helper()
	if _, err := g.ChangePlayerLife(player, delta); err != nil {
		t.Fatalf("ChangePlayerLife: %v", err)
	}
}

// b14HandCard puts a card into a player's hand and returns its ID.
func B14HandCard(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID,
	})
	return id
}
