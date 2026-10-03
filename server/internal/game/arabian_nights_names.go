package game

import "strings"

// arabian_nights_names.go — CR 206.3a's list of names originally printed
// in the Arabian Nights expansion, which exactly one card refers to: City
// in a Bottle ("Players can't cast spells or play lands with a name
// originally printed in the Arabian Nights expansion", and its first
// sentence about permanents). ADR 0109 §4 (#1895).
//
// The list is the rule's own, copied from the pinned Comprehensive Rules
// (September 25, 2026 edition). A name is matched on the card's printed
// name as the face being cast or played shows it, folded for case, the
// typographic apostrophe and the accents Scryfall and the rules disagree
// about (the rules print "Ifh-Biff Efreet", Scryfall "Ifh-Bíff Efreet").

// ArabianNightsNames is CR 206.3a's list, in the rule's order.
var ArabianNightsNames = []string{
	"Abu Ja'far", "Aladdin", "Aladdin's Lamp", "Aladdin's Ring",
	"Ali Baba", "Ali from Cairo", "Army of Allah", "Bazaar of Baghdad",
	"Bird Maiden", "Bottle of Suleiman", "Brass Man", "Camel",
	"City in a Bottle", "City of Brass", "Cuombajj Witches", "Cyclone",
	"Dancing Scimitar", "Dandân", "Desert", "Desert Nomads",
	"Desert Twister", "Diamond Valley", "Drop of Honey", "Ebony Horse",
	"Elephant Graveyard", "El-Hajjâj", "Erg Raiders", "Erhnam Djinn",
	"Eye for an Eye", "Fishliver Oil", "Flying Carpet", "Flying Men",
	"Ghazbán Ogre", "Giant Tortoise", "Guardian Beast", "Hasran Ogress",
	"Hurr Jackal", "Ifh-Biff Efreet", "Island Fish Jasconius",
	"Island of Wak-Wak", "Jandor's Ring", "Jandor's Saddlebags",
	"Jeweled Bird", "Jihad", "Junún Efreet", "Juzám Djinn",
	"Khabál Ghoul", "King Suleiman", "Kird Ape", "Library of Alexandria",
	"Magnetic Mountain", "Merchant Ship", "Metamorphosis", "Mijae Djinn",
	"Moorish Cavalry", "Nafs Asp", "Oasis", "Old Man of the Sea",
	"Oubliette", "Piety", "Pyramids", "Repentant Blacksmith",
	"Ring of Ma'rûf", "Rukh Egg", "Sandals of Abdallah", "Sandstorm",
	"Serendib Djinn", "Serendib Efreet", "Shahrazad", "Sindbad",
	"Singing Tree", "Sorceress Queen", "Stone-Throwing Devils",
	"Unstable Mutation", "War Elephant", "Wyluli Wolf", "Ydwen Efreet",
}

var arabianNightsFolded = func() map[string]bool {
	m := make(map[string]bool, len(ArabianNightsNames))
	for _, n := range ArabianNightsNames {
		m[foldCardName(n)] = true
	}
	return m
}()

// IsArabianNightsName reports whether `name` is a name originally
// printed in Arabian Nights (CR 206.3a).
func IsArabianNightsName(name string) bool {
	return arabianNightsFolded[foldCardName(name)]
}

// foldCardName lower-cases a card name and folds the typographic
// apostrophe and the accented letters CR 206.3a's names use, so the
// rules' spelling and a card database's spelling compare equal.
func foldCardName(name string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '’':
			return '\''
		case 'á', 'â', 'Á', 'Â':
			return 'a'
		case 'í', 'î', 'Í', 'Î':
			return 'i'
		case 'ú', 'û', 'Ú', 'Û':
			return 'u'
		case 'é', 'è', 'É', 'È':
			return 'e'
		}
		return r
	}, strings.ToLower(name))
}
