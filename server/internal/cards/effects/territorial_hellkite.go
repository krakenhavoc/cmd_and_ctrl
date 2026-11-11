package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Territorial Hellkite — Creature — Dragon {2}{R}{R}, 6/5 (Tarkir:
// Dragonstorm Commander):
//
//	"Flying, haste
//	 At the beginning of combat on your turn, choose an opponent at random
//	 that this creature didn't attack during your last combat. This
//	 creature attacks that player this combat if able. If you can't choose
//	 an opponent this way, tap this creature."
//
// "Your last combat" is the combat phase before this one (Player.
// LastCombatAttacks, rotated as each combat begins, so an extra combat
// reads the one before it and a turn with no combat changes nothing), and
// it counts only if THIS object attacked then (CR 400.7). An attack on a
// planeswalker or battle is not an attack on its controller, so that
// player stays a candidate. A creature that did not attack last combat
// (tapped, or not under your control then) leaves every opponent a
// candidate. The opponent is chosen from the keyed random stream among the
// opponents still in the game; players that can't be attacked can still be
// chosen (the printed ruling). The requirement is the engine's "attacks
// that player if able" (AttackPlayerIfAbleMod) until end of combat, so it
// binds the player, not a planeswalker, and gives way when the attack is
// impossible. With no candidate the Hellkite is tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1c772df4-bd88-4f7e-853c-5dd06f4cf70e",
		Name:            "Territorial Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Territorial Hellkite — choose an opponent at random that this creature didn't attack during your last combat; it attacks that player this combat if able, or tap it",
				territorialHellkiteChoose),
		},
	})
}

func territorialHellkiteChoose(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	self, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, item.SourceCardID) || ctx.isNewSourceObject(item.SourceCardID) {
		return nil
	}
	already := map[string]bool{}
	for _, id := range g.PlayersAttackedDuringLastCombatForEffect(&self) {
		already[id.String()] = true
	}
	var candidates []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if !already[opp.String()] {
			candidates = append(candidates, opp)
		}
	}
	if len(candidates) == 0 {
		return TapTarget{Target: item.SourceCardID}.Apply(ctx)
	}
	pick := g.ChooseAtRandomForEffect(randomDraw(ctx), candidates, 1)
	if len(pick) == 0 {
		return nil
	}
	dur, ok := DurationUntilEndOfCombat(ctx)
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target:   item.SourceCardID,
		Mods:     []game.Mod{game.AttackPlayerIfAbleMod(pick[0])},
		Duration: dur,
		Label:    "Territorial Hellkite — attacks that player this combat if able",
	}.Apply(ctx)
}
