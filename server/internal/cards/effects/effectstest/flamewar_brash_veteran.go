package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsfl "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/fl"
)

func FlamewarRow() cards.Card {
	return TransformerRow(cardsfl.FlamewarOracleID,
		DisturbFace{"Flamewar, Brash Veteran", "Legendary Artifact Creature — Robot", "{1}{B}{R}", "3", "2"},
		DisturbFace{"Flamewar, Streetwise Operative", "Legendary Artifact — Vehicle", "", "2", "1"},
		"More Than Meets the Eye {B}{R}\nSacrifice another artifact: Put a +1/+1 counter on Flamewar and convert it. Activate only as a sorcery.\n{1}, Discard your hand: Put all exiled cards you own with intel counters on them into your hand.",
		"Living metal\nMenace, deathtouch\nWhenever Flamewar deals combat damage to a player, exile that many cards from the top of your library face down. Put an intel counter on each of them. Convert Flamewar.",
		"Living metal", "Convert", "Menace", "Deathtouch")
}
