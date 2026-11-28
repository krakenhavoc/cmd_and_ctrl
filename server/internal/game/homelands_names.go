package game

// homelands_names.go — CR 206.3c's list of names originally printed in the
// Homelands expansion, which exactly one card refers to: Apocalypse Chime
// ("Destroy all nontoken permanents with a name originally printed in the
// Homelands expansion"). S82 OLD b827, #5635.
//
// The list is the rule's own, copied from the pinned Comprehensive Rules
// (September 25, 2026 edition) and matched like ArabianNightsNames.

// HomelandsNames is CR 206.3c's list, in the rule's order.
var HomelandsNames = []string{
	"Abbey Gargoyles", "Abbey Matron", "Aether Storm", "Aliban’s Tower",
	"Ambush", "Ambush Party", "Anaba Ancestor", "Anaba Bodyguard",
	"Anaba Shaman", "Anaba Spirit Crafter", "An-Havva Constable",
	"An-Havva Inn", "An-Havva Township", "An-Zerrin Ruins",
	"Apocalypse Chime", "Autumn Willow", "Aysen Abbey",
	"Aysen Bureaucrats", "Aysen Crusader", "Aysen Highway",
	"Baki’s Curse", "Baron Sengir", "Beast Walkers", "Black Carriage",
	"Broken Visage", "Carapace", "Castle Sengir", "Cemetery Gate",
	"Chain Stasis", "Chandler", "Clockwork Gnomes", "Clockwork Steed",
	"Clockwork Swarm", "Coral Reef", "Dark Maze", "Daughter of Autumn",
	"Death Speakers", "Didgeridoo", "Drudge Spell", "Dry Spell",
	"Dwarven Pony", "Dwarven Sea Clan", "Dwarven Trader", "Ebony Rhino",
	"Eron the Relentless", "Evaporate", "Faerie Noble",
	"Feast of the Unicorn", "Feroz’s Ban", "Folk of An-Havva", "Forget",
	"Funeral March", "Ghost Hounds", "Giant Albatross", "Giant Oyster",
	"Grandmother Sengir", "Greater Werewolf", "Hazduhr the Abbot",
	"Headstone", "Heart Wolf", "Hungry Mist", "Ihsan’s Shade",
	"Irini Sengir", "Ironclaw Curse", "Jinx", "Joven", "Joven’s Ferrets",
	"Joven’s Tools", "Koskun Falls", "Koskun Keep", "Labyrinth Minotaur",
	"Leaping Lizard", "Leeches", "Mammoth Harness", "Marjhan",
	"Memory Lapse", "Merchant Scroll", "Mesa Falcon", "Mystic Decree",
	"Narwhal", "Orcish Mine", "Primal Order", "Prophecy",
	"Rashka the Slayer", "Reef Pirates", "Renewal", "Retribution",
	"Reveka, Wizard Savant", "Root Spider", "Roots", "Roterothopter",
	"Rysorian Badger", "Samite Alchemist", "Sea Sprite", "Sea Troll",
	"Sengir Autocrat", "Sengir Bats", "Serra Aviary", "Serra Bestiary",
	"Serra Inquisitors", "Serra Paladin", "Serrated Arrows", "Shrink",
	"Soraya the Falconer", "Spectral Bears", "Timmerian Fiends",
	"Torture", "Trade Caravan", "Truce", "Veldrane of Sengir",
	"Wall of Kelp", "Willow Faerie", "Willow Priestess", "Winter Sky",
	"Wizards’ School",
}

var homelandsFolded = func() map[string]bool {
	m := make(map[string]bool, len(HomelandsNames))
	for _, n := range HomelandsNames {
		m[foldCardName(n)] = true
	}
	return m
}()

// IsHomelandsName reports whether `name` is a name originally printed in
// Homelands (CR 206.3c).
func IsHomelandsName(name string) bool {
	return homelandsFolded[foldCardName(name)]
}
