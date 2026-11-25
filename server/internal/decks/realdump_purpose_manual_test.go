package decks

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// realdump_purpose_manual_test.go — ADR 0126 §6's manual audit of
// declared purposes, against the real Scryfall dump. Gated on
// CMDCTRL_SCRYFALL_DUMP like realdump_manual_test.go, for the same
// reason: CI has no dump.
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go test ./internal/decks/ -run RealDumpPurpose -v
//
// Three checks:
//
//   - TestRealDumpPurposeCuratedSpellList: curatedInstantsAndSorceries
//     (purpose_test.go) is exactly the curated decks' instants and
//     sorceries by Scryfall's type lines, so TestCuratedDeckPurposes
//     covers every one of them offline.
//   - TestRealDumpPurposeLandsUntapped: curatedLandsUntapped
//     (purpose_test.go) is exactly the curated cards that declare Lands
//     and whose search puts a card onto the battlefield without
//     "tapped" (ADR 0136 §2). It lists, without failing, the other
//     catalog cards that declare Lands, read the same way, and declare
//     no LandsUntapped.
//   - TestRealDumpPurposeAudit: every catalog card whose oracle text
//     reads as a board wipe declares a Sweep somewhere (owner decision
//     2: every catalog wipe), or is on reviewedNotAWipe with the reason.
//     It also LISTS, without failing, the catalog cards whose text reads
//     as a draw, a tutor or a land search and that declare no purpose:
//     a review aid, since ADR 0126 declares those for the curated decks
//     alone. It lists the same way the cards whose text gives a target
//     player something ("target player draws / creates / gains") or
//     deals a fixed amount of damage to a creature or any target, and
//     that declare no target entry: ADR 0126's amendment of 2026-10-08,
//     whose PR 5 works through that list.

// The oracle-text readings. Deliberately loose: a false positive is a
// line on reviewedNotAWipe, a false negative is a wipe nobody declared.
var (
	sweepText = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(destroy|exile) (all|each) `),
		regexp.MustCompile(`(?i)damage to each (other )?(\w+ )?creature`),
		regexp.MustCompile(`(?i)\b(all|each) (other )?(non-?\w+ )?creatures?\b[^.]*\bgets? -`),
		regexp.MustCompile(`(?i)creatures (your opponents control|you don't control|target (player|opponent) controls) get -`),
		regexp.MustCompile(`(?i)\breturn (all|each) [^.]*\bto (its|their) owners?'s? hands?`),
		regexp.MustCompile(`(?i)\bsacrifices? all\b`),
		regexp.MustCompile(`(?i)change "target" in its text to "each"`),
	}
	drawText     = regexp.MustCompile(`(?i)\b(you )?draws? (a|an|one|two|three|four|five|six|seven) cards?\b`)
	tutorText    = regexp.MustCompile(`(?i)search your library for`)
	landToBfText = regexp.MustCompile(`(?i)search your library for [^.]*\bland[^.]*onto the battlefield`)
	// ADR 0136 §2: a library search that puts a card onto the
	// battlefield, and whether "tapped" follows.
	searchToBfText = regexp.MustCompile(`(?i)search your library for [^.]*?onto the battlefield( tapped)?`)
	instantSorcer  = regexp.MustCompile(`\b(Instant|Sorcery)\b`)
	// ADR 0126's amendment of 2026-10-08: a target that is given
	// something, and burn at a creature or any target.
	giftText = regexp.MustCompile(`(?i)\btarget (player|opponent) (draws|creates|gains)\b`)
	burnText = regexp.MustCompile(`(?i)deals \d+ damage to (any (other )?target|(up to \w+ )?(another )?target ([\w,-]+ )*?(creature|planeswalker))`)
)

// declaresTargetEntry reports whether a spec declares a target entry
// (Purpose.Targets) in any slot.
func declaresTargetEntry(s effects.Spec) bool {
	ps := []game.Purpose{s.Purpose}
	modes := func(m *game.ModeSpec) {
		if m != nil {
			for _, o := range m.Options {
				ps = append(ps, o.Purpose)
			}
		}
	}
	modes(s.Modes)
	for _, a := range s.AlternativeCosts {
		ps = append(ps, a.Purpose)
	}
	for _, a := range s.Activated {
		ps = append(ps, a.Purpose)
		modes(a.Modes)
	}
	for _, t := range s.Triggered {
		ps = append(ps, t.Purpose)
		modes(t.Modes)
	}
	for _, p := range ps {
		if p.Targets != nil {
			return true
		}
	}
	return false
}

// reviewedNotAWipe is every catalog card the sweep reading flags that
// declares no Sweep, with why. Each was read.
var reviewedNotAWipe = map[string]string{
	"Druid of Purification": "each player may choose one artifact or enchantment its controller doesn't control: at most one permanent a player, not a board",
	"Kaya, Geist Hunter":    "its -6 exiles the cards in graveyards, not permanents",
	"Ill-Timed Explosion":   "damages each creature only if its controller discards two cards, for the greatest mana value among them: no fixed amount a sweep can declare, and none at all when declined",
	"Abyssal Harvester":     "exiles only the Nightmare tokens its controller controls, not a board",
	"Hex Magic":             "exiles the cards in its caster's own hand, not a board",
	"Retract":               "returns only its caster's own artifacts to hand, not a board",
	"Armageddon":            "destroys lands, which no sweep class holds alone",
	// #1997: its end-step and leaves triggers destroy only the Tombspawn
	// it created itself, every player's, never another permanent. #5802:
	// Saproling Burst's leaves trigger is the same link, destroying only
	// the Saprolings it created.
	"Tombstone Stairwell":                      "destroys only the tokens it created, not a board",
	"Saproling Burst":                          "destroys only the tokens it created, not a board",
	"All Suns' Dawn":                           "\"Exile All Suns' Dawn\" is the spell exiling itself; it returns cards from your graveyard to hand, no permanents",
	"Bag of Holding":                           "returns the cards it exiled, not permanents",
	"Northampton Farm":                         "returns the cards it exiled to hand, not permanents",
	"Baneblade Scoundrel // Baneclaw Marauder": "-1/-1 only to the creatures blocking it",
	"Battle-Scarred Goblin":                    "1 damage only to the creatures blocking it",
	"Knight of Valor":                          "-1/-1 only to the creatures without flanking blocking it",
	"Bazaar of Wonders":                        "exiles graveyards",
	"Beyeen Veil // Beyeen Coast":              "-2/-0 kills nothing",
	"The Apprentice's Folly":                   "sacrifices only its controller's own Reflection tokens, not a board",
	"Bewildering Blizzard":                     "-3/-0 kills nothing",
	"Hampering Snare":                          "-2/-0 kills nothing",
	"Marsh Gas":                                "-1/-0 kills nothing",
	"Mouth of the Storm":                       "-3/-0 until your next turn kills nothing",
	"Call Forth the Tempest":                   "damage equal to spells cast this turn: no fixed amount or X for a sweep to declare",
	"Demonic Consultation":                     "exiles cards from your own library, no permanents",
	"Council's Judgment":                       "exiles the permanents with the most votes, one per voter: removal, not a wipe",
	"Crypt Incursion":                          "exiles a graveyard",
	"Dauntless Scrapbot":                       "exiles the opponents' graveyards, no permanents",
	"Eliminate the Impossible":                 "-2/-0 kills nothing",
	"Espers to Magicite":                       "exiles the opponents' graveyards, no permanents",
	"Eye of Singularity":                       "destroys only permanents sharing a name, which no class says",
	"Glorious End":                             "ends the turn",
	"Golden Argosy":                            "exiles only the creatures that crewed it and returns them tapped at the next end step: a flicker of its own crewers, not a wipe",
	"Jace, Reality Sculptor":                   "its 0 exiles all but the bottom card of each opponent's library, no permanents",
	"Jace, the Mind Sculptor":                  "exiles a library",
	"Karn Liberated":                           "restarts the game",
	"Kaya's Guile":                             "each opponent sacrifices one creature: an edict, not a wipe",
	"Last Laugh":                               "a ping engine that fires only when a permanent dies",
	"Living Death":                             "a symmetric mass reanimation (purpose_test.go's noPrintedAmount)",
	"Living End":                               "Living Death with suspend: a symmetric mass reanimation",
	"Mandate of Peace":                         "ends combat and exiles the stack",
	"Merciful Mandate":                         "untaps your creatures and may end combat, exiling the stack",
	"Mindwhisker":                              "-1/-0 to opponents' creatures under threshold kills nothing",
	"Morningtide":                              "exiles graveyards, no permanents",
	"Obeka, Brute Chronologist":                "ends the turn",
	"Opposition Agent":                         "exiles the cards an opponent finds while searching, not permanents (#3998)",
	"Rest in Peace":                            "exiles graveyards",
	"Scavenger Grounds":                        "exiles graveyards",
	"Shadow of the Enemy":                      "exiles the creature cards of one graveyard, no permanents",
	"Soul-Guide Lantern":                       "exiles graveyards",
	"Soulblast":                                "sacrifices only its caster's creatures, as a cost (#2097); the heuristic declines the cast",
	"Sundial of the Infinite":                  "ends the turn",
	"Synthetic Destiny":                        "exiles only its caster's creatures, then returns that many from the library",
	"Time Stop":                                "ends the turn",
	"Trailblazer's Torch":                      "2 damage only to the creatures blocking the equipped creature",
	"Ugin's Binding":                           "a single target, and an exile from the graveyard",
	"Ugin, the Spirit Dragon":                  "the -X sweep is not registered (its caveat says so)",
	"V'Ger, the Intruder":                      "-1/-0 to opponents' creatures kills nothing",
	"Vault 11: Voter's Dilemma":                "destroys only the creatures with the most votes, one per voter at most: removal, not a wipe",
	"Volcanic Torrent":                         "damage equal to spells cast this turn, to opposing creatures and planeswalkers: no fixed amount or X for a sweep to declare",
	"Watchdog":                                 "-1/-0 to attackers kills nothing",

	// #5479: the Invasion's enter trigger exiles its controller's hand and
	// draws that many; the back face burns, but never every creature.
	"Invasion of Kaldheim // Pyre of the World Tree": "exiles the cards in its controller's own hand and draws that many, not a board",

	// #5479: "any number of target Equipment" are only those attached to
	// the one creature it steals.
	"Awaken the Sleeper": "destroys only Equipment attached to the one creature it steals, not a board",

	// #5482: BRO batch 16.
	"Calamity's Wake": "exiles every graveyard and bars noncreature spells for the turn: it touches no permanent",
	"Urza's Command":  "one of its two chosen bullets gives opposing creatures -2/-0 until end of turn, which removes nothing",

	// #5487: its end-step sacrifice names only the tokens its own attack made.
	"Stangg, Echo Warrior": "sacrifices only the tokens its own attack trigger created, not a board",

	// #5482: one of six modes, chosen three at a time on an enters-or-attacks trigger.
	"Mishra, Lost to Phyrexia": "its -1/-1 to opposing creatures is one of six modes, three chosen each time: a mode of a trigger, not a board wipe",

	// #5495: VOW batch 6.
	"Lantern of the Lost": "exiles the cards in every graveyard and draws a card: it touches no permanent",

	// #5493: NEO batch 18.
	"Behold the Unspeakable // Vision of the Unspeakable": "chapter I gives opposing creatures -2/-0 until your next turn, which removes nothing",

	// #5499: AFR batch 14.
	"Asmodeus the Archfiend": "its \"return all cards exiled with Asmodeus to their owner's hand\" moves only the cards its own draw replacement exiled face down, never a permanent on the board",

	// S82 AFR batch 16.
	"Wild-Magic Sorcerer": "gives the first spell you cast from exile each turn cascade: no permanents are affected, only a cascade offer",

	// S82 RPR batch 15.
	"Relic of Progenitus": "its second ability exiles the cards in every graveyard and draws a card: it touches no permanent",

	// S82 RPR batch 31.
	"Angel of Glory's Rise": "exiles only the Zombies, one creature type, then returns Human creature cards from its controller's graveyard: a type-specific exile, not a board wipe",

	// S82 BNS batch 41.
	"Waker of Waves": "a static -1/-0 on opposing creatures: it removes no creature from the battlefield, so it is not a board wipe",

	// S82 CON batch 2.
	"Cumber Stone": "a static -1/-0 on opposing creatures: it removes no creature from the battlefield, so it is not a board wipe",
	// S82 DGM batch 1.
	"Haunter of Nightveil": "a static -1/-0 on opposing creatures: it removes no creature from the battlefield, so it is not a board wipe",

	// S82 BNS batch 49.
	"Uncomfortable Chill": "-2/-0 until end of turn to opposing creatures kills nothing, then draws a card",

	// S82 GAP batch 34.
	"Ruination": "destroys the nonbasic lands, which no sweep class holds alone",

	// S82 GAP batch 35.
	"Yggdrasil, Rebirth Engine": "exiles only the creature cards in its controller's graveyard, which are no permanents",
	// S82 BOK batch 1 (#5602).
	"Yukora, the Prisoner": "its leaves trigger sacrifices only its controller's own non-Ogre creatures: a drawback, not a board",

	// #5521: RNA batch 1.
	"Azorius Skyguard": "a static -1/-0 on opposing creatures: it removes no creature from the battlefield, so it is not a board wipe",
	// S82 MBS batch 1 (#5576).
	"Turn the Tide": "-2/-0 until end of turn to opposing creatures kills nothing",
	// S82 GAP batch 15.
	"Summary Dismissal": "exiles the other spells and counters the abilities on the stack: it touches no permanent",
	// S82 GAP batch 41.
	"Abstergo Entertainment": "its last ability returns a historic card then exiles the cards in every graveyard: it touches no permanent",

	// S82 GAP batch 68 (#5948).
	"Ravages of War": "destroys lands, which no sweep class holds alone (Armageddon's reading)",
	// S82 GAP batch 69 (#5948).
	"Bottled Cloister": "exiles only its controller's own hand face down on an opponent's upkeep and returns it on theirs: no permanent leaves the battlefield",
	// S82 GAP batch 80.
	"Ghostway": "exiles only the creatures its caster controls and returns them at the next end step: a blink of the caster's own board, not a wipe",
	// S82 GAP batch 107 (#5948).
	"Faerie Artisans": "exiles only the tokens it created itself (\"all other tokens created with this creature\"), never another permanent: not a board",

	// S82 GAP batch 114.
	"Day's Undoing": "shuffles hands and graveyards into libraries and ends the turn: it exiles the stack, not permanents",

	// S82 GAP batch 64 (#5948).
	"Shaun, Father of Synths": "its leaves trigger exiles only the Synth tokens its own controller controls: a cleanup of its own copies, not a board",

	// S82 GAP batch 138 (#5948).
	"Denizen of the Deep": "returns only its controller's own other creatures to hand, a cost to itself, not a board",
	"Worldgorger Dragon":  "exiles only its controller's own other permanents and brings them back when it leaves, not a board",
	"Floodgate":           "its leaves trigger deals half the Islands its controller controls to each nonblue creature without flying: no fixed amount or X for a sweep to declare",

	// S82 GAP batch 156 (#5948).
	"Ib Halfheart, Goblin Tactician": "4 damage only to the creatures blocking the Goblin it sacrifices: a blocked attacker's blockers, not a board",
	// S82 GAP batch 197 (#5948).
	"Bile Blight":      "targets one creature and shrinks it and the creatures sharing its name: a name-bound removal, not a board",
	"The Horus Heresy": "chapter III destroys only the creatures each player chose, one per player: at most four, never a board",

	// S82 GAP batch 212 (#5948).
	"Elder Brain": "exiles only the attacked player's hand, then they draw that many: it touches no permanent",

	// S82 GAP batch 222 (#5948).
	"Sea Hag // Aquatic Ingress": "-4/-0 to the opponents' creatures until end of turn kills nothing",

	// S82 GAP batch 230 (#5948).
	"Lae'zel's Acrobatics": "exiles only the nontoken creatures its caster controls and returns them (at once on a high roll, then again at the next end step): a blink of the caster's own board, not a wipe",

	// S82 GAP batch 231 (#5948).
	"Cloudkill": "-X/-X to every creature where X is the greatest mana value of a commander its caster owns on the battlefield or in the command zone: no fixed amount a sweep can declare, and zero with no commander",

	// S82 GAP2 batch 2 (#5948).
	"Orzhov Pontiff": "its -1/-1 to the creatures its controller doesn't control is one of two modes of an enters-or-haunted-dies trigger: a mode of a trigger, not a board wipe",

	// S82 GAP2 batch 07 (#5948).
	"Fires of Mount Doom": "2 damage to one target creature, the Equipment on it destroyed, and 2 damage to each PLAYER when a card is played: no creature but the target is touched",

	// S82 old sets, batch 005 (own group).
	"Nicol Bolas, the Ravager // Nicol Bolas, the Arisen": "its -12 exiles cards from one library, not permanents; the 10 damage hits one target",

	// #5948 Q3, the secret choice cards.
	"Call to the Void": "destroys only the creatures the players secretly chose, at most one of their own and one of someone else's each: a handful of chosen creatures, not a board",

	// S82 OLD batch 023 (#5462).
	"Leveler": "its enters trigger exiles the cards in its controller's own library, not permanents: not a board wipe",

	// S82 older sets batch 35 (#5501).
	"Sanctifier en-Vec": "its enters trigger exiles the black and red cards from graveyards, no permanents, and its replacement only redirects what would go to a graveyard",

	// S82 OLD batch 073 (#5506), roll-up old-4.
	"Arcane Artisan": "when it leaves, it exiles only the tokens it created itself, not anyone else's permanents: not a board wipe",

	// S82 OLD batch 058 (#5504).
	"Author of Shadows": "its enters trigger exiles the opponents' graveyards, no permanents: not a board wipe",

	// S82 OLD batch 071 (#5505).
	"Valki, God of Lies // Tibalt, Cosmic Impostor": "Tibalt's -8 exiles every graveyard, not permanents",
	// S82 OLD batch 083 (#5507).
	"Nevinyrral, Urborg Tyrant": "a 3/6 creature cast to be a creature: the destroy-all is the optional {1} reflexive trigger of its dies trigger, which nobody can plan a turn around",
	"Szat's Will":               "one mode is an edict (each opponent sacrifices one creature, the one with the greatest power) and the other exiles graveyards: no board is swept",

	// S82 OLD batch 109 (#5511).
	"Discontinuity": "ends the turn: it exiles the stack and wears off effects, but sweeps no permanents",
	"Turn to Slag":  "5 damage to one creature and a destroy of the Equipment attached to it: one target's gear, not a board",

	// S82 OLD batch 110 (#5511).
	"Idol of Endurance": "exiles only the controller's own small creature cards from their graveyard, no permanents: not a board wipe",

	// S82 OLD batch 152 (#5517).
	"Ghastly Conscription": "exiles creature cards from one graveyard and manifests them: it sweeps no permanents, it puts creatures onto the battlefield",

	// S82 OLD batch 157 (#5518).
	"Thought Distortion": "exiles noncreature, nonland cards from one opponent's hand and graveyard, no permanents: not a board wipe",
}

// anyPurpose reports whether a spec declares a purpose in any slot.
func anyPurpose(s effects.Spec) (declared, sweep bool) {
	note := func(zero bool, hasSweep bool) {
		if !zero {
			declared = true
		}
		if hasSweep {
			sweep = true
		}
	}
	// A wipe declared under Kicked or FaceUp is a declared wipe (ADR
	// 0126's amendment of 2026-10-15, #3762): DeclaresSweep reads them.
	// What a spell answers (#2877) is no amount, so a spell slot that
	// declares only Answers still lists as undeclared here.
	note(!declaresBeyondAnswers(s.Purpose), s.Purpose.DeclaresSweep())
	if s.Modes != nil {
		for _, o := range s.Modes.Options {
			note(!declaresBeyondAnswers(o.Purpose), !o.Purpose.Sweep.IsZero())
		}
	}
	for _, a := range s.AlternativeCosts {
		note(!declaresBeyondAnswers(a.Purpose), !a.Purpose.Sweep.IsZero())
	}
	for _, a := range s.Activated {
		note(a.Purpose.IsZero(), !a.Purpose.Sweep.IsZero())
		if a.Modes != nil {
			for _, o := range a.Modes.Options {
				note(o.Purpose.IsZero(), !o.Purpose.Sweep.IsZero())
			}
		}
	}
	for _, t := range s.Triggered {
		note(t.Purpose.IsZero(), !t.Purpose.Sweep.IsZero())
		if t.Modes != nil {
			for _, o := range t.Modes.Options {
				note(o.Purpose.IsZero(), !o.Purpose.Sweep.IsZero())
			}
		}
	}
	return declared, sweep
}

// loadDumpForPurpose loads the dump or skips.
func loadDumpForPurpose(t *testing.T) *cards.Index {
	t.Helper()
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	return idx
}

// printed returns a card's type line and oracle text, every face
// joined.
func printed(c cards.Card) (typeLine, text string) {
	types, texts := []string{c.TypeLine}, []string{c.OracleText}
	for _, f := range c.CardFaces {
		types = append(types, f.TypeLine)
		texts = append(texts, f.OracleText)
	}
	return strings.Join(types, " // "), strings.Join(texts, "\n")
}

func TestRealDumpPurposeCuratedSpellList(t *testing.T) {
	idx := loadDumpForPurpose(t)
	want := map[string]bool{}
	for _, d := range All() {
		for _, c := range d.Cards() {
			if c.Basic {
				continue
			}
			id, err := uuid.Parse(c.OracleID)
			if err != nil {
				t.Fatalf("%s: %v", c.Name, err)
			}
			card, ok := idx.FindByOracleID(id)
			if !ok {
				t.Errorf("%s is not in the dump", c.Name)
				continue
			}
			if tl, _ := printed(card); instantSorcer.MatchString(tl) && !strings.Contains(tl, "Land") {
				want[c.Name] = true
			}
		}
	}
	got := map[string]bool{}
	for _, n := range curatedInstantsAndSorceries {
		got[n] = true
	}
	var diff []string
	for n := range want {
		if !got[n] {
			diff = append(diff, "missing from curatedInstantsAndSorceries: "+n)
		}
	}
	for n := range got {
		if !want[n] {
			diff = append(diff, "not an instant or sorcery in any curated deck: "+n)
		}
	}
	if len(diff) > 0 {
		sort.Strings(diff)
		t.Errorf("curatedInstantsAndSorceries is wrong:\n\t%s", strings.Join(diff, "\n\t"))
	}
}

func TestRealDumpPurposeAudit(t *testing.T) {
	idx := loadDumpForPurpose(t)

	// Every Spec under one base oracle ID is one card: its faces, a
	// split card's halves.
	type entry struct {
		name                      string
		declared, sweeps, targets bool
	}
	byBase := map[string]*entry{}
	for _, s := range effects.All() {
		base, _, _ := strings.Cut(s.OracleID, "#")
		e := byBase[base]
		if e == nil {
			e = &entry{name: s.Name}
			byBase[base] = e
		}
		d, sw := anyPurpose(s)
		e.declared = e.declared || d
		e.sweeps = e.sweeps || sw
		e.targets = e.targets || declaresTargetEntry(s)
	}

	var undeclaredWipes, staleReviews []string
	listed := map[string][]string{}
	for base, e := range byBase {
		id, err := uuid.Parse(base)
		if err != nil {
			continue // a token key, which has no printing
		}
		card, ok := idx.FindByOracleID(id)
		if !ok {
			continue
		}
		typeLine, text := printed(card)
		reads := func(re *regexp.Regexp) bool { return re.MatchString(text) }
		wipe := false
		for _, re := range sweepText {
			if reads(re) {
				wipe = true
				break
			}
		}
		_, reviewed := reviewedNotAWipe[card.Name]
		switch {
		case wipe && !e.sweeps && !reviewed:
			undeclaredWipes = append(undeclaredWipes, card.Name+" ("+base+")")
		case reviewed && (e.sweeps || !wipe):
			staleReviews = append(staleReviews, card.Name)
		}
		if !e.targets {
			switch {
			case reads(giftText):
				listed["target gift"] = append(listed["target gift"], card.Name)
			case reads(burnText):
				listed["burn at a creature or any target"] = append(listed["burn at a creature or any target"], card.Name)
			}
		}
		if e.declared {
			continue
		}
		switch {
		case reads(landToBfText):
			listed["land search"] = append(listed["land search"], card.Name)
		case reads(tutorText):
			listed["tutor"] = append(listed["tutor"], card.Name)
		case reads(drawText) && instantSorcer.MatchString(typeLine):
			listed["draw (instant or sorcery)"] = append(listed["draw (instant or sorcery)"], card.Name)
		}
	}
	if len(undeclaredWipes) > 0 {
		sort.Strings(undeclaredWipes)
		t.Errorf("%d catalog card(s) read as a board wipe and declare no Sweep (ADR 0126 owner decision 2): "+
			"declare one, or add the card to reviewedNotAWipe with the reason:\n\t%s",
			len(undeclaredWipes), strings.Join(undeclaredWipes, "\n\t"))
	}
	if len(staleReviews) > 0 {
		sort.Strings(staleReviews)
		t.Errorf("reviewedNotAWipe lists cards that now declare a Sweep, or no longer read as one:\n\t%s",
			strings.Join(staleReviews, "\n\t"))
	}
	for _, class := range []struct{ name, lacks string }{
		{"land search", "purpose"},
		{"tutor", "purpose"},
		{"draw (instant or sorcery)", "purpose"},
		{"target gift", "target entry"},
		{"burn at a creature or any target", "target entry"},
	} {
		names := listed[class.name]
		sort.Strings(names)
		t.Logf("review aid: %d catalog card(s) read as a %s and declare no %s:\n\t%s",
			len(names), class.name, class.lacks, strings.Join(names, "\n\t"))
	}
}

// readsLandsUntapped reports whether a card's text puts a searched-out
// card onto the battlefield without "tapped".
func readsLandsUntapped(text string) bool {
	for _, m := range searchToBfText.FindAllStringSubmatch(text, -1) {
		if m[1] == "" {
			return true
		}
	}
	return false
}

// declaredLands sums the Lands a spec declares in every slot.
func declaredLands(s effects.Spec) int {
	n := 0
	for _, p := range specPurposes(s) {
		n += p.Lands
	}
	return n
}

func TestRealDumpPurposeLandsUntapped(t *testing.T) {
	idx := loadDumpForPurpose(t)
	curated := map[string]bool{}
	var diff []string
	for _, d := range All() {
		for _, c := range d.Cards() {
			if c.Basic || curated[c.Name] {
				continue
			}
			curated[c.Name] = true
			spec, ok := effects.Lookup(c.OracleID)
			if !ok || declaredLands(spec) == 0 {
				continue
			}
			card, ok := idx.FindByOracleID(uuid.MustParse(c.OracleID))
			if !ok {
				t.Errorf("%s is not in the dump", c.Name)
				continue
			}
			_, text := printed(card)
			_, listed := curatedLandsUntapped[c.Name]
			switch untapped := readsLandsUntapped(text); {
			case untapped && !listed:
				diff = append(diff, c.Name+" puts its lands onto the battlefield untapped and is not on curatedLandsUntapped")
			case !untapped && listed:
				diff = append(diff, c.Name+" is on curatedLandsUntapped and its lands enter tapped")
			}
		}
	}
	if len(diff) > 0 {
		sort.Strings(diff)
		t.Errorf("curatedLandsUntapped is wrong (ADR 0136 §2):\n\t%s", strings.Join(diff, "\n\t"))
	}

	var aid []string
	for _, s := range effects.All() {
		if curated[s.Name] || declaredLands(s) == 0 || landsUntappedIn(s) > 0 {
			continue
		}
		base, _, _ := strings.Cut(s.OracleID, "#")
		id, err := uuid.Parse(base)
		if err != nil {
			continue
		}
		card, ok := idx.FindByOracleID(id)
		if !ok {
			continue
		}
		if _, text := printed(card); readsLandsUntapped(text) {
			aid = append(aid, s.Name)
		}
	}
	sort.Strings(aid)
	t.Logf("review aid: %d catalog card(s) outside the curated decks declare Lands, put them onto the battlefield untapped, and declare no LandsUntapped:\n\t%s",
		len(aid), strings.Join(aid, "\n\t"))
}
