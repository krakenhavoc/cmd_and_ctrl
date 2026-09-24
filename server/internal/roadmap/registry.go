package roadmap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// registry.go — the curated list. One entry per keyword, mechanic and
// engine seam; see roadmap.go for what each field means and
// registry_test.go for what CI holds each one to.
//
// # Adding to it
//
//   - A batch PR that skips a card for a seam appends the card's name
//     to that seam's Waiting, then regenerates docs/engine-seams.md:
//     go test ./internal/roadmap/ -update. A seam with no entry yet
//     gets one here, with a player-facing Summary and Missing.
//   - An engine PR that closes a seam flips its Status here and moves
//     its prose into engine-seams.md's hand-written Closed seams list.
//     TestWaitingCardsAreNotComplete fails the day a waiting card is
//     registered complete, which is usually the first sign.
//   - Summary and Missing are read by players. Write them the way a
//     Caveat is written: a sentence, no engine vocabulary. The
//     engineering belongs in EngineNotes.
//
// Registry order is display order within a kind.

// items is the registry.
var items = []Item{
	// ── Keywords ─────────────────────────────────────────────────
	//
	// One entry per row of game.CanonicalKeywordTable (landwalk is one
	// family of six). A keyword is implemented by construction: it
	// joins that table in the same change that teaches the engine to
	// honour it.
	{
		Slug: "flying", Name: "Flying", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A creature with flying can be blocked only by creatures with flying or reach.",
		Rules:    []string{"702.9"},
		Keywords: []string{"flying"},
		Probe:    hasKeyword("flying"),
		Printed:  printedKeyword("flying"),
		Examples: []string{"Serra Angel"},
	},
	{
		Slug: "reach", Name: "Reach", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A creature with reach can block creatures with flying.",
		Rules:    []string{"702.17"},
		Keywords: []string{"reach"},
		Probe:    hasKeyword("reach"),
		Printed:  printedKeyword("reach"),
	},
	{
		Slug: "first-strike", Name: "First strike", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A creature with first strike deals its combat damage before creatures without it, in a combat damage step of its own.",
		Rules:    []string{"702.7"},
		Keywords: []string{"first strike"},
		Probe:    hasKeyword("first strike"),
		Printed:  printedKeyword("first strike"),
	},
	{
		Slug: "double-strike", Name: "Double strike", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A creature with double strike deals combat damage twice: once alongside first strike, then again as normal.",
		Rules:    []string{"702.4"},
		Keywords: []string{"double strike"},
		Probe:    hasKeyword("double strike"),
		Printed:  printedKeyword("double strike"),
	},
	{
		Slug: "deathtouch", Name: "Deathtouch", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "Any amount of damage a source with deathtouch deals to a creature is enough to destroy it.",
		Rules:    []string{"702.2"},
		Keywords: []string{"deathtouch"},
		Probe:    hasKeyword("deathtouch"),
		Printed:  printedKeyword("deathtouch"),
	},
	{
		Slug: "lifelink", Name: "Lifelink", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "Damage dealt by a source with lifelink also gains its controller that much life.",
		Rules:    []string{"702.15"},
		Keywords: []string{"lifelink"},
		Probe:    hasKeyword("lifelink"),
		Printed:  printedKeyword("lifelink"),
	},
	{
		Slug: "trample", Name: "Trample", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "An attacking creature with trample deals the damage its blockers don't need to the player or permanent it is attacking.",
		Rules:    []string{"702.19"},
		Keywords: []string{"trample"},
		Probe:    hasKeyword("trample"),
		Printed:  printedKeyword("trample"),
	},
	{
		Slug: "vigilance", Name: "Vigilance", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "Attacking doesn't cause a creature with vigilance to tap.",
		Rules:    []string{"702.20"},
		Keywords: []string{"vigilance"},
		Probe:    hasKeyword("vigilance"),
		Printed:  printedKeyword("vigilance"),
	},
	{
		Slug: "menace", Name: "Menace", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A creature with menace can't be blocked except by two or more creatures.",
		Rules:    []string{"702.111"},
		Keywords: []string{"menace"},
		Probe:    hasKeyword("menace"),
		Printed:  printedKeyword("menace"),
	},
	{
		Slug: "defender", Name: "Defender", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A creature with defender can't attack.",
		Rules:    []string{"702.3"},
		Keywords: []string{"defender"},
		Probe:    hasKeyword("defender"),
		Printed:  printedKeyword("defender"),
	},
	{
		Slug: "haste", Name: "Haste", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A creature with haste can attack and use its tap abilities the turn it comes under your control.",
		Rules:    []string{"702.10"},
		Keywords: []string{"haste"},
		Probe:    hasKeyword("haste"),
		Printed:  printedKeyword("haste"),
	},
	{
		Slug: "flash", Name: "Flash", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "You may cast a card with flash any time you could cast an instant.",
		Rules:    []string{"702.8"},
		Keywords: []string{"flash"},
		Probe:    hasKeyword("flash"),
		Printed:  printedKeyword("flash"),
	},
	{
		Slug: "hexproof", Name: "Hexproof", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A permanent with hexproof can't be the target of spells or abilities your opponents control.",
		Rules:    []string{"702.11"},
		Keywords: []string{"hexproof"},
		Probe:    hasKeyword("hexproof"),
		Printed:  printedKeyword("hexproof"),
	},
	{
		Slug: "shroud", Name: "Shroud", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A permanent with shroud can't be the target of any spell or ability, including its controller's.",
		Rules:    []string{"702.18"},
		Keywords: []string{"shroud"},
		Probe:    hasKeyword("shroud"),
		Printed:  printedKeyword("shroud"),
	},
	{
		Slug: "indestructible", Name: "Indestructible", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A permanent with indestructible isn't destroyed by lethal damage or by effects that say \"destroy\".",
		Rules:    []string{"702.12"},
		Keywords: []string{"indestructible"},
		Probe:    hasKeyword("indestructible"),
		Printed:  printedKeyword("indestructible"),
	},
	{
		Slug: "changeling", Name: "Changeling", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A card with changeling is every creature type, wherever it is.",
		Rules:    []string{"702.73"},
		Keywords: []string{game.KeywordChangeling},
		Probe:    hasKeyword(game.KeywordChangeling),
		Printed:  printedKeyword("changeling"),
	},
	{
		Slug: "landwalk", Name: "Landwalk", Kind: KindKeyword, Status: StatusImplemented,
		Summary: "A creature with landwalk (islandwalk, forestwalk and the rest) can't be blocked while the defending player controls a land of that kind.",
		Rules:   []string{"702.14"},
		Keywords: []string{
			"plainswalk", "islandwalk", "swampwalk", "mountainwalk", "forestwalk", "nonbasic landwalk",
		},
		Probe:   hasKeyword("plainswalk", "islandwalk", "swampwalk", "mountainwalk", "forestwalk", "nonbasic landwalk"),
		Printed: `(?im)^(?:[a-z' -]+, )*(?:plains|island|swamp|mountain|forest|nonbasic land)walk\b`,
		Phrases: []string{"landwalk", "snow swampwalk", "desertwalk"},
	},
	{
		Slug: "fear", Name: "Fear", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A creature with fear can be blocked only by artifact creatures and black creatures.",
		Rules:    []string{"702.36"},
		Keywords: []string{"fear"},
		Probe:    hasKeyword("fear"),
		Printed:  printedKeyword("fear"),
		Examples: []string{"Guiltfeeder"},
	},
	{
		Slug: "intimidate", Name: "Intimidate", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A creature with intimidate can be blocked only by artifact creatures and creatures that share a color with it.",
		Rules:    []string{"702.13"},
		Keywords: []string{"intimidate"},
		Probe:    hasKeyword("intimidate"),
		Printed:  printedKeyword("intimidate"),
	},
	{
		Slug: "shadow", Name: "Shadow", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A creature with shadow can block or be blocked only by other creatures with shadow.",
		Rules:    []string{"702.28"},
		Keywords: []string{"shadow"},
		Probe:    hasKeyword("shadow"),
		Printed:  printedKeyword("shadow"),
	},
	{
		Slug: "horsemanship", Name: "Horsemanship", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A creature with horsemanship can be blocked only by creatures with horsemanship.",
		Rules:    []string{"702.31"},
		Keywords: []string{"horsemanship"},
		Probe:    hasKeyword("horsemanship"),
		Printed:  printedKeyword("horsemanship"),
	},
	{
		Slug: "skulk", Name: "Skulk", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A creature with skulk can't be blocked by creatures with greater power.",
		Rules:    []string{"702.118"},
		Keywords: []string{"skulk"},
		Probe:    hasKeyword("skulk"),
		Printed:  printedKeyword("skulk"),
	},
	{
		Slug: "cycling", Name: "Cycling", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "Pay a card's cycling cost and discard it from your hand to draw a card.",
		Rules:    []string{"702.29"},
		Keywords: []string{"cycling"},
		Probe:    activated(func(ab effects.ActivatedAbility) bool { return ab.Cycling }),
		Printed:  printedLine("cycling"),
	},
	{
		Slug: "protection", Name: "Protection", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "Protection from a color, a card type or a creature type stops damage, targeting, blocking and attachments from anything with that quality.",
		Rules:    []string{"702.16"},
		ADR:      "0072-protection.md",
		Keywords: []string{game.KeywordProtection},
		Probe:    hasKeyword("protection from "),
		Printed:  printedLine("protection from"),
	},
	{
		Slug: "foretell", Name: "Foretell", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "On your turn, exile a card with foretell face down from your hand for two mana, then cast it on a later turn for its foretell cost.",
		Rules:    []string{"702.143"},
		ADR:      "0062-abilities-and-special-actions-from-the-hand.md",
		Keywords: []string{"foretell"},
		Probe:    specialAction(game.SpecialActionForetell),
		Examples: []string{"Saw It Coming"},
	},
	{
		Slug: "suspend", Name: "Suspend", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "Exile a card from your hand with time counters on it; one comes off each upkeep, and when the last one does you cast it for free.",
		Rules:    []string{"702.62"},
		ADR:      "0062-abilities-and-special-actions-from-the-hand.md",
		Keywords: []string{"suspend"},
		Probe:    specialAction(game.SpecialActionSuspend),
	},
	{
		Slug: "madness", Name: "Madness", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "When you discard a card with madness, you may cast it for its madness cost instead of putting it into your graveyard.",
		Rules:    []string{"702.35"},
		Keywords: []string{"madness"},
		Probe:    func(s effects.Spec) bool { return s.Madness != "" },
	},
	{
		Slug: "plot", Name: "Plot", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "Pay a card's plot cost to exile it from your hand as a sorcery, then cast it for free on a later turn.",
		Rules:    []string{"702.170"},
		ADR:      "0062-abilities-and-special-actions-from-the-hand.md",
		Keywords: []string{"plot"},
		Probe:    specialAction(game.SpecialActionPlot),
	},
	{
		Slug: "phasing", Name: "Phasing", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A permanent that phases out is treated as though it doesn't exist until it phases back in during its controller's next untap step.",
		Rules:    []string{"702.26"},
		ADR:      "0084-phasing.md",
		Keywords: []string{game.KeywordPhasing},
		Probe:    hasKeyword(game.KeywordPhasing),
		Printed:  `(?i)\bphases? (?:out|in)\b|` + printedKeyword("phasing"),
		Examples: []string{"Teferi's Protection"},
	},
	{
		Slug: "infect", Name: "Infect", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "Damage from a source with infect is dealt to creatures as -1/-1 counters and to players as poison counters.",
		Rules:    []string{"702.90"},
		ADR:      "0056-infect-wither-toxic.md",
		Keywords: []string{game.KeywordInfect},
		Probe:    hasKeyword(game.KeywordInfect),
		Printed:  printedKeyword("infect"),
		Examples: []string{"Plague Myr"},
	},
	{
		Slug: "wither", Name: "Wither", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "Damage from a source with wither is dealt to creatures as -1/-1 counters, which stay after the turn ends.",
		Rules:    []string{"702.80"},
		ADR:      "0056-infect-wither-toxic.md",
		Keywords: []string{game.KeywordWither},
		Probe:    hasKeyword(game.KeywordWither),
		Printed:  printedKeyword("wither"),
		Examples: []string{"Puncture Blast"},
	},
	{
		Slug: "toxic", Name: "Toxic", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "A player dealt combat damage by a creature with toxic also gets that many poison counters, and every instance of toxic a creature has adds up.",
		Rules:    []string{"702.164"},
		ADR:      "0056-infect-wither-toxic.md",
		Keywords: []string{game.KeywordToxic},
		Probe:    hasKeyword(game.KeywordToxic + " "),
		Printed:  printedLine("Toxic"),
		Examples: []string{"Karumonix, the Rat King"},
	},
	{
		Slug: "prowess", Name: "Prowess", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "Whenever you cast a noncreature spell, a creature with prowess gets +1/+1 until end of turn, once for each instance of prowess it has.",
		Rules:    []string{"702.108"},
		ADR:      "0014-combat-keywords.md",
		Keywords: []string{game.KeywordProwess},
		Mechanic: "prowess",
		Printed:  printedKeyword("prowess"),
		Examples: []string{"Ty Lee, Chi Blocker"},
	},
	{
		Slug: "split-second", Name: "Split second", Kind: KindKeyword, Status: StatusImplemented,
		Summary:  "While a spell with split second is on the stack, nobody can cast spells or activate abilities other than mana abilities. Triggered abilities and special actions still happen.",
		Rules:    []string{"702.61"},
		ADR:      "0007-stack-foundation.md",
		Keywords: []string{game.KeywordSplitSecond},
		Mechanic: "split second",
		Printed:  printedKeyword("split second"),
		Examples: []string{"Krosan Grip"},
	},

	// ── Mechanics from the coverage table ────────────────────────
	//
	// Every row of coverage.Mechanics(), by name: the probe and the
	// caveat phrases come from there, so there is one copy of each.
	{
		Slug: "trigger-doubling", Name: "Trigger doubling", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Permanents like Panharmonicon make other abilities trigger an additional time.",
		Mechanic: "trigger doubling",
		Examples: []string{"Panharmonicon"},
	},
	{
		Slug: "flashback", Name: "Flashback", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Cast a spell from your graveyard for its flashback cost; afterwards it's exiled.",
		Rules:    []string{"702.34"},
		Mechanic: "flashback",
	},
	{
		Slug: "escape", Name: "Escape", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Cast a card from your graveyard for its escape cost, exiling other cards from your graveyard to pay for it.",
		Rules:    []string{"702.138"},
		Mechanic: "escape",
	},
	{
		Slug: "warp", Name: "Warp", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Cast a permanent cheaply for its warp cost; it's exiled at the next end step and can be cast again on a later turn.",
		Rules:    []string{"702.185"},
		Mechanic: "warp",
	},
	{
		Slug: "free-cast", Name: "Free casts", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Spells you may cast without paying their mana cost when a condition holds, such as controlling your commander.",
		Mechanic: "free cast",
	},
	{
		Slug: "evoke", Name: "Evoke", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Cast a creature for its evoke cost to get its enters ability; the creature is then sacrificed.",
		Rules:    []string{"702.74"},
		Mechanic: "evoke",
	},
	{
		Slug: "overload", Name: "Overload", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Pay a spell's overload cost to have it affect everything it could target instead of one target.",
		Mechanic: "overload",
	},
	{
		Slug: "cleave", Name: "Cleave", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Pay a spell's cleave cost to cast it without the words in square brackets.",
		Mechanic: "cleave",
	},
	{
		Slug: "pitch-cost", Name: "Exile a card instead of paying mana", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Spells like Force of Will that you may cast by exiling a card from your hand instead of paying their mana cost.",
		Mechanic: "pitch cost",
	},
	{
		Slug: "pay-life-instead", Name: "Pay life instead of mana", Kind: KindMechanic, Status: StatusImplemented,
		Summary:          "Spells that let you pay life rather than their mana cost when a condition holds.",
		Mechanic:         "pay life instead",
		NoCatalogExample: "Its only catalogued card hasn't been reviewed yet, so none is listed as fully automated.",
	},
	{
		Slug: "cast-from-graveyard", Name: "Casting from your graveyard", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Cards whose own text lets you cast them from your graveyard.",
		Mechanic: "casting from the graveyard",
	},
	{
		Slug: "loses-all-abilities", Name: "Losing all abilities", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Effects that make a permanent lose all its abilities switch off everything its card text does, not just its keywords.",
		ADR:      "0046-layer-6-authoritative.md",
		Mechanic: "loses all abilities",
		Examples: []string{"Darksteel Mutation"},
	},
	{
		Slug: "self-cost-modifier", Name: "Spells that change their own cost", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Spells that cost less, or more, to cast because of their own text, such as affinity.",
		Mechanic: "a spell's own cost modifier",
	},
	{
		Slug: "surveil", Name: "Surveil", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Look at the top cards of your library and put any number of them into your graveyard and the rest back on top.",
		Rules:    []string{"701.25"},
		Mechanic: "surveil",
		Printed:  printedWords("surveil"),
	},
	{
		Slug: "converge-sunburst", Name: "Converge and sunburst", Kind: KindMechanic, Status: StatusPartial,
		Summary:          "Spells and permanents that count the colors of mana spent to cast them.",
		Missing:          "With strict mana off, the game doesn't track which mana you spent, so these cards count no colors.",
		Issue:            761,
		Phrases:          []string{"strict mana off"},
		NoCatalogExample: "Every catalogued card that counts spent colors carries the strict-mana caveat, so none is listed as fully automated.",
		ADR:              "0068-the-mana-spent-on-a-spell.md",
		Mechanic:         "converge and sunburst",
	},
	{
		Slug: "adamant", Name: "Adamant", Kind: KindMechanic, Status: StatusPartial,
		Summary:          "Spells that do more when at least three mana of one color was spent to cast them.",
		Missing:          "With strict mana off, the game doesn't track which mana you spent, so adamant never turns on.",
		Issue:            761,
		NoCatalogExample: "The catalogued adamant card carries the strict-mana caveat, so none is listed as fully automated.",
		ADR:              "0068-the-mana-spent-on-a-spell.md",
		Mechanic:         "adamant",
		Printed:          printedLine("adamant"),
	},
	{
		Slug: "counter-removal-cost", Name: "Removing counters as a cost", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Abilities that remove counters from a permanent as part of their cost.",
		ADR:      "0020-activated-abilities.md",
		Mechanic: "a counter-removal cost",
	},
	{
		Slug: "counter-to-a-zone", Name: "Counterspells that exile or bounce", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Counterspells that send the countered spell somewhere other than its owner's graveyard.",
		Mechanic: "counter to a zone",
		Examples: []string{"Hinder", "Spell Crumple", "Devious Cover-Up"},
	},
	{
		Slug: "gift", Name: "Gift", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Promise an opponent a gift as you cast the spell, and the spell does more.",
		Rules:    []string{"702.174"},
		ADR:      "0089-gift.md",
		Mechanic: "gift",
	},

	// ── Other mechanics with a shape in the catalog ──────────────
	{
		Slug: "kicker", Name: "Kicker", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Pay a spell's optional kicker cost as you cast it and it does more.",
		Rules:   []string{"702.33"},
		ADR:     "0073-optional-additional-costs-and-the-cast-gate.md",
		Probe:   optionalCost(game.KickerKey),
		Printed: printedLine("kicker"),
	},
	{
		Slug: "multikicker", Name: "Multikicker", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Pay a spell's multikicker cost as many times as you like, and it does more for each.",
		Rules:   []string{"702.33d"},
		ADR:     "0073-optional-additional-costs-and-the-cast-gate.md",
		Probe:   optionalCost(game.MultikickerKey),
		Printed: printedLine("multikicker"),
	},
	{
		Slug: "buyback", Name: "Buyback", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Pay a spell's buyback cost and it returns to your hand as it resolves.",
		Rules:   []string{"702.27"},
		ADR:     "0073-optional-additional-costs-and-the-cast-gate.md",
		Probe:   optionalCost(game.BuybackKey),
		Printed: printedLine("buyback"),
	},
	{
		Slug: "offspring", Name: "Offspring", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Pay a creature's offspring cost as you cast it to also get a 1/1 token copy of it.",
		Probe:   optionalCost(effects.OffspringKey),
		Printed: printedLine("offspring"),
	},
	{
		Slug: "spree", Name: "Spree", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Choose any number of a spree spell's modes, paying the extra cost of each one you choose.",
		Rules:   []string{"702.172"},
		Probe: func(s effects.Spec) bool {
			return s.Modes != nil && len(s.Modes.Prompt) >= 5 && s.Modes.Prompt[:5] == "Spree"
		},
		Printed: printedLine("spree"),
	},
	{
		Slug: "convoke", Name: "Convoke", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Tap your creatures to help pay for a spell with convoke; each one pays for one mana of its color or one generic mana.",
		Rules:   []string{"702.51"},
		Probe:   tapCost("convoke"),
		Printed: printedKeyword("convoke"),
	},
	{
		Slug: "waterbend", Name: "Waterbend", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Tap your artifacts and creatures to help pay a waterbend cost, each one paying for one generic mana.",
		Rules:   []string{"701.67"},
		Probe:   tapCost("waterbend"),
		Printed: printedWords("waterbend"),
	},
	{
		Slug: "ward", Name: "Ward", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "When a permanent with ward becomes the target of an opponent's spell or ability, that spell or ability is countered unless its controller pays the ward cost.",
		Rules:   []string{"702.21"},
		Printed: printedLine("ward"),
	},
	{
		Slug: "equip", Name: "Equip", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Pay an Equipment's equip cost to attach it to a creature you control, as a sorcery.",
		Rules:   []string{"702.6"},
		Probe:   activated(func(ab effects.ActivatedAbility) bool { return ab.Equip }),
		Printed: printedLine("equip"),
	},
	{
		Slug: "crew", Name: "Crew", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Tap creatures with enough total power to turn a Vehicle into an artifact creature until end of turn.",
		Rules:   []string{"702.122"},
		Probe:   activated(func(ab effects.ActivatedAbility) bool { return ab.Cost.Crew > 0 }),
		Printed: printedLine("crew"),
	},
	{
		Slug: "typecycling", Name: "Landcycling and typecycling", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Discard a card with landcycling or another typecycling from your hand to search your library for a card of that type.",
		Rules:   []string{"702.29"},
		Printed: `(?im)^[a-z ]+cycling\b`,
	},
	{
		Slug: "morph", Name: "Morph and megamorph", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Cast a creature face down as a 2/2 for three mana, then turn it face up any time for its morph cost.",
		Rules:   []string{"702.37"},
		ADR:     "0082-casting-face-down-and-turning-face-up.md",
		Probe:   altCost("morph", "megamorph"),
	},
	{
		Slug: "disguise", Name: "Disguise and cloak", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Cast a creature face down as a 2/2 with ward 2, then turn it face up for its disguise cost.",
		Rules:   []string{"702.168"},
		ADR:     "0082-casting-face-down-and-turning-face-up.md",
		Probe:   altCost("disguise"),
		Printed: printedWords("cloak"),
	},
	{
		Slug: "manifest", Name: "Manifest", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Put the top card of your library onto the battlefield face down as a 2/2; if it's a creature card you may turn it face up for its mana cost.",
		ADR:     "0082-casting-face-down-and-turning-face-up.md",
		Printed: printedWords("manifest"),
	},
	{
		Slug: "unearth", Name: "Unearth", Kind: KindMechanic, Status: StatusImplemented,
		Summary:          "Return a creature from your graveyard to the battlefield with haste for one turn; it's exiled afterwards.",
		Rules:            []string{"702.82"},
		Probe:            activatedLabel("Unearth "),
		Phrases:          []string{"unearthed"},
		NoCatalogExample: "The one catalogued unearth card has a small caveat of its own, so none is listed as fully automated yet.",
	},
	{
		Slug: "embalm-eternalize", Name: "Embalm and eternalize", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Exile a creature card from your graveyard to create a token copy of it.",
		Rules:   []string{"702.128", "702.129"},
		Probe:   anyOf(activatedLabel("Embalm "), activatedLabel("Eternalize ")),
	},
	{
		Slug: "scavenge", Name: "Scavenge", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Exile a creature card from your graveyard to put +1/+1 counters equal to its power on a creature.",
		Rules:   []string{"702.96"},
		Probe:   activatedLabel("Scavenge "),
	},
	{
		Slug: "ninjutsu", Name: "Ninjutsu", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Return an unblocked attacker you control to your hand to put a creature with ninjutsu onto the battlefield tapped and attacking.",
		Rules:   []string{"702.49"},
		Probe:   activatedLabel("Ninjutsu "),
	},
	{
		Slug: "impending", Name: "Impending", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Cast a creature for its impending cost and it arrives as a non-creature with time counters, becoming a creature when they run out.",
		Rules:   []string{"702.176"},
		Probe:   altCost(effects.AltCostKeyImpending),
	},
	{
		Slug: "cascade", Name: "Cascade", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "When you cast a spell with cascade, reveal cards from your library until you hit a cheaper nonland card, which you may cast for free.",
		Rules:    []string{"702.85"},
		Mechanic: "cascade",
		Printed:  printedKeyword("cascade"),
		Examples: []string{"Bloodbraid Elf"},
	},
	{
		Slug: "storm", Name: "Storm", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "When you cast a spell with storm, copy it for each spell cast before it this turn.",
		Rules:    []string{"702.40"},
		ADR:      "0086-storm-and-the-turns-cast-order.md",
		Mechanic: "storm",
		Printed:  printedKeyword("storm"),
		Examples: []string{"Grapeshot"},
	},
	{
		Slug: "cumulative-upkeep", Name: "Cumulative upkeep", Kind: KindMechanic, Status: StatusPartial,
		Summary:          "At your upkeep a permanent with cumulative upkeep gets an age counter, and you pay its cost once per counter or sacrifice it.",
		Missing:          "Only mana costs work: cumulative upkeep that asks for life or a sacrifice isn't supported yet.",
		Rules:            []string{"702.24"},
		Issue:            567,
		Printed:          printedLine("cumulative upkeep"),
		Phrases:          []string{"cumulative upkeep"},
		NoCatalogExample: "The one catalogued card with cumulative upkeep has an unrelated caveat, so none is listed as fully automated yet.",
	},
	{
		Slug: "emblems", Name: "Emblems", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Planeswalker ultimates that give you an emblem with a lasting ability.",
		Rules:    []string{"114"},
		ADR:      "0064-emblems.md",
		Probe:    func(s effects.Spec) bool { return s.Emblem != nil },
		Examples: []string{"Elspeth, Sun's Champion"},
	},
	{
		Slug: "classes-cases", Name: "Classes and Cases", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Class enchantments that gain abilities as you level them up, and Cases that gain one once they're solved.",
		Rules:    []string{"716", "719"},
		ADR:      "0071-designations-that-switch-abilities-on.md",
		Probe:    designation(game.DesignationClassLevel, game.DesignationCaseSolved),
		Examples: []string{"Wizard Class"},
	},
	{
		Slug: "station", Name: "Station", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Tap another creature to put charge counters on a Spacecraft or Planet; enough counters switch on its other abilities.",
		Rules:   []string{"702.184"},
		ADR:     "0071-designations-that-switch-abilities-on.md",
		Probe: anyOf(
			activated(func(ab effects.ActivatedAbility) bool { return ab.Label == effects.StationLabel }),
			designation(game.DesignationChargeCounters),
		),
	},
	{
		Slug: "harness", Name: "Harness", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Harness a permanent to switch on the ability printed after its infinity symbol.",
		Rules:   []string{"701.64"},
		Probe:   designation(game.DesignationHarnessed),
	},
	{
		Slug: "hideaway", Name: "Hideaway", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "When a permanent with hideaway enters, you exile one of your top cards face down and may later play it for free.",
		Rules:    []string{"702.75"},
		ADR:      "0091-hideaway.md",
		Printed:  printedLine("hideaway"),
		Examples: []string{"Windbrisk Heights"},
	},
	{
		Slug: "preparation", Name: "Preparation", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "A prepared creature lets you cast a copy of the spell printed on its card, once, and is unprepared when you do.",
		Rules:   []string{"722"},
		ADR:     "0090-preparation-cards.md",
		Printed: printedWords("prepared"),
	},
	{
		Slug: "adventure", Name: "Adventures", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Cast a card's Adventure half, then cast the creature later from exile.",
		Rules:    []string{"715"},
		ADR:      "0034-multi-face-cards.md",
		Examples: []string{"Foulmire Knight"},
	},
	{
		Slug: "transform", Name: "Transforming", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Double-faced permanents that turn over to their other face when an effect transforms them.",
		ADR:     "0079-transforming-a-permanent.md",
		Printed: printedWords("transform"),
	},
	{
		Slug: "sieges", Name: "Battles and Sieges", Kind: KindMechanic, Status: StatusPartial,
		Summary:          "Battles enter with defense counters for an opponent to protect; defeat a Siege and you cast its back face.",
		Missing:          "The back face can be cast only in your own main phase of the turn the Siege is defeated, so a Siege defeated on another player's turn is lost.",
		Issue:            92,
		ADR:              "0034-multi-face-cards.md",
		Probe:            func(s effects.Spec) bool { return s.Battle != nil },
		Phrases:          []string{"Siege is defeated"},
		NoCatalogExample: "Every catalogued Siege carries this caveat, so none is listed as fully automated yet.",
	},
	{
		Slug: "amass", Name: "Amass", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Put +1/+1 counters on an Army you control, creating a 0/0 Army token first if you don't have one.",
		Rules:   []string{"701.47"},
		ADR:     "0087-amass.md",
		Printed: printedWords("amass"),
	},
	{
		Slug: "earthbend", Name: "Earthbend", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Turn a land you control into a creature with counters on it; it comes back tapped if it dies or is exiled.",
		ADR:     "0081-earthbend-and-object-keyed-delayed-triggers.md",
		Printed: printedWords("earthbend"),
	},
	{
		Slug: "proliferate", Name: "Proliferate", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Choose any number of permanents and players with counters and give each another counter of a kind it already has.",
		Printed: printedWords("proliferate"),
	},
	{
		Slug: "regeneration", Name: "Regeneration", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "The next time a regenerated permanent would be destroyed this turn, it's tapped and removed from combat instead.",
		Rules:   []string{"701.19"},
		Printed: printedWords("regenerate"),
	},
	{
		Slug: "clones", Name: "Clones and copies", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Permanents that enter as a copy of another, and spells and abilities that copy a spell or make token copies.",
		ADR:      "0043-copy-effects.md",
		Printed:  `(?i)\bas a copy of\b|\bcopy (?:target|that) [a-z ]*spell\b|\btoken that's a copy\b`,
		Examples: []string{"Clone", "Twincast"},
	},
	{
		Slug: "attack-taxes", Name: "Attack taxes", Kind: KindMechanic, Status: StatusImplemented,
		Summary: "Effects like Propaganda that make opponents pay for each creature that attacks you.",
		ADR:     "0080-attack-taxes.md",
		Probe:   func(s effects.Spec) bool { return len(s.AttackTaxes) > 0 },
	},
	{
		Slug: "player-protection", Name: "Hexproof and protection for players", Kind: KindMechanic, Status: StatusImplemented,
		Summary:  "Effects that give you hexproof or protection, so opponents can't target you or damage you.",
		ADR:      "0072-protection.md",
		Printed:  `(?i)\byou (?:have|gain) (?:hexproof|protection)\b`,
		Examples: []string{"Teferi's Protection"},
	},

	// ── Engine seams ─────────────────────────────────────────────
	//
	// docs/engine-seams.md's open table is generated from the entries
	// below that are not implemented, in this order (seams.go). Each
	// was re-checked against the code on 2026-09-24 (#1386); a row
	// that turned out fully closed stays here as implemented and its
	// prose moved to that doc's Closed seams list.
	{
		Slug: "abilities-granted-to-other-permanents", Name: "Abilities granted to other permanents", Kind: KindSeam, Status: StatusPartial,
		Summary:  "Effects that give other permanents a new activated, triggered or mana ability, such as Cryptolith Rite letting your creatures tap for mana.",
		Missing:  "Permanents can give other permanents mana and activated abilities, but not triggered abilities, and a spell can't grant an ability for a while yet.",
		Issue:    754,
		Tracked:  "#754 (related: #665, #669)",
		ADR:      "0093-abilities-granted-to-other-permanents.md",
		Mechanic: "an ability granted to another permanent",
		Examples: []string{"Cryptolith Rite", "Chromatic Lantern", "Necrotic Sliver"},
		Waiting: []string{
			"Urza's Saga", "Ultima, Origin of Oblivion", "Teferi's Talent",
		},
		Phrases:     []string{"gains the ability", "have the ability"},
		EngineNotes: "**static and attached grants of mana and activated abilities ship (ADR 0093 PRs 1-2).** A layer-6 static declares `StaticAbility.GrantAbilities` (card side: `effects.GrantAbilities`, `TribalAbilityGrant`, `GrantAbilitiesToAttached`) naming an `effects.AbilityGrant` bundle; the recipient carries `Characteristic.GrantedAbilities`, `CatalogAbilityKey` composes own + layered grants, the ability readers return own + intrinsic + granted with a stable `ref` per row, and the auto-tapper offers every acceptable ability of a permanent as mutually exclusive candidates with a creature's granted mana in the last-resort tier. Shipped on Cryptolith Rite, Chromatic Lantern, Gemhide and Manaweft Sliver, Necrotic Sliver, Rishkar, Jaheira, Insidious Roots, Great Divide Guide, The World Tree, Paradise Mantle, Squirrel Nest and Springleaf Parade. Still missing: granted-trigger cards and the Dionus / Agent of the Iron Throne migration (PR 3, the LKI harvest is already wired), duration grants from a resolving spell (PR 4 — Urza's Saga, Ultima; waits on #1545's ScopedEffect decision), and the client's left-click picker (the rows reach the existing ability menus already).",
	},
	{
		Slug: "riot", Name: "Riot", Kind: KindSeam, Status: StatusMissing,
		Summary:     "Riot lets a creature enter with a +1/+1 counter or with haste, its controller's choice.",
		Missing:     "Riot isn't implemented yet, so a creature with it gets neither the counter nor haste.",
		Issue:       1556,
		Tracked:     "#1556 (moved off #754 by ADR 0093)",
		Waiting:     []string{"Rhythm of the Wild"},
		Phrases:     []string{"riot"},
		Rules:       []string{"702.136a"},
		ADR:         "0093-abilities-granted-to-other-permanents.md",
		Unblocks:    1,
		EngineNotes: "primitive: riot (CR 702.136a) is a keyword with a CR 614.12 \"as this enters\" choice, and no such keyword exists — it is not in `canonicalKeywords`. Rhythm of the Wild was listed under abilities-granted-to-other-permanents, but ADR 0093's audit found it misattributed: GRANTING a keyword already works (a layer-6 keyword grant); what is missing is riot itself, plus the CR 614.12 look-ahead for \"abilities it would have on the battlefield\" as it enters. Re-checked 2026-09-24.",
	},
	{
		Slug: "all-activated-abilities-of", Name: "Having another card's activated abilities", Kind: KindSeam, Status: StatusMissing,
		Summary:     "Effects that give a permanent all the activated abilities of other cards, such as Marvin, Murderous Mimic.",
		Missing:     "A permanent can't yet gain all the activated abilities of another card.",
		Issue:       1557,
		Tracked:     "#1557 (moved off #754 by ADR 0093 Decision 10; needs its own ADR)",
		Waiting:     []string{"Marvin, Murderous Mimic"},
		Phrases:     []string{"all activated abilities"},
		ADR:         "0093-abilities-granted-to-other-permanents.md",
		Unblocks:    1,
		EngineNotes: "primitive: \"has all activated abilities of …\" (Marvin, Necrotic Ooze, Drana and Linvala, Hazel's Brewmaster) grants ANOTHER OBJECT'S text, computed each layer pass. ADR 0093's grant is a catalog bundle with a fixed key; this grant has no bundle key at all, so it is Decision 10's out-of-scope case and needs its own ADR. Re-checked 2026-09-24.",
	},
	{
		Slug: "extra-combats", Name: "Extra combat and main phases", Kind: KindSeam, Status: StatusMissing,
		Summary:  "Spells and abilities that give you an additional combat phase, usually followed by an additional main phase.",
		Missing:  "The turn's phases are fixed, so nothing can add another combat or main phase yet.",
		Issue:    753,
		Unblocks: 28,
		Waiting: []string{
			"Relentless Assault", "Aggravated Assault", "Seize the Day", "Karlach, Fury of Avernus",
			"Hellkite Charger", "Aurelia, the Warleader", "Full Throttle",
		},
		Phrases:     []string{"additional combat", "extra combat"},
		EngineNotes: "primitive: the turn's step sequence is fixed, so no effect can add a combat phase or a main phase after it (CR 500.8), and there is no per-turn combat ordinal for \"the first combat\" or a delayed trigger bound to \"that combat\". Planned with extra turns (\"Extra turns\") as one turn-machinery ADR; overlaps #717's second combat damage step. Re-checked 2026-09-24: the step sequence is still fixed.",
	},
	{
		Slug: "infect-wither-toxic", Name: "Infect, wither and toxic", Kind: KindSeam, Status: StatusImplemented,
		Summary:  "Damage that becomes -1/-1 counters on creatures, and poison counters on players.",
		Rules:    []string{"702.80", "702.90", "702.164"},
		ADR:      "0056-infect-wither-toxic.md",
		Probe:    hasKeyword(game.KeywordInfect, game.KeywordWither, game.KeywordToxic+" "),
		Examples: []string{"Blighted Agent", "Puncture Blast"},
		Phrases:  []string{"infect", "wither", "toxic"},
	},
	{
		Slug: "win-the-game", Name: "Winning the game by an effect", Kind: KindSeam, Status: StatusImplemented,
		Summary: "Cards that say you win the game, and effects like \"you can't lose the game\" or \"your opponents can't win the game\".",
		Rules:   []string{"104.2b", "104.3", "104.4a"},
		ADR:     "0057-win-and-lose-by-effect.md",
		Issue:   749,
		// A card that prints the result is the honest probe: the win
		// is a closure in a trigger or a replacement, and a gate is a
		// Spec field only on the permanents that print one.
		Printed:  `(?i)\b(wins? the game|can't (lose|win) the game)\b`,
		Examples: []string{"Platinum Angel", "Felidar Sovereign", "Laboratory Maniac"},
		Phrases:  []string{"win the game", "wins the game", "can't lose the game"},
	},
	{
		Slug: "mana-spend-riders", Name: "Mana that does something when it's spent", Kind: KindSeam, Status: StatusPartial,
		Summary:  "Cards that read the mana spent on a spell work: converge, sunburst, adamant, \"if it was paid with Treasure\", and mana that does something when it's spent — Cavern of Souls, Pyromancer's Goggles, Hall of the Bandit Lord.",
		Missing:  "Other permanents can't yet grant a spell something based on the mana spent on it: Lux Artillery's granted sunburst and Coin of Mastery's extra counters. Satoru, the Infiltrator doesn't yet count a creature cast without spending mana.",
		Issue:    1552,
		Tracked:  "#761 (record shipped), #1212 (source snapshot + entry-side read shipped), #1547 (spend riders shipped); granted readers open in #1552",
		ADR:      "0040-mana-pipeline.md",
		Unblocks: 0,
		Waiting: []string{
			"Lux Artillery", "Satoru, the Infiltrator", "Coin of Mastery",
		},
		Phrases:          []string{"which mana you spent", "spend its mana", "mana spent"},
		NoCatalogExample: "Every catalogued card that reads spent mana carries the strict-mana caveat, so none is listed as fully automated yet.",
		EngineNotes:      "primitive: **the record shipped in #761** — `StackItem.Paid` carries the tokens that paid, and converge, sunburst, adamant and \"if no mana was spent\" all read it (see Closed seams). **The SOURCE snapshot shipped in #1212** — `ManaToken.SourceKinds`, carried onto the permanent by `Card.Provenance` (CR 400.7d) and read through `game.ManaSpent`. **SPEND RIDERS shipped in #1547** (ADR 0040 amendment 2026-09-24): `ManaToken.Riders []ManaSpendRider` — data (kind, a spend filter in the restriction vocabulary, a production id), copied from `ManaAbilityShape.SpendRiders` at mint (through `PendingChoice.ManaRiders` for a pick), fired by `applyManaSpendRidersLocked` where a payment becomes a stack object (cast and activation, manual and auto-tap alike) and stamped `Applied` on `StackItem.Paid.Mana`. Readers: `spellCantBeCounteredLocked` (Cavern of Souls, Delighted Halfling, Boseiju), `applyCastEntryCountersLocked` (Biophagus), a layer-6 gather over `Card.Provenance` (Hall of the Bandit Lord), and trigger riders queued at the spend from a key registry (Pyromancer's Goggles, Scaled Nurturer, Path of Ancestry). Still missing: a rider GRANTED by another permanent over someone else's cast — Lux Artillery's sunburst, Coin of Mastery's per-artifact-mana counters — and Satoru's \"no mana was spent to cast them\" read on an entering creature. Not catalogued and one small step away: Opal Palace (an entry rider whose count reads the command-zone tally and whose filter is \"your commander\") and Generator Servant (haste until end of turn, a duration the haste rider does not carry).",
	},
	{
		Slug: "tap-another-permanent-cost", Name: "Tapping your other permanents as a cost", Kind: KindSeam, Status: StatusImplemented,
		Summary: "Fixed-count costs that tap other untapped permanents work on activated abilities and mana abilities.",
		Issue:   758,
		Rules:   []string{"118.3", "602.2b"},
		Probe: func(s effects.Spec) bool {
			for _, ab := range s.Activated {
				if ab.Cost.TapOthers != nil {
					return true
				}
			}
			for _, ab := range s.ManaAbilities {
				if ab.Cost.TapOthers != nil {
					return true
				}
			}
			return false
		},
		Examples: []string{"Springleaf Drum", "Heritage Druid", "Relic of Legends", "The Seriema"},
	},
	{
		Slug: "variable-count-tap-others-cost", Name: "Variable-count tap-others cost", Kind: KindSeam, Status: StatusImplemented,
		Summary: "An activated ability can tap X eligible permanents as a cost, with the number picked becoming its announced X.",
		Issue:   1421,
		ADR:     "0073-optional-additional-costs-and-the-cast-gate.md",
		Rules:   []string{"107.3", "118.3", "602.2b"},
		Probe: func(s effects.Spec) bool {
			for _, ab := range s.Activated {
				if game.TapOthersCountFromX(ab.Cost.TapOthers) {
					return true
				}
			}
			return false
		},
		Examples: []string{"Secluded Starforge", "Apothecary White"},
	},
	{
		Slug: "blocking-restrictions", Name: "Conditional blocking restrictions", Kind: KindSeam, Status: StatusImplemented,
		Summary:  "Rules like \"can't be blocked except by Walls\", \"creatures with power less than this creature's power can't block creatures you control\" and limits on how many creatures can block one attacker.",
		Rules:    []string{"509.1b"},
		Issue:    750,
		ADR:      "0045-combat-restrictions.md",
		Probe:    func(s effects.Spec) bool { return len(s.BlockRules) > 0 },
		Examples: []string{"Prowler's Helm", "Legolas Greenleaf"},
		Phrases:  []string{"can't be blocked except", "can't block"},
	},
	{
		Slug: "combat-wide-limits", Name: "Combat-wide attack and block limits", Kind: KindSeam, Status: StatusImplemented,
		Summary:  "Rules that limit how many creatures can attack or block in a whole combat, such as Silent Arbiter's \"no more than one creature can block each combat\" and Crawlspace's \"no more than two creatures can attack you each combat\".",
		Rules:    []string{"508.1c", "509.1b"},
		Issue:    1507,
		ADR:      "0045-combat-restrictions.md",
		Probe:    func(s effects.Spec) bool { return len(s.AttackLimits) > 0 },
		Examples: []string{"Silent Arbiter", "Crawlspace"},
		Phrases:  []string{"no more than one creature can", "no more than two creatures can"},
	},
	{
		Slug: "conditional-combat-limits", Name: "Conditional and per-player combat limits", Kind: KindSeam, Status: StatusImplemented,
		Summary: "Combat limits that apply only under a condition or to one player or permanent, such as Mirri, Weatherlight Duelist's \"each opponent can't block with more than one creature this combat\" and The Eternal Wanderer's \"no more than one creature can attack The Eternal Wanderer each combat\".",
		Rules:   []string{"508.1c", "509.1b"},
		Issue:   1534,
		ADR:     "0045-combat-restrictions.md",
		Probe: func(s effects.Spec) bool {
			for _, l := range s.AttackLimits {
				if l.While != nil || l.Scope == game.AttackLimitAttackingThis {
					return true
				}
			}
			return false
		},
		Examples: []string{"Mirri, Weatherlight Duelist"},
	},
	{
		Slug: "extra-turns", Name: "Extra turns", Kind: KindSeam, Status: StatusPartial,
		Summary:          "Spells and abilities that let a player take an extra turn.",
		Missing:          "The turn plan cannot queue or insert extra turns yet.",
		Issue:            753,
		Unblocks:         14,
		Waiting:          []string{"Time Warp", "Time Stretch", "Avatar Kuruk"},
		Phrases:          []string{"extra turn"},
		NoCatalogExample: "The landed identity foundation is internal turn machinery; no card can take an extra turn until the queue lands.",
		EngineNotes:      "ADR 0059's identity foundation has landed: `Turn.Seq` changes every turn, `Turn.Round` is display-only, `TurnsBegun` counts per seat, and every early turn ending uses one rotation seam. The remaining blocker is ADR 0059 sub-PR 2's turn plan and queued extra-turn machinery. Nothing in `internal/game` takes an extra turn yet.",
	},
	{
		Slug: "face-down-objects", Name: "Face-down cards", Kind: KindSeam, Status: StatusPartial,
		Summary:     "Morph, megamorph, disguise, manifest, cloak, foretell, hideaway, and effects that turn permanents face down all work.",
		Missing:     "Manifesting cards from a hand isn't supported, and a manifested card that also has morph can only be turned face up for its mana cost.",
		Issue:       886,
		Tracked:     "#886 (the S43 tracker; #95 and #658 closed)",
		ADR:         "0069-face-down-objects.md",
		Waiting:     []string{"Kozilek, the Broken Reality"},
		Probe:       altCost("morph", "megamorph", "disguise"),
		Examples:    []string{"Aerie Bowmasters"},
		EngineNotes: "**Nearly all of this row has shipped** — see Closed seams for #1194 (morph, megamorph, manifest, disguise, cloak), #1209 (turning a permanent face down), #1270 / #1271 (listed characteristics, CR 613.7f), #658 (foretell) and #1331 (hideaway); the object model is [ADR 0069](decisions/0069-face-down-objects.md). Re-checked 2026-09-24, two gaps are left. (1) `ManifestForEffect` (`game/battlefield_put.go`) manifests the top card of a LIBRARY and nothing else, so \"each manifest two cards from their hands\" (Kozilek, the Broken Reality) has no primitive. (2) The double offer for a manifested card that also has morph: `TurnFaceUpOffer` offers the manifest's mana-cost route and not the card's own morph cost. Voidmage Apprentice was listed here and is not waiting on anything any more — morph and the turned-face-up trigger both exist — so it is off the row. The audit's 26 only-blocker cards were the morph / manifest / disguise families that have since shipped, so Unblocks is 0.",
	},
	{
		Slug: "rooms", Name: "Rooms", Kind: KindSeam, Status: StatusPartial,
		Summary:     "Classes and Cases work; Rooms are the third kind of card that switches abilities on as it changes state.",
		Missing:     "Rooms can't be cast yet: unlocking doors and casting either half of a Room aren't supported.",
		Rules:       []string{"709.5"},
		Issue:       886,
		ADR:         "0071-designations-that-switch-abilities-on.md",
		Waiting:     []string{"Funeral Room // Awakening Hall", "Unholy Annex // Ritual Chamber"},
		Examples:    []string{"Wizard Class"},
		EngineNotes: "**Classes and Cases shipped** ([ADR 0071](decisions/0071-designations-that-switch-abilities-on.md), #757): one `ActiveWhen Designation` field on every printed-ability slot, evaluated by one predicate in the object-level accessors (`game/designations.go`), plus `Card.ClassLevel` (CR 716.2), `Card.Solved` (CR 719.3), the `LevelUp` / `ToSolve` constructors and `class_level` / `solved` on the wire. What is left is ROOMS (CR 709.5): the two unlocked flags, the `DoorUnlocked` gate (reserved, and refused at boot until it is built — re-checked 2026-09-24, `effects.Register` still panics on it), casting either half of a split card (CR 709.3, no issue of its own), and the unlock special action. Tracked for #886. The Class and Case cards this row used to list (Caretaker's Talent, Case of the Ransacked Lab, Cleric Class, Druid Class, Artist's Talent) wait on nothing here any more, and the audit's 32 only-blocker cards were mostly Classes, so Unblocks is 0.",
	},
	{
		Slug: "ability-suppression", Name: "Stopping abilities", Kind: KindSeam, Status: StatusPartial,
		Summary:     "\"Loses all abilities\" works on the permanent an Aura or Equipment is attached to, and abilities on the stack can be countered.",
		Missing:     "Making every creature lose its abilities, taking one keyword away from a group, and stopping enters abilities from triggering (Torpor Orb) aren't supported yet.",
		Issue:       1210,
		Tracked:     "#1210",
		Waiting:     []string{"Dress Down", "Archetype of Imagination", "Tishana's Tidebinder", "Torpor Orb"},
		Examples:    []string{"Darksteel Mutation", "Stifle"},
		EngineNotes: "**the stated primitive is built and the row text was stale.** `game/ability_removal.go` (S24, refined by #668/#669/#670 and [ADR 0067](decisions/0067-layer-dependency-ordering.md)) makes CR 613.1f authoritative OVER the catalog: `CatalogAbilityKey` returns `\"\"` for an object carrying `Characteristic.AbilitiesRemoved`, and every `Catalog*` hook treats an empty key as \"no entry\" — triggered, activated, mana, static, replacements, cost modifiers, untap permissions, and since #760 cast restrictions and since #1210 activation restrictions. `effects.LoseAllAbilities(keep…)` is the card-side declaration and Darksteel Mutation, Kenrith's Transformation and Kitesail Larcenist ship on it. The row's second clause (\"a later grant can re-add a stripped keyword\") describes CR 613.6, which is CORRECT and deliberate: the layer-6 bucket is timestamp-sorted, so a Rancor cast after a Darksteel Mutation really does grant trample. What is ACTUALLY left is one gap per card, and none of them is this one: (1) a SCOPED ability removal — `LoseAllAbilities` hardcodes `AppliesTo: AttachedToSource`, so \"creatures lose all abilities\" board-wide has no helper (one line on the returned struct, no engine change); (2) removal of ONE named keyword from a scoped group (\"creatures your opponents control lose flying\"); (3) ~~COUNTERING an activated or triggered ability on the stack, which nothing can do~~ — **closed by #1211**: Stifle, Tale's End, Voidslime and Disallow ship, and Tishana's Tidebinder still waits only on (1); (4) suppressing the TRIGGERING of abilities on entry (Torpor Orb), which is a replacement on the trigger event and not an ability removal at all. Re-checked 2026-09-24: `effects.LoseAllAbilities` still hardcodes `AppliesTo: AttachedToSource`, and nothing suppresses a trigger on entry.",
	},
	{
		Slug: "set-level-sacrifice-cost", Name: "Rules about a whole set of sacrificed permanents", Kind: KindSeam, Status: StatusPartial,
		Summary:     "Choices with a rule about the whole set, such as \"total mana value 4 or less\", work when picking cards from your library.",
		Missing:     "A sacrifice cost with a rule about the whole set, like \"three artifact tokens with different names\", can't be paid yet.",
		Issue:       998,
		Tracked:     "#998 (prompt half, closed)",
		Waiting:     []string{"Transmutation Font"},
		Examples:    []string{"Ao, the Dawn Sky"},
		EngineNotes: "cost component / prompt: since #747 a sacrifice clause pays a fixed count of N permanents, but its predicate judges each permanent alone, so a restriction on the SET (\"three artifact tokens with different names\") can't be written, and the enumerator's first-N payment would have to become a search for a valid set. A set-level `Validate` in the style of #682's search validator is the likely shape. Transmutation Font ships with a caveat for it. The PROMPT half is a shorter reach and was found by #952's re-triage: `game.ChooseCardsPrompt.Validate func(picked []Card) bool` already exists, is enforced on submit and inside `legal.EnumerateFor`, and `effects.b23TotalManaValueAtMost` is already written — but `effects.PutFromLibraryOntoBattlefield` and `effects.TakeFromLibraryToHand` build their `choose_cards` prompt without forwarding it, so \"any number of nonland permanent cards with total mana value 4 or less from among them\" cannot be stated over a look at the top seven. **The PROMPT half closed with #998** (see Closed seams): `PutFromLibraryOntoBattlefield.Validate` and `TakeFromLibraryToHand.Validate` forward the field, Ao, the Dawn Sky ships `full` on it, and what is left on this row is only the COST half — a sacrifice clause whose predicate judges each permanent alone, which Transmutation Font still waits on. Re-checked 2026-09-24: the sacrifice clause's predicate still judges each permanent alone.",
	},
	{
		Slug: "draw-step-tally", Name: "\"The first card you draw each draw step\"", Kind: KindSeam, Status: StatusPartial,
		Summary:     "Effects that change how many cards you draw work, such as Thought Reflection.",
		Missing:     "The game doesn't count draws within a draw step, so \"except the first card you draw\" is read as \"except every card you draw in your draw step\".",
		Issue:       1222,
		Tracked:     "#1222 (count half, closed)",
		Waiting:     []string{"Teferi's Ageless Insight", "Alhammarret's Archive"},
		Examples:    []string{"Thought Reflection"},
		EngineNotes: "**the COUNT shipped** (#1222, [ADR 0013 §5ab](decisions/0013-replacement-effects.md)): `ReplacementEvent.DrawCount`, base one per instruction (CR 121.2), so Thought Reflection and Alhammarret's Archive are in the catalog. What is left is the OTHER half of the row — no per-draw-step draw tally on a player, so \"except the first one you draw in each of your draw steps\" is read as \"except any draw in your own draw step\" (Notion Thief's declared simplification, now shared with the Archive). Teferi's Ageless Insight prints only that clause, so it waits on the tally alone. Re-checked 2026-09-24: no per-draw-step tally on `PlayerTurnTally`.",
	},
	{
		Slug: "ability-cost-modification", Name: "Cheaper abilities and granted costs", Kind: KindSeam, Status: StatusPartial,
		Summary:     "Effects that make activated abilities cost less work, such as Boom Scholar's.",
		Missing:     "A discount that can't take an ability below one mana (Training Grounds), and granting spells an alternative cost like prowl, aren't supported yet.",
		Issue:       302,
		Tracked:     "#302 (batch issue; #1184 shipped the ability half)",
		Waiting:     []string{"Training Grounds", "Hunting Velociraptor"},
		Examples:    []string{"Boom Scholar"},
		EngineNotes: "primitive: **narrowed twice.** #1184 put activations through the CR 601.2f pass for BOARD modifiers (`CostModifier.Activations`, Boom Scholar), and #1296 gave an ability its OWN cost clause (`ActivatedAbilityShape.CostModifiers` — \"This ability costs {1} less to activate …\", including clauses that read the ability's target; see Closed seams). **Still open:** a board reduction with a FLOOR on someone else's abilities — Training Grounds' and Power Artifact's \"can't reduce the mana in that cost to less than one mana\" (`CostFloor` is the Trinisphere shape, a total-mana minimum, not a per-reduction cap) — and granted alternative costs for spells (Hunting Velociraptor)",
	},
	{
		Slug: "search-replacement", Name: "Changing how a library is searched", Kind: KindSeam, Status: StatusMissing,
		Summary:     "Effects like Aven Mindcensor that change what an opponent's library search can find.",
		Missing:     "A library search always sees the whole library; nothing can limit it to the top few cards yet.",
		Issue:       304,
		Tracked:     "#304 (batch issue)",
		Waiting:     []string{"Aven Mindcensor"},
		EngineNotes: "primitive: no search-replacement or depth limit on a library search — no `ReplacementEventKind` watches a search (re-checked 2026-09-24 against `game/replacements.go`). The non-owner-searcher half closed with #1230 (see Closed seams), which took Bribery off it. The library-top cast path used to be listed here, then had its own row, and closed with #765 (see Closed seams); the mill-as-a-replacement-event kind was the other half of this row and closed with #569 (see Closed seams), which took Bruvac the Grandiloquent off it.",
	},
	{
		Slug: "paradigm", Name: "Paradigm", Kind: KindSeam, Status: StatusMissing,
		Summary:     "After you first resolve a paradigm spell, you may cast a copy of it for free at the start of each of your first main phases.",
		Missing:     "Paradigm isn't implemented: the spell goes to your graveyard and never offers the free copies.",
		Issue:       1112,
		Tracked:     "#1112 (deck tracker)",
		Waiting:     []string{"Echocasting Symposium", "Improvisation Capstone"},
		Phrases:     []string{"paradigm"},
		EngineNotes: "keyword: three clauses and no shape for any of them. The spell EXILES ITSELF instead of going to the graveyard; a per-NAME \"have you resolved a spell with this name yet\" record is kept for the rest of the game; and from then on, at the beginning of each of the controller's first main phases, they are offered a cast of a **copy** of the card out of exile for no mana. The third clause is the one with nothing behind it: every cast permission in the engine ([ADR 0066](decisions/0066-granted-cast-and-play-permissions.md)) opens the CARD, and \"cast a copy of a card in exile\" exists nowhere — `CreateTokenCopy` mints a permanent, not a spell on the stack. The other two are reachable on their own (`CastPermission.ExileOnResolution` is the self-exile's sibling, and a per-name record is a tally), but a card shipped with only those would be all drawback and no payoff, so both Lessons defer the whole keyword.",
	},
	{
		Slug: "suspended-card-abilities", Name: "Abilities of a suspended card", Kind: KindSeam, Status: StatusPartial,
		Summary:     "Abilities that work from your hand, graveyard or exile — cycling, unearth, the Spirit Guides — all work.",
		Missing:     "An ability that works only while its card is suspended, like Greater Gargadon's, can't be activated yet.",
		Issue:       1221,
		Tracked:     "#655, #1221, #1228, #1284 (all closed)",
		ADR:         "0062-abilities-and-special-actions-from-the-hand.md",
		Waiting:     []string{"Greater Gargadon"},
		Examples:    []string{"Reassembling Skeleton", "Simian Spirit Guide"},
		EngineNotes: "Was \"Ability activatable from a non-battlefield zone\". **The ACTIVATED half closed with #1221, the MANA half with #1228, and the graveyard-return \"enters tapped\" gap with #1284** (see Closed seams) — `ManaAbilityShape.Zones` and `ManaAbilityCost.ExileSelf` carry the Spirit Guides, read by the activation path, the enumerator's walk, the view's per-seat stamp and the auto-tapper's second gather, and `ReturnFromGraveyard.Tapped` closed Reassembling Skeleton and Drownyard Temple. What is left is not a zone at all: a per-instance exile GRANT (Greater Gargadon while suspended) has no shape — [ADR 0062](decisions/0062-abilities-and-special-actions-from-the-hand.md) open question 2, unchanged. The audit's 45 only-blocker cards were the hand / graveyard activations that have shipped, so Unblocks is 0.",
	},
	{
		Slug: "resolution-lki", Name: "Remembering a permanent after it leaves", Kind: KindSeam, Status: StatusImplemented,
		Summary:  "Abilities that read a creature's power or other details when they resolve use what it looked like as it left, even if it's gone by then.",
		Rules:    []string{"608.2h"},
		ADR:      "0018-triggers-on-the-stack.md",
		Examples: []string{"Cream of the Crop", "Tribute to the World Tree", "Claustrophobia"},
	},
	{
		Slug: "departed-damage-source-keywords", Name: "Lifelink and deathtouch from a source that has left", Kind: KindSeam, Status: StatusImplemented,
		Summary:  "Damage dealt by a creature that has already left the battlefield still counts its last known power, lifelink and deathtouch.",
		ADR:      "0056-infect-wither-toxic.md",
		Examples: []string{"Warstorm Surge"},
	},
	{
		Slug: "becomes-plotted", Name: "\"When this card becomes plotted\"", Kind: KindSeam, Status: StatusImplemented,
		Summary:  "Abilities that trigger when you plot a card, whether you plotted it or an effect did.",
		Rules:    []string{"702.170"},
		ADR:      "0062-abilities-and-special-actions-from-the-hand.md",
		Examples: []string{"Longhorn Sharpshooter", "Aloe Alchemist"},
	},
	{
		Slug: "commander-ninjutsu", Name: "Commander ninjutsu", Kind: KindSeam, Status: StatusImplemented,
		Summary:  "Ninjutsu works from your hand, and commander ninjutsu also works from the command zone.",
		Rules:    []string{"702.49c"},
		ADR:      "0020-activated-abilities.md",
		Examples: []string{"Yuriko, the Tiger's Shadow"},
	},
	{
		Slug: "emblem-activation-timing", Name: "Emblems that change when you can activate", Kind: KindSeam, Status: StatusImplemented,
		Summary: "An emblem can let you activate loyalty abilities of your planeswalkers at instant speed, on any player's turn.",
		Issue:   1275,
		Rules:   []string{"114.3", "606.3"},
		ADR:     "0066-granted-cast-and-play-permissions.md",
		Probe: func(s effects.Spec) bool {
			return s.Emblem != nil && len(s.Emblem.ActivationTimings) > 0
		},
		Examples: []string{"Teferi, Temporal Archmage"},
	},

	// Seams that have fully closed. They stay so the page can say so;
	// their history is in engine-seams.md's Closed seams list.
	{
		Slug: "dice-and-coins", Name: "Dice rolls and coin flips", Kind: KindSeam, Status: StatusImplemented,
		Summary: "Cards that roll dice, flip coins or choose at random, with results that undo correctly.",
		Rules:   []string{"705", "706"},
		ADR:     "0054-dice-rolls-and-coin-flips.md",
		Printed: `(?i)\broll (?:a|two|three) d\d+|\bflips? (?:a|two|three) coins?\b|\bat random\b`,
	},
	{
		Slug: "state-dependent-abilities", Name: "Abilities that depend on your hand, life or graveyard", Kind: KindSeam, Status: StatusImplemented,
		Summary:  "Continuous abilities that depend on hand size, life total, graveyard contents, attacking or spells cast this turn stay up to date.",
		Examples: []string{"Serra Ascendant", "Stoic Sphinx", "Psychosis Crawler"},
	},
	{
		Slug: "pick-from-exile-or-graveyard", Name: "Choosing cards from exile or a graveyard", Kind: KindSeam, Status: StatusImplemented,
		Summary:  "Effects that have you choose among cards in exile or in a graveyard and then do something with them.",
		Examples: []string{"Currency Converter", "Malcolm, Alluring Scoundrel", "Helm of Obedience"},
	},
	{
		Slug: "exiled-with-and-departed-counters", Name: "\"Exiled with this\" and counters on a departed permanent", Kind: KindSeam, Status: StatusImplemented,
		Summary:  "Cards that remember what they exiled, and abilities that read the counters a permanent had when it left.",
		Examples: []string{"The Ozolith", "Valakut Exploration", "Currency Converter"},
	},
}
