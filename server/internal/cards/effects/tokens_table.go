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
//
// A TOKEN WHOSE PRINTED VERSION ALWAYS CARRIES AN ABILITY DOES NOT
// BELONG HERE AT ALL. The Pest is always "when this token dies, you
// gain 1 life"; the 0/1 black Wizard is always the noncreature-cast
// ping; Fable's Goblin Shaman is always the attack trigger. A row for
// one of those is a textless version of a token that has never been
// printed textless, and the next card to reach for it would ship
// silently weaker than printed. Those three left this table when
// ADR 0083 gave them catalog templates (tokens.go,
// fable_of_the_mirror_breaker.go); TokenCard panics on the key, and
// TestEveryTokenKeyResolves catches it at test time.
//
// THE COLOUR IN THE KEY IS THE COLOUR IN THE ROW. `Colors` is a
// characteristic the rules read — protection, lords, "destroy target
// white creature", devotion — so a row whose colour is wrong is a
// rules bug and not a cosmetic one (#1127: 21 rows declared no colour
// where the printed token is coloured, and a colourless Soldier sat in
// Elspeth's +1 for a year). Two tests hold the line:
// TestTokenTableRowsAreWellFormed pins the key against `Colors`
// hermetically, and TestRealDumpEveryTokenTemplateMatchesAPrintedToken
// pins the whole row against the printed token card in the Scryfall
// dump. Take the colour from the ORACLE TEXT of the card that makes
// the token — that is what the token is — and reuse an existing row
// rather than adding a second one that differs only in colour.

// Token returns a fresh copy of the named template — fresh slices, so
// a card that appends to Keywords or Colors cannot leak into the next
// caller's token.
func TokenCard(key string) game.Card {
	t, ok := TokenTable[key]
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
	keys := make([]string, 0, len(TokenTable))
	for k := range TokenTable {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

var TokenTable = map[string]game.Card{
	// Magic Origins, Nissa, Sage Animist (S82 older sets, batch 322).
	"4/4 green Legendary Elemental named Ashaya, the Awoken World": {Name: "Ashaya, the Awoken World", TypeLine: "Token Legendary Creature — Elemental", Power: 4, Toughness: 4, Colors: []string{"G"}},
	// Warhammer 40,000 Commander Tyranids (#5948, S82 GAP batch 192).
	"1/1 green Tyranid":                                             {Name: "Tyranid", TypeLine: "Token Creature — Tyranid", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"5/5 green Tyranid":                                             {Name: "Tyranid", TypeLine: "Token Creature — Tyranid", Power: 5, Toughness: 5, Colors: []string{"G"}},
	"3/3 green Tyranid Warrior with trample":                        {Name: "Tyranid Warrior", TypeLine: "Token Creature — Tyranid Warrior", Power: 3, Toughness: 3, Colors: []string{"G"}, Keywords: []string{"trample"}},
	"1/1 green and white Human Citizen":                             {Name: "Human Citizen", TypeLine: "Token Creature — Human Citizen", Power: 1, Toughness: 1, Colors: []string{"G", "W"}},
	"1/1 green and white Citizen":                                   {Name: "Citizen", TypeLine: "Token Creature — Citizen", Power: 1, Toughness: 1, Colors: []string{"G", "W"}},
	"1/1 green and white Kithkin":                                   {Name: "Kithkin", TypeLine: "Token Creature — Kithkin", Power: 1, Toughness: 1, Colors: []string{"G", "W"}},
	"2/2 black and green Elf":                                       {Name: "Elf", TypeLine: "Token Creature — Elf", Power: 2, Toughness: 2, Colors: []string{"B", "G"}},
	"1/1 green and white Elf Warrior":                               {Name: "Elf Warrior", TypeLine: "Token Creature — Elf Warrior", Power: 1, Toughness: 1, Colors: []string{"G", "W"}},
	"3/3 blue Serpent named Koma's Coil":                            {Name: "Koma's Coil", TypeLine: "Token Creature — Serpent", Power: 3, Toughness: 3, Colors: []string{"U"}},
	"1/1 blue Faerie Dragon with flying":                            {Name: "Faerie Dragon", TypeLine: "Token Creature — Faerie Dragon", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"1/1 blue Faerie with flying":                                   {Name: "Faerie", TypeLine: "Token Creature — Faerie", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"2/2 blue Djinn Monk with flying":                               {Name: "Djinn Monk", TypeLine: "Token Creature — Djinn Monk", Power: 2, Toughness: 2, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"2/2 blue Homunculus":                                           {Name: "Homunculus", TypeLine: "Token Creature — Homunculus", Power: 2, Toughness: 2, Colors: []string{"U"}},
	"0/0 blue Alien":                                                {Name: "Alien", TypeLine: "Token Creature — Alien", Power: 0, Toughness: 0, Colors: []string{"U"}, PrintedPTKnown: true},
	"0/0 black Phyrexian Germ":                                      {Name: "Phyrexian Germ", TypeLine: "Token Creature — Phyrexian Germ", Power: 0, Toughness: 0, Colors: []string{"B"}, PrintedPTKnown: true},
	"0/0 green and blue Fractal":                                    {Name: "Fractal", TypeLine: "Token Creature — Fractal", Power: 0, Toughness: 0, Colors: []string{"G", "U"}, PrintedPTKnown: true},
	"0/0 green and blue Fractal named Primo, the Indivisible":       {Name: "Primo, the Indivisible", TypeLine: "Token Legendary Creature — Fractal", Power: 0, Toughness: 0, Colors: []string{"G", "U"}, PrintedPTKnown: true},
	"2/2 colorless Construct artifact":                              {Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 2, Toughness: 2},
	"0/0 colorless Construct artifact":                              {Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 0, Toughness: 0},
	"0/0 red Gremlin artifact":                                      {Name: "Gremlin", TypeLine: "Token Artifact Creature — Gremlin", Power: 0, Toughness: 0, Colors: []string{"R"}, PrintedPTKnown: true},
	"0/1 black Insect":                                              {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 0, Toughness: 1, Colors: []string{"B"}},
	"0/1 green Egg with defender":                                   {Name: "Egg", TypeLine: "Token Creature — Egg", Power: 0, Toughness: 1, Colors: []string{"G"}, Keywords: []string{"defender"}},
	"0/2 green Plant with defender":                                 {Name: "Plant", TypeLine: "Token Creature — Plant", Power: 0, Toughness: 2, Colors: []string{"G"}, Keywords: []string{"defender"}},
	"0/1 green Plant":                                               {Name: "Plant", TypeLine: "Token Creature — Plant", Power: 0, Toughness: 1, Colors: []string{"G"}},
	"1/1 green Plant":                                               {Name: "Plant", TypeLine: "Token Creature — Plant", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"0/1 green Sheep":                                               {Name: "Sheep", TypeLine: "Token Creature — Sheep", Power: 0, Toughness: 1, Colors: []string{"G"}},
	"3/3 green Hippo":                                               {Name: "Hippo", TypeLine: "Token Creature — Hippo", Power: 3, Toughness: 3, Colors: []string{"G"}},                                                    // Mouth // Feed
	"1/1 green Hippo":                                               {Name: "Hippo", TypeLine: "Token Creature — Hippo", Power: 1, Toughness: 1, Colors: []string{"G"}},                                                    // Questing Phelddagrif
	"1/1 black and green Snake enchantment with deathtouch":         {Name: "Snake", TypeLine: "Token Enchantment Creature — Snake", Power: 1, Toughness: 1, Colors: []string{"B", "G"}, Keywords: []string{"deathtouch"}}, // Pharika, God of Affliction
	"2/2 black Zombie Walker":                                       {Name: "Walker", TypeLine: "Token Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}},                                                  // Hansk, Slayer Zealot
	"0/1 red Kobolds of Kher Keep":                                  {Name: "Kobolds of Kher Keep", TypeLine: "Token Creature — Kobold", Power: 0, Toughness: 1, Colors: []string{"R"}},
	"0/1 white Goat":                                                {Name: "Goat", TypeLine: "Token Creature — Goat", Power: 0, Toughness: 1, Colors: []string{"W"}},
	"0/3 blue Crab":                                                 {Name: "Crab", TypeLine: "Token Creature — Crab", Power: 0, Toughness: 3, Colors: []string{"U"}},
	"0/4 colorless Wall artifact with defender":                     {Name: "Wall", TypeLine: "Token Artifact Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}},
	"1/1 colorless Robot artifact with flying":                      {Name: "Robot", TypeLine: "Token Artifact Creature — Robot", Power: 1, Toughness: 1, Keywords: []string{"flying"}},
	"3/3 red Ape Villain with haste":                                {Name: "Ape Villain", TypeLine: "Token Creature — Ape Villain", Power: 3, Toughness: 3, Colors: []string{"R"}, Keywords: []string{"haste"}},
	"0/4 colorless Wall with defender":                              {Name: "Wall", TypeLine: "Token Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}},
	"0/3 colorless Wall with defender and reach":                    {Name: "Wall", TypeLine: "Token Creature — Wall", Power: 0, Toughness: 3, Keywords: []string{"defender", "reach"}},
	"3/1 colorless Wall artifact with defender named Stone Boulder": {Name: "Stone Boulder", TypeLine: "Token Artifact Creature — Wall", Power: 3, Toughness: 1, Keywords: []string{"defender"}},
	"0/4 white Wall with defender and flying":                       {Name: "Wall", TypeLine: "Token Creature — Wall", Power: 0, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"defender", "flying"}},
	"0/4 white Wall with defender":                                  {Name: "Wall", TypeLine: "Token Creature — Wall", Power: 0, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"defender"}},
	"1/3 white Wall with defender":                                  {Name: "Wall", TypeLine: "Token Creature — Wall", Power: 1, Toughness: 3, Colors: []string{"W"}, Keywords: []string{"defender"}},
	"0/3 white Wall with defender":                                  {Name: "Wall", TypeLine: "Token Creature — Wall", Power: 0, Toughness: 3, Colors: []string{"W"}, Keywords: []string{"defender"}},
	"1/1 black Bat with flying":                                     {Name: "Bat", TypeLine: "Token Creature — Bat", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"flying"}},
	"1/1 black Harpy with flying":                                   {Name: "Harpy", TypeLine: "Token Creature — Harpy", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"flying"}}, // Abhorrent Overlord
	"1/1 black Minion":                                              {Name: "Minion", TypeLine: "Token Creature — Minion", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 black Horror":                                              {Name: "Horror", TypeLine: "Token Creature — Horror", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 black Insect":                                              {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 black Rat":                                                 {Name: "Rat", TypeLine: "Token Creature — Rat", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 black Rat Rogue":                                           {Name: "Rat Rogue", TypeLine: "Token Creature — Rat Rogue", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 white Human Rogue":                                         {Name: "Human Rogue", TypeLine: "Token Creature — Human Rogue", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 black Rat with lifelink":                                   {Name: "Rat", TypeLine: "Token Creature — Rat", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"lifelink"}},
	"1/1 black Snail":                                               {Name: "Snail", TypeLine: "Token Creature — Snail", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 black Slug":                                                {Name: "Slug", TypeLine: "Token Creature — Slug", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 black Thrull":                                              {Name: "Thrull", TypeLine: "Token Creature — Thrull", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 blue Bird Illusion with flying":                            {Name: "Bird Illusion", TypeLine: "Token Creature — Bird Illusion", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"1/1 blue Bird with flying and vigilance":                       {Name: "Bird", TypeLine: "Token Creature — Bird", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying", "vigilance"}},
	"1/1 blue Illusion":                                             {Name: "Illusion", TypeLine: "Token Creature — Illusion", Power: 1, Toughness: 1, Colors: []string{"U"}},
	"1/1 blue Illusion with flying":                                 {Name: "Illusion", TypeLine: "Token Creature — Illusion", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"2/2 blue Illusion":                                             {Name: "Illusion", TypeLine: "Token Creature — Illusion", Power: 2, Toughness: 2, Colors: []string{"U"}},
	"1/1 blue Fish":                                                 {Name: "Fish", TypeLine: "Token Creature — Fish", Power: 1, Toughness: 1, Colors: []string{"U"}},
	"1/1 blue Squid with islandwalk":                                {Name: "Squid", TypeLine: "Token Creature — Squid", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"islandwalk"}},
	"1/1 blue Merfolk":                                              {Name: "Merfolk", TypeLine: "Token Creature — Merfolk", Power: 1, Toughness: 1, Colors: []string{"U"}},
	"1/1 blue Merfolk with hexproof":                                {Name: "Merfolk", TypeLine: "Token Creature — Merfolk", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"hexproof"}},
	"1/1 white and blue Merfolk":                                    {Name: "Merfolk", TypeLine: "Token Creature — Merfolk", Power: 1, Toughness: 1, Colors: []string{"W", "U"}},
	"1/1 blue Tentacle":                                             {Name: "Tentacle", TypeLine: "Token Creature — Tentacle", Power: 1, Toughness: 1, Colors: []string{"U"}},
	"1/1 blue Thopter artifact with flying":                         {Name: "Thopter", TypeLine: "Token Artifact Creature — Thopter", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"1/1 blue and black Faerie with flying":                         {Name: "Faerie", TypeLine: "Token Creature — Faerie", Power: 1, Toughness: 1, Colors: []string{"U", "B"}, Keywords: []string{"flying"}},
	"1/1 blue and black Faerie Rogue with flying":                   {Name: "Faerie Rogue", TypeLine: "Token Creature — Faerie Rogue", Power: 1, Toughness: 1, Colors: []string{"U", "B"}, Keywords: []string{"flying"}},
	"1/1 blue and black Zombie Wizard":                              {Name: "Zombie Wizard", TypeLine: "Token Creature — Zombie Wizard", Power: 1, Toughness: 1, Colors: []string{"U", "B"}},
	"2/2 blue and black Zombie Rogue":                               {Name: "Zombie Rogue", TypeLine: "Token Creature — Zombie Rogue", Power: 2, Toughness: 2, Colors: []string{"U", "B"}},
	"1/1 blue and red Elemental":                                    {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 1, Toughness: 1, Colors: []string{"U", "R"}},
	"1/1 blue and red Insect with flying and haste":                 {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1, Colors: []string{"U", "R"}, Keywords: []string{"flying", "haste"}},
	"1/1 black and green Insect with flying":                        {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1, Colors: []string{"B", "G"}, Keywords: []string{"flying"}},
	"1/1 black and green Insect":                                    {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1, Colors: []string{"B", "G"}},
	"1/1 white Sheep":                                               {Name: "Sheep", TypeLine: "Token Creature — Sheep", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Mouse":                                               {Name: "Mouse", TypeLine: "Token Creature — Mouse", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Ally":                                                {Name: "Ally", TypeLine: "Token Creature — Ally", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Athlete":                                             {Name: "Athlete", TypeLine: "Token Creature — Athlete", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 black Faerie Rogue with flying":                            {Name: "Faerie Rogue", TypeLine: "Token Creature — Faerie Rogue", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"flying"}},
	"1/1 colorless Construct artifact":                              {Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 1, Toughness: 1},
	"1/1 colorless Construct artifact with defender":                {Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 1, Toughness: 1, Keywords: []string{"defender"}}, // Daretti, Ingenious Iconoclast (#5462)
	"1/1 colorless Gnome artifact":                                  {Name: "Gnome", TypeLine: "Token Artifact Creature — Gnome", Power: 1, Toughness: 1},
	"1/1 colorless Hero":                                            {Name: "Hero", TypeLine: "Token Creature — Hero", Power: 1, Toughness: 1},
	"1/1 white Human":                                               {Name: "Human", TypeLine: "Token Creature — Human", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 colorless Insect artifact with flying":                     {Name: "Insect", TypeLine: "Token Artifact Creature — Insect", Power: 1, Toughness: 1, Keywords: []string{"flying"}},
	"1/1 colorless Myr artifact":                                    {Name: "Myr", TypeLine: "Token Artifact Creature — Myr", Power: 1, Toughness: 1},
	"1/1 colorless Sliver":                                          {Name: "Sliver", TypeLine: "Token Creature — Sliver", Power: 1, Toughness: 1},
	"1/1 black Assassin with deathtouch and haste":                  {Name: "Assassin", TypeLine: "Token Creature — Assassin", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"deathtouch", "haste"}},
	"1/1 black Assassin with menace":                                {Name: "Assassin", TypeLine: "Token Creature — Assassin", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"menace"}}, // Ratonhnhaké꞉ton (#5948)
	"1/1 black Snake with deathtouch":                               {Name: "Snake", TypeLine: "Token Creature — Snake", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"deathtouch"}},
	"1/1 green Snake with deathtouch":                               {Name: "Snake", TypeLine: "Token Creature — Snake", Power: 1, Toughness: 1, Colors: []string{"G"}, Keywords: []string{"deathtouch"}},
	"2/2 colorless Scarecrow artifact":                              {Name: "Scarecrow", TypeLine: "Token Artifact Creature — Scarecrow", Power: 2, Toughness: 2},
	"1/1 black Wolf with deathtouch":                                {Name: "Wolf", TypeLine: "Token Creature — Wolf", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"deathtouch"}},
	"1/1 colorless Servo artifact":                                  {Name: "Servo", TypeLine: "Token Artifact Creature — Servo", Power: 1, Toughness: 1},
	"1/1 colorless Soldier artifact":                                {Name: "Soldier", TypeLine: "Token Artifact Creature — Soldier", Power: 1, Toughness: 1},
	"3/3 colorless Zombie artifact":                                 {Name: "Zombie", TypeLine: "Token Artifact Creature — Zombie", Power: 3, Toughness: 3},
	"Scrap artifact":                                                {Name: "Scrap", TypeLine: "Token Artifact"},
	"1/1 colorless Spirit":                                          {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1},
	"2/2 red Spirit with menace":                                    {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 2, Toughness: 2, Colors: []string{"R"}, Keywords: []string{"menace"}},
	"1/1 colorless Thopter artifact with flying":                    {Name: "Thopter", TypeLine: "Token Artifact Creature — Thopter", Power: 1, Toughness: 1, Keywords: []string{"flying"}},
	"3/3 green Forest Tentacle":                                     {Name: "Forest Tentacle", TypeLine: "Token Land Creature — Forest Tentacle", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"1/1 green Forest Dryad":                                        {Name: "Forest Dryad", TypeLine: "Token Land Creature — Forest Dryad", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 green Elf Warrior":                                         {Name: "Elf Warrior", TypeLine: "Token Creature — Elf Warrior", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 green Elf":                                                 {Name: "Elf", TypeLine: "Token Creature — Elf", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 green Snake":                                               {Name: "Snake", TypeLine: "Token Creature — Snake", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 green and blue Snake":                                      {Name: "Snake", TypeLine: "Token Creature — Snake", Power: 1, Toughness: 1, Colors: []string{"G", "U"}}, // #5598: Patagia Viper
	"5/4 green Snake":                                               {Name: "Snake", TypeLine: "Token Creature — Snake", Power: 5, Toughness: 4, Colors: []string{"G"}},
	"1/1 green Frog":                                                {Name: "Frog", TypeLine: "Token Creature — Frog", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 green Insect":                                              {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 green Butterfly Insect with flying":                        {Name: "Butterfly", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1, Colors: []string{"G"}, Keywords: []string{"flying"}},
	"1/1 green Phyrexian Insect with infect":                        {Name: "Phyrexian Insect", TypeLine: "Token Creature — Phyrexian Insect", Power: 1, Toughness: 1, Colors: []string{"G"}, Keywords: []string{"infect"}},
	"1/1 green Insect with flying and deathtouch":                   {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1, Colors: []string{"G"}, Keywords: []string{"flying", "deathtouch"}},
	"2/1 white Insect with flying":                                  {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 2, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"1/1 green Phyrexian Saproling":                                 {Name: "Phyrexian Saproling", TypeLine: "Token Creature — Phyrexian Saproling", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 green Saproling":                                           {Name: "Saproling", TypeLine: "Token Creature — Saproling", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 green Squirrel":                                            {Name: "Squirrel", TypeLine: "Token Creature — Squirrel", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"1/1 red Dwarf":                                                 {Name: "Dwarf", TypeLine: "Token Creature — Dwarf", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"1/1 red Elemental":                                             {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"1/1 red Elemental with haste":                                  {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 1, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"haste"}}, // Tempt with Vengeance

	"1/1 red Hamster":                                                  {Name: "Hamster", TypeLine: "Token Creature — Hamster", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"1/2 white Soldier with defender":                                  {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"defender"}},
	"1/1 red Goblin":                                                   {Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"1/1 red Phyrexian Goblin":                                         {Name: "Phyrexian Goblin", TypeLine: "Token Creature — Phyrexian Goblin", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"1/1 red Human":                                                    {Name: "Human", TypeLine: "Token Creature — Human", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"1/1 red Goblin with haste":                                        {Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"haste"}},
	"1/1 red Warrior":                                                  {Name: "Warrior", TypeLine: "Token Creature — Warrior", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"1/1 red Survivor":                                                 {Name: "Survivor", TypeLine: "Token Creature — Survivor", Power: 1, Toughness: 1, Colors: []string{"R"}}, // Varchild, Betrayer of Kjeldor
	"1/1 red and white Soldier with haste":                             {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, Colors: []string{"R", "W"}, Keywords: []string{"haste"}},
	"1/1 blue Bird with flying":                                        {Name: "Bird", TypeLine: "Token Creature — Bird", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"1/1 white Cat Bird with flying":                                   {Name: "Cat Bird", TypeLine: "Token Creature — Cat Bird", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"1/1 white Bird with flying":                                       {Name: "Bird", TypeLine: "Token Creature — Bird", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"1/1 white Bird Soldier with flying":                               {Name: "Bird Soldier", TypeLine: "Token Creature — Bird Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"flying"}}, // #5586: Flurry of Wings
	"3/3 white Bird with flying":                                       {Name: "Bird", TypeLine: "Token Creature — Bird", Power: 3, Toughness: 3, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"3/4 white Bird with flying":                                       {Name: "Bird", TypeLine: "Token Creature — Bird", Power: 3, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"flying"}}, // S82 older sets batch 152: Wingmate Roc
	"1/1 red Goblin Wizard with prowess":                               {Name: "Goblin Wizard", TypeLine: "Token Creature — Goblin Wizard", Power: 1, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"prowess"}},
	"1/1 red Monk with prowess":                                        {Name: "Monk", TypeLine: "Token Creature — Monk", Power: 1, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"prowess"}},
	"1/1 white Monk with prowess":                                      {Name: "Monk", TypeLine: "Token Creature — Monk", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"prowess"}},
	"1/0 blue Elemental":                                               {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 1, Toughness: 0, Colors: []string{"U"}}, // Master of Waves
	"2/2 black Walker":                                                 {Name: "Walker", TypeLine: "Token Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}},       // Enkira, Hostile Scavenger
	"1/1 green Cat":                                                    {Name: "Cat", TypeLine: "Token Creature — Cat", Power: 1, Toughness: 1, Colors: []string{"G"}},             // Rin and Seri, Inseparable
	"1/1 white Cat":                                                    {Name: "Cat", TypeLine: "Token Creature — Cat", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Dog":                                                    {Name: "Dog", TypeLine: "Token Creature — Dog", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Cat Soldier with vigilance":                             {Name: "Cat Soldier", TypeLine: "Token Creature — Cat Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"vigilance"}},
	"2/2 white Cat Beast":                                              {Name: "Cat Beast", TypeLine: "Token Creature — Cat Beast", Power: 2, Toughness: 2, Colors: []string{"W"}},
	"2/2 white Unicorn":                                                {Name: "Unicorn", TypeLine: "Token Creature — Unicorn", Power: 2, Toughness: 2, Colors: []string{"W"}},
	"2/2 white Cat":                                                    {Name: "Cat", TypeLine: "Token Creature — Cat", Power: 2, Toughness: 2, Colors: []string{"W"}},
	"2/2 white Griffin with flying":                                    {Name: "Griffin", TypeLine: "Token Creature — Griffin", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"1/1 white Cat with lifelink":                                      {Name: "Cat", TypeLine: "Token Creature — Cat", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"lifelink"}},
	"1/1 white Halfling":                                               {Name: "Halfling", TypeLine: "Token Creature — Halfling", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Human Soldier":                                          {Name: "Human Soldier", TypeLine: "Token Creature — Human Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Human Warrior":                                          {Name: "Human Warrior", TypeLine: "Token Creature — Human Warrior", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white and black Human Cleric":                                 {Name: "Human Cleric", TypeLine: "Token Creature — Human Cleric", Power: 1, Toughness: 1, Colors: []string{"W", "B"}},
	"1/1 white Rabbit":                                                 {Name: "Rabbit", TypeLine: "Token Creature — Rabbit", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Kor Warrior":                                            {Name: "Kor Warrior", TypeLine: "Token Creature — Kor Warrior", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Kor Soldier":                                            {Name: "Kor Soldier", TypeLine: "Token Creature — Kor Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Kor Ally":                                               {Name: "Kor Ally", TypeLine: "Token Creature — Kor Ally", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Soldier":                                                {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Soldier Ally":                                           {Name: "Soldier Ally", TypeLine: "Token Creature — Soldier Ally", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Soldier with lifelink":                                  {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"lifelink"}},
	"3/1 white Spirit with flying":                                     {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 3, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"1/1 red Gremlin":                                                  {Name: "Gremlin", TypeLine: "Token Creature — Gremlin", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"1/1 white Spirit":                                                 {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Spirit with flying":                                     {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"2/2 white Spirit with flying":                                     {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"1/1 blue Human Wizard":                                            {Name: "Human Wizard", TypeLine: "Token Creature — Human Wizard", Power: 1, Toughness: 1, Colors: []string{"U"}},
	"3/3 white Spirit with flying":                                     {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 3, Toughness: 3, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"4/4 white Spirit with flying":                                     {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 4, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"1/1 black Vampire":                                                {Name: "Vampire", TypeLine: "Token Creature — Vampire", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 black Vampire with lifelink":                                  {Name: "Vampire", TypeLine: "Token Creature — Vampire", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"lifelink"}},
	"1/1 black Vampire Knight with lifelink":                           {Name: "Vampire Knight", TypeLine: "Token Creature — Vampire Knight", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"lifelink"}},
	"1/1 black Vampire Rogue with lifelink":                            {Name: "Vampire Rogue", TypeLine: "Token Creature — Vampire Rogue", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"lifelink"}},
	"2/2 black Vampire with flying":                                    {Name: "Vampire", TypeLine: "Token Creature — Vampire", Power: 2, Toughness: 2, Colors: []string{"B"}, Keywords: []string{"flying"}},
	"2/3 black Vampire with flying and lifelink":                       {Name: "Vampire", TypeLine: "Token Creature — Vampire", Power: 2, Toughness: 3, Colors: []string{"B"}, Keywords: []string{"flying", "lifelink"}},
	"1/1 white Vampire with lifelink":                                  {Name: "Vampire", TypeLine: "Token Creature — Vampire", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"lifelink"}},
	"1/1 white Warrior":                                                {Name: "Warrior", TypeLine: "Token Creature — Warrior", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Warrior with vigilance":                                 {Name: "Warrior", TypeLine: "Token Creature — Warrior", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"vigilance"}},
	"1/1 white Zombie":                                                 {Name: "Zombie", TypeLine: "Token Creature — Zombie", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"2/1 green Spider with reach":                                      {Name: "Spider", TypeLine: "Token Creature — Spider", Power: 2, Toughness: 1, Colors: []string{"G"}, Keywords: []string{"reach"}},
	"3/3 green and white Phyrexian Hydra with reach":                   {Name: "Phyrexian Hydra", TypeLine: "Token Creature — Phyrexian Hydra", Power: 3, Toughness: 3, Colors: []string{"G", "W"}, Keywords: []string{"reach"}},
	"3/3 green and white Phyrexian Hydra with lifelink":                {Name: "Phyrexian Hydra", TypeLine: "Token Creature — Phyrexian Hydra", Power: 3, Toughness: 3, Colors: []string{"G", "W"}, Keywords: []string{"lifelink"}},
	"2/1 green Varmint":                                                {Name: "Varmint", TypeLine: "Token Creature — Varmint", Power: 2, Toughness: 1, Colors: []string{"G"}}, // Rise of the Varmints (#4468)
	"2/1 black and green Spider with menace and reach":                 {Name: "Spider", TypeLine: "Token Creature — Spider", Power: 2, Toughness: 1, Colors: []string{"B", "G"}, Keywords: []string{"menace", "reach"}},
	"1/2 green Spider with reach":                                      {Name: "Spider", TypeLine: "Token Creature — Spider", Power: 1, Toughness: 2, Colors: []string{"G"}, Keywords: []string{"reach"}},
	"2/4 black Spider with reach":                                      {Name: "Spider", TypeLine: "Token Creature — Spider", Power: 2, Toughness: 4, Colors: []string{"B"}, Keywords: []string{"reach"}}, // Penumbra Spider (#5596)
	"1/2 blue Moonfolk with flying":                                    {Name: "Moonfolk", TypeLine: "Token Creature — Moonfolk", Power: 1, Toughness: 2, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"1/2 white Moogle with lifelink":                                   {Name: "Moogle", TypeLine: "Token Creature — Moogle", Power: 1, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"lifelink"}},
	"10/10 colorless Eldrazi":                                          {Name: "Eldrazi", TypeLine: "Token Creature — Eldrazi", Power: 10, Toughness: 10},
	"3/1 red Elemental":                                                {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 3, Toughness: 1, Colors: []string{"R"}},
	"3/1 red Elemental with haste":                                     {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 3, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"haste"}}, // Rakka Mar
	"3/1 red Knight":                                                   {Name: "Knight", TypeLine: "Token Creature — Knight", Power: 3, Toughness: 1, Colors: []string{"R"}},
	"3/1 black Mindless-One":                                           {Name: "Mindless-One", TypeLine: "Token Creature — Mindless-One", Power: 3, Toughness: 1, Colors: []string{"B"}},
	"2/1 red Elemental with trample and haste":                         {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 2, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"trample", "haste"}},
	"3/1 red Elemental with trample and haste":                         {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 3, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"trample", "haste"}}, // Valduk, Keeper of the Flame
	"7/1 red Elemental with trample and haste":                         {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 7, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"trample", "haste"}}, // Elemental Appeal
	"2/1 red Dwarf Berserker":                                          {Name: "Dwarf Berserker", TypeLine: "Token Creature — Dwarf Berserker", Power: 2, Toughness: 1, Colors: []string{"R"}},
	"2/2 red Soldier":                                                  {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 2, Toughness: 2, Colors: []string{"R"}},
	"2/2 red Dwarf":                                                    {Name: "Dwarf", TypeLine: "Token Creature — Dwarf", Power: 2, Toughness: 2, Colors: []string{"R"}}, // The Hobbit (#4475)
	"2/3 red Minotaur":                                                 {Name: "Minotaur", TypeLine: "Token Creature — Minotaur", Power: 2, Toughness: 3, Colors: []string{"R"}},
	"2/1 white and black Inkling with flying":                          {Name: "Inkling", TypeLine: "Token Creature — Inkling", Power: 2, Toughness: 1, Colors: []string{"W", "B"}, Keywords: []string{"flying"}},
	"1/1 white and black Inkling with flying":                          {Name: "Inkling", TypeLine: "Token Creature — Inkling", Power: 1, Toughness: 1, Colors: []string{"W", "B"}, Keywords: []string{"flying"}}, // Secrets of Strixhaven (#4456)
	"2/2 black Phyrexian":                                              {Name: "Phyrexian", TypeLine: "Token Creature — Phyrexian", Power: 2, Toughness: 2, Colors: []string{"B"}},
	"2/2 black Rogue":                                                  {Name: "Rogue", TypeLine: "Token Creature — Rogue", Power: 2, Toughness: 2, Colors: []string{"B"}},
	"2/2 colorless Wizard Soldier named Cadet":                         {Name: "Cadet", TypeLine: "Token Creature — Wizard Soldier", Power: 2, Toughness: 2},
	"2/2 black Zombie":                                                 {Name: "Zombie", TypeLine: "Token Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}},
	"2/2 black Phyrexian Zombie":                                       {Name: "Phyrexian Zombie", TypeLine: "Token Creature — Phyrexian Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}},
	"2/2 black Zombie with decayed":                                    {Name: "Zombie", TypeLine: "Token Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}, Keywords: []string{game.KeywordDecayed}},
	"2/2 black Zombie Druid":                                           {Name: "Zombie Druid", TypeLine: "Token Creature — Zombie Druid", Power: 2, Toughness: 2, Colors: []string{"B"}},
	"2/2 black Zombie Mutant":                                          {Name: "Zombie Mutant", TypeLine: "Token Creature — Zombie Mutant", Power: 2, Toughness: 2, Colors: []string{"B"}},
	"2/2 black Zombie Knight with menace":                              {Name: "Zombie Knight", TypeLine: "Token Creature — Zombie Knight", Power: 2, Toughness: 2, Colors: []string{"B"}, Keywords: []string{"menace"}},
	"2/2 black Zombie Berserker":                                       {Name: "Zombie Berserker", TypeLine: "Token Creature — Zombie Berserker", Power: 2, Toughness: 2, Colors: []string{"B"}},
	"2/2 black Zombie with haste named Tombspawn":                      {Name: "Tombspawn", TypeLine: "Token Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}, Keywords: []string{"haste"}},
	"2/2 black Bird with flying":                                       {Name: "Bird", TypeLine: "Token Creature — Bird", Power: 2, Toughness: 2, Colors: []string{"B"}, Keywords: []string{"flying"}},
	"2/2 blue Bird with flying":                                        {Name: "Bird", TypeLine: "Token Creature — Bird", Power: 2, Toughness: 2, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"2/2 blue Drake with flying":                                       {Name: "Drake", TypeLine: "Token Creature — Drake", Power: 2, Toughness: 2, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"2/2 colorless Assembly-Worker artifact":                           {Name: "Assembly-Worker", TypeLine: "Token Artifact Creature — Assembly-Worker", Power: 2, Toughness: 2},
	"2/2 green Monkey":                                                 {Name: "Monkey", TypeLine: "Token Creature — Monkey", Power: 2, Toughness: 2, Colors: []string{"G"}},
	"2/2 green Cat":                                                    {Name: "Cat", TypeLine: "Token Creature — Cat", Power: 2, Toughness: 2, Colors: []string{"G"}},
	"2/2 green Cat with haste":                                         {Name: "Cat", TypeLine: "Token Creature — Cat", Power: 2, Toughness: 2, Colors: []string{"G"}, Keywords: []string{"haste"}},
	"3/1 green Dog with vigilance":                                     {Name: "Dog", TypeLine: "Token Creature — Dog", Power: 3, Toughness: 1, Colors: []string{"G"}, Keywords: []string{"vigilance"}},
	"2/2 colorless Spawn artifact":                                     {Name: "Spawn", TypeLine: "Token Artifact Creature — Spawn", Power: 2, Toughness: 2},
	"2/2 white Alien":                                                  {Name: "Alien", TypeLine: "Token Creature — Alien", Power: 2, Toughness: 2, Colors: []string{"W"}},
	"4/4 green Ox":                                                     {Name: "Ox", TypeLine: "Token Creature — Ox", Power: 4, Toughness: 4, Colors: []string{"G"}},                                         // GAP b230: Contraband Livestock
	"2/2 red and green Satyr with haste":                               {Name: "Satyr", TypeLine: "Token Creature — Satyr", Power: 2, Toughness: 2, Colors: []string{"R", "G"}, Keywords: []string{"haste"}}, // GAP b230: Xenagos, the Reveler
	"2/4 white Ox":                                                     {Name: "Ox", TypeLine: "Token Creature — Ox", Power: 2, Toughness: 4, Colors: []string{"W"}},
	"2/2 white Ox":                                                     {Name: "Ox", TypeLine: "Token Creature — Ox", Power: 2, Toughness: 2, Colors: []string{"W"}},
	"2/2 white Soldier":                                                {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 2, Toughness: 2, Colors: []string{"W"}}, // Signaling Roar (#5175)
	"2/2 white Knight":                                                 {Name: "Knight", TypeLine: "Token Creature — Knight", Power: 2, Toughness: 2, Colors: []string{"W"}},
	"2/2 white Knight Ally":                                            {Name: "Knight Ally", TypeLine: "Token Creature — Knight Ally", Power: 2, Toughness: 2, Colors: []string{"W"}},
	"2/2 white Knight with vigilance":                                  {Name: "Knight", TypeLine: "Token Creature — Knight", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"vigilance"}},
	"2/2 white Fox with vigilance":                                     {Name: "Fox", TypeLine: "Token Creature — Fox", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"vigilance"}},
	"2/2 colorless Robot artifact":                                     {Name: "Robot", TypeLine: "Token Artifact Creature — Robot", Power: 2, Toughness: 2},
	"2/2 colorless Robot Villain artifact":                             {Name: "Robot Villain", TypeLine: "Token Artifact Creature — Robot Villain", Power: 2, Toughness: 2},
	"1/1 colorless Robot Hero artifact with flying":                    {Name: "Robot Hero", TypeLine: "Token Artifact Creature — Robot Hero", Power: 1, Toughness: 1, Keywords: []string{"flying"}},
	"1/1 colorless Borg artifact":                                      {Name: "Borg", TypeLine: "Token Artifact Creature — Borg", Power: 1, Toughness: 1},
	"1/1 red Officer":                                                  {Name: "Officer", TypeLine: "Token Creature — Officer", Power: 1, Toughness: 1, Colors: []string{"R"}},
	"3/3 blue Illusion Villain":                                        {Name: "Illusion Villain", TypeLine: "Token Creature — Illusion Villain", Power: 3, Toughness: 3, Colors: []string{"U"}},
	"2/1 colorless Robot Hero artifact with flying":                    {Name: "Robot Hero", TypeLine: "Token Artifact Creature — Robot Hero", Power: 2, Toughness: 1, Keywords: []string{"flying"}},
	"1/1 white Clown Robot artifact":                                   {Name: "Clown Robot", TypeLine: "Token Artifact Creature — Clown Robot", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"2/2 white Samurai with vigilance":                                 {Name: "Samurai", TypeLine: "Token Creature — Samurai", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"vigilance"}},
	"2/2 white Samurai with double strike":                             {Name: "Samurai", TypeLine: "Token Creature — Samurai", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"double strike"}},
	"2/2 green Boar":                                                   {Name: "Boar", TypeLine: "Token Creature — Boar", Power: 2, Toughness: 2, Colors: []string{"G"}},
	"3/1 green Boar":                                                   {Name: "Boar", TypeLine: "Token Creature — Boar", Power: 3, Toughness: 1, Colors: []string{"G"}},
	"3/3 green Boar":                                                   {Name: "Boar", TypeLine: "Token Creature — Boar", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"2/2 red Human Knight with trample and haste":                      {Name: "Human Knight", TypeLine: "Token Creature — Human Knight", Power: 2, Toughness: 2, Colors: []string{"R"}, Keywords: []string{"trample", "haste"}},
	"2/2 white and blue and black and red and green Elemental":         {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 2, Toughness: 2, Colors: []string{"W", "U", "B", "R", "G"}},
	"2/2 green Elemental":                                              {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 2, Toughness: 2, Colors: []string{"G"}},
	"2/2 green Spider with reach":                                      {Name: "Spider", TypeLine: "Token Creature — Spider", Power: 2, Toughness: 2, Colors: []string{"G"}, Keywords: []string{"reach"}},
	"2/2 green Wolf":                                                   {Name: "Wolf", TypeLine: "Token Creature — Wolf", Power: 2, Toughness: 2, Colors: []string{"G"}},
	"1/1 green Ooze":                                                   {Name: "Ooze", TypeLine: "Token Creature — Ooze", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"2/2 green Ooze":                                                   {Name: "Ooze", TypeLine: "Token Creature — Ooze", Power: 2, Toughness: 2, Colors: []string{"G"}},
	"3/3 green Ooze":                                                   {Name: "Ooze", TypeLine: "Token Creature — Ooze", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"0/0 green Ooze with trample":                                      {Name: "Ooze", TypeLine: "Token Creature — Ooze", Power: 0, Toughness: 0, Colors: []string{"G"}, Keywords: []string{"trample"}, PrintedPTKnown: true},
	"1/1 white Pegasus with flying":                                    {Name: "Pegasus", TypeLine: "Token Creature — Pegasus", Power: 1, Toughness: 1, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"2/2 white Pegasus with flying":                                    {Name: "Pegasus", TypeLine: "Token Creature — Pegasus", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"2/2 white Soldier with vigilance":                                 {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"vigilance"}},
	"2/1 black Skeleton":                                               {Name: "Skeleton", TypeLine: "Token Creature — Skeleton", Power: 2, Toughness: 1, Colors: []string{"B"}},
	"2/1 black Warrior":                                                {Name: "Warrior", TypeLine: "Token Creature — Warrior", Power: 2, Toughness: 1, Colors: []string{"B"}},
	"1/1 red Pirate with menace and haste":                             {Name: "Pirate", TypeLine: "Token Creature — Pirate", Power: 1, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"menace", "haste"}},
	"2/2 black Pirate with menace":                                     {Name: "Pirate", TypeLine: "Token Creature — Pirate", Power: 2, Toughness: 2, Colors: []string{"B"}, Keywords: []string{"menace"}},
	"2/2 black Skeleton Pirate":                                        {Name: "Skeleton Pirate", TypeLine: "Token Creature — Skeleton Pirate", Power: 2, Toughness: 2, Colors: []string{"B"}},
	"2/2 white and blue Knight with vigilance":                         {Name: "Knight", TypeLine: "Token Creature — Knight", Power: 2, Toughness: 2, Colors: []string{"W", "U"}, Keywords: []string{"vigilance"}},
	"2/2 white and blue Detective":                                     {Name: "Detective", TypeLine: "Token Creature — Detective", Power: 2, Toughness: 2, Colors: []string{"W", "U"}},
	"3/1 red Dinosaur":                                                 {Name: "Dinosaur", TypeLine: "Token Creature — Dinosaur", Power: 3, Toughness: 1, Colors: []string{"R"}},
	"1/1 red Dinosaur with haste":                                      {Name: "Dinosaur", TypeLine: "Token Creature — Dinosaur", Power: 1, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"haste"}}, // #5512: Forbidden Friendship
	"1/1 white Kithkin Soldier":                                        {Name: "Kithkin Soldier", TypeLine: "Token Creature — Kithkin Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 black and green Worm":                                         {Name: "Worm", TypeLine: "Token Creature — Worm", Power: 1, Toughness: 1, Colors: []string{"B", "G"}},
	"1/1 black and red Goblin":                                         {Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1, Colors: []string{"B", "R"}},
	"5/5 blue and red Elemental with flying":                           {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 5, Toughness: 5, Colors: []string{"U", "R"}, Keywords: []string{"flying"}},
	"3/3 blue and red Elemental with flying":                           {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 3, Toughness: 3, Colors: []string{"U", "R"}, Keywords: []string{"flying"}},
	"3/3 green Ape":                                                    {Name: "Ape", TypeLine: "Token Creature — Ape", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"3/3 green Beast":                                                  {Name: "Beast", TypeLine: "Token Creature — Beast", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"2/2 colorless Spirit":                                             {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 2, Toughness: 2},
	"3/3 black Beast with deathtouch":                                  {Name: "Beast", TypeLine: "Token Creature — Beast", Power: 3, Toughness: 3, Colors: []string{"B"}, Keywords: []string{"deathtouch"}},
	"3/3 green Raccoon":                                                {Name: "Raccoon", TypeLine: "Token Creature — Raccoon", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"3/3 green Elk":                                                    {Name: "Elk", TypeLine: "Token Creature — Elk", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"3/2 red Wolf":                                                     {Name: "Wolf", TypeLine: "Token Creature — Wolf", Power: 3, Toughness: 2, Colors: []string{"R"}},
	"3/2 colorless Shapeshifter with changeling":                       {Name: "Shapeshifter", TypeLine: "Token Creature — Shapeshifter", Power: 3, Toughness: 2, Keywords: []string{"changeling"}},
	"3/2 white Hero with vigilance":                                    {Name: "Hero", TypeLine: "Token Creature — Hero", Power: 3, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"vigilance"}},
	"3/3 colorless Robot Villain artifact named Doombot":               {Name: "Doombot", TypeLine: "Token Artifact Creature — Robot Villain", Power: 3, Toughness: 3},
	"3/2 colorless Eldrazi Horror":                                     {Name: "Eldrazi Horror", TypeLine: "Token Creature — Eldrazi Horror", Power: 3, Toughness: 2},
	"4/4 colorless Eldrazi Angel with flying and vigilance":            {Name: "Eldrazi Angel", TypeLine: "Token Creature — Eldrazi Angel", Power: 4, Toughness: 4, Keywords: []string{"flying", "vigilance"}},
	"3/2 red and white Spirit":                                         {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 3, Toughness: 2, Colors: []string{"R", "W"}},
	"2/2 red Rebel":                                                    {Name: "Rebel", TypeLine: "Token Creature — Rebel", Power: 2, Toughness: 2, Colors: []string{"R"}},
	"2/2 red and white Spirit":                                         {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 2, Toughness: 2, Colors: []string{"R", "W"}},
	"4/4 green Fungus Beast with trample":                              {Name: "Fungus Beast", TypeLine: "Token Creature — Fungus Beast", Power: 4, Toughness: 4, Colors: []string{"G"}, Keywords: []string{"trample"}},
	"1/1 colorless Shapeshifter with changeling":                       {Name: "Shapeshifter", TypeLine: "Token Creature — Shapeshifter", Power: 1, Toughness: 1, Keywords: []string{"changeling"}},
	"3/3 green Centaur":                                                {Name: "Centaur", TypeLine: "Token Creature — Centaur", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"3/3 green Centaur with protection from black":                     {Name: "Centaur", TypeLine: "Token Creature — Centaur", Power: 3, Toughness: 3, Colors: []string{"G"}, Keywords: []string{"protection from black"}},
	"4/4 green Rhino with trample":                                     {Name: "Rhino", TypeLine: "Token Creature — Rhino", Power: 4, Toughness: 4, Colors: []string{"G"}, Keywords: []string{"trample"}},
	"4/4 green Rhino Warrior":                                          {Name: "Rhino Warrior", TypeLine: "Token Creature — Rhino Warrior", Power: 4, Toughness: 4, Colors: []string{"G"}},
	"3/3 green Frog Lizard":                                            {Name: "Frog Lizard", TypeLine: "Token Creature — Frog Lizard", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"3/3 colorless Golem artifact":                                     {Name: "Golem", TypeLine: "Token Enchantment Artifact Creature — Golem", Power: 3, Toughness: 3},
	"4/4 colorless Construct artifact":                                 {Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 4, Toughness: 4},
	"4/4 colorless Construct artifact with flying and haste":           {Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 4, Toughness: 4, Keywords: []string{"flying", "haste"}},
	"3/3 colorless Golem artifact creature":                            {Name: "Golem", TypeLine: "Token Artifact Creature — Golem", Power: 3, Toughness: 3},
	"4/4 colorless Golem artifact creature":                            {Name: "Golem", TypeLine: "Token Artifact Creature — Golem", Power: 4, Toughness: 4},
	"9/9 colorless Golem artifact creature":                            {Name: "Golem", TypeLine: "Token Artifact Creature — Golem", Power: 9, Toughness: 9},
	"1/1 black Insect with flying":                                     {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"flying"}},
	"4/4 white and blue Golem artifact":                                {Name: "Golem", TypeLine: "Token Artifact Creature — Golem", Power: 4, Toughness: 4, Colors: []string{"W", "U"}},
	"4/4 red and white Golem artifact":                                 {Name: "Golem", TypeLine: "Token Artifact Creature — Golem", Power: 4, Toughness: 4, Colors: []string{"R", "W"}},
	"3/3 colorless Phyrexian Golem artifact":                           {Name: "Phyrexian Golem", TypeLine: "Token Artifact Creature — Phyrexian Golem", Power: 3, Toughness: 3},
	"3/3 colorless Phyrexian Wurm artifact":                            {Name: "Phyrexian Wurm", TypeLine: "Token Artifact Creature — Phyrexian Wurm", Power: 3, Toughness: 3},
	"3/3 green Dinosaur with trample":                                  {Name: "Dinosaur", TypeLine: "Token Creature — Dinosaur", Power: 3, Toughness: 3, Colors: []string{"G"}, Keywords: []string{"trample"}},
	"0/1 green Dinosaur Egg":                                           {Name: "Dinosaur Egg", TypeLine: "Token Creature — Dinosaur Egg", Power: 0, Toughness: 1, Colors: []string{"G"}},
	"3/3 green Dinosaur":                                               {Name: "Dinosaur", TypeLine: "Token Creature — Dinosaur", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"3/3 green Elephant":                                               {Name: "Elephant", TypeLine: "Token Creature — Elephant", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"3/4 green Treefolk with reach":                                    {Name: "Treefolk", TypeLine: "Token Creature — Treefolk", Power: 3, Toughness: 4, Colors: []string{"G"}, Keywords: []string{"reach"}},
	"3/3 red Ogre":                                                     {Name: "Ogre", TypeLine: "Token Creature — Ogre", Power: 3, Toughness: 3, Colors: []string{"R"}},
	"4/3 black Ogre Warrior":                                           {Name: "Ogre Warrior", TypeLine: "Token Creature — Ogre Warrior", Power: 4, Toughness: 3, Colors: []string{"B"}},
	"3/3 black Dalek artifact with menace":                             {Name: "Dalek", TypeLine: "Token Artifact Creature — Dalek", Power: 3, Toughness: 3, Colors: []string{"B"}, Keywords: []string{"menace"}},
	"3/3 black Wraith with menace":                                     {Name: "Wraith", TypeLine: "Token Creature — Wraith", Power: 3, Toughness: 3, Colors: []string{"B"}, Keywords: []string{"menace"}},
	"4/4 black Zombie Warrior with vigilance":                          {Name: "Zombie Warrior", TypeLine: "Token Creature — Zombie Warrior", Power: 4, Toughness: 4, Colors: []string{"B"}, Keywords: []string{"vigilance"}},
	"2/1 black Villain with menace":                                    {Name: "Villain", TypeLine: "Token Creature — Villain", Power: 2, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"menace"}},
	"1/1 colorless Phyrexian Myr artifact":                             {Name: "Phyrexian Myr", TypeLine: "Token Artifact Creature — Phyrexian Myr", Power: 1, Toughness: 1},
	"2/1 blue Phyrexian Myr artifact":                                  {Name: "Phyrexian Myr", TypeLine: "Token Artifact Creature — Phyrexian Myr", Power: 2, Toughness: 1, Colors: []string{"U"}},
	"1/1 black Skeleton":                                               {Name: "Skeleton", TypeLine: "Token Creature — Skeleton", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"4/1 black Skeleton with menace":                                   {Name: "Skeleton", TypeLine: "Token Creature — Skeleton", Power: 4, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"menace"}},
	"4/4 white and blue Elemental":                                     {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 4, Toughness: 4, Colors: []string{"W", "U"}}, // #4471: Hylda of the Icy Crown
	"4/4 blue and red Elemental":                                       {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 4, Toughness: 4, Colors: []string{"U", "R"}},
	"4/4 white Elemental with flying":                                  {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 4, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"flying"}}, // cmdr2 S22: Hoofprints of the Stag
	"4/4 blue Giant Wizard":                                            {Name: "Giant Wizard", TypeLine: "Token Creature — Giant Wizard", Power: 4, Toughness: 4, Colors: []string{"U"}},
	"4/4 red Giant":                                                    {Name: "Giant", TypeLine: "Token Creature — Giant", Power: 4, Toughness: 4, Colors: []string{"R"}},                                 // #5594: Pact of the Titan
	"4/4 blue Sphinx with flying":                                      {Name: "Sphinx", TypeLine: "Token Creature — Sphinx", Power: 4, Toughness: 4, Colors: []string{"U"}, Keywords: []string{"flying"}}, // S82 OLD batch 366: Hour of Need
	"2/2 green Bear":                                                   {Name: "Bear", TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}},
	"4/4 green Bear":                                                   {Name: "Bear", TypeLine: "Token Creature — Bear", Power: 4, Toughness: 4, Colors: []string{"G"}},
	"4/4 blue Elemental Bird with flying":                              {Name: "Elemental Bird", TypeLine: "Token Creature — Elemental Bird", Power: 4, Toughness: 4, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"4/4 white Bird Soldier with flying":                               {Name: "Bird Soldier", TypeLine: "Token Creature — Bird Soldier", Power: 4, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"4/4 green Beast":                                                  {Name: "Beast", TypeLine: "Token Creature — Beast", Power: 4, Toughness: 4, Colors: []string{"G"}},
	"4/4 green Beast with trample":                                     {Name: "Beast", TypeLine: "Token Creature — Beast", Power: 4, Toughness: 4, Colors: []string{"G"}, Keywords: []string{"trample"}},
	"5/5 green Beast with trample":                                     {Name: "Beast", TypeLine: "Token Creature — Beast", Power: 5, Toughness: 5, Colors: []string{"G"}, Keywords: []string{"trample"}}, // Spawning Grounds (C18)
	"4/4 green Phyrexian Beast":                                        {Name: "Phyrexian Beast", TypeLine: "Token Creature — Phyrexian Beast", Power: 4, Toughness: 4, Colors: []string{"G"}},
	"4/4 red Dragon with flying":                                       {Name: "Dragon", TypeLine: "Token Creature — Dragon", Power: 4, Toughness: 4, Colors: []string{"R"}, Keywords: []string{"flying"}},
	"4/4 red Bird with flying":                                         {Name: "Bird", TypeLine: "Token Creature — Bird", Power: 4, Toughness: 4, Colors: []string{"R"}, Keywords: []string{"flying"}}, // S82 older sets batch 620: Rukh Egg
	"4/4 red Dinosaur Dragon with flying":                              {Name: "Dinosaur Dragon", TypeLine: "Token Creature — Dinosaur Dragon", Power: 4, Toughness: 4, Colors: []string{"R"}, Keywords: []string{"flying"}},
	"5/5 red Dragon with flying":                                       {Name: "Dragon", TypeLine: "Token Creature — Dragon", Power: 5, Toughness: 5, Colors: []string{"R"}, Keywords: []string{"flying"}},
	"4/4 red Scorpion Dragon with flying and haste":                    {Name: "Scorpion Dragon", TypeLine: "Token Creature — Scorpion Dragon", Power: 4, Toughness: 4, Colors: []string{"R"}, Keywords: []string{"flying", "haste"}},
	"3/3 blue Shark":                                                   {Name: "Shark", TypeLine: "Token Creature — Shark", Power: 3, Toughness: 3, Colors: []string{"U"}},
	"3/3 blue Angel with flying":                                       {Name: "Angel", TypeLine: "Token Creature — Angel", Power: 3, Toughness: 3, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"3/3 green Legendary Dog named Mowu":                               {Name: "Mowu", TypeLine: "Token Legendary Creature — Dog", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"1/1 green Legendary Dog named Porthos":                            {Name: "Porthos", TypeLine: "Token Legendary Creature — Dog", Power: 1, Toughness: 1, Colors: []string{"G"}},
	"3/2 red Legendary Horse named Phobos":                             {Name: "Phobos", TypeLine: "Token Legendary Creature — Horse", Power: 3, Toughness: 2, Colors: []string{"R"}},
	"1/1 green and white Legendary Dog named Angelo":                   {Name: "Angelo", TypeLine: "Token Legendary Creature — Dog", Power: 1, Toughness: 1, Colors: []string{"G", "W"}},
	"2/2 green Legendary Dog Detective named Tiny with trample":        {Name: "Tiny", TypeLine: "Token Legendary Creature — Dog Detective", Power: 2, Toughness: 2, Colors: []string{"G"}, Keywords: []string{"trample"}},
	"8/8 black Legendary Zombie God named Vecna with indestructible":   {Name: "Vecna", TypeLine: "Token Legendary Creature — Zombie God", Power: 8, Toughness: 8, Colors: []string{"B"}, Keywords: []string{"indestructible"}},
	"2/2 white and black Legendary Dog named Darkstar":                 {Name: "Darkstar", TypeLine: "Token Legendary Creature — Dog", Power: 2, Toughness: 2, Colors: []string{"W", "B"}},
	"3/3 colorless Illusion artifact":                                  {Name: "Illusion", TypeLine: "Token Artifact Creature — Illusion", Power: 3, Toughness: 3},
	"3/3 white Angel with flying":                                      {Name: "Angel", TypeLine: "Token Creature — Angel", Power: 3, Toughness: 3, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"3/3 white Knight":                                                 {Name: "Knight", TypeLine: "Token Creature — Knight", Power: 3, Toughness: 3, Colors: []string{"W"}},
	"4/4 white Angel with flying":                                      {Name: "Angel", TypeLine: "Token Creature — Angel", Power: 4, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"4/4 white Angel Warrior with flying":                              {Name: "Angel Warrior", TypeLine: "Token Creature — Angel Warrior", Power: 4, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"4/4 white Angel Warrior with flying and vigilance":                {Name: "Angel Warrior", TypeLine: "Token Creature — Angel Warrior", Power: 4, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"flying", "vigilance"}},
	"4/4 white Avatar with flying":                                     {Name: "Avatar", TypeLine: "Token Creature — Avatar", Power: 4, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"flying"}},
	"4/4 white Angel with flying and vigilance":                        {Name: "Angel", TypeLine: "Token Creature — Angel", Power: 4, Toughness: 4, Colors: []string{"W"}, Keywords: []string{"flying", "vigilance"}},
	"4/2 green Plant Warrior with reach":                               {Name: "Plant Warrior", TypeLine: "Token Creature — Plant Warrior", Power: 4, Toughness: 2, Colors: []string{"G"}, Keywords: []string{"reach"}},
	"1/1 red and green and white Sand Warrior":                         {Name: "Sand Warrior", TypeLine: "Token Creature — Sand Warrior", Power: 1, Toughness: 1, Colors: []string{"R", "G", "W"}},
	"5/3 green Elemental":                                              {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 5, Toughness: 3, Colors: []string{"G"}},
	"5/5 red Dragon Spirit with flying":                                {Name: "Dragon Spirit", TypeLine: "Token Creature — Dragon Spirit", Power: 5, Toughness: 5, Colors: []string{"R"}, Keywords: []string{"flying"}},
	"5/5 white Horse":                                                  {Name: "Horse", TypeLine: "Token Creature — Horse", Power: 5, Toughness: 5, Colors: []string{"W"}},
	"5/5 green Elephant":                                               {Name: "Elephant", TypeLine: "Token Creature — Elephant", Power: 5, Toughness: 5, Colors: []string{"G"}},
	"5/5 red and green Elemental":                                      {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 5, Toughness: 5, Colors: []string{"R", "G"}},
	"4/4 red and green Beast with trample":                             {Name: "Beast", TypeLine: "Token Creature — Beast", Power: 4, Toughness: 4, Colors: []string{"R", "G"}, Keywords: []string{"trample"}},
	"8/8 blue Leviathan with hexproof":                                 {Name: "Leviathan", TypeLine: "Token Creature — Leviathan", Power: 8, Toughness: 8, Colors: []string{"U"}, Keywords: []string{"hexproof"}},
	"6/5 blue Leviathan with hexproof":                                 {Name: "Leviathan", TypeLine: "Token Creature — Leviathan", Power: 6, Toughness: 5, Colors: []string{"U"}, Keywords: []string{"hexproof"}},
	"8/8 blue Octopus":                                                 {Name: "Octopus", TypeLine: "Token Creature — Octopus", Power: 8, Toughness: 8, Colors: []string{"U"}},
	"8/8 blue Scion of the Deep":                                       {Name: "Scion of the Deep", TypeLine: "Token Legendary Creature — Octopus", Power: 8, Toughness: 8, Colors: []string{"U"}},
	"8/8 red Lizard":                                                   {Name: "Lizard", TypeLine: "Token Creature — Lizard", Power: 8, Toughness: 8, Colors: []string{"R"}},
	"9/9 blue Kraken":                                                  {Name: "Kraken", TypeLine: "Token Creature — Kraken", Power: 9, Toughness: 9, Colors: []string{"U"}},
	"8/8 blue Kraken":                                                  {Name: "Kraken", TypeLine: "Token Creature — Kraken", Power: 8, Toughness: 8, Colors: []string{"U"}},
	"4/4 black Legendary God Horror named The Atropal with deathtouch": {Name: "The Atropal", TypeLine: "Token Legendary Creature — God Horror", Power: 4, Toughness: 4, Colors: []string{"B"}, Keywords: []string{"deathtouch"}},
	"5/5 black Legendary Demon Spirit named Urami with flying":         {Name: "Urami", TypeLine: "Token Legendary Creature — Demon Spirit", Power: 5, Toughness: 5, Colors: []string{"B"}, Keywords: []string{"flying"}},
	"20/20 black Marit Lage":                                           {Name: "Marit Lage", TypeLine: "Token Legendary Creature — Avatar", Power: 20, Toughness: 20, Colors: []string{"B"}, Keywords: []string{"flying", "indestructible"}},

	// S82 MID batch 3: Hungry for More's Vampire and Rise of the Ants's Insect.
	"3/1 black and red Vampire with trample, lifelink, and haste": {Name: "Vampire", TypeLine: "Token Creature — Vampire", Power: 3, Toughness: 1, Colors: []string{"B", "R"}, Keywords: []string{"trample", "lifelink", "haste"}},
	"3/3 green Insect": {Name: "Insect", TypeLine: "Token Creature — Insect", Power: 3, Toughness: 3, Colors: []string{"G"}},

	// A key longer than every row above: kept in its own block so gofmt
	// does not re-align the whole table (and every open PR's rows) for it.
	"6/12 colorless Construct artifact with trample": {Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 6, Toughness: 12, Keywords: []string{"trample"}},

	// Same reason, and longer again (#706: Elemental Eruption).
	"4/4 red Dragon Elemental with flying and prowess": {Name: "Dragon Elemental", TypeLine: "Token Creature — Dragon Elemental", Power: 4, Toughness: 4, Colors: []string{"R"}, Keywords: []string{"flying", "prowess"}},

	// Same reason (ADR 0102: Abby, Merciless Soldier). A named token:
	// the key says the name the card prints.
	"1/1 black Fungus Zombie named Cordyceps Infected": {Name: "Cordyceps Infected", TypeLine: "Token Creature — Fungus Zombie", Power: 1, Toughness: 1, Colors: []string{"B"}},

	// ADR 0103 (Rooms): the tokens Dollmaker's Shop, Grand Entryway and
	// Ritual Chamber make.
	"1/1 white Toy artifact":        {Name: "Toy", TypeLine: "Token Artifact Creature — Toy", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"1/1 white Glimmer enchantment": {Name: "Glimmer", TypeLine: "Token Enchantment Creature — Glimmer", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"6/6 black Demon with flying":   {Name: "Demon", TypeLine: "Token Creature — Demon", Power: 6, Toughness: 6, Colors: []string{"B"}, Keywords: []string{"flying"}},

	// #2466: its own alignment block, so a long key does not re-flow the table above.
	"2/2 red and white Dinosaur with flying and haste": {Name: "Dinosaur", TypeLine: "Token Creature — Dinosaur", Power: 2, Toughness: 2, Colors: []string{"R", "W"}, Keywords: []string{"flying", "haste"}},

	// S82 OLD batch 554: Hearthcage Giant, Rebellion of the Flamekin.
	"3/1 red Elemental Shaman": {Name: "Elemental Shaman", TypeLine: "Token Creature — Elemental Shaman", Power: 3, Toughness: 1, Colors: []string{"R"}},

	// #2038: Hostility's and Worldspine Wurm's tokens.
	"3/1 red Elemental Shaman with haste": {Name: "Elemental Shaman", TypeLine: "Token Creature — Elemental Shaman", Power: 3, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"haste"}},
	"5/5 green Wurm with trample":         {Name: "Wurm", TypeLine: "Token Creature — Wurm", Power: 5, Toughness: 5, Colors: []string{"G"}, Keywords: []string{"trample"}},

	// #4457: Teenage Mutant Ninja Turtles' Robot, Mutant and Dinosaur Soldier.
	"1/1 colorless Robot artifact": {Name: "Robot", TypeLine: "Token Artifact Creature — Robot", Power: 1, Toughness: 1},
	"2/2 red Mutant":               {Name: "Mutant", TypeLine: "Token Creature — Mutant", Power: 2, Toughness: 2, Colors: []string{"R"}},
	"2/2 white Dinosaur Soldier":   {Name: "Dinosaur Soldier", TypeLine: "Token Creature — Dinosaur Soldier", Power: 2, Toughness: 2, Colors: []string{"W"}},

	// #5368: Amphin Mutineer's Salamander Warrior (Tarkir: Dragonstorm Commander).
	"4/3 blue Salamander Warrior": {Name: "Salamander Warrior", TypeLine: "Token Creature — Salamander Warrior", Power: 4, Toughness: 3, Colors: []string{"U"}},

	// #5368: Verix Bladewing's Karox Bladewing (Tarkir: Dragonstorm Commander).
	"4/4 red Legendary Dragon named Karox Bladewing with flying": {Name: "Karox Bladewing", TypeLine: "Token Legendary Creature — Dragon", Power: 4, Toughness: 4, Colors: []string{"R"}, Keywords: []string{"flying"}},

	// #4457 batch 12: The Last Ronin's Technique's Ninja Turtle Spirit.
	"1/1 white Ninja Turtle Spirit": {Name: "Ninja Turtle Spirit", TypeLine: "Token Creature — Ninja Turtle Spirit", Power: 1, Toughness: 1, Colors: []string{"W"}},
	// #4457: Lord Dregg, Insect Invader's token.
	"1/1 black Insect Warrior with flying": {Name: "Insect Warrior", TypeLine: "Token Creature — Insect Warrior", Power: 1, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"flying"}},
	// #4457 batch 3: Foot Mystic's Ninja.
	"1/1 black Ninja": {Name: "Ninja", TypeLine: "Token Creature — Ninja", Power: 1, Toughness: 1, Colors: []string{"B"}},
	// #4973 batch 5: Star Trek's Officer (U.S.S. Enterprise-A, Free Spirit).

	// #4973 batch 10: Star Trek's Klingon Warrior (Cha'DIch Investigator).
	"2/1 red Klingon Warrior with haste": {Name: "Klingon Warrior", TypeLine: "Token Creature — Klingon Warrior", Power: 2, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"haste"}},

	// #4973: Star Trek's Officer (Excalbian Gamemaster).

	// #4973 batch 10: Star Trek's Klingon Warrior (Cha'DIch Investigator).

	// #4973 batch 5: Star Trek's Officer (U.S.S. Enterprise-A, Free Spirit).

	// #4973: Star Trek's Officer (Excalbian Gamemaster).

	// #4973 batch 5: Star Trek's Officer (U.S.S. Enterprise-A, Free Spirit).
	// #4973 batch 5: Star Trek's Romulan (Cloak and Phaser).
	"2/2 black Romulan": {Name: "Romulan", TypeLine: "Token Creature — Romulan", Power: 2, Toughness: 2, Colors: []string{"B"}},
	// #4973 batch 21: Star Trek's Illusion (Hologram Stalker) is the
	// "3/3 colorless Illusion artifact" row above.

	// #4973 batch 21: Star Trek's Illusion (Hologram Stalker).

	// #4973 batch 21: Star Trek's Illusion (Hologram Stalker) uses the
	// "3/3 colorless Illusion artifact" row batch 19 added above.
	// #4973 batch 21: Star Trek's Murf (Rok-Tahk, Budding Scientist).
	"0/1 green Legendary Ooze Slug named Murf with indestructible": {Name: "Murf", TypeLine: "Token Legendary Creature — Ooze Slug", Power: 0, Toughness: 1, Colors: []string{"G"}, Keywords: []string{"indestructible"}},
	// #4974 batch 43: Star Trek Commander's Spacecraft Illusion (U.S.S. Stargazer, Here and There).
	"4/3 colorless Spacecraft Illusion artifact with flying": {Name: "Spacecraft Illusion", TypeLine: "Token Artifact Creature — Spacecraft Illusion", Power: 4, Toughness: 3, Keywords: []string{"flying"}},

	// #4459 batch 8: Avatar's Ballistic Boulder (Fire Navy Trebuchet, TTLA 13).
	"2/1 colorless Construct artifact with flying named Ballistic Boulder": {Name: "Ballistic Boulder", TypeLine: "Token Artifact Creature — Construct", Power: 2, Toughness: 1, Keywords: []string{"flying"}},

	// #5075 batch 34: Avatar Eternal's plain Soldier (Fire Nation Archers, TTLE).

	// #5368: Lost Caverns of Ixalan Commander's white and black Vampire (Glass-Cast Heart).
	"1/1 white and black Vampire with lifelink": {Name: "Vampire", TypeLine: "Token Creature — Vampire", Power: 1, Toughness: 1, Colors: []string{"W", "B"}, Keywords: []string{"lifelink"}},

	// #4470: The Lost Caverns of Ixalan's Vampire Demon (Vito, Fanatic of Aclazotz).
	"4/3 white and black Vampire Demon with flying": {Name: "Vampire Demon", TypeLine: "Token Creature — Vampire Demon", Power: 4, Toughness: 3, Colors: []string{"W", "B"}, Keywords: []string{"flying"}},

	// #5368: Bloomburrow Commander's Storm Crow Bird (Murmuration).
	"1/2 blue Bird with flying named Storm Crow": {Name: "Storm Crow", TypeLine: "Token Creature — Bird", Power: 1, Toughness: 2, Colors: []string{"U"}, Keywords: []string{"flying"}},

	// #4467: Bloomburrow's Otter (Stormchaser's Talent).
	"1/1 blue and red Otter with prowess": {Name: "Otter", TypeLine: "Token Creature — Otter", Power: 1, Toughness: 1, Colors: []string{"U", "R"}, Keywords: []string{"prowess"}},

	// #4462: Final Fantasy's Robot Warrior (Retrieve the Esper) and Horror (The Final Days).
	"3/3 blue Robot Warrior artifact": {Name: "Robot Warrior", TypeLine: "Token Artifact Creature — Robot Warrior", Power: 3, Toughness: 3, Colors: []string{"U"}},
	"2/2 black Horror":                {Name: "Horror", TypeLine: "Token Creature — Horror", Power: 2, Toughness: 2, Colors: []string{"B"}},

	// #5600: Ravnica: City of Guilds's Hunted Dragon Knights and Hunted Lammasu Horror.
	"2/2 white Knight with first strike": {Name: "Knight", TypeLine: "Token Creature — Knight", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"first strike"}},
	"4/4 black Horror":                   {Name: "Horror", TypeLine: "Token Creature — Horror", Power: 4, Toughness: 4, Colors: []string{"B"}},

	// #5368: Duskmourn's Horror (Phenomenon Investigators) and Demon (Demonic Covenant).
	"2/2 black Horror enchantment": {Name: "Horror", TypeLine: "Token Enchantment Creature — Horror", Power: 2, Toughness: 2, Colors: []string{"B"}},
	"5/5 black Demon with flying":  {Name: "Demon", TypeLine: "Token Creature — Demon", Power: 5, Toughness: 5, Colors: []string{"B"}, Keywords: []string{"flying"}},

	// #4465: Foundations' Ninja (Kaito, Cunning Infiltrator).
	"2/1 blue Ninja": {Name: "Ninja", TypeLine: "Token Creature — Ninja", Power: 2, Toughness: 1, Colors: []string{"U"}},
	// #5368: Tarkir: Dragonstorm Commander's First Mate Ragavan (Baral and Kari Zev).
	// #5368: Lost Caverns of Ixalan Commander's Ragavan (Kari Zev, Skyship Raider).
	"2/1 red Legendary Monkey named Ragavan":                   {Name: "Ragavan", TypeLine: "Token Legendary Creature — Monkey", Power: 2, Toughness: 1, Colors: []string{"R"}},
	"2/1 red Legendary Monkey Pirate named First Mate Ragavan": {Name: "First Mate Ragavan", TypeLine: "Token Legendary Creature — Monkey Pirate", Power: 2, Toughness: 1, Colors: []string{"R"}},

	// Venser, Corpse Puppet (#5653): The Hollow Sentinel.
	"3/3 colorless Legendary Phyrexian Golem artifact named The Hollow Sentinel": {Name: "The Hollow Sentinel", TypeLine: "Token Legendary Artifact Creature — Phyrexian Golem", Power: 3, Toughness: 3},

	// #5647: Saint Traft and Rem Karolus, Invasion of Segovia.
	"1/1 blue Spirit with flying":  {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying"}},
	"1/1 blue Kraken with trample": {Name: "Kraken", TypeLine: "Token Creature — Kraken", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"trample"}},

	// #5637: Fallen Empires' Citizen (Icatian Town).
	"1/1 white Citizen": {Name: "Citizen", TypeLine: "Token Creature — Citizen", Power: 1, Toughness: 1, Colors: []string{"W"}},

	// #5789: Odyssey block's Wurm (Roar of the Wurm, Crush of Wurms).
	"6/6 green Wurm": {Name: "Wurm", TypeLine: "Token Creature — Wurm", Power: 6, Toughness: 6, Colors: []string{"G"}},

	// S82 ONE batch 16: Wurmquake's X/X Phyrexian Wurm (printed X/X; a 0/0 placeholder, set to
	// the mana spent when it is created).
	"0/0 green Phyrexian Wurm with trample and toxic 1": {Name: "Phyrexian Wurm", TypeLine: "Token Creature — Phyrexian Wurm", Colors: []string{"G"}, Keywords: []string{"trample", "toxic 1"}, PrintedPTKnown: true},
	// #5479: Phyrexia: All Will Be One's Beast (Goliath Hatchery, Viral Spawning, Lukka, Bound to Ruin).
	"3/3 green Phyrexian Beast with toxic 1": {Name: "Phyrexian Beast", TypeLine: "Token Creature — Phyrexian Beast", Power: 3, Toughness: 3, Colors: []string{"G"}, Keywords: []string{"toxic 1"}},

	// S82 NEO batch 20: Kura, the Boundless Sky's X/X Spirit (printed */*; a 0/0 placeholder,
	// set to the number of lands as it is created).
	"0/0 green Spirit": {Name: "Spirit", TypeLine: "Token Creature — Spirit", Colors: []string{"G"}, PrintedPTKnown: true},
	"0/0 green Hydra":  {Name: "Hydra", TypeLine: "Token Creature — Hydra", Colors: []string{"G"}, PrintedPTKnown: true},

	// S82 OLD batch 097: Grakmaw, Skyclave Ravager's X/X black and green Hydra (printed */*; a 0/0
	// placeholder, set to the +1/+1 counters Grakmaw had as it created it).
	"0/0 black and green Hydra": {Name: "Hydra", TypeLine: "Token Creature — Hydra", Colors: []string{"B", "G"}, PrintedPTKnown: true},

	// S82 OLD batch 207: Rite of Belzenlok's Cleric, and Chainer's Torment's X/X Nightmare
	// Horror (printed */*; a 0/0 placeholder, set to half the controller's life as it is created).
	"0/1 black Cleric":           {Name: "Cleric", TypeLine: "Token Creature — Cleric", Power: 0, Toughness: 1, Colors: []string{"B"}},
	"0/0 black Nightmare Horror": {Name: "Nightmare Horror", TypeLine: "Token Creature — Nightmare Horror", Colors: []string{"B"}, PrintedPTKnown: true},

	// S82 RPR batch 32: Stitcher Geralf's X/X blue Zombie (printed */*; a 0/0 placeholder,
	// set to the total power exiled as it is created).
	"0/0 blue Zombie": {Name: "Zombie", TypeLine: "Token Creature — Zombie", Colors: []string{"U"}, PrintedPTKnown: true},

	// S82 OLD batch 667: Pure Reflection's X/X white Reflection (no token card was printed; a 0/0
	// placeholder, set to the mana value of the spell as it is created).
	"0/0 white Reflection": {Name: "Reflection", TypeLine: "Token Creature — Reflection", Colors: []string{"W"}, PrintedPTKnown: true},

	// S82 OLD batch 85: Marath, Will of the Wild's X/X green Elemental (printed */*; a 0/0
	// placeholder, set to X as it is created).
	"0/0 green Elemental": {Name: "Elemental", TypeLine: "Token Creature — Elemental", Colors: []string{"G"}, PrintedPTKnown: true},

	// S82 OLD batch 350: Kin-Tree Invocation's X/X black and green Spirit Warrior (printed */*; a
	// 0/0 placeholder, set to the greatest toughness among your creatures as it is created).
	"0/0 black and green Spirit Warrior": {Name: "Spirit Warrior", TypeLine: "Token Creature — Spirit Warrior", Colors: []string{"B", "G"}, PrintedPTKnown: true},

	// S82 OLD batch 414 (Gatecrash): Ooze Flux's and Mystic Genesis's X/X green Ooze (printed */*; a
	// 0/0 placeholder, set to X as it is created). Beckon Apparition's Spirit is the std_b43 row.
	// S82 OLD batch 411: Miming Slime's X/X green Ooze (printed */*; a 0/0 placeholder, set to the
	// greatest power among your creatures as it is created).
	"0/0 green Ooze": {Name: "Ooze", TypeLine: "Token Creature — Ooze", Colors: []string{"G"}, PrintedPTKnown: true},

	// S82 DMU batch 14: Yotia Declares War's Ornithopter.
	"0/2 colorless Thopter artifact with flying named Ornithopter": {Name: "Ornithopter", TypeLine: "Token Artifact Creature — Thopter", Power: 0, Toughness: 2, Keywords: []string{"flying"}},

	// Jared Carthalion (DMU, S82 batch 15): the printed all-colors Kavu.
	"3/3 white and blue and black and red and green Kavu with trample": {Name: "Kavu", TypeLine: "Token Creature — Kavu", Power: 3, Toughness: 3, Colors: []string{"W", "U", "B", "R", "G"}, Keywords: []string{"trample"}},

	// #5487: Dominaria United Commander's Stangg Twin (Stangg, Echo Warrior).
	"3/4 red and green Legendary Human Warrior named Stangg Twin": {Name: "Stangg Twin", TypeLine: "Token Legendary Creature — Human Warrior", Power: 3, Toughness: 4, Colors: []string{"R", "G"}},

	// S82 DMU batch 18: Dominaria United's tokens for Greensleeves, Baru, Rasputin and Jedit Ojanen.
	"3/3 green Badger": {Name: "Badger", TypeLine: "Token Creature — Badger", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"4/4 green Wurm":   {Name: "Wurm", TypeLine: "Token Creature — Wurm", Power: 4, Toughness: 4, Colors: []string{"G"}},
	"2/2 white Knight with protection from red": {Name: "Knight", TypeLine: "Token Creature — Knight", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"protection from red"}},
	"2/2 green Cat Warrior with forestwalk":     {Name: "Cat Warrior", TypeLine: "Token Creature — Cat Warrior", Power: 2, Toughness: 2, Colors: []string{"G"}, Keywords: []string{"forestwalk"}},
	"3/1 red Construct artifact with haste":     {Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 3, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"haste"}},

	// S82 GAP batch 109: Heliod, God of the Sun's Cleric.
	"2/1 white Cleric enchantment creature": {Name: "Cleric", TypeLine: "Token Enchantment Creature — Cleric", Power: 2, Toughness: 1, Colors: []string{"W"}},

	// S82 GAP batch 110: Battle at the Helvault's Avacyn.
	"8/8 white Legendary Angel named Avacyn with flying, vigilance, and indestructible": {Name: "Avacyn", TypeLine: "Token Legendary Creature — Angel", Power: 8, Toughness: 8, Colors: []string{"W"}, Keywords: []string{"flying", "vigilance", "indestructible"}},

	// S82 BNS batch 52: Psemilla, Meletian Poet's Nymph.
	"2/2 white Nymph enchantment creature": {Name: "Nymph", TypeLine: "Token Enchantment Creature — Nymph", Power: 2, Toughness: 2, Colors: []string{"W"}},
	// S82 BNS batch 54: Giant Opportunity's Giant.
	"7/7 green Giant": {Name: "Giant", TypeLine: "Token Creature — Giant", Power: 7, Toughness: 7, Colors: []string{"G"}},

	// S82 NEO batch 13: Mechtitan, the Mechtitan Core token.
	"10/10 white and blue and black and red and green Legendary Construct artifact named Mechtitan with flying, vigilance, trample, lifelink, and haste": {Name: "Mechtitan", TypeLine: "Token Legendary Artifact Creature — Construct", Power: 10, Toughness: 10, Colors: []string{"W", "U", "B", "R", "G"}, Keywords: []string{"flying", "vigilance", "trample", "lifelink", "haste"}},
	// S82 NEO batch 14: Invoke the Ancients.
	"4/5 green Spirit": {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 4, Toughness: 5, Colors: []string{"G"}},
	// #5493: Kamigawa: Neon Dynasty's Shrine (Go-Shintai of Life's Origin).
	"1/1 colorless Shrine enchantment": {Name: "Shrine", TypeLine: "Token Enchantment Creature — Shrine", Power: 1, Toughness: 1},
	// S82 AFR batch 15: Drider, Lolth, Minsc and Drizzt.
	"2/1 black Spider with menace and reach":                     {Name: "Spider", TypeLine: "Token Creature — Spider", Power: 2, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"menace", "reach"}},
	"1/1 red Legendary Hamster named Boo with trample and haste": {Name: "Boo", TypeLine: "Token Legendary Creature — Hamster", Power: 1, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"trample", "haste"}},
	"4/1 green Legendary Cat named Guenhwyvar with trample":      {Name: "Guenhwyvar", TypeLine: "Token Legendary Creature — Cat", Power: 4, Toughness: 1, Colors: []string{"G"}, Keywords: []string{"trample"}},
	// S82 BNS batch 02: Outlaws' Merriment's Warrior and Cleric.
	"3/1 red and white Human Warrior with trample and haste": {Name: "Human Warrior", TypeLine: "Token Creature — Human Warrior", Power: 3, Toughness: 1, Colors: []string{"R", "W"}, Keywords: []string{"trample", "haste"}},
	"2/1 red and white Human Cleric with lifelink and haste": {Name: "Human Cleric", TypeLine: "Token Creature — Human Cleric", Power: 2, Toughness: 1, Colors: []string{"R", "W"}, Keywords: []string{"lifelink", "haste"}},

	// S82 BNS batch 25: Foundations Jumpstart's Akroan Crusader.
	"1/1 red Soldier with haste": {Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"haste"}},
	// S82 BNS batch 41: Voice of the Woods.
	"7/7 green Elemental with trample": {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 7, Toughness: 7, Colors: []string{"G"}, Keywords: []string{"trample"}},
	// S82 BNS batch 50: Cynette, Jelly Drover's Jellyfish.
	"1/1 blue Jellyfish with flying": {Name: "Jellyfish", TypeLine: "Token Creature — Jellyfish", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying"}},
	// S82 GAP batch 4: Chief Magistrate of Mercadia's Goblin.
	"2/1 red Goblin with haste": {Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 2, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"haste"}},
	// S82 GAP batch 4: Balefang the Unslayable's Baneslayer Angel token,
	// a {3}{W}{W} 5/5 Angel with the printed card's five keywords.
	"5/5 white Angel named Baneslayer Angel": {Name: "Baneslayer Angel", TypeLine: "Token Creature — Angel", ManaCost: "{3}{W}{W}", Power: 5, Toughness: 5, Colors: []string{"W"}, Keywords: []string{"flying", "first strike", "lifelink", "protection from Demons", "protection from Dragons"}},

	// S82 GAP batch 05: Nivea, Beloved Battlemage's Akroma.
	"6/6 white Legendary Angel named Akroma, Angel of Wrath with flying, first strike, vigilance, trample, haste and protection from black and from red": {Name: "Akroma, Angel of Wrath", TypeLine: "Token Legendary Creature — Angel", Power: 6, Toughness: 6, Colors: []string{"W"}, Keywords: []string{"flying", "first strike", "vigilance", "trample", "haste", "protection from black", "protection from red"}},

	// S82 GAP batch 79: Assemble's Elf Knight and Warden's Sphinx.
	"2/2 green and white Elf Knight with vigilance":       {Name: "Elf Knight", TypeLine: "Token Creature — Elf Knight", Power: 2, Toughness: 2, Colors: []string{"G", "W"}, Keywords: []string{"vigilance"}},
	"4/4 white and blue Sphinx with flying and vigilance": {Name: "Sphinx", TypeLine: "Token Creature — Sphinx", Power: 4, Toughness: 4, Colors: []string{"W", "U"}, Keywords: []string{"flying", "vigilance"}},

	// S82 GAP batch 06: Worzel, the Protector's Scryb Sprites.
	"1/1 green Faerie with flying named Scryb Sprites": {Name: "Scryb Sprites", TypeLine: "Token Creature — Faerie", Power: 1, Toughness: 1, Colors: []string{"G"}, Keywords: []string{"flying"}},
	// #5559: Born of the Gods' enchantment creature tokens (Fated
	// Intervention, Pheres-Band Raiders, God-Favored General, Satyr Nyx-Smith).
	"3/3 green Centaur enchantment":            {Name: "Centaur", TypeLine: "Token Enchantment Creature — Centaur", Power: 3, Toughness: 3, Colors: []string{"G"}},
	"1/1 white Soldier enchantment":            {Name: "Soldier", TypeLine: "Token Enchantment Creature — Soldier", Power: 1, Toughness: 1, Colors: []string{"W"}},
	"3/1 red Elemental enchantment with haste": {Name: "Elemental", TypeLine: "Token Enchantment Creature — Elemental", Power: 3, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"haste"}},

	// Doctor Who (#5948): Iraxxa, Empress of Mars; Vrestin, Menoptra Leader.
	"2/2 red Alien Warrior":                        {Name: "Alien Warrior", TypeLine: "Token Creature — Alien Warrior", Power: 2, Toughness: 2, Colors: []string{"R"}},
	"1/1 green and white Alien Insect with flying": {Name: "Alien Insect", TypeLine: "Token Creature — Alien Insect", Power: 1, Toughness: 1, Colors: []string{"G", "W"}, Keywords: []string{"flying"}},
	// S82 GAP batch 07: Soul Separator's Zombie (printed */*, sized from the exiled card as it is created).
	"0/0 black Zombie": {Name: "Zombie", TypeLine: "Token Creature — Zombie", Colors: []string{"B"}, PrintedPTKnown: true},
	// S82 GAP batch 101: Sek'Kuar, Deathkeeper's Graveborn.
	"3/1 black and red Graveborn with haste": {Name: "Graveborn", TypeLine: "Token Creature — Graveborn", Power: 3, Toughness: 1, Colors: []string{"B", "R"}, Keywords: []string{"haste"}},

	// S82 GAP batch 49: Wurmcoil Larva's two Phyrexian Wurms.
	"1/2 black Phyrexian Wurm artifact with deathtouch": {Name: "Phyrexian Wurm", TypeLine: "Token Artifact Creature — Phyrexian Wurm", Power: 1, Toughness: 2, Colors: []string{"B"}, Keywords: []string{"deathtouch"}},
	"2/1 black Phyrexian Wurm artifact with lifelink":   {Name: "Phyrexian Wurm", TypeLine: "Token Artifact Creature — Phyrexian Wurm", Power: 2, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"lifelink"}},
	// S82 GAP batch 58: Ajani, Nacatl Pariah's Cat Warrior.
	"2/1 white Cat Warrior": {Name: "Cat Warrior", TypeLine: "Token Creature — Cat Warrior", Power: 2, Toughness: 1, Colors: []string{"W"}},
	// S82 GAP batch 62: Mr. House, President and CEO's Robot.
	"3/3 colorless Robot artifact": {Name: "Robot", TypeLine: "Token Artifact Creature — Robot", Power: 3, Toughness: 3},
	// S82 GAP batch 185: Soundwave, Superior Captain's Ravage and Laserbeak.
	"3/3 black Legendary Robot artifact named Ravage with menace and deathtouch": {Name: "Ravage", TypeLine: "Token Legendary Artifact Creature — Robot", Power: 3, Toughness: 3, Colors: []string{"B"}, Keywords: []string{"menace", "deathtouch"}},
	"2/2 blue Legendary Robot artifact named Laserbeak with flying and hexproof": {Name: "Laserbeak", TypeLine: "Token Legendary Artifact Creature — Robot", Power: 2, Toughness: 2, Colors: []string{"U"}, Keywords: []string{"flying", "hexproof"}},
	// S82 GAP batches 186, 188 and 196 (Warhammer 40,000 Commander): Necron Warrior,
	// shared by batch 186's cards, batch 188's cards and Biotransference.
	"2/2 black Necron Warrior artifact": {Name: "Necron Warrior", TypeLine: "Token Artifact Creature — Necron Warrior", Power: 2, Toughness: 2, Colors: []string{"B"}},
	// S82 GAP batches 186, 190 and 195: white Astartes Warrior with vigilance,
	// shared by batch 186's cards, Marneus Calgar and Belisarius Cawl.
	"2/2 white Astartes Warrior with vigilance": {Name: "Astartes Warrior", TypeLine: "Token Creature — Astartes Warrior", Power: 2, Toughness: 2, Colors: []string{"W"}, Keywords: []string{"vigilance"}},
	// S82 GAP batches 187 and 196: Great Unclean One's and Nurgle's Rot's Plaguebearer of Nurgle.
	"1/3 black Demon named Plaguebearer of Nurgle": {Name: "Plaguebearer of Nurgle", TypeLine: "Token Creature — Demon", Power: 1, Toughness: 3, Colors: []string{"B"}},
	// S82 GAP batch 188: Mortarion's Astartes Warrior.
	"2/2 black Astartes Warrior with menace": {Name: "Astartes Warrior", TypeLine: "Token Creature — Astartes Warrior", Power: 2, Toughness: 2, Colors: []string{"B"}, Keywords: []string{"menace"}},
	// #5948: Warhammer 40,000 Commander's Cherubael (Inquisitor Eisenhorn).
	"4/4 black Legendary Demon with flying named Cherubael": {Name: "Cherubael", TypeLine: "Token Legendary Creature — Demon", Power: 4, Toughness: 4, Colors: []string{"B"}, Keywords: []string{"flying"}},
	// S82 GAP batch 193: Tyranid Gargoyle (Gargoyle Flock, Tyranid Harridan).
	"1/1 blue Tyranid Gargoyle with flying": {Name: "Tyranid Gargoyle", TypeLine: "Token Creature — Tyranid Gargoyle", Power: 1, Toughness: 1, Colors: []string{"U"}, Keywords: []string{"flying"}},
	// S82 GAP batch 194: Magnus the Red's Spawn.
	"3/3 red Spawn": {Name: "Spawn", TypeLine: "Token Creature — Spawn", Power: 3, Toughness: 3, Colors: []string{"R"}},
	// S82 GAP batch 175: Kasmina, Enigmatic Mentor's Wizard.
	"2/2 blue Wizard": {Name: "Wizard", TypeLine: "Token Creature — Wizard", Power: 2, Toughness: 2, Colors: []string{"U"}},
	// S82 GAP batch 136: Penumbra Bobcat's Cat.
	"2/1 black Cat": {Name: "Cat", TypeLine: "Token Creature — Cat", Power: 2, Toughness: 1, Colors: []string{"B"}},
	// S82 MOR batch 1 (#5592): Walker of the Grove, Feudkiller's Verdict,
	// Weirding Shaman and Stonybrook Schoolmaster.
	"1/1 black Goblin Rogue":  {Name: "Goblin Rogue", TypeLine: "Token Creature — Goblin Rogue", Power: 1, Toughness: 1, Colors: []string{"B"}},
	"1/1 blue Merfolk Wizard": {Name: "Merfolk Wizard", TypeLine: "Token Creature — Merfolk Wizard", Power: 1, Toughness: 1, Colors: []string{"U"}},
	"4/4 green Elemental":     {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 4, Toughness: 4, Colors: []string{"G"}},
	"5/5 white Giant Warrior": {Name: "Giant Warrior", TypeLine: "Token Creature — Giant Warrior", Power: 5, Toughness: 5, Colors: []string{"W"}},
	// S82 GAP2 batch 12 (#6100): The Curse of Fenric's Mutant and The Sea
	// Devils' Alien Salamander.
	"3/3 green Mutant with deathtouch":           {Name: "Mutant", TypeLine: "Token Creature — Mutant", Power: 3, Toughness: 3, Colors: []string{"G"}, Keywords: []string{"deathtouch"}},
	"2/2 green Alien Salamander with islandwalk": {Name: "Alien Salamander", TypeLine: "Token Creature — Alien Salamander", Power: 2, Toughness: 2, Colors: []string{"G"}, Keywords: []string{"islandwalk"}},
	// S82 OLD batch 153 (#5517): Gargoyle Castle's Gargoyle.
	"3/4 colorless Gargoyle artifact with flying": {Name: "Gargoyle", TypeLine: "Token Artifact Creature — Gargoyle", Power: 3, Toughness: 4, Keywords: []string{"flying"}},
	// S82 GAP2 batch 8 (#5948): Tasha, the Witch Queen's Demon.
	"3/3 black Demon": {Name: "Demon", TypeLine: "Token Creature — Demon", Power: 3, Toughness: 3, Colors: []string{"B"}},

	// Secret Lair Drop partner commanders (#5462, S82 OLD batch 014): Tevesh Szat, Doom of Fools.
	"0/1 black Thrull": {Name: "Thrull", TypeLine: "Token Creature — Thrull", Power: 0, Toughness: 1, Colors: []string{"B"}},
	// S82 OLD batch 26 (#5462): Gisa's Favorite Shovel's Walker.
	"2/2 black Zombie named Walker": {Name: "Walker", TypeLine: "Token Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}},
	// S82 OLD batch 68 (#5505): Troll Warrior (Gnottvold Slumbermound; also batch 071's Old-Growth Troll and batch 060's Waking the Trolls).
	"4/4 green Troll Warrior with trample": {Name: "Troll Warrior", TypeLine: "Token Creature — Troll Warrior", Power: 4, Toughness: 4, Colors: []string{"G"}, Keywords: []string{"trample"}},
	// S82 OLD batch 060 (#5505): The Bloodsky Massacre's Demon Berserker.
	"2/3 red Demon Berserker with menace": {Name: "Demon Berserker", TypeLine: "Token Creature — Demon Berserker", Power: 2, Toughness: 3, Colors: []string{"R"}, Keywords: []string{"menace"}},

	// S82 OLD batch 083: Profane Transfusion's X/X Horror (printed */*; a 0/0 placeholder, set to the
	// difference between the life totals as it is created).
	"0/0 colorless Horror artifact": {Name: "Horror", TypeLine: "Token Artifact Creature — Horror", PrintedPTKnown: true},

	// S82 OLD batch 110 (#5511): Experimental Overload's X/X blue and red Weird (printed */*; a 0/0
	// placeholder, set to the instants and sorceries in your graveyard as it is created).
	"0/0 blue and red Weird": {Name: "Weird", TypeLine: "Token Creature — Weird", Colors: []string{"U", "R"}, PrintedPTKnown: true},

	// S82 OLD batch 124 (#5513): Wort, the Raidmother's Goblin Warrior and Gavi, Nest Warden's Dinosaur Cat.
	"1/1 red and green Goblin Warrior": {Name: "Goblin Warrior", TypeLine: "Token Creature — Goblin Warrior", Power: 1, Toughness: 1, Colors: []string{"R", "G"}},
	"2/2 red and white Dinosaur Cat":   {Name: "Dinosaur Cat", TypeLine: "Token Creature — Dinosaur Cat", Power: 2, Toughness: 2, Colors: []string{"R", "W"}},
	// S82 OLD batch 134 (#5514): Kiora Bests the Sea God's Kraken.
	"8/8 blue Kraken with hexproof": {Name: "Kraken", TypeLine: "Token Creature — Kraken", Power: 8, Toughness: 8, Colors: []string{"U"}, Keywords: []string{"hexproof"}},
	// Purphoros's Intervention: an X/1 whose printed token card is */1; 0/1 here, X set as it is created.
	"0/1 red Elemental with trample and haste": {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 0, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"trample", "haste"}},
	// S82 OLD batch 136 (#5514): Alirios, Enraptured's Reflection.
	"3/2 blue Reflection": {Name: "Reflection", TypeLine: "Token Creature — Reflection", Power: 3, Toughness: 2, Colors: []string{"U"}},
	// S82 OLD batch 170 (#5520): Planewide Celebration's Citizen.
	"2/2 white and blue and black and red and green Citizen": {Name: "Citizen", TypeLine: "Token Creature — Citizen", Power: 2, Toughness: 2, Colors: []string{"W", "U", "B", "R", "G"}},
	// S82 OLD batch 172 (#5520): Tolsimir, Friend to Wolves's Voja.
	"3/3 green and white Legendary Wolf named Voja, Friend to Elves": {Name: "Voja, Friend to Elves", TypeLine: "Token Legendary Creature — Wolf", Power: 3, Toughness: 3, Colors: []string{"G", "W"}},
	// S82 OLD batch 205 (#5529): Quest for the Gravelord's Zombie Giant.
	"5/5 black Zombie Giant": {Name: "Zombie Giant", TypeLine: "Token Creature — Zombie Giant", Power: 5, Toughness: 5, Colors: []string{"B"}},
	// S82 OLD batch 249 (#5535): Trial of Strength's Beast.
	"4/2 green Beast": {Name: "Beast", TypeLine: "Token Creature — Beast", Power: 4, Toughness: 2, Colors: []string{"G"}},
	// S82 OLD batch 252 (#5536): Release the Gremlins and Gremlin Infestation's Gremlin.
	"2/2 red Gremlin": {Name: "Gremlin", TypeLine: "Token Creature — Gremlin", Power: 2, Toughness: 2, Colors: []string{"R"}},
	// S82 OLD batch 316 (#5548): Hellion Eruption's Hellion.
	"4/4 red Hellion": {Name: "Hellion", TypeLine: "Token Creature — Hellion", Power: 4, Toughness: 4, Colors: []string{"R"}},
	// S82 OLD batch 333 (#5551): Corpseweft's Zombie Horror; an X/X whose printed token card is */*, so 0/0 here and X is set as it is created.
	"0/0 black Zombie Horror": {Name: "Zombie Horror", TypeLine: "Token Creature — Zombie Horror", Power: 0, Toughness: 0, Colors: []string{"B"}},

	"5/5 colorless Legendary Goblin Golem artifact named Tuktuk the Returned": {Name: "Tuktuk the Returned", TypeLine: "Token Legendary Artifact Creature — Goblin Golem", Power: 5, Toughness: 5},

	// S82 OLD b370: Journey into Nyx, Rise of Eagles.
	"2/2 blue Bird enchantment with flying": {Name: "Bird", TypeLine: "Token Enchantment Creature — Bird", Power: 2, Toughness: 2, Colors: []string{"U"}, Keywords: []string{"flying"}},

	// S82 OLD b371: Journey into Nyx, Renowned Weaver.
	"1/3 green Spider enchantment with reach": {Name: "Spider", TypeLine: "Token Enchantment Creature — Spider", Power: 1, Toughness: 3, Colors: []string{"G"}, Keywords: []string{"reach"}},

	// S82 OLD batch 375 (#5559): Forlorn Pseudamma's token (Aerie Worshippers uses b370's Bird row).
	"2/2 black Zombie enchantment creature": {Name: "Zombie", TypeLine: "Token Enchantment Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}},

	// S82 OLD batch 426 (#5566): Grove of the Guardian's Elemental.
	"8/8 green and white Elemental with vigilance": {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 8, Toughness: 8, Colors: []string{"G", "W"}, Keywords: []string{"vigilance"}},

	// S82 OLD batch 427 (#5567): Thunderheads' Weird (the printed token's Weird face).
	"3/3 blue Weird with defender and flying": {Name: "Weird", TypeLine: "Token Creature — Weird", Power: 3, Toughness: 3, Colors: []string{"U"}, Keywords: []string{"defender", "flying"}},

	// S82 OLD batch 461 (#5575): Chancellor of the Forge's Goblins (the printed token has no haste; the text gives it).
	"1/1 red Phyrexian Goblin with haste": {Name: "Phyrexian Goblin", TypeLine: "Token Creature — Phyrexian Goblin", Power: 1, Toughness: 1, Colors: []string{"R"}, Keywords: []string{"haste"}},

	// S82 OLD batch 496: Kalitas, Bloodchief of Ghet's X/X Vampire (printed */*; a 0/0
	// placeholder, set to the destroyed creature's power and toughness as it is created).
	// Zektar Shrine Expedition's 7/1 Elemental shares the existing Elemental Appeal row.
	"0/0 black Vampire": {Name: "Vampire", TypeLine: "Token Creature — Vampire", Colors: []string{"B"}, PrintedPTKnown: true},

	// S82 OLD batch 503 (#5586): Predatory Advantage's Lizard.
	"2/2 green Lizard": {Name: "Lizard", TypeLine: "Token Creature — Lizard", Power: 2, Toughness: 2, Colors: []string{"G"}},

	// S82 OLD batch 510 (#5588): Puppet Conjurer's Homunculus and Godsire's Beast.
	"0/1 blue Homunculus artifact creature": {Name: "Homunculus", TypeLine: "Token Artifact Creature — Homunculus", Power: 0, Toughness: 1, Colors: []string{"U"}},
	"8/8 red and green and white Beast":     {Name: "Beast", TypeLine: "Token Creature — Beast", Power: 8, Toughness: 8, Colors: []string{"R", "G", "W"}},

	// S82 OLD batch 533 (#5591): Giantbaiting's Giant Warrior.
	"4/4 red and green Giant Warrior with haste": {Name: "Giant Warrior", TypeLine: "Token Creature — Giant Warrior", Power: 4, Toughness: 4, Colors: []string{"R", "G"}, Keywords: []string{"haste"}},

	// S82 OLD batch 535 (#5591): Din of the Fireherd's Elemental.
	"5/5 black and red Elemental": {Name: "Elemental", TypeLine: "Token Creature — Elemental", Power: 5, Toughness: 5, Colors: []string{"B", "R"}},

	// S82 OLD batch 635 (#5614): Penumbra Kavu's Kavu, Penumbra Wurm's Wurm and
	// Haunted Angel's Angel.
	"3/3 black Kavu":              {Name: "Kavu", TypeLine: "Token Creature — Kavu", Power: 3, Toughness: 3, Colors: []string{"B"}},
	"6/6 black Wurm with trample": {Name: "Wurm", TypeLine: "Token Creature — Wurm", Power: 6, Toughness: 6, Colors: []string{"B"}, Keywords: []string{"trample"}},
	"3/3 black Angel with flying": {Name: "Angel", TypeLine: "Token Creature — Angel", Power: 3, Toughness: 3, Colors: []string{"B"}, Keywords: []string{"flying"}},
}
