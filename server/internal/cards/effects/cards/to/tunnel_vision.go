package cardsto

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tunnel Vision — Sorcery {5}{U} (Ravnica: City of Guilds, S82 OLD batch 581,
// #5600):
//
//	"Choose a card name. Target player reveals cards from the top of their
//	 library until a card with that name is revealed. If it is, that player puts
//	 the rest of the revealed cards into their graveyard and puts the card with
//	 the chosen name on top of their library. Otherwise, the player shuffles."
//
// The name is chosen as the spell resolves, by its controller (free text, as
// Slaughter Games reads one), and a target gone by resolution makes the spell do
// nothing and no name is asked (CR 608.2b). The run is measured read-only and
// announced as one reveal (Mind Funeral's shape). The named card is then the top
// card of the library already, so only the cards above it move: that many are
// put into the graveyard and it stays on top, and nothing is shuffled. With no
// such card the whole library was revealed and the player shuffles it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e134040-1fe3-410c-aac5-94ae1bd17637",
		Name:         "Tunnel Vision",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim := effects.TargetedPlayerOf(ctx)
			if victim == uuid.Nil {
				return nil
			}
			ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
				"Tunnel Vision — choose a card name", effects.WithCarry(ctx.ItemTail(tunnelVisionNamed), victim))
			return nil
		},
	})
}

// tunnelVisionNamed is the reveal once the name (the tail's Word) is known. An
// on-disk key: never renamed, never reused.
var tunnelVisionNamed = effects.CardTail("card-name/s82-old-b581-tunnel-vision", func(ctx *effects.Context, o game.TailOutcome) error {
	name := strings.TrimSpace(o.Tail.Word)
	victim := effects.TailCarry0(o)
	p := ctx.PlayerByID(victim)
	if victim == uuid.Nil || p == nil || p.Library == nil {
		return nil
	}
	var run []uuid.UUID
	found := false
	for i := len(p.Library.Cards) - 1; i >= 0; i-- {
		c := p.Library.Cards[i]
		run = append(run, c.InstanceID)
		if name != "" && game.CardNameMatches(c, name) {
			found = true
			break
		}
	}
	if err := (effects.RevealCards{Player: victim, Cards: run, Reason: "reveal from the top until a card named " + name}).Apply(ctx); err != nil {
		return err
	}
	if !found {
		return effects.ShuffleLibrary{Player: victim}.Apply(ctx)
	}
	if len(run) > 1 {
		return effects.MillToZone{Player: victim, N: len(run) - 1}.Apply(ctx)
	}
	return nil
})
