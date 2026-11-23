package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sungold Sentinel — Creature — Human Soldier {1}{W}, 3/2 (Midnight Hunt
// Commander, S82 MID batch 13):
//
//	"Whenever this creature enters or attacks, exile up to one target card
//	 from a graveyard.
//	 Coven — {1}{W}: Choose a color. This creature gains hexproof from that
//	 color until end of turn and can't be blocked by creatures of that color
//	 this turn. Activate only if you control three or more creatures with
//	 different powers."
//
// One trigger watching two events (Blossoming Tortoise's shape) with Ambush
// Wolf's "up to one target card from a graveyard". The coven ability is
// gated at activation (CR 602.5b) by three distinct powers among the
// creatures its controller controls, read through the layers
// (trekB48Coven); the colour is chosen as it resolves and the grant is
// Skrelv's pair, hexproof from the colour plus evasion against it, on this
// creature until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "410f5067-3380-4afb-9881-4b3bf0268493",
		Name:         "Sungold Sentinel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OnAny([]game.EventKind{game.EventETB, game.EventAttack},
					func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return effects.B21SelfEnteredOrAttacked(ev, source)
					},
					"Sungold Sentinel — exile up to one target card from a graveyard", effects.B27ExileChosenTarget),
				effects.TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "Coven — {1}{W}: Choose a color. This creature gains hexproof from that color until end of turn and can't be blocked by creatures of that color this turn. Activate only if you control three or more creatures with different powers.",
			Purpose:   game.Purpose{Answers: game.AnswerProtect | game.AnswerCombatGrant},
			Cost:      effects.ManaCost("{1}{W}"),
			Condition: effects.TrekB48Coven,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				effects.ChooseColorTail(game.ColorForProtection, g, item.Controller, item.SourceCardID,
					"Sungold Sentinel — choose a color", ctx.ItemTail(sungoldSentinelColorChosen))
				return nil
			},
		}},
	})
}

// sungoldSentinelColorChosen gives the source hexproof from the colour (the
// tail's Word) and evasion against it until end of turn. An on-disk key.
var sungoldSentinelColorChosen = effects.CardTail("then/sungold-sentinel-color-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" {
		return nil
	}
	return effects.ScopedEffectFor{
		Target: ctx.Source(),
		Mods: []game.Mod{
			game.AddKeywordsMod(game.HexproofFromColor(color)),
			game.CantBeBlockedByColorMod(color),
		},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Sungold Sentinel — hexproof and evasion from " + game.ColorName(color),
	}.Apply(ctx)
})
