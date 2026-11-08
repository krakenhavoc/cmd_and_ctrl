package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Synthesizer Labship — Artifact — Spacecraft {U}, 4/4 (Edge of
// Eternities, #4461):
//
//	"Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 9+.)
//	 2+ | At the beginning of combat on your turn, up to one other target
//	 artifact you control becomes an artifact creature with base power
//	 and toughness 2/2 and gains flying until end of turn.
//	 9+ | Flying, vigilance"
//
// The 2+ line is D-7 Battle Cruiser's charge-gated beginning-of-combat
// trigger (TriggeredAbility.ActiveWhen, CR 721.2a). The animation is
// Living Brain's crew shape (BecomeCreatureUntilEOT: a layer 4 type
// change and a layer 7b base P/T set, pinned to the object, CR 400.7),
// so counters and other pumps still apply on top, with a flying grant
// for the same turn. "Other" leaves the Labship out of its own trigger.
// The 9+ line is a 4/4 artifact creature with flying and vigilance.
//
// No simplification.
func init() {
	combat := Targeting(AtBeginningOfYourCombat("Synthesizer Labship — up to one other target artifact you control becomes a 2/2 artifact creature with flying until end of turn",
		synthesizerLabshipAnimate),
		Another(TargetPermanent("up to one other target artifact you control", Artifact(), YouControl()).WithCount(0, 1)))
	combat.ActiveWhen = AtChargeCounters(2)
	Register(Spec{
		OracleID:     "a37314cf-bbbd-4003-b94d-0e6e60edfc15",
		Name:         "Synthesizer Labship",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{combat},
		Activated:    []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(9, 4, 4),
			ThresholdKeywords(9, "flying", "vigilance"),
		),
	})
}

// synthesizerLabshipAnimate animates each still-legal target and gives it
// flying until end of turn.
func synthesizerLabshipAnimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (BecomeCreatureUntilEOT{
			Target:       t.ID,
			SetPower:     2,
			SetToughness: 2,
			Label:        "Synthesizer Labship — becomes a 2/2 artifact creature until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"},
			Label: "Synthesizer Labship — flying until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
