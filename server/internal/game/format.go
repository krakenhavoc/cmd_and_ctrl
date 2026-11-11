package game

import (
	"errors"
	"fmt"
)

// format.go is the table's format (ADR 0222 §1): which kind of Magic a
// table plays. A format is data, a FormatSpec row, so the next
// constructed format is a new row and not new code.
//
// Commander is the default and what every table, saved setup and
// snapshot written before the field existed reads as: the stored zero
// value is Commander (TableSettings.EffectiveFormat), which also keeps
// every Commander snapshot byte-identical.

// Format names a format on the wire, in a snapshot and in a saved deck.
type Format string

const (
	// FormatCommander is the default format (CR 903).
	FormatCommander Format = "commander"
	// FormatStandard is constructed Standard, one against one
	// (ADR 0222 owner decision 2).
	FormatStandard Format = "standard"
	// FormatLimited is limited play with a drafted deck, one against
	// one (ADR 0222 §2 and §6, #5440). A table of it is opened by a
	// draft for one of its matches, never made by hand (DraftOnly).
	//
	//	CR 100.2b: "In limited play (a way of playing in which each
	//	player gets the same quantity of unopened Magic product such as
	//	booster packs and creates their own deck using only this product
	//	and basic land cards), each deck has a minimum deck size of 40
	//	cards. A limited deck may contain as many duplicates of a card
	//	as are included with the product."
	//
	//	CR 100.4b: "In limited play involving individual players, all
	//	cards in a player's card pool not included in their deck are in
	//	that player's sideboard."
	FormatLimited Format = "limited"
)

// DeckRules is a format's deck construction rules, as numbers. The deck
// checker (internal/deck) reads them; nothing in the engine does.
type DeckRules struct {
	// MinCards is the smallest legal deck.
	MinCards int
	// ExactCards says the deck is exactly MinCards cards (Commander's
	// 100, CR 903.5a), not at least that many.
	ExactCards bool
	// MaxCopies is the most cards of one English name a deck may hold,
	// basic lands aside: 1 in Commander (CR 903.5b), 4 in constructed
	// (CR 100.2a).
	MaxCopies int
	// SideboardSize is the most cards a sideboard may hold (CR 100.4a),
	// 0 for a format that uses none (CR 903.5e).
	SideboardSize int
	// AnySideboard says the sideboard has no size limit: in limited
	// play it is whatever of the card pool the deck leaves out (CR
	// 100.4b), so SideboardSize is not read.
	AnySideboard bool
	// FromPool says a deck is built from the player's own card pool
	// and basic lands (CR 100.2b), not from a format's card pool: the
	// deck check is deck.ValidateLimited, which takes that pool, and
	// Legality is not read.
	FromPool bool
	// Legality is the key of Scryfall's legalities map the format's
	// card pool is read from.
	Legality string
}

// FormatSpec is one format: what a table of it starts with.
type FormatSpec struct {
	Format Format
	// Name is the format's display name.
	Name string
	// StartingLife is each seat's life total at Start unless the host
	// sets another (CR 103.4; 40 in Commander, CR 903.7).
	StartingLife int
	// MaxSeats is the most seats a table of the format has. Never
	// above MaxPlayers, which stays the global bound.
	MaxSeats int
	// CommanderExpected says a deck of the format has a commander.
	CommanderExpected bool
	// Matches says a table of the format plays a match (ADR 0222 §3,
	// match.go): best-of-one unless the table asks for best-of-three.
	// False for Commander, whose tables have no match at all.
	Matches bool
	// DraftOnly says a table of the format is opened by a draft for one
	// of its matches (ADR 0222 §6) and cannot be made or chosen by
	// hand: nobody has a limited deck without a card pool to check it
	// against.
	DraftOnly bool
	// Deck is the format's deck construction rules.
	Deck DeckRules
}

// formats is every format, in the order a picker lists them. Commander
// first: it is the primary format (ADR 0222 owner decision 1).
var formats = []FormatSpec{
	{
		Format:            FormatCommander,
		Name:              "Commander",
		StartingLife:      StartingLife,
		MaxSeats:          MaxPlayers,
		CommanderExpected: true,
		Deck:              DeckRules{MinCards: 100, ExactCards: true, MaxCopies: 1, Legality: "commander"},
	},
	{
		Format:       FormatStandard,
		Name:         "Standard",
		StartingLife: 20,
		MaxSeats:     2,
		Matches:      true,
		Deck:         DeckRules{MinCards: 60, MaxCopies: 4, SideboardSize: 15, Legality: "standard"},
	},
	{
		// 20 life (CR 103.4), two seats (ADR 0222 owner decision 5), no
		// commander. No copy limit beyond the pool's own (CR 100.2b),
		// which MaxCopies 0 leaves to the pool check.
		Format:       FormatLimited,
		Name:         "Limited",
		StartingLife: 20,
		MaxSeats:     2,
		Matches:      true,
		DraftOnly:    true,
		Deck:         DeckRules{MinCards: 40, AnySideboard: true, FromPool: true},
	},
}

// Formats lists every format in picker order. The slice is a copy.
func Formats() []FormatSpec {
	return append([]FormatSpec(nil), formats...)
}

// SideboardCap is the most sideboard cards a seat of the format may
// bring to the table: the engine's cap of MaxSideboardSize (ADR 0191
// §5), or no cap at all in a format whose sideboard is the rest of a
// card pool (CR 100.4b). ok is false for no cap.
func (s FormatSpec) SideboardCap() (n int, ok bool) {
	if s.Deck.AnySideboard {
		return 0, false
	}
	return MaxSideboardSize, true
}

// LookupFormat is the spec of f. The empty Format is Commander, which
// is how a table, a snapshot or a saved deck from before formats
// reads. False for a name no row has.
func LookupFormat(f Format) (FormatSpec, bool) {
	if f == "" {
		f = FormatCommander
	}
	for _, s := range formats {
		if s.Format == f {
			return s, true
		}
	}
	return FormatSpec{}, false
}

// Spec is the spec of f, and Commander's for a name no row has: a
// snapshot naming a format this build does not know still restores and
// plays, as the primary format.
func (f Format) Spec() FormatSpec {
	if s, ok := LookupFormat(f); ok {
		return s
	}
	return formats[0]
}

// SettingFormat is the format's settings key (see the other Setting
// constants in settings.go).
const SettingFormat = "format"

var (
	// ErrFormatLocked is returned by UpdateSettings when a patch
	// changes the format after Start: a format is chosen when the table
	// is made and fixed once the game starts (ADR 0222 §1). Nothing in
	// the patch is applied.
	ErrFormatLocked = errors.New("game: the format cannot change once the game has started")
	// ErrFormatSeats is returned by UpdateSettings when a patch changes
	// the format to one with fewer seats than are already taken.
	// Nothing in the patch is applied.
	ErrFormatSeats = errors.New("game: more seats are taken than that format has")
)

// EffectiveFormat is the table's format: Commander when the stored
// field is empty or names a format this build does not know.
func (s TableSettings) EffectiveFormat() Format {
	return s.Format.Spec().Format
}

// FormatSpec is the spec of the table's format.
func (s TableSettings) FormatSpec() FormatSpec {
	return s.Format.Spec()
}

// storedFormat is how TableSettings keeps f: Commander as the zero
// value, so a Commander table's snapshot has no format field at all.
func storedFormat(f Format) Format {
	if f == FormatCommander {
		return ""
	}
	return f
}

// validateFormat range-checks a patch's format.
func validateFormat(f Format) error {
	if _, ok := LookupFormat(f); !ok {
		return fmt.Errorf("%w: %s %q", ErrInvalidSetting, SettingFormat, f)
	}
	return nil
}

// SeatCap is how many seats the table has: its format's (ADR 0222 §1).
func (g *Game) SeatCap() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Settings.FormatSpec().MaxSeats
}

// HasCommander reports whether any seat brought a commander: the
// client hides the commander UI when none did (ADR 0222 slice 1). It
// reads the cards, not the format, so a Commander table is unchanged
// and a table of any format with a commander still shows it.
func (g *Game) HasCommander() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.hasCommanderLocked()
}

// HasCommanderLocked is HasCommander for a caller that already holds
// g.mu, as the view builder does inside ReadSnapshot.
func (g *Game) HasCommanderLocked() bool { return g.hasCommanderLocked() }

// hasCommanderLocked is HasCommander. In the lobby the commanders are
// the command-zone cards the deck installed; once the game is under
// way a commander may be anywhere, and keeps its designation through
// every zone change (CR 903.3), so every zone is walked. A token is
// never a commander (CommandersOf says why).
//
// Caller must hold g.mu.
func (g *Game) hasCommanderLocked() bool {
	has := func(z *Zone) bool {
		if z == nil {
			return false
		}
		for i := range z.Cards {
			if c := &z.Cards[i]; c.IsCommander && !c.IsToken() {
				return true
			}
		}
		return false
	}
	for _, p := range g.Seats {
		if p != nil && has(p.Command) {
			return true
		}
	}
	if g.State == StateLobby {
		return false
	}
	// A phased-out commander is still a commander; the indexed walk
	// leaves that zone out (CR 702.26b).
	if has(g.PhasedOut) {
		return true
	}
	found := false
	g.eachIndexedZoneLocked(func(_ cardSlot, _ int, z *Zone) bool {
		found = has(z)
		return found
	})
	return found
}

// OpenLimited makes an empty table in its lobby a limited table for one
// match of a draft (ADR 0222 §6): the format, its starting life and the
// match length. It is the only way a table gets a DraftOnly format;
// UpdateSettings refuses one. bestOf is 1 or 3.
func (g *Game) OpenLimited(bestOf int) error {
	if err := validateBestOf(bestOf); err != nil {
		return err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateLobby || len(g.Seats) > 0 {
		return ErrGameNotInLobby
	}
	spec := FormatLimited.Spec()
	g.Settings.Format = storedFormat(spec.Format)
	g.Settings.StartingLife = spec.StartingLife
	g.Settings.BestOf = storedBestOf(bestOf)
	return nil
}
