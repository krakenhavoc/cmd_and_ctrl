package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hot Pursuit — Enchantment {1}{R}:
//
//	"When this enchantment enters, suspect target creature an opponent
//	 controls. As long as this enchantment remains on the battlefield,
//	 that creature is also goaded.
//	 At the beginning of combat on your turn, if two or more players
//	 have lost the game, gain control of all goaded and/or suspected
//	 creatures until end of turn. Untap them. They gain haste until end
//	 of turn."
//
// The goad is a continuous effect from the resolved enters trigger
// (#2733, ADR 0071 amendment 2026-10-10): a record pinned to the
// creature as the object it is now, carrying goad's two requirements
// for you (game.GoadMod), for as long as this enchantment remains on
// the battlefield. It does not end at your next turn, as a goad from a
// spell does; it ends when Hot Pursuit leaves, or when the creature
// does (CR 400.7). It holds whether or not the suspect took (a creature
// that was already suspected, or can't become suspected, is still
// goaded), and a Hot Pursuit that left before its trigger resolved
// goads nothing, because the duration has already run out.
//
// The combat trigger has an intervening if (CR 603.4) on players who
// have lost the game, read as it would trigger and again as it
// resolves. "Goaded" is any goad, a spell's or a continuous one
// (Card.Goaded), so the creature this card goads is one of them even
// once it stops being suspected. A creature you already control is
// untapped and gains haste too.
//
// No simplification.
func init() {
	combat := AtBeginningOfYourCombat("Hot Pursuit — gain control of all goaded and suspected creatures", hotPursuitRoundUp)
	timing := combat.AppliesTo
	combat.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return timing(ev, source, ch, g) && twoOrMorePlayersHaveLost(g)
	}
	Register(Spec{
		OracleID:     "00bf9859-d5bf-455a-a4be-965ea4250c7e",
		Name:         "Hot Pursuit",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Hot Pursuit — suspect target creature an opponent controls; it's goaded while this remains", hotPursuitEnters),
				TargetCreature("target creature an opponent controls", OpponentControls()),
			),
			combat,
		},
	})
}

// hotPursuitEnters suspects the target and goads it for as long as Hot
// Pursuit remains on the battlefield.
func hotPursuitEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (Suspect{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		dur, ok := DurationWhileSourceRemains(ctx, ctx.Source())
		if !ok {
			continue
		}
		if err := (ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.GoadMod(ctx.Controller())},
			Duration: dur,
			Label:    "Hot Pursuit — goaded while Hot Pursuit remains",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// hotPursuitRoundUp is the combat trigger's resolution: every goaded
// or suspected creature is yours until end of turn, untapped, with
// haste.
func hotPursuitRoundUp(g *game.Game, item *game.StackItem) error {
	if !twoOrMorePlayersHaveLost(g) { // CR 603.4: re-checked on resolution
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && (c.Suspected || c.Goaded()) {
			ids = append(ids, c.InstanceID)
		}
	}
	ctx := NewContext(g, item)
	for _, id := range ids {
		if err := threatenAndGrant(ctx, id, "Hot Pursuit", "haste"); err != nil {
			return err
		}
	}
	return nil
}

// twoOrMorePlayersHaveLost is "if two or more players have lost the
// game": a player who conceded or lost to a state-based action or an
// effect has left the game (Player.Eliminated, ADR 0057).
func twoOrMorePlayersHaveLost(g *game.Game) bool {
	n := 0
	for _, p := range g.Seats {
		if p != nil && p.Eliminated {
			n++
		}
	}
	return n >= 2
}
