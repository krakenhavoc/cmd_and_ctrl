package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tiller Engine — Artifact Creature — Construct {2}, 1/3 (EDHREC
// rank 3309):
//
//	"Whenever a land you control enters tapped, choose one —
//	 • Untap that land.
//	 • Tap target nonland permanent an opponent controls."
//
// Amulet of Vigor with a second mode. "Enters tapped" is read off
// the land at the moment EventETB fires (b31LandYouControlEnteredTapped
// — every entry path stamps Tapped before it announces the entry),
// and each land that enters tapped is its own trigger, as printed.
//
// The modal trigger has no mode prompt of its own — a triggered
// ability cannot carry a "choose one" — so the choice is made
// through the trigger's target prompt instead (Voracious Hydra's
// posture): the trigger targets "up to one nonland permanent an
// opponent controls", and picking one is the tap mode while picking
// none is the untap mode (b31TapChosenOrUntapLandEffect). It is declared
// TWICE for the reason Hazel's Brewmaster gives: the engine drops a
// targeted trigger whose legal set is empty, which would drop the
// untap with it, so the targeted entry fires only while an opponent
// controls a nonland permanent and the untargeted untap-only entry
// fires only when none does. Exactly one applies to any entry, and
// exactly one mode ever happens. A chosen permanent that became
// illegal in response counters the trigger (CR 608.2b) — no tap AND
// no untap, which is what the printed card does when its chosen
// mode's target is gone.
//
// Sandbox simplification, declared: the mode is chosen through the
// target prompt rather than a mode picker.
func init() {
	Register(Spec{
		OracleID:     "62f6baba-da3b-45b8-a3c1-efb75763cca8",
		Name:         "Tiller Engine",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The mode is chosen through the target prompt: pick a nonland permanent an opponent controls to tap it, or pick no target to untap the land instead.",
		},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				Key:     b31TillerEngineLabel,
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b31LandYouControlEnteredTapped(ev, source, g) && b31OpponentControlsNonlandPermanent(g, source.Controller)
				},
				Targets: TargetPermanent("up to one target nonland permanent an opponent controls to tap (none: untap that land)",
					Nonland(), OpponentControls()).WithCount(0, 1),
				Effect: b31TapChosenOrUntapLandEffect,
			},
			{
				Watches: []game.EventKind{game.EventETB},
				Key:     b31TillerEngineLabel,
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b31LandYouControlEnteredTapped(ev, source, g) && !b31OpponentControlsNonlandPermanent(g, source.Controller)
				},
				Effect: b31TapChosenOrUntapLandEffect,
			},
		},
	})
}
