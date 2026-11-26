package game

import (
	"slices"
	"strings"

	"github.com/google/uuid"
)

// untap_type_choice.go — "each player chooses artifact, creature, or
// land during their untap step. That player can untap only permanents
// of the chosen type this step" (Storage Matrix, #6218; CR 502.3).
//
// CR 502.3: "the active player determines which permanents they control
// will untap. Then they untap them all simultaneously." untap_choice.go
// (ADR 0070) made the determination a real decision for a cap and an
// opt-out. This is the third shape: a card-type pick that narrows the
// whole CR 502.3 set before anything else is asked. It is an
// UntapTypeChoice on the card's definition (plain data: the types
// offered, and whether it binds only while its source is untapped),
// gathered at the step like the caps.
//
// The question is an ordinary keyed option pick (option_pick_keyed.go),
// asked of the active player once per live source, so two Storage
// Matrices are two choices and only a permanent of BOTH chosen types
// untaps. Everything the rest of the step needs rides the prompt as
// data: Carry is the sources still to ask about and then the untap set,
// Nums[0] is how many of the carried IDs are sources, Words is this
// source's options as lowercase types. Nothing untaps until the last
// answer; then the narrowed set goes on to the ADR 0070 determination
// (determineUntapLocked), which may ask again, and the step ends.
//
// "As long as this artifact is untapped" is read when the step begins,
// as Winter Orb's is.

// UntapTypeChoice declares "each player chooses <types> during their
// untap step; that player can untap only permanents of the chosen type
// this step" (Storage Matrix). Plain data on CardDef.UntapTypeChoices.
type UntapTypeChoice struct {
	// Types are the options, lowercase permanent types, in the order the
	// card prints them.
	Types []string
	// WhileUntapped is "as long as this <permanent> is untapped".
	WhileUntapped bool
	// Label names the clause for logs and the card panel.
	Label string
}

// CatalogUntapTypeChoices returns the untap type choices declared by
// the given ability key, or nil. Nil hook: no catalog, nothing asked.
var CatalogUntapTypeChoices func(oracleID string) []UntapTypeChoice

// untapTypeChoiceThen is the answer to one source's type pick. An
// on-disk key: never renamed, never reused. Registered from init for
// untapChoiceThen's reason (it re-enters the step machinery).
var untapTypeChoiceThen OptionPickThen

func init() {
	untapTypeChoiceThen = RegisterOptionPickThen("option/untap-type-choice", untapTypeChoiceAnswered)
}

// liveUntapTypeChoiceSourcesLocked lists the permanents whose type
// choice binds this untap step, in battlefield order, one entry per
// declared choice (a card declaring two would be asked twice).
//
// Caller must hold g.mu in write mode.
func (g *Game) liveUntapTypeChoiceSourcesLocked() []uuid.UUID {
	if g.Battlefield == nil || CatalogUntapTypeChoices == nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		key := catalogAbilityKeyOf(c)
		if key == "" {
			continue
		}
		for _, ch := range CatalogUntapTypeChoices(key) {
			if len(ch.Types) == 0 || (ch.WhileUntapped && c.Tapped) {
				continue
			}
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// untapTypeChoiceOf is the choice `source` declares, or nil.
func untapTypeChoiceOf(source *Card) *UntapTypeChoice {
	if source == nil || CatalogUntapTypeChoices == nil {
		return nil
	}
	key := catalogAbilityKeyOf(source)
	if key == "" {
		return nil
	}
	for _, ch := range CatalogUntapTypeChoices(key) {
		if len(ch.Types) > 0 {
			ch := ch
			return &ch
		}
	}
	return nil
}

// askUntapTypeChoiceLocked asks `activePlayer` the type pick of the
// first of `sources` that still declares one, carrying the rest and the
// untap set. Reports false when nothing was asked (no source left, or
// the chooser has left the game), and the caller goes on with `set`.
//
// Caller must hold g.mu in write mode.
func (g *Game) askUntapTypeChoiceLocked(activePlayer uuid.UUID, sources, set []uuid.UUID) bool {
	for len(sources) > 0 {
		src := sources[0]
		sources = sources[1:]
		ch := untapTypeChoiceOf(findBattlefieldCard(g, src))
		if ch == nil {
			continue
		}
		opts := make([]ChoiceOption, len(ch.Types))
		for i, t := range ch.Types {
			opts[i] = ChoiceOption{Label: titleWord(t)}
		}
		carry := make([]uuid.UUID, 0, len(sources)+len(set))
		carry = append(append(carry, sources...), set...)
		id := g.QueueOptionPickForEffect(OptionPickPrompt{
			Chooser:  activePlayer,
			Source:   src,
			Question: "Untap step — choose the card type you can untap this step",
			Options:  opts,
			ThenKey:  untapTypeChoiceThen,
			Carry:    carry,
			Words:    copyStrings(ch.Types),
			Nums:     []int{len(sources)},
		})
		return id != uuid.Nil
	}
	return false
}

// titleWord capitalises a lowercase type for an option label.
func titleWord(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func untapTypeChoiceAnswered(g *Game, r OptionPicked) error {
	nSources := 0
	if len(r.Nums) == 1 {
		nSources = r.Nums[0]
	}
	if nSources < 0 || nSources > len(r.Carry) {
		nSources = 0
	}
	sources := append([]uuid.UUID(nil), r.Carry[:nSources]...)
	set := append([]uuid.UUID(nil), r.Carry[nSources:]...)
	active := g.activePlayerIDLocked()
	if r.Option != nil && r.Index >= 0 && r.Index < len(r.Words) {
		chosen := r.Words[r.Index]
		set = slices.DeleteFunc(set, func(id uuid.UUID) bool {
			c := findBattlefieldCard(g, id)
			return c == nil || !hasCardType(c, chosen)
		})
	}
	if g.askUntapTypeChoiceLocked(active, sources, set) {
		return nil
	}
	if g.determineUntapLocked(active, set) {
		return nil
	}
	if g.State != StateActive || g.Turn.Step != StepUntap || g.MulligansOpen {
		return nil
	}
	g.exitUntapStepLocked()
	return nil
}

// determineUntapLocked is CR 502.3 from the untap set on: the ADR 0070
// determination when a cap or an opt-out makes it a question (paused
// true), otherwise the whole set untaps at once.
//
// Caller must hold g.mu in write mode.
func (g *Game) determineUntapLocked(activePlayer uuid.UUID, ids []uuid.UUID) (paused bool) {
	if plan := g.untapChoicePlanLocked(activePlayer, ids); plan != nil {
		g.queueUntapChoiceLocked(activePlayer, plan)
		return true
	}
	for _, id := range ids {
		g.untapPermanentByIDLocked(id)
	}
	return false
}
