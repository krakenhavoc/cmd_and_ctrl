package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments_batch3_test.go covers the third S24 attachment batch —
// thirteen Equipment and ten Auras on the surface #374 / #379 / #511
// built.
//
// The plain cards are not retested here. Bonesplitter already pins
// "equip attaches and the 7c static reaches the host", and Loxodon
// Warhammer pins "a grant and a pump move together", so Sword of
// Vengeance and Behemoth Sledge would only re-prove those. What IS
// here is everything this batch does that no earlier card did:
//
//   - a static whose bonus is COUNTED from the board on every
//     recompute rather than fixed (Blackblade, All That Glitters,
//     Ancestral Mask, Blanchwood Armor)
//   - a grant with a CONDITION on the host (Champion's Helm)
//   - an attach ability that is NOT an equip ability, because it has
//     no sorcery-speed gate (Cranial Plating)
//   - a second equip ability with a tighter target clause (Blackblade)
//   - "whenever equipped creature attacks" (Sword of the Animist,
//     Argentum Armor) and a targeted version of it
//   - a damage condition wider than the Swords' (Curiosity, Spirit
//     Link)
//   - ward granted to the HOST rather than printed on the source
//     (Lavaspur Boots)

const (
	SwordOfVengeanceOracle  = "e366afb3-c447-4bde-b358-41c8568142d5"
	ChampionsHelmOracle     = "c01aafb8-3da5-4eb1-8731-2a223e747d63"
	BlackbladeOracle        = "dca51281-fb21-45b6-beb4-1f13397caee2"
	CranialPlatingOracle    = "75564721-e4e9-463a-b49e-f0c7cd6f53a7"
	SwordOfTheAnimistOracle = "d79cbc61-6c15-48ea-bbba-3cffb819ccba"
	ArgentumArmorOracle     = "d0a1f39b-cda4-4925-83f8-2161f575edfb"
	QuietusSpikeOracle      = "0ff0a309-9b4a-4f9e-9ea6-c4d07906ebec"
	GoldveinPickOracle      = "c1624d10-8838-4af8-aea1-a96c0fe6fd6b"
	LavaspurBootsOracle     = "63bfb9ca-d9d5-4c17-be39-82eb115bc20c"
	AllThatGlittersOracle   = "a4d751e0-41c1-4e90-853d-512f385acd81"
	AncestralMaskOracle     = "db5380ed-ba28-4ea2-abc3-4998e2022903"
	BlanchwoodArmorOracle   = "80ea56ad-e741-4a85-b4e8-ce62e7d593d5"
	CuriosityOracle         = "223fa044-d387-4884-bf4e-75f1b61c6a46"
	SpiritLinkOracle        = "c77ff526-c0a8-45c7-9730-2e306a0d01b8"
	ShadowspearOracle       = "8b27326f-e7b8-4a4d-b589-df459246d19a"
	AngelicGiftOracle       = "e5e04968-d9b7-4bd5-b826-be9502360cd3"
)

// seedLegendaryCreature is a 2/2 with the Legendary supertype, which
// is the half Champion's Helm and Blackblade Reforged both read.
func SeedLegendaryCreature(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Human Soldier",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// seedPermanent puts an arbitrary non-creature permanent on the
// battlefield — the things the counted pumps count.
func SeedPermanent(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner,
	})
}

func CountTreasuresControlledBy(g *game.Game, controller uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == "Treasure" {
				n++
			}
		}
	})
	return n
}

func EmitNoncombatDamage(g *game.Game, source, victim uuid.UUID, amount int) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind:   game.EventDealDamage,
			Source: source,
			Target: victim,
			Amount: amount,
		})
	})
}
