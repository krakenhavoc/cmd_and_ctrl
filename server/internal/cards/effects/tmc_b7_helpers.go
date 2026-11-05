package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b7_helpers.go — shapes TMNT Commander batch 7 shares (#5011).

// tmcB7APlayerCasts is "Whenever a player casts a <spell> spell": any
// player's cast, the source's controller's included. The spell's type
// is a word of the source's text (CR 612.2), read through it as YouCast
// and AnOpponentCast read theirs.
func tmcB7APlayerCasts(spell CardPredicate) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventCast || ev.Actor == uuid.Nil {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && spell(g, ev.Actor, game.ThroughTextOf(source, c))
	}
}

// tmcB7ArtifactInstantOrSorcery is "artifact, instant, or sorcery".
func tmcB7ArtifactInstantOrSorcery() CardPredicate {
	return Or(Artifact(), Instant(), Sorcery())
}

// tmcB7EachModeADifferentPlayer is "Choose one or both. Each mode must
// target a different player." (the TMNT Commander team-up cycle): one
// or both bullets, each bullet's "target player" clause Distinct, so
// with both chosen the second bullet's pick may not repeat the first's
// (CR 601.2c lets one player fill two "target" words unless the card
// says otherwise; this card says otherwise). Both clauses are marked,
// so the rule holds whichever order the chosen bullets are announced
// in.
func tmcB7EachModeADifferentPlayer(options ...game.ModeOption) *game.ModeSpec {
	for i := range options {
		Distinct(options[i].Targets)
	}
	return ChooseN("Choose one or both", 1, 2, options...)
}

// tmcB7TargetPlayer is a modal bullet's "target player" clause.
func tmcB7TargetPlayer() *game.TargetSpec { return TargetPlayer("target player") }

// tmcB7ModePlayer is the still-legal player a bullet targets (CR
// 608.2b), or false when it has none.
func tmcB7ModePlayer(ctx *Context, occ int) (uuid.UUID, bool) {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return uuid.Nil, false
	}
	return t.ID, true
}

// untapEachLegalTarget is "untap up to N target <permanents>": each target
// still legal as the ability resolves is untapped (CR 608.2b). Teferi,
// Temporal Archmage's −1 and Triceraton Regenta's attack trigger.
func untapEachLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetCards(item, g) {
		if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
