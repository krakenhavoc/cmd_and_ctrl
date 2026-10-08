package lobby

import (
	"math/rand/v2"
	"strings"
	"sync"
	"unicode/utf8"
)

// The table-name generator (#2630): a blank name on POST /games, and
// GET /games/name-suggestion, get a Magic-flavoured mash-up from the
// hand-written lists below. Generic game vocabulary only: no card
// names, card text or art (AGENTS.md section 8). Keep every entry
// affectionate and PG.

// maxTableNameLen is the longest a table name may be (CreateWith trims
// to it).
const maxTableNameLen = 80

// maxPersonalNameLen is the longest display name, in runes, a personal
// template will use. A longer one skips the personal templates rather
// than being cut mid-word.
const maxPersonalNameLen = 32

var nameAdjectives = []string{
	"Grumpy", "Suspicious", "Tapped-Out", "Overextended", "Sleepy", "Reckless",
	"Questionable", "Unlikely", "Sneaky", "Cursed", "Glorious", "Mildly Haunted",
	"Overprepared", "Underdressed", "Chaotic", "Sentimental", "Unstoppable",
	"Wobbly", "Dramatic", "Shiny", "Greedy", "Budget", "Legendary", "Feral",
	"Polite", "Smug", "Hasty", "Timid", "Ambitious", "Mischievous", "Unlucky",
	"Fearless", "Enchanted", "Haunted", "Wandering", "Unfair", "Ancient",
	"Cheerful", "Reluctant", "Colourless", "Topdecked", "Unblockable",
	"Miscounted", "Vigilant",
}

// nameNouns and namePlurals are aligned by index: namePlurals[i] is the
// plural of nameNouns[i].
var nameNouns = []string{
	"Ornithopter", "Goblin", "Treasure", "Island", "Dragon", "Mulligan",
	"Commander", "Wizard", "Sphinx", "Elf", "Zombie", "Angel", "Golem",
	"Library", "Graveyard", "Planeswalker", "Mana Rock", "Spellbook",
	"Dungeon", "Familiar", "Artifact", "Sorcerer", "Hydra", "Merfolk",
	"Knight", "Vampire", "Shapeshifter", "Cleric", "Rogue", "Druid",
	"Beast", "Phoenix", "Kraken", "Wurm", "Squirrel", "Sliver", "Treefolk",
	"Gargoyle", "Dice Bag", "Sideboard", "Battlefield", "Stack",
}

var namePlurals = []string{
	"Ornithopters", "Goblins", "Treasures", "Islands", "Dragons", "Mulligans",
	"Commanders", "Wizards", "Sphinxes", "Elves", "Zombies", "Angels", "Golems",
	"Libraries", "Graveyards", "Planeswalkers", "Mana Rocks", "Spellbooks",
	"Dungeons", "Familiars", "Artifacts", "Sorcerers", "Hydras", "Merfolk",
	"Knights", "Vampires", "Shapeshifters", "Clerics", "Rogues", "Druids",
	"Beasts", "Phoenixes", "Krakens", "Wurms", "Squirrels", "Slivers", "Treefolk",
	"Gargoyles", "Dice Bags", "Sideboards", "Battlefields", "Stacks",
}

var nameGroups = []string{
	"Accord", "Tribunal", "Conspiracy", "Gambit", "Summit", "Council",
	"Coalition", "Alliance", "Committee", "Syndicate", "Pact", "Assembly",
}

var nameNumbers = []string{
	"Two", "Three", "Four", "Five", "Six", "Seven", "Nine", "Thirteen",
}

var nameTroubles = []string{
	"Behaving Badly", "Doing Their Best", "Out of Position", "Running Late",
	"Looking for Value", "Making Questionable Choices", "Taking a Day Off",
	"Counting Their Outs", "Minding Their Own Business",
}

// A template is a format with these markers, filled in by fill:
//
//	{A} an adjective        {N} a singular noun   {P} its plural
//	{G} a group word        {#} a number word     {T} a "trouble" phrase
//	{U} the creator's name (a personal template: skipped without one)
var nameTemplates = []string{
	"The {A} {N} {G}",
	"{#} Untapped {P}",
	"A {A} Pile of {P}",
	"The {A} {G}",
	"{P} {T}",
	"{A} {P} Anonymous",
	"The Great {N} Heist",
	"Last Call for {P}",
	"{#} {A} {P}",
	"Dinner With {P}",
	"The {N} Who Knew Too Much",
	"Attack of the {A} {P}",
	"{A} Business at the {N} Table",
	"{U}'s {A} {N}",
	"{U}'s {A} {G}",
	"{U} and the {A} {P}",
	"Everyone Versus {U}",
	"{U}'s {N} Problem",
}

// TableNameGenerator builds table names. Use NewTableNameGenerator; a
// test injects a seeded rng for a fixed sequence. It is safe for
// concurrent use.
type TableNameGenerator struct {
	mu  sync.Mutex
	rng *rand.Rand
}

// NewTableNameGenerator returns a generator drawing from rng, which it
// then owns.
func NewTableNameGenerator(rng *rand.Rand) *TableNameGenerator {
	return &TableNameGenerator{rng: rng}
}

// Generate returns a name no longer than maxTableNameLen bytes. display
// is the creator's display name, or "" for none. A blank, over-long
// (maxPersonalNameLen runes) or control-character name leaves the
// personal templates out, so a name is never cut.
func (g *TableNameGenerator) Generate(display string) string {
	g.mu.Lock()
	defer g.mu.Unlock()
	display = strings.TrimSpace(display)
	usable := personalNameUsable(display)
	for {
		t := nameTemplates[g.rng.IntN(len(nameTemplates))]
		if !usable && strings.Contains(t, "{U}") {
			// "Sometimes personal": four of the eighteen templates use
			// the creator's name, and none does without a usable one.
			continue
		}
		if out := g.fill(t, display); len(out) <= maxTableNameLen {
			return out
		}
	}
}

func personalNameUsable(s string) bool {
	if s == "" || utf8.RuneCountInString(s) > maxPersonalNameLen {
		return false
	}
	for _, r := range s {
		if r < ' ' || r == 0x7f {
			return false
		}
	}
	return true
}

func (g *TableNameGenerator) fill(t, display string) string {
	// One noun per template, so {N} and {P} agree.
	n := g.rng.IntN(len(nameNouns))
	var b strings.Builder
	for i := 0; i < len(t); i++ {
		if t[i] != '{' {
			b.WriteByte(t[i])
			continue
		}
		end := strings.IndexByte(t[i:], '}')
		switch t[i+1 : i+end] {
		case "A":
			b.WriteString(nameAdjectives[g.rng.IntN(len(nameAdjectives))])
		case "N":
			b.WriteString(nameNouns[n])
		case "P":
			b.WriteString(namePlurals[n])
		case "G":
			b.WriteString(nameGroups[g.rng.IntN(len(nameGroups))])
		case "#":
			b.WriteString(nameNumbers[g.rng.IntN(len(nameNumbers))])
		case "T":
			b.WriteString(nameTroubles[g.rng.IntN(len(nameTroubles))])
		case "U":
			b.WriteString(display)
		}
		i += end
	}
	return b.String()
}

// tableNames is the process's generator, seeded from the runtime's
// entropy. Tests build their own with a fixed seed.
var tableNames = NewTableNameGenerator(rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())))

// SuggestTableName is a fresh table name for display, a signed-in
// creator's display name, or "" for a purely random one.
func SuggestTableName(display string) string { return tableNames.Generate(display) }
