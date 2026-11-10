package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Opportunistic Dragon — Creature — Dragon {2}{R}{R}, 4/3 (Tarkir:
// Dragonstorm Commander):
//
//	"Flying
//	 When this creature enters, choose target Human or artifact an
//	 opponent controls. For as long as this creature remains on the
//	 battlefield, gain control of that permanent, it loses all
//	 abilities, and it can't attack or block."
//
// Kitesail Larcenist's shape: one record at one timestamp on the chosen
// permanent, pinned to the object it is when the ability resolves, lasting
// as long as the Dragon remains on the battlefield (DurationWhileSourceRemains).
// A Dragon that left before the trigger resolved changes nothing (the
// printed ruling). The record carries three mods: layer 2 gives you
// control, layer 6 removes every ability, and the can't-attack and
// can't-block restriction. The effect does not end if the Dragon changes
// controllers or loses its abilities, only when it is gone, and a stolen
// permanent that changes hands again stays stripped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "46f46bdd-52b8-436f-9a39-f5360a84b2cd",
		Name:            "Opportunistic Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Opportunistic Dragon — gain control of the target Human or artifact; it loses all abilities and can't attack or block",
				opportunisticDragonSteals),
				TargetPermanent("target Human or artifact an opponent controls",
					Or(HasSubtype("Human"), Artifact()), OpponentControls())),
		},
	})
}

func opportunisticDragonSteals(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	dur, ok := DurationWhileSourceRemains(ctx, ctx.Source())
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || !onBattlefield(g, t.ID) {
			continue
		}
		return ScopedEffectFor{
			Target: t.ID,
			Mods: []game.Mod{
				game.SetControllerMod(item.Controller),
				game.LoseAllAbilitiesMod(),
				game.AddRestrictionsMod(game.CantAttack | game.CantBlock),
			},
			Duration: dur,
			Label:    "Opportunistic Dragon — controlled, no abilities, can't attack or block",
		}.Apply(ctx)
	}
	return nil
}
