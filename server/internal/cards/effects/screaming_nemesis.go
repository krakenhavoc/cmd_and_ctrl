package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Screaming Nemesis — Creature — Spirit {2}{R}, 3/3 (EDHREC rank
// 3726):
//
//	"Haste
//	 Whenever this creature is dealt damage, it deals that much damage
//	 to any other target. If a player is dealt damage this way, they
//	 can't gain life for the rest of the game."
//
// Red's Stuffy Doll with legs. Haste rides PrintedKeywords. The
// trigger is Wrathful Red Dragon's shape narrowed to the source
// itself: EventDealDamage names the damaged permanent in Target, and
// the Nemesis is a targeted ability — the controller picks any
// target as it goes on the stack and the Nemesis deals the event's
// amount to it on resolution. The SOURCE is the Nemesis, as printed,
// so its colour is what a prevention check sees; a Nemesis that
// died to the damage still deals it, because the damage path does
// not need the source on the battlefield. "Any OTHER target" is
// enforced at resolution (Aang, Swift Savior's posture — a target
// clause cannot name the instance it hangs off): a Nemesis chosen as
// its own target is skipped rather than damaged.
//
// Two declared simplifications, both weaker than printed:
//
//   - "They can't gain life for the rest of the game" is NOT
//     implemented. It is a permanent-duration player-scoped
//     replacement, and the engine has no registry for a continuous
//     effect that outlives its source and no per-player flag the
//     life-gain path consults (Sulfuric Vortex's gap). A player the
//     Nemesis hits gains life as normal.
//   - The engine emits one damage event per SOURCE, so a Nemesis
//     blocked by two creatures fires twice — once per blocker's
//     damage, each with its own target — where the printed card
//     fires once for the total (Wrathful Red Dragon's posture). The
//     same damage is reflected either way; only the split differs.
func init() {
	Register(Spec{
		OracleID:     "fcb7c93c-46ab-49b5-a6e0-35d73f3be8f0",
		Name:         "Screaming Nemesis",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A player it damages can still gain life — the \"can't gain life for the rest of the game\" clause isn't implemented.",
			"If two or more sources damage it at the same time, it reflects each source's damage separately instead of the total in one go.",
		},
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b35SelfWasDealtDamage(ev, source)
			},
			Targets: b35TargetAnyOther(),
			Key:     "Screaming Nemesis — deal that much damage to any other target",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Screaming Nemesis — deal that much damage to any other target")
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: b35RedirectDamageToChosen,
		}},
	})
}
