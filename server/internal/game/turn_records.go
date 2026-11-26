package game

import (
	"strings"

	"github.com/google/uuid"
)

// turn_records.go — #6217: four things a card asks about this turn that
// the turn tally (turn_tally.go) did not keep, recorded as they happen and
// reset with the rest of the tally when the turn begins.
//
//	Grim Reminder    "Each opponent who cast a spell this turn with the
//	                 same name as that card loses 6 life."
//	                 → TurnTally.SpellNames
//	Mana Maze        "Players can't cast spells that share a color with
//	                 the spell most recently cast this turn."
//	                 → TurnTally.LastSpellCast
//	Desolation       "each player who tapped a land for mana this turn
//	                 sacrifices a land of their choice"
//	                 → PlayerTurnTally.TappedLandForMana
//	Vodalian War     "destroy all Merfolk tapped this turn to pay for its
//	Machine          abilities" → TurnTally.TappedToPay
//
// The fifth clause of the issue, Jandor's Ring's "Discard the last card
// you drew this turn", reads Game.DrawnThisTurn, which already kept every
// draw in order (DiscardCost.LastDrawn).

// CastNameRecord is one name of a spell cast this turn, with its caster.
type CastNameRecord struct {
	Player uuid.UUID `json:"player"`
	Name   string    `json:"name"`
}

// CastColorRecord is the colours of one spell cast this turn, with its caster
// and the spell's instance ID, which a permanent that spell becomes keeps.
type CastColorRecord struct {
	Player uuid.UUID `json:"player"`
	Spell  uuid.UUID `json:"spell"`
	// Colors are the colour letters the spell had as it was cast (CR 601.2i),
	// in WUBRG order ("UR"); empty for a colourless spell.
	Colors string `json:"colors,omitempty"`
}

// LastSpellCastRecord is the spell most recently cast this turn.
type LastSpellCastRecord struct {
	// Colors are its colours as it was cast (CR 105.2, read off the spell
	// on the stack). Empty for a colourless spell, which shares a colour
	// with nothing.
	Colors []string `json:"colors,omitempty"`
}

// TappedToPayRecord is one permanent tapped to pay a cost of an activated
// ability of Source this turn.
type TappedToPayRecord struct {
	// Source is the ability's source, as the object it was (CR 400.7): a
	// War Machine that left and came back is a new object whose abilities
	// were never paid for.
	Source ObjectRef `json:"source"`
	// Tapped is the permanent that was tapped, as the object it was.
	Tapped ObjectRef `json:"tapped"`
}

// noteSpellCastLocked records a spell just cast (CR 601.2i): its name for
// its caster, and its colours as the turn's most recently cast spell.
// Caller must hold g.mu.
func (g *Game) noteSpellCastLocked(playerID, cardID uuid.UUID) {
	if g.Stack == nil {
		return
	}
	for i := range g.Stack.Cards {
		sc := &g.Stack.Cards[i]
		if sc.InstanceID != cardID {
			continue
		}
		ch := SourceCharacteristics(sc)
		// CR 201.2a: a split spell has two names, and either is "the
		// same name". The name the spell has on the stack, as a
		// copy effect or a face left it.
		named := *sc
		named.Name = ch.Name
		names := NamesOf(named)
		for _, n := range names {
			g.TurnTally.SpellNames = append(g.TurnTally.SpellNames, CastNameRecord{Player: playerID, Name: n})
		}
		g.TurnTally.SpellColors = append(g.TurnTally.SpellColors,
			CastColorRecord{Player: playerID, Spell: cardID, Colors: unionColorLetters("", ch.Colors)})
		g.TurnTally.LastSpellCast = &LastSpellCastRecord{Colors: append([]string(nil), ch.Colors...)}
		return
	}
}

// noteManaAbilityTappedLocked records CR 106.12's "tapped a land for mana"
// for the activator: a mana ability with {T} in its cost, of a land.
// Caller must hold g.mu.
func (g *Game) noteManaAbilityTappedLocked(playerID, cardID uuid.UUID, tapSymbol bool) {
	if !tapSymbol {
		return
	}
	c := g.findCardByIDLocked(cardID)
	if c == nil || !c.IsLand() {
		return
	}
	g.bumpPlayerTally(playerID, func(t *PlayerTurnTally) { t.TappedLandForMana = true })
}

// noteTappedToPayLocked records the permanents an activated ability of
// `source` tapped for its cost. Caller must hold g.mu.
func (g *Game) noteTappedToPayLocked(source uuid.UUID, taps []PaidTap) {
	if len(taps) == 0 {
		return
	}
	src := ObjectRef{ID: source, Epoch: g.objectEpochLocked(source)}
	for _, t := range taps {
		g.TurnTally.TappedToPay = append(g.TurnTally.TappedToPay, TappedToPayRecord{Source: src, Tapped: ObjectRef{ID: t.ID, Epoch: t.Epoch}})
	}
}

// isLastDrawnThisTurnLocked reports whether `cardID` is the last card
// `playerID` drew this turn. Caller must hold g.mu.
func (g *Game) isLastDrawnThisTurnLocked(playerID, cardID uuid.UUID) bool {
	drawn := g.DrawnThisTurn[playerID]
	return len(drawn) > 0 && drawn[len(drawn)-1] == cardID
}

// CastSpellNamedThisTurnForEffect reports whether `playerID` cast a spell
// this turn with the same name as `card` (CR 201.2a: at least one name in
// common; a card with no name shares a name with nothing). Grim
// Reminder's reader. Caller must hold g.mu.
func (g *Game) CastSpellNamedThisTurnForEffect(playerID uuid.UUID, card Card) bool {
	for _, r := range g.TurnTally.SpellNames {
		if r.Player == playerID && r.Name != "" && HasName(card, r.Name) {
			return true
		}
	}
	return false
}

// CastAnotherSpellOfColorThisTurnForEffect reports whether `playerID` cast a
// spell of the colour this turn other than the spell or permanent `except`
// ("if you've cast another blue spell this turn", CR 700.2: another means
// not the object itself). A spell cast and then countered still counts, as
// it was cast; a copy was never cast (CR 707.10). Caller must hold g.mu.
func (g *Game) CastAnotherSpellOfColorThisTurnForEffect(playerID, except uuid.UUID, color string) bool {
	for _, r := range g.TurnTally.SpellColors {
		if r.Player == playerID && r.Spell != except && strings.Contains(r.Colors, color) {
			return true
		}
	}
	return false
}

// LastSpellCastColorsThisTurnForEffect is the colours of the spell most
// recently cast this turn, and whether any spell has been cast this turn.
// Mana Maze's reader. Caller must hold g.mu.
func (g *Game) LastSpellCastColorsThisTurnForEffect() ([]string, bool) {
	if g.TurnTally.LastSpellCast == nil {
		return nil, false
	}
	return append([]string(nil), g.TurnTally.LastSpellCast.Colors...), true
}

// TappedLandForManaThisTurnForEffect reports whether `playerID` tapped a
// land for mana this turn (CR 106.12). Desolation's reader.
func (g *Game) TappedLandForManaThisTurnForEffect(playerID uuid.UUID) bool {
	return g.TurnTally.Players[playerID].TappedLandForMana
}

// TappedToPaySourceThisTurnForEffect is the permanents tapped this turn to
// pay costs of activated abilities of `source` (the object, CR 400.7), in
// the order they were tapped, each once. Vodalian War Machine's reader:
// its dies trigger names its source as it last existed.
func (g *Game) TappedToPaySourceThisTurnForEffect(source ObjectRef) []ObjectRef {
	var out []ObjectRef
	seen := map[ObjectRef]bool{}
	for _, r := range g.TurnTally.TappedToPay {
		if r.Source != source || seen[r.Tapped] {
			continue
		}
		seen[r.Tapped] = true
		out = append(out, r.Tapped)
	}
	return out
}
