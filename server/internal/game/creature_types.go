package game

// creature_types.go is the S26 creature-type vocabulary: the CR 205.3m
// list, the membership test, and the "do these two creatures share a
// type" question Coat of Arms is built out of.
//
// Why a list at all. Two rules need to enumerate creature types rather
// than test one: changeling (CR 702.73a — "this card is every creature
// type"), and "shares a creature type with", which has to know that
// Forest on Dryad Arbor is a LAND type and does not count. Neither can
// be answered from a type line alone.
//
// The list is generated from the Scryfall bulk dump rather than typed
// out: every subtype appearing after the em-dash on a card whose left
// half says Creature, minus every subtype that also appears on a
// non-creature line (which strips the basic land types off Dryad
// Arbor, Equipment off a Kindred Artifact, and so on), plus the
// hand-classified handful that appear ONLY on Kindred lines — Elf and
// Goblin are creature types even though the only place the dump shows
// them outside a creature is "Kindred Instant — Shapeshifter"-style
// text.
//
// Accuracy in the two directions is not symmetric, which is why the
// derivation errs the way it does. A type wrongly IN the list makes a
// changeling that type too — harmless, since nothing keys off a type
// no card prints. A type wrongly OUT makes a changeling NOT that type,
// which is a rules error a player would see. So the generator prefers
// to over-include, and the funny sets are excluded because their
// subtypes are jokes rather than because including them would break
// anything.
//
// Refreshing it after a new set is a manual step; there is no
// generator checked in, because the list has grown by a handful of
// entries a year and a stale entry is invisible until someone plays a
// card from the set that added it.

// AllCreatureTypes is every creature type the engine knows (CR
// 205.3m), sorted. Read-only: callers must not append to it or sort
// it. Ranged over by the changeling paths and by any effect that
// needs to enumerate types.
var AllCreatureTypes = []string{
	"Advisor", "Aetherborn", "Alien", "Ally", "Andorian", "Angel",
	"Antelope", "Ape", "Archer", "Archon", "Armadillo", "Armored",
	"Army", "Artificer", "Assassin", "Assembly-Worker", "Astartes",
	"Athlete", "Atog", "Aurochs", "Automaton", "Avatar", "Azra",
	"Badger", "Balloon", "Barbarian", "Bard", "Basilisk", "Bat",
	"Bear", "Beast", "Beaver", "Beeble", "Beholder", "Berserker",
	"Bird", "Bison", "Boar", "Borg", "Brainiac", "Bringer",
	"Brushwagg", "C'tan", "Caitian", "Camel", "Capybara", "Carrier",
	"Cat", "Centaur", "Champion", "Chimera", "Citizen", "Cleric",
	"Clown", "Cockatrice", "Construct", "Cow", "Coward", "Coyote",
	"Crab", "Crocodile", "Custodes", "Cyberman", "Cyborg",
	"Cyclops", "Dalek", "Dauthi", "Demigod", "Demon", "Designer",
	"Detective", "Devil", "Dinosaur", "Djinn", "Doctor", "Dog",
	"Dragon", "Drake", "Dreadnought", "Drix", "Drone", "Druid",
	"Dryad", "Dwarf", "Echidna", "Efreet", "Egg", "Elder",
	"Eldrazi", "Elemental", "Elephant", "Elf", "Elk", "Employee",
	"Eternal", "Eye", "Faerie", "Ferret", "Fish", "Flagbearer",
	"Fox", "Fractal", "Frog", "Fungus", "Gamer", "Gamma",
	"Gargoyle", "Germ", "Giant", "Giraffe", "Gith", "Glimmer",
	"Gnoll", "Gnome", "Goat", "Goblin", "God", "Golem", "Gorgon",
	"Gorn", "Graveborn", "Gremlin", "Griffin", "Hag", "Halfling",
	"Hamster", "Harpy", "Hedgehog", "Hellion", "Hero", "Hippo",
	"Hippogriff", "Homarid", "Homunculus", "Horror", "Horse",
	"Human", "Hydra", "Hyena", "Illusion", "Imp", "Incarnation",
	"Inhuman", "Inkling", "Inquisitor", "Insect", "Jackal",
	"Jellyfish", "Juggernaut", "Kangaroo", "Kavu", "Kelpien",
	"Kirin", "Kithkin", "Klingon", "Knight", "Kobold", "Kor",
	"Kraken", "Kree", "Lamia", "Lammasu", "Lanthanite", "Leech",
	"Lemur", "Leviathan", "Lhurgoyf", "Licid", "Lizard", "Lobster",
	"Lord", "Manticore", "Masticore", "Mercenary", "Merfolk",
	"Metathran", "Mime", "Minion", "Minotaur", "Mite", "Mole",
	"Monger", "Mongoose", "Monk", "Monkey", "Moogle", "Moonfolk",
	"Mount", "Mouse", "Mutant", "Myr", "Mystic", "Naga", "Nautilus",
	"Necron", "Nephilim", "Nightmare", "Nightstalker", "Ninja",
	"Noble", "Noggle", "Nomad", "Nymph", "Octopus", "Officer",
	"Ogre", "Ooze", "Orc", "Orgg", "Orion", "Otter", "Ouphe", "Ox",
	"Oyster", "Pangolin", "Peasant", "Pegasus", "Pentavite",
	"Performer", "Pest", "Phelddagrif", "Phoenix", "Phyrexian",
	"Pilot", "Pirate", "Plant", "Platypus", "Pony", "Porcupine",
	"Possum", "Praetor", "Primarch", "Processor", "Q", "Qu",
	"Rabbit", "Raccoon", "Ranger", "Rat", "Rebel", "Reflection",
	"Rhino", "Rigger", "Robot", "Rogue", "Rukh", "Sable",
	"Salamander", "Samurai", "Sand", "Saproling", "Satyr",
	"Scarecrow", "Scientist", "Scion", "Scorpion", "Scout",
	"Sculpture", "Seal", "Serf", "Serpent", "Servo", "Shade",
	"Shaman", "Shapeshifter", "Shark", "Sheep", "Shi'ar", "Siren",
	"Skeleton", "Skrull", "Skunk", "Slith", "Sliver", "Sloth",
	"Slug", "Snail", "Snake", "Soldier", "Soltari", "Sorcerer",
	"Spawn", "Specter", "Spellshaper", "Sphinx", "Spider", "Spike",
	"Spirit", "Sponge", "Spuzzem", "Spy", "Squid", "Squirrel",
	"Starfish", "Surrakar", "Survivor", "Symbiote", "Synth",
	"Talosian", "Teddy", "Tellarite", "Tentacle", "Thalakos",
	"Tholian", "Thopter", "Thrull", "Tiefling", "Time", "Tosk",
	"Toy", "Treefolk", "Trilobite", "Triskelavite", "Troll",
	"Turtle", "Tyranid", "Unicorn", "Urzan", "Utrom", "Vampire",
	"Varmint", "Vedalken", "Villain", "Volver", "Vorta", "Vulcan",
	"Wall", "Walrus", "Warlock", "Warrior", "Weasel", "Weird",
	"Werewolf", "Whale", "Wizard", "Wolf", "Wolverine", "Wombat",
	"Worm", "Wraith", "Wurm", "Xindi", "Yeti", "Zombie", "Zubera",
}

// creatureTypeSet maps a creature type's lowercase form to its
// canonical spelling, so IsCreatureType is a map hit rather than a
// 345-entry scan and CanonicalCreatureType can normalise a client's
// input. Built once at package init.
var creatureTypeSet = func() map[string]string {
	out := make(map[string]string, len(AllCreatureTypes))
	for _, t := range AllCreatureTypes {
		out[lowerASCII(t)] = t
	}
	return out
}()

// IsCreatureType reports whether `s` names a creature type,
// case-insensitively. False for land types ("Forest"), artifact types
// ("Equipment"), enchantment types ("Aura") and for anything the
// vocabulary has not heard of.
func IsCreatureType(s string) bool {
	_, ok := CanonicalCreatureType(s)
	return ok
}

// CanonicalCreatureType normalises a creature type to the spelling in
// AllCreatureTypes, reporting whether it is one at all.
//
// Normalising rather than accepting the caller's string is what keeps
// Card.NamedTribe comparable: the type is chosen once through a wire
// action and then compared against subtype lists on every layer
// recompute, and "elf" vs "Elf" has to not be the difference between
// a lord pumping and not. The comparison itself is still
// case-insensitive (typeListHas folds), so this is belt and braces —
// but it is also what the prompt echoes back on the wire, and a
// lower-case tribe name in the UI reads as a bug.
func CanonicalCreatureType(s string) (string, bool) {
	if s == "" {
		return "", false
	}
	t, ok := creatureTypeSet[lowerASCII(s)]
	return t, ok
}

// lowerASCII lowercases an ASCII string, allocating only when the
// input actually has an upper-case byte. Type names come out of
// Scryfall in title case, so the allocation is taken once per
// distinct lookup key and never on the hot membership path for an
// already-lowercase needle.
func lowerASCII(s string) string {
	needs := false
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			needs = true
			break
		}
	}
	if !needs {
		return s
	}
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}

// HasAllCreatureTypes reports whether the card is every creature type
// right now (CR 702.73a).
//
// The engine carries "is every creature type" as the `changeling`
// keyword in the ability list — for printed changeling because that is
// literally what the card says, and for a GRANT of the same property
// (Maskwood Nexus) because expressing it as ~345 entries appended to
// Characteristic.Subtypes would make the wire type line unreadable and
// every subtype loop in the engine quadratic for no gain.
//
// The consequence worth knowing: this is true in EVERY zone for a
// printed changeling, because HasKeyword falls back to the card's own
// Keywords off the battlefield, and only on the battlefield for a
// granted one, because CR 113.6 says a static ability's continuous
// effect applies only while its source is on the battlefield. Both are
// the correct answer, for different reasons.
func HasAllCreatureTypes(c *Card) bool {
	return HasKeyword(c, KeywordChangeling)
}

// CreatureTypesOf returns the creature types `c` currently has, in
// the order they appear on the card, filtered to the CR 205.3m
// vocabulary — so Dryad Arbor answers ["Dryad"] and not ["Forest",
// "Dryad"].
//
// A changeling returns AllCreatureTypes, which is a large slice; the
// callers that only need "do these overlap" should use
// SharesCreatureType, which answers without materialising either
// side.
//
// Reads effective subtypes, so a Layer-4 type grant is included.
func CreatureTypesOf(c *Card) []string {
	if c == nil {
		return nil
	}
	if HasAllCreatureTypes(c) {
		return AllCreatureTypes
	}
	subs := c.Effective().Subtypes
	out := make([]string, 0, len(subs))
	for _, s := range subs {
		if IsCreatureType(s) {
			out = append(out, s)
		}
	}
	return out
}

// SharesCreatureType reports whether `a` and `b` have at least one
// creature type in common (CR 702.73a-aware). This is Coat of Arms'
// whole rule and Path of Ancestry's, and it is deliberately NOT a
// subtype-slice intersection:
//
//   - Land and artifact types don't count. Two Dryad Arbors share
//     Dryad, not Forest, and the answer happens to be the same; a
//     Forest that a Kormus Bell animated and a Dryad Arbor share
//     nothing, and a raw intersection would say they share Forest.
//   - A changeling shares a type with every creature that has ANY
//     creature type — including another changeling, and NOT
//     including a creature printed with no subtype at all (there is
//     no type for them to share).
//
// nil on either side is false.
func SharesCreatureType(a, b *Card) bool {
	if a == nil || b == nil {
		return false
	}
	aAll, bAll := HasAllCreatureTypes(a), HasAllCreatureTypes(b)
	if aAll && bAll {
		// Both are every type, and the vocabulary is non-empty, so
		// they share all of it.
		return true
	}
	if aAll {
		return hasAnyCreatureType(b)
	}
	if bAll {
		return hasAnyCreatureType(a)
	}
	for _, t := range a.Effective().Subtypes {
		if !IsCreatureType(t) {
			continue
		}
		if typeListHas(b.Effective().Subtypes, t) {
			return true
		}
	}
	return false
}

// hasAnyCreatureType reports whether the card has at least one
// creature type. Separate from len(CreatureTypesOf(c)) > 0 so the
// changeling case doesn't allocate a copy of the whole vocabulary to
// answer a yes/no question.
func hasAnyCreatureType(c *Card) bool {
	if HasAllCreatureTypes(c) {
		return true
	}
	for _, t := range c.Effective().Subtypes {
		if IsCreatureType(t) {
			return true
		}
	}
	return false
}
