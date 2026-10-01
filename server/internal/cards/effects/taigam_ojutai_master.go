package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Taigam, Ojutai Master — Legendary Creature — Human Monk {2}{W}{U}, 3/4:
//
//	"Instant, sorcery, and Dragon spells you control can't be countered.
//	 Whenever you cast an instant or sorcery spell from your hand, if
//	 Taigam attacked this turn, that spell gains rebound. (Exile the
//	 spell as it resolves. At the beginning of your next upkeep, you may
//	 cast that card from exile without paying its mana cost.)"
//
// Attack with Taigam, then every instant and sorcery you cast from hand
// for the rest of the turn comes back next upkeep, and none of them can
// be countered.
//
//   - THE SHIELD is ADR 0106 §4's battlefield static, read live at the
//     counter gate (CR 613.11). "Dragon spells" are spells with the
//     Dragon subtype.
//   - THE TRIGGER has an intervening "if" (CR 603.4): it triggers only
//     if Taigam attacked this turn, and checks again as it resolves. At
//     resolution "Taigam" is the object the trigger came from, read from
//     its last known information if it has left (CR 608.2h), so a
//     Taigam that attacked and then died still gives the spell rebound.
//   - "THAT SPELL GAINS REBOUND" is a layer-6 effect on the spell
//     (CR 613.1f, ADR 0107 §3), pinned to it on the stack. The spell
//     resolves with rebound, so a spell cast from hand is exiled and
//     can be cast free at your next upkeep. The trigger resolves before
//     the spell, and a spell countered or resolved first gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4cf3710-8600-4f95-abb2-faefdc25693d",
		Name:         "Taigam, Ojutai Master",
		Completeness: CompletenessFull,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			SpellsYouControlCantBeCountered("Instant, sorcery, and Dragon spells you control can't be countered.",
				Or(Instant(), Sorcery(), Subtype("Dragon"))),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, taigamCastFromHandAfterAttacking,
				"Taigam, Ojutai Master — that spell gains rebound", taigamGrantRebound),
		},
	})
}

// taigamCastFromHandAfterAttacking is the trigger condition and its
// intervening "if": you cast an instant or sorcery from your hand, and
// this Taigam attacked this turn.
func taigamCastFromHandAfterAttacking(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return YouCastFromYourHand(Or(Instant(), Sorcery()))(ev, source, lki, g) &&
		g.AttackedThisTurn(source.InstanceID)
}

// taigamGrantRebound rechecks the "if" against the object the trigger
// came from (CR 603.4, 608.2h), then gives the spell rebound.
func taigamGrantRebound(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok || !g.ObjectAttackedThisTurn(ref) {
		return nil
	}
	return ThatSpellGains{Keywords: []string{game.KeywordRebound}}.Apply(ctx)
}
