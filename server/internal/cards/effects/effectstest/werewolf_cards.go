package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// werewolf_cards_test.go — the daybound // nightbound werewolves in the
// catalogue (#2561, ADR 0132), on real imported cards.
//
// They go through deck.ToGameCard for transform_cards_test.go's reason,
// sharper here: daybound and nightbound reach a face only through the
// importer's per-face keyword stamping (the dump's `keywords` array
// narrowed by each face's own text), and a fixture built by hand would
// skip exactly the path a player's deck takes.

const (
	fearfulVillagerOracle   = "5fd09dbc-8bcd-4fe0-91b5-b00e721fa7eb"
	infestationExpertOracle = "bf69dc2a-9aec-4181-bbe9-70875055ec03"
	VillageWatchOracle      = "b2ecaae4-41ee-4c61-b5ce-db4364b307fc"
)

// werewolfRow builds a Scryfall `transform` record the way the bulk dump
// has every werewolf: the top-level `keywords` array names both
// keywords, and each face's oracle text has its own keyword lines.
func WerewolfRow(oracleID, front, frontType, frontCost, frontText, back, backType, backText string, fp, ft, bp, bt string, colors []string, extraKeywords ...string) cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     front + " // " + back,
		Layout:   "transform",
		TypeLine: frontType + " // " + backType,
		OracleID: uuid.MustParse(oracleID),
		Keywords: append([]string{"Daybound", "Nightbound"}, extraKeywords...),
		CardFaces: []cards.CardFace{
			{Name: front, TypeLine: frontType, ManaCost: frontCost, Colors: colors, Power: fp, Toughness: ft, OracleText: frontText},
			{Name: back, TypeLine: backType, Colors: colors, Power: bp, Toughness: bt, OracleText: backText},
		},
	}
}

func FearfulVillagerRow() cards.Card {
	return WerewolfRow(fearfulVillagerOracle,
		"Fearful Villager", "Creature — Human Werewolf", "{2}{R}",
		"Menace (This creature can't be blocked except by two or more creatures.)\nDaybound (If a player casts no spells during their own turn, it becomes night next turn.)",
		"Fearsome Werewolf", "Creature — Werewolf",
		"Menace (This creature can't be blocked except by two or more creatures.)\nNightbound (If a player casts at least two spells during their own turn, it becomes day next turn.)",
		"2", "3", "4", "3", []string{"R"}, "Menace")
}

func InfestationExpertRow() cards.Card {
	return WerewolfRow(infestationExpertOracle,
		"Infestation Expert", "Creature — Human Werewolf", "{4}{G}",
		"Whenever this creature enters or attacks, create a 1/1 green Insect creature token.\nDaybound (If a player casts no spells during their own turn, it becomes night next turn.)",
		"Infested Werewolf", "Creature — Werewolf",
		"Whenever this creature enters or attacks, create two 1/1 green Insect creature tokens.\nNightbound (If a player casts at least two spells during their own turn, it becomes day next turn.)",
		"3", "4", "4", "5", []string{"G"})
}
