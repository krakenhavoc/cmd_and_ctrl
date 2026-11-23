package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b226_helpers.go — shared shapes of the GAP batch 226 cards
// (#5948), in a file of their own so concurrent batches never collide.

// CreatureTypesAmongYourCreaturesMax5 is Valiant Changeling's reduction:
// the number of distinct creature types among the creatures its caster
// controls (CR 205.3m), at most five. Types are read through the layers,
// and a changeling (CR 702.73a) is every one of them.
func CreatureTypesAmongYourCreaturesMax5(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	q.Game.RecomputeLayersIfStaleLocked()
	seen := map[string]bool{}
	for _, c := range q.Game.BattlefieldCardsForEffect() {
		if c.Controller != q.Controller || !c.IsCreature() {
			continue
		}
		for _, t := range game.CreatureTypesOf(&c) {
			seen[t] = true
		}
	}
	return min(len(seen), 5)
}

// IfYouHaveAFullParty narrows a trigger to the intervening "if you have a
// full party" (CR 700.8c, 603.4): it triggers only then, and does nothing if
// the party is no longer full as it resolves (Squad Commander).
func IfYouHaveAFullParty(t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && FinB9HasFullParty(g, source.Controller) && (when == nil || when(ev, source, lki, g))
	}
	return InterveningIf(FinB9FullParty, t)
}

// FirstDragonSpellThisTurn passes on a Dragon spell while its caster has cast
// no Dragon spell yet this turn (Acolyte of Bahamut). The tally is bumped
// after a spell is priced (CR 601.2f), so no Dragon spell so far means this
// one is the first.
func FirstDragonSpellThisTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && q.Card.HasSubtype(q.Source.TextSubtype("Dragon")) && q.Game.CastTallyFor(q.Controller).DragonSpells == 0
	}
}
