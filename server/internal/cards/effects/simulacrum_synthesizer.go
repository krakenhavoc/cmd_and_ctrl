package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Simulacrum Synthesizer — Artifact {2}{U} (EDHREC rank 1718):
//
//	"When this artifact enters, scry 2.
//	 Whenever another artifact you control with mana value 3 or
//	 greater enters, create a 0/0 colorless Construct artifact
//	 creature token with "This token gets +1/+1 for each artifact
//	 you control.""
//
// The artifact deck's Construct factory: every real rock or
// equipment after it is a Karnstruct. The entry scry is Temple's;
// the Construct trigger reads the entering artifact's mana value off
// the permanent (b15AnotherBigArtifactYouControlEntered — a token
// with no cost is zero, so Treasures make nothing).
//
// Sandbox simplification, and the reason the card carries a
// caveat: the Construct's printed "+1/+1 for each artifact you
// control" is an ability OF THE TOKEN, and a token template has no
// static-ability slot and no oracle ID for the catalog to key one
// on. So the Synthesizer carries it on the tokens' behalf — a layer
// 7c static over the 0/0 Construct tokens its controller controls,
// counting that controller's artifacts, applied by the FIRST
// Synthesizer the controller controls only, so two Synthesizers
// size a Construct once (b15IsFirstOfItsNameControlledBy). The
// observable difference is what happens when the last Synthesizer
// leaves: the Constructs lose their sizing and shrink to 0/0, where
// printed Constructs keep it for good and are never smaller than
// 1/1. Weaker than printed, never stronger. (They shrink rather
// than die: the engine's toughness state-based action skips a
// printed 0/0 with no counters as a placeholder — the CurrentToughness
// convention — so a shrunken Construct lingers as a 0/0 body. The
// batch 15 test pins that so a change in the convention shows.)
func init() {
	Register(Spec{
		OracleID:     "eb7a1f21-a66d-415b-8520-710b44890bb6",
		Name:         "Simulacrum Synthesizer",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Constructs get their +1/+1 per artifact from the Synthesizer rather than on their own, so if it leaves the battlefield they shrink to 0/0."},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && IsToken(*target) && target.Name == "Construct" &&
					target.IsCreature() && target.Power == 0 && target.Toughness == 0 &&
					b15IsFirstOfItsNameControlledBy(g, source)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b15ArtifactsControlled(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Simulacrum Synthesizer — scry 2", Do(Scry{N: 2})),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15AnotherBigArtifactYouControlEntered(ev, source, g)
			}, "Simulacrum Synthesizer — create a Construct", Do(CreateToken{Template: TokenCard("0/0 colorless Construct artifact"), N: 1})),
		},
	})
}
