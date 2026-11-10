package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fear of Falling — Enchantment Creature — Nightmare {3}{U}{U}, 4/4
// (Duskmourn):
//
//	"Flying
//	 Whenever this creature attacks, target creature defending player
//	 controls gets -2/-0 and loses flying until your next turn."
//
// The target clause reads the defending player off the Fear's own attack
// (TargetCreatureDefendingPlayerControls). Both halves are one continuous
// effect from the resolving ability (CR 611.2a), pinned to that creature
// (CR 400.7), ending as its controller's next turn begins (CR 611.2b). Losing
// flying is "loses", not "doesn't have": another effect can still give it
// flying later, as the ruling says (the keyword removal is a layer-6 change
// that a later-timestamped grant overrides).
func init() {
	t := WheneverThisAttacks("Fear of Falling — target creature defending player controls gets -2/-0 and loses flying until your next turn",
		fearOfFallingShrink)
	t.TargetsFrom = TargetCreatureDefendingPlayerControls
	Register(Spec{
		OracleID:        "00ba96e0-42a7-432a-a500-5e607d75a358",
		Name:            "Fear of Falling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{t},
	})
}

// fearOfFallingShrink gives the target -2/-0 and takes its flying away
// until the Fear's controller's next turn.
func fearOfFallingShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.ModifyPTMod(-2, 0), game.RemoveKeywordsMod("flying")},
			Duration: DurationUntilYourNextTurn(ctx, item.Controller),
			Label:    "Fear of Falling — -2/-0 and loses flying until your next turn",
		}.Apply(ctx)
	}
	return nil
}
