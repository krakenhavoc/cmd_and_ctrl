package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tangle Wire — Artifact {3}:
//
//	"Fading 4 (This artifact enters with four fade counters on it. At the
//	 beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)
//	 At the beginning of each player's upkeep, that player taps an
//	 untapped artifact, creature, or land they control for each fade
//	 counter on this artifact."
//
// Fading is Fading/FadingUpkeep (fading.go, CR 702.32a). The tapping
// trigger fires at every player's upkeep, and the number is read as it
// resolves: the fade counters on the Wire then, or as it last existed on
// the battlefield if it has left (CR 608.2h). At its controller's upkeep
// the two triggers are both theirs, so they order them (CR 603.3b):
// fading first taps one fewer.
//
// The upkeep's player chooses which of their untapped artifacts,
// creatures and lands to tap, exactly as many as the count, or all of
// them when they have no more than that (then nothing is asked). The Wire
// is an artifact, so its controller may tap the Wire itself. Each pick is
// re-read as it is tapped, so one that became tapped or changed hands
// while the prompt was open is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4ea0931-43e0-4214-9e0b-213eee3a7e69",
		Name:         "Tangle Wire",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Fading(4)},
		Triggered: []game.TriggeredAbility{
			FadingUpkeep(),
			On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			}, "Tangle Wire — that player taps an untapped artifact, creature, or land they control for each fade counter on it",
				tangleWireAsk),
		},
	})
}

// tangleWireTappable is CR's "untapped artifact, creature, or land"
// the player controls.
func tangleWireTappable(c game.Card, player uuid.UUID) bool {
	return c.Controller == player && !c.Tapped && (c.IsArtifact() || c.IsCreature() || c.IsLand())
}

// tangleWireAsk counts the fade counters and has the upkeep's player
// tap that many of their permanents.
func tangleWireAsk(g *game.Game, item *game.StackItem) error {
	player := triggeringActor(item)
	if player == uuid.Nil {
		return nil
	}
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	n := info.Counters[game.CounterFade]
	if n <= 0 {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if tangleWireTappable(c, player) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) <= n {
		// Every one of them: nothing to choose.
		return tangleWireTap(ctx, player, candidates)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: tangleWireQuestion(n),
		Cards:    candidates,
		Min:      n,
		Max:      n,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  tangleWireTapChosen,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// tangleWireQuestion is the prompt's header for a count of n.
func tangleWireQuestion(n int) string {
	if n == 1 {
		return "Tangle Wire — tap an untapped artifact, creature, or land you control"
	}
	return fmt.Sprintf("Tangle Wire — tap %d untapped artifacts, creatures, or lands you control", n)
}

// tangleWireTapChosen taps the upkeep player's picks. An on-disk key:
// never renamed, never reused.
var tangleWireTapChosen = askedPick("choose-cards/tangle-wire-tap", func(ctx *Context, r game.CardsChosen) error {
	return tangleWireTap(ctx, r.Chooser, r.Picked)
})

// tangleWireTap taps each of ids that is still an untapped artifact,
// creature or land `player` controls.
func tangleWireTap(ctx *Context, player uuid.UUID, ids []uuid.UUID) error {
	for _, id := range ids {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !onBattlefield(ctx.Game, id) || !tangleWireTappable(c, player) {
			continue
		}
		if err := (TapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
