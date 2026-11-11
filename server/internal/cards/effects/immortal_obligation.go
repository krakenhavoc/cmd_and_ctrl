package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Immortal Obligation — Instant {1}{W}:
//
//	"Return target creature card from an opponent's graveyard to the
//	 battlefield under their control with a duty counter on it. For as long
//	 as that creature has a duty counter on it, it is goaded, can't attack
//	 you or a permanent you control, and can't block creatures you control."
//
// The creature returns under the control of the graveyard's owner, the
// opponent (not the caster), and ENTERS with the duty counter (the keyed
// return with counters, so an entry replacement sees it). Once it has
// entered, three rules are pinned to it for as long as the counter is on it
// (CR 611.2b: it ends the moment the last duty counter leaves, for good):
//
//   - goaded by the caster (game.GoadMod): it attacks each combat if able and
//     attacks a player other than the caster if able (CR 701.15b);
//   - can't attack the caster or any permanent they control: the player, the
//     planeswalkers they control and the battles they control
//     (game.CantAttackPlayerOrPermanentsMod, CR 508.1c);
//   - can't block creatures the caster controls (game.CantBlockPlayersCreaturesMod,
//     CR 509.1b), read live against each attacker's controller.
//
// The target is re-judged at resolution (CR 608.2b): a card that left the
// graveyard in response fizzles the spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b34080fe-69d1-4cc9-9feb-c37762f26a23",
		Name:         "Immortal Obligation",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card in an opponent's graveyard", Creature(), Not(YouOwn())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ids := legalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(ids[0])
			if !ok {
				return nil
			}
			return ctx.Game.ReturnFromGraveyardWithCountersKeyedForEffect(ids[0], card.Owner, false,
				map[string]int{immortalObligationCounter: 1}, ctx.ItemTail(immortalObligationEntered))
		},
	})
}

// immortalObligationCounter is the duty counter's kind.
const immortalObligationCounter = "duty"

// immortalObligationEntered is the entry's tail as a keyed tail: it pins the
// three rules to the returned creature for as long as it has the counter.
// An on-disk key: never renamed, never reused.
var immortalObligationEntered = EnteredTail("immortal-obligation/duty-rules", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	you := ctx.Controller()
	d, ok := DurationWhileItHasCounter(ctx, entered, immortalObligationCounter)
	if !ok {
		return nil
	}
	if err := (ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.GoadMod(you), game.CantAttackPlayerOrPermanentsMod(you)},
		Duration: d,
		Label:    "Immortal Obligation — goaded and can't attack its caster or their permanents while it has a duty counter",
	}).Apply(ctx); err != nil {
		return err
	}
	d, ok = DurationWhileItHasCounter(ctx, entered, immortalObligationCounter)
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.CantBlockPlayersCreaturesMod(you, "it can't block creatures its goader controls")},
		Duration: d,
		Label:    "Immortal Obligation — can't block creatures its caster controls while it has a duty counter",
	}.Apply(ctx)
})
