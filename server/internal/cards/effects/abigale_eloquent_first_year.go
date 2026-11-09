package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abigale, Eloquent First-Year — Legendary Creature — Bird Bard {W/B}{W/B}
// (Lorwyn Eclipsed, #4458):
//
//	"Flying, first strike, lifelink
//	 When Abigale enters, up to one other target creature loses all
//	 abilities. Put a flying counter, a first strike counter, and a
//	 lifelink counter on that creature."
//
// The loss has no duration: it is a scoped effect pinned to the target
// object, lasting until that object leaves the battlefield (the ruling: it
// doesn't end at cleanup or when Abigale leaves). The counters go on after
// it, so their keywords are gained later and kept (the ruling).
func init() {
	Register(Spec{
		OracleID:        "48a773d0-c433-4463-b50f-b6d4604a042c",
		Name:            "Abigale, Eloquent First-Year",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Abigale — up to one other target creature loses all abilities; put a flying counter, a first strike counter and a lifelink counter on it",
					abigaleResolve),
				Another(TargetCreature("up to one other target creature").WithCount(0, 1))),
		},
	})
}

func abigaleResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.LoseAllAbilitiesMod()},
		Duration: g.PinnedTo(game.IndefiniteDuration(), id),
		Label:    "Abigale — loses all abilities",
	}).Apply(ctx); err != nil {
		return err
	}
	for _, kind := range []string{game.CounterFlying, game.CounterFirstStrike, game.CounterLifelink} {
		if err := (AddCounter{Target: id, Kind: kind, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
