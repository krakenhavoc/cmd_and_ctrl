package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b21_helpers.go — shared pieces of the Streets of New Capenna
// Commander spell-copy batch (S82, #5491). Append-only.

// s82SncEquipmentAttachedTo counts the Equipment attached to the permanent
// `host`, read through the layered view of each attachment.
func S82SncEquipmentAttachedTo(g *game.Game, host uuid.UUID) int {
	n := 0
	for _, a := range g.BattlefieldCardsForEffect() {
		if a.IsAttachedTo(host) && a.HasSubtype("Equipment") {
			n++
		}
	}
	return n
}

// s82SncCounterKindsAmongYourPermanents is "each kind of counter among
// permanents you control": the distinct counter names with at least one
// counter on a permanent `player` controls, sorted so a log reads the same
// way every time.
func S82SncCounterKindsAmongYourPermanents(g *game.Game, player uuid.UUID) []string {
	seen := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player {
			continue
		}
		for kind, n := range c.Counters {
			if n > 0 {
				seen[kind] = true
			}
		}
	}
	kinds := make([]string, 0, len(seen))
	for k := range seen {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	return kinds
}

// s82SncReplicateTrigger is the replicate trigger of a spell a permanent
// gave the keyword to (CR 702.56a): "When you cast this spell, if a
// replicate cost was paid for it, copy it for each time its replicate cost
// was paid." It is Hugh, First of Many's trigger (trekB39HughReplicate) for
// any spell the grant reaches: `reaches` is the grant's own predicate, read
// against the spell as it was cast, and only the permanent the engine filed
// the grant from copies it (trekB39HughIsTheGrantor).
func S82SncReplicateTrigger(label string, reaches func(g *game.Game, caster uuid.UUID, spell game.Card) bool) game.TriggeredAbility {
	return game.TriggeredAbility{
		Keyword: KeywordReplicate,
		Key:     label,
		Watches: []game.EventKind{game.EventCast},
		AppliesTo: socB32YouCastASpellThat(func(spell game.Card, source *game.Card, g *game.Game) bool {
			return reaches(g, source.Controller, spell) && trekB39HughPaid(g, spell.InstanceID) > 0 &&
				trekB39HughIsTheGrantor(g, source, source.Controller, spell)
		}),
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Controller, item.Owner = ev.Actor, ev.Actor
			item.Params.Object = game.ObjectRef{ID: ev.CardID}
			item.Params.Amount = trekB39HughPaid(g, ev.CardID)
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			if item.Params.Amount <= 0 {
				return nil
			}
			return CopySpell{
				StackID:          item.Params.Object.ID,
				Controller:       item.Controller,
				Count:            item.Params.Amount,
				ChooseNewTargets: true,
				FromLastKnown:    true,
			}.Apply(NewContext(g, item))
		},
	}
}

// s82SncFirstInstantOrSorceryThisTurn is "the first instant or sorcery spell
// you cast each turn", read as the spell is announced, before the cast is
// tallied (CR 601.2b): the card is an instant or sorcery and the caster has
// cast none yet this turn.
func S82SncFirstInstantOrSorceryThisTurn(g *game.Game, caster uuid.UUID, spell game.Card) bool {
	return (spell.IsInstant() || spell.IsSorcery()) && g.CastTallyFor(caster).InstantSorcery == 0
}

// s82SncExactlyThreeColors is "a spell that's exactly three colors".
func S82SncExactlyThreeColors(_ *game.Game, _ uuid.UUID, spell game.Card) bool {
	return ColorsOf(spell) == 3
}

// s82SncSpellWasNotCast is "a spell that wasn't cast": a copy created on the
// stack (CR 707.10), which is never cast. A prepared spell's copy is the one
// copy that is cast (CR 722.3c), so it is not one.
func S82SncSpellWasNotCast(g *game.Game, _ uuid.UUID, spell game.Card) bool {
	if spell.PrepareCopy {
		return false
	}
	item := g.StackItemForEffect(spell.InstanceID)
	return item != nil && item.IsCopy
}
