package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b035_helpers.go — shared pieces of S82 older-sets batch 35
// (Modern Horizons 2: removal).

// OldB035CouldProduceColorless is "could produce {C}" for a land
// (CR 106.7): the mana its abilities could make right now includes colorless.
// Break the Ice and Obsidian Charmaw read it. The predicates run under the
// game lock, which ProducibleManaLocked asks for.
func OldB035CouldProduceColorless(g *game.Game, c game.Card) bool {
	if g == nil || !c.IsLand() {
		return false
	}
	for _, m := range g.ProducibleManaLocked(c) {
		if m == "C" {
			return true
		}
	}
	return false
}

// OldB035SnowOrCouldProduceColorless is Break the Ice's clause "land that is
// snow or could produce {C}".
func OldB035SnowOrCouldProduceColorless() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsLand() && (c.HasSupertype("snow") || OldB035CouldProduceColorless(g, c))
	}
}

// OldB035OpponentLandsThatCouldProduceColorless counts the lands the spell's
// caster's opponents control that could produce {C}: Obsidian Charmaw's
// reduction.
func OldB035OpponentLandsThatCouldProduceColorless(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	n := 0
	for _, c := range q.Game.Battlefield.Cards {
		if c.Controller == q.Controller || !c.IsLand() {
			continue
		}
		if OldB035CouldProduceColorless(q.Game, c) {
			n++
		}
	}
	return n
}

// OldB035BlackOrRed is "a black or red <object>", read through the source's
// text (CR 612.2): the two colour words are words of Sanctifier en-Vec's text.
func OldB035BlackOrRed(source *game.Card, c game.Card) bool {
	return c.HasColorWord(source.TextColor("B")) || c.HasColorWord(source.TextColor("R"))
}

// OldB035SanctifierReplacement is "If a black or red permanent, spell, or card
// not on the battlefield would be put into a graveyard, exile it instead": the
// graveyard-becomes-exile replacement for every graveyard, narrowed to black or
// red objects (a token permanent included: a permanent is not only a card).
func OldB035SanctifierReplacement() game.ReplacementEffect {
	r := GraveyardBecomesExile{Label: "Sanctifier en-Vec: exile black and red cards instead of a graveyard"}.Build()
	applies := r.AppliesTo
	r.AppliesTo = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
		if !applies(ev, g, src) {
			return false
		}
		moving, ok := g.LookupCardForEffect(ev.CardID)
		return ok && src != nil && OldB035BlackOrRed(src, moving)
	}
	return r
}

// OldB035SanctifierExile is "When this creature enters, exile all cards that
// are black or red from all graveyards." The IDs are taken before the first
// exile, since each one removes a card from the pile.
func OldB035SanctifierExile(g *game.Game, item *game.StackItem) error {
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if OldB035BlackOrRed(&src, c) {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	ctx := NewContext(g, item)
	for _, id := range ids {
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// OldB035ForecastTapLabel words Piercing Rays' forecast.
const OldB035ForecastTapLabel = "Forecast — {2}{W}, Reveal this card from your hand: Tap target untapped creature."
