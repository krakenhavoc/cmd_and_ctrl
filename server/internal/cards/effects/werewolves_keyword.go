package effects

// werewolves_keyword.go — the Midnight Hunt / Crimson Vow werewolves
// whose two faces print nothing but combat keywords beside daybound and
// nightbound (#2561, ADR 0132; three more in #2586). All of them are a
// table, for the reason temples.go is: near-identical files are places
// to get one face wrong.
//
//	Fearful Villager // Fearsome Werewolf              menace, menace
//	Harvesttide Infiltrator // Harvesttide Assailant   trample, trample
//	Bird Admirer // Wing Shredder                      reach, reach
//	Shady Traveler // Stalking Predator                menace, menace
//	Tireless Hauler // Dire-Strain Brawler             vigilance, vigilance
//	Tavern Ruffian // Tavern Smasher                   (nothing)
//
// The engine does all of it. Daybound and nightbound are canonical
// keywords the deck importer stamps off each face's own text, and the
// day/night rules (game/daynight.go) turn the permanent over: night
// turns the front over, day turns the back over, a daybound card cast at
// night enters on its back face, and nothing else can transform it. A
// card file exists so the catalogue can say the whole card is automated
// and the combat keywords ride PrintedKeywords like every other
// creature's.
//
// No simplification.
type keywordWerewolf struct {
	oracleID string
	front    string
	back     string
	keyword  string
}

var keywordWerewolves = []keywordWerewolf{
	{"5fd09dbc-8bcd-4fe0-91b5-b00e721fa7eb", "Fearful Villager", "Fearsome Werewolf", "menace"},
	{"8669f2e1-3e98-4fa5-ba4f-a0860b92c609", "Harvesttide Infiltrator", "Harvesttide Assailant", "trample"},
	{"58bd02ae-2676-4c9c-b24e-2bd51be8bde7", "Bird Admirer", "Wing Shredder", "reach"},
	{"10be1b27-bc9f-4c6e-ac85-f1fa8b2a34d6", "Shady Traveler", "Stalking Predator", "menace"},
	{"c31e9db3-5d9d-470a-871a-b4b5b0536db5", "Tireless Hauler", "Dire-Strain Brawler", "vigilance"},
	{"73a3b9a1-37a0-469a-9557-8c118a1ee78f", "Tavern Ruffian", "Tavern Smasher", ""},
}

func init() {
	for _, w := range keywordWerewolves {
		// keyword is empty for a werewolf whose faces are vanilla (Tavern
		// Ruffian): the only printed keyword is then daybound / nightbound.
		with := func(bound string) []string {
			if w.keyword == "" {
				return []string{bound}
			}
			return []string{w.keyword, bound}
		}
		Register(Spec{
			OracleID:        w.oracleID,
			Name:            w.front,
			Completeness:    CompletenessFull,
			PrintedKeywords: with("daybound"),
		})
		Register(Spec{
			OracleID:        w.oracleID + "#1",
			Name:            w.back,
			Completeness:    CompletenessFull,
			PrintedKeywords: with("nightbound"),
		})
	}
}
