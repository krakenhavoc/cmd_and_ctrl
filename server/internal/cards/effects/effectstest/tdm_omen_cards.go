package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsbl "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/bl"
	cardsdi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/di"
	cardsex "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ex"
	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	cardsri "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ri"
	cardss "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/s"
	cardsst "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/st"
	cardsto "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/to"
	cardsw "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/w"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_omen_cards_test.go — Tarkir: Dragonstorm's omen cards (#4463,
// #5175). Each test plays the real catalog entries and asserts what the
// printed text says. Every Omen half that resolves is checked for
// CR 720.3d: the card is in its owner's library, not in exile and not in
// a graveyard. The engine-level lifecycle is internal/game/omen_test.go.

// tdmOmenCard builds an omen card the way deck.toGameCard does: layout
// `adventure`, the Dragon as face 0 and the Omen as face 1, then
// SetFace(0).
func tdmOmenCard(oracle string, dragon, omen game.Face) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Layout:     game.LayoutAdventure,
		Faces:      []game.Face{dragon, omen},
	}
	c.SetFace(0)
	return c
}

func tdmDragon(name, cost string, power, toughness int) game.Face {
	return game.Face{Name: name, TypeLine: "Creature — Dragon", ManaCost: cost,
		Colors: game.ColorsInManaCost(cost), Power: power, Toughness: toughness}
}

func tdmOmen(name, typeLine, cost string) game.Face {
	return game.Face{Name: name, TypeLine: typeLine + " — Omen", ManaCost: cost,
		Colors: game.ColorsInManaCost(cost)}
}

var TdmOmenCards = map[string]game.Card{
	"Bloomvine Regent": tdmOmenCard(cardsbl.BloomvineRegentOracleID,
		tdmDragon("Bloomvine Regent", "{3}{G}{G}", 4, 5), tdmOmen("Claim Territory", "Sorcery", "{2}{G}")),
	"Dirgur Island Dragon": tdmOmenCard(cardsdi.DirgurIslandDragonOracleID,
		tdmDragon("Dirgur Island Dragon", "{5}{U}", 4, 4), tdmOmen("Skimming Strike", "Instant", "{1}{U}")),
	"Disruptive Stormbrood": tdmOmenCard(cardsdi.DisruptiveStormbroodOracleID,
		tdmDragon("Disruptive Stormbrood", "{4}{G}", 3, 3), tdmOmen("Petty Revenge", "Sorcery", "{1}{B}")),
	"Feral Deathgorger": tdmOmenCard(cardsex.FeralDeathgorgerOracleID,
		tdmDragon("Feral Deathgorger", "{5}{B}", 3, 5), tdmOmen("Dusk Sight", "Sorcery", "{1}{B}")),
	"Marang River Regent": tdmOmenCard(cardslu.MarangRiverRegentOracleID,
		tdmDragon("Marang River Regent", "{4}{U}{U}", 6, 7), tdmOmen("Coil and Catch", "Instant", "{3}{U}")),
	"Purging Stormbrood": tdmOmenCard(effects.PurgingStormbroodOracleID,
		tdmDragon("Purging Stormbrood", "{4}{B}", 4, 4), tdmOmen("Absorb Essence", "Instant", "{1}{W}")),
	"Riling Dawnbreaker": tdmOmenCard(cardsri.RilingDawnbreakerOracleID,
		tdmDragon("Riling Dawnbreaker", "{4}{W}", 3, 4), tdmOmen("Signaling Roar", "Sorcery", "{1}{W}")),
	"Runescale Stormbrood": tdmOmenCard(cardsri.RunescaleStormbroodOracleID,
		tdmDragon("Runescale Stormbrood", "{3}{R}", 2, 4), tdmOmen("Chilling Screech", "Instant", "{1}{U}")),
	"Sagu Wildling": tdmOmenCard(cardss.SaguWildlingOracleID,
		tdmDragon("Sagu Wildling", "{4}{G}", 3, 3), tdmOmen("Roost Seek", "Sorcery", "{G}")),
	"Scavenger Regent": tdmOmenCard(cardss.ScavengerRegentOracleID,
		tdmDragon("Scavenger Regent", "{3}{B}", 4, 4), tdmOmen("Exude Toxin", "Sorcery", "{X}{B}{B}")),
	"Stormshriek Feral": tdmOmenCard(cardsst.StormshriekFeralOracleID,
		tdmDragon("Stormshriek Feral", "{4}{R}", 3, 3), tdmOmen("Flush Out", "Sorcery", "{1}{R}")),
	"Twinmaw Stormbrood": tdmOmenCard(cardsto.TwinmawStormbroodOracleID,
		tdmDragon("Twinmaw Stormbrood", "{5}{W}", 5, 4), tdmOmen("Charring Bite", "Sorcery", "{1}{R}")),
	"Whirlwing Stormbrood": tdmOmenCard(cardsw.WhirlwingStormbroodOracleID,
		tdmDragon("Whirlwing Stormbrood", "{4}{U}", 4, 3), tdmOmen("Dynamic Soar", "Sorcery", "{2}{G}")),
}
