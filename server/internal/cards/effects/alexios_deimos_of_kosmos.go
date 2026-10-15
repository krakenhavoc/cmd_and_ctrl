package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alexios, Deimos of Kosmos (#3275, ADR 0106 §2) — Legendary Creature —
// Human Berserker {3}{R}, 4/4:
//
//	"Trample
//	 Alexios attacks each combat if able, can't be sacrificed, and can't
//	 attack its owner.
//	 At the beginning of each player's upkeep, that player gains control
//	 of Alexios, untaps it, and puts a +1/+1 counter on it. It gains
//	 haste until end of turn."
//
// The three restrictions are the creature's own statics, so they go with
// its abilities (CR 613.1f): AttacksEachCombat (CR 508.1d),
// CantBeSacrificed (#3275: no cost, edict or delayed sacrifice can take
// it, CR 101.2) and CantAttackItsOwner (ADR 0106 §2, the owner read live,
// CR 108.3). The upkeep's player is read off the trigger's event; the
// control change has no stated duration, so it lasts until another one
// replaces it (CR 611.2a), and the next upkeep's trigger hands Alexios
// on. The haste is until cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "81c6c4f0-81ad-463c-a42b-60a73e6d70f3",
		Name:            "Alexios, Deimos of Kosmos",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{AttacksEachCombat(), CantBeSacrificed(), CantAttackItsOwner()},
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Alexios, Deimos of Kosmos — that player gains control of Alexios, untaps it and puts a +1/+1 counter on it; it gains haste",
				alexiosUpkeep),
		},
	})
}

func alexiosUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	alexios := item.SourceCardID
	if ctx.isNewSourceObject(alexios) || !onBattlefield(g, alexios) {
		return nil
	}
	if err := (GainControl{
		Target:     alexios,
		Controller: item.Trigger.Event.Actor,
		Duration:   game.IndefiniteDuration(),
		Label:      "Alexios, Deimos of Kosmos — gain control",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (UntapTarget{Target: alexios}).Apply(ctx); err != nil {
		return err
	}
	if err := (AddCounter{Target: alexios, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Target:   alexios,
		Keywords: []string{"haste"},
		Label:    "Alexios, Deimos of Kosmos — gains haste until end of turn",
	}.Apply(ctx)
}
