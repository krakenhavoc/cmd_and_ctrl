package effects

import (
	"sort"
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tokens_table.go — every plain token template in the catalog, as
// data (#581, Discussion #559 item 4). A key reads like the printed
// text: "1/1 white Soldier", "2/2 black Zombie", "1/1 blue Bird with
// flying", "4/4 green Rhino with trample". Tokens that carry a mana
// ability or an activated ability (Treasure, Food, Clue, Gold, …) are
// behaviour, not data, and keep their constructors in tokens.go.
//
// Ask for one with TokenCard("…"). A key that is not in the table panics,
// and TestEveryTokenKeyResolves checks every literal in the package at
// test time so the panic can only happen for a key built at runtime.
// Add a token by adding a row; if a card needs a variant (a tapped
// entry, counters), wrap the template in a TokenSpec.

// Token returns a fresh copy of the named template — fresh slices, so
// a card that appends to Keywords or Colors cannot leak into the next
// caller's token.
func TokenCard(key string) game.Card {
	t, ok := tokenTable[key]
	if !ok {
		panic("effects.TokenCard: no token template named " + strconv.Quote(key))
	}
	t.Colors = append([]string(nil), t.Colors...)
	t.Keywords = append([]string(nil), t.Keywords...)
	return t
}

// TokenKeys lists every key in the table, sorted. For tests and the
// catalogue page.
func TokenKeys() []string {
	keys := make([]string, 0, len(tokenTable))
	for k := range tokenTable {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

var tokenTable = map[string]game.Card{
	"0/0 colorless Construct artifact":              {Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 0, Toughness: 0},
	"0/0 white Spirit Cleric":                       {Name: "Spirit Cleric", TypeLine: "Token Creature — Spirit Cleric", Power: 0, Toughness: 0, Colors: []string{"W"}},
	"0/1 black Wizard":                              {Name: "Wizard", TypeLine: "Token Creature — Wizard", Power: 0, Toughness: 1, Colors: []string{"B"}},
	"0/1 colorless Plant":                           {Name: "Plant", TypeLine: "Token Creature — Plant", Power: 0, Toughness: 1},
	"0/1 green Plant":                               {Name: "Plant", TypeLine: "Token Creature — Plant", Power: 0, Toughness: 1, Colors: []string{"G"}},
	"0/1 red Kobolds of Kher Keep":                  {Name: "Kobolds of Kher Keep", TypeLine: "Token Creature — Kobold", Power: 0, Toughness: 1, Colors: []string{"R"}},
	"0/1 white Goat":                                {Name: "Goat", TypeLine: "Token Creature — Goat", Power: 0, Toughness: 1, Colors: []string{"W"}},
	"0/4 colorless Wall artifact with defender":     {Name: "Wall", TypeLine: "Token Artifact Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}},
	"1/1 black Bat with flying":                     {Name: "Bat", TypeLine: "Token Creature — Bat", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"flying"}},
	"1/1 black Insect":                              {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 black Rat":                                 {Name: "Rat", TypeLine: "Token Creature — Rat", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 black Slug":                                {Name: "Slug", TypeLine: "Token Creature — Slug", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 black and green Pest":                      {Name: "Pest", TypeLine: "Token Creature — Pest", Power: 1, Toughness: 1, Colors: []string{"B", "G"}},
	"1/1 blue Bird Illusion with flying":            {Name: "Bird Illusion", TypeLine: "Token Creature — Bird Illusion", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"1/1 blue Bird with flying and vigilance":       {Name: "Bird", TypeLine: "Token Creature — Bird", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying", "vigilance"}},
	"1/1 blue Fish":                                 {Name: "Fish", TypeLine: "Token Creature — Fish", Power: 1, Toughness: 1, Colors: []string{"U"}},
	"1/1 blue Tentacle":                             {Name: "Tentacle", TypeLine: "Token Creature — Tentacle", Power: 1, Toughness: 1, Colors: []string{"U"}},
	"1/1 blue Thopter artifact with flying":         {Name: "Thopter", TypeLine: "Token Artifact Creature — Thopter", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"1/1 blue and black Faerie with flying":         {Name: "Faerie", TypeLine: "Token Creature — Faerie", Power: 1, Toughness: 1, Colors: []string{"U", "B"}, Keywords: []string{"flying"}},
	"1/1 blue and red Insect with flying and haste": {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1, Colors: []string{"U", "R"}, Keywords: []string{"flying", "haste"}},
	"1/1 colorless Ally":                            {Name: "Ally", TypeLine: "Token Creature — Ally", Power: 1, Toughness: 1},
	"1/1 colorless Faerie Rogue with flying":        {Name: "Faerie Rogue", TypeLine: "Token Creature — Faerie Rogue", Power: 1, Toughness: 1, Keywords: []string{"flying"}},
	"1/1 colorless Gnome artifact":                  {Name: "Gnome", TypeLine: "Token Artifact Creature — Gnome", Power: 1, Toughness: 1},
	"1/1 colorless Human":                           {Name: "Human", TypeLine: "Token Creature — Human", Power: 1, Toughness: 1},
	"1/1 colorless Human Soldier":                   {Name: "Human Soldier", TypeLine: "Token Creature — Human Soldier", Power: 1, Toughness: 1},
	"1/1 colorless Insect with flying and haste":    {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1, Keywords: []string{"flying", "haste"}},
	"1/1 colorless Myr artifact":                    {Name: "Myr", TypeLine: "Token Artifact Creature — Myr", Power: 1, Toughness: 1},
	"1/1 colorless Snake with deathtouch":           {Name: "Snake", TypeLine: "Token Creature — Snake", Power: 1, Toughness: 1, Keywords: []string{"deathtouch"}},
	"1/1 colorless Soldier":                         {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1},
	"1/1 colorless Soldier artifact":                {Name: "Soldier", TypeLine: "Token Artifact Creature — Soldier", Power: 1, Toughness: 1},
	"1/1 colorless Spirit":                          {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1},
	"1/1 colorless Spirit with flying":              {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1, Keywords: []string{"flying"}},
	"1/1 colorless Thopter artifact with flying":    {Name: "Thopter", TypeLine: "Token Artifact Creature — Thopter", Power: 1, Toughness: 1, Keywords: []string{"flying"}},
	"1/1 green Dryad":                               {Name: "Dryad", TypeLine: "Token Land Creature — Forest Dryad", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 green Elf Warrior":                         {Name: "Elf Warrior", TypeLine: "Token Creature — Elf Warrior", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 green Insect with flying and deathtouch":   {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1, Colors: []string{"G"}, Keywords: []string{"flying", "deathtouch"}},
	"1/1 green Saproling":                           {Name: "Saproling", TypeLine: "Token Creature — Saproling", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 green Squirrel":                            {Name: "Squirrel", TypeLine: "Token Creature — Squirrel", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 red Dwarf":                                 {Name: "Dwarf", TypeLine: "Token Creature — Dwarf", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"1/1 red Elemental":                             {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"1/1 red Goblin":                                {Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"1/1 red Warrior":                               {Name: "Warrior", TypeLine: "Token Creature — Warrior", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"1/1 red and white Soldier with haste":          {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, Colors: []string{"R", "W"}, Keywords: []string{"haste"}},
	"1/1 white Bird with flying":                    {Name: "Bird", TypeLine: "Token Creature — Bird", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"1/1 white Cat":                                 {Name: "Cat", TypeLine: "Token Creature — Cat", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Cat Soldier with vigilance":          {Name: "Cat Soldier", TypeLine: "Token Creature — Cat Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"vigilance"}},
	"1/1 white Cat with lifelink":                   {Name: "Cat", TypeLine: "Token Creature — Cat", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"lifelink"}},
	"1/1 white Halfling":                            {Name: "Halfling", TypeLine: "Token Creature — Halfling", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Human Soldier":                       {Name: "Human Soldier", TypeLine: "Token Creature — Human Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Human Warrior":                       {Name: "Human Warrior", TypeLine: "Token Creature — Human Warrior", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Rabbit":                              {Name: "Rabbit", TypeLine: "Token Creature — Rabbit", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Soldier":                             {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Soldier with lifelink":               {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"lifelink"}},
	"1/1 white Spirit with flying":                  {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"1/1 white Vampire with lifelink":               {Name: "Vampire", TypeLine: "Token Creature — Vampire", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"lifelink"}},
	"1/1 white Warrior":                             {Name: "Warrior", TypeLine: "Token Creature — Warrior", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Zombie":                              {Name: "Zombie", TypeLine: "Token Creature — Zombie", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/2 green Spider with reach":                   {Name: "Spider", TypeLine: "Token Creature — Spider", Power: 1, Toughness: 2, Colors: []string{"G"}, Keywords: []string{"reach"}},
	"2/1 white and black Inkling with flying":       {Name: "Inkling", TypeLine: "Token Creature — Inkling", Power: 2, Toughness: 1, Colors: []string{"W", "B"}, Keywords: []string{"flying"}},
	"2/2 black Zombie":                              {Name: "Zombie", TypeLine: "Token Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}},
	"2/2 black Zombie Druid":                        {Name: "Zombie Druid", TypeLine: "Token Creature — Zombie Druid", Power: 2, Toughness: 2, Colors: []string{"B"}},
	"2/2 blue Bird with flying":                     {Name: "Bird", TypeLine: "Token Creature — Bird", Power: 2, Toughness: 2, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"2/2 blue Drake with flying":                    {Name: "Drake", TypeLine: "Token Creature — Drake", Power: 2, Toughness: 2, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"2/2 blue Spirit with vigilance":                {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 2, Toughness: 2, Colors: []string{"U"}, Keywords: []string{"vigilance"}},
	"2/2 colorless Bird with flying":                {Name: "Bird", TypeLine: "Token Creature — Bird", Power: 2, Toughness: 2, Keywords: []string{"flying"}},
	"2/2 colorless Cat":                             {Name: "Cat", TypeLine: "Token Creature — Cat", Power: 2, Toughness: 2},
	"2/2 colorless Knight with vigilance":           {Name: "Knight", TypeLine: "Token Creature — Knight", Power: 2, Toughness: 2, Keywords: []string{"vigilance"}},
	"2/2 colorless Samurai with vigilance":          {Name: "Samurai", TypeLine: "Token Creature — Samurai", Power: 2, Toughness: 2, Keywords: []string{"vigilance"}},
	"2/2 colorless Zombie":                          {Name: "Zombie", TypeLine: "Token Creature — Zombie", Power: 2, Toughness: 2},
	"2/2 green Boar":                                {Name: "Boar", TypeLine: "Token Creature — Boar", Power: 2, Toughness: 2, Colors: []string{"G"}},
	"2/2 green Elemental":                           {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 2, Toughness: 2, Colors: []string{"G"}},
	"2/2 green Spider with reach":                   {Name: "Spider", TypeLine: "Token Creature — Spider", Power: 2, Toughness: 2, Colors: []string{"G"}, Keywords: []string{"reach"}},
	"2/2 white Pegasus with flying":                 {Name: "Pegasus", TypeLine: "Token Creature — Pegasus", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"2/2 white Soldier with vigilance":              {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"vigilance"}},
	"3/1 red Dinosaur":                              {Name: "Dinosaur", TypeLine: "Token Creature — Dinosaur", Power: 3, Toughness: 1, Colors: []string{"R"}},
	"3/3 colorless Ape":                             {Name: "Ape", TypeLine: "Token Creature — Ape", Power: 3, Toughness: 3},
	"3/3 colorless Beast":                           {Name: "Beast", TypeLine: "Token Creature — Beast", Power: 3, Toughness: 3},
	"3/3 colorless Elephant":                        {Name: "Elephant", TypeLine: "Token Creature — Elephant", Power: 3, Toughness: 3},
	"3/3 colorless Frog Lizard":                     {Name: "Frog Lizard", TypeLine: "Token Creature — Frog Lizard", Power: 3, Toughness: 3},
	"3/3 colorless Golem artifact":                  {Name: "Golem", TypeLine: "Token Enchantment Artifact Creature — Golem", Power: 3, Toughness: 3},
	"3/3 colorless Phyrexian Wurm artifact":         {Name: "Phyrexian Wurm", TypeLine: "Token Artifact Creature — Phyrexian Wurm", Power: 3, Toughness: 3},
	"3/3 green Dinosaur with trample":               {Name: "Dinosaur", TypeLine: "Token Creature — Dinosaur", Power: 3, Toughness: 3, Colors: []string{"G"}, Keywords: []string{"trample"}},
	"3/3 green Elephant":                            {Name: "Elephant", TypeLine: "Token Creature — Elephant", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"3/3 red Ogre":                                  {Name: "Ogre", TypeLine: "Token Creature — Ogre", Power: 3, Toughness: 3, Colors: []string{"R"}},
	"4/4 black Zombie Warrior with vigilance":       {Name: "Zombie Warrior", TypeLine: "Token Creature — Zombie Warrior", Power: 4, Toughness: 4, Colors: []string{"B"}, Keywords: []string{"vigilance"}},
	"4/4 blue and red Elemental":                    {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 4, Toughness: 4, Colors: []string{"U", "R"}},
	"4/4 colorless Angel with flying and vigilance": {Name: "Angel", TypeLine: "Token Creature — Angel", Power: 4, Toughness: 4, Keywords: []string{"flying", "vigilance"}},
	"4/4 colorless Beast":                           {Name: "Beast", TypeLine: "Token Creature — Beast", Power: 4, Toughness: 4},
	"4/4 green Bear":                                {Name: "Bear", TypeLine: "Token Creature — Bear", Power: 4, Toughness: 4, Colors: []string{"G"}},
	"4/4 green Phyrexian Beast":                     {Name: "Phyrexian Beast", TypeLine: "Token Creature — Phyrexian Beast", Power: 4, Toughness: 4, Colors: []string{"G"}},
	"4/4 white Angel with flying":                   {Name: "Angel", TypeLine: "Token Creature — Angel", Power: 4, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"4/4 white Angel with flying and vigilance":     {Name: "Angel", TypeLine: "Token Creature — Angel", Power: 4, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"flying", "vigilance"}},
	"5/3 green Elemental":                           {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 5, Toughness: 3, Colors: []string{"G"}},
	"5/5 red and green Elemental":                   {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 5, Toughness: 5, Colors: []string{"R", "G"}},
	"8/8 blue Scion of the Deep":                    {Name: "Scion of the Deep", TypeLine: "Token Legendary Creature — Octopus", Power: 8, Toughness: 8, Colors: []string{"U"}},
	"Munitions":                                     {Name: "Munitions", TypeLine: "Token Artifact"},

	// A key longer than every row above: kept in its own block so gofmt
	// does not re-align the whole table (and every open PR's rows) for it.
	"6/12 colorless Construct artifact with trample": {Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 6, Toughness: 12, Keywords: []string{"trample"}},
}
