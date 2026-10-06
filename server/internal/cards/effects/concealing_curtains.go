package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// concealingCurtainsOracle is shared by both faces: the front registers
// under it, the back under concealingCurtainsOracle + "#1".
const concealingCurtainsOracle = "a3a56b02-7290-4352-a862-f2de608f2a86"

// Concealing Curtains // Revealing Eye — a transforming creature:
//
//	Concealing Curtains {B}, Creature — Wall, 0/4
//	  "Defender
//	   {2}{B}: Transform this creature. Activate only as a sorcery."
//
//	Revealing Eye, Creature — Eye Horror, 3/4
//	  "Menace
//	   When this creature transforms into Revealing Eye, target opponent
//	   reveals their hand. You may choose a nonland card from it. If you
//	   do, that player discards that card, then draws a card."
//
// The activation turns the permanent over in place (Transform, ADR
// 0079). The back face's trigger watches that transform: it triggers
// only when THIS permanent transforms and is Revealing Eye immediately
// afterwards (CR 701.27e), so a permanent that enters with its back
// face up has not transformed and does not trigger.
//
// The trigger is the optional revealed-hand pick (#2115, ADR 0116's
// 2026-10-05 amendment): the whole table sees the hand (CR 701.20a),
// and you may choose a nonland card or nothing. "If you do" is a keyed
// continuation run after the discard: the player draws a card only when
// a card was chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        concealingCurtainsOracle,
		Name:            "Concealing Curtains",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:        "{2}{B}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{2}{B}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        concealingCurtainsOracle + "#1",
		Name:            "Revealing Eye",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventTransform},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				// CR 701.27e: this permanent transformed, and its back
				// face — Revealing Eye — is now up.
				return ev.CardID == source.InstanceID && ev.Amount == 1
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Key:     "Revealing Eye — target opponent reveals their hand",
			Effect:  revealingEyeReveal,
		}},
	})
}

// revealingEyeReveal is the trigger's effect.
func revealingEyeReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return ChooseFromRevealedHand{
		Player:   TargetedPlayer(ctx),
		Filter:   Nonland(),
		Label:    "nonland card",
		Optional: true,
		Then:     revealingEyeIfYouDo,
	}.Apply(ctx)
}

// revealingEyeIfYouDo is "If you do, that player discards that card,
// then draws a card": the discard is the pick's own, and the draw
// follows it only when a card was chosen.
var revealingEyeIfYouDo = RevealedPickThen("revealed-pick/revealing-eye-draw",
	func(ctx *Context, pick game.RevealedPick) error {
		if len(pick.Chosen) == 0 {
			return nil
		}
		return ctx.Game.DrawNForEffect(pick.FromPlayer, 1)
	})
